#!/bin/bash
# Sentinel Hub Setup Script
# Usage: ./scripts/setup.sh

set -e

echo "🚀 Sentinel Hub Setup"
echo "══════════════════════════════════════════════════════════════"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

echo "✅ Docker and Docker Compose found"

# Create .env if not exists
if [ ! -f .env ]; then
    echo ""
    echo "📝 Creating .env file..."
    
    # Generate secrets
    DB_PASSWORD=$(openssl rand -base64 32 | tr -d '/+=' | head -c 32)
    JWT_SECRET=$(openssl rand -base64 64 | tr -d '/+=' | head -c 64)
    
    cat > .env << EOF
# Sentinel Hub Configuration
# Generated on $(date)

# Database
DB_PASSWORD=${DB_PASSWORD}

# Security
JWT_SECRET=${JWT_SECRET}

# CORS (set to your domain in production)
CORS_ORIGIN=*

# Worker settings
WORKER_CONCURRENCY=4
EOF

    echo "✅ Created .env file with generated secrets"
else
    echo "✅ .env file already exists"
fi

# Build images
echo ""
echo "🔨 Building Docker images..."
docker-compose build

# Start services
echo ""
echo "🚀 Starting services..."
docker-compose up -d

# Wait for database
echo ""
echo "⏳ Waiting for database to be ready..."
sleep 5

# Check health
echo ""
echo "🔍 Checking service health..."

MAX_RETRIES=30
RETRY=0
while [ "$RETRY" -lt "$MAX_RETRIES" ]; do
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ API server is healthy"
        break
    fi
    RETRY=$((RETRY + 1))
    echo "   Waiting for API server... ($RETRY/$MAX_RETRIES)"
    sleep 2
done

if [ "$RETRY" -eq "$MAX_RETRIES" ]; then
    echo "❌ API server failed to start. Check logs: docker-compose logs api"
    exit 1
fi

# Create default organization and project
echo ""
echo "📦 Setting up default organization..."

ORG_RESPONSE=$(curl -sf -X POST http://localhost:8080/api/v1/admin/organizations \
    -H "Content-Type: application/json" \
    -d '{"name": "Default Organization"}')

if [ $? -eq 0 ]; then
    ORG_ID=$(echo "$ORG_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "✅ Created organization: $ORG_ID"
    
    echo ""
    echo "📁 Creating default project..."
    
    PROJECT_RESPONSE=$(curl -sf -X POST http://localhost:8080/api/v1/admin/projects \
        -H "Content-Type: application/json" \
        -d "{\"org_id\": \"$ORG_ID\", \"name\": \"Default Project\"}")
    
    if [ $? -eq 0 ]; then
        API_KEY=$(echo "$PROJECT_RESPONSE" | grep -o '"api_key":"[^"]*"' | cut -d'"' -f4)
        PROJECT_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        
        echo "✅ Created project: $PROJECT_ID"
        echo ""
        echo "══════════════════════════════════════════════════════════════"
        echo "🎉 Setup Complete!"
        echo "══════════════════════════════════════════════════════════════"
        echo ""
        echo "API URL:  http://localhost:8080"
        echo "API Key:  $API_KEY"
        echo ""
        echo "Save this API key! You'll need it to configure Sentinel Agent."
        echo ""
        echo "To configure Agent, add to .sentinelsrc:"
        echo ""
        echo '  "hub": {'
        echo "    \"url\": \"http://localhost:8080\","
        echo "    \"apiKey\": \"$API_KEY\""
        echo '  }'
        echo ""
        echo "Or set environment variable:"
        echo "  export SENTINEL_API_KEY=\"$API_KEY\""
        echo ""
    fi
else
    echo "⚠️  Could not create default organization (may already exist)"
fi

# Pull Ollama model
echo ""
echo "📥 Pulling LLM model (this may take a while)..."
docker exec sentinel-hub-ollama-1 ollama pull llama2 2>/dev/null || \
docker exec hub-ollama-1 ollama pull llama2 2>/dev/null || \
echo "⚠️  Could not pull LLM model. Run manually: docker exec <ollama-container> ollama pull llama2"

echo ""
echo "🏁 Done! Hub is running at http://localhost:8080"

