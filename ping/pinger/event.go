package pinger

import "github.com/sarchlab/akita/v4/sim"

// A PingEvent triggers a PingReq to be sent.
type PingEvent struct {
	src, dst sim.Component
	time     sim.VTimeInSec
}

// NewPingEvent creates a new PingEvent.
func NewPingEvent(src, dst sim.Component, time sim.VTimeInSec) *PingEvent {
	return &PingEvent{
		src:  src,
		dst:  dst,
		time: time,
	}
}

// Time returns the time when the event should be triggered.
func (e *PingEvent) Time() sim.VTimeInSec {
	return e.time
}

// Handler returns the handler that should handle the event.
func (e *PingEvent) Handler() sim.Handler {
	return e.src
}

// IsSecondary always returns false.
func (e *PingEvent) IsSecondary() bool {
	return false
}
