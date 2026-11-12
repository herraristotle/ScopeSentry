// ct/models-----------------------------------
// @file      : event.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - event and callback definitions
// ============================================

package models

import (
	"fmt"
	"time"
)

// EventType is the event type
type EventType string

const (
	// Subdomain discovery event
	EventTypeSubdomainDiscovered EventType = "subdomain_discovered"

	// Certificate processing event
	EventTypeCertificateProcessed EventType = "certificate_processed"

	// Error event
	EventTypeErrorOccurred EventType = "error_occurred"

	// Watcher status event
	EventTypeWatcherStatusChanged EventType = "watcher_status_changed"

	// SDK lifecycle events
	EventTypeSDKStarted EventType = "sdk_started"
	EventTypeSDKStopped EventType = "sdk_stopped"
)

// Event is the base event interface
type Event interface {
	GetType() EventType
	GetTimestamp() time.Time
	GetID() string
}

// BaseEvent is the base event structure
type BaseEvent struct {
	Type      EventType `json:"type"`      // Event type
	ID        string    `json:"id"`        // Unique event ID
	Timestamp time.Time `json:"timestamp"` // Event timestamp
}

// GetType implements the Event interface
func (e *BaseEvent) GetType() EventType {
	return e.Type
}

// GetTimestamp implements the Event interface
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetID implements the Event interface
func (e *BaseEvent) GetID() string {
	return e.ID
}

// SubdomainDiscoveredEvent is the subdomain discovery event
type SubdomainDiscoveredEvent struct {
	BaseEvent
	SubdomainEvent SubdomainEvent `json:"subdomain_event"`
}

// SubdomainEvent is the subdomain event details
type SubdomainEvent struct {
	Subdomain    string    `json:"subdomain"`     // Discovered subdomain
	RootDomain   string    `json:"root_domain"`   // Root domain
	Certificate  *CertInfo `json:"certificate"`   // Certificate information (simplified)
	DiscoveredAt time.Time `json:"discovered_at"` // Discovery time
	EventID      string    `json:"event_id"`      // Event ID
	IsNew        bool      `json:"is_new"`        // Whether it is a new discovery
}

// CertInfo is the certificate information (simplified, only essential fields)
type CertInfo struct {
	LogURL     string    `json:"log_url"`     // CT log URL
	LogID      string    `json:"log_id"`      // Log ID
	Index      int64     `json:"index"`       // Certificate index
	Timestamp  time.Time `json:"timestamp"`   // Certificate timestamp
	CommonName string    `json:"common_name"` // Certificate CN
	AllDomains []string  `json:"all_domains"` // All domains in the certificate
}

// CertificateProcessedEvent is the certificate processing event
type CertificateProcessedEvent struct {
	BaseEvent
	Certificate    *CertInfo     `json:"certificate"`     // Certificate information
	ProcessingTime time.Duration `json:"processing_time"` // Processing time
	Success        bool          `json:"success"`         // Whether it succeeded
	Error          string        `json:"error,omitempty"` // Error message
}

// ErrorOccurredEvent is the error event
type ErrorOccurredEvent struct {
	BaseEvent
	ErrorType    string                 `json:"error_type"`    // Error type
	ErrorMessage string                 `json:"error_message"` // Error message
	Context      map[string]interface{} `json:"context"`       // Error context
	Severity     string                 `json:"severity"`      // Severity (low/medium/high/critical)
	Recoverable  bool                   `json:"recoverable"`   // Whether it is recoverable
}

// WatcherStatusChangedEvent is the watcher status change event
type WatcherStatusChangedEvent struct {
	BaseEvent
	LogURL    string `json:"log_url"`    // CT log URL
	OldStatus string `json:"old_status"` // Old status
	NewStatus string `json:"new_status"` // New status
	Reason    string `json:"reason"`     // Reason for the change
}

// SDKLifecycleEvent is the SDK lifecycle event
type SDKLifecycleEvent struct {
	BaseEvent
	Action   string         `json:"action"`             // Action (start/stop/restart)
	Reason   string         `json:"reason"`             // Reason
	Duration *time.Duration `json:"duration,omitempty"` // Duration (on stop)
	Stats    *SDKStats      `json:"stats,omitempty"`    // Current statistics
}

// SDKStats holds SDK statistics
type SDKStats struct {
	TotalProcessedCertificates int64         `json:"total_processed_certificates"` // Total processed certificates
	TotalDiscoveredSubdomains  int64         `json:"total_discovered_subdomains"`  // Total discovered subdomains
	TotalErrors                int64         `json:"total_errors"`                 // Total errors
	Uptime                     time.Duration `json:"uptime"`                       // Uptime
}

// Callback function type definitions

// SubdomainCallback is the subdomain discovery callback function
type SubdomainCallback func(event SubdomainEvent)

// CertificateCallback is the certificate processing callback function
type CertificateCallback func(cert *CertInfo, success bool, err error)

// ErrorCallback is the error callback function
type ErrorCallback func(err error, context map[string]interface{})

// StatusCallback is the status change callback function
type StatusCallback func(logURL string, oldStatus, newStatus string)

// Event handler interface definitions

// EventHandler is the event handler interface
type EventHandler interface {
	HandleEvent(event Event) error
}

// EventBus is the event bus
type EventBus struct {
	handlers map[EventType][]EventHandler
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe subscribes to an event
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish publishes an event
func (eb *EventBus) Publish(event Event) error {
	handlers, exists := eb.handlers[event.GetType()]
	if !exists {
		return nil
	}

	for _, handler := range handlers {
		if err := handler.HandleEvent(event); err != nil {
			// Record the error but do not interrupt the other handlers
			continue
		}
	}
	return nil
}

// GenerateEventID generates an event ID
func GenerateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
