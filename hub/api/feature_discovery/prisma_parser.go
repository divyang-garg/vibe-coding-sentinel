// Package feature_discovery provides Prisma schema parsing
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package feature_discovery

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// discoverPrismaTables discovers Prisma schema tables with relationships
func discoverPrismaTables(codebasePath string, featureName string) ([]TableInfo, []RelationshipInfo) {
	tables := []TableInfo{}
	relationships := []RelationshipInfo{}

	schemaPath := filepath.Join(codebasePath, "prisma", "schema.prisma")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return tables, relationships
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var currentTable *TableInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "model ") {
			// Save previous table if exists
			if currentTable != nil {
				tables = append(tables, *currentTable)
			}

			// Extract model name
			modelName := strings.TrimSpace(strings.TrimPrefix(line, "model "))
			modelName = strings.Fields(modelName)[0]

			if matchesFeature(modelName, featureName) {
				currentTable = &TableInfo{
					Name:     modelName,
					Source:   "prisma",
					File:     schemaPath,
					Columns:  []ColumnInfo{},
					Indexes:  []IndexInfo{},
					Metadata: make(map[string]string),
				}
			} else {
				currentTable = nil
			}

		} else if currentTable != nil {
			// Parse table contents
			if strings.Contains(line, "@@") {
				// Table-level directives (indexes, unique constraints)
				if strings.Contains(line, "@@index") {
					index := parsePrismaIndex(line)
					if index != nil {
						currentTable.Indexes = append(currentTable.Indexes, *index)
					}
				} else if strings.Contains(line, "@@unique") {
					index := parsePrismaUnique(line)
					if index != nil {
						currentTable.Indexes = append(currentTable.Indexes, *index)
					}
				}
			} else if strings.Contains(line, "  ") && !strings.HasPrefix(line, "@@") && !strings.HasPrefix(line, "}") {
				// Check for relationships first
				if strings.Contains(line, "@relation") {
					// Extract field name
					parts := strings.Fields(strings.TrimSpace(line))
					if len(parts) > 0 {
						fieldName := parts[0]
						if rel := parsePrismaRelationship(line, currentTable.Name, fieldName); rel != nil {
							relationships = append(relationships, *rel)
						}
					}
				} else {
					// Column definition
					column := parsePrismaColumn(line)
					if column != nil {
						currentTable.Columns = append(currentTable.Columns, *column)
					}
				}
			} else if strings.HasPrefix(line, "}") && currentTable != nil {
				// End of model
				tables = append(tables, *currentTable)
				currentTable = nil
			}
		}
	}

	// Add final table if exists
	if currentTable != nil {
		tables = append(tables, *currentTable)
	}

	return tables, relationships
}

// parsePrismaColumn parses a Prisma column definition
func parsePrismaColumn(line string) *ColumnInfo {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "//") {
		return nil
	}

	// Extract column name and type
	re := regexp.MustCompile(`(\w+)\s+([^@\s]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	column := &ColumnInfo{
		Name:     matches[1],
		Type:     matches[2],
		Nullable: strings.Contains(line, "?"),
		Metadata: make(map[string]string),
	}

	// Check for modifiers
	if strings.Contains(line, "@id") {
		column.PrimaryKey = true
	}
	if strings.Contains(line, "@unique") {
		column.Unique = true
	}
	if strings.Contains(line, "@default") {
		// Extract default value
		defaultRe := regexp.MustCompile(`@default\(([^)]+)\)`)
		if defaultMatch := defaultRe.FindStringSubmatch(line); len(defaultMatch) > 1 {
			column.DefaultValue = defaultMatch[1]
		}
	}

	// Check for auto-increment
	if strings.Contains(line, "@default(autoincrement())") {
		column.AutoIncrement = true
	}

	// Parse type details (length, precision, etc.)
	column.Type = parsePrismaType(column.Type)

	return column
}

// parsePrismaRelationship parses relationship annotations
func parsePrismaRelationship(line string, sourceTable string, sourceColumn string) *RelationshipInfo {
	// Check for @relation annotation
	if !strings.Contains(line, "@relation") {
		return nil
	}

	relation := &RelationshipInfo{
		SourceTable:  sourceTable,
		SourceColumn: sourceColumn,
		Metadata:     make(map[string]string),
	}

	// Parse relation details
	re := regexp.MustCompile(`@relation\(([^)]+)\)`)
	if match := re.FindStringSubmatch(line); len(match) > 1 {
		params := match[1]

		// Extract fields
		if fieldsMatch := regexp.MustCompile(`fields:\s*\[([^\]]+)\]`).FindStringSubmatch(params); len(fieldsMatch) > 1 {
			relation.SourceColumn = strings.Trim(fieldsMatch[1], "\"'")
		}

		if referencesMatch := regexp.MustCompile(`references:\s*\[([^\]]+)\]`).FindStringSubmatch(params); len(referencesMatch) > 1 {
			relation.TargetColumn = strings.Trim(referencesMatch[1], "\"'")
		}

		// Determine relationship type
		if strings.Contains(params, "fields:") && strings.Contains(params, "references:") {
			relation.Type = "many-to-one" // Foreign key relationship
			relation.SourceCardinality = "*"
			relation.TargetCardinality = "1"
		}
	}

	// Extract target table from type annotation
	typeRe := regexp.MustCompile(`(\w+)\[\]|\w+`)
	if typeMatch := typeRe.FindStringSubmatch(line); len(typeMatch) > 1 {
		relation.TargetTable = typeMatch[1]
	}

	if relation.TargetTable != "" && relation.Type != "" {
		return relation
	}

	return nil
}

// parsePrismaIndex parses @@index directive
func parsePrismaIndex(line string) *IndexInfo {
	re := regexp.MustCompile(`@@index\(\[([^\]]+)\](?:,\s*name:\s*['"]([^'"]+)['"])?\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}

	columnsStr := matches[1]
	name := "unnamed_index"
	if len(matches) > 2 && matches[2] != "" {
		name = matches[2]
	}

	columns := parseColumnList(columnsStr)

	return &IndexInfo{
		Name:    name,
		Columns: columns,
		Unique:  false,
		Type:    "BTREE",
	}
}

// parsePrismaUnique parses @@unique directive
func parsePrismaUnique(line string) *IndexInfo {
	re := regexp.MustCompile(`@@unique\(\[([^\]]+)\](?:,\s*name:\s*['"]([^'"]+)['"])?\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}

	columnsStr := matches[1]
	name := "unnamed_unique"
	if len(matches) > 2 && matches[2] != "" {
		name = matches[2]
	}

	columns := parseColumnList(columnsStr)

	return &IndexInfo{
		Name:    name,
		Columns: columns,
		Unique:  true,
		Type:    "BTREE",
	}
}
