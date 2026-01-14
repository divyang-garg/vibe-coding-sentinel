# Telemetry Granularity Feature

**Status**: ✅ Implemented (Phase 9.5)  
**Purpose**: Track all hook interactions with detailed events while maintaining performance through batching and aggregation.

---

## Problem Statement

**Current Limitation**: The existing telemetry system only tracks `audit_complete` events, providing high-level metrics but missing critical details about:
- Hook execution patterns
- User interaction choices
- Override frequencies and reasons
- Baseline addition patterns
- Performance metrics (duration)

**Impact**: Organizations cannot:
- Measure hook effectiveness
- Identify problematic patterns
- Track compliance improvements
- Make data-driven decisions about hook policies

---

## Solution Overview

**Two-Tier Telemetry System**:
1. **Detailed Events**: Stored locally, sent in batches
2. **Summary Events**: Sent immediately to Hub
3. **Hub Aggregation**: Daily/weekly/monthly summaries for dashboard

---

## Event Schemas

### Hook Execution Event

**Event Type**: `hook_execution`

**Schema**:
```json
{
  "agent_id": "uuid",
  "org_id": "uuid",
  "hook_type": "pre-commit" | "pre-push",
  "result": "blocked" | "allowed" | "overridden",
  "override_reason": "string (optional)",
  "findings_summary": {
    "total": 5,
    "critical": 0,
    "warning": 3,
    "info": 2
  },
  "user_actions": ["viewed_details", "proceeded_anyway"],
  "duration_ms": 1234,
  "timestamp": "2024-12-06T10:00:00Z"
}
```

**Fields**:
- `agent_id`: Unique identifier for the agent
- `org_id`: Organization ID (from Hub config)
- `hook_type`: Type of hook that executed
- `result`: Final outcome of hook execution
- `override_reason`: Reason for override (if applicable)
- `findings_summary`: Aggregated findings counts
- `user_actions`: Array of actions user took
- `duration_ms`: Total execution time in milliseconds
- `timestamp`: ISO 8601 timestamp

### Hook Baseline Event

**Event Type**: `hook_baseline` (Future - Phase 9.5C)

**Schema**:
```json
{
  "agent_id": "uuid",
  "org_id": "uuid",
  "baseline_entry": {
    "file": "src/file.js",
    "line": 123,
    "pattern": "console.log",
    "reason": "Added from hook",
    "date": "2024-12-06T10:00:00Z"
  },
  "source": "hook",
  "hook_type": "pre-commit",
  "reviewed": false
}
```

### Hook Override Event

**Event Type**: `hook_override` (Future - Phase 9.5C)

**Schema**:
```json
{
  "agent_id": "uuid",
  "org_id": "uuid",
  "hook_type": "pre-commit",
  "override_reason": "Temporary fix, will address in next PR",
  "findings_count": 3,
  "severity": "warning"
}
```

---

## Batching Strategy

### Local Queue

**Purpose**: Store events locally when Hub is unavailable

**Storage**: `.sentinel/telemetry-queue.json`

**Structure**:
```json
{
  "events": [
    {
      "event": "hook_execution",
      "timestamp": "...",
      "metrics": {...}
    }
  ]
}
```

**Flush Strategy**:
- Flush on next successful Hub connection
- Flush on agent shutdown
- Flush when queue reaches 100 events
- Flush every 5 minutes (background)

### Batch Sending

**Batch Size**: 10-50 events per request

**Format**:
```json
[
  {
    "event_type": "hook_execution",
    "payload": {...}
  }
]
```

**Error Handling**:
- Network errors: Queue for retry
- Server errors (4xx): Log and skip
- Server errors (5xx): Queue for retry

---

## Performance Considerations

### Async Sending

**Strategy**: Non-blocking telemetry sending

**Implementation**:
- Send telemetry in background goroutine
- Don't block hook execution
- Queue on failure

**Timeout**: 5 seconds per request

### Rate Limiting

**Client-Side**:
- Max 10 requests per minute
- Batch events to reduce requests

**Server-Side**:
- Rate limit: 100 requests per minute per agent
- Return 429 (Too Many Requests) on limit

### Caching

**Policy Cache**:
- Cache duration: 5 minutes
- Invalidate on policy update
- Fallback to default policy on error

---

## Hub Aggregation

### Daily Summaries

**Purpose**: Aggregate hook metrics by day

**Query**:
```sql
SELECT 
  DATE(created_at) as date,
  COUNT(*) as total_executions,
  COUNT(*) FILTER (WHERE result = 'blocked') as blocked_count,
  COUNT(*) FILTER (WHERE result = 'overridden') as overridden_count,
  AVG(duration_ms) as avg_duration_ms
FROM hook_executions
WHERE org_id = $1
  AND created_at >= $2
  AND created_at < $3
GROUP BY DATE(created_at)
ORDER BY date DESC
```

### Weekly/Monthly Summaries

**Similar aggregation** with appropriate date ranges

### Most Blocked Patterns

**Query**:
```sql
SELECT 
  findings_summary->>'pattern' as pattern,
  COUNT(*) as count
FROM hook_executions
WHERE result = 'blocked'
  AND org_id = $1
  AND created_at >= $2
GROUP BY pattern
ORDER BY count DESC
LIMIT 10
```

---

## Dashboard Integration

### Metrics Views

**Hook Effectiveness Dashboard**:
- Total executions (day/week/month)
- Block rate
- Override rate
- Average duration
- Most blocked patterns
- Team-level breakdown

**Team Hook Metrics**:
- Per-team execution counts
- Per-team override rates
- Per-team compliance trends

**Policy Insights**:
- Policy effectiveness
- Override patterns
- Baseline addition patterns

---

## Implementation Status

✅ **Phase 9.5A**: Basic hook telemetry implemented
- `sendHookTelemetry()` function
- Direct Hub endpoint integration
- Event schema defined

⏳ **Phase 9.5B**: Hub aggregation (in progress)
- Database schema created
- Metrics API endpoint implemented
- Aggregation queries defined

⏳ **Phase 9.5C**: Advanced tracking (planned)
- Baseline event tracking
- Override event tracking
- Policy compliance tracking

---

## Usage Examples

### Agent Side

**Automatic**: Hook telemetry is sent automatically when hooks execute

**Manual** (for testing):
```bash
# Not directly callable - integrated into hook execution
```

### Hub Side

**Query Metrics**:
```bash
curl -H "Authorization: Bearer $API_KEY" \
  "https://hub.example.com/api/v1/hooks/metrics?org_id=$ORG_ID&start_date=2024-12-01"
```

**Response**:
```json
{
  "total_executions": 150,
  "blocked_count": 12,
  "allowed_count": 120,
  "overridden_count": 18,
  "override_rate": 12.0,
  "avg_duration_ms": 2345.6,
  "most_blocked_patterns": [
    {"pattern": "console.log", "count": 5},
    {"pattern": "file_size", "count": 3}
  ]
}
```

---

## Future Enhancements

1. **Real-time Dashboard**: WebSocket updates for live metrics
2. **Alerting**: Notify on threshold breaches
3. **Export**: CSV/JSON export for external analysis
4. **Retention**: Configurable data retention policies
5. **Anonymization**: PII removal for compliance

---

## Related Documentation

- [Interactive Hooks Analysis](./INTERACTIVE_HOOKS_ANALYSIS.md)
- [Architecture](./ARCHITECTURE.md) - Telemetry section
- [Technical Spec](./TECHNICAL_SPEC.md) - Hook types and endpoints












