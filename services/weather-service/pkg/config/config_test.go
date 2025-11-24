package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("default values without environment variables", func(t *testing.T) {
		// Clear any existing env vars
		os.Unsetenv("SERVICE_NAME")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("PORT")
		os.Unsetenv("GRPC_PORT")

		cfg := Load()

		if cfg == nil {
			t.Fatal("Expected config to be loaded, got nil")
		}
		if cfg.ServiceName != "weather-service" {
			t.Errorf("Expected ServiceName 'weather-service', got %s", cfg.ServiceName)
		}
		if cfg.Environment != "development" {
			t.Errorf("Expected Environment 'development', got %s", cfg.Environment)
		}
		if cfg.Port != "8080" {
			t.Errorf("Expected Port '8080', got %s", cfg.Port)
		}
		if cfg.GRPCPort != "9090" {
			t.Errorf("Expected GRPCPort '9090', got %s", cfg.GRPCPort)
		}
	})

	t.Run("custom values from environment variables", func(t *testing.T) {
		// Set custom env vars
		os.Setenv("SERVICE_NAME", "custom-weather-service")
		os.Setenv("ENVIRONMENT", "production")
		os.Setenv("PORT", "3000")
		os.Setenv("GRPC_PORT", "5000")

		cfg := Load()

		if cfg.ServiceName != "custom-weather-service" {
			t.Errorf("Expected ServiceName 'custom-weather-service', got %s", cfg.ServiceName)
		}
		if cfg.Environment != "production" {
			t.Errorf("Expected Environment 'production', got %s", cfg.Environment)
		}
		if cfg.Port != "3000" {
			t.Errorf("Expected Port '3000', got %s", cfg.Port)
		}
		if cfg.GRPCPort != "5000" {
			t.Errorf("Expected GRPCPort '5000', got %s", cfg.GRPCPort)
		}

		// Clean up
		os.Unsetenv("SERVICE_NAME")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("PORT")
		os.Unsetenv("GRPC_PORT")
	})

	t.Run("partial environment variables", func(t *testing.T) {
		// Set only some env vars
		os.Setenv("SERVICE_NAME", "partial-service")
		os.Setenv("PORT", "4000")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("GRPC_PORT")

		cfg := Load()

		if cfg.ServiceName != "partial-service" {
			t.Errorf("Expected ServiceName 'partial-service', got %s", cfg.ServiceName)
		}
		if cfg.Port != "4000" {
			t.Errorf("Expected Port '4000', got %s", cfg.Port)
		}
		// Should use defaults for unset vars
		if cfg.Environment != "development" {
			t.Errorf("Expected default Environment 'development', got %s", cfg.Environment)
		}
		if cfg.GRPCPort != "9090" {
			t.Errorf("Expected default GRPCPort '9090', got %s", cfg.GRPCPort)
		}

		// Clean up
		os.Unsetenv("SERVICE_NAME")
		os.Unsetenv("PORT")
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("returns environment variable when set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")

		result := getEnv("TEST_VAR", "default_value")

		if result != "test_value" {
			t.Errorf("Expected 'test_value', got %s", result)
		}

		os.Unsetenv("TEST_VAR")
	})

	t.Run("returns default value when environment variable not set", func(t *testing.T) {
		os.Unsetenv("NONEXISTENT_VAR")

		result := getEnv("NONEXISTENT_VAR", "default_value")

		if result != "default_value" {
			t.Errorf("Expected 'default_value', got %s", result)
		}
	})

	t.Run("returns empty string from environment over default", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")

		result := getEnv("EMPTY_VAR", "default_value")

		// Empty string should return default
		if result != "default_value" {
			t.Errorf("Expected 'default_value' for empty env var, got %s", result)
		}

		os.Unsetenv("EMPTY_VAR")
	})

	t.Run("handles special characters in value", func(t *testing.T) {
		os.Setenv("SPECIAL_VAR", "value-with-special_chars:123/test")

		result := getEnv("SPECIAL_VAR", "default")

		if result != "value-with-special_chars:123/test" {
			t.Errorf("Expected 'value-with-special_chars:123/test', got %s", result)
		}

		os.Unsetenv("SPECIAL_VAR")
	})
}

func TestConfig_AllFields(t *testing.T) {
	cfg := &Config{
		ServiceName: "test-service",
		Environment: "test",
		Port:        "9999",
		GRPCPort:    "8888",
	}

	if cfg.ServiceName != "test-service" {
		t.Errorf("Expected ServiceName 'test-service', got %s", cfg.ServiceName)
	}
	if cfg.Environment != "test" {
		t.Errorf("Expected Environment 'test', got %s", cfg.Environment)
	}
	if cfg.Port != "9999" {
		t.Errorf("Expected Port '9999', got %s", cfg.Port)
	}
	if cfg.GRPCPort != "8888" {
		t.Errorf("Expected GRPCPort '8888', got %s", cfg.GRPCPort)
	}
}

func TestConfig_ZeroValues(t *testing.T) {
	cfg := &Config{}

	if cfg.ServiceName != "" {
		t.Error("Zero value ServiceName should be empty string")
	}
	if cfg.Environment != "" {
		t.Error("Zero value Environment should be empty string")
	}
	if cfg.Port != "" {
		t.Error("Zero value Port should be empty string")
	}
	if cfg.GRPCPort != "" {
		t.Error("Zero value GRPCPort should be empty string")
	}
}
