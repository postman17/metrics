package main

import (
	"flag"
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	RunAddr        string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	RateLimit      int64  `env:"RATE_LIMIT"`
}

func addPrefix(s string) string {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "http://" + s
	}
	return s
}

func parseFlags() Config {
	var (
		reportInterval int64
		pollInterval   int64
		runAddr        string
		key            string
		rateLimit      int64
	)

	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", 10, "reportInterval")
	flag.Int64Var(&pollInterval, "p", 2, "pollInterval")
	flag.StringVar(&key, "k", "", "key")
	flag.Int64Var(&rateLimit, "l", 5, "pollInterval")
	flag.Parse()

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Error parse envs:", "error", err)
	}
	if cfg.RunAddr != "" {
		runAddr = cfg.RunAddr
	}
	if cfg.ReportInterval != 0 {
		reportInterval = cfg.ReportInterval
	}
	if cfg.PollInterval != 0 {
		pollInterval = cfg.PollInterval
	}
	if cfg.Key != "" {
		key = cfg.Key
	}
	if cfg.RateLimit != 0 {
		rateLimit = cfg.RateLimit
	}

	return Config{
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		RunAddr:        addPrefix(runAddr),
		Key:            key,
		RateLimit:      rateLimit,
	}
}
