// Package single_ping contains a benchmark that perform ping once.
package single_ping

import (
	"fmt"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/yuzawa_example/ping/pinger"
)

// The Benchmark struct is the benchmark that perform ping once.
type Benchmark struct {
	Name string
	simulation *sim.Simulation
	senderNames []string
	receiverName string
}

// Run runs the benchmark.
func (b *Benchmark) Run() {
	engine := b.simulation.GetEngine()
	senders := b.simulation.GetComponentByName(b.senderNames[0])
	receiver := b.simulation.GetComponentByName(b.receiverName)

	evt := pinger.NewPingEvent(senders, receiver, 0)
	engine.Schedule(evt)

	engine.Run()

	fmt.Printf("End time: %.10f seconds\n",
		engine.CurrentTime())
}
