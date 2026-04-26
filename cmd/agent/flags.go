package main

import (
	"flag"
	"strings"
)

var (
	flagReportInterval int64
	flagPollInterval   int64
	flagRunAddr        httpAddr = "http://localhost:8080"
)

type httpAddr string

func (h *httpAddr) Set(s string) error {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "http://" + s
	}
	*h = httpAddr(s)
	return nil
}

func (h *httpAddr) String() string {
	return string(*h)
}

func parseFlags() {
	flag.Var(&flagRunAddr, "a", "address and port to run server")
	flag.Int64Var(&flagReportInterval, "r", 10, "reportInterval")
	flag.Int64Var(&flagPollInterval, "p", 2, "pollInterval")
	flag.Parse()
}
