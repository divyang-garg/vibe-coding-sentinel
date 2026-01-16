// Package feature_discovery provides tests for database schema analysis
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package feature_discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestDiscoverDatabaseTables_Prismatests Prisma schema analysis
func TestDiscoverDatabaseTables_Prismat(t *testing.T) {
	tempDir := t.TempDir()

	// Create Prisma schema directory
	prismaDir := filepath.Join(tempDir, "prisma")
	err := os.MkdirAll(prismaDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create prisma directory: %v", err)
	}

	// Create schema.prisma file
	schemaContent := `model User {
  id        Int      @id @default(autoincrement())
  email     String   @unique
  name      String
  posts     Post[]
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  @@index([email])
  @@map("users")
}

model Post {
  id        Int      @id @default(autoincrement())
  title     String
  content   String?
  published Boolean  @default(false)
  authorId  Int
  author    User     @relation(fields: [authorId], references: [id], onDelete: Cascade)
  tags      Tag[]

  @@index([authorId])
  @@map("posts")
}

model Tag {
  id    Int    @id @default(autoincrement())
  name  String @unique
  posts Post[]

  @@map("tags")
}`

	schemaPath := filepath.Join(prismaDir, "schema.prisma")
	err = os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create schema file: %v", err)
	}

	// Test database table discovery
	result, err := discoverDatabaseTables(context.Background(), tempDir, "", "prisma")
	if err != nil {
		t.Fatalf("discoverDatabaseTables returned error: %v", err)
	}

	// Debug: check what was found
	t.Logf("Found %d tables", len(result.Tables))
	for _, table := range result.Tables {
		t.Logf("Table: %s, Columns: %d, Indexes: %d", table.Name, len(table.Columns), len(table.Indexes))
	}
	t.Logf("Found %d relationships", len(result.Relationships))

	if result.ORMType != "prisma" {
		t.Errorf("Expected ORM type 'prisma', got '%s'", result.ORMType)
	}

	if len(result.Tables) == 0 {
		t.Errorf("Expected to find tables, but found none")
	}

	// Check for User table
	foundUser := false
	for _, table := range result.Tables {
		if table.Name == "User" {
			foundUser = true

			// Check columns
			if len(table.Columns) < 3 {
				t.Errorf("Expected User table to have at least 3 columns, got %d", len(table.Columns))
			}

			// Check for primary key
			foundID := false
			for _, col := range table.Columns {
				if col.Name == "id" && col.PrimaryKey {
					foundID = true
					if !col.AutoIncrement {
						t.Errorf("Expected id column to be auto-increment")
					}
				}
				if col.Name == "email" && !col.Unique {
					t.Errorf("Expected email column to be unique")
				}
			}
			if !foundID {
				t.Errorf("Expected to find id column as primary key")
			}

			// Check indexes
			if len(table.Indexes) == 0 {
				t.Errorf("Expected User table to have indexes")
			}

			break
		}
	}

	if !foundUser {
		t.Errorf("Expected to find User table")
	}

	// TODO: Add relationship parsing tests once relationship detection is fully implemented
	// For now, focus on table and column detection
	_ = result.Relationships // Avoid unused variable warning
}

// TestDiscoverDatabaseTables_TypeORM tests TypeORM entity analysis
func TestDiscoverDatabaseTables_TypeORM(t *testing.T) {
	tempDir := t.TempDir()

	// Create TypeORM entity file
	entityContent := `import { Entity, PrimaryGeneratedColumn, Column, OneToMany, ManyToOne } from 'typeorm';
import { Post } from './Post';

@Entity('users')
export class User {
  @PrimaryGeneratedColumn()
  id: number;

  @Column({ unique: true })
  email: string;

  @Column()
  name: string;

  @Column({ type: 'timestamp', default: () => 'CURRENT_TIMESTAMP' })
  createdAt: Date;

  @OneToMany(() => Post, post => post.author)
  posts: Post[];
}`

	entityPath := filepath.Join(tempDir, "user.entity.ts")
	err := os.WriteFile(entityPath, []byte(entityContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create entity file: %v", err)
	}

	// Test database table discovery
	result, err := discoverDatabaseTables(context.Background(), tempDir, "user", "typeorm")
	if err != nil {
		t.Fatalf("discoverDatabaseTables returned error: %v", err)
	}

	if len(result.Tables) == 0 {
		t.Errorf("Expected to find tables, but found none")
	}

	// Check for users table
	foundUsers := false
	for _, table := range result.Tables {
		if table.Name == "users" {
			foundUsers = true
			break
		}
	}

	if !foundUsers {
		t.Errorf("Expected to find users table")
		// Debug: show what tables were found
		for _, table := range result.Tables {
			t.Logf("Found table: %s", table.Name)
		}
	}
}

// TestDiscoverDatabaseTables_SQL tests SQL migration analysis
func TestDiscoverDatabaseTables_SQL(t *testing.T) {
	tempDir := t.TempDir()

	// Create migrations directory
	migrationsDir := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Create SQL migration file
	sqlContent := `CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title VARCHAR(255) NOT NULL,
  content TEXT,
  user_id INTEGER NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_posts_user_id ON posts(user_id);`

	sqlPath := filepath.Join(migrationsDir, "001_initial.sql")
	err = os.WriteFile(sqlPath, []byte(sqlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create SQL migration file: %v", err)
	}

	// Test database table discovery
	result, err := discoverDatabaseTables(context.Background(), tempDir, "user", "raw_sql")
	if err != nil {
		t.Fatalf("discoverDatabaseTables returned error: %v", err)
	}

	if len(result.Tables) == 0 {
		t.Errorf("Expected to find tables, but found none")
	}

	// Check for users table
	foundUsers := false
	for _, table := range result.Tables {
		if table.Name == "users" {
			foundUsers = true
			break
		}
	}

	if !foundUsers {
		t.Errorf("Expected to find users table")
		// Debug: show what tables were found
		for _, table := range result.Tables {
			t.Logf("Found table: %s", table.Name)
		}
	}
}

// TestDiscoverDatabaseTables_AutoDetect tests automatic ORM detection
func TestDiscoverDatabaseTables_AutoDetect(t *testing.T) {
	tempDir := t.TempDir()

	// Create both Prisma and TypeORM files
	prismaDir := filepath.Join(tempDir, "prisma")
	err := os.MkdirAll(prismaDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create prisma directory: %v", err)
	}

	schemaContent := `model Product {
  id    Int    @id @default(autoincrement())
  name  String
  price Float
}`

	err = os.WriteFile(filepath.Join(prismaDir, "schema.prisma"), []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create schema file: %v", err)
	}

	entityContent := `import { Entity, PrimaryGeneratedColumn, Column } from 'typeorm';

@Entity('categories')
export class Category {
  @PrimaryGeneratedColumn()
  id: number;

  @Column()
  name: string;
}`

	err = os.WriteFile(filepath.Join(tempDir, "Category.ts"), []byte(entityContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create entity file: %v", err)
	}

	// Test with auto-detect (empty ormType)
	result, err := discoverDatabaseTables(context.Background(), tempDir, "", "")
	if err != nil {
		t.Fatalf("discoverDatabaseTables returned error: %v", err)
	}

	// Should find at least the Prisma Product table
	if len(result.Tables) == 0 {
		t.Errorf("Expected to find at least 1 table, got %d", len(result.Tables))
	}

	foundProduct := false
	for _, table := range result.Tables {
		if table.Name == "Product" {
			foundProduct = true
			break
		}
	}

	if !foundProduct {
		t.Errorf("Expected to find Product table from Prisma")
		// Debug: show what tables were found
		for _, table := range result.Tables {
			t.Logf("Found table: %s", table.Name)
		}
	}
}

// TestDiscoverDatabaseTables_NoTables tests case with no matching tables
func TestDiscoverDatabaseTables_NoTables(t *testing.T) {
	tempDir := t.TempDir()

	// Create a non-database file
	randomFile := filepath.Join(tempDir, "utils.js")
	utilsContent := `export const formatDate = (date) => {
  return date.toISOString();
};`

	err := os.WriteFile(randomFile, []byte(utilsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create utils file: %v", err)
	}

	// Test database table discovery with non-matching feature name
	result, err := discoverDatabaseTables(context.Background(), tempDir, "nonexistent", "prisma")
	if err != nil {
		t.Fatalf("discoverDatabaseTables returned error: %v", err)
	}

	if len(result.Tables) != 0 {
		t.Errorf("Expected to find no tables, but found %d", len(result.Tables))
	}
}
