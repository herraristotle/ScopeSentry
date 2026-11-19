// configuration-------------------------------------
// @file      : searchapi.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2026/11/19 10:00
// -------------------------------------------

package configuration

import (
	"encoding/json"

	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/services/configuration"
	"github.com/gin-gonic/gin"
)

var searchApiConfigService *configuration.Service

// searchApiRequest is the credential document for the asset-search plugins.
type searchApiRequest struct {
	Fofa struct {
		Key string `json:"key"`
	} `json:"fofa"`
	Shodan struct {
		Key string `json:"key"`
	} `json:"shodan"`
}

// GetSearchApiData @Summary Get asset-search credentials
// @Tags        configuration
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} response.SuccessResponse{data=object{fofa=object{key=string},shodan=object{key=string}}}
// @Router      /configuration/searchapi/data [get]
func GetSearchApiData(c *gin.Context) {
	content, err := searchApiConfigService.GetSearchApiContent(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	// Hand the client a parsed document so the form can bind directly, and fall
	// back to an empty shape when the record is missing or malformed.
	var cfg searchApiRequest
	if content != "" {
		if err := json.Unmarshal([]byte(content), &cfg); err != nil {
			cfg = searchApiRequest{}
		}
	}
	response.Success(c, gin.H{"fofa": cfg.Fofa, "shodan": cfg.Shodan}, "api.success")
}

// SaveSearchApiData @Summary Save asset-search credentials
// @Tags        configuration
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body searchApiRequest true "FOFA and Shodan keys"
// @Success     200 {object} response.SuccessResponse
// @Router      /configuration/searchapi/save [post]
func SaveSearchApiData(c *gin.Context) {
	var req searchApiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	if err := searchApiConfigService.SaveSearchApiContent(c, string(encoded)); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func init() {
	searchApiConfigService = configuration.NewService()
}
