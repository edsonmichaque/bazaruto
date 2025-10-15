package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/edsonmichaque/bazaruto/internal/config"
)

// Limiter defines the interface for rate limiting implementations.
type Limiter interface {
	Allow(key string) bool
	Close()
}

// LimiterFactory creates a new limiter instance.
type LimiterFactory interface {
	Create(rateLimitCfg config.RateLimitConfig, redisCfg config.RedisConfig, perMinute, burst int) (Limiter, error)
	Provider() string
}

// LimiterRegistry manages available rate limiting implementations.
type LimiterRegistry struct {
	factories map[string]LimiterFactory
}

// NewLimiterRegistry creates a new limiter registry with default implementations.
func NewLimiterRegistry() *LimiterRegistry {
	registry := &LimiterRegistry{
		factories: make(map[string]LimiterFactory),
	}

	// Register default implementations
	registry.Register(&MemoryLimiterFactory{})
	registry.Register(&RedisLimiterFactory{})

	return registry
}

// Register adds a new limiter factory to the registry.
func (r *LimiterRegistry) Register(factory LimiterFactory) {
	r.factories[factory.Provider()] = factory
}

// Create creates a limiter using the specified provider.
func (r *LimiterRegistry) Create(provider string, rateLimitCfg config.RateLimitConfig, redisCfg config.RedisConfig, perMinute, burst int) (Limiter, error) {
	factory, exists := r.factories[strings.ToLower(provider)]
	if !exists {
		return nil, fmt.Errorf("unknown rate limiting provider: %s", provider)
	}

	return factory.Create(rateLimitCfg, redisCfg, perMinute, burst)
}

// ListProviders returns a list of available providers.
func (r *LimiterRegistry) ListProviders() []string {
	var providers []string
	for provider := range r.factories {
		providers = append(providers, provider)
	}
	return providers
}

// MemoryLimiterFactory creates in-memory rate limiters.
type MemoryLimiterFactory struct{}

func (f *MemoryLimiterFactory) Provider() string {
	return "memory"
}

func (f *MemoryLimiterFactory) Create(rateLimitCfg config.RateLimitConfig, redisCfg config.RedisConfig, perMinute, burst int) (Limiter, error) {
	return NewInMemoryLimiter(perMinute, burst, rateLimitCfg.TTL, rateLimitCfg.GCInterval), nil
}

// RedisLimiterFactory creates Redis-based rate limiters.
type RedisLimiterFactory struct{}

func (f *RedisLimiterFactory) Provider() string {
	return "redis"
}

func (f *RedisLimiterFactory) Create(rateLimitCfg config.RateLimitConfig, redisCfg config.RedisConfig, perMinute, burst int) (Limiter, error) {
	return NewRedisLimiter(redisCfg, rateLimitCfg, perMinute, burst)
}

// PolicyEngine manages rate limiting policies.
type PolicyEngine struct {
	defaultLimiter Limiter
	policies       map[string]*PolicyLimiter
	keyFunc        KeyFunc
	registry       *LimiterRegistry
}

// PolicyLimiter represents a rate limiter for a specific policy.
type PolicyLimiter struct {
	policy  config.RatePolicy
	limiter Limiter
}

// KeyFunc extracts a key from the request for rate limiting.
type KeyFunc func(r *http.Request) string

// NewPolicyEngine creates a new policy engine.
func NewPolicyEngine(rateLimitCfg config.RateLimitConfig, redisCfg config.RedisConfig, keyFunc KeyFunc) (*PolicyEngine, error) {
	registry := NewLimiterRegistry()

	engine := &PolicyEngine{
		policies: make(map[string]*PolicyLimiter),
		keyFunc:  keyFunc,
		registry: registry,
	}

	// Create default limiter
	defaultLimiter, err := registry.Create(rateLimitCfg.Provider, rateLimitCfg, redisCfg, rateLimitCfg.PerMinute, rateLimitCfg.Burst)
	if err != nil {
		return nil, fmt.Errorf("failed to create default limiter: %w", err)
	}
	engine.defaultLimiter = defaultLimiter

	// Create policy-specific limiters
	for _, policy := range rateLimitCfg.Policies {
		// Find the most restrictive limit for this policy
		var maxPerMinute, maxBurst int
		for _, limit := range policy.Limits {
			if limit.PerMinute > maxPerMinute {
				maxPerMinute = limit.PerMinute
			}
			if limit.Burst > maxBurst {
				maxBurst = limit.Burst
			}
		}

		// Use default values if no limits specified
		if maxPerMinute == 0 {
			maxPerMinute = rateLimitCfg.PerMinute
		}
		if maxBurst == 0 {
			maxBurst = rateLimitCfg.Burst
		}

		limiter, err := engine.registry.Create(rateLimitCfg.Provider, rateLimitCfg, redisCfg, maxPerMinute, maxBurst)
		if err != nil {
			return nil, fmt.Errorf("failed to create limiter for policy %s: %w", policy.Name, err)
		}

		engine.policies[policy.Name] = &PolicyLimiter{
			policy:  policy,
			limiter: limiter,
		}
	}

	return engine, nil
}

// RegisterLimiter allows registering custom rate limiting implementations.
// This enables the Open/Closed Principle - new implementations can be added
// without modifying existing code.
func (e *PolicyEngine) RegisterLimiter(factory LimiterFactory) {
	e.registry.Register(factory)
}

// ListAvailableProviders returns a list of available rate limiting providers.
func (e *PolicyEngine) ListAvailableProviders() []string {
	return e.registry.ListProviders()
}

// Allow checks if a request should be allowed based on rate limiting policies.
func (e *PolicyEngine) Allow(r *http.Request) bool {
	// Find matching policy
	policy := e.findMatchingPolicy(r)
	if policy == nil {
		// Use default limiter
		key := e.keyFunc(r)
		return e.defaultLimiter.Allow(key)
	}

	// Extract key based on policy scope
	key := e.extractKeyForPolicy(r, policy)
	return policy.limiter.Allow(key)
}

// Close closes all limiters.
func (e *PolicyEngine) Close() {
	if e.defaultLimiter != nil {
		e.defaultLimiter.Close()
	}
	for _, policy := range e.policies {
		if policy.limiter != nil {
			policy.limiter.Close()
		}
	}
}

// findMatchingPolicy finds the first policy that matches the request.
func (e *PolicyEngine) findMatchingPolicy(r *http.Request) *PolicyLimiter {
	for _, policy := range e.policies {
		if e.matchesPolicy(r, &policy.policy) {
			return policy
		}
	}
	return nil
}

// matchesPolicy checks if a request matches a policy.
func (e *PolicyEngine) matchesPolicy(r *http.Request, policy *config.RatePolicy) bool {
	// Check path prefix
	if policy.Match.PathPrefix != "" {
		if !strings.HasPrefix(r.URL.Path, policy.Match.PathPrefix) {
			return false
		}
	}

	// Check methods
	if len(policy.Match.Methods) > 0 {
		method := strings.ToUpper(r.Method)
		found := false
		for _, allowedMethod := range policy.Match.Methods {
			if strings.ToUpper(allowedMethod) == method {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// extractKeyForPolicy extracts a key for rate limiting based on policy scope.
func (e *PolicyEngine) extractKeyForPolicy(r *http.Request, policy *PolicyLimiter) string {
	// Use the most restrictive scope from the policy limits
	for _, limit := range policy.policy.Limits {
		switch limit.Scope {
		case "global":
			return "global"
		case "ip":
			return e.extractIPKey(r)
		case "header":
			if limit.Header != "" {
				return e.extractHeaderKey(r, limit.Header)
			}
		}
	}

	// Default to IP-based key
	return e.extractIPKey(r)
}

// extractIPKey extracts the client IP address.
func (e *PolicyEngine) extractIPKey(r *http.Request) string {
	// Check X-Forwarded-For header first
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// extractHeaderKey extracts a key from a specific header.
func (e *PolicyEngine) extractHeaderKey(r *http.Request, headerName string) string {
	value := r.Header.Get(headerName)
	if value == "" {
		// Fall back to IP if header is not present
		return e.extractIPKey(r)
	}
	return fmt.Sprintf("%s:%s", headerName, value)
}

// RateLimitByPolicy is a middleware that applies rate limiting based on policies.
func RateLimitByPolicy(engine *PolicyEngine) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !engine.Allow(r) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BuildPolicyEngine builds a policy engine from configuration.
func BuildPolicyEngine(rateLimitCfg config.RateLimitConfig, redisCfg config.RedisConfig) (*PolicyEngine, func(), error) {
	// Create key function based on configuration
	var keyFunc KeyFunc
	switch strings.ToLower(rateLimitCfg.KeyStrategy) {
	case "ip":
		keyFunc = func(r *http.Request) string {
			// Extract IP address
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ips := strings.Split(forwarded, ",")
				if len(ips) > 0 {
					return strings.TrimSpace(ips[0])
				}
			}
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				return realIP
			}
			ip := r.RemoteAddr
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}
			return ip
		}
	case "header":
		keyFunc = func(r *http.Request) string {
			headerValue := r.Header.Get(rateLimitCfg.KeyHeader)
			if headerValue == "" {
				// Fall back to IP if header is not present
				ip := r.RemoteAddr
				if idx := strings.LastIndex(ip, ":"); idx != -1 {
					ip = ip[:idx]
				}
				return ip
			}
			return fmt.Sprintf("%s:%s", rateLimitCfg.KeyHeader, headerValue)
		}
	default:
		keyFunc = func(r *http.Request) string {
			return "default"
		}
	}

	engine, err := NewPolicyEngine(rateLimitCfg, redisCfg, keyFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create policy engine: %w", err)
	}

	closer := func() {
		engine.Close()
	}

	return engine, closer, nil
}
