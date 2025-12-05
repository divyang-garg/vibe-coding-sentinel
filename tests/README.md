# Sentinel Test Suite

## Directory Structure

```
tests/
├── fixtures/                    # Test data and sample files
│   ├── projects/               # Sample projects for pattern detection
│   │   ├── javascript/         # JS/TS project with camelCase
│   │   ├── python/             # Python project with snake_case
│   │   └── shell/              # Shell scripts with best practices
│   ├── security/               # Vulnerable code for security scanning
│   │   ├── secrets_vulnerable.js      # Hardcoded secrets
│   │   ├── sql_injection_vulnerable.php # SQL injection
│   │   ├── shell_vulnerable.sh        # Shell vulnerabilities
│   │   ├── nosql_vulnerable.js        # NoSQL injection
│   │   └── clean_code.js              # Clean file (no issues)
│   ├── config/                 # Configuration file samples
│   │   ├── valid_config.json   # Valid configuration
│   │   ├── minimal_config.json # Minimal valid config
│   │   └── invalid_config.json # Invalid configuration
│   ├── documents/              # Sample documents for ingestion
│   │   └── (placeholder PDFs)
│   └── patterns/               # Pattern detection samples
│
├── unit/                       # Unit tests
│   ├── scanning_test.go        # Security scanning tests
│   ├── config_test.go          # Configuration tests
│   ├── baseline_test.go        # Baseline system tests
│   ├── patterns_test.go        # Pattern detection tests
│   └── fix_test.go             # Auto-fix tests
│
└── integration/                # Integration tests
    ├── workflow_test.go        # End-to-end workflow
    ├── hooks_test.go           # Git hooks tests
    └── hub_test.go             # Hub integration tests
```

## Running Tests

### Prerequisites

- Go 1.21+
- Sentinel binary compiled

### Run All Tests

```bash
# From project root
go test -v ./tests/...
```

### Run Specific Test Suite

```bash
# Unit tests only
go test -v ./tests/unit/...

# Integration tests only
go test -v ./tests/integration/...
```

### Run with Coverage

```bash
go test -v -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out -o coverage.html
```

## Test Fixtures

### Security Fixtures

The `security/` directory contains intentionally vulnerable code for testing Sentinel's detection capabilities.

| File | Issues | Expected Findings |
|------|--------|-------------------|
| `secrets_vulnerable.js` | Hardcoded API keys, AWS creds, passwords | ~8 critical, 3 warning |
| `sql_injection_vulnerable.php` | SQL injection, eval, XXE | ~9 critical |
| `shell_vulnerable.sh` | Unquoted vars, eval, hardcoded paths | ~15 issues |
| `nosql_vulnerable.js` | $where, NoSQL injection | ~10 issues |
| `clean_code.js` | None | 0 findings |

### Project Fixtures

Sample projects with consistent naming conventions for pattern detection testing.

| Project | Language | Naming Style | Structure |
|---------|----------|--------------|-----------|
| `javascript/` | JS/JSX | camelCase | src/, components/ |
| `python/` | Python | snake_case | src/, services/ |
| `shell/` | Bash | snake_case | scripts/ |

### Config Fixtures

| File | Purpose |
|------|---------|
| `valid_config.json` | Complete valid configuration |
| `minimal_config.json` | Minimum required fields |
| `invalid_config.json` | Invalid types for error testing |

## Writing Tests

### Unit Test Template

```go
package unit

import (
    "testing"
)

func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Test Template

```go
package integration

import (
    "os/exec"
    "testing"
)

func TestWorkflow(t *testing.T) {
    // Setup
    
    // Execute sentinel command
    cmd := exec.Command("./sentinel", "audit")
    output, err := cmd.CombinedOutput()
    
    // Assertions
    if err != nil {
        t.Fatalf("Command failed: %v, output: %s", err, output)
    }
}
```

## Coverage Targets

| Component | Target | Notes |
|-----------|--------|-------|
| Core scanning | >90% | Critical path |
| Pattern detection | >85% | Complex logic |
| Configuration | >90% | Must be robust |
| Baseline | >90% | Data integrity |
| Fix engine | >95% | Modifies code |

## CI Integration

Tests are run automatically on:
- Pull requests
- Push to main branch
- Release tags

See `.github/workflows/ci.yml` for configuration.

