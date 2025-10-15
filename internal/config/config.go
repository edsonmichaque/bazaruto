package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root application configuration.
type Config struct {
	Server ServerConfig    `mapstructure:"server"`
	DB     DBConfig        `mapstructure:"db"`
	Redis  RedisConfig     `mapstructure:"redis"`
	Rate   RateLimitConfig `mapstructure:"rate"`
	Jobs   JobsConfig      `mapstructure:"jobs"`

	// Observability fields (flattened from ObservabilityConfig)
	LogLevel       string        `mapstructure:"log_level"`
	LogFormat      string        `mapstructure:"log_format"`
	MetricsEnabled bool          `mapstructure:"metrics_enabled"`
	MetricsPath    string        `mapstructure:"metrics_path"`
	Tracing        TracingConfig `mapstructure:"tracing"`
}

// ServerConfig defines HTTP server options.
type ServerConfig struct {
	Addr         string        `mapstructure:"addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DBConfig defines database connection parameters.
type DBConfig struct {
	DSN            string        `mapstructure:"dsn"`
	MaxConnections int           `mapstructure:"max_connections"`
	MinConnections int           `mapstructure:"min_connections"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	AcquireTimeout time.Duration `mapstructure:"acquire_timeout"`
	MaxLifetime    time.Duration `mapstructure:"max_lifetime"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
}

// RedisConfig defines Redis connection parameters.
type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// RateLimitConfig defines rate limiting configuration.
type RateLimitConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	Provider    string        `mapstructure:"provider"` // "redis" or "memory"
	PerMinute   int           `mapstructure:"per_minute"`
	Burst       int           `mapstructure:"burst"`
	KeyStrategy string        `mapstructure:"key_strategy"` // "ip" or "header"
	KeyHeader   string        `mapstructure:"key_header"`
	KeyPrefix   string        `mapstructure:"key_prefix"`
	TTL         time.Duration `mapstructure:"ttl"`
	GCInterval  time.Duration `mapstructure:"gc_interval"`
	Policies    []RatePolicy  `mapstructure:"policies"`
}

// RatePolicy defines a rate limiting policy.
type RatePolicy struct {
	Name   string      `mapstructure:"name"`
	Match  PolicyMatch `mapstructure:"match"`
	Limits []RateLimit `mapstructure:"limits"`
}

// PolicyMatch defines matching criteria for a rate policy.
type PolicyMatch struct {
	PathPrefix string   `mapstructure:"path_prefix"`
	Methods    []string `mapstructure:"methods"`
}

// RateLimit defines rate limiting rules.
type RateLimit struct {
	Scope     string `mapstructure:"scope"`  // "global", "ip", "header"
	Header    string `mapstructure:"header"` // for header scope
	PerMinute int    `mapstructure:"per_minute"`
	Burst     int    `mapstructure:"burst"`
}

// TracingConfig defines distributed tracing settings.
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
}

// JobsConfig defines background job processing settings.
type JobsConfig struct {
	Adapter      string        `mapstructure:"adapter"`       // "memory", "redis", "database"
	Queues       []string      `mapstructure:"queues"`        // ["default", "mailers", "processing"]
	Concurrency  int           `mapstructure:"concurrency"`   // Worker pool size
	PollInterval time.Duration `mapstructure:"poll_interval"` // Dequeue poll interval
	MaxRetries   int           `mapstructure:"max_retries"`   // Default max retries
	Timeout      time.Duration `mapstructure:"timeout"`       // Default job timeout

	// Redis-specific
	Redis RedisConfig `mapstructure:"redis"`

	// Database-specific
	Database DBConfig `mapstructure:"database"`
}

// Load reads configuration from environment variables and config files.
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file (optional)
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/bazaruto")
	v.AddConfigPath("$HOME/.config/bazaruto")

	// Environment
	v.SetEnvPrefix("BAZARUTO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file if present
	if err := v.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s\n", v.ConfigFileUsed())
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values.
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "60s")
	v.SetDefault("server.idle_timeout", "90s")

	// Database defaults
	v.SetDefault("db.max_connections", 25)
	v.SetDefault("db.min_connections", 5)
	v.SetDefault("db.connect_timeout", "5s")
	v.SetDefault("db.acquire_timeout", "3s")
	v.SetDefault("db.max_lifetime", "30m")
	v.SetDefault("db.idle_timeout", "10m")

	// Redis defaults
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.dial_timeout", "2s")
	v.SetDefault("redis.read_timeout", "1s")
	v.SetDefault("redis.write_timeout", "1s")

	// Rate limiting defaults
	v.SetDefault("rate.enabled", true)
	v.SetDefault("rate.provider", "memory")
	v.SetDefault("rate.per_minute", 120)
	v.SetDefault("rate.burst", 30)
	v.SetDefault("rate.key_strategy", "ip")
	v.SetDefault("rate.key_prefix", "rl:")
	v.SetDefault("rate.ttl", "10m")
	v.SetDefault("rate.gc_interval", "5m")

	// Observability defaults (flattened)
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "json")
	v.SetDefault("metrics_enabled", true)
	v.SetDefault("metrics_path", "/metrics")

	// Tracing defaults (nested)
	v.SetDefault("tracing.enabled", true)
	v.SetDefault("tracing.endpoint", "localhost:4317")
	v.SetDefault("tracing.service_name", "bazaruto")

	// Jobs defaults
	v.SetDefault("jobs.adapter", "memory")
	v.SetDefault("jobs.queues", []string{"default", "mailers", "processing", "heavy"})
	v.SetDefault("jobs.concurrency", 5)
	v.SetDefault("jobs.poll_interval", "1s")
	v.SetDefault("jobs.max_retries", 3)
	v.SetDefault("jobs.timeout", "5m")
}
