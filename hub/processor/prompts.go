// Phase 13: Enhanced Extraction Prompts
// Provides structured prompts for LLM knowledge extraction

package main

import (
	"fmt"
	"strings"
)

// PromptTemplate represents a prompt template with system and user prompts
type PromptTemplate struct {
	SystemPrompt string
	UserPrompt   string
}

// getBusinessRulePrompt returns an enhanced prompt for extracting business rules
func getBusinessRulePrompt(text string) PromptTemplate {
	systemPrompt := `You are an expert at analyzing business documents and extracting structured business rules. 
You must extract business rules following the standardized JSON schema exactly.
Return only valid JSON arrays. Each business rule must include:
- Constraints with pseudocode and boundary specification (inclusive/exclusive)
- Test requirements (minimum 2: happy_path + error_case)
- Traceability information (source document, section, page, quote)
- Ambiguity flags if any constraint is unclear`

	userPrompt := fmt.Sprintf(`Extract business rules from this document. A business rule is a conditional statement about how the business operates.

<document>
%s
</document>

For EACH business rule found:

1. IDENTIFY the rule:
   - Trigger: What initiates this rule?
   - Preconditions: What must be true before?
   - Constraints: What are the EXACT conditions?
     - For numeric values: specify boundary (< vs <=)
     - For time: specify units and reference point
   - Exceptions: Who/what is exempt?
   - Side effects: What else must happen?
   - Error cases: What can go wrong?

2. For EVERY constraint:
   - Write pseudocode that can be verified
   - Specify if boundary is inclusive (<=) or exclusive (<)
   - If AMBIGUOUS: flag as "NEEDS_CLARIFICATION" and list interpretations

3. Generate TEST REQUIREMENTS:
   - Minimum: happy_path + error_case for each rule
   - Include: boundary tests for numeric constraints
   - Include: exception tests if exceptions exist

4. Add TRACEABILITY:
   - Source document name
   - Section/page number
   - Quote the original text

5. If any constraint is ambiguous, add to ambiguity_flags array with:
   - Field name
   - Possible interpretations
   - Clarification question

Return a JSON array with this exact format:
[
  {
    "id": "BR-001",
    "version": "1.0.0",
    "status": "active",
    "title": "Rule Name",
    "description": "Detailed rule description",
    "category": "orders",
    "priority": "high",
    "specification": {
      "trigger": "What initiates this rule",
      "preconditions": ["Condition 1", "Condition 2"],
      "constraints": [
        {
          "id": "C1",
          "type": "time_based",
          "expression": "Human readable expression",
          "pseudocode": "Date.now() - order.createdAt < 24 * 60 * 60 * 1000",
          "boundary": "exclusive",
          "unit": "hours"
        }
      ],
      "exceptions": [
        {
          "id": "E1",
          "condition": "When exception applies",
          "modified_constraint": "How constraint changes",
          "applies_to": ["user.tier === 'premium'"],
          "source": "Document reference"
        }
      ],
      "side_effects": [
        {
          "action": "Function/operation to trigger",
          "condition": "When to trigger",
          "parameters": {},
          "async": true,
          "required": true
        }
      ],
      "error_cases": [
        {
          "condition": "When error occurs",
          "error_code": "ERR_SNAKE_CASE_NAME",
          "error_message": "Human readable message",
          "http_status": 400,
          "recoverable": false
        }
      ]
    },
    "test_requirements": [
      {
        "id": "BR-001-T1",
        "name": "test_happy_path",
        "type": "happy_path",
        "priority": "critical",
        "scenario": "Description of what is being tested",
        "setup": {
          "entities": {},
          "state": {}
        },
        "action": "Function call or API request",
        "expected": {
          "success": true,
          "return_value": {},
          "side_effects": [],
          "error": null
        },
        "assertions_required": [
          "expect(result.success).toBe(true)"
        ]
      },
      {
        "id": "BR-001-T2",
        "name": "test_error_case",
        "type": "error_case",
        "priority": "critical",
        "scenario": "Error scenario description",
        "expected": {
          "success": false,
          "error": "ERR_ERROR_CODE"
        }
      }
    ],
    "traceability": {
      "source_document": "filename",
      "source_section": "Section number or name",
      "source_page": 12,
      "source_quote": "Original text from document"
    },
    "ambiguity_flags": []
  }
]

Only include clear, actionable business rules. Return empty array [] if none found.
For numeric constraints, ALWAYS specify if boundary is inclusive (<=) or exclusive (<).
If a constraint is ambiguous, flag it in ambiguity_flags with possible interpretations.`, text)

	return PromptTemplate{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
}

// getEntityPrompt returns a prompt for extracting entities
func getEntityPrompt(text string) PromptTemplate {
	systemPrompt := `You are an expert at analyzing business documents and extracting domain entities. 
Return only valid JSON arrays following the standardized schema.`

	userPrompt := fmt.Sprintf(`Extract business entities from this document. An entity is a key object in the business domain with specific attributes.

<document>
%s
</document>

Return a JSON array with this exact format:
[
  {
    "id": "ENT-001",
    "version": "1.0.0",
    "status": "active",
    "name": "EntityName",
    "description": "Entity description",
    "category": "domain category",
    "fields": [
      {
        "name": "fieldName",
        "type": "string",
        "required": true,
        "unique": false,
        "description": "Field description",
        "validation": {
          "pattern": "regex pattern",
          "minLength": 1,
          "maxLength": 255
        }
      }
    ],
    "relationships": [
      {
        "entity": "RelatedEntity",
        "type": "one-to-many",
        "foreign_key": "fieldName",
        "cascade": "delete"
      }
    ],
    "traceability": {
      "source_document": "filename",
      "source_section": "section"
    }
  }
]

Only include clear entities. Return empty array [] if none found.`, text)

	return PromptTemplate{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
}

// getGlossaryPrompt returns a prompt for extracting glossary terms
func getGlossaryPrompt(text string) PromptTemplate {
	systemPrompt := `You are an expert at analyzing business documents and extracting glossary terms. 
Return only valid JSON arrays.`

	userPrompt := fmt.Sprintf(`Extract glossary terms from this document. A glossary term is a domain-specific word or phrase that needs definition.

<document>
%s
</document>

Return a JSON array with this exact format:
[
  {
    "id": "GL-001",
    "term": "Term",
    "definition": "Clear definition",
    "context": "When this term applies",
    "related_terms": ["GL-002"],
    "examples": ["Usage example"],
    "traceability": {
      "source_document": "filename",
      "source_section": "section"
    }
  }
]

Only include clear definitions. Return empty array [] if none found.`, text)

	return PromptTemplate{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
}

// getAPIContractPrompt returns a prompt for extracting API contracts
func getAPIContractPrompt(text string) PromptTemplate {
	systemPrompt := `You are an expert at analyzing API documentation and extracting API contracts. 
Return only valid JSON arrays following the standardized schema.`

	userPrompt := fmt.Sprintf(`Extract API contracts from this document. An API contract defines an endpoint, method, request/response format.

<document>
%s
</document>

Return a JSON array with this exact format:
[
  {
    "id": "API-001",
    "version": "1.0.0",
    "status": "active",
    "endpoint": "/api/path/:param",
    "method": "POST",
    "description": "What this endpoint does",
    "authentication": {
      "required": true,
      "type": "bearer",
      "scopes": ["read:orders"]
    },
    "request": {
      "params": {},
      "body": {}
    },
    "response": {
      "200": {
        "description": "Success response",
        "schema": {}
      }
    },
    "implements_rules": ["BR-001"],
    "traceability": {
      "source_document": "filename"
    }
  }
]

Only include clear API contracts. Return empty array [] if none found.`, text)

	return PromptTemplate{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
}

// getUserJourneyPrompt returns a prompt for extracting user journeys
func getUserJourneyPrompt(text string) PromptTemplate {
	systemPrompt := `You are an expert at analyzing user workflows and extracting user journeys. 
Return only valid JSON arrays following the standardized schema.`

	userPrompt := fmt.Sprintf(`Extract user journeys from this document. A user journey is a sequence of steps a user takes to achieve a goal.

<document>
%s
</document>

Return a JSON array with this exact format:
[
  {
    "id": "UJ-001",
    "version": "1.0.0",
    "status": "active",
    "name": "Journey Name",
    "actor": "User role",
    "goal": "What the user wants to achieve",
    "description": "Journey description",
    "preconditions": ["Condition 1"],
    "steps": [
      {
        "id": "S1",
        "action": "What the user does",
        "system_response": "What the system does",
        "decision_points": [
          {
            "condition": "When this decision applies",
            "next_step": "S2"
          }
        ]
      }
    ],
    "postconditions": ["Condition 1"],
    "implements_rules": ["BR-001"],
    "traceability": {
      "source_document": "filename"
    }
  }
]

Only include clear user journeys. Return empty array [] if none found.`, text)

	return PromptTemplate{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
}

// getPromptByType returns the appropriate prompt template based on knowledge type
func getPromptByType(extractType string, text string) PromptTemplate {
	switch extractType {
	case "business_rule":
		return getBusinessRulePrompt(text)
	case "entity":
		return getEntityPrompt(text)
	case "glossary":
		return getGlossaryPrompt(text)
	case "api_contract":
		return getAPIContractPrompt(text)
	case "user_journey":
		return getUserJourneyPrompt(text)
	default:
		// Default to business rule prompt
		return getBusinessRulePrompt(text)
	}
}

// formatPromptForProvider formats a prompt template for a specific LLM provider
func formatPromptForProvider(template PromptTemplate, providerName string) (string, string) {
	// For Azure AI Foundry, use system/user message format
	if strings.Contains(providerName, "azure") {
		return template.SystemPrompt, template.UserPrompt
	}
	
	// For Ollama, combine into single prompt
	combinedPrompt := template.SystemPrompt + "\n\n" + template.UserPrompt
	return "", combinedPrompt
}











