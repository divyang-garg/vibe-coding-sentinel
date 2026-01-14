# Security Guide

## API Key Management

### ⚠️ Security Warning

**Never commit API keys to version control.** The `.sentinelsrc` file may contain sensitive credentials and should be kept secure.

### Recommended Practices

1. **Use Environment Variables (Recommended)**
   ```bash
   export SENTINEL_HUB_URL="https://your-hub-url.com"
   export SENTINEL_API_KEY="your-api-key-here"
   ```

2. **For CI/CD Pipelines**
   ```bash
   ./sentinel audit --api-key "$CI_API_KEY"
   ```
   Use the `--api-key` flag to pass the API key directly without storing it in files.

3. **File-Based Configuration (Less Secure)**
   If you must use `.sentinelsrc`:
   ```json
   {
     "hubUrl": "https://your-hub-url.com",
     "apiKey": "your-api-key-here"
   }
   ```
   
   **Important**: Set restrictive file permissions:
   ```bash
   chmod 600 .sentinelsrc
   ```
   
   This ensures only the file owner can read/write the file.

### File Permissions

The Sentinel Agent will warn you if `.sentinelsrc` has overly permissive permissions:
- ✅ **Good**: `600` (owner read/write only)
- ⚠️ **Warning**: `644` (world-readable)
- ❌ **Dangerous**: `666` (world-readable/writable)

### Priority Order

API keys are loaded in this order (first found wins):
1. `--api-key` command-line flag
2. `SENTINEL_API_KEY` environment variable
3. `.sentinelsrc` file (with security warnings)

### Rotating API Keys

If an API key is compromised:
1. Generate a new API key from the Hub dashboard
2. Update environment variables or `.sentinelsrc`
3. Revoke the old API key
4. Verify the new key works: `./sentinel audit`

### Best Practices

- ✅ Use environment variables in production
- ✅ Use `--api-key` flag in CI/CD pipelines
- ✅ Set `.sentinelsrc` permissions to `600` if used
- ✅ Never commit `.sentinelsrc` to git (already in `.gitignore`)
- ✅ Rotate API keys regularly
- ❌ Don't share API keys via email or chat
- ❌ Don't hardcode API keys in scripts
- ❌ Don't use the same API key across multiple projects

## Network Security

### HTTPS

Always use HTTPS for Hub URLs in production:
```bash
export SENTINEL_HUB_URL="https://hub.example.com"
```

### Firewall Rules

If running Hub API on-premises:
- Restrict access to Hub API port (default: 8080)
- Use VPN or private network for Hub access
- Implement IP whitelisting if possible

## Input Validation

The Sentinel Agent validates and sanitizes all inputs:
- File paths are checked for directory traversal (`../`)
- Codebase paths are validated to exist
- Malicious patterns are rejected

## Rate Limiting

The Hub API implements rate limiting:
- Comprehensive analysis: 2 requests/second
- Other endpoints: 10 requests/second
- Rate limit headers: `Retry-After` is returned on 429 responses

## Security Headers

The Hub API sets security headers:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security` (when using HTTPS)

## Reporting Security Issues

If you discover a security vulnerability:
1. **Do not** create a public issue
2. Email security@example.com (replace with your security contact)
3. Include details about the vulnerability
4. Allow time for the issue to be addressed before disclosure

## Compliance

- All API keys are stored securely
- No sensitive data is logged
- Audit trails are maintained for security analysis
- Regular security audits are performed










