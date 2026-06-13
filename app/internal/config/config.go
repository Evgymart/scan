package config

import (
	"errors"
	"os"
)

type Config struct {
	ServerPort string
}

func Load() (*Config, error) {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return nil, errors.New("SERVER_PORT is not specified")
	}

	return &Config{
		ServerPort: serverPort,
	}, nil
}
