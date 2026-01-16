// Package feature_discovery provides comprehensive database schema analysis
// Complies with CODING_STANDARDS.md: Database schema max 300 lines
package feature_discovery

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// discoverDatabaseTables discovers database tables and relationships
// Supports Prisma, TypeORM, SQL migrations with comprehensive analysis
func discoverDatabaseTables(ctx context.Context, codebasePath string, featureName string, ormType string) (*DatabaseLayerTables, error) {
	tables := []TableInfo{}
	relationships := []RelationshipInfo{}
	constraints := []ConstraintInfo{}

	switch ormType {
	case "prisma":
		tables, relationships = discoverPrismaTables(codebasePath, featureName)
	case "typeorm":
		tables, relationships = discoverTypeORMTables(codebasePath, featureName)
	case "raw_sql":
		tables, relationships, constraints = discoverSQLTables(codebasePath, featureName)
	default:
		// Try all methods for comprehensive discovery
		if prismaTables, prismaRels := discoverPrismaTables(codebasePath, featureName); len(prismaTables) > 0 {
			tables = append(tables, prismaTables...)
			relationships = append(relationships, prismaRels...)
		}
		if typeormTables, typeormRels := discoverTypeORMTables(codebasePath, featureName); len(typeormTables) > 0 {
			tables = append(tables, typeormTables...)
			relationships = append(relationships, typeormRels...)
		}
		if sqlTables, sqlRels, sqlConstraints := discoverSQLTables(codebasePath, featureName); len(sqlTables) > 0 {
			tables = append(tables, sqlTables...)
			relationships = append(relationships, sqlRels...)
			constraints = append(constraints, sqlConstraints...)
		}
	}

	return &DatabaseLayerTables{
		Tables:        tables,
		Relationships: relationships,
		Constraints:   constraints,
		ORMType:       ormType,
	}, nil
}

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
						} else {
							// Debug: relationship parsing failed
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

// parsePrismaType parses Prisma type annotations
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

// discoverSQLTables discovers tables from SQL migration files
func discoverSQLTables(codebasePath string, featureName string) ([]TableInfo, []RelationshipInfo, []ConstraintInfo) {
	tables := []TableInfo{}
	relationships := []RelationshipInfo{}
	constraints := []ConstraintInfo{}

	migrationDirs := []string{
		filepath.Join(codebasePath, "migrations"),
		filepath.Join(codebasePath, "db", "migrations"),
		filepath.Join(codebasePath, "database", "migrations"),
		filepath.Join(codebasePath, "migrations", "versions"),
	}

	for _, dir := range migrationDirs {
		if _, err := os.Stat(dir); err == nil {
			sqlFiles, _ := filepath.Glob(filepath.Join(dir, "*.sql"))
			pyFiles, _ := filepath.Glob(filepath.Join(dir, "*.py"))

			allFiles := append(sqlFiles, pyFiles...)

			for _, file := range allFiles {
				data, err := os.ReadFile(file)
				if err != nil {
					continue
				}

				content := string(data)

				// Parse CREATE TABLE statements
				fileTables, fileRels, fileConstraints := parseSQLContent(content, file, featureName)
				tables = append(tables, fileTables...)
				relationships = append(relationships, fileRels...)
				constraints = append(constraints, fileConstraints...)
			}
		}
	}

	return tables, relationships, constraints
}

// parseSQLContent parses SQL content for tables, relationships, and constraints
func parseSQLContent(content string, filePath string, featureName string) ([]TableInfo, []RelationshipInfo, []ConstraintInfo) {
	tables := []TableInfo{}
	relationships := []RelationshipInfo{}
	constraints := []ConstraintInfo{}

	lines := strings.Split(content, "\n")
	var currentTable *TableInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		lineUpper := strings.ToUpper(line)

		if strings.Contains(lineUpper, "CREATE TABLE") {
			// Save previous table
			if currentTable != nil {
				tables = append(tables, *currentTable)
			}

			// Extract table name
			tableName := extractSQLTableName(line)
			if tableName != "" && matchesFeature(tableName, featureName) {
				currentTable = &TableInfo{
					Name:     tableName,
					Source:   "migration",
					File:     filePath,
					Columns:  []ColumnInfo{},
					Indexes:  []IndexInfo{},
					Metadata: make(map[string]string),
				}
			} else {
				currentTable = nil
			}

		} else if currentTable != nil {
			if strings.Contains(lineUpper, "PRIMARY KEY") {
				constraint := parseSQLPrimaryKey(line, currentTable.Name)
				if constraint != nil {
					constraints = append(constraints, *constraint)
				}
			} else if strings.Contains(lineUpper, "FOREIGN KEY") {
				rel, constraint := parseSQLForeignKey(line, currentTable.Name)
				if rel != nil {
					relationships = append(relationships, *rel)
				}
				if constraint != nil {
					constraints = append(constraints, *constraint)
				}
			} else if strings.Contains(lineUpper, "UNIQUE") {
				constraint := parseSQLUnique(line, currentTable.Name)
				if constraint != nil {
					constraints = append(constraints, *constraint)
				}
			} else if strings.Contains(line, "`") || strings.Contains(line, "\"") {
				// Potential column definition
				column := parseSQLColumn(line)
				if column != nil {
					currentTable.Columns = append(currentTable.Columns, *column)
				}
			} else if strings.Contains(line, ");") || strings.Contains(lineUpper, "ENGINE=") {
				// End of table definition
				tables = append(tables, *currentTable)
				currentTable = nil
			}
		}
	}

	// Add final table
	if currentTable != nil {
		tables = append(tables, *currentTable)
	}

	return tables, relationships, constraints
}

// extractSQLTableName extracts table name from CREATE TABLE statement
func extractSQLTableName(line string) string {
	// Handle various formats: CREATE TABLE `table`, CREATE TABLE "table", CREATE TABLE table
	re := regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:` + "`" + `([^` + "`" + `]+)` + "`" + `|'([^']+)'|"([^"]+)"|(\w+))`)
	matches := re.FindStringSubmatch(line)

	for i := 1; i < len(matches); i++ {
		if matches[i] != "" {
			return matches[i]
		}
	}

	return ""
}

// parseSQLColumn parses a SQL column definition
func parseSQLColumn(line string) *ColumnInfo {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "//") {
		return nil
	}

	// Extract column name and type
	re := regexp.MustCompile(`(` + "`" + `[^` + "`" + `]+` + "`" + `|["'][^"']+["']|\w+)\s+([^\s,(]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	columnName := strings.Trim(matches[1], "`\"'")
	columnType := strings.ToUpper(matches[2])

	column := &ColumnInfo{
		Name:     columnName,
		Type:     parseSQLType(columnType),
		Nullable: !strings.Contains(strings.ToUpper(line), "NOT NULL"),
		Metadata: make(map[string]string),
	}

	// Check for PRIMARY KEY
	if strings.Contains(strings.ToUpper(line), "PRIMARY KEY") {
		column.PrimaryKey = true
	}

	// Check for AUTO_INCREMENT/AUTOINCREMENT
	if strings.Contains(strings.ToUpper(line), "AUTO_INCREMENT") || strings.Contains(strings.ToUpper(line), "AUTOINCREMENT") {
		column.AutoIncrement = true
	}

	// Check for UNIQUE
	if strings.Contains(strings.ToUpper(line), "UNIQUE") {
		column.Unique = true
	}

	// Extract default value
	if defaultRe := regexp.MustCompile(`DEFAULT\s+(['"]([^'"]+)['"]|\w+|\d+)`); defaultRe.MatchString(line) {
		if match := defaultRe.FindStringSubmatch(line); len(match) > 1 {
			column.DefaultValue = match[1]
		}
	}

	return column
}

// parseSQLType parses SQL type definitions
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

// parseSQLPrimaryKey parses PRIMARY KEY constraints
func parseSQLPrimaryKey(line string, tableName string) *ConstraintInfo {
	re := regexp.MustCompile(`PRIMARY\s+KEY\s*\(([^)]+)\)`)
	if match := re.FindStringSubmatch(line); len(match) > 1 {
		columnsStr := match[1]
		columns := parseColumnList(columnsStr)

		return &ConstraintInfo{
			Name:    "PRIMARY",
			Type:    "PRIMARY KEY",
			Table:   tableName,
			Columns: columns,
		}
	}

	return nil
}

// parseSQLForeignKey parses FOREIGN KEY constraints
func parseSQLForeignKey(line string, tableName string) (*RelationshipInfo, *ConstraintInfo) {
	// This is a simplified implementation
	re := regexp.MustCompile(`FOREIGN\s+KEY\s*\(([^)]+)\)\s*REFERENCES\s+(\w+)\s*\(([^)]+)\)`)
	if match := re.FindStringSubmatch(line); len(match) >= 4 {
		sourceColumns := parseColumnList(match[1])
		targetTable := match[2]
		targetColumns := parseColumnList(match[3])

		if len(sourceColumns) > 0 && len(targetColumns) > 0 {
			relationship := &RelationshipInfo{
				Type:              "many-to-one",
				SourceTable:       tableName,
				SourceColumn:      sourceColumns[0],
				TargetTable:       targetTable,
				TargetColumn:      targetColumns[0],
				SourceCardinality: "*",
				TargetCardinality: "1",
			}

			constraint := &ConstraintInfo{
				Name:    "FOREIGN_KEY_" + sourceColumns[0],
				Type:    "FOREIGN KEY",
				Table:   tableName,
				Columns: sourceColumns,
			}

			return relationship, constraint
		}
	}

	return nil, nil
}

// parseSQLUnique parses UNIQUE constraints
func parseSQLUnique(line string, tableName string) *ConstraintInfo {
	re := regexp.MustCompile(`UNIQUE\s*\(([^)]+)\)`)
	if match := re.FindStringSubmatch(line); len(match) > 1 {
		columnsStr := match[1]
		columns := parseColumnList(columnsStr)

		return &ConstraintInfo{
			Name:    "UNIQUE_" + strings.Join(columns, "_"),
			Type:    "UNIQUE",
			Table:   tableName,
			Columns: columns,
		}
	}

	return nil
}
