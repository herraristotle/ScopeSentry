package app

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/app"
	"github.com/gin-gonic/gin"
)

var appService app.Service

func init() {
	appService = app.NewService()
}

// GetAppData godoc
// @Summary Get application data
// @Description Get the application data list
// @Tags Application Management
// @Accept json
// @Produce json
// @Param request body models.SearchRequest true "Request parameters"
// @Success 200 {object} response.Response{data=models.AppResponse} "Success"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 500 {object} response.Response "Server error"
// @Router /api/assets/app [post]
func GetAppData(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := appService.GetAppData(c, req)
	if err != nil {
		logger.Error(fmt.Sprintf("Get app data error: %s", err.Error()))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
