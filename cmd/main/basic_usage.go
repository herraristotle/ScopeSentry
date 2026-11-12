//go:build ignore

// examples----------------------------------
// @file      : basic_usage.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - basic usage example
// ============================================

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/utils/ct"
	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/models"
	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/types"
)

func main() {
	// Create SDK configuration
	opts := &types.SDKOptions{
		// CT log server configuration
		LogServers: []*types.LogServerConfig{
			{
				URL:           "https://ct.googleapis.com/logs/us1/argon2026h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/us1/argon2026h2/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/us1/argon2027h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/eu1/xenon2026h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/eu1/xenon2026h2/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/eu1/xenon2027h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.cloudflare.com/logs/nimbus2026/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.cloudflare.com/logs/nimbus2027/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
		},
		// Initially watched domains
		Domains: []string{
			"example.com",
			"google.com",
			"github.com",
			"scope-sentry.top",
			"jgeai.site",
		},
		// Domain matcher configuration
		DomainMatcher: &types.DomainMatcherConfig{
			EnableBloomFilter: true,
			BloomFilterSize:   1000000,
			BloomFilterHash:   5,
			EnableLRUCache:    true,
			LRUCacheSize:      10000,
			CaseSensitive:     false,
			MaxDomainLength:   253,
		},
		// Performance configuration
		Performance: &types.PerformanceConfig{
			MaxWorkers:        10,
			ChannelBufferSize: 100,
			ProcessingTimeout: 5 * time.Second,
			BatchProcessSize:  50,
		},
		// Redis configuration
		Redis: &types.RedisConfig{
			KeyPrefix: "ct_watcher:",
		},
	}

	// Create SDK instance
	sdk, err := ct.NewCTWatcherSDK(opts)
	if err != nil {
		log.Fatalf("Failed to create SDK: %v", err)
	}
	defer sdk.Close()

	// Register subdomain discovery callback
	sdk.OnSubdomainFound(func(event models.SubdomainEvent) {
		fmt.Printf("[Subdomain found] %s -> %s | Certificate index: %d | Time: %s\n",
			event.RootDomain,
			event.Subdomain,
			event.Certificate.Index,
			event.DiscoveredAt.Format("2006-01-02 15:04:05"))
	})

	// Register error callback
	sdk.OnError(func(err error, context map[string]interface{}) {
		errorType := "unknown"
		if ctxType, ok := context["error_type"]; ok {
			if et, ok := ctxType.(string); ok {
				errorType = et
			}
		}
		fmt.Printf("[Error] Type: %s | Message: %v\n", errorType, err)
	})

	// Create context to control lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start SDK
	fmt.Println("Starting CT log listener...")
	if err := sdk.Start(ctx); err != nil {
		log.Fatalf("Failed to start SDK: %v", err)
	}

	fmt.Println("CT listener started, watching for subdomain discovery...")
	fmt.Println("Press Ctrl+C to stop listening")

	// Start status monitoring goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// Variables for calculating incremental QPS
		var lastProcessedCount int64 = 0
		var lastReportTime = time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				currentTime := time.Now()
				status := sdk.GetStatus()
				stats := sdk.GetStats()

				// Calculate overall average QPS
				uptimeSeconds := status.Uptime.Seconds()
				avgQps := float64(0)
				if uptimeSeconds > 0 {
					avgQps = float64(status.ProcessedCertificates) / uptimeSeconds
				}

				// Calculate incremental QPS (actual processing rate over the last 30 seconds)
				timeDiff := currentTime.Sub(lastReportTime).Seconds()
				countDiff := status.ProcessedCertificates - lastProcessedCount
				incrementalQps := float64(0)
				if timeDiff > 0 {
					incrementalQps = float64(countDiff) / timeDiff
				}

				// Update the previous state
				lastProcessedCount = status.ProcessedCertificates
				lastReportTime = currentTime

				fmt.Printf("\n=== Listener status report ===\n")
				fmt.Printf("Uptime: %v\n", status.Uptime)
				fmt.Printf("Total certificates processed: %d\n", status.ProcessedCertificates)
				fmt.Printf("Total subdomains discovered: %d\n", status.DiscoveredSubdomains)
				fmt.Printf("Average QPS: %.2f certs/sec\n", avgQps)
				fmt.Printf("Incremental QPS: %.2f certs/sec (last 30 seconds)\n", incrementalQps)
				fmt.Printf("Active listeners: %d\n", status.ActiveWatchers)
				fmt.Printf("Total domains: %d\n", stats.TotalDomains)

				// Show processing statistics for each listener
				if len(status.WatcherStatuses) > 0 {
					fmt.Printf("\n--- Listener details ---\n")
					totalWatcherProcessed := int64(0)
					for url, watcherStatus := range status.WatcherStatuses {
						fmt.Printf("Listener: %s\n", url)
						fmt.Printf("  Status: %s\n", watcherStatus.Status)
						fmt.Printf("  Certificates processed: %d\n", watcherStatus.ProcessedCount)
						fmt.Printf("  Last index: %d\n", watcherStatus.LastIndex)
						fmt.Printf("  Error count: %d\n", watcherStatus.ErrorCount)
						totalWatcherProcessed += watcherStatus.ProcessedCount
					}
					fmt.Printf("\nTotal certificates processed by all listeners: %d\n", totalWatcherProcessed)
				}

				fmt.Printf("========================\n\n")
			}
		}
	}()

	// Example of dynamically adding domains
	go func() {
		time.Sleep(2 * time.Minute)
		fmt.Println("Dynamically adding domain: test.com")
		if err := sdk.AddDomains([]string{"test.com"}); err != nil {
			fmt.Printf("Failed to add domain: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block waiting for signal
	<-sigChan
	fmt.Println("\nStop signal received, stopping CT listener...")

	// Cancel context to trigger graceful shutdown
	cancel()

	// Wait some time to ensure cleanup completes
	time.Sleep(5 * time.Second)

	fmt.Println("CT listener fully stopped")
}

// Usage example:
// func main() {
//     runCTExample()
// }
