package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original environment variables
	origPort := os.Getenv("PORT")
	origDataDir := os.Getenv("DATA_DIR")
	origAccessKey := os.Getenv("ACCESS_KEY")
	origSecretKey := os.Getenv("SECRET_KEY")
	origRegion := os.Getenv("REGION")
	origLogLevel := os.Getenv("LOG_LEVEL")
	origBaseDomain := os.Getenv("BASE_DOMAIN")
	origDisableAuth := os.Getenv("DISABLE_AUTH")

	// Clean up environment after test
	defer func() {
		os.Setenv("PORT", origPort)
		os.Setenv("DATA_DIR", origDataDir)
		os.Setenv("ACCESS_KEY", origAccessKey)
		os.Setenv("SECRET_KEY", origSecretKey)
		os.Setenv("REGION", origRegion)
		os.Setenv("LOG_LEVEL", origLogLevel)
		os.Setenv("BASE_DOMAIN", origBaseDomain)
		os.Setenv("DISABLE_AUTH", origDisableAuth)
	}()

	// Test default values
	os.Unsetenv("PORT")
	os.Unsetenv("DATA_DIR")
	os.Unsetenv("ACCESS_KEY")
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("REGION")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("BASE_DOMAIN")
	os.Unsetenv("DISABLE_AUTH")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("Expected default port 3000, got %d", cfg.Port)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("Expected default data dir './data', got %s", cfg.DataDir)
	}
	if cfg.AccessKey != "test" {
		t.Errorf("Expected default access key 'test', got %s", cfg.AccessKey)
	}
	if cfg.SecretKey != "test123456789" {
		t.Errorf("Expected default secret key 'test123456789', got %s", cfg.SecretKey)
	}
	if cfg.Region != "ap-southeast-3" {
		t.Errorf("Expected default region 'ap-southeast-3', got %s", cfg.Region)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected default log level 'debug', got %s", cfg.LogLevel)
	}
	if cfg.BaseDomain != "localhost" {
		t.Errorf("Expected default base domain 'localhost', got %s", cfg.BaseDomain)
	}
	if cfg.DisableAuth != false {
		t.Errorf("Expected default disable auth 'false', got %t", cfg.DisableAuth)
	}

	// Test custom values
	os.Setenv("PORT", "8080")
	os.Setenv("DATA_DIR", "/tmp/data")
	os.Setenv("ACCESS_KEY", "custom-key")
	os.Setenv("SECRET_KEY", "custom-secret")
	os.Setenv("REGION", "us-west-1")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("BASE_DOMAIN", "example.com")
	os.Setenv("DISABLE_AUTH", "true")

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}
	if cfg.DataDir != "/tmp/data" {
		t.Errorf("Expected data dir '/tmp/data', got %s", cfg.DataDir)
	}
	if cfg.AccessKey != "custom-key" {
		t.Errorf("Expected access key 'custom-key', got %s", cfg.AccessKey)
	}
	if cfg.SecretKey != "custom-secret" {
		t.Errorf("Expected secret key 'custom-secret', got %s", cfg.SecretKey)
	}
	if cfg.Region != "us-west-1" {
		t.Errorf("Expected region 'us-west-1', got %s", cfg.Region)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got %s", cfg.LogLevel)
	}
	if cfg.BaseDomain != "example.com" {
		t.Errorf("Expected base domain 'example.com', got %s", cfg.BaseDomain)
	}
	if cfg.DisableAuth != true {
		t.Errorf("Expected disable auth 'true', got %t", cfg.DisableAuth)
	}
}

func TestGetEnvFunctions(t *testing.T) {
	// Test getEnv
	os.Setenv("TEST_ENV", "test-value")
	if getEnv("TEST_ENV", "default") != "test-value" {
		t.Errorf("Expected getEnv to return 'test-value'")
	}
	if getEnv("NON_EXISTENT_ENV", "default") != "default" {
		t.Errorf("Expected getEnv to return default value 'default'")
	}

	// Test getEnvAsInt
	os.Setenv("TEST_INT", "42")
	if getEnvAsInt("TEST_INT", 0) != 42 {
		t.Errorf("Expected getEnvAsInt to return 42")
	}
	if getEnvAsInt("NON_EXISTENT_INT", 99) != 99 {
		t.Errorf("Expected getEnvAsInt to return default value 99")
	}
	os.Setenv("INVALID_INT", "not-an-int")
	if getEnvAsInt("INVALID_INT", 100) != 100 {
		t.Errorf("Expected getEnvAsInt to return default value 100 for invalid int")
	}

	// Test getEnvAsBool
	os.Setenv("TEST_BOOL_TRUE", "true")
	os.Setenv("TEST_BOOL_FALSE", "false")
	if getEnvAsBool("TEST_BOOL_TRUE", false) != true {
		t.Errorf("Expected getEnvAsBool to return true")
	}
	if getEnvAsBool("TEST_BOOL_FALSE", true) != false {
		t.Errorf("Expected getEnvAsBool to return false")
	}
	if getEnvAsBool("NON_EXISTENT_BOOL", true) != true {
		t.Errorf("Expected getEnvAsBool to return default value true")
	}
	os.Setenv("INVALID_BOOL", "not-a-bool")
	if getEnvAsBool("INVALID_BOOL", true) != true {
		t.Errorf("Expected getEnvAsBool to return default value true for invalid bool")
	}
}
