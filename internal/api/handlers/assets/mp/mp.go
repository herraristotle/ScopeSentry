package mp

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/mp"
	"github.com/gin-gonic/gin"
)

var mpService mp.Service

func init() {
	mpService = mp.NewService()
}

// GetMPData godoc
// @Summary Get mini-program data
// @Description Get the mini-program data list
// @Tags Mini-Program Management
// @Accept json
// @Produce json
// @Param request body models.SearchRequest true "Request parameters"
// @Success 200 {object} response.Response{data=models.MPResponse} "Success"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 500 {object} response.Response "Server error"
// @Router /api/assets/mp [post]
func GetMPData(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := mpService.GetMPData(c, req)
	if err != nil {
		logger.Error(fmt.Sprintf("Get mp data error: %s", err.Error()))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
