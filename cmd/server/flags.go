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
	Database_DSN    string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

func parseFlags() Config {
	var (
		runAddr         string
		storeInterval   int64
		fileStoragePath string
		restore         bool
		dbDsn           string
		key             string
	)
	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&storeInterval, "i", 300, "store interval in seconds")
	// ./file.json
	flag.StringVar(&fileStoragePath, "f", "", "file storage path")
	flag.BoolVar(&restore, "r", false, "load previous values")
	// postgres://metrics_user:metrics_user_password@localhost:5432/metrics
	flag.StringVar(&dbDsn, "d", "", "dsn database address")
	flag.StringVar(&key, "k", "", "key")
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
	if cfg.Database_DSN != "" {
		dbDsn = cfg.Database_DSN
	}
	if cfg.Key != "" {
		key = cfg.Key
	}

	return Config{
		RunAddr:         runAddr,
		StoreInterval:   &storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         &restore,
		Database_DSN:    dbDsn,
		Key:             key,
	}
}
