# Applying Migration 006: Knowledge Sync Metadata

## Quick Apply

If you have PostgreSQL running and the database configured:

```bash
# Option 1: Using DATABASE_URL environment variable
export DATABASE_URL="postgres://user:password@localhost:5432/sentinel?sslmode=disable"
psql "$DATABASE_URL" -f hub/migrations/006_add_knowledge_sync_metadata.sql

# Option 2: Using individual environment variables
export DB_HOST=localhost
export DB_USER=sentinel
export DB_PASSWORD=sentinel
export DB_NAME=sentinel
export DB_PORT=5432
psql "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" -f hub/migrations/006_add_knowledge_sync_metadata.sql

# Option 3: Using the migration script
./scripts/apply_migration_006.sh

# Option 4: Direct connection (replace with your credentials)
psql -h localhost -U sentinel -d sentinel -f hub/migrations/006_add_knowledge_sync_metadata.sql
```

## Prerequisites

1. **PostgreSQL must be running**
   ```bash
   # Check if PostgreSQL is running
   pg_isready
   # or
   psql --version
   ```

2. **Database must exist**
   ```sql
   -- Connect to PostgreSQL
   psql -U postgres
   
   -- Create database if it doesn't exist
   CREATE DATABASE sentinel;
   
   -- Create user if it doesn't exist
   CREATE USER sentinel WITH PASSWORD 'sentinel';
   
   -- Grant privileges
   GRANT ALL PRIVILEGES ON DATABASE sentinel TO sentinel;
   \q
   ```

3. **Previous migrations must be applied**
   - Migration 002 (`002_create_core_tables.sql`) must be applied first
   - This creates the `knowledge_items` table

## Verification

After applying the migration, verify it was successful:

```sql
-- Connect to database
psql -h localhost -U sentinel -d sentinel

-- Check columns exist
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns 
WHERE table_name = 'knowledge_items' 
  AND column_name IN ('updated_at', 'approved_by', 'approved_at', 'last_synced_at', 'sync_version', 'sync_status')
ORDER BY column_name;

-- Check indexes exist
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'knowledge_items' 
  AND indexname LIKE 'idx_knowledge%'
ORDER BY indexname;

-- Check existing data
SELECT 
    COUNT(*) as total_items,
    COUNT(updated_at) as items_with_updated_at,
    COUNT(last_synced_at) as items_with_sync_time,
    COUNT(sync_version) as items_with_sync_version
FROM knowledge_items;

\q
```

## Expected Output

After successful migration, you should see:
- 6 new columns added to `knowledge_items` table
- 4 new indexes created
- Existing rows updated with default values

## Troubleshooting

### Error: "relation knowledge_items does not exist"
- Run migration 002 first: `psql -h localhost -U sentinel -d sentinel -f hub/migrations/002_create_core_tables.sql`

### Error: "permission denied"
- Grant privileges: `GRANT ALL PRIVILEGES ON DATABASE sentinel TO sentinel;`
- Or run as postgres superuser

### Error: "column already exists"
- This is safe - the migration uses `IF NOT EXISTS`
- The migration can be run multiple times safely

### Error: "database does not exist"
- Create the database first (see Prerequisites)

## Rollback (if needed)

If you need to rollback this migration:

```sql
-- Connect to database
psql -h localhost -U sentinel -d sentinel

-- Drop indexes
DROP INDEX IF EXISTS idx_knowledge_last_synced;
DROP INDEX IF EXISTS idx_knowledge_sync_status;
DROP INDEX IF EXISTS idx_knowledge_type_status;
DROP INDEX IF EXISTS idx_knowledge_status;

-- Drop columns
ALTER TABLE knowledge_items 
DROP COLUMN IF EXISTS sync_status,
DROP COLUMN IF EXISTS sync_version,
DROP COLUMN IF EXISTS last_synced_at,
DROP COLUMN IF EXISTS approved_at,
DROP COLUMN IF EXISTS approved_by,
DROP COLUMN IF EXISTS updated_at;
```

**Note:** Rollback will lose sync metadata. Only rollback if absolutely necessary.
