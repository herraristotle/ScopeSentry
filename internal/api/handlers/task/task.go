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

	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/common"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/task"
	"github.com/gin-gonic/gin"
)

var taskService task.Service
var commonService common.Service

// TaskDetailRequest task detail request
type TaskDetailRequest struct {
	ID string `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
}

// List gets the task list (pagination + fuzzy search)
// @Summary      Get task list
// @Description  Supports fuzzy search on the name field and pagination
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      ListRequest  true  "Query parameters"
// @Success      200  {object}  response.SuccessResponse{data=object{list=[]models.Task,total=int}}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task [post]
func List(c *gin.Context) {
	var req models.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	go func() {
		err := taskService.TaskProgress(context.Background())
		if err != nil {
			logger.Error(fmt.Sprintf("%v", err))
		}
	}()
	tasks, total, err := taskService.List(c, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		logger.Error(fmt.Sprintf("%v", err))
		response.InternalServerError(c, "api.task.list.error", err)
		return
	}
	resp := map[string]interface{}{
		"list":  tasks,
		"total": total,
	}
	response.Success(c, resp, "")
}

// TaskDetail gets task details
// @Summary      Get task details
// @Description  Get detailed information of a task by task ID
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      TaskDetailRequest  true  "Task ID"
// @Success      200  {object}  response.SuccessResponse{data=object{name=string,target=string,ignore=string,node=[]string,allNode=bool,scheduledTasks=bool,hour=int,duplicates=string,template=string,day=int,minute=int,project=[]string,search=string,cycleType=string}}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/detail [post]
func TaskDetail(c *gin.Context) {
	var req TaskDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Check whether the ID is provided
	if req.ID == "" {
		response.BadRequest(c, "api.task.id_missing", nil)
		return
	}

	// Get task details
	taskDetail, err := taskService.GetTaskDetail(c, req.ID)
	if err != nil {
		logger.Error(fmt.Sprintf("TaskDetail %v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}

	if taskDetail == nil {
		response.NotFound(c, "api.task.not_found", nil)
		return
	}
	response.Success(c, taskDetail, "")
}

// AddTask adds a task
// @Summary Add task
// @Description Add a new scan task
// @Tags Task
// @Accept json
// @Produce json
// @Param request body AddTaskRequest true "Task information"
// @Success 200 {object} response.Response "Success"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 500 {object} response.Response "Server error"
// @Router /api/task/add [post]
func AddTask(c *gin.Context) {
	var req models.Task
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Check whether the task name already exists
	exists, err := taskService.CheckTaskNameExists(c, req.Name)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	if exists {
		response.BadRequest(c, "api.task.name_exists", nil)
		return
	}

	// Validate required parameters
	if req.Name == "" || len(req.Node) == 0 {
		response.BadRequest(c, "api.task.target_null", nil)
		return
	}

	// Create the task
	_, err = commonService.Insert(context.Background(), &req)
	if err != nil {
		logger.Error(fmt.Sprintf("AddTask %v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}

	// If it is a scheduled task, create the scheduled task
	if req.ScheduledTasks {
		err := SchedulerService.CreateScheduledTask(c, &req)
		if err != nil {
			response.InternalServerError(c, "api.error", err)
			return
		}
	}

	response.Success(c, nil, "api.task.add_success")
}

// DeleteTaskRequest delete task request
// Supports batch deletion and optional asset deletion
// ids: list of task IDs
// delA: whether to also delete assets
type DeleteTaskRequest struct {
	IDs  []string `json:"ids" binding:"required" example:"[\"507f1f77bcf86cd799439011\",\"507f1f77bcf86cd799439012\"]"`
	DelA bool     `json:"delA" binding:"omitempty" example:"false"`
}

// RetestTaskRequest retest task request
type RetestTaskRequest struct {
	ID string `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
}

// DeleteTask deletes tasks
// @Summary      Delete tasks
// @Description  Batch delete tasks, optionally delete the assets they produced, and clean up schedules and cache
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      DeleteTaskRequest  true  "Delete parameters"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/delete [post]
func DeleteTask(c *gin.Context) {
	var req DeleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("ids is required"))
		return
	}

	if err := taskService.DeleteTasks(c, req.IDs, req.DelA); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// RetestTask retests a task
// @Summary      Retest task
// @Description  Recreate the scan task and reset its progress
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      RetestTaskRequest  true  "Task ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/retest [post]
func RetestTask(c *gin.Context) {
	var req RetestTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if req.ID == "" {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("id is required"))
		return
	}

	if err := taskService.RetestTask(c, req.ID); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// StopTaskRequest stop task request
type StopTaskRequest struct {
	IDs []string `json:"ids" binding:"required" example:"[\"507f1f77bcf86cd799439011\"]"`
}

// StartTaskRequest start task request
type StartTaskRequest struct {
	IDs []string `json:"ids" binding:"required" example:"[\"507f1f77bcf86cd799439011\"]"`
}

// StopTask stops tasks
// @Summary      Stop tasks
// @Description  Batch stop tasks, broadcast the stop to nodes and set the status to 2
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      StopTaskRequest  true  "List of task IDs"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/stop [post]
func StopTask(c *gin.Context) {
	var req StopTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("ids is required"))
		return
	}
	if err := taskService.StopTasks(c, req.IDs); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// StartTask starts tasks (resume from pause)
// @Summary      Start tasks
// @Description  Batch start tasks, resuming the scan from a paused state
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      StartTaskRequest  true  "List of task IDs"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/start [post]
func StartTask(c *gin.Context) {
	var req StartTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("ids is required"))
		return
	}
	if err := taskService.StartTasks(c, req.IDs); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// ProgressInfoRequest task progress info request
type ProgressInfoRequest struct {
	ID        string `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
	PageIndex int    `json:"pageIndex" binding:"omitempty,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"omitempty,min=1" example:"10"`
}

// ProgressInfo gets task progress info
// @Summary      Get task progress info
// @Description  Get detailed progress info for a given task, including timestamps for each scan stage
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      ProgressInfoRequest  true  "Task ID and pagination parameters"
// @Success      200  {object}  response.SuccessResponse{data=object{list=[]models.TaskProgress,total=int}}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/progress/info [post]
func ProgressInfo(c *gin.Context) {
	var req ProgressInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Set default pagination parameters
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	progressData, err := taskService.GetTaskProgress(c, req.ID, req.PageIndex, req.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, progressData, "")
}

// GetAllTaskNames gets all task names
// @Summary      Get all task names
// @Description  Return all task names, ordered by id descending
// @Tags         task
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=[]object{id=string,name=string}}
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/names [post]
func GetAllTaskNames(c *gin.Context) {
	tasks, err := taskService.GetAllTaskNames(c)
	if err != nil {
		logger.Error(fmt.Sprintf("GetAllTaskNames %v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}

	// Build the response data, containing only id and name
	result := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, map[string]interface{}{
			"id":   t.ID.Hex(),
			"name": t.Name,
		})
	}

	response.Success(c, result, "")
}

func init() {
	taskService = task.NewService()
	commonService = common.NewService()
}
