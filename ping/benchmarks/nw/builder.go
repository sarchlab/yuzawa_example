package nw

import (
	"fmt"

	"github.com/sarchlab/akita/v4/mem/mem"
	mgpusimnw "github.com/sarchlab/mgpusim/v4/amd/benchmarks/rodinia/nw"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
)

type Builder struct {
	sim     *simulation.Simulation
	length  int
	penalty int
}

// MakeBuilder creates a builder with default parameters.
func MakeBuilder() *Builder {
	return &Builder{
		length:  256,
		penalty: 10,
	}
}

// WithSimulation sets the simulation to use.
func (b *Builder) WithSimulation(sim *simulation.Simulation) *Builder {
	b.sim = sim
	return b
}

// WithLength sets the sequence length for Needleman-Wunsch.
func (b *Builder) WithLength(l int) *Builder {
	b.length = l
	return b
}

// WithPenalty sets the gap penalty.
func (b *Builder) WithPenalty(p int) *Builder {
	b.penalty = p
	return b
}

// Build creates an NW benchmark with the given parameters.
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

	d.RegisterGPU(commandProcessor.GetPortByName("ToDriver"), driver.DeviceProperties{
		CUCount:  cuCount,
		DRAMSize: 4 * mem.GB,
	})

	nwbm := mgpusimnw.NewBenchmark(d)
	nwbm.SetLength(b.length)
	nwbm.SetPenalty(b.penalty)
	nwbm.SelectGPU([]int{1})

	return &Benchmark{
		name: name,
		sim:  b.sim,
		nw:   nwbm,
	}
}
