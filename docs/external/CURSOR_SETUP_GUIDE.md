# Cursor IDE Setup Guide

This guide explains how to configure Cursor IDE to use Sentinel's MCP (Model Context Protocol) server for enhanced AI-assisted development.

## Overview

Sentinel provides 19 MCP tools that integrate directly with Cursor IDE, allowing Cursor to:
- Access business context and security rules
- Validate code before generation
- Check task status and dependencies
- Get test requirements
- Analyze code quality and security

## Prerequisites

- Cursor IDE installed
- Sentinel binary built (`./synapsevibsentinel.sh`)
- Hub API configured (if using Hub features)
- Environment variables set:
  - `SENTINEL_HUB_URL` (optional, for Hub features)
  - `SENTINEL_API_KEY` (optional, for Hub features)

## Configuration Steps

### Step 1: Locate Cursor Configuration Directory

**macOS:**
```
~/.cursor/mcp.json
```

**Linux:**
```
~/.config/cursor/mcp.json
```

**Windows:**
```
%APPDATA%\Cursor\mcp.json
```

### Step 2: Create or Edit MCP Configuration

Create or edit the MCP configuration file with the following content:

```json
{
  "mcpServers": {
    "sentinel": {
      "command": "/absolute/path/to/sentinel",
      "args": ["mcp-server"],
      "env": {
        "SENTINEL_HUB_URL": "https://your-hub-url.com",
        "SENTINEL_API_KEY": "your-api-key"
      }
    }
  }
}
```

**Important Notes:**
- Use the **absolute path** to the Sentinel binary
- On Windows, use forward slashes or escaped backslashes: `C:/path/to/sentinel.exe`
- Environment variables are optional if not using Hub features
- If Hub URL is not set, Sentinel will work in local-only mode

### Step 3: Find Sentinel Binary Path

**macOS/Linux:**
```bash
which sentinel
# or
realpath ./sentinel
```

**Windows:**
```cmd
where sentinel.exe
```

**Example paths:**
- macOS: `/Users/username/projects/VicecodingSentinel/sentinel`
- Linux: `/home/username/projects/VicecodingSentinel/sentinel`
- Windows: `C:\Users\username\projects\VicecodingSentinel\sentinel.exe`

### Step 4: Restart Cursor IDE

After saving the configuration file:
1. Close Cursor IDE completely
2. Reopen Cursor IDE
3. Cursor will automatically load the MCP server

### Step 5: Verify MCP Server Connection

1. Open Cursor IDE
2. Open the Cursor Settings (Cmd/Ctrl + ,)
3. Navigate to "Features" → "MCP Servers"
4. Verify "sentinel" appears in the list
5. Check that the status shows "Connected" or "Ready"

Alternatively, check Cursor's developer console:
- macOS: `Cmd + Shift + P` → "Developer: Toggle Developer Tools"
- Look for MCP-related messages in the console

## Available MCP Tools

Once configured, Cursor can use these Sentinel MCP tools:

### Knowledge & Context Tools
- `sentinel_get_business_context` - Get business rules and context
- `sentinel_get_knowledge_items` - Query knowledge base
- `sentinel_search_knowledge` - Search knowledge items

### Security Tools
- `sentinel_get_security_context` - Get security rules and compliance status
- `sentinel_validate_code` - Validate code using AST analysis
- `sentinel_apply_fix` - Apply security/style/performance fixes

### Test Tools
- `sentinel_get_test_requirements` - Get test requirements and coverage
- `sentinel_validate_test` - Validate test file quality

### Task Management Tools
- `sentinel_get_task_status` - Get task status and details
- `sentinel_verify_task` - Verify task completion
- `sentinel_list_tasks` - List all tasks

### Analysis Tools
- `sentinel_analyze_intent` - Analyze user intent from prompts
- `sentinel_check_file_size` - Check file size and get split suggestions

## Troubleshooting

### MCP Server Not Connecting

**Issue:** Sentinel MCP server doesn't appear in Cursor

**Solutions:**
1. Verify binary path is absolute and correct
2. Check file permissions: `chmod +x sentinel` (Unix)
3. Test binary manually: `./sentinel mcp-server` (should start and wait for input)
4. Check Cursor logs for errors

### Permission Denied

**Issue:** Cursor cannot execute Sentinel binary

**Solutions:**
- macOS/Linux: `chmod +x /path/to/sentinel`
- Windows: Ensure binary is not blocked by Windows Defender
- Verify user has execute permissions

### Hub Connection Errors

**Issue:** MCP tools that require Hub fail

**Solutions:**
1. Verify `SENTINEL_HUB_URL` is set correctly
2. Verify `SENTINEL_API_KEY` is valid
3. Test Hub connection: `curl -H "Authorization: Bearer $SENTINEL_API_KEY" $SENTINEL_HUB_URL/health`
4. Check network connectivity

### Binary Not Found

**Issue:** Cursor reports binary not found

**Solutions:**
1. Use absolute path, not relative
2. Verify binary exists: `ls -la /path/to/sentinel`
3. On Windows, include `.exe` extension
4. Check path for typos

### MCP Server Crashes

**Issue:** MCP server starts but crashes

**Solutions:**
1. Check Cursor console for error messages
2. Test binary manually: `echo '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{}}' | ./sentinel mcp-server`
3. Verify Go runtime is compatible
4. Check system logs for errors

## Testing MCP Server Manually

Test the MCP server directly:

```bash
# Send initialize request
echo '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./sentinel mcp-server

# Should return JSON-RPC response with server info
```

## Environment Variables

### Required (for Hub features)
- `SENTINEL_HUB_URL` - Hub API URL (e.g., `https://hub.example.com`)
- `SENTINEL_API_KEY` - API key for Hub authentication

### Optional
- `SENTINEL_LOG_LEVEL` - Log level (DEBUG, INFO, WARN, ERROR)
- `SENTINEL_LOG_FORMAT` - Log format (json, text)

## Configuration Examples

### Minimal Configuration (Local Only)

```json
{
  "mcpServers": {
    "sentinel": {
      "command": "/Users/john/projects/sentinel/sentinel",
      "args": ["mcp-server"]
    }
  }
}
```

### Full Configuration (With Hub)

```json
{
  "mcpServers": {
    "sentinel": {
      "command": "/Users/john/projects/sentinel/sentinel",
      "args": ["mcp-server"],
      "env": {
        "SENTINEL_HUB_URL": "https://hub.example.com",
        "SENTINEL_API_KEY": "sk-abc123...",
        "SENTINEL_LOG_LEVEL": "INFO"
      }
    }
  }
}
```

### Multiple MCP Servers

```json
{
  "mcpServers": {
    "sentinel": {
      "command": "/path/to/sentinel",
      "args": ["mcp-server"],
      "env": {
        "SENTINEL_HUB_URL": "https://hub.example.com",
        "SENTINEL_API_KEY": "your-key"
      }
    },
    "other-server": {
      "command": "/path/to/other",
      "args": ["server"]
    }
  }
}
```

## Verification Checklist

- [ ] MCP configuration file created
- [ ] Absolute path to Sentinel binary set
- [ ] Environment variables configured (if using Hub)
- [ ] Binary has execute permissions
- [ ] Cursor IDE restarted
- [ ] MCP server shows as connected in Cursor
- [ ] Test MCP tool call works (e.g., `sentinel_get_business_context`)

## Additional Resources

- [MCP Task Tools Guide](./MCP_TASK_TOOLS_GUIDE.md) - Detailed guide for task management tools
- [Hub Deployment Guide](./HUB_DEPLOYMENT_GUIDE.md) - Hub setup and configuration
- [Technical Specification](./TECHNICAL_SPEC.md) - Technical details

## Support

If you encounter issues:
1. Check Cursor's developer console for errors
2. Test Sentinel binary manually
3. Verify configuration file syntax (valid JSON)
4. Check file permissions and paths
5. Review Sentinel logs (if logging enabled)



