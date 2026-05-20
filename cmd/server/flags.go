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
	var (
		runAddr         string = cfg.RunAddr
		storeInterval   int64
		fileStoragePath string = cfg.FileStoragePath
		restore         bool
	)
	if cfg.RunAddr == "" {
		flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
		flag.Parse()
	}
	if cfg.StoreInterval == nil {
		flag.Int64Var(&storeInterval, "i", 300, "store interval in seconds")
		flag.Parse()
	}
	if cfg.FileStoragePath == "" {
		flag.StringVar(&fileStoragePath, "f", "./file.json", "file storage path")
		flag.Parse()
	}
	if cfg.Restore == nil {
		flag.BoolVar(&restore, "r", false, "load previous values")
		flag.Parse()
	}
	config := Config{
		RunAddr:         runAddr,
		StoreInterval:   &storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         &restore,
	}
	return config
}
