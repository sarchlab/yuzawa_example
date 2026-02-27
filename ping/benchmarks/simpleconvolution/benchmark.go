package simpleconvolution

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/simpleconvolution"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	sc   *simpleconvolution.Benchmark
}

func (b *Benchmark) Run() {
	if b.sc == nil {
		log.Panic("SimpleConvolution benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)

	d.Run()
	b.sc.Run()
	log.Println("MGPUSim simple convolution benchmark completed")

	log.Println("Verifying simple convolution benchmark results...")
	b.sc.Verify()
	log.Println("Simple convolution benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
