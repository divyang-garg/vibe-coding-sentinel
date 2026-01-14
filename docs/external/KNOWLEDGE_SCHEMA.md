# Knowledge Schema Specification

> **For AI Agents**: This document defines the standardized schema for all knowledge extracted from project documents. Follow this schema exactly when processing documents or generating knowledge items.

## Overview

The Knowledge Schema provides a consistent, machine-readable format for business knowledge extracted from project documents. This standardization ensures:

1. **Unambiguous interpretation** - No room for LLM misinterpretation
2. **Executable assertions** - Rules can be verified against code
3. **Test generation** - Test cases derived automatically
4. **Traceability** - Link from requirement to code to test
5. **Change management** - Track versions and modifications

---

## Schema Version

```
Version: 1.0.0
Last Updated: 2024-12-05
Status: Active
```

---

## 1. Document Metadata

Every knowledge document must include metadata:

```json
{
  "$schema": "https://sentinel.dev/schemas/knowledge/v1",
  "metadata": {
    "project_id": "string (required)",
    "document_version": "semver (required)",
    "created_at": "ISO 8601 date (required)",
    "last_updated": "ISO 8601 date (required)",
    "status": "draft | active | deprecated (required)",
    "reviewed_by": "email (optional)",
    "reviewed_at": "ISO 8601 date (optional)",
    "source_documents": [
      {
        "name": "string",
        "type": "pdf | docx | xlsx | txt | md | eml | image",
        "hash": "sha256:... (for integrity)",
        "uploaded_at": "ISO 8601 date",
        "pages_extracted": [1, 5, 7]
      }
    ]
  }
}
```

---

## 2. Business Rules Schema

### 2.1 Full Schema

```json
{
  "business_rules": [
    {
      "id": "BR-XXX (required, auto-generated if not provided)",
      "version": "semver (required)",
      "status": "draft | active | deprecated | superseded",
      "superseded_by": "BR-XXX (if status is superseded)",
      
      "title": "string (required, max 100 chars)",
      "description": "string (required, max 500 chars)",
      "category": "string (optional, e.g., 'orders', 'payments', 'users')",
      "priority": "critical | high | medium | low",
      
      "specification": {
        "trigger": "string - What initiates this rule",
        
        "preconditions": [
          "string - Condition that must be true before rule applies"
        ],
        
        "constraints": [
          {
            "id": "C1",
            "type": "time_based | value_based | state_based | relationship_based",
            "expression": "Human readable expression",
            "pseudocode": "Machine-parseable expression",
            "boundary": "inclusive | exclusive",
            "unit": "hours | minutes | days | currency | count"
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
            "condition": "When to trigger (or 'always')",
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
          "id": "BR-XXX-T1",
          "name": "test_snake_case_name",
          "type": "happy_path | error_case | edge_case | exception_case",
          "priority": "critical | high | medium | low",
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
        }
      ],
      
      "implementation_hints": {
        "suggested_location": "path/to/file.ts",
        "function_name": "cancelOrder",
        "dependencies": ["OrderRepository", "PaymentService"],
        "patterns": ["Service layer", "Repository pattern"],
        "complexity_estimate": "low | medium | high"
      },
      
      "traceability": {
        "source_document": "filename",
        "source_section": "Section number or name",
        "source_page": 12,
        "source_quote": "Original text from document",
        "stakeholder": "Role or name",
        "approved_date": "ISO 8601 date",
        "related_rules": ["BR-002", "BR-005"],
        "implements": ["code location once implemented"]
      },
      
      "metadata": {
        "created_at": "ISO 8601 date",
        "created_by": "system | email",
        "confidence": 0.92,
        "needs_clarification": false,
        "clarification_questions": []
      }
    }
  ]
}
```

### 2.2 Minimal Required Fields

```json
{
  "id": "BR-001",
  "version": "1.0",
  "status": "active",
  "title": "Order Cancellation Window",
  "description": "Orders can only be cancelled within 24 hours of creation",
  "specification": {
    "constraints": [
      {
        "type": "time_based",
        "expression": "Order age < 24 hours",
        "pseudocode": "Date.now() - order.createdAt < 24 * 60 * 60 * 1000"
      }
    ]
  },
  "test_requirements": [
    {
      "id": "BR-001-T1",
      "name": "test_cancel_within_window",
      "type": "happy_path",
      "scenario": "Cancel order within 24 hours"
    },
    {
      "id": "BR-001-T2", 
      "name": "test_cancel_after_window",
      "type": "error_case",
      "scenario": "Reject cancel after 24 hours"
    }
  ],
  "traceability": {
    "source_document": "Requirements.docx"
  }
}
```

### 2.3 Example: Complete Business Rule

```json
{
  "id": "BR-001",
  "version": "1.0.0",
  "status": "active",
  "title": "Order Cancellation Window",
  "description": "Orders can only be cancelled within 24 hours of creation, with exceptions for premium users",
  "category": "orders",
  "priority": "high",
  
  "specification": {
    "trigger": "User requests to cancel an order",
    
    "preconditions": [
      "Order exists in the system",
      "Order status is 'pending' or 'processing'",
      "User is the order owner or an admin"
    ],
    
    "constraints": [
      {
        "id": "C1",
        "type": "time_based",
        "expression": "Order must be less than 24 hours old",
        "pseudocode": "Date.now() - order.createdAt < 24 * 60 * 60 * 1000",
        "boundary": "exclusive",
        "unit": "hours"
      }
    ],
    
    "exceptions": [
      {
        "id": "E1",
        "condition": "User is a premium tier member",
        "modified_constraint": "48 hours instead of 24 hours",
        "applies_to": ["user.tier === 'premium'"],
        "source": "Premium Policy Document, Section 3.2"
      },
      {
        "id": "E2",
        "condition": "User is a business account",
        "modified_constraint": "Can cancel until order is shipped",
        "applies_to": ["user.type === 'business'"],
        "source": "Business Account Terms, Section 5.1"
      }
    ],
    
    "side_effects": [
      {
        "action": "initiateRefund",
        "condition": "order.paymentStatus === 'paid'",
        "parameters": {"orderId": "order.id", "amount": "order.total"},
        "async": true,
        "required": true
      },
      {
        "action": "restoreInventory",
        "condition": "always",
        "parameters": {"items": "order.items"},
        "async": true,
        "required": true
      },
      {
        "action": "notifyWarehouse",
        "condition": "order.status === 'processing'",
        "parameters": {"orderId": "order.id", "action": "cancel"},
        "async": true,
        "required": false
      },
      {
        "action": "sendNotification",
        "condition": "always",
        "parameters": {"template": "order_cancelled", "userId": "order.userId"},
        "async": true,
        "required": true
      }
    ],
    
    "error_cases": [
      {
        "condition": "Order does not exist",
        "error_code": "ERR_ORDER_NOT_FOUND",
        "error_message": "Order not found",
        "http_status": 404,
        "recoverable": false
      },
      {
        "condition": "Order already shipped or delivered",
        "error_code": "ERR_ORDER_SHIPPED",
        "error_message": "Cannot cancel shipped or delivered orders",
        "http_status": 400,
        "recoverable": false
      },
      {
        "condition": "Outside cancellation window",
        "error_code": "ERR_CANCEL_WINDOW_EXPIRED",
        "error_message": "Cancellation window has expired",
        "http_status": 400,
        "recoverable": false
      },
      {
        "condition": "User does not own order",
        "error_code": "ERR_FORBIDDEN",
        "error_message": "You do not have permission to cancel this order",
        "http_status": 403,
        "recoverable": false
      }
    ]
  },
  
  "test_requirements": [
    {
      "id": "BR-001-T1",
      "name": "test_cancel_within_24h_success",
      "type": "happy_path",
      "priority": "critical",
      "scenario": "User successfully cancels order created 1 hour ago",
      "setup": {
        "entities": {
          "user": {"id": "user-1", "tier": "standard"},
          "order": {"id": "order-1", "userId": "user-1", "createdAt": "now - 1h", "status": "pending", "paymentStatus": "paid"}
        }
      },
      "action": "cancelOrder(order.id, user.id)",
      "expected": {
        "success": true,
        "return_value": {"orderId": "order-1", "status": "cancelled"},
        "side_effects": ["refund initiated", "inventory restored", "email sent"]
      },
      "assertions_required": [
        "expect(result.success).toBe(true)",
        "expect(order.status).toBe('cancelled')",
        "expect(refundService.initiate).toHaveBeenCalled()",
        "expect(inventoryService.restore).toHaveBeenCalled()"
      ]
    },
    {
      "id": "BR-001-T2",
      "name": "test_cancel_after_24h_fails",
      "type": "error_case",
      "priority": "critical",
      "scenario": "User tries to cancel order created 25 hours ago",
      "setup": {
        "entities": {
          "user": {"id": "user-1", "tier": "standard"},
          "order": {"id": "order-1", "userId": "user-1", "createdAt": "now - 25h", "status": "pending"}
        }
      },
      "action": "cancelOrder(order.id, user.id)",
      "expected": {
        "success": false,
        "error": "ERR_CANCEL_WINDOW_EXPIRED"
      },
      "assertions_required": [
        "expect(result.success).toBe(false)",
        "expect(result.error).toBe('ERR_CANCEL_WINDOW_EXPIRED')",
        "expect(order.status).toBe('pending')"
      ]
    },
    {
      "id": "BR-001-T3",
      "name": "test_cancel_at_boundary_24h",
      "type": "edge_case",
      "priority": "high",
      "scenario": "User tries to cancel order at exactly 24 hours",
      "setup": {
        "entities": {
          "order": {"createdAt": "now - 24h exactly"}
        }
      },
      "action": "cancelOrder(order.id, user.id)",
      "expected": {
        "success": false,
        "error": "ERR_CANCEL_WINDOW_EXPIRED"
      },
      "notes": "Boundary is exclusive, so exactly 24h should fail"
    },
    {
      "id": "BR-001-T4",
      "name": "test_cancel_premium_48h_success",
      "type": "exception_case",
      "priority": "high",
      "scenario": "Premium user cancels order created 30 hours ago",
      "setup": {
        "entities": {
          "user": {"id": "user-1", "tier": "premium"},
          "order": {"id": "order-1", "userId": "user-1", "createdAt": "now - 30h", "status": "pending"}
        }
      },
      "action": "cancelOrder(order.id, user.id)",
      "expected": {
        "success": true
      }
    },
    {
      "id": "BR-001-T5",
      "name": "test_cancel_shipped_fails",
      "type": "error_case",
      "priority": "critical",
      "scenario": "User tries to cancel shipped order",
      "setup": {
        "entities": {
          "order": {"status": "shipped", "createdAt": "now - 1h"}
        }
      },
      "action": "cancelOrder(order.id, user.id)",
      "expected": {
        "success": false,
        "error": "ERR_ORDER_SHIPPED"
      }
    }
  ],
  
  "implementation_hints": {
    "suggested_location": "src/services/order/cancellation.ts",
    "function_name": "cancelOrder",
    "dependencies": ["OrderRepository", "PaymentService", "InventoryService", "NotificationService"],
    "patterns": ["Service layer", "Repository pattern", "Event-driven side effects"],
    "complexity_estimate": "medium"
  },
  
  "traceability": {
    "source_document": "Requirements_v2.docx",
    "source_section": "4.2 Order Management",
    "source_page": 12,
    "source_quote": "Customers may cancel their order within 24 hours of placing it...",
    "stakeholder": "Product Owner",
    "approved_date": "2024-11-15",
    "related_rules": ["BR-002", "BR-003", "BR-010"],
    "implements": []
  },
  
  "metadata": {
    "created_at": "2024-12-01T10:00:00Z",
    "created_by": "system",
    "confidence": 0.92,
    "needs_clarification": false
  }
}
```

---

## 3. Entity Schema

### 3.1 Full Schema

```json
{
  "entities": [
    {
      "id": "ENT-XXX",
      "version": "semver",
      "status": "draft | active | deprecated",
      
      "name": "PascalCase entity name",
      "description": "What this entity represents",
      "category": "domain category",
      
      "fields": [
        {
          "name": "camelCase field name",
          "type": "string | number | boolean | date | datetime | enum | object | array",
          "required": true,
          "unique": false,
          "default": null,
          "description": "Field description",
          
          "validation": {
            "pattern": "regex pattern",
            "min": 0,
            "max": 100,
            "minLength": 1,
            "maxLength": 255,
            "enum_values": ["value1", "value2"],
            "custom": "validation expression"
          }
        }
      ],
      
      "relationships": [
        {
          "entity": "Related entity name",
          "type": "one-to-one | one-to-many | many-to-many",
          "foreign_key": "field name",
          "inverse": "field on related entity",
          "cascade": "delete | nullify | restrict"
        }
      ],
      
      "invariants": [
        {
          "name": "invariant_name",
          "condition": "Expression that must always be true",
          "message": "Error message if violated"
        }
      ],
      
      "business_rules": ["BR-001", "BR-005"],
      
      "traceability": {
        "source_document": "filename",
        "source_section": "section"
      }
    }
  ]
}
```

### 3.2 Example: User Entity

```json
{
  "id": "ENT-001",
  "version": "1.0.0",
  "status": "active",
  
  "name": "User",
  "description": "Represents a registered customer in the system",
  "category": "identity",
  
  "fields": [
    {
      "name": "id",
      "type": "string",
      "required": true,
      "unique": true,
      "description": "Unique identifier (UUID)",
      "validation": {
        "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
      }
    },
    {
      "name": "email",
      "type": "string",
      "required": true,
      "unique": true,
      "description": "User's email address",
      "validation": {
        "pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
        "maxLength": 255
      }
    },
    {
      "name": "passwordHash",
      "type": "string",
      "required": true,
      "description": "Bcrypt hashed password",
      "validation": {
        "minLength": 60,
        "maxLength": 60
      }
    },
    {
      "name": "firstName",
      "type": "string",
      "required": true,
      "validation": {
        "minLength": 1,
        "maxLength": 100
      }
    },
    {
      "name": "lastName",
      "type": "string",
      "required": true,
      "validation": {
        "minLength": 1,
        "maxLength": 100
      }
    },
    {
      "name": "tier",
      "type": "enum",
      "required": true,
      "default": "standard",
      "description": "User's subscription tier",
      "validation": {
        "enum_values": ["standard", "premium", "enterprise"]
      }
    },
    {
      "name": "role",
      "type": "enum",
      "required": true,
      "default": "customer",
      "validation": {
        "enum_values": ["customer", "support", "admin"]
      }
    },
    {
      "name": "createdAt",
      "type": "datetime",
      "required": true,
      "description": "When the user registered"
    },
    {
      "name": "lastLoginAt",
      "type": "datetime",
      "required": false
    }
  ],
  
  "relationships": [
    {
      "entity": "Order",
      "type": "one-to-many",
      "foreign_key": "userId",
      "inverse": "orders",
      "cascade": "restrict"
    },
    {
      "entity": "Address",
      "type": "one-to-many",
      "foreign_key": "userId",
      "inverse": "addresses",
      "cascade": "delete"
    }
  ],
  
  "invariants": [
    {
      "name": "email_immutable",
      "condition": "on_update: email === previous.email",
      "message": "Email cannot be changed after registration"
    },
    {
      "name": "password_hashed",
      "condition": "passwordHash.startsWith('$2')",
      "message": "Password must be bcrypt hashed"
    }
  ],
  
  "business_rules": ["BR-001", "BR-005", "BR-012"]
}
```

---

## 4. API Contract Schema

### 4.1 Full Schema

```json
{
  "api_contracts": [
    {
      "id": "API-XXX",
      "version": "semver",
      "status": "draft | active | deprecated",
      
      "endpoint": "/api/path/:param",
      "method": "GET | POST | PUT | PATCH | DELETE",
      "description": "What this endpoint does",
      
      "authentication": {
        "required": true,
        "type": "bearer | api_key | basic",
        "scopes": ["read:orders", "write:orders"]
      },
      
      "rate_limiting": {
        "enabled": true,
        "requests_per_minute": 60
      },
      
      "request": {
        "params": {
          "paramName": {
            "type": "string | number",
            "required": true,
            "validation": {}
          }
        },
        "query": {},
        "headers": {},
        "body": {
          "fieldName": {
            "type": "string",
            "required": true,
            "validation": {}
          }
        }
      },
      
      "response": {
        "200": {
          "description": "Success response",
          "schema": {}
        },
        "400": {
          "description": "Validation error",
          "schema": {},
          "examples": []
        },
        "401": {
          "description": "Unauthorized"
        },
        "403": {
          "description": "Forbidden"
        },
        "404": {
          "description": "Not found"
        }
      },
      
      "implements_rules": ["BR-001", "BR-002"],
      "security_rules": ["SEC-001", "SEC-003"]
    }
  ]
}
```

### 4.2 Example: Cancel Order API

```json
{
  "id": "API-005",
  "version": "1.0.0",
  "status": "active",
  
  "endpoint": "/api/v1/orders/:orderId/cancel",
  "method": "POST",
  "description": "Cancel an existing order",
  
  "authentication": {
    "required": true,
    "type": "bearer",
    "scopes": ["write:orders"]
  },
  
  "rate_limiting": {
    "enabled": true,
    "requests_per_minute": 10
  },
  
  "request": {
    "params": {
      "orderId": {
        "type": "string",
        "required": true,
        "description": "Order UUID",
        "validation": {
          "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
        }
      }
    },
    "body": {
      "reason": {
        "type": "string",
        "required": false,
        "description": "Cancellation reason",
        "validation": {
          "maxLength": 500
        }
      }
    }
  },
  
  "response": {
    "200": {
      "description": "Order successfully cancelled",
      "schema": {
        "success": "boolean",
        "order": {
          "id": "string",
          "status": "string",
          "cancelledAt": "datetime"
        },
        "refund": {
          "initiated": "boolean",
          "amount": "number",
          "estimatedDays": "number"
        }
      }
    },
    "400": {
      "description": "Cannot cancel order",
      "schema": {
        "success": "boolean",
        "error": "string",
        "code": "string"
      },
      "examples": [
        {
          "success": false,
          "error": "Cancellation window has expired",
          "code": "ERR_CANCEL_WINDOW_EXPIRED"
        },
        {
          "success": false,
          "error": "Cannot cancel shipped orders",
          "code": "ERR_ORDER_SHIPPED"
        }
      ]
    },
    "401": {
      "description": "Not authenticated"
    },
    "403": {
      "description": "User does not own this order"
    },
    "404": {
      "description": "Order not found"
    }
  },
  
  "implements_rules": ["BR-001", "BR-002"],
  "security_rules": ["SEC-001", "SEC-003", "SEC-004"]
}
```

---

## 5. Security Rules Schema

### 5.1 Full Schema

```json
{
  "security_rules": [
    {
      "id": "SEC-XXX",
      "version": "semver",
      "status": "active",
      
      "name": "Rule name",
      "type": "authorization | authentication | injection | validation | cryptography | transport",
      "severity": "critical | high | medium | low",
      "description": "What this rule enforces",
      
      "detection": {
        "endpoints": ["patterns to match"],
        "resources": ["affected resource types"],
        "required_checks": ["checks that must be present"],
        "patterns_forbidden": ["regex patterns that must not appear"],
        "patterns_required": ["regex patterns that must appear"]
      },
      
      "ast_check": {
        "function_contains": ["function names to check"],
        "must_have_before_response": "required code pattern"
      },
      
      "auto_fix": {
        "available": true,
        "code_template": "code to insert"
      },
      
      "test_requirements": [
        {
          "id": "SEC-XXX-T1",
          "scenario": "Test scenario",
          "type": "bypass_attempt | valid_request"
        }
      ]
    }
  ]
}
```

---

## 6. User Journey Schema

### 6.1 Full Schema

```json
{
  "user_journeys": [
    {
      "id": "UJ-XXX",
      "version": "semver",
      "status": "active",
      
      "name": "Journey name",
      "actor": "User role performing journey",
      "goal": "What the user wants to achieve",
      "description": "Journey description",
      
      "preconditions": [
        "Conditions that must be true before starting"
      ],
      
      "steps": [
        {
          "step": 1,
          "actor_action": "What the user does",
          "system_response": "What the system does",
          "validation": "Condition to verify",
          "business_rules": ["BR-XXX"],
          "api_calls": ["API-XXX"]
        }
      ],
      
      "postconditions": [
        "Conditions that must be true after completion"
      ],
      
      "error_paths": [
        {
          "from_step": 2,
          "condition": "When this error occurs",
          "response": "System response",
          "recovery": "How to recover"
        }
      ],
      
      "metrics": {
        "expected_duration": "seconds",
        "success_rate_target": 0.95
      }
    }
  ]
}
```

---

## 7. Glossary Schema

### 7.1 Full Schema

```json
{
  "glossary": [
    {
      "id": "GL-XXX",
      "term": "Term name",
      "definition": "Clear definition",
      "synonyms": ["alternative terms"],
      "related_terms": ["GL-YYY"],
      "examples": ["Usage examples"],
      "anti_examples": ["What it does NOT mean"],
      "context": "When this term applies",
      "source": "Where this definition comes from"
    }
  ]
}
```

---

## 8. Requirements Lifecycle Schema

### 8.1 Change Request Schema

```json
{
  "change_requests": [
    {
      "id": "CR-XXX",
      "type": "new | modification | deprecation",
      "status": "draft | pending_approval | approved | rejected | implemented",
      "priority": "critical | high | medium | low",
      
      "target": {
        "type": "business_rule | entity | api_contract",
        "id": "BR-001"
      },
      
      "requested_by": "email",
      "requested_at": "ISO 8601 date",
      
      "current_state": {
        "summary": "Current behavior"
      },
      
      "proposed_state": {
        "summary": "Proposed behavior"
      },
      
      "justification": "Why this change is needed",
      
      "impact_analysis": {
        "affected_code": ["file:lines"],
        "affected_tests": ["test files"],
        "affected_rules": ["related rules"],
        "estimated_effort": "hours",
        "risk_level": "low | medium | high"
      },
      
      "approval": {
        "required_approvers": ["roles or emails"],
        "approvals": [
          {
            "approver": "email",
            "approved_at": "date",
            "comments": "any comments"
          }
        ]
      },
      
      "implementation": {
        "status": "not_started | in_progress | completed",
        "implemented_by": "email",
        "implemented_at": "date",
        "commits": ["commit hashes"]
      }
    }
  ]
}
```

---

## 9. Project Profile Schema

### 9.1 Adaptive Configuration

```json
{
  "project_profile": {
    "detected": {
      "size": "small | medium | large | enterprise",
      "files": 127,
      "lines_of_code": 18500,
      "languages": {
        "typescript": 0.85,
        "python": 0.15
      },
      "frameworks": ["express", "react"]
    },
    
    "configuration": {
      "file_size_threshold": 300,
      "test_coverage_target": 80,
      "rule_coverage_target": 100,
      "documentation_level": "standard"
    },
    
    "thresholds": {
      "file_size": {
        "warning": 300,
        "critical": 500,
        "maximum": 1000
      },
      "function_length": {
        "warning": 50,
        "critical": 100
      },
      "complexity": {
        "warning": 10,
        "critical": 20
      }
    }
  }
}
```

---

## 10. Extraction Guidelines

### 10.1 LLM Extraction Prompt Template

When extracting knowledge from documents, use this prompt:

```
You are extracting business knowledge from a project document.

For EACH piece of knowledge found:

1. IDENTIFY the type:
   - Business Rule (BR): Something the system MUST or MUST NOT do
   - Entity (ENT): A thing the system manages (User, Order, Product)
   - API Contract (API): An endpoint definition
   - User Journey (UJ): A sequence of steps a user takes
   - Glossary (GL): A term definition

2. For BUSINESS RULES, extract:
   - Trigger: What initiates this rule?
   - Preconditions: What must be true before?
   - Constraints: What are the EXACT conditions?
     - For numeric values: specify boundary (< vs <=)
     - For time: specify units and reference point
   - Exceptions: Who/what is exempt?
   - Side effects: What else must happen?
   - Error cases: What can go wrong?

3. For EVERY constraint:
   - Write pseudocode that can be verified
   - Specify if boundary is inclusive or exclusive
   - If AMBIGUOUS: flag as "NEEDS_CLARIFICATION" and list interpretations

4. Generate TEST REQUIREMENTS:
   - Minimum: happy_path + error_case for each rule
   - Include: boundary tests for numeric constraints
   - Include: exception tests if exceptions exist

5. Add TRACEABILITY:
   - Source document name
   - Section/page number
   - Quote the original text

OUTPUT FORMAT: Use the JSON schema defined in KNOWLEDGE_SCHEMA.md

Document text:
{{document_text}}
```

### 10.2 Ambiguity Handling

When the source document is ambiguous:

```json
{
  "id": "BR-015",
  "title": "Refund Processing Time",
  "status": "needs_clarification",
  
  "specification": {
    "constraints": [
      {
        "type": "time_based",
        "expression": "Refund processed within 5 days",
        "ambiguous": true,
        "possible_interpretations": [
          {
            "interpretation": "5 business days",
            "pseudocode": "businessDays(refund.requestedAt, now) <= 5"
          },
          {
            "interpretation": "5 calendar days",
            "pseudocode": "calendarDays(refund.requestedAt, now) <= 5"
          }
        ]
      }
    ]
  },
  
  "metadata": {
    "needs_clarification": true,
    "clarification_questions": [
      "Is the 5-day refund window in business days or calendar days?",
      "Does the clock start from request submission or approval?"
    ]
  }
}
```

---

## Appendix: Quick Reference

### Required Fields by Type

| Type | Required Fields |
|------|-----------------|
| Business Rule | id, version, status, title, description, specification.constraints, test_requirements (min 2), traceability.source_document |
| Entity | id, version, status, name, description, fields (min 1) |
| API Contract | id, version, status, endpoint, method, response.200 |
| User Journey | id, version, status, name, goal, steps (min 1) |
| Security Rule | id, version, status, name, type, severity, detection |
| Glossary | id, term, definition |

### Status Values

| Status | Meaning |
|--------|---------|
| draft | Being written, not reviewed |
| active | Approved and in effect |
| deprecated | Being phased out |
| superseded | Replaced by another rule |

### Priority/Severity Values

| Level | Meaning |
|-------|---------|
| critical | Must be implemented/fixed immediately |
| high | Important, should be addressed soon |
| medium | Normal priority |
| low | Nice to have |












