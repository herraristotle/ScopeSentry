// models-------------------------------------
// @file      : route.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/8 21:09
// -------------------------------------------

package models

import "github.com/gin-gonic/gin"

// RouteGroup route group configuration
type RouteGroup struct {
	Path        string
	Middlewares []gin.HandlerFunc
	Routes      []Route
	Groups      []RouteGroup
}

// Route route configuration
type Route struct {
	Method      string
	Path        string
	Handler     gin.HandlerFunc
	Middlewares []gin.HandlerFunc
}
