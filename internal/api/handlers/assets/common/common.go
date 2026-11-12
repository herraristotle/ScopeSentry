// common-------------------------------------
// @file      : common.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/10 23:33
// -------------------------------------------

package common

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/common"
	"github.com/gin-gonic/gin"
)

var commonService common.Service

func init() {
	commonService = common.NewService()
}

// DeleteData deletes data
// @Summary Delete data
// @Description Delete data from the specified collection by a list of IDs
// @Tags Common Operations
// @Accept json
// @Produce json
// @Param request body common.DeleteRequest true "Delete request parameters"
// @Success 200 {object} response.Response
// @Router /api/asset/common/delete [post]
func DeleteData(c *gin.Context) {
	var req models.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := commonService.DeleteData(c, &req); err != nil {
		switch err {
		case common.ErrInvalidCollection:
			response.NotFound(c, "api.not_found", err)
		case common.ErrNoValidIDs, common.ErrNoDocumentsDeleted:
			response.NotFound(c, "api.not_found", err)
		default:
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, nil, "api.success")
}

// AddTag adds a tag
// @Summary Add a tag
// @Description Add a tag to the specified document
// @Tags Common Operations
// @Accept json
// @Produce json
// @Param request body common.TagRequest true "Tag request parameters"
// @Success 200 {object} response.Response
// @Router /api/assets/common/add_tag [post]
func AddTag(c *gin.Context) {
	var req models.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := commonService.AddTag(c, &req); err != nil {
		switch err {
		case common.ErrInvalidRequest, common.ErrInvalidID:
			response.NotFound(c, "api.not_found", err)
		default:
			logger.Error(fmt.Sprintf("add_tag error: %v", err))
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, nil, "api.success")
}

// DeleteTag deletes a tag
// @Summary Delete a tag
// @Description Remove a tag from the specified document
// @Tags Common Operations
// @Accept json
// @Produce json
// @Param request body common.TagRequest true "Tag request parameters"
// @Success 200 {object} response.Response
// @Router /api/common/delete_tag [post]
func DeleteTag(c *gin.Context) {
	var req models.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := commonService.DeleteTag(c, &req); err != nil {
		switch err {
		case common.ErrInvalidRequest, common.ErrInvalidID:
			response.NotFound(c, "api.not_found", err)
		default:
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, nil, "api.success")
}

// UpdateStatus updates status
// @Summary Update status
// @Description Update the status of the specified document
// @Tags Common Operations
// @Accept json
// @Produce json
// @Param request body common.StatusRequest true "Status request parameters"
// @Success 200 {object} response.Response
// @Router /api/common/update_status [post]
func UpdateStatus(c *gin.Context) {
	var req models.StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := commonService.UpdateStatus(c, &req); err != nil {
		switch err {
		case common.ErrInvalidRequest, common.ErrInvalidID:
			response.NotFound(c, "api.not_found", err)
		default:
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, nil, "api.success")
}

// TotalData gets the total count
// @Summary Get total count
// @Description Get the total number of data entries in the specified collection
// @Tags Common Operations
// @Accept json
// @Produce json
// @Param request body common.TotalRequest true "Total request parameters"
// @Success 200 {object} response.Response
// @Router /api/assets/common/total [post]
func TotalData(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	count, err := commonService.TotalData(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case common.ErrInvalidCollection:
			response.NotFound(c, "api.not_found", err)
		default:
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, gin.H{"total": count}, "api.success")
}
