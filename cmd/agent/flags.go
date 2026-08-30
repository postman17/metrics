package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/postman17/metrics/internal/config"
)

type Config struct {
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	RunAddr        string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	CryptoKey      string `env:"CRYPTO_KEY"`
}

type agentJSONConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
	Key            string `json:"key"`
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
		jsonConfig     string
	)

	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", 10, "reportInterval")
	flag.Int64Var(&pollInterval, "p", 2, "pollInterval")
	flag.StringVar(&key, "k", "", "key")
	flag.StringVar(&cryptoKey, "crypto-key", "", "crypto key path")
	flag.StringVar(&jsonConfig, "c", "", "json config file")
	flag.StringVar(&jsonConfig, "config", "", "json config file")
	flag.Parse()

	explicitFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		explicitFlags[f.Name] = true
	})

	configPath := jsonConfig
	if p := os.Getenv("CONFIG"); p != "" {
		configPath = p
	}

	var jsonCfg agentJSONConfig
	if configPath != "" {
		if err := config.LoadJSON(configPath, &jsonCfg); err != nil {
			slog.Error("Error loading JSON config:", "error", err)
		}
	}

	if jsonCfg.Address != "" && !explicitFlags["a"] {
		runAddr = jsonCfg.Address
	}
	if jsonCfg.ReportInterval != "" && !explicitFlags["r"] {
		if secs, err := config.ParseDurationSeconds(jsonCfg.ReportInterval); err == nil {
			reportInterval = secs
		} else {
			slog.Error("Error parsing report_interval from JSON config:", "error", err)
		}
	}
	if jsonCfg.PollInterval != "" && !explicitFlags["p"] {
		if secs, err := config.ParseDurationSeconds(jsonCfg.PollInterval); err == nil {
			pollInterval = secs
		} else {
			slog.Error("Error parsing poll_interval from JSON config:", "error", err)
		}
	}
	if jsonCfg.CryptoKey != "" && !explicitFlags["crypto-key"] {
		cryptoKey = jsonCfg.CryptoKey
	}
	if jsonCfg.Key != "" && !explicitFlags["k"] {
		key = jsonCfg.Key
	}

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
