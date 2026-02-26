package fir

import (
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/heteromark/fir"
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

// WithLength sets the number of elements for the FIR benchmark.
func (b *Builder) WithLength(l int) *Builder {
	b.length = l
	return b
}

// Build creates a FIR benchmark with the given parameters.
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

	f := fir.NewBenchmark(d)
	f.Length = b.length
	f.SelectGPU([]int{1})

	bm := &Benchmark{
		name: name,
		sim:  b.sim,
		fir:  f,
	}
	return bm
}
