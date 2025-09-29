package ideal_mem_controller

import (
	"fmt"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
	"github.com/sarchlab/yuzawa_example/ping/memaccessagent"
)

// Benchmark is a benchmark that tests the IdealMemController.
type Benchmark struct {
	Name       string
	simulation *simulation.Simulation
	numAccess  int
	maxAddress uint64
}

// Run executes the benchmark.
func (b *Benchmark) Run() {
	// Set up metrics reporter
	metricsReporter := metrics_reporter.NewReporter(b.simulation)

	engine := b.simulation.GetEngine()
	agent := b.simulation.GetComponentByName("MemAgent").(*memaccessagent.MemAccessAgent)

	agent.WriteLeft = b.numAccess
	agent.ReadLeft = b.numAccess
	agent.MaxAddress = b.maxAddress

	agent.TickLater()
	err := engine.Run()
	if err != nil {
		panic(err)
	}

	if len(agent.PendingWriteReq) > 0 || len(agent.PendingReadReq) > 0 {
		panic(fmt.Errorf("there are still pending requests"))
	}

	if agent.WriteLeft > 0 || agent.ReadLeft > 0 {
		panic(fmt.Errorf("there are still requests left"))
	}

	fmt.Printf("End time: %.10f seconds\n",
		engine.CurrentTime())

	// Report metrics before completing
	metricsReporter.Report()
}
