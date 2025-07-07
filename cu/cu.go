package cu

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/yuzawa_example/protocol"
)

type ComputeUnit struct {
	*sim.TickingComponent

	ctrlPort  sim.Port // <- command-processor
	instPort sim.Port // -> L1IROB
	scalarPort sim.Port // -> L1SROB
	vectorPort sim.Port // -> L1VROB
}

func (cu *ComputeUnit) Tick() bool {
	madeProgress := false

	if msg := cu.ctrlPort.RetrieveIncoming(); msg != nil {
        if req, ok := msg.(*protocol.LaunchReq); ok {
            var out sim.Port
            switch req.Kind {
            case protocol.InstOp:
                out = cu.instPort
            case protocol.ScalarOp:
                out = cu.scalarPort
            case protocol.VectorOp:
                out = cu.vectorPort
            }

            if err := out.Send(req); err == nil {
                madeProgress = true
            } else {
                cu.ctrlPort.Send(req) 
            }
        }
    }
	// if msg := cu.ctrlPort.RetrieveIncoming(); msg != nil {
	// 	return true // made progress
	// }

	for _, p := range []sim.Port{cu.instPort, cu.scalarPort, cu.vectorPort} {
        if msg := p.RetrieveIncoming(); msg != nil {
            if err := cu.ctrlPort.Send(msg); err == nil {
                madeProgress = true
            } else {
                // couldn’t send; push back to the same port
                p.Send(msg)
            }
        }
    }

	return madeProgress
	// return false 
}

func NewComputeUnit(name string, engine sim.Engine, freq sim.Freq) *ComputeUnit {
	cu := new(ComputeUnit)
	cu.TickingComponent = sim.NewTickingComponent(name, engine, freq, cu)

	cu.ctrlPort = sim.NewPort(cu, 1, 1, name+".ToCP")
	cu.instPort = sim.NewPort(cu, 1, 1, name+".ToInst")
	cu.scalarPort = sim.NewPort(cu, 1, 1, name+".ToScalar")
	cu.vectorPort = sim.NewPort(cu, 1, 1, name+".ToVector")

	cu.AddPort("Top", cu.ctrlPort)
	cu.AddPort("InstPort", cu.instPort)
	cu.AddPort("ScalarPort", cu.scalarPort)
	cu.AddPort("VectorPort", cu.vectorPort)

	return cu
}
