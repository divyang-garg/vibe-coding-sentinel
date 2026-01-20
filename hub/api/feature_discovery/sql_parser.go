// Package feature_discovery provides SQL migration parsing
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package feature_discovery

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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
