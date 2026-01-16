// Package feature_discovery provides tests for ORM detection
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package feature_discovery

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectDatabaseORM_Prismatests Prisma detection
func TestDetectDatabaseORM_Prismat(t *testing.T) {
	// Create temporary directory with Prisma schema
	tempDir := t.TempDir()
	prismaDir := filepath.Join(tempDir, "prisma")
	err := os.MkdirAll(prismaDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create prisma directory: %v", err)
	}

	schemaPath := filepath.Join(prismaDir, "schema.prisma")
	schemaContent := `model User { id Int @id @default(autoincrement()) }`
	err = os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create schema file: %v", err)
	}

	// Test detection
	orm, err := detectDatabaseORM(tempDir)
	if err != nil {
		t.Fatalf("detectDatabaseORM returned error: %v", err)
	}
	if orm != "prisma" {
		t.Errorf("Expected prisma, got %s", orm)
	}
}

// TestDetectDatabaseORM_TypeORM tests TypeORM detection
func TestDetectDatabaseORM_TypeORM(t *testing.T) {
	tempDir := t.TempDir()
	entityPath := filepath.Join(tempDir, "user.entity.ts")
	entityContent := `@Entity() export class User { @PrimaryGeneratedColumn() id: number; }`
	err := os.WriteFile(entityPath, []byte(entityContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create entity file: %v", err)
	}

	orm, err := detectDatabaseORM(tempDir)
	if err != nil {
		t.Fatalf("detectDatabaseORM returned error: %v", err)
	}
	if orm != "typeorm" {
		t.Errorf("Expected typeorm, got %s", orm)
	}
}

// TestDetectDatabaseORM_SQLAlchemy tests SQLAlchemy detection
func TestDetectDatabaseORM_SQLAlchemy(t *testing.T) {
	tempDir := t.TempDir()
	reqPath := filepath.Join(tempDir, "requirements.txt")
	reqContent := `sqlalchemy==1.4.0
flask==2.0.0`
	err := os.WriteFile(reqPath, []byte(reqContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create requirements file: %v", err)
	}

	orm, err := detectDatabaseORM(tempDir)
	if err != nil {
		t.Fatalf("detectDatabaseORM returned error: %v", err)
	}
	if orm != "sqlalchemy" {
		t.Errorf("Expected sqlalchemy, got %s", orm)
	}
}

// TestDetectDatabaseORM_GORM tests GORM detection
func TestDetectDatabaseORM_GORM(t *testing.T) {
	tempDir := t.TempDir()
	goFile := filepath.Join(tempDir, "main.go")
	goContent := `package main
import "gorm.io/gorm"
func main() { db, _ := gorm.Open() }`
	err := os.WriteFile(goFile, []byte(goContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go file: %v", err)
	}

	orm, err := detectDatabaseORM(tempDir)
	if err != nil {
		t.Fatalf("detectDatabaseORM returned error: %v", err)
	}
	if orm != "gorm" {
		t.Errorf("Expected gorm, got %s", orm)
	}
}

// TestDetectDatabaseORM_Unknown tests unknown ORM detection
func TestDetectDatabaseORM_Unknown(t *testing.T) {
	tempDir := t.TempDir()

	orm, err := detectDatabaseORM(tempDir)
	if err != nil {
		t.Fatalf("detectDatabaseORM returned error: %v", err)
	}
	if orm != "unknown" {
		t.Errorf("Expected unknown, got %s", orm)
	}
}
