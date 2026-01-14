// Configuration
// Centralized configuration management with environment variable support

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ServerConfig holds all server configuration
type ServerConfig struct {
	// Basic server configuration (from original Config)
	Port            string
	DatabaseURL     string
	DocumentStorage string
	BinaryStorage   string
	RulesStorage    string
	AdminAPIKey     string
	JWTSecret       string
	CORSOrigin      string

	// Advanced configuration
	Timeouts TimeoutConfig
	Limits   LimitsConfig
	Cache    CacheConfig
	Retry    RetryConfig
}

// TimeoutConfig holds timeout configurations
type TimeoutConfig struct {
	Query    time.Duration // Database query timeout
	Analysis time.Duration // Analysis operation timeout
	HTTP     time.Duration // HTTP request timeout
	Context  time.Duration // Default context timeout
}

// LimitsConfig holds size and rate limit configurations
type LimitsConfig struct {
	MaxFileSize              int64 // Maximum file size in bytes
	MaxStringLength          int   // Maximum string length
	MaxRequestSize           int64 // Maximum HTTP request size in bytes
	RateLimitRPS             int   // Rate limit requests per second
	RateLimitBurst           int   // Rate limit burst size
	MaxTaskTitleLength       int   // Maximum task title length
	MaxTaskDescriptionLength int   // Maximum task description length
	DefaultTaskListLimit     int   // Default task list limit
	MaxTaskListLimit         int   // Maximum task list limit
	DefaultDateRangeDays     int   // Default date range in days
}

// CacheConfig holds cache configurations
type CacheConfig struct {
	DefaultTTL      time.Duration // Default cache TTL
	TaskCacheTTL    time.Duration // Task cache TTL
	VerificationTTL time.Duration // Verification cache TTL
	DependencyTTL   time.Duration // Dependency cache TTL
	MaxSize         int           // Maximum cache size
	CleanupInterval time.Duration // Cache cleanup interval
}

// RetryConfig holds retry configurations
type RetryConfig struct {
	MaxRetries        int           // Maximum number of retries
	InitialBackoff    time.Duration // Initial backoff duration
	MaxBackoff        time.Duration // Maximum backoff duration
	BackoffMultiplier float64       // Backoff multiplier
}

var serverConfig *ServerConfig

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*ServerConfig, error) {
	if serverConfig != nil {
		return serverConfig, nil
	}

	config := &ServerConfig{
		// Basic server configuration
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable"),
		DocumentStorage: getEnv("DOCUMENT_STORAGE", "/data/documents"),
		BinaryStorage:   getEnv("BINARY_STORAGE", "/data/binaries"),
		RulesStorage:    getEnv("RULES_STORAGE", "/data/rules"),
		AdminAPIKey:     getEnv("ADMIN_API_KEY", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		CORSOrigin:      getEnv("CORS_ORIGIN", "*"),

		Timeouts: TimeoutConfig{
			Query:    parseDuration("SENTINEL_DB_TIMEOUT", 10*time.Second),
			Analysis: parseDuration("SENTINEL_ANALYSIS_TIMEOUT", 60*time.Second),
			HTTP:     parseDuration("SENTINEL_HTTP_TIMEOUT", 30*time.Second),
			Context:  parseDuration("SENTINEL_CONTEXT_TIMEOUT", 30*time.Second),
		},
		Limits: LimitsConfig{
			MaxFileSize:              parseInt64("SENTINEL_MAX_FILE_SIZE", 100*1024*1024),   // 100MB
			MaxStringLength:          parseInt("SENTINEL_MAX_STRING_LENGTH", 1000000),       // 1MB chars
			MaxRequestSize:           parseInt64("SENTINEL_MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
			RateLimitRPS:             parseInt("SENTINEL_RATE_LIMIT_RPS", 100),
			RateLimitBurst:           parseInt("SENTINEL_RATE_LIMIT_BURST", 200),
			MaxTaskTitleLength:       parseInt("SENTINEL_MAX_TASK_TITLE_LENGTH", 500),
			MaxTaskDescriptionLength: parseInt("SENTINEL_MAX_TASK_DESCRIPTION_LENGTH", 5000),
			DefaultTaskListLimit:     parseInt("SENTINEL_DEFAULT_TASK_LIST_LIMIT", 50),
			MaxTaskListLimit:         parseInt("SENTINEL_MAX_TASK_LIST_LIMIT", 1000),
			DefaultDateRangeDays:     parseInt("SENTINEL_DEFAULT_DATE_RANGE_DAYS", 30),
		},
		Cache: CacheConfig{
			DefaultTTL:      parseDuration("SENTINEL_CACHE_TTL", 5*time.Minute),
			TaskCacheTTL:    parseDuration("SENTINEL_TASK_CACHE_TTL", 5*time.Minute),
			VerificationTTL: parseDuration("SENTINEL_VERIFICATION_CACHE_TTL", 1*time.Hour),
			DependencyTTL:   parseDuration("SENTINEL_DEPENDENCY_CACHE_TTL", 10*time.Minute),
			MaxSize:         parseInt("SENTINEL_CACHE_MAX_SIZE", 10000),
			CleanupInterval: parseDuration("SENTINEL_CACHE_CLEANUP_INTERVAL", 5*time.Minute),
		},
		Retry: RetryConfig{
			MaxRetries:        parseInt("SENTINEL_MAX_RETRIES", 3),
			InitialBackoff:    parseDuration("SENTINEL_INITIAL_BACKOFF", 100*time.Millisecond),
			MaxBackoff:        parseDuration("SENTINEL_MAX_BACKOFF", 5*time.Second),
			BackoffMultiplier: parseFloat("SENTINEL_BACKOFF_MULTIPLIER", 2.0),
		},
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	serverConfig = config
	return config, nil
}

// GetConfig returns the current server configuration
func GetConfig() *ServerConfig {
	if serverConfig == nil {
		// Load config if not already loaded
		config, err := LoadConfig()
		if err != nil {
			// Return default config if loading fails (shouldn't happen in production)
			return &ServerConfig{
				Port:            "8080",
				DatabaseURL:     "postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable",
				DocumentStorage: "/data/documents",
				BinaryStorage:   "/data/binaries",
				RulesStorage:    "/data/rules",
				AdminAPIKey:     "",
				JWTSecret:       "",
				CORSOrigin:      "*",

				Timeouts: TimeoutConfig{
					Query:    10 * time.Second,
					Analysis: 60 * time.Second,
					HTTP:     30 * time.Second,
					Context:  30 * time.Second,
				},
				Limits: LimitsConfig{
					MaxFileSize:              100 * 1024 * 1024,
					MaxStringLength:          1000000,
					MaxRequestSize:           10 * 1024 * 1024,
					RateLimitRPS:             100,
					RateLimitBurst:           200,
					MaxTaskTitleLength:       500,
					MaxTaskDescriptionLength: 5000,
					DefaultTaskListLimit:     50,
					MaxTaskListLimit:         1000,
					DefaultDateRangeDays:     30,
				},
				Cache: CacheConfig{
					DefaultTTL:      5 * time.Minute,
					TaskCacheTTL:    5 * time.Minute,
					VerificationTTL: 1 * time.Hour,
					DependencyTTL:   10 * time.Minute,
					MaxSize:         10000,
					CleanupInterval: 5 * time.Minute,
				},
				Retry: RetryConfig{
					MaxRetries:        3,
					InitialBackoff:    100 * time.Millisecond,
					MaxBackoff:        5 * time.Second,
					BackoffMultiplier: 2.0,
				},
			}
		}
		return config
	}
	return serverConfig
}

// validateConfig validates all configuration values
func validateConfig(config *ServerConfig) error {
	// Validate timeouts
	if config.Timeouts.Query <= 0 {
		return fmt.Errorf("query timeout must be > 0")
	}
	if config.Timeouts.Analysis <= 0 {
		return fmt.Errorf("analysis timeout must be > 0")
	}
	if config.Timeouts.HTTP <= 0 {
		return fmt.Errorf("HTTP timeout must be > 0")
	}
	if config.Timeouts.Context <= 0 {
		return fmt.Errorf("context timeout must be > 0")
	}

	// Validate limits
	if config.Limits.MaxFileSize <= 0 {
		return fmt.Errorf("MaxFileSize must be > 0")
	}
	if config.Limits.MaxStringLength <= 0 {
		return fmt.Errorf("MaxStringLength must be > 0")
	}
	if config.Limits.MaxRequestSize <= 0 {
		return fmt.Errorf("MaxRequestSize must be > 0")
	}
	if config.Limits.RateLimitRPS <= 0 {
		return fmt.Errorf("RateLimitRPS must be > 0")
	}
	if config.Limits.RateLimitBurst <= 0 {
		return fmt.Errorf("RateLimitBurst must be > 0")
	}
	if config.Limits.MaxTaskTitleLength <= 0 {
		return fmt.Errorf("MaxTaskTitleLength must be > 0")
	}
	if config.Limits.MaxTaskDescriptionLength <= 0 {
		return fmt.Errorf("MaxTaskDescriptionLength must be > 0")
	}
	if config.Limits.DefaultTaskListLimit <= 0 {
		return fmt.Errorf("DefaultTaskListLimit must be > 0")
	}
	if config.Limits.MaxTaskListLimit <= 0 {
		return fmt.Errorf("MaxTaskListLimit must be > 0")
	}
	if config.Limits.DefaultDateRangeDays <= 0 {
		return fmt.Errorf("DefaultDateRangeDays must be > 0")
	}

	// Validate cache
	if config.Cache.DefaultTTL <= 0 {
		return fmt.Errorf("DefaultTTL must be > 0")
	}
	if config.Cache.TaskCacheTTL <= 0 {
		return fmt.Errorf("TaskCacheTTL must be > 0")
	}
	if config.Cache.VerificationTTL <= 0 {
		return fmt.Errorf("VerificationTTL must be > 0")
	}
	if config.Cache.DependencyTTL <= 0 {
		return fmt.Errorf("DependencyTTL must be > 0")
	}
	if config.Cache.MaxSize <= 0 {
		return fmt.Errorf("cache MaxSize must be > 0")
	}
	if config.Cache.CleanupInterval <= 0 {
		return fmt.Errorf("CleanupInterval must be > 0")
	}

	// Validate retry
	if config.Retry.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries must be >= 0")
	}
	if config.Retry.InitialBackoff <= 0 {
		return fmt.Errorf("InitialBackoff must be > 0")
	}
	if config.Retry.MaxBackoff < config.Retry.InitialBackoff {
		return fmt.Errorf("MaxBackoff must be >= InitialBackoff")
	}
	if config.Retry.BackoffMultiplier <= 1.0 {
		return fmt.Errorf("BackoffMultiplier must be > 1.0")
	}

	return nil
}

func parseDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func parseInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
