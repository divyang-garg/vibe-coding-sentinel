// Package llm provides rate limiting and quota management
// Complies with CODING_STANDARDS.md: Rate limiting max 250 lines
package llm

import (
	"math"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64 // Current token count
	capacity   float64 // Maximum tokens
	refillRate float64 // Tokens per second
	lastRefill time.Time
}

// QuotaManager tracks usage quotas and limits
type QuotaManager struct {
	mu            sync.RWMutex
	dailyUsage    map[string]int       // project -> tokens used today
	dailyLimits   map[string]int       // project -> daily token limit
	resetTimes    map[string]time.Time // project -> next reset time
	monthlyUsage  map[string]int       // project -> tokens used this month
	monthlyLimits map[string]int       // project -> monthly token limit
}

// NewRateLimiter creates a rate limiter with specified capacity and refill rate
func NewRateLimiter(capacity float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request can proceed and consumes a token
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens = math.Min(rl.capacity, rl.tokens+(elapsed*rl.refillRate))
	rl.lastRefill = now

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// TokensRemaining returns the current token count
func (rl *RateLimiter) TokensRemaining() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens = math.Min(rl.capacity, rl.tokens+(elapsed*rl.refillRate))
	rl.lastRefill = now

	return rl.tokens
}

// NewQuotaManager creates a quota manager for tracking usage limits
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		dailyUsage:    make(map[string]int),
		dailyLimits:   make(map[string]int),
		resetTimes:    make(map[string]time.Time),
		monthlyUsage:  make(map[string]int),
		monthlyLimits: make(map[string]int),
	}
}

// CheckQuota verifies if a request can proceed within quota limits
func (qm *QuotaManager) CheckQuota(projectID string, tokens int) bool {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// Check daily limit
	if dailyLimit, exists := qm.dailyLimits[projectID]; exists {
		if qm.dailyUsage[projectID]+tokens > dailyLimit {
			return false
		}
	}

	// Check monthly limit
	if monthlyLimit, exists := qm.monthlyLimits[projectID]; exists {
		if qm.monthlyUsage[projectID]+tokens > monthlyLimit {
			return false
		}
	}

	return true
}

// RecordUsage records token usage for quota tracking
func (qm *QuotaManager) RecordUsage(projectID string, tokens int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.dailyUsage[projectID] += tokens
	qm.monthlyUsage[projectID] += tokens
}

// SetLimits configures quota limits for a project
func (qm *QuotaManager) SetLimits(projectID string, dailyLimit int, monthlyLimit int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.dailyLimits[projectID] = dailyLimit
	qm.monthlyLimits[projectID] = monthlyLimit
	qm.resetTimes[projectID] = time.Now().Add(24 * time.Hour)
}

// ResetQuotas resets daily quotas (should be called daily)
func (qm *QuotaManager) ResetQuotas() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	now := time.Now()
	for projectID, resetTime := range qm.resetTimes {
		if now.After(resetTime) {
			qm.dailyUsage[projectID] = 0
			qm.resetTimes[projectID] = now.Add(24 * time.Hour)
		}
	}
}

// GetUsage returns current usage statistics
func (qm *QuotaManager) GetUsage(projectID string) (dailyUsed int, dailyLimit int, monthlyUsed int, monthlyLimit int) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	dailyUsed = qm.dailyUsage[projectID]
	dailyLimit = qm.dailyLimits[projectID]
	monthlyUsed = qm.monthlyUsage[projectID]
	monthlyLimit = qm.monthlyLimits[projectID]

	return
}
