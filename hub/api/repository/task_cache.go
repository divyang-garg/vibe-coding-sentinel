// Phase 14E: Task Cache Layer
// In-memory caching for task management
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package repository

import (
	"sync"

	"sentinel-hub-api/models"
)

// Task is an alias to models.Task for local use
type Task = models.Task

// Simple in-memory cache for tasks
var (
	taskCache   = make(map[string]*models.Task)
	taskCacheMu sync.RWMutex
)

// GetCachedTask retrieves a task from cache
func GetCachedTask(taskID string) (*models.Task, bool) {
	taskCacheMu.RLock()
	defer taskCacheMu.RUnlock()
	task, found := taskCache[taskID]
	return task, found
}

// SetCachedTask stores a task in cache
func SetCachedTask(taskID string, task *models.Task) {
	if task == nil {
		return
	}
	taskCacheMu.Lock()
	defer taskCacheMu.Unlock()
	taskCache[taskID] = task
}

// ClearTaskCache clears the task cache
func ClearTaskCache() {
	taskCacheMu.Lock()
	defer taskCacheMu.Unlock()
	taskCache = make(map[string]*models.Task)
}

// InvalidateTaskCache removes a task from cache
func InvalidateTaskCache(taskID string) {
	taskCacheMu.Lock()
	defer taskCacheMu.Unlock()
	delete(taskCache, taskID)
}
