package main

import (
	"log"
	"os"

	"github.com/sarchlab/akita/v4/mem/cache/writeback"
	"github.com/sarchlab/akita/v4/mem/cache/writethrough"
	"github.com/sarchlab/akita/v4/mem/idealmemcontroller"
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/mem/trace"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"

	"github.com/sarchlab/akita/v4/mem/vm/addresstranslator"
	"github.com/sarchlab/akita/v4/mem/vm/tlb"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/multi_stage_memory"
	"github.com/sarchlab/yuzawa_example/ping/memaccessagent"
	"github.com/sarchlab/yuzawa_example/ping/mmu"
	"github.com/sarchlab/yuzawa_example/ping/rob"
)

func main() {
	// simulation := sim.NewSimulation()
	// engine := sim.NewParallelEngine()
	// simulation.RegisterEngine(engine)

	simBuilder := simulation.MakeBuilder().Build()
	// simulation := simBuilder.Build()
	engine := simBuilder.GetEngine()

	MemCtrl := idealmemcontroller.MakeBuilder().
		WithEngine(engine).
		WithNewStorage(4 * mem.GB).
		WithLatency(100).
		Build("MemCtrl")
	simBuilder.RegisterComponent(MemCtrl)

	L2Cache := writeback.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).
		WithNumReqPerCycle(2).
		WithAddressMapperType("single").
		WithRemotePorts(MemCtrl.GetPortByName("Top").AsRemote()).
		Build("L2Cache")
	simBuilder.RegisterComponent(L2Cache)

	L1Cache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(2).
		WithAddressMapperType("single").
		WithRemotePorts(L2Cache.GetPortByName("Top").AsRemote()).
		Build("L1Cache")
	simBuilder.RegisterComponent(L1Cache)

	// for _, p := range L1Cache.Ports() {
	// 	fmt.Println("L1 port name:", p.Name())
	// }

	IoMMU := mmu.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithMaxNumReqInFlight(16).
		WithPageWalkingLatency(10).
		Build("IoMMU")
	simBuilder.RegisterComponent(IoMMU)

	L2TLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(64).
		WithNumSets(64).
		WithPageSize(4096).
		WithNumReqPerCycle(4).
		WithRemotePorts(IoMMU.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("L2TLB")
	simBuilder.RegisterComponent(L2TLB)

	TLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithPageSize(4096).
		WithNumReqPerCycle(2).
		WithRemotePorts(L2TLB.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("TLB")
	simBuilder.RegisterComponent(TLB)

	AT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithTranslationProvider(TLB.GetPortByName("Top").AsRemote()).
		WithRemotePorts(L1Cache.GetPortByName("Top").AsRemote()).
		WithAddressMapperType("single").
		Build("AT")
	simBuilder.RegisterComponent(AT)

	ROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		Build("ROB")
	simBuilder.RegisterComponent(ROB)

	ROB.BottomUnit = AT.GetPortByName("Top")
	if ROB.BottomUnit == nil {
		panic("Failed to assign BottomUnit: Top port not found")
	}

	MemAgent := memaccessagent.MakeBuilder().
		WithFreq(1 * sim.GHz).
		WithMaxAddress(1 * mem.GB).
		WithWriteLeft(100000).
		WithReadLeft(100000).
		WithEngine(engine).
		Build("MemAgent")
	simBuilder.RegisterComponent(MemAgent)
	MemAgent.LowModule = ROB.GetPortByName("Top")
	if MemAgent.LowModule == nil {
		panic("Failed to assign LowModule: Top port not found")
	}
	
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
	WithSimulation(simBuilder).
	WithNumAccess(100000).
	WithMaxAddress(1 * mem.GB).
	Build("Benchmark")
	benchmark.Run()
}
