package user

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/handlers/user"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/common"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers user-related routes
func RegisterUserRoutes(api *gin.RouterGroup) {
	// Define the user route group
	userRoutes := models.RouteGroup{
		Path: "/user",
		Routes: []models.Route{
			// Public routes
			{
				Method:  "POST",
				Path:    "/login",
				Handler: user.Login,
			},
			// Change password (requires authentication)
			{
				Method:      "POST",
				Path:        "/changePassword",
				Handler:     user.ChangePassword,
				Middlewares: common.WithAuth(),
			},
			//{
			//	Method:  "POST",
			//	Path:    "/register",
			//	Handler: user.Register,
			//},
			//// Routes requiring authentication and user:read permission
			//{
			//	Method:      "GET",
			//	Path:        "",
			//	Handler:     user.List,
			//	Middlewares: common.WithAuthAndPermission(constants.PermissionUserRead),
			//},
			//{
			//	Method:      "GET",
			//	Path:        "/:id",
			//	Handler:     user.Get,
			//	Middlewares: common.WithAuthAndPermission(constants.PermissionUserRead),
			//},
			//// Routes requiring authentication and user:update permission
			//{
			//	Method:      "PUT",
			//	Path:        "/:id",
			//	Handler:     user.Update,
			//	Middlewares: common.WithAuthAndPermission(constants.PermissionUserUpdate),
			//},
			//// Routes requiring authentication and user:delete permission
			//{
			//	Method:      "DELETE",
			//	Path:        "/:id",
			//	Handler:     user.Delete,
			//	Middlewares: common.WithAuthAndPermission(constants.PermissionUserDelete),
			//},
		},
	}

	// Register the user route group
	common.RegisterRouteGroup(api, userRoutes)
}
