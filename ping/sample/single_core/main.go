package main

import (
	"log"
	"os"

	"github.com/sarchlab/akita/v4/mem/cache/writeback"
	"github.com/sarchlab/akita/v4/mem/cache/writethrough"
	"github.com/sarchlab/akita/v4/mem/idealmemcontroller"
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/mem/trace"
	"github.com/sarchlab/akita/v4/mem/vm"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"

	"github.com/sarchlab/akita/v4/mem/vm/addresstranslator"
	"github.com/sarchlab/akita/v4/mem/vm/tlb"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/tracing"

	"github.com/sarchlab/yuzawa_example/ping/benchmarks/relu"
	"github.com/sarchlab/yuzawa_example/ping/mmu"
	"github.com/sarchlab/yuzawa_example/ping/rob"

	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cu"
)

// --- Adapter so *cu.ComputeUnit satisfies cp.CUInterfaceForCP ---
type cuForCP struct{ inner *cu.ComputeUnit }

// CP expects RemotePort; CU exposes a local Port → convert with AsRemote().
func (w *cuForCP) ControlPort() sim.RemotePort     { return w.inner.GetPortByName("Ctrl").AsRemote() }
func (w *cuForCP) DispatchingPort() sim.RemotePort { return w.inner.GetPortByName("Ctrl").AsRemote() }

// Resource queries (reasonable defaults; tune if your CU builder exposes getters)
func (w *cuForCP) LDSBytes() int  { return 64 * 1024 } // 64 KiB LDS
func (w *cuForCP) SRegCount() int { return 1024 }      // scalar registers
func (w *cuForCP) VRegCounts() []int {
	// one entry per SIMD/cluster; single entry is fine for single-SIMD configs
	return []int{65536}
}
func (w *cuForCP) WfPoolSizes() []int {
	// Wavefront pool sizes per SIMD unit
	return []int{40} // typical wavefront pool size
}

func main() {
	// Build simulation & engine
	s := simulation.MakeBuilder().Build()
	engine := s.GetEngine()

	// -------------------------
	// Memory controllers + L2 caches (separate V/S/I paths)
	// -------------------------
	// Create shared storage for all memory controllers
	sharedStorage := mem.NewStorage(4 * mem.GB)

	MemCtrlV := idealmemcontroller.MakeBuilder().
		WithEngine(engine).WithStorage(sharedStorage).WithLatency(100).
		Build("MemCtrlV")
	s.RegisterComponent(MemCtrlV)

	MemCtrlS := idealmemcontroller.MakeBuilder().
		WithEngine(engine).WithStorage(sharedStorage).WithLatency(100).
		Build("MemCtrlS")
	s.RegisterComponent(MemCtrlS)

	MemCtrlI := idealmemcontroller.MakeBuilder().
		WithEngine(engine).WithStorage(sharedStorage).WithLatency(100).
		Build("MemCtrlI")
	s.RegisterComponent(MemCtrlI)

	L2VCache := writeback.MakeBuilder().
		WithEngine(engine).WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).WithNumReqPerCycle(8).
		WithAddressMapperType("single").
		WithRemotePorts(MemCtrlV.GetPortByName("Top").AsRemote()).
		Build("L2VCache")
	s.RegisterComponent(L2VCache)

	L2SCache := writeback.MakeBuilder().
		WithEngine(engine).WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).WithNumReqPerCycle(8).
		WithAddressMapperType("single").
		WithRemotePorts(MemCtrlS.GetPortByName("Top").AsRemote()).
		Build("L2SCache")
	s.RegisterComponent(L2SCache)

	L2ICache := writeback.MakeBuilder().
		WithEngine(engine).WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).WithNumReqPerCycle(8).
		WithAddressMapperType("single").
		WithRemotePorts(MemCtrlI.GetPortByName("Top").AsRemote()).
		Build("L2ICache")
	s.RegisterComponent(L2ICache)

	L1VCache := writethrough.MakeBuilder().
		WithEngine(engine).WithFreq(1 * sim.GHz).WithWayAssociativity(2).
		WithAddressMapperType("single").
		WithRemotePorts(L2VCache.GetPortByName("Top").AsRemote()).
		Build("L1VCache")
	s.RegisterComponent(L1VCache)

	L1SCache := writethrough.MakeBuilder().
		WithEngine(engine).WithFreq(1 * sim.GHz).WithWayAssociativity(2).
		WithAddressMapperType("single").
		WithRemotePorts(L2SCache.GetPortByName("Top").AsRemote()).
		Build("L1SCache")
	s.RegisterComponent(L1SCache)

	L1ICache := writethrough.MakeBuilder().
		WithEngine(engine).WithFreq(1 * sim.GHz).WithWayAssociativity(2).
		WithAddressMapperType("single").
		WithRemotePorts(L2ICache.GetPortByName("Top").AsRemote()).
		Build("L1ICache")
	s.RegisterComponent(L1ICache)

	// -------------------------
	// IoMMUs + L2TLBs + L1 TLBs + ATs (per path)
	// -------------------------
	IoMMUV := mmu.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).WithMaxNumReqInFlight(16).WithPageWalkingLatency(10).
		Build("IoMMUV")
	s.RegisterComponent(IoMMUV)

	IoMMUS := mmu.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).WithMaxNumReqInFlight(16).WithPageWalkingLatency(10).
		Build("IoMMUS")
	s.RegisterComponent(IoMMUS)

	IoMMUI := mmu.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).WithMaxNumReqInFlight(16).WithPageWalkingLatency(10).
		Build("IoMMUI")
	s.RegisterComponent(IoMMUI)

	L2VTLB := tlb.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumWays(64).WithNumSets(64).WithPageSize(4096).WithNumReqPerCycle(8).
		WithRemotePorts(IoMMUV.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("L2VTLB")
	s.RegisterComponent(L2VTLB)

	L2STLB := tlb.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumWays(64).WithNumSets(64).WithPageSize(4096).WithNumReqPerCycle(8).
		WithRemotePorts(IoMMUS.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("L2STLB")
	s.RegisterComponent(L2STLB)

	L2ITLB := tlb.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumWays(64).WithNumSets(64).WithPageSize(4096).WithNumReqPerCycle(8).
		WithRemotePorts(IoMMUI.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("L2ITLB")
	s.RegisterComponent(L2ITLB)

	VTLB := tlb.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumWays(8).WithNumSets(8).WithPageSize(4096).WithNumReqPerCycle(2).
		WithRemotePorts(L2VTLB.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("VTLB")
	s.RegisterComponent(VTLB)

	STLB := tlb.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumWays(8).WithNumSets(8).WithPageSize(4096).WithNumReqPerCycle(2).
		WithRemotePorts(L2STLB.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("STLB")
	s.RegisterComponent(STLB)

	ITLB := tlb.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumWays(8).WithNumSets(8).WithPageSize(4096).WithNumReqPerCycle(2).
		WithRemotePorts(L2ITLB.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("ITLB")
	s.RegisterComponent(ITLB)

	VAT := addresstranslator.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithTranslationProvider(VTLB.GetPortByName("Top").AsRemote()).
		WithRemotePorts(L1VCache.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("VAT")
	s.RegisterComponent(VAT)

	SAT := addresstranslator.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithTranslationProvider(STLB.GetPortByName("Top").AsRemote()).
		WithRemotePorts(L1SCache.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("SAT")
	s.RegisterComponent(SAT)

	IAT := addresstranslator.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithTranslationProvider(ITLB.GetPortByName("Top").AsRemote()).
		WithRemotePorts(L1ICache.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("IAT")
	s.RegisterComponent(IAT)

	// -------------------------
	// ROBs
	// -------------------------
	VROB := rob.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).WithBufferSize(128).
		WithBottomUnit(VAT.GetPortByName("Top")). // issue into VAT
		Build("VROB")
	s.RegisterComponent(VROB)

	SROB := rob.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).WithBufferSize(128).
		WithBottomUnit(SAT.GetPortByName("Top")).
		Build("SROB")
	s.RegisterComponent(SROB)

	IROB := rob.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).WithBufferSize(128).
		WithBottomUnit(IAT.GetPortByName("Top")).
		Build("IROB")
	s.RegisterComponent(IROB)

	// -------------------------
	// GPU front-end: CP + CU + Driver
	// -------------------------
	CUBuild := cu.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).
		WithInstMem(IROB.GetPortByName("Top")).
		WithVGPRCount([]int{32768, 32768, 32768, 32768})
	CU := CUBuild.Build("CU")
	s.RegisterComponent(CU)

	CP := cp.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("CP")
	s.RegisterComponent(CP)

	// Tell CP about CU (via adapter)
	CP.RegisterCU(&cuForCP{inner: CU})

	pt := vm.NewPageTable(12)
	Driver := driver.MakeBuilder().
		WithEngine(engine).WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).WithPageTable(pt).
		WithGlobalStorage(sharedStorage).
		WithMagicMemoryCopyMiddleware().
		Build("Driver")
	s.RegisterComponent(Driver)

	Driver.RegisterGPU(
		CP.ToDriver,
		driver.DeviceProperties{CUCount: 1, DRAMSize: 4 * mem.GB},
	)

	// -------------------------
	// Wiring
	// -------------------------
	// Driver <-> CP
	Conn1 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn1")
	Conn1.PlugIn(Driver.GetPortByName("GPU"))
	Conn1.PlugIn(CP.ToDriver)

	// CP <-> CU control
	Conn2 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn2")
	Conn2.PlugIn(CP.ToCUs)
	Conn2.PlugIn(CU.GetPortByName("Ctrl"))

	// CU <-> ROBs
	Conn3 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn3")
	Conn3.PlugIn(CU.GetPortByName("VectorMem"))
	Conn3.PlugIn(VROB.GetPortByName("Top"))

	Conn11 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn11")
	Conn11.PlugIn(CU.GetPortByName("ScalarMem"))
	Conn11.PlugIn(SROB.GetPortByName("Top"))

	Conn15 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn15")
	Conn15.PlugIn(CU.GetPortByName("InstMem"))
	Conn15.PlugIn(IROB.GetPortByName("Top"))

	// ROBs -> ATs
	Conn4 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn4")
	Conn4.PlugIn(VROB.GetPortByName("Bottom"))
	Conn4.PlugIn(VAT.GetPortByName("Top"))

	Conn12 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn12")
	Conn12.PlugIn(SROB.GetPortByName("Bottom"))
	Conn12.PlugIn(SAT.GetPortByName("Top"))

	Conn16 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn16")
	Conn16.PlugIn(IROB.GetPortByName("Bottom"))
	Conn16.PlugIn(IAT.GetPortByName("Top"))

	// ATs -> L1 TLBs (translation reqs)
	Conn5 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn5")
	Conn5.PlugIn(VAT.GetPortByName("Translation"))
	Conn5.PlugIn(VTLB.GetPortByName("Top"))

	Conn13 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn13")
	Conn13.PlugIn(SAT.GetPortByName("Translation"))
	Conn13.PlugIn(STLB.GetPortByName("Top"))

	Conn17 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn17")
	Conn17.PlugIn(IAT.GetPortByName("Translation"))
	Conn17.PlugIn(ITLB.GetPortByName("Top"))

	// Close TLB chains: L1 bottoms -> L2 tops; L2 bottoms -> IoMMUs
	Conn6 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn6")
	Conn6.PlugIn(VTLB.GetPortByName("Bottom"))
	Conn6.PlugIn(L2VTLB.GetPortByName("Top"))

	Conn19a := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn19a")
	Conn19a.PlugIn(STLB.GetPortByName("Bottom"))
	Conn19a.PlugIn(L2STLB.GetPortByName("Top"))

	Conn19b := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn19b")
	Conn19b.PlugIn(ITLB.GetPortByName("Bottom"))
	Conn19b.PlugIn(L2ITLB.GetPortByName("Top"))

	Conn7a := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn7a")
	Conn7a.PlugIn(L2VTLB.GetPortByName("Bottom"))
	Conn7a.PlugIn(IoMMUV.GetPortByName("Top"))

	Conn7b := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn7b")
	Conn7b.PlugIn(L2STLB.GetPortByName("Bottom"))
	Conn7b.PlugIn(IoMMUS.GetPortByName("Top"))

	Conn7c := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn7c")
	Conn7c.PlugIn(L2ITLB.GetPortByName("Bottom"))
	Conn7c.PlugIn(IoMMUI.GetPortByName("Top"))

	// ATs -> L1 caches -> L2 -> MemCtrl (data path)
	Conn8 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn8")
	Conn8.PlugIn(VAT.GetPortByName("Bottom"))
	Conn8.PlugIn(L1VCache.GetPortByName("Top"))

	Conn14 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn14")
	Conn14.PlugIn(SAT.GetPortByName("Bottom"))
	Conn14.PlugIn(L1SCache.GetPortByName("Top"))

	Conn18 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn18")
	Conn18.PlugIn(IAT.GetPortByName("Bottom"))
	Conn18.PlugIn(L1ICache.GetPortByName("Top"))

	Conn9 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn9")
	Conn9.PlugIn(L1VCache.GetPortByName("Bottom"))
	Conn9.PlugIn(L2VCache.GetPortByName("Top"))

	Conn9b := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn9b")
	Conn9b.PlugIn(L1SCache.GetPortByName("Bottom"))
	Conn9b.PlugIn(L2SCache.GetPortByName("Top"))

	Conn9c := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn9c")
	Conn9c.PlugIn(L1ICache.GetPortByName("Bottom"))
	Conn9c.PlugIn(L2ICache.GetPortByName("Top"))

	Conn10a := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn10a")
	Conn10a.PlugIn(L2VCache.GetPortByName("Bottom"))
	Conn10a.PlugIn(MemCtrlV.GetPortByName("Top"))

	Conn10b := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn10b")
	Conn10b.PlugIn(L2SCache.GetPortByName("Bottom"))
	Conn10b.PlugIn(MemCtrlS.GetPortByName("Top"))

	Conn10c := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn10c")
	Conn10c.PlugIn(L2ICache.GetPortByName("Bottom"))
	Conn10c.PlugIn(MemCtrlI.GetPortByName("Top"))

	// -------------------------
	// CP control fabrics (important so CP can configure TLBs/ATs/caches)
	// -------------------------
	ConnCTL_TLB := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("ConnCTLTLB")
	ConnCTL_TLB.PlugIn(CP.ToTLBs)
	ConnCTL_TLB.PlugIn(VTLB.GetPortByName("Control"))
	ConnCTL_TLB.PlugIn(STLB.GetPortByName("Control"))
	ConnCTL_TLB.PlugIn(ITLB.GetPortByName("Control"))

	ConnCTL_AT := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("ConnCTLAT")
	ConnCTL_AT.PlugIn(CP.ToAddressTranslators)
	ConnCTL_AT.PlugIn(VAT.GetPortByName("Control"))
	ConnCTL_AT.PlugIn(SAT.GetPortByName("Control"))
	ConnCTL_AT.PlugIn(IAT.GetPortByName("Control"))

	ConnCTL_Cache := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("ConnCTLCache")
	ConnCTL_Cache.PlugIn(CP.ToCaches)
	ConnCTL_Cache.PlugIn(L1VCache.GetPortByName("Control"))
	ConnCTL_Cache.PlugIn(L1SCache.GetPortByName("Control"))
	ConnCTL_Cache.PlugIn(L1ICache.GetPortByName("Control"))
	ConnCTL_Cache.PlugIn(L2VCache.GetPortByName("Control"))
	ConnCTL_Cache.PlugIn(L2SCache.GetPortByName("Control"))
	ConnCTL_Cache.PlugIn(L2ICache.GetPortByName("Control"))

	// -------------------------
	// Tracing (helpful to confirm activity)
	// -------------------------
	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	defer traceFile.Close()
	logger := log.New(traceFile, "", 0)
	memTracer := trace.NewTracer(logger, engine)

	tracing.CollectTrace(MemCtrlV, memTracer)
	tracing.CollectTrace(MemCtrlS, memTracer)
	tracing.CollectTrace(MemCtrlI, memTracer)
	tracing.CollectTrace(L2VCache, memTracer)
	tracing.CollectTrace(L2SCache, memTracer)
	tracing.CollectTrace(L2ICache, memTracer)
	tracing.CollectTrace(L1VCache, memTracer)
	tracing.CollectTrace(L1SCache, memTracer)
	tracing.CollectTrace(L1ICache, memTracer)
	tracing.CollectTrace(IoMMUV, memTracer)
	tracing.CollectTrace(IoMMUS, memTracer)
	tracing.CollectTrace(IoMMUI, memTracer)

	// -------------------------
	// Benchmark + runner pattern
	// -------------------------
	benchmark := relu.MakeBuilder().
		WithSimulation(s).
		WithLength(1 << 20). // 1,048,576 elements = 16,384 work-groups
		Build("ReLU")

	// Run the benchmark directly with proper engine management
	log.Println("Starting GPU simulation with ReLU benchmark...")

	// Start driver in background
	go Driver.Run()

	// Run benchmark in main thread
	log.Println("Starting ReLU benchmark...")
	log.Printf("Benchmark length: %d", benchmark.GetUnderlyingBenchmark().Length)

	// Add timeout mechanism with debug output
	done := make(chan bool, 1)
	go func() {
		log.Println("Starting benchmark execution...")
		benchmark.Run()
		log.Println("Benchmark execution completed")
		done <- true
	}()

	<-done
	log.Println("Benchmark completed successfully!")

	// Terminate everything
	log.Println("Terminating driver...")
	Driver.Terminate()
	s.Terminate()

	log.Println("Simulation completed successfully.")
}
