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
| `PORT` | HTTP server port | `8080` | `8080` |
| `DOCUMENT_STORAGE` | Path for document storage | `/data/documents` | `/var/lib/sentinel/documents` |

### Optional Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `JWT_SECRET` | Secret for JWT tokens | `change-me-in-production` | `your-secret-key` |
| `ADMIN_API_KEY` | Admin API key for admin endpoints | (required) | `hex-32-chars` |
| `OLLAMA_HOST` | Ollama API host for LLM | `http://localhost:11434` | `http://ollama:11434` |
| `CORS_ORIGIN` | Allowed CORS origin | `*` | `https://app.example.com` |
| `HUB_URL` | Hub URL for self-reference | `http://localhost:8080` | `https://hub.example.com` |
| `ENVIRONMENT` | Environment name | `development` | `production` |
| `SENTINEL_DB_TIMEOUT` | Database query timeout | `10s` | `30s` |
| `SENTINEL_ANALYSIS_TIMEOUT` | Analysis timeout | `60s` | `120s` |

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
- ✅ 15/18 MCP tools functional
- ✅ Intent analysis and pattern learning

### Stub Implementations (MVP Limitations)
- ⚠️ `POST /api/v1/validate/code` - Returns empty violations (stub)
- ⚠️ `POST /api/v1/fixes/apply` - Returns original code (stub)
- ⚠️ `sentinel_analyze_intent` MCP tool - Handler missing (Hub endpoint exists)

**Note**: These stubs don't prevent MVP deployment but limit functionality. Plan to fix before full production deployment.

## Production Considerations

### Security Hardening

1. **Database Security**
   - Use SSL/TLS for database connections (`sslmode=require`)
   - Use strong passwords
   - Limit database user permissions
   - Enable PostgreSQL logging

2. **API Security**
   - Set `CORS_ORIGIN` to specific domains (not `*`)
   - Use strong `JWT_SECRET`
   - Enable HTTPS/TLS termination (use reverse proxy)
   - Implement rate limiting (already included)
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

## Stub Implementations

The following endpoints are implemented but have stub behavior:

### POST /api/v1/validate/code
- **Status**: Stub implementation
- **Location**: `hub/api/main.go:1408-1455`
- **Current**: Always returns empty violations
- **Expected**: Should call `analyzeAST()` and return actual violations
- **Impact**: Code validation always reports as valid

### POST /api/v1/fixes/apply
- **Status**: Stub implementation
- **Location**: `hub/api/main.go:1546-1606`
- **Current**: Returns original code unchanged
- **Expected**: Should apply fixes based on fixType
- **Impact**: Fix tool doesn't actually fix code

**Recommendation**: Fix stubs before full production deployment (P0 priority).

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


