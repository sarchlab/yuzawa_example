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
	"github.com/sarchlab/akita/v4/tracing"

	"github.com/sarchlab/yuzawa_example/ping/benchmarks/ideal_mem_controller"
	"github.com/sarchlab/yuzawa_example/ping/memaccessagent"
)

func main() {
	// Seed random number generator
	seed := int64(0)
	if seed == 0 {
		seed = int64(time.Now().UnixNano())
	}
	rand.New(rand.NewSource(seed))
	log.Printf("Seed: %d\n", seed)

	// Create simulation and engine
	simulation := sim.NewSimulation()
	engine := sim.NewParallelEngine()
	simulation.RegisterEngine(engine)

	// Instantiate MemAccessAgent using builder
	agent := memaccessagent.MakeBuilder().
		WithEngine(engine).
		WithName("MemAgent").
		WithFreq(1 * sim.GHz).
		WithMaxAddress(1 * mem.GB).
		WithWriteLeft(100000).
		WithReadLeft(100000).
		Build("MemAgent")
	simulation.RegisterComponent(agent)

	// Create IdealMemoryController
	idealmemcontroller := idealmemcontroller.MakeBuilder().
		WithEngine(engine).
		WithNewStorage(4 * mem.GB).
		WithLatency(100).
		Build("IdealMemoryController")
	simulation.RegisterComponent(idealmemcontroller)

	topPort := idealmemcontroller.GetPortByName("Top")
	if topPort == nil {
		panic("Error: IdealMemoryController GetPortByName('Top') returned nil")
	}

	// Assign LowModule for MemAccessAgent
	agent.LowModule = topPort
	if agent.LowModule == nil {
		panic("Failed to assign LowModule: Port does not exist")
	}

	// Connect ports
	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")

	// Ensure agent MemPort exists
	agentPort := agent.GetPortByName("Mem")
	if agentPort == nil {
		panic("Error: MemPort does not exist")
	}

	conn.PlugIn(agent.GetPortByName("Mem"))
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
		WithSimulation(simulation).
		WithNumAccess(100000).
		WithMaxAddress(1 * mem.GB).
		Build("Benchmark")
	benchmark.Run()
}
