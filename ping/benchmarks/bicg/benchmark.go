package bicg

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/polybench/bicg"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	bicg *bicg.Benchmark
}

func (b *Benchmark) Run() {
	if b.bicg == nil {
		log.Panic("BICG benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)
	d.Run()

	b.bicg.Run()
	log.Println("MGPUSim BICG benchmark completed")

	log.Println("Verifying BICG benchmark results...")
	b.bicg.Verify()
	log.Println("BICG benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
