// Package models - Repository management data models
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import "time"

// Repository represents a code repository
type Repository struct {
	ID            string     `json:"id"`
	OrgID         string     `json:"org_id"`
	Name          string     `json:"name"`
	FullName      string     `json:"full_name"`
	Description   string     `json:"description,omitempty"`
	URL           string     `json:"url,omitempty"`
	CloneURL      string     `json:"clone_url,omitempty"`
	SSHURL        string     `json:"ssh_url,omitempty"`
	DefaultBranch string     `json:"default_branch"`
	Language      string     `json:"language,omitempty"`
	SizeBytes     int64      `json:"size_bytes,omitempty"`
	StarsCount    int        `json:"stars_count"`
	ForksCount    int        `json:"forks_count"`
	WatchersCount int        `json:"watchers_count"`
	IsPrivate     bool       `json:"is_private"`
	IsArchived    bool       `json:"is_archived"`
	IsTemplate    bool       `json:"is_template"`
	IsFork        bool       `json:"is_fork"`
	ParentRepoID  string     `json:"parent_repo_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastSyncedAt  *time.Time `json:"last_synced_at,omitempty"`
	SyncStatus    string     `json:"sync_status"`
}

// CrossRepoAnalysisRequest represents a request for cross-repository analysis
type CrossRepoAnalysisRequest struct {
	RepositoryIDs []string               `json:"repository_ids" validate:"required,min=2"`
	AnalysisType  string                 `json:"analysis_type" validate:"required"`
	Timeframe     string                 `json:"timeframe,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
}
