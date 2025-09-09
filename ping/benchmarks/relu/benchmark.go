package relu

import (
	"log"

	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/mgpusim/v4/amd/benchmarks/dnn/layer_benchmarks/relu"
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

	b.relu.Run()
	log.Println("MGPUSim ReLU benchmark completed")

	log.Println("Verifying ReLU benchmark results...")
	b.relu.Verify()
	log.Println("ReLU benchmark verification completed successfully!")
}
