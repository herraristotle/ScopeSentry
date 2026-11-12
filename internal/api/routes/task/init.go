// task-------------------------------------
// @file      : task.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/10 15:33
// -------------------------------------------

package task

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/common"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/gin-gonic/gin"
)

func RegisterTaskRoutes(api *gin.RouterGroup) {
	// Define the route group
	taskRoutes := models.RouteGroup{
		Path: "/task",
		Groups: []models.RouteGroup{
			registerTaskRoutes(),
			registerTemplateRoutes(),
			registerSchedulerTaskRoutes(),
		},
	}

	// Register the route group
	common.RegisterRouteGroup(api, taskRoutes)
}
