// Phase 14E: Task Alerting System
// Provides alerting services for task-related events

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// AlertService interface for sending alerts
type AlertService interface {
	SendTaskCompletionAlert(ctx context.Context, task *Task) error
	SendCriticalTaskAlert(ctx context.Context, task Task) error
	SendDependencyBlockingAlert(ctx context.Context, task Task, blockingTaskIDs []string) error
}

// LoggingAlertService implements AlertService using logging
type LoggingAlertService struct{}

// SendTaskCompletionAlert sends task completion alert via logging
func (s *LoggingAlertService) SendTaskCompletionAlert(ctx context.Context, task *Task) error {
	LogInfo(ctx, "📢 Task %s auto-completed (confidence: %.2f%%)", task.ID, task.VerificationConfidence*100)
	return nil
}

// SendCriticalTaskAlert sends critical task alert via logging
func (s *LoggingAlertService) SendCriticalTaskAlert(ctx context.Context, task Task) error {
	LogInfo(ctx, "🚨 CRITICAL: Task %s is incomplete (status: %s, confidence: %.2f%%)",
		task.ID, task.Status, task.VerificationConfidence*100)
	return nil
}

// SendDependencyBlockingAlert sends dependency blocking alert via logging
func (s *LoggingAlertService) SendDependencyBlockingAlert(ctx context.Context, task Task, blockingTaskIDs []string) error {
	LogInfo(ctx, "🚫 Task %s is blocked by dependencies: %v", task.ID, blockingTaskIDs)
	return nil
}

// WebhookAlertService implements AlertService using HTTP webhooks
type WebhookAlertService struct {
	WebhookURL string
	Client     *http.Client
}

// NewWebhookAlertService creates a new webhook alert service
func NewWebhookAlertService(webhookURL string) *WebhookAlertService {
	return &WebhookAlertService{
		WebhookURL: webhookURL,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// SendTaskCompletionAlert sends task completion alert via webhook
func (s *WebhookAlertService) SendTaskCompletionAlert(ctx context.Context, task *Task) error {
	payload := map[string]interface{}{
		"type":       "task_completion",
		"task_id":    task.ID,
		"title":      task.Title,
		"status":     task.Status,
		"confidence": task.VerificationConfidence,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	return s.sendWebhook(ctx, payload)
}

// SendCriticalTaskAlert sends critical task alert via webhook
func (s *WebhookAlertService) SendCriticalTaskAlert(ctx context.Context, task Task) error {
	payload := map[string]interface{}{
		"type":       "critical_task",
		"task_id":    task.ID,
		"title":      task.Title,
		"status":     task.Status,
		"priority":   task.Priority,
		"confidence": task.VerificationConfidence,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	return s.sendWebhook(ctx, payload)
}

// SendDependencyBlockingAlert sends dependency blocking alert via webhook
func (s *WebhookAlertService) SendDependencyBlockingAlert(ctx context.Context, task Task, blockingTaskIDs []string) error {
	payload := map[string]interface{}{
		"type":           "dependency_blocking",
		"task_id":        task.ID,
		"title":          task.Title,
		"blocking_tasks": blockingTaskIDs,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	return s.sendWebhook(ctx, payload)
}

// sendWebhook sends HTTP POST request to webhook URL
func (s *WebhookAlertService) sendWebhook(ctx context.Context, payload map[string]interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// getAlertService returns the configured alert service
func getAlertService() AlertService {
	// Check for webhook URL in environment
	webhookURL := os.Getenv("SENTINEL_ALERT_WEBHOOK_URL")
	if webhookURL != "" {
		return NewWebhookAlertService(webhookURL)
	}

	// Default to logging service
	return &LoggingAlertService{}
}

// Global alert service instance
var alertService AlertService

// initAlertService initializes the alert service
func initAlertService() {
	alertService = getAlertService()
}
