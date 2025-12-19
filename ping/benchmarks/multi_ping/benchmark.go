// Package multi_ping contains a benchmark that performs multiple pings.
package multi_ping

import (
	"fmt"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
	"github.com/sarchlab/yuzawa_example/ping/pinger"
)

// The Benchmark struct is the benchmark that performs multiple pings.
type Benchmark struct {
	Name              string
	simulation        *simulation.Simulation
	senderNames       []string
	receiverName      string
	numPingsPerSender int
}

// Run runs the multi-ping benchmark.
func (b *Benchmark) Run() {
	// Set up metrics reporter
	metricsReporter := metrics_reporter.NewReporter(b.simulation)

	engine := b.simulation.GetEngine()
	receiver := b.simulation.GetComponentByName(b.receiverName)

	// Send multiple pings from each sender
	for senderIndex, senderName := range b.senderNames {
		sender := b.simulation.GetComponentByName(senderName)
		for i := 0; i < b.numPingsPerSender; i++ {
			// Stagger events by both the ping index and the sender index.
			offset := 0.04*float64(i+1) + 0.001*float64(senderIndex)
			evt := pinger.NewPingEvent(sender, receiver, engine.CurrentTime()+sim.VTimeInSec(offset))
			engine.Schedule(evt)
		}
	}

	engine.Run()

	fmt.Printf("End time: %.10f seconds\n",
		engine.CurrentTime())

	// Report metrics before completing
	metricsReporter.Report()
}
