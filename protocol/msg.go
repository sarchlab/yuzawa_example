package protocol

import (
	"github.com/sarchlab/akita/v4/sim"
)

type OpKind uint8

const (
	InstOp OpKind = iota
	ScalarOp 
	VectorOp
)

type LaunchReq struct {
	sim.MsgMeta
	Kind OpKind
	UID string
}

type LaunchRsp struct {
	sim.MsgMeta
	UID string
}

func (m *LaunchReq) Meta() *sim.MsgMeta {
	return &m.MsgMeta 
}

func (m *LaunchReq) Clone() sim.Msg {
	c := *m
	return &c 
}

func (m *LaunchRsp) Meta() *sim.MsgMeta {
	return &m.MsgMeta 
}

func (m *LaunchRsp) Clone() sim.Msg {
	c := *m
	return &c 
}

