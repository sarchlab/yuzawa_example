package multi_stage_memory

import (
	"github.com/sarchlab/akita/v4/simulation"
)

// Builder helps in setting up the memory controller simulation.
type Builder struct {
	simulation *simulation.Simulation
	numAccess int
	maxAddress uint64
}

// MakeBuilder creates a new Builder.
func MakeBuilder() *Builder {
	return &Builder{}
}

// WithSimulation sets the simulation for the benchmark.
func (b *Builder) WithSimulation(s *simulation.Simulation) *Builder {
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
		name:              name,
		simulation: b.simulation,
		numAccess: b.numAccess,
		maxAddress: b.maxAddress,
	}
}