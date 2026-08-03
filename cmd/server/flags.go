package main

import (
	"flag"
	"log/slog"

	"github.com/caarlos0/env/v11"
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
	)
	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&storeInterval, "i", 300, "store interval in seconds")
	// ./file.json
	flag.StringVar(&fileStoragePath, "f", "", "file storage path")
	flag.BoolVar(&restore, "r", false, "load previous values")
	// postgres://metrics_user:metrics_user_password@localhost:5432/metrics
	flag.StringVar(&dbDsn, "d", "", "dsn database address")
	flag.StringVar(&key, "k", "", "key")
	flag.StringVar(&auditFile, "h", "", "audit file")
	flag.StringVar(&auditURL, "u", "", "audit url")
	flag.Parse()

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

	return Config{
		RunAddr:         runAddr,
		StoreInterval:   &storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         &restore,
		DatabaseDSN:     dbDsn,
		Key:             key,
		AuditFile:       auditFile,
		AuditURL:        auditURL,
	}
}
