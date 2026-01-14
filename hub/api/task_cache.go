// Phase 14E: Task System Caching
// Performance optimization through caching

package main

import (
	"sync"
	"time"
)

// TaskCacheEntry represents a cached task result
type TaskCacheEntry struct {
	Task      *Task
	Timestamp time.Time
	TTL       time.Duration
}

// TaskVerificationCacheEntry represents a cached verification result
type TaskVerificationCacheEntry struct {
	Verification *VerifyTaskResponse
	Timestamp    time.Time
	TTL          time.Duration
}

// TaskDependencyCacheEntry represents a cached dependency graph
type TaskDependencyCacheEntry struct {
	Dependencies *DependencyGraphResponse
	Timestamp    time.Time
	TTL          time.Duration
}

// Task cache storage
var (
	taskCache                  = make(map[string]*TaskCacheEntry)
	taskVerificationCache      = make(map[string]*TaskVerificationCacheEntry)
	taskDependencyCache        = make(map[string]*TaskDependencyCacheEntry)
	taskCacheMutex             sync.RWMutex
	taskVerificationCacheMutex sync.RWMutex
	taskDependencyCacheMutex   sync.RWMutex
)

// Cache TTL constants - now loaded from config
func getTaskCacheTTL() time.Duration {
	return GetConfig().Cache.TaskCacheTTL
}

func getTaskVerificationCacheTTL() time.Duration {
	return GetConfig().Cache.VerificationTTL
}

func getTaskDependencyCacheTTL() time.Duration {
	return GetConfig().Cache.DependencyTTL
}

// GetCachedTask retrieves a task from cache
func GetCachedTask(taskID string) (*Task, bool) {
	taskCacheMutex.RLock()
	defer taskCacheMutex.RUnlock()

	entry, exists := taskCache[taskID]
	if !exists {
		return nil, false
	}

	// Check if cache entry is still valid
	if time.Since(entry.Timestamp) > entry.TTL {
		return nil, false
	}

	return entry.Task, true
}

// SetCachedTask stores a task in cache
func SetCachedTask(taskID string, task *Task) {
	taskCacheMutex.Lock()
	defer taskCacheMutex.Unlock()

	taskCache[taskID] = &TaskCacheEntry{
		Task:      task,
		Timestamp: time.Now(),
		TTL:       getTaskCacheTTL(),
	}
}

// GetCachedVerification retrieves a verification result from cache
func GetCachedVerification(taskID string) (*VerifyTaskResponse, bool) {
	taskVerificationCacheMutex.RLock()
	defer taskVerificationCacheMutex.RUnlock()

	entry, exists := taskVerificationCache[taskID]
	if !exists {
		return nil, false
	}

	// Check if cache entry is still valid
	if time.Since(entry.Timestamp) > entry.TTL {
		return nil, false
	}

	return entry.Verification, true
}

// SetCachedVerification stores a verification result in cache
func SetCachedVerification(taskID string, verification *VerifyTaskResponse) {
	taskVerificationCacheMutex.Lock()
	defer taskVerificationCacheMutex.Unlock()

	taskVerificationCache[taskID] = &TaskVerificationCacheEntry{
		Verification: verification,
		Timestamp:    time.Now(),
		TTL:          getTaskVerificationCacheTTL(),
	}
}

// GetCachedDependencies retrieves dependencies from cache
func GetCachedDependencies(taskID string) (*DependencyGraphResponse, bool) {
	taskDependencyCacheMutex.RLock()
	defer taskDependencyCacheMutex.RUnlock()

	entry, exists := taskDependencyCache[taskID]
	if !exists {
		return nil, false
	}

	// Check if cache entry is still valid
	if time.Since(entry.Timestamp) > entry.TTL {
		return nil, false
	}

	return entry.Dependencies, true
}

// SetCachedDependencies stores dependencies in cache
func SetCachedDependencies(taskID string, dependencies *DependencyGraphResponse) {
	taskDependencyCacheMutex.Lock()
	defer taskDependencyCacheMutex.Unlock()

	taskDependencyCache[taskID] = &TaskDependencyCacheEntry{
		Dependencies: dependencies,
		Timestamp:    time.Now(),
		TTL:          getTaskDependencyCacheTTL(),
	}
}

// InvalidateTaskCache invalidates cache for a task
func InvalidateTaskCache(taskID string) {
	taskCacheMutex.Lock()
	defer taskCacheMutex.Unlock()

	delete(taskCache, taskID)
	delete(taskVerificationCache, taskID)
	delete(taskDependencyCache, taskID)
}

// CleanupTaskCache removes expired cache entries
func CleanupTaskCache() {
	taskCacheMutex.Lock()
	defer taskCacheMutex.Unlock()

	now := time.Now()
	for key, entry := range taskCache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(taskCache, key)
		}
	}

	taskVerificationCacheMutex.Lock()
	defer taskVerificationCacheMutex.Unlock()

	for key, entry := range taskVerificationCache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(taskVerificationCache, key)
		}
	}

	taskDependencyCacheMutex.Lock()
	defer taskDependencyCacheMutex.Unlock()

	for key, entry := range taskDependencyCache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(taskDependencyCache, key)
		}
	}
}

// StartTaskCacheCleanup starts background goroutine for cache cleanup
func StartTaskCacheCleanup() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			CleanupTaskCache()
		}
	}()
}
