package main

import (
	"fmt"
	"regexp"
	"strings"
)

func extractTypeORMTableName(content string) string {
	// Check @Entity decorator for table name
	entityRe := regexp.MustCompile(`@Entity\(\{[^}]*table:\s*['"]([^'"]+)['"]`)
	if match := entityRe.FindStringSubmatch(content); len(match) > 1 {
		fmt.Printf("Found table name in decorator: %s\n", match[1])
		return match[1]
	}

	// Fallback: extract class name
	classRe := regexp.MustCompile(`export class (\w+)`)
	if match := classRe.FindStringSubmatch(content); len(match) > 1 {
		fmt.Printf("Found class name: %s\n", match[1])
		return match[1]
	}

	fmt.Println("No table name found")
	return ""
}

func main() {
	content := `import { Entity, PrimaryGeneratedColumn, Column } from 'typeorm';

@Entity('users')
export class Category {
  @PrimaryGeneratedColumn()
  id: number;

  @Column()
  name: string;
}`

	tableName := extractTypeORMTableName(content)
	fmt.Printf("Final table name: %s\n", tableName)
}
EOF && go run debug_typeorm.go && rm debug_typeorm.go