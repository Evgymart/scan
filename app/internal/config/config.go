package config

import (
	"errors"
	"os"
	"strconv"
)

const defaultMaxUploadSize = 100 << 20 // 100MB

const defaultBadgerPath = "./data/badger"

const defaultScanDir = "./data/scan"

const defaultClamAVHost = "clamav"

const defaultClamAVPort = 3310

type Config struct {
	ServerPort    string
	MaxUploadSize int64
	BadgerDBPath  string
	ScanDir       string
	ClamAVHost    string
	ClamAVPort    int
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

	scanDir := defaultScanDir
	if val := os.Getenv("SCAN_DIR"); val != "" {
		scanDir = val
	}

	clamAVHost := defaultClamAVHost
	if val := os.Getenv("CLAMAV_HOST"); val != "" {
		clamAVHost = val
	}

	clamAVPort := defaultClamAVPort
	if val := os.Getenv("CLAMAV_PORT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			clamAVPort = parsed
		}
	}

	return &Config{
		ServerPort:    serverPort,
		MaxUploadSize: maxUploadSize,
		BadgerDBPath:  badgerDBPath,
		ScanDir:       scanDir,
		ClamAVHost:    clamAVHost,
		ClamAVPort:    clamAVPort,
	}, nil
}
