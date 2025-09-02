// runner/runner.go
package runner

import (
	"log"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"
)

// Runner mirrors MGPUSim's pattern: it doesn't build hardware;
// it only runs the engine and terminates when the workload is done.
type Runner struct {
	sim    *simulation.Simulation
	engine sim.Engine
}

func New(sim *simulation.Simulation) *Runner {
	return &Runner{sim: sim, engine: sim.GetEngine()}
}

// Run starts the user-provided workload (e.g., benchmark.Run()) in a goroutine,
// then runs the Akita engine until Terminate() is called.
func (r *Runner) Run(startWork func()) {
	log.Println("Runner: Starting workload goroutine...")
	go func() {
		log.Println("Runner: Executing workload...")
		startWork()
		log.Println("Runner: Workload completed, terminating simulation...")
		r.sim.Terminate()
	}()
	log.Println("Runner: Starting engine...")
	r.engine.Run()
	log.Println("Runner: Engine finished.")
}
