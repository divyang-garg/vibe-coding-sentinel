# Sentinel Deployment Guide for VicecodingSentinel Project

## Deployment Readiness Assessment

### ✅ Prerequisites Check

1. **Go Compiler**: Required for building Sentinel binary
   - Status: ⚠️ **Not Installed**
   - Action: Install Go from https://go.dev/doc/install
   - Verify: Run `go version` after installation

2. **Git Repository**: Required for git hooks
   - Status: ⚠️ **Check Required**
   - Action: Run `git init` if not already a git repo

3. **Project Structure**: Ready for Sentinel
   - Status: ✅ **Ready**
   - The project has the build script (`synapsevibsentinel.sh`)
   - Windows wrappers are present (`sentinel.ps1`, `sentinel.bat`)

## Deployment Steps

### Step 1: Install Go (if not installed)

**macOS:**
```bash
brew install go
```

**Linux:**
```bash
sudo apt-get update
sudo apt-get install golang-go
```

**Windows:**
Download from https://go.dev/dl/ and install

**Verify installation:**
```bash
go version
```

### Step 2: Build Sentinel Binary

```bash
cd /Users/divyanggarg/VicecodingSentinel
chmod +x synapsevibsentinel.sh
./synapsevibsentinel.sh
```

Expected output:
```
⚙️  Compiling The Ultimate Sentinel...
🔨 Compiling Binary...
✅ Binary compiled successfully
🔒 Source Deleted.
✅ SENTINEL v24 READY.
```

### Step 3: Initialize Sentinel

```bash
./sentinel init
```

This will:
- Create `.cursor/rules/` directory with governance rules
- Create `.github/workflows/sentinel.yml` CI workflow
- Create `docs/knowledge/` directory
- Create `.sentinelsrc` configuration file
- Update `.gitignore` to exclude rules

**Non-interactive mode:**
```bash
./sentinel init --stack web --db none --non-interactive
```

### Step 4: Run Initial Audit

```bash
./sentinel audit
```

This will scan the project for:
- Security vulnerabilities
- Code quality issues
- Secrets detection
- Custom patterns (if configured)

### Step 5: Install Git Hooks (Optional but Recommended)

```bash
./sentinel install-hooks
```

This installs:
- `pre-commit` hook: Runs audit before commit
- `pre-push` hook: Runs audit before push
- `commit-msg` hook: Validates commit message format

### Step 6: Verify Deployment

```bash
# Verify hooks
./sentinel verify-hooks

# List rules
./sentinel list-rules

# Validate rules
./sentinel validate-rules
```

## Project-Specific Configuration

### Recommended `.sentinelsrc` Configuration

Since this is a Sentinel project itself, you may want to customize:

```json
{
  "scanDirs": [],
  "excludePaths": [
    "node_modules",
    ".git",
    "vendor",
    "dist",
    "build",
    ".next",
    "*.test.*",
    "*_test.go",
    "sentinel",
    "sentinel.exe",
    ".cursor/rules"
  ],
  "severityLevels": {
    "secrets": "critical",
    "console.log": "warning"
  },
  "customPatterns": {
    "hardcoded-path": "/Users/divyanggarg/",
    "temp-files": "\\.tmp$|\\.bak$"
  },
  "ruleLocations": [".cursor/rules"]
}
```

### Baseline Management

If you get false positives, add them to baseline:

```bash
# Example: Add a false positive
./sentinel baseline add synapsevibsentinel.sh 123 "pattern" "Reason: This is acceptable"

# List baselined findings
./sentinel baseline list

# Remove from baseline if fixed
./sentinel baseline remove synapsevibsentinel.sh 123
```

## CI/CD Integration

The `init` command creates `.github/workflows/sentinel.yml` which:
- Runs on every push and pull request
- Caches the binary for faster runs
- Fails the build if audit finds critical issues

**To enable:**
1. Commit the workflow file
2. Push to GitHub
3. GitHub Actions will automatically run audits

## Testing Deployment

### Test Commands

```bash
# Test audit with different outputs
./sentinel audit --output json --output-file audit-report.json
./sentinel audit --output html --output-file audit-report.html
./sentinel audit --output markdown --output-file audit-report.md

# Test with debug logging
./sentinel --debug audit

# Test docs generation
./sentinel docs

# Test rules management
./sentinel list-rules
./sentinel validate-rules
```

### Expected Results

After deployment, you should see:
- ✅ Binary compiled successfully
- ✅ Rules directory created
- ✅ Configuration file created
- ✅ Git hooks installed (if git repo exists)
- ✅ Audit runs without errors

## Troubleshooting Deployment

### Issue: "Go is required"
**Solution:** Install Go compiler (see Step 1)

### Issue: "Permission Denied"
**Solution:**
```bash
chmod +x synapsevibsentinel.sh
chmod +x sentinel
```

### Issue: "Not a git repository"
**Solution:**
```bash
git init
# Then run install-hooks again
```

### Issue: Binary not found after build
**Solution:**
- Check if build succeeded (look for "Binary compiled successfully")
- Verify `sentinel` file exists: `ls -la sentinel`
- Check file permissions: `chmod +x sentinel`

### Issue: Rules not working in Cursor
**Solution:**
1. Verify rules exist: `ls -la .cursor/rules/`
2. Check file extensions are `.md`: `file .cursor/rules/*.md`
3. Validate syntax: `./sentinel validate-rules`
4. Restart Cursor IDE

## Post-Deployment Checklist

- [ ] Go compiler installed and verified
- [ ] Sentinel binary built successfully
- [ ] `init` command completed without errors
- [ ] Rules directory created with files
- [ ] Configuration file created
- [ ] Initial audit runs successfully
- [ ] Git hooks installed (if using git)
- [ ] CI workflow file created
- [ ] Documentation updated

## Next Steps After Deployment

1. **Customize Rules**: Edit `.cursor/rules/*.md` files for project-specific governance
2. **Configure Patterns**: Add custom security patterns in `.sentinelsrc`
3. **Set Up Baseline**: Add known false positives to baseline
4. **Enable CI**: Commit and push to enable GitHub Actions
5. **Team Onboarding**: Share README.md with team members

## Deployment Status

**Current Status**: ⚠️ **Ready for Deployment** (requires Go installation)

**Blockers**:
- Go compiler not installed

**Ready**:
- ✅ Build script present
- ✅ Windows wrappers present
- ✅ Project structure suitable
- ✅ All code changes implemented

Once Go is installed, deployment should take < 5 minutes.



