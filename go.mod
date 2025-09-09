module github.com/sarchlab/yuzawa_example

go 1.25

require (
	github.com/sarchlab/akita/v4 v4.6.1
	github.com/sarchlab/mgpusim/v4 v4.1.3
)

require (
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/pprof v0.0.0-20250820193118-f64d9cf942d6 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.32 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/syifan/goseth v0.1.2 // indirect
	github.com/tebeka/atexit v0.3.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/sys v0.35.0 // indirect
)

// replace github.com/sarchlab/akita/v4 => ../akita

// replace (
// 	github.com/sarchlab/akita/v4 => /Users/sabilaaljannat/akita
// 	github.com/sarchlab/yuzawa_example => /Users/sabilaaljannat/yuzawa_example
// )

replace (
	github.com/sarchlab/akita/v4 => /Users/sabilaaljannat/akita
	github.com/sarchlab/mgpusim/v4 => /Users/sabilaaljannat/mgpusim
)
