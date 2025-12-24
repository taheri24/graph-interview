package config_test

import (
	"os"
	"testing"

	"taheri24.ir/graph1/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Store original env vars
	origDBHost := os.Getenv("DB_HOST")
	origDBPort := os.Getenv("DB_PORT")
	origDBUser := os.Getenv("DB_USER")
	origDBPassword := os.Getenv("DB_PASSWORD")
	origDBName := os.Getenv("DB_NAME")
	origDBSSLMode := os.Getenv("DB_SSLMODE")
	origServerPort := os.Getenv("SERVER_PORT")

	// Clean up after test
	defer func() {
		os.Setenv("DB_HOST", origDBHost)
		os.Setenv("DB_PORT", origDBPort)
		os.Setenv("DB_USER", origDBUser)
		os.Setenv("DB_PASSWORD", origDBPassword)
		os.Setenv("DB_NAME", origDBName)
		os.Setenv("DB_SSLMODE", origDBSSLMode)
		os.Setenv("SERVER_PORT", origServerPort)
	}()

	// Test with default values (no env vars set)
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")
	os.Unsetenv("SERVER_PORT")

	cfg := config.Load()

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "", cfg.Database.Password)
	assert.Equal(t, "taskdb", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "8080", cfg.Server.Port)

	// Test with custom env vars
	os.Setenv("DB_HOST", "customhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "customuser")
	os.Setenv("DB_PASSWORD", "custompass")
	os.Setenv("DB_NAME", "customdb")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("SERVER_PORT", "9000")

	cfg = config.Load()

	assert.Equal(t, "customhost", cfg.Database.Host)
	assert.Equal(t, "5433", cfg.Database.Port)
	assert.Equal(t, "customuser", cfg.Database.User)
	assert.Equal(t, "custompass", cfg.Database.Password)
	assert.Equal(t, "customdb", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "9000", cfg.Server.Port)
}

func TestDatabaseConfigString(t *testing.T) {
	// Test DSN without password
	dbCfg := config.DatabaseConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "",
		DBName:   "testdb",
		SSLMode:  "disable",
	}
	dsn := dbCfg.String()
	expected := "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)

	// Test DSN with password
	dbCfg.Password = "password"
	dsn = dbCfg.String()
	expected = "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)

	// Test DSN with custom values
	dbCfg.Host = "customhost"
	dbCfg.Port = "5433"
	dbCfg.User = "customuser"
	dbCfg.Password = "custompass"
	dbCfg.DBName = "customdb"
	dbCfg.SSLMode = "require"
	dsn, expected = dbCfg.String(), `host=customhost port=5433 user=customuser password=custompass dbname=customdb sslmode=require`
	assert.Equal(t, expected, dsn)

	// Test SQLite DSN
	dbCfg.Type = "sqlite"
	dbCfg.DBName = "testdb"
	dsn = dbCfg.String()
	assert.Equal(t, "testdb.db", dsn)

	// Test SQLite DSN with :memory
	dbCfg.DBName = ":memory:"
	dbCfg.Type = "sqlite"
	assert.Equal(t, ":memory:", dbCfg.String())
	// Test SQLite DSN with :memory
	dbCfg.DSN = "postgresql://user:password@localhost:5432/dbname?sslmode=disable"
	dbCfg.Type = "postgresql"
	assert.Equal(t, "postgresql://user:password@localhost:5432/dbname?sslmode=disable", dbCfg.String())

}

func TestGetEnv(t *testing.T) {
	// Test getting existing environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	// Since getEnv is not exported, we test it indirectly through Load
	origValue := os.Getenv("DB_HOST")
	defer os.Setenv("DB_HOST", origValue)

	os.Setenv("DB_HOST", "test_host")
	cfg := config.Load()
	assert.Equal(t, "test_host", cfg.Database.Host)

	// Test getting non-existing environment variable (should return default)
	os.Unsetenv("DB_HOST")
	cfg = config.Load()
	assert.Equal(t, "localhost", cfg.Database.Host) // default value
}

func TestGetEnvAsInt(t *testing.T) {
	// Test getting existing integer environment variable
	os.Setenv("TEST_INT_VAR", "42")
	defer os.Unsetenv("TEST_INT_VAR")

	// Since getEnvAsInt is not exported, we test it indirectly through Load
	origValue := os.Getenv("REDIS_DB")
	defer os.Setenv("REDIS_DB", origValue)

	os.Setenv("REDIS_DB", "5")
	cfg := config.Load()
	assert.Equal(t, 5, cfg.Redis.DB)

	// Test getting non-existing environment variable (should return default)
	os.Unsetenv("REDIS_DB")
	cfg = config.Load()
	assert.Equal(t, 0, cfg.Redis.DB) // default value

	// Test invalid integer (should return default)
	os.Setenv("REDIS_DB", "invalid")
	cfg = config.Load()
	assert.Equal(t, 0, cfg.Redis.DB) // default value
}
