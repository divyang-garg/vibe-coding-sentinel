# Hub Deployment Guide

This guide provides comprehensive instructions for deploying the Sentinel Hub API server, including database setup, Docker deployment, environment configuration, and production considerations.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Database Setup](#database-setup)
3. [Environment Variables](#environment-variables)
4. [Docker Deployment](#docker-deployment)
5. [Manual Deployment](#manual-deployment)
6. [Production Considerations](#production-considerations)
7. [Monitoring and Health Checks](#monitoring-and-health-checks)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

- **PostgreSQL 12+** - Database server
- **Go 1.21+** - For building from source (optional)
- **Docker & Docker Compose** - For containerized deployment (recommended)

### System Requirements

- **CPU**: 2+ cores recommended
- **Memory**: 4GB+ RAM recommended
- **Storage**: 10GB+ for documents and database
- **Network**: Port 8080 (configurable) accessible

## Database Setup

### PostgreSQL Installation

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### macOS (Homebrew)
```bash
brew install postgresql
brew services start postgresql
```

#### Docker
```bash
docker run --name sentinel-postgres \
  -e POSTGRES_PASSWORD=sentinel \
  -e POSTGRES_USER=sentinel \
  -e POSTGRES_DB=sentinel \
  -p 5432:5432 \
  -v sentinel-data:/var/lib/postgresql/data \
  -d postgres:15
```

### Database Creation

```bash
# Connect to PostgreSQL
sudo -u postgres psql

# Create database and user
CREATE DATABASE sentinel;
CREATE USER sentinel WITH PASSWORD 'your-secure-password';
GRANT ALL PRIVILEGES ON DATABASE sentinel TO sentinel;

# For UUID extension
\c sentinel
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  # For text search

# Exit
\q
```

### Database Connection String

Format: `postgres://username:password@host:port/database?sslmode=mode`

**Examples:**
- Local: `postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable`
- Production: `postgres://sentinel:password@db.example.com:5432/sentinel?sslmode=require`
- Docker: `postgres://sentinel:sentinel@postgres:5432/sentinel?sslmode=disable`

## Environment Variables

The Hub API uses the following environment variables for configuration:

### Required Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable` | `postgres://user:pass@host:5432/db` |
| `ADMIN_API_KEY` | Admin API key for admin endpoints | (required) | `hex-32-chars` |
| `JWT_SECRET` | Secret for JWT tokens | (required) | `your-secret-key` |

### Storage Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DOCUMENT_STORAGE` | Path for document storage | `/data/documents` | `/var/lib/sentinel/documents` |
| `BINARY_STORAGE` | Path for binary storage | `/data/binaries` | `/var/lib/sentinel/binaries` |
| `RULES_STORAGE` | Path for rules storage | `/data/rules` | `/var/lib/sentinel/rules` |

### Server Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `PORT` | HTTP server port | `8080` | `8080` |
| `CORS_ORIGIN` | Allowed CORS origin | `*` | `https://app.example.com` |
| `ENVIRONMENT` | Environment name | `development` | `production` |

### LLM Configuration (Optional)

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `AZURE_AI_KEY` | Azure AI Foundry API key | - | `your-azure-key` |

### Performance & Limits

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `SENTINEL_DB_TIMEOUT` | Database query timeout | `10s` | `30s` |
| `SENTINEL_ANALYSIS_TIMEOUT` | Analysis timeout | `60s` | `120s` |
| `SENTINEL_HTTP_TIMEOUT` | HTTP request timeout | `30s` | `60s` |
| `SENTINEL_CONTEXT_TIMEOUT` | Default context timeout | `30s` | `60s` |
| `SENTINEL_MAX_FILE_SIZE` | Maximum file size (bytes) | `104857600` (100MB) | `209715200` (200MB) |
| `SENTINEL_MAX_STRING_LENGTH` | Maximum string length | `1000000` | `2000000` |
| `SENTINEL_MAX_REQUEST_SIZE` | Maximum HTTP request size | `10485760` (10MB) | `20971520` (20MB) |
| `SENTINEL_RATE_LIMIT_RPS` | Rate limit requests/sec | `100` | `200` |
| `SENTINEL_RATE_LIMIT_BURST` | Rate limit burst size | `200` | `400` |

### Task Management

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `SENTINEL_MAX_TASK_TITLE_LENGTH` | Max task title length | `500` | `1000` |
| `SENTINEL_MAX_TASK_DESCRIPTION_LENGTH` | Max task description length | `5000` | `10000` |
| `SENTINEL_DEFAULT_TASK_LIST_LIMIT` | Default task list limit | `50` | `100` |
| `SENTINEL_MAX_TASK_LIST_LIMIT` | Maximum task list limit | `1000` | `2000` |
| `SENTINEL_DEFAULT_DATE_RANGE_DAYS` | Default date range (days) | `30` | `90` |

### Cache Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `SENTINEL_CACHE_TTL` | Default cache TTL | `5m` | `10m` |
| `SENTINEL_TASK_CACHE_TTL` | Task cache TTL | `5m` | `10m` |
| `SENTINEL_VERIFICATION_CACHE_TTL` | Verification cache TTL | `1h` | `2h` |
| `SENTINEL_DEPENDENCY_CACHE_TTL` | Dependency cache TTL | `10m` | `20m` |
| `SENTINEL_CACHE_MAX_SIZE` | Maximum cache size | `10000` | `20000` |
| `SENTINEL_CACHE_CLEANUP_INTERVAL` | Cache cleanup interval | `5m` | `10m` |

### Retry Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `SENTINEL_MAX_RETRIES` | Maximum retry attempts | `3` | `5` |
| `SENTINEL_INITIAL_BACKOFF` | Initial backoff duration | `100ms` | `200ms` |
| `SENTINEL_MAX_BACKOFF` | Maximum backoff duration | `5s` | `10s` |
| `SENTINEL_BACKOFF_MULTIPLIER` | Backoff multiplier | `2.0` | `1.5` |

### Security Recommendations

1. **Never use default values in production**
   - Change `JWT_SECRET` to a strong random string
   - Set `ADMIN_API_KEY` to a strong random string (required for admin endpoints)
   - Set `CORS_ORIGIN` to specific allowed origins (not `*`)
   - Use SSL for database connections (`sslmode=require`)

2. **Use secrets management**
   - Kubernetes: Use Secrets
   - Docker Compose: Use Docker secrets or environment files
   - Cloud: Use AWS Secrets Manager, Azure Key Vault, etc.

### Admin Authentication

Admin endpoints (`/api/v1/admin/*`) require authentication via admin API key. This includes:
- Creating organizations
- Creating projects
- Uploading binary versions

**Authentication Methods:**

1. **X-Admin-API-Key header** (recommended):
   ```bash
   curl -X POST https://hub.example.com/api/v1/admin/binary/upload \
     -H "X-Admin-API-Key: your-admin-key-here" \
     -F "version=1.2.3" \
     -F "platform=linux-amd64" \
     -F "binary=@sentinel-linux-amd64"
   ```

2. **Authorization Bearer header**:
   ```bash
   curl -X POST https://hub.example.com/api/v1/admin/binary/upload \
     -H "Authorization: Bearer your-admin-key-here" \
     -F "version=1.2.3" \
     -F "platform=linux-amd64" \
     -F "binary=@sentinel-linux-amd64"
   ```

**Generate Admin Key:**

```bash
openssl rand -hex 32
```

**Set in Environment:**

```bash
export ADMIN_API_KEY=$(openssl rand -hex 32)
```

**Security Notes:**
- Admin key must be kept secret and never committed to version control
- Rotate admin key regularly (recommended: every 90 days)
- Use different admin keys for different environments
- Restrict access to admin key storage

### Environment File Example

Create `.env` file:
```bash
# Database
DATABASE_URL=postgres://sentinel:secure-password@localhost:5432/sentinel?sslmode=require

# Server
PORT=8080
ENVIRONMENT=production

# Storage
DOCUMENT_STORAGE=/var/lib/sentinel/documents

# Security
JWT_SECRET=$(openssl rand -hex 32)
ADMIN_API_KEY=$(openssl rand -hex 32)
CORS_ORIGIN=https://app.example.com

# External Services
OLLAMA_HOST=http://ollama:11434
HUB_URL=https://hub.example.com
```

## Docker Deployment

### Dockerfile

Create `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY hub/api/ ./hub/api/
COPY hub/api/go.mod ./hub/api/go.mod
COPY hub/api/go.sum ./hub/api/go.sum

WORKDIR /app/hub/api
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o sentinel-hub main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/hub/api/sentinel-hub .
COPY --from=builder /app/hub/api/migrations ./migrations

EXPOSE 8080
CMD ["./sentinel-hub"]
```

### Docker Compose

Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: sentinel-postgres
    environment:
      POSTGRES_DB: sentinel
      POSTGRES_USER: sentinel
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-sentinel}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U sentinel"]
      interval: 10s
      timeout: 5s
      retries: 5

  hub-api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: sentinel-hub-api
    environment:
      DATABASE_URL: postgres://sentinel:${POSTGRES_PASSWORD:-sentinel}@postgres:5432/sentinel?sslmode=disable
      PORT: 8080
      DOCUMENT_STORAGE: /data/documents
      JWT_SECRET: ${JWT_SECRET:-change-me-in-production}
      CORS_ORIGIN: ${CORS_ORIGIN:-*}
      OLLAMA_HOST: ${OLLAMA_HOST:-http://ollama:11434}
      HUB_URL: ${HUB_URL:-http://localhost:8080}
      ENVIRONMENT: ${ENVIRONMENT:-production}
    volumes:
      - document-storage:/data/documents
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  postgres-data:
  document-storage:
```

### Deployment Steps

1. **Clone repository**
   ```bash
   git clone https://github.com/yourorg/sentinel-hub.git
   cd sentinel-hub
   ```

2. **Create environment file**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start services**
   ```bash
   docker-compose up -d
   ```

4. **Verify deployment**
   ```bash
   # Check logs
   docker-compose logs -f hub-api
   
   # Check health
   curl http://localhost:8080/health
   ```

5. **Create initial organization and project**
   ```bash
   # Using API (requires admin access)
   curl -X POST http://localhost:8080/api/v1/admin/organizations \
     -H "Content-Type: application/json" \
     -d '{"name": "My Organization"}'
   
   curl -X POST http://localhost:8080/api/v1/admin/projects \
     -H "Content-Type: application/json" \
     -d '{
       "org_id": "org-uuid-here",
       "name": "My Project"
     }'
   ```

## Manual Deployment

### Building from Source

1. **Install dependencies**
   ```bash
   cd hub/api
   go mod download
   ```

2. **Build binary**
   ```bash
   go build -o sentinel-hub main.go
   ```

3. **Run migrations**
   ```bash
   # Migrations run automatically on startup
   # Or run manually if needed
   ./sentinel-hub migrate
   ```

4. **Start server**
   ```bash
   export DATABASE_URL="postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable"
   export PORT=8080
   export DOCUMENT_STORAGE="/data/documents"
   ./sentinel-hub
   ```

### Systemd Service

Create `/etc/systemd/system/sentinel-hub.service`:
```ini
[Unit]
Description=Sentinel Hub API
After=network.target postgresql.service

[Service]
Type=simple
User=sentinel
WorkingDirectory=/opt/sentinel-hub
ExecStart=/opt/sentinel-hub/sentinel-hub
Restart=always
RestartSec=10

Environment="DATABASE_URL=postgres://sentinel:password@localhost:5432/sentinel?sslmode=require"
Environment="PORT=8080"
Environment="DOCUMENT_STORAGE=/var/lib/sentinel/documents"
Environment="JWT_SECRET=your-secret-key"
Environment="CORS_ORIGIN=https://app.example.com"
Environment="ENVIRONMENT=production"

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable sentinel-hub
sudo systemctl start sentinel-hub
sudo systemctl status sentinel-hub
```

## MVP Deployment

For MVP (Minimum Viable Product) deployment, the following features are fully functional:

### Working Features
- ✅ Document ingestion and processing
- ✅ Knowledge extraction and management
- ✅ AST analysis and security scanning
- ✅ Test requirement generation and validation
- ✅ Comprehensive feature analysis
- ✅ 15/19 MCP tools functional
- ✅ Intent analysis and pattern learning

### Production-Ready Endpoints
- ✅ `POST /api/v1/validate/code` - Returns actual AST-based violations
- ✅ `POST /api/v1/fixes/apply` - Applies security and code quality fixes
- ✅ All MCP tools - Fully functional with handlers implemented

**Note**: All endpoints are now production-ready with full functionality.

## Production Considerations

### Security Hardening

1. **Database Security**
   - Use SSL/TLS for database connections (`sslmode=require`)
   - Use strong passwords
   - Limit database user permissions
   - Enable PostgreSQL logging

2. **API Security**
   - Set `CORS_ORIGIN` to specific domains (not `*`)
   - Use strong `JWT_SECRET` and `ADMIN_API_KEY`
   - Enable HTTPS/TLS termination (use reverse proxy)
   - Implement rate limiting (already included)
   - Enhanced API key validation (length, format, patterns)
   - Security headers (CSP, HSTS, X-Frame-Options)
   - Content-Type validation for file uploads
   - Regular security updates

3. **Network Security**
   - Use firewall rules to restrict access
   - Use VPN or private networks for database access
   - Implement DDoS protection

### High Availability

1. **Database**
   - Use PostgreSQL replication (master-slave)
   - Implement connection pooling
   - Regular backups

2. **Application**
   - Deploy multiple instances behind load balancer
   - Use health checks for load balancer
   - Implement graceful shutdown

3. **Storage**
   - Use shared storage (NFS, S3, etc.) for documents
   - Implement backup strategy

### Performance Optimization

1. **Database**
   - Monitor query performance
   - Add indexes as needed (already included)
   - Tune PostgreSQL configuration
   - Use connection pooling (already implemented)

2. **Application**
   - Monitor metrics via `/api/v1/metrics/prometheus`
   - Optimize slow endpoints
   - Use caching where appropriate

3. **Storage**
   - Use fast storage for document storage
   - Implement cleanup policies for old documents

## Binary Upload Endpoint

### POST /api/v1/admin/binary/upload

Uploads a new binary version for distribution to Sentinel clients.

**Authentication:** Requires admin API key via `X-Admin-API-Key` header or `Authorization: Bearer <admin-key>`.

**Request Format:** `multipart/form-data`

**Required Fields:**
- `version`: Semver format (e.g., `1.2.3` or `v1.2.3`)
- `platform`: One of `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`, `windows-amd64`
- `binary`: Binary file (max 100MB)

**Optional Fields:**
- `release_notes`: Release notes (max 10KB, sanitized)
- `is_stable`: `true` or `false` (default: `false`)
- `is_latest`: `true` or `false` (default: `false`)

**Validation:**
- Version must match semver format: `^v?\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`
- Platform must be from allowed list
- Release notes are sanitized (control characters removed, length limited)

**Response:** JSON with `success`, `version`, `platform`, and `checksum` fields.

**Error Handling:**
- Missing/invalid admin key → 401 Unauthorized
- Invalid version format → 400 Bad Request
- Invalid platform → 400 Bad Request
- File operation errors → 500 Internal Server Error (with cleanup)
- Database errors → 500 Internal Server Error (logged with context)

## Production-Ready Endpoints

All endpoints are now fully functional and production-ready:

### POST /api/v1/validate/code
- **Status**: ✅ Fully functional
- **Implementation**: Calls `analyzeAST()` and returns actual violations
- **Features**: AST-based code analysis with security and quality checks

### POST /api/v1/fixes/apply
- **Status**: ✅ Fully functional
- **Implementation**: Applies security, style, and quality fixes based on fixType
- **Features**: Automated code improvement with targeted fix application

**Status**: All endpoints are production-ready with no stub implementations remaining.

## API Key Management Endpoints

The Hub API provides endpoints for managing project API keys. These endpoints allow you to generate, view information about, and revoke API keys.

### POST /api/v1/projects/{id}/api-key

Generates a new API key for a project. The old key (if any) is automatically revoked.

**Authentication:** Requires valid API key (admin or project key)

**Request:**
```bash
curl -X POST https://hub.example.com/api/v1/projects/proj_123/api-key \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-admin-key"
```

**Response (200 OK):**
```json
{
  "api_key": "xK9mP2qR7vT4wY8zA1bC3dE5fG6hI0j",
  "api_key_prefix": "xK9mP2qR",
  "message": "API key generated successfully. Save this key - it will not be shown again.",
  "warning": "This is the only time you will see this key. Store it securely."
}
```

**Important:** The full API key is only returned once. Save it immediately!

### GET /api/v1/projects/{id}/api-key

Returns API key information (prefix only, for security). The full key is never returned.

**Authentication:** Requires valid API key

**Request:**
```bash
curl -X GET https://hub.example.com/api/v1/projects/proj_123/api-key \
  -H "X-API-Key: your-admin-key"
```

**Response (200 OK):**
```json
{
  "has_api_key": true,
  "api_key_prefix": "xK9mP2qR",
  "message": "Full API key is never returned for security reasons. Use POST to generate a new key."
}
```

### DELETE /api/v1/projects/{id}/api-key

Revokes a project's API key. After revocation, all requests using that key will fail.

**Authentication:** Requires valid API key (admin or project key)

**Request:**
```bash
curl -X DELETE https://hub.example.com/api/v1/projects/proj_123/api-key \
  -H "X-API-Key: your-admin-key"
```

**Response (200 OK):**
```json
{
  "message": "API key revoked successfully"
}
```

### API Key Security Features

1. **Secure Generation:** API keys are generated using `crypto/rand` for cryptographic security
2. **Hashed Storage:** Keys are stored as SHA-256 hashes in the database, never in plaintext
3. **One-Time Display:** Full keys are only shown once when generated
4. **Prefix Display:** Only the first 8 characters are shown in subsequent requests
5. **Audit Logging:** All API key operations are logged for security auditing

### Initial Project Setup

When deploying the Hub, you'll need to create an initial project and API key:

```bash
# 1. Create a project (auto-generates API key)
PROJECT_RESPONSE=$(curl -s -X POST https://hub.example.com/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $ADMIN_API_KEY" \
  -d '{"name": "Default Project"}')

# 2. Extract and save the API key
PROJECT_ID=$(echo $PROJECT_RESPONSE | jq -r '.id')
API_KEY=$(echo $PROJECT_RESPONSE | jq -r '.api_key')

# 3. Store in environment or secret management
export SENTINEL_API_KEY="$API_KEY"

# 4. Verify the key works
curl -X GET https://hub.example.com/api/v1/projects/$PROJECT_ID \
  -H "X-API-Key: $API_KEY"
```

For detailed API key management documentation, see:
- `docs/API_KEY_MANAGEMENT_GUIDE.md` - User guide for API key management
- `docs/API_KEY_IMPLEMENTATION_FLOW.md` - Technical implementation details

### Backup Strategy

1. **Database Backups**
   ```bash
   # Daily backup script
   pg_dump -U sentinel -d sentinel > backup-$(date +%Y%m%d).sql
   ```

2. **Document Storage Backups**
   ```bash
   # Backup document storage
   tar -czf documents-$(date +%Y%m%d).tar.gz /data/documents
   ```

3. **Automated Backups**
   - Use cron jobs or scheduled tasks
   - Store backups off-site
   - Test restore procedures regularly

## Monitoring and Health Checks

### Health Check Endpoints

- **GET /health** - Basic health check
- **GET /health/db** - Database connectivity check
- **GET /health/ready** - Readiness check (database + storage)

### Metrics Endpoint

- **GET /api/v1/metrics/prometheus** - Prometheus metrics (requires API key)

See [MONITORING_GUIDE.md](./MONITORING_GUIDE.md) for detailed monitoring setup.

### Logging

The Hub API logs to stdout/stderr. In production:
- Use log aggregation (ELK, Loki, etc.)
- Configure log rotation
- Monitor error logs

## Troubleshooting

### Common Issues

#### Database Connection Failed

**Symptoms:**
- Health check `/health/db` returns unhealthy
- Logs show connection errors

**Solutions:**
1. Verify database is running:
   ```bash
   sudo systemctl status postgresql
   # or
   docker ps | grep postgres
   ```

2. Check connection string:
   ```bash
   echo $DATABASE_URL
   ```

3. Test connection manually:
   ```bash
   psql $DATABASE_URL -c "SELECT 1"
   ```

4. Check firewall rules:
   ```bash
   sudo ufw status
   ```

#### Storage Directory Issues

**Symptoms:**
- `/health/ready` returns not_ready
- Document upload fails

**Solutions:**
1. Verify directory exists:
   ```bash
   ls -la $DOCUMENT_STORAGE
   ```

2. Check permissions:
   ```bash
   sudo chown -R sentinel:sentinel $DOCUMENT_STORAGE
   sudo chmod -R 755 $DOCUMENT_STORAGE
   ```

3. Check disk space:
   ```bash
   df -h $DOCUMENT_STORAGE
   ```

#### High Memory Usage

**Symptoms:**
- Application crashes
- Slow performance

**Solutions:**
1. Check database connection pool:
   ```bash
   curl http://localhost:8080/health/db
   ```

2. Review connection pool settings in code
3. Monitor metrics for memory leaks
4. Increase container/system memory if needed

#### CORS Errors

**Symptoms:**
- Browser shows CORS errors
- API requests fail from frontend

**Solutions:**
1. Verify `CORS_ORIGIN` environment variable:
   ```bash
   echo $CORS_ORIGIN
   ```

2. Set to specific origin (not `*`):
   ```bash
   export CORS_ORIGIN=https://app.example.com
   ```

3. Restart application

### Debug Mode

Enable debug logging:
```bash
export ENVIRONMENT=development
# Application will show more detailed logs
```

### Getting Help

1. Check logs:
   ```bash
   docker-compose logs hub-api
   # or
   journalctl -u sentinel-hub -f
   ```

2. Check health endpoints:
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/health/db
   curl http://localhost:8080/health/ready
   ```

3. Review metrics:
   ```bash
   curl -H "Authorization: Bearer $API_KEY" \
     http://localhost:8080/api/v1/metrics/prometheus
   ```

## Additional Resources

- [Monitoring Guide](./MONITORING_GUIDE.md)
- [Security Guide](./SECURITY_GUIDE.md)
- [API Reference](./HUB_API_REFERENCE.md) (when available)
- [Architecture Documentation](./ARCHITECTURE.md)


