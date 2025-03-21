package main

import (
	"log"
	"os"
	"time"
	"math/rand"

	"github.com/sarchlab/akita/v4/mem/acceptancetests"
	"github.com/sarchlab/akita/v4/mem/idealmemcontroller"
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/mem/trace"
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/imc"
)

func main() {
	// Seed random number generator
	seed := int64(0)
	if seed == 0 {
		seed = int64(time.Now().UnixNano())
	}
	// rand.Seed(seed)
	rand.New(rand.NewSource(seed))

	simulation := sim.NewSimulation()
	engine := sim.NewParallelEngine()
	simulation.RegisterEngine(engine)

	// Create memory access agent
	agent := acceptancetests.NewMemAccessAgent(engine)
	agent.MaxAddress = 1 * mem.GB
	agent.WriteLeft = 100000
	agent.ReadLeft = 100000
	simulation.RegisterComponent(agent)

	// Create ideal memory controller
	dram := idealmemcontroller.MakeBuilder().
		WithEngine(engine).
		WithNewStorage(4 * mem.GB).
		WithLatency(100).
		Build("DRAM")
	simulation.RegisterComponent(dram)

	if dram == nil {
		panic("Error: Memory controller failed to initialize")
	}

	topPort := dram.GetPortByName("Top")
	if topPort == nil {
		panic("Error: DRAM GetPortByName('Top') returned nil")
	}

	// Assign LowModule for MemAccessAgent
	agent.LowModule = topPort

	if agent.LowModule == nil {
		panic("Failed to assign LowModule: Port does not exist")
	}

	// Create connection
	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")

	// Ensure agent MemPort exists
	agentMemPort := agent.GetPortByName("Mem")
	if agentMemPort == nil {
		panic("Error: MemPort does not exist")
	}

	conn.PlugIn(agentMemPort)
	conn.PlugIn(dram.GetPortByName("Top"))

	// Enable tracing if --trace flag is provided
	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	logger := log.New(traceFile, "", 0)
	tracer := trace.NewTracer(logger, engine)
	tracing.CollectTrace(dram, tracer)

	benchmarkBuilder := imc.NewBuilder().
		WithSimulation(simulation).
		WithNumAccess(1000).
		WithMaxAddress(1 * mem.GB)
	benchmark := benchmarkBuilder.Build()

	if benchmark == nil {
		panic("Error: Benchmark failed to initialize")
	}

	benchmark.Run()
}
