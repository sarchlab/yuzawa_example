module github.com/sarchlab/yuzawa_example

go 1.23

toolchain go1.24.0

require github.com/sarchlab/akita/v4 v4.0.0

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/tebeka/atexit v0.3.0 // indirect
)

// replace github.com/sarchlab/akita/v4 => ../akita

// replace github.com/sarchlab/yuzawa_example/ping/benchmarks/multi_ping => /Users/sabilaaljannat/yuzawa_example/ping/benchmarks/multi_ping

replace github.com/sarchlab/yuzawa_example/ping/benchmarks/ideal_mem_controller => /Users/sabilaaljannat/yuzawa_example/ping/benchmarks/imc
