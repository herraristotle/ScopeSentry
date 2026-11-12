// Package response -----------------------------
// @file      : response.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/4 22:04
// -------------------------------------------
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response struct
type Response struct {
	Code    int         `json:"code"`    // Status code
	Message string      `json:"message"` // Message
	Data    interface{} `json:"data"`    // Data
}

// Ok success response
func Ok(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
	})
}

// OkWithMessage success response with message
func OkWithMessage(message string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
	})
}

// OkWithData success response with data
func OkWithData(data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// OkWithMessageAndData success response with message and data
func OkWithMessageAndData(message string, data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// Fail failure response
func Fail(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    500,
		Message: "fail",
	})
}

// FailWithMessage failure response with message
func FailWithMessage(message string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    500,
		Message: message,
	})
}

// FailWithCodeAndMessage failure response with status code and message
func FailWithCodeAndMessage(code int, message string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// FailWithCodeAndMessageAndData failure response with status code, message and data
func FailWithCodeAndMessageAndData(code int, message string, data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
