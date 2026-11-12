// ct-----------------------------------------
// @file      : checkpoint.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/24 21:00
// -------------------------------------------
// CT log monitoring SDK - Redis checkpoint mechanism
// ============================================

package ct

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	sasRedis "github.com/Autumn-27/ScopeSentry/internal/database/redis"
	"github.com/Autumn-27/ScopeSentry/internal/utils/ct/models"
	"github.com/redis/go-redis/v9"
)

// RedisCheckpointManager is the Redis checkpoint manager
type RedisCheckpointManager struct {
	client    *redis.Client // uses the project-wide Redis client
	keyPrefix string
	ctx       context.Context
}

// NewRedisCheckpointManager creates a new Redis checkpoint manager
func NewRedisCheckpointManager(keyPrefix string) (*RedisCheckpointManager, error) {
	// Use the project-wide Redis client
	redisClient := sasRedis.Client
	if redisClient == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}

	// Test the connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis connection failed: %w", err)
	}

	return &RedisCheckpointManager{
		client:    redisClient,
		keyPrefix: keyPrefix,
		ctx:       ctx,
	}, nil
}

// SaveCheckpoint saves a checkpoint
func (rcm *RedisCheckpointManager) SaveCheckpoint(cp *models.Checkpoint) error {
	if cp == nil || !cp.IsValid() {
		return fmt.Errorf("invalid checkpoint data")
	}

	key := rcm.makeKey(cp.URL)
	data, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("failed to serialize the checkpoint: %w", err)
	}

	// Save to Redis (no expiration)
	err = rcm.client.Set(rcm.ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save the checkpoint to Redis: %w", err)
	}

	return nil
}

// LoadCheckpoint loads a checkpoint
func (rcm *RedisCheckpointManager) LoadCheckpoint(url string) (*models.Checkpoint, error) {
	key := rcm.makeKey(url)

	data, err := rcm.client.Get(rcm.ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			// No checkpoint found, return the default value
			return &models.Checkpoint{
				URL:                url,
				LastProcessedIndex: 0,
				Timestamp:          time.Now(),
				Version:            1,
			}, nil
		}
		return nil, fmt.Errorf("failed to load the checkpoint from Redis: %w", err)
	}

	var cp models.Checkpoint
	if err := json.Unmarshal([]byte(data), &cp); err != nil {
		return nil, fmt.Errorf("failed to deserialize the checkpoint: %w", err)
	}

	return &cp, nil
}

// DeleteCheckpoint deletes a checkpoint
func (rcm *RedisCheckpointManager) DeleteCheckpoint(url string) error {
	key := rcm.makeKey(url)

	err := rcm.client.Del(rcm.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete the checkpoint: %w", err)
	}

	return nil
}

// ListCheckpoints lists all checkpoints
func (rcm *RedisCheckpointManager) ListCheckpoints() ([]*models.Checkpoint, error) {
	pattern := rcm.keyPrefix + "checkpoint:*"

	keys, err := rcm.client.Keys(rcm.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get the checkpoint key list: %w", err)
	}

	checkpoints := make([]*models.Checkpoint, 0, len(keys))

	for _, key := range keys {
		data, err := rcm.client.Get(rcm.ctx, key).Result()
		if err != nil {
			continue // skip problematic keys
		}

		var cp models.Checkpoint
		if err := json.Unmarshal([]byte(data), &cp); err != nil {
			continue // skip checkpoints that cannot be parsed
		}

		checkpoints = append(checkpoints, &cp)
	}

	return checkpoints, nil
}

// CleanupExpired cleans up expired checkpoints
func (rcm *RedisCheckpointManager) CleanupExpired(olderThan time.Duration) error {
	checkpoints, err := rcm.ListCheckpoints()
	if err != nil {
		return fmt.Errorf("failed to get the checkpoint list: %w", err)
	}

	cutoffTime := time.Now().Add(-olderThan)
	deletedCount := 0

	for _, cp := range checkpoints {
		if cp.Timestamp.Before(cutoffTime) {
			if err := rcm.DeleteCheckpoint(cp.URL); err != nil {
				continue // skip checkpoints that fail to delete
			}
			deletedCount++
		}
	}

	return nil
}

// BatchSaveCheckpoints saves checkpoints in batches
func (rcm *RedisCheckpointManager) BatchSaveCheckpoints(checkpoints []*models.Checkpoint) error {
	if len(checkpoints) == 0 {
		return nil
	}

	pipe := rcm.client.Pipeline()

	for _, cp := range checkpoints {
		if cp == nil || !cp.IsValid() {
			continue
		}

		key := rcm.makeKey(cp.URL)
		data, err := json.Marshal(cp)
		if err != nil {
			continue // skip checkpoints that fail to serialize
		}

		pipe.Set(rcm.ctx, key, data, 0)
	}

	_, err := pipe.Exec(rcm.ctx)
	if err != nil {
		return fmt.Errorf("failed to save checkpoints in batch: %w", err)
	}

	return nil
}

// GetCheckpointStats returns the checkpoint statistics
func (rcm *RedisCheckpointManager) GetCheckpointStats() (*CheckpointStats, error) {
	checkpoints, err := rcm.ListCheckpoints()
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint stats: %w", err)
	}

	stats := &CheckpointStats{
		TotalCheckpoints: len(checkpoints),
	}

	if len(checkpoints) > 0 {
		oldest := checkpoints[0].Timestamp
		newest := checkpoints[0].Timestamp

		totalIndexes := int64(0)

		for _, cp := range checkpoints {
			if cp.Timestamp.Before(oldest) {
				oldest = cp.Timestamp
			}
			if cp.Timestamp.After(newest) {
				newest = cp.Timestamp
			}
			totalIndexes += cp.LastProcessedIndex
		}

		stats.OldestCheckpoint = oldest
		stats.NewestCheckpoint = newest
		stats.AverageIndex = totalIndexes / int64(len(checkpoints))
	}

	return stats, nil
}

// Close closes the checkpoint manager
// Note: does not close the Redis connection, because it uses the project-wide client
func (rcm *RedisCheckpointManager) Close() error {
	return nil
}

// makeKey generates a Redis key
func (rcm *RedisCheckpointManager) makeKey(url string) string {
	return rcm.keyPrefix + "checkpoint:" + url
}

// CheckpointStats holds the checkpoint statistics
type CheckpointStats struct {
	TotalCheckpoints int       `json:"total_checkpoints"`
	OldestCheckpoint time.Time `json:"oldest_checkpoint"`
	NewestCheckpoint time.Time `json:"newest_checkpoint"`
	AverageIndex     int64     `json:"average_index"`
}

// AutoCheckpointSaver is the automatic checkpoint saver
type AutoCheckpointSaver struct {
	manager     *RedisCheckpointManager
	interval    time.Duration
	checkpoints map[string]*models.Checkpoint
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	isRunning   bool
}

// NewAutoCheckpointSaver creates an automatic checkpoint saver
func NewAutoCheckpointSaver(manager *RedisCheckpointManager, interval time.Duration) *AutoCheckpointSaver {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutoCheckpointSaver{
		manager:     manager,
		interval:    interval,
		checkpoints: make(map[string]*models.Checkpoint),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts automatic saving
func (acs *AutoCheckpointSaver) Start() error {
	if acs.isRunning {
		return fmt.Errorf("the automatic saver is already running")
	}

	acs.isRunning = true
	acs.saveAllCheckpoints()
	go func() {
		ticker := time.NewTicker(acs.interval)
		defer ticker.Stop()

		for {
			select {
			case <-acs.ctx.Done():
				// Save all checkpoints before exiting
				acs.saveAllCheckpoints()
				return
			case <-ticker.C:
				// Save on schedule
				acs.saveAllCheckpoints()
			}
		}
	}()

	return nil
}

// Stop stops automatic saving
func (acs *AutoCheckpointSaver) Stop() error {
	if !acs.isRunning {
		return nil
	}

	acs.cancel()
	acs.isRunning = false

	return nil
}

// UpdateCheckpoint updates a checkpoint (in memory)
func (acs *AutoCheckpointSaver) UpdateCheckpoint(cp *models.Checkpoint) {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	if cp != nil && cp.IsValid() {
		cp.Timestamp = time.Now()
		acs.checkpoints[cp.URL] = cp
	}
}

// RemoveCheckpoint removes a checkpoint
func (acs *AutoCheckpointSaver) RemoveCheckpoint(url string) {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	delete(acs.checkpoints, url)
}

// saveAllCheckpoints saves all checkpoints to Redis
func (acs *AutoCheckpointSaver) saveAllCheckpoints() {
	acs.mu.RLock()
	checkpoints := make([]*models.Checkpoint, 0, len(acs.checkpoints))
	for _, cp := range acs.checkpoints {
		checkpoints = append(checkpoints, cp)
	}
	acs.mu.RUnlock()

	if len(checkpoints) > 0 {
		// Save in batch
		if err := acs.manager.BatchSaveCheckpoints(checkpoints); err != nil {
			// Error handling can be added here, such as retrying or alerting
		}
	}
}

// GetPendingCheckpoints returns the number of checkpoints pending save
func (acs *AutoCheckpointSaver) GetPendingCheckpoints() int {
	acs.mu.RLock()
	defer acs.mu.RUnlock()
	return len(acs.checkpoints)
}
