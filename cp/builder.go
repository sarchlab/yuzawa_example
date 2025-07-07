package cp

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

func MakeBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) WithSimulation(sim *simulation.Simulation) *Builder {
	b.sim = sim
	return b
}

func (b *Builder) WithName(name string) *Builder {
	b.name = name
	return b
}

func (b *Builder) WithFreq(freq sim.Freq) *Builder {
	b.freq = freq
	return b
}

func (b *Builder) Build(name string) *CommandProcessor {
	if b.sim == nil {
		panic("Simulation must be set before building CommandProcessor")
	}

	cp := NewCommandProcessor(
		name,
		b.engine,
		b.freq,
	)

	return cp
}