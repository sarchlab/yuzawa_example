package fir

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/heteromark/fir"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	fir  *fir.Benchmark
}

func (b *Benchmark) Run() {
	if b.fir == nil {
		log.Panic("FIR benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)

	d.Run()
	b.fir.Run()
	log.Println("MGPUSim FIR benchmark completed")

	log.Println("Verifying FIR benchmark results...")
	b.fir.Verify()
	log.Println("FIR benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
