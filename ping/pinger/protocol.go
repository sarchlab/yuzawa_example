package pinger

import (
	"errors"

	"github.com/sarchlab/akita/v4/sim"
)

// PingProtocol is the protocol for the ping simulation
type PingProtocol struct {
}

// CreateMsg creates a new message of the given type
func (p *PingProtocol) CreateMsg(msgType string) (sim.Msg, error) {
	switch msgType {
	case "PingReq":
		return &PingReq{
			MsgMeta: sim.MsgMeta{
				ID: sim.GetIDGenerator().Generate(),
			},
		}, nil
	case "PingRsp":
		return &sim.GeneralRsp{
			MsgMeta: sim.MsgMeta{
				ID: sim.GetIDGenerator().Generate(),
			},
		}, nil
	default:
		return nil, errors.New("Unknown message type")
	}
}

// PingReq is the request message for the ping protocol
type PingReq struct {
	MsgMeta sim.MsgMeta
}

// Meta returns the meta data associated with the message
func (m *PingReq) Meta() *sim.MsgMeta {
	return &m.MsgMeta
}
