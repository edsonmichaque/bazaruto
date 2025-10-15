package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CalculatePremiumJob represents a job for calculating insurance premiums
type CalculatePremiumJob struct {
	ID           uuid.UUID              `json:"id"`
	QuoteID      uuid.UUID              `json:"quote_id"`
	QuoteService *services.QuoteService `json:"-"` // Injected dependency
	Attempts     int                    `json:"attempts"`
	RunAtTime    time.Time              `json:"run_at_time"`
}

// Perform executes the premium calculation job
func (j *CalculatePremiumJob) Perform(ctx context.Context) error {
	log := logger.NewLogger("info", "json")

	// Calculate premium using the quote service
	quote, err := j.QuoteService.CalculatePremium(ctx, j.QuoteID)
	if err != nil {
		log.Error("Failed to calculate premium",
			zap.Error(err),
			zap.String("quote_id", j.QuoteID.String()))
		return fmt.Errorf("failed to calculate premium: %w", err)
	}

	log.Info("Premium calculated successfully",
		zap.String("quote_id", j.QuoteID.String()),
		zap.Float64("premium", quote.FinalPrice))

	return nil
}

// CalculatePremiumJob interface methods
func (j *CalculatePremiumJob) Queue() string               { return job.QueueProcessing }
func (j *CalculatePremiumJob) MaxRetries() int             { return 3 }
func (j *CalculatePremiumJob) RetryBackoff() time.Duration { return time.Second }
func (j *CalculatePremiumJob) Priority() int               { return 1 } // High priority for premium calculations
func (j *CalculatePremiumJob) Type() string                { return "jobs.CalculatePremiumJob" }
func (j *CalculatePremiumJob) SetID(id uuid.UUID)          { j.ID = id }
func (j *CalculatePremiumJob) GetID() uuid.UUID            { return j.ID }
func (j *CalculatePremiumJob) SetAttempts(attempts int)    { j.Attempts = attempts }
func (j *CalculatePremiumJob) GetAttempts() int            { return j.Attempts }
func (j *CalculatePremiumJob) SetRunAt(t time.Time)        { j.RunAtTime = t }
func (j *CalculatePremiumJob) GetRunAt() time.Time         { return j.RunAtTime }
func (j *CalculatePremiumJob) Timeout() time.Duration      { return job.DefaultTimeout }
