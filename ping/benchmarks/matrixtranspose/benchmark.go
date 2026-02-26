package matrixtranspose

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/matrixtranspose"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	mt   *matrixtranspose.Benchmark
}

func (b *Benchmark) Run() {
	if b.mt == nil {
		log.Panic("Matrix transpose benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)

	d.Run()
	b.mt.Run()
	log.Println("MGPUSim matrix transpose benchmark completed")

	log.Println("Verifying matrix transpose benchmark results...")
	b.mt.Verify()
	log.Println("Matrix transpose benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
