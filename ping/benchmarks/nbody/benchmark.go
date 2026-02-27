package nbody

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/nbody"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name  string
	sim   *simulation.Simulation
	nbody *nbody.Benchmark
}

func (b *Benchmark) Run() {
	if b.nbody == nil {
		log.Panic("NBody benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)
	d.Run()

	b.nbody.Run()
	log.Println("MGPUSim NBody benchmark completed")

	log.Println("Verifying NBody benchmark results...")
	b.nbody.Verify()
	log.Println("NBody benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
