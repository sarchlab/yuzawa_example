package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/sarchlab/akita/v4/mem/cache/writeback"
	"github.com/sarchlab/akita/v4/mem/cache/writethrough"
	"github.com/sarchlab/akita/v4/mem/idealmemcontroller"
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/mem/trace"
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/ideal_mem_controller"
	"github.com/sarchlab/yuzawa_example/ping/memaccessagent"
)

func main() {
	seed := int64(0)
	if seed == 0 {
		seed = int64(time.Now().UnixNano())
	}
	rand.New(rand.NewSource(seed))
	log.Printf("Seed: %d\n", seed)

	simulation := sim.NewSimulation()
	engine := sim.NewParallelEngine()
	simulation.RegisterEngine(engine)

	core := memaccessagent.MakeBuilder().
		WithName("Core").
		WithFreq(1 * sim.GHz).
		WithMaxAddress(1 * mem.GB).
		WithWriteLeft(100000).
		WithReadLeft(100000).
		WithEngine(engine).
		Build("Core")
	simulation.RegisterComponent(core)

	l1 := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(2).
		Build("L1Cache")
	simulation.RegisterComponent(l1)

	l2 := writeback.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).
		WithNumReqPerCycle(2).
		Build("L2Cache")
	simulation.RegisterComponent(l2)

	memCtrl := idealmemcontroller.MakeBuilder().
		WithEngine(engine).
		WithNewStorage(4 * mem.GB).
		WithLatency(100).
		Build("MemCtrl")
	simulation.RegisterComponent(memCtrl)

	l1.SetAddressToPortMapper(&mem.SinglePortMapper{Port: l2.GetPortByName("Top").AsRemote()})
	l2.SetAddressToPortMapper(&mem.SinglePortMapper{Port: memCtrl.GetPortByName("Top").AsRemote()})

	core.LowModule = l1.GetPortByName("Top")

	conn1 := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn1")
	conn1.PlugIn(core.GetPortByName("Mem"))
	conn1.PlugIn(l1.GetPortByName("Top"))

	conn2 := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn2")
	conn2.PlugIn(l1.GetPortByName("Bottom"))
	conn2.PlugIn(l2.GetPortByName("Top"))

	conn3 := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn3")
	conn3.PlugIn(l2.GetPortByName("Bottom"))
	conn3.PlugIn(memCtrl.GetPortByName("Top"))

	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	logger := log.New(traceFile, "", 0)
	tracer := trace.NewTracer(logger, engine)
	tracing.CollectTrace(memCtrl, tracer)

	benchmark := ideal_mem_controller.MakeBuilder().
		WithSimulation(simulation).
		WithNumAccess(100000).
		WithMaxAddress(1 * mem.GB).
		Build("Benchmark")

	benchmark.Run()

}
