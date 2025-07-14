package main

import (
	"testing"

	"locals3/internal/config"
	"locals3/internal/handlers"

	"github.com/gorilla/mux"
)

func TestSetupLogging(t *testing.T) {
	// Test different log levels
	logLevels := []string{"debug", "info", "warn", "error", "invalid"}

	for _, level := range logLevels {
		setupLogging(level)
		// Simply ensure the function runs without panicking
	}
}

func TestSetupRouter(t *testing.T) {
	// Mock minimal handler
	h := &handlers.Handler{}

	// Test router setup
	router := setupRouter(h)

	// Ensure router is created
	if router == nil {
		t.Fatal("Expected router to be non-nil")
	}

	// Test that router responds to expected paths using the proper name lookup
	// In gorilla/mux, routes are typically named or matched by pattern, not by direct path
	routes := []*mux.Route{}
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		routes = append(routes, route)
		return nil
	})

	// Look for a route that matches the health check path
	foundHealthRoute := false
	for _, route := range routes {
		// Check if this route would match the health check path
		if path, _ := route.GetPathTemplate(); path == "/health" {
			foundHealthRoute = true
			break
		}
	}

	if !foundHealthRoute {
		t.Error("Health check route not found in router")
	}
}

func TestConfigLoad(t *testing.T) {
	// Test config loading
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check that config has default values
	if cfg.Port <= 0 {
		t.Errorf("Expected positive port number, got %d", cfg.Port)
	}

	if cfg.DataDir == "" {
		t.Error("Expected non-empty data directory")
	}
}
