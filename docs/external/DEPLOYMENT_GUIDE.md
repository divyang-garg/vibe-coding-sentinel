# Deployment Guide

## Overview

This guide covers deployment strategies for the Sentinel Vibe Coding Platform with a focus on simplicity and minimal manual intervention.

## Components

| Component | Description | Deployment |
|-----------|-------------|------------|
| **Agent** | Go binary on developer machines | Self-contained binary |
| **Hub** | Central server for metrics/dashboard | Docker Compose |

---

## Part 1: Agent Deployment

### Strategy: Self-Contained Binary Distribution

No package managers, no dependencies, no installation process.

### Build Process (CI/CD)

GitHub Actions builds binaries on every release:

```
sentinel-darwin-amd64       (macOS Intel)
sentinel-darwin-arm64       (macOS Apple Silicon)
sentinel-linux-amd64        (Linux x64)
sentinel-linux-arm64        (Linux ARM)
sentinel-windows-amd64.exe  (Windows)
```

### GitHub Actions Workflow

```yaml
# .github/workflows/release.yml
name: Release Agent

on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: darwin
            goarch: amd64
          - goos: darwin
            goarch: arm64
          - goos: linux
            goarch: amd64
          - goos: linux
            goarch: arm64
          - goos: windows
            goarch: amd64

    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          EXT=""
          if [ "$GOOS" = "windows" ]; then EXT=".exe"; fi
          OUTPUT="sentinel-${{ matrix.goos }}-${{ matrix.goarch }}${EXT}"
          go build -ldflags="-s -w -X main.Version=${{ github.ref_name }}" \
            -o "$OUTPUT" ./cmd/sentinel
          chmod +x "$OUTPUT"
      
      - name: Upload to Release
        uses: softprops/action-gh-release@v1
        with:
          files: sentinel-*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Distribution Options

#### Option A: Internal URL (Recommended for Organizations)

Host the install script and binaries on internal infrastructure.

**Install Script** (`install.sh`):

```bash
#!/bin/bash
# Sentinel Agent Installer
# Usage: curl -fsSL https://internal.company.com/sentinel/install.sh | sh

set -e

# Configuration
VERSION="${SENTINEL_VERSION:-latest}"
BASE_URL="${SENTINEL_URL:-https://internal.company.com/sentinel/releases}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect Architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Binary name
BINARY="sentinel-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY="${BINARY}.exe"
fi

echo "📦 Installing Sentinel for ${OS}/${ARCH}..."

# Get version if latest
if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -fsSL "${BASE_URL}/latest" 2>/dev/null || echo "latest")
fi

# Download
DOWNLOAD_URL="${BASE_URL}/${VERSION}/${BINARY}"
echo "   Downloading from: $DOWNLOAD_URL"

if ! curl -fsSL "$DOWNLOAD_URL" -o /tmp/sentinel; then
    echo "❌ Download failed"
    exit 1
fi

chmod +x /tmp/sentinel

# Install location
if [ -d ".git" ]; then
    # Project-local installation
    mv /tmp/sentinel ./sentinel
    INSTALL_PATH="./sentinel"
    echo "✅ Installed to ./sentinel (project-local)"
else
    # Global installation
    INSTALL_DIR="/usr/local/bin"
    if [ "$OS" = "darwin" ] || [ "$OS" = "linux" ]; then
        if [ -w "$INSTALL_DIR" ]; then
            mv /tmp/sentinel "$INSTALL_DIR/sentinel"
        else
            sudo mv /tmp/sentinel "$INSTALL_DIR/sentinel"
        fi
        INSTALL_PATH="$INSTALL_DIR/sentinel"
    else
        # Windows - install to user directory
        INSTALL_DIR="$HOME/bin"
        mkdir -p "$INSTALL_DIR"
        mv /tmp/sentinel "$INSTALL_DIR/sentinel.exe"
        INSTALL_PATH="$INSTALL_DIR/sentinel.exe"
    fi
    echo "✅ Installed to $INSTALL_PATH (global)"
fi

# Configure hub endpoint if provided
if [ -n "$SENTINEL_HUB_URL" ]; then
    "$INSTALL_PATH" config set telemetry.endpoint "$SENTINEL_HUB_URL"
    "$INSTALL_PATH" config set telemetry.enabled true
    echo "✅ Configured hub endpoint: $SENTINEL_HUB_URL"
fi

# Verify installation
echo ""
echo "🎉 Installation complete!"
echo ""
"$INSTALL_PATH" --version
echo ""
echo "Next steps:"
echo "  1. cd your-project"
echo "  2. sentinel init"
echo ""
```

**Developer usage**:
```bash
# One command to install
curl -fsSL https://internal.company.com/sentinel/install.sh | sh

# With custom hub URL
SENTINEL_HUB_URL=https://sentinel.company.com/api \
  curl -fsSL https://internal.company.com/sentinel/install.sh | sh
```

#### Option B: GitHub Releases

```bash
# Direct download from GitHub
curl -fsSL https://github.com/yourorg/sentinel/releases/latest/download/sentinel-darwin-arm64 -o sentinel
chmod +x sentinel
```

#### Option C: Package Repository (npm/homebrew)

For wider distribution:

```bash
# Homebrew (macOS)
brew install yourorg/tap/sentinel

# npm (cross-platform)
npm install -g @yourorg/sentinel
```

### Auto-Update Feature

Build auto-update into the agent:

```go
// cmd/sentinel/update.go
func runUpdate(args []string) {
    current := Version
    latest := fetchLatestVersion()
    
    if current == latest {
        fmt.Println("✅ Already up to date:", current)
        return
    }
    
    fmt.Printf("📦 Updating %s → %s\n", current, latest)
    
    // Download new binary
    binary := downloadBinary(latest)
    
    // Replace current binary
    replaceBinary(binary)
    
    fmt.Println("✅ Updated successfully")
}
```

**Usage**:
```bash
# Check for updates
./sentinel update --check

# Update to latest
./sentinel update

# Auto-update on startup (configurable)
./sentinel config set autoUpdate true
```

---

## Part 2: Hub Deployment

### Strategy: Docker Compose + Optional Managed Services

Start simple, scale when needed.

### Prerequisites

- Server with Docker and Docker Compose
- Domain name (e.g., `sentinel.company.com`)
- 2 CPU, 4GB RAM, 50GB disk minimum

### Quick Start

```bash
# 1. Clone hub repository
git clone https://github.com/yourorg/sentinel-hub.git
cd sentinel-hub

# 2. Create environment file
cp .env.example .env

# 3. Generate secrets
cat >> .env << EOF
DB_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
DOMAIN=sentinel.company.com
EOF

# 4. Start services
docker-compose up -d

# 5. Initialize database
docker-compose exec api ./sentinel-hub migrate

# 6. Create admin user
docker-compose exec api ./sentinel-hub create-admin \
  --email admin@company.com \
  --password "secure-password-here"

# 7. Access dashboard
echo "Dashboard: https://sentinel.company.com"
```

### Docker Compose Configuration

> **Document Processing**: Hub includes document processing service with all
> dependencies (poppler, tesseract, LLM). Developers don't need to install anything.
> See [Architecture Decision](./ARCHITECTURE_DOCUMENT_PROCESSING.md).

```yaml
# docker-compose.yml
version: '3.8'

services:
  # ===================
  # API Server
  # ===================
  api:
    image: sentinel-hub-api:${VERSION:-latest}
    build:
      context: ./api
      dockerfile: Dockerfile
    environment:
      - DATABASE_URL=postgres://sentinel:${DB_PASSWORD}@db:5432/sentinel?sslmode=disable
      - JWT_SECRET=${JWT_SECRET}
      - CORS_ORIGIN=https://${DOMAIN}
      - LOG_LEVEL=info
      - DOCUMENT_STORAGE=/data/documents
      - OLLAMA_HOST=http://ollama:11434
    ports:
      - "8080:8080"
    volumes:
      - document_storage:/data/documents
    depends_on:
      db:
        condition: service_healthy
      ollama:
        condition: service_started
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  # ===================
  # Document Processor
  # ===================
  # Handles PDF, images, and other document formats
  # Dependencies installed here, NOT on developer machines
  processor:
    image: sentinel-hub-processor:${VERSION:-latest}
    build:
      context: ./processor
      dockerfile: Dockerfile
    environment:
      - DATABASE_URL=postgres://sentinel:${DB_PASSWORD}@db:5432/sentinel?sslmode=disable
      - DOCUMENT_STORAGE=/data/documents
      - OLLAMA_HOST=http://ollama:11434
      - WORKER_CONCURRENCY=4
      # Azure AI Foundry (Optional - for Claude Opus 4.5)
      # If not configured, system falls back to Ollama
      - AZURE_AI_ENDPOINT=${AZURE_AI_ENDPOINT:-}
      - AZURE_AI_KEY=${AZURE_AI_KEY:-}
      - AZURE_AI_DEPLOYMENT=${AZURE_AI_DEPLOYMENT:-claude-opus-4-5}
      - AZURE_AI_API_VERSION=${AZURE_AI_API_VERSION:-2024-02-01}
    volumes:
      - document_storage:/data/documents
    depends_on:
      - db
      - ollama
    restart: unless-stopped
    # Note: This container includes poppler, tesseract, libreoffice

  # ===================
  # LLM Service (Ollama)
  # ===================
  # Local LLM for knowledge extraction - no data leaves your server
  # Falls back to Ollama if Azure AI Foundry is not configured
  ollama:
    image: ollama/ollama:latest
    ports:
      - "127.0.0.1:11434:11434"  # Only localhost access
    volumes:
      - ollama_models:/root/.ollama
    restart: unless-stopped
    # Optional: GPU support for faster inference
    # deploy:
    #   resources:
    #     reservations:
    #       devices:
    #         - driver: nvidia
    #           count: 1
    #           capabilities: [gpu]

  # ===================
  # Dashboard (Frontend)
  # ===================
  dashboard:
    image: sentinel-hub-dashboard:${VERSION:-latest}
    build:
      context: ./dashboard
      dockerfile: Dockerfile
      args:
        - VITE_API_URL=https://${DOMAIN}/api
    ports:
      - "3000:80"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:80"]
      interval: 30s
      timeout: 10s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  # ===================
  # Database
  # ===================
  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=sentinel
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=sentinel
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    ports:
      - "127.0.0.1:5432:5432"  # Only localhost access
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U sentinel -d sentinel"]
      interval: 10s
      timeout: 5s
      retries: 5

  # ===================
  # Reverse Proxy (Nginx)
  # ===================
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/certs:/etc/nginx/certs:ro
      - ./nginx/certbot:/var/www/certbot:ro
    depends_on:
      - api
      - dashboard
    restart: unless-stopped

  # ===================
  # SSL Certificate Renewal
  # ===================
  certbot:
    image: certbot/certbot
    volumes:
      - ./nginx/certs:/etc/letsencrypt
      - ./nginx/certbot:/var/www/certbot
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait $${!}; done;'"
    restart: unless-stopped

volumes:
  postgres_data:
    driver: local
  document_storage:
    driver: local
  ollama_models:
    driver: local

networks:
  default:
    name: sentinel-network
```

### Azure AI Foundry Configuration (Optional)

For higher-quality knowledge extraction using Claude Opus 4.5, configure Azure AI Foundry:

1. **Set up Azure AI Foundry** (see [AZURE_SETUP_GUIDE.md](./AZURE_SETUP_GUIDE.md))
2. **Add to `.env` file**:
   ```bash
   AZURE_AI_ENDPOINT=https://your-resource.services.ai.azure.com
   AZURE_AI_KEY=your-api-key-here
   AZURE_AI_DEPLOYMENT=claude-opus-4-5
   AZURE_AI_API_VERSION=2024-02-01
   ```
3. **Restart processor**:
   ```bash
   docker-compose restart processor
   ```

**Provider Fallback**: If Azure is unavailable, the system automatically falls back to Ollama. This ensures knowledge extraction always works, even if Azure is down.

### Document Processor Dockerfile

The processor container includes all document parsing dependencies:

```dockerfile
# processor/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o processor ./cmd/processor

FROM ubuntu:22.04

# Install document processing dependencies
# (These are only installed on the server, not developer machines)
RUN apt-get update && apt-get install -y --no-install-recommends \
    poppler-utils \
    tesseract-ocr \
    tesseract-ocr-eng \
    tesseract-ocr-spa \
    tesseract-ocr-fra \
    tesseract-ocr-deu \
    libreoffice-writer \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/processor /usr/local/bin/processor

CMD ["processor", "worker"]
```

### Nginx Configuration

```nginx
# nginx/nginx.conf
events {
    worker_connections 1024;
}

http {
    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent"';
    
    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;

    # SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_session_cache shared:SSL:10m;

    # Upstream servers
    upstream api {
        server api:8080;
    }

    upstream dashboard {
        server dashboard:80;
    }

    # Redirect HTTP to HTTPS
    server {
        listen 80;
        server_name sentinel.company.com;
        
        # Let's Encrypt challenge
        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }
        
        location / {
            return 301 https://$host$request_uri;
        }
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name sentinel.company.com;

        ssl_certificate /etc/nginx/certs/live/sentinel.company.com/fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/live/sentinel.company.com/privkey.pem;

        # API routes
        location /api/ {
            proxy_pass http://api/;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # Timeout settings
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        # Dashboard (frontend)
        location / {
            proxy_pass http://dashboard;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # Health check endpoint
        location /health {
            access_log off;
            return 200 "OK";
            add_header Content-Type text/plain;
        }
    }
}
```

### Initial SSL Setup

```bash
# First-time SSL certificate setup
docker-compose run --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d sentinel.company.com \
  --email admin@company.com \
  --agree-tos \
  --no-eff-email

# Restart nginx to load certificates
docker-compose restart nginx
```

### Environment Configuration

```bash
# .env.example

# ===================
# Required
# ===================

# Database password (generate with: openssl rand -base64 32)
DB_PASSWORD=change-me-to-random-string

# JWT secret for authentication (generate with: openssl rand -base64 64)
JWT_SECRET=change-me-to-long-random-string

# Your domain
DOMAIN=sentinel.company.com

# ===================
# Optional
# ===================

# Version tag for images
VERSION=latest

# External database (if not using local)
# DATABASE_URL=postgres://user:pass@host:5432/sentinel

# SMTP for email alerts
# SMTP_HOST=smtp.company.com
# SMTP_PORT=587
# SMTP_USER=sentinel@company.com
# SMTP_PASS=smtp-password

# Slack for alerts
# SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx

# OpenAI for document ingestion (agents use this)
# OPENAI_API_KEY=sk-xxx
```

---

## Part 3: CI/CD Automation

### Hub Deployment Pipeline

```yaml
# .github/workflows/deploy-hub.yml
name: Deploy Hub

on:
  push:
    branches: [main]
    paths:
      - 'hub/**'
      - 'docker-compose.yml'
      - '.github/workflows/deploy-hub.yml'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: |
          cd hub/api
          go test -v ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Login to Registry
        uses: docker/login-action@v2
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Build and push API
        uses: docker/build-push-action@v4
        with:
          context: ./hub/api
          push: true
          tags: |
            ghcr.io/${{ github.repository }}/sentinel-hub-api:${{ github.sha }}
            ghcr.io/${{ github.repository }}/sentinel-hub-api:latest
      
      - name: Build and push Dashboard
        uses: docker/build-push-action@v4
        with:
          context: ./hub/dashboard
          push: true
          tags: |
            ghcr.io/${{ github.repository }}/sentinel-hub-dashboard:${{ github.sha }}
            ghcr.io/${{ github.repository }}/sentinel-hub-dashboard:latest

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: Deploy to server
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.HUB_HOST }}
          username: ${{ secrets.HUB_USER }}
          key: ${{ secrets.HUB_SSH_KEY }}
          script: |
            cd /opt/sentinel-hub
            
            # Pull new images
            docker-compose pull api dashboard
            
            # Rolling restart (zero downtime)
            docker-compose up -d --no-deps api
            sleep 10
            docker-compose up -d --no-deps dashboard
            
            # Cleanup old images
            docker image prune -f
            
            # Verify health
            curl -f http://localhost:8080/health || exit 1
            
            echo "✅ Deployed: ${{ github.sha }}"
```

### Deployment Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    AUTOMATED DEPLOYMENT FLOW                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  AGENT UPDATES                                                          │
│  ═════════════                                                          │
│                                                                          │
│  Developer                                                              │
│      │                                                                   │
│      ▼                                                                   │
│  git tag v1.2.0 && git push --tags                                     │
│      │                                                                   │
│      ▼                                                                   │
│  GitHub Actions                                                         │
│  ├── Build for darwin-amd64                                            │
│  ├── Build for darwin-arm64                                            │
│  ├── Build for linux-amd64                                             │
│  ├── Build for linux-arm64                                             │
│  └── Build for windows-amd64                                           │
│      │                                                                   │
│      ▼                                                                   │
│  Upload to GitHub Releases                                              │
│      │                                                                   │
│      ▼                                                                   │
│  Developers: ./sentinel update                                          │
│                                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  HUB UPDATES                                                            │
│  ═══════════                                                            │
│                                                                          │
│  Developer                                                              │
│      │                                                                   │
│      ▼                                                                   │
│  git push origin main (hub/* changes)                                  │
│      │                                                                   │
│      ▼                                                                   │
│  GitHub Actions                                                         │
│  ├── Run tests                                                         │
│  ├── Build API image                                                   │
│  ├── Build Dashboard image                                             │
│  └── Push to container registry                                        │
│      │                                                                   │
│      ▼                                                                   │
│  SSH to production server                                              │
│  ├── docker-compose pull                                               │
│  └── docker-compose up -d                                              │
│      │                                                                   │
│      ▼                                                                   │
│  Hub updated with zero downtime                                        │
│                                                                          │
│  ════════════════════════════════════════════════════════════════════  │
│                    ZERO MANUAL INTERVENTION                             │
│  ════════════════════════════════════════════════════════════════════  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Part 4: Operations

### Server Provisioning Script

```bash
#!/bin/bash
# provision-server.sh
# Run on a fresh Ubuntu 22.04 server

set -e

echo "🚀 Provisioning Sentinel Hub server..."

# Update system
apt-get update
apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
usermod -aG docker $USER

# Install Docker Compose
DOCKER_COMPOSE_VERSION="2.24.0"
curl -fsSL "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-linux-x86_64" \
  -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create directory structure
mkdir -p /opt/sentinel-hub
cd /opt/sentinel-hub

# Clone repository (or copy files)
git clone https://github.com/yourorg/sentinel-hub.git .

# Generate secrets
cat > .env << EOF
DB_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
DOMAIN=${DOMAIN:-sentinel.example.com}
VERSION=latest
EOF

# Start services
docker-compose up -d

echo "✅ Server provisioned!"
echo ""
echo "Next steps:"
echo "1. Configure DNS: ${DOMAIN} → $(curl -s ifconfig.me)"
echo "2. Run SSL setup: docker-compose run --rm certbot certonly ..."
echo "3. Create admin: docker-compose exec api ./sentinel-hub create-admin ..."
```

### Backup Script

```bash
#!/bin/bash
# backup.sh - Run daily via cron

set -e

BACKUP_DIR="/opt/sentinel-hub/backups"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7

# Create backup directory
mkdir -p "$BACKUP_DIR"

echo "📦 Starting backup: $DATE"

# Backup database
echo "   Backing up database..."
docker-compose exec -T db pg_dump -U sentinel sentinel | \
  gzip > "$BACKUP_DIR/db_$DATE.sql.gz"

# Backup configuration
echo "   Backing up configuration..."
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" \
  .env \
  docker-compose.yml \
  nginx/ \
  --exclude='nginx/certs'

# Backup SSL certificates (separately, less frequently)
if [ "$(date +%u)" = "1" ]; then  # Only on Mondays
  echo "   Backing up SSL certificates..."
  tar -czf "$BACKUP_DIR/certs_$DATE.tar.gz" nginx/certs/
fi

# Remove old backups
echo "   Cleaning up old backups..."
find "$BACKUP_DIR" -name "db_*.sql.gz" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "config_*.tar.gz" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "certs_*.tar.gz" -mtime +30 -delete

# Report
BACKUP_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
echo "✅ Backup completed: $DATE"
echo "   Total backup size: $BACKUP_SIZE"
```

Add to crontab:
```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /opt/sentinel-hub/backup.sh >> /var/log/sentinel-backup.log 2>&1
```

### Restore Script

```bash
#!/bin/bash
# restore.sh - Restore from backup

set -e

BACKUP_DIR="/opt/sentinel-hub/backups"

# List available backups
echo "Available database backups:"
ls -lh "$BACKUP_DIR"/db_*.sql.gz | tail -10

echo ""
read -p "Enter backup filename (e.g., db_20240115_020000.sql.gz): " BACKUP_FILE

if [ ! -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
    echo "❌ Backup file not found"
    exit 1
fi

echo "⚠️  This will replace the current database!"
read -p "Continue? [y/N]: " CONFIRM

if [ "$CONFIRM" != "y" ]; then
    echo "Cancelled."
    exit 0
fi

echo "🔄 Restoring from $BACKUP_FILE..."

# Stop API to prevent writes
docker-compose stop api

# Restore database
gunzip -c "$BACKUP_DIR/$BACKUP_FILE" | \
  docker-compose exec -T db psql -U sentinel sentinel

# Restart services
docker-compose up -d

echo "✅ Restore completed!"
```

### Health Check Script

```bash
#!/bin/bash
# health-check.sh

set -e

API_URL="${API_URL:-http://localhost:8080}"
DASHBOARD_URL="${DASHBOARD_URL:-http://localhost:3000}"

echo "🏥 Sentinel Health Check"
echo "========================"
echo ""

# Check API
echo -n "API Server: "
if curl -sf "$API_URL/health" > /dev/null; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

# Check Dashboard
echo -n "Dashboard:  "
if curl -sf "$DASHBOARD_URL" > /dev/null; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

# Check Database
echo -n "Database:   "
if docker-compose exec -T db pg_isready -U sentinel > /dev/null 2>&1; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

echo ""
echo "Container Status:"
docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "Resource Usage:"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"
```

### Log Management

```bash
# View logs
docker-compose logs -f                    # All services
docker-compose logs -f api                # API only
docker-compose logs -f --tail=100 api     # Last 100 lines

# Export logs
docker-compose logs --no-color > logs_$(date +%Y%m%d).txt

# Clear old logs (handled by Docker's log rotation)
# Configured in docker-compose.yml with max-size and max-file
```

---

## Part 5: Monitoring (Optional)

### Add Prometheus + Grafana

```yaml
# Add to docker-compose.yml

  # ===================
  # Prometheus (Metrics)
  # ===================
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.retention.time=30d'
    ports:
      - "127.0.0.1:9090:9090"
    restart: unless-stopped

  # ===================
  # Grafana (Dashboards)
  # ===================
  grafana:
    image: grafana/grafana:latest
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
    ports:
      - "127.0.0.1:3001:3000"
    depends_on:
      - prometheus
    restart: unless-stopped

volumes:
  prometheus_data:
  grafana_data:
```

### Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'sentinel-api'
    static_configs:
      - targets: ['api:8080']
    metrics_path: '/metrics'

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'postgres'
    static_configs:
      - targets: ['db:5432']
```

---

## Part 6: Scaling Guide

### When to Scale

| Agents | Recommendation |
|--------|----------------|
| < 100 | Single server, Docker Compose |
| 100-500 | Optimize database, add indexes |
| 500-2000 | Multiple API replicas |
| 2000-5000 | Managed database, load balancer |
| 5000+ | Kubernetes |

### Horizontal Scaling (Docker Compose)

```yaml
# Scale API to 3 instances
docker-compose up -d --scale api=3
```

Update nginx for load balancing:
```nginx
upstream api {
    least_conn;
    server api:8080;
}
```

### Database Optimization

```sql
-- Add indexes for common queries
CREATE INDEX CONCURRENTLY idx_telemetry_agent_created 
  ON telemetry(agent_id, created_at DESC);

CREATE INDEX CONCURRENTLY idx_telemetry_org_event 
  ON telemetry(org_id, event_type, created_at DESC);

-- Analyze tables after bulk inserts
ANALYZE telemetry;

-- Set up table partitioning for large deployments
-- (Requires PostgreSQL 11+)
```

### Move to Managed Database

When self-hosted Postgres becomes a burden:

```bash
# Update .env
DATABASE_URL=postgres://user:pass@your-rds-instance.region.rds.amazonaws.com:5432/sentinel

# Remove db service from docker-compose.yml
# Restart with external database
docker-compose up -d
```

---

## Part 7: Cost Estimates

### Minimal Setup (Recommended Start)

| Component | Service | Monthly Cost |
|-----------|---------|--------------|
| Hub Server | DigitalOcean Droplet (2 CPU, 4GB) | $24 |
| Domain | Your existing | $0 |
| SSL | Let's Encrypt | $0 |
| DNS | Cloudflare Free | $0 |
| **Total** | | **~$24/month** |

### Production Setup

| Component | Service | Monthly Cost |
|-----------|---------|--------------|
| Hub Server | AWS EC2 t3.medium | $30 |
| Database | AWS RDS PostgreSQL (db.t3.micro) | $15 |
| Load Balancer | AWS ALB | $20 |
| Storage | S3 (agent binaries) | $5 |
| Monitoring | CloudWatch | $10 |
| **Total** | | **~$80/month** |

### Enterprise Setup

| Component | Service | Monthly Cost |
|-----------|---------|--------------|
| Kubernetes | AWS EKS | $75 |
| Database | RDS Multi-AZ | $50 |
| Load Balancer | ALB | $20 |
| Storage | S3 | $10 |
| Monitoring | Datadog | $50 |
| Backups | S3 + lifecycle | $5 |
| **Total** | | **~$210/month** |

---

## Deployment Checklist

### Initial Hub Deployment

- [ ] Provision server (2 CPU, 4GB RAM, 50GB disk)
- [ ] Install Docker and Docker Compose
- [ ] Clone hub repository
- [ ] Configure `.env` with secure secrets
- [ ] Start services: `docker-compose up -d`
- [ ] Configure DNS to point to server IP
- [ ] Set up SSL with Let's Encrypt
- [ ] Create admin user
- [ ] Test dashboard access
- [ ] Test API health endpoint

### Agent Distribution Setup

- [ ] Set up GitHub Actions for multi-platform builds
- [ ] Host install script on internal server
- [ ] Test installation on macOS, Linux, Windows
- [ ] Configure auto-update endpoint
- [ ] Document installation for developers
- [ ] Test hub connectivity from agent

### Operations Setup

- [ ] Configure automated backups (cron)
- [ ] Test backup restore procedure
- [ ] Set up log rotation
- [ ] Configure health check alerts
- [ ] Document runbooks for common issues
- [ ] Set up monitoring (optional)

### Security Checklist

- [ ] Strong passwords in `.env`
- [ ] Database not exposed publicly
- [ ] HTTPS enforced
- [ ] API rate limiting configured
- [ ] Firewall rules in place
- [ ] SSH key-based auth only

---

## Troubleshooting

### Hub Won't Start

```bash
# Check container status
docker-compose ps

# View logs
docker-compose logs api
docker-compose logs db

# Common fix: wait for database
docker-compose restart api
```

### Database Connection Failed

```bash
# Check if database is ready
docker-compose exec db pg_isready

# Check connection string
docker-compose exec api env | grep DATABASE

# Reset database
docker-compose down -v  # WARNING: Deletes data
docker-compose up -d
```

### SSL Certificate Issues

```bash
# Check certificate status
docker-compose exec certbot certbot certificates

# Force renewal
docker-compose run --rm certbot renew --force-renewal

# Restart nginx
docker-compose restart nginx
```

### Agent Can't Connect

```bash
# Test from agent machine
curl -v https://sentinel.company.com/api/health

# Check agent config
./sentinel config get telemetry.endpoint

# Test with verbose output
./sentinel audit --verbose
```

### Out of Disk Space

```bash
# Clean Docker
docker system prune -a

# Check disk usage
df -h
du -sh /var/lib/docker

# Clean old backups
find /opt/sentinel-hub/backups -mtime +7 -delete
```

---

## Support

For issues:
1. Check logs: `docker-compose logs`
2. Run health check: `./health-check.sh`
3. Review this guide's troubleshooting section
4. Open an issue with logs attached

