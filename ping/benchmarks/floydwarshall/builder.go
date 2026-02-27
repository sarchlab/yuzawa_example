package floydwarshall

import (
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/floydwarshall"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
)

type Builder struct {
	sim            *simulation.Simulation
	numNodes       int
	numIterations  int
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

// WithNumNodes sets the number of graph nodes.
func (b *Builder) WithNumNodes(n int) *Builder {
	b.numNodes = n
	return b
}

// WithNumIterations sets the number of iterations (0 = use NumNodes).
func (b *Builder) WithNumIterations(n int) *Builder {
	b.numIterations = n
	return b
}

// Build creates a Floyd-Warshall benchmark with the given parameters.
func (b *Builder) Build(name string) *Benchmark {
	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	cpComp := b.sim.GetComponentByName("CP")
	if cpComp == nil {
		panic("CP component not found in simulation!")
	}
	cp := cpComp.(*cp.CommandProcessor)

	d.RegisterGPU(cp.GetPortByName("ToDriver"), driver.DeviceProperties{
		CUCount:  1,
		DRAMSize: 4 * mem.GB,
	})

	bm := floydwarshall.NewBenchmark(d)
	bm.NumNodes = uint32(b.numNodes)
	bm.NumIterations = uint32(b.numIterations)
	bm.SelectGPU([]int{1})

	return &Benchmark{
		name: name,
		sim:  b.sim,
		fw:   bm,
	}
}
