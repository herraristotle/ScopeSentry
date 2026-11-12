// ct/types------------------------------------
// @file      : options.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - configuration options and parameter definitions
// ============================================

package types

import (
	"fmt"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/models"
)

// SDKOptions is the SDK configuration options
type SDKOptions struct {
	// CT log server configuration list
	LogServers []*LogServerConfig `json:"log_servers"`

	// Initial list of watched domains
	Domains []string `json:"domains"`

	// Domain matcher configuration
	DomainMatcher *DomainMatcherConfig `json:"domain_matcher"`

	// Performance configuration
	Performance *PerformanceConfig `json:"performance"`

	// Redis configuration (used for checkpoint storage)
	Redis *RedisConfig `json:"redis"`
}

// LogServerConfig is the CT log server configuration
type LogServerConfig struct {
	// Basic configuration
	URL string `json:"url"` // CT log URL (required)

	// Scanner configuration
	BatchSize     int `json:"batch_size"`     // Batch fetch size (default 100)
	ParallelFetch int `json:"parallel_fetch"` // Parallel fetch count (default 1)
	BufferSize    int `json:"buffer_size"`    // Buffer size (default 1000)

	// Network configuration
	Timeout    time.Duration `json:"timeout"`     // Request timeout (default 30s)
	MaxRetries int           `json:"max_retries"` // Maximum retry count (default 3)
}

// DomainMatcherConfig is the domain matcher configuration
type DomainMatcherConfig struct {
	// Filter configuration
	EnableBloomFilter bool `json:"enable_bloom_filter"` // Enable the bloom filter (default true)
	BloomFilterSize   uint `json:"bloom_filter_size"`   // Bloom filter size (default 1000000)
	BloomFilterHash   uint `json:"bloom_filter_hash"`   // Number of hash functions (default 5)

	// Cache configuration
	EnableLRUCache bool `json:"enable_lru_cache"` // Enable the LRU cache (default true)
	LRUCacheSize   int  `json:"lru_cache_size"`   // LRU cache size (default 10000)

	// Matching configuration
	CaseSensitive   bool `json:"case_sensitive"`    // Case sensitive (default false)
	MaxDomainLength int  `json:"max_domain_length"` // Maximum domain length (default 253)
}

// PerformanceConfig is the performance configuration
type PerformanceConfig struct {
	// Goroutine pool configuration
	MaxWorkers        int `json:"max_workers"`         // Maximum worker goroutines (default 10)
	ChannelBufferSize int `json:"channel_buffer_size"` // Channel buffer size (default 100)

	// Processing configuration
	ProcessingTimeout time.Duration `json:"processing_timeout"` // Per-certificate processing timeout (default 5s)
	BatchProcessSize  int           `json:"batch_process_size"` // Batch processing size (default 50)

	// Resource limits
	MaxMemoryUsage int64   `json:"max_memory_usage"` // Maximum memory usage (MB) (default 512)
	MaxCPUUsage    float64 `json:"max_cpu_usage"`    // Maximum CPU usage (default 0.8)
}

// RedisConfig is the Redis configuration
type RedisConfig struct {

	// Key configuration
	KeyPrefix string `json:"key_prefix"` // Key prefix (default "ct_watcher:")
}

// DefaultSDKOptions returns the default SDK configuration
func DefaultSDKOptions() *SDKOptions {
	return &SDKOptions{
		LogServers: []*LogServerConfig{
			{
				URL:           "https://ct.googleapis.com/logs/argon2023/",
				BatchSize:     100,
				ParallelFetch: 1,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
		},
		Domains: []string{},
		DomainMatcher: &DomainMatcherConfig{
			EnableBloomFilter: true,
			BloomFilterSize:   1000000,
			BloomFilterHash:   5,
			EnableLRUCache:    true,
			LRUCacheSize:      10000,
			CaseSensitive:     false,
			MaxDomainLength:   253,
		},
		Performance: &PerformanceConfig{
			MaxWorkers:        10,
			ChannelBufferSize: 100,
			ProcessingTimeout: 5 * time.Second,
			BatchProcessSize:  50,
			MaxMemoryUsage:    512,
			MaxCPUUsage:       0.8,
		},
		Redis: &RedisConfig{
			KeyPrefix: "ct_watcher:",
		},
	}
}

// Validate validates the configuration
func (opts *SDKOptions) Validate() error {
	if len(opts.LogServers) == 0 {
		return &ValidationError{Field: "LogServers", Message: "at least one CT log server must be configured"}
	}

	for i, server := range opts.LogServers {
		if server.URL == "" {
			return &ValidationError{
				Field:   "LogServers",
				Index:   i,
				Message: "CT log URL must not be empty",
			}
		}

		if server.BatchSize <= 0 {
			server.BatchSize = 100
		}
		if server.ParallelFetch <= 0 {
			server.ParallelFetch = 1
		}
		if server.BufferSize <= 0 {
			server.BufferSize = 1000
		}
		if server.Timeout <= 0 {
			server.Timeout = 30 * time.Second
		}
		if server.MaxRetries < 0 {
			server.MaxRetries = 3
		}
	}

	// Validate the domain matcher configuration
	if opts.DomainMatcher == nil {
		opts.DomainMatcher = &DomainMatcherConfig{
			EnableBloomFilter: true,
			BloomFilterSize:   1000000,
			BloomFilterHash:   5,
			EnableLRUCache:    true,
			LRUCacheSize:      10000,
			CaseSensitive:     false,
			MaxDomainLength:   253,
		}
	}

	// Validate the performance configuration
	if opts.Performance == nil {
		opts.Performance = &PerformanceConfig{
			MaxWorkers:        10,
			ChannelBufferSize: 100,
			ProcessingTimeout: 5 * time.Second,
			BatchProcessSize:  50,
			MaxMemoryUsage:    512,
			MaxCPUUsage:       0.8,
		}
	}

	// Validate the Redis configuration
	if opts.Redis == nil {
		opts.Redis = &RedisConfig{
			KeyPrefix: "ct_watcher:",
		}
	}

	return nil
}

// ValidationError is a configuration validation error
type ValidationError struct {
	Field   string
	Index   int
	Message string
}

// Status is the SDK status
type Status struct {
	// Running status
	IsRunning bool          `json:"is_running"` // Whether the SDK is running
	StartTime time.Time     `json:"start_time"` // Start time
	Uptime    time.Duration `json:"uptime"`     // Uptime

	// Processing statistics
	ProcessedCertificates int64 `json:"processed_certificates"` // Total processed certificates
	DiscoveredSubdomains  int64 `json:"discovered_subdomains"`  // Total discovered subdomains
	ActiveWatchers        int   `json:"active_watchers"`        // Number of active watchers

	// Error statistics
	TotalErrors  int64    `json:"total_errors"`  // Total errors
	RecentErrors []string `json:"recent_errors"` // Recent errors

	// Resource usage
	MemoryUsage    int64   `json:"memory_usage"`    // Memory usage
	CPUUsage       float64 `json:"cpu_usage"`       // CPU usage
	GoroutineCount int     `json:"goroutine_count"` // Goroutine count

	// Watcher status
	WatcherStatuses map[string]*models.LogStatus `json:"watcher_statuses"` // Status of each watcher
}

func (e *ValidationError) Error() string {
	if e.Index >= 0 {
		return fmt.Sprintf("configuration validation failed [%s[%d]]: %s", e.Field, e.Index, e.Message)
	}
	return fmt.Sprintf("configuration validation failed [%s]: %s", e.Field, e.Message)
}
