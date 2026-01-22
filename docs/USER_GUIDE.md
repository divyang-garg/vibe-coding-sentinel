# Sentinel User Guide

## Welcome to Sentinel

Sentinel is an AI-powered code analysis and security tool that helps development teams maintain high-quality, secure codebases. This guide will walk you through getting started and making the most of Sentinel's capabilities.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Core Concepts](#core-concepts)
3. [Daily Workflow](#daily-workflow)
4. [Advanced Features](#advanced-features)
5. [Team Collaboration](#team-collaboration)
6. [Troubleshooting](#troubleshooting)
7. [Best Practices](#best-practices)

## Quick Start

### Installation

```bash
# Download the latest release
curl -L https://github.com/your-org/sentinel/releases/download/v24/sentinel-linux-amd64 -o sentinel
chmod +x sentinel

# Or for macOS
curl -L https://github.com/your-org/sentinel/releases/download/v24/sentinel-darwin-amd64 -o sentinel
chmod +x sentinel
```

### First Project

```bash
# Navigate to your project
cd my-project

# Initialize Sentinel
./sentinel init

# Learn your code patterns
./sentinel learn

# Run your first audit
./sentinel audit --offline

# Fix any issues found
./sentinel fix --safe
```

That's it! Sentinel is now monitoring your codebase.

## Core Concepts

### Local vs Hub Mode

**Local Mode** (Recommended for individual developers):
- No external dependencies
- Fast, offline operation
- Core security and quality checks
- Pattern learning and auto-fixing

**Hub Mode** (Recommended for teams):
- Server-side processing
- Cross-repository analysis
- Team collaboration features
- Advanced AI capabilities

### Key Components

#### Pattern Learning
Sentinel analyzes your codebase to understand:
- Programming languages used
- Framework preferences
- Naming conventions
- Project structure
- Code quality patterns

#### Security Scanning
Comprehensive security analysis including:
- Secrets detection (API keys, passwords)
- SQL injection vulnerabilities
- XSS and other injection attacks
- Insecure configurations
- Deprecated API usage

#### Auto-Fix
Automated code improvements:
- Remove debug statements
- Fix formatting issues
- Sort imports
- Clean up unused code
- Apply consistent styling

## Daily Workflow

### Morning Setup

```bash
# Check project status
./sentinel status

# Update patterns if team made changes
./sentinel learn

# Run security audit
./sentinel audit --offline
```

### During Development

```bash
# Before committing
./sentinel audit --ci

# Fix any issues
./sentinel fix --safe

# Commit with confidence
git commit -m "feat: add user authentication"
```

### End of Day

```bash
# Final audit
./sentinel audit --offline

# Update documentation
./sentinel docs

# Check task status
./sentinel tasks list
```

## Advanced Features

### Configuration

Create a `.sentinelsrc` file for project-specific settings:

```json
{
  "hubUrl": "https://your-team-hub.com",
  "apiKey": "your-api-key",
  "scanDirs": ["src", "lib", "packages"],
  "excludePaths": [".git", "node_modules", "dist", "build"],
  "severityLevels": {
    "secrets": "critical",
    "console.log": "warning",
    "eval": "critical"
  }
}
```

### Custom Rules

Sentinel generates Cursor-compatible rules in `.cursor/rules/project-patterns.md`:

```markdown
# Project Code Patterns

## Naming Conventions
- Functions and variables: camelCase
- Classes: PascalCase
- Constants: UPPER_SNAKE_CASE

## Frameworks
- React for UI components
- Express.js for API server
- Jest for testing
```

### Task Management

```bash
# Discover new tasks in code
./sentinel tasks scan

# List all tasks
./sentinel tasks list

# Focus on high-priority tasks
./sentinel tasks list --priority critical

# Mark task complete
./sentinel tasks complete TASK-123 --reason "Implemented user auth"
```

### Deep Analysis

For comprehensive analysis:

```bash
# Deep security scan
./sentinel audit --deep

# Include code quality checks
./sentinel audit --vibe-check

# Generate detailed report
./sentinel audit --output json --output-file security-report.json
```

## Team Collaboration

### Hub Setup

1. **Deploy Sentinel Hub** (optional)
   ```bash
   docker run -p 8080:8080 sentinel/hub:latest
   ```

2. **Create a Project and Get API Key**
   
   When you create a project in the Hub, an API key is automatically generated:
   ```bash
   # Create project via Hub API
   curl -X POST https://hub.yourcompany.com/api/v1/projects \
     -H "Content-Type: application/json" \
     -H "X-API-Key: admin-key" \
     -d '{"name": "My Project"}'
   
   # Response includes api_key - SAVE THIS!
   # {
   #   "id": "proj_123",
   #   "api_key": "xK9mP2qR7vT4wY8zA1bC3dE5fG6hI0j",  # ⚠️ Save immediately!
   #   "api_key_prefix": "xK9mP2qR",
   #   ...
   # }
   ```

3. **Configure team access**
   ```bash
   export SENTINEL_HUB_URL="https://hub.yourcompany.com"
   export SENTINEL_API_KEY="your-project-api-key"  # From step 2
   ```

4. **Share patterns across repositories**
   ```bash
   ./sentinel learn  # Patterns automatically sync to Hub
   ```

### API Key Management

If you need to manage your API keys:

**Generate a new API key:**
```bash
curl -X POST https://hub.yourcompany.com/api/v1/projects/proj_123/api-key \
  -H "X-API-Key: admin-key"
# Response includes new api_key - save it!
```

**Check API key status:**
```bash
curl -X GET https://hub.yourcompany.com/api/v1/projects/proj_123/api-key \
  -H "X-API-Key: admin-key"
# Returns prefix only (for security)
```

**Revoke an API key:**
```bash
curl -X DELETE https://hub.yourcompany.com/api/v1/projects/proj_123/api-key \
  -H "X-API-Key: admin-key"
```

**Important Security Notes:**
- API keys are only shown once when generated
- Store keys securely (environment variables, secret management)
- Never commit keys to version control
- Rotate keys regularly for security

For detailed API key management, see [API Key Management Guide](../API_KEY_MANAGEMENT_GUIDE.md).

### Cross-Repository Analysis

```bash
# Analyze multiple repos
./sentinel audit --repos "frontend,backend,shared"

# Team-wide task tracking
./sentinel tasks list --team

# Organization-wide patterns
./sentinel learn --org
```

### Code Review Integration

```bash
# Pre-commit hooks
./sentinel audit --ci

# Pull request checks
./sentinel audit --pr --base main

# Automated fixes for PRs
./sentinel fix --pr 123
```

## Troubleshooting

### Common Issues

#### "Audit failed with exit code 1"
This is normal! It means Sentinel found security issues that need attention.

**Solution:**
```bash
# See what was found
./sentinel audit --offline

# Fix automatically where possible
./sentinel fix --safe

# Review remaining issues manually
```

#### "Pattern learning found nothing"
Your project might be too small or use unsupported languages.

**Solution:**
- Ensure you have at least 5-10 code files
- Check supported languages: JavaScript, TypeScript, Python, Go, Java, etc.
- Run in a subdirectory: `cd src && ../sentinel learn`

#### "Hub connection failed"
Network or authentication issue.

**Solution:**
```bash
# Test connectivity
curl -H "X-API-Key: $SENTINEL_API_KEY" $SENTINEL_HUB_URL/api/health

# Verify API key is valid
curl -X GET $SENTINEL_HUB_URL/api/v1/projects/your-project-id/api-key \
  -H "X-API-Key: $SENTINEL_API_KEY"

# If key is invalid, generate a new one
curl -X POST $SENTINEL_HUB_URL/api/v1/projects/your-project-id/api-key \
  -H "X-API-Key: admin-key"

# Use offline mode if Hub is unavailable
./sentinel audit --offline

# Check configuration
cat .sentinelsrc
```

#### "Fix didn't change files"
Using safe mode or permission issues.

**Solution:**
```bash
# Use force mode
./sentinel fix --safe --yes

# Check permissions
ls -la files-to-fix/

# Run as appropriate user
sudo ./sentinel fix
```

### Performance Issues

#### Slow Audits
- Use `--offline` for local development
- Limit scan directories in `.sentinelsrc`
- Exclude large directories: `node_modules`, `dist`, etc.

#### High Memory Usage
- Reduce scan scope
- Use `--deep` only when needed
- Close other memory-intensive applications

#### Storage Issues
- Clean old backups: `rm -rf .sentinel/backups/old-*`
- Limit backup retention in configuration
- Use external storage for large projects

### Debug Mode

Enable detailed logging:

```bash
export SENTINEL_LOG_LEVEL=DEBUG
./sentinel audit 2>&1 | tee audit.log
```

Check the logs for specific error messages and stack traces.

## Best Practices

### Development Workflow

1. **Initialize Early**
   ```bash
   ./sentinel init  # Run once per project
   ```

2. **Learn Patterns Regularly**
   ```bash
   ./sentinel learn  # Run weekly or when patterns change
   ```

3. **Audit Frequently**
   ```bash
   ./sentinel audit --offline  # Run daily, before commits
   ```

4. **Fix Incrementally**
   ```bash
   ./sentinel fix --safe  # Review changes before applying
   ```

### Team Standards

#### Code Review Checklist
- [ ] `sentinel audit --ci` passes
- [ ] No critical security issues
- [ ] Patterns match team standards
- [ ] Documentation updated

#### CI/CD Pipeline
```yaml
# .github/workflows/quality.yml
name: Code Quality
on: [push, pull_request]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install Sentinel
        run: |
          curl -L https://github.com/your-org/sentinel/releases/download/v24/sentinel -o sentinel
          chmod +x sentinel
      - name: Security Audit
        run: ./sentinel audit --ci --offline
      - name: Pattern Check
        run: ./sentinel learn --verify
      - name: Auto-Fix
        run: ./sentinel fix --safe --yes
```

### Configuration Management

#### Project-Specific Settings
```json
{
  "scanDirs": ["src", "lib"],
  "excludePaths": ["test", "docs"],
  "customPatterns": {
    "no-todos": "TODO|FIXME|HACK"
  }
}
```

#### Organization Standards
```json
{
  "severityLevels": {
    "secrets": "critical",
    "unused-imports": "warning",
    "long-lines": "info"
  },
  "requiredPatterns": ["camelCase", "PascalCase"],
  "forbiddenPatterns": ["snake_case"]
}
```

### Security Considerations

#### API Key Management
- Never commit API keys to version control
- Use environment variables: `SENTINEL_API_KEY`
- Rotate keys regularly
- Use different keys for different environments

#### File Permissions
```bash
# Secure configuration
chmod 600 .sentinelsrc

# Secure backups
chmod 700 .sentinel/backups
```

#### Network Security
- Use HTTPS for Hub communication
- Implement IP whitelisting
- Monitor API usage
- Regular security audits

## Support and Resources

### Getting Help

- **Documentation**: See [API Reference](./api/API_REFERENCE.md)
- **Community**: Join our Discord/Slack community
- **Issues**: Report bugs on GitHub
- **Security**: Report vulnerabilities privately

### Version Compatibility

| Sentinel Version | Hub Version | Features |
|------------------|-------------|----------|
| v24 | v24+ | Full feature set |
| v23 | v23+ | Core features |
| v22 | v22+ | Basic functionality |

### Migration Guide

#### Upgrading from v23
```bash
# Backup configuration
cp .sentinelsrc .sentinelsrc.backup

# Update binary
curl -L https://github.com/your-org/sentinel/releases/download/v24/sentinel -o sentinel.new
chmod +x sentinel.new
mv sentinel.new sentinel

# Re-learn patterns
./sentinel learn

# Test configuration
./sentinel audit --offline
```

### Roadmap

#### Upcoming Features
- **Real-time monitoring** (v25)
- **IDE integrations** (v25)
- **Advanced AI analysis** (v26)
- **Multi-language support** (v26)

#### Contributing
We welcome contributions! See our [Contributing Guide](./CONTRIBUTING.md).

---

**Happy coding with Sentinel!** 🛡️

*Keep your code secure, your patterns consistent, and your team productive.*



