package relu

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/dnn/layer_benchmarks/relu"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
)

type Benchmark struct {
	name string
	sim  *simulation.Simulation
	relu *relu.Benchmark
}

func (b *Benchmark) Run() {
	if b.relu == nil {
		log.Panic("ReLU benchmark not initialized!")
	}

	driverComp := b.sim.GetComponentByName("Driver")
	if driverComp == nil {
		log.Panic("Driver component not found in simulation!")
	}
	d := driverComp.(*driver.Driver)

	// Start the driver
	d.Run()

	// Run the benchmark
	b.relu.Run()
	log.Println("MGPUSim ReLU benchmark completed")

	// Verify results
	log.Println("Verifying ReLU benchmark results...")
	b.relu.Verify()
	log.Println("ReLU benchmark verification completed successfully!")

	// Terminate the driver
	d.Terminate()

	log.Println("Simulation completed")
}
