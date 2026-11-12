// fingerprint-------------------------------------
// @file      : fingerprint.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/10/28 20:30
// -------------------------------------------

package fingerprint

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	service "github.com/Autumn-27/ScopeSentry/internal/services/fingerprint"
	"github.com/Autumn-27/ScopeSentry/internal/utils/random"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var fingerprintService service.Service

type listRequest struct {
	Search    string `json:"search" binding:"omitempty" example:"nginx"`
	PageIndex int    `json:"pageIndex" binding:"required,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"required,min=1" example:"10"`
}

type updateRequest struct {
	ID      string `json:"id" binding:"required" example:"6642b0..."`
	Content string `json:"content" binding:"required" example:"name: Nginx\ncategory: web"`
	State   bool   `json:"state"`
}

type addRequest struct {
	Content string `json:"content" binding:"required" example:"name: Nginx\ncategory: web"`
	State   bool   `json:"state" example:"1"`
}

type deleteRequest struct {
	IDs []string `json:"ids" binding:"required,min=1,dive,required"`
}

type batchAddItem struct {
	ID         string `json:"id" example:"SAS_xxxx"`
	Name       string `json:"name" example:"druid-server"`
	Content    string `json:"content" binding:"required" example:"fingerprint:\n  name: \"druid-server\"\n  category: \"Service\""`
	UpdateTime string `json:"updateTime" example:"2026-01-19 23:50"`
}

type batchAddRequest struct {
	Items []batchAddItem `json:"items" binding:"required,min=1,dive"`
}

// Fingerprint YAML fingerprint definition
type Fingerprint struct {
	Name           string   `yaml:"name"`
	Id             string   `yaml:"id"`
	Tags           []string `yaml:"tags"`
	Category       string   `yaml:"category"`
	ParentCategory string   `yaml:"parent_category"`
	Company        string   `yaml:"company"`
	Rules          []Rule   `yaml:"rules"`
}

type FingerprintYaml struct {
	Fingerprint Fingerprint `yaml:"fingerprint"`
}

// Rule rule definition
type Rule struct {
	Logic      string      `yaml:"logic"` // AND or OR
	Conditions []Condition `yaml:"conditions"`
}

// Condition condition definition (supports plain conditions and nested condition groups)
type Condition struct {
	// plain condition fields
	Location    string `yaml:"location,omitempty"`   // body, header, title, request
	MatchType   string `yaml:"match_type,omitempty"` // regex, contains, extract, active
	Pattern     string `yaml:"pattern,omitempty"`
	Group       int    `yaml:"group,omitempty"`
	SaveAs      string `yaml:"save_as,omitempty"`
	Path        string `yaml:"path,omitempty"`
	DynamicPath string `yaml:"dynamic_path,omitempty"`
	Method      string `yaml:"method,omitempty"`
	// nested condition group fields
	Logic      string      `yaml:"logic,omitempty"`      // AND or OR (used for nested groups)
	Conditions []Condition `yaml:"conditions,omitempty"` // child conditions or nested groups
}

// Data @Summary      Query fingerprint rules
// @Description  Fuzzy search by name with pagination
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      listRequest  true  "Query parameters"
// @Success      200      {object}  response.SuccessResponse{data=object{list=[]models.FingerprintRule,total=int}}
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/data [post]
func Data(c *gin.Context) {
	var req listRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	list, total, err := fingerprintService.List(c, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"list": list, "total": total}, "api.success")
}

// Update @Summary      Update fingerprint rule
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      updateRequest  true  "Data to update"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/update [post]
func Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Parse the YAML content
	var fp FingerprintYaml
	if err := yaml.Unmarshal([]byte(req.Content), &fp); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Extract fields from the parsed struct
	data := models.FingerprintRule{
		Name:           fp.Fingerprint.Name,
		Rule:           req.Content, // store the full YAML content
		Category:       fp.Fingerprint.Category,
		ParentCategory: fp.Fingerprint.ParentCategory,
		State:          req.State,
		Company:        fp.Fingerprint.Company,
	}

	if err := fingerprintService.Update(c, req.ID, data); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// Add @Summary      Add fingerprint rule
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      addRequest  true  "Data to add"
// @Success      200      {object}  response.SuccessResponse{data=object{id=string}}
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/add [post]
func Add(c *gin.Context) {
	var req addRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Parse the YAML content
	var fp FingerprintYaml
	if err := yaml.Unmarshal([]byte(req.Content), &fp); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Generate an ID and add it to the Fingerprint struct (sas_ + 8 random uppercase/lowercase letters and digits)
	FingerId := "SAA_" + random.GenerateRandomString(12)

	// Extract fields from the parsed struct
	data := models.FingerprintRule{
		Name:           fp.Fingerprint.Name,
		Rule:           req.Content, // store the full YAML content including the id
		FingerprintId:  FingerId,    // store the id field from the YAML
		Category:       fp.Fingerprint.Category,
		ParentCategory: fp.Fingerprint.ParentCategory,
		State:          req.State,
		Company:        fp.Fingerprint.Company,
	}

	id, err := fingerprintService.Add(c, data)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"id": id}, "api.success")
}

// Delete @Summary      Batch delete fingerprint rules
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      deleteRequest  true  "List of IDs to delete"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/delete [post]
func Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	_, err := fingerprintService.Delete(c, req.IDs)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// GetVersion @Summary      Get fingerprint version
// @Tags         fingerprint
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200      {object}  response.SuccessResponse{data=object{version=string}}
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/version [get]
func GetVersion(c *gin.Context) {
	version, err := fingerprintService.GetVersion(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"version": version}, "api.success")
}

// BatchAdd @Summary      Batch add fingerprint rules
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      batchAddRequest  true  "Data to add in batch"
// @Success      200      {object}  response.SuccessResponse{data=object{success=int,failed=int,results=[]object{id=int,success=bool,error=string}}}
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/batch-add [post]
func BatchAdd(c *gin.Context) {
	var req batchAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Convert types
	items := make([]service.BatchAddItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.BatchAddItem{
			ID:         item.ID,
			Name:       item.Name,
			Content:    item.Content,
			UpdateTime: item.UpdateTime,
		}
	}

	results, successCount, failedCount, err := fingerprintService.BatchAdd(c, items)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, gin.H{
		"success": successCount,
		"failed":  failedCount,
		"results": results,
	}, "api.success")
}

func init() {
	fingerprintService = service.NewService()
}
