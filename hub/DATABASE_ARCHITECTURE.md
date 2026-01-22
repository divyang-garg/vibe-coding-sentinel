# Sentinel Hub Database Architecture

## Overview

The Sentinel Hub uses **PostgreSQL** as its primary database, deployed via **Docker Compose**. The database is tightly integrated with the Hub API service.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Sentinel Hub Stack                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐         ┌──────────────┐                │
│  │  Hub API     │◄───────►│  PostgreSQL  │                │
│  │  Service     │         │  Database    │                │
│  │  :8080       │         │  :5432       │                │
│  └──────────────┘         └──────────────┘                │
│         │                        │                          │
│         │                        │                          │
│         ▼                        ▼                          │
│  ┌──────────────┐         ┌──────────────┐                │
│  │   Ollama     │         │   Volumes    │                │
│  │   (LLM)      │         │  (Persistent)│                │
│  │   :11434     │         └──────────────┘                │
│  └──────────────┘                                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Database Deployment

### Production/Development Environment

**Location:** `hub/docker-compose.yml`

**Database Service:**
- **Image:** `postgres:15-alpine`
- **User:** `sentinel`
- **Database:** `sentinel`
- **Port:** `127.0.0.1:5433:5432` (mapped to localhost:5433)
- **Volume:** `postgres_data` (persistent storage)

**Connection String:**
```
postgres://sentinel:${DB_PASSWORD}@db:5432/sentinel?sslmode=disable
```

**Key Points:**
- Database runs in Docker container named `db`
- API service connects via Docker network (`db:5432`)
- Port 5433 exposed on localhost for external access
- Password configured via environment variable

### Test Environment

**Location:** `hub/docker-compose.test.yml`

**Test Database:**
- **Database:** `sentinel_test`
- **Port:** `127.0.0.1:5433:5432`
- **Auto-initialization:** Migrations run on container startup

**Features:**
- Separate test database to avoid data conflicts
- Migrations automatically applied on container start
- Isolated network for testing

## Migration Strategy

### Current Setup

**Migration Files Location:** `hub/migrations/`

**Files:**
1. `001_create_hook_tables.sql`
2. `002_create_core_tables.sql`
3. `003_create_workflow_tables.sql`
4. `004_create_task_dependencies.sql`
5. `005_create_llm_usage_table.sql`
6. `001_add_api_key_hashing.sql` (NEW - security remediation)

### How Migrations Are Applied

#### Option 1: Automatic (Test Environment)
In `docker-compose.test.yml`, migrations are mounted as initialization scripts:
```yaml
volumes:
  - ./migrations:/docker-entrypoint-initdb.d/02-migrations
```
**Behavior:** Migrations run automatically when the test database container starts for the first time.

#### Option 2: Manual (Production/Development)
Migrations must be applied manually:

```bash
# Connect to running database container
docker exec -i sentinel-postgres psql -U sentinel -d sentinel < hub/migrations/001_add_api_key_hashing.sql

# Or from host (if port is exposed)
psql -h localhost -p 5433 -U sentinel -d sentinel -f hub/migrations/001_add_api_key_hashing.sql
```

## Database Schema

### Core Tables (from `002_create_core_tables.sql`):
- `organizations` - Customer organizations
- `projects` - Projects within organizations (API keys stored here)
- `users` - User accounts
- `teams` - Teams within organizations

### Projects Table Structure

**Current:**
```sql
CREATE TABLE projects (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    api_key VARCHAR(255),              -- Legacy: plaintext (deprecated)
    api_key_hash VARCHAR(64),          -- NEW: SHA-256 hash
    api_key_prefix VARCHAR(8),         -- NEW: First 8 chars
    created_at TIMESTAMP NOT NULL
);
```

**Migration:** The `001_add_api_key_hashing.sql` migration adds the hash columns.

## Is This Expected?

### ✅ YES - This is the Correct Architecture

**Why the migration should be applied to the Hub database:**

1. **Centralized Authentication:** API keys are stored in the `projects` table, which is part of the Hub's core schema
2. **Service Integration:** The `OrganizationService.ValidateAPIKey()` method queries this database
3. **Hub Responsibilities:** The Hub manages:
   - Organization and project data
   - API key generation and validation
   - Authentication and authorization
   - Team collaboration features

4. **Database Ownership:** The Hub API service owns and manages the database schema
5. **Migration Location:** Migration files are correctly placed in `hub/migrations/`

### Architecture Decision

The **Hub API** and **PostgreSQL database** form a **tightly coupled unit**:
- Hub API = Application Layer (business logic, HTTP API)
- PostgreSQL = Data Layer (persistence, relationships)
- Together = Complete Hub service

The database is **NOT** shared with external services. It's the Hub's dedicated database.

## Applying the Migration

### Step 1: Ensure Database is Running

```bash
cd /Users/divyanggarg/VicecodingSentinel/hub
docker-compose ps  # Check if db container is running
```

### Step 2: Apply Migration

**Option A: Via Docker Exec (Recommended)**
```bash
# Copy migration into container and run
docker cp hub/migrations/001_add_api_key_hashing.sql $(docker-compose ps -q db):/tmp/
docker exec -i $(docker-compose ps -q db) psql -U sentinel -d sentinel -f /tmp/001_add_api_key_hashing.sql
```

**Option B: Via Exposed Port**
```bash
# If using the root docker-compose.yml (port 5433)
export PGPASSWORD=${DB_PASSWORD}  # From .env file
psql -h localhost -p 5433 -U sentinel -d sentinel -f hub/migrations/001_add_api_key_hashing.sql
```

**Option C: Using docker-compose exec**
```bash
cd hub
docker-compose exec db psql -U sentinel -d sentinel -f /tmp/001_add_api_key_hashing.sql
# (First copy file into container or mount migrations directory)
```

### Step 3: Verify Migration

```bash
# Connect to database
docker exec -it $(docker-compose ps -q db) psql -U sentinel -d sentinel

# Check columns
\d projects

# Should show:
# - api_key_hash VARCHAR(64)
# - api_key_prefix VARCHAR(8)

# Check indexes
\di idx_projects_api_key_hash
\di idx_projects_api_key_prefix

# Verify existing keys were migrated
SELECT COUNT(*) as total, 
       COUNT(api_key_hash) as hashed_keys 
FROM projects WHERE api_key IS NOT NULL;
```

## Integration with Hub API

### Service Layer
The `OrganizationService` uses the `ProjectRepository` to:
- Generate API keys → Store hash in `projects.api_key_hash`
- Validate API keys → Query by `api_key_hash`
- Revoke API keys → Clear both `api_key` and `api_key_hash`

### Middleware
The `AuthMiddleware` calls `OrganizationService.ValidateAPIKey()`, which queries this database.

### Data Flow

```
HTTP Request with API Key
        │
        ▼
AuthMiddleware
        │
        ▼
OrganizationService.ValidateAPIKey()
        │
        ▼
ProjectRepository.FindByAPIKeyHash()
        │
        ▼
PostgreSQL Database (projects table)
        │
        ▼
Return Project (with org_id, project_id)
        │
        ▼
Inject into Context → Continue Request
```

## Conclusion

**✅ This migration is correctly scoped:**
- Applies to the Hub's PostgreSQL database
- Part of the Hub's core schema
- Required for the Hub's authentication system
- Follows the existing migration pattern

**✅ The database architecture is correct:**
- Docker-based deployment
- Hub API owns the database
- Centralized authentication
- Proper separation of concerns

The migration should be applied to the database running in the Hub's Docker Compose stack.
