package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppPort         string
	RequestTimeout  time.Duration
	DatabaseURL     string
	DBMaxOpenConns  int32
	DBMaxIdleConns  int32
	BesuRPCURL      string
	ChainID         int64
	PrivateKey      string
	FromAddress     string
	ContractAddress string

	ConfirmationsRequired int64
	SyncInterval          time.Duration
	SyncBatchSize         int
}

func Load() (Config, error) {
	var cfg Config
	var errs []string

	cfg.AppPort = strings.TrimSpace(os.Getenv("APP_PORT"))
	if cfg.AppPort == "" {
		errs = append(errs, "APP_PORT is required")
	}

	timeoutMS, err := requiredInt("REQUEST_TIMEOUT_MS")
	if err != nil {
		errs = append(errs, err.Error())
	} else if timeoutMS <= 0 {
		errs = append(errs, "REQUEST_TIMEOUT_MS must be > 0")
	} else {
		cfg.RequestTimeout = time.Duration(timeoutMS) * time.Millisecond
	}

	cfg.DatabaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if cfg.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL is required")
	}

	dbOpen, err := requiredInt("DB_MAX_OPEN_CONNS")
	if err != nil {
		errs = append(errs, err.Error())
	} else {
		cfg.DBMaxOpenConns = int32(dbOpen)
	}

	dbIdle, err := requiredInt("DB_MAX_IDLE_CONNS")
	if err != nil {
		errs = append(errs, err.Error())
	} else {
		cfg.DBMaxIdleConns = int32(dbIdle)
	}

	cfg.BesuRPCURL = strings.TrimSpace(os.Getenv("BESU_RPC_URL"))
	if cfg.BesuRPCURL == "" {
		errs = append(errs, "BESU_RPC_URL is required")
	}

	chainID, err := requiredInt64("CHAIN_ID")
	if err != nil {
		errs = append(errs, err.Error())
	} else {
		cfg.ChainID = chainID
	}

	cfg.PrivateKey = strings.TrimPrefix(strings.TrimSpace(os.Getenv("PRIVATE_KEY")), "0x")
	if cfg.PrivateKey == "" {
		errs = append(errs, "PRIVATE_KEY is required")
	}

	cfg.FromAddress = strings.TrimSpace(os.Getenv("FROM_ADDRESS"))
	if cfg.FromAddress == "" {
		errs = append(errs, "FROM_ADDRESS is required")
	}

	cfg.ContractAddress = strings.TrimSpace(os.Getenv("CONTRACT_ADDRESS"))
	if cfg.ContractAddress == "" {
		errs = append(errs, "CONTRACT_ADDRESS is required")
	}

	confirms, err := requiredInt64("CONFIRMATIONS_REQUIRED")
	if err != nil {
		errs = append(errs, err.Error())
	} else if confirms < 0 {
		errs = append(errs, "CONFIRMATIONS_REQUIRED must be >= 0")
	} else {
		cfg.ConfirmationsRequired = confirms
	}

	syncSec, err := requiredInt("SYNC_INTERVAL_SECONDS")
	if err != nil {
		errs = append(errs, err.Error())
	} else if syncSec <= 0 {
		errs = append(errs, "SYNC_INTERVAL_SECONDS must be > 0")
	} else {
		cfg.SyncInterval = time.Duration(syncSec) * time.Second
	}

	batch, err := requiredInt("SYNC_BATCH_SIZE")
	if err != nil {
		errs = append(errs, err.Error())
	} else if batch <= 0 {
		errs = append(errs, "SYNC_BATCH_SIZE must be > 0")
	} else {
		cfg.SyncBatchSize = batch
	}

	if len(errs) > 0 {
		return Config{}, errors.New(strings.Join(errs, "; "))
	}

	return cfg, nil
}

func requiredInt(name string) (int, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return v, nil
}

func requiredInt64(name string) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return v, nil
}
