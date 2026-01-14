// Phase 13: Schema Validation
// Validates knowledge items against JSON schema

package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// validateKnowledgeItem validates a knowledge item against the JSON schema
func validateKnowledgeItem(schemaPath string, itemJSON []byte) error {
	// Load schema
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewBytesLoader(itemJSON)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("failed to validate schema: %w", err)
	}

	if result.Valid() {
		return nil
	}

	// Convert validation errors to ValidationErrors
	var validationErrors ValidationErrors
	for _, desc := range result.Errors() {
		validationErrors = append(validationErrors, ValidationError{
			Field:   desc.Field(),
			Message: desc.Description(),
			Value:   desc.Value(),
		})
	}

	return validationErrors
}

// validateBusinessRule validates a business rule knowledge item
func validateBusinessRule(itemJSON []byte) error {
	schemaPath, err := filepath.Abs("schemas/knowledge_schema.json")
	if err != nil {
		return fmt.Errorf("failed to resolve schema path: %w", err)
	}
	return validateKnowledgeItem(schemaPath, itemJSON)
}

// validateEntity validates an entity knowledge item
func validateEntity(itemJSON []byte) error {
	schemaPath, err := filepath.Abs("schemas/knowledge_schema.json")
	if err != nil {
		return fmt.Errorf("failed to resolve schema path: %w", err)
	}
	return validateKnowledgeItem(schemaPath, itemJSON)
}

// validateAPIContract validates an API contract knowledge item
func validateAPIContract(itemJSON []byte) error {
	schemaPath, err := filepath.Abs("schemas/knowledge_schema.json")
	if err != nil {
		return fmt.Errorf("failed to resolve schema path: %w", err)
	}
	return validateKnowledgeItem(schemaPath, itemJSON)
}

// validateUserJourney validates a user journey knowledge item
func validateUserJourney(itemJSON []byte) error {
	schemaPath, err := filepath.Abs("schemas/knowledge_schema.json")
	if err != nil {
		return fmt.Errorf("failed to resolve schema path: %w", err)
	}
	return validateKnowledgeItem(schemaPath, itemJSON)
}

// validateGlossary validates a glossary knowledge item
func validateGlossary(itemJSON []byte) error {
	schemaPath, err := filepath.Abs("schemas/knowledge_schema.json")
	if err != nil {
		return fmt.Errorf("failed to resolve schema path: %w", err)
	}
	return validateKnowledgeItem(schemaPath, itemJSON)
}

// validateByType validates a knowledge item by its type
func validateByType(itemType string, itemJSON []byte) error {
	switch itemType {
	case "business_rule":
		return validateBusinessRule(itemJSON)
	case "entity":
		return validateEntity(itemJSON)
	case "api_contract":
		return validateAPIContract(itemJSON)
	case "user_journey":
		return validateUserJourney(itemJSON)
	case "glossary":
		return validateGlossary(itemJSON)
	default:
		return fmt.Errorf("unknown knowledge item type: %s", itemType)
	}
}

// validateStructuredKnowledgeItem validates a StructuredKnowledgeItem
func validateStructuredKnowledgeItem(item *StructuredKnowledgeItem) error {
	// Marshal to JSON
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal knowledge item: %w", err)
	}

	// Validate by type
	return validateByType(item.Type, itemJSON)
}











