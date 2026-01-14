# Sentinel Hub

Central server for document processing, metrics, and organization management.

## Overview

Sentinel Hub processes documents uploaded by Sentinel Agents, eliminating the need for
developers to install document parsing dependencies (poppler, tesseract) on their machines.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Sentinel Hub                             │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   API       │  │  Processor  │  │  Dependencies       │  │
│  │   Server    │  │  Worker     │  │  ├── poppler-utils  │  │
│  │             │  │             │  │  ├── tesseract-ocr  │  │
│  │  Port 8080  │  │  Parallel   │  │  └── pandoc         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│         │                │                                   │
│         └────────────────┼──────────────────────────────────│
│                          │                                   │
│                  ┌───────▼───────┐  ┌─────────────────────┐ │
│                  │  PostgreSQL   │  │  Ollama (LLM)       │ │
│                  │  Database     │  │  Knowledge Extract  │ │
│                  └───────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# 1. Clone and enter directory
cd hub

# 2. Run setup script
./scripts/setup.sh

# 3. Note the API key displayed at the end
```

## Manual Setup

```bash
# 1. Create environment file
cp .env.example .env

# 2. Generate secrets
DB_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
echo "DB_PASSWORD=$DB_PASSWORD" >> .env
echo "JWT_SECRET=$JWT_SECRET" >> .env

# 3. Start services
docker-compose up -d

# 4. Create organization and project
curl -X POST http://localhost:8080/api/v1/admin/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "My Organization"}'

curl -X POST http://localhost:8080/api/v1/admin/projects \
  -H "Content-Type: application/json" \
  -d '{"org_id": "<org-id>", "name": "My Project"}'

# 5. Note the API key from the response
```

## API Endpoints

### Document Processing

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/documents/ingest` | Upload document |
| GET | `/api/v1/documents/{id}/status` | Check processing status |
| GET | `/api/v1/documents/{id}/extracted` | Get extracted text |
| GET | `/api/v1/documents/{id}/knowledge` | Get knowledge items |
| GET | `/api/v1/documents` | List all documents |

### Administration

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/admin/organizations` | Create organization |
| POST | `/api/v1/admin/projects` | Create project |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

## Agent Configuration

Add Hub configuration to `.sentinelsrc`:

```json
{
  "hub": {
    "url": "http://localhost:8080",
    "apiKey": "sk_live_xxxxx"
  }
}
```

Or use environment variables:

```bash
export SENTINEL_HUB_URL="http://localhost:8080"
export SENTINEL_API_KEY="sk_live_xxxxx"
```

## Supported Document Formats

| Format | Extension | Parser |
|--------|-----------|--------|
| Text | .txt, .md | Native |
| PDF | .pdf | pdftotext |
| Word | .docx | pandoc/native |
| Excel | .xlsx | native |
| Email | .eml | native |
| Images | .png, .jpg | tesseract |

## Services

### API Server (`api/`)

- Go-based REST API
- Handles authentication, uploads, queries
- Stores metadata in PostgreSQL

### Document Processor (`processor/`)

- Worker pool for parallel processing
- Contains all parsing dependencies
- Integrates with Ollama for LLM extraction

### Database (PostgreSQL)

- Stores documents, extractions, knowledge items
- Handles queue for processing

### LLM (Ollama)

- Local LLM for knowledge extraction
- No data leaves your server
- Optional GPU acceleration

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | - |
| `JWT_SECRET` | JWT signing secret | - |
| `DOCUMENT_STORAGE` | Document storage path | /data/documents |
| `OLLAMA_HOST` | Ollama server URL | http://ollama:11434 |
| `CORS_ORIGIN` | Allowed CORS origin | * |
| `WORKER_CONCURRENCY` | Number of processor workers | 4 |

## Security

- All API requests require Bearer token authentication
- Documents encrypted at rest
- TLS recommended for production
- API keys scoped per project

## Development

```bash
# Run API server locally
cd api
go run main.go

# Run processor locally
cd processor
go run main.go

# Run tests
go test ./...
```

## Production Deployment

See [DEPLOYMENT_GUIDE.md](../docs/external/DEPLOYMENT_GUIDE.md) for production setup including:

- SSL/TLS configuration
- Reverse proxy (nginx)
- Managed database options
- Scaling considerations












