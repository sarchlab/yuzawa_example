package atax

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/polybench/atax"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	atax *atax.Benchmark
}

func (b *Benchmark) Run() {
	if b.atax == nil {
		log.Panic("ATAX benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)
	d.Run()

	b.atax.Run()
	log.Println("MGPUSim ATAX benchmark completed")

	log.Println("Verifying ATAX benchmark results...")
	b.atax.Verify()
	log.Println("ATAX benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
