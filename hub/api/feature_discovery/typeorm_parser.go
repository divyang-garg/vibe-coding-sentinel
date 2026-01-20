// Package feature_discovery provides TypeORM entity parsing
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package feature_discovery

import (
	"os"
	"regexp"
	"strings"
)

// discoverTypeORMTables discovers TypeORM entity tables with relationships
func discoverTypeORMTables(codebasePath string, featureName string) ([]TableInfo, []RelationshipInfo) {
	tables := []TableInfo{}
	relationships := []RelationshipInfo{}

	entityFiles, _ := findFilesRecursively(codebasePath, "*.entity.ts")

	for _, file := range entityFiles {
		if isExcludedPath(file) {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)

		// Look for @Entity decorator
		if strings.Contains(content, "@Entity") {
			table, rels := parseTypeORMEntity(content, file, featureName)
			if table != nil {
				tables = append(tables, *table)
				relationships = append(relationships, rels...)
			}
		}
	}

	return tables, relationships
}

// parseTypeORMEntity parses a TypeORM entity class
func parseTypeORMEntity(content string, filePath string, featureName string) (*TableInfo, []RelationshipInfo) {
	tableName := extractTypeORMTableName(content)
	if tableName == "" || !matchesFeature(tableName, featureName) {
		return nil, nil
	}

	table := &TableInfo{
		Name:     tableName,
		Source:   "typeorm",
		File:     filePath,
		Columns:  []ColumnInfo{},
		Indexes:  []IndexInfo{},
		Metadata: make(map[string]string),
	}

	relationships := []RelationshipInfo{}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "@Column") || strings.Contains(line, "@PrimaryColumn") || strings.Contains(line, "@PrimaryGeneratedColumn") {
			column := parseTypeORMColumn(line)
			if column != nil {
				table.Columns = append(table.Columns, *column)
			}
		} else if strings.Contains(line, "@OneToMany") || strings.Contains(line, "@ManyToOne") || strings.Contains(line, "@OneToOne") || strings.Contains(line, "@ManyToMany") {
			rel := parseTypeORMRelationship(line, tableName)
			if rel != nil {
				relationships = append(relationships, *rel)
			}
		} else if strings.Contains(line, "@Index") {
			index := parseTypeORMIndex(line)
			if index != nil {
				table.Indexes = append(table.Indexes, *index)
			}
		}
	}

	return table, relationships
}

// extractTypeORMTableName extracts table name from TypeORM entity
func extractTypeORMTableName(content string) string {
	// Check @Entity decorator for table name - options object format
	entityRe := regexp.MustCompile(`@Entity\(\{[^}]*table:\s*['"]([^'"]+)['"]`)
	if match := entityRe.FindStringSubmatch(content); len(match) > 1 {
		return match[1]
	}

	// Check @Entity decorator for table name - string format
	entityStringRe := regexp.MustCompile(`@Entity\(\s*['"]([^'"]+)['"]`)
	if match := entityStringRe.FindStringSubmatch(content); len(match) > 1 {
		return match[1]
	}

	// Fallback: extract class name
	classRe := regexp.MustCompile(`export class (\w+)`)
	if match := classRe.FindStringSubmatch(content); len(match) > 1 {
		return match[1]
	}

	return ""
}

// parseTypeORMColumn parses a TypeORM column decorator
func parseTypeORMColumn(line string) *ColumnInfo {
	// Extract property name and type
	propRe := regexp.MustCompile(`(\w+):\s*([^;]+);`)
	propMatch := propRe.FindStringSubmatch(line)
	if len(propMatch) < 3 {
		return nil
	}

	column := &ColumnInfo{
		Name:     propMatch[1],
		Type:     parseTypeScriptType(propMatch[2]),
		Nullable: strings.Contains(line, "nullable: true"),
		Metadata: make(map[string]string),
	}

	// Check for decorators
	if strings.Contains(line, "@PrimaryColumn") || strings.Contains(line, "@PrimaryGeneratedColumn") {
		column.PrimaryKey = true
		if strings.Contains(line, "@PrimaryGeneratedColumn") {
			column.AutoIncrement = true
		}
	}

	return column
}

// parseTypeORMRelationship parses TypeORM relationship decorators
func parseTypeORMRelationship(line string, sourceTable string) *RelationshipInfo {
	relation := &RelationshipInfo{
		SourceTable: sourceTable,
		Metadata:    make(map[string]string),
	}

	if strings.Contains(line, "@OneToMany") {
		relation.Type = "one-to-many"
		relation.SourceCardinality = "1"
		relation.TargetCardinality = "*"
	} else if strings.Contains(line, "@ManyToOne") {
		relation.Type = "many-to-one"
		relation.SourceCardinality = "*"
		relation.TargetCardinality = "1"
	} else if strings.Contains(line, "@OneToOne") {
		relation.Type = "one-to-one"
		relation.SourceCardinality = "1"
		relation.TargetCardinality = "1"
	} else if strings.Contains(line, "@ManyToMany") {
		relation.Type = "many-to-many"
		relation.SourceCardinality = "*"
		relation.TargetCardinality = "*"
	}

	// Extract target type
	typeRe := regexp.MustCompile(`:\s*(\w+)`)
	if typeMatch := typeRe.FindStringSubmatch(line); len(typeMatch) > 1 {
		relation.TargetTable = typeMatch[1]
	}

	return relation
}

// parseTypeORMIndex parses TypeORM @Index decorator
func parseTypeORMIndex(line string) *IndexInfo {
	// This is a simplified implementation
	return &IndexInfo{
		Name:   "typeorm_index",
		Unique: strings.Contains(line, "isUnique: true"),
		Type:   "BTREE",
	}
}
