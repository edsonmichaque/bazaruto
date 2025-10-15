package middleware

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// InMemoryLimiter implements rate limiting using an in-memory token bucket.
type InMemoryLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	ttl      time.Duration
	gcTicker *time.Ticker
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewInMemoryLimiter creates a new in-memory rate limiter.
func NewInMemoryLimiter(perMinute, burst int, ttl, gcInterval time.Duration) *InMemoryLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	limiter := &InMemoryLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(float64(perMinute) / 60.0), // Convert per minute to per second
		burst:    burst,
		ttl:      ttl,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start garbage collection
	if gcInterval > 0 {
		limiter.gcTicker = time.NewTicker(gcInterval)
		go limiter.garbageCollect()
	}

	return limiter
}

// Allow checks if a request should be allowed.
func (l *InMemoryLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[key] = limiter
	}

	return limiter.Allow()
}

// Close stops the limiter and cleans up resources.
func (l *InMemoryLimiter) Close() {
	l.cancel()
	if l.gcTicker != nil {
		l.gcTicker.Stop()
	}
}

// garbageCollect periodically removes old limiters to prevent memory leaks.
func (l *InMemoryLimiter) garbageCollect() {
	for {
		select {
		case <-l.ctx.Done():
			return
		case <-l.gcTicker.C:
			l.cleanup()
		}
	}
}

// cleanup removes limiters that haven't been used recently.
func (l *InMemoryLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Simple cleanup: remove limiters that haven't been used in the last TTL period
	// In a production system, you might want to track last access time
	// For now, we'll just limit the total number of limiters
	if len(l.limiters) > 10000 {
		// Remove half of the limiters (simple strategy)
		count := 0
		for key := range l.limiters {
			if count >= len(l.limiters)/2 {
				break
			}
			delete(l.limiters, key)
			count++
		}
	}
}
