// Package pinger defines the simulation component that can send and respond to
// ping messages.
package pinger

import (
	"github.com/sarchlab/akita/v4/sim"
)

type processingMsg struct {
	pingReq   *PingReq
	cycleLeft int
}

// Comp is the simulation component that can send and respond to ping messages.
type Comp struct {
	*sim.TickingComponent

	remotePort   sim.RemotePort
	port         sim.Port
	pingProtocol *PingProtocol

	latency int

	processingMsgs []*processingMsg
}

func (c *Comp) Handle(e sim.Event) error {
	switch e := e.(type) {
	case *PingEvent:
		return c.handlePingEvent(e)
	case sim.TickEvent:
		return c.TickingComponent.Handle(e)
	default:
		panic("Unknown event type")
	}
}

func (c *Comp) handlePingEvent(e *PingEvent) error {
	pingReq, err := c.pingProtocol.CreateMsg("PingReq")
	if err != nil {
		return err
	}

	pingReq.Meta().Src = c.remotePort
	pingReq.Meta().Dst = e.dst.GetPortByName("PingPort").AsRemote()

	sendError := c.port.Send(pingReq)
	if sendError != nil {
		panic(sendError)
	}

	return nil
}

// Tick updates the component state
func (c *Comp) Tick() (madeProgress bool) {
	madeProgress = c.respond() || madeProgress
	madeProgress = c.update() || madeProgress
	madeProgress = c.receive() || madeProgress

	return madeProgress
}

func (c *Comp) receive() bool {
	msg := c.port.RetrieveIncoming()
	if msg == nil {
		return false
	}

	switch msg := msg.(type) {
	case *PingReq:
		c.processingMsgs = append(c.processingMsgs, &processingMsg{
			pingReq:   msg,
			cycleLeft: c.latency,
		})

		return true
	case *sim.GeneralRsp:
		return true
	default:
		panic("Unknown message type")
	}
}

func (c *Comp) update() bool {
	madeProgress := false

	for i := 0; i < len(c.processingMsgs); i++ {
		msg := c.processingMsgs[i]
		if msg.cycleLeft == 0 {
			continue
		}

		msg.cycleLeft--
		madeProgress = true
	}

	return madeProgress
}

func (c *Comp) respond() bool {
	madeProgress := false

	for i := 0; i < len(c.processingMsgs); i++ {
		msg := c.processingMsgs[i]
		if msg.cycleLeft != 0 {
			continue
		}

		rsp := sim.GeneralRspBuilder{}.
			WithSrc(c.remotePort).
			WithDst(msg.pingReq.Meta().Src).
			WithOriginalReq(msg.pingReq).
			Build()

		c.port.Send(rsp)
		c.processingMsgs = append(
			c.processingMsgs[:i], c.processingMsgs[i+1:]...)
		i--
		madeProgress = true
	}

	return madeProgress
}
