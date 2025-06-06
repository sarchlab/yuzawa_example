package main

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/single_ping"
	"github.com/sarchlab/yuzawa_example/ping/pinger"
)

func main() {
	// simulation := sim.NewSimulation()

	// engine := sim.NewSerialEngine()
	// simulation.RegisterEngine(engine)

	simBuilder := simulation.MakeBuilder().Build()
	// simulation := simBuilder.Build()
	engine := simBuilder.GetEngine()

	pingBuilder := pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	sender := pingBuilder.Build("Sender")
	simBuilder.RegisterComponent(sender)

	pingBuilder = pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	receiver := pingBuilder.Build("Receiver")
	simBuilder.RegisterComponent(receiver)

	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")
	conn.PlugIn(sender.GetPortByName("PingPort"))
	conn.PlugIn(receiver.GetPortByName("PingPort"))

	benchmarkBuilder := single_ping.MakeBuilder().
		WithSimulation(simBuilder).
		WithSender([]string{"Sender"}).
		WithReceiver("Receiver")
	benchmark := benchmarkBuilder.Build("Benchmark")

	benchmark.Run()
}
