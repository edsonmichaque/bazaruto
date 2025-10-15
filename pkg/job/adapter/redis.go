package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisAdapter implements a Redis-based job queue using sorted sets and lists
type RedisAdapter struct {
	client *redis.Client
	prefix string
}

// NewRedisAdapter creates a new Redis adapter
func NewRedisAdapter(config RedisAdapterConfig) (*RedisAdapter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	prefix := config.Prefix
	if prefix == "" {
		prefix = "bazaruto:jobs"
	}

	return &RedisAdapter{
		client: client,
		prefix: prefix,
	}, nil
}

// Enqueue adds a job to the queue for immediate processing
func (r *RedisAdapter) Enqueue(ctx context.Context, job *SerializedJob) error {
	return r.EnqueueAt(ctx, job, time.Now())
}

// EnqueueAt schedules a job to be processed at a specific time
func (r *RedisAdapter) EnqueueAt(ctx context.Context, job *SerializedJob, at time.Time) error {
	// Serialize job data
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Use sorted set for scheduling (score = timestamp)
	queueKey := r.queueKey(job.Queue)
	score := float64(at.Unix())

	// Add to sorted set
	if err := r.client.ZAdd(ctx, queueKey, &redis.Z{
		Score:  score,
		Member: job.ID.String(),
	}).Err(); err != nil {
		return fmt.Errorf("failed to add job to queue: %w", err)
	}

	// Store job data
	jobKey := r.jobKey(job.ID)
	if err := r.client.Set(ctx, jobKey, jobData, 0).Err(); err != nil {
		// Clean up from queue if storage fails
		r.client.ZRem(ctx, queueKey, job.ID.String())
		return fmt.Errorf("failed to store job data: %w", err)
	}

	return nil
}

// Dequeue retrieves the next job from the queue
func (r *RedisAdapter) Dequeue(ctx context.Context, queueName string) (*SerializedJob, error) {
	queueKey := r.queueKey(queueName)

	// Use Lua script for atomic dequeue operation
	script := `
		local queue_key = KEYS[1]
		local now = tonumber(ARGV[1])
		local worker_id = ARGV[2]
		
		-- Get jobs ready to run (score <= now)
		local jobs = redis.call('ZRANGEBYSCORE', queue_key, '-inf', now, 'LIMIT', 0, 1)
		
		if #jobs == 0 then
			return nil
		end
		
		local job_id = jobs[1]
		
		-- Remove from queue
		redis.call('ZREM', queue_key, job_id)
		
		-- Get job data
		local job_data = redis.call('GET', 'bazaruto:jobs:data:' .. job_id)
		
		if not job_data then
			return nil
		end
		
		-- Parse job data and update lock info
		local job = cjson.decode(job_data)
		job.locked_at = now * 1000  -- Convert to milliseconds
		job.locked_by = worker_id
		
		-- Update job data
		redis.call('SET', 'bazaruto:jobs:data:' .. job_id, cjson.encode(job))
		
		return job_data
	`

	workerID := fmt.Sprintf("worker-%d", time.Now().UnixNano())
	result, err := r.client.Eval(ctx, script, []string{queueKey}, time.Now().Unix(), workerID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("no jobs available in queue %s", queueName)
	}

	// Parse job data
	jobData, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("invalid job data format")
	}

	var job SerializedJob
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// Complete marks a job as successfully completed
func (r *RedisAdapter) Complete(ctx context.Context, jobID uuid.UUID) error {
	jobKey := r.jobKey(jobID)
	return r.client.Del(ctx, jobKey).Err()
}

// Retry marks a job for retry with updated attempt count
func (r *RedisAdapter) Retry(ctx context.Context, job *SerializedJob) error {
	// Increment attempt count
	job.Attempts++

	// Calculate backoff delay (exponential backoff with jitter)
	delay := time.Duration(job.Attempts*job.Attempts) * time.Second
	job.RunAt = time.Now().Add(delay)

	// Clear lock
	job.LockedAt = nil
	job.LockedBy = ""
	job.FailedAt = nil
	job.Error = ""

	// Re-enqueue the job
	return r.EnqueueAt(ctx, job, job.RunAt)
}

// Dead moves a job to the dead letter queue
func (r *RedisAdapter) Dead(ctx context.Context, job *SerializedJob) error {
	// Mark as failed
	now := time.Now()
	job.FailedAt = &now
	job.LockedAt = nil
	job.LockedBy = ""

	// Store in dead jobs set
	deadKey := r.deadJobsKey()
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal dead job: %w", err)
	}

	return r.client.HSet(ctx, deadKey, job.ID.String(), jobData).Err()
}

// Stats returns statistics for all queues
func (r *RedisAdapter) Stats(ctx context.Context) (map[string]*QueueStats, error) {
	// Get all queue keys
	pattern := r.prefix + ":queue:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue keys: %w", err)
	}

	stats := make(map[string]*QueueStats)
	now := time.Now().Unix()

	for _, key := range keys {
		queueName := r.extractQueueName(key)

		// Count pending jobs (ready to run)
		pending, err := r.client.ZCount(ctx, key, "-inf", fmt.Sprintf("%d", now)).Result()
		if err != nil {
			continue
		}

		// Count scheduled jobs (future)
		_, err = r.client.ZCount(ctx, key, fmt.Sprintf("%d", now+1), "+inf").Result()
		if err != nil {
			continue
		}

		stats[queueName] = &QueueStats{
			Queue:      queueName,
			Pending:    pending,
			Processing: 0, // Redis adapter doesn't track processing separately
			Failed:     0, // Would need to count dead jobs
			Completed:  0, // Redis adapter doesn't track completed
		}
	}

	return stats, nil
}

// Clear removes all jobs from a specific queue
func (r *RedisAdapter) Clear(ctx context.Context, queue string) error {
	if queue == "" {
		// Clear all queues
		pattern := r.prefix + ":queue:*"
		keys, err := r.client.Keys(ctx, pattern).Result()
		if err != nil {
			return fmt.Errorf("failed to get queue keys: %w", err)
		}

		for _, key := range keys {
			if err := r.client.Del(ctx, key).Err(); err != nil {
				return fmt.Errorf("failed to clear queue %s: %w", key, err)
			}
		}
	} else {
		queueKey := r.queueKey(queue)
		return r.client.Del(ctx, queueKey).Err()
	}

	return nil
}

// Close closes the Redis connection
func (r *RedisAdapter) Close() error {
	return r.client.Close()
}

// Helper methods

func (r *RedisAdapter) queueKey(queue string) string {
	return fmt.Sprintf("%s:queue:%s", r.prefix, queue)
}

func (r *RedisAdapter) jobKey(jobID uuid.UUID) string {
	return fmt.Sprintf("%s:data:%s", r.prefix, jobID.String())
}

func (r *RedisAdapter) deadJobsKey() string {
	return fmt.Sprintf("%s:dead", r.prefix)
}

func (r *RedisAdapter) extractQueueName(key string) string {
	// Extract queue name from key like "bazaruto:jobs:queue:default"
	parts := len(r.prefix) + 8 // ":queue:" = 8 characters
	if len(key) > parts {
		return key[parts:]
	}
	return "unknown"
}
