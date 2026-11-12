// ct/models-----------------------------------
// @file      : checkpoint.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - checkpoint and status models
// ============================================

package models

import (
	"time"
)

// Checkpoint is the checkpoint data structure
type Checkpoint struct {
	// Basic information
	URL         string    `json:"url" redis:"url"`                 // CT log URL
	LogID       string    `json:"log_id" redis:"log_id"`           // Log ID

	// Processing progress
	LastProcessedIndex int64 `json:"last_processed_index" redis:"last_processed_index"` // Last processed certificate index
	STHTreeSize        int64 `json:"sth_tree_size" redis:"sth_tree_size"`               // STH tree size

	// Metadata
	Timestamp   time.Time `json:"timestamp" redis:"timestamp"`     // Checkpoint save time
	Version     int       `json:"version" redis:"version"`         // Checkpoint version
	SDKVersion  string    `json:"sdk_version" redis:"sdk_version"` // SDK version
}

// NewCheckpoint creates a new checkpoint
func NewCheckpoint(url string) *Checkpoint {
	return &Checkpoint{
		URL:       url,
		Version:   1,
		Timestamp: time.Now(),
	}
}

// Update updates the checkpoint
func (cp *Checkpoint) Update(lastIndex, sthSize int64) {
	cp.LastProcessedIndex = lastIndex
	cp.STHTreeSize = sthSize
	cp.Timestamp = time.Now()
}

// IsValid checks whether the checkpoint is valid
func (cp *Checkpoint) IsValid() bool {
	return cp.URL != "" && cp.LastProcessedIndex >= 0
}

// CheckpointManager is the checkpoint manager interface
type CheckpointManager interface {
	// SaveCheckpoint saves a checkpoint
	SaveCheckpoint(cp *Checkpoint) error

	// LoadCheckpoint loads a checkpoint
	LoadCheckpoint(url string) (*Checkpoint, error)

	// DeleteCheckpoint deletes a checkpoint
	DeleteCheckpoint(url string) error

	// ListCheckpoints lists all checkpoints
	ListCheckpoints() ([]*Checkpoint, error)

	// CleanupExpired cleans up expired checkpoints
	CleanupExpired(olderThan time.Duration) error
}

// WatcherStatus is the watcher status
type WatcherStatus struct {
	// Basic information
	LogServer   *LogServerConfig `json:"log_server"`    // Log server configuration
	LogID       string           `json:"log_id"`         // Log ID

	// Connection status
	IsConnected bool   `json:"is_connected"`  // Whether the connection succeeded
	Status      string `json:"status"`        // Current status (connecting/connected/disconnected/error)

	// Processing statistics
	LastProcessedIndex int64     `json:"last_processed_index"` // Last processed index
	ProcessedCount     int64     `json:"processed_count"`     // Number of processed certificates
	ErrorCount         int64     `json:"error_count"`         // Error count
	RetryCount         int64     `json:"retry_count"`         // Retry count

	// Time information
	StartedAt      time.Time     `json:"started_at"`       // Start time
	LastActivity   time.Time     `json:"last_activity"`    // Last activity time
	LastErrorTime  *time.Time    `json:"last_error_time"`  // Last error time
	NextRetryTime  *time.Time    `json:"next_retry_time"`  // Next retry time

	// Performance metrics
	ProcessingRate float64 `json:"processing_rate"` // Processing rate (certificates/second)
	AverageLatency time.Duration `json:"average_latency"` // Average latency
}

// NewWatcherStatus creates a new watcher status
func NewWatcherStatus(logServer *LogServerConfig) *WatcherStatus {
	now := time.Now()
	return &WatcherStatus{
		LogServer:   logServer,
		Status:      "disconnected",
		StartedAt:   now,
		LastActivity: now,
	}
}

// UpdateStatus updates the status
func (ws *WatcherStatus) UpdateStatus(status string) {
	ws.Status = status
	ws.LastActivity = time.Now()

	if status == "connected" {
		ws.IsConnected = true
	} else if status == "disconnected" || status == "error" {
		ws.IsConnected = false
	}
}

// RecordProcessing records processing statistics
func (ws *WatcherStatus) RecordProcessing(count int64) {
	ws.ProcessedCount += count
	ws.LastActivity = time.Now()
}

// RecordError records an error
func (ws *WatcherStatus) RecordError() {
	now := time.Now()
	ws.ErrorCount++
	ws.LastErrorTime = &now
	ws.LastActivity = now
}

// SDKStatus is the overall SDK status
type SDKStatus struct {
	// Running status
	IsRunning   bool      `json:"is_running"`   // Whether the SDK is running
	StartTime   time.Time `json:"start_time"`   // Start time
	Uptime      time.Duration `json:"uptime"`   // Uptime

	// Global statistics
	TotalProcessedCertificates int64 `json:"total_processed_certificates"` // Total processed certificates
	TotalDiscoveredSubdomains  int64 `json:"total_discovered_subdomains"`  // Total discovered subdomains
	TotalErrors                int64 `json:"total_errors"`                // Total errors

	// Resource usage
	MemoryUsage int64   `json:"memory_usage"` // Memory usage (bytes)
	CPUUsage    float64 `json:"cpu_usage"`    // CPU usage (0-1)
	GoroutineCount int  `json:"goroutine_count"` // Goroutine count

	// Watcher status
	ActiveWatchers int                        `json:"active_watchers"` // Number of active watchers
	WatcherStatuses map[string]*WatcherStatus `json:"watcher_statuses"` // Status of each watcher

	// Domain statistics
	TotalDomains     int64 `json:"total_domains"`      // Total watched domains
	MatchedDomains   int64 `json:"matched_domains"`    // Number of matched domains
	MatchRate        float64 `json:"match_rate"`       // Match rate
}

// NewSDKStatus creates a new SDK status
func NewSDKStatus() *SDKStatus {
	return &SDKStatus{
		StartTime:       time.Now(),
		WatcherStatuses: make(map[string]*WatcherStatus),
	}
}

// UpdateStats updates the statistics
func (ss *SDKStatus) UpdateStats() {
	ss.Uptime = time.Since(ss.StartTime)

	// Count the active watchers
	activeCount := 0
	for _, status := range ss.WatcherStatuses {
		if status.IsConnected {
			activeCount++
		}
	}
	ss.ActiveWatchers = activeCount

	// Calculate the match rate
	if ss.TotalProcessedCertificates > 0 {
		ss.MatchRate = float64(ss.TotalDiscoveredSubdomains) / float64(ss.TotalProcessedCertificates)
	}
}

// LogStatus is the watcher status (used for API responses)
type LogStatus struct {
	LogServer      *LogServerConfig `json:"log_server"`       // Log server configuration
	IsConnected    bool             `json:"is_connected"`     // Whether the connection succeeded
	LastIndex      int64            `json:"last_index"`       // Last processed index
	ErrorCount     int64            `json:"error_count"`      // Error count
	ProcessedCount int64            `json:"processed_count"`  // Number of processed certificates
	Status         string           `json:"status"`           // Current status
}

// LogServerConfig is the log server configuration (simplified, used for status display)
type LogServerConfig struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
}

// ProcessingStats holds processing statistics
type ProcessingStats struct {
	// Time window statistics
	TimeWindow time.Duration `json:"time_window"` // Statistics time window

	// Processing metrics
	CertificatesPerSecond float64       `json:"certificates_per_second"` // Certificate processing rate
	SubdomainsPerSecond   float64       `json:"subdomains_per_second"`   // Subdomain discovery rate
	ErrorRate             float64       `json:"error_rate"`             // Error rate
	AverageLatency        time.Duration `json:"average_latency"`        // Average processing latency

	// Resource usage
	MemoryUsage           int64   `json:"memory_usage"`           // Memory usage
	CPUUsage              float64 `json:"cpu_usage"`              // CPU usage
	GCStats               GCStats `json:"gc_stats"`               // GC statistics
}

// GCStats holds GC statistics
type GCStats struct {
	NumGC      uint32        `json:"num_gc"`       // GC count
	LastGCTime time.Time     `json:"last_gc_time"` // Last GC time
	GCCPUFraction float64    `json:"gc_cpu_fraction"` // GC CPU fraction
}
