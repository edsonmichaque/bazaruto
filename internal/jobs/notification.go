package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/google/uuid"
)

// PushNotificationJob represents a job for sending push notifications
type PushNotificationJob struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Attempts  int       `json:"attempts"`
	RunAtTime time.Time `json:"run_at_time"`
}

// Perform executes the push notification job
func (j *PushNotificationJob) Perform(ctx context.Context) error {
	// TODO: Implement actual push notification logic
	// This would typically:
	// 1. Fetch user's device tokens from database
	// 2. Send push notification through FCM, APNS, etc.
	// 3. Handle delivery status and failures
	// 4. Update notification delivery status

	fmt.Printf("Sending push notification to user %s: %s - %s\n",
		j.UserID.String(), j.Title, j.Body)

	// Simulate push notification sending delay
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(300 * time.Millisecond):
	}

	return nil
}

// SendSMSJob represents a job for sending SMS notifications
type SendSMSJob struct {
	ID        uuid.UUID `json:"id"`
	Phone     string    `json:"phone"`
	Message   string    `json:"message"`
	Attempts  int       `json:"attempts"`
	RunAtTime time.Time `json:"run_at_time"`
}

// Perform executes the SMS sending job
func (j *SendSMSJob) Perform(ctx context.Context) error {
	// TODO: Implement actual SMS sending logic
	// This would typically:
	// 1. Validate phone number format
	// 2. Send SMS through SMS provider (Twilio, AWS SNS, etc.)
	// 3. Log delivery status
	// 4. Handle delivery failures

	fmt.Printf("Sending SMS to %s: %s\n", j.Phone, j.Message)

	// Simulate SMS sending delay
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(200 * time.Millisecond):
	}

	return nil
}

// PushNotificationJob interface methods
func (j *PushNotificationJob) Queue() string               { return job.QueueNotifications }
func (j *PushNotificationJob) MaxRetries() int             { return 3 }
func (j *PushNotificationJob) RetryBackoff() time.Duration { return time.Second }
func (j *PushNotificationJob) Priority() int               { return 0 }
func (j *PushNotificationJob) Type() string                { return "jobs.PushNotificationJob" }
func (j *PushNotificationJob) SetID(id uuid.UUID)          { j.ID = id }
func (j *PushNotificationJob) GetID() uuid.UUID            { return j.ID }
func (j *PushNotificationJob) SetAttempts(attempts int)    { j.Attempts = attempts }
func (j *PushNotificationJob) GetAttempts() int            { return j.Attempts }
func (j *PushNotificationJob) SetRunAt(t time.Time)        { j.RunAtTime = t }
func (j *PushNotificationJob) GetRunAt() time.Time         { return j.RunAtTime }
func (j *PushNotificationJob) Timeout() time.Duration      { return job.DefaultTimeout }

// SendSMSJob interface methods
func (j *SendSMSJob) Queue() string               { return job.QueueNotifications }
func (j *SendSMSJob) MaxRetries() int             { return 3 }
func (j *SendSMSJob) RetryBackoff() time.Duration { return time.Second }
func (j *SendSMSJob) Priority() int               { return 0 }
func (j *SendSMSJob) Type() string                { return "jobs.SendSMSJob" }
func (j *SendSMSJob) SetID(id uuid.UUID)          { j.ID = id }
func (j *SendSMSJob) GetID() uuid.UUID            { return j.ID }
func (j *SendSMSJob) SetAttempts(attempts int)    { j.Attempts = attempts }
func (j *SendSMSJob) GetAttempts() int            { return j.Attempts }
func (j *SendSMSJob) SetRunAt(t time.Time)        { j.RunAtTime = t }
func (j *SendSMSJob) GetRunAt() time.Time         { return j.RunAtTime }
func (j *SendSMSJob) Timeout() time.Duration      { return job.DefaultTimeout }
