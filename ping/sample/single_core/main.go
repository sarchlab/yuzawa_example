package main

import (
	"log"
	// "os"

	"github.com/sarchlab/akita/v4/mem/cache/writeback"
	"github.com/sarchlab/akita/v4/mem/cache/writethrough"
	"github.com/sarchlab/akita/v4/mem/idealmemcontroller"
	"github.com/sarchlab/akita/v4/mem/mem"
	// "github.com/sarchlab/akita/v4/mem/trace"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"

	"github.com/sarchlab/akita/v4/mem/vm/addresstranslator"
	"github.com/sarchlab/akita/v4/mem/vm/tlb"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	// "github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/relu"
	"github.com/sarchlab/yuzawa_example/ping/mmu"
	"github.com/sarchlab/yuzawa_example/ping/rob"
	"github.com/sarchlab/yuzawa_example/cp"
	"github.com/sarchlab/yuzawa_example/cu"
	// "github.com/sarchlab/yuzawa_example/driver"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	// "github.com/sarchlab/mgpusim/v4/amd/timing/cp"
	// "github.com/sarchlab/mgpusim/v4/amd/timing/cu"
)

func main() {
	s := simulation.MakeBuilder().Build()
	engine := s.GetEngine()

	MemCtrl := idealmemcontroller.MakeBuilder().
		WithEngine(engine).
		WithNewStorage(4 * mem.GB).
		WithLatency(100).
		Build("MemCtrl")
	s.RegisterComponent(MemCtrl)

	L2Cache := writeback.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).
		WithNumReqPerCycle(8).
		WithAddressMapperType("single").
		WithRemotePorts(MemCtrl.GetPortByName("Top").AsRemote()).
		Build("L2Cache")
	s.RegisterComponent(L2Cache)

	L1VCache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(2).
		WithAddressMapperType("single").
		WithRemotePorts(L2Cache.GetPortByName("Top").AsRemote()).
		Build("L1VCache")
	s.RegisterComponent(L1VCache)

	IoMMU := mmu.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithMaxNumReqInFlight(16).
		WithPageWalkingLatency(10).
		Build("IoMMU")
	s.RegisterComponent(IoMMU)

	L2TLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(64).
		WithNumSets(64).
		WithPageSize(4096).
		WithNumReqPerCycle(8).
		WithRemotePorts(IoMMU.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("L2TLB")
	s.RegisterComponent(L2TLB)

	VTLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithPageSize(4096).
		WithNumReqPerCycle(2).
		WithRemotePorts(L2TLB.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("VTLB")
	s.RegisterComponent(VTLB)

	VAT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithTranslationProvider(VTLB.GetPortByName("Top").AsRemote()).
		WithRemotePorts(L1VCache.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("VAT")
	s.RegisterComponent(VAT)

	VROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		WithBottomUnit(VAT.GetPortByName("Top")).
		Build("VROB")
	s.RegisterComponent(VROB)

	L1SCache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(2).
		WithAddressMapperType("single").
		WithRemotePorts(L2Cache.GetPortByName("Top").AsRemote()).
		Build("L1SCache")
	s.RegisterComponent(L1SCache)

	STLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithPageSize(4096).
		WithNumReqPerCycle(2).
		WithRemotePorts(L2TLB.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("STLB")
	s.RegisterComponent(STLB)

	SAT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithTranslationProvider(STLB.GetPortByName("Top").AsRemote()).
		WithRemotePorts(L1SCache.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("SAT")
	s.RegisterComponent(SAT)

	SROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		WithBottomUnit(SAT.GetPortByName("Top")).
		Build("SROB")
	s.RegisterComponent(SROB)

	L1ICache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(2).
		WithAddressMapperType("single").
		WithRemotePorts(L2Cache.GetPortByName("Top").AsRemote()).
		Build("L1ICache")
	s.RegisterComponent(L1ICache)

	ITLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithPageSize(4096).
		WithNumReqPerCycle(2).
		WithRemotePorts(L2TLB.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("ITLB")
	s.RegisterComponent(ITLB)

	IAT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithTranslationProvider(ITLB.GetPortByName("Top").AsRemote()).
		WithRemotePorts(L1ICache.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("IAT")
	s.RegisterComponent(IAT)

	IROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		WithBottomUnit(IAT.GetPortByName("Top")).
		Build("IROB")
	s.RegisterComponent(IROB)

	CU := cu.MakeBuilder().
		WithSimulation(s).
		WithFreq(1 * sim.GHz).
		Build("CU")
	s.RegisterComponent(CU)

	CP := cp.MakeBuilder().
		WithSimulation(s).
		WithFreq(1 * sim.GHz).
		Build("CP")
	s.RegisterComponent(CP)

	Driver := driver.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Driver")
	s.RegisterComponent(Driver)

	Conn1 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn1")
	Conn1.PlugIn(Driver.GetPortByName("GPU"))
	Conn1.PlugIn(CP.GetPortByName("Top"))

	Conn2 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn2")
	Conn2.PlugIn(CP.GetPortByName("RequestPort"))
	Conn2.PlugIn(CU.GetPortByName("Top"))

	Conn3 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn3")
	Conn3.PlugIn(CU.GetPortByName("VectorPort"))
	Conn3.PlugIn(VROB.GetPortByName("Top"))

	Conn4 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn4")
	Conn4.PlugIn(VROB.GetPortByName("Bottom"))
	Conn4.PlugIn(VAT.GetPortByName("Top"))

	Conn5 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn5")
	Conn5.PlugIn(VAT.GetPortByName("Translation"))
	Conn5.PlugIn(VTLB.GetPortByName("Top"))

	Conn6 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn6")
	Conn6.PlugIn(VTLB.GetPortByName("Bottom"))
	Conn6.PlugIn(L2TLB.GetPortByName("Top"))

	Conn7 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn7")
	Conn7.PlugIn(L2TLB.GetPortByName("Bottom"))
	Conn7.PlugIn(IoMMU.GetPortByName("Top"))

	Conn8 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn8")
	Conn8.PlugIn(VAT.GetPortByName("Bottom"))
	Conn8.PlugIn(L1VCache.GetPortByName("Top"))

	Conn9 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn9")
	Conn9.PlugIn(L1VCache.GetPortByName("Bottom"))
	Conn9.PlugIn(L2Cache.GetPortByName("Top"))

	Conn10 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn10")
	Conn10.PlugIn(L2Cache.GetPortByName("Bottom"))
	Conn10.PlugIn(MemCtrl.GetPortByName("Top"))

	Conn11 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn11")
	Conn11.PlugIn(CU.GetPortByName("ScalarPort"))
	Conn11.PlugIn(SROB.GetPortByName("Top"))

	Conn12 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn12")
	Conn12.PlugIn(SROB.GetPortByName("Bottom"))
	Conn12.PlugIn(SAT.GetPortByName("Top"))

	Conn13 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn13")
	Conn13.PlugIn(SAT.GetPortByName("Translation"))
	Conn13.PlugIn(STLB.GetPortByName("Top"))

	// Conn14 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn14")
	// Conn14.PlugIn(STLB.GetPortByName("Bottom"))
	// Conn14.PlugIn(L2TLB.GetPortByName("Top"))

	Conn15 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn15")
	Conn15.PlugIn(CU.GetPortByName("InstPort"))
	Conn15.PlugIn(IROB.GetPortByName("Top"))

	Conn16 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn16")
	Conn16.PlugIn(IROB.GetPortByName("Bottom"))
	Conn16.PlugIn(IAT.GetPortByName("Top"))

	Conn17 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn17")
	Conn17.PlugIn(IAT.GetPortByName("Translation"))
	Conn17.PlugIn(ITLB.GetPortByName("Top"))

	// Conn18 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn18")
	// Conn18.PlugIn(ITLB.GetPortByName("Bottom"))
	// Conn18.PlugIn(L2TLB.GetPortByName("Top"))

	// traceFile, err := os.Create("trace.log")
	// if err != nil {
	// 	panic("Error: Failed to create trace file")
	// }
	// logger := log.New(traceFile, "", 0)
	// tracer := trace.NewTracer(logger, engine)
	// tracing.CollectTrace(MemCtrl, tracer)

	benchmark := relu.MakeBuilder().
		WithSimulation(s).
		Build("Benchmark")
	benchmark.Run()

	s.Terminate()
	log.Println("Simulation completed successfully.")
}