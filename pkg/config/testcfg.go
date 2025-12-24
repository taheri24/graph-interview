package config

func NewTestConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Type:   "sqlite",
			DBName: ":memory:",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   0,
		},
		CacheEnabled: true,
		Server: struct {
			Port string
		}{
			Port: "8080",
		},
	}
}
