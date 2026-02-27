package fastwalshtransform

import (
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/fastwalshtransform"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
)

type Builder struct {
	sim    *simulation.Simulation
	length uint32
}

// MakeBuilder creates a builder with default parameters.
func MakeBuilder() *Builder {
	return &Builder{
		length: 256,
	}
}

// WithSimulation sets the simulation to use.
func (b *Builder) WithSimulation(sim *simulation.Simulation) *Builder {
	b.sim = sim
	return b
}

// WithLength sets the transform length (must be power of 2).
func (b *Builder) WithLength(l uint32) *Builder {
	b.length = l
	return b
}

// Build creates a fast Walsh transform benchmark with the given parameters.
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

	fwt := fastwalshtransform.NewBenchmark(d)
	fwt.Length = b.length
	fwt.SelectGPU([]int{1})

	return &Benchmark{
		name: name,
		sim:  b.sim,
		fwt:  fwt,
	}
}
