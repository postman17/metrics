package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type httpAddr string
type Config struct {
	ReportInterval int64    `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int64    `env:"POLL_INTERVAL" envDefault:"2"`
	RunAddr        httpAddr `env:"ADDRESS" envDefault:"http://localhost:8080"`
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
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Println("Error:", err)
	}
	var (
		reportInterval int64
		pollInterval   int64
		runAddr        httpAddr = cfg.RunAddr
	)
	flag.Var(&runAddr, "a", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", cfg.ReportInterval, "reportInterval")
	flag.Int64Var(&pollInterval, "p", cfg.PollInterval, "pollInterval")
	flag.Parse()

	config := Config{
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		RunAddr:        runAddr,
	}

	return config
}
