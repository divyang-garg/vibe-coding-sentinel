// Package config provides centralized configuration management.
// Complies with CODING_STANDARDS.md: Configuration management
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Security SecurityConfig `json:"security"`
	Logging  LoggingConfig  `json:"logging"`
	Services ServicesConfig `json:"services"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	User         string        `json:"user"`
	Password     string        `json:"password"`
	Database     string        `json:"database"`
	SSLMode      string        `json:"ssl_mode"`
	MaxOpenConns int           `json:"max_open_conns"`
	MaxIdleConns int           `json:"max_idle_conns"`
	MaxLifetime  time.Duration `json:"max_lifetime"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	APIKeys            []string      `json:"api_keys"`
	RateLimitMax       float64       `json:"rate_limit_max"`
	RateLimitRefill    float64       `json:"rate_limit_refill"`
	JWTSecret          string        `json:"jwt_secret"`
	CORSAllowedOrigins []string      `json:"cors_allowed_origins"`
	SessionTimeout     time.Duration `json:"session_timeout"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	FilePath   string `json:"file_path,omitempty"`
	MaxSize    int    `json:"max_size_mb,omitempty"`
	MaxBackups int    `json:"max_backups,omitempty"`
	MaxAge     int    `json:"max_age_days,omitempty"`
}

// ServicesConfig holds external service configurations
type ServicesConfig struct {
	Timeout        time.Duration        `json:"timeout"`
	RetryCount     int                  `json:"retry_count"`
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker"`
	Cache          CacheConfig          `json:"cache"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled bool          `json:"enabled"`
	TTL     time.Duration `json:"ttl"`
	MaxSize int           `json:"max_size_mb"`
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	config := &Config{}

	// Load from config file if exists
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		if err := loadFromFile(configFile, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	// Set defaults for missing values
	setDefaults(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromFile loads configuration from a JSON file
func loadFromFile(filename string, config *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(config)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Server config
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	// Database config
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if database := os.Getenv("DB_NAME"); database != "" {
		config.Database.Database = database
	}

	// Security config
	if apiKeys := os.Getenv("API_KEYS"); apiKeys != "" {
		config.Security.APIKeys = []string{apiKeys} // Simplified, should parse comma-separated
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Security.JWTSecret = jwtSecret
	}
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	// Server defaults
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30 * time.Second
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 120 * time.Second
	}

	// Database defaults
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.User == "" {
		config.Database.User = "postgres"
	}
	if config.Database.Database == "" {
		config.Database.Database = "sentinel"
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 5
	}
	if config.Database.MaxLifetime == 0 {
		config.Database.MaxLifetime = 5 * time.Minute
	}

	// Security defaults
	if len(config.Security.APIKeys) == 0 {
		config.Security.APIKeys = []string{"dev-api-key-123", "test-api-key-456"}
	}
	if config.Security.RateLimitMax == 0 {
		config.Security.RateLimitMax = 100
	}
	if config.Security.RateLimitRefill == 0 {
		config.Security.RateLimitRefill = 10
	}
	if config.Security.JWTSecret == "" {
		config.Security.JWTSecret = "dev-jwt-secret-change-in-production"
	}
	if len(config.Security.CORSAllowedOrigins) == 0 {
		config.Security.CORSAllowedOrigins = []string{"*"}
	}
	if config.Security.SessionTimeout == 0 {
		config.Security.SessionTimeout = 24 * time.Hour
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	// Services defaults
	if config.Services.Timeout == 0 {
		config.Services.Timeout = 30 * time.Second
	}
	if config.Services.RetryCount == 0 {
		config.Services.RetryCount = 3
	}
	if config.Services.CircuitBreaker.FailureThreshold == 0 {
		config.Services.CircuitBreaker.FailureThreshold = 5
	}
	if config.Services.CircuitBreaker.ResetTimeout == 0 {
		config.Services.CircuitBreaker.ResetTimeout = 60 * time.Second
	}
	if config.Services.Cache.TTL == 0 {
		config.Services.Cache.TTL = 5 * time.Minute
	}
	if config.Services.Cache.MaxSize == 0 {
		config.Services.Cache.MaxSize = 100
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.MaxOpenConns < config.Database.MaxIdleConns {
		return fmt.Errorf("max_open_conns must be >= max_idle_conns")
	}

	if len(config.Security.APIKeys) == 0 {
		return fmt.Errorf("at least one API key must be configured")
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", config.Logging.Level)
	}

	return nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if this is a development environment
func (c *Config) IsDevelopment() bool {
	return c.Logging.Level == "debug"
}

// IsAPIKeyValid validates an API key
func (c *Config) IsAPIKeyValid(apiKey string) bool {
	for _, key := range c.Security.APIKeys {
		if key == apiKey {
			return true
		}
	}
	return false
}
