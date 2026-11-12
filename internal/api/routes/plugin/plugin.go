package plugin

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/handlers/plugin"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/common"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/gin-gonic/gin"
)

// RegisterPluginRoutes registers plugin-related routes
func RegisterPluginRoutes(api *gin.RouterGroup) {
	// Define the route group
	pluginRoutes := models.RouteGroup{
		Path: "/plugin",
		Routes: []models.Route{
			// Routes requiring authentication and user:read permission
			{
				Method:      "POST",
				Path:        "",
				Handler:     plugin.List,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/detail",
				Handler:     plugin.Detail,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/save",
				Handler:     plugin.Save,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/delete",
				Handler:     plugin.Delete,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/reinstall",
				Handler:     plugin.Reinstall,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/recheck",
				Handler:     plugin.Recheck,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/uninstall",
				Handler:     plugin.Uninstall,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/import",
				Handler:     plugin.Import,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/log",
				Handler:     plugin.GetLogs,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/log/clean",
				Handler:     plugin.CleanLogs,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/log/clean/all",
				Handler:     plugin.CleanAllLogs,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/module",
				Handler:     plugin.ListByModule,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/key/check",
				Handler:     plugin.CheckKey,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/remote/search",
				Handler:     plugin.SearchRemotePlugins,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/import/data",
				Handler:     plugin.ImportByData,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/status",
				Handler:     plugin.Status,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/run",
				Handler:     plugin.Run,
				Middlewares: common.WithAuth(),
			},
		},
	}

	// Register the route group
	common.RegisterRouteGroup(api, pluginRoutes)
}
