// ct-----------------------------------------
// @file      : sdk.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - main interface implementation
// ============================================

package ct

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/models"
	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/types"
)

// CTWatcherSDK is the CT log watcher SDK
type CTWatcherSDK struct {
	// Configuration
	opts *types.SDKOptions

	// Core components
	domainMatcher     *DomainMatcher
	streamProcessor   *StreamProcessor
	checkpointSaver   *AutoCheckpointSaver
	checkpointManager *RedisCheckpointManager

	// Watcher management
	watchers map[string]*CTWatcher
	wg       sync.WaitGroup
	mu       sync.RWMutex

	// State management
	isRunning int32
	status    *models.SDKStatus
	ctx       context.Context
	cancel    context.CancelFunc

	// Event handling
	eventBus *models.EventBus

	// Callbacks
	subdomainCallbacks []models.SubdomainCallback
	errorCallbacks     []models.ErrorCallback
}

// NewCTWatcherSDK creates a new CT watcher SDK
func NewCTWatcherSDK(opts *types.SDKOptions) (*CTWatcherSDK, error) {
	if opts == nil {
		opts = types.DefaultSDKOptions()
	}

	// Validate the configuration
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sdk := &CTWatcherSDK{
		opts:               opts,
		watchers:           make(map[string]*CTWatcher),
		status:             models.NewSDKStatus(),
		ctx:                ctx,
		cancel:             cancel,
		eventBus:           models.NewEventBus(),
		subdomainCallbacks: make([]models.SubdomainCallback, 0),
		errorCallbacks:     make([]models.ErrorCallback, 0),
	}

	// Initialize the Redis checkpoint manager
	checkpointManager, err := NewRedisCheckpointManager(opts.Redis.KeyPrefix)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize the Redis checkpoint manager: %w", err)
	}
	sdk.checkpointManager = checkpointManager

	// Initialize the automatic checkpoint saver
	sdk.checkpointSaver = NewAutoCheckpointSaver(checkpointManager, 10*time.Minute)

	// Initialize the domain matcher
	sdk.domainMatcher = NewDomainMatcher(opts.DomainMatcher)

	// Initialize the stream processor
	sdk.streamProcessor = NewStreamProcessor(sdk.domainMatcher, sdk.eventBus)

	// Add the initial domains
	for _, domain := range opts.Domains {
		if err := sdk.domainMatcher.AddDomain(domain); err != nil {
			// Log a warning but do not abort initialization
			continue
		}
	}

	// Register the event handlers
	sdk.eventBus.Subscribe(models.EventTypeSubdomainDiscovered, &SubdomainEventHandler{sdk: sdk})
	sdk.eventBus.Subscribe(models.EventTypeCertificateProcessed, &CertificateEventHandler{sdk: sdk})
	sdk.eventBus.Subscribe(models.EventTypeErrorOccurred, &ErrorEventHandler{sdk: sdk})

	return sdk, nil
}

// Start starts the SDK
func (sdk *CTWatcherSDK) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&sdk.isRunning, 0, 1) {
		return fmt.Errorf("SDK is already running")
	}

	sdk.status.StartTime = time.Now()
	sdk.status.IsRunning = true

	// Create the derived context
	sdkCtx, cancel := context.WithCancel(ctx)
	sdk.cancel = cancel

	// Create a watcher for each configured log server
	for _, logServer := range sdk.opts.LogServers {
		if err := sdk.addWatcher(logServer); err != nil {
			// Log the error but continue creating the other watchers
			sdk.handleError(fmt.Errorf("failed to create watcher [%s]: %w", logServer.URL, err), nil)
			continue
		}
	}

	// Start the checkpoint saver
	if err := sdk.checkpointSaver.Start(); err != nil {
		atomic.StoreInt32(&sdk.isRunning, 0)
		return fmt.Errorf("failed to start the checkpoint saver: %w", err)
	}

	if len(sdk.watchers) == 0 {
		atomic.StoreInt32(&sdk.isRunning, 0)
		sdk.checkpointSaver.Stop()
		return fmt.Errorf("no watcher was created successfully")
	}

	// Wait for context cancellation or all watchers to exit
	go func() {
		<-sdkCtx.Done()
		sdk.stopInternal()
	}()

	return nil
}

// Stop stops the SDK
func (sdk *CTWatcherSDK) Stop() error {
	if !atomic.CompareAndSwapInt32(&sdk.isRunning, 1, 0) {
		return nil
	}

	sdk.cancel()
	return nil
}

// stopInternal is the internal stop method
func (sdk *CTWatcherSDK) stopInternal() {
	sdk.mu.Lock()
	defer sdk.mu.Unlock()

	// Stop all watchers
	for url, watcher := range sdk.watchers {
		if err := watcher.Stop(); err != nil {
			sdk.handleError(fmt.Errorf("failed to stop watcher [%s]: %w", url, err), nil)
		}
	}

	// Stop the checkpoint saver
	sdk.checkpointSaver.Stop()

	// Close the checkpoint manager
	sdk.checkpointManager.Close()

	sdk.status.IsRunning = false
}

// AddDomains dynamically adds domains to watch
func (sdk *CTWatcherSDK) AddDomains(domains []string) error {
	for _, domain := range domains {
		if err := sdk.domainMatcher.AddDomain(domain); err != nil {
			return fmt.Errorf("failed to add domain [%s]: %w", domain, err)
		}
	}
	return nil
}

// RemoveDomains dynamically removes watched domains
func (sdk *CTWatcherSDK) RemoveDomains(domains []string) error {
	for _, domain := range domains {
		if err := sdk.domainMatcher.RemoveDomain(domain); err != nil {
			return fmt.Errorf("failed to remove domain [%s]: %w", domain, err)
		}
	}
	return nil
}

// GetWatchedDomains fetches the currently watched domains
func (sdk *CTWatcherSDK) GetWatchedDomains() []string {
	return sdk.domainMatcher.GetAllDomains()
}

// GetStatus fetches the SDK status
func (sdk *CTWatcherSDK) GetStatus() *types.Status {
	sdk.mu.RLock()
	defer sdk.mu.RUnlock()

	status := &types.Status{
		IsRunning:             atomic.LoadInt32(&sdk.isRunning) == 1,
		StartTime:             sdk.status.StartTime,
		Uptime:                time.Since(sdk.status.StartTime),
		ProcessedCertificates: sdk.status.TotalProcessedCertificates,
		DiscoveredSubdomains:  sdk.status.TotalDiscoveredSubdomains,
		ActiveWatchers:        len(sdk.watchers),
		RecentErrors:          make([]string, 0),
		WatcherStatuses:       make(map[string]*models.LogStatus),
	}

	// Update the watcher statuses
	for url, watcher := range sdk.watchers {
		watcherStatus := watcher.GetStatus()
		status.WatcherStatuses[url] = &models.LogStatus{
			LogServer:      watcherStatus.LogServer,
			IsConnected:    watcherStatus.IsConnected,
			LastIndex:      watcherStatus.LastProcessedIndex,
			ErrorCount:     watcherStatus.ErrorCount,
			ProcessedCount: watcherStatus.ProcessedCount,
			Status:         watcherStatus.Status,
		}
	}

	return status
}

// GetStats fetches statistics
func (sdk *CTWatcherSDK) GetStats() *models.DomainMatcherStats {
	if sdk.domainMatcher == nil {
		return &models.DomainMatcherStats{}
	}

	return sdk.domainMatcher.GetStats()
}

// OnSubdomainFound registers the subdomain discovery callback
func (sdk *CTWatcherSDK) OnSubdomainFound(callback models.SubdomainCallback) error {
	if callback == nil {
		return fmt.Errorf("callback must not be empty")
	}

	sdk.subdomainCallbacks = append(sdk.subdomainCallbacks, callback)
	return nil
}

// OnError registers the error callback
func (sdk *CTWatcherSDK) OnError(callback models.ErrorCallback) error {
	if callback == nil {
		return fmt.Errorf("callback must not be empty")
	}

	sdk.errorCallbacks = append(sdk.errorCallbacks, callback)
	return nil
}

// addWatcher adds a watcher
func (sdk *CTWatcherSDK) addWatcher(logServer *types.LogServerConfig) error {
	sdk.mu.Lock()
	defer sdk.mu.Unlock()

	if _, exists := sdk.watchers[logServer.URL]; exists {
		return fmt.Errorf("watcher already exists: %s", logServer.URL)
	}

	watcher := NewCTWatcher(
		logServer,
		sdk.opts.Performance,
		sdk.streamProcessor,
		sdk.checkpointSaver,
	)

	sdk.watchers[logServer.URL] = watcher

	// Start the watcher
	if err := watcher.Start(); err != nil {
		delete(sdk.watchers, logServer.URL)
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	return nil
}

// SubdomainEventHandler is the subdomain event handler
type SubdomainEventHandler struct {
	sdk *CTWatcherSDK
}

func (h *SubdomainEventHandler) HandleEvent(event models.Event) error {
	if subdomainEvent, ok := event.(*models.SubdomainDiscoveredEvent); ok {
		atomic.AddInt64(&h.sdk.status.TotalDiscoveredSubdomains, 1)

		// Trigger the user-registered callbacks
		for _, callback := range h.sdk.subdomainCallbacks {
			go func(cb models.SubdomainCallback) {
				defer func() {
					if r := recover(); r != nil {
						h.sdk.handleError(fmt.Errorf("subdomain callback execution failed: %v", r), nil)
					}
				}()
				cb(subdomainEvent.SubdomainEvent)
			}(callback)
		}
	}
	return nil
}

// CertificateEventHandler is the certificate processing event handler
type CertificateEventHandler struct {
	sdk *CTWatcherSDK
}

func (h *CertificateEventHandler) HandleEvent(event models.Event) error {
	if certEvent, ok := event.(*models.CertificateProcessedEvent); ok {
		atomic.AddInt64(&h.sdk.status.TotalProcessedCertificates, 1)

		// Certificate processing callbacks can be added here if needed in the future
		_ = certEvent // Currently unused, avoids a compile warning
	}
	return nil
}

// ErrorEventHandler is the error event handler
type ErrorEventHandler struct {
	sdk *CTWatcherSDK
}

func (h *ErrorEventHandler) HandleEvent(event models.Event) error {
	if errorEvent, ok := event.(*models.ErrorOccurredEvent); ok {
		atomic.AddInt64(&h.sdk.status.TotalErrors, 1)

		// Trigger the user-registered callbacks
		for _, callback := range h.sdk.errorCallbacks {
			go func(cb models.ErrorCallback) {
				defer func() {
					if r := recover(); r != nil {
						h.sdk.handleError(fmt.Errorf("error callback execution failed: %v", r), nil)
					}
				}()
				cb(fmt.Errorf(errorEvent.ErrorMessage), map[string]interface{}{
					"error_type": errorEvent.ErrorType,
					"severity":   errorEvent.Severity,
				})
			}(callback)
		}
	}
	return nil
}

// handleError handles internal errors
func (sdk *CTWatcherSDK) handleError(err error, context map[string]interface{}) {
	errorEvent := &models.ErrorOccurredEvent{
		BaseEvent: models.BaseEvent{
			Type:      models.EventTypeErrorOccurred,
			ID:        models.GenerateEventID(),
			Timestamp: time.Now(),
		},
		ErrorType:    "internal_error",
		ErrorMessage: err.Error(),
		Context:      context,
		Severity:     "medium",
		Recoverable:  true,
	}

	sdk.eventBus.Publish(errorEvent)
}

// IsRunning checks whether the SDK is running
func (sdk *CTWatcherSDK) IsRunning() bool {
	return atomic.LoadInt32(&sdk.isRunning) == 1
}

// Close closes the SDK and releases resources
func (sdk *CTWatcherSDK) Close() error {
	if sdk.IsRunning() {
		sdk.Stop()
	}

	if sdk.checkpointManager != nil {
		sdk.checkpointManager.Close()
	}

	return nil
}
