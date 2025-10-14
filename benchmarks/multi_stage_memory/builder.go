package multi_stage_memory

import (
	"github.com/sarchlab/akita/v4/mem/vm"
	"github.com/sarchlab/akita/v4/simulation"
)

// Builder helps in setting up the memory controller simulation.
type Builder struct {
	simulation        *simulation.Simulation
	numAccess         int
	maxAddress        uint64
	pageTable         *vm.PageTable
	deviceID          uint64
	useVirtualAddress bool
}

// MakeBuilder creates a new Builder.
func MakeBuilder() *Builder {
	return &Builder{
		useVirtualAddress: false, // Default to physical addressing
		deviceID:          1,
	}
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

// WithPageTable sets the page table for virtual addressing.
func (b *Builder) WithPageTable(pt *vm.PageTable) *Builder {
	b.pageTable = pt
	return b
}

// WithDeviceID sets the device ID for page mapping.
func (b *Builder) WithDeviceID(id uint64) *Builder {
	b.deviceID = id
	return b
}

// WithUseVirtualAddress sets whether to use virtual addressing.
func (b *Builder) WithUseVirtualAddress(use bool) *Builder {
	b.useVirtualAddress = use
	return b
}

// Build creates a new Benchmark.
func (b *Builder) Build(name string) *Benchmark {
	return &Benchmark{
		name:              name,
		simulation:        b.simulation,
		numAccess:         b.numAccess,
		maxAddress:        b.maxAddress,
		pageTable:         b.pageTable,
		deviceID:          b.deviceID,
		useVirtualAddress: b.useVirtualAddress,
		ioMMUName:         "IoMMU", // Set the default IoMMU name
	}
}
