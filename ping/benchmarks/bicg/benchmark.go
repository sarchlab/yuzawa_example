package bicg

import (
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/polybench/bicg"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	bicg *bicg.Benchmark
}

func MakeBuilder() *Builder {
	return &Builder{}
}

type Builder struct {
	sim *simulation.Simulation
	nx  int
	ny  int
}

func (b *Builder) WithSimulation(sim *simulation.Simulation) *Builder {
	b.sim = sim
	return b
}

func (b *Builder) WithSize(nx, ny int) *Builder {
	b.nx = nx
	b.ny = ny
	return b
}

func (b *Builder) Build(name string) *Benchmark {
	driver := b.sim.GetComponentByName("Driver").(*driver.Driver)

	bm := bicg.NewBenchmark(driver)
	bm.NX = b.nx
	bm.NY = b.ny
	bm.SelectGPU([]int{1}) // Use GPU index 1 (MGPUSim driver expects IDs starting from 1)

	return &Benchmark{
		name: name,
		sim:  b.sim,
		bicg: bm,
	}
}

func (b *Benchmark) Run() {
	b.bicg.Run()
	b.bicg.Verify()
} 