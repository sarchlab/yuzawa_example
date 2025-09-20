package relu

import (
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/dnn/layer_benchmarks/relu"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
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
	d := b.sim.GetComponentByName("Driver").(*driver.Driver)
	cp := b.sim.GetComponentByName("CP").(*cp.CommandProcessor)

	// CP-Driver connection
	cp.Driver = d.GetPortByName("GPU")

	// Register GPU with driver
	d.RegisterGPU(cp.GetPortByName("ToDriver"), driver.DeviceProperties{
		CUCount:  1,
		DRAMSize: 4 * mem.GB,
	})

	r := relu.NewBenchmark(d)
	r.Length = b.length
	r.SelectGPU([]int{1})

	bm := &Benchmark{
		name: name,
		sim:  b.sim,
		relu: r,
	}

	return bm
}
