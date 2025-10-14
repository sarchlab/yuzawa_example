package metrics_reporter

import (
	"sort"
	"strings"
	"sync"

	"github.com/sarchlab/akita/v4/datarecording"
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cu"
	"github.com/sarchlab/mgpusim/v4/amd/timing/rdma"
)

const (
	tableName = "mgpusim_metrics"
)

type metric struct {
	Location string
	What     string
	Value    float64
	Unit     string
}

type kernelTimeTracer struct {
	tracer *tracing.BusyTimeTracer
	comp   tracing.NamedHookable
}

type instCountTracer struct {
	tracer *instTracer
	cu     tracing.NamedHookable
}

type cacheLatencyTracer struct {
	tracer *tracing.AverageTimeTracer
	cache  tracing.NamedHookable
}

type cacheHitRateTracer struct {
	tracer *tracing.StepCountTracer
	cache  tracing.NamedHookable
}

type tlbHitRateTracer struct {
	tracer *tracing.StepCountTracer
	tlb    tracing.NamedHookable
}

type dramTransactionCountTracer struct {
	tracer *dramTracer
	dram   tracing.NamedHookable
}

type rdmaTransactionCountTracer struct {
	outgoingTracer *tracing.AverageTimeTracer
	incomingTracer *tracing.AverageTimeTracer
	rdmaEngine     *rdma.Comp
}

type simdBusyTimeTracer struct {
	tracer *tracing.BusyTimeTracer
	simd   tracing.NamedHookable
}

type cuCPIStackTracer struct {
	cu     tracing.NamedHookable
	tracer *cu.CPIStackTracer
}

type reporter struct {
	dataRecorder datarecording.DataRecorder

	kernelTimeTracer        *kernelTimeTracer
	perGPUKernelTimeTracers []*kernelTimeTracer
	instCountTracers        []*instCountTracer
	cacheLatencyTracers     []*cacheLatencyTracer
	cacheHitRateTracers     []*cacheHitRateTracer
	tlbHitRateTracers       []*tlbHitRateTracer
	dramTracers             []*dramTransactionCountTracer
	rdmaTransactionCounters []*rdmaTransactionCountTracer
	simdBusyTimeTracers     []*simdBusyTimeTracer
	cuCPITraces             []*cuCPIStackTracer

	ReportInstCount            bool
	ReportCacheLatency         bool
	ReportCacheHitRate         bool
	ReportTLBHitRate           bool
	ReportRDMATransactionCount bool
	ReportDRAMTransactionCount bool
	ReportSIMDBusyTime         bool
	ReportCPIStack             bool
}

func NewReporter(s *simulation.Simulation) *reporter {
	r := &reporter{
		dataRecorder: s.GetDataRecorder(),
		// Enable all metrics by default for comprehensive reporting
		ReportInstCount:            true,
		ReportCacheLatency:         true,
		ReportCacheHitRate:         true,
		ReportTLBHitRate:           true,
		ReportRDMATransactionCount: true,
		ReportDRAMTransactionCount: true,
		ReportSIMDBusyTime:         true,
		ReportCPIStack:             true,
	}

	r.injectTracers(s)

	r.dataRecorder.CreateTable(tableName, metric{})

	return r
}

func (r *reporter) injectTracers(s *simulation.Simulation) {
	r.injectKernelTimeTracer(s)
	r.injectInstCountTracer(s)
	r.injectCUCPIHook(s)
	r.injectCacheLatencyTracer(s)
	r.injectCacheHitRateTracer(s)
	r.injectTLBHitRateTracer(s)
	r.injectRDMAEngineTracer(s)
	r.injectDRAMTracer(s)
	r.injectSIMDBusyTimeTracer(s)
}

func (r *reporter) injectKernelTimeTracer(s *simulation.Simulation) {
	// Driver tracer for kernel launch commands
	driverComp := s.GetComponentByName("Driver")
	if driverComp != nil {
		tracer := tracing.NewBusyTimeTracer(
			s.GetEngine(),
			func(task tracing.Task) bool {
				return task.What == "*driver.LaunchKernelCommand"
			})
		tracing.CollectTrace(
			driverComp.(tracing.NamedHookable),
			tracer)
		r.kernelTimeTracer = &kernelTimeTracer{
			tracer: tracer,
			comp:   driverComp.(tracing.NamedHookable),
		}
	}

	// CommandProcessor tracers for kernel requests
	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "CommandProcessor") || strings.Contains(comp.Name(), "CP") {
			tracer := tracing.NewBusyTimeTracer(
				s.GetEngine(),
				func(task tracing.Task) bool {
					return task.What == "*protocol.LaunchKernelReq"
				})
			tracing.CollectTrace(
				comp.(tracing.NamedHookable),
				tracer)
			r.perGPUKernelTimeTracers = append(
				r.perGPUKernelTimeTracers,
				&kernelTimeTracer{
					tracer: tracer,
					comp:   comp.(tracing.NamedHookable),
				})
		}
	}
}

func (r *reporter) injectInstCountTracer(s *simulation.Simulation) {
	if !r.ReportInstCount {
		return
	}

	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "CU") {
			tracer := newInstTracer()
			r.instCountTracers = append(r.instCountTracers,
				&instCountTracer{
					tracer: tracer,
					cu:     comp.(tracing.NamedHookable),
				})
			tracing.CollectTrace(comp.(tracing.NamedHookable), tracer)
		}
	}
}

func (r *reporter) injectCUCPIHook(s *simulation.Simulation) {
	if !r.ReportCPIStack {
		return
	}

	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "CU") {
			if cuComp, ok := comp.(*cu.ComputeUnit); ok {
				tracer := cu.NewCPIStackInstHook(cuComp, s.GetEngine())
				tracing.CollectTrace(comp.(tracing.NamedHookable), tracer)

				r.cuCPITraces = append(r.cuCPITraces,
					&cuCPIStackTracer{
						tracer: tracer,
						cu:     comp.(tracing.NamedHookable),
					})
			}
		}
	}
}

func (r *reporter) injectCacheLatencyTracer(s *simulation.Simulation) {
	if !r.ReportCacheLatency {
		return
	}

	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "Cache") {
			tracer := tracing.NewAverageTimeTracer(
				s.GetEngine(),
				func(task tracing.Task) bool {
					return task.Kind == "req_in"
				})
			r.cacheLatencyTracers = append(r.cacheLatencyTracers,
				&cacheLatencyTracer{
					tracer: tracer,
					cache:  comp.(tracing.NamedHookable),
				})
			tracing.CollectTrace(comp.(tracing.NamedHookable), tracer)
		}
	}
}

func (r *reporter) injectCacheHitRateTracer(s *simulation.Simulation) {
	if !r.ReportCacheHitRate {
		return
	}

	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "Cache") {
			tracer := tracing.NewStepCountTracer(
				func(task tracing.Task) bool { return true })
			r.cacheHitRateTracers = append(r.cacheHitRateTracers,
				&cacheHitRateTracer{
					tracer: tracer,
					cache:  comp.(tracing.NamedHookable),
				})
			tracing.CollectTrace(comp.(tracing.NamedHookable), tracer)
		}
	}
}

func (r *reporter) injectTLBHitRateTracer(s *simulation.Simulation) {
	if !r.ReportTLBHitRate {
		return
	}

	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "TLB") {
			tracer := tracing.NewStepCountTracer(
				func(task tracing.Task) bool { return true })
			r.tlbHitRateTracers = append(r.tlbHitRateTracers,
				&tlbHitRateTracer{
					tracer: tracer,
					tlb:    comp.(tracing.NamedHookable),
				})
			tracing.CollectTrace(comp.(tracing.NamedHookable), tracer)
		}
	}
}

func (r *reporter) injectRDMAEngineTracer(s *simulation.Simulation) {
	if !r.ReportRDMATransactionCount {
		return
	}

	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "RDMA") {
			if rdmaComp, ok := comp.(*rdma.Comp); ok {
				t := &rdmaTransactionCountTracer{}
				t.rdmaEngine = rdmaComp
				t.incomingTracer = tracing.NewAverageTimeTracer(
					s.GetEngine(),
					func(task tracing.Task) bool {
						if task.Kind != "req_in" {
							return false
						}

						isFromOutside := strings.Contains(
							string(task.Detail.(sim.Msg).Meta().Src), "RDMA")
						if !isFromOutside {
							return false
						}

						return true
					})
				t.outgoingTracer = tracing.NewAverageTimeTracer(
					s.GetEngine(),
					func(task tracing.Task) bool {
						if task.Kind != "req_in" {
							return false
						}

						isFromOutside := strings.Contains(
							string(task.Detail.(sim.Msg).Meta().Src), "RDMA")
						if isFromOutside {
							return false
						}

						return true
					})

				tracing.CollectTrace(t.rdmaEngine, t.incomingTracer)
				tracing.CollectTrace(t.rdmaEngine, t.outgoingTracer)

				r.rdmaTransactionCounters = append(r.rdmaTransactionCounters, t)
			}
		}
	}
}

func (r *reporter) injectDRAMTracer(s *simulation.Simulation) {
	if !r.ReportDRAMTransactionCount {
		return
	}

	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "DRAM") {
			t := &dramTransactionCountTracer{}
			t.dram = comp.(tracing.NamedHookable)
			t.tracer = newDramTracer(s.GetEngine())

			tracing.CollectTrace(t.dram, t.tracer)

			r.dramTracers = append(r.dramTracers, t)
		}
	}
}

func (r *reporter) injectSIMDBusyTimeTracer(s *simulation.Simulation) {
	if !r.ReportSIMDBusyTime {
		return
	}

	for _, comp := range s.Components() {
		if strings.Contains(comp.Name(), "SIMD") {
			perSIMDBusyTimeTracer := tracing.NewBusyTimeTracer(
				s.GetEngine(),
				func(task tracing.Task) bool {
					return task.Kind == "pipeline"
				})
			r.simdBusyTimeTracers = append(r.simdBusyTimeTracers,
				&simdBusyTimeTracer{
					tracer: perSIMDBusyTimeTracer,
					simd:   comp.(tracing.NamedHookable),
				})
			tracing.CollectTrace(comp.(tracing.NamedHookable), perSIMDBusyTimeTracer)
		}
	}
}

func (r *reporter) Report() {
	r.reportKernelTime()
	r.reportInstCount()
	r.reportCPIStack()
	r.reportSIMDBusyTime()
	r.reportCacheLatency()
	r.reportCacheHitRate()
	r.reportTLBHitRate()
	r.reportRDMATransactionCount()
	r.reportDRAMTransactionCount()
}

func (r *reporter) reportKernelTime() {
	// Report Driver kernel time
	if r.kernelTimeTracer != nil {
		kernelTime := float64(r.kernelTimeTracer.tracer.BusyTime())
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: r.kernelTimeTracer.comp.Name(),
				What:     "kernel_time",
				Value:    kernelTime,
				Unit:     "second",
			},
		)
	}

	// Report CommandProcessor kernel times
	for _, t := range r.perGPUKernelTimeTracers {
		kernelTime := float64(t.tracer.BusyTime())
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.comp.Name(),
				What:     "kernel_time",
				Value:    kernelTime,
				Unit:     "second",
			},
		)
	}
}

func (r *reporter) reportInstCount() {
	if r.kernelTimeTracer == nil {
		return
	}

	kernelTime := float64(r.kernelTimeTracer.tracer.BusyTime())
	for _, t := range r.instCountTracers {
		if cuComp, ok := t.cu.(*cu.ComputeUnit); ok {
			cuFreq := float64(cuComp.Freq)
			numCycle := kernelTime * cuFreq

			r.dataRecorder.InsertData(
				tableName,
				metric{
					Location: t.cu.Name(),
					What:     "cu_inst_count",
					Value:    float64(t.tracer.count),
					Unit:     "count",
				},
			)

			if t.tracer.count > 0 {
				r.dataRecorder.InsertData(
					tableName,
					metric{
						Location: t.cu.Name(),
						What:     "cu_CPI",
						Value:    numCycle / float64(t.tracer.count),
						Unit:     "cycles/inst",
					},
				)
			}

			r.dataRecorder.InsertData(
				tableName,
				metric{
					Location: t.cu.Name(),
					What:     "simd_inst_count",
					Value:    float64(t.tracer.simdCount),
					Unit:     "count",
				},
			)

			if t.tracer.simdCount > 0 {
				r.dataRecorder.InsertData(
					tableName,
					metric{
						Location: t.cu.Name(),
						What:     "simd_CPI",
						Value:    numCycle / float64(t.tracer.simdCount),
						Unit:     "cycles/inst",
					},
				)
			}
		}
	}
}

func (r *reporter) reportCPIStack() {
	for _, t := range r.cuCPITraces {
		cu := t.cu
		hook := t.tracer

		r.reportCPIStackEntries(hook, cu, false)
		r.reportCPIStackEntries(hook, cu, true)
	}
}

func (r *reporter) reportCPIStackEntries(
	hook *cu.CPIStackTracer,
	cu tracing.NamedHookable,
	simdStack bool,
) {
	cpiStack := hook.GetCPIStack()
	if simdStack {
		cpiStack = hook.GetSIMDCPIStack()
	}

	keys := make([]string, 0, len(cpiStack))
	for k := range cpiStack {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	stackTypeName := "CPIStack"
	if simdStack {
		stackTypeName = "SIMDCPIStack"
	}

	for _, name := range keys {
		value := cpiStack[name]
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: cu.Name(),
				What:     stackTypeName + "." + name,
				Value:    value,
				Unit:     "cycles/inst",
			},
		)
	}
}

func (r *reporter) reportSIMDBusyTime() {
	for _, t := range r.simdBusyTimeTracers {
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.simd.Name(),
				What:     "busy_time",
				Value:    float64(t.tracer.BusyTime()),
				Unit:     "second",
			},
		)
	}
}

func (r *reporter) reportCacheLatency() {
	for _, tracer := range r.cacheLatencyTracers {
		if tracer.tracer.AverageTime() == 0 {
			continue
		}

		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: tracer.cache.Name(),
				What:     "req_average_latency",
				Value:    float64(tracer.tracer.AverageTime()),
				Unit:     "second",
			},
		)
	}
}

func (r *reporter) reportCacheHitRate() {
	for _, tracer := range r.cacheHitRateTracers {
		readHit := tracer.tracer.GetStepCount("read-hit")
		readMiss := tracer.tracer.GetStepCount("read-miss")
		readMSHRHit := tracer.tracer.GetStepCount("read-mshr-hit")
		writeHit := tracer.tracer.GetStepCount("write-hit")
		writeMiss := tracer.tracer.GetStepCount("write-miss")
		writeMSHRHit := tracer.tracer.GetStepCount("write-mshr-hit")

		totalTransaction := readHit + readMiss + readMSHRHit +
			writeHit + writeMiss + writeMSHRHit

		if totalTransaction == 0 {
			continue
		}

		r.dataRecorder.InsertData(tableName, metric{
			Location: tracer.cache.Name(),
			What:     "read-hit",
			Value:    float64(readHit),
			Unit:     "count",
		})
		r.dataRecorder.InsertData(tableName, metric{
			Location: tracer.cache.Name(),
			What:     "read-miss",
			Value:    float64(readMiss),
			Unit:     "count",
		})
		r.dataRecorder.InsertData(tableName, metric{
			Location: tracer.cache.Name(),
			What:     "read-mshr-hit",
			Value:    float64(readMSHRHit),
			Unit:     "count",
		})
		r.dataRecorder.InsertData(tableName, metric{
			Location: tracer.cache.Name(),
			What:     "write-hit",
			Value:    float64(writeHit),
			Unit:     "count",
		})
		r.dataRecorder.InsertData(tableName, metric{
			Location: tracer.cache.Name(),
			What:     "write-miss",
			Value:    float64(writeMiss),
			Unit:     "count",
		})
		r.dataRecorder.InsertData(tableName, metric{
			Location: tracer.cache.Name(),
			What:     "write-mshr-hit",
			Value:    float64(writeMSHRHit),
			Unit:     "count",
		})
	}
}

func (r *reporter) reportTLBHitRate() {
	for _, tracer := range r.tlbHitRateTracers {
		hit := tracer.tracer.GetStepCount("hit")
		miss := tracer.tracer.GetStepCount("miss")
		mshrHit := tracer.tracer.GetStepCount("mshr-hit")

		totalTransaction := hit + miss + mshrHit

		if totalTransaction == 0 {
			continue
		}

		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: tracer.tlb.Name(),
				What:     "hit",
				Value:    float64(hit),
				Unit:     "count",
			},
		)
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: tracer.tlb.Name(),
				What:     "miss",
				Value:    float64(miss),
				Unit:     "count",
			},
		)
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: tracer.tlb.Name(),
				What:     "mshr-hit",
				Value:    float64(mshrHit),
				Unit:     "count",
			},
		)
	}
}

func (r *reporter) reportRDMATransactionCount() {
	for _, t := range r.rdmaTransactionCounters {
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.rdmaEngine.Name(),
				What:     "outgoing_trans_count",
				Value:    float64(t.outgoingTracer.TotalCount()),
				Unit:     "count",
			},
		)
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.rdmaEngine.Name(),
				What:     "incoming_trans_count",
				Value:    float64(t.incomingTracer.TotalCount()),
				Unit:     "count",
			},
		)
	}
}

func (r *reporter) reportDRAMTransactionCount() {
	for _, t := range r.dramTracers {
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.dram.Name(),
				What:     "read_trans_count",
				Value:    float64(t.tracer.readCount),
				Unit:     "count",
			},
		)
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.dram.Name(),
				What:     "write_trans_count",
				Value:    float64(t.tracer.writeCount),
				Unit:     "count",
			},
		)
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.dram.Name(),
				What:     "read_avg_latency",
				Value:    float64(t.tracer.readAvgLatency),
				Unit:     "second",
			},
		)
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.dram.Name(),
				What:     "write_avg_latency",
				Value:    float64(t.tracer.writeAvgLatency),
				Unit:     "second",
			},
		)
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.dram.Name(),
				What:     "read_size",
				Value:    float64(t.tracer.readSize),
				Unit:     "bytes",
			},
		)
		r.dataRecorder.InsertData(
			tableName,
			metric{
				Location: t.dram.Name(),
				What:     "write_size",
				Value:    float64(t.tracer.writeSize),
				Unit:     "bytes",
			},
		)
	}
}

// instTracer can trace the number of instruction completed.
type instTracer struct {
	count     uint64
	simdInst  bool
	simdCount uint64
	maxCount  uint64

	inflightInst map[string]tracing.Task
}

// newInstTracer creates a tracer that can count the number of instructions.
func newInstTracer() *instTracer {
	t := &instTracer{
		inflightInst: map[string]tracing.Task{},
	}
	return t
}

func (t *instTracer) StartTask(task tracing.Task) {
	if task.Kind != "inst" {
		return
	}

	if task.What == "VALU" {
		t.simdInst = true
	} else {
		t.simdInst = false
	}

	t.inflightInst[task.ID] = task
}

func (t *instTracer) StepTask(task tracing.Task) {
	// Do nothing
}

func (t *instTracer) AddMilestone(milestone tracing.Milestone) {
	// Do nothing
}

func (t *instTracer) EndTask(task tracing.Task) {
	_, found := t.inflightInst[task.ID]
	if !found {
		return
	}

	if t.simdInst {
		t.simdCount++
	}

	delete(t.inflightInst, task.ID)

	t.count++
}

// dramTracer can trace DRAM activities.
type dramTracer struct {
	sync.Mutex
	sim.TimeTeller

	inflightTasks map[string]tracing.Task

	readCount       int
	writeCount      int
	readAvgLatency  sim.VTimeInSec
	writeAvgLatency sim.VTimeInSec
	readSize        uint64
	writeSize       uint64
}

func newDramTracer(timeTeller sim.TimeTeller) *dramTracer {
	return &dramTracer{
		TimeTeller:    timeTeller,
		inflightTasks: make(map[string]tracing.Task),
	}
}

// StartTask records the task start time
func (t *dramTracer) StartTask(task tracing.Task) {
	t.Lock()
	defer t.Unlock()

	task.StartTime = t.TimeTeller.CurrentTime()

	t.inflightTasks[task.ID] = task
}

// StepTask does nothing
func (t *dramTracer) StepTask(task tracing.Task) {
	// Do nothing
}

// AddMilestone does nothing
func (t *dramTracer) AddMilestone(milestone tracing.Milestone) {
	// Do nothing
}

// EndTask records the end of the task
func (t *dramTracer) EndTask(task tracing.Task) {
	t.Lock()
	defer t.Unlock()

	originalTask, ok := t.inflightTasks[task.ID]
	if !ok {
		return
	}

	task.EndTime = t.TimeTeller.CurrentTime()
	taskTime := task.EndTime - originalTask.StartTime

	switch originalTask.What {
	case "*mem.ReadReq":
		t.readAvgLatency = sim.VTimeInSec(
			(float64(t.readAvgLatency)*float64(t.readCount) +
				float64(taskTime)) / float64(t.readCount+1))
		t.readCount++
		t.readSize += originalTask.Detail.(*mem.ReadReq).AccessByteSize
	case "*mem.WriteReq":
		t.writeAvgLatency = sim.VTimeInSec(
			(float64(t.writeAvgLatency)*float64(t.writeCount) +
				float64(taskTime)) / float64(t.writeCount+1))
		t.writeCount++
		t.writeSize += uint64(len(originalTask.Detail.(*mem.WriteReq).Data))
	}

	delete(t.inflightTasks, task.ID)
}
