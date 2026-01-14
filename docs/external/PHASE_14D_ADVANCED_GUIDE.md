# Phase 14D: Cost Optimization Advanced Guide

Advanced guide for optimizing LLM costs in Sentinel Hub, covering advanced caching strategies, progressive depth optimization, model selection tuning, and performance optimization.

## Overview

This guide covers advanced techniques for maximizing cost savings while maintaining analysis quality. It assumes familiarity with basic Phase 14D features covered in [PHASE_14D_GUIDE.md](./PHASE_14D_GUIDE.md).

## Table of Contents

1. [Advanced Caching Strategies](#advanced-caching-strategies)
2. [Progressive Depth Optimization](#progressive-depth-optimization)
3. [Model Selection Tuning](#model-selection-tuning)
4. [Cost Monitoring and Alerts](#cost-monitoring-and-alerts)
5. [Performance Optimization Tips](#performance-optimization-tips)
6. [Cost Analysis Examples](#cost-analysis-examples)

## Advanced Caching Strategies

### Cache Key Design

Cache keys are constructed from:
- Project ID
- Feature name hash
- Analysis depth
- Analysis mode
- Codebase hash (for code changes)

**Best Practice**: Use consistent feature names to maximize cache hits.

```json
{
  "feature": "user-authentication",  // Consistent naming
  "depth": "medium",
  "mode": "auto"
}
```

### Cache Warming

Pre-populate cache for frequently analyzed features:

```bash
# Warm cache for critical features
curl -X POST "http://localhost:8080/api/v1/analyze/comprehensive" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "feature": "user-authentication",
    "codebasePath": ".",
    "depth": "medium"
  }'
```

### Cache Invalidation Strategy

Cache is automatically invalidated when:
- TTL expires (default: 24 hours)
- Codebase changes detected
- Manual invalidation via API

**Manual Invalidation:**
```bash
# Clear cache for specific project
curl -X DELETE "http://localhost:8080/api/v1/cache/project/{projectId}" \
  -H "Authorization: Bearer $API_KEY"
```

### Cache Size Management

Monitor cache size and adjust TTL:

```bash
# Get cache metrics
curl -X GET "http://localhost:8080/api/v1/metrics/cache?project_id={id}" \
  -H "Authorization: Bearer $API_KEY"
```

**Response:**
```json
{
  "success": true,
  "project_id": "uuid",
  "hit_rate": 0.75,
  "total_hits": 150,
  "total_misses": 50,
  "cache_size": 250,
  "cache_ttl_hours": 24
}
```

**Optimization Tips:**
- Increase TTL for stable codebases (48-72 hours)
- Decrease TTL for frequently changing codebases (12 hours)
- Monitor cache size to prevent memory issues

## Progressive Depth Optimization

### Depth Selection Strategy

Choose depth based on use case:

#### Surface Depth (No LLM Calls)

**Use Cases:**
- CI/CD pipelines
- Pre-commit hooks
- Frequent validations
- Quick checks

**Cost**: $0 (AST analysis only)

**Example:**
```json
{
  "feature": "user-authentication",
  "depth": "surface",
  "mode": "auto"
}
```

#### Medium Depth (Cheaper Models)

**Use Cases:**
- Standard code reviews
- Non-critical features
- Development phase analysis

**Models Used**: gpt-3.5-turbo, claude-3-haiku

**Cost**: ~$0.001-0.01 per analysis

**Example:**
```json
{
  "feature": "user-authentication",
  "depth": "medium",
  "mode": "auto"
}
```

#### Deep Depth (Expensive Models)

**Use Cases:**
- Security analysis
- Critical business logic
- Production releases
- Final validation

**Models Used**: gpt-4, claude-3-opus

**Cost**: ~$0.01-0.10 per analysis

**Example:**
```json
{
  "feature": "user-authentication",
  "depth": "deep",
  "mode": "auto"
}
```

### Progressive Depth Workflow

1. **Start with Surface**: Quick validation
2. **Escalate to Medium**: If issues found
3. **Use Deep**: Only for critical issues

**Example Workflow:**
```bash
# Step 1: Surface analysis (free)
curl -X POST "http://localhost:8080/api/v1/analyze/comprehensive" \
  -d '{"feature": "auth", "depth": "surface"}'

# Step 2: If issues found, escalate to medium
curl -X POST "http://localhost:8080/api/v1/analyze/comprehensive" \
  -d '{"feature": "auth", "depth": "medium"}'

# Step 3: Only use deep for critical issues
curl -X POST "http://localhost:8080/api/v1/analyze/comprehensive" \
  -d '{"feature": "auth", "depth": "deep"}'
```

## Model Selection Tuning

### Cost Limit Configuration

Set maximum cost per request:

```json
{
  "cost_optimization": {
    "max_cost_per_request": 0.05,  // $0.05 max per request
    "use_cache": true,
    "progressive_depth": true
  }
}
```

**Behavior:**
- If estimated cost exceeds limit, cheaper model is selected
- Request fails if even cheapest model exceeds limit
- Cost limit applies per request, not cumulative

### Model Selection Logic

The system selects models based on:

1. **Task Criticality**: Critical tasks → expensive models
2. **Analysis Depth**: Deep depth → expensive models
3. **Cost Limits**: Enforce maximum cost
4. **Cache Availability**: Use cache if available

**Model Selection Priority:**
```
1. Check cache → Return cached result (if available)
2. Check cost limit → Select cheaper model if needed
3. Check depth → Select appropriate model for depth
4. Check criticality → Upgrade model for critical tasks
```

### Custom Model Selection

Override automatic selection:

```json
{
  "feature": "user-authentication",
  "depth": "deep",
  "mode": "manual",
  "model_override": {
    "provider": "openai",
    "model": "gpt-3.5-turbo"  // Force cheaper model
  }
}
```

## Cost Monitoring and Alerts

### Cost Metrics Endpoint

Monitor costs over time:

```bash
GET /api/v1/metrics/cost?project_id={id}&period=monthly
```

**Response:**
```json
{
  "success": true,
  "project_id": "uuid",
  "period": "monthly",
  "total_cost": 45.50,
  "cost_savings": 18.20,
  "savings_percentage": 28.6,
  "cache_hit_savings": 12.00,
  "model_selection_savings": 6.20,
  "total_requests": 500
}
```

### Cost Breakdown by Provider

```bash
GET /api/v1/llm/usage/cost-breakdown?project_id={id}&period=monthly
```

**Response:**
```json
{
  "success": true,
  "breakdown": {
    "openai": {
      "cost": 30.00,
      "requests": 300,
      "tokens": 150000
    },
    "anthropic": {
      "cost": 15.50,
      "requests": 200,
      "tokens": 80000
    }
  }
}
```

### Cost Trends

Track cost trends over time:

```bash
GET /api/v1/llm/usage/trends?project_id={id}&period=monthly
```

**Response:**
```json
{
  "success": true,
  "trends": [
    {
      "date": "2024-12-01",
      "cost": 1.50,
      "requests": 20
    },
    {
      "date": "2024-12-02",
      "cost": 2.00,
      "requests": 25
    }
  ]
}
```

### Setting Up Alerts

Configure cost alerts (requires webhook integration):

```json
{
  "alerts": {
    "daily_cost_limit": 10.00,
    "monthly_cost_limit": 200.00,
    "cost_increase_threshold": 0.20  // 20% increase triggers alert
  }
}
```

## Performance Optimization Tips

### 1. Maximize Cache Hits

**Strategy**: Use consistent feature names and analysis parameters

```json
// Good: Consistent naming
{"feature": "user-authentication"}
{"feature": "user-authentication"}  // Cache hit!

// Bad: Inconsistent naming
{"feature": "user-auth"}
{"feature": "authentication"}  // Cache miss
```

### 2. Use Surface Depth When Possible

**Strategy**: Start with surface depth, escalate only when needed

```bash
# Surface depth for quick checks
depth="surface"  # No LLM calls

# Medium depth for standard analysis
depth="medium"   # Cheaper models

# Deep depth only for critical analysis
depth="deep"     # Expensive models
```

### 3. Batch Analysis Requests

**Strategy**: Analyze multiple features in parallel

```bash
# Analyze multiple features
for feature in auth payment shipping; do
  curl -X POST "http://localhost:8080/api/v1/analyze/comprehensive" \
    -d "{\"feature\": \"$feature\", \"depth\": \"medium\"}" &
done
wait
```

### 4. Optimize Cache TTL

**Strategy**: Adjust TTL based on codebase stability

```json
{
  "cost_optimization": {
    "cache_ttl_hours": 48  // Increase for stable codebases
  }
}
```

### 5. Monitor and Adjust

**Strategy**: Regularly review cost metrics and adjust strategy

```bash
# Weekly cost review
curl -X GET "http://localhost:8080/api/v1/metrics/cost?project_id={id}&period=weekly"
```

## Cost Analysis Examples

### Example 1: Optimizing Cache Hit Rate

**Scenario**: Low cache hit rate (30%)

**Analysis:**
```bash
GET /api/v1/metrics/cache?project_id={id}
```

**Findings:**
- Hit rate: 0.30
- Total requests: 1000
- Cache hits: 300
- Cache misses: 700

**Optimization:**
1. Increase cache TTL from 24 to 48 hours
2. Use consistent feature naming
3. Pre-warm cache for frequently analyzed features

**Expected Result**: Hit rate increases to 60-70%

### Example 2: Reducing Model Costs

**Scenario**: High model costs ($50/month)

**Analysis:**
```bash
GET /api/v1/llm/usage/cost-breakdown?project_id={id}&period=monthly
```

**Findings:**
- gpt-4 usage: $40 (80% of cost)
- gpt-3.5-turbo usage: $10 (20% of cost)

**Optimization:**
1. Use progressive depth (surface → medium → deep)
2. Set cost limits to prefer cheaper models
3. Use medium depth for non-critical features

**Expected Result**: 40-50% cost reduction

### Example 3: Balancing Cost and Quality

**Scenario**: Need to reduce costs while maintaining quality

**Strategy:**
1. Use surface depth for CI/CD (free)
2. Use medium depth for development (cheap)
3. Use deep depth for production releases (expensive but necessary)

**Configuration:**
```json
{
  "cost_optimization": {
    "use_cache": true,
    "cache_ttl_hours": 48,
    "progressive_depth": true,
    "max_cost_per_request": 0.05
  }
}
```

**Result**: 60% cost reduction with maintained quality for critical paths

## Advanced Configuration

### Custom Cost Limits

Set different limits for different scenarios:

```json
{
  "cost_optimization": {
    "max_cost_per_request": 0.05,
    "max_cost_daily": 10.00,
    "max_cost_monthly": 200.00
  }
}
```

### Provider-Specific Limits

Set limits per provider:

```json
{
  "cost_optimization": {
    "provider_limits": {
      "openai": {
        "max_cost_per_request": 0.03
      },
      "anthropic": {
        "max_cost_per_request": 0.05
      }
    }
  }
}
```

### Cache Strategy Configuration

Fine-tune caching behavior:

```json
{
  "cost_optimization": {
    "use_cache": true,
    "cache_ttl_hours": 48,
    "cache_invalidation": {
      "on_code_change": true,
      "on_feature_change": true
    }
  }
}
```

## Troubleshooting

### High Costs Despite Caching

**Possible Causes:**
1. Cache TTL too short
2. Inconsistent feature naming
3. Codebase changes too frequently

**Solutions:**
1. Increase cache TTL
2. Standardize feature naming
3. Use codebase hash for cache keys

### Low Cache Hit Rate

**Possible Causes:**
1. Feature names inconsistent
2. Analysis parameters vary
3. Cache TTL too short

**Solutions:**
1. Use consistent naming conventions
2. Standardize analysis parameters
3. Increase cache TTL

### Cost Limits Not Working

**Possible Causes:**
1. Cost estimation inaccurate
2. Model selection override
3. Configuration not applied

**Solutions:**
1. Review cost estimation logic
2. Check for model overrides
3. Verify configuration is loaded

## Additional Resources

- [PHASE_14D_GUIDE.md](./PHASE_14D_GUIDE.md) - Basic Phase 14D guide
- [HUB_API_REFERENCE.md](./HUB_API_REFERENCE.md) - Complete API reference
- [MONITORING_GUIDE.md](./MONITORING_GUIDE.md) - Monitoring and observability









