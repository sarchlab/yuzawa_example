package rob

import (
	"container/list"

	"github.com/sarchlab/akita/v4/sim"
)

// A Builder can build ReorderBuffers.
type Builder struct {
	engine         sim.Engine
	freq           sim.Freq
	numReqPerCycle int
	bufferSize     int
	bottomUnit     sim.Port
}

// MakeBuilder creates a builder with default parameters.
func MakeBuilder() Builder {
	return Builder{
		freq:           1 * sim.GHz,
		numReqPerCycle: 4,
		bufferSize:     128,
	}
}

// WithEngine sets the engine to use.
func (b Builder) WithEngine(engine sim.Engine) Builder {
	b.engine = engine
	return b
}

// WithFreq sets the frequency that the ReorderBuffer works at.
func (b Builder) WithFreq(freq sim.Freq) Builder {
	b.freq = freq
	return b
}

// WithNumReqPerCycle sets the number of request that the ReorderBuffer can
// handle in each cycle.
func (b Builder) WithNumReqPerCycle(n int) Builder {
	b.numReqPerCycle = n
	return b
}

// WithBufferSize sets the number of transactions that the buffer can handle.
func (b Builder) WithBufferSize(n int) Builder {
	b.bufferSize = n
	return b
}

func (b Builder) WithBottomUnit(port sim.Port) Builder { 
    b.bottomUnit = port
    return b
}

// Build creates a ReorderBuffer with the given parameters.
func (b Builder) Build(name string) *ReorderBuffer {
	rb := &ReorderBuffer{}

	rb.TickingComponent = sim.NewTickingComponent(name, b.engine, b.freq, rb)

	rb.transactions = list.New()
	rb.transactions.Init()
	rb.toBottomReqIDToTransactionTable = make(map[string]*list.Element)

	rb.bufferSize = b.bufferSize
	rb.numReqPerCycle = b.numReqPerCycle

	b.createPorts(name, rb)

	if b.bottomUnit != nil {           
		rb.bottomUnit = b.bottomUnit
	}

	return rb
}

func (b *Builder) createPorts(name string, rb *ReorderBuffer) {
	rb.topPort = sim.NewPort(
		rb,
		2*b.numReqPerCycle,
		2*b.numReqPerCycle,
		name+".TopPort",
	)
	rb.AddPort("Top", rb.topPort)

	rb.bottomPort = sim.NewPort(
		rb,
		2*b.numReqPerCycle,
		2*b.numReqPerCycle,
		name+".BottomPort",
	)
	rb.AddPort("Bottom", rb.bottomPort)

	rb.controlPort = sim.NewPort(
		rb,
		1,
		1,
		name+".ControlPort",
	)
	rb.AddPort("Control", rb.controlPort)
}