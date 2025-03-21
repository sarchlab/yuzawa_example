package idealmemcontroller

import (
	"fmt"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/mem/acceptancetests"
)

// Benchmark is a benchmark that tests the IdealMemController.
type Benchmark struct {
	simulation *sim.Simulation
	numAccess int
	maxAddress uint64
}

// Run executes the benchmark.
func (b *Benchmark) Run() {
	engine := b.simulation.GetEngine()
	agent := b.simulation.GetComponentByName("MemAgent").(*acceptancetests.MemAccessAgent)

	agent.WriteLeft = b.numAccess
	agent.ReadLeft = b.numAccess
	agent.MaxAddress = b.maxAddress

	agent.TickLater()
	err := engine.Run()
	if err != nil {
		panic(err)
	}

	if len(agent.PendingWriteReq) > 0 || len(agent.PendingReadReq) > 0 {
		panic(fmt.Errorf("There are still pending requests"))
	}

	if agent.WriteLeft > 0 || agent.ReadLeft > 0 {
		panic(fmt.Errorf("There are still requests left"))
	}

	fmt.Printf("End time: %.10f seconds\n",
		engine.CurrentTime())
}

