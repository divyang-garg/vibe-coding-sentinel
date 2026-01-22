# Test Database Setup Guide

## Overview

This guide explains how to set up the test database for running integration and migration tests.

## Quick Setup

Run the setup script:

```bash
./scripts/setup_test_db.sh
```

This will:
- Create the `sentinel` user with password `sentinel`
- Create the `sentinel_test` database
- Set up proper permissions
- Run all migration files to create tables

## Manual Setup

If you prefer to set up manually:

### 1. Create User and Database

```bash
psql -h localhost -p 5432 -U postgres -c "CREATE USER sentinel WITH PASSWORD 'sentinel';"
psql -h localhost -p 5432 -U postgres -c "ALTER USER sentinel CREATEDB;"
psql -h localhost -p 5432 -U postgres -c "CREATE DATABASE sentinel_test OWNER sentinel;"
```

### 2. Run Migrations

```bash
cd hub/migrations
psql -h localhost -p 5432 -U sentinel -d sentinel_test -f 001_create_hook_tables.sql
psql -h localhost -p 5432 -U sentinel -d sentinel_test -f 002_create_core_tables.sql
psql -h localhost -p 5432 -U sentinel -d sentinel_test -f 003_create_workflow_tables.sql
psql -h localhost -p 5432 -U sentinel -d sentinel_test -f 004_create_task_dependencies.sql
psql -h localhost -p 5432 -U sentinel -d sentinel_test -f 005_create_llm_usage_table.sql
psql -h localhost -p 5432 -U sentinel -d sentinel_test -f 001_add_api_key_hashing.sql
```

## Environment Variables

The test helpers use these environment variables (with defaults):

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=sentinel
export TEST_DB_PASSWORD=sentinel
export TEST_DB_NAME=sentinel_test
export TEST_DB_SSLMODE=disable
```

You can also source the `.env.test` file:

```bash
source .env.test
```

## Running Tests

After setup, run integration tests:

```bash
cd hub/api
go test ./services -run TestIntegrationSuite -v
go test ./database -run TestMigration -v
```

## Database Schema

The test database includes these tables:
- `organizations` - Organization management
- `projects` - Project management
- `tasks` - Task tracking
- `documents` - Document storage
- `knowledge_items` - Knowledge base
- `task_dependencies` - Task relationships
- `llm_usage` - LLM usage tracking
- `workflows` - Workflow definitions
- `workflow_executions` - Workflow execution history
- `hook_policies` - Git hook policies
- `hook_executions` - Hook execution logs
- `hook_baselines` - Hook baselines
- `error_reports` - Error reporting
- `llm_configurations` - LLM configuration

## Troubleshooting

### Connection Issues

If you get "password authentication failed":
1. Check PostgreSQL is running: `pg_isready -h localhost`
2. Verify user exists: `psql -h localhost -U postgres -c "\du sentinel"`
3. Reset password: `psql -h localhost -U postgres -c "ALTER USER sentinel WITH PASSWORD 'sentinel';"`

### Missing Tables

If tests fail with "relation does not exist":
1. Check migrations ran: `psql -h localhost -p 5432 -U sentinel -d sentinel_test -c "\dt"`
2. Re-run migrations in order (see Manual Setup above)

### Port Issues

The default port is 5432. If your PostgreSQL runs on a different port:
```bash
export TEST_DB_PORT=5433  # or your port
```

## Cleanup

To drop and recreate the test database:

```bash
psql -h localhost -p 5432 -U postgres -c "DROP DATABASE IF EXISTS sentinel_test;"
./scripts/setup_test_db.sh
```
