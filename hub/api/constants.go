// Constants for status values and enums across all modules
// Ensures consistency and prevents typos

package main

// Task Status Constants (Phase 14E)
const (
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusBlocked    = "blocked"
)

// Change Request Status Constants (Phase 12)
const (
	ChangeRequestStatusPendingApproval = "pending_approval"
	ChangeRequestStatusApproved        = "approved"
	ChangeRequestStatusRejected        = "rejected"
)

// Change Request Implementation Status Constants (Phase 12)
const (
	ImplementationStatusPending    = "pending"
	ImplementationStatusInProgress = "in_progress"
	ImplementationStatusCompleted  = "completed"
	ImplementationStatusBlocked    = "blocked"
)

// Knowledge Item Status Constants (Phase 4)
const (
	KnowledgeItemStatusPending    = "pending"
	KnowledgeItemStatusApproved   = "approved"
	KnowledgeItemStatusRejected   = "rejected"
	KnowledgeItemStatusActive     = "active"
	KnowledgeItemStatusDeprecated = "deprecated"
)

// Document Status Constants (Phase 3)
const (
	DocumentStatusQueued     = "queued"
	DocumentStatusProcessing = "processing"
	DocumentStatusCompleted  = "completed"
	DocumentStatusFailed     = "failed"
)

// Task Link Type Constants (Phase 14E)
const (
	LinkTypeChangeRequest         = "change_request"
	LinkTypeKnowledgeItem         = "knowledge_item"
	LinkTypeComprehensiveAnalysis = "comprehensive_analysis"
	LinkTypeTestRequirement       = "test_requirement"
)

// Task Dependency Type Constants (Phase 14E)
const (
	DependencyTypeExplicit    = "explicit"
	DependencyTypeImplicit    = "implicit"
	DependencyTypeIntegration = "integration"
	DependencyTypeFeature     = "feature"
)

// Task Verification Type Constants (Phase 14E)
const (
	VerificationTypeCodeExistence = "code_existence"
	VerificationTypeCodeUsage     = "code_usage"
	VerificationTypeTestCoverage  = "test_coverage"
	VerificationTypeIntegration   = "integration"
)

// Task Verification Status Constants (Phase 14E)
const (
	VerificationStatusPending  = "pending"
	VerificationStatusVerified = "verified"
	VerificationStatusFailed   = "failed"
)

// Task Priority Constants (Phase 14E)
const (
	TaskPriorityLow      = "low"
	TaskPriorityMedium   = "medium"
	TaskPriorityHigh     = "high"
	TaskPriorityCritical = "critical"
)

// Test Requirement Priority Constants (Phase 10)
const (
	TestRequirementPriorityCritical = "critical"
	TestRequirementPriorityHigh     = "high"
	TestRequirementPriorityMedium   = "medium"
	TestRequirementPriorityLow      = "low"
)

// Gap Severity Constants (Phase 12)
const (
	GapSeverityCritical = "critical"
	GapSeverityHigh     = "high"
	GapSeverityMedium   = "medium"
	GapSeverityLow      = "low"
)
