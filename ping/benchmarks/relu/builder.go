package relu

import (
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/dnn/layer_benchmarks/relu"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
)

type Builder struct {
	sim    *simulation.Simulation
	length int
}

// MakeBuilder creates a builder with default parameters.
func MakeBuilder() *Builder {
	return &Builder{}
}

// WithSimulation sets the simulation to use.
func (b *Builder) WithSimulation(sim *simulation.Simulation) *Builder {
	b.sim = sim
	return b
}

// WithLength sets the number of element to perform ReLU operation on.
func (b *Builder) WithLength(l int) *Builder {
	b.length = l
	return b
}

// Build creates a ReLU benchmark with the given parameters.
func (b *Builder) Build(name string) *Benchmark {
	driver := b.sim.GetComponentByName("Driver").(*driver.Driver)

	r := relu.NewBenchmark(driver)
	r.Length = b.length
	r.SelectGPU([]int{1}) // Use GPU index 1 (MGPUSim driver expects IDs starting from 1)

	bm := &Benchmark{
		name: name,
		sim:  b.sim,
		relu: r,
	}

	return bm
}
