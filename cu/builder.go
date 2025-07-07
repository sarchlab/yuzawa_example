package cu

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"
)


type Builder struct {
	sim *simulation.Simulation
	name string
	freq sim.Freq
	engine sim.Engine
}

func MakeBuilder() Builder {
	return Builder{
		freq: 1 * sim.GHz,
	}
}

func (b Builder) WithSimulation(sim *simulation.Simulation) Builder {
	b.sim = sim
	return b
}

func (b Builder) WithName(name string) Builder {
	b.name = name
	return b
}

func (b Builder) WithFreq(freq sim.Freq) Builder {
	b.freq = freq
	return b
}


func (b Builder) Build(name string) *ComputeUnit {
	if b.name == "" {
		b.name = name
	}

	if b.sim == nil {
		panic("Simulation must be set before building ComputeUnit")
	}

	engine := b.sim.GetEngine()
	if b.engine == nil {
		if engine == nil {
			panic("Engine must be set before building ComputeUnit")
		}
		b.engine = engine
	}

	cu := NewComputeUnit(b.name, b.engine, b.freq)

	return cu
}
