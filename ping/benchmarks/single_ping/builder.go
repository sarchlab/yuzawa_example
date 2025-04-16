package single_ping

import "github.com/sarchlab/akita/v4/sim"

// A Builder can build a benchmark
type Builder struct {
	Name 	  string
	simulation               *sim.Simulation
	senderNames  []string 
	receiverName string
}

// NewBuilder creates a new builder
func NewBuilder() *Builder {
	return &Builder{}
}

// WithSimulation sets the simulation for the builder
func (b *Builder) WithSimulation(simulation *sim.Simulation) *Builder {
	b.simulation = simulation
	return b
}

// WithSender sets the sender for the builder
func (b *Builder) WithSender(senders []string) *Builder {
	b.senderNames = senders
	return b
}

// WithReceiver sets the receiver for the builder
func (b *Builder) WithReceiver(receiver string) *Builder {
	b.receiverName = receiver
	return b
}

// Build builds the benchmark
func (b *Builder) Build(name string) *Benchmark {
	return &Benchmark{
		Name:         name,
		simulation:   b.simulation,
		senderNames:  b.senderNames,
		receiverName: b.receiverName,
	}
}
