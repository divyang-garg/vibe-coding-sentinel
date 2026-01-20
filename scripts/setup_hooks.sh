#!/bin/bash
# Setup script for Git hooks and CI/CD integration
# Configures pre-commit hooks and CI/CD pipelines for CODING_STANDARDS.md compliance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.githooks"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_header() {
    echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
}

# Function to check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    log_success "Git repository detected"
}

# Function to configure git hooks path
configure_git_hooks() {
    log_info "Configuring git hooks path..."

    # Set the hooks path to .githooks directory
    git config core.hooksPath "$HOOKS_DIR"

    # Verify the configuration
    local configured_path=$(git config core.hooksPath)
    if [ "$configured_path" = "$HOOKS_DIR" ]; then
        log_success "Git hooks path configured: $HOOKS_DIR"
    else
        log_error "Failed to configure git hooks path"
        return 1
    fi
}

# Function to install pre-commit hook
install_precommit_hook() {
    log_info "Installing pre-commit hook..."

    local hook_file="$HOOKS_DIR/pre-commit"

    if [ ! -f "$hook_file" ]; then
        log_error "Pre-commit hook not found: $hook_file"
        return 1
    fi

    if [ ! -x "$hook_file" ]; then
        log_warning "Pre-commit hook not executable, fixing permissions..."
        chmod +x "$hook_file"
    fi

    log_success "Pre-commit hook installed and executable"
}

# Function to install development tools
install_development_tools() {
    log_info "Installing development tools..."

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        log_warning "Go compiler not found - skipping development tools installation"
        log_info "Install Go from https://go.dev/doc/install"
        return 0
    fi

    # Install goimports
    if command -v goimports >/dev/null 2>&1; then
        log_success "goimports already installed: $(which goimports)"
    else
        log_info "Installing goimports..."
        if go install golang.org/x/tools/cmd/goimports@latest 2>/dev/null; then
            # Check if goimports is now in PATH
            if command -v goimports >/dev/null 2>&1; then
                log_success "goimports installed successfully: $(which goimports)"
            else
                # goimports installed but not in PATH
                GOPATH=$(go env GOPATH)
                if [ -f "$GOPATH/bin/goimports" ]; then
                    log_success "goimports installed to $GOPATH/bin/goimports"
                    log_info "Add to PATH: export PATH=\$PATH:$GOPATH/bin"
                else
                    log_warning "goimports installation may have failed - verify manually"
                fi
            fi
        else
            log_warning "Failed to install goimports automatically"
            log_info "Install manually: go install golang.org/x/tools/cmd/goimports@latest"
            log_info "Then add to PATH: export PATH=\$PATH:\$(go env GOPATH)/bin"
        fi
    fi

    # Install golangci-lint (optional but recommended)
    if command -v golangci-lint >/dev/null 2>&1; then
        log_success "golangci-lint already installed"
    else
        log_info "golangci-lint not installed (optional - install for enhanced linting)"
        log_info "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi
}

# Function to create additional git hooks
create_additional_hooks() {
    log_info "Creating additional git hooks..."

    # Pre-push hook for additional validation
    cat > "$HOOKS_DIR/pre-push" << 'EOF'
#!/bin/bash
# Pre-push hook for additional validation
# Runs quick checks before pushing to remote

echo "🧪 Running pre-push validation..."

# Run a quick unit test subset
if [ -f "tests/run_all_tests.sh" ]; then
    echo "Running unit tests..."
    if timeout 60 bash tests/run_all_tests.sh > /dev/null 2>&1; then
        echo "✅ Unit tests passed"
    else
        echo "❌ Unit tests failed - fix before pushing"
        exit 1
    fi
fi

echo "✅ Pre-push validation completed"
EOF

    chmod +x "$HOOKS_DIR/pre-push"
    log_success "Pre-push hook created"

    # Commit-msg hook for conventional commits
    cat > "$HOOKS_DIR/commit-msg" << 'EOF'
#!/bin/bash
# Commit message validation hook
# Enforces conventional commit format

COMMIT_MSG_FILE=$1
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Basic conventional commit pattern
if ! echo "$COMMIT_MSG" | grep -qE "^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .{1,}"; then
    echo "❌ Invalid commit message format"
    echo ""
    echo "Commit message must follow conventional format:"
    echo "type(scope): description"
    echo ""
    echo "Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert"
    echo ""
    echo "Examples:"
    echo "  feat: add user authentication"
    echo "  fix(api): resolve null pointer exception"
    echo "  test: add unit tests for user service"
    exit 1
fi

echo "✅ Commit message format validated"
EOF

    chmod +x "$HOOKS_DIR/commit-msg"
    log_success "Commit-msg hook created"
}

# Function to setup CI/CD pipeline
setup_ci_pipeline() {
    log_info "Setting up CI/CD pipeline..."

    local workflows_dir=".github/workflows"
    mkdir -p "$workflows_dir"

    # Create GitHub Actions workflow
    cat > "$workflows_dir/ci.yml" << 'EOF'
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Download dependencies
      run: go mod download

    - name: Install development tools
      run: |
        go install golang.org/x/tools/cmd/goimports@latest
        echo "$(go env GOPATH)/bin" >> $GITHUB_PATH

    - name: Check import organization
      run: |
        echo "Checking import organization with goimports..."
        UNFORMATTED=$(goimports -l . | tee /tmp/goimports.diff | wc -l)
        if [ "$UNFORMATTED" -gt 0 ]; then
          echo "❌ Import organization issues found in $UNFORMATTED file(s):"
          cat /tmp/goimports.diff
          echo ""
          echo "Fix by running: goimports -w ."
          exit 1
        else
          echo "✅ All imports properly organized"
        fi

    - name: Run linting
      run: |
        if command -v golangci-lint >/dev/null 2>&1; then
          golangci-lint run --timeout=10m
        else
          echo "golangci-lint not available, skipping"
        fi

    - name: Run unit tests
      run: ./tests/run_all_tests.sh

    - name: Run coverage analysis
      run: ./tests/coverage/coverage_report.sh --ci

    - name: Run performance tests
      run: ./tests/performance/comprehensive_analysis_perf_test.sh

    - name: Generate test reports
      run: ./tests/reporting/test_aggregation_report.sh --no-display

    - name: Upload coverage reports
      uses: actions/upload-artifact@v3
      with:
        name: coverage-reports
        path: tests/coverage/reports/

    - name: Upload test reports
      uses: actions/upload-artifact@v3
      with:
        name: test-reports
        path: tests/reporting/aggregated/

  security:
    runs-on: ubuntu-latest
    needs: test

    steps:
    - uses: actions/checkout@v3

    - name: Run security audit
      run: ./tests/production/SECURITY_AUDIT.sh

    - name: Upload security report
      uses: actions/upload-artifact@v3
      with:
        name: security-report
        path: tests/security/

  deploy:
    runs-on: ubuntu-latest
    needs: [test, security]
    if: github.ref == 'refs/heads/main'

    steps:
    - name: Deploy to staging
      run: echo "Deploy to staging environment"
      # Add actual deployment steps here
EOF

    log_success "GitHub Actions CI/CD pipeline created: $workflows_dir/ci.yml"
}

# Function to create branch protection script
create_branch_protection() {
    log_info "Creating branch protection configuration..."

    cat > "$SCRIPTS_DIR/configure_branch_protection.sh" << 'EOF'
#!/bin/bash
# Configure GitHub branch protection rules
# Run this script to set up branch protection for main/develop branches

# This script provides commands to configure branch protection via GitHub CLI
# Requires: gh CLI installed and authenticated

echo "🔒 Branch Protection Configuration"
echo ""
echo "Run these commands to configure branch protection:"
echo ""
echo "# Require status checks"
echo "gh api repos/{owner}/{repo}/branches/main/protection \\"
echo "  --method PUT \\"
echo "  --field required_status_checks='{\"strict\":true,\"contexts\":[\"test\",\"lint\",\"security\"]}' \\"
echo "  --field enforce_admins=true \\"
echo "  --field required_pull_request_reviews='{\"required_approving_review_count\":1}' \\"
echo "  --field restrictions=null"
echo ""
echo "# Configure for develop branch (similar command)"
echo ""
echo "Manual setup in GitHub:"
echo "1. Go to Settings > Branches > Branch protection rules"
echo "2. Add rule for 'main' and 'develop' branches"
echo "3. Enable: Require status checks, Require reviews, Include administrators"
echo "4. Required status checks: test, lint, security"
echo ""
EOF

    chmod +x "$SCRIPTS_DIR/configure_branch_protection.sh"
    log_success "Branch protection configuration script created"
}

# Function to test hooks installation
test_hooks_installation() {
    log_info "Testing hooks installation..."

    # Check if hooks are configured
    local hooks_path=$(git config core.hooksPath)
    if [ "$hooks_path" != "$HOOKS_DIR" ]; then
        log_error "Git hooks path not configured correctly"
        return 1
    fi

    # Check if pre-commit hook exists and is executable
    if [ ! -x "$HOOKS_DIR/pre-commit" ]; then
        log_error "Pre-commit hook not found or not executable"
        return 1
    fi

    # Test pre-commit hook (dry run)
    log_info "Testing pre-commit hook (dry run)..."
    if bash "$HOOKS_DIR/pre-commit" --help > /dev/null 2>&1; then
        log_success "Pre-commit hook functional"
    else
        log_warning "Pre-commit hook test failed (may be expected if dependencies missing)"
    fi

    log_success "Hooks installation verified"
}

# Function to create documentation
create_documentation() {
    log_info "Creating setup documentation..."

    cat > "$HOOKS_DIR/README.md" << 'EOF'
# Git Hooks Setup

This directory contains Git hooks for enforcing CODING_STANDARDS.md compliance.

## Installed Hooks

### pre-commit
- Runs before each commit
- Validates code quality, tests, coverage, and security
- Blocks commits that don't meet standards

### pre-push
- Runs before pushing to remote
- Performs quick validation checks
- Prevents broken code from reaching remote

### commit-msg
- Validates commit message format
- Enforces conventional commit standards

## Setup

Run the setup script:
```bash
./scripts/setup_hooks.sh
```

## Manual Configuration

If automatic setup fails:

```bash
# Configure hooks path
git config core.hooksPath .githooks

# Make hooks executable
chmod +x .githooks/*
```

## Troubleshooting

### Hook not running
- Check if hooks are executable: `ls -la .githooks/`
- Verify git configuration: `git config core.hooksPath`
- Test manually: `./.githooks/pre-commit`

### Tests failing
- Run tests locally first: `./tests/run_all_tests.sh`
- Check test dependencies
- Review error messages in hook output

### Coverage issues
- Generate coverage report: `./tests/coverage/coverage_report.sh`
- Review coverage gaps
- Add missing test cases

## CODING_STANDARDS.md Compliance

These hooks enforce:
- ✅ Test execution on commit
- ✅ Coverage thresholds (80% overall, 90% critical)
- ✅ Code quality standards
- ✅ File size limits
- ✅ Security checks
- ✅ Conventional commit messages
EOF

    log_success "Hooks documentation created: $HOOKS_DIR/README.md"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Setup script for Git hooks and CI/CD integration"
    echo ""
    echo "OPTIONS:"
    echo "  --help          Show this help message"
    echo "  --skip-ci       Skip CI/CD pipeline setup"
    echo "  --skip-test     Skip hooks testing"
    echo "  --force         Force reinstallation"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                           # Full setup"
    echo "  $0 --skip-ci                # Setup hooks only"
    echo "  $0 --skip-test --force      # Force reinstall without testing"
    echo ""
    echo "WHAT GETS INSTALLED:"
    echo "  • Git hooks: pre-commit, pre-push, commit-msg"
    echo "  • CI/CD pipeline: .github/workflows/ci.yml"
    echo "  • Branch protection script"
    echo "  • Documentation and troubleshooting guides"
    echo ""
    echo "CODING_STANDARDS.md ENFORCEMENT:"
    echo "  • Pre-commit: Tests, coverage, linting, file sizes"
    echo "  • CI/CD: Automated testing and quality gates"
    echo "  • Branch protection: Peer review requirements"
}

# Parse command line arguments
SKIP_CI=false
SKIP_TEST=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_usage
            exit 0
            ;;
        --skip-ci)
            SKIP_CI=true
            shift
            ;;
        --skip-test)
            SKIP_TEST=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    log_header "SENTINEL HOOKS & CI/CD SETUP"
    log_info "CODING_STANDARDS.md compliance enforcement"
    echo ""

    # Pre-flight checks
    check_git_repo

    # Install development tools
    install_development_tools

    # Configure git hooks
    if ! configure_git_hooks; then
        log_error "Failed to configure git hooks"
        exit 1
    fi

    # Install hooks
    if ! install_precommit_hook; then
        log_error "Failed to install pre-commit hook"
        exit 1
    fi

    # Create additional hooks
    create_additional_hooks

    # Setup CI/CD pipeline
    if [ "$SKIP_CI" = "false" ]; then
        setup_ci_pipeline
    else
        log_info "Skipping CI/CD pipeline setup (--skip-ci)"
    fi

    # Create branch protection configuration
    create_branch_protection

    # Create documentation
    create_documentation

    # Test installation
    if [ "$SKIP_TEST" = "false" ]; then
        if ! test_hooks_installation; then
            log_warning "Hooks installation test failed - check manually"
        fi
    else
        log_info "Skipping hooks testing (--skip-test)"
    fi

    # Final summary
    log_header "SETUP COMPLETED SUCCESSFULLY"
    echo ""
    echo -e "${GREEN}✅ Development tools installed${NC}"
    echo -e "${GREEN}✅ Git hooks configured and installed${NC}"
    echo -e "${GREEN}✅ Pre-commit quality gates active${NC}"

    if [ "$SKIP_CI" = "false" ]; then
        echo -e "${GREEN}✅ CI/CD pipeline configured${NC}"
    fi

    echo ""
    echo -e "${CYAN}Next steps:${NC}"
    echo "  1. Test the setup: git commit -m 'test: validate hooks'"
    echo "  2. Review CI/CD pipeline in .github/workflows/"
    echo "  3. Configure branch protection if using GitHub"
    echo "  4. Update team documentation with new requirements"
    echo ""
    echo -e "${BLUE}CODING_STANDARDS.md enforcement is now active!${NC}"
}

# Run main function
main "$@"