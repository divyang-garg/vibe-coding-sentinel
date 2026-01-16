// Package feature_discovery provides comprehensive framework detection
// Complies with CODING_STANDARDS.md: ORM detection max 250 lines
package feature_discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// findFilesRecursively finds files recursively matching a pattern
func findFilesRecursively(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			matched, err := filepath.Match(pattern, info.Name())
			if err != nil {
				return nil
			}
			if matched {
				matches = append(matches, path)
			}
		}
		return nil
	})
	return matches, err
}

// detectDatabaseORM detects the database ORM used in the codebase
// Supports SQLAlchemy, Django ORM, GORM, Prisma, TypeORM, Mongoose, Sequelize
func detectDatabaseORM(codebasePath string) (string, error) {
	// Check for Prisma
	prismaSchemaPath := filepath.Join(codebasePath, "prisma", "schema.prisma")
	if _, err := os.Stat(prismaSchemaPath); err == nil {
		return "prisma", nil
	}

	// Check for TypeORM entities (TypeScript files with @Entity decorator)
	tsFiles, _ := findFilesRecursively(codebasePath, "*.entity.ts")
	if len(tsFiles) > 0 {
		return "typeorm", nil
	}

	// Check for Sequelize models
	packageJSONPath := filepath.Join(codebasePath, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
					if _, ok := deps["sequelize"]; ok {
						return "sequelize", nil
					}
				}
			}
		}
	}

	// Check for Mongoose (MongoDB ODM)
	if _, err := os.Stat(packageJSONPath); err == nil {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
					if _, ok := deps["mongoose"]; ok {
						return "mongoose", nil
					}
				}
			}
		}
	}

	// Check for SQLAlchemy (Python)
	requirementsPath := filepath.Join(codebasePath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		data, err := os.ReadFile(requirementsPath)
		if err == nil {
			content := string(data)
			if strings.Contains(content, "sqlalchemy") {
				return "sqlalchemy", nil
			}
		}
	}

	// Check for Django ORM
	if _, err := os.Stat(requirementsPath); err == nil {
		data, err := os.ReadFile(requirementsPath)
		if err == nil {
			content := string(data)
			if strings.Contains(content, "django") {
				return "django_orm", nil
			}
		}
	}

	// Check for GORM (Go)
	goFiles, _ := findFilesRecursively(codebasePath, "*.go")
	if len(goFiles) > 0 {
		for _, file := range goFiles {
			data, err := os.ReadFile(file)
			if err == nil {
				content := string(data)
				if strings.Contains(content, "gorm.io/gorm") || strings.Contains(content, "github.com/jinzhu/gorm") {
					return "gorm", nil
				}
			}
		}
	}

	// Check for SQL migrations (indicates raw SQL or ORM)
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
			if len(sqlFiles) > 0 || len(pyFiles) > 0 {
				return "raw_sql", nil
			}
		}
	}

	return "unknown", nil
}
