package cp

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/yuzawa_example/protocol"
	// "github.com/sarchlab/yuzawa_example/cu"
)

type CommandProcessor struct {
	*sim.TickingComponent

	topPort    sim.Port  // -> Driver
	reqPort    sim.Port   // CP -> CU
	rspPort    sim.Port // CU -> CP
}

func (cp *CommandProcessor) Tick() bool {
	made := false

	if msg := cp.topPort.RetrieveIncoming(); msg != nil {
		if _, ok := msg.(*protocol.LaunchReq); ok {
			if err := cp.reqPort.Send(msg); err == nil {
				made = true
			} else { 
				cp.topPort.Send(msg)
			}
		}
	}

	if msg := cp.rspPort.RetrieveIncoming(); msg != nil {
		
		if err := cp.topPort.Send(msg); err == nil {
			made = true
		} else { // can’t send this cycle
			cp.rspPort.Send(msg)
		}
	}

	// if cu.vectorPort.CanSend() {
	// 	_ = cu.vectorPort.Send(msg)
	// }
	
	return made
}

func NewCommandProcessor(name string, eng sim.Engine, freq sim.Freq) *CommandProcessor {
    cp := &CommandProcessor{}
    cp.TickingComponent = sim.NewTickingComponent(name, eng, freq, cp)

    cp.topPort  = sim.NewPort(cp, 1, 1, name+".Top")
    cp.reqPort  = sim.NewPort(cp, 1, 1, name+".Req")
    cp.rspPort  = sim.NewPort(cp, 1, 1, name+".Ret")

    cp.AddPort("Top",  cp.topPort)
    cp.AddPort("RequestPort",  cp.reqPort)
    cp.AddPort("ResponsePort",  cp.rspPort)
    return cp
}
