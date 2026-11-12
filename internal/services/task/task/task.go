// Package task -----------------------------
// @file      : task.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/4 22:04
// -------------------------------------------
package task

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/services/plugin"
	"math"
	"strings"

	"github.com/Autumn-27/ScopeSentry/internal/logger"

	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/repositories/task/task"
	schedulerCore "github.com/Autumn-27/ScopeSentry/internal/scheduler"
	"github.com/Autumn-27/ScopeSentry/internal/services/node"
	commonTask "github.com/Autumn-27/ScopeSentry/internal/services/task/common"
	schedulerSvc "github.com/Autumn-27/ScopeSentry/internal/services/task/scheduler"
	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service defines the task service interface
type Service interface {
	List(ctx *gin.Context, search string, pageIndex, pageSize int) ([]models.Task, int64, error)
	CheckTaskNameExists(ctx *gin.Context, name string) (bool, error)
	GetTaskDetail(ctx *gin.Context, taskID string) (*models.Task, error)
	DeleteTasks(ctx *gin.Context, ids []string, delA bool) error
	RetestTask(ctx context.Context, id string) error
	StopTasks(ctx *gin.Context, ids []string) error
	StartTasks(ctx *gin.Context, ids []string) error
	GetTaskProgress(ctx *gin.Context, taskID string, pageIndex, pageSize int) (map[string]interface{}, error)
	TaskProgressNumber(ctx *gin.Context) error
	TaskProgress(ctx context.Context) error
	ProcessTaskProgress(ctx context.Context, task models.Task) error
	GetAllTaskNames(ctx *gin.Context) ([]models.Task, error)
}

// service implements the Service interface
type service struct {
	taskRepo         task.Repository
	commonService    commonTask.Service
	schedulerService schedulerSvc.Service
	nodeService      node.Service
	pluginService    plugin.Service
}

// NewService creates a task service instance
func NewService() Service {
	return &service{
		taskRepo:         task.NewRepository(),
		commonService:    commonTask.NewService(),
		schedulerService: schedulerSvc.NewService(),
		nodeService:      node.NewService(),
		pluginService:    plugin.NewService(),
	}
}

// List gets the task list
func (s *service) List(ctx *gin.Context, search string, pageIndex, pageSize int) ([]models.Task, int64, error) {
	filter := bson.M{}
	if search != "" {
		filter["name"] = bson.M{"$regex": search, "$options": "i"}
	}

	// Compute the total count
	total, err := s.taskRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	projection := bson.M{
		"_id":       1,
		"status":    1,
		"name":      1,
		"taskNum":   1,
		"progress":  1,
		"creatTime": 1,
		"endTime":   1,
	}

	// Paginated query
	opts := options.Find().
		SetProjection(projection).
		SetSkip(int64((pageIndex - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "creatTime", Value: -1}})

	tasks, err := s.taskRepo.Find(ctx.Request.Context(), filter, opts)
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// CheckTaskNameExists checks whether a task name already exists
func (s *service) CheckTaskNameExists(ctx *gin.Context, name string) (bool, error) {
	return s.taskRepo.CheckTaskNameExists(ctx, name)
}

// GetTaskDetail gets task details
func (s *service) GetTaskDetail(ctx *gin.Context, taskID string) (*models.Task, error) {
	// Convert the string ID to an ObjectID
	objectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	// Query the task details
	task, err := s.taskRepo.FindByID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTasks batch-deletes tasks
func (s *service) DeleteTasks(ctx *gin.Context, ids []string, delA bool) error {
	// 1. Remove scheduled tasks and configuration
	// First try removing the task from the custom scheduler
	for _, id := range ids {
		if id == "" {
			continue
		}
		// Remove the gocron scheduled task and the scheduled collection record
		_ = schedulerCore.GetGlobalScheduler().RemoveJob(id)

		// Clean up the Redis cache
		_ = s.taskRepo.ClearTaskCache(ctx, id)
	}

	// Delete the configuration in the ScheduledTasks collection
	err := s.schedulerService.DeleteScheduledTasks(ctx, ids)
	if err != nil {
		logger.Error(fmt.Sprintf("error while deleting tasks: %v", err))
	}

	// 2. Optional: delete assets (by taskName)
	if delA {
		// Get the list of task names
		var objIDs []primitive.ObjectID
		for _, id := range ids {
			if oid, err := primitive.ObjectIDFromHex(id); err == nil {
				objIDs = append(objIDs, oid)
			}
		}
		if len(objIDs) > 0 {
			// Query only the name field
			opts := options.Find().SetProjection(bson.M{"name": 1})
			tasks, err := s.taskRepo.Find(ctx.Request.Context(), bson.M{"_id": bson.M{"$in": objIDs}}, opts)
			if err == nil {
				nameSet := make(map[string]struct{})
				for _, t := range tasks {
					if t.Name != "" {
						nameSet[t.Name] = struct{}{}
					}
				}
				if len(nameSet) > 0 {
					var names []string
					for n := range nameSet {
						names = append(names, n)
					}
					// Broadcast to all nodes for follow-up cleanup (compatible with the original refresh_config behavior)
					_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "delete_task", Content: strings.Join(ids, ",")})

					// Delete asset collection data by taskName
					// Note: in some collections taskName is an array field, which $in can also match
					collections := []string{"asset", "DirScanResult", "SensitiveResult", "SubdomainTakerResult", "UrlScan", "crawler", "subdomain", "vulnerability", "PageMonitoring", "app", "RootDomain", "mp", "IPAsset"}
					filter := bson.M{"taskName": bson.M{"$in": names}}
					// Execute the batch delete directly through the underlying database
					db := mongodb.DB
					for _, coll := range collections {
						_, _ = db.Collection(coll).DeleteMany(ctx.Request.Context(), filter)
					}
				}
			}
		}
	} else {
		// Still issue a node config refresh to stay consistent with the Python side
		_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "delete_task", Content: strings.Join(ids, ",")})
	}

	// 3. Delete the tasks in the task collection
	var objIDs2 []primitive.ObjectID
	for _, id := range ids {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs2 = append(objIDs2, oid)
		}
	}
	if len(objIDs2) == 0 {
		return nil
	}
	if _, err := s.taskRepo.DeleteMany(ctx, objIDs2); err != nil {
		return err
	}
	return nil
}

// RetestTask retests a task
func (s *service) RetestTask(ctx context.Context, id string) error {
	// Get the task
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	t, err := s.taskRepo.FindByID(ctx, objectID)
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("task not found")
	}

	// Create the scan task
	if _, err := s.commonService.CreateTaskScan(ctx, *t, id, false); err != nil {
		return err
	}

	// Update the task status
	update := bson.M{"$set": bson.M{
		"progress":  0,
		"creatTime": helper.GetNowTimeString(),
		"endTime":   "",
		"status":    1,
	}}
	if err := s.taskRepo.UpdateOne(ctx, bson.M{"_id": objectID}, update); err != nil {
		return err
	}
	return nil
}

// StopTasks stops tasks: broadcast the stop command and set status to 2
func (s *service) StopTasks(ctx *gin.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		// Broadcast the task stop
		_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "stop_task", Content: id})
		// Set the status to 2
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			_ = s.taskRepo.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"status": 2}})
		}
	}
	return nil
}

// StartTasks starts tasks: resume scanning from a pause
func (s *service) StartTasks(ctx *gin.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		t, err := s.taskRepo.FindByID(ctx, oid)
		if err != nil || t == nil {
			continue
		}
		if t.Progress == 100 {
			// Already finished; do not start again
			continue
		}
		t.IsStart = true
		if _, err := s.commonService.CreateTaskScan(ctx, *t, id, true); err != nil {
			continue
		}
		_ = s.taskRepo.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"status": 1}})
	}
	return nil
}

// GetTaskProgress gets task progress information
func (s *service) GetTaskProgress(ctx *gin.Context, taskID string, pageIndex, pageSize int) (map[string]interface{}, error) {
	// 1. Get basic task information
	objectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task ID: %w", err)
	}

	task, err := s.taskRepo.FindByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// 2. Process the target list and pagination
	targets, err := s.processTargetsWithPagination(ctx, task.Target, task.Ignore, pageIndex, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to process targets: %w", err)
	}

	// 3. Get progress data from Redis
	progressData, err := s.getProgressFromRedis(ctx, taskID, targets)
	if err != nil {
		return nil, fmt.Errorf("failed to get progress data: %w", err)
	}

	return map[string]interface{}{
		"list":  progressData,
		"total": task.TaskNum,
	}, nil
}

// processTargetsWithPagination processes the target list and paginates it
func (s *service) processTargetsWithPagination(ctx *gin.Context, targetStr, ignoreStr string, pageIndex, pageSize int) ([]string, error) {
	if targetStr == "" {
		return []string{}, nil
	}

	// Split the target string
	targetLines := strings.Split(targetStr, "\n")
	var allTargets []string

	// Process each target line
	for _, line := range targetLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Use the helper to get the target list (applying ignore rules)
		targets, err := helper.GetTargetList(line, ignoreStr)
		if err != nil {
			continue
		}
		allTargets = append(allTargets, targets...)
	}

	// Paginate
	startIndex := (pageIndex - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= len(allTargets) {
		return []string{}, nil
	}

	if endIndex > len(allTargets) {
		endIndex = len(allTargets)
	}

	return allTargets[startIndex:endIndex], nil
}

// getProgressFromRedis gets progress data from Redis
func (s *service) getProgressFromRedis(ctx *gin.Context, taskID string, targets []string) ([]models.TaskProgressDetail, error) {
	var progressList []models.TaskProgressDetail

	for _, target := range targets {
		progress := s.createProgressResult(target)

		// Get the progress information for this target from Redis
		redisKey := fmt.Sprintf("TaskInfo:progress:%s:%s", taskID, target)
		progressData, err := s.taskRepo.GetProgressFromRedis(ctx, redisKey)
		if err == nil && progressData != nil {
			s.populateProgressFromRedis(progress, progressData)
		}

		progressList = append(progressList, *progress)
	}

	return progressList, nil
}

// createProgressResult creates the progress result struct
func (s *service) createProgressResult(target string) *models.TaskProgressDetail {
	return &models.TaskProgressDetail{
		ID:                  primitive.NewObjectID(),
		TargetHandler:       []string{"", ""},
		SubdomainScan:       []string{"", ""},
		SubdomainSecurity:   []string{"", ""},
		PortScanPreparation: []string{"", ""},
		PortScan:            []string{"", ""},
		PortFingerprint:     []string{"", ""},
		AssetMapping:        []string{"", ""},
		AssetHandle:         []string{"", ""},
		URLScan:             []string{"", ""},
		WebCrawler:          []string{"", ""},
		URLSecurity:         []string{"", ""},
		DirScan:             []string{"", ""},
		VulnerabilityScan:   []string{"", ""},
		All:                 []string{"", ""},
		Target:              target,
		Node:                "",
	}
}

// populateProgressFromRedis fills in progress information from Redis data
func (s *service) populateProgressFromRedis(progress *models.TaskProgressDetail, redisData map[string]string) {
	// Map Redis fields to the progress struct
	fieldMappings := map[string]*[]string{
		"TargetHandler_start":       &progress.TargetHandler,
		"SubdomainScan_start":       &progress.SubdomainScan,
		"SubdomainSecurity_start":   &progress.SubdomainSecurity,
		"PortScanPreparation_start": &progress.PortScanPreparation,
		"PortScan_start":            &progress.PortScan,
		"PortFingerprint_start":     &progress.PortFingerprint,
		"AssetMapping_start":        &progress.AssetMapping,
		"AssetHandle_start":         &progress.AssetHandle,
		"URLScan_start":             &progress.URLScan,
		"WebCrawler_start":          &progress.WebCrawler,
		"URLSecurity_start":         &progress.URLSecurity,
		"DirScan_start":             &progress.DirScan,
		"VulnerabilityScan_start":   &progress.VulnerabilityScan,
		"scan_start":                &progress.All,
	}

	// Fill in the start time
	for redisField, progressField := range fieldMappings {
		if startTime, exists := redisData[redisField]; exists {
			(*progressField)[0] = startTime
		}
	}

	// Fill in the end time
	endFieldMappings := map[string]*[]string{
		"TargetHandler_end":       &progress.TargetHandler,
		"SubdomainScan_end":       &progress.SubdomainScan,
		"SubdomainSecurity_end":   &progress.SubdomainSecurity,
		"PortScanPreparation_end": &progress.PortScanPreparation,
		"PortScan_end":            &progress.PortScan,
		"PortFingerprint_end":     &progress.PortFingerprint,
		"AssetMapping_end":        &progress.AssetMapping,
		"AssetHandle_end":         &progress.AssetHandle,
		"URLScan_end":             &progress.URLScan,
		"WebCrawler_end":          &progress.WebCrawler,
		"URLSecurity_end":         &progress.URLSecurity,
		"DirScan_end":             &progress.DirScan,
		"VulnerabilityScan_end":   &progress.VulnerabilityScan,
		"scan_end":                &progress.All,
	}

	for redisField, progressField := range endFieldMappings {
		if endTime, exists := redisData[redisField]; exists {
			(*progressField)[1] = endTime
		}
	}

	// Set node information
	if node, exists := redisData["node"]; exists {
		progress.Node = node
	}
}

func (s *service) TaskProgressNumber(ctx *gin.Context) error {
	filter := bson.M{
		"progress": bson.M{"$ne": 100},
		"status":   1,
	}
	// Call the repository
	tasks, err := s.taskRepo.Find(ctx.Request.Context(), filter, nil)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		fmt.Printf("task ID: %s, progress: %d, status: %d\n", task.ID.Hex(), task.Progress, task.Status)
		id := task.ID.Hex()
		key := fmt.Sprintf("TaskInfo:tmp:%s", id)
		fmt.Printf("<UNK>ID: %s\n", key)
	}
	return nil
}

// TaskProgress updates task progress
func (s *service) TaskProgress(ctx context.Context) error {
	// Query all tasks whose progress is not 100 and whose status is 1 (running)
	filter := bson.M{
		"progress": bson.M{"$ne": 100},
		"status":   1,
	}

	tasks, err := s.taskRepo.Find(ctx, filter, nil)
	if err != nil {
		return fmt.Errorf("failed to find running tasks: %w", err)
	}

	// If no tasks are running, return immediately
	if len(tasks) == 0 {
		return nil
	}

	// Process each task
	for _, task := range tasks {
		if err := s.ProcessTaskProgress(ctx, task); err != nil {
			logger.Error(fmt.Sprintf("failed to process task progress for task %s: %v", task.ID.Hex(), err))
			continue
		}
	}

	return nil
}

// processTaskProgress handles the progress update for a single task
func (s *service) ProcessTaskProgress(ctx context.Context, task models.Task) error {
	taskID := task.ID.Hex()
	tmpKey := fmt.Sprintf("TaskInfo:tmp:%s", taskID)
	timeKey := fmt.Sprintf("TaskInfo:time:%s", taskID)

	// Check whether a temporary statistics key exists in Redis
	exists, err := s.taskRepo.Exists(ctx, tmpKey)
	if err != nil {
		return fmt.Errorf("failed to check tmp key existence: %w", err)
	}

	if exists {
		// Get the number of elements in the collection
		count, err := s.taskRepo.SCard(ctx, tmpKey)
		if err != nil {
			return fmt.Errorf("failed to get set cardinality: %w", err)
		}

		// Compute the progress
		progressTmp := float64(count) / float64(task.TaskNum)
		progressTmp = math.Round(progressTmp*100*10) / 10 // keep one decimal place

		// Ensure the progress does not exceed 100
		if progressTmp > 100 {
			progressTmp = 100
		}

		// If progress reaches 100, update the task status to completed
		if progressTmp == 100 {
			// Get the end time
			endTime, err := s.taskRepo.Get(ctx, timeKey)
			if err != nil {
				return fmt.Errorf("failed to get end time: %w", err)
			}

			// Update the task status to completed (status 3) and set the end time
			update := bson.M{
				"$set": bson.M{
					"endTime":  endTime,
					"status":   3,
					"progress": int(progressTmp),
				},
			}
			if err := s.taskRepo.UpdateOne(ctx, bson.M{"_id": task.ID}, update); err != nil {
				return fmt.Errorf("failed to update task status: %w", err)
			}

			// Delete the Redis statistics
			if err := s.taskRepo.Del(ctx, tmpKey, timeKey); err != nil {
				logger.Error(fmt.Sprintf("failed to delete redis keys for completed task %s: %v", taskID, err))
			}

			// Call the server-side plugin
			go func() {
				err := s.pluginService.RunTaskEnd(task)
				if err != nil {
					return
				}
			}()

		} else {
			// Update the task progress
			update := bson.M{
				"$set": bson.M{
					"progress": int(progressTmp),
				},
			}
			if err := s.taskRepo.UpdateOne(ctx, bson.M{"_id": task.ID}, update); err != nil {
				return fmt.Errorf("failed to update task progress: %w", err)
			}
		}
	} else {
		// The Redis key does not exist; set the progress to 0
		update := bson.M{
			"$set": bson.M{
				"progress": 0,
			},
		}
		if err := s.taskRepo.UpdateOne(ctx, bson.M{"_id": task.ID}, update); err != nil {
			return fmt.Errorf("failed to reset task progress: %w", err)
		}
	}

	return nil
}

// GetAllTaskNames gets all task names, ordered by id descending
func (s *service) GetAllTaskNames(ctx *gin.Context) ([]models.Task, error) {
	filter := bson.M{}

	// Query only the id and name fields
	projection := bson.M{
		"_id":  1,
		"name": 1,
	}

	// Sort by id descending
	opts := options.Find().
		SetProjection(projection).
		SetSort(bson.D{{Key: "_id", Value: -1}})

	tasks, err := s.taskRepo.Find(ctx.Request.Context(), filter, opts)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
