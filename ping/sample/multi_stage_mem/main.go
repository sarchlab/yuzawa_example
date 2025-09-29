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
	"github.com/sarchlab/akita/v4/mem/vm/mmu"
	"github.com/sarchlab/mgpusim/v4/amd/timing/rob"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"

	"github.com/sarchlab/akita/v4/mem/vm/addresstranslator"
	"github.com/sarchlab/akita/v4/mem/vm/tlb"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/multi_stage_memory"
	"github.com/sarchlab/yuzawa_example/ping/memaccessagent"
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
		WithNumReqPerCycle(2).
		WithAddressMapperType("single").
		WithRemotePorts(MemCtrl.GetPortByName("Top").AsRemote()).
		Build("L2Cache")
	s.RegisterComponent(L2Cache)

	L1Cache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(2).
		WithLog2BlockSize(6).
		WithAddressMapperType("single").
		WithRemotePorts(L2Cache.GetPortByName("Top").AsRemote()).
		Build("L1Cache")
	s.RegisterComponent(L1Cache)

	pageTable := vm.NewPageTable(12)
	IoMMU := mmu.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithMaxNumReqInFlight(16).
		WithPageWalkingLatency(10).
		WithPageTable(pageTable).
		Build("IoMMU")
	s.RegisterComponent(IoMMU)

	L2TLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(64).
		WithNumSets(64).
		WithLog2PageSize(12).
		WithNumReqPerCycle(4).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(IoMMU.GetPortByName("Top").AsRemote()).
		Build("L2TLB")
	s.RegisterComponent(L2TLB)

	TLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithLog2PageSize(12).
		WithNumReqPerCycle(2).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(L2TLB.GetPortByName("Top").AsRemote()).
		Build("TLB")
	s.RegisterComponent(TLB)

	AT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithDeviceID(1).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(TLB.GetPortByName("Top").AsRemote()).
		WithMemoryProviderType("single").
		WithMemoryProviders(L1Cache.GetPortByName("Top").AsRemote()).
		Build("AT")
	s.RegisterComponent(AT)

	ROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		WithBottomUnit(AT.GetPortByName("Top").AsRemote()).
		Build("ROB")
	s.RegisterComponent(ROB)
	// if ROB.BottomUnit == nil {
	// 	panic("Failed to assign BottomUnit: Top port not found")
	// }

	MemAgent := memaccessagent.MakeBuilder().
		WithFreq(1 * sim.GHz).
		WithMaxAddress(1 * mem.GB).
		WithWriteLeft(100000).
		WithReadLeft(100000).
		WithEngine(engine).
		WithLowModule(L1Cache.GetPortByName("Top")).
		Build("MemAgent")
	s.RegisterComponent(MemAgent)

	Conn1 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn1")
	Conn1.PlugIn(MemAgent.GetPortByName("Mem"))
	Conn1.PlugIn(ROB.GetPortByName("Top"))

	Conn2 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn2")
	Conn2.PlugIn(ROB.GetPortByName("Bottom"))
	Conn2.PlugIn(AT.GetPortByName("Top"))

	Conn3 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn3")
	Conn3.PlugIn(AT.GetPortByName("Translation"))
	Conn3.PlugIn(TLB.GetPortByName("Top"))

	Conn4 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn4")
	Conn4.PlugIn(TLB.GetPortByName("Bottom"))
	Conn4.PlugIn(L2TLB.GetPortByName("Top"))

	Conn5 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn5")
	Conn5.PlugIn(L2TLB.GetPortByName("Bottom"))
	Conn5.PlugIn(IoMMU.GetPortByName("Top"))

	Conn6 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn6")
	Conn6.PlugIn(AT.GetPortByName("Bottom"))
	Conn6.PlugIn(L1Cache.GetPortByName("Top"))

	Conn7 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn7")
	Conn7.PlugIn(L1Cache.GetPortByName("Bottom"))
	Conn7.PlugIn(L2Cache.GetPortByName("Top"))

	Conn8 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn8")
	Conn8.PlugIn(L2Cache.GetPortByName("Bottom"))
	Conn8.PlugIn(MemCtrl.GetPortByName("Top"))

	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	logger := log.New(traceFile, "", 0)
	tracer := trace.NewTracer(logger, engine)
	tracing.CollectTrace(MemCtrl, tracer)

	benchmark := multi_stage_memory.MakeBuilder().
		WithSimulation(s).
		WithNumAccess(100000).
		WithMaxAddress(1 * mem.GB).
		Build("Benchmark")
	benchmark.Run()
}
