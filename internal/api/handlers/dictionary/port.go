package dictionary

import (
	"fmt"

	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	dictservice "github.com/Autumn-27/ScopeSentry/internal/services/dictionary"
	"github.com/gin-gonic/gin"
)

var portService dictservice.PortService

func init() {
	portService = dictservice.NewPortService()
}

// PortDataRequest request body
// Used for /data pagination and search
// swagger:ignore
// Used only for binding
type PortDataRequest struct {
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
	Search    string `json:"search"`
}

// GetPortData get port dictionary data
// @Summary Get port dictionary
// @Tags Dictionary Management
// @Accept json
// @Produce json
// @Param request body PortDataRequest true "Request parameters"
// @Success 200 {object} response.Response{data=object{list=[]models.PortDoc,total=int}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security ApiKeyAuth
// @Router /api/dictionary/port/data [post]
func GetPortData(c *gin.Context) {
	var req PortDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	list, total, err := portService.Get(c, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"list": list, "total": total}, "api.success")
}

// UpgradePortDict update port dictionary entry
// @Summary Update port dictionary
// @Tags Dictionary Management
// @Accept json
// @Produce json
// @Param request body models.PortDoc true "Request parameters"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security ApiKeyAuth
// @Router /api/dictionary/port/upgrade [post]
func UpgradePortDict(c *gin.Context) {
	var req models.UpdatePort
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if req.ID == "" {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("id is required"))
		return
	}
	if err := portService.Update(c, req.ID, req.Name, req.Value); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"code": 200, "message": "SensitiveRule updated successfully"}, "api.success")
}

// AddPortDict add port dictionary entry
// @Summary Add port dictionary
// @Tags Dictionary Management
// @Accept json
// @Produce json
// @Param request body models.Port true "Request parameters"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security ApiKeyAuth
// @Router /api/dictionary/port/add [post]
func AddPortDict(c *gin.Context) {
	var req models.Port
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := portService.Add(c, req.Name, req.Value); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	response.Success(c, gin.H{"code": 200, "message": "Port dict added successfully"}, "api.success")
}

// DeletePortDict delete port dictionary entry
// @Summary Delete port dictionary
// @Tags Dictionary Management
// @Accept json
// @Produce json
// @Param request body object{ids=[]string} true "Request parameters"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security ApiKeyAuth
// @Router /api/dictionary/port/delete [post]
func DeletePortDict(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("ids cannot be empty"))
		return
	}
	if err := portService.Delete(c, req.IDs); err != nil {
		logger.Error(err.Error())
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"code": 200, "message": "Port dict deleted successfully"}, "api.success")
}
