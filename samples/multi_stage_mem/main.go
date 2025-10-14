package main

import (
	"log"
	"os"

	"github.com/sarchlab/akita/v4/mem/cache/writearound"
	"github.com/sarchlab/akita/v4/mem/cache/writethrough"
	"github.com/sarchlab/akita/v4/mem/idealmemcontroller"
	"github.com/sarchlab/akita/v4/mem/mem"
	"github.com/sarchlab/akita/v4/mem/trace"
	"github.com/sarchlab/akita/v4/mem/vm"
	"github.com/sarchlab/akita/v4/mem/vm/mmu"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/rob"

	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/simulation"

	"github.com/sarchlab/akita/v4/mem/vm/addresstranslator"
	"github.com/sarchlab/akita/v4/mem/vm/tlb"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/yuzawa_example/benchmarks/multi_stage_memory"
	"github.com/sarchlab/yuzawa_example/components/memaccessagent"
)

func main() {
	s := simulation.MakeBuilder().Build()
	engine := s.GetEngine()

	sharedStorage := mem.NewStorage(16 * mem.GB)

	MemCtrl := idealmemcontroller.MakeBuilder().
		WithEngine(engine).
		WithStorage(sharedStorage).
		WithLatency(10).
		Build("MemCtrl")
	s.RegisterComponent(MemCtrl)

	L2Cache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).
		WithLog2BlockSize(6).
		WithAddressMapperType("single").
		WithRemotePorts(MemCtrl.GetPortByName("Top").AsRemote()).
		Build("L2Cache")
	s.RegisterComponent(L2Cache)

	L1Cache := writearound.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).
		WithNumBanks(1).
		WithLog2BlockSize(6).
		WithTotalByteSize(16 * mem.KB).
		WithBankLatency(60).
		WithNumMSHREntry(16).
		WithAddressMapperType("single").
		WithRemotePorts(L2Cache.GetPortByName("Top").AsRemote()).
		Build("L1Cache")
	s.RegisterComponent(L1Cache)

	pageTable := vm.NewPageTable(12)
	IoMMU := mmu.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithMaxNumReqInFlight(16).
		WithPageWalkingLatency(10).
		WithPageTable(pageTable).
		Build("IoMMU")
	s.RegisterComponent(IoMMU)

	L2TLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(64).
		WithNumSets(64).
		WithLog2PageSize(12).
		WithNumReqPerCycle(4).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(IoMMU.GetPortByName("Top").AsRemote()).
		Build("L2TLB")
	s.RegisterComponent(L2TLB)

	TLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithLog2PageSize(12).
		WithNumReqPerCycle(2).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(L2TLB.GetPortByName("Top").AsRemote()).
		Build("TLB")
	s.RegisterComponent(TLB)

	AT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithDeviceID(1).
		WithTranslationProviders(TLB.GetPortByName("Top").AsRemote()).
		WithMemoryProviderType("single").
		WithMemoryProviders(L1Cache.GetPortByName("Top").AsRemote()).
		WithTranslationProviderMapperType("single").
		Build("AT")
	s.RegisterComponent(AT)

	ROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		WithBottomUnit(AT.GetPortByName("Top").AsRemote()).
		Build("ROB")
	s.RegisterComponent(ROB)

	MemAgent := memaccessagent.MakeBuilder().
		WithFreq(1 * sim.GHz).
		WithMaxAddress(1 * mem.GB).
		WithWriteLeft(100000).
		WithReadLeft(100000).
		WithEngine(engine).
		WithLowModule(ROB.GetPortByName("Top")).
		Build("MemAgent")
	s.RegisterComponent(MemAgent)

	Driver := driver.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithGlobalStorage(sharedStorage).
		WithPageTable(pageTable).
		WithLog2PageSize(12).
		WithMagicMemoryCopyMiddleware().
		Build("Driver")
	s.RegisterComponent(Driver)

	ConnMemAgentToROB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnMemAgentToROB")
	ConnMemAgentToROB.PlugIn(MemAgent.GetPortByName("Mem"))
	ConnMemAgentToROB.PlugIn(ROB.GetPortByName("Top"))

	ConnROBToAT := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnROBToAT")
	ConnROBToAT.PlugIn(ROB.GetPortByName("Bottom"))
	ConnROBToAT.PlugIn(AT.GetPortByName("Top"))

	ConnATTranslationToTLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnATTranslationToTLB")
	ConnATTranslationToTLB.PlugIn(AT.GetPortByName("Translation"))
	ConnATTranslationToTLB.PlugIn(TLB.GetPortByName("Top"))

	ConnATToL1Cache := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnATToL1Cache")
	ConnATToL1Cache.PlugIn(AT.GetPortByName("Bottom"))
	ConnATToL1Cache.PlugIn(L1Cache.GetPortByName("Top"))

	ConnTLBToL2TLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnTLBToL2TLB")
	ConnTLBToL2TLB.PlugIn(TLB.GetPortByName("Bottom"))
	ConnTLBToL2TLB.PlugIn(L2TLB.GetPortByName("Top"))

	ConnL2TLBToIoMMU := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL2TLBToIoMMU")
	ConnL2TLBToIoMMU.PlugIn(L2TLB.GetPortByName("Bottom"))
	ConnL2TLBToIoMMU.PlugIn(IoMMU.GetPortByName("Top"))

	ConnL1ToL2 := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL1ToL2")
	ConnL1ToL2.PlugIn(L1Cache.GetPortByName("Bottom"))
	ConnL1ToL2.PlugIn(L2Cache.GetPortByName("Top"))

	ConnL2ToMemCtrl := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL2ToMemCtrl")
	ConnL2ToMemCtrl.PlugIn(L2Cache.GetPortByName("Bottom"))
	ConnL2ToMemCtrl.PlugIn(MemCtrl.GetPortByName("Top"))

	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	logger := log.New(traceFile, "", 0)
	tracer := trace.NewTracer(logger, engine)
	tracing.CollectTrace(MemCtrl, tracer)

	benchmark := multi_stage_memory.MakeBuilder().
		WithSimulation(s).
		WithNumAccess(100000).
		WithMaxAddress(1 * mem.GB).
		Build("Benchmark")
	benchmark.Run()
	s.Terminate()
}
