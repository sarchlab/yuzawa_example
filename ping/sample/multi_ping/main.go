package main

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/multi_ping"
	"github.com/sarchlab/yuzawa_example/ping/pinger"
)

func main() {
	simBuilder := simulation.MakeBuilder().Build()
	engine := simBuilder.GetEngine()

	// Create two senders
	pingBuilder := pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	sender1 := pingBuilder.Build("Sender1")
	simBuilder.RegisterComponent(sender1)

	pingBuilder = pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	sender2 := pingBuilder.Build("Sender2")
	simBuilder.RegisterComponent(sender2)

	// Create a receiver
	pingBuilder = pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	receiver := pingBuilder.Build("Receiver")
	simBuilder.RegisterComponent(receiver)

	// Create connection
	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")
	conn.PlugIn(sender1.GetPortByName("PingPort"))
	conn.PlugIn(sender2.GetPortByName("PingPort"))
	conn.PlugIn(receiver.GetPortByName("PingPort"))

	// Run multiple pings
	benchmarkBuilder := multi_ping.MakeBuilder().
		WithSimulation(simBuilder).
		WithSenders([]string{"Sender1", "Sender2"}).
		WithReceiver("Receiver").
		WithNumPings(5) // Each sender sends 5 pings
	benchmark := benchmarkBuilder.Build("Benchmark")

	benchmark.Run()
}
