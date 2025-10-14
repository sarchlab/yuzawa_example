package multi_stage_memory

import (
	"fmt"

	"github.com/sarchlab/akita/v4/mem/vm"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/yuzawa_example/components/memaccessagent"
	"github.com/sarchlab/yuzawa_example/components/metrics_reporter"
)

type Benchmark struct {
	name              string
	simulation        *simulation.Simulation
	numAccess         int
	maxAddress        uint64
	pageTable         *vm.PageTable
	deviceID          uint64
	useVirtualAddress bool

	ioMMUName string
}

func (b *Benchmark) Run() {
	// Set up metrics reporter
	metricsReporter := metrics_reporter.NewReporter(b.simulation)

	engine := b.simulation.GetEngine()
	agent := b.simulation.GetComponentByName("MemAgent").(*memaccessagent.MemAccessAgent)

	// Map pages if using virtual addressing
	if b.useVirtualAddress && b.pageTable != nil {
		pageSize := uint64(1 << 12) // 4KB pages (2^12)
		for vAddr := uint64(0); vAddr < b.maxAddress; vAddr += pageSize {
			pAddr := vAddr // Identity mapping for simplicity
			b.pageTable.Update(vm.Page{
				PID:      1,
				VAddr:    vAddr,
				PAddr:    pAddr,
				PageSize: pageSize,
				Valid:    true,
				DeviceID: b.deviceID,
			})
		}
	}

	agent.WriteLeft = b.numAccess
	agent.ReadLeft = b.numAccess
	agent.MaxAddress = b.maxAddress
	agent.UseVirtualAddress = b.useVirtualAddress
	if b.useVirtualAddress {
		agent.PID = 1
	}

	agent.TickLater()
	err := engine.Run()
	if err != nil {
		panic(err)
	}

	if len(agent.PendingWriteReq) > 0 || len(agent.PendingReadReq) > 0 {
		panic(fmt.Errorf("there are still pending requests"))
	}

	if agent.WriteLeft > 0 || agent.ReadLeft > 0 {
		panic(fmt.Errorf("there are still requests left"))
	}

	fmt.Printf("End time: %.10f seconds\n",
		engine.CurrentTime())

	// Report metrics before completing
	metricsReporter.Report()
}
