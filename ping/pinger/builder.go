package pinger

import "github.com/sarchlab/akita/v4/sim"

// Builder is a builder for the Ping Component.
type Builder struct {
	name   string
	engine sim.Engine
	freq   sim.Freq
}

// NewBuilder creates a new builder.
func NewBuilder() *Builder {
	return &Builder{
		name: "Pinger",
		freq: 1 * sim.GHz,
	}
}

// WithEngine sets the engine for the builder.
func (b *Builder) WithEngine(engine sim.Engine) *Builder {
	b.engine = engine
	return b
}

// WithFreq sets the frequency for the builder.
func (b *Builder) WithFreq(freq sim.Freq) *Builder {
	b.freq = freq
	return b
}

// WithName sets the name for the component to build.
func (b *Builder) WithName(name string) *Builder {
	b.name = name
	return b
}

// Build creates a new Ping Component.
func (b *Builder) Build() *Comp {
	c := &Comp{}

	c.TickingComponent = sim.NewTickingComponent(
		b.name, b.engine, b.freq, c)

	c.pingProtocol = &PingProtocol{}

	c.port = sim.NewLimitNumMsgPort(c, 1, b.name+".PingPort")
	c.AddPort("PingPort", c.port)

	return c
}
