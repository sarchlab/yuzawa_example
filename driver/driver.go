package driver

import (
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/yuzawa_example/protocol"
)

type Driver struct {
	*sim.TickingComponent
	
	memPort sim.Port 
	pendingReq []sim.Msg
	waitingReq map[string]bool
}

func (d *Driver) Enqueue(msg sim.Msg) {
	d.pendingReq = append(d.pendingReq, msg)
	if lr, ok := msg.(*protocol.LaunchReq); ok {
		d.waitingReq[lr.UID] = true
	}
}

func (d *Driver) Tick() (madeProgress bool) {
	progress := false

	// Process incoming messages
	if len(d.pendingReq) > 0 {
		if err := d.memPort.Send(d.pendingReq[0]); err == nil {
			d.pendingReq = d.pendingReq[1:]
			progress = true
		}
	}

	if rsp := d.memPort.RetrieveIncoming(); rsp != nil {
		if lrsp, ok := rsp.(*protocol.LaunchRsp); ok {
			delete(d.waitingReq, lrsp.UID)
		}
		progress = true
	}

	return progress
}

func NewDriver(name string, engine sim.Engine, freq sim.Freq) *Driver {
	d := &Driver{
		waitingReq: map[string]bool{},
	}
	d.TickingComponent = sim.NewTickingComponent(name, engine, freq, d)

	d.memPort = sim.NewPort(d, 2, 2, name+".MemPort")

	d.AddPort("Mem", d.memPort)

	return d
}



