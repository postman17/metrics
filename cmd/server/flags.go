package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/postman17/metrics/internal/config"
)

type Config struct {
	RunAddr         string `env:"ADDRESS"`
	StoreInterval   *int64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	CryptoKey       string `env:"CRYPTO_KEY"`
}

type serverJSONConfig struct {
	Address       string `json:"address"`
	Restore       *bool  `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	Key           string `json:"key"`
	AuditFile     string `json:"audit_file"`
	AuditURL      string `json:"audit_url"`
}

func parseFlags() Config {
	var (
		runAddr         string
		storeInterval   int64
		fileStoragePath string
		restore         bool
		dbDsn           string
		key             string
		auditFile       string
		auditURL        string
		cryptoKey       string
		jsonConfig      string
	)
	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&storeInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&fileStoragePath, "f", "", "file storage path")
	flag.BoolVar(&restore, "r", false, "load previous values")
	flag.StringVar(&dbDsn, "d", "", "dsn database address")
	flag.StringVar(&key, "k", "", "key")
	flag.StringVar(&auditFile, "h", "", "audit file")
	flag.StringVar(&auditURL, "u", "", "audit url")
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

	var jsonCfg serverJSONConfig
	if configPath != "" {
		if err := config.LoadJSON(configPath, &jsonCfg); err != nil {
			slog.Error("Error loading JSON config:", "error", err)
		}
	}

	if jsonCfg.Address != "" && !explicitFlags["a"] {
		runAddr = jsonCfg.Address
	}
	if jsonCfg.StoreInterval != "" && !explicitFlags["i"] {
		if secs, err := config.ParseDurationSeconds(jsonCfg.StoreInterval); err == nil {
			storeInterval = secs
		} else {
			slog.Error("Error parsing store_interval from JSON config:", "error", err)
		}
	}
	if jsonCfg.StoreFile != "" && !explicitFlags["f"] {
		fileStoragePath = jsonCfg.StoreFile
	}
	if jsonCfg.Restore != nil && !explicitFlags["r"] {
		restore = *jsonCfg.Restore
	}
	if jsonCfg.DatabaseDSN != "" && !explicitFlags["d"] {
		dbDsn = jsonCfg.DatabaseDSN
	}
	if jsonCfg.Key != "" && !explicitFlags["k"] {
		key = jsonCfg.Key
	}
	if jsonCfg.AuditFile != "" && !explicitFlags["h"] {
		auditFile = jsonCfg.AuditFile
	}
	if jsonCfg.AuditURL != "" && !explicitFlags["u"] {
		auditURL = jsonCfg.AuditURL
	}
	if jsonCfg.CryptoKey != "" && !explicitFlags["crypto-key"] {
		cryptoKey = jsonCfg.CryptoKey
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Error parse envs:", "error", err)
	}
	if cfg.RunAddr != "" {
		runAddr = cfg.RunAddr
	}
	if cfg.StoreInterval != nil {
		storeInterval = *cfg.StoreInterval
	}
	if cfg.FileStoragePath != "" {
		fileStoragePath = cfg.FileStoragePath
	}
	if cfg.Restore != nil {
		restore = *cfg.Restore
	}
	if cfg.DatabaseDSN != "" {
		dbDsn = cfg.DatabaseDSN
	}
	if cfg.Key != "" {
		key = cfg.Key
	}
	if cfg.AuditFile != "" {
		auditFile = cfg.AuditFile
	}
	if cfg.AuditURL != "" {
		auditURL = cfg.AuditURL
	}
	if cfg.CryptoKey != "" {
		cryptoKey = cfg.CryptoKey
	}

	return Config{
		RunAddr:         runAddr,
		StoreInterval:   &storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         &restore,
		DatabaseDSN:     dbDsn,
		Key:             key,
		AuditFile:       auditFile,
		AuditURL:        auditURL,
		CryptoKey:       cryptoKey,
	}
}
