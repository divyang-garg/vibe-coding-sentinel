# Comprehensive Feature Analysis Solution

> **For AI Agents**: This document specifies the complete solution for comprehensive feature analysis across all layers of an application. This addresses the critical gap where vibe coding misses important aspects during analysis.

## Executive Summary

**Problem**: Vibe coding often performs only surface-level checks, missing critical aspects like end-to-end flows, cross-layer integration, business logic correctness, and comprehensive validation.

**Solution**: A mandatory comprehensive analysis system that automatically discovers features, analyzes all 7 layers (Business, UI, API, Database, Logic, Integration, Tests), verifies end-to-end flows, and provides actionable checklists.

**Key Innovation**: Feature-level analysis with automatic discovery, layer-specific deep analysis, and business context integration.

---

## 1. Problem Statement

### Current Limitations

1. **Surface-Level Analysis**: Only checks immediate code, not full context
2. **Missing Cross-Layer Verification**: Doesn't verify UI → API → Database → Integration flows
3. **No Business Context**: Doesn't validate against business rules and requirements
4. **Incomplete Integration Checks**: Misses external API contracts and side effects
5. **No End-to-End Validation**: Doesn't verify complete user journeys
6. **Missing Edge Cases**: Doesn't identify missing error handling or edge cases

### Real-World Impact

- **Production Bugs**: Features work in isolation but fail in production
- **Integration Failures**: Components work but don't integrate correctly
- **Business Logic Errors**: Code works but violates business rules
- **Security Gaps**: Security checks exist but miss edge cases
- **User Experience Issues**: UI works but doesn't match business requirements

---

## 2. Solution Overview

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    COMPREHENSIVE ANALYSIS FLOW                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Cursor (MCP Request)                                             │
│      │                                                            │
│      ▼                                                            │
│  Agent: sentinel_analyze_feature_comprehensive                    │
│      │                                                            │
│      ├── Feature Discovery (Auto/Manual)                          │
│      │   ├── UI Components                                        │
│      │   ├── API Endpoints                                        │
│      │   ├── Database Tables                                      │
│      │   ├── Business Logic Functions                             │
│      │   ├── Integration Points                                   │
│      │   └── Test Files                                           │
│      │                                                            │
│      ▼                                                            │
│  Hub: POST /api/v1/analyze/comprehensive                          │
│      │                                                            │
│      ├── Layer-Specific Analysis (7 Layers)                       │
│      │   ├── Business Context (Rules, Journeys, Entities)         │
│      │   ├── UI Layer (Components, Forms, Validation)             │
│      │   ├── API Layer (Endpoints, Security, Middleware)         │
│      │   ├── Database Layer (Schema, Migrations, Integrity)      │
│      │   ├── Business Logic (AST, Cross-File, Semantic)          │
│      │   ├── Integration Layer (External APIs, Contracts)         │
│      │   └── Test Layer (Coverage, Quality, Edge Cases)          │
│      │                                                            │
│      ├── End-to-End Flow Verification                             │
│      │   ├── Flow Detection Across Layers                         │
│      │   ├── Breakpoint Identification                            │
│      │   └── Integration Verification                             │
│      │                                                            │
│      ├── LLM Semantic Analysis                                    │
│      │   ├── Business Logic Correctness                          │
│      │   ├── Requirement Compliance                               │
│      │   └── Edge Case Identification                            │
│      │                                                            │
│      ▼                                                            │
│  Result Aggregation & Checklist Generation                        │
│      │                                                            │
│      ▼                                                            │
│  Hub Storage + URL Generation                                      │
│      │                                                            │
│      ▼                                                            │
│  Agent Response to Cursor                                          │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Key Components

1. **Feature Discovery Engine**: Auto-discovers features across all layers
2. **Layer-Specific Analyzers**: 7 specialized analyzers for each layer
3. **End-to-End Flow Verifier**: Validates complete user journeys
4. **LLM Integration Layer**: Semantic analysis with business context
5. **Result Aggregator**: Combines findings into actionable checklists
6. **Hub Storage**: Stores results for reference and trending

---

## 3. Feature Discovery Algorithm

### Auto-Discovery

**Purpose**: Automatically identify all components of a feature across layers.

**Algorithm**:

1. **UI Layer Discovery**:
   - Scan for component files (React, Vue, Angular, etc.)
   - Identify form components, buttons, modals
   - Extract component names and props
   - Map to feature keywords

2. **API Layer Discovery**:
   - Scan route definitions (Express, FastAPI, Django, etc.)
   - Identify endpoints matching feature keywords
   - Extract HTTP methods, paths, handlers
   - Map to UI components via naming conventions

3. **Database Layer Discovery**:
   - Scan migration files and schema definitions
   - Identify tables matching feature keywords
   - Extract relationships and constraints
   - Map to API endpoints via naming conventions

4. **Business Logic Discovery**:
   - Scan service/domain layer files
   - Identify functions matching feature keywords
   - Extract function signatures and dependencies
   - Map to API handlers and database operations

5. **Integration Layer Discovery**:
   - Scan for external API calls (HTTP clients, SDKs)
   - Identify third-party service integrations
   - Extract API contracts and configurations
   - Map to business logic functions

6. **Test Layer Discovery**:
   - Scan test files matching feature keywords
   - Identify test cases and coverage
   - Extract test scenarios and assertions
   - Map to components being tested

**Example**:
```
Feature: "Order Cancellation"
    │
    ├── UI: CancelButton.tsx, OrderDetails.tsx
    ├── API: DELETE /api/orders/:id, POST /api/orders/:id/cancel
    ├── Database: orders table, order_status enum
    ├── Logic: cancelOrder(), refundOrder(), restoreInventory()
    ├── Integration: PaymentGateway.refund(), WarehouseAPI.cancel()
    └── Tests: OrderCancellation.test.ts, OrderService.test.ts
```

### Manual File Specification

**Purpose**: Allow developers to explicitly specify files for analysis.

**Format**:
```json
{
  "feature": "Order Cancellation",
  "files": {
    "ui": ["src/components/OrderDetails.tsx", "src/components/CancelButton.tsx"],
    "api": ["src/routes/orders.ts", "src/controllers/orderController.ts"],
    "database": ["migrations/001_create_orders.sql", "models/Order.ts"],
    "logic": ["src/services/orderService.ts", "src/services/refundService.ts"],
    "integration": ["src/integrations/paymentGateway.ts", "src/integrations/warehouse.ts"],
    "tests": ["tests/orderCancellation.test.ts", "tests/orderService.test.ts"]
  }
}
```

---

## 4. Layer-Specific Analysis

### 4.1 Business Context Analysis

**Purpose**: Validate feature against business rules, user journeys, and entity definitions.

**Checks**:
- ✅ Business rules compliance (from knowledge base)
- ✅ User journey adherence (from knowledge base)
- ✅ Entity validation (from knowledge base)
- ✅ Business logic correctness (semantic analysis)
- ✅ Requirement coverage (from documentation)

**Sources**:
- `docs/knowledge/business-rules.json`
- `docs/knowledge/user-journeys.json`
- `docs/knowledge/entities.json`
- `docs/knowledge/requirements.json`

**Example**:
```json
{
  "layer": "business",
  "findings": [
    {
      "type": "business_rule_violation",
      "rule": "Refunds must be processed within 5 business days",
      "location": "src/services/refundService.ts:45",
      "issue": "No time limit check in refund processing",
      "severity": "critical"
    },
    {
      "type": "user_journey_mismatch",
      "journey": "Order Cancellation",
      "location": "src/components/CancelButton.tsx",
      "issue": "Missing confirmation step before cancellation",
      "severity": "high"
    }
  ]
}
```

### 4.2 UI Layer Analysis

**Purpose**: Validate UI components, forms, validation, and user experience.

**Checks**:
- ✅ Component structure and props
- ✅ Form validation (client-side)
- ✅ Error handling and display
- ✅ Loading states and feedback
- ✅ Accessibility (a11y) compliance
- ✅ Responsive design
- ✅ State management

**Example**:
```json
{
  "layer": "ui",
  "findings": [
    {
      "type": "missing_validation",
      "location": "src/components/CancelButton.tsx:12",
      "issue": "No confirmation dialog before cancellation",
      "severity": "high"
    },
    {
      "type": "missing_error_handling",
      "location": "src/components/OrderDetails.tsx:45",
      "issue": "No error message display for failed cancellation",
      "severity": "medium"
    }
  ]
}
```

### 4.3 API Layer Analysis

**Purpose**: Validate API endpoints, security, middleware, and contracts.

**Checks**:
- ✅ Endpoint definition and routing
- ✅ Authentication and authorization
- ✅ Input validation (server-side)
- ✅ Error handling and status codes
- ✅ Rate limiting and throttling
- ✅ API contract compliance (OpenAPI/Swagger)
- ✅ Middleware application

**Example**:
```json
{
  "layer": "api",
  "findings": [
    {
      "type": "missing_auth",
      "location": "src/routes/orders.ts:45",
      "issue": "DELETE /api/orders/:id endpoint missing authentication middleware",
      "severity": "critical"
    },
    {
      "type": "missing_validation",
      "location": "src/controllers/orderController.ts:78",
      "issue": "No input validation for order ID parameter",
      "severity": "high"
    }
  ]
}
```

### 4.4 Database Layer Analysis

**Purpose**: Validate database schema, migrations, and data integrity.

**Checks**:
- ✅ Schema definition and relationships
- ✅ Migration consistency
- ✅ Foreign key constraints
- ✅ Index optimization
- ✅ Data integrity rules
- ✅ Transaction handling

**Example**:
```json
{
  "layer": "database",
  "findings": [
    {
      "type": "missing_constraint",
      "location": "migrations/001_create_orders.sql:12",
      "issue": "orders table missing foreign key constraint on user_id",
      "severity": "high"
    },
    {
      "type": "missing_index",
      "location": "migrations/001_create_orders.sql:15",
      "issue": "orders table missing index on status column (used in WHERE clauses)",
      "severity": "medium"
    }
  ]
}
```

### 4.5 Business Logic Analysis

**Purpose**: Validate business logic correctness, cross-file dependencies, and semantic issues.

**Checks**:
- ✅ AST-based code analysis (duplicates, orphaned code, unused variables)
- ✅ Cross-file reference resolution
- ✅ Function signature mismatches
- ✅ Import/export mismatches
- ✅ Business logic correctness (LLM semantic analysis)
- ✅ Error handling completeness

**Example**:
```json
{
  "layer": "logic",
  "findings": [
    {
      "type": "semantic_error",
      "location": "src/services/refundService.ts:45",
      "issue": "Refund amount calculation doesn't account for partial refunds",
      "severity": "critical"
    },
    {
      "type": "missing_error_handling",
      "location": "src/services/orderService.ts:78",
      "issue": "cancelOrder() doesn't handle case where order is already cancelled",
      "severity": "high"
    }
  ]
}
```

### 4.6 Integration Layer Analysis

**Purpose**: Validate external API integrations, contracts, and side effects.

**Checks**:
- ✅ External API calls and contracts
- ✅ Error handling for external APIs
- ✅ Retry logic and timeouts
- ✅ API contract compliance
- ✅ Side effect verification (e.g., payment gateway refunds)
- ✅ Integration test coverage

**Important Clarification**: When analyzing features like "order cancellation", Sentinel identifies payment gateway integrations as part of the FEATURE being analyzed. This is analysis of the FEATURE's integrations, not Sentinel's own functionality.

**Example**:
```json
{
  "layer": "integration",
  "findings": [
    {
      "type": "missing_error_handling",
      "location": "src/integrations/paymentGateway.ts:45",
      "issue": "PaymentGateway.refund() call missing error handling for failed refunds",
      "severity": "critical"
    },
    {
      "type": "missing_retry",
      "location": "src/integrations/paymentGateway.ts:52",
      "issue": "PaymentGateway.refund() call missing retry logic for transient failures",
      "severity": "high"
    },
    {
      "type": "contract_mismatch",
      "location": "src/integrations/paymentGateway.ts:45",
      "issue": "Refund API call doesn't match documented contract (missing refund_reason parameter)",
      "severity": "medium"
    }
  ]
}
```

**Note**: Sentinel does NOT process payments, handle refunds, or integrate with payment gateways. It only analyzes whether the feature correctly integrates with payment gateways.

### 4.7 Test Layer Analysis

**Purpose**: Validate test coverage, quality, and edge case handling.

**Checks**:
- ✅ Test coverage (unit, integration, e2e)
- ✅ Test quality (assertions, scenarios)
- ✅ Edge case coverage
- ✅ Error scenario testing
- ✅ Integration test coverage
- ✅ Test correctness (mutation testing)

**Example**:
```json
{
  "layer": "tests",
  "findings": [
    {
      "type": "missing_coverage",
      "location": "tests/orderCancellation.test.ts",
      "issue": "No test for cancellation of already-cancelled order",
      "severity": "high"
    },
    {
      "type": "weak_assertion",
      "location": "tests/orderService.test.ts:45",
      "issue": "Test only checks return value, doesn't verify side effects (inventory restoration)",
      "severity": "medium"
    }
  ]
}
```

---

## 5. End-to-End Flow Verification

### Flow Detection

**Purpose**: Identify complete user journeys across all layers.

**Algorithm**:
1. Start from UI component (user action)
2. Trace to API endpoint
3. Trace to business logic function
4. Trace to database operations
5. Trace to integration calls
6. Verify return path (response → UI update)

**Example Flow**:
```
User clicks "Cancel Order" button (UI)
    │
    ▼
POST /api/orders/:id/cancel (API)
    │
    ▼
cancelOrder(orderId) (Business Logic)
    │
    ├── Check order status (Database)
    ├── Call PaymentGateway.refund() (Integration)
    ├── Call WarehouseAPI.cancel() (Integration)
    ├── Update order status (Database)
    └── Send notification (Integration)
    │
    ▼
Return success response (API)
    │
    ▼
Update UI with success message (UI)
```

### Breakpoint Identification

**Purpose**: Identify points where flow might break.

**Checks**:
- ✅ Missing error handling at each step
- ✅ Missing validation at boundaries
- ✅ Missing rollback for partial failures
- ✅ Missing timeout handling
- ✅ Missing retry logic

### Integration Verification

**Purpose**: Verify all integration points are correctly connected.

**Checks**:
- ✅ API endpoints match UI calls
- ✅ Database operations match business logic
- ✅ Integration calls match business logic
- ✅ Response formats match expectations
- ✅ Error propagation is correct

---

## 6. LLM Integration

### API Key Management Model

**Principle**: Users/Organizations subscribe to LLM providers separately and provide API keys to Sentinel. Sentinel does NOT handle billing or payments.

**Architecture**:
```
User/Organization
    │
    ├── Option 1: Subscribes to OpenAI API
    │   └── Provides API key to Hub
    │
    ├── Option 2: Subscribes to Codex Pro ($200/month)
    │   └── Uses Codex Pro via Cursor (if API available, provides key)
    │
    └── Option 3: Organization provides shared API key
        └── Hub uses for all team members (organization pays)
            │
            ▼
    Hub stores API key (encrypted)
            │
            ▼
    Hub uses API key for LLM calls
            │
            ▼
    User/Organization pays LLM provider directly
```

### Dual Access Model

**Option 1: User-Provided API Keys (Recommended)**
- **How**: User subscribes to LLM provider, provides API key to Hub
- **Storage**: Encrypted in Hub database
- **Billing**: User pays LLM provider directly
- **Tracking**: Hub tracks usage for reporting only (not billing)
- **Best For**: Individual developers, small teams
- **Pros**: Simple, user controls costs, no payment processing
- **Cons**: Each user needs subscription

**Option 2: Organization-Shared API Key (Optional)**
- **How**: Organization subscribes, provides shared key
- **Storage**: Encrypted in Hub database
- **Billing**: Organization pays LLM provider directly
- **Tracking**: Hub tracks usage per user/project for allocation
- **Best For**: Large organizations, centralized billing
- **Pros**: Centralized billing, cost allocation
- **Cons**: Organization manages key distribution

**Option 3: Codex Pro Subscription (For Individuals)**
- **How**: User has ChatGPT Pro subscription ($200/month)
- **Usage**: Codex Pro used via Cursor (if API available)
- **Fallback**: Sentinel uses API access if Codex Pro unavailable
- **Best For**: Individual developers with Codex Pro
- **Pros**: Convenient if already subscribed
- **Cons**: Usage limits, may need API fallback

### Model Selection Strategy

**Critical Tasks** (Use High-Accuracy Models):
- Business logic correctness analysis
- End-to-end flow verification
- Security analysis
- Requirement compliance checking
- **Models**: GPT-5.1-Codex-Max, GPT-5.1 Thinking

**Non-Critical Tasks** (Use Cheaper Models):
- Simple pattern matching
- Basic validation checks
- Formatting verification
- **Models**: GPT-5.1 Instant, codex-mini-latest

**Progressive Depth**:
1. **Level 1**: Fast checks (pattern-based, AST-based) - No LLM
2. **Level 2**: Medium-depth analysis (cheaper models) - GPT-5.1 Instant
3. **Level 3**: Deep analysis (high-accuracy models) - GPT-5.1-Codex-Max

### Cost Optimization

**Caching Strategy**:
- Cache analysis results by file hash
- Cache business context queries
- Cache LLM responses for identical code patterns
- **Target**: 70% cache hit rate

**Progressive Depth**:
- Start with fast checks (no LLM)
- Only use LLM for complex analysis
- Skip LLM if AST/pattern analysis finds issues
- **Target**: 50% reduction in LLM calls

**Smart Model Selection**:
- Use cheaper models for simple tasks
- Use expensive models only for critical tasks
- **Target**: 40% cost reduction

### Token Usage Tracking

**Purpose**: Track usage for reporting and optimization, NOT for billing.

**What's Tracked**:
- Token usage per analysis
- Cost per analysis (calculated, not charged)
- Usage per project/user
- Trends and patterns

**What's NOT Done**:
- ❌ No billing to users
- ❌ No payment processing
- ❌ No credit system
- ❌ No refunds (users pay providers directly)

**Usage Reports**:
```json
{
  "period": "2024-12",
  "totalAnalyses": 200,
  "totalTokens": 1,280,000,
  "estimatedCost": 12.80,
  "byProject": {
    "project-1": {
      "analyses": 100,
      "tokens": 640,000,
      "estimatedCost": 6.40
    }
  },
  "byUser": {
    "user-1": {
      "analyses": 50,
      "tokens": 320,000,
      "estimatedCost": 3.20
    }
  }
}
```

### Security Considerations

**API Key Storage**:
- Encrypted at rest (AES-256)
- Encrypted in transit (TLS)
- Never logged
- Never exposed in responses
- Rotatable via Hub interface

**Access Control**:
- Only Hub administrators can configure API keys
- Users cannot see API keys (only test connection)
- API keys scoped to organization/project

---

## 7. Configuration System

### Hub-Based Configuration Interface

**Location**: Hub Dashboard → Settings → LLM Configuration

**UI Mockup**:
```
┌─────────────────────────────────────────────────────────────┐
│  LLM PROVIDER CONFIGURATION                                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Provider Selection                                          │
│  ─────────────────                                           │
│  Provider: [OpenAI ▼]  (OpenAI, Anthropic, Azure)          │
│  Model: [GPT-5.1-Codex-Max ▼]                               │
│                                                              │
│  API Key Management                                          │
│  ───────────────────                                         │
│  Key Type: [User-Provided ▼]  (User-Provided, Org-Shared)   │
│                                                              │
│  User-Provided API Key:                                       │
│  API Key: [••••••••••••••••] [Test Connection]              │
│  [✓] Encrypt key in database                                 │
│  Note: You must subscribe to OpenAI API separately.          │
│        Sentinel does not handle billing.                     │
│                                                              │
│  Organization-Shared API Key:                                │
│  API Key: [••••••••••••••••] [Test Connection]              │
│  [✓] Track usage per project/user                            │
│  Note: Organization pays LLM provider directly.                │
│                                                              │
│  Codex Pro Integration (Optional)                            │
│  ───────────────────────────                                │
│  [✓] Enable Codex Pro detection                               │
│  [✓] Fallback to API if Codex Pro unavailable               │
│  Note: Requires ChatGPT Pro subscription ($200/month).        │
│        Use Codex Pro via Cursor, API as fallback.            │
│                                                              │
│  Cost Optimization                                           │
│  ────────────────                                            │
│  [✓] Enable caching (70% target hit rate)                    │
│  [✓] Progressive depth (skip LLM when possible)              │
│  [✓] Smart model selection (cheaper for simple tasks)        │
│                                                              │
│  Cost Tracking (Reporting Only)                              │
│  ─────────────────────────────                               │
│  [✓] Track token usage per analysis                          │
│  [✓] Track costs per project/user                            │
│  [✓] Generate usage reports                                  │
│  Note: Tracking is for reporting only, not billing.          │
│                                                              │
│  [Save Configuration] [Test Connection]                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Configuration Schema

```json
{
  "llmProvider": {
    "type": "user-provided", // "user-provided", "organization-shared"
    "provider": "openai", // "openai", "anthropic", "azure"
    "apiKey": "encrypted_key", // Stored encrypted in Hub
    "model": "gpt-5.1-codex-max",
    "endpoint": "https://api.openai.com/v1",
    "codexPro": {
      "enabled": false, // Enable if user has Codex Pro
      "fallbackToAPI": true // Use API if Codex Pro unavailable
    },
    "usageTracking": {
      "enabled": true,
      "allocation": "per-project" // "per-project", "per-user", "none"
    },
    "costOptimization": {
      "caching": {
        "enabled": true,
        "targetHitRate": 0.70
      },
      "progressiveDepth": {
        "enabled": true,
        "skipLLMForPatternMatches": true
      },
      "modelSelection": {
        "enabled": true,
        "criticalTasks": ["gpt-5.1-codex-max"],
        "nonCriticalTasks": ["gpt-5.1-instant"]
      }
    }
  }
}
```

---

## 8. MCP Integration

### Tool: `sentinel_analyze_feature_comprehensive`

**Purpose**: Request comprehensive analysis of a feature across all layers.

**Request Format**:
```json
{
  "name": "sentinel_analyze_feature_comprehensive",
  "description": "Perform comprehensive analysis of a feature across all layers (UI, API, Database, Logic, Integration, Tests) with business context validation",
  "inputSchema": {
    "type": "object",
    "properties": {
      "feature": {
        "type": "string",
        "description": "Feature name or description (e.g., 'Order Cancellation')"
      },
      "mode": {
        "type": "string",
        "enum": ["auto", "manual"],
        "description": "Auto-discover feature components or use manual file specification",
        "default": "auto"
      },
      "files": {
        "type": "object",
        "description": "Manual file specification (required if mode='manual')",
        "properties": {
          "ui": {"type": "array", "items": {"type": "string"}},
          "api": {"type": "array", "items": {"type": "string"}},
          "database": {"type": "array", "items": {"type": "string"}},
          "logic": {"type": "array", "items": {"type": "string"}},
          "integration": {"type": "array", "items": {"type": "string"}},
          "tests": {"type": "array", "items": {"type": "string"}}
        }
      },
      "depth": {
        "type": "string",
        "enum": ["surface", "medium", "deep"],
        "description": "Analysis depth (surface=fast, medium=balanced, deep=comprehensive)",
        "default": "medium"
      },
      "includeBusinessContext": {
        "type": "boolean",
        "description": "Include business rules, journeys, and entities validation",
        "default": true
      }
    },
    "required": ["feature"]
  }
}
```

**Response Format**:
```json
{
  "validationId": "val_abc123",
  "feature": "Order Cancellation",
  "status": "completed",
  "hubUrl": "https://hub.example.com/validations/val_abc123",
  "summary": {
    "totalFindings": 12,
    "critical": 2,
    "high": 5,
    "medium": 3,
    "low": 2,
    "layersAnalyzed": 7,
    "flowsVerified": 3
  },
  "checklist": [
    {
      "id": "chk_001",
      "category": "business",
      "severity": "critical",
      "title": "Business rule violation: Refunds must be processed within 5 business days",
      "description": "Refund processing doesn't check time limit",
      "location": "src/services/refundService.ts:45",
      "remediation": "Add time limit check before processing refund",
      "autoFixable": false
    },
    {
      "id": "chk_002",
      "category": "api",
      "severity": "critical",
      "title": "Missing authentication on DELETE /api/orders/:id",
      "description": "Endpoint missing authentication middleware",
      "location": "src/routes/orders.ts:45",
      "remediation": "Add authentication middleware to route",
      "autoFixable": true
    }
  ],
  "layerAnalysis": {
    "business": {
      "findings": 2,
      "critical": 1,
      "high": 1
    },
    "ui": {
      "findings": 1,
      "high": 1
    },
    "api": {
      "findings": 2,
      "critical": 1,
      "high": 1
    },
    "database": {
      "findings": 1,
      "medium": 1
    },
    "logic": {
      "findings": 3,
      "high": 2,
      "medium": 1
    },
    "integration": {
      "findings": 2,
      "critical": 1,
      "high": 1
    },
    "tests": {
      "findings": 1,
      "high": 1
    }
  },
  "endToEndFlows": [
    {
      "flow": "Order Cancellation Flow",
      "status": "broken",
      "breakpoints": [
        {
          "layer": "api",
          "location": "src/routes/orders.ts:45",
          "issue": "Missing authentication"
        }
      ]
    }
  ]
}
```

**Error Handling**:
```json
{
  "error": {
    "code": "ANALYSIS_FAILED",
    "message": "Analysis failed: Hub unavailable",
    "fallback": "Using Cursor default auto mode",
    "details": "Hub health check failed: connection timeout"
  }
}
```

**Fallback Strategy**: If Hub is unavailable or analysis fails, Sentinel falls back to Cursor's default auto mode and notifies the user.

---

## 9. Results and Reporting

### Response Format

**Structure**:
- `validationId`: Unique ID for this analysis
- `hubUrl`: URL to view detailed results in Hub
- `summary`: High-level statistics
- `checklist`: Prioritized list of findings
- `layerAnalysis`: Findings per layer
- `endToEndFlows`: Flow verification results

### Checklist System

**Prioritization**:
1. **Critical**: Security issues, business rule violations, broken flows
2. **High**: Missing error handling, missing validation, missing tests
3. **Medium**: Performance issues, code quality, missing optimizations
4. **Low**: Style issues, documentation, minor improvements

**Format**:
- `id`: Unique checklist item ID
- `category`: Layer or category (business, api, ui, etc.)
- `severity`: critical, high, medium, low
- `title`: Short description
- `description`: Detailed explanation
- `location`: File and line number
- `remediation`: How to fix
- `autoFixable`: Whether Sentinel can auto-fix

### Hub Storage

**Storage**:
- Results stored in Hub database
- Accessible via Hub dashboard
- URL format: `https://hub.example.com/validations/{validationId}`

**Features**:
- View detailed findings
- Filter by layer, severity, category
- Track trends over time
- Export reports (JSON, PDF)
- Share with team members

---

## 10. Implementation Plan

### Phase 14A: Foundation (Feature Discovery & Layer Analyzers)

**Goal**: Implement core feature discovery and layer-specific analyzers.

**Tasks**:
1. Feature discovery algorithm (2 days)
   - UI layer discovery
   - API layer discovery
   - Database layer discovery
   - Business logic discovery
   - Integration layer discovery
   - Test layer discovery

2. Layer-specific analyzers (3 days)
   - Business context analyzer
   - UI layer analyzer
   - API layer analyzer
   - Database layer analyzer
   - Business logic analyzer
   - Integration layer analyzer
   - Test layer analyzer

3. End-to-end flow verification (2 days)
   - Flow detection algorithm
   - Breakpoint identification
   - Integration verification

4. Hub LLM integration (2 days)
   - API key management
   - Model selection logic
   - Cost optimization (caching, progressive depth)
   - Token tracking

5. Result aggregation (1 day)
   - Checklist generation
   - Prioritization algorithm
   - Summary generation

6. Database schema (1 day)
   - `comprehensive_validations` table
   - `analysis_configurations` table
   - Indexes for performance

7. API endpoints (1 day)
   - `POST /api/v1/analyze/comprehensive`
   - `GET /api/v1/validations/{id}`
   - `GET /api/v1/validations?project={id}`

**Subtotal**: 12 days

### Phase 14B: MCP Integration

**Goal**: Integrate comprehensive analysis into Cursor via MCP.

**Tasks**:
1. MCP tool implementation (2 days)
   - `sentinel_analyze_feature_comprehensive` tool
   - Request/response handling
   - Error handling and fallback

2. Agent integration (1 day)
   - Agent command handler
   - Hub communication
   - Response formatting

3. Testing (2 days)
   - Unit tests
   - Integration tests
   - End-to-end tests

**Subtotal**: 5 days

### Phase 14C: Hub Configuration Interface

**Goal**: Build Hub UI for LLM configuration and cost tracking.

**Tasks**:
1. Configuration UI (2 days)
   - Provider selection
   - API key input (encrypted)
   - Model selection
   - Cost optimization settings

2. Usage tracking dashboard (2 days)
   - Token usage charts
   - Cost reports
   - Usage by project/user

3. Testing (1.5 days)
   - UI tests
   - Integration tests

**Subtotal**: 5.5 days

### Phase 14D: Cost Optimization

**Goal**: Implement advanced cost optimization features.

**Tasks**:
1. Caching system (2 days)
   - Result caching by file hash
   - Business context caching
   - LLM response caching

2. Progressive depth (1.5 days)
   - Level 1: Fast checks (no LLM)
   - Level 2: Medium-depth (cheaper models)
   - Level 3: Deep analysis (expensive models)

3. Smart model selection (1.5 days)
   - Task classification
   - Model routing
   - Cost tracking

**Subtotal**: 5 days

**Total**: 27.5 days (~5.5 weeks)

---

## 11. Cost Analysis

### Important: No Sentinel Billing

**Sentinel does NOT**:
- Bill users for LLM usage
- Process payments
- Manage subscriptions
- Handle refunds

**Users/Organizations**:
- Subscribe to LLM providers separately (OpenAI, Anthropic, Azure)
- Provide API keys to Sentinel Hub
- Pay LLM providers directly
- Sentinel only tracks usage for reporting

### Cost Model: User Pays LLM Provider Directly

**Individual Developer**:
- Subscribes to: OpenAI API or Codex Pro
- Provides: API key to Hub
- Pays: LLM provider directly
- Sentinel: Tracks usage for reporting only

**Team**:
- Option 1: Each developer provides their own API key
- Option 2: Organization provides shared API key
- Pays: LLM provider directly
- Sentinel: Tracks usage per project/user for allocation

**Organization**:
- Subscribes to: LLM provider (enterprise plan)
- Provides: Shared API key to Hub
- Pays: LLM provider directly
- Sentinel: Tracks usage for cost allocation and reporting

### Estimated LLM Costs (Paid to Provider, Not Sentinel)

**Per Comprehensive Analysis**:
- Tokens: ~6,400 (with 70% caching: ~1,920)
- Cost (GPT-5.1-Codex-Max): ~$0.40 (with caching: ~$0.12)
- Cost (codex-mini-latest): ~$0.16 (with caching: ~$0.05)

**Monthly Usage Examples**:
- 10 analyses/day × 20 days = 200 analyses/month
- Cost: 200 × $0.12 = $24/month (with caching)
- User pays OpenAI directly, not Sentinel

### Usage Tracking (Reporting Only)

**What Sentinel Tracks**:
- Token usage per analysis
- Estimated cost per analysis (for reporting)
- Usage trends
- Cost allocation per project/user

**What Sentinel Does NOT Do**:
- Charge users
- Process payments
- Manage credits
- Handle refunds

---

## 12. Performance Considerations

### Analysis Time by Mode

**Surface Mode** (Fast checks only):
- Time: ~5-10 seconds
- LLM Calls: 0-1 (only for business context)
- Tokens: ~500-1,000

**Medium Mode** (Balanced):
- Time: ~30-60 seconds
- LLM Calls: 3-5
- Tokens: ~3,000-5,000

**Deep Mode** (Comprehensive):
- Time: ~2-5 minutes
- LLM Calls: 10-15
- Tokens: ~10,000-15,000

### Caching Strategy

**Target**: 70% cache hit rate

**What's Cached**:
- Analysis results by file hash
- Business context queries
- LLM responses for identical code patterns

**Cache Invalidation**:
- On file change (hash mismatch)
- On business context update
- TTL: 24 hours

### Scalability

**Hub Architecture**:
- Async processing via Redis queue
- Worker pool for parallel analysis
- Horizontal scaling support

**Agent Architecture**:
- Lightweight (no LLM calls)
- Fast response (cached results)
- Minimal resource usage

---

## 13. Risk Assessment

### Technical Risks

**Risk 1: LLM API Rate Limits**
- **Impact**: High
- **Probability**: Medium
- **Mitigation**: Implement retry logic, fallback to cheaper models, queue system

**Risk 2: High Token Costs**
- **Impact**: Medium
- **Probability**: Medium
- **Mitigation**: Aggressive caching, progressive depth, smart model selection

**Risk 3: Analysis Timeout**
- **Impact**: Medium
- **Probability**: Low
- **Mitigation**: Timeout handling, partial results, async processing

### Operational Risks

**Risk 1: Hub Unavailability**
- **Impact**: High
- **Probability**: Low
- **Mitigation**: Fallback to Cursor default auto mode, offline mode

**Risk 2: API Key Compromise**
- **Impact**: High
- **Probability**: Low
- **Mitigation**: Encrypted storage, key rotation, access control

**Risk 3: False Positives**
- **Impact**: Medium
- **Probability**: Medium
- **Mitigation**: Tuning, user feedback, baseline system

---

## 14. Success Criteria

### Functional Requirements

- ✅ Auto-discover features across all 7 layers
- ✅ Analyze each layer with appropriate depth
- ✅ Verify end-to-end flows
- ✅ Integrate business context (rules, journeys, entities)
- ✅ Generate prioritized checklists
- ✅ Store results in Hub for reference
- ✅ Support manual file specification
- ✅ Support multiple analysis depths

### Non-Functional Requirements

- ✅ Analysis completes in <5 minutes (deep mode)
- ✅ 70% cache hit rate
- ✅ 40% cost reduction via optimization
- ✅ Hub availability >99.9%
- ✅ Secure API key storage
- ✅ Comprehensive error handling

### User Experience

- ✅ Simple MCP tool invocation
- ✅ Clear, actionable checklist
- ✅ Hub dashboard for detailed results
- ✅ Configurable via Hub interface
- ✅ Transparent cost tracking (reporting only)

---

## 15. References

- [FEATURES.md](./FEATURES.md) - Complete feature specification
- [ARCHITECTURE.md](./ARCHITECTURE.md) - System architecture
- [IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md) - Implementation timeline
- [TECHNICAL_SPEC.md](./TECHNICAL_SPEC.md) - Technical specifications
- [USER_GUIDE.md](./USER_GUIDE.md) - User documentation

---

**Document Version**: 1.0  
**Last Updated**: 2024-12-XX  
**Status**: ⏳ Pending Implementation (Phase 14A)












