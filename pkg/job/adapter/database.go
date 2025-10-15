package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DatabaseAdapter implements a database-based job queue using GORM
type DatabaseAdapter struct {
	db *gorm.DB
}

// NewDatabaseAdapter creates a new database adapter
func NewDatabaseAdapter(config DatabaseAdapterConfig) (*DatabaseAdapter, error) {
	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Auto-migrate job tables
	if err := db.AutoMigrate(&SerializedJob{}); err != nil {
		return nil, fmt.Errorf("failed to migrate job tables: %w", err)
	}

	return &DatabaseAdapter{
		db: db,
	}, nil
}

// Enqueue adds a job to the queue for immediate processing
func (d *DatabaseAdapter) Enqueue(ctx context.Context, job *SerializedJob) error {
	return d.EnqueueAt(ctx, job, time.Now())
}

// EnqueueAt schedules a job to be processed at a specific time
func (d *DatabaseAdapter) EnqueueAt(ctx context.Context, job *SerializedJob, at time.Time) error {
	job.RunAt = at
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	return d.db.WithContext(ctx).Create(job).Error
}

// Dequeue retrieves the next job from the queue using SKIP LOCKED
func (d *DatabaseAdapter) Dequeue(ctx context.Context, queueName string) (*SerializedJob, error) {
	var job SerializedJob

	// Use SKIP LOCKED to handle concurrent workers
	err := d.db.WithContext(ctx).
		Where("queue = ? AND run_at <= ? AND failed_at IS NULL AND locked_at IS NULL",
			queueName, time.Now()).
		Order("priority DESC, run_at ASC").
		First(&job).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no jobs available in queue %s", queueName)
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	// Lock the job
	now := time.Now()
	workerID := fmt.Sprintf("worker-%d", time.Now().UnixNano())

	updateResult := d.db.WithContext(ctx).
		Model(&job).
		Where("id = ? AND locked_at IS NULL", job.ID).
		Updates(map[string]interface{}{
			"locked_at":  &now,
			"locked_by":  workerID,
			"updated_at": now,
		})

	if updateResult.Error != nil {
		return nil, fmt.Errorf("failed to lock job: %w", updateResult.Error)
	}

	if updateResult.RowsAffected == 0 {
		// Job was locked by another worker
		return nil, fmt.Errorf("job was locked by another worker")
	}

	job.LockedAt = &now
	job.LockedBy = workerID

	return &job, nil
}

// Complete marks a job as successfully completed
func (d *DatabaseAdapter) Complete(ctx context.Context, jobID uuid.UUID) error {
	return d.db.WithContext(ctx).Delete(&SerializedJob{}, jobID).Error
}

// Retry marks a job for retry with updated attempt count
func (d *DatabaseAdapter) Retry(ctx context.Context, job *SerializedJob) error {
	// Increment attempt count
	job.Attempts++

	// Calculate backoff delay (exponential backoff with jitter)
	delay := time.Duration(job.Attempts*job.Attempts) * time.Second
	job.RunAt = time.Now().Add(delay)

	// Clear lock and error
	job.LockedAt = nil
	job.LockedBy = ""
	job.FailedAt = nil
	job.Error = ""
	job.UpdatedAt = time.Now()

	return d.db.WithContext(ctx).Save(job).Error
}

// Dead moves a job to the dead letter queue
func (d *DatabaseAdapter) Dead(ctx context.Context, job *SerializedJob) error {
	// Create dead job record
	now := time.Now()
	deadJob := &DeadJob{
		ID:         job.ID,
		Type:       job.Type,
		Payload:    job.Payload,
		Queue:      job.Queue,
		Priority:   job.Priority,
		MaxRetries: job.MaxRetries,
		Attempts:   job.Attempts,
		RunAt:      job.RunAt,
		FailedAt:   now,
		Error:      job.Error,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  now,
		DeadAt:     now,
	}

	// Start transaction
	tx := d.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create dead job
	if err := tx.Create(deadJob).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create dead job: %w", err)
	}

	// Delete original job
	if err := tx.Delete(job).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete original job: %w", err)
	}

	return tx.Commit().Error
}

// Stats returns statistics for all queues
func (d *DatabaseAdapter) Stats(ctx context.Context) (map[string]*QueueStats, error) {
	var results []struct {
		Queue      string
		Pending    int64
		Processing int64
		Failed     int64
	}

	// Get queue statistics
	err := d.db.WithContext(ctx).
		Table("jobs").
		Select(`
			queue,
			COUNT(CASE WHEN failed_at IS NULL AND locked_at IS NULL THEN 1 END) as pending,
			COUNT(CASE WHEN failed_at IS NULL AND locked_at IS NOT NULL THEN 1 END) as processing,
			COUNT(CASE WHEN failed_at IS NOT NULL THEN 1 END) as failed
		`).
		Group("queue").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}

	stats := make(map[string]*QueueStats)
	for _, result := range results {
		stats[result.Queue] = &QueueStats{
			Queue:      result.Queue,
			Pending:    result.Pending,
			Processing: result.Processing,
			Failed:     result.Failed,
			Completed:  0, // Would need to track completed jobs separately
		}
	}

	return stats, nil
}

// Clear removes all jobs from a specific queue
func (d *DatabaseAdapter) Clear(ctx context.Context, queue string) error {
	if queue == "" {
		// Clear all queues
		return d.db.WithContext(ctx).Where("1 = 1").Delete(&SerializedJob{}).Error
	} else {
		return d.db.WithContext(ctx).Where("queue = ?", queue).Delete(&SerializedJob{}).Error
	}
}

// Close closes the database connection
func (d *DatabaseAdapter) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
