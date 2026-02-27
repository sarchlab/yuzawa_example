package fastwalshtransform

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/fastwalshtransform"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	fwt  *fastwalshtransform.Benchmark
}

func (b *Benchmark) Run() {
	if b.fwt == nil {
		log.Panic("FastWalshTransform benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)

	d.Run()
	b.fwt.Run()
	log.Println("MGPUSim fast Walsh transform benchmark completed")

	log.Println("Verifying fast Walsh transform benchmark results...")
	b.fwt.Verify()
	log.Println("Fast Walsh transform benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
