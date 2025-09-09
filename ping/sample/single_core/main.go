package main

import (
	"log"
	"os"
	"time"

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
	"github.com/sarchlab/mgpusim/v4/amd/emu"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cp"
	"github.com/sarchlab/mgpusim/v4/amd/timing/cu"
	"github.com/sarchlab/mgpusim/v4/amd/timing/rob"
	"github.com/sarchlab/yuzawa_example/ping/benchmarks/relu"
)

func main() {
	log.Printf("Starting main function...")

	// Build simulation & engine
	log.Printf("Building simulation...")
	s := simulation.MakeBuilder().Build()
	engine := s.GetEngine()
	log.Printf("Simulation built successfully")

	// Create shared storage for all components to ensure memory consistency
	log.Printf("Creating shared storage...")
	sharedStorage := mem.NewStorage(16 * mem.GB)
	log.Printf("Shared storage created")

	// Memory hierarchy components
	log.Printf("Creating memory controller...")
	MemCtrl := idealmemcontroller.MakeBuilder().
		WithEngine(engine).
		WithStorage(sharedStorage).
		WithLatency(10). // Reduced from 100 to 10 cycles for faster simulation
		Build("MemCtrl")
	s.RegisterComponent(MemCtrl)
	log.Printf("Memory controller created and registered")

	log.Printf("Creating L2 cache (write-through)...")
	L2Cache := writethrough.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithWayAssociativity(4).
		WithLog2BlockSize(6).
		WithAddressMapperType("single").
		WithRemotePorts(MemCtrl.GetPortByName("Top").AsRemote()).
		Build("L2Cache")
	s.RegisterComponent(L2Cache)

	log.Printf("Creating L1 VCache...")
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

	log.Printf("Creating L2 TLB...")
	// GPU-specific memory management components
	L2TLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(64).
		WithNumSets(64).
		WithLog2PageSize(4096).
		WithNumReqPerCycle(4).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(IoMMU.GetPortByName("Top").AsRemote()).
		Build("L2TLB")
	s.RegisterComponent(L2TLB)

	log.Printf("Creating VTLB...")
	VTLB := tlb.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumWays(8).
		WithNumSets(8).
		WithLog2PageSize(4096).
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
		WithLog2PageSize(4096).
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
		WithLog2PageSize(4096).
		WithNumReqPerCycle(2).
		WithTranslationProviderMapperType("single").
		WithTranslationProviders(L2TLB.GetPortByName("Top").AsRemote()).
		Build("ITLB")
	s.RegisterComponent(ITLB)

	// Create Address Translators after TLBs are ready
	log.Printf("Creating VAT...")
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

	log.Printf("Creating SAT...")
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

	log.Printf("Creating IAT...")
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

	log.Printf("Creating VROB...")
	// Reorder Buffers
	VROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		Build("VROB")
	s.RegisterComponent(VROB)

	// Set BottomUnit for VROB
	VROB.BottomUnit = VAT.GetPortByName("Top")

	// VectorMemModules will be configured after CU is created

	log.Printf("Creating SROB...")
	SROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		Build("SROB")
	s.RegisterComponent(SROB)

	// Set BottomUnit for SROB
	SROB.BottomUnit = SAT.GetPortByName("Top")

	log.Printf("Creating IROB...")
	IROB := rob.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithNumReqPerCycle(4).
		WithBufferSize(128).
		Build("IROB")
	s.RegisterComponent(IROB)

	// Set BottomUnit for IROB
	IROB.BottomUnit = IAT.GetPortByName("Top")

	storageAccessor := emu.NewStorageAccessor(sharedStorage, pageTable, 12, nil)

	log.Printf("Creating CU...")
	CU := cu.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithVGPRCount([]int{32768, 32768, 32768, 32768}).
		WithStorageAccessor(storageAccessor). // This is the key addition!
		WithCPIntegration().                  // Enable CP integration
		Build("CU")
	s.RegisterComponent(CU)

	// WithInstMem(IROB.GetPortByName("Top")).

	// Configure VectorMemModules to map all addresses to the VROB port
	// This is critical for the compute unit to access memory
	// Note: We'll set this after VROB is created to avoid circular reference

	// Configure ScalarMem to point to the SROB's top port
	// This is required for the ScalarUnit to send scalar memory requests
	CU.ScalarMem = SROB.GetPortByName("Top")

	// Configure VectorMemModules to send through VROB (CU → VROB → VAT → caches)
	CU.VectorMemModules = &mem.SinglePortMapper{
		Port: VROB.GetPortByName("Top").AsRemote(),
	}

	log.Printf("Creating CP...")
	// Command Processor
	CP := cp.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("CP")
		// WithNumDispatchers(1). // Single dispatcher for single core
		// Build("CP")
	s.RegisterComponent(CP)

	log.Printf("Creating DMA Engine...")
	// Create DMA Engine for CP with a simple configuration
	localDataSource := &mem.SinglePortMapper{
		Port: MemCtrl.GetPortByName("Top").AsRemote(),
	}
	DMAEngine := cp.NewDMAEngine("DMAEngine", engine, localDataSource)
	s.RegisterComponent(DMAEngine)

	// Set DMA Engine in CP
	CP.DMAEngine = DMAEngine.ToCP
	log.Printf("DMA Engine set in CP")

	// Connect DMA Engine to CP port
	ConnCPToDMA := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCPToDMA")
	ConnCPToDMA.PlugIn(CP.GetPortByName("ToDispatcher"))
	ConnCPToDMA.PlugIn(DMAEngine.ToCP)
	log.Printf("DMA Engine connected to CP port")

	log.Printf("Creating Driver...")
	// Driver
	Driver := driver.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithGlobalStorage(sharedStorage).
		WithPageTable(pageTable).
		WithLog2PageSize(12).
		Build("Driver")
	s.RegisterComponent(Driver)
	log.Printf("Driver created and registered")

	// Configure Driver port for CP to send responses back
	// This is required for the Command Processor to communicate with the Driver
	CP.Driver = Driver.GetPortByName("GPU")

	// Register CP as GPU with Driver
	Driver.RegisterGPU(CP.GetPortByName("ToDriver"), driver.DeviceProperties{
		CUCount:  1,
		DRAMSize: 4 * mem.GB,
	})
	log.Printf("CP registered as GPU with Driver")

	CP.Driver = Driver.GetPortByName("GPU")

	log.Printf("Creating ConnGPU1...")
	ConnGPU1 := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnGPU1")
	ConnGPU1.PlugIn(CP.GetPortByName("ToDriver"))
	ConnGPU1.PlugIn(Driver.GetPortByName("GPU"))

	log.Printf("Registering CU with CP...")
	CP.RegisterCU(CU)
	log.Printf("CU registered with CP")

	log.Printf("Creating ConnCPToCU...")
	ConnCPToCU := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCPToCU")
	ConnCPToCU.PlugIn(CP.GetPortByName("ToCUs"))
	ConnCPToCU.PlugIn(CU.GetPortByName("Top"))
	ConnCPToCU.PlugIn(CU.GetPortByName("Ctrl"))

	log.Printf("Creating ConnCUToVROB...")
	ConnCUToVROB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCUToVROB")
	ConnCUToVROB.PlugIn(CU.GetPortByName("VectorMem"))
	ConnCUToVROB.PlugIn(VROB.GetPortByName("Top"))

	log.Printf("Creating ConnCUToSROB...")
	ConnCUToSROB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCUToSROB")
	ConnCUToSROB.PlugIn(CU.GetPortByName("ScalarMem"))
	ConnCUToSROB.PlugIn(SROB.GetPortByName("Top"))

	log.Printf("Creating ConnCUToIROB...")
	ConnCUToIROB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnCUToIROB")
	ConnCUToIROB.PlugIn(CU.GetPortByName("InstMem"))
	ConnCUToIROB.PlugIn(IROB.GetPortByName("Top"))

	log.Printf("Creating ConnVROBToVAT...")
	ConnVROBToVAT := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnVROBToVAT")
	ConnVROBToVAT.PlugIn(VROB.GetPortByName("Bottom"))
	ConnVROBToVAT.PlugIn(VAT.GetPortByName("Top"))

	log.Printf("Creating ConnSROBToSAT...")
	ConnSROBToSAT := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnSROBToSAT")
	ConnSROBToSAT.PlugIn(SROB.GetPortByName("Bottom"))
	ConnSROBToSAT.PlugIn(SAT.GetPortByName("Top"))

	log.Printf("Creating ConnIROBToIAT...")
	ConnIROBToIAT := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnIROBToIAT")
	ConnIROBToIAT.PlugIn(IROB.GetPortByName("Bottom"))
	ConnIROBToIAT.PlugIn(IAT.GetPortByName("Top"))

	log.Printf("Creating ConnVATTranslationToVTLB...")
	ConnVATTranslationToVTLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnVATTranslationToVTLB")
	ConnVATTranslationToVTLB.PlugIn(VAT.GetPortByName("Translation"))
	ConnVATTranslationToVTLB.PlugIn(VTLB.GetPortByName("Top"))

	log.Printf("Creating ConnSATTranslationToSTLB...")
	ConnSATTranslationToSTLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnSATTranslationToSTLB")
	ConnSATTranslationToSTLB.PlugIn(SAT.GetPortByName("Translation"))
	ConnSATTranslationToSTLB.PlugIn(STLB.GetPortByName("Top"))

	log.Printf("Creating ConnIATTranslationToITLB...")
	ConnIATTranslationToITLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnIATTranslationToITLB")
	ConnIATTranslationToITLB.PlugIn(IAT.GetPortByName("Translation"))
	ConnIATTranslationToITLB.PlugIn(ITLB.GetPortByName("Top"))

	log.Printf("Creating ConnVATToL1VCache...")
	ConnVATToL1VCache := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnVATToL1VCache")
	ConnVATToL1VCache.PlugIn(VAT.GetPortByName("Bottom"))
	ConnVATToL1VCache.PlugIn(L1VCache.GetPortByName("Top"))

	log.Printf("Creating ConnSATToL1SCache...")
	ConnSATToL1SCache := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnSATToL1SCache")
	ConnSATToL1SCache.PlugIn(SAT.GetPortByName("Bottom"))
	ConnSATToL1SCache.PlugIn(L1SCache.GetPortByName("Top"))

	log.Printf("Creating ConnIATToL1ICache...")
	ConnIATToL1ICache := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnIATToL1ICache")
	ConnIATToL1ICache.PlugIn(IAT.GetPortByName("Bottom"))
	ConnIATToL1ICache.PlugIn(L1ICache.GetPortByName("Top"))

	log.Printf("Creating ConnL1ToL2...")
	ConnL1ToL2 := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL1ToL2")
	ConnL1ToL2.PlugIn(L1VCache.GetPortByName("Bottom"))
	ConnL1ToL2.PlugIn(L1SCache.GetPortByName("Bottom"))
	ConnL1ToL2.PlugIn(L1ICache.GetPortByName("Bottom"))
	ConnL1ToL2.PlugIn(L2Cache.GetPortByName("Top"))

	log.Printf("Creating ConnL2AndDMAToMemCtrl...")
	ConnL2AndDMAToMemCtrl := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL2AndDMAToMemCtrl")
	ConnL2AndDMAToMemCtrl.PlugIn(L2Cache.GetPortByName("Bottom"))
	ConnL2AndDMAToMemCtrl.PlugIn(DMAEngine.ToMem)
	ConnL2AndDMAToMemCtrl.PlugIn(MemCtrl.GetPortByName("Top"))

	log.Printf("Creating ConnTLBToL2TLB...")
	ConnTLBToL2TLB := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnVTLBToL2TLB")
	ConnTLBToL2TLB.PlugIn(VTLB.GetPortByName("Bottom"))
	ConnTLBToL2TLB.PlugIn(STLB.GetPortByName("Bottom"))
	ConnTLBToL2TLB.PlugIn(ITLB.GetPortByName("Bottom"))
	ConnTLBToL2TLB.PlugIn(L2TLB.GetPortByName("Top"))

	log.Printf("Creating ConnL2TLBToIoMMU...")
	ConnL2TLBToIoMMU := directconnection.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("ConnL2TLBToIoMMU")
	ConnL2TLBToIoMMU.PlugIn(L2TLB.GetPortByName("Bottom"))
	ConnL2TLBToIoMMU.PlugIn(IoMMU.GetPortByName("Top"))

	log.Printf("DMA engine connected to memory controller via ConnL2AndDMAToMemCtrl")

	log.Printf("Tracing...")
	// Tracing
	traceFile, err := os.Create("trace.log")
	if err != nil {
		panic("Error: Failed to create trace file")
	}
	logger := log.New(traceFile, "", 0)
	tracer := trace.NewTracer(logger, engine)
	tracing.CollectTrace(MemCtrl, tracer)

	log.Printf("Creating ReLU...")
	// Benchmark - reduced size for faster testing
	benchmark := relu.MakeBuilder().
		WithSimulation(s).
		WithLength(4). // 4 elements = minimal test case
		Build("ReLU")

	// Run simulation
	start := time.Now()
	log.Printf("Starting simulation with 4 elements...")

	log.Printf("Calling Driver.Run()...")
	Driver.Run()
	log.Printf("Driver.Run() completed")

	log.Printf("Calling benchmark.Run()...")
	benchmark.Run()
	log.Printf("benchmark.Run() completed")

	log.Printf("Calling Driver.Terminate()...")
	Driver.Terminate()
	log.Printf("Driver.Terminate() completed")

	log.Printf("Calling s.Terminate()...")
	s.Terminate()
	log.Printf("s.Terminate() completed")

	duration := time.Since(start)
	log.Printf("Simulation completed in %v", duration)
}
