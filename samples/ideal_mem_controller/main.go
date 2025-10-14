package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/sarchlab/akita/v4/mem/idealmemcontroller"
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/mem/trace"
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/akita/v4/tracing"

	"github.com/sarchlab/yuzawa_example/benchmarks/ideal_mem_controller"
	"github.com/sarchlab/yuzawa_example/components/memaccessagent"
)

func main() {
	// Seed random number generator
	seed := int64(0)
	if seed == 0 {
		seed = int64(time.Now().UnixNano())
	}
	rand.New(rand.NewSource(seed))
	log.Printf("Seed: %d\n", seed)

	s := simulation.MakeBuilder().Build()
	engine := s.GetEngine()

	// Create IdealMemoryController
	idealmemcontroller := idealmemcontroller.MakeBuilder().
		WithEngine(engine).
		WithNewStorage(4 * mem.GB).
		WithLatency(100).
		Build("IdealMemoryController")
	s.RegisterComponent(idealmemcontroller)

	// Instantiate MemAccessAgent using builder
	MemAgent := memaccessagent.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithMaxAddress(1 * mem.GB).
		WithWriteLeft(100000).
		WithReadLeft(100000).
		WithLowModule(idealmemcontroller.GetPortByName("Top")).
		Build("MemAgent")
	s.RegisterComponent(MemAgent)

	// Connect ports
	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")

	conn.PlugIn(MemAgent.GetPortByName("Mem"))
	conn.PlugIn(idealmemcontroller.GetPortByName("Top"))

	// Optional tracing
	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	logger := log.New(traceFile, "", 0)
	tracer := trace.NewTracer(logger, engine)
	tracing.CollectTrace(idealmemcontroller, tracer)

	// Run benchmark
	benchmark := ideal_mem_controller.MakeBuilder().
		WithSimulation(s).
		WithNumAccess(100000).
		WithMaxAddress(1 * mem.GB).
		Build("Benchmark")
	benchmark.Run()
}
