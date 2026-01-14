# Cursor IDE Compatibility Analysis

## File Format Verification

### Current Implementation
The script generates files with `.mdc` extension:
```
.cursor/rules/00-constitution.mdc
.cursor/rules/01-firewall.mdc
.cursor/rules/web.mdc
```

### Cursor Expected Format
Cursor IDE recognizes rules in these formats:
1. **Single file**: `.cursorrules` in project root
2. **Multiple files**: `.cursor/rules/*.md` (Markdown files)

**⚠️ CRITICAL**: The `.mdc` extension is **NOT** standard. Cursor expects `.md` files.

### Frontmatter Format
The YAML frontmatter format used appears correct:
```yaml
---
description: Universal Laws.
globs: ["**/*"]
alwaysApply: true
---
```

However, verify these fields are supported:
- `description`: Likely supported (metadata)
- `globs`: Should work (file matching)
- `alwaysApply`: **Needs verification** - may not be a standard Cursor field

### Testing Cursor Compatibility

To verify if rules work:

1. **Manual Test**:
   ```bash
   # After running ./sentinel init
   # Rename files to .md
   cd .cursor/rules
   for f in *.mdc; do mv "$f" "${f%.mdc}.md"; done
   ```

2. **Check Cursor Recognition**:
   - Open Cursor IDE
   - Open a file matching the glob patterns
   - Use Cursor's AI chat
   - Verify if rules are being applied (check if AI follows the rules)

3. **Alternative Format Check**:
   Cursor may also support `.cursorrules` file in root:
   ```bash
   # Create single file version
   cat .cursor/rules/*.md > .cursorrules
   ```

## Compatibility Status

| Feature | Status | Notes |
|---------|--------|-------|
| Directory `.cursor/rules/` | ✅ Correct | Standard location |
| File Extension | ✅ Fixed | Changed to `.md` |
| Frontmatter Format | ⚠️ Unknown | Needs testing |
| Glob Patterns | ⚠️ Unknown | Syntax needs verification |
| Numbered Prefixes | ✅ Likely OK | Common pattern for ordering |
| alwaysApply Field | ⚠️ Unknown | May not be supported |

## Recommended Fixes

### Completed
1. ✅ Changed file extension from `.mdc` to `.md` in `writeFile()` calls (2024-12-10)
2. ✅ Updated `.gitignore` pattern to match `.md` files (already correct)
3. ⚠️ Test with actual Cursor IDE installation (pending manual verification)

### Verification Steps
1. Install Cursor IDE
2. Run `./sentinel init` on a test project
3. Manually rename `.mdc` → `.md`
4. Open project in Cursor
5. Test if rules are applied in AI chat
6. Check Cursor's rules panel (if available)

## Alternative Approach

If `.mdc` is intentional (custom format), consider:
1. Document why custom extension is used
2. Add conversion script: `.mdc` → `.md`
3. Or use Cursor's native `.cursorrules` single-file format





