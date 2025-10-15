package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WebhookJob represents a job for sending webhook notifications to external systems
type WebhookJob struct {
	ID        uuid.UUID              `json:"id"`
	URL       string                 `json:"url"`
	Method    string                 `json:"method"`
	Headers   map[string]string      `json:"headers"`
	Payload   map[string]interface{} `json:"payload"`
	EventType string                 `json:"event_type"`
	EventID   uuid.UUID              `json:"event_id"`
	Attempts  int                    `json:"attempts"`
	RunAtTime time.Time              `json:"run_at_time"`
}

// Perform executes the webhook job
func (j *WebhookJob) Perform(ctx context.Context) error {
	log := logger.NewLogger("info", "json")

	log.Info("Sending webhook notification",
		zap.String("webhook_url", j.URL),
		zap.String("event_type", j.EventType),
		zap.String("event_id", j.EventID.String()))

	// Prepare the request payload
	payloadBytes, err := json.Marshal(j.Payload)
	if err != nil {
		log.Error("Failed to marshal webhook payload",
			zap.Error(err),
			zap.String("webhook_url", j.URL))
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, j.Method, j.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Error("Failed to create HTTP request",
			zap.Error(err),
			zap.String("webhook_url", j.URL))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Bazaruto-Webhook/1.0")
	req.Header.Set("X-Event-Type", j.EventType)
	req.Header.Set("X-Event-ID", j.EventID.String())
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))

	// Set custom headers
	for key, value := range j.Headers {
		req.Header.Set(key, value)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send webhook request",
			zap.Error(err),
			zap.String("webhook_url", j.URL))
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("Webhook request failed with non-2xx status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("webhook_url", j.URL))
		return fmt.Errorf("webhook request failed with status %d", resp.StatusCode)
	}

	log.Info("Webhook notification sent successfully",
		zap.Int("status_code", resp.StatusCode),
		zap.String("webhook_url", j.URL),
		zap.String("event_type", j.EventType))

	return nil
}

// WebhookDeliveryJob represents a job for sending webhook notifications with persistent retry tracking
// Inspired by Stripe's webhook retry mechanism
type WebhookDeliveryJob struct {
	ID             uuid.UUID                `json:"id"`
	DeliveryID     uuid.UUID                `json:"delivery_id"`
	URL            string                   `json:"url"`
	Method         string                   `json:"method"`
	Headers        map[string]string        `json:"headers"`
	Payload        map[string]interface{}   `json:"payload"`
	EventType      string                   `json:"event_type"`
	EventID        uuid.UUID                `json:"event_id"`
	WebhookService *services.WebhookService `json:"-"` // Injected dependency
	Attempts       int                      `json:"attempts"`
	RunAtTime      time.Time                `json:"run_at_time"`
}

// Perform executes the webhook delivery job with Stripe-inspired retry logic
func (j *WebhookDeliveryJob) Perform(ctx context.Context) error {
	log := logger.NewLogger("info", "json")

	log.Info("Sending webhook delivery",
		zap.String("delivery_id", j.DeliveryID.String()),
		zap.String("webhook_url", j.URL),
		zap.String("event_type", j.EventType),
		zap.String("event_id", j.EventID.String()),
		zap.Int("attempt", j.Attempts))

	// Get the delivery record
	delivery, err := j.WebhookService.GetWebhookDelivery(ctx, j.DeliveryID)
	if err != nil {
		log.Error("Failed to get webhook delivery record",
			zap.Error(err),
			zap.String("delivery_id", j.DeliveryID.String()))
		return fmt.Errorf("failed to get delivery record: %w", err)
	}

	// Update attempt count
	delivery.AttemptCount++
	now := time.Now()
	delivery.LastAttemptAt = &now

	// Prepare the request payload
	payloadBytes, err := json.Marshal(j.Payload)
	if err != nil {
		log.Error("Failed to marshal webhook payload",
			zap.Error(err),
			zap.String("delivery_id", j.DeliveryID.String()))

		delivery.Status = models.WebhookStatusFailed
		delivery.ErrorMessage = fmt.Sprintf("Failed to marshal payload: %v", err)
		delivery.FailedAt = &now
		_ = j.WebhookService.UpdateWebhookDelivery(ctx, delivery)

		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, j.Method, j.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Error("Failed to create HTTP request",
			zap.Error(err),
			zap.String("delivery_id", j.DeliveryID.String()))

		delivery.Status = models.WebhookStatusFailed
		delivery.ErrorMessage = fmt.Sprintf("Failed to create request: %v", err)
		delivery.FailedAt = &now
		_ = j.WebhookService.UpdateWebhookDelivery(ctx, delivery)

		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Bazaruto-Webhook/1.0")
	req.Header.Set("X-Event-Type", j.EventType)
	req.Header.Set("X-Event-ID", j.EventID.String())
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))

	// Set custom headers
	for key, value := range j.Headers {
		req.Header.Set(key, value)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send webhook request",
			zap.Error(err),
			zap.String("delivery_id", j.DeliveryID.String()))

		// Check if this is a retryable error
		if j.isRetryableError(err) {
			delivery.Status = models.WebhookStatusPending // Keep as pending for retry
			delivery.ErrorMessage = fmt.Sprintf("Retryable error: %v", err)
		} else {
			delivery.Status = models.WebhookStatusFailed
			delivery.ErrorMessage = fmt.Sprintf("Non-retryable error: %v", err)
			delivery.FailedAt = &now
		}
		_ = j.WebhookService.UpdateWebhookDelivery(ctx, delivery)

		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	var responseBody string
	if resp.Body != nil {
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		if n > 0 {
			responseBody = string(buf[:n])
		}
	}

	// Update delivery record with response details
	delivery.ResponseStatus = resp.StatusCode
	delivery.ResponseBody = responseBody

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("Webhook request failed with non-2xx status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("delivery_id", j.DeliveryID.String()),
			zap.String("response_body", responseBody))

		// Check if this is a retryable status code
		if j.isRetryableStatusCode(resp.StatusCode) {
			delivery.Status = models.WebhookStatusPending // Keep as pending for retry
			delivery.ErrorMessage = fmt.Sprintf("Retryable status code: %d", resp.StatusCode)
		} else {
			delivery.Status = models.WebhookStatusFailed
			delivery.ErrorMessage = fmt.Sprintf("Non-retryable status code: %d", resp.StatusCode)
			delivery.FailedAt = &now
		}
		_ = j.WebhookService.UpdateWebhookDelivery(ctx, delivery)

		return fmt.Errorf("webhook request failed with status %d", resp.StatusCode)
	}

	// Success - update delivery status
	delivery.Status = models.WebhookStatusDelivered
	delivery.DeliveredAt = &now
	delivery.ErrorMessage = ""

	if err := j.WebhookService.UpdateWebhookDelivery(ctx, delivery); err != nil {
		log.Error("Failed to update webhook delivery status",
			zap.Error(err),
			zap.String("delivery_id", j.DeliveryID.String()))
		// Don't return error here as the webhook was successfully sent
	}

	log.Info("Webhook delivery completed successfully",
		zap.Int("status_code", resp.StatusCode),
		zap.String("delivery_id", j.DeliveryID.String()),
		zap.String("webhook_url", j.URL),
		zap.String("event_type", j.EventType))

	return nil
}

// isRetryableError determines if an error is retryable
func (j *WebhookDeliveryJob) isRetryableError(err error) bool {
	// Network errors, timeouts, and connection issues are typically retryable
	// This is a simplified implementation - in production you'd check for specific error types
	return true // For now, consider all errors retryable
}

// isRetryableStatusCode determines if an HTTP status code is retryable
func (j *WebhookDeliveryJob) isRetryableStatusCode(statusCode int) bool {
	// Stripe-inspired retry logic:
	// - 4xx client errors (except 400, 401, 403, 404) are retryable
	// - 5xx server errors are retryable
	// - 400, 401, 403, 404 are not retryable as they indicate client issues

	switch {
	case statusCode >= 500:
		return true // Server errors are retryable
	case statusCode == 400, statusCode == 401, statusCode == 403, statusCode == 404:
		return false // Client errors that shouldn't be retried
	case statusCode >= 400:
		return true // Other 4xx errors might be retryable
	default:
		return false
	}
}

// calculateRetryDelay calculates the delay for the next retry attempt using exponential backoff
// Inspired by Stripe's retry schedule: 2s, 4s, 8s, 16s, 32s, 64s, 128s, 256s, 512s, 1024s
func (j *WebhookDeliveryJob) calculateRetryDelay(attempt int) time.Duration {
	// Stripe uses exponential backoff with jitter
	baseDelay := time.Duration(math.Pow(2, float64(attempt))) * time.Second

	// Cap at 24 hours (86400 seconds)
	maxDelay := 24 * time.Hour
	if baseDelay > maxDelay {
		baseDelay = maxDelay
	}

	return baseDelay
}

// WebhookJob interface methods
func (j *WebhookJob) Queue() string               { return job.QueueNotifications }
func (j *WebhookJob) MaxRetries() int             { return 5 }               // Webhooks should have more retries due to network issues
func (j *WebhookJob) RetryBackoff() time.Duration { return 2 * time.Second } // Exponential backoff starting at 2 seconds
func (j *WebhookJob) Priority() int               { return 2 }               // Medium priority for webhooks
func (j *WebhookJob) Type() string                { return "jobs.WebhookJob" }
func (j *WebhookJob) SetID(id uuid.UUID)          { j.ID = id }
func (j *WebhookJob) GetID() uuid.UUID            { return j.ID }
func (j *WebhookJob) SetAttempts(attempts int)    { j.Attempts = attempts }
func (j *WebhookJob) GetAttempts() int            { return j.Attempts }
func (j *WebhookJob) SetRunAt(t time.Time)        { j.RunAtTime = t }
func (j *WebhookJob) GetRunAt() time.Time         { return j.RunAtTime }
func (j *WebhookJob) Timeout() time.Duration      { return 30 * time.Second } // 30 second timeout for webhook requests

// WebhookDeliveryJob interface methods
func (j *WebhookDeliveryJob) Queue() string               { return job.QueueNotifications }
func (j *WebhookDeliveryJob) MaxRetries() int             { return 10 }                                // More retries for persistent webhook delivery
func (j *WebhookDeliveryJob) RetryBackoff() time.Duration { return j.calculateRetryDelay(j.Attempts) } // Dynamic backoff
func (j *WebhookDeliveryJob) Priority() int               { return 2 }                                 // Medium priority for webhooks
func (j *WebhookDeliveryJob) Type() string                { return "jobs.WebhookDeliveryJob" }
func (j *WebhookDeliveryJob) SetID(id uuid.UUID)          { j.ID = id }
func (j *WebhookDeliveryJob) GetID() uuid.UUID            { return j.ID }
func (j *WebhookDeliveryJob) SetAttempts(attempts int)    { j.Attempts = attempts }
func (j *WebhookDeliveryJob) GetAttempts() int            { return j.Attempts }
func (j *WebhookDeliveryJob) SetRunAt(t time.Time)        { j.RunAtTime = t }
func (j *WebhookDeliveryJob) GetRunAt() time.Time         { return j.RunAtTime }
func (j *WebhookDeliveryJob) Timeout() time.Duration      { return 30 * time.Second } // 30 second timeout for webhook requests
