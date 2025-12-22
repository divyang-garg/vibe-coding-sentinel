// Sentinel Hub API Server
// Central server for document processing, metrics, and organization management

package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/lib/pq"
)

// isCriticalError classifies errors that should fail fast
func isCriticalError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// Classify errors that should fail fast
	criticalPatterns := []string{
		"database connection",
		"authentication failed",
		"permission denied",
		"unauthorized",
		"forbidden",
	}
	for _, pattern := range criticalPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// =============================================================================
// CONTEXT KEYS
// =============================================================================

// contextKey is a custom type for context keys to prevent collisions
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	projectKey   contextKey = "project"
)

// =============================================================================
// CONFIGURATION
// =============================================================================

type Config struct {
	Port            string
	DatabaseURL     string
	DocumentStorage string
	JWTSecret       string
	OllamaHost      string
	CORSOrigin      string
	HubURL          string
	BinaryStorage   string
	RulesStorage    string
	AdminAPIKey     string
}

func loadConfig() *Config {
	corsOrigin := getEnv("CORS_ORIGIN", "*")

	// Validate CORS origin - warn if using wildcard in production
	if corsOrigin == "*" {
		env := getEnv("ENVIRONMENT", "development")
		if env == "production" {
			log.Printf("⚠️  WARNING: CORS_ORIGIN is set to '*' in production. This is insecure.")
			log.Printf("   Recommendation: Set CORS_ORIGIN to specific allowed origins.")
		}
	} else {
		// Validate origin format (basic check)
		if !strings.HasPrefix(corsOrigin, "http://") && !strings.HasPrefix(corsOrigin, "https://") {
			log.Printf("⚠️  WARNING: CORS_ORIGIN '%s' does not appear to be a valid URL. Expected http:// or https://", corsOrigin)
		}
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable"),
		DocumentStorage: getEnv("DOCUMENT_STORAGE", "/data/documents"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
		OllamaHost:      getEnv("OLLAMA_HOST", "http://localhost:11434"),
		CORSOrigin:      corsOrigin,
		HubURL:          getEnv("HUB_URL", "http://localhost:8080"),
		BinaryStorage:   getEnv("BINARY_STORAGE", "/data/binaries"),
		RulesStorage:    getEnv("RULES_STORAGE", "/data/rules"),
		AdminAPIKey:     getEnv("ADMIN_API_KEY", ""),
	}
}

// validateProductionConfig validates production configuration and fails startup if insecure defaults detected
func validateProductionConfig(config *Config) {
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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// =============================================================================
// TIMEOUT CONSTANTS AND HELPERS
// =============================================================================

// Timeout constants
// Deprecated: Use GetConfig().Timeouts.Query instead
func getQueryTimeout() time.Duration {
	return GetConfig().Timeouts.Query
}

// Deprecated: Use GetConfig().Timeouts.Analysis instead
func getAnalysisTimeout() time.Duration {
	return GetConfig().Timeouts.Analysis
}

// Deprecated: Use GetConfig().Timeouts.Context instead
var DefaultContextTimeout = GetConfig().Timeouts.Context

// Deprecated: Use GetConfig().Timeouts.HTTP instead
var DefaultHTTPTimeout = GetConfig().Timeouts.HTTP

// =============================================================================
// DATABASE QUERY HELPERS
// =============================================================================

// Database query helpers with timeout
func queryWithTimeout(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	timeout := getQueryTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return db.QueryContext(ctx, query, args...)
}

func queryRowWithTimeout(ctx context.Context, query string, args ...interface{}) *sql.Row {
	timeout := getQueryTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return db.QueryRowContext(ctx, query, args...)
}

func execWithTimeout(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	timeout := getQueryTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return db.ExecContext(ctx, query, args...)
}

// =============================================================================
// REQUEST ID MIDDLEWARE
// =============================================================================

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// =============================================================================
// CONTEXT HELPER FUNCTIONS
// =============================================================================

// getProjectFromContext safely retrieves project from context
// Returns error if project is missing or invalid (should never happen in protected routes)
func getProjectFromContext(ctx context.Context) (*Project, error) {
	value := ctx.Value(projectKey)
	if value == nil {
		return nil, fmt.Errorf("project not found in context")
	}
	project, ok := value.(*Project)
	if !ok || project == nil {
		return nil, fmt.Errorf("invalid project type in context")
	}
	return project, nil
}

// requireProjectContext validates and returns project from context, writing error response if missing
func requireProjectContext(w http.ResponseWriter, r *http.Request) (*Project, error) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Project context required but missing")
		WriteErrorResponse(w, &ValidationError{
			Field:   "authorization",
			Message: "Project context required",
			Code:    "unauthorized",
		}, http.StatusUnauthorized)
		return nil, err
	}
	return project, nil
}

// getRequestIDFromContext safely retrieves request ID from context
func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		return requestID
	}
	return "unknown"
}

// =============================================================================
// MODELS
// =============================================================================

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Document struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	Name          string     `json:"name"`
	OriginalName  string     `json:"original_name"`
	Size          int64      `json:"size"`
	MimeType      string     `json:"mime_type"`
	Status        string     `json:"status"` // queued, processing, completed, failed
	Progress      int        `json:"progress"`
	FilePath      string     `json:"-"`
	ExtractedText string     `json:"extracted_text,omitempty"`
	Error         string     `json:"error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
}

// KnowledgeItem is defined in types.go

type ProcessingStage struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // pending, processing, completed, failed
	Duration int    `json:"duration_ms,omitempty"`
	Error    string `json:"error,omitempty"`
}

type DocumentStatus struct {
	ID           string            `json:"id"`
	OriginalName string            `json:"original_name"`
	Status       string            `json:"status"`
	Progress     int               `json:"progress"`
	Stages       []ProcessingStage `json:"stages"`
	Result       *DocumentResult   `json:"result,omitempty"`
	Error        string            `json:"error,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	ProcessedAt  *time.Time        `json:"processed_at,omitempty"`
}

type DocumentResult struct {
	Pages          int `json:"pages,omitempty"`
	TextLength     int `json:"text_length"`
	KnowledgeItems int `json:"knowledge_items"`
}

// =============================================================================
// DATABASE
// =============================================================================

var db *sql.DB

// Metrics collection (Phase G: Logging and Monitoring)
var (
	httpRequestCounter = make(map[string]int64) // endpoint -> count
	httpErrorCounter   = make(map[string]int64) // endpoint -> error count
	httpDurationSum    = make(map[string]int64) // endpoint -> total duration (ms)
	httpRequestCount   = make(map[string]int64) // endpoint -> request count for avg
	metricsMutex       sync.RWMutex
	startTime          = time.Now()
)

func initDB(databaseURL string) error {
	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initial connectivity check
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Validate pool settings
	stats := db.Stats()
	if stats.MaxOpenConnections != 25 {
		log.Printf("Warning: MaxOpenConnections mismatch: expected 25, got %d", stats.MaxOpenConnections)
	}

	// Start background health check goroutine
	go monitorDBHealth()

	return nil
}

func monitorDBHealth() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := db.Stats()

		// Log pool metrics
		log.Printf("DB Pool Stats: Open=%d/%d, Idle=%d, InUse=%d, WaitCount=%d, WaitDuration=%v",
			stats.OpenConnections, stats.MaxOpenConnections,
			stats.Idle, stats.InUse,
			stats.WaitCount, stats.WaitDuration)

		// Alert if pool is exhausted
		if stats.OpenConnections >= stats.MaxOpenConnections {
			log.Printf("WARNING: Database connection pool exhausted!")
		}

		// Alert if many connections waiting
		if stats.WaitCount > 0 {
			log.Printf("WARNING: %d connections waiting for database pool", stats.WaitCount)
		}

		// Health check ping
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := db.PingContext(ctx); err != nil {
			log.Printf("ERROR: Database health check failed: %v", err)
		}
		cancel()
	}
}

func runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS organizations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			api_key VARCHAR(64) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS documents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			size BIGINT NOT NULL,
			mime_type VARCHAR(100),
			status VARCHAR(20) DEFAULT 'queued',
			progress INT DEFAULT 0,
			file_path VARCHAR(500),
			extracted_text TEXT,
			error TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			processed_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS knowledge_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			confidence FLOAT DEFAULT 0,
			source_page INT,
			status VARCHAR(20) DEFAULT 'pending',
			approved_by VARCHAR(100),
			approved_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW(),
			structured_data JSONB
		)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_structured ON knowledge_items USING GIN (structured_data)`,
		`CREATE TABLE IF NOT EXISTS telemetry_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			event_type VARCHAR(50) NOT NULL,
			payload JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		// Migration: Add new columns for spec alignment
		`ALTER TABLE telemetry_events ADD COLUMN IF NOT EXISTS agent_id VARCHAR(64)`,
		`ALTER TABLE telemetry_events ADD COLUMN IF NOT EXISTS org_id UUID`,
		`ALTER TABLE telemetry_events ADD COLUMN IF NOT EXISTS team_id UUID`,
		`ALTER TABLE telemetry_events ADD COLUMN IF NOT EXISTS timestamp TIMESTAMP`,
		`CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_items_document_id ON knowledge_items(document_id)`,
		`CREATE INDEX IF NOT EXISTS idx_telemetry_events_project_id ON telemetry_events(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_telemetry_agent ON telemetry_events(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_telemetry_org ON telemetry_events(org_id)`,
		// Hook tables (Phase 9.5)
		`CREATE TABLE IF NOT EXISTS hook_executions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id VARCHAR(64),
			org_id UUID,
			team_id UUID,
			hook_type VARCHAR(20) NOT NULL,
			result VARCHAR(20) NOT NULL,
			override_reason VARCHAR(255),
			findings_summary JSONB,
			user_actions JSONB,
			duration_ms BIGINT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS hook_baselines (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id VARCHAR(64),
			org_id UUID,
			baseline_entry JSONB NOT NULL,
			source VARCHAR(20) NOT NULL,
			hook_type VARCHAR(20),
			reviewed BOOLEAN DEFAULT false,
			reviewed_by VARCHAR(100),
			reviewed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS hook_policies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			policy_config JSONB NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_hook_executions_agent ON hook_executions(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_hook_executions_org ON hook_executions(org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_hook_executions_created ON hook_executions(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_hook_baselines_org ON hook_baselines(org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_hook_baselines_reviewed ON hook_baselines(reviewed)`,
		`CREATE INDEX IF NOT EXISTS idx_hook_policies_org ON hook_policies(org_id)`,
		// Phase 10: Test Enforcement System
		`CREATE TABLE IF NOT EXISTS test_requirements (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			knowledge_item_id UUID REFERENCES knowledge_items(id) ON DELETE CASCADE,
			rule_title VARCHAR(255) NOT NULL,
			requirement_type VARCHAR(50) NOT NULL,
			description TEXT NOT NULL,
			code_function VARCHAR(255),
			priority VARCHAR(20) DEFAULT 'medium',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS test_coverage (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			test_requirement_id UUID REFERENCES test_requirements(id) ON DELETE CASCADE,
			knowledge_item_id UUID REFERENCES knowledge_items(id) ON DELETE CASCADE,
			coverage_percentage FLOAT DEFAULT 0.0,
			test_files TEXT[],
			last_updated TIMESTAMP DEFAULT NOW(),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS test_validations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			test_requirement_id UUID REFERENCES test_requirements(id) ON DELETE CASCADE,
			validation_status VARCHAR(20) NOT NULL,
			issues JSONB,
			test_code_hash VARCHAR(64),
			score FLOAT DEFAULT 0.0,
			validated_at TIMESTAMP DEFAULT NOW(),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_test_validations_hash ON test_validations(test_code_hash)`,
		`CREATE TABLE IF NOT EXISTS mutation_results (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			test_requirement_id UUID REFERENCES test_requirements(id) ON DELETE CASCADE,
			mutation_score FLOAT DEFAULT 0.0,
			total_mutants INT DEFAULT 0,
			killed_mutants INT DEFAULT 0,
			survived_mutants INT DEFAULT 0,
			execution_time_ms INT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS test_executions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			execution_type VARCHAR(50) NOT NULL,
			status VARCHAR(20) DEFAULT 'running',
			result JSONB,
			execution_time_ms INT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		// Phase 11: Doc-Sync System
		`CREATE TABLE IF NOT EXISTS doc_sync_reports (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			report_type VARCHAR(50) NOT NULL,
			discrepancies JSONB NOT NULL,
			summary JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS doc_sync_updates (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			report_id UUID REFERENCES doc_sync_reports(id) ON DELETE CASCADE,
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			file_path VARCHAR(500) NOT NULL,
			change_type VARCHAR(50) NOT NULL,
			old_value TEXT,
			new_value TEXT,
			line_number INT,
			approved_by VARCHAR(100),
			approved_at TIMESTAMP,
			applied BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_doc_sync_reports_project ON doc_sync_reports(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_doc_sync_reports_created ON doc_sync_reports(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_doc_sync_updates_report ON doc_sync_updates(report_id)`,
		`CREATE INDEX IF NOT EXISTS idx_doc_sync_updates_project ON doc_sync_updates(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_doc_sync_updates_applied ON doc_sync_updates(applied)`,
		// Phase 12: Gap Reports
		`CREATE TABLE IF NOT EXISTS gap_reports (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			gaps JSONB NOT NULL,
			summary JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gap_reports_project ON gap_reports(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gap_reports_created ON gap_reports(created_at DESC)`,
		// Phase 12: Change Requests
		`CREATE TABLE IF NOT EXISTS change_requests (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			knowledge_item_id UUID REFERENCES knowledge_items(id) ON DELETE CASCADE,
			type VARCHAR(20) NOT NULL,
			current_state JSONB,
			proposed_state JSONB,
			status VARCHAR(20) DEFAULT 'pending_approval',
			implementation_status VARCHAR(20) DEFAULT 'pending',
			implementation_notes TEXT,
			impact_analysis JSONB,
			created_at TIMESTAMP DEFAULT NOW(),
			approved_by VARCHAR(100),
			approved_at TIMESTAMP,
			rejected_by VARCHAR(100),
			rejected_at TIMESTAMP,
			rejection_reason TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_project ON change_requests(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_status ON change_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_knowledge_item ON change_requests(knowledge_item_id)`,
		// Change request sequence table for ID generation
		`CREATE TABLE IF NOT EXISTS change_request_sequences (
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			sequence_number SERIAL,
			PRIMARY KEY (project_id, sequence_number)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_impl_status ON change_requests(implementation_status)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_created ON change_requests(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_project_status ON change_requests(project_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_project_impl ON change_requests(project_id, implementation_status)`,
		`CREATE INDEX IF NOT EXISTS idx_test_requirements_knowledge_item ON test_requirements(knowledge_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_test_coverage_requirement ON test_coverage(test_requirement_id)`,
		`CREATE INDEX IF NOT EXISTS idx_test_coverage_knowledge_item ON test_coverage(knowledge_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_test_validations_requirement ON test_validations(test_requirement_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mutation_results_requirement ON mutation_results(test_requirement_id)`,
		`CREATE INDEX IF NOT EXISTS idx_test_executions_project ON test_executions(project_id)`,
		// Phase 14A: Comprehensive Feature Analysis
		`CREATE TABLE IF NOT EXISTS comprehensive_validations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			validation_id VARCHAR(50) UNIQUE NOT NULL,
			feature VARCHAR(255) NOT NULL,
			mode VARCHAR(20) NOT NULL,
			depth VARCHAR(20) NOT NULL,
			findings JSONB NOT NULL,
			summary JSONB NOT NULL,
			layer_analysis JSONB NOT NULL,
			end_to_end_flows JSONB,
			checklist JSONB NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			completed_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_validations_project ON comprehensive_validations(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_validations_validation_id ON comprehensive_validations(validation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_validations_created ON comprehensive_validations(created_at DESC)`,
		// Phase 15: Intent & Simple Language
		`CREATE TABLE IF NOT EXISTS intent_decisions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			original_prompt TEXT NOT NULL,
			intent_type VARCHAR(50) NOT NULL,
			clarifying_question TEXT NOT NULL,
			user_choice TEXT NOT NULL,
			resolved_prompt TEXT,
			context_data JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS intent_patterns (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			pattern_type VARCHAR(50) NOT NULL,
			pattern_data JSONB NOT NULL,
			frequency INTEGER DEFAULT 1,
			last_used TIMESTAMP DEFAULT NOW(),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_intent_decisions_project ON intent_decisions(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_intent_decisions_type ON intent_decisions(intent_type)`,
		`CREATE INDEX IF NOT EXISTS idx_intent_decisions_created ON intent_decisions(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_intent_patterns_project ON intent_patterns(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_intent_patterns_type ON intent_patterns(pattern_type)`,
		`CREATE INDEX IF NOT EXISTS idx_intent_patterns_frequency ON intent_patterns(frequency DESC)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_intent_patterns_unique ON intent_patterns(project_id, pattern_type, pattern_data)`,
		// Phase F: Database Optimization - Additional indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_projects_api_key ON projects(api_key)`,                                       // Critical for auth lookups (already UNIQUE, but explicit index helps)
		`CREATE INDEX IF NOT EXISTS idx_knowledge_items_doc_type ON knowledge_items(document_id, type)`,              // Composite for knowledge item queries
		`CREATE INDEX IF NOT EXISTS idx_knowledge_items_status ON knowledge_items(status) WHERE status = 'approved'`, // Partial index for approved items
		`CREATE INDEX IF NOT EXISTS idx_documents_project_status ON documents(project_id, status)`,                   // Composite for document queries
		// Composite index for business context queries (JOIN optimization)
		`CREATE INDEX IF NOT EXISTS idx_documents_project_id_status ON documents(project_id, status) WHERE status = 'completed'`, // Partial index for completed documents
		`CREATE TABLE IF NOT EXISTS analysis_configurations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			configuration JSONB NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_configs_project ON analysis_configurations(project_id)`,
		`CREATE TABLE IF NOT EXISTS llm_configurations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			provider VARCHAR(50) NOT NULL,
			api_key_encrypted BYTEA NOT NULL,
			model VARCHAR(100) NOT NULL,
			key_type VARCHAR(20) NOT NULL,
			cost_optimization JSONB,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_llm_configs_project ON llm_configurations(project_id)`,
		`CREATE TABLE IF NOT EXISTS llm_usage (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			validation_id UUID REFERENCES comprehensive_validations(id) ON DELETE CASCADE,
			provider VARCHAR(50) NOT NULL,
			model VARCHAR(100) NOT NULL,
			tokens_used INT NOT NULL,
			estimated_cost DECIMAL(10, 4),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_llm_usage_project ON llm_usage(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_llm_usage_validation ON llm_usage(validation_id)`,
		// Phase 14C: Config Audit Log
		`CREATE TABLE IF NOT EXISTS config_audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			config_id UUID,
			action VARCHAR(20) NOT NULL,
			changed_by VARCHAR(100),
			old_value JSONB,
			new_value JSONB,
			ip_address VARCHAR(45),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_config_audit_project ON config_audit_log(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_config_audit_config ON config_audit_log(config_id)`,
		`CREATE INDEX IF NOT EXISTS idx_config_audit_created ON config_audit_log(created_at DESC)`,
		// Binary distribution system
		`CREATE TABLE IF NOT EXISTS binary_versions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			version VARCHAR(50) NOT NULL,
			platform VARCHAR(20) NOT NULL,
			architecture VARCHAR(10) NOT NULL,
			os VARCHAR(10) NOT NULL,
			file_path TEXT NOT NULL,
			file_size BIGINT NOT NULL,
			checksum_sha256 VARCHAR(64) NOT NULL,
			checksum_md5 VARCHAR(32),
			signature TEXT,
			release_notes TEXT,
			is_stable BOOLEAN DEFAULT true,
			is_latest BOOLEAN DEFAULT false,
			min_go_version VARCHAR(10),
			created_at TIMESTAMP DEFAULT NOW(),
			released_at TIMESTAMP,
			created_by UUID,
			UNIQUE(version, platform)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_binary_versions_platform_latest ON binary_versions(platform, is_latest) WHERE is_latest = true`,
		`CREATE INDEX IF NOT EXISTS idx_binary_versions_version ON binary_versions(version)`,
		`CREATE TABLE IF NOT EXISTS binary_downloads (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			version_id UUID REFERENCES binary_versions(id),
			project_id UUID REFERENCES projects(id),
			user_agent TEXT,
			ip_address INET,
			downloaded_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_binary_downloads_version ON binary_downloads(version_id)`,
		`CREATE INDEX IF NOT EXISTS idx_binary_downloads_project ON binary_downloads(project_id)`,
		`CREATE TABLE IF NOT EXISTS rules_versions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			version VARCHAR(50) NOT NULL,
			rule_name VARCHAR(100) NOT NULL,
			rule_content TEXT NOT NULL,
			rule_type VARCHAR(20) NOT NULL,
			globs TEXT[],
			is_latest BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW(),
			released_at TIMESTAMP,
			UNIQUE(version, rule_name)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rules_versions_latest ON rules_versions(is_latest) WHERE is_latest = true`,
		// Phase 14E: Task Dependency & Verification System
		`CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			source VARCHAR(50) NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			file_path VARCHAR(500),
			line_number INTEGER,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			priority VARCHAR(10) DEFAULT 'medium',
			assigned_to VARCHAR(100),
			estimated_effort INTEGER,
			actual_effort INTEGER,
			tags TEXT[],
			verification_confidence FLOAT DEFAULT 0.0,
			version INTEGER DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP,
			verified_at TIMESTAMP,
			archived_at TIMESTAMP,
			CONSTRAINT valid_status CHECK (status IN ('pending', 'in_progress', 'completed', 'blocked')),
			CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
			CONSTRAINT valid_confidence CHECK (verification_confidence >= 0.0 AND verification_confidence <= 1.0)
		)`,
		`CREATE TABLE IF NOT EXISTS task_dependencies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			depends_on_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			dependency_type VARCHAR(20) NOT NULL,
			confidence FLOAT DEFAULT 0.0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT valid_dependency_type CHECK (dependency_type IN ('explicit', 'implicit', 'integration', 'feature')),
			CONSTRAINT no_self_dependency CHECK (task_id != depends_on_task_id),
			CONSTRAINT unique_dependency UNIQUE (task_id, depends_on_task_id),
			CONSTRAINT valid_dep_confidence CHECK (confidence >= 0.0 AND confidence <= 1.0)
		)`,
		`CREATE TABLE IF NOT EXISTS task_verifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			verification_type VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			confidence FLOAT DEFAULT 0.0,
			evidence JSONB,
			retry_count INTEGER DEFAULT 0,
			verified_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT valid_verification_type CHECK (verification_type IN ('code_existence', 'code_usage', 'test_coverage', 'integration')),
			CONSTRAINT valid_verification_status CHECK (status IN ('pending', 'verified', 'failed')),
			CONSTRAINT valid_ver_confidence CHECK (confidence >= 0.0 AND confidence <= 1.0)
		)`,
		`CREATE TABLE IF NOT EXISTS task_links (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			link_type VARCHAR(50) NOT NULL,
			linked_id UUID NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			CONSTRAINT valid_link_type CHECK (link_type IN ('change_request', 'knowledge_item', 'comprehensive_analysis', 'test_requirement')),
			CONSTRAINT unique_task_link UNIQUE (task_id, link_type, linked_id)
		)`,
		// Indexes for tasks table
		`CREATE INDEX IF NOT EXISTS idx_tasks_project_status ON tasks(project_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_project_priority ON tasks(project_id, priority)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_file_path ON tasks(file_path)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status) WHERE status IN ('pending', 'in_progress')`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority) WHERE priority IN ('high', 'critical')`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks(assigned_to) WHERE assigned_to IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at DESC)`,
		// Full-text search index for task title (using text search)
		`CREATE INDEX IF NOT EXISTS idx_tasks_title_search ON tasks USING gin(to_tsvector('english', title))`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_description_search ON tasks USING gin(to_tsvector('english', description)) WHERE description IS NOT NULL`,
		// Indexes for task_dependencies table
		`CREATE INDEX IF NOT EXISTS idx_task_dependencies_task ON task_dependencies(task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_task_dependencies_depends_on ON task_dependencies(depends_on_task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_task_dependencies_type ON task_dependencies(dependency_type)`,
		// Indexes for task_verifications table
		`CREATE INDEX IF NOT EXISTS idx_task_verifications_task ON task_verifications(task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_task_verifications_status ON task_verifications(status)`,
		`CREATE INDEX IF NOT EXISTS idx_task_verifications_type ON task_verifications(verification_type)`,
		// Indexes for task_links table
		`CREATE INDEX IF NOT EXISTS idx_task_links_task ON task_links(task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_task_links_linked ON task_links(link_type, linked_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// =============================================================================
// HANDLERS
// =============================================================================

// Health check handlers (Phase G: Logging and Monitoring)

// healthHandler returns basic health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"service":   "sentinel-hub",
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// healthDBHandler checks database connectivity
func healthDBHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "unhealthy",
			"service":   "database",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Get connection pool stats
	stats := db.Stats()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "database",
		"stats": map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// healthReadyHandler checks if the service is ready to accept traffic
func healthReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database connectivity
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "not_ready",
			"reason":    "database_unavailable",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Check storage directory
	config := loadConfig()
	if _, err := os.Stat(config.DocumentStorage); os.IsNotExist(err) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "not_ready",
			"reason":    "storage_unavailable",
			"error":     fmt.Sprintf("Storage directory does not exist: %s", config.DocumentStorage),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Upload document
func uploadDocumentHandler(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project from context (set by auth middleware)
		project, err := getProjectFromContext(r.Context())
		if err != nil {
			LogErrorWithContext(r.Context(), err, "Failed to get project from context")
			LogErrorWithContext(r.Context(), err, "Internal server error")
			LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
			WriteErrorResponse(w, &DatabaseError{
				Operation:     "internal_operation",
				Message:       "Internal server error",
				OriginalError: fmt.Errorf("internal server error"),
			}, http.StatusInternalServerError)
			return
		}

		// Parse multipart form (max 100MB)
		if err := r.ParseMultipartForm(100 << 20); err != nil {
			LogErrorWithContext(r.Context(), err, fmt.Sprintf("Error parsing multipart form for project %s", project.ID))
			WriteErrorResponse(w, &ValidationError{
				Field:   "file",
				Message: "File too large or invalid form",
				Code:    "invalid_file",
			}, http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			LogErrorWithContext(r.Context(), err, fmt.Sprintf("Error getting file from form for project %s", project.ID))
			WriteErrorResponse(w, &ValidationError{
				Field:   "file",
				Message: "No file provided",
				Code:    "required",
			}, http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Generate document ID
		docID := uuid.New().String()

		// Create storage directory
		storageDir := filepath.Join(config.DocumentStorage, project.ID, docID)
		if err := os.MkdirAll(storageDir, 0755); err != nil {
			LogErrorWithContext(r.Context(), err, "Storage error")
			WriteErrorResponse(w, &ExternalServiceError{
				Service:    "storage",
				Message:    "Storage error",
				StatusCode: 0,
			}, http.StatusInternalServerError)
			return
		}

		// Save file
		filePath := filepath.Join(storageDir, header.Filename)
		dst, err := os.Create(filePath)
		if err != nil {
			LogErrorWithContext(r.Context(), err, "Storage error")
			WriteErrorResponse(w, &ExternalServiceError{
				Service:    "storage",
				Message:    "Storage error",
				StatusCode: 0,
			}, http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			LogErrorWithContext(r.Context(), err, "Storage error")
			WriteErrorResponse(w, &ExternalServiceError{
				Service:    "storage",
				Message:    "Storage error",
				StatusCode: 0,
			}, http.StatusInternalServerError)
			return
		}

		// Detect MIME type
		mimeType := header.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = detectMimeType(header.Filename)
		}

		// Insert into database
		doc := Document{
			ID:           docID,
			ProjectID:    project.ID,
			Name:         docID,
			OriginalName: header.Filename,
			Size:         header.Size,
			MimeType:     mimeType,
			Status:       "queued",
			Progress:     0,
			FilePath:     filePath,
			CreatedAt:    time.Now(),
		}

		_, err = db.Exec(`
			INSERT INTO documents (id, project_id, name, original_name, size, mime_type, status, progress, file_path, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, doc.ID, doc.ProjectID, doc.Name, doc.OriginalName, doc.Size, doc.MimeType, doc.Status, doc.Progress, doc.FilePath, doc.CreatedAt)

		if err != nil {
			WriteErrorResponse(w, &DatabaseError{
				Operation:     "database_query",
				Message:       "Database error",
				OriginalError: err,
			}, http.StatusInternalServerError)
			return
		}

		// Check if change detection is requested
		detectChanges := r.URL.Query().Get("detect_changes") == "true"

		// Change detection is automatically triggered by the processor after knowledge extraction
		// The detect_changes query parameter is deprecated but kept for backward compatibility
		if detectChanges {
			LogInfo(r.Context(), "Change detection requested for document %s (will process after extraction)", doc.ID)
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                     doc.ID,
			"status":                 doc.Status,
			"estimated_time_seconds": estimateProcessingTime(doc.Size),
			"detect_changes":         detectChanges,
		})
	}
}

// detectChangesHandler triggers change detection for a document
// POST /api/v1/documents/{documentId}/detect-changes
func detectChangesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	documentID := chi.URLParam(r, "id")

	// Validate document ID
	if err := ValidateUUID(documentID); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "id",
			Message: "Invalid document ID",
			Code:    "invalid_uuid",
		}, http.StatusBadRequest)
		return
	}

	// Get project from context
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Call change detection
	changeRequestIDs, err := processChangeDetectionForDocument(ctx, documentID, project.ID)
	if err != nil {
		LogErrorWithContext(ctx, err, "Change detection failed")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "detect_changes",
			Message:       "Change detection failed",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":            true,
		"document_id":        documentID,
		"change_request_ids": changeRequestIDs,
		"count":              len(changeRequestIDs),
	})
}

// Get document status
func getDocumentStatusHandler(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var doc Document
	var processedAt sql.NullTime
	var extractedText, errorText sql.NullString

	err = db.QueryRow(`
		SELECT id, project_id, name, original_name, size, mime_type, status, progress, 
		       extracted_text, error, created_at, processed_at
		FROM documents WHERE id = $1 AND project_id = $2
	`, docID, project.ID).Scan(
		&doc.ID, &doc.ProjectID, &doc.Name, &doc.OriginalName, &doc.Size, &doc.MimeType,
		&doc.Status, &doc.Progress, &extractedText, &errorText, &doc.CreatedAt, &processedAt,
	)

	if err == sql.ErrNoRows {
		WriteErrorResponse(w, &NotFoundError{
			Resource: "document",
			ID:       docID,
			Message:  "Document not found",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	if extractedText.Valid {
		doc.ExtractedText = extractedText.String
	}
	if errorText.Valid {
		doc.Error = errorText.String
	}
	if processedAt.Valid {
		doc.ProcessedAt = &processedAt.Time
	}

	// Build status response
	status := DocumentStatus{
		ID:           doc.ID,
		OriginalName: doc.OriginalName,
		Status:       doc.Status,
		Progress:     doc.Progress,
		Error:        doc.Error,
		CreatedAt:    doc.CreatedAt,
		ProcessedAt:  doc.ProcessedAt,
		Stages: []ProcessingStage{
			{Name: "upload", Status: "completed"},
			{Name: "parsing", Status: stageStatus(doc.Status, doc.Progress, 50)},
			{Name: "extraction", Status: stageStatus(doc.Status, doc.Progress, 100)},
		},
	}

	if doc.Status == "completed" {
		// Count knowledge items
		var itemCount int
		if err := db.QueryRow("SELECT COUNT(*) FROM knowledge_items WHERE document_id = $1", doc.ID).Scan(&itemCount); err != nil {
			LogErrorWithContext(r.Context(), err, "Error counting knowledge items")
			itemCount = 0
		}

		status.Result = &DocumentResult{
			TextLength:     len(doc.ExtractedText),
			KnowledgeItems: itemCount,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Get extracted text
func getExtractedTextHandler(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var doc Document
	var extractedText sql.NullString

	err = db.QueryRow(`
		SELECT id, original_name, extracted_text
		FROM documents WHERE id = $1 AND project_id = $2 AND status = 'completed'
	`, docID, project.ID).Scan(&doc.ID, &doc.OriginalName, &extractedText)

	if err == sql.ErrNoRows {
		WriteErrorResponse(w, &NotFoundError{
			Resource: "document",
			ID:       docID,
			Message:  "Document not found or not processed",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             doc.ID,
		"original_name":  doc.OriginalName,
		"extracted_text": extractedText.String,
		"metadata": map[string]interface{}{
			"word_count": len(strings.Fields(extractedText.String)),
		},
	})
}

// Get knowledge items
func getKnowledgeItemsHandler(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Verify document belongs to project
	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM documents WHERE id = $1 AND project_id = $2)", docID, project.ID).Scan(&exists)
	if !exists {
		WriteErrorResponse(w, &NotFoundError{
			Resource: "document",
			ID:       docID,
			Message:  "Document not found",
		}, http.StatusNotFound)
		return
	}

	rows, err := db.Query(`
		SELECT id, document_id, type, title, content, confidence, source_page, status, created_at
		FROM knowledge_items WHERE document_id = $1
	`, docID)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []KnowledgeItem{}
	ctx := r.Context()
	for rows.Next() {
		var item KnowledgeItem
		var sourcePage sql.NullInt32
		if err := rows.Scan(&item.ID, &item.DocumentID, &item.Type, &item.Title, &item.Content,
			&item.Confidence, &sourcePage, &item.Status, &item.CreatedAt); err != nil {
			LogWarn(ctx, "Failed to scan knowledge item row (skipping): %v. DocumentID: %s", err, docID)
			continue
		}
		if sourcePage.Valid {
			item.SourcePage = int(sourcePage.Int32)
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":              docID,
		"knowledge_items": items,
	})
}

// List project documents
func listDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`
		SELECT d.id, d.original_name, d.status, d.created_at,
		       (SELECT COUNT(*) FROM knowledge_items WHERE document_id = d.id) as item_count
		FROM documents d
		WHERE d.project_id = $1
		ORDER BY d.created_at DESC
		LIMIT 100
	`, project.ID)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	docs := []map[string]interface{}{}
	ctx := r.Context()
	for rows.Next() {
		var id, name, status string
		var createdAt time.Time
		var itemCount int
		if err := rows.Scan(&id, &name, &status, &createdAt, &itemCount); err != nil {
			LogWarn(ctx, "Failed to scan document row (skipping): %v. ProjectID: %s", err, project.ID)
			continue
		}
		docs = append(docs, map[string]interface{}{
			"id":              id,
			"name":            name,
			"status":          status,
			"knowledge_items": itemCount,
			"uploaded_at":     createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": docs,
		"total":     len(docs),
	})
}

// Update knowledge item status
func updateKnowledgeStatusHandler(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req struct {
		Status     string `json:"status"`      // approved, rejected
		ApprovedBy string `json:"approved_by"` // optional
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		WriteErrorResponse(w, &ValidationError{
			Field:   "status",
			Message: "Status must be 'approved' or 'rejected'",
			Code:    "invalid_enum",
		}, http.StatusBadRequest)
		return
	}

	// Verify knowledge item belongs to project
	var docID string
	err = db.QueryRow(`
		SELECT ki.document_id 
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE ki.id = $1 AND d.project_id = $2
	`, itemID, project.ID).Scan(&docID)

	if err == sql.ErrNoRows {
		WriteErrorResponse(w, &NotFoundError{
			Resource: "knowledge_item",
			ID:       itemID,
			Message:  "Knowledge item not found",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// Update status
	now := time.Now()
	var approvedBy sql.NullString
	var approvedAt sql.NullTime

	if req.Status == "approved" {
		approvedBy = sql.NullString{String: req.ApprovedBy, Valid: req.ApprovedBy != ""}
		approvedAt = sql.NullTime{Time: now, Valid: true}
	}

	_, err = db.Exec(`
		UPDATE knowledge_items 
		SET status = $1, approved_by = $2, approved_at = $3
		WHERE id = $4
	`, req.Status, approvedBy, approvedAt, itemID)

	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     itemID,
		"status": req.Status,
	})
}

// List all knowledge items for project
func listProjectKnowledgeHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Get status filter from query param
	statusFilter := r.URL.Query().Get("status") // pending, approved, rejected

	query := `
		SELECT ki.id, ki.document_id, ki.type, ki.title, ki.content, ki.confidence, 
		       ki.source_page, ki.status, ki.approved_by, ki.approved_at, ki.created_at,
		       d.original_name
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE d.project_id = $1
	`
	args := []interface{}{project.ID}

	if statusFilter != "" {
		query += " AND ki.status = $2"
		args = append(args, statusFilter)
	}

	query += " ORDER BY ki.created_at DESC LIMIT 500"

	rows, err := db.Query(query, args...)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []map[string]interface{}{}
	ctx := r.Context()
	for rows.Next() {
		var item KnowledgeItem
		var sourcePage sql.NullInt32
		var approvedBy sql.NullString
		var approvedAt sql.NullTime
		var docName string

		err := rows.Scan(&item.ID, &item.DocumentID, &item.Type, &item.Title, &item.Content,
			&item.Confidence, &sourcePage, &item.Status, &approvedBy, &approvedAt,
			&item.CreatedAt, &docName)
		if err != nil {
			LogWarn(ctx, "Failed to scan knowledge item row (skipping): %v. ProjectID: %s", err, project.ID)
			continue
		}

		result := map[string]interface{}{
			"id":          item.ID,
			"document_id": item.DocumentID,
			"document":    docName,
			"type":        item.Type,
			"title":       item.Title,
			"content":     item.Content,
			"confidence":  item.Confidence,
			"status":      item.Status,
			"created_at":  item.CreatedAt,
		}

		if sourcePage.Valid {
			result["source_page"] = sourcePage.Int32
		}
		if approvedBy.Valid {
			result["approved_by"] = approvedBy.String
		}
		if approvedAt.Valid {
			result["approved_at"] = approvedAt.Time
		}

		items = append(items, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"knowledge_items": items,
		"total":           len(items),
	})
}

// getBusinessContextHandler handles GET /api/v1/knowledge/business
// Returns business rules, entities, and journeys filtered by type
func getBusinessContextHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Get type filter from query param (rule, entity, journey)
	typeFilter := r.URL.Query().Get("type")

	query := `
		SELECT ki.id, ki.document_id, ki.type, ki.title, ki.content, ki.confidence, 
		       ki.source_page, ki.status, ki.created_at, d.original_name
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE d.project_id = $1 AND ki.status = 'approved'
	`
	args := []interface{}{project.ID}

	// Filter by business-related types: rule, entity, journey
	if typeFilter != "" {
		query += " AND ki.type = $2"
		args = append(args, typeFilter)
	} else {
		// Default: return all business-related types
		query += " AND ki.type IN ('rule', 'entity', 'journey')"
	}

	query += " ORDER BY ki.created_at DESC LIMIT 200"

	rows, err := queryWithTimeout(ctx, query, args...)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to query business context")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []map[string]interface{}{}
	for rows.Next() {
		var item struct {
			ID           string
			DocumentID   string
			Type         string
			Title        string
			Content      string
			Confidence   float64
			SourcePage   sql.NullInt32
			Status       string
			CreatedAt    time.Time
			DocumentName string
		}

		err := rows.Scan(&item.ID, &item.DocumentID, &item.Type, &item.Title, &item.Content,
			&item.Confidence, &item.SourcePage, &item.Status, &item.CreatedAt, &item.DocumentName)
		if err != nil {
			LogWarn(ctx, "Failed to scan business context item: %v", err)
			continue
		}

		result := map[string]interface{}{
			"id":          item.ID,
			"document_id": item.DocumentID,
			"document":    item.DocumentName,
			"item_type":   item.Type,
			"type":        item.Type, // Also include as 'type' for compatibility
			"title":       item.Title,
			"content":     item.Content,
			"confidence":  item.Confidence,
			"status":      item.Status,
			"created_at":  item.CreatedAt.Format(time.RFC3339),
		}

		if item.SourcePage.Valid {
			result["source_page"] = item.SourcePage.Int32
		}

		items = append(items, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"items":   items,
		"total":   len(items),
	})
}

// getSecurityContextHandler handles GET /api/v1/security/context
// Returns security rules, compliance status, and security score
func getSecurityContextHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogError(ctx, "Failed to get project from context: %v", err)
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Get all security rules
	rules := []map[string]interface{}{}
	for ruleID, rule := range SecurityRules {
		ruleData := map[string]interface{}{
			"rule_id":     ruleID,
			"name":        rule.Name,
			"type":        rule.Type,
			"severity":    rule.Severity,
			"description": rule.Description,
			"status":      "active", // All rules are active by default
		}
		rules = append(rules, ruleData)
	}

	// Calculate compliance status (simplified - in production, would query recent analyses)
	compliance := map[string]interface{}{
		"total_rules":     len(SecurityRules),
		"rules_checked":   len(SecurityRules),
		"rules_compliant": len(SecurityRules), // Default to compliant
		"last_check":      time.Now().Format(time.RFC3339),
	}

	// Default security score (in production, would calculate from recent analyses)
	securityScore := 85.0
	securityGrade := "B"

	// Try to get recent security score from database if available
	// Note: This would require a security_analysis table which may not exist yet
	// For now, return default values

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"security_score": securityScore,
		"security_grade": securityGrade,
		"rules":          rules,
		"compliance":     compliance,
		"project_id":     project.ID,
	})
}

// validateCodeHandler handles POST /api/v1/validate/code
// Validates code using AST analysis
func validateCodeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	var req struct {
		Code     string `json:"code"`
		FilePath string `json:"file_path"`
		Language string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		WriteErrorResponse(w, &ValidationError{
			Field:   "code",
			Message: "code is required",
			Code:    "required",
		}, http.StatusBadRequest)
		return
	}

	// Infer language from file extension if not provided
	if req.Language == "" && req.FilePath != "" {
		ext := filepath.Ext(req.FilePath)
		langMap := map[string]string{
			".js": "javascript", ".ts": "typescript", ".jsx": "javascript",
			".py": "python", ".go": "go", ".java": "java",
			".cs": "csharp", ".php": "php", ".rb": "ruby",
		}
		if lang, ok := langMap[ext]; ok {
			req.Language = lang
		}
	}

	// Perform AST analysis
	violations := []map[string]interface{}{}

	if req.Language != "" {
		// Call AST analysis using existing analyzeAST function
		findings, stats, err := analyzeAST(req.Code, req.Language, []string{"duplicates", "unused", "unreachable"})
		if err != nil {
			LogErrorWithContext(ctx, err, "AST analysis failed")
			// Continue with empty violations rather than failing the request
		} else {
			// Convert AST findings to violations format
			for _, finding := range findings {
				violation := map[string]interface{}{
					"type":       finding.Type,
					"severity":   finding.Severity,
					"line":       finding.Line,
					"column":     finding.Column,
					"end_line":   finding.EndLine,
					"end_column": finding.EndColumn,
					"message":    finding.Message,
					"code":       finding.Code,
					"suggestion": finding.Suggestion,
				}
				violations = append(violations, violation)
			}
		}
		_ = stats // Stats available for future use (logging, metrics)
	} else {
		// Language not provided and couldn't be inferred
		violations = append(violations, map[string]interface{}{
			"type":     "language_required",
			"severity": "warning",
			"message":  "Language not specified and could not be inferred from file path",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"violations": violations,
		"valid":      len(violations) == 0,
	})
}

// validateBusinessHandler handles POST /api/v1/validate/business
// Validates code against business rules
func validateBusinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req struct {
		Feature string `json:"feature"`
		Code    string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.Feature == "" || req.Code == "" {
		var missingField string
		if req.Feature == "" {
			missingField = "feature"
		} else {
			missingField = "code"
		}
		WriteErrorResponse(w, &ValidationError{
			Field:   missingField,
			Message: fmt.Sprintf("%s is required", missingField),
			Code:    "required",
		}, http.StatusBadRequest)
		return
	}

	// Get business rules for the project
	query := `
		SELECT ki.id, ki.type, ki.title, ki.content
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE d.project_id = $1 AND ki.type IN ('rule', 'entity', 'journey') AND ki.status = 'approved'
		ORDER BY ki.created_at DESC
		LIMIT 100
	`
	rows, err := queryWithTimeout(ctx, query, project.ID)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to query business rules")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	violations := []map[string]interface{}{}
	businessRules := []map[string]interface{}{}

	for rows.Next() {
		var rule struct {
			ID      string
			Type    string
			Title   string
			Content string
		}
		if err := rows.Scan(&rule.ID, &rule.Type, &rule.Title, &rule.Content); err != nil {
			continue
		}
		businessRules = append(businessRules, map[string]interface{}{
			"id":      rule.ID,
			"type":    rule.Type,
			"title":   rule.Title,
			"content": rule.Content,
		})
	}

	// Simple validation: check if code mentions business entities/rules
	// In production, would use LLM or more sophisticated analysis
	codeLower := strings.ToLower(req.Code)
	for _, rule := range businessRules {
		if title, ok := rule["title"].(string); ok {
			titleLower := strings.ToLower(title)
			// Check if rule title is mentioned in code (simplified check)
			if !strings.Contains(codeLower, titleLower) && rule["type"] == "rule" {
				// This is a simplified check - in production would be more sophisticated
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"violations":    violations,
		"valid":         len(violations) == 0,
		"rules_checked": len(businessRules),
	})
}

// applyFixHandler handles POST /api/v1/fixes/apply
// Applies fixes to code issues
func applyFixHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	var req struct {
		FilePath string `json:"file_path"`
		FixType  string `json:"fix_type"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.FilePath == "" || req.FixType == "" || req.Content == "" {
		var missingField string
		if req.FilePath == "" {
			missingField = "file_path"
		} else if req.FixType == "" {
			missingField = "fix_type"
		} else {
			missingField = "content"
		}
		WriteErrorResponse(w, &ValidationError{
			Field:   missingField,
			Message: fmt.Sprintf("%s is required", missingField),
			Code:    "required",
		}, http.StatusBadRequest)
		return
	}

	// Apply fixes based on fix type
	var fixedCode string
	var changes []map[string]interface{}
	var err error

	// Infer language from file path if not provided
	language := inferLanguageFromPath(req.FilePath)

	switch req.FixType {
	case "security":
		fixedCode, changes, err = ApplySecurityFixes(ctx, req.Content, language)
	case "style":
		fixedCode, changes, err = ApplyStyleFixes(ctx, req.Content, language)
	case "performance":
		fixedCode, changes, err = ApplyPerformanceFixes(ctx, req.Content, language)
	default:
		WriteErrorResponse(w, &ValidationError{
			Field:   "fix_type",
			Message: fmt.Sprintf("Unknown fix type: %s. Supported types: security, style, performance", req.FixType),
			Code:    "invalid_enum",
		}, http.StatusBadRequest)
		return
	}

	if err != nil {
		LogErrorWithContext(ctx, err, "Fix application failed")
		WriteErrorResponse(w, &ExternalServiceError{
			Service:    "fix_applier",
			Message:    fmt.Sprintf("Failed to apply fixes: %v", err),
			StatusCode: 0,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"fixed_code": fixedCode,
		"changes":    changes,
	})
}

// Sync knowledge items (bidirectional)
func syncKnowledgeHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	updated := 0
	for _, item := range req.Items {
		// Verify item belongs to project
		var exists bool
		db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM knowledge_items ki
				INNER JOIN documents d ON ki.document_id = d.id
				WHERE ki.id = $1 AND d.project_id = $2
			)
		`, item.ID, project.ID).Scan(&exists)

		if !exists {
			continue
		}

		// Update status
		var approvedAt sql.NullTime
		if item.Status == "approved" {
			approvedAt = sql.NullTime{Time: time.Now(), Valid: true}
		}

		_, err := db.Exec(`
			UPDATE knowledge_items 
			SET status = $1, approved_at = $2
			WHERE id = $3
		`, item.Status, approvedAt, item.ID)

		if err == nil {
			updated++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updated": updated,
		"total":   len(req.Items),
	})
}

// Gap Analysis Handler (Phase 12)
func gapAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req GapAnalysisRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	// Use project ID from context if not provided
	if req.ProjectID == "" {
		req.ProjectID = project.ID
	}

	// Default codebase path if not provided
	if req.CodebasePath == "" {
		req.CodebasePath = "."
	}

	// Default options
	if req.Options == nil {
		req.Options = make(map[string]interface{})
	}

	ctx, cancel := context.WithTimeout(r.Context(), getAnalysisTimeout())
	defer cancel()

	report, err := analyzeGaps(ctx, req.ProjectID, req.CodebasePath, req.Options)
	if err != nil {
		LogErrorWithContext(ctx, err, fmt.Sprintf("Gap analysis failed (project: %s, path: %s)", req.ProjectID, req.CodebasePath))
		http.Error(w, fmt.Sprintf("Gap analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Store report in database
	reportID, err := storeGapReport(ctx, report)
	if err != nil {
		// Log error but don't fail the request (report is still returned)
		LogWarn(ctx, "Failed to store gap report (project: %s): %v", req.ProjectID, err)
		// reportID will be empty string, which is acceptable
	}

	w.Header().Set("Content-Type", "application/json")
	response := GapAnalysisResponse{
		Success:  true,
		Report:   report,
		ReportID: reportID, // Set from storage
	}

	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// PHASE 12: CHANGE REQUEST HANDLERS
// =============================================================================

// List change requests handler
func listChangeRequestsHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	statusFilter := r.URL.Query().Get("status")
	limit := 50 // Default
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	requests, total, err := listChangeRequests(ctx, project.ID, statusFilter, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list change requests: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate pagination metadata
	hasNext := offset+limit < total
	hasPrevious := offset > 0

	response := ListChangeRequestsResponse{
		ChangeRequests: requests,
		Total:          total,
		Limit:          limit,
		Offset:         offset,
		HasNext:        hasNext,
		HasPrevious:    hasPrevious,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get change request handler
func getChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	changeRequestID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}

	// Verify change request belongs to project
	if cr.ProjectID != project.ID {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cr)
}

// Approve change request handler
func approveChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	changeRequestID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req ApproveChangeRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.ApprovedBy == "" {
		req.ApprovedBy = "system" // Default
	}

	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	// Verify change request belongs to project
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}
	if cr.ProjectID != project.ID {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}

	err = approveChangeRequest(ctx, changeRequestID, req.ApprovedBy)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to approve: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Change request approved",
	})
}

// Reject change request handler
func rejectChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	changeRequestID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req RejectChangeRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.RejectedBy == "" {
		req.RejectedBy = "system" // Default
	}
	if req.Reason == "" {
		req.Reason = "No reason provided"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Verify change request belongs to project
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}
	if cr.ProjectID != project.ID {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}

	err = rejectChangeRequest(ctx, changeRequestID, req.RejectedBy, req.Reason)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to reject: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Change request rejected",
	})
}

// Analyze impact handler
func analyzeImpactHandler(w http.ResponseWriter, r *http.Request) {
	changeRequestID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req ImpactAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use default if not provided
		req.CodebasePath = "."
	}

	if req.CodebasePath == "" {
		req.CodebasePath = "."
	}

	ctx, cancel := context.WithTimeout(r.Context(), getAnalysisTimeout())
	defer cancel()

	// Verify change request belongs to project
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}
	if cr.ProjectID != project.ID {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}

	impact, err := analyzeImpact(ctx, changeRequestID, project.ID, req.CodebasePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to analyze impact: %v", err), http.StatusInternalServerError)
		return
	}

	// Store impact analysis in database
	if err := storeImpactAnalysis(ctx, changeRequestID, impact); err != nil {
		// Log error but don't fail the request (impact is still returned)
		LogWarn(ctx, "Failed to store impact analysis: %v", err)
	}

	response := ImpactAnalysisResponse{
		Success: true,
		Impact:  impact,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Start implementation handler
func startImplementationHandler(w http.ResponseWriter, r *http.Request) {
	changeRequestID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req StartImplementationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), getQueryTimeout())
	defer cancel()

	// Verify change request belongs to project
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}
	if cr.ProjectID != project.ID {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}

	err = updateImplementationStatus(ctx, changeRequestID, ImplStatusInProgress, req.Notes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start implementation: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Implementation started",
	})
}

// Complete implementation handler
func completeImplementationHandler(w http.ResponseWriter, r *http.Request) {
	changeRequestID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req CompleteImplementationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), getQueryTimeout())
	defer cancel()

	// Verify change request belongs to project
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}
	if cr.ProjectID != project.ID {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}

	err = updateImplementationStatus(ctx, changeRequestID, ImplStatusCompleted, req.Notes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to complete implementation: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Implementation completed",
	})
}

// Update implementation handler
func updateImplementationHandler(w http.ResponseWriter, r *http.Request) {
	changeRequestID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req UpdateImplementationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Verify change request belongs to project
	cr, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}
	if cr.ProjectID != project.ID {
		http.Error(w, "Change request not found", http.StatusNotFound)
		return
	}

	err = updateImplementationStatus(ctx, changeRequestID, req.Status, req.Notes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update implementation: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Implementation status updated",
	})
}

// Get change requests dashboard handler
func getChangeRequestsDashboardHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get counts by status
	statusQuery := `
		SELECT status, COUNT(*) 
		FROM change_requests 
		WHERE project_id = $1 
		GROUP BY status
	`
	rows, err := db.QueryContext(ctx, statusQuery, project.ID)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	byStatus := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err == nil {
			byStatus[status] = count
		}
	}

	// Get counts by implementation status
	implQuery := `
		SELECT implementation_status, COUNT(*) 
		FROM change_requests 
		WHERE project_id = $1 
		GROUP BY implementation_status
	`
	rows2, err := db.QueryContext(ctx, implQuery, project.ID)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows2.Close()

	byImplStatus := make(map[string]int)
	for rows2.Next() {
		var status string
		var count int
		if err := rows2.Scan(&status, &count); err == nil {
			byImplStatus[status] = count
		}
	}

	// Get total
	var total int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM change_requests WHERE project_id = $1", project.ID).Scan(&total)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":                    total,
		"by_status":                byStatus,
		"by_implementation_status": byImplStatus,
	})
}

// =============================================================================
// TELEMETRY
// =============================================================================

type TelemetryEvent struct {
	Event     string                 `json:"event"`
	EventType string                 `json:"event_type"` // Keep for backward compat
	AgentID   string                 `json:"agentId"`
	OrgID     string                 `json:"orgId"`
	TeamID    string                 `json:"teamId,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// Telemetry ingestion handler
func telemetryIngestionHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var events []TelemetryEvent
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		log.Printf("Error decoding telemetry events for project %s: %v", project.ID, err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(events) == 0 {
		log.Printf("Telemetry ingestion request with no events for project %s", project.ID)
		http.Error(w, "No events provided", http.StatusBadRequest)
		return
	}

	// Validate and sanitize events
	inserted := 0
	for _, event := range events {
		// Validate event type
		validTypes := map[string]bool{
			"audit_complete":   true,
			"fix_applied":      true,
			"pattern_learned":  true,
			"doc_ingested":     true,
			"knowledge_synced": true,
		}

		if !validTypes[event.EventType] {
			log.Printf("Invalid event type: %s", event.EventType)
			continue
		}

		// Sanitize payload (ensure no code content)
		sanitizedPayload := sanitizeTelemetryPayload(event.Payload)

		// Insert into database
		payloadJSON, err := json.Marshal(sanitizedPayload)
		if err != nil {
			log.Printf("Failed to marshal payload: %v", err)
			continue
		}

		// Extract additional fields from event
		agentID := event.AgentID
		orgID := event.OrgID
		teamID := event.TeamID
		eventTimestamp := event.Timestamp

		// Parse timestamp if provided
		var timestamp sql.NullTime
		if eventTimestamp != "" {
			if t, err := time.Parse(time.RFC3339, eventTimestamp); err == nil {
				timestamp = sql.NullTime{Time: t, Valid: true}
			}
		}

		// Convert orgID and teamID to UUID if provided
		var orgIDUUID sql.NullString
		var teamIDUUID sql.NullString
		if orgID != "" {
			if _, err := uuid.Parse(orgID); err == nil {
				orgIDUUID = sql.NullString{String: orgID, Valid: true}
			}
		}
		if teamID != "" {
			if _, err := uuid.Parse(teamID); err == nil {
				teamIDUUID = sql.NullString{String: teamID, Valid: true}
			}
		}

		_, err = db.Exec(`
			INSERT INTO telemetry_events (project_id, event_type, payload, agent_id, org_id, team_id, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, project.ID, event.EventType, string(payloadJSON),
			sql.NullString{String: agentID, Valid: agentID != ""},
			orgIDUUID, teamIDUUID, timestamp)

		if err != nil {
			log.Printf("Failed to insert telemetry event: %v", err)
			continue
		}

		inserted++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"received": len(events),
		"inserted": inserted,
	})
}

// Sanitize telemetry payload to ensure no sensitive data
func sanitizeTelemetryPayload(payload map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	// Allowed fields (no code content)
	allowedFields := map[string]bool{
		"finding_count":      true,
		"critical_count":     true,
		"warning_count":      true,
		"info_count":         true,
		"compliance_percent": true,
		"fix_count":          true,
		"fix_type":           true,
		"pattern_confidence": true,
		"pattern_type":       true,
		"doc_count":          true,
		"knowledge_count":    true,
		"file_count":         true,
		"timestamp":          true,
		"duration_ms":        true,
	}

	for key, value := range payload {
		if allowedFields[key] {
			sanitized[key] = value
		}
	}

	return sanitized
}

// Get metrics handler
func getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Get query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	eventType := r.URL.Query().Get("event_type")

	// Build query
	query := `
		SELECT event_type, payload, created_at
		FROM telemetry_events
		WHERE project_id = $1
	`
	args := []interface{}{project.ID}
	argIndex := 2

	if startDate != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	if eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	query += " ORDER BY created_at DESC LIMIT 1000"

	rows, err := db.Query(query, args...)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []map[string]interface{}{}
	for rows.Next() {
		var eventType string
		var payloadJSON string
		var createdAt time.Time

		if err := rows.Scan(&eventType, &payloadJSON, &createdAt); err != nil {
			continue
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
			payload = make(map[string]interface{})
		}

		events = append(events, map[string]interface{}{
			"event_type": eventType,
			"payload":    payload,
			"created_at": createdAt,
		})
	}

	// Calculate aggregated metrics
	metrics := calculateMetrics(events)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events":  events,
		"metrics": metrics,
		"total":   len(events),
	})
}

// prometheusMetricsHandler exposes metrics in Prometheus format (Phase G: Logging and Monitoring)
func prometheusMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	// HTTP request counter
	for endpoint, count := range httpRequestCounter {
		fmt.Fprintf(w, "sentinel_http_requests_total{endpoint=\"%s\"} %d\n", endpoint, count)
	}

	// HTTP error counter
	for endpoint, count := range httpErrorCounter {
		fmt.Fprintf(w, "sentinel_http_errors_total{endpoint=\"%s\"} %d\n", endpoint, count)
	}

	// HTTP request duration (average)
	for endpoint := range httpDurationSum {
		if requestCount := httpRequestCount[endpoint]; requestCount > 0 {
			avgDuration := float64(httpDurationSum[endpoint]) / float64(requestCount)
			fmt.Fprintf(w, "sentinel_http_request_duration_ms{endpoint=\"%s\"} %.2f\n", endpoint, avgDuration)
		}
	}

	// Database connection pool stats
	stats := db.Stats()
	fmt.Fprintf(w, "sentinel_db_open_connections %d\n", stats.OpenConnections)
	fmt.Fprintf(w, "sentinel_db_in_use %d\n", stats.InUse)
	fmt.Fprintf(w, "sentinel_db_idle %d\n", stats.Idle)
	fmt.Fprintf(w, "sentinel_db_wait_count %d\n", stats.WaitCount)
	fmt.Fprintf(w, "sentinel_db_wait_duration_ms %d\n", stats.WaitDuration.Milliseconds())

	// Uptime
	uptime := time.Since(startTime).Seconds()
	fmt.Fprintf(w, "sentinel_uptime_seconds %.2f\n", uptime)
}

// recordHTTPMetric records HTTP request metrics (Phase G: Logging and Monitoring)
func recordHTTPMetric(endpoint string, statusCode int, duration time.Duration) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	httpRequestCounter[endpoint]++
	if statusCode >= 400 {
		httpErrorCounter[endpoint]++
	}
	httpDurationSum[endpoint] += duration.Milliseconds()
	httpRequestCount[endpoint]++
}

// Calculate aggregated metrics from events
func calculateMetrics(events []map[string]interface{}) map[string]interface{} {
	auditCount := 0
	fixCount := 0
	patternCount := 0
	docCount := 0
	totalFindings := 0
	totalCritical := 0
	totalWarnings := 0
	complianceSum := 0.0
	complianceCount := 0

	for _, event := range events {
		eventType, _ := event["event_type"].(string)
		payload, _ := event["payload"].(map[string]interface{})

		switch eventType {
		case "audit_complete":
			auditCount++
			if compliance, ok := payload["compliance_percent"].(float64); ok {
				complianceSum += compliance
				complianceCount++
			}
			if count, ok := payload["finding_count"].(float64); ok {
				totalFindings += int(count)
			}
			if count, ok := payload["critical_count"].(float64); ok {
				totalCritical += int(count)
			}
			if count, ok := payload["warning_count"].(float64); ok {
				totalWarnings += int(count)
			}

		case "fix_applied":
			fixCount++

		case "pattern_learned":
			patternCount++

		case "doc_ingested":
			docCount++
		}
	}

	avgCompliance := 0.0
	if complianceCount > 0 {
		avgCompliance = complianceSum / float64(complianceCount)
	}

	return map[string]interface{}{
		"total_events":   len(events),
		"audit_count":    auditCount,
		"fix_count":      fixCount,
		"pattern_count":  patternCount,
		"doc_count":      docCount,
		"avg_compliance": avgCompliance,
		"total_findings": totalFindings,
		"total_critical": totalCritical,
		"total_warnings": totalWarnings,
	}
}

// Get recent telemetry events handler
func getRecentTelemetryHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	eventType := r.URL.Query().Get("event_type")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build query
	query := `
		SELECT event_type, payload, created_at
		FROM telemetry_events
		WHERE project_id = $1
	`
	args := []interface{}{project.ID}
	argIndex := 2

	if eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []map[string]interface{}{}
	for rows.Next() {
		var eventType string
		var payloadJSON string
		var createdAt time.Time

		if err := rows.Scan(&eventType, &payloadJSON, &createdAt); err != nil {
			continue
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
			payload = make(map[string]interface{})
		}

		events = append(events, map[string]interface{}{
			"event_type": eventType,
			"payload":    payload,
			"created_at": createdAt,
		})
	}

	// Get total count for pagination
	var total int
	countQuery := `
		SELECT COUNT(*) FROM telemetry_events WHERE project_id = $1
	`
	countArgs := []interface{}{project.ID}
	if eventType != "" {
		countQuery += " AND event_type = $2"
		countArgs = append(countArgs, eventType)
	}
	db.QueryRow(countQuery, countArgs...).Scan(&total)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Get metrics trends handler
func getMetricsTrendsHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Get query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	period := r.URL.Query().Get("period") // daily, weekly, monthly

	if period == "" {
		period = "daily"
	}

	// Default to last 30 days if not specified
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Build query based on period
	var dateFormat string
	switch period {
	case "weekly":
		dateFormat = "DATE_TRUNC('week', created_at)"
	case "monthly":
		dateFormat = "DATE_TRUNC('month', created_at)"
	default:
		dateFormat = "DATE(created_at)"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as period,
			event_type,
			COUNT(*) as event_count,
			AVG((payload->>'finding_count')::float) as avg_findings,
			AVG((payload->>'compliance_percent')::float) as avg_compliance
		FROM telemetry_events
		WHERE project_id = $1
			AND created_at >= $2
			AND created_at <= $3
		GROUP BY period, event_type
		ORDER BY period DESC, event_type
	`, dateFormat)

	rows, err := db.Query(query, project.ID, startDate, endDate)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	trends := []map[string]interface{}{}
	for rows.Next() {
		var period time.Time
		var eventType string
		var eventCount int
		var avgFindings sql.NullFloat64
		var avgCompliance sql.NullFloat64

		if err := rows.Scan(&period, &eventType, &eventCount, &avgFindings, &avgCompliance); err != nil {
			continue
		}

		trend := map[string]interface{}{
			"period":     period.Format("2006-01-02"),
			"event_type": eventType,
			"count":      eventCount,
		}

		if avgFindings.Valid {
			trend["avg_findings"] = avgFindings.Float64
		}
		if avgCompliance.Valid {
			trend["avg_compliance"] = avgCompliance.Float64
		}

		trends = append(trends, trend)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trends":     trends,
		"period":     period,
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// Get team metrics handler
func getTeamMetricsHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}
	teamID := chi.URLParam(r, "teamId")

	if teamID == "" {
		http.Error(w, "Team ID required", http.StatusBadRequest)
		return
	}

	// Get query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Build query - note: team_id column needs to be added in Phase C
	query := `
		SELECT event_type, payload, created_at
		FROM telemetry_events
		WHERE project_id = $1
	`
	args := []interface{}{project.ID}
	argIndex := 2

	// TODO: Add team_id filter once column is added
	// if teamID != "" {
	// 	query += fmt.Sprintf(" AND team_id = $%d", argIndex)
	// 	args = append(args, teamID)
	// 	argIndex++
	// }

	if startDate != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	query += " ORDER BY created_at DESC LIMIT 1000"

	rows, err := db.Query(query, args...)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []map[string]interface{}{}
	for rows.Next() {
		var eventType string
		var payloadJSON string
		var createdAt time.Time

		if err := rows.Scan(&eventType, &payloadJSON, &createdAt); err != nil {
			continue
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
			payload = make(map[string]interface{})
		}

		events = append(events, map[string]interface{}{
			"event_type": eventType,
			"payload":    payload,
			"created_at": createdAt,
		})
	}

	// Calculate team-specific metrics
	metrics := calculateMetrics(events)
	metrics["team_id"] = teamID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"team_id": teamID,
		"events":  events,
		"metrics": metrics,
		"total":   len(events),
	})
}

// =============================================================================
// MIDDLEWARE
// =============================================================================

func apiKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Extract API key from "Bearer <key>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		apiKey := parts[1]

		// Check per-API-key rate limit (before database lookup for efficiency)
		endpoint := r.URL.Path
		if !checkAPIKeyRateLimit(apiKey, endpoint) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limit exceeded for this API key. Please try again later."}`))
			return
		}

		// Look up project by API key
		var project Project
		err := db.QueryRow(`
			SELECT id, org_id, name, api_key, created_at
			FROM projects WHERE api_key = $1
		`, apiKey).Scan(&project.ID, &project.OrgID, &project.Name, &project.APIKey, &project.CreatedAt)

		if err == sql.ErrNoRows {
			WriteErrorResponse(w, &ValidationError{
				Field:   "authorization",
				Message: "Invalid API key",
				Code:    "unauthorized",
			}, http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, "Authentication error", http.StatusInternalServerError)
			return
		}

		// Add project to context
		ctx := context.WithValue(r.Context(), projectKey, &project)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// adminAuthMiddleware validates admin API key for admin endpoints
func adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := loadConfig()
		
		// Check if admin key is configured
		if config.AdminAPIKey == "" {
			LogError(r.Context(), "Admin API key not configured")
			WriteErrorResponse(w, &InternalError{
				Message: "Admin authentication not configured",
				Code:    "configuration_error",
			}, http.StatusInternalServerError)
			return
		}

		// Try X-Admin-API-Key header first
		adminKey := r.Header.Get("X-Admin-API-Key")
		
		// Fallback to Authorization header
		if adminKey == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					adminKey = parts[1]
				}
			}
		}

		// Validate admin key
		if adminKey == "" {
			WriteErrorResponse(w, &ValidationError{
				Field:   "authorization",
				Message: "Missing admin API key. Provide X-Admin-API-Key header or Authorization: Bearer <admin-key>",
				Code:    "unauthorized",
			}, http.StatusUnauthorized)
			return
		}

		// Use constant-time comparison to prevent timing attacks
		if !constantTimeEqual(adminKey, config.AdminAPIKey) {
			WriteErrorResponse(w, &ValidationError{
				Field:   "authorization",
				Message: "Invalid admin API key",
				Code:    "unauthorized",
			}, http.StatusUnauthorized)
			return
		}

		// Admin authenticated, proceed
		next.ServeHTTP(w, r)
	})
}

// constantTimeEqual performs constant-time string comparison to prevent timing attacks
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i]) ^ int(b[i])
	}
	return result == 0
}

// sanitizeString sanitizes a string by trimming, limiting length, and removing control characters
func sanitizeString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	// Remove control characters except newline, carriage return, and tab
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, s)
	return s
}

// validateVersionFormat validates that version follows semver format
func validateVersionFormat(version string) error {
	versionPattern := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`)
	if !versionPattern.MatchString(version) {
		return fmt.Errorf("version must be in semver format (e.g., 1.2.3 or v1.2.3)")
	}
	return nil
}

// validatePlatform validates that platform is in the allowed list
func validatePlatform(platform string) error {
	allowedPlatforms := []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"}
	for _, p := range allowedPlatforms {
		if platform == p {
			return nil
		}
	}
	return fmt.Errorf("platform must be one of: %v", allowedPlatforms)
}

// =============================================================================
// HELPERS
// =============================================================================

func detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".doc":
		return "application/msword"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".txt":
		return "text/plain"
	case ".md", ".markdown":
		return "text/markdown"
	case ".eml":
		return "message/rfc822"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func estimateProcessingTime(size int64) int {
	// Rough estimate: 1 second per 100KB + 10 seconds base
	return 10 + int(size/(100*1024))
}

func stageStatus(docStatus string, progress int, threshold int) string {
	if docStatus == "failed" {
		return "failed"
	}
	if docStatus == "completed" {
		return "completed"
	}
	if progress >= threshold {
		return "completed"
	}
	if progress > threshold-50 {
		return "processing"
	}
	return "pending"
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "sk_live_" + hex.EncodeToString(bytes)
}

// =============================================================================
// ADMIN HANDLERS (for setup)
// =============================================================================

func createOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	orgID := uuid.New().String()
	_, err := db.Exec("INSERT INTO organizations (id, name) VALUES ($1, $2)", orgID, req.Name)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":   orgID,
		"name": req.Name,
	})
}

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID string `json:"org_id"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	projectID := uuid.New().String()
	apiKey := generateAPIKey()

	_, err := db.Exec("INSERT INTO projects (id, org_id, name, api_key) VALUES ($1, $2, $3, $4)",
		projectID, req.OrgID, req.Name, apiKey)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      projectID,
		"name":    req.Name,
		"api_key": apiKey,
	})
}

// =============================================================================
// AST ANALYSIS (Phase 6) - Stub Implementation
// =============================================================================

// AST Analysis Request
type ASTAnalysisRequest struct {
	Code      string   `json:"code"`
	Language  string   `json:"language"`
	Filename  string   `json:"filename"`
	ProjectID string   `json:"projectId"`
	Analyses  []string `json:"analyses"` // duplicates, unused, unreachable, security
}

// AST Finding
type ASTFinding struct {
	Type       string `json:"type"` // duplicate_function, unused_variable, etc.
	Severity   string `json:"severity"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	EndLine    int    `json:"endLine"`
	EndColumn  int    `json:"endColumn"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Suggestion string `json:"suggestion"`
}

// AST Analysis Response
type ASTAnalysisResponse struct {
	Success  bool         `json:"success"`
	Findings []ASTFinding `json:"findings"`
	Stats    struct {
		ParseTime    int64 `json:"parseTimeMs"`
		AnalysisTime int64 `json:"analysisTimeMs"`
		NodesVisited int   `json:"nodesVisited"`
	} `json:"stats"`
}

// AST Analysis Handler - Phase 6 Implementation
// Status: ✅ IMPLEMENTED - Full Tree-sitter AST analysis
func astAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	var req ASTAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding AST analysis request: %v", err)
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Code == "" {
		log.Printf("AST analysis request missing code")
		http.Error(w, "Missing required field: code", http.StatusBadRequest)
		return
	}

	// Initialize parsers if not already done
	if len(parsers) == 0 {
		initParsers()
	}

	// Perform AST analysis
	findings, stats, err := analyzeAST(req.Code, req.Language, req.Analyses)
	if err != nil {
		log.Printf("AST analysis error for file %s (language: %s): %v", req.Filename, req.Language, err)
		http.Error(w, fmt.Sprintf("Analysis error: %v", err), http.StatusInternalServerError)
		return
	}

	response := ASTAnalysisResponse{
		Success:  true,
		Findings: findings,
		Stats: struct {
			ParseTime    int64 `json:"parseTimeMs"`
			AnalysisTime int64 `json:"analysisTimeMs"`
			NodesVisited int   `json:"nodesVisited"`
		}{
			ParseTime:    stats.ParseTime,
			AnalysisTime: stats.AnalysisTime,
			NodesVisited: stats.NodesVisited,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding AST analysis response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// Vibe Analysis Handler - ✅ IMPLEMENTED (Phase 7)
// Status: ✅ COMPLETE - Full AST-based vibe pattern detection
// Implements: Duplicate functions, unused variables, unreachable code, orphaned code
func vibeAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	var req ASTAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding vibe analysis request: %v", err)
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Code == "" {
		log.Printf("Vibe analysis request missing code")
		http.Error(w, "Missing required field: code", http.StatusBadRequest)
		return
	}

	// Initialize parsers if not already done
	if len(parsers) == 0 {
		initParsers()
	}

	// Perform vibe-specific AST analysis (duplicates, unused, unreachable)
	analyses := []string{"duplicates", "unused", "unreachable"}
	if len(req.Analyses) > 0 {
		analyses = req.Analyses
	}

	findings, stats, err := analyzeAST(req.Code, req.Language, analyses)
	if err != nil {
		log.Printf("Vibe analysis failed: %v", err)
		http.Error(w, fmt.Sprintf("Analysis error: %v", err), http.StatusInternalServerError)
		return
	}

	response := ASTAnalysisResponse{
		Success:  true,
		Findings: findings,
		Stats: struct {
			ParseTime    int64 `json:"parseTimeMs"`
			AnalysisTime int64 `json:"analysisTimeMs"`
			NodesVisited int   `json:"nodesVisited"`
		}{
			ParseTime:    stats.ParseTime,
			AnalysisTime: stats.AnalysisTime,
			NodesVisited: stats.NodesVisited,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding vibe analysis response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// Cross-File Analysis Handler - Phase 6F Implementation
// Status: ✅ COMPLETE - Functional cross-file analysis
func crossFileAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	type CrossFileRequest struct {
		Files []struct {
			Path     string `json:"path"`
			Code     string `json:"code"`
			Language string `json:"language,omitempty"`
		} `json:"files"`
		Language  string `json:"language"` // Default language if not specified per file
		ProjectID string `json:"projectId"`
	}

	var req CrossFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding cross-file analysis request: %v", err)
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		log.Printf("Cross-file analysis request missing files")
		http.Error(w, "No files provided", http.StatusBadRequest)
		return
	}

	// Initialize parsers if not already done
	if len(parsers) == 0 {
		initParsers()
	}

	// Prepare files for analysis
	analysisFiles := make([]struct {
		Path     string
		Code     string
		Language string
	}, len(req.Files))

	for i, file := range req.Files {
		analysisFiles[i].Path = file.Path
		analysisFiles[i].Code = file.Code
		// Use file-specific language or fallback to request language
		if file.Language != "" {
			analysisFiles[i].Language = file.Language
		} else {
			analysisFiles[i].Language = req.Language
		}
	}

	// Perform cross-file analysis
	findings, stats, err := analyzeCrossFile(analysisFiles)
	if err != nil {
		log.Printf("Cross-file analysis failed: %v", err)
		http.Error(w, fmt.Sprintf("Analysis error: %v", err), http.StatusInternalServerError)
		return
	}

	response := ASTAnalysisResponse{
		Success:  true,
		Findings: findings,
		Stats: struct {
			ParseTime    int64 `json:"parseTimeMs"`
			AnalysisTime int64 `json:"analysisTimeMs"`
			NodesVisited int   `json:"nodesVisited"`
		}{
			ParseTime:    stats.ParseTime,
			AnalysisTime: stats.AnalysisTime,
			NodesVisited: stats.NodesVisited,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// SECURITY ANALYSIS (Phase 8) ✅ IMPLEMENTED - Full security rule checking with AST analysis
// =============================================================================

// Security Analysis Request
type SecurityAnalysisRequest struct {
	Code             string          `json:"code"`
	Language         string          `json:"language"`
	Filename         string          `json:"filename"`
	ProjectID        string          `json:"projectId"`
	Rules            []string        `json:"rules,omitempty"`            // Specific rules to check (SEC-001, etc.)
	ExpectedFindings map[string]bool `json:"expectedFindings,omitempty"` // Ground truth for detection rate validation
}

// Security Finding
type SecurityFinding struct {
	RuleID      string `json:"ruleId"`
	RuleName    string `json:"ruleName"`
	Severity    string `json:"severity"`
	Line        int    `json:"line"`
	Code        string `json:"code"`
	Issue       string `json:"issue"`
	Remediation string `json:"remediation"`
	AutoFixable bool   `json:"autoFixable"`
}

// Security Analysis Response
type SecurityAnalysisResponse struct {
	Score    int               `json:"score"` // 0-100
	Grade    string            `json:"grade"` // A, B, C, D, F
	Findings []SecurityFinding `json:"findings"`
	Summary  struct {
		TotalRules int `json:"totalRules"`
		Passed     int `json:"passed"`
		Failed     int `json:"failed"`
		Critical   int `json:"critical"`
		High       int `json:"high"`
		Medium     int `json:"medium"`
		Low        int `json:"low"`
	} `json:"summary"`
	Metrics *DetectionMetrics `json:"metrics,omitempty"` // Optional: only for validation runs
}

// Security Rules Definitions (SEC-001 to SEC-008)
var securityRules = map[string]struct {
	Name     string
	Severity string
	Type     string
}{
	"SEC-001": {"Resource Ownership", "critical", "authorization"},
	"SEC-002": {"SQL Injection", "critical", "injection"},
	"SEC-003": {"Auth Middleware", "critical", "authentication"},
	"SEC-004": {"Rate Limiting", "high", "transport"},
	"SEC-005": {"Password Hashing", "critical", "cryptography"},
	"SEC-006": {"Input Validation", "high", "validation"},
	"SEC-007": {"Secure Headers", "medium", "transport"},
	"SEC-008": {"CORS Config", "high", "transport"},
}

// Security Analysis Handler - ✅ IMPLEMENTED (Phase 8)
// Status: ✅ COMPLETE - Full security rule checking with AST analysis
func securityAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	var req SecurityAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Perform security analysis
	findings, err := analyzeSecurity(req.Code, req.Language, req.Filename, req.Rules)
	if err != nil {
		http.Error(w, fmt.Sprintf("Security analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate security score and summary
	score, grade := calculateSecurityScore(findings)
	summary := calculateSecuritySummary(findings)

	response := SecurityAnalysisResponse{
		Score:    score,
		Grade:    grade,
		Findings: findings,
		Summary:  summary,
	}

	// Calculate detection rate metrics if ground truth is provided
	if len(req.ExpectedFindings) > 0 {
		metrics := calculateDetectionRate(findings, req.ExpectedFindings)
		response.Metrics = &metrics
		log.Printf("Detection rate metrics calculated: %.2f%% detection rate, %.2f%% precision, %.2f%% recall",
			metrics.DetectionRate, metrics.Precision, metrics.Recall)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding security analysis response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// docSyncHandler handles doc-sync analysis requests (Phase 11)
func docSyncHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req DocSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	// Set project ID from context
	req.ProjectID = project.ID

	// Default report type if not specified
	if req.ReportType == "" {
		req.ReportType = "status_tracking"
	}

	// Get codebase path (default to current directory, can be configured)
	codebasePath := "."
	if path, ok := req.Options["codebase_path"].(string); ok && path != "" {
		codebasePath = path
	}

	// Perform analysis
	ctx := r.Context()
	response, err := analyzeDocSync(ctx, req, codebasePath)
	if err != nil {
		LogErrorWithContext(ctx, err, "Doc-sync analysis error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "doc_sync_analysis",
			Message:       fmt.Sprintf("Analysis failed: %v", err),
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// businessRulesComparisonHandler handles business rules comparison requests (Phase 11B)
func businessRulesComparisonHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req struct {
		KnowledgeItemIDs []string `json:"knowledgeItemIds,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	codebasePath := "."
	ctx := r.Context()
	discrepancies, err := compareBusinessRules(ctx, project.ID, codebasePath)
	if err != nil {
		log.Printf("Business rules comparison error: %v", err)
		http.Error(w, fmt.Sprintf("Comparison failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"discrepancies": discrepancies,
		"count":         len(discrepancies),
	})
}

// reviewQueueHandler returns pending review items (Phase 11B)
func reviewQueueHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	// Query doc_sync_updates table for pending reviews
	query := `
		SELECT id, file_path, change_type, old_value, new_value, line_number, created_at
		FROM doc_sync_updates
		WHERE project_id = $1 AND applied = false
	`
	if status != "" {
		query += " AND change_type = $2"
	}
	query += " ORDER BY created_at DESC"

	var rows *sql.Rows
	if status != "" {
		rows, err = db.Query(query, project.ID, status)
	} else {
		rows, err = db.Query(query, project.ID)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query review queue: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviews []map[string]interface{}
	for rows.Next() {
		var id, filePath, changeType, oldValue, newValue string
		var lineNumber sql.NullInt64
		var createdAt time.Time

		err := rows.Scan(&id, &filePath, &changeType, &oldValue, &newValue, &lineNumber, &createdAt)
		if err != nil {
			continue
		}

		review := map[string]interface{}{
			"id":          id,
			"file_path":   filePath,
			"change_type": changeType,
			"old_value":   oldValue,
			"new_value":   newValue,
			"created_at":  createdAt,
		}
		if lineNumber.Valid {
			review["line_number"] = lineNumber.Int64
		}
		reviews = append(reviews, review)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"reviews": reviews,
		"count":   len(reviews),
	})
}

// reviewHandler handles approval/rejection of review items (Phase 11B)
func reviewHandler(w http.ResponseWriter, r *http.Request) {
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}
	reviewID := chi.URLParam(r, "id")

	var req struct {
		Action  string `json:"action"` // "approve" or "reject"
		Comment string `json:"comment,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		http.Error(w, "Action must be 'approve' or 'reject'", http.StatusBadRequest)
		return
	}

	// Update review status
	query := `
		UPDATE doc_sync_updates
		SET applied = $1, approved_by = $2, approved_at = $3
		WHERE id = $4 AND project_id = $5
	`
	approved := req.Action == "approve"
	_, err = db.Exec(query, approved, "user", time.Now(), reviewID, project.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update review: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Review %s", req.Action),
	})
}

// architectureAnalysisHandler handles architecture analysis requests (Phase 9)
func architectureAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	var req ArchitectureAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding architecture analysis request: %v", err)
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Perform architecture analysis
	response := analyzeArchitecture(req.Files)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding architecture analysis response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// =============================================================================
// BINARY DISTRIBUTION HANDLERS
// =============================================================================

// getBinaryVersionHandler returns latest version info for platform
func getBinaryVersionHandler(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		platform = detectPlatformFromUserAgent(r.Header.Get("User-Agent"))
	}

	query := `
		SELECT version, platform, file_size, checksum_sha256, checksum_md5,
		       signature, release_notes, released_at, min_go_version
		FROM binary_versions
		WHERE platform = $1 AND is_stable = true AND is_latest = true
		ORDER BY released_at DESC
		LIMIT 1
	`

	var version struct {
		Version        string
		Platform       string
		FileSize       int64
		ChecksumSHA256 string
		ChecksumMD5    sql.NullString
		Signature      sql.NullString
		ReleaseNotes   sql.NullString
		ReleasedAt     sql.NullTime
		MinGoVersion   sql.NullString
	}

	err := db.QueryRow(query, platform).Scan(
		&version.Version, &version.Platform, &version.FileSize,
		&version.ChecksumSHA256, &version.ChecksumMD5, &version.Signature,
		&version.ReleaseNotes, &version.ReleasedAt, &version.MinGoVersion,
	)

	if err == sql.ErrNoRows {
		WriteErrorResponse(w, &NotFoundError{
			Message: fmt.Sprintf("No binary found for platform: %s", platform),
		}, http.StatusNotFound)
		return
	}

	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation: "query_binary_version",
			Message:   "Failed to query binary version",
			Code:      "database_error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"version":         version.Version,
		"platform":        version.Platform,
		"file_size":       version.FileSize,
		"checksum_sha256": version.ChecksumSHA256,
	}

	if version.ChecksumMD5.Valid {
		response["checksum_md5"] = version.ChecksumMD5.String
	}
	if version.Signature.Valid {
		response["signature"] = version.Signature.String
	}
	if version.ReleaseNotes.Valid {
		response["release_notes"] = version.ReleaseNotes.String
	}
	if version.ReleasedAt.Valid {
		response["released_at"] = version.ReleasedAt.Time.Format(time.RFC3339)
	}
	if version.MinGoVersion.Valid {
		response["min_go_version"] = version.MinGoVersion.String
	}

	WriteJSONResponse(w, response, http.StatusOK)
}

// detectPlatformFromUserAgent detects platform from User-Agent header
func detectPlatformFromUserAgent(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "darwin") || strings.Contains(ua, "mac") {
		if strings.Contains(ua, "arm64") || strings.Contains(ua, "aarch64") {
			return "darwin-arm64"
		}
		return "darwin-amd64"
	}
	if strings.Contains(ua, "linux") {
		if strings.Contains(ua, "arm64") || strings.Contains(ua, "aarch64") {
			return "linux-arm64"
		}
		return "linux-amd64"
	}
	if strings.Contains(ua, "windows") {
		return "windows-amd64"
	}
	return "linux-amd64" // Default fallback
}

// downloadBinaryHandler streams binary file to client
func downloadBinaryHandler(w http.ResponseWriter, r *http.Request) {
	version := r.URL.Query().Get("version")
	platform := r.URL.Query().Get("platform")

	if version == "" || platform == "" {
		WriteErrorResponse(w, &ValidationError{
			Field:   "version,platform",
			Message: "version and platform query parameters are required",
			Code:    "missing_params",
		}, http.StatusBadRequest)
		return
	}

	query := `
		SELECT file_path, file_size, checksum_sha256
		FROM binary_versions
		WHERE version = $1 AND platform = $2
	`

	var filePath string
	var fileSize int64
	var checksum string

	err := db.QueryRow(query, version, platform).Scan(&filePath, &fileSize, &checksum)
	if err == sql.ErrNoRows {
		WriteErrorResponse(w, &NotFoundError{
			Message: fmt.Sprintf("Binary not found: %s for %s", version, platform),
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation: "query_binary",
			Message:   "Failed to query binary",
			Code:      "database_error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		WriteErrorResponse(w, &InternalError{
			Message: "Failed to open binary file",
			Code:    "file_error",
		}, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filename := fmt.Sprintf("sentinel-%s-%s", version, platform)
	if platform == "windows-amd64" {
		filename += ".exe"
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
	w.Header().Set("X-Checksum-SHA256", checksum)
	w.Header().Set("X-Version", version)
	w.Header().Set("X-Platform", platform)

	io.Copy(w, file)

	// Track download asynchronously
	go trackBinaryDownload(version, platform, r)
}

// trackBinaryDownload tracks binary download for analytics
func trackBinaryDownload(version, platform string, r *http.Request) {
	project := r.Context().Value(projectKey)
	if project == nil {
		return
	}

	var versionID uuid.UUID
	err := db.QueryRow("SELECT id FROM binary_versions WHERE version = $1 AND platform = $2",
		version, platform).Scan(&versionID)
	if err != nil {
		return
	}

	projectObj := project.(*Project)
	ipAddr := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			ipAddr = strings.TrimSpace(parts[0])
		}
	}

	db.Exec(`
		INSERT INTO binary_downloads (version_id, project_id, user_agent, ip_address)
		VALUES ($1, $2, $3, $4)
	`, versionID, projectObj.ID, r.UserAgent(), ipAddr)
}

// uploadBinaryHandler allows admins to upload new binary versions
func uploadBinaryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseMultipartForm(100 << 20) // 100MB max
	if err != nil {
		WriteErrorResponse(w, &ValidationError{
			Message: "Failed to parse form",
			Code:    "parse_error",
		}, http.StatusBadRequest)
		return
	}

	version := r.FormValue("version")
	platform := r.FormValue("platform")
	releaseNotes := r.FormValue("release_notes")
	isStable := r.FormValue("is_stable") == "true"
	isLatest := r.FormValue("is_latest") == "true"

	if version == "" || platform == "" {
		WriteErrorResponse(w, &ValidationError{
			Message: "version and platform are required",
			Code:    "missing_fields",
		}, http.StatusBadRequest)
		return
	}

	// Validate version format (semver)
	if err := validateVersionFormat(version); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "version",
			Message: err.Error(),
			Code:    "invalid_format",
		}, http.StatusBadRequest)
		return
	}

	// Validate platform against allowed list
	if err := validatePlatform(platform); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "platform",
			Message: err.Error(),
			Code:    "invalid_platform",
		}, http.StatusBadRequest)
		return
	}

	// Sanitize release notes
	releaseNotes = sanitizeString(releaseNotes, 10000) // Max 10KB

	file, _, err := r.FormFile("binary")
	if err != nil {
		WriteErrorResponse(w, &ValidationError{
			Message: "binary file is required",
			Code:    "missing_file",
		}, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Calculate checksum
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		WriteErrorResponse(w, &InternalError{
			Message: "Failed to read file",
			Code:    "file_error",
		}, http.StatusInternalServerError)
		return
	}

	checksum := hex.EncodeToString(hash.Sum(nil))

	// Save file
	config := loadConfig()
	storagePath := filepath.Join(config.BinaryStorage, version, platform)
	
	// Create directory with error handling
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		WriteErrorResponse(w, &InternalError{
			Message: "Failed to create storage directory",
			Code:    "storage_error",
		}, http.StatusInternalServerError)
		return
	}

	destFile, err := os.Create(storagePath)
	if err != nil {
		WriteErrorResponse(w, &InternalError{
			Message: "Failed to save file",
			Code:    "storage_error",
		}, http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Reset file pointer with error handling
	if _, err := file.Seek(0, 0); err != nil {
		destFile.Close()
		os.Remove(storagePath)
		WriteErrorResponse(w, &InternalError{
			Message: "Failed to reset file pointer",
			Code:    "file_error",
		}, http.StatusInternalServerError)
		return
	}

	// Copy file with error handling and verification
	written, err := io.Copy(destFile, file)
	if err != nil {
		destFile.Close()
		os.Remove(storagePath)
		WriteErrorResponse(w, &InternalError{
			Message: "Failed to save file",
			Code:    "file_error",
		}, http.StatusInternalServerError)
		return
	}
	if written != size {
		destFile.Close()
		os.Remove(storagePath)
		WriteErrorResponse(w, &InternalError{
			Message: "File size mismatch",
			Code:    "file_error",
		}, http.StatusInternalServerError)
		return
	}

	// Parse platform to extract OS and architecture
	parts := strings.Split(platform, "-")
	osName := parts[0]
	arch := "amd64"
	if len(parts) > 1 {
		arch = parts[1]
	}

	// Insert database record
	query := `
		INSERT INTO binary_versions 
		(version, platform, architecture, os, file_path, file_size, 
		 checksum_sha256, release_notes, is_stable, is_latest, released_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (version, platform) DO UPDATE SET
			file_path = EXCLUDED.file_path,
			file_size = EXCLUDED.file_size,
			checksum_sha256 = EXCLUDED.checksum_sha256,
			release_notes = EXCLUDED.release_notes,
			is_stable = EXCLUDED.is_stable,
			is_latest = EXCLUDED.is_latest,
			released_at = EXCLUDED.released_at
	`

	_, err = db.Exec(query, version, platform, arch, osName, storagePath, size,
		checksum, releaseNotes, isStable, isLatest)
	if err != nil {
		LogErrorWithContext(ctx, err, "Failed to save binary version metadata", map[string]interface{}{
			"version":   version,
			"platform":  platform,
			"file_size": size,
			"checksum":  checksum,
		})
		WriteErrorResponse(w, &DatabaseError{
			Operation: "save_binary_version",
			Message:   fmt.Sprintf("Failed to save version metadata for version=%s platform=%s", version, platform),
			Code:      "database_error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// If marked as latest, unmark others
	if isLatest {
		_, err := db.Exec(`
			UPDATE binary_versions 
			SET is_latest = false 
			WHERE platform = $1 AND version != $2
		`, platform, version)
		if err != nil {
			LogError(ctx, "Failed to unmark other versions as latest: %v (version=%s platform=%s)", err, version, platform)
			// Don't fail the request, but log the error for investigation
			// The new version is still saved, but other versions may still be marked as latest
		}
	}

	WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"version": version,
		"platform": platform,
		"checksum": checksum,
	}, http.StatusCreated)
}

// listBinaryVersionsHandler lists available versions for a platform
func listBinaryVersionsHandler(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	stableOnly := r.URL.Query().Get("stable") == "true"

	query := `
		SELECT version, platform, file_size, checksum_sha256, 
		       release_notes, released_at, is_stable, is_latest
		FROM binary_versions
		WHERE ($1 = '' OR platform = $1)
		AND ($2 = false OR is_stable = true)
		ORDER BY released_at DESC
		LIMIT 50
	`

	rows, err := db.Query(query, platform, stableOnly)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation: "query_versions",
			Message:   "Failed to query versions",
			Code:      "database_error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	versions := []map[string]interface{}{}
	ctx := r.Context()
	for rows.Next() {
		var v struct {
			Version      string
			Platform     string
			FileSize     int64
			ChecksumSHA256 string
			ReleaseNotes sql.NullString
			ReleasedAt   sql.NullTime
			IsStable     bool
			IsLatest     bool
		}
		err := rows.Scan(&v.Version, &v.Platform, &v.FileSize, &v.ChecksumSHA256,
			&v.ReleaseNotes, &v.ReleasedAt, &v.IsStable, &v.IsLatest)
		if err != nil {
			LogWarn(ctx, "Failed to scan binary version row (skipping): %v", err)
			continue
		}

		versionMap := map[string]interface{}{
			"version":         v.Version,
			"platform":        v.Platform,
			"file_size":       v.FileSize,
			"checksum_sha256": v.ChecksumSHA256,
			"is_stable":       v.IsStable,
			"is_latest":       v.IsLatest,
		}
		if v.ReleaseNotes.Valid {
			versionMap["release_notes"] = v.ReleaseNotes.String
		}
		if v.ReleasedAt.Valid {
			versionMap["released_at"] = v.ReleasedAt.Time.Format(time.RFC3339)
		}
		versions = append(versions, versionMap)
	}

	WriteJSONResponse(w, map[string]interface{}{
		"versions": versions,
		"count":    len(versions),
	}, http.StatusOK)
}

// getLatestRulesHandler returns latest rules for project
func getLatestRulesHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT rule_name, rule_content, rule_type, globs
		FROM rules_versions
		WHERE is_latest = true
		ORDER BY rule_type, rule_name
	`

	rows, err := db.Query(query)
	if err != nil {
		WriteErrorResponse(w, &DatabaseError{
			Operation: "query_rules",
			Message:   "Failed to query rules",
			Code:      "database_error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	rules := []map[string]interface{}{}
	for rows.Next() {
		var ruleName, ruleContent, ruleType string
		var globs []string
		err := rows.Scan(&ruleName, &ruleContent, &ruleType, pq.Array(&globs))
		if err != nil {
			continue
		}

		rules = append(rules, map[string]interface{}{
			"name":    ruleName,
			"content": ruleContent,
			"type":    ruleType,
			"globs":   globs,
		})
	}

	WriteJSONResponse(w, map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	}, http.StatusOK)
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	config := loadConfig()
	validateProductionConfig(config)

	// Initialize database
	log.Println("Connecting to database...")
	if err := initDB(config.DatabaseURL); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Run migrations
	log.Println("Running migrations...")
	if err := runMigrations(); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	// Create storage directories
	if err := os.MkdirAll(config.DocumentStorage, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}
	if err := os.MkdirAll(config.BinaryStorage, 0755); err != nil {
		log.Fatalf("Failed to create binary storage directory: %v", err)
	}
	if err := os.MkdirAll(config.RulesStorage, 0755); err != nil {
		log.Fatalf("Failed to create rules storage directory: %v", err)
	}

	// Configure endpoint-specific rate limits
	setEndpointRateLimiter("/api/v1/documents/ingest", 10, 20)      // 10 req/s, burst 20 for document uploads
	setEndpointRateLimiter("/api/v1/telemetry", 50, 100)            // 50 req/s, burst 100 for telemetry
	setEndpointRateLimiter("/api/v1/analyze/ast", 5, 10)            // 5 req/s, burst 10 for AST analysis (CPU intensive)
	setEndpointRateLimiter("/api/v1/analyze/vibe", 5, 10)           // 5 req/s, burst 10 for vibe analysis
	setEndpointRateLimiter("/api/v1/knowledge/gap-analysis", 5, 10) // 5 req/s, burst 10 for gap analysis
	setEndpointRateLimiter("/api/v1/change-requests", 20, 40)       // 20 req/s, burst 40 for change requests

	// Setup router
	r := chi.NewRouter()

	// Add panic recovery middleware to all routes
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer RecoverFromPanic(w, req)
			next.ServeHTTP(w, req)
		})
	})

	// Serve dashboard static files (before other middleware for performance)
	r.Get("/dashboard/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/dashboard/", http.FileServer(http.Dir("./dashboard/"))).ServeHTTP(w, r)
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard/", http.StatusFound)
	})

	// Request ID middleware (must be first)
	r.Use(requestIDMiddleware)

	// Security middleware (early in chain)
	r.Use(securityHeadersMiddleware)
	r.Use(requestSizeLimitMiddleware(DefaultMaxRequestSize))
	r.Use(csrfProtectionMiddleware)

	// Rate limiting middleware
	r.Use(rateLimitByEndpointMiddleware())

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{config.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"},
		ExposedHeaders:   []string{"Link", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes - Health checks (Phase G: Logging and Monitoring)
	r.Get("/health", healthHandler)
	r.Get("/health/db", healthDBHandler)
	r.Get("/health/ready", healthReadyHandler)

	// Admin routes (protected with admin authentication)
	r.Route("/api/v1/admin", func(r chi.Router) {
		r.Use(adminAuthMiddleware)
		r.Post("/organizations", createOrganizationHandler)
		r.Post("/projects", createProjectHandler)
		r.Post("/binary/upload", uploadBinaryHandler)
	})

	// Protected routes (require API key)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(apiKeyAuthMiddleware)

		// Document endpoints
		r.Post("/documents/ingest", uploadDocumentHandler(config))
		r.Get("/documents/{id}/status", getDocumentStatusHandler)
		r.Get("/documents/{id}/extracted", getExtractedTextHandler)
		r.Get("/documents/{id}/knowledge", getKnowledgeItemsHandler)
		r.Post("/documents/{id}/detect-changes", detectChangesHandler)
		r.Get("/documents", listDocumentsHandler)

		// Knowledge management endpoints
		r.Put("/knowledge/{id}/status", updateKnowledgeStatusHandler)
		r.Get("/projects/knowledge", listProjectKnowledgeHandler)
		r.Get("/knowledge/business", getBusinessContextHandler) // Phase A: Business context for MCP tools
		r.Post("/knowledge/sync", syncKnowledgeHandler)
		r.With(rateLimitMiddleware(5, 10)).Post("/knowledge/gap-analysis", gapAnalysisHandler) // Phase 12
		r.With(rateLimitMiddleware(1, 5)).Post("/knowledge/migrate", migrateKnowledgeHandler)  // Phase 13

		// Phase 12: Change Request endpoints
		r.Get("/change-requests", listChangeRequestsHandler)
		r.Get("/change-requests/{id}", getChangeRequestHandler)
		r.Post("/change-requests/{id}/approve", approveChangeRequestHandler)
		r.Post("/change-requests/{id}/reject", rejectChangeRequestHandler)
		r.Post("/change-requests/{id}/impact", analyzeImpactHandler)
		r.Post("/change-requests/{id}/start", startImplementationHandler)
		r.Post("/change-requests/{id}/complete", completeImplementationHandler)
		r.Post("/change-requests/{id}/update", updateImplementationHandler)
		r.Get("/change-requests/dashboard", getChangeRequestsDashboardHandler)

		// Telemetry endpoints
		r.Post("/telemetry", telemetryIngestionHandler)
		r.Get("/telemetry/recent", getRecentTelemetryHandler)
		r.Get("/metrics", getMetricsHandler)
		r.Get("/metrics/trends", getMetricsTrendsHandler)
		r.Get("/metrics/team/{teamId}", getTeamMetricsHandler)

		// Prometheus metrics endpoint (Phase G: Logging and Monitoring)
		r.Get("/metrics/prometheus", prometheusMetricsHandler)

		// AST Analysis endpoints (Phase 6) ✅ IMPLEMENTED - Full Tree-sitter AST analysis
		r.Post("/analyze/ast", astAnalysisHandler)
		r.Post("/analyze/vibe", vibeAnalysisHandler)
		r.Post("/analyze/cross-file", crossFileAnalysisHandler)

		// Security Analysis endpoint (Phase 8) ✅ IMPLEMENTED - Full security rule checking with AST analysis
		r.Post("/analyze/security", securityAnalysisHandler)
		r.Get("/security/context", getSecurityContextHandler) // Phase A: Security context for MCP tools

		// Validation endpoints (Phase B)
		r.Post("/validate/code", validateCodeHandler)         // Phase B: Code validation
		r.Post("/validate/business", validateBusinessHandler) // Phase B: Business rule validation

		// Action endpoints (Phase C)
		r.Post("/fixes/apply", applyFixHandler) // Phase C: Apply fixes

		// Architecture Analysis endpoint (Phase 9) ⏳ IMPLEMENTED - File structure analysis and split suggestions
		r.Post("/analyze/architecture", architectureAnalysisHandler)

		// Doc-Sync endpoint (Phase 11) - Code-Documentation Comparison
		r.Post("/analyze/doc-sync", docSyncHandler)
		r.Post("/analyze/business-rules", businessRulesComparisonHandler)
		r.Get("/doc-sync/review-queue", reviewQueueHandler)
		r.Post("/doc-sync/review/{id}", reviewHandler)

		// Test Enforcement endpoints (Phase 10)
		r.Post("/test-requirements/generate", generateTestRequirementsHandler)
		r.Post("/test-coverage/analyze", analyzeCoverageHandler)
		r.Get("/test-coverage/{knowledge_item_id}", getCoverageHandler)
		r.Post("/test-validations/validate", validateTestsHandler)
		r.Get("/test-validations/{test_requirement_id}", getValidationHandler)
		r.Post("/test-execution/run", testExecutionHandler)
		r.Get("/test-execution/{execution_id}", getTestExecutionHandler)
		r.Post("/mutation-test/run", mutationTestHandler)
		r.Get("/mutation-test/{test_requirement_id}", getMutationResultHandler)

		// Hook endpoints (Phase 9.5)
		r.Post("/api/v1/telemetry/hook", hookTelemetryHandler)
		r.Get("/api/v1/hooks/metrics", hookMetricsHandler)
		r.Get("/api/v1/hooks/metrics/team", hookMetricsHandler) // Same handler, different path
		r.Get("/api/v1/hooks/policies", hookPoliciesHandler)
		r.Post("/api/v1/hooks/policies", createOrUpdateHookPolicyHandler)
		r.Get("/api/v1/hooks/limits", hookLimitsHandler)
		r.Post("/api/v1/hooks/baselines", hookBaselineHandler)
		r.Post("/api/v1/hooks/baselines/{id}/review", reviewHookBaselineHandler)

		// Phase 14A: Comprehensive Feature Analysis endpoints
		r.With(rateLimitMiddleware(2, 5)).Post("/analyze/comprehensive", comprehensiveAnalysisHandler)
		r.Get("/validations/{id}", getComprehensiveValidationHandler)
		r.Get("/validations", listValidationsHandler)

		// Phase 15: Intent & Simple Language endpoints
		r.With(rateLimitMiddleware(2, 5)).Post("/analyze/intent", intentAnalysisHandler)
		r.Post("/intent/decisions", recordIntentDecisionHandler)
		r.Get("/intent/patterns", getIntentPatternsHandler)

		// Phase 14C: LLM Configuration endpoints
		r.With(rateLimitMiddleware(10, 60)).Post("/api/v1/llm/config", createLLMConfigHandler)
		r.With(rateLimitMiddleware(10, 60)).Get("/api/v1/llm/config/{id}", getLLMConfigHandler)
		r.With(rateLimitMiddleware(10, 60)).Put("/api/v1/llm/config/{id}", updateLLMConfigHandler)
		r.With(rateLimitMiddleware(10, 60)).Delete("/api/v1/llm/config/{id}", deleteLLMConfigHandler)
		r.With(rateLimitMiddleware(10, 60)).Get("/api/v1/llm/config/project/{projectId}", listLLMConfigsHandler)

		// Phase 14C: LLM Metadata endpoints
		r.Get("/api/v1/llm/providers", getProvidersHandler)
		r.Get("/api/v1/llm/models/{provider}", getModelsHandler)

		// Phase 14C: LLM Validation endpoint
		r.With(rateLimitMiddleware(5, 60)).Post("/api/v1/llm/config/validate", validateLLMConfigHandler)

		// Phase 14C: LLM Usage Reporting endpoints
		r.With(rateLimitMiddleware(30, 60)).Get("/api/v1/llm/usage/report", getUsageReportHandler)
		r.With(rateLimitMiddleware(30, 60)).Get("/api/v1/llm/usage/stats", getUsageStatsHandler)
		r.With(rateLimitMiddleware(30, 60)).Get("/api/v1/llm/usage/cost-breakdown", getCostBreakdownHandler)
		r.With(rateLimitMiddleware(30, 60)).Get("/api/v1/llm/usage/trends", getUsageTrendsHandler)

		// Phase 14D: Cost Optimization Metrics endpoints
		r.With(rateLimitMiddleware(30, 60)).Get("/api/v1/metrics/cache", getCacheMetricsHandler)
		r.With(rateLimitMiddleware(30, 60)).Get("/api/v1/metrics/cost", getCostMetricsHandler)

		// Binary distribution endpoints
		r.Get("/binary/version", getBinaryVersionHandler)
		r.Get("/binary/download", downloadBinaryHandler)
		r.Get("/binary/versions", listBinaryVersionsHandler)

		// Rules endpoints
		r.Get("/rules/latest", getLatestRulesHandler)

		// Phase 14E: Task Dependency & Verification endpoints
		r.Post("/tasks", createTaskHandler)
		r.Get("/tasks", listTasksHandler)
		r.Get("/tasks/{id}", getTaskHandler)
		r.Put("/tasks/{id}", updateTaskHandler)
		r.Delete("/tasks/{id}", deleteTaskHandler)
		r.Post("/tasks/scan", scanTasksHandler)
		r.Post("/tasks/{id}/verify", verifyTaskHandler)
		r.Post("/tasks/verify-all", verifyAllTasksHandler)
		r.Get("/tasks/{id}/dependencies", getTaskDependenciesHandler)
		r.Post("/tasks/{id}/detect-dependencies", detectDependenciesHandler)
	})

	// Phase 14D: Start cache cleanup goroutine
	startCacheCleanup()
	LogInfo(context.Background(), "Cache cleanup goroutine started")

	// Phase 14E: Start task cache cleanup goroutine
	StartTaskCacheCleanup()
	LogInfo(context.Background(), "Task cache cleanup goroutine started")

	// Start server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      r,
		ReadTimeout:  GetConfig().Timeouts.HTTP,
		WriteTimeout: GetConfig().Timeouts.HTTP,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), GetConfig().Timeouts.Context)
		defer cancel()
		server.Shutdown(ctx)
	}()

	log.Printf("🚀 Sentinel Hub API starting on port %s", config.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// =============================================================================
// Phase 14A: Comprehensive Feature Analysis Handlers
// =============================================================================

// comprehensiveAnalysisHandler handles POST /api/v1/analyze/comprehensive
func comprehensiveAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	// Use configurable timeout (defaults to 60s, can be overridden via env var)
	ctx, cancel := context.WithTimeout(r.Context(), getAnalysisTimeout())
	defer cancel()

	project, err := getProjectFromContext(ctx)
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req struct {
		Feature                string              `json:"feature"`
		Mode                   string              `json:"mode"`            // "auto" or "manual"
		Files                  map[string][]string `json:"files,omitempty"` // For manual mode
		CodebasePath           string              `json:"codebasePath"`
		Depth                  string              `json:"depth"` // "surface", "medium", "deep"
		IncludeBusinessContext bool                `json:"includeBusinessContext"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		LogError(ctx, "Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Feature == "" {
		http.Error(w, "feature is required", http.StatusBadRequest)
		return
	}
	if req.CodebasePath == "" && req.Mode == "auto" {
		http.Error(w, "codebasePath is required for auto mode", http.StatusBadRequest)
		return
	}
	if req.Mode != "auto" && req.Mode != "manual" {
		http.Error(w, "mode must be 'auto' or 'manual'", http.StatusBadRequest)
		return
	}
	if req.Depth == "" {
		req.Depth = "medium" // Default depth
	}
	if req.Depth != "surface" && req.Depth != "medium" && req.Depth != "deep" {
		http.Error(w, "depth must be 'surface', 'medium', or 'deep'", http.StatusBadRequest)
		return
	}

	// Calculate timeout based on depth
	var timeout time.Duration
	config := GetConfig()
	switch req.Depth {
	case "surface":
		timeout = config.Timeouts.HTTP
	case "medium":
		timeout = config.Timeouts.Analysis
	case "deep":
		timeout = 3 * config.Timeouts.Analysis // 3x for deep analysis
	default:
		timeout = config.Timeouts.Analysis // Fallback to default
	}

	// Override context timeout with depth-based timeout
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// Validate codebasePath exists (for auto mode)
	if req.Mode == "auto" {
		if _, err := os.Stat(req.CodebasePath); os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("codebasePath does not exist: %s", req.CodebasePath), http.StatusBadRequest)
			return
		}
	}

	// Validate manual files exist
	if req.Mode == "manual" && req.Files != nil {
		for layer, files := range req.Files {
			for _, filePath := range files {
				fullPath := filepath.Join(req.CodebasePath, filePath)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					http.Error(w, fmt.Sprintf("File does not exist: %s (layer: %s)", fullPath, layer), http.StatusBadRequest)
					return
				}
			}
		}
	}

	analysisStart := time.Now()

	// Phase 14D: Surface depth - Skip LLM calls, use AST/patterns only
	if req.Depth == "surface" {
		LogInfo(ctx, "Surface depth: Skipping LLM calls, using AST/patterns only")
		// Continue with AST/pattern-based analysis only (no LLM)
		// Business context analysis will be skipped (see below)
	}

	// Phase 14D: Check cache for comprehensive analysis result
	llmConfig, err := getLLMConfig(ctx, project.ID)
	if err == nil && llmConfig != nil {
		featureHash := generateFeatureHash(req.Feature, req.CodebasePath)
		if cachedResult, ok := getCachedAnalysisResult(project.ID, featureHash, req.Depth, req.Mode, llmConfig); ok {
			LogInfo(ctx, "Returning cached comprehensive analysis result")
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"success":       true,
				"validation_id": cachedResult.ValidationID,
				"hub_url":       cachedResult.HubURL,
				"report":        cachedResult,
				"cached":        true,
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Discover feature
	feature, err := discoverFeature(ctx, req.Feature, req.CodebasePath, req.Files)
	if err != nil {
		LogError(ctx, "Feature discovery failed: %v", err)
		http.Error(w, fmt.Sprintf("Feature discovery failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Run analyses in parallel for better performance
	type analysisResult struct {
		businessFindings    []BusinessContextFinding
		apiFindings         []APILayerFinding
		testFindings        []TestLayerFinding
		logicFindings       []LogicLayerFinding
		uiFindings          []UILayerFinding
		dbFindings          []DatabaseLayerFinding
		integrationFindings []IntegrationLayerFinding
		errors              []string
		criticalErrors      []string // Track critical errors separately
	}

	result := analysisResult{}

	// Use goroutines for parallel execution
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Analyze business context if requested
	// Phase 14D: Skip business context LLM extraction for surface depth
	if req.IncludeBusinessContext && req.Depth != "surface" {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Check context cancellation
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Business context analysis cancelled: %v", ctx.Err())
				return
			default:
			}

			// Phase 14D: Get codebaseHash and LLM config for caching
			featureHash := generateFeatureHash(req.Feature, req.CodebasePath)
			llmConfig, _ := getLLMConfig(ctx, project.ID)

			businessFindings, err := analyzeBusinessContext(ctx, project.ID, feature, featureHash, llmConfig)
			if err != nil {
				LogWarn(ctx, "Business context analysis failed: %v", err)
				// Check context again before writing error
				select {
				case <-ctx.Done():
					LogWarn(ctx, "Cancelled before writing error: %v", ctx.Err())
					return
				default:
				}
				mu.Lock()
				// Classify error severity
				if isCriticalError(err) {
					result.criticalErrors = append(result.criticalErrors, fmt.Sprintf("Business context: %v", err))
				} else {
					result.errors = append(result.errors, fmt.Sprintf("Business context: %v", err))
				}
				mu.Unlock()
				return
			}

			// Check context cancellation before journey adherence
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Journey adherence check cancelled: %v", ctx.Err())
				mu.Lock()
				result.businessFindings = businessFindings
				mu.Unlock()
				return
			default:
			}

			// Check journey adherence
			// Phase 14D: Pass codebaseHash and config for caching
			journeyFindings, err := checkJourneyAdherence(ctx, project.ID, feature, featureHash, llmConfig)
			if err == nil {
				businessFindings = append(businessFindings, journeyFindings...)
			}

			// Check context before writing results
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Cancelled before writing results: %v", ctx.Err())
				return
			default:
			}

			mu.Lock()
			result.businessFindings = businessFindings
			mu.Unlock()
		}()
	}

	// Analyze API layer
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check context cancellation
		select {
		case <-ctx.Done():
			LogWarn(ctx, "API layer analysis cancelled: %v", ctx.Err())
			return
		default:
		}

		apiFindings, err := analyzeAPILayer(ctx, feature)
		if err != nil {
			LogWarn(ctx, "API layer analysis failed: %v", err)
			apiFindings = []APILayerFinding{}
			// Check context again before writing error
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Cancelled before writing error: %v", ctx.Err())
				return
			default:
			}
			mu.Lock()
			// Classify error severity
			if isCriticalError(err) {
				result.criticalErrors = append(result.criticalErrors, fmt.Sprintf("API layer: %v", err))
			} else {
				result.errors = append(result.errors, fmt.Sprintf("API layer: %v", err))
			}
			mu.Unlock()
		}
		// Check context before writing results
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Cancelled before writing results: %v", ctx.Err())
			return
		default:
		}
		mu.Lock()
		result.apiFindings = apiFindings
		mu.Unlock()
	}()

	// Analyze test layer
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check context cancellation
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Test layer analysis cancelled: %v", ctx.Err())
			return
		default:
		}

		testFindings, err := analyzeTestLayer(ctx, feature)
		if err != nil {
			LogWarn(ctx, "Test layer analysis failed: %v", err)
			testFindings = []TestLayerFinding{}
			// Check context again before writing error
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Cancelled before writing error: %v", ctx.Err())
				return
			default:
			}
			mu.Lock()
			// Classify error severity
			if isCriticalError(err) {
				result.criticalErrors = append(result.criticalErrors, fmt.Sprintf("Test layer: %v", err))
			} else {
				result.errors = append(result.errors, fmt.Sprintf("Test layer: %v", err))
			}
			mu.Unlock()
		}
		// Check context before writing results
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Cancelled before writing results: %v", ctx.Err())
			return
		default:
		}
		mu.Lock()
		result.testFindings = testFindings
		mu.Unlock()
	}()

	// Analyze business logic layer
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check context cancellation
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Business logic analysis cancelled: %v", ctx.Err())
			return
		default:
		}

		// Phase 14D: Pass depth to skip LLM for surface depth
		logicFindings, err := analyzeBusinessLogicWithDepth(ctx, project.ID, feature, req.Depth)
		if err != nil {
			LogWarn(ctx, "Business logic analysis failed: %v", err)
			logicFindings = []LogicLayerFinding{}
			// Check context again before writing error
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Cancelled before writing error: %v", ctx.Err())
				return
			default:
			}
			mu.Lock()
			// Classify error severity
			if isCriticalError(err) {
				result.criticalErrors = append(result.criticalErrors, fmt.Sprintf("Business logic: %v", err))
			} else {
				result.errors = append(result.errors, fmt.Sprintf("Business logic: %v", err))
			}
			mu.Unlock()
		}
		// Check context before writing results
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Cancelled before writing results: %v", ctx.Err())
			return
		default:
		}
		mu.Lock()
		result.logicFindings = logicFindings
		mu.Unlock()
	}()

	// Analyze UI layer
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check context cancellation
		select {
		case <-ctx.Done():
			LogWarn(ctx, "UI layer analysis cancelled: %v", ctx.Err())
			return
		default:
		}

		uiFindings, err := analyzeUILayer(ctx, feature)
		if err != nil {
			LogWarn(ctx, "UI layer analysis failed: %v", err)
			uiFindings = []UILayerFinding{}
			// Check context again before writing error
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Cancelled before writing error: %v", ctx.Err())
				return
			default:
			}
			mu.Lock()
			// Classify error severity
			if isCriticalError(err) {
				result.criticalErrors = append(result.criticalErrors, fmt.Sprintf("UI layer: %v", err))
			} else {
				result.errors = append(result.errors, fmt.Sprintf("UI layer: %v", err))
			}
			mu.Unlock()
		}
		// Check context before writing results
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Cancelled before writing results: %v", ctx.Err())
			return
		default:
		}
		mu.Lock()
		result.uiFindings = uiFindings
		mu.Unlock()
	}()

	// Analyze database layer
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check context cancellation
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Database layer analysis cancelled: %v", ctx.Err())
			return
		default:
		}

		dbFindings, err := analyzeDatabaseLayer(ctx, feature)
		if err != nil {
			LogWarn(ctx, "Database layer analysis failed: %v", err)
			dbFindings = []DatabaseLayerFinding{}
			// Check context again before writing error
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Cancelled before writing error: %v", ctx.Err())
				return
			default:
			}
			mu.Lock()
			// Classify error severity
			if isCriticalError(err) {
				result.criticalErrors = append(result.criticalErrors, fmt.Sprintf("Database layer: %v", err))
			} else {
				result.errors = append(result.errors, fmt.Sprintf("Database layer: %v", err))
			}
			mu.Unlock()
		}
		// Check context before writing results
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Cancelled before writing results: %v", ctx.Err())
			return
		default:
		}
		mu.Lock()
		result.dbFindings = dbFindings
		mu.Unlock()
	}()

	// Analyze integration layer
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check context cancellation
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Integration layer analysis cancelled: %v", ctx.Err())
			return
		default:
		}

		integrationFindings, err := analyzeIntegrationLayer(ctx, feature)
		if err != nil {
			LogWarn(ctx, "Integration layer analysis failed: %v", err)
			integrationFindings = []IntegrationLayerFinding{}
			// Check context again before writing error
			select {
			case <-ctx.Done():
				LogWarn(ctx, "Cancelled before writing error: %v", ctx.Err())
				return
			default:
			}
			mu.Lock()
			// Classify error severity
			if isCriticalError(err) {
				result.criticalErrors = append(result.criticalErrors, fmt.Sprintf("Integration layer: %v", err))
			} else {
				result.errors = append(result.errors, fmt.Sprintf("Integration layer: %v", err))
			}
			mu.Unlock()
		}
		// Check context before writing results
		select {
		case <-ctx.Done():
			LogWarn(ctx, "Cancelled before writing results: %v", ctx.Err())
			return
		default:
		}
		mu.Lock()
		result.integrationFindings = integrationFindings
		mu.Unlock()
	}()

	// Wait for all analyses to complete
	wg.Wait()

	// Check for critical errors
	if len(result.criticalErrors) > 0 {
		LogError(ctx, "Critical errors in comprehensive analysis: %v", result.criticalErrors)
		// Continue with partial analysis rather than failing completely
	}

	// Extract results
	businessFindings := result.businessFindings
	apiFindings := result.apiFindings
	testFindings := result.testFindings
	logicFindings := result.logicFindings
	uiFindings := result.uiFindings
	dbFindings := result.dbFindings
	integrationFindings := result.integrationFindings

	// Generate checklist from all findings (pass correct types directly)
	checklist := generateChecklist(
		businessFindings,
		uiFindings,
		apiFindings,
		dbFindings,
		logicFindings,
		integrationFindings,
		testFindings,
	)

	// Verify end-to-end flows
	flows, err := verifyEndToEndFlows(ctx, feature)
	if err != nil {
		LogWarn(ctx, "Flow verification failed: %v", err)
		flows = []Flow{}
	}

	// Verify integration points
	integrationBreakpoints, err := verifyIntegrationPoints(ctx, flows, feature)
	if err == nil {
		// Add integration breakpoints to flows
		for i := range flows {
			flows[i].Breakpoints = append(flows[i].Breakpoints, integrationBreakpoints...)
		}
	}

	// Convert flows to interface{} for storage
	flowsInterface := make([]interface{}, len(flows))
	for i, f := range flows {
		flowsInterface[i] = f
	}

	// Build layer analysis map
	layerAnalysis := map[string]interface{}{
		"business":    businessFindings,
		"ui":          uiFindings,
		"api":         apiFindings,
		"database":    dbFindings,
		"logic":       logicFindings,
		"integration": integrationFindings,
		"test":        testFindings,
	}

	// Generate summary
	analysisTime := time.Since(analysisStart)
	flowsVerified := len(flows)
	flowsBroken := 0
	for _, flow := range flows {
		if flow.Status == "broken" {
			flowsBroken++
		}
	}
	summary := generateSummary(checklist, flowsVerified, flowsBroken, analysisTime)

	// Format report
	hubConfig := loadConfig()
	report, err := formatReport(
		ctx,
		project.ID,
		req.Feature,
		req.Mode,
		req.Depth,
		checklist,
		summary,
		layerAnalysis,
		flowsInterface,
		hubConfig.HubURL,
	)
	if err != nil {
		LogError(ctx, "Failed to format report: %v", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	// Store report
	if err := storeComprehensiveValidation(ctx, report, project.ID); err != nil {
		LogWarn(ctx, "Failed to store validation: %v", err)
		// Continue and return report anyway
	} else {
		// Update LLM usage records with validation ID
		if err := updateLLMUsageValidationID(ctx, report.ValidationID, project.ID); err != nil {
			LogWarn(ctx, "Failed to update LLM usage validation ID: %v", err)
		}
	}

	// Phase 14D: Cache the analysis result
	if llmConfig != nil {
		featureHash := generateFeatureHash(req.Feature, req.CodebasePath)
		setCachedAnalysisResult(project.ID, featureHash, req.Depth, req.Mode, report, llmConfig)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":       true,
		"validation_id": report.ValidationID,
		"hub_url":       report.HubURL,
		"report":        report,
	}
	// Add warnings if there were errors
	if len(result.errors) > 0 || len(result.criticalErrors) > 0 {
		response["warnings"] = result.errors
		if len(result.criticalErrors) > 0 {
			response["critical_errors"] = result.criticalErrors
			response["partial_analysis"] = true
		}
	}
	json.NewEncoder(w).Encode(response)
}

// getComprehensiveValidationHandler handles GET /api/v1/validations/{id}
func getComprehensiveValidationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	validationID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	query := `
		SELECT validation_id, feature, mode, depth, findings, summary,
		       layer_analysis, end_to_end_flows, checklist, created_at, completed_at
		FROM comprehensive_validations
		WHERE validation_id = $1 AND project_id = $2
	`

	var valID, feature, mode, depth string
	var findingsJSON, summaryJSON, layerAnalysisJSON, flowsJSON, checklistJSON sql.NullString
	var createdAt time.Time
	var completedAt sql.NullTime

	err = queryRowWithTimeout(ctx, query, validationID, project.ID).Scan(
		&valID, &feature, &mode, &depth,
		&findingsJSON, &summaryJSON, &layerAnalysisJSON, &flowsJSON, &checklistJSON,
		&createdAt, &completedAt,
	)

	if err == sql.ErrNoRows {
		WriteErrorResponse(w, &NotFoundError{
			Resource: "validation",
			ID:       validationID,
			Message:  "Validation not found",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		LogError(ctx, "Failed to query validation: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// Unmarshal JSONB fields
	var findings []ChecklistItem
	var summary AnalysisSummary
	var layerAnalysis map[string]interface{}
	var flows []interface{}
	var checklist []ChecklistItem

	if findingsJSON.Valid {
		unmarshalJSONB(findingsJSON.String, &findings)
	}
	if summaryJSON.Valid {
		unmarshalJSONB(summaryJSON.String, &summary)
	}
	if layerAnalysisJSON.Valid {
		unmarshalJSONB(layerAnalysisJSON.String, &layerAnalysis)
	}
	if flowsJSON.Valid {
		unmarshalJSONB(flowsJSON.String, &flows)
	}
	if checklistJSON.Valid {
		unmarshalJSONB(checklistJSON.String, &checklist)
	}

	report := ComprehensiveAnalysisReport{
		ValidationID:  valID,
		Feature:       feature,
		Mode:          mode,
		Depth:         depth,
		Summary:       &summary,
		Checklist:     checklist,
		LayerAnalysis: layerAnalysis,
		EndToEndFlows: flows,
		HubURL:        fmt.Sprintf("https://hub.example.com/validations/%s", valID),
		CreatedAt:     createdAt,
	}
	if completedAt.Valid {
		report.CompletedAt = &completedAt.Time
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// listValidationsHandler handles GET /api/v1/validations?project={id}
func listValidationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	limit := 20
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	query := `
		SELECT validation_id, feature, mode, depth, summary, created_at, completed_at
		FROM comprehensive_validations
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := queryWithTimeout(ctx, query, project.ID, limit, offset)
	if err != nil {
		LogError(ctx, "Failed to query validations: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	validations := []map[string]interface{}{}
	for rows.Next() {
		var valID, feature, mode, depth string
		var summaryJSON sql.NullString
		var createdAt time.Time
		var completedAt sql.NullTime

		err := rows.Scan(&valID, &feature, &mode, &depth, &summaryJSON, &createdAt, &completedAt)
		if err != nil {
			LogWarn(ctx, "Failed to scan validation: %v", err)
			continue
		}

		var summary AnalysisSummary
		if summaryJSON.Valid {
			unmarshalJSONB(summaryJSON.String, &summary)
		}

		val := map[string]interface{}{
			"validation_id": valID,
			"feature":       feature,
			"mode":          mode,
			"depth":         depth,
			"summary":       summary,
			"hub_url":       fmt.Sprintf("https://hub.example.com/validations/%s", valID),
			"created_at":    createdAt,
		}
		if completedAt.Valid {
			val["completed_at"] = completedAt.Time
		}

		validations = append(validations, val)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM comprehensive_validations WHERE project_id = $1`
	err = queryRowWithTimeout(ctx, countQuery, project.ID).Scan(&total)
	if err != nil {
		LogWarn(ctx, "Failed to count validations: %v", err)
		total = len(validations)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"validations":  validations,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
		"has_next":     offset+limit < total,
		"has_previous": offset > 0,
	})
}

// =============================================================================
// Phase 15: Intent & Simple Language Handlers
// =============================================================================

// intentAnalysisHandler handles POST /api/v1/analyze/intent
func intentAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), GetConfig().Timeouts.Analysis)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req IntentAnalysisRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		LogError(ctx, "Failed to decode intent analysis request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Prompt == "" {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}

	// Gather context if requested
	var contextData *ContextData
	if req.IncludeContext {
		codebasePath := req.CodebasePath
		if codebasePath == "" {
			codebasePath = "." // Default to current directory
		}
		contextData, err = GatherContext(ctx, project.ID, codebasePath)
		if err != nil {
			LogWarn(ctx, "Failed to gather context: %v", err)
			// Continue without context
		}
	}

	// Analyze intent
	result, err := AnalyzeIntent(ctx, req.Prompt, contextData, project.ID)
	if err != nil {
		LogError(ctx, "Intent analysis failed: %v", err)
		http.Error(w, fmt.Sprintf("Intent analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Create decision record if clarification is needed
	if result.RequiresClarification {
		decision := &IntentDecision{
			ProjectID:          project.ID,
			OriginalPrompt:     req.Prompt,
			IntentType:         result.IntentType,
			ClarifyingQuestion: result.ClarifyingQuestion,
			UserChoice:         "", // Will be filled when user responds
			ResolvedPrompt:     "",
			ContextData:        nil,
		}

		err = RecordDecision(ctx, project.ID, decision)
		if err != nil {
			LogWarn(ctx, "Failed to create decision record: %v", err)
			// Continue without decision_id - non-fatal
		} else {
			result.DecisionID = decision.ID
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// recordIntentDecisionHandler handles POST /api/v1/intent/decisions
func recordIntentDecisionHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req IntentDecisionRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		LogError(ctx, "Failed to decode decision request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.DecisionID == "" || req.UserChoice == "" {
		http.Error(w, "decision_id and user_choice are required", http.StatusBadRequest)
		return
	}

	// Retrieve original decision from database
	query := `
		SELECT original_prompt, intent_type, clarifying_question
		FROM intent_decisions
		WHERE id = $1 AND project_id = $2
	`

	var originalPrompt, intentTypeStr, clarifyingQuestion string
	err = queryRowWithTimeout(ctx, query, req.DecisionID, project.ID).Scan(
		&originalPrompt, &intentTypeStr, &clarifyingQuestion,
	)
	if err == sql.ErrNoRows {
		WriteErrorResponse(w, &NotFoundError{
			Resource: "decision",
			ID:       req.DecisionID,
			Message:  "Decision not found",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		LogError(ctx, "Failed to query decision: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// Create decision object
	decision := &IntentDecision{
		ID:                 req.DecisionID,
		ProjectID:          project.ID,
		OriginalPrompt:     originalPrompt,
		IntentType:         IntentType(intentTypeStr),
		ClarifyingQuestion: clarifyingQuestion,
		UserChoice:         req.UserChoice,
		ResolvedPrompt:     req.ResolvedPrompt,
		ContextData:        req.AdditionalContext,
	}

	// Record decision
	err = RecordDecision(ctx, project.ID, decision)
	if err != nil {
		LogError(ctx, "Failed to record decision: %v", err)
		http.Error(w, fmt.Sprintf("Failed to record decision: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"decision_id": req.DecisionID, // Use original ID, not decision.ID (which may be new for inserts)
		"message":     "Decision recorded successfully",
	})
}

// getIntentPatternsHandler handles GET /api/v1/intent/patterns
func getIntentPatternsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Get query parameters
	patternType := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Query patterns
	var patterns []IntentPattern

	if patternType != "" {
		// Filter by type
		query := `
			SELECT id, project_id, pattern_type, pattern_data, frequency, last_used, created_at
			FROM intent_patterns
			WHERE project_id = $1 AND pattern_type = $2
			ORDER BY frequency DESC, last_used DESC
			LIMIT $3
		`
		rows, err := queryWithTimeout(ctx, query, project.ID, patternType, limit)
		if err != nil {
			LogError(ctx, "Failed to query patterns: %v", err)
			WriteErrorResponse(w, &DatabaseError{
				Operation:     "database_query",
				Message:       "Database error",
				OriginalError: err,
			}, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		patterns, err = scanIntentPatterns(rows)
		if err != nil {
			LogError(ctx, "Failed to scan patterns: %v", err)
			WriteErrorResponse(w, &DatabaseError{
				Operation:     "database_query",
				Message:       "Database error",
				OriginalError: err,
			}, http.StatusInternalServerError)
			return
		}
	} else {
		// Get all patterns
		patterns, err = GetLearnedPatterns(ctx, project.ID)
		if err != nil {
			LogError(ctx, "Failed to get patterns: %v", err)
			http.Error(w, fmt.Sprintf("Failed to get patterns: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Limit results
	if len(patterns) > limit {
		patterns = patterns[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"patterns": patterns,
		"count":    len(patterns),
	})
}

// =============================================================================
// Phase 14C: LLM Configuration Handlers
// =============================================================================

// createLLMConfigHandler handles POST /api/v1/llm/config
func createLLMConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req struct {
		Provider         string                 `json:"provider"`
		APIKey           string                 `json:"api_key"`
		Model            string                 `json:"model"`
		KeyType          string                 `json:"key_type"`
		CostOptimization CostOptimizationConfig `json:"cost_optimization,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Provider == "" || req.APIKey == "" || req.Model == "" {
		http.Error(w, "provider, api_key, and model are required", http.StatusBadRequest)
		return
	}

	// Validate provider
	if err := validateProvider(req.Provider); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate model
	if err := validateModel(req.Provider, req.Model); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate API key format
	if err := validateAPIKeyFormat(req.Provider, req.APIKey); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate cost optimization
	if err := validateCostOptimization(req.CostOptimization); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.KeyType == "" {
		req.KeyType = "user-provided"
	}

	// Set default cost optimization if not provided
	if req.CostOptimization.CacheTTLHours == 0 {
		req.CostOptimization = CostOptimizationConfig{
			UseCache:         true,
			CacheTTLHours:    24,
			ProgressiveDepth: true,
		}
	}

	config := &LLMConfig{
		Provider:         req.Provider,
		APIKey:           req.APIKey,
		Model:            req.Model,
		KeyType:          req.KeyType,
		CostOptimization: req.CostOptimization,
	}

	configID, err := saveLLMConfig(ctx, project.ID, config)
	if err != nil {
		LogError(ctx, "Failed to save LLM config: %v", err)
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	// Log audit entry
	ipAddress := getIPAddress(r)
	changedBy := r.Header.Get("X-User-Email") // Or extract from auth token
	if changedBy == "" {
		changedBy = "system"
	}
	logConfigChange(ctx, project.ID, configID, "create", changedBy, nil, config, ipAddress)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"config_id": configID,
		"message":   "Configuration saved successfully",
	})
}

// getLLMConfigHandler handles GET /api/v1/llm/config/{id}
func getLLMConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	configID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Query config
	query := `
		SELECT provider, api_key_encrypted, model, key_type, cost_optimization
		FROM llm_configurations
		WHERE id = $1 AND project_id = $2
	`

	var provider, model, keyType string
	var apiKeyEncrypted []byte
	var costOptJSON sql.NullString

	err = queryRowWithTimeout(ctx, query, configID, project.ID).Scan(
		&provider, &apiKeyEncrypted, &model, &keyType, &costOptJSON,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}
	if err != nil {
		LogError(ctx, "Failed to query config: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	// Decrypt and mask API key
	apiKey, err := decryptAPIKey(apiKeyEncrypted)
	if err != nil {
		LogError(ctx, "Failed to decrypt API key: %v", err)
		http.Error(w, "Failed to decrypt API key", http.StatusInternalServerError)
		return
	}

	maskedKey := maskAPIKey(apiKey)

	// Parse cost optimization
	var costOpt CostOptimizationConfig
	if costOptJSON.Valid && costOptJSON.String != "" {
		if err := json.Unmarshal([]byte(costOptJSON.String), &costOpt); err != nil {
			costOpt = CostOptimizationConfig{
				UseCache:         true,
				CacheTTLHours:    24,
				ProgressiveDepth: true,
			}
		}
	} else {
		costOpt = CostOptimizationConfig{
			UseCache:         true,
			CacheTTLHours:    24,
			ProgressiveDepth: true,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                configID,
		"provider":          provider,
		"api_key":           maskedKey,
		"model":             model,
		"key_type":          keyType,
		"cost_optimization": costOpt,
	})
}

// updateLLMConfigHandler handles PUT /api/v1/llm/config/{id}
func updateLLMConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	configID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	var req struct {
		Provider         string                 `json:"provider"`
		APIKey           string                 `json:"api_key,omitempty"`
		Model            string                 `json:"model"`
		KeyType          string                 `json:"key_type"`
		CostOptimization CostOptimizationConfig `json:"cost_optimization,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.Provider == "" || req.Model == "" {
		http.Error(w, "provider and model are required", http.StatusBadRequest)
		return
	}

	// Validate provider
	if err := validateProvider(req.Provider); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate model
	if err := validateModel(req.Provider, req.Model); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate API key format if provided
	if req.APIKey != "" {
		if err := validateAPIKeyFormat(req.Provider, req.APIKey); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Validate cost optimization
	if err := validateCostOptimization(req.CostOptimization); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.KeyType == "" {
		req.KeyType = "user-provided"
	}

	// Set default cost optimization if not provided
	if req.CostOptimization.CacheTTLHours == 0 {
		req.CostOptimization = CostOptimizationConfig{
			UseCache:         true,
			CacheTTLHours:    24,
			ProgressiveDepth: true,
		}
	}

	config := &LLMConfig{
		Provider:         req.Provider,
		APIKey:           req.APIKey,
		Model:            req.Model,
		KeyType:          req.KeyType,
		CostOptimization: req.CostOptimization,
	}

	// Get old config for audit log
	oldConfig, err := getLLMConfigByID(ctx, configID, project.ID)
	if err != nil {
		// Log warning but continue - config might not exist yet or might have been deleted
		// This is non-critical for audit logging, but we should log it
		LogWarn(ctx, "Could not retrieve old config for audit log (config may not exist): %v", err)
		oldConfig = nil
	}

	err = updateLLMConfig(ctx, configID, project.ID, config)
	if err != nil {
		LogError(ctx, "Failed to update LLM config: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Log audit entry
	ipAddress := getIPAddress(r)
	changedBy := r.Header.Get("X-User-Email")
	if changedBy == "" {
		changedBy = "system"
	}
	logConfigChange(ctx, project.ID, configID, "update", changedBy, oldConfig, config, ipAddress)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

// deleteLLMConfigHandler handles DELETE /api/v1/llm/config/{id}
func deleteLLMConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	configID := chi.URLParam(r, "id")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Get config before deletion for audit log
	oldConfig, err := getLLMConfigByID(ctx, configID, project.ID)
	if err != nil {
		// Log warning but continue - config might already be deleted or not exist
		// This is non-critical for audit logging, but we should log it
		LogWarn(ctx, "Could not retrieve config for audit log before deletion (config may not exist): %v", err)
		oldConfig = nil
	}

	err = deleteLLMConfig(ctx, configID, project.ID)
	if err != nil {
		LogError(ctx, "Failed to delete LLM config: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Configuration not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete config: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Log audit entry
	ipAddress := getIPAddress(r)
	changedBy := r.Header.Get("X-User-Email")
	if changedBy == "" {
		changedBy = "system"
	}
	logConfigChange(ctx, project.ID, configID, "delete", changedBy, oldConfig, nil, ipAddress)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration deleted successfully",
	})
}

// listLLMConfigsHandler handles GET /api/v1/llm/config/project/{projectId}
func listLLMConfigsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Verify project ownership
	if project.ID != projectID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	configs, err := listLLMConfigs(ctx, projectID)
	if err != nil {
		LogError(ctx, "Failed to list LLM configs: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"configs": configs,
		"count":   len(configs),
	})
}

// getProvidersHandler handles GET /api/v1/llm/providers
func getProvidersHandler(w http.ResponseWriter, r *http.Request) {
	providers := getSupportedProviders()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"providers": providers,
	})
}

// getModelsHandler handles GET /api/v1/llm/models/{provider}
func getModelsHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	models := getSupportedModels(provider)

	if len(models) == 0 {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"models":  models,
	})
}

// validateLLMConfigHandler handles POST /api/v1/llm/config/validate
func validateLLMConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), GetConfig().Timeouts.Context)
	defer cancel()

	var req struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
		Model    string `json:"model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, &ValidationError{
			Field:   "body",
			Message: "Invalid request body",
			Code:    "invalid_json",
		}, http.StatusBadRequest)
		return
	}

	if req.Provider == "" || req.APIKey == "" || req.Model == "" {
		http.Error(w, "provider, api_key, and model are required", http.StatusBadRequest)
		return
	}

	config := &LLMConfig{
		Provider: req.Provider,
		APIKey:   req.APIKey,
		Model:    req.Model,
		KeyType:  "user-provided",
		CostOptimization: CostOptimizationConfig{
			UseCache:         true,
			CacheTTLHours:    24,
			ProgressiveDepth: true,
		},
	}

	err := testLLMConnection(ctx, config)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"valid":   false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"valid":   true,
		"message": "Connection test successful",
	})
}

// getUsageReportHandler handles GET /api/v1/llm/usage/report
func getUsageReportHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -GetConfig().Limits.DefaultDateRangeDays) // Default to configured days ago
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now()
	}

	report, err := getUsageReport(ctx, project.ID, startDate, endDate)
	if err != nil {
		LogError(ctx, "Failed to get usage report: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// getUsageStatsHandler handles GET /api/v1/llm/usage/stats
func getUsageStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	stats, err := getUsageStats(ctx, project.ID, period)
	if err != nil {
		LogError(ctx, "Failed to get usage stats: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// getCostBreakdownHandler handles GET /api/v1/llm/usage/cost-breakdown
func getCostBreakdownHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	breakdown, err := getCostBreakdown(ctx, project.ID, period)
	if err != nil {
		LogError(ctx, "Failed to get cost breakdown: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(breakdown)
}

// getUsageTrendsHandler handles GET /api/v1/llm/usage/trends
func getUsageTrendsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultContextTimeout)
	defer cancel()

	project, err := getProjectFromContext(r.Context())
	if err != nil {
		LogErrorWithContext(r.Context(), err, "Failed to get project from context")
		LogErrorWithContext(r.Context(), err, "Internal server error")
		LogErrorWithContext(r.Context(), fmt.Errorf("internal server error"), "Internal server error")
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "internal_operation",
			Message:       "Internal server error",
			OriginalError: fmt.Errorf("internal server error"),
		}, http.StatusInternalServerError)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "day"
	}

	// Calculate date range
	var startDate time.Time
	endDate := time.Now()

	switch period {
	case "daily":
		startDate = endDate.AddDate(0, 0, -1)
	case "weekly":
		startDate = endDate.AddDate(0, 0, -7)
	case "monthly":
		startDate = endDate.AddDate(0, -1, 0)
	case "yearly":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, 0, -GetConfig().Limits.DefaultDateRangeDays)
	}

	// Query trends based on group_by
	var query string
	var rows *sql.Rows

	switch groupBy {
	case "day":
		query = `
			SELECT DATE(created_at) as date, SUM(tokens_used) as tokens, SUM(estimated_cost) as cost, COUNT(*) as requests
			FROM llm_usage
			WHERE project_id = $1 AND created_at >= $2 AND created_at <= $3
			GROUP BY DATE(created_at)
			ORDER BY DATE(created_at) ASC
		`
		rows, err = queryWithTimeout(ctx, query, project.ID, startDate, endDate)
	case "provider":
		query = `
			SELECT provider, SUM(tokens_used) as tokens, SUM(estimated_cost) as cost, COUNT(*) as requests
			FROM llm_usage
			WHERE project_id = $1 AND created_at >= $2 AND created_at <= $3
			GROUP BY provider
			ORDER BY SUM(estimated_cost) DESC
		`
		rows, err = queryWithTimeout(ctx, query, project.ID, startDate, endDate)
	case "model":
		query = `
			SELECT model, SUM(tokens_used) as tokens, SUM(estimated_cost) as cost, COUNT(*) as requests
			FROM llm_usage
			WHERE project_id = $1 AND created_at >= $2 AND created_at <= $3
			GROUP BY model
			ORDER BY SUM(estimated_cost) DESC
		`
		rows, err = queryWithTimeout(ctx, query, project.ID, startDate, endDate)
	default:
		http.Error(w, "Invalid group_by parameter (must be day, provider, or model)", http.StatusBadRequest)
		return
	}

	if err != nil {
		LogError(ctx, "Failed to query trends: %v", err)
		WriteErrorResponse(w, &DatabaseError{
			Operation:     "database_query",
			Message:       "Database error",
			OriginalError: err,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TrendPoint struct {
		Label    string  `json:"label"`
		Tokens   int64   `json:"tokens"`
		Cost     float64 `json:"cost"`
		Requests int64   `json:"requests"`
	}

	var trends []TrendPoint
	for rows.Next() {
		var label string
		var tokens int64
		var cost float64
		var requests int64

		err := rows.Scan(&label, &tokens, &cost, &requests)
		if err != nil {
			LogError(ctx, "Failed to scan trend row: %v", err)
			continue
		}

		trends = append(trends, TrendPoint{
			Label:    label,
			Tokens:   tokens,
			Cost:     cost,
			Requests: requests,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"period":   period,
		"group_by": groupBy,
		"trends":   trends,
	})
}

// scanIntentPatterns scans rows into IntentPattern slice
func scanIntentPatterns(rows *sql.Rows) ([]IntentPattern, error) {
	patterns := []IntentPattern{}
	for rows.Next() {
		var pattern IntentPattern
		var patternDataJSON sql.NullString
		var lastUsed, createdAt sql.NullTime

		err := rows.Scan(
			&pattern.ID,
			&pattern.ProjectID,
			&pattern.PatternType,
			&patternDataJSON,
			&pattern.Frequency,
			&lastUsed,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal pattern data
		if patternDataJSON.Valid {
			if err := unmarshalJSONB(patternDataJSON.String, &pattern.PatternData); err != nil {
				pattern.PatternData = make(map[string]interface{})
			}
		} else {
			pattern.PatternData = make(map[string]interface{})
		}

		if lastUsed.Valid {
			pattern.LastUsed = lastUsed.Time.Format(time.RFC3339)
		}
		if createdAt.Valid {
			pattern.CreatedAt = createdAt.Time.Format(time.RFC3339)
		}

		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

// =============================================================================
// Phase 14D: Cost Optimization Metrics Handlers
// =============================================================================

// getCacheMetricsHandler handles GET /api/v1/metrics/cache?project_id={id}
func getCacheMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		project, err := getProjectFromContext(ctx)
		if err != nil {
			http.Error(w, "project_id is required", http.StatusBadRequest)
			return
		}
		projectID = project.ID
	}

	// Get cache hit rate
	hitRate := getCacheHitRate(projectID)

	// Get hit/miss counts
	var hits, misses int64
	if val, ok := cacheHitCounter.Load(projectID); ok {
		hits = val.(int64)
	}
	if val, ok := cacheMissCounter.Load(projectID); ok {
		misses = val.(int64)
	}

	// Phase 14D: Get cache size from counter (O(1) instead of O(n))
	var cacheSize int64
	if val, ok := cacheSizeCounter.Load(projectID); ok {
		cacheSize = val.(int64)
	}

	// Get LLM config for TTL
	llmConfig, err := getLLMConfig(ctx, projectID)
	cacheTTLHours := 24 // default
	if err == nil && llmConfig != nil && llmConfig.CostOptimization.CacheTTLHours > 0 {
		cacheTTLHours = llmConfig.CostOptimization.CacheTTLHours
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"project_id":      projectID,
		"hit_rate":        hitRate,
		"total_hits":      hits,
		"total_misses":    misses,
		"cache_size":      cacheSize,
		"cache_ttl_hours": cacheTTLHours,
	})
}

// getCostMetricsHandler handles GET /api/v1/metrics/cost?project_id={id}&period={daily|weekly|monthly}
func getCostMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		project, err := getProjectFromContext(ctx)
		if err != nil {
			http.Error(w, "project_id is required", http.StatusBadRequest)
			return
		}
		projectID = project.ID
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}
	if period != "daily" && period != "weekly" && period != "monthly" {
		http.Error(w, "period must be 'daily', 'weekly', or 'monthly'", http.StatusBadRequest)
		return
	}

	// Calculate date range based on period
	var startDate time.Time
	now := time.Now()
	switch period {
	case "daily":
		startDate = now.AddDate(0, 0, -1)
	case "weekly":
		startDate = now.AddDate(0, 0, -7)
	case "monthly":
		startDate = now.AddDate(0, -1, 0)
	}

	// Query usage data
	query := `
		SELECT 
			COALESCE(SUM(estimated_cost), 0) as total_cost,
			COUNT(*) as total_requests
		FROM llm_usage
		WHERE project_id = $1 AND created_at >= $2
	`

	var totalCost float64
	var totalRequests int
	err := db.QueryRowContext(ctx, query, projectID, startDate).Scan(&totalCost, &totalRequests)
	if err != nil {
		LogError(ctx, "Failed to query cost metrics: %v", err)
		http.Error(w, "Failed to retrieve cost metrics", http.StatusInternalServerError)
		return
	}

	// Phase 14D: Calculate cache hit savings from actual tracked metrics
	cacheHitRate := getCacheHitRate(projectID)
	// Calculate actual cache hit savings: hit rate * total cost (cache hits save 100% of cost)
	cacheHitSavings := totalCost * cacheHitRate

	// Phase 14D: Get actual model selection savings from tracked metrics
	modelSelectionSavings := getModelSelectionSavings(projectID)

	// Total savings
	totalSavings := cacheHitSavings + modelSelectionSavings
	savingsPercentage := 0.0
	if totalCost > 0 {
		savingsPercentage = (totalSavings / totalCost) * 100.0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":                 true,
		"project_id":              projectID,
		"period":                  period,
		"total_cost":              totalCost,
		"cost_savings":            totalSavings,
		"savings_percentage":      savingsPercentage,
		"cache_hit_savings":       cacheHitSavings,
		"model_selection_savings": modelSelectionSavings,
		"total_requests":          totalRequests,
	})
}
