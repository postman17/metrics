package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddr         string `env:"ADDRESS"`
	StoreInterval   *int64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
}

func parseFlags() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Println("Error:", err)
	}
	runAddr := cfg.RunAddr
	if runAddr == "" {
		flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
		flag.Parse()
	}

	storeInterval := int64(300)
	if cfg.StoreInterval != nil {
		storeInterval = *cfg.StoreInterval
	} else {
		flag.Int64Var(&storeInterval, "i", 300, "store interval in seconds")
		flag.Parse()
	}

	fileStoragePath := cfg.FileStoragePath
	if fileStoragePath == "" {
		flag.StringVar(&fileStoragePath, "f", "./file.json", "file storage path")
		flag.Parse()
	}

	restore := false
	if cfg.Restore != nil {
		restore = *cfg.Restore
	} else {
		flag.BoolVar(&restore, "r", false, "load previous values")
		flag.Parse()
	}

	return Config{
		RunAddr:         runAddr,
		StoreInterval:   &storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         &restore,
	}
}
