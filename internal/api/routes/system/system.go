// Package routes -----------------------------
// @file      : system.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/3 22:17
// -------------------------------------------
package system

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/handlers/system"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/common"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/gin-gonic/gin"
)

func RegisterSystemRoutes(api *gin.RouterGroup) {
	// Define the route group
	systemRoutes := models.RouteGroup{
		Path: "/system",
		Routes: []models.Route{
			// Routes requiring authentication and user:read permission
			{
				Method:      "GET",
				Path:        "version",
				Handler:     system.GetSystemVersion,
				Middlewares: common.WithAuth(),
			},
		},
	}

	// Register the route group
	common.RegisterRouteGroup(api, systemRoutes)
}
