#!/bin/bash
# Generate secure secrets for production deployment

set -e

echo "🔐 Generating secure secrets for production..."

# Generate DB password (32 characters, alphanumeric + special chars)
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
echo "DB_PASSWORD=$DB_PASSWORD"

# Generate JWT secret (64 hex characters)
JWT_SECRET=$(openssl rand -hex 32)
echo "JWT_SECRET=$JWT_SECRET"

# Generate Admin API key (64 hex characters)
ADMIN_API_KEY=$(openssl rand -hex 32)
echo "ADMIN_API_KEY=$ADMIN_API_KEY"

# Generate .env file
cat > .env <<EOF
# Generated on $(date)
# SECURITY: Keep this file secret. Never commit to git.

# Database Configuration
DATABASE_URL=postgres://sentinel:${DB_PASSWORD}@db:5432/sentinel?sslmode=require
DB_PASSWORD=${DB_PASSWORD}

# Security
JWT_SECRET=${JWT_SECRET}
ADMIN_API_KEY=${ADMIN_API_KEY}
CORS_ORIGIN=https://yourdomain.com

# Server Configuration
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info

# Storage
DOCUMENT_STORAGE=/data/documents

# External Services
OLLAMA_HOST=http://ollama:11434
HUB_URL=https://yourdomain.com

# Worker Configuration
WORKER_CONCURRENCY=4
EOF

echo ""
echo "✅ Secrets generated in .env file"
echo "⚠️  IMPORTANT:"
echo "   1. Review .env file and update CORS_ORIGIN and HUB_URL"
echo "   2. Never commit .env to git"
echo "   3. Store .env securely (use secrets management in production)"


