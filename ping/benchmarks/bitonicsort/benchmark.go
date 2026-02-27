package bitonicsort

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	mgpusimbitonic "github.com/sarchlab/mgpusim/v4/amd/benchmarks/amdappsdk/bitonicsort"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/yuzawa_example/metrics_reporter"
)

type Benchmark struct {
	name    string
	sim     *simulation.Simulation
	bitonic *mgpusimbitonic.Benchmark
}

func (b *Benchmark) Run() {
	if b.bitonic == nil {
		log.Panic("BitonicSort benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	metricsReporter := metrics_reporter.NewReporter(b.sim)
	d.Run()

	b.bitonic.Run()
	log.Println("MGPUSim BitonicSort benchmark completed")

	log.Println("Verifying BitonicSort benchmark results...")
	b.bitonic.Verify()
	log.Println("BitonicSort benchmark verification completed successfully!")

	d.Terminate()
	metricsReporter.Report()
	log.Println("Simulation completed")
}
