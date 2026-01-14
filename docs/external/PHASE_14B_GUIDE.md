# Phase 14B: MCP Integration for Comprehensive Analysis Guide

## Overview

Phase 14B integrates Phase 14A's comprehensive feature analysis into Cursor IDE via Model Context Protocol (MCP). This enables real-time feature analysis directly from Cursor, allowing developers to get comprehensive insights across all 7 layers (Business, UI, API, Database, Logic, Integration, Tests) without leaving their IDE.

## Features

- **MCP Server**: JSON-RPC 2.0 protocol implementation over stdio
- **Comprehensive Analysis Tool**: `sentinel_analyze_feature_comprehensive` tool for Cursor
- **Hub Integration**: Seamless communication with Sentinel Hub API
- **Error Handling**: Graceful fallback with helpful error messages
- **Parameter Validation**: Comprehensive validation of all inputs

## Setup

### Prerequisites

- Sentinel binary built (`./sentinel` exists)
- Hub API running and accessible (optional, but required for full functionality)
- Cursor IDE installed

### Cursor MCP Configuration

1. **Create or edit `~/.cursor/mcp.json`**:

```json
{
  "mcpServers": {
    "sentinel": {
      "command": "/path/to/sentinel",
      "args": ["mcp-server"],
      "env": {
        "SENTINEL_HUB_URL": "http://localhost:8080",
        "SENTINEL_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

2. **Replace placeholders**:
   - `/path/to/sentinel`: Full path to your Sentinel binary (e.g., `/Users/username/project/sentinel`)
   - `http://localhost:8080`: Your Hub API URL
   - `your-api-key-here`: Your Hub API key

3. **Restart Cursor IDE** to load the MCP configuration

### Environment Variables

You can also configure via environment variables (takes precedence over config file):

```bash
export SENTINEL_HUB_URL="http://localhost:8080"
export SENTINEL_API_KEY="your-api-key"
```

Or set in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Sentinel Hub Configuration
export SENTINEL_HUB_URL="http://localhost:8080"
export SENTINEL_API_KEY="your-api-key"
```

## Usage

### From Cursor IDE

Once configured, you can use the comprehensive analysis tool directly in Cursor chat:

**Example 1: Auto-discover feature**
```
Use sentinel_analyze_feature_comprehensive to analyze the user authentication feature
```

**Example 2: Specify analysis depth**
```
Analyze the payment processing feature with deep analysis using sentinel_analyze_feature_comprehensive
```

**Example 3: Include business context**
```
Use sentinel_analyze_feature_comprehensive to analyze order cancellation with business context validation
```

### Tool Parameters

The `sentinel_analyze_feature_comprehensive` tool accepts the following parameters:

- **feature** (required): Feature name or description (e.g., "Order Cancellation")
- **mode** (optional): "auto" or "manual" (default: "auto")
  - `auto`: Automatically discover feature components across all layers
  - `manual`: Use provided file specification
- **codebasePath** (optional): Path to codebase root (default: current working directory)
- **depth** (optional): Analysis depth level (default: "medium")
  - `surface`: Fast checks only (< 10 seconds)
  - `medium`: Balanced analysis with LLM caching (< 30 seconds)
  - `deep`: Comprehensive analysis (< 60 seconds)
- **includeBusinessContext** (optional): Include business rule validation (default: false)
- **files** (optional, for manual mode): Map of layer to file paths
  ```json
  {
    "ui": ["src/components/Profile.jsx"],
    "api": ["api/profile.go"],
    "database": ["migrations/001_profile.sql"]
  }
  ```

### Response Format

The tool returns a formatted response with:

- **validation_id**: Unique identifier for this analysis
- **feature**: Feature name analyzed
- **summary**: Summary statistics
  - `total_findings`: Total number of issues found
  - `by_severity`: Breakdown by severity (critical, high, medium, low)
  - `flows_verified`: Number of end-to-end flows verified
  - `flows_broken`: Number of broken flows
- **checklist**: Prioritized list of actionable items
- **hub_url**: URL to view detailed results in Hub dashboard
- **layer_summary**: Summary of findings per layer
- **flows**: End-to-end flow status

## Error Handling

### Common Errors

#### Hub Not Configured (Error Code: -32002)

**Cause**: `SENTINEL_HUB_URL` or `SENTINEL_API_KEY` not set

**Solution**:
1. Check your `~/.cursor/mcp.json` configuration
2. Verify environment variables are set
3. Ensure Hub API is running

#### Hub Unavailable (Error Code: -32000)

**Cause**: Hub API is not reachable or returned an error

**Solution**:
1. Verify Hub API is running: `curl http://localhost:8080/health`
2. Check network connectivity
3. Verify Hub URL is correct
4. Check Hub API logs for errors

**Fallback**: Use Cursor's default analysis or configure Hub connection

#### Hub Timeout (Error Code: -32001)

**Cause**: Analysis took too long (>60 seconds)

**Solution**:
- Try with `depth: "surface"` for faster results
- Check Hub API performance
- Verify codebase size is reasonable

#### Invalid Parameters (Error Code: -32602)

**Cause**: Missing or invalid tool parameters

**Solution**:
- Ensure `feature` parameter is provided
- Verify `codebasePath` exists if specified
- Check parameter types match expected format

#### Method Not Found (Error Code: -32601)

**Cause**: Unknown MCP method called

**Solution**: This should not happen with proper Cursor integration. If it does, check Cursor MCP configuration.

## Troubleshooting

### MCP Server Not Starting

**Symptoms**: Cursor shows error about MCP server

**Solutions**:
1. Verify Sentinel binary exists and is executable: `./sentinel --version`
2. Check binary path in `~/.cursor/mcp.json` is absolute and correct
3. Test MCP server manually:
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./sentinel mcp-server
   ```

### No Tools Available in Cursor

**Symptoms**: Cursor doesn't show `sentinel_analyze_feature_comprehensive` tool

**Solutions**:
1. Restart Cursor IDE after configuration changes
2. Check Cursor MCP logs for errors
3. Verify MCP server responds to `tools/list`:
   ```bash
   echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./sentinel mcp-server
   ```

### Analysis Always Fails

**Symptoms**: Tool calls always return errors

**Solutions**:
1. Verify Hub API is running and accessible
2. Check API key is valid
3. Test Hub API directly:
   ```bash
   curl -X POST http://localhost:8080/api/v1/analyze/comprehensive \
     -H "Authorization: Bearer $SENTINEL_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"feature":"test","mode":"auto","codebasePath":".","depth":"surface"}'
   ```

### Slow Analysis

**Symptoms**: Analysis takes too long

**Solutions**:
1. Use `depth: "surface"` for faster results
2. Check Hub API performance
3. Verify codebase size is reasonable
4. Consider using manual mode with specific files

## Examples

### Example 1: Quick Feature Check

**Request**:
```
Analyze the login feature with surface depth
```

**Cursor will call**:
```json
{
  "name": "sentinel_analyze_feature_comprehensive",
  "arguments": {
    "feature": "login",
    "depth": "surface"
  }
}
```

### Example 2: Deep Analysis with Business Context

**Request**:
```
Perform deep analysis of the payment processing feature including business context
```

**Cursor will call**:
```json
{
  "name": "sentinel_analyze_feature_comprehensive",
  "arguments": {
    "feature": "payment processing",
    "depth": "deep",
    "includeBusinessContext": true
  }
}
```

### Example 3: Manual File Specification

**Request**:
```
Analyze the user profile feature using these files: UI: src/components/Profile.jsx, API: api/profile.go
```

**Cursor will call**:
```json
{
  "name": "sentinel_analyze_feature_comprehensive",
  "arguments": {
    "feature": "user profile",
    "mode": "manual",
    "files": {
      "ui": ["src/components/Profile.jsx"],
      "api": ["api/profile.go"]
    }
  }
}
```

## Performance Considerations

### Analysis Time by Depth

- **Surface**: ~5-10 seconds (pattern-based only)
- **Medium**: ~30-60 seconds (includes LLM with caching)
- **Deep**: ~2-5 minutes (comprehensive LLM analysis)

### Optimization Tips

1. **Use surface depth** for quick checks during development
2. **Use medium depth** for regular analysis
3. **Use deep depth** only for critical features or before releases
4. **Enable LLM caching** in Hub configuration (Phase 14C) to reduce costs
5. **Use manual mode** with specific files for faster analysis of known components

## Security Considerations

### API Key Management

- **Never commit API keys** to version control
- **Use environment variables** or secure configuration files
- **Rotate API keys** regularly
- **Use different keys** for development and production

### Network Security

- **Use HTTPS** for Hub API in production
- **Validate SSL certificates** (automatic in Go)
- **Implement rate limiting** on Hub API (already configured)

## Related Documentation

- [Phase 14A Guide](./PHASE_14A_GUIDE.md) - Comprehensive analysis foundation
- [Comprehensive Analysis Solution](./COMPREHENSIVE_ANALYSIS_SOLUTION.md) - Solution specification
- [Technical Spec](./TECHNICAL_SPEC.md) - MCP protocol specification
- [Architecture](./ARCHITECTURE.md) - System architecture
- [User Guide](./USER_GUIDE.md) - General user documentation

## Support

For issues or questions:

1. Check this troubleshooting guide
2. Review Hub API logs
3. Test MCP server manually using examples above
4. Check Cursor MCP logs for client-side issues

---

**Document Version**: 1.0  
**Last Updated**: 2024-12-XX  
**Status**: Complete (Phase 14B)










