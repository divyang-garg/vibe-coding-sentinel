// Package config provides configuration management
// Complies with CODING_STANDARDS.md: Configuration files max 200 lines
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
	LLM      LLMConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWTSecret          string
	JWTExpiration      time.Duration
	BcryptCost         int
	RateLimitRequests  int
	RateLimitWindow    time.Duration
	CORSAllowedOrigins []string
}

// LLMConfig holds LLM service configuration
type LLMConfig struct {
	OllamaHost         string
	AzureAIEndpoint    string
	AzureAIKey         string
	AzureAIDeployment  string
	AzureAPIVersion    string
	RequestTimeout     time.Duration
	MaxRetries         int
	RateLimitPerMinute int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("HOST", "0.0.0.0"),
			Port:         getEnvAsInt("PORT", 8080),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://sentinel:password@localhost/sentinel?sslmode=disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Security: SecurityConfig{
			JWTSecret:          getEnv("JWT_SECRET", "your-super-secure-jwt-secret-change-in-production"),
			JWTExpiration:      getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
			BcryptCost:         getEnvAsInt("BCRYPT_COST", 12),
			RateLimitRequests:  getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:    getEnvAsDuration("RATE_LIMIT_WINDOW", 15*time.Minute),
			CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
		LLM: LLMConfig{
			OllamaHost:         getEnv("OLLAMA_HOST", "http://localhost:11434"),
			AzureAIEndpoint:    getEnv("AZURE_AI_ENDPOINT", ""),
			AzureAIKey:         getEnv("AZURE_AI_KEY", ""),
			AzureAIDeployment:  getEnv("AZURE_AI_DEPLOYMENT", "claude-opus-4-5"),
			AzureAPIVersion:    getEnv("AZURE_AI_API_VERSION", "2024-02-01"),
			RequestTimeout:     getEnvAsDuration("LLM_REQUEST_TIMEOUT", 60*time.Second),
			MaxRetries:         getEnvAsInt("LLM_MAX_RETRIES", 3),
			RateLimitPerMinute: getEnvAsInt("LLM_RATE_LIMIT_PER_MINUTE", 60),
		},
	}

	return cfg, nil
}

// Helper functions for environment variable parsing

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

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// In production, consider more robust parsing
		return []string{value}
	}
	return defaultValue
}

// GetHubConfig reads Hub configuration from environment variables
// Returns Hub URL and API key (empty string if not set)
func GetHubConfig() (url, apiKey string) {
	url = getEnv("SENTINEL_HUB_URL", "http://localhost:8080")
	apiKey = getEnv("SENTINEL_HUB_API_KEY", "")
	return url, apiKey
}
