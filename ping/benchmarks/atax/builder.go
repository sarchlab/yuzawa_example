package atax

import (
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/polybench/atax"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
)

type Builder struct {
	sim    *simulation.Simulation
	nx, ny int
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

// WithNx sets the first dimension size.
func (b *Builder) WithNx(nx int) *Builder {
	b.nx = nx
	return b
}

// WithNy sets the second dimension size.
func (b *Builder) WithNy(ny int) *Builder {
	b.ny = ny
	return b
}

// Build creates an ATAX benchmark with the given parameters.
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

	bm := atax.NewBenchmark(d)
	bm.NX = b.nx
	bm.NY = b.ny
	bm.SelectGPU([]int{1})

	return &Benchmark{
		name: name,
		sim:  b.sim,
		atax: bm,
	}
}
