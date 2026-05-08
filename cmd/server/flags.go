package main

import (
	"flag"
)

type Config struct {
	runAddr string
}

func parseFlags() Config {
	var (
		runAddr string
	)
	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	config := Config{runAddr: runAddr}
	return config
}
