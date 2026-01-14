# Sentinel API Reference

## Overview

The Sentinel API provides comprehensive code analysis, security scanning, and automated fixing capabilities through a command-line interface and optional Hub integration.

## Architecture

Sentinel operates in two modes:
- **Local Mode**: Standalone operation with no external dependencies
- **Hub Mode**: Enhanced capabilities with server-side processing and team collaboration

## Core Commands

### Initialization

#### `sentinel init`
Initialize a new Sentinel project in the current directory.

```bash
sentinel init
```

**What it does:**
- Creates `.sentinel/` directory for local data
- Initializes basic configuration
- Sets up project structure detection

### Code Analysis

#### `sentinel audit [options] [path]`
Perform comprehensive security and code quality analysis.

```bash
# Basic audit
sentinel audit

# Offline mode (no Hub communication)
sentinel audit --offline

# CI/CD mode (non-interactive)
sentinel audit --ci

# Custom output format
sentinel audit --output json --output-file results.json

# Deep analysis (advanced scanning)
sentinel audit --deep

# Vibe analysis (code quality)
sentinel audit --vibe-check
```

**Parameters:**
- `path`: Optional path to scan (defaults to current directory)
- `--offline`: Disable Hub communication
- `--ci`: CI/CD mode with appropriate exit codes
- `--output`: Output format (text, json, xml)
- `--output-file`: Save results to file
- `--deep`: Enable advanced recursive analysis
- `--vibe-check`: Include code quality analysis

**Exit Codes:**
- `0`: Audit passed (no critical issues)
- `1`: Audit failed (critical issues found)
- `2`: Audit error (configuration or execution error)

### Pattern Learning

#### `sentinel learn [options]`
Analyze codebase patterns and generate development guidelines.

```bash
# Full pattern learning
sentinel learn

# Naming conventions only
sentinel learn --naming
```

**What it generates:**
- `.sentinel/patterns.json`: Structured pattern data
- `.cursor/rules/project-patterns.md`: Cursor-compatible rules

### Auto-Fix

#### `sentinel fix [options] [path]`
Automatically fix common code issues.

```bash
# Safe mode (dry-run)
sentinel fix --safe

# Force fixes
sentinel fix --safe --yes

# Fix specific path
sentinel fix src/
```

**Fixes Applied:**
- Console.log statement removal
- Debugger statement removal
- Trailing whitespace cleanup
- Import sorting and organization
- Unused import detection

**Options:**
- `--safe`: Dry-run mode (no file modifications)
- `--yes`: Force modifications even in safe mode
- `path`: Specific path to fix (defaults to current directory)

### Task Management

#### `sentinel tasks <subcommand> [options]`
Manage development tasks and track progress.

```bash
# List all tasks
sentinel tasks list

# List tasks by status
sentinel tasks list --status pending

# Scan codebase for new tasks
sentinel tasks scan

# Verify task completion
sentinel tasks verify <task-id>

# Complete a task
sentinel tasks complete <task-id> --reason "Implementation complete"
```

**Subcommands:**
- `list`: Display all tasks
- `scan`: Discover new tasks in codebase
- `verify`: Check task implementation status
- `complete`: Mark task as completed
- `dependencies`: Show task relationships

### Documentation

#### `sentinel docs [options]`
Generate and update project documentation.

```bash
# Generate documentation
sentinel docs

# Check documentation-code synchronization
sentinel doc-sync

# Auto-fix documentation issues
sentinel doc-sync --fix
```

### Status & Monitoring

#### `sentinel status`
Display current project status and health metrics.

```bash
sentinel status
```

**Shows:**
- Pattern learning status
- Recent audit results
- Task completion metrics
- Configuration health

### Configuration

#### Configuration File (.sentinelsrc)

```json
{
  "hubUrl": "https://your-hub-instance.com",
  "apiKey": "your-api-key",
  "scanDirs": ["src", "lib", "tests"],
  "excludePaths": [".git", "node_modules", "dist"],
  "severityLevels": {
    "secrets": "critical",
    "console.log": "warning",
    "NOLOCK": "critical"
  }
}
```

#### Environment Variables

- `SENTINEL_HUB_URL`: Hub server URL
- `SENTINEL_API_KEY`: Authentication key for Hub
- `SENTINEL_LOG_LEVEL`: Logging verbosity (DEBUG, INFO, WARN, ERROR)
- `SENTINEL_CONFIG`: Custom config file path

## Hub Integration APIs

When connected to a Sentinel Hub, additional capabilities are available:

### Comprehensive Analysis

```bash
# Advanced analysis with Hub
sentinel audit  # Automatically uses Hub if configured
```

**Hub Features:**
- Cross-repository analysis
- Team collaboration
- Historical trend analysis
- Advanced AI-powered detection

### Document Synchronization

```bash
# Check doc-code alignment
sentinel doc-sync --report
```

### Task Collaboration

```bash
# Team task management
sentinel tasks list --assigned-to team
```

## Error Handling

### Common Exit Codes

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | None |
| 1 | Security issues found | Review and fix issues |
| 2 | Configuration error | Check .sentinelsrc and environment |
| 3 | Network error | Verify Hub connectivity |
| 4 | File access error | Check permissions and paths |
| 5 | Internal error | Report to development team |

### Error Messages

**Configuration Issues:**
```
❌ Hub not configured. Set SENTINEL_HUB_URL and SENTINEL_API_KEY
```

**Network Issues:**
```
❌ Hub request failed: connection refused
```

**File Issues:**
```
❌ Cannot read file: permission denied
```

## Best Practices

### Development Workflow

1. **Initialize**: `sentinel init`
2. **Learn Patterns**: `sentinel learn`
3. **Regular Audits**: `sentinel audit --offline`
4. **Fix Issues**: `sentinel fix --safe`
5. **Track Tasks**: `sentinel tasks scan && sentinel tasks list`

### CI/CD Integration

```yaml
# .github/workflows/security.yml
name: Security Audit
on: [push, pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Sentinel Audit
        run: |
          curl -L https://github.com/your-org/sentinel/releases/download/v24/sentinel -o sentinel
          chmod +x sentinel
          ./sentinel audit --ci --offline
```

### Performance Optimization

- Use `--offline` for local development
- Limit scan directories with `.sentinelsrc`
- Use `--deep` only when needed
- Run pattern learning periodically, not on every commit

## Troubleshooting

### Common Issues

**"Command not found"**
- Ensure sentinel is in PATH or use `./sentinel`
- Check file permissions: `chmod +x sentinel`

**"Hub connection failed"**
- Verify `SENTINEL_HUB_URL` and `SENTINEL_API_KEY`
- Check network connectivity
- Confirm Hub is running and accessible

**"No patterns detected"**
- Run `sentinel learn` first
- Check that project has sufficient code files
- Verify file extensions are supported

**"Fix didn't work"**
- Use `--safe` first to preview changes
- Check file permissions
- Ensure no external processes have files locked

### Debug Mode

Enable verbose logging:
```bash
export SENTINEL_LOG_LEVEL=DEBUG
sentinel audit
```

### Getting Help

```bash
# Show all commands
sentinel --help

# Command-specific help
sentinel audit --help
sentinel tasks --help
```

## Version Information

- **Current Version**: v24 (Ultimate)
- **Architecture**: Go-based cross-platform binary
- **Dependencies**: None (self-contained)
- **Hub Compatibility**: v24+ API

---

For additional support, see the [User Guide](./USER_GUIDE.md) or visit the project repository.



