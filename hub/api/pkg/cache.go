// Phase 12: Caching Module
// Provides caching for gap analysis results to improve performance

package pkg

import (
	"strings"
	"sync"
	"time"
)

// GapAnalysisReport represents gap analysis results (local definition for pkg)
type GapAnalysisReport struct {
	ProjectID string                 `json:"project_id"`
	Gaps      []Gap                  `json:"gaps"`
	Summary   map[string]interface{} `json:"summary"`
	CreatedAt string                 `json:"created_at"`
}

// Gap represents a single gap (local definition for pkg)
type Gap struct {
	Type            string                 `json:"type"`
	KnowledgeItemID string                 `json:"knowledge_item_id"`
	RuleTitle       string                 `json:"rule_title"`
	FilePath        string                 `json:"file_path,omitempty"`
	LineNumber      int                    `json:"line_number,omitempty"`
	Description     string                 `json:"description"`
	Evidence        map[string]interface{} `json:"evidence,omitempty"`
	Recommendation  string                 `json:"recommendation"`
	Severity        string                 `json:"severity"`
}

// CachedGapAnalysis represents a cached gap analysis result
type CachedGapAnalysis struct {
	Report    *GapAnalysisReport
	ExpiresAt time.Time
}

var (
	gapAnalysisCache = sync.Map{} // map[string]*CachedGapAnalysis
)

const GapAnalysisCacheTTL = 5 * time.Minute

// getCachedGapAnalysis retrieves a cached gap analysis if available and not expired
func getCachedGapAnalysis(projectID, codebasePath string) (*GapAnalysisReport, bool) {
	cacheKey := projectID + ":" + codebasePath

	if cached, ok := gapAnalysisCache.Load(cacheKey); ok {
		cachedAnalysis := cached.(*CachedGapAnalysis)
		if time.Now().Before(cachedAnalysis.ExpiresAt) {
			return cachedAnalysis.Report, true
		}
		// Expired, remove from cache
		gapAnalysisCache.Delete(cacheKey)
	}

	return nil, false
}

// setCachedGapAnalysis stores a gap analysis result in cache
func setCachedGapAnalysis(projectID, codebasePath string, report *GapAnalysisReport) {
	cacheKey := projectID + ":" + codebasePath
	cached := &CachedGapAnalysis{
		Report:    report,
		ExpiresAt: time.Now().Add(GapAnalysisCacheTTL),
	}
	gapAnalysisCache.Store(cacheKey, cached)
}

// invalidateGapAnalysisCache invalidates all cached gap analyses for a project
func invalidateGapAnalysisCache(projectID string) {
	gapAnalysisCache.Range(func(key, value interface{}) bool {
		cacheKey := key.(string)
		if strings.HasPrefix(cacheKey, projectID+":") {
			gapAnalysisCache.Delete(cacheKey)
		}
		return true
	})
}
