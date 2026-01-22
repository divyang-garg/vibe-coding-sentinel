# Database Migration Instructions
## API Key Hashing Migration

### Prerequisites

Before running the migration, ensure:

1. **PostgreSQL is installed and running**
   ```bash
   psql --version
   ```

2. **Database and user exist** (or create them):
   ```bash
   # Connect as postgres superuser
   psql -U postgres
   
   # Create database
   CREATE DATABASE sentinel;
   
   # Create user (if it doesn't exist)
   CREATE USER sentinel WITH PASSWORD 'password';
   
   # Grant privileges
   GRANT ALL PRIVILEGES ON DATABASE sentinel TO sentinel;
   \q
   ```

3. **Existing migrations are applied** (if this is a new setup, ensure core tables exist):
   ```bash
   # Apply core migrations first
   psql -h localhost -U sentinel -d sentinel -f hub/migrations/002_create_core_tables.sql
   ```

### Running the Migration

#### Option 1: Using psql directly
```bash
cd /Users/divyanggarg/VicecodingSentinel
psql -h localhost -U sentinel -d sentinel -f hub/migrations/001_add_api_key_hashing.sql
```

#### Option 2: Using environment variables
```bash
export PGPASSWORD=password
psql -h localhost -U sentinel -d sentinel -f hub/migrations/001_add_api_key_hashing.sql
unset PGPASSWORD
```

#### Option 3: Using connection string
```bash
psql "postgres://sentinel:password@localhost:5432/sentinel?sslmode=disable" -f hub/migrations/001_add_api_key_hashing.sql
```

### Verification

After running the migration, verify it was successful:

```sql
-- Connect to database
psql -h localhost -U sentinel -d sentinel

-- Check columns exist
\d projects

-- Should show:
-- - api_key (existing)
-- - api_key_hash (new)
-- - api_key_prefix (new)

-- Verify indexes were created
\di idx_projects_api_key_hash
\di idx_projects_api_key_prefix

-- Check if any existing keys were hashed
SELECT 
    COUNT(*) as total_projects,
    COUNT(api_key) as projects_with_plaintext_key,
    COUNT(api_key_hash) as projects_with_hashed_key
FROM projects;

-- Exit
\q
```

### Rollback (if needed)

If you need to rollback this migration:

```sql
-- Drop indexes
DROP INDEX IF EXISTS idx_projects_api_key_prefix;
DROP INDEX IF EXISTS idx_projects_api_key_hash;

-- Drop columns
ALTER TABLE projects 
DROP COLUMN IF EXISTS api_key_prefix,
DROP COLUMN IF EXISTS api_key_hash;
```

### Troubleshooting

**Error: "role sentinel does not exist"**
- Create the user: `CREATE USER sentinel WITH PASSWORD 'password';`

**Error: "database sentinel does not exist"**
- Create the database: `CREATE DATABASE sentinel;`

**Error: "permission denied"**
- Grant privileges: `GRANT ALL PRIVILEGES ON DATABASE sentinel TO sentinel;`
- Ensure the projects table exists (run core migrations first)

**Error: "relation projects does not exist"**
- Run the core table creation migration first:
  ```bash
  psql -h localhost -U sentinel -d sentinel -f hub/migrations/002_create_core_tables.sql
  ```

### Post-Migration Steps

1. **Verify existing API keys were hashed:**
   ```sql
   SELECT id, name, 
          CASE WHEN api_key_hash IS NOT NULL THEN 'hashed' ELSE 'not hashed' END as key_status
   FROM projects
   WHERE api_key IS NOT NULL;
   ```

2. **Test API key generation:**
   - Generate a new API key through the service
   - Verify it's stored as a hash in the database
   - Verify the plaintext key is NOT stored

3. **Test API key validation:**
   - Use the generated key to authenticate
   - Verify authentication works with hash-based lookup

### Production Notes

- This migration is **safe to run** on production with zero downtime
- Existing API keys will continue to work during migration
- New keys will use hash-based storage
- Old plaintext keys will be migrated on first use (via ValidateAPIKey fallback)
