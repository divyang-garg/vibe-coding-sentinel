// Package services provides repository management business logic.
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"sentinel-hub-api/models"
)

// RepositoryServiceImpl implements RepositoryService
type RepositoryServiceImpl struct {
	// In production, this would integrate with actual Git hosting services
	// For now, we'll simulate repository data
	repositories map[string]*models.Repository
	nextID       int
}

// NewRepositoryService creates a new repository service instance
func NewRepositoryService() RepositoryService {
	return &RepositoryServiceImpl{
		repositories: make(map[string]*models.Repository),
		nextID:       1,
	}
}

// ListRepositories retrieves repositories based on criteria
func (s *RepositoryServiceImpl) ListRepositories(ctx context.Context, language string, limit int) ([]interface{}, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Generate mock repositories for demonstration
	repos := s.generateMockRepositories(language, limit)

	var result []interface{}
	for _, repo := range repos {
		result = append(result, map[string]interface{}{
			"id":             repo.ID,
			"name":           repo.Name,
			"full_name":      repo.FullName,
			"description":    repo.Description,
			"language":       repo.Language,
			"stars_count":    repo.StarsCount,
			"forks_count":    repo.ForksCount,
			"size_bytes":     repo.SizeBytes,
			"is_private":     repo.IsPrivate,
			"created_at":     repo.CreatedAt,
			"last_synced_at": repo.LastSyncedAt,
		})
	}

	return result, nil
}

// GetRepositoryImpact analyzes repository impact and relationships
func (s *RepositoryServiceImpl) GetRepositoryImpact(ctx context.Context, id string) (interface{}, error) {
	repo, exists := s.repositories[id]
	if !exists {
		// Generate mock data for non-existent repos
		repo = s.generateMockRepository(id)
	}

	impact := map[string]interface{}{
		"repository_id":          repo.ID,
		"name":                   repo.Name,
		"direct_dependencies":    rand.Intn(20) + 5,
		"dependent_repositories": rand.Intn(15) + 3,
		"impact_score":           rand.Float64()*10 + 85,
		"criticality_level":      s.calculateCriticality(repo),
		"change_risk_assessment": s.assessChangeRisk(repo),
		"recommended_reviewers":  []string{"architect", "lead-developer", "qa-lead"},
		"estimated_review_time":  "4-6 hours",
		"last_analyzed":          time.Now(),
	}

	return impact, nil
}

// GetRepositoryCentrality calculates repository centrality in the ecosystem
func (s *RepositoryServiceImpl) GetRepositoryCentrality(ctx context.Context, id string) (interface{}, error) {
	repo, exists := s.repositories[id]
	if !exists {
		repo = s.generateMockRepository(id)
	}

	centrality := map[string]interface{}{
		"repository_id":          repo.ID,
		"name":                   repo.Name,
		"degree_centrality":      rand.Float64()*0.8 + 0.2,
		"betweenness_centrality": rand.Float64()*0.6 + 0.1,
		"closeness_centrality":   rand.Float64()*0.7 + 0.3,
		"eigenvector_centrality": rand.Float64()*0.9 + 0.1,
		"overall_centrality":     rand.Float64() * 100,
		"connection_count":       rand.Intn(50) + 10,
		"influence_score":        rand.Float64()*50 + 50,
		"calculated_at":          time.Now(),
	}

	return centrality, nil
}

// GetRepositoryNetwork provides repository network visualization data
func (s *RepositoryServiceImpl) GetRepositoryNetwork(ctx context.Context) (interface{}, error) {
	// Generate mock network data
	nodes := []map[string]interface{}{}
	edges := []map[string]interface{}{}

	// Create sample nodes
	for i := 1; i <= 10; i++ {
		nodes = append(nodes, map[string]interface{}{
			"id":       fmt.Sprintf("repo_%d", i),
			"name":     fmt.Sprintf("repository-%d", i),
			"group":    rand.Intn(3) + 1,
			"size":     rand.Intn(20) + 5,
			"language": []string{"go", "python", "javascript"}[rand.Intn(3)],
		})
	}

	// Create sample edges
	for i := 0; i < 15; i++ {
		source := rand.Intn(10) + 1
		target := rand.Intn(10) + 1
		for source == target {
			target = rand.Intn(10) + 1
		}

		edges = append(edges, map[string]interface{}{
			"source": fmt.Sprintf("repo_%d", source),
			"target": fmt.Sprintf("repo_%d", target),
			"weight": rand.Float64()*5 + 1,
			"type":   []string{"dependency", "collaboration", "shared"}[rand.Intn(3)],
		})
	}

	network := map[string]interface{}{
		"nodes":                nodes,
		"edges":                edges,
		"node_count":           len(nodes),
		"edge_count":           len(edges),
		"density":              float64(len(edges)) / float64(len(nodes)*(len(nodes)-1)/2),
		"connected_components": rand.Intn(3) + 1,
		"generated_at":         time.Now(),
	}

	return network, nil
}

// GetRepositoryClusters identifies repository clusters and groupings
func (s *RepositoryServiceImpl) GetRepositoryClusters(ctx context.Context) ([]interface{}, error) {
	clusters := []interface{}{}

	// Generate mock clusters
	clusterTypes := []string{"authentication", "data-processing", "api-gateway", "monitoring", "utilities"}

	for i, clusterType := range clusterTypes {
		repos := []string{}
		repoCount := rand.Intn(8) + 3

		for j := 0; j < repoCount; j++ {
			repos = append(repos, fmt.Sprintf("%s-service-%d", clusterType, j+1))
		}

		clusters = append(clusters, map[string]interface{}{
			"id":               fmt.Sprintf("cluster_%d", i+1),
			"type":             clusterType,
			"name":             fmt.Sprintf("%s Cluster", clusterType),
			"repositories":     repos,
			"repository_count": len(repos),
			"cohesion_score":   rand.Float64()*30 + 70,
			"description":      fmt.Sprintf("Group of repositories handling %s functionality", clusterType),
		})
	}

	return clusters, nil
}

// AnalyzeCrossRepoImpact performs cross-repository impact analysis
func (s *RepositoryServiceImpl) AnalyzeCrossRepoImpact(ctx context.Context, req interface{}) (interface{}, error) {
	request, ok := req.(models.CrossRepoAnalysisRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	if len(request.RepositoryIDs) < 2 {
		return nil, fmt.Errorf("at least 2 repositories required for cross-repo analysis")
	}

	// Perform mock analysis
	affectedRepos := []map[string]interface{}{}
	totalImpact := 0.0

	for _, repoID := range request.RepositoryIDs {
		impact := map[string]interface{}{
			"repository_id":       repoID,
			"impact_level":        []string{"low", "medium", "high"}[rand.Intn(3)],
			"change_propagation":  rand.Float64(),
			"breaking_changes":    rand.Intn(5),
			"affected_components": []string{"api", "database", "frontend", "backend"}[rand.Intn(4):],
		}
		affectedRepos = append(affectedRepos, impact)

		if impact["impact_level"] == "high" {
			totalImpact += 3.0
		} else if impact["impact_level"] == "medium" {
			totalImpact += 2.0
		} else {
			totalImpact += 1.0
		}
	}

	analysis := map[string]interface{}{
		"analysis_id":              fmt.Sprintf("cross_analysis_%d", time.Now().Unix()),
		"analysis_type":            request.AnalysisType,
		"repository_count":         len(request.RepositoryIDs),
		"timeframe":                request.Timeframe,
		"affected_repositories":    affectedRepos,
		"overall_impact_score":     totalImpact / float64(len(request.RepositoryIDs)),
		"risk_assessment":          s.assessCrossRepoRisk(totalImpact),
		"recommendations":          s.generateCrossRepoRecommendations(totalImpact),
		"estimated_migration_time": fmt.Sprintf("%d days", int(totalImpact)+1),
		"analyzed_at":              time.Now(),
	}

	return analysis, nil
}

// Helper methods

func (s *RepositoryServiceImpl) generateMockRepositories(language string, limit int) []*models.Repository {
	var repos []*models.Repository

	for i := 1; i <= limit; i++ {
		repo := &models.Repository{
			ID:          fmt.Sprintf("repo_%d", s.nextID),
			Name:        fmt.Sprintf("repository-%d", i),
			FullName:    fmt.Sprintf("org/repository-%d", i),
			Description: fmt.Sprintf("A sample repository %d", i),
			Language:    language,
			SizeBytes:   int64(rand.Intn(10000000) + 1000000),
			StarsCount:  rand.Intn(1000),
			ForksCount:  rand.Intn(500),
			IsPrivate:   rand.Intn(10) == 0, // 10% private
			CreatedAt:   time.Now().Add(-time.Duration(rand.Intn(365)) * 24 * time.Hour),
		}

		if rand.Intn(10) < 7 { // 70% have been synced recently
			syncedAt := time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
			repo.LastSyncedAt = &syncedAt
		}

		s.repositories[repo.ID] = repo
		repos = append(repos, repo)
		s.nextID++
	}

	return repos
}

func (s *RepositoryServiceImpl) generateMockRepository(id string) *models.Repository {
	return &models.Repository{
		ID:          id,
		Name:        fmt.Sprintf("repo-%s", id),
		FullName:    fmt.Sprintf("org/repo-%s", id),
		Description: "Mock repository for analysis",
		Language:    "go",
		StarsCount:  rand.Intn(500) + 50,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
	}
}

func (s *RepositoryServiceImpl) calculateCriticality(repo *models.Repository) string {
	score := repo.StarsCount + repo.ForksCount*2

	if score > 500 {
		return "high"
	} else if score > 100 {
		return "medium"
	}
	return "low"
}

func (s *RepositoryServiceImpl) assessChangeRisk(repo *models.Repository) map[string]interface{} {
	riskLevel := "low"
	if repo.StarsCount > 500 {
		riskLevel = "high"
	} else if repo.StarsCount > 100 {
		riskLevel = "medium"
	}

	return map[string]interface{}{
		"risk_level":            riskLevel,
		"change_frequency":      "moderate",
		"test_coverage":         rand.Float64()*40 + 60, // 60-100%
		"documentation_quality": rand.Float64()*30 + 70, // 70-100%
	}
}

func (s *RepositoryServiceImpl) assessCrossRepoRisk(totalImpact float64) string {
	if totalImpact > 6.0 {
		return "high"
	} else if totalImpact > 3.0 {
		return "medium"
	}
	return "low"
}

func (s *RepositoryServiceImpl) generateCrossRepoRecommendations(impact float64) []string {
	recommendations := []string{
		"Perform thorough testing across all affected repositories",
		"Update documentation to reflect changes",
	}

	if impact > 4.0 {
		recommendations = append(recommendations,
			"Schedule dedicated testing phase",
			"Consider phased rollout strategy",
			"Prepare rollback procedures")
	}

	return recommendations
}
