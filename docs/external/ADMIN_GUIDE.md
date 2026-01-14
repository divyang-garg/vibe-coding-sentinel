# Sentinel Hub Administrator Guide

**Version:** 1.0.0
**Last Updated:** January 8, 2026

## Overview

This guide provides comprehensive instructions for administrators managing the Sentinel Hub, including installation, configuration, monitoring, security, and maintenance procedures.

## Table of Contents

1. [Installation & Setup](#installation--setup)
2. [Configuration Management](#configuration-management)
3. [User & Organization Management](#user--organization-management)
4. [Security Administration](#security-administration)
5. [Monitoring & Logging](#monitoring--logging)
6. [Backup & Recovery](#backup--recovery)
7. [Performance Tuning](#performance-tuning)
8. [Troubleshooting](#troubleshooting)
9. [Upgrade Procedures](#upgrade-procedures)

---

## Installation & Setup

### Prerequisites

**System Requirements:**
- **OS:** Linux, macOS, or Windows Server
- **CPU:** 2+ cores (4+ recommended for production)
- **RAM:** 4GB minimum (8GB+ recommended)
- **Storage:** 20GB+ available disk space
- **Network:** Stable internet connection

**Software Dependencies:**
- **Docker & Docker Compose:** For containerized deployment
- **PostgreSQL 15+:** Database server
- **Git:** For version control and hooks
- **curl/wget:** For health checks and API testing

### Quick Start Installation

```bash
# 1. Clone the repository
git clone https://github.com/your-org/sentinel-hub.git
cd sentinel-hub

# 2. Configure environment
cp .env.example .env
# Edit .env with your settings

# 3. Start services
docker-compose up -d

# 4. Verify installation
curl http://localhost:8080/health
```

### Production Deployment

```bash
# Use production overrides
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Enable SSL/TLS termination
# Configure reverse proxy (nginx/caddy/traefik)
# Set up monitoring and alerting
```

---

## Configuration Management

### Environment Variables

**Core Configuration:**
```bash
# Server
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info

# Database
DATABASE_URL=postgres://user:password@db:5432/sentinel?sslmode=require

# Security
JWT_SECRET=your-256-bit-secret-here
ADMIN_API_KEY=your-admin-key-here
CORS_ORIGIN=https://yourdomain.com

# Storage
DOCUMENT_STORAGE=/data/documents
BINARY_STORAGE=/data/binaries
RULES_STORAGE=/data/rules

# External Services
OLLAMA_HOST=http://ollama:11434
HUB_URL=https://yourdomain.com
```

### Admin API Key Management

**Generating Secure API Keys:**
```bash
# Use openssl for cryptographically secure keys
openssl rand -hex 32

# Or use /dev/urandom
head -c 32 /dev/urandom | xxd -p -c 32
```

**Key Rotation Procedure:**
```bash
# 1. Generate new admin key
NEW_KEY=$(openssl rand -hex 32)

# 2. Update environment
echo "ADMIN_API_KEY=$NEW_KEY" >> .env

# 3. Restart services
docker-compose restart api

# 4. Update client configurations
# 5. Remove old key from documentation
```

---

## User & Organization Management

### Creating Organizations

```bash
# Create organization via API
curl -X POST http://localhost:8080/api/v1/admin/organizations \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "description": "Manufacturing company",
    "contact_email": "admin@acme.com"
  }'
```

### Managing Projects

**Project Creation:**
```bash
curl -X POST http://localhost:8080/api/v1/admin/projects \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "org-uuid",
    "name": "E-commerce Platform",
    "description": "Customer-facing web application"
  }'
```

**Project API Key Management:**
```bash
# Get project details including API key
curl http://localhost:8080/api/v1/admin/projects/$PROJECT_ID \
  -H "X-Admin-API-Key: $ADMIN_KEY"

# Rotate project API key
curl -X POST http://localhost:8080/api/v1/admin/projects/$PROJECT_ID/rotate-key \
  -H "X-Admin-API-Key: $ADMIN_KEY"
```

### Organization Policies

**Setting Organization Policies:**
```bash
curl -X PUT http://localhost:8080/api/v1/admin/organizations/$ORG_ID/policy \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "max_projects": 50,
    "max_storage_gb": 100,
    "allowed_file_types": ["pdf", "docx", "xlsx"],
    "require_approval": true
  }'
```

---

## Security Administration

### Access Control

**Admin Role Management:**
- Admin API keys provide full system access
- Project API keys are scoped to specific projects
- Rate limiting applies to all endpoints
- IP whitelisting can be implemented at reverse proxy level

### Security Monitoring

**Monitoring Security Events:**
```bash
# Check recent security events
curl http://localhost:8080/api/v1/admin/security/events \
  -H "X-Admin-API-Key: $ADMIN_KEY"

# Monitor failed authentication attempts
curl http://localhost:8080/api/v1/admin/security/failed-auth \
  -H "X-Admin-API-Key: $ADMIN_KEY"
```

### Binary Upload Security

**Binary Validation:**
```bash
# Upload binary with security validation
curl -X POST http://localhost:8080/api/v1/admin/binary/upload \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -F "file=@sentinel-binary" \
  -F "version=1.2.3" \
  -F "platform=linux-amd64"
```

**Binary Signing (Future):**
```bash
# Sign binary before upload
gpg --detach-sign sentinel-binary
curl -X POST http://localhost:8080/api/v1/admin/binary/upload \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -F "file=@sentinel-binary" \
  -F "signature=@sentinel-binary.sig"
```

---

## Monitoring & Logging

### Health Checks

**System Health:**
```bash
# Overall system health
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed

# Database connectivity
curl http://localhost:8080/health/database
```

### Metrics Collection

**Available Metrics:**
```bash
# System metrics
curl http://localhost:8080/metrics

# Request statistics
curl http://localhost:8080/api/v1/admin/metrics/requests

# Storage usage
curl http://localhost:8080/api/v1/admin/metrics/storage

# LLM usage tracking
curl http://localhost:8080/api/v1/admin/metrics/llm-usage
```

### Log Management

**Log Configuration:**
```bash
# Environment variables
LOG_LEVEL=info  # debug, info, warn, error
LOG_FORMAT=json # text, json

# Structured logging output
{
  "timestamp": "2026-01-08T10:30:45Z",
  "level": "INFO",
  "request_id": "req-12345",
  "message": "Task created successfully",
  "user_id": "user-789",
  "project_id": "proj-456"
}
```

**Log Rotation:**
```bash
# Docker logging
docker-compose logs -f api

# Log aggregation (external)
# Configure ELK stack, Splunk, or similar
```

---

## Backup & Recovery

### Database Backup

**Automated Backup:**
```bash
# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="sentinel_backup_$TIMESTAMP.sql"

docker-compose exec -T db pg_dump -U sentinel sentinel > "$BACKUP_FILE"
gzip "$BACKUP_FILE"
echo "Backup created: $BACKUP_FILE.gz"
EOF

# Schedule with cron
0 2 * * * /path/to/backup.sh
```

**Manual Backup:**
```bash
# Stop services for consistent backup
docker-compose stop api

# Create backup
docker-compose exec db pg_dump -U sentinel sentinel > backup.sql

# Restart services
docker-compose start api
```

### Document Storage Backup

**File System Backup:**
```bash
# Backup document storage
tar -czf documents_backup_$(date +%Y%m%d).tar.gz /data/documents/

# Backup binary storage
tar -czf binaries_backup_$(date +%Y%m%d).tar.gz /data/binaries/
```

### Recovery Procedures

**Database Recovery:**
```bash
# Stop services
docker-compose stop

# Drop and recreate database
docker-compose exec db psql -U sentinel -c "DROP DATABASE sentinel;"
docker-compose exec db psql -U sentinel -c "CREATE DATABASE sentinel;"

# Restore from backup
gunzip backup.sql.gz
docker-compose exec -T db psql -U sentinel sentinel < backup.sql

# Restart services
docker-compose start
```

**Full System Recovery:**
```bash
# 1. Restore database
# 2. Restore document files
# 3. Restore binary files
# 4. Verify system health
# 5. Update DNS/load balancer if needed
```

---

## Performance Tuning

### Database Optimization

**Connection Pooling:**
```bash
# PostgreSQL configuration
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
```

**Query Optimization:**
```bash
# Analyze query performance
docker-compose exec db psql -U sentinel sentinel -c "EXPLAIN ANALYZE SELECT * FROM documents LIMIT 10;"

# Create indexes for common queries
CREATE INDEX idx_documents_project_status ON documents(project_id, status);
CREATE INDEX idx_knowledge_items_type_status ON knowledge_items(type, status);
```

### Memory Management

**Go Runtime Tuning:**
```bash
# Environment variables for Go
GOGC=100        # Garbage collection target percentage
GOMAXPROCS=4    # Maximum CPU cores to use
```

**Container Limits:**
```yaml
# docker-compose.prod.yml
services:
  api:
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '2.0'
        reservations:
          memory: 512M
          cpus: '1.0'
```

### Caching Strategies

**Response Caching:**
```bash
# Enable response caching for MCP tools
CACHE_TTL_MINUTES=15
CACHE_SIZE_MB=100
```

**Database Query Caching:**
```bash
# Cache frequently accessed data
# Implement Redis for distributed caching
# Cache LLM configurations and patterns
```

---

## Troubleshooting

### Common Issues

**Database Connection Issues:**
```bash
# Check database connectivity
docker-compose exec db psql -U sentinel -c "SELECT 1;"

# Check database logs
docker-compose logs db

# Verify connection string
echo $DATABASE_URL
```

**High Memory Usage:**
```bash
# Check memory usage
docker stats

# Check Go garbage collection
curl http://localhost:8080/debug/pprof/heap

# Restart service if needed
docker-compose restart api
```

**Slow Response Times:**
```bash
# Check system load
uptime
df -h

# Check database performance
docker-compose exec db psql -U sentinel -c "SELECT * FROM pg_stat_activity;"

# Enable query logging
# Add to postgresql.conf: log_statement = 'all'
```

**API Key Issues:**
```bash
# Validate API key format
echo "your-api-key" | grep -E '^[a-zA-Z0-9\-_\.]+$'

# Check API key in database
docker-compose exec db psql -U sentinel -c "SELECT id, name FROM projects WHERE api_key = 'your-api-key';"
```

### Diagnostic Commands

**System Diagnostics:**
```bash
# Full system status
curl http://localhost:8080/api/v1/admin/diagnostics

# Database diagnostics
curl http://localhost:8080/api/v1/admin/diagnostics/database

# Cache diagnostics
curl http://localhost:8080/api/v1/admin/diagnostics/cache
```

**Log Analysis:**
```bash
# Search for errors in logs
docker-compose logs api | grep ERROR

# Check recent requests
docker-compose logs api --tail 100

# Analyze response times
docker-compose logs api | grep "response_time"
```

---

## Upgrade Procedures

### Minor Version Upgrades

```bash
# 1. Backup data
./backup.sh

# 2. Pull new images
docker-compose pull

# 3. Update environment if needed
# Edit .env file for new configuration options

# 4. Run database migrations
docker-compose run --rm api migrate up

# 5. Restart services
docker-compose up -d

# 6. Verify upgrade
curl http://localhost:8080/health
```

### Major Version Upgrades

```bash
# 1. Review release notes
# Check for breaking changes

# 2. Create full backup
./full-backup.sh

# 3. Test upgrade in staging environment
# Clone production environment
# Run upgrade procedures
# Validate functionality

# 4. Schedule maintenance window
# Notify users of downtime

# 5. Perform upgrade
docker-compose down
docker-compose pull
# Update configuration
docker-compose up -d

# 6. Run post-upgrade checks
# Verify all endpoints work
# Check data integrity
# Monitor error rates
```

### Rollback Procedures

```bash
# Quick rollback (last 24 hours)
docker-compose down
docker-compose run --rm api migrate down 1
docker-compose up -d

# Full rollback (from backup)
docker-compose down
./restore-from-backup.sh yesterday-backup.sql.gz
docker-compose up -d
```

---

## Emergency Procedures

### Service Outage Response

**Immediate Actions:**
1. Check service status: `docker-compose ps`
2. Check logs: `docker-compose logs --tail 50 api`
3. Restart failed services: `docker-compose restart api`
4. Check database connectivity
5. Notify team if outage > 5 minutes

### Data Loss Recovery

**Critical Data Loss:**
1. Stop all services immediately
2. Assess scope of data loss
3. Restore from most recent backup
4. Verify data integrity
5. Re-enable services gradually
6. Notify affected users

### Security Incident Response

**Suspected Breach:**
1. Isolate affected systems
2. Preserve logs and evidence
3. Rotate all API keys and secrets
4. Notify security team
5. Conduct forensic analysis
6. Implement remediation measures

---

## Monitoring Dashboard

### Key Metrics to Monitor

**System Health:**
- Service uptime (target: 99.9%)
- Response time percentiles (P50, P95, P99)
- Error rates by endpoint
- Database connection pool usage

**Business Metrics:**
- Active projects and organizations
- Document processing volume
- API request volume by project
- Storage utilization

**Security Metrics:**
- Failed authentication attempts
- Rate limit hits
- Suspicious request patterns
- API key usage patterns

### Alert Configuration

**Critical Alerts:**
- Service down (> 5 minutes)
- Database unavailable (> 1 minute)
- High error rate (> 5%)
- Storage > 90% capacity

**Warning Alerts:**
- Response time > 2 seconds (P95)
- Memory usage > 80%
- Disk space > 75%
- Failed requests > 1%

---

## Best Practices

### Operational Excellence

1. **Regular Backups:** Daily automated backups with offsite storage
2. **Monitoring:** 24/7 monitoring with alerting for critical issues
3. **Log Retention:** 90-day log retention with automated archiving
4. **Security Updates:** Monthly security updates and patches
5. **Performance Reviews:** Quarterly performance optimization reviews

### Security Best Practices

1. **Principle of Least Privilege:** Minimal required permissions
2. **Regular Key Rotation:** API keys rotated every 90 days
3. **Network Security:** All services behind firewall/reverse proxy
4. **Access Logging:** All administrative actions logged
5. **Incident Response:** Documented procedures for security incidents

### Scalability Planning

1. **Resource Monitoring:** Track usage patterns for capacity planning
2. **Load Testing:** Regular load testing before major releases
3. **Horizontal Scaling:** Plan for multiple Hub instances
4. **Database Sharding:** Consider sharding for high-volume deployments
5. **CDN Integration:** Use CDN for static binary distribution

This administrator guide provides comprehensive coverage of Sentinel Hub management. Regular review and updates to procedures are recommended as the system evolves.



