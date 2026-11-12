package root_domain

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/root_domain"
	"github.com/gin-gonic/gin"
)

var rootDomainService root_domain.Service

func init() {
	rootDomainService = root_domain.NewService()
}

// GetRootDomainData godoc
// @Summary Get root domain data
// @Description Get the root domain data list with pagination
// @Tags Root Domain Management
// @Accept json
// @Produce json
// @Param request body GetRootDomainDataRequest true "Request parameters"
// @Success 200 {object} response.Response{data=map[string]interface{}} "Success"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 500 {object} response.Response "Internal server error"
// @Security ApiKeyAuth
// @Router /api/assets/root_domain [post]
func GetRootDomainData(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := rootDomainService.GetRootDomainData(c, req)
	if err != nil {
		logger.Error(fmt.Sprintf("Get root domain data error: %s", err.Error()))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
