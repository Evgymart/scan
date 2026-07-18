package config

import (
	"errors"
	"os"
	"strconv"
)

const defaultMaxUploadSize = 100 << 20 // 100MB

const defaultBadgerPath = "./data/badger"

type Config struct {
	ServerPort    string
	MaxUploadSize int64
	BadgerDBPath  string
}

func Load() (*Config, error) {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return nil, errors.New("SERVER_PORT is not specified")
	}

	maxUploadSize := int64(defaultMaxUploadSize)
	if val := os.Getenv("MAX_UPLOAD_SIZE"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			maxUploadSize = parsed
		}
	}

	badgerDBPath := defaultBadgerPath
	if val := os.Getenv("BADGER_DB_PATH"); val != "" {
		badgerDBPath = val
	}

	return &Config{
		ServerPort:    serverPort,
		MaxUploadSize: maxUploadSize,
		BadgerDBPath:  badgerDBPath,
	}, nil
}
