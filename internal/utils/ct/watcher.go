// ct-----------------------------------------
// @file      : watcher.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - Scanner watcher wrapper
// ============================================

package ct

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/global"
	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/models"
	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/types"

	ct "github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/client"
	"github.com/google/certificate-transparency-go/jsonclient"
	"github.com/google/certificate-transparency-go/scanner"
	"go.uber.org/zap"
)

// CTWatcher is the CT log watcher
type CTWatcher struct {
	// Configuration
	config      *types.LogServerConfig
	performance *types.PerformanceConfig

	// Core components
	streamProcessor *StreamProcessor
	checkpointSaver *AutoCheckpointSaver

	// Scanner-related
	httpClient *http.Client
	ctClient   *client.LogClient
	scanner    *scanner.Scanner

	// State management
	status     *models.WatcherStatus
	checkpoint *models.Checkpoint
	isRunning  int32
	lastIndex  int64

	// Logger
	logger *zap.Logger

	// Context control
	ctx    context.Context
	cancel context.CancelFunc

	// Statistics
	stats struct {
		totalProcessed int64
		totalErrors    int64
		startTime      time.Time
	}
}

// NewCTWatcher creates a new CT watcher
func NewCTWatcher(
	config *types.LogServerConfig,
	performance *types.PerformanceConfig,
	streamProcessor *StreamProcessor,
	checkpointSaver *AutoCheckpointSaver,
) *CTWatcher {

	ctx, cancel := context.WithCancel(context.Background())

	watcher := &CTWatcher{
		config:          config,
		performance:     performance,
		streamProcessor: streamProcessor,
		checkpointSaver: checkpointSaver,
		status: models.NewWatcherStatus(&models.LogServerConfig{
			URL:         config.URL,
			Description: config.URL, // Use the URL as the default description
			Operator:    "Unknown",  // default operator
		}),
		logger: global.Log, // use the global logger
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize the HTTP client
	watcher.initHTTPClient()

	return watcher
}

// Start starts the watcher
func (w *CTWatcher) Start() error {
	if !atomic.CompareAndSwapInt32(&w.isRunning, 0, 1) {
		return fmt.Errorf("watcher is already running")
	}

	w.stats.startTime = time.Now()
	w.status.UpdateStatus("connecting")

	// Initialize the CT client
	if err := w.initCTClient(); err != nil {
		atomic.StoreInt32(&w.isRunning, 0)
		w.status.UpdateStatus("error")
		return fmt.Errorf("failed to initialize the CT client: %w", err)
	}

	// Load the checkpoint
	if err := w.loadCheckpoint(); err != nil {
		atomic.StoreInt32(&w.isRunning, 0)
		w.status.UpdateStatus("error")
		return fmt.Errorf("failed to load the checkpoint: %w", err)
	}

	// Create the scanner
	if err := w.createScanner(); err != nil {
		atomic.StoreInt32(&w.isRunning, 0)
		w.status.UpdateStatus("error")
		return fmt.Errorf("failed to create the scanner: %w", err)
	}

	w.status.UpdateStatus("connected")

	// Start the watch loop
	go w.watchLoop()

	return nil
}

// Stop stops the watcher
func (w *CTWatcher) Stop() error {
	if !atomic.CompareAndSwapInt32(&w.isRunning, 1, 0) {
		return nil
	}

	w.status.UpdateStatus("disconnected")

	// Cancel the context
	w.cancel()

	// Save the final checkpoint
	w.saveCheckpoint()

	return nil
}

// GetStatus returns the watcher status
func (w *CTWatcher) GetStatus() *models.WatcherStatus {
	// Update the live statistics
	w.status.ProcessedCount = atomic.LoadInt64(&w.stats.totalProcessed)
	w.status.LastProcessedIndex = atomic.LoadInt64(&w.lastIndex)
	w.status.LastActivity = time.Now()

	return w.status
}

// watchLoop is the watch loop (core method)
func (w *CTWatcher) watchLoop() {
	defer func() {
		atomic.StoreInt32(&w.isRunning, 0)
		w.status.UpdateStatus("stopped")
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		// Perform one scan
		if err := w.performScan(); err != nil {
			w.status.RecordError()
			atomic.AddInt64(&w.stats.totalErrors, 1)

			// Wait before retrying
			select {
			case <-w.ctx.Done():
				return
			case <-time.After(w.calculateRetryDelay()):
				continue
			}
		}

		// After a successful scan, wait for the next scan
		select {
		case <-w.ctx.Done():
			return
		case <-time.After(30 * time.Second): // scan every 30 seconds
			continue
		}
	}
}

// performScan performs one scan
func (w *CTWatcher) performScan() error {
	if w.scanner == nil {
		return fmt.Errorf("scanner is not initialized")
	}

	w.status.UpdateStatus("scanning")

	// Create a context with timeout
	scanCtx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()

	// Perform the scan
	err := w.scanner.Scan(scanCtx, w.foundCertCallback, w.foundPrecertCallback)
	if err != nil {
		w.status.UpdateStatus("error")
		return fmt.Errorf("scan failed: %w", err)
	}

	w.status.UpdateStatus("connected")
	return nil
}

// foundCertCallback is the certificate discovery callback
func (w *CTWatcher) foundCertCallback(rawEntry *ct.RawLogEntry) {
	w.processRawEntry(rawEntry, "X509LogEntry")
}

// foundPrecertCallback is the precertificate discovery callback
func (w *CTWatcher) foundPrecertCallback(rawEntry *ct.RawLogEntry) {
	w.processRawEntry(rawEntry, "PrecertLogEntry")
}

// processRawEntry processes a raw log entry
func (w *CTWatcher) processRawEntry(rawEntry *ct.RawLogEntry, entryType string) {
	// Parse the timestamp
	timestamp := time.Unix(int64(rawEntry.Leaf.TimestampedEntry.Timestamp)/1000, 0)

	// Update the processing index (mark as processed regardless of the outcome of any further processing)
	for {
		currentIndex := atomic.LoadInt64(&w.lastIndex)
		if rawEntry.Index <= currentIndex {
			// The index is not greater than the current index, skip the update
			break
		}
		if atomic.CompareAndSwapInt64(&w.lastIndex, currentIndex, rawEntry.Index) {
			// Index updated successfully
			break
		}
		// CAS failed, retry
	}

	// Get the data based on the certificate type
	var certData []byte
	switch entryType {
	case "X509LogEntry":
		if rawEntry.Leaf.TimestampedEntry.X509Entry == nil {
			w.logger.Warn("X509Entry is empty, skipping processing", zap.Int64("index", rawEntry.Index))
			atomic.AddInt64(&w.stats.totalProcessed, 1)
			w.status.RecordProcessing(1)
			w.updateCheckpoint(rawEntry.Index)
			return
		}
		certData = rawEntry.Leaf.TimestampedEntry.X509Entry.Data
	case "PrecertLogEntry":
		// Skip precertificate processing for now, because the domain information may be incomplete
		w.logger.Debug("skipping precertificate processing", zap.Int64("index", rawEntry.Index))
		atomic.AddInt64(&w.stats.totalProcessed, 1)
		w.status.RecordProcessing(1)
		w.updateCheckpoint(rawEntry.Index)
		return
	default:
		w.logger.Warn("unknown certificate type", zap.String("type", entryType), zap.Int64("index", rawEntry.Index))
		atomic.AddInt64(&w.stats.totalProcessed, 1)
		w.status.RecordProcessing(1)
		w.updateCheckpoint(rawEntry.Index)
		return
	}

	// Stream-process the certificate
	err := w.streamProcessor.ProcessCertificate(
		certData,
		rawEntry.Index,
		timestamp,
		w.config.URL,
	)

	atomic.AddInt64(&w.stats.totalProcessed, 1)
	w.status.RecordProcessing(1)

	if err != nil {
		atomic.AddInt64(&w.stats.totalErrors, 1)
		w.status.RecordError()
		// Update the checkpoint even if processing fails (mark as attempted)
		w.updateCheckpoint(rawEntry.Index)
		return
	}

	// Processing succeeded, update the checkpoint
	w.updateCheckpoint(rawEntry.Index)
}

// updateCheckpoint updates the checkpoint
func (w *CTWatcher) updateCheckpoint(index int64) {
	if w.checkpoint != nil {
		w.checkpoint.Update(index, 0) // STH size is temporarily set to 0
		w.checkpointSaver.UpdateCheckpoint(w.checkpoint)
	}
}

// saveCheckpoint saves the checkpoint
func (w *CTWatcher) saveCheckpoint() {
	if w.checkpointSaver != nil && w.checkpoint != nil {
		w.checkpointSaver.UpdateCheckpoint(w.checkpoint)
	}
}

// loadCheckpoint loads the checkpoint
func (w *CTWatcher) loadCheckpoint() error {
	if w.checkpointSaver == nil {
		return fmt.Errorf("checkpoint saver is not initialized")
	}

	cp, err := w.checkpointSaver.manager.LoadCheckpoint(w.config.URL)
	if err != nil {
		return err
	}

	w.checkpoint = cp

	// Decide the starting index based on the checkpoint state
	if cp.LastProcessedIndex == 0 {
		// The checkpoint loaded from Redis shows LastProcessedIndex of 0, meaning this is the first run
		if w.ctClient != nil {
			sth, err := w.ctClient.GetSTH(context.Background())
			if err != nil {
				w.logger.Warn("failed to get the STH, continuing to watch with the default index", zap.Error(err))
				// When getting the STH fails, continue running with the default index 0 instead of stopping the watch
				atomic.StoreInt64(&w.lastIndex, 0)
			} else {
				// Start watching new certificates from the current tree size
				treeSize := int64(sth.TreeSize)
				atomic.StoreInt64(&w.lastIndex, treeSize)
				w.logger.Info("first run, starting to watch from the current tree size",
					zap.Int64("tree_size", treeSize),
					zap.String("url", w.config.URL))
			}
		} else {
			// No CT client, use the default index
			atomic.StoreInt64(&w.lastIndex, 0)
			w.logger.Info("first run, starting to watch from the default index",
				zap.String("url", w.config.URL))
		}
	} else {
		// The checkpoint exists, resume from the previous position
		atomic.StoreInt64(&w.lastIndex, cp.LastProcessedIndex)
		w.logger.Info("resuming the previous run, restoring from the checkpoint",
			zap.Int64("last_index", cp.LastProcessedIndex),
			zap.String("url", w.config.URL))
	}

	return nil
}

// initHTTPClient initializes the HTTP client
func (w *CTWatcher) initHTTPClient() {
	w.httpClient = &http.Client{
		Timeout: w.config.Timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment, // explicitly use the proxy settings from environment variables
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// initCTClient initializes the CT client
func (w *CTWatcher) initCTClient() error {
	userAgent := "CT-Watcher/1.0"

	jsonClient, err := client.New(w.config.URL, w.httpClient, jsonclient.Options{
		UserAgent: userAgent,
	})
	if err != nil {
		return fmt.Errorf("failed to create the JSON client: %w", err)
	}

	w.ctClient = jsonClient
	return nil
}

// createScanner creates the Scanner instance
func (w *CTWatcher) createScanner() error {
	if w.ctClient == nil {
		return fmt.Errorf("CT client is not initialized")
	}

	scannerOptions := &scanner.ScannerOptions{
		FetcherOptions: scanner.FetcherOptions{
			BatchSize:     w.config.BatchSize,
			ParallelFetch: w.config.ParallelFetch,
			StartIndex:    w.lastIndex,
			Continuous:    true, // continuous watching mode
		},
		Matcher:     scanner.MatchAll{}, // match all certificates
		PrecertOnly: false,              // process all certificate types
		NumWorkers:  1,                  // number of scan worker goroutines
		BufferSize:  w.config.BufferSize,
	}

	w.scanner = scanner.NewScanner(w.ctClient, *scannerOptions)

	return nil
}

// calculateRetryDelay calculates the retry delay
func (w *CTWatcher) calculateRetryDelay() time.Duration {
	// Simple exponential backoff strategy
	retryCount := w.status.ErrorCount
	if retryCount > 10 {
		retryCount = 10
	}

	delay := time.Duration(retryCount) * 5 * time.Second
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}

	return delay
}

// GetStats returns the watcher statistics
func (w *CTWatcher) GetStats() *WatcherStats {
	return &WatcherStats{
		TotalProcessed: atomic.LoadInt64(&w.stats.totalProcessed),
		TotalErrors:    atomic.LoadInt64(&w.stats.totalErrors),
		Uptime:         time.Since(w.stats.startTime),
		CurrentIndex:   atomic.LoadInt64(&w.lastIndex),
	}
}

// WatcherStats holds the watcher statistics
type WatcherStats struct {
	TotalProcessed int64         `json:"total_processed"`
	TotalErrors    int64         `json:"total_errors"`
	Uptime         time.Duration `json:"uptime"`
	CurrentIndex   int64         `json:"current_index"`
}
