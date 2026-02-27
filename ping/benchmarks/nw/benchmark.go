package nw

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	mgpusimnw "github.com/sarchlab/mgpusim/v4/amd/benchmarks/rodinia/nw"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	nw   *mgpusimnw.Benchmark
}

func (b *Benchmark) Run() {
	if b.nw == nil {
		log.Panic("NW (Needleman-Wunsch) benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)

	d.Run()
	b.nw.Run()
	log.Println("MGPUSim NW benchmark completed")

	log.Println("Verifying NW benchmark results...")
	b.nw.Verify()
	log.Println("NW benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
