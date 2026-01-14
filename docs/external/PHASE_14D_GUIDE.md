# Phase 14D: Cost Optimization Guide

## Overview

Phase 14D implements advanced cost optimization features to reduce LLM API costs by up to 40% while maintaining analysis quality. The optimization system includes:

- **Enhanced Caching**: Comprehensive analysis results and business context caching
- **Progressive Depth**: Skip LLM calls for surface-level analysis
- **Smart Model Selection**: Automatically choose cheaper models when appropriate
- **Cost Limits**: Enforce maximum cost per request

## Features

### 1. Enhanced Caching System

The caching system stores comprehensive analysis results and business context to avoid redundant LLM calls.

#### Configuration

Enable caching in your LLM configuration:

```json
{
  "provider": "openai",
  "model": "gpt-4",
  "cost_optimization": {
    "use_cache": true,
    "cache_ttl_hours": 24
  }
}
```

- `use_cache`: Enable/disable caching (default: true)
- `cache_ttl_hours`: How long to cache results (default: 24 hours)

#### How It Works

1. **Analysis Result Caching**: When you run comprehensive analysis, the result is cached using a key based on:
   - Project ID
   - Feature name hash
   - Analysis depth
   - Analysis mode

2. **Business Context Caching**: Business rules, entities, and journeys are cached separately to avoid re-extraction.

3. **Cache Hit**: If a cached result exists and hasn't expired, it's returned immediately without LLM calls.

#### Cache Metrics

Monitor cache performance via the metrics endpoint:

```bash
GET /api/v1/metrics/cache?project_id={id}
```

Response:
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

### 2. Progressive Depth Analysis

Progressive depth allows you to control analysis depth and cost:

- **Surface Depth**: AST and pattern matching only, no LLM calls
- **Medium Depth**: Cheaper models (gpt-3.5-turbo, claude-3-haiku)
- **Deep Depth**: High-accuracy models (gpt-4, claude-3-opus)

#### Usage

```bash
POST /api/v1/analyze/comprehensive
{
  "feature": "user-authentication",
  "codebasePath": "/path/to/codebase",
  "depth": "surface",  // or "medium" or "deep"
  "mode": "auto"
}
```

#### When to Use Each Depth

- **Surface**: Quick checks, CI/CD pipelines, frequent validations
- **Medium**: Standard code reviews, non-critical features
- **Deep**: Security analysis, critical business logic, production releases

### 3. Smart Model Selection

The system automatically selects the most cost-effective model based on:

- Task criticality
- Analysis depth
- Cost limits
- User preferences

#### Model Selection Logic

1. **User Model**: If you specify a model, it's used (unless cost limit exceeded)
2. **Task Criticality**: Critical tasks (security, business rules) use expensive models
3. **Depth Consideration**: Medium depth + non-critical tasks use cheaper models
4. **Cost Limits**: If estimated cost exceeds limit, cheaper model is used

#### Model Cost Database

Current model costs (per 1K tokens):

**OpenAI:**
- gpt-4: $0.03
- gpt-3.5-turbo: $0.0015

**Anthropic:**
- claude-3-opus: $0.015
- claude-3-haiku: $0.00025

### 4. Cost Limits

Set maximum cost per request to prevent unexpected charges:

```json
{
  "cost_optimization": {
    "max_cost_per_request": 0.10  // $0.10 per request
  }
}
```

#### How It Works

1. Before making an LLM call, the system estimates cost based on prompt size
2. If estimated cost exceeds limit:
   - System tries cheaper model
   - If still exceeds, request is rejected with error
3. Cost estimation uses rough token count (1 token ≈ 4 characters)

### 5. Cost Metrics

Track cost savings and optimization effectiveness:

```bash
GET /api/v1/metrics/cost?project_id={id}&period=monthly
```

Response:
```json
{
  "success": true,
  "project_id": "uuid",
  "period": "monthly",
  "total_cost": 45.50,
  "cost_savings": 18.20,
  "savings_percentage": 40.0,
  "cache_hit_savings": 12.30,
  "model_selection_savings": 5.90,
  "total_requests": 150
}
```

## Best Practices

### 1. Cache Configuration

- **Enable caching** for production workloads
- **Set appropriate TTL** based on code change frequency:
  - Active development: 12-24 hours
  - Stable codebase: 48-72 hours
- **Monitor hit rate**: Aim for >70% cache hit rate

### 2. Depth Selection

- Use **surface depth** for:
  - Pre-commit hooks
  - CI/CD pipelines
  - Frequent validations
- Use **medium depth** for:
  - Standard code reviews
  - Non-critical features
- Use **deep depth** for:
  - Security audits
  - Production releases
  - Critical business logic

### 3. Cost Limits

- Set **realistic limits** based on your budget
- Start conservative and adjust based on usage
- Monitor cost metrics regularly
- Use cost alerts if available

### 4. Model Selection

- **Trust the system**: Smart model selection is optimized for cost/quality balance
- **Override when needed**: Specify model for critical analyses
- **Monitor savings**: Check cost metrics to see optimization impact

## Troubleshooting

### Low Cache Hit Rate

**Problem**: Cache hit rate is below 50%

**Solutions**:
1. Check if `use_cache` is enabled
2. Verify cache TTL is appropriate
3. Ensure feature/codebase hasn't changed significantly
4. Check cache size limits

### Cost Limits Too Restrictive

**Problem**: Requests frequently rejected due to cost limits

**Solutions**:
1. Increase `max_cost_per_request`
2. Use surface/medium depth for non-critical analyses
3. Enable caching to reduce LLM calls
4. Review prompt sizes

### Unexpected Model Selection

**Problem**: System using different model than expected

**Solutions**:
1. Check cost limits - may be forcing cheaper model
2. Verify task criticality classification
3. Review depth parameter
4. Check user-configured model override

## API Reference

### Cache Metrics Endpoint

```
GET /api/v1/metrics/cache?project_id={id}
```

**Query Parameters**:
- `project_id` (optional): Project ID (uses authenticated project if not provided)

**Response**:
- `hit_rate`: Cache hit rate (0.0 to 1.0)
- `total_hits`: Total cache hits
- `total_misses`: Total cache misses
- `cache_size`: Number of cached entries
- `cache_ttl_hours`: Cache TTL in hours

### Cost Metrics Endpoint

```
GET /api/v1/metrics/cost?project_id={id}&period={daily|weekly|monthly}
```

**Query Parameters**:
- `project_id` (optional): Project ID (uses authenticated project if not provided)
- `period` (optional): Time period (default: monthly)

**Response**:
- `total_cost`: Total LLM costs for period
- `cost_savings`: Estimated savings from optimization
- `savings_percentage`: Percentage saved
- `cache_hit_savings`: Savings from cache hits
- `model_selection_savings`: Savings from smart model selection

## Examples

### Example 1: Enable Caching

```bash
curl -X POST https://hub.example.com/api/v1/llm/config \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "api_key": "sk-...",
    "model": "gpt-4",
    "cost_optimization": {
      "use_cache": true,
      "cache_ttl_hours": 24,
      "max_cost_per_request": 0.10
    }
  }'
```

### Example 2: Surface Depth Analysis

```bash
curl -X POST https://hub.example.com/api/v1/analyze/comprehensive \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "feature": "user-login",
    "codebasePath": "/app",
    "depth": "surface",
    "mode": "auto"
  }'
```

### Example 3: Check Cost Metrics

```bash
curl -X GET "https://hub.example.com/api/v1/metrics/cost?project_id=uuid&period=monthly" \
  -H "Authorization: Bearer $TOKEN"
```

## Performance Impact

### Expected Savings

- **Cache Hit Rate**: 70%+ (saves 70% of LLM calls)
- **Model Selection**: 20-30% cost reduction
- **Progressive Depth**: 50-80% cost reduction for surface depth
- **Overall**: 40%+ total cost reduction

### Latency Impact

- **Cache Hit**: <10ms (vs 1-5s for LLM call)
- **Surface Depth**: <100ms (vs 1-5s for LLM call)
- **Cache Miss**: No additional latency

## Migration Guide

### From Phase 14C

If you're upgrading from Phase 14C:

1. **Enable caching** in existing LLM configs:
   ```json
   {
     "cost_optimization": {
       "use_cache": true,
       "cache_ttl_hours": 24
     }
   }
   ```

2. **Add depth parameter** to comprehensive analysis calls:
   ```json
   {
     "depth": "medium"  // Add this field
   }
   ```

3. **Set cost limits** if desired:
   ```json
   {
     "cost_optimization": {
       "max_cost_per_request": 0.10
     }
   }
   ```

4. **Monitor metrics** to verify optimization:
   - Check cache hit rate
   - Review cost savings
   - Adjust configuration as needed

## Support

For issues or questions:
- Check troubleshooting section above
- Review API reference
- Contact support with metrics data









