// @title           ScopeSentry API
// @version         1.0
// @description     ScopeSentry is a security scanning platform
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8082
// @BasePath  /api

// @securityDefinitions.apikey  ApiKeyAuth
// @in                         header
// @name                       Authorization
// @description               Bearer token for authentication

package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/Autumn-27/ScopeSentry/internal/worker"

	"github.com/Autumn-27/ScopeSentry/internal/bootstrap"
	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/scheduler"
	"github.com/Autumn-27/ScopeSentry/internal/update"

	"github.com/Autumn-27/ScopeSentry/internal/logger"

	_ "github.com/Autumn-27/ScopeSentry/internal/database/mongodb"

	_ "github.com/Autumn-27/ScopeSentry/internal/database/redis"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/constants"

	_ "github.com/Autumn-27/ScopeSentry/internal/bootstrap"

	"github.com/Autumn-27/ScopeSentry/internal/api/routes"
	scopesentrymcp "github.com/Autumn-27/ScopeSentry/internal/mcp"
	redisLogSubscriber "github.com/Autumn-27/ScopeSentry/internal/services/redis_log_subscriber"

	"github.com/Autumn-27/ScopeSentry/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

//go:embed static/*
var embeddedFiles embed.FS

func main() {
	// Set Gin mode
	if config.GlobalConfig.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	err := os.Setenv("TZ", config.GlobalConfig.System.Timezone)
	if err != nil {
		return
	}
	if constants.Version == "" {
		constants.Version = "dev"
	}
	fmt.Printf("version: %v\n", constants.Version)
	PrintPlugin()
	err = update.Update()
	if err != nil {
		logger.Error("Update failed", zap.Error(err))
	}
	bootstrap.GetProjectList()
	// Initialize scheduled tasks
	scheduler.InitializeGlobalScheduler()
	scheduler.GetGlobalScheduler().Start()
	defer scheduler.GetGlobalScheduler().Stop()

	// Start Redis log subscription service (running in background)
	go redisLogSubscriber.SubscribeLogChannel()
	// Create router
	router := routes.SetupRouter()
	scopesentrymcp.RegisterRoutes(router)
	gin.DisableConsoleColor()
	//router.Use(logger.GinLogger(), logger.GinRecovery())

	// Prepare frontend filesystem
	frontendFS, _ := fs.Sub(embeddedFiles, "static")
	// Prepare static assets filesystem (assets directory)
	assetsFS, _ := fs.Sub(embeddedFiles, "static/assets")
	// Prepare other static assets filesystems
	cssFS, _ := fs.Sub(embeddedFiles, "static/css")
	dracoFS, _ := fs.Sub(embeddedFiles, "static/draco")
	jsFS, _ := fs.Sub(embeddedFiles, "static/js")
	libFS, _ := fs.Sub(embeddedFiles, "static/lib")
	pluginsFS, _ := fs.Sub(embeddedFiles, "static/plugins")
	pngFS, _ := fs.Sub(embeddedFiles, "static/png")

	// Register Swagger routes (before static files to avoid conflicts)
	docs.SwaggerInfo.BasePath = "/api"
	router.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register static assets route (/assets maps to static/assets directory)
	router.StaticFS("/assets", http.FS(assetsFS)) // /assets corresponds to frontend js/css

	// Register other static assets routes
	router.StaticFS("/css", http.FS(cssFS))         // /css corresponds to static/css directory
	router.StaticFS("/draco", http.FS(dracoFS))     // /draco corresponds to static/draco directory
	router.StaticFS("/js", http.FS(jsFS))           // /js corresponds to static/js directory
	router.StaticFS("/lib", http.FS(libFS))         // /lib corresponds to static/lib directory
	router.StaticFS("/plugins", http.FS(pluginsFS)) // /plugins corresponds to static/plugins directory
	router.StaticFS("/png", http.FS(pngFS))         // /png corresponds to static/png directory

	// Register /map route mapped to map.html
	router.GET("/map", func(c *gin.Context) {
		data, err := fs.ReadFile(frontendFS, "map.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read map.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Register /favicon.ico route
	router.GET("/favicon.ico", func(c *gin.Context) {
		data, err := fs.ReadFile(frontendFS, "favicon.ico")
		if err != nil {
			c.String(http.StatusNotFound, "Favicon not found")
			return
		}
		c.Data(http.StatusOK, "image/x-icon", data)
	})

	// Register uploaded files static route (/images maps to external uploads directory)
	router.Static("/images", config.GlobalConfig.System.ImgDir) // /uploads corresponds to uploaded image files

	// Register root path route, returning file content directly to avoid redirects
	router.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Handle all unmatched routes, returning the frontend homepage (for SPA routing)
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Exclude API paths and static asset paths
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/mcp") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		// If a static asset is requested but not found, return 404 instead of the homepage
		if strings.HasPrefix(path, "/assets") || strings.HasPrefix(path, "/uploads") ||
			strings.HasPrefix(path, "/css") || strings.HasPrefix(path, "/draco") ||
			strings.HasPrefix(path, "/js") || strings.HasPrefix(path, "/lib") ||
			strings.HasPrefix(path, "/plugins") || strings.HasPrefix(path, "/png") ||
			strings.HasPrefix(path, "/images") {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		// Other paths return the frontend homepage (SPA routing)
		data, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler: router,
	}

	// Run asset processing
	worker.Run()

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.Int("port", config.GlobalConfig.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Set shutdown timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shut down server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}

func Banner() {
	banner := "   _____                         _____            _              \n  / ____|                       / ____|          | |             \n | (___   ___ ___  _ __   ___  | (___   ___ _ __ | |_ _ __ _   _ \n  \\___ \\ / __/ _ \\| '_ \\ / _ \\  \\___ \\ / _ \\ '_ \\| __| '__| | | |\n  ____) | (_| (_) | |_) |  __/  ____) |  __/ | | | |_| |  | |_| |\n |_____/ \\___\\___/| .__/ \\___| |_____/ \\___|_| |_|\\__|_|   \\__, |\n                  | |                                       __/ |\n                  |_|                                      |___/ "
	fmt.Println(banner)
}

func PrintPlugin() {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("✨✨✨ IMPORTANT NOTICE: Please review the Plugin Key below ✨✨✨")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("🔑 Plugin Key: %s\n", config.GlobalConfig.System.PluginKey)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("✅ Ensure the Plugin Key is correctly copied!")
}
