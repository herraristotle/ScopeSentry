// subdomain-------------------------------------
// @file      : taker.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/14 21:53
// -------------------------------------------

package subdomain

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/subdomain"
	"github.com/gin-gonic/gin"
)

var takerService subdomain.TakerService

func init() {
	takerService = subdomain.NewTakerService()
}

// GetSubdomainTakerData godoc
// @Summary Get subdomain takeover data
// @Description Get the subdomain takeover data list
// @Tags Subdomain Takeover
// @Accept json
// @Produce json
// @Param request body subdomain.SubdomainTakerRequest true "Request parameters"
// @Success 200 {object} response.Response{data=subdomain.SubdomainTakerResponse} "Success"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 500 {object} response.Response "Server error"
// @Router /api/assets/subdomain/taker [post]
func GetSubdomainTakerData(c *gin.Context) {
	var request models.SearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := takerService.GetSubdomainTakerData(c, request)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
