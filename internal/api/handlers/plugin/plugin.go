package plugin

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/plugin"
	"github.com/Autumn-27/ScopeSentry/internal/utils/random"
	"github.com/gin-gonic/gin"
)

var pluginService plugin.Service

func init() {
	pluginService = plugin.NewService()
}

// List gets the plugin list
// @Summary Get plugin list
// @Description Get the plugin list with pagination
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginListRequest true "Request parameters"
// @Success 200 {object} response.Response{data=plugin.PluginListResponse}
// @Router /api/plugin [post]
func List(c *gin.Context) {
	var req models.PluginListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pluginService.List(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// ListByModule gets the plugin list by module
// @Summary Get plugin list by module
// @Description Get the plugin list for the specified module
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body map[string]string true "Request parameters"
// @Success 200 {object} response.Response{data=[]plugin.Plugin}
// @Router /api/plugin/module [post]
func ListByModule(c *gin.Context) {
	var req struct {
		Module string `json:"module" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pluginService.ListByModule(c, req.Module)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, map[string]interface{}{
		"list": result,
	}, "api.success")
}

// Detail gets plugin details
// @Summary Get plugin details
// @Description Get detailed information about the specified plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginDetailRequest true "Request parameters"
// @Success 200 {object} response.Response{data=plugin.Plugin}
// @Router /api/plugin/detail [post]
func Detail(c *gin.Context) {
	var req models.PluginDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pluginService.Detail(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// Save saves a plugin
// @Summary Save plugin
// @Description Create or update a plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginSaveRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/save [post]
func Save(c *gin.Context) {
	var req models.PluginSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Save(c, &req); err != nil {
		response.InternalServerError(c, "api.plugin.key.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// Delete deletes a plugin
// @Summary Delete plugin
// @Description Delete the specified plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginDeleteRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/delete [post]
func Delete(c *gin.Context) {
	var req models.PluginDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Delete(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// GetLogs gets plugin logs
// @Summary Get plugin logs
// @Description Get the logs of the specified plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginLogRequest true "Request parameters"
// @Success 200 {object} response.Response{data=string}
// @Router /api/plugin/log [post]
func GetLogs(c *gin.Context) {
	var req models.PluginLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pluginService.GetLogs(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, gin.H{
		"data": result,
	}, "api.success")
}

// CleanLogs cleans plugin logs
// @Summary Clean plugin logs
// @Description Clean the logs of the specified plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginLogRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/log/clean [post]
func CleanLogs(c *gin.Context) {
	var req models.PluginLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.CleanLogs(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// CleanAllLogs clears all plugin logs
// @Summary Clear all plugin logs
// @Description Clear the logs of all plugins
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/plugin/log/clean/all [post]
func CleanAllLogs(c *gin.Context) {
	if err := pluginService.CleanAllLogs(c); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// Import imports a plugin
// @Summary Import plugin
// @Description Import a new plugin
// @Tags Plugin Management
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Plugin file"
// @Param key query string true "Plugin key"
// @Success 200 {object} response.Response
// @Router /api/plugin/import [post]
func Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	key := c.Query("key")
	if key == "" {
		response.BadRequest(c, "api.bad_request", nil)
		return
	}
	// Save the uploaded file
	filePath := "uploads/" + random.GenerateString(5) + ".zip"
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	if err := pluginService.Import(c, filePath, key); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// Reinstall reinstalls a plugin
// @Summary Reinstall plugin
// @Description Reinstall the specified plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginReinstallRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/reinstall [post]
func Reinstall(c *gin.Context) {
	var req models.PluginReinstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Reinstall(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// Recheck rechecks a plugin
// @Summary Recheck plugin
// @Description Recheck the specified plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginRecheckRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/recheck [post]
func Recheck(c *gin.Context) {
	var req models.PluginRecheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Recheck(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// Uninstall uninstalls a plugin
// @Summary Uninstall plugin
// @Description Uninstall the specified plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginUninstallRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/uninstall [post]
func Uninstall(c *gin.Context) {
	var req models.PluginUninstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Uninstall(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// CheckKey checks a plugin key
// @Summary Check plugin key
// @Description Check whether the plugin key is valid
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body plugin.PluginKeyCheckRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/key/check [post]
func CheckKey(c *gin.Context) {
	var req models.PluginKeyCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.CheckKey(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// SearchRemotePlugins searches remote plugins
// @Summary Search remote plugins
// @Description Get the remote plugin list, marking installed plugins and plugins needing updates
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/plugin/remote/search [post]
func SearchRemotePlugins(c *gin.Context) {
	result, err := pluginService.SearchRemotePlugins(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// ImportByData imports a plugin via POST JSON data
// @Summary Import plugin via POST JSON data
// @Description Import a plugin by sending JSON data via POST, including the serialized string of info.json, the source field, and the isSystem field
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body models.PluginImportByDataRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/import/data [post]
func ImportByData(c *gin.Context) {
	var req models.PluginImportByDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.ImportByData(c, &req); err != nil {
		response.InternalServerError(c, "api.plugin.key.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// Status updates the plugin status
// @Summary Update plugin status
// @Description Update the enabled/disabled status of the specified plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body models.PluginStatusRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/status [post]
func Status(c *gin.Context) {
	var req models.PluginStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.UpdateStatus(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// Run runs a plugin once
// @Summary Run plugin once
// @Description Manually trigger execution of the specified server-side plugin
// @Tags Plugin Management
// @Accept json
// @Produce json
// @Param request body models.PluginRunRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/run [post]
func Run(c *gin.Context) {
	var req models.PluginRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Run(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}
