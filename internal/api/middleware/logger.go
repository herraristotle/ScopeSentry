package middleware

import (
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		start := time.Now()
		// Request path
		path := c.Request.URL.Path
		// Request parameters
		query := c.Request.URL.RawQuery

		// Process the request
		c.Next()

		// End time
		cost := time.Since(start)
		// Status code
		status := c.Writer.Status()

		// Record the log
		logger.Info("Request",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("cost", cost),
		)
	}
}
