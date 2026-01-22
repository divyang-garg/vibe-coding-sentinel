#!/bin/bash
# Setup script for test database
# Creates sentinel user and sentinel_test database for integration tests

set -e

echo "🔧 Setting up test database for Sentinel..."

# Database configuration
DB_USER="sentinel"
DB_PASSWORD="sentinel"
DB_NAME="sentinel_test"
DB_HOST="localhost"
DB_PORT="5432"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if PostgreSQL is running
if ! pg_isready -h $DB_HOST -p $DB_PORT > /dev/null 2>&1; then
    echo -e "${RED}❌ PostgreSQL is not running on $DB_HOST:$DB_PORT${NC}"
    echo "Please start PostgreSQL and try again."
    exit 1
fi

echo -e "${GREEN}✅ PostgreSQL is running on $DB_HOST:$DB_PORT${NC}"

# Try to connect as postgres user (or current user if postgres doesn't exist)
if psql -h $DB_HOST -p $DB_PORT -U postgres -c "SELECT 1;" > /dev/null 2>&1; then
    ADMIN_USER="postgres"
elif psql -h $DB_HOST -p $DB_PORT -U $(whoami) -c "SELECT 1;" > /dev/null 2>&1; then
    ADMIN_USER=$(whoami)
else
    echo -e "${RED}❌ Cannot connect to PostgreSQL. Please check your PostgreSQL setup.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Connected as user: $ADMIN_USER${NC}"

# Create user if it doesn't exist
echo "Creating user '$DB_USER'..."
if psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'" | grep -q 1; then
    echo -e "${YELLOW}⚠️  User '$DB_USER' already exists${NC}"
    # Update password
    psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -c "ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" > /dev/null 2>&1
    echo -e "${GREEN}✅ Password updated for user '$DB_USER'${NC}"
else
    psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" > /dev/null 2>&1
    echo -e "${GREEN}✅ User '$DB_USER' created${NC}"
fi

# Grant necessary privileges
echo "Granting privileges..."
psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -c "ALTER USER $DB_USER CREATEDB;" > /dev/null 2>&1
echo -e "${GREEN}✅ Privileges granted${NC}"

# Drop database if it exists (for clean setup)
if psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo -e "${YELLOW}⚠️  Database '$DB_NAME' already exists${NC}"
    read -p "Do you want to drop and recreate it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -c "DROP DATABASE IF EXISTS $DB_NAME;" > /dev/null 2>&1
        echo -e "${GREEN}✅ Database '$DB_NAME' dropped${NC}"
    else
        echo -e "${YELLOW}⚠️  Keeping existing database${NC}"
    fi
fi

# Create database if it doesn't exist
if ! psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo "Creating database '$DB_NAME'..."
    psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" > /dev/null 2>&1
    echo -e "${GREEN}✅ Database '$DB_NAME' created${NC}"
fi

# Grant all privileges on database to user
psql -h $DB_HOST -p $DB_PORT -U $ADMIN_USER -d $DB_NAME -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" > /dev/null 2>&1

# Test connection
echo "Testing connection..."
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Connection test successful!${NC}"
else
    echo -e "${RED}❌ Connection test failed${NC}"
    exit 1
fi

# Set environment variable for tests
export TEST_DB_HOST=$DB_HOST
export TEST_DB_PORT=$DB_PORT
export TEST_DB_USER=$DB_USER
export TEST_DB_PASSWORD=$DB_PASSWORD
export TEST_DB_NAME=$DB_NAME
export TEST_DB_SSLMODE="disable"

echo ""
echo -e "${GREEN}✅ Test database setup complete!${NC}"
echo ""
echo "Database Configuration:"
echo "  Host:     $DB_HOST"
echo "  Port:     $DB_PORT"
echo "  User:     $DB_USER"
echo "  Database: $DB_NAME"
echo ""
echo "Connection String:"
echo "  postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
echo ""
echo "Environment Variables (add to your shell profile if needed):"
echo "  export TEST_DB_HOST=$DB_HOST"
echo "  export TEST_DB_PORT=$DB_PORT"
echo "  export TEST_DB_USER=$DB_USER"
echo "  export TEST_DB_PASSWORD=$DB_PASSWORD"
echo "  export TEST_DB_NAME=$DB_NAME"
echo "  export TEST_DB_SSLMODE=disable"
echo ""
echo -e "${GREEN}You can now run integration tests!${NC}"
