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

## Required Secrets

- `DB_PASSWORD`: Database password (32+ characters recommended)
- `JWT_SECRET`: JWT signing secret (64+ characters recommended)
- `ADMIN_API_KEY`: Admin API key for admin endpoints (64 hex characters recommended)
- `CORS_ORIGIN`: Allowed CORS origin (must be specific domain, not `*`)
- `AZURE_AI_KEY`: Azure AI Foundry API key (if using)

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

