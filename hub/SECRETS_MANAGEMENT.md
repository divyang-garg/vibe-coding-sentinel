# Secrets Management Guide

## Overview
All sensitive configuration values are managed via environment variables loaded from `.env` file.

## Setup

### Development
1. Copy `.env.example` to `.env`
2. Fill in values (defaults are acceptable for local development)
3. `.env` is gitignored and will not be committed

### Production
1. Use `scripts/generate-secrets.sh` to generate secure secrets
2. Update `CORS_ORIGIN` and `HUB_URL` with actual production values
3. Store `.env` securely:
   - Docker Swarm: Use Docker secrets
   - Kubernetes: Use Kubernetes secrets
   - Cloud: Use cloud secrets manager (AWS Secrets Manager, Azure Key Vault, etc.)

## Required Environment Variables

### Database Configuration
- `DATABASE_URL`: PostgreSQL connection string (default: `postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable`)
- `DB_PASSWORD`: Database password (32+ characters recommended) - part of DATABASE_URL

### Authentication & Security
- `ADMIN_API_KEY`: Admin API key for admin endpoints (64 hex characters recommended)
- `JWT_SECRET`: JWT signing secret (64+ characters recommended)
- `CORS_ORIGIN`: Allowed CORS origin (default: `*` - change to specific domain in production)

### Storage Paths
- `DOCUMENT_STORAGE`: Path for document storage (default: `/data/documents`)
- `BINARY_STORAGE`: Path for binary storage (default: `/data/binaries`)
- `RULES_STORAGE`: Path for rules storage (default: `/data/rules`)

### Server Configuration
- `PORT`: Server port (default: `8080`)

### LLM Provider (Optional)
- `AZURE_AI_KEY`: Azure AI Foundry API key (if using Azure AI)

## Advanced Configuration (Optional)

### Timeout Settings
- `SENTINEL_DB_TIMEOUT`: Database query timeout (default: `10s`)
- `SENTINEL_ANALYSIS_TIMEOUT`: Analysis operation timeout (default: `60s`)
- `SENTINEL_HTTP_TIMEOUT`: HTTP request timeout (default: `30s`)
- `SENTINEL_CONTEXT_TIMEOUT`: Default context timeout (default: `30s`)

### Limits Configuration
- `SENTINEL_MAX_FILE_SIZE`: Maximum file size in bytes (default: `104857600` - 100MB)
- `SENTINEL_MAX_STRING_LENGTH`: Maximum string length (default: `1000000` - 1M chars)
- `SENTINEL_MAX_REQUEST_SIZE`: Maximum HTTP request size (default: `10485760` - 10MB)
- `SENTINEL_RATE_LIMIT_RPS`: Rate limit requests per second (default: `100`)
- `SENTINEL_RATE_LIMIT_BURST`: Rate limit burst size (default: `200`)

### Task Management Limits
- `SENTINEL_MAX_TASK_TITLE_LENGTH`: Maximum task title length (default: `500`)
- `SENTINEL_MAX_TASK_DESCRIPTION_LENGTH`: Maximum task description length (default: `5000`)
- `SENTINEL_DEFAULT_TASK_LIST_LIMIT`: Default task list limit (default: `50`)
- `SENTINEL_MAX_TASK_LIST_LIMIT`: Maximum task list limit (default: `1000`)
- `SENTINEL_DEFAULT_DATE_RANGE_DAYS`: Default date range in days (default: `30`)

### Cache Configuration
- `SENTINEL_CACHE_TTL`: Default cache TTL (default: `5m`)
- `SENTINEL_TASK_CACHE_TTL`: Task cache TTL (default: `5m`)
- `SENTINEL_VERIFICATION_CACHE_TTL`: Verification cache TTL (default: `1h`)
- `SENTINEL_DEPENDENCY_CACHE_TTL`: Dependency cache TTL (default: `10m`)
- `SENTINEL_CACHE_MAX_SIZE`: Maximum cache size (default: `10000`)
- `SENTINEL_CACHE_CLEANUP_INTERVAL`: Cache cleanup interval (default: `5m`)

### Retry Configuration
- `SENTINEL_MAX_RETRIES`: Maximum retry attempts (default: `3`)
- `SENTINEL_INITIAL_BACKOFF`: Initial backoff duration (default: `100ms`)
- `SENTINEL_MAX_BACKOFF`: Maximum backoff duration (default: `5s`)
- `SENTINEL_BACKOFF_MULTIPLIER`: Backoff multiplier (default: `2.0`)

## Generating Secrets

### Using the Script (Recommended)
```bash
cd hub
./scripts/generate-secrets.sh
```

This will:
- Generate cryptographically secure random secrets
- Create `.env` file with all required variables
- Set appropriate defaults for production

### Manual Generation
```bash
# Generate DB password
openssl rand -base64 32 | tr -d "=+/" | cut -c1-32

# Generate JWT secret
openssl rand -hex 32

# Generate Admin API key
openssl rand -hex 32
```

## Security Best Practices

1. **Never commit `.env` files to git**
   - `.env` is already in `.gitignore`
   - Use pre-commit hooks to prevent accidental commits

2. **Rotate secrets regularly**
   - Rotate every 90 days
   - Rotate immediately if compromised
   - Admin API key should be rotated more frequently (every 60 days recommended)

3. **Use different secrets for each environment**
   - Development, staging, and production must have different secrets
   - Never reuse production secrets in development

4. **Use secrets management service in production**
   - Docker Swarm: `docker secret create`
   - Kubernetes: `kubectl create secret`
   - AWS: AWS Secrets Manager
   - Azure: Azure Key Vault
   - GCP: Secret Manager

5. **Restrict access to secrets**
   - Principle of least privilege
   - Only grant access to those who need it
   - Audit secret access regularly

6. **Validate production configuration**
   - Application validates production config on startup
   - Fails if insecure defaults detected
   - See `validateProductionConfig()` in `main.go`

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_PASSWORD` | Database password | `aB3$kL9mN2pQ5rS7tU1vW4xY6zA8` |
| `JWT_SECRET` | JWT signing secret | `a1b2c3d4e5f6...` (64 hex chars) |
| `ADMIN_API_KEY` | Admin API key for admin endpoints | `a1b2c3d4e5f6...` (64 hex chars) |
| `CORS_ORIGIN` | Allowed CORS origin | `https://app.example.com` |
| `DATABASE_URL` | Full database connection string | `postgres://user:pass@host:5432/db?sslmode=require` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `ENVIRONMENT` | Environment name | `development` |
| `LOG_LEVEL` | Logging level | `info` |
| `DOCUMENT_STORAGE` | Document storage path | `/data/documents` |
| `OLLAMA_HOST` | Ollama API host | `http://ollama:11434` |
| `HUB_URL` | Hub URL for self-reference | `http://localhost:8080` |

## Production Checklist

Before deploying to production:

- [ ] `.env` file created with secure secrets
- [ ] `CORS_ORIGIN` set to specific domain (not `*`)
- [ ] `JWT_SECRET` changed from default value
- [ ] `ADMIN_API_KEY` set to secure random value
- [ ] `DB_PASSWORD` changed from default value
- [ ] Database connection uses SSL (`sslmode=require`)
- [ ] `.env` file is gitignored
- [ ] Secrets stored in secrets management service
- [ ] Access to secrets restricted appropriately
- [ ] Production validation passes (`ENVIRONMENT=production`)

## Troubleshooting

### Application fails to start in production
**Error**: `PRODUCTION CONFIGURATION ERRORS`

**Solution**: 
1. Check `.env` file exists
2. Verify all required variables are set
3. Ensure no default values are used
4. Check SSL is enabled for database

### Secrets not loading
**Issue**: Environment variables not found

**Solution**:
1. Verify `.env` file exists in `hub/` directory
2. Check `env_file` directive in `docker-compose.yml`
3. Verify file permissions (should be readable)
4. Check for typos in variable names

### CORS errors in production
**Issue**: CORS_ORIGIN is `*` or incorrect

**Solution**:
1. Set `CORS_ORIGIN` to specific domain in `.env`
2. Restart application
3. Verify production validation passes

## Additional Resources

- [Production Checklist](../docs/external/PRODUCTION_CHECKLIST.md)
- [Deployment Guide](../docs/external/HUB_DEPLOYMENT_GUIDE.md)
- [Security Guide](../docs/external/SECURITY_GUIDE.md)

