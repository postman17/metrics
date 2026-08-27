package main

import (
	"flag"
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config содержит параметры конфигурации агента: интервалы опроса/отправки,
// адрес сервера и ключ подписи.
type Config struct {
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	RunAddr        string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	CryptoKey      string `env:"CRYPTO_KEY"`
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
		cryptoKey      string
	)

	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", 10, "reportInterval")
	flag.Int64Var(&pollInterval, "p", 2, "pollInterval")
	flag.StringVar(&key, "k", "", "key")
	flag.StringVar(&cryptoKey, "crypto-key", "", "crypto key path")
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
	if cfg.CryptoKey != "" {
		cryptoKey = cfg.CryptoKey
	}

	return Config{
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		RunAddr:        addPrefix(runAddr),
		Key:            key,
		CryptoKey:      cryptoKey,
	}
}
