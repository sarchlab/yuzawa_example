package nbody

import (
	"fmt"

	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/nbody"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
)

const nbodyGroupSize = 256

type Builder struct {
	sim             *simulation.Simulation
	numParticles    int
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

// WithNumParticles sets the number of particles.
func (b *Builder) WithNumParticles(n int) *Builder {
	b.numParticles = n
	return b
}

// WithNumIterations sets the number of simulation iterations.
func (b *Builder) WithNumIterations(n int) *Builder {
	b.numIterations = n
	return b
}

// Build creates an NBody benchmark with the given parameters.
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

	cuCount := 0
	for i := 0; ; i++ {
		cuName := fmt.Sprintf("CU[%d]", i)
		comp := b.sim.GetComponentByName(cuName)
		if comp == nil || comp.Name() != cuName {
			break
		}
		cuCount++
	}
	if cuCount == 0 {
		cuCount = 1
	}

	d.RegisterGPU(cp.GetPortByName("ToDriver"), driver.DeviceProperties{
		CUCount:  cuCount,
		DRAMSize: 4 * mem.GB,
	})

	bm := nbody.NewBenchmark(d)
	np := int32(b.numParticles)
	if np < nbodyGroupSize {
		np = nbodyGroupSize
	}
	np = (np / nbodyGroupSize) * nbodyGroupSize
	bm.NumParticles = np
	bm.NumIterations = int32(b.numIterations)
	bm.SelectGPU([]int{1})

	return &Benchmark{
		name:   name,
		sim:    b.sim,
		nbody:  bm,
	}
}
