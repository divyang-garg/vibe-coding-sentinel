// Package services - Fix application service
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"

	"sentinel-hub-api/models"
)

// FixServiceImpl implements FixService interface
type FixServiceImpl struct {
	// Dependencies can be added here if needed
}

// NewFixService creates a new fix service
func NewFixService() FixService {
	return &FixServiceImpl{}
}

// ApplyFix applies fixes to code based on fix type
func (s *FixServiceImpl) ApplyFix(ctx context.Context, req models.ApplyFixRequest) (*models.ApplyFixResponse, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var fixedCode string
	var changes []map[string]interface{}
	var err error

	// Apply fixes based on type using the fix applier functions
	switch req.FixType {
	case "security":
		fixedCode, changes, err = applySecurityFixes(ctx, req.Code, req.Language)
	case "style":
		fixedCode, changes, err = applyStyleFixes(ctx, req.Code, req.Language)
	case "performance":
		fixedCode, changes, err = applyPerformanceFixes(ctx, req.Code, req.Language)
	default:
		return nil, fmt.Errorf("unsupported fix type: %s", req.FixType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to apply fixes: %w", err)
	}

	// Generate summary
	summary := fmt.Sprintf("Applied %d %s fixes", len(changes), req.FixType)

	return &models.ApplyFixResponse{
		FixedCode: fixedCode,
		Changes:   changes,
		Summary:   summary,
	}, nil
}

// Note: The actual fix application functions (applySecurityFixes, applyStyleFixes,
// applyPerformanceFixes) are implemented in fix_applier.go in this package.
