// Package extraction provides LLM-powered knowledge extraction
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package extraction

import "fmt"

// PromptBuilder constructs LLM prompts for extraction
type PromptBuilder interface {
	BuildBusinessRulesPrompt(text string) string
	BuildEntitiesPrompt(text string) string
	BuildAPIContractsPrompt(text string) string
	BuildUserJourneysPrompt(text string) string
	BuildGlossaryPrompt(text string) string
}

// promptBuilder implements PromptBuilder
type promptBuilder struct{}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() PromptBuilder {
	return &promptBuilder{}
}

// BuildBusinessRulesPrompt creates a prompt for business rule extraction
func (p *promptBuilder) BuildBusinessRulesPrompt(text string) string {
	return fmt.Sprintf(`You are extracting business rules from a project document.

For EACH business rule found, extract:

1. IDENTIFY rules that describe what the system MUST or MUST NOT do
2. Extract the trigger (what initiates this rule)
3. Extract preconditions (what must be true before)
4. Extract constraints with EXACT conditions:
   - For numeric values: specify boundary (< vs <=)
   - For time: specify units and reference point
5. Extract exceptions (who/what is exempt)
6. Extract error cases (what can go wrong)

For EVERY constraint, write pseudocode that can be verified.
If AMBIGUOUS: flag as needs_clarification.

OUTPUT FORMAT (strict JSON):
{
  "business_rules": [
    {
      "id": "BR-XXX",
      "version": "1.0",
      "status": "draft",
      "title": "Short descriptive title",
      "description": "Detailed description",
      "priority": "high|medium|low",
      "specification": {
        "trigger": "What initiates this rule",
        "preconditions": ["condition1", "condition2"],
        "constraints": [
          {
            "id": "C1",
            "type": "time_based|value_based|state_based",
            "expression": "Human readable expression",
            "pseudocode": "machine_parseable_expression",
            "boundary": "inclusive|exclusive",
            "unit": "hours|minutes|days|currency|count"
          }
        ],
        "exceptions": [
          {
            "id": "E1",
            "condition": "When exception applies",
            "modified_constraint": "How constraint changes"
          }
        ],
        "error_cases": [
          {
            "condition": "When error occurs",
            "error_code": "ERR_SNAKE_CASE",
            "error_message": "Human readable message",
            "http_status": 400
          }
        ]
      },
      "traceability": {
        "source_document": "document name",
        "source_quote": "original text from document"
      },
      "confidence": 0.85
    }
  ]
}

DOCUMENT TEXT:
%s

Return ONLY valid JSON. Do not include markdown code fences.`, text)
}

// BuildEntitiesPrompt creates a prompt for entity extraction
func (p *promptBuilder) BuildEntitiesPrompt(text string) string {
	return fmt.Sprintf(`Extract entities from the document. For each entity: name, type (domain_entity|value_object|aggregate_root), fields with types/validation, relationships, traceability.

OUTPUT (strict JSON):
{"entities":[{"id":"ENT-XXX","version":"1.0","status":"draft","name":"EntityName","description":"...","fields":[{"name":"field","type":"string","required":true}],"relationships":[{"entity":"Related","type":"one-to-many"}],"traceability":{"source_document":"..."}}]}

DOCUMENT:
%s

Return ONLY valid JSON, no markdown fences.`, text)
}

// BuildAPIContractsPrompt creates a prompt for API contract extraction
func (p *promptBuilder) BuildAPIContractsPrompt(text string) string {
	return fmt.Sprintf(`Extract API contracts. For each endpoint: path, method (GET|POST|PUT|PATCH|DELETE), request (params/query/body), response status codes, traceability.

OUTPUT (strict JSON):
{"api_contracts":[{"id":"API-XXX","version":"1.0","status":"draft","endpoint":"/api/path","method":"GET","description":"...","request":{"params":{}},"response":{"status_codes":{"200":{"description":"Success"}}},"traceability":{"source_document":"..."}}]}

DOCUMENT:
%s

Return ONLY valid JSON, no markdown fences.`, text)
}

// BuildUserJourneysPrompt creates a prompt for user journey extraction
func (p *promptBuilder) BuildUserJourneysPrompt(text string) string {
	return fmt.Sprintf(`Extract user journeys. For each: name, actor (user role), goal, preconditions, sequential steps (actor action, system response, validation), postconditions, related rules/APIs, traceability.

OUTPUT (strict JSON):
{"user_journeys":[{"id":"UJ-XXX","version":"1.0","status":"draft","name":"Journey","actor":"User","goal":"...","description":"...","preconditions":[],"steps":[{"step":1,"actor_action":"...","system_response":"..."}],"postconditions":[],"traceability":{"source_document":"..."}}]}

DOCUMENT:
%s

Return ONLY valid JSON, no markdown fences.`, text)
}

// BuildGlossaryPrompt creates a prompt for glossary term extraction
func (p *promptBuilder) BuildGlossaryPrompt(text string) string {
	return fmt.Sprintf(`Extract glossary terms. For each: term name, definition, synonyms, related terms, examples, context, traceability.

OUTPUT (strict JSON):
{"glossary":[{"id":"GL-XXX","term":"Term","definition":"...","synonyms":[],"related_terms":[],"examples":[],"context":"...","traceability":{"source_document":"..."}}]}

DOCUMENT:
%s

Return ONLY valid JSON, no markdown fences.`, text)
}
