package main

import (
	"flag"
	"strings"
)

type httpAddr string
type Config struct {
	reportInterval int64
	pollInterval   int64
	runAddr        httpAddr
}

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

func parseFlags() Config {
	var (
		reportInterval int64
		pollInterval   int64
		runAddr        httpAddr = "http://localhost:8080"
	)

	flag.Var(&runAddr, "a", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", 10, "reportInterval")
	flag.Int64Var(&pollInterval, "p", 2, "pollInterval")
	flag.Parse()

	config := Config{
		reportInterval: reportInterval,
		pollInterval:   pollInterval,
		runAddr:        runAddr,
	}

	return config
}
