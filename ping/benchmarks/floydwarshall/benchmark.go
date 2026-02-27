package floydwarshall

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/floydwarshall"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	fw   *floydwarshall.Benchmark
}

func (b *Benchmark) Run() {
	if b.fw == nil {
		log.Panic("Floyd-Warshall benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)
	d.Run()

	b.fw.Run()
	log.Println("MGPUSim Floyd-Warshall benchmark completed")

	log.Println("Verifying Floyd-Warshall benchmark results...")
	b.fw.Verify()
	log.Println("Floyd-Warshall benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
