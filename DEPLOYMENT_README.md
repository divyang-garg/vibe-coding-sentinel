# 🚀 Sentinel Hub API - Deployment Guide

## Overview

This guide covers the deployment of the Sentinel Hub API, a quality control gate for vibe coding practices. The deployment follows CODING_STANDARDS.md and includes development, staging, and production configurations.

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Nginx LB      │    │  Sentinel API   │    │   PostgreSQL    │
│   (Port 80/443) │────│   (Port 8080)   │────│   (Port 5432)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │     Redis       │
                       │   (Port 6379)   │
                       └─────────────────┘
```

## 📋 Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- Go 1.21+ (for local development)
- PostgreSQL 15+ (if not using Docker)
- 4GB RAM minimum
- 10GB disk space

## 🚀 Quick Start (Development)

### 1. Clone and Setup

```bash
git clone <repository-url>
cd sentinel-hub-api

# Copy development configuration
cp env/development.env .env
```

### 2. Start Services

```bash
# Start all services
docker-compose up -d

# Or for development with live reload
docker-compose -f docker-compose.yml up --build
```

### 3. Verify Deployment

```bash
# Check service health
curl http://localhost:8080/health

# View logs
docker-compose logs -f api

# Access pgAdmin (development only)
open http://localhost:5050
# Email: admin@sentinel.com
# Password: admin
```

### 4. Test API

```bash
# Create a test user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User","password":"password123"}'

# Get health check
curl http://localhost:8080/health
```

## 🏭 Production Deployment

### 1. Environment Setup

```bash
# Create production environment file
cp env/production.env .env.production

# Edit with your production values
vim .env.production
```

### Required Production Environment Variables:

```bash
# Database
POSTGRES_DB=sentinel
POSTGRES_USER=sentinel
POSTGRES_PASSWORD=<secure-password>

# Security (generate strong secrets)
JWT_SECRET=<64-character-random-string>
BCRYPT_COST=12

# CORS (restrict to your domains)
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# LLM (if using external services)
AZURE_AI_ENDPOINT=https://your-resource.openai.azure.com
AZURE_AI_KEY=<your-azure-key>
OLLAMA_HOST=<your-ollama-host>
```

### 2. SSL Certificate Setup

```bash
# Create SSL directory
mkdir -p nginx/ssl

# Place your certificates
# nginx/ssl/cert.pem
# nginx/ssl/key.pem
```

### 3. Deploy to Production

```bash
# Run deployment script
./scripts/deploy.sh production

# Or manually with docker-compose
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d
```

### 4. Verify Production Deployment

```bash
# Check all services are running
docker-compose -f docker-compose.prod.yml ps

# Verify health endpoint
curl https://yourdomain.com/health

# Check logs
docker-compose -f docker-compose.prod.yml logs -f api
```

## 📊 Monitoring & Observability

### Health Checks

The API includes comprehensive health checks:

```bash
# Application health
GET /health

# Database health
GET /health/db

# Full health check
GET /health/ready
```

### Metrics (Future Enhancement)

```json
{
  "uptime": "2h30m45s",
  "requests_total": 1250,
  "requests_per_second": 12.5,
  "error_rate": 0.02,
  "database_connections": 8,
  "memory_usage": "256MB"
}
```

### Logging

Structured logging with configurable levels:

```bash
# Development: debug level with colors
LOG_LEVEL=debug

# Production: info level, structured JSON
LOG_LEVEL=info
```

## 🔧 Configuration Management

### Environment Variables

| Variable | Development | Production | Description |
|----------|-------------|------------|-------------|
| `HOST` | `0.0.0.0` | `0.0.0.0` | Server bind address |
| `PORT` | `8080` | `8080` | Server port |
| `DATABASE_URL` | `postgres://...` | `postgres://...` | PostgreSQL connection |
| `JWT_SECRET` | `dev-secret` | `<secure>` | JWT signing secret |
| `BCRYPT_COST` | `8` | `12` | Password hashing cost |
| `RATE_LIMIT_REQUESTS` | `1000` | `100` | Rate limit per window |
| `CORS_ALLOWED_ORIGINS` | `*` | `yourdomain.com` | Allowed CORS origins |

### Database Configuration

The application uses PostgreSQL with the following schema:

- `users` - User accounts and authentication
- `tasks` - Development tasks and tracking
- `llm_configurations` - LLM service settings
- `audit_logs` - Compliance and security audit trail

Database migrations are handled automatically via SQL scripts.

## 🚦 Scaling & Performance

### Horizontal Scaling

```yaml
# docker-compose.prod.yml
services:
  api:
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
```

### Vertical Scaling

```yaml
# Resource limits
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 1G
    reservations:
      cpus: '1.0'
      memory: 512M
```

### Database Scaling

```yaml
# Connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

## 🔒 Security Considerations

### Production Security Checklist

- [ ] Strong JWT secrets (64+ characters)
- [ ] SSL/TLS certificates configured
- [ ] CORS restricted to allowed domains
- [ ] Database credentials in environment variables
- [ ] Rate limiting configured appropriately
- [ ] Audit logging enabled
- [ ] Regular security updates
- [ ] Network segmentation implemented

### Security Headers

The API automatically includes security headers:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000
```

## 🧪 Testing Deployment

### Automated Testing

```bash
# Run all tests
go test ./...

# Run integration tests
go test ./tests/integration/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Load Testing

```bash
# Install hey (load testing tool)
go install github.com/rakyll/hey@latest

# Test API endpoints
hey -n 1000 -c 10 http://localhost:8080/health
hey -n 500 -c 5 -m POST -H "Content-Type: application/json" \
  -d '{"email":"loadtest@example.com","name":"Load Test","password":"test123"}' \
  http://localhost:8080/api/v1/users
```

### Performance Benchmarks

```bash
# Database performance
go test -bench=. ./internal/repository/

# API performance
go test -bench=. ./internal/api/handlers/
```

## 📚 Troubleshooting

### Common Issues

#### 1. Database Connection Failed

```bash
# Check database logs
docker-compose logs postgres

# Verify connection string
docker-compose exec postgres psql -U sentinel -d sentinel -c "SELECT 1;"

# Reset database
docker-compose down -v
docker-compose up -d postgres
```

#### 2. API Not Starting

```bash
# Check API logs
docker-compose logs api

# Verify environment variables
docker-compose exec api env | grep -E "(DATABASE|JWT)"

# Check dependencies
docker-compose ps
```

#### 3. High Memory Usage

```bash
# Check container resources
docker stats

# Adjust memory limits in docker-compose.yml
deploy:
  resources:
    limits:
      memory: 512M
```

#### 4. Rate Limiting Issues

```bash
# Check rate limiter configuration
docker-compose exec api env | grep RATE_LIMIT

# Adjust limits in environment
RATE_LIMIT_REQUESTS=200
RATE_LIMIT_WINDOW=30m
```

### Logs and Debugging

```bash
# View all logs
docker-compose logs

# Follow logs in real-time
docker-compose logs -f

# View specific service logs
docker-compose logs api

# Export logs for analysis
docker-compose logs api > api_logs.txt
```

## 🔄 Backup & Recovery

### Database Backup

```bash
# Create backup
docker-compose exec postgres pg_dump -U sentinel sentinel > backup.sql

# Restore backup
docker-compose exec -T postgres psql -U sentinel sentinel < backup.sql
```

### Configuration Backup

```bash
# Backup environment files
tar -czf env_backup.tar.gz env/
cp .env.production .env.production.backup
```

### Rollback Deployment

```bash
# Rollback to previous version
./scripts/deploy.sh production --rollback

# Or manually
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

## 📞 Support & Maintenance

### Regular Maintenance Tasks

- [ ] Monitor resource usage
- [ ] Update Docker images
- [ ] Rotate JWT secrets
- [ ] Review audit logs
- [ ] Update SSL certificates
- [ ] Run security scans

### Health Monitoring

```bash
# Quick health check script
#!/bin/bash
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
if [ $STATUS -eq 200 ]; then
    echo "✅ API is healthy"
else
    echo "❌ API is unhealthy (HTTP $STATUS)"
    exit 1
fi
```

---

## 🎯 Success Criteria

✅ **Deployment successful** when:
- All services are running and healthy
- API responds to health checks
- Database connections are established
- Authentication and authorization work
- Rate limiting is functional
- SSL certificates are valid (production)
- Monitoring is configured
- Logs are being generated

The Sentinel Hub API is now deployed and ready to serve as a quality control gate for vibe coding practices! 🎉