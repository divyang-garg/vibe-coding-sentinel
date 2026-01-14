// Package services provides unit tests for repository service.
// Complies with CODING_STANDARDS.md: Test files max 500 lines
package services

import (
	"context"
	"testing"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryServiceImpl_ListRepositories(t *testing.T) {
	tests := []struct {
		name     string
		language string
		limit    int
		wantErr  bool
	}{
		{
			name:     "list Go repositories",
			language: "go",
			limit:    5,
			wantErr:  false,
		},
		{
			name:     "list all repositories",
			language: "",
			limit:    10,
			wantErr:  false,
		},
		{
			name:    "zero limit",
			limit:   0,
			wantErr: false,
		},
		{
			name:     "large limit gets capped",
			language: "python",
			limit:    200,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewRepositoryService()

			result, err := service.ListRepositories(context.Background(), tt.language, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Should return a slice of interfaces
				assert.IsType(t, []interface{}{}, result)

				// Check limit is respected
				expectedLimit := tt.limit
				if expectedLimit == 0 {
					expectedLimit = 50 // default
				}
				if expectedLimit > 100 {
					expectedLimit = 100 // max
				}
				assert.LessOrEqual(t, len(result), expectedLimit)

				// Validate each repository structure
				for _, repo := range result {
					repoMap, ok := repo.(map[string]interface{})
					assert.True(t, ok, "each repository should be a map")

					assert.Contains(t, repoMap, "id")
					assert.Contains(t, repoMap, "name")
					assert.Contains(t, repoMap, "full_name")
					assert.Contains(t, repoMap, "description")
					assert.Contains(t, repoMap, "language")
					assert.Contains(t, repoMap, "stars_count")
					assert.Contains(t, repoMap, "forks_count")
					assert.Contains(t, repoMap, "size_bytes")
					assert.Contains(t, repoMap, "is_private")
					assert.Contains(t, repoMap, "created_at")

					// Validate types
					assert.IsType(t, "", repoMap["id"])
					assert.IsType(t, "", repoMap["name"])
					assert.IsType(t, 0, repoMap["stars_count"])
					assert.IsType(t, 0, repoMap["forks_count"])
					assert.IsType(t, int64(0), repoMap["size_bytes"])
					assert.IsType(t, false, repoMap["is_private"])
				}
			}
		})
	}
}

func TestRepositoryServiceImpl_GetRepositoryImpact(t *testing.T) {
	service := NewRepositoryService()

	// Test with a mock repository ID
	repoID := "test-repo-123"

	result, err := service.GetRepositoryImpact(context.Background(), repoID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Validate impact analysis structure
	assert.Contains(t, resultMap, "repository_id")
	assert.Contains(t, resultMap, "name")
	assert.Contains(t, resultMap, "direct_dependencies")
	assert.Contains(t, resultMap, "dependent_repositories")
	assert.Contains(t, resultMap, "impact_score")
	assert.Contains(t, resultMap, "criticality_level")
	assert.Contains(t, resultMap, "change_risk_assessment")
	assert.Contains(t, resultMap, "recommended_reviewers")
	assert.Contains(t, resultMap, "estimated_review_time")
	assert.Contains(t, resultMap, "last_analyzed")

	assert.Equal(t, repoID, resultMap["repository_id"])

	impactScore, ok := resultMap["impact_score"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, impactScore, 85.0)
	assert.LessOrEqual(t, impactScore, 100.0)

	criticalityLevel, ok := resultMap["criticality_level"].(string)
	assert.True(t, ok)
	assert.Contains(t, []string{"low", "medium", "high"}, criticalityLevel)
}

func TestRepositoryServiceImpl_GetRepositoryCentrality(t *testing.T) {
	service := NewRepositoryService()

	repoID := "central-repo-456"

	result, err := service.GetRepositoryCentrality(context.Background(), repoID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Validate centrality metrics
	assert.Contains(t, resultMap, "repository_id")
	assert.Contains(t, resultMap, "name")
	assert.Contains(t, resultMap, "degree_centrality")
	assert.Contains(t, resultMap, "betweenness_centrality")
	assert.Contains(t, resultMap, "closeness_centrality")
	assert.Contains(t, resultMap, "eigenvector_centrality")
	assert.Contains(t, resultMap, "overall_centrality")
	assert.Contains(t, resultMap, "connection_count")
	assert.Contains(t, resultMap, "influence_score")
	assert.Contains(t, resultMap, "calculated_at")

	assert.Equal(t, repoID, resultMap["repository_id"])

	// Validate centrality scores are reasonable
	degree, ok := resultMap["degree_centrality"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, degree, 0.0)
	assert.LessOrEqual(t, degree, 1.0)

	betweenness, ok := resultMap["betweenness_centrality"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, betweenness, 0.0)
	assert.LessOrEqual(t, betweenness, 1.0)

	overall, ok := resultMap["overall_centrality"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, overall, 0.0)
	assert.LessOrEqual(t, overall, 100.0)
}

func TestRepositoryServiceImpl_GetRepositoryNetwork(t *testing.T) {
	service := NewRepositoryService()

	result, err := service.GetRepositoryNetwork(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Validate network structure
	assert.Contains(t, resultMap, "nodes")
	assert.Contains(t, resultMap, "edges")
	assert.Contains(t, resultMap, "node_count")
	assert.Contains(t, resultMap, "edge_count")
	assert.Contains(t, resultMap, "density")
	assert.Contains(t, resultMap, "connected_components")
	assert.Contains(t, resultMap, "generated_at")

	nodes, ok := resultMap["nodes"].([]map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, nodes)

	edges, ok := resultMap["edges"].([]map[string]interface{})
	assert.True(t, ok)

	nodeCount, ok := resultMap["node_count"].(int)
	assert.True(t, ok)
	assert.Equal(t, len(nodes), nodeCount)

	edgeCount, ok := resultMap["edge_count"].(int)
	assert.True(t, ok)
	assert.Equal(t, len(edges), edgeCount)

	// Validate node structure
	for _, node := range nodes {
		assert.Contains(t, node, "id")
		assert.Contains(t, node, "name")
		assert.Contains(t, node, "group")
		assert.Contains(t, node, "size")
		assert.Contains(t, node, "language")
	}

	// Validate edge structure
	for _, edge := range edges {
		assert.Contains(t, edge, "source")
		assert.Contains(t, edge, "target")
		assert.Contains(t, edge, "weight")
		assert.Contains(t, edge, "type")
	}
}

func TestRepositoryServiceImpl_GetRepositoryClusters(t *testing.T) {
	service := NewRepositoryService()

	result, err := service.GetRepositoryClusters(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)

	clusters := result
	assert.NotEmpty(t, clusters)

	// Should have at least a few clusters
	assert.GreaterOrEqual(t, len(clusters), 3)

	// Validate cluster structure
	for _, cluster := range clusters {
		clusterMap, ok := cluster.(map[string]interface{})
		assert.True(t, ok)

		assert.Contains(t, clusterMap, "id")
		assert.Contains(t, clusterMap, "type")
		assert.Contains(t, clusterMap, "name")
		assert.Contains(t, clusterMap, "repositories")
		assert.Contains(t, clusterMap, "repository_count")
		assert.Contains(t, clusterMap, "cohesion_score")
		assert.Contains(t, clusterMap, "description")

		repos, ok := clusterMap["repositories"].([]string)
		assert.True(t, ok)
		assert.NotEmpty(t, repos)

		repoCount, ok := clusterMap["repository_count"].(int)
		assert.True(t, ok)
		assert.Equal(t, len(repos), repoCount)

		cohesionScore, ok := clusterMap["cohesion_score"].(float64)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, cohesionScore, 70.0)
		assert.LessOrEqual(t, cohesionScore, 100.0)
	}
}

func TestRepositoryServiceImpl_AnalyzeCrossRepoImpact(t *testing.T) {
	tests := []struct {
		name         string
		repoIDs      []string
		analysisType string
		timeframe    string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid cross-repo analysis",
			repoIDs:      []string{"repo-1", "repo-2", "repo-3"},
			analysisType: "dependency",
			timeframe:    "30d",
			wantErr:      false,
		},
		{
			name:         "too few repositories",
			repoIDs:      []string{"repo-1"},
			analysisType: "dependency",
			wantErr:      true,
			errMsg:       "at least 2 repositories",
		},
		{
			name:         "empty repository list",
			repoIDs:      []string{},
			analysisType: "dependency",
			wantErr:      true,
			errMsg:       "at least 2 repositories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewRepositoryService()

			req := models.CrossRepoAnalysisRequest{
				RepositoryIDs: tt.repoIDs,
				AnalysisType:  tt.analysisType,
				Timeframe:     tt.timeframe,
			}

			result, err := service.AnalyzeCrossRepoImpact(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				assert.Contains(t, resultMap, "analysis_id")
				assert.Contains(t, resultMap, "analysis_type")
				assert.Contains(t, resultMap, "repository_count")
				assert.Contains(t, resultMap, "timeframe")
				assert.Contains(t, resultMap, "affected_repositories")
				assert.Contains(t, resultMap, "overall_impact_score")
				assert.Contains(t, resultMap, "risk_assessment")
				assert.Contains(t, resultMap, "recommendations")
				assert.Contains(t, resultMap, "estimated_migration_time")
				assert.Contains(t, resultMap, "analyzed_at")

				assert.Equal(t, tt.analysisType, resultMap["analysis_type"])
				assert.Equal(t, len(tt.repoIDs), resultMap["repository_count"])

				affectedRepos, ok := resultMap["affected_repositories"].([]map[string]interface{})
				assert.True(t, ok)
				assert.Len(t, affectedRepos, len(tt.repoIDs))

				// Validate each affected repository
				for _, repo := range affectedRepos {
					assert.Contains(t, repo, "repository_id")
					assert.Contains(t, repo, "impact_level")
					assert.Contains(t, repo, "change_propagation")
					assert.Contains(t, repo, "breaking_changes")
					assert.Contains(t, repo, "affected_components")

					impactLevel, ok := repo["impact_level"].(string)
					assert.True(t, ok)
					assert.Contains(t, []string{"low", "medium", "high"}, impactLevel)
				}
			}
		})
	}
}
