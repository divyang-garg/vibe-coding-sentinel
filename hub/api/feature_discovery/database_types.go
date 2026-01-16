// Package feature_discovery provides database schema analysis types
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package feature_discovery

// DatabaseLayerTables represents discovered database tables and relationships
type DatabaseLayerTables struct {
	Tables        []TableInfo        `json:"tables"`
	Relationships []RelationshipInfo `json:"relationships,omitempty"`
	ORMType       string             `json:"orm_type,omitempty"`
	Constraints   []ConstraintInfo   `json:"constraints,omitempty"`
}

// TableInfo contains comprehensive information about a database table
type TableInfo struct {
	Name          string             `json:"name"`
	Schema        string             `json:"schema,omitempty"` // Database schema name
	Columns       []ColumnInfo       `json:"columns,omitempty"`
	Indexes       []IndexInfo        `json:"indexes,omitempty"`
	Relationships []RelationshipInfo `json:"relationships,omitempty"`
	Source        string             `json:"source"` // "migration", "prisma", "typeorm", "sql"
	File          string             `json:"file,omitempty"`
	Metadata      map[string]string  `json:"metadata,omitempty"`
}

// ColumnInfo contains detailed information about a database column
type ColumnInfo struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Length        int               `json:"length,omitempty"`    // For VARCHAR, etc.
	Precision     int               `json:"precision,omitempty"` // For DECIMAL
	Scale         int               `json:"scale,omitempty"`     // For DECIMAL
	Nullable      bool              `json:"nullable"`
	DefaultValue  string            `json:"default_value,omitempty"`
	PrimaryKey    bool              `json:"primary_key,omitempty"`
	Unique        bool              `json:"unique,omitempty"`
	AutoIncrement bool              `json:"auto_increment,omitempty"`
	ForeignKey    *ForeignKeyInfo   `json:"foreign_key,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// ForeignKeyInfo contains foreign key relationship information
type ForeignKeyInfo struct {
	ReferencedTable  string `json:"referenced_table"`
	ReferencedColumn string `json:"referenced_column"`
	OnDelete         string `json:"on_delete,omitempty"` // CASCADE, SET NULL, etc.
	OnUpdate         string `json:"on_update,omitempty"`
}

// IndexInfo contains information about database indexes
type IndexInfo struct {
	Name     string            `json:"name"`
	Columns  []string          `json:"columns"`
	Unique   bool              `json:"unique"`
	Type     string            `json:"type,omitempty"` // BTREE, HASH, etc.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RelationshipInfo contains information about table relationships
type RelationshipInfo struct {
	Type              string            `json:"type"` // "one-to-one", "one-to-many", "many-to-one", "many-to-many"
	SourceTable       string            `json:"source_table"`
	SourceColumn      string            `json:"source_column,omitempty"`
	TargetTable       string            `json:"target_table"`
	TargetColumn      string            `json:"target_column,omitempty"`
	ForeignKeyName    string            `json:"foreign_key_name,omitempty"`
	JoinTable         string            `json:"join_table,omitempty"` // For many-to-many
	SourceCardinality string            `json:"source_cardinality"`   // "1" or "*"
	TargetCardinality string            `json:"target_cardinality"`   // "1" or "*"
	Bidirectional     bool              `json:"bidirectional,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// ConstraintInfo contains information about database constraints
type ConstraintInfo struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"` // PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK
	Table      string            `json:"table"`
	Columns    []string          `json:"columns,omitempty"`
	Expression string            `json:"expression,omitempty"` // For CHECK constraints
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// MigrationInfo contains information about database migrations
type MigrationInfo struct {
	Version   string            `json:"version"`
	Timestamp string            `json:"timestamp,omitempty"`
	File      string            `json:"file"`
	Changes   []MigrationChange `json:"changes,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// MigrationChange represents a single change in a migration
type MigrationChange struct {
	Type     string            `json:"type"` // CREATE_TABLE, ALTER_TABLE, DROP_TABLE, etc.
	Table    string            `json:"table,omitempty"`
	Column   string            `json:"column,omitempty"`
	Details  string            `json:"details,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
