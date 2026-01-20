// Package feature_discovery provides shared types and utilities for schema parsing
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package feature_discovery

import (
	"regexp"
	"strings"
)

// parseColumnList parses a comma-separated list of columns
func parseColumnList(columnsStr string) []string {
	columns := []string{}
	columnNames := strings.Split(columnsStr, ",")

	for _, col := range columnNames {
		col = strings.TrimSpace(strings.Trim(col, "\"'"))
		if col != "" {
			columns = append(columns, col)
		}
	}

	return columns
}

// parsePrismaType parses Prisma type annotations and converts to SQL types
func parsePrismaType(typeStr string) string {
	// Handle parameterized types like String(255), Decimal(10,2)
	re := regexp.MustCompile(`(\w+)\(([^)]+)\)`)
	if matches := re.FindStringSubmatch(typeStr); len(matches) >= 3 {
		baseType := matches[1]
		params := matches[2]

		switch baseType {
		case "String":
			return "VARCHAR(" + params + ")"
		case "Decimal":
			return "DECIMAL(" + params + ")"
		default:
			return strings.ToUpper(baseType) + "(" + params + ")"
		}
	}

	// Standard type mapping
	switch strings.ToLower(typeStr) {
	case "string":
		return "VARCHAR"
	case "int":
		return "INTEGER"
	case "bigint":
		return "BIGINT"
	case "float":
		return "FLOAT"
	case "decimal":
		return "DECIMAL"
	case "boolean":
		return "BOOLEAN"
	case "datetime":
		return "TIMESTAMP"
	case "date":
		return "DATE"
	case "json":
		return "JSON"
	default:
		return strings.ToUpper(typeStr)
	}
}

// parseTypeScriptType converts TypeScript types to SQL types
func parseTypeScriptType(tsType string) string {
	tsType = strings.TrimSpace(strings.ToLower(tsType))

	switch tsType {
	case "string":
		return "VARCHAR"
	case "number":
		return "INTEGER"
	case "boolean":
		return "BOOLEAN"
	case "date":
		return "DATE"
	default:
		return strings.ToUpper(tsType)
	}
}

// parseSQLType parses SQL type definitions and normalizes them
func parseSQLType(sqlType string) string {
	// Handle parameterized types
	if strings.Contains(sqlType, "(") {
		return sqlType
	}

	// Standard type mappings
	switch sqlType {
	case "VARCHAR", "TEXT", "CHAR", "NVARCHAR":
		return sqlType + "(255)" // Default length
	case "INT", "INTEGER":
		return "INTEGER"
	case "BIGINT":
		return "BIGINT"
	case "FLOAT", "DOUBLE":
		return "FLOAT"
	case "DECIMAL", "NUMERIC":
		return "DECIMAL(10,2)"
	case "BOOLEAN", "BOOL":
		return "BOOLEAN"
	case "DATE":
		return "DATE"
	case "DATETIME", "TIMESTAMP":
		return "TIMESTAMP"
	case "JSON", "JSONB":
		return "JSON"
	default:
		return sqlType
	}
}
