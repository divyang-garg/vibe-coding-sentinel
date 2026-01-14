// Phase 14A: Feature Discovery Algorithm
// Automatically discovers features across all layers (UI, API, Database, Logic, Integration, Tests)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// FeatureDiscovery contains framework and technology detection results
type FeatureDiscovery struct {
	UIFramework    string            `json:"ui_framework"`
	UIFrameworkVer string            `json:"ui_framework_version,omitempty"`
	APIFramework   string            `json:"api_framework"`
	DatabaseORM    string            `json:"database_orm"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// AnalysisContext stores framework and discovery context
type AnalysisContext struct {
	FrameworkInfo *FeatureDiscovery `json:"framework_info"`
	CodebasePath  string            `json:"codebase_path"`
	Language      string            `json:"language,omitempty"`
}

// UILayerComponents represents discovered UI components
type UILayerComponents struct {
	Components []ComponentInfo `json:"components"`
	Framework  string          `json:"framework"`
}

// ComponentInfo contains information about a UI component
type ComponentInfo struct {
	Name     string            `json:"name"`
	Path     string            `json:"path"`
	Type     string            `json:"type"` // "component", "form", "page"
	Props    []string          `json:"props,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// APILayerEndpoints represents discovered API endpoints
type APILayerEndpoints struct {
	Endpoints []EndpointInfo `json:"endpoints"`
	Framework string         `json:"framework"`
}

// EndpointInfo contains information about an API endpoint
type EndpointInfo struct {
	Method   string            `json:"method"` // GET, POST, PUT, DELETE
	Path     string            `json:"path"`
	Handler  string            `json:"handler,omitempty"`
	File     string            `json:"file"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// DatabaseLayerTables represents discovered database tables
type DatabaseLayerTables struct {
	Tables  []TableInfo `json:"tables"`
	ORMType string      `json:"orm_type,omitempty"`
}

// TableInfo contains information about a database table
type TableInfo struct {
	Name          string             `json:"name"`
	Columns       []ColumnInfo       `json:"columns,omitempty"`
	Relationships []RelationshipInfo `json:"relationships,omitempty"`
	Source        string             `json:"source"` // "migration", "prisma", "typeorm"
	File          string             `json:"file,omitempty"`
	Metadata      map[string]string  `json:"metadata,omitempty"`
}

// ColumnInfo contains information about a database column
type ColumnInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	PrimaryKey bool   `json:"primary_key,omitempty"`
}

// RelationshipInfo contains information about table relationships
type RelationshipInfo struct {
	Type        string `json:"type"` // "one-to-many", "many-to-one", "many-to-many"
	TargetTable string `json:"target_table"`
	ForeignKey  string `json:"foreign_key,omitempty"`
}

// LogicLayerFunctions represents discovered business logic functions
type LogicLayerFunctions struct {
	Functions []BusinessLogicFunctionInfo `json:"functions"`
	Language  string                      `json:"language"`
}

// BusinessLogicFunctionInfo contains information about a business logic function
type BusinessLogicFunctionInfo struct {
	Name       string            `json:"name"`
	Signature  string            `json:"signature,omitempty"`
	File       string            `json:"file"`
	LineNumber int               `json:"line_number,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// IntegrationLayerAPIs represents discovered external API integrations
type IntegrationLayerAPIs struct {
	Integrations []IntegrationInfo `json:"integrations"`
}

// IntegrationInfo contains information about an external API integration
type IntegrationInfo struct {
	Service    string            `json:"service,omitempty"` // "payment", "warehouse", etc.
	Endpoint   string            `json:"endpoint"`
	Method     string            `json:"method"`
	File       string            `json:"file"`
	LineNumber int               `json:"line_number,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// TestLayerFiles represents discovered test files
type TestLayerFiles struct {
	TestFiles []TestFileInfo `json:"test_files"`
}

// TestFileInfo contains information about a test file
type TestFileInfo struct {
	Path      string            `json:"path"`
	Framework string            `json:"framework,omitempty"` // "jest", "mocha", "pytest", etc.
	TestCases []string          `json:"test_cases,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// DiscoveredFeature contains all discovered components for a feature
type DiscoveredFeature struct {
	FeatureName      string                `json:"feature_name"`
	UILayer          *UILayerComponents    `json:"ui_layer,omitempty"`
	APILayer         *APILayerEndpoints    `json:"api_layer,omitempty"`
	DatabaseLayer    *DatabaseLayerTables  `json:"database_layer,omitempty"`
	LogicLayer       *LogicLayerFunctions  `json:"logic_layer,omitempty"`
	IntegrationLayer *IntegrationLayerAPIs `json:"integration_layer,omitempty"`
	TestLayer        *TestLayerFiles       `json:"test_layer,omitempty"`
	Context          *AnalysisContext      `json:"context"`
}

// detectUIFramework detects the UI framework used in the codebase
func detectUIFramework(codebasePath string) (string, string, error) {
	packageJSONPath := filepath.Join(codebasePath, "package.json")

	// Check package.json for React, Vue, Angular
	if _, err := os.Stat(packageJSONPath); err == nil {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
					// Check for React
					if react, ok := deps["react"].(string); ok {
						// Check for Next.js
						if _, hasNext := deps["next"]; hasNext {
							return "nextjs", react, nil
						}
						return "react", react, nil
					}
					// Check for Vue
					if vue, ok := deps["vue"].(string); ok {
						return "vue", vue, nil
					}
					// Check for Angular
					if angular, ok := deps["@angular/core"].(string); ok {
						return "angular", angular, nil
					}
				}
			}
		}
	}

	// Check for framework-specific config files
	if _, err := os.Stat(filepath.Join(codebasePath, "vite.config.js")); err == nil {
		return "vite", "", nil // Could be Vue or React with Vite
	}
	if _, err := os.Stat(filepath.Join(codebasePath, "vite.config.ts")); err == nil {
		return "vite", "", nil
	}
	if _, err := os.Stat(filepath.Join(codebasePath, "next.config.js")); err == nil {
		return "nextjs", "", nil
	}
	if _, err := os.Stat(filepath.Join(codebasePath, "angular.json")); err == nil {
		return "angular", "", nil
	}

	return "unknown", "", nil
}

// detectAPIFramework detects the API framework used in the codebase
func detectAPIFramework(codebasePath string) (string, error) {
	packageJSONPath := filepath.Join(codebasePath, "package.json")

	// Check package.json for Express (Node.js)
	if _, err := os.Stat(packageJSONPath); err == nil {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
					if _, ok := deps["express"]; ok {
						return "express", nil
					}
				}
			}
		}
	}

	// Check for Python frameworks
	requirementsPath := filepath.Join(codebasePath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		data, err := os.ReadFile(requirementsPath)
		if err == nil {
			content := string(data)
			if strings.Contains(content, "fastapi") {
				return "fastapi", nil
			}
			if strings.Contains(content, "django") {
				return "django", nil
			}
			if strings.Contains(content, "flask") {
				return "flask", nil
			}
		}
	}

	// Check for Go files (Gin router)
	goFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.go"))
	if len(goFiles) > 0 {
		// Check for Gin imports
		for _, file := range goFiles {
			data, err := os.ReadFile(file)
			if err == nil {
				content := string(data)
				if strings.Contains(content, "github.com/gin-gonic/gin") {
					return "gin", nil
				}
				if strings.Contains(content, "github.com/go-chi/chi") {
					return "chi", nil
				}
			}
		}
	}

	return "unknown", nil
}

// detectDatabaseORM detects the database ORM used in the codebase
func detectDatabaseORM(codebasePath string) (string, error) {
	// Check for Prisma
	prismaSchemaPath := filepath.Join(codebasePath, "prisma", "schema.prisma")
	if _, err := os.Stat(prismaSchemaPath); err == nil {
		return "prisma", nil
	}

	// Check for TypeORM entities (TypeScript files with @Entity decorator)
	tsFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.entity.ts"))
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

	// Check for SQL migrations (indicates raw SQL or ORM)
	migrationDirs := []string{
		filepath.Join(codebasePath, "migrations"),
		filepath.Join(codebasePath, "db", "migrations"),
		filepath.Join(codebasePath, "database", "migrations"),
	}
	for _, dir := range migrationDirs {
		if _, err := os.Stat(dir); err == nil {
			sqlFiles, _ := filepath.Glob(filepath.Join(dir, "*.sql"))
			if len(sqlFiles) > 0 {
				return "raw_sql", nil
			}
		}
	}

	return "unknown", nil
}

// detectFrameworks detects all frameworks in the codebase
func detectFrameworks(codebasePath string) (*FeatureDiscovery, error) {
	uiFramework, uiVersion, err := detectUIFramework(codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect UI framework: %w", err)
	}

	apiFramework, err := detectAPIFramework(codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect API framework: %w", err)
	}

	orm, err := detectDatabaseORM(codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect database ORM: %w", err)
	}

	return &FeatureDiscovery{
		UIFramework:    uiFramework,
		UIFrameworkVer: uiVersion,
		APIFramework:   apiFramework,
		DatabaseORM:    orm,
		Metadata:       make(map[string]string),
	}, nil
}

// discoverUIComponents discovers UI components in the codebase
func discoverUIComponents(ctx context.Context, codebasePath string, featureName string, framework string) (*UILayerComponents, error) {
	components := []ComponentInfo{}

	// Determine file extensions based on framework
	var extensions []string
	switch framework {
	case "react", "nextjs":
		extensions = []string{"*.tsx", "*.jsx"}
	case "vue":
		extensions = []string{"*.vue"}
	case "angular":
		extensions = []string{"*.ts", "*.component.ts"}
	default:
		// Try all common extensions
		extensions = []string{"*.tsx", "*.jsx", "*.vue", "*.ts"}
	}

	// Search for component files
	for _, ext := range extensions {
		pattern := filepath.Join(codebasePath, "**", ext)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range matches {
			// Skip node_modules and build directories
			if strings.Contains(file, "node_modules") || strings.Contains(file, "build") || strings.Contains(file, "dist") {
				continue
			}

			// Check if file name matches feature keywords
			fileName := filepath.Base(file)
			if matchesFeature(fileName, featureName) {
				component := extractComponentInfo(file, framework, featureName)
				if component != nil {
					components = append(components, *component)
				}
			}
		}
	}

	return &UILayerComponents{
		Components: components,
		Framework:  framework,
	}, nil
}

// extractComponentInfo extracts component information from a file
func extractComponentInfo(filePath string, framework string, featureName string) *ComponentInfo {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	content := string(data)
	fileName := filepath.Base(filePath)
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	component := &ComponentInfo{
		Name:     name,
		Path:     filePath,
		Type:     "component",
		Props:    []string{},
		Metadata: make(map[string]string),
	}

	// Extract component-specific information based on framework
	switch framework {
	case "react", "nextjs":
		// Check for function component or class component
		if strings.Contains(content, "function "+name) || strings.Contains(content, "const "+name+" =") {
			// Extract props from function parameters
			// This is simplified - in production, use AST
			if strings.Contains(content, "props") || strings.Contains(content, "{"+name+"Props") {
				component.Props = append(component.Props, "props")
			}
		}
		// Check if it's a form component
		if strings.Contains(content, "react-hook-form") || strings.Contains(content, "formik") {
			component.Type = "form"
		}
	case "vue":
		// Check for component name in script section
		if strings.Contains(content, "export default") {
			component.Type = "component"
		}
		// Check for form validation
		if strings.Contains(content, "vuelidate") || strings.Contains(content, "vee-validate") {
			component.Type = "form"
		}
	case "angular":
		// Check for @Component decorator
		if strings.Contains(content, "@Component") {
			component.Type = "component"
		}
		// Check for form validation
		if strings.Contains(content, "FormGroup") || strings.Contains(content, "ngForm") {
			component.Type = "form"
		}
	}

	return component
}

// matchesFeature checks if a file name matches feature keywords
func matchesFeature(fileName string, featureName string) bool {
	fileNameLower := strings.ToLower(fileName)
	featureLower := strings.ToLower(featureName)

	// Extract keywords from feature name
	keywords := extractFeatureKeywords(featureLower)

	for _, keyword := range keywords {
		if strings.Contains(fileNameLower, keyword) {
			return true
		}
	}

	return false
}

// extractFeatureKeywords extracts keywords from a feature name
func extractFeatureKeywords(featureName string) []string {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
	}

	words := strings.Fields(featureName)
	var keywords []string
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && !stopWords[strings.ToLower(word)] {
			keywords = append(keywords, strings.ToLower(word))
		}
	}

	return keywords
}

// discoverAPIEndpoints discovers API endpoints in the codebase
func discoverAPIEndpoints(ctx context.Context, codebasePath string, featureName string, framework string) (*APILayerEndpoints, error) {
	endpoints := []EndpointInfo{}

	switch framework {
	case "express":
		endpoints = discoverExpressEndpoints(codebasePath, featureName)
	case "fastapi":
		endpoints = discoverFastAPIEndpoints(codebasePath, featureName)
	case "django":
		endpoints = discoverDjangoEndpoints(codebasePath, featureName)
	case "gin", "chi":
		endpoints = discoverGoEndpoints(codebasePath, featureName, framework)
	}

	return &APILayerEndpoints{
		Endpoints: endpoints,
		Framework: framework,
	}, nil
}

// discoverExpressEndpoints discovers Express.js endpoints
func discoverExpressEndpoints(codebasePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// Search for route files
	jsFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.js"))
	tsFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.ts"))

	allFiles := append(jsFiles, tsFiles...)

	for _, file := range allFiles {
		if strings.Contains(file, "node_modules") {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)

		// Look for Express route definitions
		methods := []string{"get", "post", "put", "delete", "patch"}
		for _, method := range methods {
			// Use regex or simple string matching
			if strings.Contains(content, "app."+method+"(") {
				// Extract path (simplified - use regex in production)
				// This is a simplified version
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					if strings.Contains(line, "app."+method+"(") {
						// Extract path from line
						// Simplified extraction
						endpoint := EndpointInfo{
							Method:   strings.ToUpper(method),
							Path:     extractPathFromLine(line),
							File:     file,
							Metadata: make(map[string]string),
						}
						if matchesFeature(filepath.Base(file), featureName) || matchesFeature(endpoint.Path, featureName) {
							endpoints = append(endpoints, endpoint)
						}
					}
				}
			}
		}
	}

	return endpoints
}

// discoverFastAPIEndpoints discovers FastAPI endpoints
func discoverFastAPIEndpoints(codebasePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	pyFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.py"))

	for _, file := range pyFiles {
		if strings.Contains(file, "__pycache__") {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)

		// Look for FastAPI route decorators
		methods := []string{"get", "post", "put", "delete", "patch"}
		for _, method := range methods {
			if strings.Contains(content, "@app."+method+"(") {
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					if strings.Contains(line, "@app."+method+"(") {
						endpoint := EndpointInfo{
							Method:   strings.ToUpper(method),
							Path:     extractPathFromLine(line),
							File:     file,
							Metadata: make(map[string]string),
						}
						if matchesFeature(filepath.Base(file), featureName) || matchesFeature(endpoint.Path, featureName) {
							endpoints = append(endpoints, endpoint)
						}
					}
				}
			}
		}
	}

	return endpoints
}

// discoverDjangoEndpoints discovers Django endpoints
func discoverDjangoEndpoints(codebasePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// Look for urls.py files
	urlFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/urls.py"))

	for _, file := range urlFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)

		// Look for URL patterns
		if strings.Contains(content, "path(") || strings.Contains(content, "url(") {
			// Simplified extraction - use regex in production
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "path(") || strings.Contains(line, "url(") {
					endpoint := EndpointInfo{
						Method:   "GET", // Django defaults to GET
						Path:     extractPathFromLine(line),
						File:     file,
						Metadata: make(map[string]string),
					}
					if matchesFeature(filepath.Base(file), featureName) || matchesFeature(endpoint.Path, featureName) {
						endpoints = append(endpoints, endpoint)
					}
				}
			}
		}
	}

	return endpoints
}

// discoverGoEndpoints discovers Go (Gin/Chi) endpoints
func discoverGoEndpoints(codebasePath string, featureName string, framework string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	goFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.go"))

	for _, file := range goFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)

		// Look for Gin or Chi router methods
		var prefix string
		if framework == "gin" {
			prefix = "r."
		} else if framework == "chi" {
			prefix = "r."
		}

		methods := []string{"Get", "Post", "Put", "Delete", "Patch"}
		for _, method := range methods {
			pattern := prefix + method + "("
			if strings.Contains(content, pattern) {
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					if strings.Contains(line, pattern) {
						endpoint := EndpointInfo{
							Method:   strings.ToUpper(method),
							Path:     extractPathFromLine(line),
							File:     file,
							Metadata: make(map[string]string),
						}
						if matchesFeature(filepath.Base(file), featureName) || matchesFeature(endpoint.Path, featureName) {
							endpoints = append(endpoints, endpoint)
						}
					}
				}
			}
		}
	}

	return endpoints
}

// extractPathFromLine extracts a path from a route definition line (simplified)
func extractPathFromLine(line string) string {
	// Simplified extraction - use regex in production
	// Look for quoted strings
	start := strings.Index(line, "\"")
	if start == -1 {
		start = strings.Index(line, "'")
		if start == -1 {
			return "/"
		}
	}

	end := strings.Index(line[start+1:], "\"")
	if end == -1 {
		end = strings.Index(line[start+1:], "'")
		if end == -1 {
			return "/"
		}
	}

	return line[start+1 : start+1+end]
}

// discoverDatabaseTables discovers database tables in the codebase
func discoverDatabaseTables(ctx context.Context, codebasePath string, featureName string, ormType string) (*DatabaseLayerTables, error) {
	tables := []TableInfo{}

	switch ormType {
	case "prisma":
		tables = discoverPrismaTables(codebasePath, featureName)
	case "typeorm":
		tables = discoverTypeORMTables(codebasePath, featureName)
	case "raw_sql":
		tables = discoverSQLTables(codebasePath, featureName)
	default:
		// Try all methods
		tables = append(tables, discoverPrismaTables(codebasePath, featureName)...)
		tables = append(tables, discoverTypeORMTables(codebasePath, featureName)...)
		tables = append(tables, discoverSQLTables(codebasePath, featureName)...)
	}

	return &DatabaseLayerTables{
		Tables:  tables,
		ORMType: ormType,
	}, nil
}

// discoverPrismaTables discovers Prisma schema tables
func discoverPrismaTables(codebasePath string, featureName string) []TableInfo {
	tables := []TableInfo{}

	schemaPath := filepath.Join(codebasePath, "prisma", "schema.prisma")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return tables
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var currentTable *TableInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "model ") {
			// Extract model name
			modelName := strings.TrimSpace(strings.TrimPrefix(line, "model "))
			modelName = strings.Fields(modelName)[0]

			if matchesFeature(modelName, featureName) {
				currentTable = &TableInfo{
					Name:     modelName,
					Source:   "prisma",
					File:     schemaPath,
					Columns:  []ColumnInfo{},
					Metadata: make(map[string]string),
				}
			}
		} else if currentTable != nil && strings.Contains(line, "  ") {
			// Extract column (simplified)
			if strings.Contains(line, "@id") {
				// Primary key
				parts := strings.Fields(line)
				if len(parts) > 0 {
					colName := parts[0]
					currentTable.Columns = append(currentTable.Columns, ColumnInfo{
						Name:       colName,
						PrimaryKey: true,
						Nullable:   false,
					})
				}
			} else if strings.Contains(line, "?") {
				// Nullable column
				parts := strings.Fields(line)
				if len(parts) > 0 {
					colName := parts[0]
					currentTable.Columns = append(currentTable.Columns, ColumnInfo{
						Name:     colName,
						Nullable: true,
					})
				}
			} else if !strings.HasPrefix(line, "@") && !strings.HasPrefix(line, "}") {
				// Regular column
				parts := strings.Fields(line)
				if len(parts) > 0 {
					colName := parts[0]
					currentTable.Columns = append(currentTable.Columns, ColumnInfo{
						Name:     colName,
						Nullable: false,
					})
				}
			}
		} else if currentTable != nil && strings.HasPrefix(line, "}") {
			tables = append(tables, *currentTable)
			currentTable = nil
		}
	}

	return tables
}

// discoverTypeORMTables discovers TypeORM entity tables
func discoverTypeORMTables(codebasePath string, featureName string) []TableInfo {
	tables := []TableInfo{}

	entityFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.entity.ts"))

	for _, file := range entityFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)

		// Look for @Entity decorator
		if strings.Contains(content, "@Entity") {
			// Extract entity name (simplified)
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "@Entity") {
					// Extract table name from decorator or class name
					// Simplified extraction
					tableName := extractTableNameFromEntity(line, content)
					if matchesFeature(tableName, featureName) {
						table := TableInfo{
							Name:     tableName,
							Source:   "typeorm",
							File:     file,
							Columns:  []ColumnInfo{},
							Metadata: make(map[string]string),
						}
						// Extract columns (simplified)
						table.Columns = extractTypeORMColumns(content)
						tables = append(tables, table)
					}
					break
				}
			}
		}
	}

	return tables
}

// discoverSQLTables discovers SQL migration tables
func discoverSQLTables(codebasePath string, featureName string) []TableInfo {
	tables := []TableInfo{}

	migrationDirs := []string{
		filepath.Join(codebasePath, "migrations"),
		filepath.Join(codebasePath, "db", "migrations"),
		filepath.Join(codebasePath, "database", "migrations"),
	}

	for _, dir := range migrationDirs {
		sqlFiles, _ := filepath.Glob(filepath.Join(dir, "*.sql"))
		for _, file := range sqlFiles {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			content := string(data)

			// Look for CREATE TABLE statements
			if strings.Contains(strings.ToUpper(content), "CREATE TABLE") {
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					lineUpper := strings.ToUpper(line)
					if strings.Contains(lineUpper, "CREATE TABLE") {
						// Extract table name (simplified)
						parts := strings.Fields(lineUpper)
						for i, part := range parts {
							if part == "TABLE" && i+1 < len(parts) {
								tableName := strings.Trim(parts[i+1], "`\"'")
								if matchesFeature(tableName, featureName) {
									table := TableInfo{
										Name:     tableName,
										Source:   "migration",
										File:     file,
										Columns:  []ColumnInfo{},
										Metadata: make(map[string]string),
									}
									// Extract columns (simplified - would use SQL parser in production)
									tables = append(tables, table)
								}
								break
							}
						}
					}
				}
			}
		}
	}

	return tables
}

// extractTableNameFromEntity extracts table name from TypeORM entity (simplified)
func extractTableNameFromEntity(entityLine string, content string) string {
	// Look for @Entity("table_name") or use class name
	if strings.Contains(entityLine, "@Entity(") {
		start := strings.Index(entityLine, "(")
		if start != -1 {
			end := strings.Index(entityLine[start+1:], ")")
			if end != -1 {
				name := strings.Trim(entityLine[start+1:start+1+end], "\"'")
				if name != "" {
					return name
				}
			}
		}
	}

	// Fallback: extract class name
	if strings.Contains(content, "export class ") {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "export class ") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "class" && i+1 < len(parts) {
						return parts[i+1]
					}
				}
			}
		}
	}

	return "unknown"
}

// extractTypeORMColumns extracts columns from TypeORM entity (simplified)
func extractTypeORMColumns(content string) []ColumnInfo {
	columns := []ColumnInfo{}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "@Column") {
			// Extract column name (simplified)
			parts := strings.Fields(line)
			if len(parts) > 0 {
				colName := parts[0]
				nullable := strings.Contains(line, "nullable: true")
				primaryKey := strings.Contains(line, "@PrimaryColumn") || strings.Contains(line, "@PrimaryGeneratedColumn")

				columns = append(columns, ColumnInfo{
					Name:       colName,
					Nullable:   nullable,
					PrimaryKey: primaryKey,
				})
			}
		}
	}

	return columns
}

// discoverBusinessLogic discovers business logic functions using AST analyzer
func discoverBusinessLogic(ctx context.Context, codebasePath string, featureName string, language string) (*LogicLayerFunctions, error) {
	functions := []BusinessLogicFunctionInfo{}

	// Use AST analyzer to find functions
	// Determine language from codebase
	if language == "" {
		// Auto-detect language
		goFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.go"))
		jsFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.js"))
		tsFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.ts"))
		pyFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.py"))

		if len(goFiles) > 0 {
			language = "go"
		} else if len(tsFiles) > 0 {
			language = "typescript"
		} else if len(jsFiles) > 0 {
			language = "javascript"
		} else if len(pyFiles) > 0 {
			language = "python"
		}
	}

	// Search for service/domain layer files
	var codeFiles []string
	switch language {
	case "go":
		codeFiles, _ = filepath.Glob(filepath.Join(codebasePath, "**/*.go"))
	case "typescript", "javascript":
		codeFiles, _ = filepath.Glob(filepath.Join(codebasePath, "**/*.{ts,js}"))
	case "python":
		codeFiles, _ = filepath.Glob(filepath.Join(codebasePath, "**/*.py"))
	}

	// Filter by feature keywords
	keywords := extractFeatureKeywords(strings.ToLower(featureName))

	for _, file := range codeFiles {
		if strings.Contains(file, "node_modules") || strings.Contains(file, "test") {
			continue
		}

		// Check if file matches feature
		if !matchesFeature(filepath.Base(file), featureName) {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Use AST analyzer to extract functions
		findings, _, err := analyzeAST(string(data), language, []string{"functions"})
		if err != nil {
			continue
		}

		// Filter functions by feature keywords
		for _, finding := range findings {
			// Extract function name from message or code
			funcName := extractFunctionNameFromFinding(finding)
			if funcName != "" {
				funcNameLower := strings.ToLower(funcName)
				for _, keyword := range keywords {
					if strings.Contains(funcNameLower, keyword) {
						functions = append(functions, BusinessLogicFunctionInfo{
							Name:       funcName,
							Signature:  funcName, // Simplified
							File:       file,
							LineNumber: finding.Line,
							Metadata:   make(map[string]string),
						})
						break
					}
				}
			}
		}
	}

	return &LogicLayerFunctions{
		Functions: functions,
		Language:  language,
	}, nil
}

// discoverIntegrations discovers external API integrations
func discoverIntegrations(ctx context.Context, codebasePath string, featureName string) (*IntegrationLayerAPIs, error) {
	integrations := []IntegrationInfo{}

	// Search for HTTP client calls
	codeFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.{js,ts,py,go}"))

	for _, file := range codeFiles {
		if strings.Contains(file, "node_modules") || strings.Contains(file, "test") {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)

		// Look for HTTP client calls
		httpPatterns := []struct {
			pattern string
			method  string
		}{
			{"axios.get(", "GET"},
			{"axios.post(", "POST"},
			{"fetch(", "GET"},
			{"http.Get(", "GET"},
			{"http.Post(", "POST"},
			{"requests.get(", "GET"},
			{"requests.post(", "POST"},
		}

		lines := strings.Split(content, "\n")
		for lineNum, line := range lines {
			for _, pattern := range httpPatterns {
				if strings.Contains(line, pattern.pattern) {
					// Extract endpoint (simplified)
					endpoint := extractEndpointFromLine(line)
					if endpoint != "" && matchesFeature(endpoint, featureName) {
						integrations = append(integrations, IntegrationInfo{
							Service:    identifyService(endpoint),
							Endpoint:   endpoint,
							Method:     pattern.method,
							File:       file,
							LineNumber: lineNum + 1,
							Metadata:   make(map[string]string),
						})
					}
				}
			}
		}
	}

	return &IntegrationLayerAPIs{
		Integrations: integrations,
	}, nil
}

// discoverTests discovers test files
func discoverTests(ctx context.Context, codebasePath string, featureName string) (*TestLayerFiles, error) {
	testFiles := []TestFileInfo{}

	// Search for test files
	testPatterns := []string{
		"**/*.test.{js,ts,py}",
		"**/*.spec.{js,ts,py}",
		"**/test_*.py",
		"**/*_test.py",
		"**/*_test.go",
	}

	for _, pattern := range testPatterns {
		matches, _ := filepath.Glob(filepath.Join(codebasePath, pattern))
		for _, file := range matches {
			if strings.Contains(file, "node_modules") {
				continue
			}

			// Check if test file matches feature
			if matchesFeature(filepath.Base(file), featureName) {
				framework := detectTestFramework(file)
				testFile := TestFileInfo{
					Path:      file,
					Framework: framework,
					TestCases: []string{},
					Metadata:  make(map[string]string),
				}

				// Extract test cases (simplified)
				data, err := os.ReadFile(file)
				if err == nil {
					testFile.TestCases = extractTestCases(string(data), framework)
				}

				testFiles = append(testFiles, testFile)
			}
		}
	}

	return &TestLayerFiles{
		TestFiles: testFiles,
	}, nil
}

// mapManualFilesToLayers maps manually provided files to appropriate layers
func mapManualFilesToLayers(ctx context.Context, feature *DiscoveredFeature, manualFiles map[string][]string, codebasePath string) *DiscoveredFeature {
	frameworkInfo := feature.Context.FrameworkInfo
	if frameworkInfo == nil {
		// Detect frameworks if not available
		detected, _ := detectFrameworks(codebasePath)
		frameworkInfo = detected
		feature.Context.FrameworkInfo = frameworkInfo
	}

	// Initialize layer structures
	uiComponents := []ComponentInfo{}
	apiEndpoints := []EndpointInfo{}
	dbTables := []TableInfo{}
	logicFunctions := []BusinessLogicFunctionInfo{}
	integrations := []IntegrationInfo{}
	testFiles := []TestFileInfo{}

	// Process files by layer category if provided
	if uiFiles, ok := manualFiles["ui"]; ok {
		for _, filePath := range uiFiles {
			component := extractComponentInfo(filePath, codebasePath, frameworkInfo.UIFramework)
			if component != nil {
				uiComponents = append(uiComponents, *component)
			}
		}
		if len(uiComponents) > 0 {
			feature.UILayer = &UILayerComponents{
				Framework:  frameworkInfo.UIFramework,
				Components: uiComponents,
			}
		}
	}

	if apiFiles, ok := manualFiles["api"]; ok {
		for _, filePath := range apiFiles {
			// Simple endpoint extraction - would use proper parsing in production
			endpoint := &EndpointInfo{
				Path:   filePath,
				Method: "POST", // Default, would be extracted from code
				File:   filePath,
			}
			apiEndpoints = append(apiEndpoints, *endpoint)
		}
		if len(apiEndpoints) > 0 {
			feature.APILayer = &APILayerEndpoints{
				Framework: frameworkInfo.APIFramework,
				Endpoints: apiEndpoints,
			}
		}
	}

	if dbFiles, ok := manualFiles["database"]; ok {
		for _, filePath := range dbFiles {
			// Simple table extraction - would use proper parsing in production
			table := &TableInfo{
				Name: filepath.Base(filePath),
				File: filePath,
			}
			dbTables = append(dbTables, *table)
		}
		if len(dbTables) > 0 {
			feature.DatabaseLayer = &DatabaseLayerTables{
				ORMType: frameworkInfo.DatabaseORM,
				Tables:  dbTables,
			}
		}
	}

	if logicFiles, ok := manualFiles["logic"]; ok {
		for _, filePath := range logicFiles {
			// Extract functions from file
			functions := extractFunctionsFromFile(filePath, codebasePath)
			for _, f := range functions {
				logicFunctions = append(logicFunctions, BusinessLogicFunctionInfo{
					Name:       f.Name,
					File:       f.File,
					LineNumber: f.LineNumber,
				})
			}
		}
		if len(logicFunctions) > 0 {
			feature.LogicLayer = &LogicLayerFunctions{
				Functions: logicFunctions,
			}
		}
	}

	if integrationFiles, ok := manualFiles["integration"]; ok {
		for _, filePath := range integrationFiles {
			// Simple integration extraction
			integration := &IntegrationInfo{
				Service:  filepath.Base(filePath),
				Endpoint: filePath,
				Method:   "POST",
				File:     filePath,
			}
			integrations = append(integrations, *integration)
		}
		if len(integrations) > 0 {
			feature.IntegrationLayer = &IntegrationLayerAPIs{
				Integrations: integrations,
			}
		}
	}

	if testFilesList, ok := manualFiles["tests"]; ok {
		for _, filePath := range testFilesList {
			// Simple test file extraction
			testFile := &TestFileInfo{
				Path: filePath,
			}
			testFiles = append(testFiles, *testFile)
		}
		if len(testFiles) > 0 {
			feature.TestLayer = &TestLayerFiles{
				TestFiles: testFiles,
			}
		}
	}

	// If no layer categories provided, auto-detect based on file patterns
	if len(manualFiles) == 0 || (len(uiComponents) == 0 && len(apiEndpoints) == 0 && len(dbTables) == 0 && len(logicFunctions) == 0 && len(integrations) == 0 && len(testFiles) == 0) {
		// Try to auto-categorize files from a generic "files" list
		if allFiles, ok := manualFiles["files"]; ok {
			for _, filePath := range allFiles {
				// Categorize based on file path and extension
				if isUIFile(filePath) {
					component := extractComponentInfo(filePath, codebasePath, frameworkInfo.UIFramework)
					if component != nil {
						uiComponents = append(uiComponents, *component)
					}
				} else if isAPIFile(filePath) {
					endpoint := &EndpointInfo{
						Path:   filePath,
						Method: "POST",
						File:   filePath,
					}
					apiEndpoints = append(apiEndpoints, *endpoint)
				} else if isDatabaseFile(filePath) {
					table := &TableInfo{
						Name: filepath.Base(filePath),
						File: filePath,
					}
					dbTables = append(dbTables, *table)
				} else if isTestFile(filePath) {
					testFile := &TestFileInfo{
						Path: filePath,
					}
					testFiles = append(testFiles, *testFile)
				} else {
					// Default to logic layer
					functions := extractFunctionsFromFile(filePath, codebasePath)
					for _, f := range functions {
						logicFunctions = append(logicFunctions, BusinessLogicFunctionInfo{
							Name:       f.Name,
							File:       f.File,
							LineNumber: f.LineNumber,
						})
					}
				}
			}

			// Update feature with discovered layers
			if len(uiComponents) > 0 {
				feature.UILayer = &UILayerComponents{
					Framework:  frameworkInfo.UIFramework,
					Components: uiComponents,
				}
			}
			if len(apiEndpoints) > 0 {
				feature.APILayer = &APILayerEndpoints{
					Framework: frameworkInfo.APIFramework,
					Endpoints: apiEndpoints,
				}
			}
			if len(dbTables) > 0 {
				feature.DatabaseLayer = &DatabaseLayerTables{
					ORMType: frameworkInfo.DatabaseORM,
					Tables:  dbTables,
				}
			}
			if len(logicFunctions) > 0 {
				feature.LogicLayer = &LogicLayerFunctions{
					Functions: logicFunctions,
				}
			}
			if len(testFiles) > 0 {
				feature.TestLayer = &TestLayerFiles{
					TestFiles: testFiles,
				}
			}
		}
	}

	return feature
}

// Helper functions for file categorization
func isUIFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	path := strings.ToLower(filePath)
	return ext == ".jsx" || ext == ".tsx" || ext == ".vue" || ext == ".js" || ext == ".ts" ||
		strings.Contains(path, "/components/") || strings.Contains(path, "/ui/") ||
		strings.Contains(path, "/views/") || strings.Contains(path, "/pages/")
}

func isAPIFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	path := strings.ToLower(filePath)
	return ext == ".go" && (strings.Contains(path, "/api/") || strings.Contains(path, "/routes/") ||
		strings.Contains(path, "/handlers/") || strings.Contains(path, "/controllers/")) ||
		strings.Contains(path, "/endpoints/") || strings.Contains(path, "/routes/")
}

func isDatabaseFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	path := strings.ToLower(filePath)
	return ext == ".sql" || ext == ".prisma" || strings.Contains(path, "/migrations/") ||
		strings.Contains(path, "/schema/") || strings.Contains(path, "/models/") ||
		strings.Contains(path, "schema.ts") || strings.Contains(path, "schema.js")
}

func isTestFile(filePath string) bool {
	path := strings.ToLower(filePath)
	return strings.Contains(path, "_test.") || strings.Contains(path, ".test.") ||
		strings.Contains(path, ".spec.") || strings.Contains(path, "/test/") ||
		strings.Contains(path, "/tests/") || strings.Contains(path, "__tests__")
}

// extractFunctionsFromFile extracts function information from a file
func extractFunctionsFromFile(filePath string, codebasePath string) []BusinessLogicFunctionInfo {
	functions := []BusinessLogicFunctionInfo{}

	// Read file content
	fullPath := filepath.Join(codebasePath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return functions
	}

	// Try to use AST if available
	ext := strings.ToLower(filepath.Ext(filePath))
	language := ""
	switch ext {
	case ".go":
		language = "go"
	case ".js", ".jsx":
		language = "javascript"
	case ".ts", ".tsx":
		language = "typescript"
	case ".py":
		language = "python"
	default:
		// Fallback to pattern matching
		return extractFunctionsPatternMatch(string(content), filePath)
	}

	// Use AST extraction if parser available
	parser, err := getParser(language)
	if err != nil {
		// Fallback to pattern matching
		return extractFunctionsPatternMatch(string(content), filePath)
	}

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return extractFunctionsPatternMatch(string(content), filePath)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	codeStr := string(content)

	// Traverse AST to find functions
	traverseAST(rootNode, func(node *sitter.Node) bool {
		var funcName string
		var isFunction bool
		var startLine int

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" || child.Type() == "field_identifier" {
							funcName = codeStr[child.StartByte():child.EndByte()]
							isFunction = true
							startLine = int(node.StartPoint().Row) + 1
							break
						}
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" || child.Type() == "property_identifier" {
							funcName = codeStr[child.StartByte():child.EndByte()]
							isFunction = true
							startLine = int(node.StartPoint().Row) + 1
							break
						}
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						funcName = codeStr[child.StartByte():child.EndByte()]
						isFunction = true
						startLine = int(node.StartPoint().Row) + 1
						break
					}
				}
			}
		}

		if isFunction && funcName != "" {
			functions = append(functions, BusinessLogicFunctionInfo{
				Name:       funcName,
				File:       filePath,
				LineNumber: startLine,
			})
		}

		return true
	})

	return functions
}

// extractFunctionsPatternMatch extracts functions using pattern matching (fallback)
func extractFunctionsPatternMatch(content string, filePath string) []BusinessLogicFunctionInfo {
	functions := []BusinessLogicFunctionInfo{}
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Pattern matching for common function declarations
		if strings.Contains(line, "func ") && strings.Contains(line, "(") {
			// Extract function name
			parts := strings.Fields(line)
			for j, part := range parts {
				if part == "func" && j+1 < len(parts) {
					funcName := strings.Split(parts[j+1], "(")[0]
					if funcName != "" {
						functions = append(functions, BusinessLogicFunctionInfo{
							Name:       funcName,
							File:       filePath,
							LineNumber: i + 1,
						})
					}
					break
				}
			}
		}
	}

	return functions
}

// discoverFeature is the main orchestrator that combines all discovery functions
func discoverFeature(ctx context.Context, featureName string, codebasePath string, manualFiles map[string][]string) (*DiscoveredFeature, error) {
	// Detect frameworks
	frameworkInfo, err := detectFrameworks(codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect frameworks: %w", err)
	}

	context := &AnalysisContext{
		FrameworkInfo: frameworkInfo,
		CodebasePath:  codebasePath,
	}

	feature := &DiscoveredFeature{
		FeatureName: featureName,
		Context:     context,
	}

	// If manual files provided, skip discovery and use provided files
	if len(manualFiles) > 0 {
		// Map provided files to appropriate layers based on file patterns and paths
		feature = mapManualFilesToLayers(ctx, feature, manualFiles, codebasePath)
		return feature, nil
	}

	// Discover UI layer
	if frameworkInfo.UIFramework != "unknown" {
		uiLayer, err := discoverUIComponents(ctx, codebasePath, featureName, frameworkInfo.UIFramework)
		if err == nil {
			feature.UILayer = uiLayer
		}
	}

	// Discover API layer
	if frameworkInfo.APIFramework != "unknown" {
		apiLayer, err := discoverAPIEndpoints(ctx, codebasePath, featureName, frameworkInfo.APIFramework)
		if err == nil {
			feature.APILayer = apiLayer
		}
	}

	// Discover Database layer
	if frameworkInfo.DatabaseORM != "unknown" {
		dbLayer, err := discoverDatabaseTables(ctx, codebasePath, featureName, frameworkInfo.DatabaseORM)
		if err == nil {
			feature.DatabaseLayer = dbLayer
		}
	}

	// Discover Logic layer
	logicLayer, err := discoverBusinessLogic(ctx, codebasePath, featureName, "")
	if err == nil {
		feature.LogicLayer = logicLayer
		context.Language = logicLayer.Language
	}

	// Discover Integration layer
	integrationLayer, err := discoverIntegrations(ctx, codebasePath, featureName)
	if err == nil {
		feature.IntegrationLayer = integrationLayer
	}

	// Discover Test layer
	testLayer, err := discoverTests(ctx, codebasePath, featureName)
	if err == nil {
		feature.TestLayer = testLayer
	}

	return feature, nil
}

// Helper functions

// extractEndpointFromLine extracts endpoint URL from HTTP call line (simplified)
func extractEndpointFromLine(line string) string {
	// Look for quoted strings (URLs)
	start := strings.Index(line, "\"")
	if start == -1 {
		start = strings.Index(line, "'")
		if start == -1 {
			return ""
		}
	}

	end := strings.Index(line[start+1:], "\"")
	if end == -1 {
		end = strings.Index(line[start+1:], "'")
		if end == -1 {
			return ""
		}
	}

	return line[start+1 : start+1+end]
}

// identifyService identifies the service type from endpoint URL
func identifyService(endpoint string) string {
	endpointLower := strings.ToLower(endpoint)

	if strings.Contains(endpointLower, "payment") || strings.Contains(endpointLower, "stripe") || strings.Contains(endpointLower, "paypal") {
		return "payment"
	}
	if strings.Contains(endpointLower, "warehouse") || strings.Contains(endpointLower, "inventory") {
		return "warehouse"
	}
	if strings.Contains(endpointLower, "email") || strings.Contains(endpointLower, "mail") {
		return "email"
	}
	if strings.Contains(endpointLower, "sms") || strings.Contains(endpointLower, "twilio") {
		return "sms"
	}

	return "external"
}

// detectTestFramework detects the test framework from file
func detectTestFramework(filePath string) string {
	ext := filepath.Ext(filePath)

	if ext == ".go" {
		return "go-test"
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}

	content := string(data)

	if strings.Contains(content, "describe(") || strings.Contains(content, "it(") {
		if strings.Contains(content, "jest") {
			return "jest"
		}
		return "mocha"
	}
	if strings.Contains(content, "test(") {
		return "jest"
	}
	if strings.Contains(content, "def test_") {
		return "pytest"
	}
	if strings.Contains(content, "unittest") {
		return "unittest"
	}

	return "unknown"
}

// extractTestCases extracts test case names from test file (simplified)
func extractTestCases(content string, framework string) []string {
	testCases := []string{}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		switch framework {
		case "jest", "mocha":
			if strings.Contains(line, "it(") || strings.Contains(line, "test(") {
				// Extract test name
				start := strings.Index(line, "(")
				if start != -1 {
					end := strings.Index(line[start+1:], ",")
					if end == -1 {
						end = strings.Index(line[start+1:], ")")
					}
					if end != -1 {
						name := strings.Trim(line[start+1:start+1+end], "\"' ")
						if name != "" {
							testCases = append(testCases, name)
						}
					}
				}
			}
		case "pytest":
			if strings.Contains(line, "def test_") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "test_") {
						testCases = append(testCases, part)
						break
					}
				}
			}
		case "go-test":
			if strings.Contains(line, "func Test") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "Test") {
						testCases = append(testCases, part)
						break
					}
				}
			}
		}
	}

	return testCases
}

// extractFunctionNameFromFinding extracts function name from ASTFinding
func extractFunctionNameFromFinding(finding ASTFinding) string {
	// Try to extract from message (e.g., "Duplicate function: processOrder")
	if strings.Contains(finding.Message, "function:") {
		parts := strings.Split(finding.Message, "function:")
		if len(parts) > 1 {
			name := strings.TrimSpace(parts[1])
			// Remove any trailing punctuation
			name = strings.Trim(name, ".,!?;:")
			return name
		}
	}

	// Try to extract from code snippet
	if finding.Code != "" {
		// Look for function definition patterns
		lines := strings.Split(finding.Code, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Check for common function patterns
			if strings.Contains(line, "func ") {
				// Go function
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "func" && i+1 < len(parts) {
						funcName := parts[i+1]
						// Remove receiver if present
						if strings.Contains(funcName, "(") {
							continue
						}
						return funcName
					}
				}
			} else if strings.Contains(line, "function ") {
				// JavaScript function
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "function" && i+1 < len(parts) {
						return parts[i+1]
					}
				}
			} else if strings.Contains(line, "def ") {
				// Python function
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "def" && i+1 < len(parts) {
						funcName := parts[i+1]
						// Remove parentheses
						funcName = strings.Split(funcName, "(")[0]
						return funcName
					}
				}
			}
		}
	}

	return ""
}
