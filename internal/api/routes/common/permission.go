// common-------------------------------------
// @file      : permission.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/8 21:11
// -------------------------------------------

package common

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// WithAuth adds the authentication middleware
func WithAuth() []gin.HandlerFunc {
	return []gin.HandlerFunc{middleware.AuthMiddleware()}
}

// WithPermission adds the permission middleware
func WithPermission(permission string) []gin.HandlerFunc {
	return []gin.HandlerFunc{middleware.PermissionMiddleware(permission)}
}

// WithAuthAndPermission adds the authentication and permission middleware
func WithAuthAndPermission(permission string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthMiddleware(),
		middleware.PermissionMiddleware(permission),
	}
}
