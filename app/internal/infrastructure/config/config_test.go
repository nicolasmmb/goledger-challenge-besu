package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	t.Setenv("APP_PORT", "8080")
	t.Setenv("REQUEST_TIMEOUT_MS", "15000")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/goledger?sslmode=disable")
	t.Setenv("DB_MAX_OPEN_CONNS", "10")
	t.Setenv("DB_MAX_IDLE_CONNS", "5")
	t.Setenv("BESU_RPC_URL", "http://localhost:8545")
	t.Setenv("CHAIN_ID", "1337")
	t.Setenv("PRIVATE_KEY", "0xabc")
	t.Setenv("FROM_ADDRESS", "0xfrom")
	t.Setenv("CONTRACT_ADDRESS", "0xcontract")
	t.Setenv("CONFIRMATIONS_REQUIRED", "2")
	t.Setenv("SYNC_INTERVAL_SECONDS", "5")
	t.Setenv("SYNC_BATCH_SIZE", "50")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "8080", cfg.AppPort)
	require.EqualValues(t, 1337, cfg.ChainID)
	require.Equal(t, 50, cfg.SyncBatchSize)
}

func TestLoad_MissingRequired(t *testing.T) {
	clearRelevantEnv(t)

	_, err := Load()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "APP_PORT is required"))
}

func clearRelevantEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"APP_PORT", "REQUEST_TIMEOUT_MS", "DATABASE_URL", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"BESU_RPC_URL", "CHAIN_ID", "PRIVATE_KEY", "FROM_ADDRESS", "CONTRACT_ADDRESS",
		"CONFIRMATIONS_REQUIRED", "SYNC_INTERVAL_SECONDS", "SYNC_BATCH_SIZE",
	}
	for _, k := range keys {
		_ = os.Unsetenv(k)
	}
}
