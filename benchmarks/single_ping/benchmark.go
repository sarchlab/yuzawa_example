// Package single_ping contains a benchmark that perform ping once.
package single_ping

import (
	"fmt"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/yuzawa_example/components/metrics_reporter"
	"github.com/sarchlab/yuzawa_example/components/pinger"
)

// The Benchmark struct is the benchmark that perform ping once.
type Benchmark struct {
	Name         string
	simulation   *simulation.Simulation
	senderNames  []string
	receiverName string
}

// Run runs the benchmark.
func (b *Benchmark) Run() {
	// Set up metrics reporter
	metricsReporter := metrics_reporter.NewReporter(b.simulation)

	engine := b.simulation.GetEngine()
	senders := b.simulation.GetComponentByName(b.senderNames[0])
	receiver := b.simulation.GetComponentByName(b.receiverName)

	evt := pinger.NewPingEvent(senders, receiver, 0)
	engine.Schedule(evt)

	engine.Run()

	fmt.Printf("End time: %.10f seconds\n",
		engine.CurrentTime())

	// Report metrics before completing
	metricsReporter.Report()
}
