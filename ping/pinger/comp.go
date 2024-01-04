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

	pingReq.Meta().Src = c.port
	pingReq.Meta().Dst = e.dst.GetPortByName("PingPort")
	pingReq.Meta().SendTime = e.time

	sendError := c.port.Send(pingReq)
	if sendError != nil {
		panic(sendError)
	}

	return nil
}

// Tick updates the component state
func (c *Comp) Tick(now sim.VTimeInSec) (madeProgress bool) {
	madeProgress = c.respond(now) || madeProgress
	madeProgress = c.update(now) || madeProgress
	madeProgress = c.receive(now) || madeProgress

	return madeProgress
}

func (c *Comp) receive(now sim.VTimeInSec) bool {
	req := c.port.Retrieve(now)
	if req == nil {
		return false
	}

	pingReq := req.(*PingReq)
	c.processingMsgs = append(c.processingMsgs, &processingMsg{
		pingReq:   pingReq,
		cycleLeft: c.latency,
	})

	return false
}

func (c *Comp) update(now sim.VTimeInSec) bool {
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

func (c *Comp) respond(now sim.VTimeInSec) bool {
	madeProgress := false

	for i := 0; i < len(c.processingMsgs); i++ {
		msg := c.processingMsgs[i]
		if msg.cycleLeft != 0 {
			continue
		}

		rsp := sim.GeneralRspBuilder{}.
			WithSrc(c.port).
			WithDst(msg.pingReq.Meta().Src).
			WithSendTime(now).
			WithOriginalReq(msg.pingReq).
			Build()

		c.port.Send(rsp)
		c.processingMsgs = append(c.processingMsgs[:i], c.processingMsgs[i+1:]...)
		i--
		madeProgress = true
	}

	return madeProgress
}
