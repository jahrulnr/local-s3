package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port        int
	DataDir     string
	AccessKey   string
	SecretKey   string
	Region      string
	LogLevel    string
	BaseDomain  string
	DisableAuth bool
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnvAsInt("PORT", 3000),
		DataDir:     getEnv("DATA_DIR", "./data"),
		AccessKey:   getEnv("ACCESS_KEY", "test"),
		SecretKey:   getEnv("SECRET_KEY", "test123456789"),
		Region:      getEnv("REGION", "ap-southeast-3"),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),
		BaseDomain:  getEnv("BASE_DOMAIN", "localhost"),
		DisableAuth: getEnvAsBool("DISABLE_AUTH", false),
	}

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
