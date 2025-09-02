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
	// Don't call driver.Run() and driver.Terminate() here - handled in main.go
	log.Printf("Running ReLU benchmark with length: %d", b.relu.Length)
	log.Println("About to call underlying MGPUSim ReLU benchmark...")
	b.relu.Run()
	log.Println("Underlying MGPUSim ReLU benchmark completed")
}

// GetUnderlyingBenchmark returns the underlying MGPUSim benchmark
func (b *Benchmark) GetUnderlyingBenchmark() *relu.Benchmark {
	return b.relu
}
