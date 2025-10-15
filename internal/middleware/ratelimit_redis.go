package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements rate limiting using Redis with Lua scripts.
type RedisLimiter struct {
	rdb    *redis.Client
	prefix string
	rate   float64 // tokens per second
	cap    int     // burst capacity
	ttl    time.Duration
	ctx    context.Context
	script *redis.Script
}

// NewRedisLimiter creates a new Redis-based rate limiter.
func NewRedisLimiter(redisCfg config.RedisConfig, rateLimitCfg config.RateLimitConfig, perMinute, burst int) (*RedisLimiter, error) {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisCfg.Addr,
		Password:     redisCfg.Password,
		DB:           redisCfg.DB,
		DialTimeout:  redisCfg.DialTimeout,
		ReadTimeout:  redisCfg.ReadTimeout,
		WriteTimeout: redisCfg.WriteTimeout,
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Calculate rate (tokens per second)
	rate := float64(perMinute) / 60.0

	// Lua script for token bucket algorithm
	luaScript := `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])
local ttl_ms = tonumber(ARGV[4])

local tokens_key = key .. ":tokens"
local ts_key = key .. ":ts"

local tokens = tonumber(redis.call("GET", tokens_key))
if tokens == nil then tokens = capacity end
local ts = tonumber(redis.call("GET", ts_key))
if ts == nil then ts = now end

local delta = math.max(0, now - ts)
local refill = delta * rate / 1000.0
tokens = math.min(capacity, tokens + refill)
local allowed = 0
if tokens >= 1.0 then
  tokens = tokens - 1.0
  allowed = 1
end

redis.call("SET", tokens_key, tokens, "PX", ttl_ms)
redis.call("SET", ts_key, now, "PX", ttl_ms)

return allowed
`

	script := redis.NewScript(luaScript)

	return &RedisLimiter{
		rdb:    rdb,
		prefix: rateLimitCfg.KeyPrefix,
		rate:   rate,
		cap:    burst,
		ttl:    rateLimitCfg.TTL,
		ctx:    ctx,
		script: script,
	}, nil
}

// Allow checks if a request should be allowed.
func (l *RedisLimiter) Allow(key string) bool {
	now := time.Now().UnixMilli()
	ttlMs := l.ttl.Milliseconds()
	if ttlMs <= 0 {
		ttlMs = 600000 // default 10 minutes
	}

	fullKey := l.prefix + key

	res, err := l.script.Run(l.ctx, l.rdb, []string{fullKey},
		strconv.FormatInt(now, 10),
		strconv.FormatFloat(l.rate, 'f', 6, 64),
		strconv.Itoa(l.cap),
		strconv.FormatInt(ttlMs, 10),
	).Int()

	if err != nil {
		// Fail-open policy: allow request if Redis is unavailable
		return true
	}

	return res == 1
}

// Close closes the Redis connection.
func (l *RedisLimiter) Close() {
	_ = l.rdb.Close()
}
