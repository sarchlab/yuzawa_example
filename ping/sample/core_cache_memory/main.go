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
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/ideal_mem_controller"
	"github.com/sarchlab/yuzawa_example/ping/memaccessagent"
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

	MemAgent := memaccessagent.MakeBuilder().
		WithFreq(1 * sim.GHz).
		WithMaxAddress(1 * mem.GB).
		WithWriteLeft(100000).
		WithReadLeft(100000).
		WithEngine(engine).
		Build("MemAgent")
	simBuilder.RegisterComponent(MemAgent)

	MemAgent.LowModule = L1Cache.GetPortByName("Top")
	if MemAgent.LowModule == nil {
		panic("Failed to assign LowModule: Top port not found")
	}

	Conn1 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn1")
	Conn1.PlugIn(MemAgent.GetPortByName("Mem"))
	Conn1.PlugIn(L1Cache.GetPortByName("Top"))

	Conn2 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn2")
	Conn2.PlugIn(L1Cache.GetPortByName("Bottom"))
	Conn2.PlugIn(L2Cache.GetPortByName("Top"))

	Conn3 := directconnection.MakeBuilder().WithEngine(engine).WithFreq(1 * sim.GHz).Build("Conn3")
	Conn3.PlugIn(L2Cache.GetPortByName("Bottom"))
	Conn3.PlugIn(MemCtrl.GetPortByName("Top"))

	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	logger := log.New(traceFile, "", 0)
	tracer := trace.NewTracer(logger, engine)
	tracing.CollectTrace(MemCtrl, tracer)

	benchmark := ideal_mem_controller.MakeBuilder().
		WithSimulation(simBuilder).
		WithNumAccess(100000).
		WithMaxAddress(1 * mem.GB).
		Build("Benchmark")
	benchmark.Run()
}
