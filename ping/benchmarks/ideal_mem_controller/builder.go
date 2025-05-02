package ideal_mem_controller

import (
	"github.com/sarchlab/akita/v4/sim"
)

// Builder helps in setting up the memory controller simulation.
type Builder struct {
	simulation *sim.Simulation
	numAccess int
	maxAddress uint64
}

// MakeBuilder creates a new Builder.
func MakeBuilder() *Builder {
	return &Builder{}
}

// WithSimulation sets the simulation for the benchmark.
func (b *Builder) WithSimulation(s *sim.Simulation) *Builder {
	b.simulation = s
	return b
}

// WithNumAccess sets the number of memory access requests.
func (b *Builder) WithNumAccess(n int) *Builder {
	b.numAccess = n
	return b
}

// WithMaxAddress sets the maximum address for the memory access requests.
func (b *Builder) WithMaxAddress(a uint64) *Builder {
	b.maxAddress = a
	return b
}

// Build creates a new Benchmark.
func (b *Builder) Build(name string) *Benchmark {
	return &Benchmark{
		Name:              name,
		simulation: b.simulation,
		numAccess: b.numAccess,
		maxAddress: b.maxAddress,
	}
}