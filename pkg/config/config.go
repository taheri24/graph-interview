package config

import (
	"fmt"
	"os"
	"strconv"
)

type DatabaseConfig struct {
	DSN      string // Database connection string (overrides individual fields)
	Type     string // Database type: "postgres" or "sqlite"
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   struct {
		Port string
	}
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			DSN:      getEnv("DB_URL", ""),
			Type:     getEnv("DATABASE_TYPE", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "taskdb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Server: struct {
			Port string
		}{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}
}

func (d *DatabaseConfig) String() string {
	if d.DSN != "" {
		return d.DSN
	}

	switch d.Type {
	case "postgres":
		if d.Password == "" {
			return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
				d.Host,
				d.Port,
				d.User,
				d.DBName,
				d.SSLMode,
			)
		}
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			d.Host,
			d.Port,
			d.User,
			d.Password,
			d.DBName,
			d.SSLMode,
		)
	case "sqlite":
		if d.DBName == ":memory" {
			return d.DBName
		}
		// Default to SQLite
		return fmt.Sprintf("%s.db", d.DBName)
	default:
		panic(fmt.Sprintf("invalid database type: %s", d.Type))
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
