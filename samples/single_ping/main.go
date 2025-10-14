package main

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/yuzawa_example/benchmarks/single_ping"
	"github.com/sarchlab/yuzawa_example/components/pinger"
)

func main() {
	s := simulation.MakeBuilder().Build()
	engine := s.GetEngine()

	pingBuilder := pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	sender := pingBuilder.Build("Sender")
	s.RegisterComponent(sender)

	pingBuilder = pinger.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz)
	receiver := pingBuilder.Build("Receiver")
	s.RegisterComponent(receiver)

	conn := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("Conn")
	conn.PlugIn(sender.GetPortByName("PingPort"))
	conn.PlugIn(receiver.GetPortByName("PingPort"))

	benchmarkBuilder := single_ping.MakeBuilder().
		WithSimulation(s).
		WithSender([]string{"Sender"}).
		WithReceiver("Receiver")
	benchmark := benchmarkBuilder.Build("Benchmark")

	benchmark.Run()
}
