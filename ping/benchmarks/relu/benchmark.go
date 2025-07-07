package relu

import (
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/dnn/layer_benchmarks/relu"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
)

type Benchmark struct {
	name string
	sim *simulation.Simulation
	relu *relu.Benchmark
}

func (b *Benchmark) Run() {
	driver := b.sim.GetComponentByName("Driver").(*driver.Driver)

	driver.Run()
	b.relu.Run()

	driver.Terminate()
}