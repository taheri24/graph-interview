package main

import (
"os"
"testing"

"taheri24.ir/graph1/pkg/config"

"github.com/stretchr/testify/assert"
)

func TestConfigLoading(t *testing.T) {
t.Run("ConfigLoading", func(t *testing.T) {
originalPort := os.Getenv("SERVER_PORT")
originalDBHost := os.Getenv("DB_HOST")

os.Setenv("SERVER_PORT", "9090")
os.Setenv("DB_HOST", "testhost")

cfg := config.Load()
assert.Equal(t, "9090", cfg.Server.Port)
assert.Equal(t, "testhost", cfg.Database.Host)

if originalPort != "" {
os.Setenv("SERVER_PORT", originalPort)
} else {
os.Unsetenv("SERVER_PORT")
}
if originalDBHost != "" {
os.Setenv("DB_HOST", originalDBHost)
} else {
os.Unsetenv("DB_HOST")
}
})

t.Run("ConfigDefaults", func(t *testing.T) {
os.Unsetenv("SERVER_PORT")
os.Unsetenv("DB_HOST")

cfg := config.Load()
assert.Equal(t, "8080", cfg.Server.Port)
assert.Equal(t, "localhost", cfg.Database.Host)
})
}

func TestConfigurationValidation(t *testing.T) {
t.Run("DatabaseDSNGeneration", func(t *testing.T) {
cfg := &config.Config{}

cfg.Database.DATABASE_URL = ""
cfg.Database.Host = "localhost"
cfg.Database.Port = "5432"
cfg.Database.User = "user"
cfg.Database.Password = ""
cfg.Database.DBName = "testdb"
cfg.Database.SSLMode = "disable"

dsn := cfg.GetDatabaseDSN()
expected := "host=localhost port=5432 user=user dbname=testdb sslmode=disable"
assert.Equal(t, expected, dsn)

cfg.Database.Password = "password"
dsn = cfg.GetDatabaseDSN()
expected = "host=localhost port=5432 user=user password=password dbname=testdb sslmode=disable"
assert.Equal(t, expected, dsn)

cfg.Database.DATABASE_URL = "postgres://user:pass@localhost:5432/testdb"
dsn = cfg.GetDatabaseDSN()
assert.Equal(t, "postgres://user:pass@localhost:5432/testdb", dsn)
})
}

func TestEnvironmentVariableHandling(t *testing.T) {
t.Run("EnvVariableOverrides", func(t *testing.T) {
originalVars := map[string]string{
"SERVER_PORT":    os.Getenv("SERVER_PORT"),
"DB_HOST":        os.Getenv("DB_HOST"),
"DB_PORT":        os.Getenv("DB_PORT"),
"DB_USER":        os.Getenv("DB_USER"),
"DB_PASSWORD":    os.Getenv("DB_PASSWORD"),
"DB_NAME":        os.Getenv("DB_NAME"),
"DB_SSLMODE":     os.Getenv("DB_SSLMODE"),
"REDIS_HOST":     os.Getenv("REDIS_HOST"),
"REDIS_PORT":     os.Getenv("REDIS_PORT"),
"REDIS_PASSWORD": os.Getenv("REDIS_PASSWORD"),
"REDIS_DB":       os.Getenv("REDIS_DB"),
}

testEnvVars := map[string]string{
"SERVER_PORT":    "9999",
"DB_HOST":        "testhost",
"DB_PORT":        "5433",
"DB_USER":        "testuser",
"DB_PASSWORD":    "testpass",
"DB_NAME":        "testdb",
"DB_SSLMODE":     "require",
"REDIS_HOST":     "redishost",
"REDIS_PORT":     "6380",
"REDIS_PASSWORD": "redispass",
"REDIS_DB":       "1",
}

for key, value := range testEnvVars {
os.Setenv(key, value)
}

cfg := config.Load()

assert.Equal(t, "9999", cfg.Server.Port)
assert.Equal(t, "testhost", cfg.Database.Host)
assert.Equal(t, "5433", cfg.Database.Port)
assert.Equal(t, "testuser", cfg.Database.User)
assert.Equal(t, "testpass", cfg.Database.Password)
assert.Equal(t, "testdb", cfg.Database.DBName)
assert.Equal(t, "require", cfg.Database.SSLMode)
assert.Equal(t, "redishost", cfg.Redis.Host)
assert.Equal(t, "6380", cfg.Redis.Port)
assert.Equal(t, "redispass", cfg.Redis.Password)
assert.Equal(t, 1, cfg.Redis.DB)

for key, value := range originalVars {
if value != "" {
os.Setenv(key, value)
} else {
os.Unsetenv(key)
}
}
})
}

func TestApplicationFlow(t *testing.T) {
t.Run("MainComponentIntegration", func(t *testing.T) {
cfg := config.Load()
assert.NotNil(t, cfg)

dsn := cfg.GetDatabaseDSN()
assert.NotEmpty(t, dsn)
})
}
