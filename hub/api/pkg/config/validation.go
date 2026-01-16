// Package config provides configuration validation utilities
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package config

import (
	"log"
	"os"
	"strings"
)

// ProductionConfig represents configuration for production validation
type ProductionConfig struct {
	CORSOrigin  string
	JWTSecret   string
	DatabaseURL string
}

// ValidateProductionConfig validates production configuration and fails startup if insecure defaults detected
func ValidateProductionConfig(config *ProductionConfig) {
	env := getEnv("ENVIRONMENT", "development")
	if env != "production" {
		return // Skip validation in non-production
	}

	var errors []string

	// Check CORS
	if config.CORSOrigin == "*" {
		errors = append(errors, "CORS_ORIGIN cannot be '*' in production")
	}

	// Check JWT Secret
	if config.JWTSecret == "change-me-in-production" {
		errors = append(errors, "JWT_SECRET must be changed from default value")
	}

	// Check Database SSL
	if strings.Contains(config.DatabaseURL, "sslmode=disable") {
		errors = append(errors, "Database connection must use SSL (sslmode=require) in production")
	}

	// Check for default password in connection string
	if strings.Contains(config.DatabaseURL, "sentinel:sentinel@") {
		errors = append(errors, "Database password must not be default 'sentinel' in production")
	}

	if len(errors) > 0 {
		log.Fatalf("❌ PRODUCTION CONFIGURATION ERRORS:\n%s\n\nPlease fix these issues before starting in production mode.", strings.Join(errors, "\n"))
	}

	log.Println("✅ Production configuration validated")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
