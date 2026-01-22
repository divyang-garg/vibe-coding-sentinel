# API Key Management Guide

**Date:** January 21, 2026  
**Status:** Production Ready

---

## Overview

This guide explains how to generate, manage, and use API keys for the Sentinel Hub API.

---

## Quick Start

### 1. Create a Project (Auto-generates API Key)

When you create a project, an API key is automatically generated:

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-admin-key" \
  -d '{"name": "My Project"}'
```

**Response:**
```json
{
  "id": "proj_123",
  "name": "My Project",
  "api_key": "xK9mP2qR7vT4wY8zA1bC3dE5fG6hI0j",  // ⚠️ SAVE THIS!
  "api_key_prefix": "xK9mP2qR",
  "created_at": "2026-01-21T12:00:00Z"
}
```

**Important:** The `api_key` is only returned once. Save it immediately!

---

## API Key Management Endpoints

### Generate New API Key

**When to use:** When you need to regenerate an API key (e.g., if it was compromised or lost).

**Endpoint:** `POST /api/v1/projects/{id}/api-key`

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/projects/proj_123/api-key \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-admin-key"
```

**Response:**
```json
{
  "api_key": "new-key-here",
  "api_key_prefix": "new-key",
  "message": "API key generated successfully. Save this key - it will not be shown again.",
  "warning": "This is the only time you will see this key. Store it securely."
}
```

**Note:** The old API key is automatically revoked when a new one is generated.

---

### Get API Key Information

**When to use:** To check if a project has an API key and see the prefix (for identification).

**Endpoint:** `GET /api/v1/projects/{id}/api-key`

**Example:**
```bash
curl -X GET http://localhost:8080/api/v1/projects/proj_123/api-key \
  -H "X-API-Key: your-admin-key"
```

**Response:**
```json
{
  "has_api_key": true,
  "api_key_prefix": "xK9mP2qR",
  "message": "Full API key is never returned for security reasons. Use POST to generate a new key."
}
```

**Security Note:** The full API key is never returned. Only the prefix is shown for identification purposes.

---

### Revoke API Key

**When to use:** When you need to immediately invalidate an API key (e.g., security breach, key compromise).

**Endpoint:** `DELETE /api/v1/projects/{id}/api-key`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/projects/proj_123/api-key \
  -H "X-API-Key: your-admin-key"
```

**Response:**
```json
{
  "message": "API key revoked successfully"
}
```

**Note:** After revocation, all requests using that API key will fail with `401 Unauthorized`.

---

## Complete Workflow Example

```bash
# 1. Create a project (auto-generates API key)
PROJECT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "X-API-Key: admin-key" \
  -d '{"name": "My Project"}')

PROJECT_ID=$(echo $PROJECT_RESPONSE | jq -r '.id')
INITIAL_KEY=$(echo $PROJECT_RESPONSE | jq -r '.api_key')

echo "Project ID: $PROJECT_ID"
echo "Initial API Key: $INITIAL_KEY"
# ⚠️ Save $INITIAL_KEY securely!

# 2. Get API key info (prefix only)
curl -X GET http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key"

# 3. Generate a new API key (revokes the old one)
NEW_KEY_RESPONSE=$(curl -s -X POST \
  http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key")

NEW_KEY=$(echo $NEW_KEY_RESPONSE | jq -r '.api_key')
echo "New API Key: $NEW_KEY"
# ⚠️ Save $NEW_KEY securely!

# 4. Verify old key no longer works
curl -X GET http://localhost:8080/api/v1/projects/$PROJECT_ID \
  -H "X-API-Key: $INITIAL_KEY"
# Should return 401 Unauthorized

# 5. Verify new key works
curl -X GET http://localhost:8080/api/v1/projects/$PROJECT_ID \
  -H "X-API-Key: $NEW_KEY"
# Should return 200 OK with project data

# 6. Revoke the API key
curl -X DELETE http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key"

# 7. Verify key is revoked
curl -X GET http://localhost:8080/api/v1/projects/$PROJECT_ID \
  -H "X-API-Key: $NEW_KEY"
# Should return 401 Unauthorized
```

---

## Security Best Practices

### 1. Store Keys Securely

- **Never commit API keys to version control**
- **Use environment variables or secret management systems**
- **Rotate keys regularly**
- **Use different keys for different environments (dev, staging, prod)**

### 2. Key Rotation

```bash
# Rotate keys periodically (e.g., every 90 days)
# 1. Generate new key
NEW_KEY=$(curl -s -X POST \
  http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key" | jq -r '.api_key')

# 2. Update your application configuration
export SENTINEL_API_KEY="$NEW_KEY"

# 3. Verify new key works
curl -X GET http://localhost:8080/api/v1/projects/$PROJECT_ID \
  -H "X-API-Key: $NEW_KEY"

# 4. Old key is automatically revoked
```

### 3. Key Revocation

If a key is compromised:

```bash
# Immediately revoke the key
curl -X DELETE http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key"

# Generate a new key
NEW_KEY=$(curl -s -X POST \
  http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key" | jq -r '.api_key')

# Update all systems using the old key
```

### 4. Key Identification

Use the prefix to identify keys without exposing the full key:

```bash
# Get key info
curl -X GET http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key"

# Response shows prefix: "xK9mP2qR"
# You can match this with your stored keys to identify which key belongs to which project
```

---

## Common Scenarios

### Scenario 1: Lost API Key

**Problem:** You lost the API key and can't authenticate.

**Solution:**
1. Use an admin API key to generate a new key
2. Save the new key securely
3. Update your application configuration

```bash
# Generate new key
curl -X POST http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key"
```

### Scenario 2: Key Compromise

**Problem:** API key was exposed or compromised.

**Solution:**
1. Immediately revoke the compromised key
2. Generate a new key
3. Update all systems

```bash
# Revoke compromised key
curl -X DELETE http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key"

# Generate new key
curl -X POST http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key"
```

### Scenario 3: Key Rotation

**Problem:** Need to rotate keys for security compliance.

**Solution:**
1. Generate new key (automatically revokes old one)
2. Update application configuration
3. Verify new key works
4. Monitor for any issues

```bash
# Generate new key
NEW_KEY=$(curl -s -X POST \
  http://localhost:8080/api/v1/projects/$PROJECT_ID/api-key \
  -H "X-API-Key: admin-key" | jq -r '.api_key')

# Update configuration
export SENTINEL_API_KEY="$NEW_KEY"

# Verify
curl -X GET http://localhost:8080/api/v1/projects/$PROJECT_ID \
  -H "X-API-Key: $NEW_KEY"
```

---

## Error Handling

### Common Errors

**401 Unauthorized**
- Invalid or missing API key
- Key has been revoked
- Solution: Check API key or generate a new one

**404 Not Found**
- Project ID doesn't exist
- Solution: Verify project ID

**400 Bad Request**
- Missing required parameters
- Invalid request format
- Solution: Check request format and parameters

**500 Internal Server Error**
- Server-side error
- Solution: Contact support or check server logs

---

## Related Documentation

- `docs/api/API_REFERENCE.md` - Complete API reference
- `docs/API_KEY_IMPLEMENTATION_FLOW.md` - Technical implementation details
- `API_KEY_ENDPOINTS_IMPLEMENTATION.md` - Endpoint implementation details

---

## Support

For issues or questions:
1. Check the error message and status code
2. Verify your API key is valid and not revoked
3. Check the project ID is correct
4. Review the related documentation
5. Contact support if issues persist
