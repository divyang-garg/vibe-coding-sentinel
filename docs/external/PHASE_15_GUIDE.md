# Phase 15: Intent & Simple Language Guide

## Overview

Phase 15 adds intent analysis and simple language handling to gracefully handle unclear prompts. When a user provides a vague or ambiguous request, Sentinel analyzes the intent and generates clarifying questions to help understand what the user really wants. This improves the developer experience by reducing back-and-forth and supporting non-English speakers.

## Features

- **Intent Analysis**: Automatically detects unclear prompts and determines what clarification is needed
- **Simple Language Templates**: Pre-defined templates for common clarification scenarios
- **Context Gathering**: Collects relevant context (recent files, git status, business rules) to better understand intent
- **Decision Recording**: Records user choices to learn patterns and improve future suggestions
- **Pattern Learning**: Learns from past decisions to refine templates and suggestions
- **MCP Integration**: Available as `sentinel_check_intent` tool in Cursor IDE

## Use Cases

1. **Vague Prompts**: "Add something here" → Clarifies what to add and where
2. **Ambiguous Requests**: "Fix the bug" → Asks which bug and how to fix
3. **Location Unclear**: "Create a new file" → Asks where to create it
4. **Entity Unclear**: "Update the user" → Asks which user and what to update
5. **Action Confirmation**: "Delete this" → Confirms before destructive action

## Setup

### Prerequisites

- Sentinel binary built (`./sentinel` exists)
- Hub API running and accessible (optional, but required for full functionality)
- Cursor IDE installed (for MCP integration)

### Cursor MCP Configuration

If you haven't already configured MCP for Phase 14B, add the `sentinel_check_intent` tool to your existing configuration:

1. **Edit `~/.cursor/mcp.json`** (or create if it doesn't exist):

```json
{
  "mcpServers": {
    "sentinel": {
      "command": "/path/to/sentinel",
      "args": ["mcp-server"],
      "env": {
        "SENTINEL_HUB_URL": "http://localhost:8080",
        "SENTINEL_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

2. **Restart Cursor IDE** to load the new tool

### Environment Variables

```bash
export SENTINEL_HUB_URL="http://localhost:8080"
export SENTINEL_API_KEY="your-api-key"
```

## Usage

### From Cursor IDE

Use the `sentinel_check_intent` tool directly in Cursor chat:

**Example 1: Check if prompt needs clarification**
```
Use sentinel_check_intent to analyze: "add a button"
```

**Example 2: With context gathering**
```
Check intent for: "fix the error" with context
```

**Example 3: Specify codebase path**
```
Analyze intent for: "create a new component" in /path/to/project
```

### Tool Parameters

The `sentinel_check_intent` tool accepts:

- **prompt** (required): User prompt to analyze for clarity
- **codebasePath** (optional): Path to codebase root (for context gathering)
- **includeContext** (optional): Include context gathering (default: true)
  - Gathers recent files, git status, project structure, business rules

### API Usage

#### 1. Analyze Intent

**Endpoint**: `POST /api/v1/analyze/intent`

**Request**:
```json
{
  "prompt": "add a new feature",
  "codebasePath": "/path/to/project",
  "includeContext": true
}
```

**Response**:
```json
{
  "success": true,
  "requires_clarification": true,
  "intent_type": "location_unclear",
  "confidence": 0.75,
  "clarifying_question": "Where should this go?\n1. src/components/\n2. src/features/",
  "options": ["src/components/", "src/features/"]
}
```

#### 2. Record Decision

**Endpoint**: `POST /api/v1/intent/decisions`

**Request**:
```json
{
  "decision_id": "uuid-of-original-decision",
  "user_choice": "src/components/",
  "resolved_prompt": "add a new feature in src/components/",
  "additional_context": {
    "feature_name": "UserProfile"
  }
}
```

**Response**:
```json
{
  "success": true,
  "decision_id": "uuid",
  "message": "Decision recorded successfully"
}
```

#### 3. Get Learned Patterns

**Endpoint**: `GET /api/v1/intent/patterns?type=location_unclear&limit=10`

**Response**:
```json
{
  "success": true,
  "patterns": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "pattern_type": "location_unclear",
      "pattern_data": {
        "intent_type": "location_unclear",
        "user_choice": "src/components/"
      },
      "frequency": 5,
      "last_used": "2024-12-10T10:00:00Z",
      "created_at": "2024-12-01T10:00:00Z"
    }
  ],
  "count": 1
}
```

## Intent Types

### `location_unclear`
The prompt doesn't specify where something should be placed.

**Example**: "add a button" → "Where should this go? 1. src/components/ 2. src/features/"

### `entity_unclear`
The prompt doesn't specify which entity is being referenced.

**Example**: "update the user" → "Which user? 1. User model 2. User profile component"

### `action_confirm`
The prompt describes a potentially destructive action that needs confirmation.

**Example**: "delete this file" → "I will delete this file. Correct? [Y/n]"

### `ambiguous`
The prompt is too vague to determine specific intent.

**Example**: "fix it" → "What needs to be fixed? 1. Bug in login 2. Performance issue"

### `clear`
The prompt is clear and doesn't need clarification.

**Example**: "create src/components/UserProfile.tsx with name and email fields"

## Simple Language Templates

Templates are used to generate clarifying questions in a consistent, user-friendly format:

1. **Location Unclear Template**: "Where should this go?\n1. {option1}\n2. {option2}"
2. **Entity Unclear Template**: "Which {entity_type}?\n1. {option1}\n2. {option2}"
3. **Action Confirm Template**: "I will {action}. Correct? [Y/n]"

Templates are extensible and can be customized per project.

## Context Gathering

When `includeContext: true`, the system gathers:

1. **Recent Files**: Last 10 modified files from git or filesystem
2. **Git Status**: Current branch, modified files count
3. **Project Structure**: Common directories (src, lib, app, etc.), file extensions
4. **Business Rules**: Approved business rules from knowledge_items
5. **Code Patterns**: Common directory patterns from recent files

This context helps generate more relevant clarifying questions.

## Decision Recording & Learning

When a user makes a choice, the system:

1. Records the decision in `intent_decisions` table
2. Updates pattern frequency in `intent_patterns` table
3. Uses learned patterns to improve future suggestions

**Example**: If users consistently choose `src/components/` for UI components, future prompts will suggest this location first.

## Error Handling

### Hub Unavailable
If the Hub API is unavailable, the MCP tool returns:
```json
{
  "error": {
    "code": -32000,
    "message": "Hub unavailable",
    "data": {
      "fallback": "Hub is not reachable. Please check SENTINEL_HUB_URL and network connectivity."
    }
  }
}
```

### Invalid Parameters
If required parameters are missing:
```json
{
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "prompt is required and must be a string"
  }
}
```

### Timeout
If the request times out:
```json
{
  "error": {
    "code": -32001,
    "message": "Hub request timeout",
    "data": {
      "fallback": "Analysis timed out. Please try again."
    }
  }
}
```

## Troubleshooting

### Tool Not Available in Cursor

1. **Check MCP configuration**: Verify `~/.cursor/mcp.json` is correct
2. **Restart Cursor**: MCP servers are loaded on startup
3. **Check binary path**: Ensure the path to `sentinel` is correct
4. **Check permissions**: Ensure `sentinel` is executable (`chmod +x sentinel`)

### Hub API Errors

1. **Check Hub URL**: Verify `SENTINEL_HUB_URL` is correct
2. **Check API Key**: Verify `SENTINEL_API_KEY` is valid
3. **Check Hub Status**: Ensure Hub API is running (`curl http://localhost:8080/health`)
4. **Check Network**: Ensure network connectivity to Hub

### No Clarifying Questions Generated

1. **Check prompt clarity**: Very clear prompts won't generate questions
2. **Check LLM configuration**: Ensure LLM is configured in Hub (for LLM-based analysis)
3. **Check context**: Try with `includeContext: true` for better results

## Examples

### Example 1: Location Unclear

**User Prompt**: "add a new component"

**Response**:
```json
{
  "requires_clarification": true,
  "intent_type": "location_unclear",
  "confidence": 0.8,
  "clarifying_question": "Where should this go?\n1. src/components/\n2. src/features/",
  "options": ["src/components/", "src/features/"]
}
```

### Example 2: Clear Prompt

**User Prompt**: "create src/components/UserProfile.tsx with name and email fields"

**Response**:
```json
{
  "requires_clarification": false,
  "intent_type": "clear",
  "confidence": 1.0,
  "suggested_action": "create src/components/UserProfile.tsx with name and email fields",
  "resolved_prompt": "create src/components/UserProfile.tsx with name and email fields"
}
```

### Example 3: Entity Unclear

**User Prompt**: "update the user"

**Response**:
```json
{
  "requires_clarification": true,
  "intent_type": "entity_unclear",
  "confidence": 0.7,
  "clarifying_question": "Which user?\n1. User model (database)\n2. User profile component",
  "options": ["User model (database)", "User profile component"]
}
```

## Best Practices

1. **Use Context**: Always enable `includeContext: true` for better results
2. **Record Decisions**: Record user choices to improve future suggestions
3. **Review Patterns**: Periodically review learned patterns via `/api/v1/intent/patterns`
4. **Customize Templates**: Adjust templates per project for better UX
5. **Monitor Performance**: Track intent analysis performance and adjust as needed

## Integration with Other Phases

- **Phase 14A**: Uses LLM integration for intent analysis
- **Phase 14B**: MCP integration enables Cursor IDE usage
- **Phase 12**: Uses business rules from knowledge_items for context
- **Phase 13**: Uses standardized knowledge schema for business rules

## Next Steps

- **Phase 14C**: Cost optimization can reduce LLM costs for intent analysis
- **Phase 14D**: Further cost optimization and caching improvements

## Support

For issues or questions:
1. Check this guide's troubleshooting section
2. Review Hub API logs for errors
3. Check MCP server logs (if running in verbose mode)
4. Open an issue on the project repository










