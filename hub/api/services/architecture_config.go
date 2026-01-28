// Package architecture_config - Configuration for architecture analysis
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

// DefaultArchitectureConfig returns default architecture thresholds.
func DefaultArchitectureConfig() ArchitectureConfig {
	return ArchitectureConfig{
		WarningLines:  300,
		CriticalLines: 500,
		MaxLines:      1000,
		MaxFanOut:     15,
	}
}

// GetArchitectureConfig returns architecture config (from service config or defaults).
func GetArchitectureConfig() ArchitectureConfig {
	cfg := GetConfig()
	if cfg != nil && cfg.Architecture.WarningLines > 0 {
		return cfg.Architecture
	}
	return DefaultArchitectureConfig()
}
