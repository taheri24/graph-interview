package config

import (
	"fmt"
	"os"
)

type DatabaseConfig struct {
	DATABASE_URL string

	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Config struct {
	Database DatabaseConfig
	Server   struct {
		Port string
	}
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			DATABASE_URL: getEnv("DB_URL", ""),
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", ""),
			DBName:       getEnv("DB_NAME", "taskdb"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
		},
		Server: struct {
			Port string
		}{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}
}

func (c *Config) GetDatabaseDSN() string {
	if c.Database.DATABASE_URL != "" {
		return c.Database.DATABASE_URL
	}
	if c.Database.Password == "" {
		return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
			c.Database.Host,
			c.Database.Port,
			c.Database.User,
			c.Database.DBName,
			c.Database.SSLMode,
		)
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
