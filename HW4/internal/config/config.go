package config

import "os"

type Config struct {
	HTTPAddr string
}

func Load() Config {
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	return Config{HTTPAddr: address}
}
