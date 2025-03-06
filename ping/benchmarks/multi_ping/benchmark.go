// Package multi_ping contains a benchmark that performs multiple pings.
package multi_ping

import (
	"fmt"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/yuzawa_example/ping/pinger"
)

// The Benchmark struct is the benchmark that performs multiple pings.
type Benchmark struct {
	simulation             *sim.Simulation
	senderNames            []string
	receiverName           string
	numPingsPerSender      int
}

// Run runs the multi-ping benchmark.
func (b *Benchmark) Run() {
	engine := b.simulation.GetEngine()
	receiver := b.simulation.GetComponentByName(b.receiverName)

	// Send multiple pings from each sender
	for _, senderName := range b.senderNames {
		sender := b.simulation.GetComponentByName(senderName)
		for i := 0; i < b.numPingsPerSender; i++ {
			evt := pinger.NewPingEvent(sender, receiver, 0)
			engine.Schedule(evt)
		}
	}

	engine.Run()

	fmt.Printf("End time: %.10f seconds\n",
		engine.CurrentTime())
}
