package dictionary

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	dictservice "github.com/Autumn-27/ScopeSentry/internal/services/dictionary"
	"github.com/gin-gonic/gin"
)

var manageService dictservice.ManageService

func init() {
	manageService = dictservice.NewManageService()
}

// DeleteRequest delete request
type DeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// List get all dictionary metadata
// @Summary Get dictionary file list
// @Tags Dictionary Management
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=object{list=[]DictionaryMeta}}
// @Failure 500 {object} response.Response
// @Security ApiKeyAuth
// @Router /api/dictionary/manage/list [get]
func List(c *gin.Context) {
	metas, err := manageService.List(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	resp := map[string]interface{}{
		"list": metas,
	}
	response.Success(c, resp, "api.success")
}

// Create add a dictionary file
// @Summary Add dictionary file
// @Tags Dictionary Management
// @Accept mpfd
// @Produce json
// @Param file formData file true "File"
// @Param name formData string true "Name"
// @Param category formData string true "Category, e.g. dir, subdomain"
// @Success 200 {object} response.Response{data=object{message=string,code=int}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security ApiKeyAuth
// @Router /api/dictionary/manage/create [post]
func Create(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("missing file: %w", err))
		return
	}
	name := c.PostForm("name")
	category := c.PostForm("category")
	if name == "" || category == "" {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("name and category are required"))
		return
	}

	// Read the file content
	opened, err := fileHeader.Open()
	if err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	defer opened.Close()
	content, err := io.ReadAll(opened)
	if err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := manageService.Create(c, name, category, content); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"message": "file added successfully", "code": 200}, "api.success")
}

// Download download a dictionary file
// @Summary Download dictionary file
// @Tags Dictionary Management
// @Accept json
// @Produce application/octet-stream
// @Param id query string true "File ID"
// @Security ApiKeyAuth
// @Router /api/dictionary/manage/download [get]
func Download(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("id is required"))
		return
	}

	filename, data, err := manageService.Download(c, id)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "application/octet-stream", data)
}

// Delete delete multiple dictionary files
// @Summary Delete dictionary files
// @Tags Dictionary Management
// @Accept json
// @Produce json
// @Param request body DeleteRequest true "Request parameters"
// @Success 200 {object} response.Response{data=object{message=string,code=int}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security ApiKeyAuth
// @Router /api/dictionary/manage/delete [post]
func Delete(c *gin.Context) {
	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("ids cannot be empty"))
		return
	}

	if err := manageService.Delete(c, req.IDs); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"code": 200, "message": "delete file successful"}, "api.success")
}

// Save replace the content of a specified dictionary file
// @Summary Update dictionary file
// @Tags Dictionary Management
// @Accept mpfd
// @Produce json
// @Param id query string true "File ID"
// @Param file formData file true "File"
// @Success 200 {object} response.Response{data=object{message=string,code=int}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security ApiKeyAuth
// @Router /api/dictionary/manage/save [post]
func Save(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("id is required"))
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("missing file: %w", err))
		return
	}
	var content []byte
	if err := readFileContent(fileHeader, &content); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := manageService.Save(c, id, content); err != nil {
		logger.Error(err.Error())
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"code": 200, "message": "upload successful"}, "api.success")
}

func readFileContent(fileHeader *multipart.FileHeader, out *[]byte) error {
	f, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	*out = data
	return nil
}
