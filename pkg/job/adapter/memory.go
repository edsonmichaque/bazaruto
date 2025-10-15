package adapter

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryAdapter implements an in-memory job queue using channels and priority queues
type MemoryAdapter struct {
	queues    map[string]*priorityQueue
	mu        sync.RWMutex
	shutdown  chan struct{}
	waitGroup sync.WaitGroup
}

// NewMemoryAdapter creates a new in-memory adapter
func NewMemoryAdapter() *MemoryAdapter {
	return &MemoryAdapter{
		queues:   make(map[string]*priorityQueue),
		shutdown: make(chan struct{}),
	}
}

// Enqueue adds a job to the queue for immediate processing
func (m *MemoryAdapter) Enqueue(ctx context.Context, job *SerializedJob) error {
	return m.EnqueueAt(ctx, job, time.Now())
}

// EnqueueAt schedules a job to be processed at a specific time
func (m *MemoryAdapter) EnqueueAt(ctx context.Context, job *SerializedJob, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Set the run time
	job.RunAt = at

	// Get or create queue
	queue, exists := m.queues[job.Queue]
	if !exists {
		queue = &priorityQueue{}
		heap.Init(queue)
		m.queues[job.Queue] = queue
	}

	// Add job to priority queue
	heap.Push(queue, &jobItem{
		job:      job,
		priority: job.Priority,
		runAt:    at,
	})

	return nil
}

// Dequeue retrieves the next job from the queue
func (m *MemoryAdapter) Dequeue(ctx context.Context, queueName string) (*SerializedJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.queues[queueName]
	if !exists || queue.Len() == 0 {
		return nil, fmt.Errorf("no jobs available in queue %s", queueName)
	}

	// Check if the next job is ready to run
	item := (*queue)[0]
	if item.runAt.After(time.Now()) {
		return nil, fmt.Errorf("no jobs ready to run in queue %s", queueName)
	}

	// Remove job from queue
	jobItem := heap.Pop(queue).(*jobItem)
	job := jobItem.job

	// Mark as locked
	now := time.Now()
	job.LockedAt = &now
	job.LockedBy = "memory-worker"

	return job, nil
}

// Complete marks a job as successfully completed
func (m *MemoryAdapter) Complete(ctx context.Context, jobID uuid.UUID) error {
	// In memory adapter, we don't need to track completed jobs
	// They are simply removed from the queue when dequeued
	return nil
}

// Retry marks a job for retry with updated attempt count
func (m *MemoryAdapter) Retry(ctx context.Context, job *SerializedJob) error {
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
	return m.EnqueueAt(ctx, job, job.RunAt)
}

// Dead moves a job to the dead letter queue
func (m *MemoryAdapter) Dead(ctx context.Context, job *SerializedJob) error {
	// In memory adapter, we just mark it as failed
	now := time.Now()
	job.FailedAt = &now
	job.LockedAt = nil
	job.LockedBy = ""

	// We could maintain a separate dead jobs map here if needed
	return nil
}

// Stats returns statistics for all queues
func (m *MemoryAdapter) Stats(ctx context.Context) (map[string]*QueueStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]*QueueStats)

	for queueName, queue := range m.queues {
		ready := 0
		scheduled := 0

		for _, item := range *queue {
			if item.runAt.After(time.Now()) {
				scheduled++
			} else {
				ready++
			}
		}

		stats[queueName] = &QueueStats{
			Queue:      queueName,
			Pending:    int64(ready),
			Processing: 0, // Memory adapter doesn't track processing
			Failed:     0, // Memory adapter doesn't track failed
			Completed:  0, // Memory adapter doesn't track completed
		}
	}

	return stats, nil
}

// Clear removes all jobs from a specific queue
func (m *MemoryAdapter) Clear(ctx context.Context, queue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if queue == "" {
		// Clear all queues
		m.queues = make(map[string]*priorityQueue)
	} else {
		delete(m.queues, queue)
	}

	return nil
}

// Close closes the adapter and cleans up resources
func (m *MemoryAdapter) Close() error {
	close(m.shutdown)
	m.waitGroup.Wait()
	return nil
}

// jobItem represents an item in the priority queue
type jobItem struct {
	job      *SerializedJob
	priority int
	runAt    time.Time
	index    int
}

// priorityQueue implements heap.Interface for priority-based job scheduling
type priorityQueue []*jobItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// First by run time, then by priority (higher priority first)
	if pq[i].runAt.Equal(pq[j].runAt) {
		return pq[i].priority > pq[j].priority
	}
	return pq[i].runAt.Before(pq[j].runAt)
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*jobItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
