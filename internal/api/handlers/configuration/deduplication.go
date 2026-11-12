// configuration-------------------------------------
// @file      : deduplication.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/10/29 20:58
// -------------------------------------------

package configuration

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/configuration"
	"github.com/gin-gonic/gin"
)

var dedupService *configuration.Service

type saveDedupRequest struct {
	RunNow bool `json:"runNow"`
	// Other fields are received dynamically, hence carried by a map
}

// GetDeduplicationConfig @Summary Get deduplication config
// @Tags        configuration
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} response.SuccessResponse{data=object}
// @Router      /configuration/deduplication/config [get]
func GetDeduplicationConfig(c *gin.Context) {
	data, err := dedupService.GetDeduplicationConfig(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "api.success")
}

// SaveDeduplicationConfig @Summary Save deduplication config
// @Tags        configuration
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body map[string]interface{} true "Deduplication config"
// @Success     200 {object} response.SuccessResponse
// @Router      /configuration/deduplication/save [post]
func SaveDeduplicationConfig(c *gin.Context) {
	var body models.DepConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := dedupService.SaveDeduplicationConfig(c, body); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func init() {
	dedupService = configuration.NewService()
}
