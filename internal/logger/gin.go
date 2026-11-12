// Package logger -----------------------------
// @file      : gin.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/4/24 13:55
// -------------------------------------------
package logger

//
//import (
//	"os"
//	"time"
//
//	"github.com/gin-gonic/gin"
//	"go.uber.org/zap"
//)
//
//// GinLogger middleware, replacing the default logger with zap
//func GinLogger() gin.HandlerFunc {
//	return gin.LoggerWithWriter(zapWriter{Logger})
//}
//
//// zapWriter struct, used to write gin logs into zap
//type zapWriter struct {
//	l *zap.Logger
//}
//
//func (z zapWriter) Write(p []byte) (n int, err error) {
//	z.l.Info(string(p)) // You can also add structured log handling here
//	return len(p), nil
//}
//
//// GinRecovery replaces the default panic recovery logic and logs errors
//func GinRecovery() gin.HandlerFunc {
//	return gin.CustomRecoveryWithWriter(os.Stderr, func(c *gin.Context, err interface{}) {
//		Logger.Error("Panic recovered",
//			zap.Any("error", err),
//			zap.String("path", c.Request.URL.Path),
//			zap.Time("time", time.Now()),
//		)
//		c.AbortWithStatusJSON(500, gin.H{
//			"message": "Server Error",
//		})
//	})
//}
