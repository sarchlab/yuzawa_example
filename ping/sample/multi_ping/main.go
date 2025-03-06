package main

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/multi_ping"
	"github.com/sarchlab/yuzawa_example/ping/pinger"
)

func main() {
	simulation := sim.NewSimulation()

	engine := sim.NewSerialEngine()
	simulation.RegisterEngine(engine)

	// Create two senders
	pingBuilder := pinger.NewBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithName("Sender1")
	sender1 := pingBuilder.Build()
	simulation.RegisterComponent(sender1)

	pingBuilder = pinger.NewBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithName("Sender2")
	sender2 := pingBuilder.Build()
	simulation.RegisterComponent(sender2)

	// Create a receiver
	pingBuilder = pinger.NewBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithName("Receiver")
	receiver := pingBuilder.Build()
	simulation.RegisterComponent(receiver)

	// Create connection
	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")
	conn.PlugIn(sender1.GetPortByName("PingPort"))
	conn.PlugIn(sender2.GetPortByName("PingPort"))
	conn.PlugIn(receiver.GetPortByName("PingPort"))

	// Run multiple pings
	benchmarkBuilder := multi_ping.NewBuilder().
		WithSimulation(simulation).
		WithSenders([]string{"Sender1", "Sender2"}).
		WithReceiver("Receiver").
		WithNumPings(5) // Each sender sends 5 pings
	benchmark := benchmarkBuilder.Build()

	benchmark.Run()
}
