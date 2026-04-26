package main

import (
	"flag"
)

var (
	flagRunAddr        string
	flagReportInterval int64
	flagPollInterval   int64
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "http://localhost:8080", "address and port to run server")
	flag.Int64Var(&flagReportInterval, "r", 10, "reportInterval")
	flag.Int64Var(&flagPollInterval, "p", 2, "pollInterval")
	flag.Parse()
}
