package relu

import (
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/dnn/layer_benchmarks/relu"

)

type Builder struct {
	sim *simulation.Simulation
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

// Build creates a ReLU benchmark with the given parameters.
func (b *Builder) Build(name string) *Benchmark {
	r := relu.NewBenchmark(nil)

	return &Benchmark{
		name: name,
		sim: b.sim,
		relu: r,
	}
}