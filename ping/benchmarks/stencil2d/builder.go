package stencil2d

import (
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	mgpusimstencil "github.com/sarchlab/mgpusim/v4/amd/benchmarks/shoc/stencil2d"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
)

type Builder struct {
	sim          *simulation.Simulation
	numRows      int
	numCols      int
	numIteration int
}

// MakeBuilder creates a builder with default parameters.
func MakeBuilder() *Builder {
	return &Builder{
		numRows:      256,
		numCols:      256,
		numIteration: 10,
	}
}

// WithSimulation sets the simulation to use.
func (b *Builder) WithSimulation(sim *simulation.Simulation) *Builder {
	b.sim = sim
	return b
}

// WithNumRows sets the number of rows.
func (b *Builder) WithNumRows(n int) *Builder {
	b.numRows = n
	return b
}

// WithNumCols sets the number of columns.
func (b *Builder) WithNumCols(n int) *Builder {
	b.numCols = n
	return b
}

// WithNumIteration sets the number of stencil iterations.
func (b *Builder) WithNumIteration(n int) *Builder {
	b.numIteration = n
	return b
}

// Build creates a stencil2d benchmark with the given parameters.
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
	commandProcessor := cpComp.(*cp.CommandProcessor)

	d.RegisterGPU(commandProcessor.GetPortByName("ToDriver"), driver.DeviceProperties{
		CUCount:  1,
		DRAMSize: 4 * mem.GB,
	})

	bm := mgpusimstencil.NewBenchmark(d)
	bm.NumRows = b.numRows
	bm.NumCols = b.numCols
	bm.NumIteration = b.numIteration
	bm.SelectGPU([]int{1})

	return &Benchmark{
		name:      name,
		sim:       b.sim,
		stencil2d: bm,
	}
}
