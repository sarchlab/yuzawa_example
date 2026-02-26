package matrixmult

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/matrixmultiplication"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	mm   *matrixmultiplication.Benchmark
}

func (b *Benchmark) Run() {
	if b.mm == nil {
		log.Panic("Matrix multiplication benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)

	d.Run()
	b.mm.Run()
	log.Println("MGPUSim matrix multiplication benchmark completed")

	log.Println("Verifying matrix multiplication benchmark results...")
	b.mm.Verify()
	log.Println("Matrix multiplication benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
