# 🛡️ Synapse Sentinel v24

**Production-Ready Governance Engine for Cursor IDE**

Synapse Sentinel is a compiled, tamper-proof governance binary that automates the setup, security, and maintenance of AI-augmented software projects. It enforces Agency Standards across Web, Mobile, Commerce, and AI service lines without exposing proprietary prompt engineering.

## 🚀 Quick Start

### Installation

1. **Build the binary:**
   ```bash
   chmod +x synapsevibsentinel.sh
   ./synapsevibsentinel.sh
   ```

2. **Initialize your project:**
   ```bash
   ./sentinel init
   ```

3. **Run an audit:**
   ```bash
   ./sentinel audit
   ```

### Windows Installation

1. **Build using WSL or Git Bash:**
   ```bash
   ./synapsevibsentinel.sh
   ```

2. **Use PowerShell wrapper:**
   ```powershell
   .\sentinel.ps1 init
   .\sentinel.ps1 audit
   ```

   Or use batch file:
   ```cmd
   sentinel.bat init
   sentinel.bat audit
   ```

## 📖 Commands

### `init` - Bootstrap Project

Initializes Sentinel in your project with interactive or non-interactive mode.

**Interactive mode:**
```bash
./sentinel init
```

**Non-interactive mode:**
```bash
./sentinel init --stack web --db sql --protocol soap
```

**Environment variables:**
```bash
export SENTINEL_STACK=web
export SENTINEL_DB=sql
export SENTINEL_PROTOCOL=soap
./sentinel init --non-interactive
```

**Options:**
- `--stack`: Service line (web, mobile-cross, mobile-native, commerce, ai)
- `--db`: Database type (sql, nosql, none)
- `--protocol`: Protocol support (soap, none)
- `--non-interactive`, `-y`: Skip interactive prompts
- `--config`: Path to config file

### `audit` - Security & Logic Scan

Scans codebase for security vulnerabilities and code quality issues.

**Basic usage:**
```bash
./sentinel audit
```

**Output formats:**
```bash
# JSON output
./sentinel audit --output json --output-file report.json

# HTML report
./sentinel audit --output html --output-file report.html

# Markdown report
./sentinel audit --output markdown --output-file report.md
```

**Environment variables:**
- `SENTINEL_OUTPUT`: Output format (text, json, html, markdown)
- `SENTINEL_OUTPUT_FILE`: Output file path
- `SENTINEL_LOG_LEVEL`: Log level (debug, info, warn, error)

### `docs` - Update Context Map

Updates the AI's context window with current project structure.

```bash
./sentinel docs
```

### `list-rules` - List Active Rules

Lists all active Cursor rules with their frontmatter.

```bash
./sentinel list-rules
```

### `validate-rules` - Validate Rule Syntax

Validates all rule files for correct frontmatter format.

```bash
./sentinel validate-rules
```

### `verify-hooks` - Verify Git Hooks

Verifies that git hooks are properly installed and functional.

```bash
./sentinel verify-hooks
```

### `baseline` - Manage Baseline/Allowlist

Manage accepted findings (false positives or known issues) to prevent them from appearing in audit reports.

**Add finding to baseline:**
```bash
./sentinel baseline add src/file.js 42 "console.log" "Debug code, will remove later"
```

**List baselined findings:**
```bash
./sentinel baseline list
```

**Remove finding from baseline:**
```bash
./sentinel baseline remove src/file.js 42
```

The baseline is stored in `.sentinel-baseline.json` and automatically filters findings during audits.

### `update-rules` - Update Rules

Update Cursor rules without recompiling the binary.

```bash
./sentinel update-rules
```

This validates existing rules and prepares for future rule update mechanisms.

## ⚙️ Configuration

Sentinel uses `.sentinelsrc` configuration file (JSON format).

**Example `.sentinelsrc`:**
```json
{
  "scanDirs": ["src", "lib"],
  "excludePaths": ["node_modules", ".git", "vendor", "dist", "build", ".next", "*.test.*", "*_test.go"],
  "severityLevels": {
    "secrets": "critical",
    "console.log": "warning",
    "NOLOCK": "critical",
    "$where": "critical",
    "simplexml_load_string": "warning",
    "custom-pattern-1": "critical"
  },
  "customPatterns": {
    "custom-pattern-1": "(?i)password\\s*=\\s*['\"][^'\"]+['\"]",
    "custom-pattern-2": "TODO.*FIXME"
  },
  "ruleLocations": [".cursor/rules"]
}
```

**Custom Patterns:**
You can define project-specific security patterns in `customPatterns`. Each pattern will be scanned during audits with the severity level specified in `severityLevels`.

**Configuration locations (checked in order):**
1. `.sentinelsrc` in project root
2. `~/.sentinelsrc` in home directory
3. Environment variables (`SENTINEL_*`)

## 🔍 Security Scans

Sentinel performs comprehensive security scans:

- **Secrets Detection**: API keys, tokens, passwords (with entropy checking)
- **SQL Injection**: String concatenation, dynamic SQL execution
- **XSS Vulnerabilities**: innerHTML usage, eval() detection
- **Database Safety**: MSSQL NOLOCK, MongoDB $where patterns
- **XXE Vulnerabilities**: simplexml_load_string detection
- **Insecure Random**: Math.random(), insecure PRNG usage
- **Hardcoded Credentials**: Credentials in URLs
- **Debug Code**: console.log detection
- **Custom Patterns**: Project-specific patterns defined in `.sentinelsrc`

### Scan Features

- **Binary File Detection**: Automatically skips binary files (images, executables, etc.)
- **Symlink Handling**: Safely handles symlinks, preventing traversal outside project directory
- **File Size Limits**: Skips files larger than 10MB for performance
- **Smart Filtering**: Excludes test files, comments, and baselined findings
- **Cross-Platform**: Works identically on Linux, macOS, and Windows

## 📊 Reporting

Sentinel supports multiple output formats:

- **Text**: Human-readable console output (default)
- **JSON**: Machine-readable structured data
- **HTML**: Formatted HTML report with styling
- **Markdown**: Markdown-formatted report

Reports include:
- File paths and line numbers
- Severity levels (critical, warning, info)
- Code context around findings
- Summary statistics

## 🛠️ Development

### Debug Mode

Enable debug logging:
```bash
./sentinel --debug audit
# or
export SENTINEL_LOG_LEVEL=debug
./sentinel audit
```

### Build from Source

The script compiles a Go binary. To rebuild:
```bash
./synapsevibsentinel.sh
```

Build optimization includes:
- Binary freshness checking (skips rebuild if up-to-date)
- Size optimization flags (`-ldflags="-s -w"`)
- Automatic cleanup of source files

### CI/CD Integration

The generated `.github/workflows/sentinel.yml` includes:
- Binary caching to avoid unnecessary rebuilds
- Automatic audit on push and pull requests
- Proper error handling and build failure on audit issues

The CI workflow uses GitHub Actions cache to store the compiled binary, significantly speeding up subsequent runs.

## 🔐 Security Features

- **Tamper Proof**: Compiled binary prevents modification of audit logic
- **IP Protection**: Rules are embedded in binary, not visible in plain text
- **Cross-Platform**: Works on Linux, macOS, and Windows (including Windows `docs` command)
- **Configurable**: Flexible configuration system with custom patterns
- **Comprehensive**: 12+ built-in security scan patterns + custom patterns
- **Baseline System**: Manage false positives and accepted findings
- **Error Handling**: Comprehensive error handling with detailed logging
- **Concurrent Execution Protection**: File-based locking prevents multiple instances
- **Path Validation**: Prevents path traversal attacks
- **Binary Detection**: Automatically skips binary files to prevent false positives

## 📁 Project Structure

After running `init`, Sentinel creates:

```
.cursor/rules/          # Cursor IDE rules (hidden from git)
  ├── 00-constitution.md
  ├── 01-firewall.md
  ├── web.md (or mobile.md, etc.)
  └── db-sql.md (or db-nosql.md)

.github/workflows/
  └── sentinel.yml      # CI/CD workflow (with caching)

docs/knowledge/
  ├── client-brief.md
  └── file-structure.txt

.sentinelsrc            # Configuration file
.sentinel-baseline.json  # Baseline/allowlist (created when using baseline command)
```

## 🐛 Troubleshooting

### "Permission Denied"
```bash
chmod +x sentinel
```

### "Go is required"
Install Go from https://go.dev/doc/install

### "No source directories found"
Check your `.sentinelsrc` configuration or ensure source directories exist.

### Rules not working in Cursor
1. Verify rules are `.md` files (not `.mdc`)
2. Run `./sentinel validate-rules` to check syntax
3. Ensure frontmatter format is correct

### Git hooks not working
1. Run `./sentinel verify-hooks` to check hook status
2. Ensure binary is accessible (check PATH or use `./sentinel install-hooks` again)
3. On Windows, ensure `.exe` extension is handled correctly

### Too many false positives
1. Use `./sentinel baseline add` to mark acceptable findings
2. Customize patterns in `.sentinelsrc` to match your project needs
3. Adjust severity levels in configuration

### "Another Sentinel instance is running"
This is normal - Sentinel uses file-based locking to prevent concurrent execution. Wait for the other instance to complete or remove `/tmp/sentinel.lock` (or `%TEMP%\sentinel.lock` on Windows) if the previous instance crashed.

### Performance issues on large codebases
- Sentinel automatically skips files larger than 10MB
- Binary files are automatically detected and skipped
- Use `excludePaths` in `.sentinelsrc` to exclude large directories
- Consider using `scanDirs` to limit scanning to specific directories

## 📝 License

Property of Synapse Engineering Strategy.

## ✨ Recent Improvements (v24 Enhanced)

### Critical Fixes
- ✅ Fixed root-level file detection (filepath.Glob brace expansion)
- ✅ Cross-platform `docs` command (replaced Unix `find` with Go-native implementation)
- ✅ Dynamic git hooks binary detection (works with PATH and Windows)
- ✅ Comprehensive error handling throughout
- ✅ CI workflow optimization with caching

### New Features
- ✅ Custom pattern scanning from configuration
- ✅ Baseline/allowlist system for managing false positives
- ✅ Git hooks verification command
- ✅ Rule update mechanism
- ✅ Binary file detection and skipping
- ✅ Symlink handling with security checks
- ✅ Concurrent execution protection
- ✅ File size limits (10MB) for performance

### Improvements
- ✅ Enhanced error messages with context
- ✅ Better logging with debug mode
- ✅ Improved cross-platform compatibility
- ✅ More robust file handling

## 🤝 Contributing

This is an internal tool. For issues or improvements, contact Synapse Engineering.

---

**Maintained by Synapse Engineering.**

**Version**: v24 Enhanced  
**Last Updated**: 2024

