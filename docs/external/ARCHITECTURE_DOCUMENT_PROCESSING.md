# Server-Side Document Processing Architecture

## Executive Summary

This document outlines the architectural decision to move document ingestion and processing from local developer machines to a centralized Sentinel Hub server. This approach eliminates dependency management issues, ensures consistent processing across all developers, and enables advanced LLM-powered knowledge extraction.

---

## Problem Statement

### Original Approach: Local Processing

The initial implementation processed documents locally on each developer's machine:

```
Developer Machine
├── Sentinel Agent
├── pdftotext (requires installation)
├── tesseract (requires installation)
└── docs/knowledge/extracted/
```

**Issues Identified:**

| Problem | Impact |
|---------|--------|
| Dependency management | Each developer must install poppler, tesseract |
| Platform inconsistency | Different OS versions = different behavior |
| IT restrictions | Enterprise environments may block installations |
| Maintenance burden | Updates needed on 50+ machines |
| No LLM extraction | Can't run AI models locally |
| Quality variance | Different tesseract versions = different OCR quality |

---

## Solution: Server-Side Processing

Move all document processing to Sentinel Hub, keeping the Agent lightweight.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SENTINEL HUB (Central Server)                      │
│                                                                              │
│  ┌─────────────────┐  ┌──────────────────┐  ┌─────────────────────────────┐ │
│  │  API Server     │  │ Document Service │  │  Dependencies (Docker)      │ │
│  │  /api/v1/       │  │                  │  │                             │ │
│  │  ├── ingest     │◄─┤  ├── Job Queue   │◄─┤  ├── poppler-utils ✓        │ │
│  │  ├── status     │  │  ├── PDF Parser  │  │  ├── tesseract-ocr ✓        │ │
│  │  ├── download   │  │  ├── DOCX Parser │  │  ├── libreoffice ✓          │ │
│  │  └── knowledge  │  │  ├── OCR Engine  │  │  └── LLM (Ollama/OpenAI) ✓  │ │
│  └─────────────────┘  │  └── LLM Extract │  └─────────────────────────────┘ │
│           │           └──────────────────┘                                   │
│           │                    │                                             │
│           │           ┌────────▼────────┐                                    │
│           │           │   PostgreSQL     │                                   │
│           │           │  ├── documents   │                                   │
│           └──────────►│  ├── extractions │                                   │
│                       │  └── knowledge   │                                   │
│                       └─────────────────┘                                    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                    Encrypted Document Storage                            ││
│  │  /data/documents/{org_id}/{project_id}/{doc_id}/                        ││
│  └─────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────┘
                                    ▲
                                    │ HTTPS (TLS 1.3)
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
        ┌───────────▼───────────┐       ┌──────────▼────────────┐
        │  Developer Machine A   │       │  Developer Machine B   │
        │                        │       │                        │
        │  ┌─────────────────┐   │       │  ┌─────────────────┐   │
        │  │ Sentinel Agent  │   │       │  │ Sentinel Agent  │   │
        │  │ (Lightweight)   │   │       │  │ (Lightweight)   │   │
        │  │                 │   │       │  │                 │   │
        │  │ No dependencies │   │       │  │ No dependencies │   │
        │  │ required!       │   │       │  │ required!       │   │
        │  └─────────────────┘   │       │  └─────────────────┘   │
        │                        │       │                        │
        │  docs/knowledge/       │       │  docs/knowledge/       │
        │  └── (synced)          │       │  └── (synced)          │
        └────────────────────────┘       └────────────────────────┘
```

---

## Benefits

| Aspect | Local Processing | Server Processing |
|--------|-----------------|-------------------|
| **Dependencies** | Each developer installs | Hub only (once) |
| **PDF Support** | Manual setup | ✅ Always available |
| **OCR Support** | Manual setup | ✅ Always available |
| **LLM Extraction** | ❌ Not possible | ✅ Built-in |
| **Consistency** | Varies by machine | Identical for all |
| **Maintenance** | N machines | 1 server |
| **Updates** | Push to all devs | Single deployment |
| **Quality** | Variable | Controlled |

---

## Data Flow

### Document Upload Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Developer  │     │    Agent     │     │     Hub      │     │   Storage    │
│              │     │              │     │              │     │              │
│  Drop file   │────►│  Validate    │────►│  Receive     │────►│  Encrypt &   │
│              │     │  & Upload    │     │  & Queue     │     │  Store       │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                                                 │
                                                 ▼
                     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
                     │   Notify     │◄────│  Extract     │◄────│   Parse      │
                     │   Agent      │     │  Knowledge   │     │   Document   │
                     └──────────────┘     └──────────────┘     └──────────────┘
                            │
                            ▼
                     ┌──────────────┐     ┌──────────────┐
                     │   Sync to    │────►│  Human       │
                     │   Local      │     │  Validation  │
                     └──────────────┘     └──────────────┘
```

### Processing Pipeline

```
1. UPLOAD
   └── Agent uploads file to Hub via HTTPS
   └── Hub validates file type, size, virus scan
   └── File stored encrypted at rest

2. QUEUE
   └── Job added to processing queue
   └── Priority: high (small files) → normal → low (large files)

3. PARSE
   └── PDF  → pdftotext (preserves layout)
   └── DOCX → XML extraction (native)
   └── XLSX → Cell extraction (native)
   └── IMG  → Tesseract OCR
   └── EML  → Header + body extraction

4. EXTRACT (LLM)
   └── Extracted text → LLM prompt
   └── Identify: Business rules, entities, glossary
   └── Generate structured knowledge items
   └── Confidence scoring

5. STORE
   └── Raw text saved
   └── Knowledge items saved
   └── Linked to project

6. SYNC
   └── Agent polls for completion
   └── Downloads results to docs/knowledge/
   └── Marks as pending human validation
```

---

## API Specification

### Authentication

All API requests require authentication via API key:

```http
Authorization: Bearer sk_live_xxxxxxxxxxxxx
```

API keys are scoped per project and can have permissions:
- `documents:write` - Upload documents
- `documents:read` - Download extractions
- `knowledge:write` - Approve knowledge items

### Endpoints

#### Upload Document

```http
POST /api/v1/documents/ingest
Content-Type: multipart/form-data

Parameters:
  file: <binary>              # Required: Document file
  project_id: string          # Required: Project identifier
  extract_knowledge: boolean  # Optional: Run LLM extraction (default: true)
  priority: string            # Optional: "high" | "normal" | "low"
  callback_url: string        # Optional: Webhook for completion

Response (202 Accepted):
{
  "id": "doc_abc123xyz",
  "status": "queued",
  "estimated_time_seconds": 30,
  "position_in_queue": 3
}
```

#### Check Status

```http
GET /api/v1/documents/{id}/status

Response (200 OK):
{
  "id": "doc_abc123xyz",
  "original_name": "requirements.pdf",
  "status": "completed",  // queued | processing | completed | failed
  "progress_percent": 100,
  "stages": {
    "upload": { "status": "completed", "duration_ms": 1234 },
    "parsing": { "status": "completed", "duration_ms": 5678 },
    "extraction": { "status": "completed", "duration_ms": 12000 }
  },
  "result": {
    "pages": 12,
    "text_length": 45230,
    "knowledge_items": 8
  },
  "created_at": "2024-12-04T10:30:00Z",
  "completed_at": "2024-12-04T10:30:45Z"
}
```

#### Download Extracted Content

```http
GET /api/v1/documents/{id}/extracted

Response (200 OK):
{
  "id": "doc_abc123xyz",
  "original_name": "requirements.pdf",
  "extracted_text": "# Project Requirements...",
  "metadata": {
    "pages": 12,
    "word_count": 4523,
    "language": "en"
  }
}
```

#### Download Knowledge Items

```http
GET /api/v1/documents/{id}/knowledge

Response (200 OK):
{
  "id": "doc_abc123xyz",
  "knowledge_items": [
    {
      "id": "ki_001",
      "type": "business_rule",
      "title": "Order Cancellation Policy",
      "content": "Orders can be cancelled within 24 hours of creation...",
      "confidence": 0.92,
      "source_page": 5,
      "status": "pending_review"
    },
    {
      "id": "ki_002", 
      "type": "entity",
      "title": "User",
      "content": "Represents a registered customer with email, name, role...",
      "confidence": 0.88,
      "source_page": 8,
      "status": "pending_review"
    }
  ]
}
```

#### List Project Documents

```http
GET /api/v1/projects/{project_id}/documents

Response (200 OK):
{
  "documents": [
    {
      "id": "doc_abc123",
      "name": "requirements.pdf",
      "status": "completed",
      "knowledge_items": 8,
      "uploaded_at": "2024-12-04T10:30:00Z"
    }
  ],
  "total": 15,
  "page": 1,
  "per_page": 20
}
```

---

## Agent Commands

### Upload Documents

```bash
# Upload single document
./sentinel ingest /path/to/requirements.pdf

# Upload directory
./sentinel ingest /path/to/project-docs/

# Upload with high priority
./sentinel ingest /path/to/urgent.pdf --priority high

# Upload without LLM extraction (text only)
./sentinel ingest /path/to/doc.pdf --no-extract
```

### Check Status

```bash
# Show all pending documents
./sentinel ingest --status

# Output:
# 📊 Document Processing Status
# ══════════════════════════════════════════════════════════════
# 
# Project: ecommerce-platform
# 
# │ Document              │ Status      │ Progress │ Items │
# ├───────────────────────┼─────────────┼──────────┼───────┤
# │ requirements.pdf      │ ✅ Completed │ 100%     │ 8     │
# │ scope.docx            │ ⏳ Processing│ 60%      │ -     │
# │ architecture.png      │ 📋 Queued   │ -        │ -     │
# └───────────────────────┴─────────────┴──────────┴───────┘
```

### Sync Results

```bash
# Sync all completed documents
./sentinel ingest --sync

# Output:
# 📥 Syncing from Hub...
# 
# ✅ requirements.pdf → docs/knowledge/extracted/requirements.txt
#    └── 8 knowledge items → docs/knowledge/business/requirements/
# 
# Synced: 1 document, 8 knowledge items
```

### Offline Mode

```bash
# Force local processing (limited formats)
./sentinel ingest /path/to/doc.txt --offline

# Check what's supported offline
./sentinel ingest --offline-info

# Output:
# 📴 Offline Mode Capabilities
# ══════════════════════════════════════════════════════════════
# 
# ✅ Supported (no dependencies):
#    • .txt, .md, .markdown (text files)
#    • .docx (Word documents)
#    • .xlsx (Excel spreadsheets)
#    • .eml (Email files)
# 
# ⚠️ Requires local dependencies:
#    • .pdf → Install: brew install poppler
#    • .png, .jpg → Install: brew install tesseract
# 
# ❌ Not available offline:
#    • LLM knowledge extraction
```

---

## Configuration

### Agent Configuration (.sentinelsrc)

```json
{
  "hub": {
    "url": "https://sentinel-hub.company.com",
    "apiKey": "${SENTINEL_API_KEY}",
    "projectId": "ecommerce-platform",
    "timeout": 30
  },
  "ingest": {
    "autoSync": true,
    "syncInterval": 60,
    "offlineFallback": true,
    "extractKnowledge": true,
    "maxFileSize": "50MB",
    "allowedTypes": ["pdf", "docx", "xlsx", "txt", "md", "eml", "png", "jpg"]
  }
}
```

### Environment Variables

```bash
# API authentication (recommended over config file)
export SENTINEL_API_KEY="sk_live_xxxxxxxxxxxxx"

# Hub URL (if not in config)
export SENTINEL_HUB_URL="https://sentinel-hub.company.com"

# Force offline mode
export SENTINEL_OFFLINE=true
```

---

## Security

### Data Protection

| Layer | Protection |
|-------|------------|
| Transport | TLS 1.3 (HTTPS only) |
| Storage | AES-256 encryption at rest |
| Access | API key + project scoping |
| Isolation | Separate storage per organization |
| Retention | Auto-delete after 30 days (configurable) |

### Document Handling

```
Upload → Virus Scan → Encrypt → Store → Process → Delete Original
                                            │
                                            ▼
                                    Keep extracted text only
                                    (original deleted after processing)
```

### Audit Logging

All operations are logged:

```json
{
  "timestamp": "2024-12-04T10:30:00Z",
  "action": "document.upload",
  "user_id": "user_123",
  "project_id": "ecommerce",
  "document_id": "doc_abc",
  "ip_address": "10.0.1.50",
  "user_agent": "Sentinel-Agent/24.0"
}
```

---

## Deployment

### Hub Deployment (Docker Compose)

```yaml
version: '3.8'

services:
  hub:
    image: sentinel-hub:latest
    ports:
      - "443:8080"
    environment:
      - DATABASE_URL=postgres://sentinel:${DB_PASS}@db:5432/sentinel
      - STORAGE_PATH=/data/documents
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
      - LLM_PROVIDER=ollama
      - OLLAMA_HOST=http://ollama:11434
    volumes:
      - document_storage:/data/documents
      - ./certs:/etc/ssl/certs
    depends_on:
      - db
      - ollama

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=sentinel
      - POSTGRES_PASSWORD=${DB_PASS}
      - POSTGRES_DB=sentinel
    volumes:
      - postgres_data:/var/lib/postgresql/data

  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_models:/root/.ollama
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]

volumes:
  document_storage:
  postgres_data:
  ollama_models:
```

### System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 4 cores | 8+ cores |
| RAM | 8 GB | 16+ GB |
| Storage | 100 GB SSD | 500+ GB SSD |
| GPU | - | NVIDIA (for LLM) |

---

## Comparison: Before vs After

| Scenario | Before (Local) | After (Server) |
|----------|---------------|----------------|
| New developer onboarding | Install Go, poppler, tesseract | Install Agent only |
| Processing a PDF | `brew install poppler` first | Just works |
| OCR an image | `brew install tesseract` first | Just works |
| Extract business rules | Manual reading | Automatic LLM |
| Update parser | Update on all machines | Single deployment |
| Offline work | Full functionality | Basic formats only |
| Processing speed | Depends on dev machine | Consistent, fast |

---

## Migration Path

### For Existing Local Installations

1. **Update Agent** to latest version
2. **Configure Hub** URL and API key
3. **Existing local docs** remain accessible
4. **New uploads** go to Hub automatically
5. **Gradual migration** of existing documents (optional)

```bash
# Migrate existing extracted docs to Hub
./sentinel ingest --migrate-local

# This will:
# 1. Upload docs from docs/knowledge/source-documents/
# 2. Re-process with LLM extraction
# 3. Sync enhanced knowledge back
```

---

## Conclusion

Server-side document processing provides:

1. **Zero-dependency developer experience** - No manual setup required
2. **Consistent quality** - Same processing for all team members
3. **Advanced capabilities** - LLM extraction not possible locally
4. **Centralized management** - Single point of maintenance
5. **Better security** - Controlled environment, audit logging
6. **Scalability** - Handle any team size

The trade-off is requiring network connectivity, addressed by offline fallback for basic formats.












