// Package plugin plugin-------------------------------------
// @file      : ct_monitor.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2026/1/25 13:05
// -------------------------------------------
package plugin

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"

	"github.com/Autumn-27/ScopeSentry/internal/models"

	"github.com/Autumn-27/ScopeSentry/internal/utils/ct"
	ct_models "github.com/Autumn-27/ScopeSentry/internal/utils/ct/models"

	"github.com/Autumn-27/ScopeSentry/internal/options"
	"github.com/Autumn-27/ScopeSentry/internal/utils"
	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/types"
	"go.mongodb.org/mongo-driver/bson"
)

func GetName() string {
	return "CT Monitor"
}

// Cycle run cycle
func Cycle() string {
	return "1"
}

func Install() error {
	return nil
}

// Exclude domains we do not care about
var blacklist = []string{
	"google.com",
}

// Exclude domains containing these strings
var subdomainBlacklist = []string{
	".prod.",
}
var CTSDK *ct.CTWatcherSDK
var OP options.PluginOption

// List of already processed subdomains, used for deduplication
var (
	processedSubdomains = make(map[string]struct{})
	subdomainMu         sync.RWMutex
)

func Execute(op options.PluginOption) error {
	// Get the plugin context
	ctx := utils.GetContext(op.PluginHash)

	// Monitoring log configuration
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
			//{
			//	URL:           "https://ct.googleapis.com/logs/us1/argon2026h2/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			//{
			//	URL:           "https://ct.googleapis.com/logs/us1/argon2027h1/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			{
				URL:           "https://ct.googleapis.com/logs/eu1/xenon2026h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			//{
			//	URL:           "https://ct.googleapis.com/logs/eu1/xenon2026h2/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			//{
			//	URL:           "https://ct.googleapis.com/logs/eu1/xenon2027h1/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			//{
			//	URL:           "https://ct.cloudflare.com/logs/nimbus2026/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			//{
			//	URL:           "https://ct.cloudflare.com/logs/nimbus2027/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
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

	// Get the domains of interest, excluding the blacklist; only the domain field is returned
	filter := bson.M{
		"domain": bson.M{
			"$nin": blacklist,
		},
	}
	var rootDomains []bson.M
	err := op.CommonRepo.FindField(ctx, "RootDomain", filter, 0, []string{"domain"}, &rootDomains)
	if err != nil {
		return err
	}

	// Convert all returned domains into a string slice
	domains := make([]string, 0, len(rootDomains))
	for _, doc := range rootDomains {
		if domain, ok := doc["domain"].(string); ok {
			domains = append(domains, domain)
		}
	}
	// Set the domains of interest
	opts.Domains = domains

	// Create the SDK instance
	sdk, err := ct.NewCTWatcherSDK(opts)
	if err != nil {
		log.Fatalf("Failed to create SDK: %v", err)
	}
	defer sdk.Close()

	CTSDK = sdk
	OP = op

	// Check whether the scan template has been created
	taskTemplateId := ""
	ctTmp, err := op.TemplateService.List(context.Background(), 1, 20, "CT_Monitor")
	if err != nil {
		op.Log(fmt.Sprintf("Search Template error: %s", err.Error()), "e")
	}
	if ctTmp != nil {
		if len(ctTmp.List) == 0 {
			// Template not created yet; create it now
			template := models.ScanTemplate{
				Name:         "CT_Monitor",
				AssetMapping: []string{"3a0d994a12305cb15a5cb7104d819623"}, // httpx
				Parameters: models.Parameters{
					AssetMapping: map[string]string{
						"3a0d994a12305cb15a5cb7104d819623": "-cdncheck false -screenshot true -tlsprobe false -st 10 -fr true -bh false -t 30",
					},
				},
				ParameterLists: models.Parameters{
					AssetMapping: map[string]string{
						"3a0d994a12305cb15a5cb7104d819623": "[{\"name\":\"cdncheck\",\"type\":\"bool\",\"defaultValue\":\"false\"},{\"name\":\"screenshot\",\"type\":\"bool\",\"defaultValue\":\"true\"},{\"name\":\"tlsprobe\",\"type\":\"bool\",\"defaultValue\":\"false\"},{\"name\":\"st\",\"type\":\"string\",\"defaultValue\":\"10\"},{\"name\":\"fr\",\"type\":\"bool\",\"defaultValue\":\"true\"},{\"name\":\"bh\",\"type\":\"bool\",\"defaultValue\":\"false\"},{\"name\":\"t\",\"type\":\"string\",\"defaultValue\":\"30\"}]",
					},
				},
			}
			templateID, err := op.TemplateService.Save(context.Background(), "", &template)
			if err != nil {
				op.Log(fmt.Sprintf("Failed to create template: %s", err.Error()), "e")
				return err
			}
			taskTemplateId = templateID
		} else {
			taskTemplateId = ctTmp.List[0].ID
		}
	}

	// Register the subdomain discovery callback
	sdk.OnSubdomainFound(func(event ct_models.SubdomainEvent) {
		// Check whether the subdomain has already been processed (read lock, supports concurrent reads)
		subdomainMu.RLock()
		_, exists := processedSubdomains[event.Subdomain]
		subdomainMu.RUnlock()

		if exists {
			// Already exists; skip processing
			return
		}
		// Exclude domain
		for _, domain := range subdomainBlacklist {
			if strings.Contains(event.Subdomain, domain) {
				return
			}
		}
		// Mark as processed (write lock)
		subdomainMu.Lock()
		processedSubdomains[event.Subdomain] = struct{}{}
		subdomainMu.Unlock()

		// After a subdomain is discovered, store it in Redis (the Set type deduplicates automatically)
		subdomainKey := "ct_watcher:subdomains"
		err := op.CommonRepo.SAdd(ctx, subdomainKey, event.Subdomain)
		if err != nil {
			op.Log(fmt.Sprintf("[ERROR] Failed to store subdomain in Redis: %s, error: %v", event.Subdomain, err), "e")
		}
		op.Log(fmt.Sprintf("[Subdomain Found] %s -> %s | Certificate Index: %d | Time: %s\n",
			fmt.Sprintf("<color value=\"#ff6b6b\" bold>%s</color>", event.RootDomain),
			fmt.Sprintf("<color value=\"#ff6b6b\" bold>%s</color>", event.Subdomain),
			event.Certificate.Index,
			event.DiscoveredAt.Format("2006-01-02 15:04:05")))
		// Send notification
		notificationMsg := fmt.Sprintf("New subdomain discovered: \n%v", event.Subdomain)
		op.Notification(notificationMsg)
	})

	// Register the error callback
	sdk.OnError(func(err error, context map[string]interface{}) {
		errorType := "unknown"
		if ctxType, ok := context["error_type"]; ok {
			if et, ok := ctxType.(string); ok {
				errorType = et
			}
		}
		op.Log(fmt.Sprintf("[ERROR] Type: %s | Message: %v\n", errorType, err), "e")
	})

	// Start the SDK
	op.Log("Starting CT log listener...")
	if err := sdk.Start(ctx); err != nil {
		log.Fatalf("Failed to start SDK: %v", err)
		op.Log(fmt.Sprintf("Failed to start SDK: %v", err), "e")
		return err
	}
	op.Log("CT listener started, monitoring for subdomain discoveries...")
	// Start the status monitoring goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		// Variables used to compute the incremental QPS
		var lastProcessedCount int64 = 0
		var lastReportTime = time.Now()

		// Status report function
		reportStatus := func() {
			currentTime := time.Now()
			status := sdk.GetStatus()
			stats := sdk.GetStats()

			// Compute the overall average QPS
			uptimeSeconds := status.Uptime.Seconds()
			avgQps := float64(0)
			if uptimeSeconds > 0 {
				avgQps = float64(status.ProcessedCertificates) / uptimeSeconds
			}

			// Compute the incremental QPS (actual processing rate over the last 30 seconds)
			timeDiff := currentTime.Sub(lastReportTime).Seconds()
			countDiff := status.ProcessedCertificates - lastProcessedCount
			incrementalQps := float64(0)
			if timeDiff > 0 {
				incrementalQps = float64(countDiff) / timeDiff
			}

			// Update the previous state
			lastProcessedCount = status.ProcessedCertificates
			lastReportTime = currentTime

			// Build the status report log
			var logContent string
			logContent += "\n=== Listener Status Report ===\n"
			logContent += fmt.Sprintf("Uptime: %v\n", status.Uptime)
			logContent += fmt.Sprintf("Total certificates processed: %d\n", status.ProcessedCertificates)
			logContent += fmt.Sprintf("Total subdomains discovered: %d\n", status.DiscoveredSubdomains)
			logContent += fmt.Sprintf("Average QPS: %.2f certs/sec\n", avgQps)
			logContent += fmt.Sprintf("Incremental QPS: %.2f certs/sec (last 30 seconds)\n", incrementalQps)
			logContent += fmt.Sprintf("Active watchers: %d\n", status.ActiveWatchers)
			logContent += fmt.Sprintf("Total domains: %d\n", stats.TotalDomains)

			// Show per-watcher processing statistics
			if len(status.WatcherStatuses) > 0 {
				totalWatcherProcessed := int64(0)
				for url, watcherStatus := range status.WatcherStatuses {
					logContent += fmt.Sprintf("\nWatcher: %s\n", url)
					logContent += fmt.Sprintf("  Status: %s\n", watcherStatus.Status)
					logContent += fmt.Sprintf("  Certificates processed: %d\n", watcherStatus.ProcessedCount)
					logContent += fmt.Sprintf("  Last index: %d\n", watcherStatus.LastIndex)
					logContent += fmt.Sprintf("  Error count: %d\n", watcherStatus.ErrorCount)
					totalWatcherProcessed += watcherStatus.ProcessedCount
				}
				logContent += fmt.Sprintf("\nTotal certificates processed by watchers: %d\n", totalWatcherProcessed)
			}
			logContent += "========================\n\n"
			op.Log(logContent)
		}

		// Run one status report immediately
		reportStatus()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reportStatus()
			}
		}
	}()

	// Fetch subdomains from Redis to create tasks
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				processSubdomains(ctx, op, taskTemplateId)
			}
		}
	}()
	// Wait for the context to be canceled; monitoring ends
	<-ctx.Done()
	op.Log("CT listener stopped")
	return nil
}

// processSubdomains processes the subdomains stored in Redis
func processSubdomains(ctx context.Context, op options.PluginOption, taskTemplateId string) {
	subdomainKey := "ct_watcher:subdomains"

	// Check whether the Set is empty
	count, err := op.CommonRepo.SCard(ctx, subdomainKey)
	if err != nil {
		op.Log(fmt.Sprintf("[ERROR] Failed to check subdomain Set: %v", err), "e")
		return
	}

	// If empty, return immediately
	if count == 0 {
		op.Log("Subdomain Set in Redis is empty, skipping processing")
		return
	}

	// Retrieve all subdomains
	subdomains, err := op.CommonRepo.SMembers(ctx, subdomainKey)
	if err != nil {
		op.Log(fmt.Sprintf("[ERROR] Failed to get subdomain list: %v", err), "e")
		return
	}

	op.Log(fmt.Sprintf("Retrieved %d subdomains from Redis, preparing to create a task", len(subdomains)))
	target := ""
	for _, subdomain := range subdomains {
		if subdomain != "" {
			target += subdomain + "\n"
		}
	}
	// Create the task
	if target != "" {
		task := models.Task{
			Name:           fmt.Sprintf("[ct monitor]-%v[plugin]", helper.GetNowTimeString()),
			AllNode:        true,
			Duplicates:     "None",
			ScheduledTasks: false,
			Template:       taskTemplateId,
			TargetSource:   "general",
			Target:         target,
		}
		_, err := op.TaskCommonService.Insert(context.Background(), &task)
		if err != nil {
			op.Log(fmt.Sprintf("[CT Monitor] failed to create task: %v", err), "e")
			return
		}

		// After the task is created successfully, delete the processed subdomains from Redis
		members := make([]interface{}, len(subdomains))
		for i, subdomain := range subdomains {
			members[i] = subdomain
		}
		err = op.CommonRepo.SRem(ctx, subdomainKey, members...)
		if err != nil {
			op.Log(fmt.Sprintf("[ERROR] Failed to delete subdomains from Redis: %v", err), "e")
		} else {
			op.Log(fmt.Sprintf("Successfully deleted %d processed subdomains from Redis", len(subdomains)))
		}
	}

}

func TaskEnd(task models.Task, plgHash string) error {
	if CTSDK == nil {
		OP.Log("CTSDK is null, cannot continue...")
		return fmt.Errorf("CTSDK is null")
	}
	// Tasks created by the plugin usually already have data, so no need to create again
	if strings.Contains(task.Name, "[plugin]") {
		return nil
	}
	filter := bson.M{
		"taskName": task.Name,
	}
	var rootDomains []bson.M
	err := OP.CommonRepo.FindField(context.Background(), "RootDomain", filter, 0, []string{"domain"}, &rootDomains)
	if err != nil {
		return err
	}

	// Convert all returned domains into a string slice
	domains := make([]string, 0, len(rootDomains))
	for _, doc := range rootDomains {
		if domain, ok := doc["domain"].(string); ok {
			domains = append(domains, domain)
		}
	}
	err = CTSDK.AddDomains(domains)
	if err != nil {
		OP.Log(fmt.Sprintf("CTSDK failed to add domain: %v", err), "e")
		return err
	}
	OP.Log(fmt.Sprintf("CT monitor added %v domains of interest", len(domains)))
	return nil
}
