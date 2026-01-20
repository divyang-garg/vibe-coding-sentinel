// Package main - LLM cache (re-exports for backward compatibility)
// This file maintains backward compatibility by re-exporting functions from refactored modules.
// All implementation has been moved to:
//   - llm_cache_core.go: Core LLM caching operations
//   - llm_cache_analysis.go: Analysis result caching and progressive depth
//   - llm_cache_context.go: Business context caching
//   - llm_cache_metrics.go: Cache metrics tracking
//   - llm_cache_cleanup.go: Cache cleanup operations
//   - llm_cache_prompts.go: Prompt generation
//
// All types and functions are defined in the above files and are accessible
// from this package since they are all in package main.

package main

// Re-export types and functions for backward compatibility
// All types and functions are defined in llm_cache_core.go, llm_cache_analysis.go,
// llm_cache_context.go, llm_cache_metrics.go, llm_cache_cleanup.go, and llm_cache_prompts.go
