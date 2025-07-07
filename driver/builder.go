package driver

import (
	"github.com/sarchlab/akita/v4/sim"
)

type Builder struct {
	engine sim.Engine
	name  string
	freq sim.Freq
}

func MakeBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) WithEngine(engine sim.Engine) *Builder {
	b.engine = engine
	return b
}

func (b *Builder) WithFreq(freq sim.Freq) *Builder {
	b.freq = freq
	return b
}

func (b *Builder) WithName(name string) *Builder {
	b.name = name
	return b
}

func (b *Builder) Build(name string) *Driver {
	if b.engine == nil {
		panic("Engine must be set before building Driver")
	}

	if b.freq == 0 {
		b.freq = 1 * sim.GHz // Default frequency
	}

	driver := NewDriver(name, b.engine, b.freq)

	return driver
}


