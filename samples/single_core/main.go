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
	"github.com/sarchlab/akita/v4/mem/vm/addresstranslator"
	"github.com/sarchlab/akita/v4/mem/vm/mmu"
	"github.com/sarchlab/akita/v4/mem/vm/tlb"
	"github.com/sarchlab/akita/v4/sim"
	"github.com/sarchlab/akita/v4/sim/directconnection"
	"github.com/sarchlab/akita/v4/simulation"
	"github.com/sarchlab/akita/v4/tracing"
	"github.com/sarchlab/mgpusim/v4/amd/driver"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cu"
	"github.com/sarchlab/mgpusim/v4/amd/timing/rob"
	"github.com/sarchlab/yuzawa_example/benchmarks/relu"
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

	L1VCache := writearound.MakeBuilder().
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
		Build("L1VCache")
	s.RegisterComponent(L1VCache)

	L1SCache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(2).
		WithLog2BlockSize(6).
		WithAddressMapperType("single").
		WithRemotePorts(L2Cache.GetPortByName("Top").AsRemote()).
		Build("L1SCache")
	s.RegisterComponent(L1SCache)

	L1ICache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(2).
		WithLog2BlockSize(6).
		WithAddressMapperType("single").
		WithRemotePorts(L2Cache.GetPortByName("Top").AsRemote()).
		Build("L1ICache")
	s.RegisterComponent(L1ICache)

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

	VTLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithLog2PageSize(12).
		WithNumReqPerCycle(2).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(L2TLB.GetPortByName("Top").AsRemote()).
		Build("VTLB")
	s.RegisterComponent(VTLB)

	STLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithLog2PageSize(12).
		WithNumReqPerCycle(2).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(L2TLB.GetPortByName("Top").AsRemote()).
		Build("STLB")
	s.RegisterComponent(STLB)

	ITLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithLog2PageSize(12).
		WithNumReqPerCycle(2).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(L2TLB.GetPortByName("Top").AsRemote()).
		Build("ITLB")
	s.RegisterComponent(ITLB)

	VAT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithDeviceID(1).
		WithTranslationProviders(VTLB.GetPortByName("Top").AsRemote()).
		WithMemoryProviderType("single").
		WithMemoryProviders(L1VCache.GetPortByName("Top").AsRemote()).
		WithTranslationProviderMapperType("single").
		Build("VAT")
	s.RegisterComponent(VAT)

	SAT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithDeviceID(1).
		WithTranslationProviders(STLB.GetPortByName("Top").AsRemote()).
		WithMemoryProviderType("single").
		WithMemoryProviders(L1SCache.GetPortByName("Top").AsRemote()).
		WithTranslationProviderMapperType("single").
		Build("SAT")
	s.RegisterComponent(SAT)

	IAT := addresstranslator.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithLog2PageSize(12).
		WithDeviceID(1).
		WithTranslationProviders(ITLB.GetPortByName("Top").AsRemote()).
		WithMemoryProviderType("single").
		WithMemoryProviders(L1ICache.GetPortByName("Top").AsRemote()).
		WithTranslationProviderMapperType("single").
		Build("IAT")
	s.RegisterComponent(IAT)

	VROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		WithBottomUnit(VAT.GetPortByName("Top").AsRemote()).
		Build("VROB")
	s.RegisterComponent(VROB)

	SROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		WithBottomUnit(SAT.GetPortByName("Top").AsRemote()).
		Build("SROB")
	s.RegisterComponent(SROB)

	IROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		WithBottomUnit(IAT.GetPortByName("Top").AsRemote()).
		Build("IROB")
	s.RegisterComponent(IROB)

	CU := cu.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithVGPRCount([]int{32768, 32768, 32768, 32768}).
		WithInstMem(IROB.GetPortByName("Top")).
		WithScalarMem(SROB.GetPortByName("Top")).
		WithVectorMemModules(&mem.SinglePortMapper{
			Port: VROB.GetPortByName("Top").AsRemote(),
		}).
		WithSIMDCount(4).
		Build("CU")
	s.RegisterComponent(CU)

	Driver := driver.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithGlobalStorage(sharedStorage).
		WithPageTable(pageTable).
		WithLog2PageSize(12).
		WithMagicMemoryCopyMiddleware().
		Build("Driver")
	s.RegisterComponent(Driver)

	CP := cp.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithDriver(Driver.GetPortByName("GPU")).
		WithCU(CU).
		Build("CP")
	s.RegisterComponent(CP)

	ConnGPU1 := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnGPU1")
	ConnGPU1.PlugIn(CP.GetPortByName("ToDriver"))
	ConnGPU1.PlugIn(Driver.GetPortByName("GPU"))

	ConnCPToCU := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCPToCU")
	ConnCPToCU.PlugIn(CP.GetPortByName("ToCUs"))
	ConnCPToCU.PlugIn(CU.GetPortByName("Top"))
	ConnCPToCU.PlugIn(CU.GetPortByName("Ctrl"))

	ConnCUToVROB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCUToVROB")
	ConnCUToVROB.PlugIn(CU.GetPortByName("VectorMem"))
	ConnCUToVROB.PlugIn(VROB.GetPortByName("Top"))

	ConnCUToSROB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCUToSROB")
	ConnCUToSROB.PlugIn(CU.GetPortByName("ScalarMem"))
	ConnCUToSROB.PlugIn(SROB.GetPortByName("Top"))

	ConnCUToIROB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCUToIROB")
	ConnCUToIROB.PlugIn(CU.GetPortByName("InstMem"))
	ConnCUToIROB.PlugIn(IROB.GetPortByName("Top"))

	ConnVROBToVAT := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnVROBToVAT")
	ConnVROBToVAT.PlugIn(VROB.GetPortByName("Bottom"))
	ConnVROBToVAT.PlugIn(VAT.GetPortByName("Top"))

	ConnSROBToSAT := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnSROBToSAT")
	ConnSROBToSAT.PlugIn(SROB.GetPortByName("Bottom"))
	ConnSROBToSAT.PlugIn(SAT.GetPortByName("Top"))

	ConnIROBToIAT := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnIROBToIAT")
	ConnIROBToIAT.PlugIn(IROB.GetPortByName("Bottom"))
	ConnIROBToIAT.PlugIn(IAT.GetPortByName("Top"))

	ConnVATTranslationToVTLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnVATTranslationToVTLB")
	ConnVATTranslationToVTLB.PlugIn(VAT.GetPortByName("Translation"))
	ConnVATTranslationToVTLB.PlugIn(VTLB.GetPortByName("Top"))

	ConnSATTranslationToSTLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnSATTranslationToSTLB")
	ConnSATTranslationToSTLB.PlugIn(SAT.GetPortByName("Translation"))
	ConnSATTranslationToSTLB.PlugIn(STLB.GetPortByName("Top"))

	ConnIATTranslationToITLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnIATTranslationToITLB")
	ConnIATTranslationToITLB.PlugIn(IAT.GetPortByName("Translation"))
	ConnIATTranslationToITLB.PlugIn(ITLB.GetPortByName("Top"))

	ConnVATToL1VCache := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnVATToL1VCache")
	ConnVATToL1VCache.PlugIn(VAT.GetPortByName("Bottom"))
	ConnVATToL1VCache.PlugIn(L1VCache.GetPortByName("Top"))

	ConnSATToL1SCache := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnSATToL1SCache")
	ConnSATToL1SCache.PlugIn(SAT.GetPortByName("Bottom"))
	ConnSATToL1SCache.PlugIn(L1SCache.GetPortByName("Top"))

	ConnIATToL1ICache := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnIATToL1ICache")
	ConnIATToL1ICache.PlugIn(IAT.GetPortByName("Bottom"))
	ConnIATToL1ICache.PlugIn(L1ICache.GetPortByName("Top"))

	ConnL1ToL2 := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL1ToL2")
	ConnL1ToL2.PlugIn(L1VCache.GetPortByName("Bottom"))
	ConnL1ToL2.PlugIn(L1SCache.GetPortByName("Bottom"))
	ConnL1ToL2.PlugIn(L1ICache.GetPortByName("Bottom"))
	ConnL1ToL2.PlugIn(L2Cache.GetPortByName("Top"))

	ConnL2AndDMAToMemCtrl := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL2AndDMAToMemCtrl")
	ConnL2AndDMAToMemCtrl.PlugIn(L2Cache.GetPortByName("Bottom"))
	ConnL2AndDMAToMemCtrl.PlugIn(MemCtrl.GetPortByName("Top"))

	ConnTLBToL2TLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnVTLBToL2TLB")
	ConnTLBToL2TLB.PlugIn(VTLB.GetPortByName("Bottom"))
	ConnTLBToL2TLB.PlugIn(STLB.GetPortByName("Bottom"))
	ConnTLBToL2TLB.PlugIn(ITLB.GetPortByName("Bottom"))
	ConnTLBToL2TLB.PlugIn(L2TLB.GetPortByName("Top"))

	ConnL2TLBToIoMMU := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL2TLBToIoMMU")
	ConnL2TLBToIoMMU.PlugIn(L2TLB.GetPortByName("Bottom"))
	ConnL2TLBToIoMMU.PlugIn(IoMMU.GetPortByName("Top"))

	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	logger := log.New(traceFile, "", 0)
	tracer := trace.NewTracer(logger, engine)
	tracing.CollectTrace(MemCtrl, tracer)

	benchmark := relu.MakeBuilder().
		WithSimulation(s).
		WithLength(4).
		Build("ReLU")

	benchmark.Run()
	s.Terminate()
}
