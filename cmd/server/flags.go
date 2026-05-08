package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddr string `env:"ADDRESS"`
}

func parseFlags() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Println("Error:", err)
	}
	var (
		runAddr string = cfg.RunAddr
	)
	if cfg.RunAddr == "" {
		flag.StringVar(&runAddr, "a", "", "address and port to run server")
		flag.Parse()
	}
	if runAddr == "" {
		runAddr = "localhost:8080"
	}
	config := Config{RunAddr: runAddr}
	return config
}
