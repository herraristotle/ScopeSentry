// Package redis_log_subscriber -----------------------------
// @file      : log_subscriber.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/01/XX XX:XX
// -------------------------------------------
package redis_log_subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/services/project"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/task"
	"strings"
	"sync"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	redisdb "github.com/Autumn-27/ScopeSentry/internal/database/redis"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	// GET_LOG_NAME stores the list of node names whose logs need to be fetched
	GET_LOG_NAME []string
	logNameMutex sync.RWMutex

	// LOG_INFO stores log information for each node
	LOG_INFO     map[string][]string
	logInfoMutex sync.RWMutex
)
var taskService task.Service
var projectService project.Service

func init() {
	LOG_INFO = make(map[string][]string)
	GET_LOG_NAME = make([]string, 0)
	taskService = task.NewService()
	projectService = project.NewService()
}

// SubscribeLogChannel subscribes to the Redis log channel and processes messages
func SubscribeLogChannel() {
	channelName := "logs"
	logger.Info("Subscribed to channel", zap.String("channel", channelName))

	for {
		trySubscribe(channelName)
		// If the connection fails, wait and retry
		time.Sleep(2 * time.Second)
	}
}

// trySubscribe attempts to subscribe to a Redis channel
func trySubscribe(channelName string) {
	ctx := context.Background()
	pubsub := redisdb.Client.Subscribe(ctx, channelName)
	defer pubsub.Close()

	// Verify the subscription
	_, err := pubsub.Receive(ctx)
	if err != nil {
		logger.Error("Failed to subscribe to channel", zap.String("channel", channelName), zap.Error(err))
		return
	}

	ch := pubsub.Channel()
	logger.Info("Successfully subscribed to channel", zap.String("channel", channelName))

	for msg := range ch {
		if err := handleMessage(ctx, msg.Payload); err != nil {
			logger.Error("Failed to handle message", zap.Error(err))
			continue
		}
	}
}

var (
	logCounters   = make(map[string]int)
	logCountersMu sync.Mutex
)

// handleMessage processes a received message
func handleMessage(ctx context.Context, payload string) error {
	// Parse the JSON data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	logName, ok := data["name"].(string)
	if !ok {
		return fmt.Errorf("invalid log name in message")
	}

	logContent, ok := data["log"].(string)
	if !ok {
		return fmt.Errorf("invalid log content in message")
	}

	//// Check whether the node is in GET_LOG_NAME
	//logNameMutex.RLock()
	//inList := contains(GET_LOG_NAME, logName)
	//logNameMutex.RUnlock()
	//
	//if inList {
	//	logInfoMutex.Lock()
	//	if LOG_INFO[logName] == nil {
	//		LOG_INFO[logName] = make([]string, 0)
	//	}
	//	LOG_INFO[logName] = append(LOG_INFO[logName], logContent)
	//	logInfoMutex.Unlock()
	//}
	logger.Info(fmt.Sprintf("%s %s", logName, logContent))
	// If it contains "Register Success", check the node's tasks
	if strings.Contains(logContent, "Register Success") {
		if err := checkNodeTask(ctx, logName); err != nil {
			logger.Error("Failed to check node task", zap.String("node", logName), zap.Error(err))
		}
	}
	// Recompute progress on per-target completion too. Waiting for "Task end"
	// meant a running task reported 0% until every target had finished.
	if strings.Contains(logContent, "module end scanning the target") ||
		strings.Contains(logContent, "Task end") {
		go func() {
			err := taskService.TaskProgress(context.Background())
			if err != nil {
				return
			}
			// Task artifacts have been persisted: self-heal project asset associations and backfill AssetCount (see issue #338).
			// Scan results are persisted without a project field; if the association is not stamped at task end,
			// MCP/API filtering by project will never return data.
			if err := projectService.RefreshProjectAssociations(context.Background()); err != nil {
				logger.Error("Failed to refresh project asset associations", zap.Error(err))
			}
		}()
	}

	// Push the log to a Redis list
	logKey := fmt.Sprintf("log:%s", logName)
	if err := redisdb.Client.RPush(ctx, logKey, logContent).Err(); err != nil {
		return fmt.Errorf("failed to push log to Redis: %w", err)
	}

	logCountersMu.Lock()
	defer logCountersMu.Unlock()

	// Update the local counter
	logCounters[logName]++
	if logCounters[logName] > config.GlobalConfig.Log.TotalLogs {
		// Over the limit: clear Redis directly and reset the local counter
		if err := redisdb.Client.Del(ctx, logKey).Err(); err != nil {
			logger.Error("Failed to delete log key", zap.String("key", logKey), zap.Error(err))
		}
		logCounters[logName] = 0
	}

	return nil
}

// checkNodeTask checks the node's tasks and pushes task data
func checkNodeTask(ctx context.Context, nodeName string) error {
	// Query unfinished tasks with status 1
	filter := bson.M{
		"progress": bson.M{"$ne": 100},
		"status":   1,
		"$or": []bson.M{
			{"node": nodeName},
			{"allNode": true},
		},
	}

	opts := options.Find()
	cursor, err := mongodb.DB.Collection("task").Find(ctx, filter, opts)
	if err != nil {
		return fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return fmt.Errorf("failed to decode tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	// Get the task service
	taskService := common.NewService()

	// Process each task
	for _, taskDoc := range tasks {
		// Get the task template data
		scanTemplate, err := taskService.GetScanTemplate(context.Background(), &taskDoc)
		if err != nil {
			logger.Error("Failed to get scan template", zap.String("task_id", taskDoc.ID.Hex()), zap.Error(err))
			continue
		}

		// Serialize the task data to JSON
		taskDataJSON, err := json.Marshal(scanTemplate)
		if err != nil {
			logger.Error("Failed to marshal task data", zap.String("task_id", taskDoc.ID.Hex()), zap.Error(err))
			continue
		}

		// Push to Redis
		nodeTaskKey := fmt.Sprintf("NodeTask:%s", nodeName)
		if err := redisdb.Client.RPush(ctx, nodeTaskKey, taskDataJSON).Err(); err != nil {
			logger.Error("Failed to push node task", zap.String("node", nodeName), zap.Error(err))
			continue
		}
	}

	return nil
}

// contains checks whether a string slice contains the given string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// AddLogName adds a node name whose logs need to be fetched
func AddLogName(name string) {
	logNameMutex.Lock()
	defer logNameMutex.Unlock()

	if !contains(GET_LOG_NAME, name) {
		GET_LOG_NAME = append(GET_LOG_NAME, name)
	}
}

// RemoveLogName removes a node name whose logs need to be fetched
func RemoveLogName(name string) {
	logNameMutex.Lock()
	defer logNameMutex.Unlock()

	for i, n := range GET_LOG_NAME {
		if n == name {
			GET_LOG_NAME = append(GET_LOG_NAME[:i], GET_LOG_NAME[i+1:]...)
			break
		}
	}
}

// GetLogInfo gets the log information of a node
func GetLogInfo(name string) []string {
	logInfoMutex.RLock()
	defer logInfoMutex.RUnlock()

	logs := LOG_INFO[name]
	if logs == nil {
		return []string{}
	}
	return logs
}

// PopLogInfo pops the log information of a node
func PopLogInfo(name string) string {
	logInfoMutex.Lock()
	defer logInfoMutex.Unlock()

	logs := LOG_INFO[name]
	if logs == nil || len(logs) == 0 {
		return ""
	}

	log := logs[0]
	LOG_INFO[name] = logs[1:]
	return log
}
