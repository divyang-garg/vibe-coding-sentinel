// Phase 14A: Database Layer Analyzer
// Analyzes database schema for constraints, indexes, and data integrity

package services

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// DatabaseLayerFinding represents a finding from database layer analysis
type DatabaseLayerFinding struct {
	Type     string `json:"type"`     // "missing_constraint", "missing_index", "data_integrity_issue"
	Location string `json:"location"` // Schema file or migration file
	Issue    string `json:"issue"`
	Severity string `json:"severity"` // "critical", "high", "medium", "low"
}

// analyzeDatabaseLayer analyzes database schema
func analyzeDatabaseLayer(ctx context.Context, feature *DiscoveredFeature) ([]DatabaseLayerFinding, error) {
	findings := []DatabaseLayerFinding{}

	if feature.DatabaseLayer == nil {
		return findings, nil
	}

	ormType := feature.DatabaseLayer.ORMType

	// Analyze based on ORM type
	switch ormType {
	case "prisma":
		findings = analyzePrismaSchema(ctx, feature.DatabaseLayer)
	case "typeorm":
		findings = analyzeTypeORMSchema(ctx, feature.DatabaseLayer)
	case "raw_sql":
		findings = analyzeSQLSchema(ctx, feature.DatabaseLayer)
	default:
		// Try all methods
		prismaFindings := analyzePrismaSchema(ctx, feature.DatabaseLayer)
		typeormFindings := analyzeTypeORMSchema(ctx, feature.DatabaseLayer)
		sqlFindings := analyzeSQLSchema(ctx, feature.DatabaseLayer)
		findings = append(findings, prismaFindings...)
		findings = append(findings, typeormFindings...)
		findings = append(findings, sqlFindings...)
	}

	return findings, nil
}

// analyzeSchema checks for foreign keys, indexes, and constraints
func analyzeSchema(tables []TableInfo) []DatabaseLayerFinding {
	findings := []DatabaseLayerFinding{}

	for _, table := range tables {
		// Check for primary key
		hasPrimaryKey := false
		for _, col := range table.Columns {
			if col.PrimaryKey {
				hasPrimaryKey = true
				break
			}
		}

		if !hasPrimaryKey {
			findings = append(findings, DatabaseLayerFinding{
				Type:     "missing_constraint",
				Location: table.File,
				Issue:    fmt.Sprintf("Table %s is missing a primary key", table.Name),
				Severity: "critical",
			})
		}

		// Check for foreign key relationships
		if len(table.Relationships) == 0 && strings.Contains(strings.ToLower(table.Name), "id") {
			// Potential foreign key missing (heuristic)
			findings = append(findings, DatabaseLayerFinding{
				Type:     "missing_constraint",
				Location: table.File,
				Issue:    fmt.Sprintf("Table %s may be missing foreign key constraints", table.Name),
				Severity: "medium",
			})
		}

		// Check for nullable columns that should not be nullable
		for _, col := range table.Columns {
			if col.Nullable && (strings.Contains(strings.ToLower(col.Name), "id") ||
				strings.Contains(strings.ToLower(col.Name), "email") ||
				strings.Contains(strings.ToLower(col.Name), "name")) {
				findings = append(findings, DatabaseLayerFinding{
					Type:     "data_integrity_issue",
					Location: table.File,
					Issue:    fmt.Sprintf("Table %s column %s is nullable but may need to be required", table.Name, col.Name),
					Severity: "medium",
				})
			}
		}
	}

	return findings
}

// analyzePrismaSchema analyzes Prisma schema files
func analyzePrismaSchema(ctx context.Context, dbLayer *DatabaseLayerTables) []DatabaseLayerFinding {
	findings := []DatabaseLayerFinding{}

	// Find Prisma schema file
	schemaFile := ""
	for _, table := range dbLayer.Tables {
		if table.Source == "prisma" && table.File != "" {
			schemaFile = table.File
			break
		}
	}

	if schemaFile == "" {
		return findings
	}

	// Read schema file
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		LogWarn(ctx, "Failed to read Prisma schema file %s: %v", schemaFile, err)
		return findings
	}

	content := string(data)

	// Check for relationships
	for _, table := range dbLayer.Tables {
		if len(table.Relationships) == 0 {
			// Check if table name appears in relation definitions
			hasRelation := strings.Contains(content, fmt.Sprintf("model %s", table.Name)) &&
				strings.Contains(content, "@relation")

			if !hasRelation && !strings.Contains(strings.ToLower(table.Name), "join") {
				findings = append(findings, DatabaseLayerFinding{
					Type:     "missing_constraint",
					Location: schemaFile,
					Issue:    fmt.Sprintf("Table %s may be missing relationship definitions", table.Name),
					Severity: "medium",
				})
			}
		}
	}

	// Run general schema analysis
	schemaFindings := analyzeSchema(dbLayer.Tables)
	findings = append(findings, schemaFindings...)

	return findings
}

// analyzeTypeORMSchema analyzes TypeORM entity files
func analyzeTypeORMSchema(ctx context.Context, dbLayer *DatabaseLayerTables) []DatabaseLayerFinding {
	findings := []DatabaseLayerFinding{}

	// Analyze each entity file
	for _, table := range dbLayer.Tables {
		if table.Source == "typeorm" && table.File != "" {
			data, err := os.ReadFile(table.File)
			if err != nil {
				LogWarn(ctx, "Failed to read TypeORM entity file %s: %v", table.File, err)
				continue
			}

			content := string(data)

			// Check for @OneToMany, @ManyToOne, @ManyToMany decorators
			hasRelations := strings.Contains(content, "@OneToMany") ||
				strings.Contains(content, "@ManyToOne") ||
				strings.Contains(content, "@ManyToMany") ||
				strings.Contains(content, "@OneToOne")

			if !hasRelations && len(table.Relationships) == 0 {
				findings = append(findings, DatabaseLayerFinding{
					Type:     "missing_constraint",
					Location: table.File,
					Issue:    fmt.Sprintf("Entity %s may be missing relationship decorators", table.Name),
					Severity: "medium",
				})
			}

			// Check for @PrimaryGeneratedColumn or @PrimaryColumn
			hasPrimaryKey := strings.Contains(content, "@PrimaryGeneratedColumn") ||
				strings.Contains(content, "@PrimaryColumn")

			if !hasPrimaryKey {
				findings = append(findings, DatabaseLayerFinding{
					Type:     "missing_constraint",
					Location: table.File,
					Issue:    fmt.Sprintf("Entity %s is missing primary key decorator", table.Name),
					Severity: "critical",
				})
			}

			// Check for @Index decorators
			hasIndexes := strings.Contains(content, "@Index")

			// Check if columns that might need indexes exist
			if !hasIndexes && strings.Contains(content, "@Column") {
				// Check for common indexed columns (email, username, etc.)
				if strings.Contains(content, "email") || strings.Contains(content, "username") {
					findings = append(findings, DatabaseLayerFinding{
						Type:     "missing_index",
						Location: table.File,
						Issue:    fmt.Sprintf("Entity %s may benefit from indexes on frequently queried columns", table.Name),
						Severity: "low",
					})
				}
			}
		}
	}

	// Run general schema analysis
	schemaFindings := analyzeSchema(dbLayer.Tables)
	findings = append(findings, schemaFindings...)

	return findings
}

// analyzeSQLSchema analyzes raw SQL migration files
func analyzeSQLSchema(ctx context.Context, dbLayer *DatabaseLayerTables) []DatabaseLayerFinding {
	findings := []DatabaseLayerFinding{}

	// Analyze each migration file
	for _, table := range dbLayer.Tables {
		if table.Source == "migration" && table.File != "" {
			data, err := os.ReadFile(table.File)
			if err != nil {
				LogWarn(ctx, "Failed to read SQL migration file %s: %v", table.File, err)
				continue
			}

			content := strings.ToUpper(string(data))

			// Check for PRIMARY KEY
			if !strings.Contains(content, "PRIMARY KEY") && strings.Contains(content, fmt.Sprintf("CREATE TABLE %s", strings.ToUpper(table.Name))) {
				findings = append(findings, DatabaseLayerFinding{
					Type:     "missing_constraint",
					Location: table.File,
					Issue:    fmt.Sprintf("Table %s is missing PRIMARY KEY constraint", table.Name),
					Severity: "critical",
				})
			}

			// Check for FOREIGN KEY
			if !strings.Contains(content, "FOREIGN KEY") && !strings.Contains(strings.ToLower(table.Name), "join") {
				findings = append(findings, DatabaseLayerFinding{
					Type:     "missing_constraint",
					Location: table.File,
					Issue:    fmt.Sprintf("Table %s may be missing FOREIGN KEY constraints", table.Name),
					Severity: "medium",
				})
			}

			// Check for indexes
			if !strings.Contains(content, "CREATE INDEX") && !strings.Contains(content, "CREATE UNIQUE INDEX") {
				// Check for columns that typically need indexes
				if strings.Contains(content, "EMAIL") || strings.Contains(content, "USERNAME") {
					findings = append(findings, DatabaseLayerFinding{
						Type:     "missing_index",
						Location: table.File,
						Issue:    fmt.Sprintf("Table %s may benefit from indexes on frequently queried columns", table.Name),
						Severity: "low",
					})
				}
			}
		}
	}

	// Run general schema analysis
	schemaFindings := analyzeSchema(dbLayer.Tables)
	findings = append(findings, schemaFindings...)

	return findings
}
