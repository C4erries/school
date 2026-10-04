package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/infrastructure/api/config"
)

func TestConfig_LoadDefaults(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "development", cfg.App.Env)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.Equal(t, "debug", cfg.App.LogLevel)
	assert.Equal(t, 15*time.Second, cfg.App.Timeout)

	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "school", cfg.Postgres.DB)
	assert.Equal(t, "postgres://school:school_dev_password@localhost:5432/school?sslmode=disable", cfg.Postgres.DSN())

	assert.Equal(t, "localhost:6379", cfg.Valkey.Addr())
	assert.Equal(t, "localhost:9000", cfg.MinIO.Endpoint)
}

func TestConfig_LoadEnvOverride(t *testing.T) {
	_ = os.Setenv("APP_PORT", "9999")
	_ = os.Setenv("APP_ENV", "production")
	defer func() {
		_ = os.Unsetenv("APP_PORT")
		_ = os.Unsetenv("APP_ENV")
	}()

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 9999, cfg.App.Port)
	assert.Equal(t, "production", cfg.App.Env)
}
