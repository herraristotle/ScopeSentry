package user

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/services/user"
	"github.com/gin-gonic/gin"
)

var userService user.Service

// ChangePasswordRequest change password request
type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required" example:"newStrongPassword123"`
}

// LoginRequest login request
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// RegisterRequest register request
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Role     string `json:"role" binding:"required" example:"admin"`
}

// UpdateRequest update request
type UpdateRequest struct {
	Email string `json:"email" binding:"omitempty,email" example:"admin@example.com"`
	Role  string `json:"role" binding:"omitempty" example:"admin"`
}

// Login user login
// @Summary      User login
// @Description  User login endpoint
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      LoginRequest  true  "Login credentials"
// @Success      200  {object}  response.SuccessResponse{data=object{access_token=string}}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /users/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	token, err := userService.Login(c, req.Username, req.Password)
	if err != nil {
		response.BadRequest(c, "api.user.login.failed", err)
		return
	}

	response.Success(c, gin.H{"access_token": token}, "api.user.login.success")
}

// ChangePassword changes the current user's password
// @Summary      Change current user's password
// @Description  Get user info from the JWT and change their password
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        changePasswordRequest  body      ChangePasswordRequest  true  "Password change information"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      401  {object}  response.UnauthorizedResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /user/changePassword [post]
func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// Get the userID from the context written by the JWT middleware
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "Unauthorized", nil)
		return
	}
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		response.Unauthorized(c, "Invalid token payload", nil)
		return
	}

	if err := userService.ChangePassword(c, userID, req.NewPassword); err != nil {
		response.InternalServerError(c, "api.user.change_password.failed", err)
		return
	}
	response.Success(c, nil, "api.user.change_password.success")
}

//// Register @Summary      User registration
//// @Description  User registration endpoint
//// @Tags         user
//// @Accept       json
//// @Produce      json
//// @Param        registerRequest  body      RegisterRequest  true  "Registration information"
//// @Success      200  {object}  response.SuccessResponse{data=models.User}
//// @Failure      400  {object}  response.BadRequestResponse
//// @Failure      401  {object}  response.UnauthorizedResponse
//// @Failure      404  {object}  response.NotFoundResponse
//// @Failure      500  {object}  response.InternalServerErrorResponse
//// @Router       /users/register [post]
//func Register(c *gin.Context) {
//	var req RegisterRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		response.BadRequest(c, "api.bad_request", err)
//		return
//	}
//
//	user := &models.User{
//		Username: req.Username,
//		Password: req.Password,
//		Email:    req.Email,
//		Role:     req.Role,
//	}
//
//	if err := userService.Register(c, user); err != nil {
//		response.InternalServerError(c, "api.user.create.failed", err)
//		return
//	}
//
//	response.Success(c, user, "api.user.create.success")
//}
//
//// List @Summary      Get user list
//// @Description  Get user list endpoint
//// @Tags         user
//// @Accept       json
//// @Produce      json
//// @Security     ApiKeyAuth
//// @Param        page  query     int  false  "Page number"  default(1)
//// @Param        size  query     int  false  "Items per page"  default(10)
//// @Success      200   {object}  response.SuccessResponse{data=[]models.User}
//// @Failure      401   {object}  response.UnauthorizedResponse
//// @Failure      500   {object}  response.InternalServerErrorResponse
//// @Router       /users [get]
//func List(c *gin.Context) {
//	users, err := userService.GetAllUsers(c)
//	if err != nil {
//		response.InternalServerError(c, "api.error", err)
//		return
//	}
//
//	response.Success(c, users, "api.success")
//}
//
//// Get @Summary      Get a single user
//// @Description  Get a single user endpoint
//// @Tags         user
//// @Accept       json
//// @Produce      json
//// @Security     ApiKeyAuth
//// @Param        id   path      string  true  "User ID"
//// @Success      200  {object}  response.SuccessResponse{data=models.User}
//// @Failure      401  {object}  response.UnauthorizedResponse
//// @Failure      404  {object}  response.NotFoundResponse
//// @Failure      500  {object}  response.InternalServerErrorResponse
//// @Router       /users/{id} [get]
//func Get(c *gin.Context) {
//	id := c.Param("id")
//	user, err := userService.GetUserByID(c, id)
//	if err != nil {
//		response.InternalServerError(c, "api.error", err)
//		return
//	}
//	if user == nil {
//		response.NotFound(c, "api.user.not_found", nil)
//		return
//	}
//
//	response.Success(c, user, "api.success")
//}
//
//// Update @Summary      Update user
//// @Description  Update user endpoint
//// @Tags         user
//// @Accept       json
//// @Produce      json
//// @Security     ApiKeyAuth
//// @Param        id            path      string         true  "User ID"
//// @Param        updateRequest  body      UpdateRequest  true  "Update information"
//// @Success      200           {object}  response.SuccessResponse{data=models.User}
//// @Failure      400           {object}  response.BadRequestResponse
//// @Failure      401           {object}  response.UnauthorizedResponse
//// @Failure      404           {object}  response.NotFoundResponse
//// @Failure      500           {object}  response.InternalServerErrorResponse
//// @Router       /users/{id} [put]
//func Update(c *gin.Context) {
//	id := c.Param("id")
//	var req UpdateRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		response.BadRequest(c, "api.bad_request", err)
//		return
//	}
//
//	user, err := userService.GetUserByID(c, id)
//	if err != nil {
//		response.InternalServerError(c, "api.error", err)
//		return
//	}
//	if user == nil {
//		response.NotFound(c, "api.user.not_found", nil)
//		return
//	}
//
//	user.Email = req.Email
//	user.Role = req.Role
//
//	if err := userService.UpdateUser(c, id, user); err != nil {
//		response.InternalServerError(c, "api.user.update.failed", err)
//		return
//	}
//
//	response.Success(c, user, "api.user.update.success")
//}
//
//// Delete @Summary      Delete user
//// @Description  Delete user endpoint
//// @Tags         user
//// @Accept       json
//// @Produce      json
//// @Security     ApiKeyAuth
//// @Param        id  path      string  true  "User ID"
//// @Success      200  {object}  response.SuccessResponse
//// @Failure      401  {object}  response.UnauthorizedResponse
//// @Failure      404  {object}  response.NotFoundResponse
//// @Failure      500  {object}  response.InternalServerErrorResponse
//// @Router       /users/{id} [delete]
//func Delete(c *gin.Context) {
//	id := c.Param("id")
//	if err := userService.DeleteUser(c, id); err != nil {
//		response.InternalServerError(c, "api.user.delete.failed", err)
//		return
//	}
//
//	response.Success(c, nil, "api.user.delete.success")
//}

func init() {
	userService = user.NewService()
}
