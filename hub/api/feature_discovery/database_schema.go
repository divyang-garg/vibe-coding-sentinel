// Package feature_discovery provides comprehensive database schema analysis
// Complies with CODING_STANDARDS.md: Database schema max 300 lines
package feature_discovery

import (
	"context"
)

// discoverDatabaseTables discovers database tables and relationships
// Supports Prisma, TypeORM, SQL migrations with comprehensive analysis
// This is the main orchestrator function that delegates to specific parsers
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
