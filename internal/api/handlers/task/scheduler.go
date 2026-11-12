// task-------------------------------------
// @file      : scheduler.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/6/8 22:58
// -------------------------------------------

package task

import (
	"fmt"

	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/scheduler"

	"github.com/gin-gonic/gin"
)

var SchedulerService scheduler.Service

func init() {
	SchedulerService = scheduler.NewService()
}

// GetScheduledData gets the scheduled task list
// @Summary      Get scheduled task list
// @Description  Supports fuzzy search on the name field and pagination
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      models.ScheduledTaskListRequest  true  "Query parameters"
// @Success      200  {object}  response.SuccessResponse{data=object{list=[]models.Task,total=int}}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/scheduled [post]
func GetScheduledData(c *gin.Context) {
	var req models.ScheduledTaskListRequest
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

	tasks, total, err := SchedulerService.GetScheduledData(c, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	resp := map[string]interface{}{
		"list":  tasks,
		"total": total,
	}
	response.Success(c, resp, "")
}

// CreateScheduledTask creates a scheduled task
// @Summary      Create scheduled task
// @Description  Create a new scheduled task
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      models.ScheduledTaskAddRequest  true  "Task information"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/scheduled/add [post]
func CreateScheduledTask(c *gin.Context) {
	var req models.Task
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Check whether the task name already exists
	exists, err := SchedulerService.CheckScheduledTaskNameExists(c, req.Name)
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

	if err := SchedulerService.CreateScheduledTask(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// UpdateScheduledTask updates a scheduled task
// @Summary      Update scheduled task
// @Description  Update the scheduled task with the specified ID
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      models.ScheduledTaskUpdateRequest  true  "Task information"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/scheduled/update [post]
func UpdateScheduledTask(c *gin.Context) {
	var req models.Task
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := SchedulerService.UpdateScheduledTask(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// DeleteScheduledTask deletes scheduled tasks
// @Summary      Delete scheduled tasks
// @Description  Batch delete scheduled tasks
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      models.ScheduledTaskDeleteRequest  true  "Task ID list"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/scheduled/delete [post]
func DeleteScheduledTask(c *gin.Context) {
	var req models.ScheduledTaskDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("ids is required"))
		return
	}

	if err := SchedulerService.DeleteScheduledTasks(c, req.IDs); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// GetScheduledTaskDetail gets scheduled task details
// @Summary      Get scheduled task details
// @Description  Get detailed information of a scheduled task by task ID
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      models.ScheduledTaskDetailRequest  true  "Task ID"
// @Success      200  {object}  response.SuccessResponse{data=models.ScheduledTask}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/scheduled/detail [post]
func GetScheduledTaskDetail(c *gin.Context) {
	var req models.ScheduledTaskDetailRequest
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
	taskDetail, err := SchedulerService.GetScheduledTaskDetail(c, req.ID)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	if taskDetail == nil {
		response.NotFound(c, "api.task.not_found", nil)
		return
	}
	response.Success(c, taskDetail, "")
}

// UpdatePageMonitScheduledTask
// @Router       /api/task/scheduled/pagemonit/update [post]
func UpdatePageMonitScheduledTask(c *gin.Context) {
	var req models.Task
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	err := SchedulerService.UpdatePageMonitScheduledTask(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// GetPageMonitData gets the page monitoring data list
// @Summary      Get page monitoring data list
// @Description  Supports fuzzy search on the url field and pagination
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      models.PageMonitoringListRequest  true  "Query parameters"
// @Success      200  {object}  response.SuccessResponse{data=object{list=[]models.PageMonitoringTask,total=int64}}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/scheduled/pagemonit/data [post]
func GetPageMonitData(c *gin.Context) {
	var req models.PageMonitoringListRequest
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

	tasks, total, err := SchedulerService.GetPageMonitData(c, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	resp := map[string]interface{}{
		"list":  tasks,
		"total": total,
	}
	response.Success(c, resp, "")
}

// AddPageMonitTask adds a page monitoring task
// @Summary      Add page monitoring task
// @Description  Add a new page monitoring task
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      models.PageMonitoringAddRequest  true  "URL information"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/scheduled/pagemonit/add [post]
func AddPageMonitTask(c *gin.Context) {
	var req models.PageMonitoringAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if req.URL == "" {
		response.BadRequest(c, "api.task.url_null", nil)
		return
	}

	if err := SchedulerService.AddPageMonitTask(c, req.URL); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// DeletePageMonitTask deletes page monitoring tasks
// @Summary      Delete page monitoring tasks
// @Description  Batch delete page monitoring tasks
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        data  body      models.PageMonitoringDeleteRequest  true  "Task ID list"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/task/scheduled/pagemonit/delete [post]
func DeletePageMonitTask(c *gin.Context) {
	var req models.PageMonitoringDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("ids is required"))
		return
	}

	if err := SchedulerService.DeletePageMonitTasks(c, req.IDs); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}
