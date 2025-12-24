package config_test

import (
	"testing"

	"taheri24.ir/graph1/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestNewTestConfig(t *testing.T) {
	cfg := config.NewTestConfig()

	assert.NotNil(t, cfg)

	// Check Database config
	assert.Equal(t, "sqlite", cfg.Database.Type)
	assert.Equal(t, ":memory:", cfg.Database.DBName)

	// Check Redis config
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, 0, cfg.Redis.DB)

	// Check Server config
	assert.Equal(t, "8080", cfg.Server.Port)
}
