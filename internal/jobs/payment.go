package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProcessPaymentJob represents a job for processing payments
type ProcessPaymentJob struct {
	ID             uuid.UUID                `json:"id"`
	PaymentID      uuid.UUID                `json:"payment_id"`
	PaymentService *services.PaymentService `json:"-"` // Injected dependency
	Attempts       int                      `json:"attempts"`
	RunAtTime      time.Time                `json:"run_at_time"`
}

// Perform executes the payment processing job
func (j *ProcessPaymentJob) Perform(ctx context.Context) error {
	log := logger.NewLogger("info", "json")

	// Process payment using the payment service
	payment, err := j.PaymentService.ProcessPayment(ctx, j.PaymentID)
	if err != nil {
		log.Error("Payment processing failed",
			zap.Error(err),
			zap.String("payment_id", j.PaymentID.String()))
		return fmt.Errorf("payment processing failed: %w", err)
	}

	log.Info("Payment processed successfully",
		zap.String("payment_id", j.PaymentID.String()),
		zap.String("transaction_id", payment.TransactionID),
		zap.Float64("amount", payment.Amount))

	return nil
}

// ProcessPaymentJob interface methods
func (j *ProcessPaymentJob) Queue() string               { return job.QueuePayments }
func (j *ProcessPaymentJob) MaxRetries() int             { return 5 } // Payment processing is critical, more retries
func (j *ProcessPaymentJob) RetryBackoff() time.Duration { return 2 * time.Second }
func (j *ProcessPaymentJob) Priority() int               { return 2 } // High priority for payments
func (j *ProcessPaymentJob) Type() string                { return "jobs.ProcessPaymentJob" }
func (j *ProcessPaymentJob) SetID(id uuid.UUID)          { j.ID = id }
func (j *ProcessPaymentJob) GetID() uuid.UUID            { return j.ID }
func (j *ProcessPaymentJob) SetAttempts(attempts int)    { j.Attempts = attempts }
func (j *ProcessPaymentJob) GetAttempts() int            { return j.Attempts }
func (j *ProcessPaymentJob) SetRunAt(t time.Time)        { j.RunAtTime = t }
func (j *ProcessPaymentJob) GetRunAt() time.Time         { return j.RunAtTime }
func (j *ProcessPaymentJob) Timeout() time.Duration      { return job.DefaultTimeout }


