package task

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/template"
	"github.com/gin-gonic/gin"
)

var templateService template.Service

func init() {
	templateService = template.NewService()
}

// ListRequest template list request
type ListRequest struct {
	PageIndex int    `json:"pageIndex" binding:"required" example:"1"`
	PageSize  int    `json:"pageSize" binding:"required" example:"10"`
	Query     string `json:"query" example:"test"`
}

// TemplateList gets the template list
// @Summary Get template list
// @Description Get the scan template list
// @Tags Template management
// @Accept json
// @Produce json
// @Param request body ListRequest true "Request parameters"
// @Success 200 {object} response.Response{data=task.TemplateList} "Success"
// @Router /api/task/template [post]
func TemplateList(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := templateService.List(c.Request.Context(), req.PageIndex, req.PageSize, req.Query)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// DetailRequest template detail request
type DetailRequest struct {
	ID string `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
}

// TemplateDetail gets template details
// @Summary Get template details
// @Description Get the details of a scan template
// @Tags Template management
// @Accept json
// @Produce json
// @Param request body DetailRequest true "Request parameters"
// @Success 200 {object} response.Response{data=task.Template} "Success"
// @Router /api/task/template/detail [post]
func TemplateDetail(c *gin.Context) {
	var req DetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := templateService.Detail(c, req.ID)
	if err != nil {
		logger.Error(fmt.Sprintf("TemplateDetail error", err))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// SaveRequest template save request
type SaveRequest struct {
	ID     string              `json:"id" example:"507f1f77bcf86cd799439011"`
	Result models.ScanTemplate `json:"result" binding:"required"`
}

// Save saves a template
// @Summary Save template
// @Description Save or update a scan template
// @Tags Template management
// @Accept json
// @Produce json
// @Param request body SaveRequest true "Request parameters"
// @Success 200 {object} response.Response "Success"
// @Router /api/task/template/save [post]
func Save(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	_, err := templateService.Save(c.Request.Context(), req.ID, &req.Result)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// DeleteRequest template delete request
type DeleteRequest struct {
	IDs []string `json:"ids" binding:"required" example:"['507f1f77bcf86cd799439011']"`
}

// Delete deletes a template
// @Summary Delete template
// @Description Delete a scan template
// @Tags Template management
// @Accept json
// @Produce json
// @Param request body DeleteRequest true "Request parameters"
// @Success 200 {object} response.Response "Success"
// @Router /api/task/template/delete [post]
func Delete(c *gin.Context) {
	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	err := templateService.Delete(c, req.IDs)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}
