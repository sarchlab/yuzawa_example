// Package single_ping contains a benchmark that perform ping once.
package single_ping

import (
	"fmt"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/yuzawa_example/ping/pinger"
)

// The Benchmark struct is the benchmark that perform ping once.
type Benchmark struct {
	simulation *sim.Simulation

	senderName, receiverName string
}

// Run runs the benchmark.
func (b *Benchmark) Run() {
	engine := b.simulation.GetEngine()
	sender := b.simulation.GetComponentByName(b.senderName)
	receiver := b.simulation.GetComponentByName(b.receiverName)

	evt := pinger.NewPingEvent(sender, receiver, 0)
	engine.Schedule(evt)

	engine.Run()

	fmt.Printf("End time: %.10f seconds\n",
		engine.CurrentTime())
}
