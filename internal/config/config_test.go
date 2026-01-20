// Package config provides tests for configuration management
package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Run("loads default configuration", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Errorf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Error("config should not be nil")
		}
	})

	t.Run("has correct server defaults", func(t *testing.T) {
		cfg, _ := Load()

		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
		}
		if cfg.Server.Port != 8080 {
			t.Errorf("expected port 8080, got %d", cfg.Server.Port)
		}
		if cfg.Server.ReadTimeout != 30*time.Second {
			t.Errorf("expected read timeout 30s, got %v", cfg.Server.ReadTimeout)
		}
	})

	t.Run("has correct database defaults", func(t *testing.T) {
		cfg, _ := Load()

		if cfg.Database.MaxOpenConns != 25 {
			t.Errorf("expected max open conns 25, got %d", cfg.Database.MaxOpenConns)
		}
		if cfg.Database.MaxIdleConns != 5 {
			t.Errorf("expected max idle conns 5, got %d", cfg.Database.MaxIdleConns)
		}
	})

	t.Run("has correct security defaults", func(t *testing.T) {
		cfg, _ := Load()

		if cfg.Security.BcryptCost != 12 {
			t.Errorf("expected bcrypt cost 12, got %d", cfg.Security.BcryptCost)
		}
		if cfg.Security.RateLimitRequests != 100 {
			t.Errorf("expected rate limit 100, got %d", cfg.Security.RateLimitRequests)
		}
	})

	t.Run("has correct LLM defaults", func(t *testing.T) {
		cfg, _ := Load()

		if cfg.LLM.OllamaHost != "http://localhost:11434" {
			t.Errorf("expected ollama host http://localhost:11434, got %s", cfg.LLM.OllamaHost)
		}
		if cfg.LLM.MaxRetries != 3 {
			t.Errorf("expected max retries 3, got %d", cfg.LLM.MaxRetries)
		}
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("returns default when env not set", func(t *testing.T) {
		result := getEnv("NONEXISTENT_VAR_12345", "default")
		if result != "default" {
			t.Errorf("expected default, got %s", result)
		}
	})

	t.Run("returns env value when set", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "custom_value")
		defer os.Unsetenv("TEST_ENV_VAR")

		result := getEnv("TEST_ENV_VAR", "default")
		if result != "custom_value" {
			t.Errorf("expected custom_value, got %s", result)
		}
	})
}

func TestGetEnvAsInt(t *testing.T) {
	t.Run("returns default when env not set", func(t *testing.T) {
		result := getEnvAsInt("NONEXISTENT_INT_VAR", 42)
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
	})

	t.Run("returns parsed int when set", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR", "100")
		defer os.Unsetenv("TEST_INT_VAR")

		result := getEnvAsInt("TEST_INT_VAR", 42)
		if result != 100 {
			t.Errorf("expected 100, got %d", result)
		}
	})

	t.Run("returns default for invalid int", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR_INVALID", "not_a_number")
		defer os.Unsetenv("TEST_INT_VAR_INVALID")

		result := getEnvAsInt("TEST_INT_VAR_INVALID", 42)
		if result != 42 {
			t.Errorf("expected 42 for invalid int, got %d", result)
		}
	})
}

func TestGetEnvAsDuration(t *testing.T) {
	t.Run("returns default when env not set", func(t *testing.T) {
		result := getEnvAsDuration("NONEXISTENT_DURATION_VAR", 5*time.Minute)
		if result != 5*time.Minute {
			t.Errorf("expected 5m, got %v", result)
		}
	})

	t.Run("returns parsed duration when set", func(t *testing.T) {
		os.Setenv("TEST_DURATION_VAR", "10s")
		defer os.Unsetenv("TEST_DURATION_VAR")

		result := getEnvAsDuration("TEST_DURATION_VAR", 5*time.Minute)
		if result != 10*time.Second {
			t.Errorf("expected 10s, got %v", result)
		}
	})

	t.Run("parses minutes correctly", func(t *testing.T) {
		os.Setenv("TEST_DURATION_MIN", "15m")
		defer os.Unsetenv("TEST_DURATION_MIN")

		result := getEnvAsDuration("TEST_DURATION_MIN", time.Hour)
		if result != 15*time.Minute {
			t.Errorf("expected 15m, got %v", result)
		}
	})

	t.Run("parses hours correctly", func(t *testing.T) {
		os.Setenv("TEST_DURATION_HOUR", "2h")
		defer os.Unsetenv("TEST_DURATION_HOUR")

		result := getEnvAsDuration("TEST_DURATION_HOUR", time.Minute)
		if result != 2*time.Hour {
			t.Errorf("expected 2h, got %v", result)
		}
	})

	t.Run("returns default for invalid duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION_INVALID", "not_a_duration")
		defer os.Unsetenv("TEST_DURATION_INVALID")

		result := getEnvAsDuration("TEST_DURATION_INVALID", 5*time.Minute)
		if result != 5*time.Minute {
			t.Errorf("expected 5m for invalid duration, got %v", result)
		}
	})
}

func TestGetEnvAsSlice(t *testing.T) {
	t.Run("returns default when env not set", func(t *testing.T) {
		defaultSlice := []string{"a", "b", "c"}
		result := getEnvAsSlice("NONEXISTENT_SLICE_VAR", defaultSlice)
		if len(result) != 3 {
			t.Errorf("expected 3 elements, got %d", len(result))
		}
	})

	t.Run("returns value when set", func(t *testing.T) {
		os.Setenv("TEST_SLICE_VAR", "http://localhost:3000")
		defer os.Unsetenv("TEST_SLICE_VAR")

		result := getEnvAsSlice("TEST_SLICE_VAR", []string{"default"})
		if len(result) == 0 {
			t.Error("result should not be empty")
		}
		if result[0] != "http://localhost:3000" {
			t.Errorf("expected http://localhost:3000, got %s", result[0])
		}
	})
}

func TestServerConfig(t *testing.T) {
	cfg := ServerConfig{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
}

func TestDatabaseConfig(t *testing.T) {
	cfg := DatabaseConfig{
		URL:             "postgres://user:pass@localhost/db",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	if cfg.MaxOpenConns != 25 {
		t.Errorf("expected max open conns 25, got %d", cfg.MaxOpenConns)
	}
}

func TestSecurityConfig(t *testing.T) {
	cfg := SecurityConfig{
		JWTSecret:          "secret",
		JWTExpiration:      24 * time.Hour,
		BcryptCost:         12,
		RateLimitRequests:  100,
		RateLimitWindow:    15 * time.Minute,
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}

	if cfg.BcryptCost != 12 {
		t.Errorf("expected bcrypt cost 12, got %d", cfg.BcryptCost)
	}
	if len(cfg.CORSAllowedOrigins) != 1 {
		t.Error("expected 1 CORS origin")
	}
}

func TestLLMConfig(t *testing.T) {
	cfg := LLMConfig{
		OllamaHost:         "http://localhost:11434",
		AzureAIEndpoint:    "https://azure.endpoint",
		AzureAIKey:         "azure-key",
		AzureAIDeployment:  "claude",
		AzureAPIVersion:    "2024-02-01",
		RequestTimeout:     60 * time.Second,
		MaxRetries:         3,
		RateLimitPerMinute: 60,
	}

	if cfg.OllamaHost != "http://localhost:11434" {
		t.Errorf("expected ollama host, got %s", cfg.OllamaHost)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected max retries 3, got %d", cfg.MaxRetries)
	}
}

func TestLoadWithEnvironmentOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("PORT", "9090")
	os.Setenv("BCRYPT_COST", "14")
	defer func() {
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
		os.Unsetenv("BCRYPT_COST")
	}()

	cfg, err := Load()
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Security.BcryptCost != 14 {
		t.Errorf("expected bcrypt cost 14, got %d", cfg.Security.BcryptCost)
	}
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("writes file successfully", func(t *testing.T) {
		path := tmpDir + "/test.txt"
		err := WriteFile(path, "test content")
		if err != nil {
			t.Errorf("WriteFile() error = %v", err)
		}

		content, _ := os.ReadFile(path)
		if string(content) != "test content" {
			t.Errorf("expected 'test content', got '%s'", string(content))
		}
	})

	t.Run("handles write error", func(t *testing.T) {
		// Try to write to a non-existent directory
		err := WriteFile("/nonexistent/dir/file.txt", "content")
		if err == nil {
			t.Error("should return error for invalid path")
		}
	})
}

func TestSecureGitIgnore(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	t.Run("appends to existing gitignore", func(t *testing.T) {
		os.WriteFile(".gitignore", []byte("node_modules/\n"), 0644)

		err := SecureGitIgnore()
		if err != nil {
			t.Errorf("SecureGitIgnore() error = %v", err)
		}

		content, _ := os.ReadFile(".gitignore")
		if !containsString(string(content), "Sentinel Rules") {
			t.Error("should add Sentinel entries")
		}
		if !containsString(string(content), "node_modules") {
			t.Error("should preserve existing entries")
		}
	})

	t.Run("creates gitignore if missing", func(t *testing.T) {
		os.Remove(".gitignore")

		err := SecureGitIgnore()
		if err != nil {
			t.Errorf("SecureGitIgnore() error = %v", err)
		}

		if _, err := os.Stat(".gitignore"); os.IsNotExist(err) {
			t.Error("should create .gitignore")
		}
	})
}

func TestCreateCI(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	t.Run("creates workflow file", func(t *testing.T) {
		os.MkdirAll(".github/workflows", 0755)

		err := CreateCI()
		if err != nil {
			t.Errorf("CreateCI() error = %v", err)
		}

		if _, err := os.Stat(".github/workflows/sentinel.yml"); os.IsNotExist(err) {
			t.Error("should create sentinel.yml")
		}
	})
}

// Helper function for string contains check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
