package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ReportInterval int64  `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int64  `env:"POLL_INTERVAL" envDefault:"2"`
	RunAddr        string `env:"ADDRESS" envDefault:"http://localhost:8080"`
}

func addPrefix(s string) string {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "http://" + s
	}
	return s
}

func parseFlags() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Println("Error:", err)
	}
	var (
		reportInterval int64
		pollInterval   int64
		runAddr        string
	)
	flag.StringVar(&runAddr, "a", cfg.RunAddr, "address and port to run server")
	flag.Int64Var(&reportInterval, "r", cfg.ReportInterval, "reportInterval")
	flag.Int64Var(&pollInterval, "p", cfg.PollInterval, "pollInterval")
	flag.Parse()

	config := Config{
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		RunAddr:        addPrefix(runAddr),
	}

	return config
}
