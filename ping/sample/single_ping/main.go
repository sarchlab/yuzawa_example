package main

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/single_ping"
	"github.com/sarchlab/yuzawa_example/ping/pinger"
)

func main() {
	simulation := sim.NewSimulation()

	engine := sim.NewSerialEngine()
	simulation.RegisterEngine(engine)

	pingBuilder := pinger.NewBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithName("Sender")
	sender := pingBuilder.Build()
	simulation.RegisterComponent(sender)

	pingBuilder = pinger.NewBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithName("Receiver")
	receiver := pingBuilder.Build()
	simulation.RegisterComponent(receiver)

	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")
	conn.PlugIn(sender.GetPortByName("PingPort"))
	conn.PlugIn(receiver.GetPortByName("PingPort"))

	benchmarkBuilder := single_ping.NewBuilder().
		WithSimulation(simulation).
		WithSender([]string{"Sender"}).
		WithReceiver("Receiver")
	benchmark := benchmarkBuilder.Build()

	benchmark.Run()
}
