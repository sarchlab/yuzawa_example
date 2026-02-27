package bitonicsort

import (
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/simulation"
	mgpusimbitonic "github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/bitonicsort"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
)

type Builder struct {
	sim            *simulation.Simulation
	length         int
	orderAscending bool
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

// WithLength sets the number of elements to sort (should be power of 2).
func (b *Builder) WithLength(n int) *Builder {
	b.length = n
	return b
}

// WithOrderAscending sets sort direction (true = ascending).
func (b *Builder) WithOrderAscending(asc bool) *Builder {
	b.orderAscending = asc
	return b
}

// Build creates a BitonicSort benchmark with the given parameters.
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

	bm := mgpusimbitonic.NewBenchmark(d)
	bm.Length = b.length
	bm.OrderAscending = b.orderAscending
	bm.SelectGPU([]int{1})

	return &Benchmark{
		name:   name,
		sim:    b.sim,
		bitonic: bm,
	}
}
