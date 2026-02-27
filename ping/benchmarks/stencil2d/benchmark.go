package stencil2d

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	mgpusimstencil "github.com/sarchlab/mgpusim/v4/amd/benchmarks/shoc/stencil2d"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	stencil2d *mgpusimstencil.Benchmark
}

func (b *Benchmark) Run() {
	if b.stencil2d == nil {
		log.Panic("Stencil2D benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)

	d.Run()
	b.stencil2d.Run()
	log.Println("MGPUSim stencil2d benchmark completed")

	log.Println("Verifying stencil2d benchmark results...")
	b.stencil2d.Verify()
	log.Println("Stencil2d benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
