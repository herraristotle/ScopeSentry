package subdomain

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/subdomain"
	"github.com/gin-gonic/gin"
)

var subdomainService subdomain.Service

func init() {
	subdomainService = subdomain.NewService()
}

// GetSubdomains godoc
// @Summary Get subdomain list
// @Description Get the subdomain list, supporting pagination, sorting, and filtering
// @Tags Subdomain
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.Subdomain}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/subdomain [post]
func GetSubdomains(c *gin.Context) {
	var query models.SearchRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	subdomains, err := subdomainService.GetSubdomains(c, query)
	if err != nil {
		response.InternalServerError(c, "api.subdomain.list.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list": subdomains,
	}, "api.subdomain.list.success")
}
