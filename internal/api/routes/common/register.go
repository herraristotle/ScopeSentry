// common-------------------------------------
// @file      : register.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/8 21:10
// -------------------------------------------

package common

import (
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/gin-gonic/gin"
)

// RegisterRouteGroup registers a route group
func RegisterRouteGroup(group *gin.RouterGroup, routeGroup models.RouteGroup) {
	// Create a child route group
	subGroup := group.Group(routeGroup.Path)

	// Apply group-level middleware
	for _, handlerFunc := range routeGroup.Middlewares {
		subGroup.Use(handlerFunc)
	}

	// Register routes
	for _, route := range routeGroup.Routes {
		// Create the route handler chain
		handlers := make([]gin.HandlerFunc, 0)

		// Add route-level middleware
		handlers = append(handlers, route.Middlewares...)

		// Add the route handler
		handlers = append(handlers, route.Handler)

		// Register the route according to the HTTP method
		switch route.Method {
		case "GET":
			subGroup.GET(route.Path, handlers...)
		case "POST":
			subGroup.POST(route.Path, handlers...)
		case "PUT":
			subGroup.PUT(route.Path, handlers...)
		case "DELETE":
			subGroup.DELETE(route.Path, handlers...)
		case "PATCH":
			subGroup.PATCH(route.Path, handlers...)
		}
	}

	for _, childGroup := range routeGroup.Groups {
		RegisterRouteGroup(subGroup, childGroup)
	}
}
