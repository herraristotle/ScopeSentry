package response

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"net/http"
	"runtime"

	"github.com/Autumn-27/ScopeSentry/internal/i18n"
	"github.com/gin-gonic/gin"
)

// Response unified response structure
type Response struct {
	Code    int         `json:"code"`           // HTTP status code
	Message string      `json:"message"`        // Response message
	Data    interface{} `json:"data,omitempty"` // Response data
}

// SuccessResponse success response example
type SuccessResponse struct {
	Code    int         `json:"code" example:"200"`        // Success status code
	Message string      `json:"message" example:"Operation successful"` // Success message
	Data    interface{} `json:"data,omitempty"`            // Response data
}

// BadRequestResponse bad request response example
type BadRequestResponse struct {
	Code    int    `json:"code" example:"400"`         // Bad request status code
	Message string `json:"message" example:"Bad request"` // Error message
}

// UnauthorizedResponse unauthorized response example
type UnauthorizedResponse struct {
	Code    int    `json:"code" example:"401"`           // Unauthorized status code
	Message string `json:"message" example:"Unauthorized access"` // Error message
}

// NotFoundResponse not found response example
type NotFoundResponse struct {
	Code    int    `json:"code" example:"404"`         // Not found status code
	Message string `json:"message" example:"Resource not found"` // Error message
}

// InternalServerErrorResponse internal server error response example
type InternalServerErrorResponse struct {
	Code    int    `json:"code" example:"500"`             // Internal server error status code
	Message string `json:"message" example:"Internal server error"` // Error message
}

type PageResponse struct {
	Total    int64       `json:"total"`
	List     interface{} `json:"list"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// getLocale gets the language of the request
func getLocale(c *gin.Context) string {
	locale := c.GetHeader("Accept-Language")
	if locale == "" {
		locale = "zh-CN" // Default to Chinese
	}
	return locale
}

func Success(c *gin.Context, data interface{}, msgKey string) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}, msgKey string) {
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    data,
	})
}

func BadRequest(c *gin.Context, msgKey string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	c.JSON(http.StatusBadRequest, Response{
		Code:    http.StatusBadRequest,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    errMsg,
	})
}

func NotFound(c *gin.Context, msgKey string, err error) {
	c.JSON(http.StatusNotFound, Response{
		Code:    http.StatusNotFound,
		Message: i18n.Translate(getLocale(c), msgKey),
	})
}

func InternalServerError(c *gin.Context, msgKey string, err error) {
	pc, file, line, ok := runtime.Caller(1)
	callerInfo := ""
	if ok {
		fn := runtime.FuncForPC(pc)
		callerInfo = fmt.Sprintf("%s:%d (%s)", file, line, fn.Name())
	}

	logger.Error(fmt.Sprintf("InternalServerError at %s: %v", callerInfo, err))
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    fmt.Sprintf("%v", err),
	})
}

func Unauthorized(c *gin.Context, msgKey string, err error) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    http.StatusUnauthorized,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    fmt.Sprintf("%v", err),
	})
}
