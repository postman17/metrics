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
}

func parseFlags() Config {
	var (
		runAddr         string
		storeInterval   int64
		fileStoragePath string
		restore         bool
	)
	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&storeInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&fileStoragePath, "f", "./file.json", "file storage path")
	flag.BoolVar(&restore, "r", false, "load previous values")
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

	return Config{
		RunAddr:         runAddr,
		StoreInterval:   &storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         &restore,
	}
}
