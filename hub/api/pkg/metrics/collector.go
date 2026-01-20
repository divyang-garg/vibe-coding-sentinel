// Package metrics provides Prometheus metrics
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package metrics

import (
	"runtime"
	"time"

	"sentinel-hub-api/pkg"
)

// StartSystemMetricsCollection starts a goroutine that periodically collects system metrics
func StartSystemMetricsCollection(m *Metrics) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if pkg.IsShuttingDown() {
			return
		}

		// Update goroutine count
		m.GoroutineCount.Set(float64(runtime.NumGoroutine()))

		// Update memory usage
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		m.MemoryUsage.Set(float64(memStats.Alloc))
	}
}
