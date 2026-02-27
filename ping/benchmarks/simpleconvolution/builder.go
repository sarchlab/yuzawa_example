package simpleconvolution

import (
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/simpleconvolution"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
)

type Builder struct {
	sim      *simulation.Simulation
	width    uint32
	height   uint32
	maskSize uint32
}

// MakeBuilder creates a builder with default parameters.
func MakeBuilder() *Builder {
	return &Builder{
		width:    256,
		height:   256,
		maskSize: 3,
	}
}

// WithSimulation sets the simulation to use.
func (b *Builder) WithSimulation(sim *simulation.Simulation) *Builder {
	b.sim = sim
	return b
}

// WithWidth sets the image width.
func (b *Builder) WithWidth(w uint32) *Builder {
	b.width = w
	return b
}

// WithHeight sets the image height.
func (b *Builder) WithHeight(h uint32) *Builder {
	b.height = h
	return b
}

// WithMaskSize sets the convolution mask size (e.g. 3 for 3x3).
func (b *Builder) WithMaskSize(m uint32) *Builder {
	b.maskSize = m
	return b
}

// Build creates a simple convolution benchmark with the given parameters.
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

	sc := simpleconvolution.NewBenchmark(d)
	sc.Width = b.width
	sc.Height = b.height
	sc.SetMaskSize(b.maskSize)
	sc.SelectGPU([]int{1})

	return &Benchmark{
		name: name,
		sim:  b.sim,
		sc:   sc,
	}
}
