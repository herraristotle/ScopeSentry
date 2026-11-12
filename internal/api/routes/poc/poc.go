package poc

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/handlers/poc"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/common"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/gin-gonic/gin"
)

// RegisterPocRoutes registers POC-related routes
func RegisterPocRoutes(api *gin.RouterGroup) {
	// Define the POC route group
	pocRoutes := models.RouteGroup{
		Path: "/poc",
		Routes: []models.Route{
			// Routes requiring authentication and user:read permission
			{
				Method:      "POST",
				Path:        "",
				Handler:     poc.GetPocList,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/detail",
				Handler:     poc.GetPocDetail,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/update",
				Handler:     poc.UpdatePoc,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/add",
				Handler:     poc.AddPoc,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/delete",
				Handler:     poc.DeletePoc,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "GET",
				Path:        "/data/all",
				Handler:     poc.GetAllPocData,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/data/import",
				Handler:     poc.ImportPoc,
				Middlewares: common.WithAuth(),
			},
		},
	}

	// Register the POC route group
	common.RegisterRouteGroup(api, pocRoutes)
}
