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

	pingBuilder := pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	sender := pingBuilder.Build("Sender")
	simulation.RegisterComponent(sender)

	pingBuilder = pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	receiver := pingBuilder.Build("Receiver")
	simulation.RegisterComponent(receiver)

	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")
	conn.PlugIn(sender.GetPortByName("PingPort"))
	conn.PlugIn(receiver.GetPortByName("PingPort"))

	benchmarkBuilder := single_ping.MakeBuilder().
		WithSimulation(simulation).
		WithSender([]string{"Sender"}).
		WithReceiver("Receiver")
	benchmark := benchmarkBuilder.Build("Benchmark")

	benchmark.Run()
}
