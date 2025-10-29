**JSON representation for single_ping example**

```json
{
  "simulation": {
    "engine": {
      "package": "github.com/sarchlab/akita/v4/simulation"
    },
    "components": [
      {
        "builder_package": ["github.com/sarchlab/yuzawa_example/ping/pinger"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Sender",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "port": "PingPort"
      },
      {
        "builder_package": ["github.com/sarchlab/yuzawa_example/ping/pinger"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Receiver",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "port": "PingPort"
      }
    ],

    "connections": [
      {
        "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Conn",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "plugs": [
          { "component": "Sender",   "port": "PingPort", "bandwidth": 1 },
          { "component": "Receiver", "port": "PingPort", "bandwidth": 1 }
        ]
      }
    ]
  },

  "benchmark": {
    "builder_package": ["github.com/sarchlab/yuzawa_example/ping/benchmarks/single_ping"],
    "params": [
      { "name": "Sender",   "value": ["Sender"] },
      { "name": "Receiver", "value": "Receiver" }
    ]
  },

  "trace": { "enabled": false }
}

```

**JSON representation for multi_ping example**

```json
{
  "simulation": {
    "engine": {
      "package": "github.com/sarchlab/akita/v4/simulation"
    },
    "components": [
      {
        "builder_package": ["github.com/sarchlab/yuzawa_example/ping/pinger"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Sender1",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "port": "PingPort"
      },
      {
        "builder_package": ["github.com/sarchlab/yuzawa_example/ping/pinger"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Sender2",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "port": "PingPort"
      },
      {
        "builder_package": ["github.com/sarchlab/yuzawa_example/ping/pinger"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Receiver",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "port": "PingPort"
      }
    ],

    "connections": [
      {
        "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Conn",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "plugs": [
          { "component": "Sender1", "port": "PingPort", "bandwidth": 1 },
          { "component": "Sender2", "port": "PingPort", "bandwidth": 1 },
          { "component": "Receiver", "port": "PingPort", "bandwidth": 1 }
        ]
      }
    ]
  },

  "benchmark": {
    "builder_package": ["github.com/sarchlab/yuzawa_example/ping/benchmarks/multi_ping"],
    "params": [
      { "name": "Senders",  "value": ["Sender1", "Sender2"] },
      { "name": "Receiver", "value": "Receiver" },
      { "name": "NumPings", "value": 5 }
    ]
  },

  "trace": { "enabled": false }
}
```

**JSON representation for ideal_mem_control example**

```json
{
  "simulation": {
    "engine": {
      "package": "github.com/sarchlab/akita/v4/simulation"
    },
    "components": [
      {
        "builder_package": ["github.com/sarchlab/yuzawa_example/ping/memaccessagent"],
        "package": [
          "github.com/sarchlab/akita/v4/sim",
          "github.com/sarchlab/akita/v4/mem/mem"
        ],
        "name": "MemAgent",
        "params": [
          { "name": "Freq",        "value": 1, "unit": "GHz" },
          { "name": "MaxAddress",  "value": 1, "unit": "GB"  },
          { "name": "WriteLeft",   "value": 100000          },
          { "name": "ReadLeft",    "value": 100000          },
          { "name": "LowModule",   "value": "IdealMemoryController", "port": "Top" }
        ],
        "port": "Mem"
      },
      {
        "builder_package": ["github.com/sarchlab/akita/v4/mem/idealmemcontroller"],
        "package": ["github.com/sarchlab/akita/v4/mem/mem"],
        "name": "IdealMemoryController",
        "params": [
          { "name": "NewStorage", "value": 4, "unit": "GB" },
          { "name": "Latency",    "value": 100            }
        ],
        "port": "Top"
      }
    ],

    "connections": [
      {
        "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Conn",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "plugs": [
          { "component": "MemAgent",             "port": "Mem", "bandwidth": 1 },
          { "component": "IdealMemoryController","port": "Top", "bandwidth": 1 }
        ]
      }
    ]
  },

  "benchmark": {
    "builder_package": ["github.com/sarchlab/yuzawa_example/ping/benchmarks/ideal_mem_controller"],
    "package": ["github.com/sarchlab/akita/v4/mem/mem"],
    "params": [
      { "name": "NumAccess",  "value": 100000 },
      { "name": "MaxAddress", "value": 1, "unit": "GB" }
    ]
  },

  "trace": {
    "enabled":   true,
    "component": "IdealMemoryController",
    "file":      "trace.log",
    "package": [
      "log",
      "os",
      "github.com/sarchlab/akita/v4/mem/trace",
      "github.com/sarchlab/akita/v4/tracing"
    ]
  },

  "seed": 0

```

**JSON representation for core_cache_memory example**

```json
{
  "simulation": {
    "engine": {
      "package": "github.com/sarchlab/akita/v4/simulation"
    },
    "components": [
      {
        "builder_package": ["github.com/sarchlab/yuzawa_example/ping/memaccessagent"],
        "package": [
          "github.com/sarchlab/akita/v4/mem/mem",
          "github.com/sarchlab/akita/v4/sim"
        ],
        "name": "MemAgent",
        "params": [
          { "name": "Freq",        "value": 1, "unit": "GHz" },
          { "name": "MaxAddress",  "value": 1, "unit": "GB"  },
          { "name": "WriteLeft",   "value": 100000          },
          { "name": "ReadLeft",    "value": 100000          },
          { "name": "LowModule",   "value": "L1Cache", "port": "Top" }
        ],
        "port": "Mem"
      },

      {
        "builder_package": ["github.com/sarchlab/akita/v4/mem/cache/writethrough"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "L1Cache",
        "params": [
          { "name": "Freq",              "value": 1, "unit": "GHz" },
          { "name": "WayAssociativity",  "value": 2                 },
          { "name": "AddressMapperType", "value": "single"          },
          { "name": "RemotePorts",       "value": "L2Cache", "port": "Top" }
        ],
        "port": "Top"
      },

      {
        "builder_package": ["github.com/sarchlab/akita/v4/mem/cache/writeback"],
        "package": ["github.com/sarchlab/akita/v4/mem/mem"],
        "name": "L2Cache",
        "params": [
          { "name": "Freq",              "value": 1, "unit": "GHz" },
          { "name": "WayAssociativity",  "value": 4                 },
          { "name": "NumReqPerCycle",    "value": 2                 },
          { "name": "AddressMapperType", "value": "single"          },
          { "name": "RemotePorts",       "value": "MemCtrl", "port": "Top" }
        ],
        "port": "Top"
      },

      {
        "builder_package": ["github.com/sarchlab/akita/v4/mem/idealmemcontroller"],
        "package": ["github.com/sarchlab/akita/v4/mem/mem"],
        "name": "MemCtrl",
        "params": [
          { "name": "NewStorage", "value": 4, "unit": "GB" },
          { "name": "Latency",    "value": 100            }
        ],
        "port": "Top"
      }
    ],

    "connections": [
      {
        "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Conn1",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "plugs": [
          { "component": "MemAgent", "port": "Mem",  "bandwidth": 1 },
          { "component": "L1Cache",  "port": "Top",  "bandwidth": 1 }
        ]
      },
      {
        "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Conn2",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "plugs": [
          { "component": "L1Cache", "port": "Bottom", "bandwidth": 1 },
          { "component": "L2Cache", "port": "Top",    "bandwidth": 1 }
        ]
      },
      {
        "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
        "package": ["github.com/sarchlab/akita/v4/sim"],
        "name": "Conn3",
        "params": [
          { "name": "Freq", "value": 1, "unit": "GHz" }
        ],
        "plugs": [
          { "component": "L2Cache", "port": "Bottom", "bandwidth": 1 },
          { "component": "MemCtrl", "port": "Top",    "bandwidth": 1 }
        ]
      }
    ]
  },

  "benchmark": {
    "builder_package": ["github.com/sarchlab/yuzawa_example/ping/benchmarks/ideal_mem_controller"],
    "package": ["github.com/sarchlab/akita/v4/mem/mem"],
    "params": [
      { "name": "NumAccess",  "value": 100000 },
      { "name": "MaxAddress", "value": 1, "unit": "GB" }
    ]
  },

  "trace": {
    "enabled":   true,
    "component": "MemCtrl",
    "file":      "trace.log",
    "package": [
      "log",
      "os",
      "github.com/sarchlab/akita/v4/mem/trace",
      "github.com/sarchlab/akita/v4/tracing"
    ]
  },

  "seed": 0
}

```

**JSON representation for multi_stage_mem example**
```json
{
    "simulation": {
      "engine": {
        "package": "github.com/sarchlab/akita/v4/simulation"
      },
      "components": [
        {
          "builder_package": ["github.com/sarchlab/yuzawa_example/ping/memaccessagent"],
          "package": [
            "github.com/sarchlab/akita/v4/mem/mem",
            "github.com/sarchlab/akita/v4/sim"
          ],
          "name": "MemAgent",
          "params": [
            { "name": "Freq",        "value": 1, "unit": "GHz" },
            { "name": "MaxAddress",  "value": 1, "unit": "GB" },
            { "name": "WriteLeft",   "value": 100000 },
            { "name": "ReadLeft",    "value": 100000 },
            { "name": "LowModule",   "value": "ROB", "port": "Top" }
          ],
          "port": "Mem"
        },
  
        {
          "builder_package": ["github.com/sarchlab/yuzawa_example/ping/rob"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "ROB",
          "params": [
            { "name": "Freq",            "value": 1, "unit": "GHz" },
            { "name": "NumReqPerCycle",  "value": 4 },
            { "name": "BufferSize",      "value": 128 },
            { "name": "BottomUnit",      "value": "AT", "port": "Top" }
          ],
          "port": "Top"
        },
  
        {
          "builder_package": ["github.com/sarchlab/akita/v4/mem/vm/addresstranslator"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "AT",
          "params": [
            { "name": "Freq",                "value": 1, "unit": "GHz" },
            { "name": "Log2PageSize",        "value": 12 },
            { "name": "TranslationProvider", "value": "TLB",   "port": "Top" },
            { "name": "RemotePorts",         "value": "L1Cache","port": "Top" },
            { "name": "AddressMapperType",   "value": "single" }
          ],
          "port": "Top"
        },
  
        {
          "builder_package": ["github.com/sarchlab/akita/v4/mem/vm/tlb"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "TLB",
          "params": [
            { "name": "Freq",              "value": 1, "unit": "GHz" },
            { "name": "NumWays",           "value": 8 },
            { "name": "NumSets",           "value": 8 },
            { "name": "PageSize",          "value": 4096 },
            { "name": "NumReqPerCycle",    "value": 2 },
            { "name": "AddressMapperType", "value": "single" },
            { "name": "RemotePorts",       "value": "L2TLB", "port": "Top" }
          ],
          "port": "Top"
        },
  
        {
          "builder_package": ["github.com/sarchlab/akita/v4/mem/vm/tlb"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "L2TLB",
          "params": [
            { "name": "Freq",              "value": 1, "unit": "GHz" },
            { "name": "NumWays",           "value": 64 },
            { "name": "NumSets",           "value": 64 },
            { "name": "PageSize",          "value": 4096 },
            { "name": "NumReqPerCycle",    "value": 4 },
            { "name": "AddressMapperType", "value": "single" },
            { "name": "RemotePorts",       "value": "IoMMU", "port": "Top" }
          ],
          "port": "Top"
        },
  
        {
          "builder_package": ["github.com/sarchlab/yuzawa_example/ping/mmu"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "IoMMU",
          "params": [
            { "name": "Freq",               "value": 1, "unit": "GHz" },
            { "name": "Log2PageSize",       "value": 12 },
            { "name": "MaxNumReqInFlight",  "value": 16 },
            { "name": "PageWalkingLatency", "value": 10 }
          ],
          "port": "Top"
        },
  
        {
          "builder_package": ["github.com/sarchlab/akita/v4/mem/cache/writethrough"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "L1Cache",
          "params": [
            { "name": "Freq",              "value": 1, "unit": "GHz" },
            { "name": "WayAssociativity",  "value": 2 },
            { "name": "AddressMapperType", "value": "single" },
            { "name": "RemotePorts",       "value": "L2Cache", "port": "Top" }
          ],
          "port": "Top"
        },
  
        {
          "builder_package": ["github.com/sarchlab/akita/v4/mem/cache/writeback"],
          "package": ["github.com/sarchlab/akita/v4/mem/mem"],
          "name": "L2Cache",
          "params": [
            { "name": "Freq",              "value": 1, "unit": "GHz" },
            { "name": "WayAssociativity",  "value": 4 },
            { "name": "NumReqPerCycle",    "value": 2 },
            { "name": "AddressMapperType", "value": "single" },
            { "name": "RemotePorts",       "value": "MemCtrl", "port": "Top" }
          ],
          "port": "Top"
        },
  
        {
          "builder_package": ["github.com/sarchlab/akita/v4/mem/idealmemcontroller"],
          "package": ["github.com/sarchlab/akita/v4/mem/mem"],
          "name": "MemCtrl",
          "params": [
            { "name": "NewStorage", "value": 4, "unit": "GB" },
            { "name": "Latency",    "value": 100 }
          ],
          "port": "Top"
        }
      ],
  
      "connections": [
        {
          "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "Conn1",
          "params": [
            { "name": "Freq", "value": 1, "unit": "GHz" }
          ],
          "plugs": [
            { "component": "MemAgent", "port": "Mem", "bandwidth": 1 },
            { "component": "ROB",      "port": "Top", "bandwidth": 1 }
          ]
        },
        {
          "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "Conn2",
          "params": [
            { "name": "Freq", "value": 1, "unit": "GHz" }
          ],
          "plugs": [
            { "component": "ROB", "port": "Bottom", "bandwidth": 1 },
            { "component": "AT",  "port": "Top",    "bandwidth": 1 }
          ]
        },
        {
          "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "Conn3",
          "params": [
            { "name": "Freq", "value": 1, "unit": "GHz" }
          ],
          "plugs": [
            { "component": "AT",  "port": "Translation", "bandwidth": 1 },
            { "component": "TLB", "port": "Top",         "bandwidth": 1 }
          ]
        },
        {
          "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "Conn4",
          "params": [
            { "name": "Freq", "value": 1, "unit": "GHz" }
          ],
          "plugs": [
            { "component": "TLB",   "port": "Bottom", "bandwidth": 1 },
            { "component": "L2TLB", "port": "Top",    "bandwidth": 1 }
          ]
        },
        {
          "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "Conn5",
          "params": [
            { "name": "Freq", "value": 1, "unit": "GHz" }
          ],
          "plugs": [
            { "component": "L2TLB", "port": "Bottom", "bandwidth": 1 },
            { "component": "IoMMU", "port": "Top",    "bandwidth": 1 }
          ]
        },
        {
          "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "Conn6",
          "params": [
            { "name": "Freq", "value": 1, "unit": "GHz" }
          ],
          "plugs": [
            { "component": "AT",      "port": "Bottom", "bandwidth": 1 },
            { "component": "L1Cache", "port": "Top",    "bandwidth": 1 }
          ]
        },
        {
          "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "Conn7",
          "params": [
            { "name": "Freq", "value": 1, "unit": "GHz" }
          ],
          "plugs": [
            { "component": "L1Cache", "port": "Bottom", "bandwidth": 1 },
            { "component": "L2Cache", "port": "Top",    "bandwidth": 1 }
          ]
        },
        {
          "builder_package": ["github.com/sarchlab/akita/v4/sim/directconnection"],
          "package": ["github.com/sarchlab/akita/v4/sim"],
          "name": "Conn8",
          "params": [
            { "name": "Freq", "value": 1, "unit": "GHz" }
          ],
          "plugs": [
            { "component": "L2Cache", "port": "Bottom", "bandwidth": 1 },
            { "component": "MemCtrl", "port": "Top",    "bandwidth": 1 }
          ]
        }
      ]
    },
  
    "benchmark": {
      "builder_package": ["github.com/sarchlab/yuzawa_example/ping/benchmarks/multi_stage_memory"],
      "package": ["github.com/sarchlab/akita/v4/mem/mem"],
      "params": [
        { "name": "NumAccess",  "value": 100000 },
        { "name": "MaxAddress", "value": 1, "unit": "GB" }
      ]
    },
  
    "trace": {
      "enabled":   true,
      "component": "MemCtrl",
      "file":      "trace.log",
      "package": [
        "log",
        "os",
        "github.com/sarchlab/akita/v4/mem/trace",
        "github.com/sarchlab/akita/v4/tracing"
      ]
    },
  
    "seed": 0
  }
```