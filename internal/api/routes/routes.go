package routes

import (
	"strings"

	"github.com/Autumn-27/ScopeSentry/internal/api/middleware"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/assets"
	confRoutes "github.com/Autumn-27/ScopeSentry/internal/api/routes/configuration"
	dictRoutes "github.com/Autumn-27/ScopeSentry/internal/api/routes/dictionary"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/export"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/fingerprint"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/plugin"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/poc"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/project"
	sroutes "github.com/Autumn-27/ScopeSentry/internal/api/routes/sensitive_rule"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/system"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/task"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes/user"
	"github.com/gin-gonic/gin"

	"github.com/Autumn-27/ScopeSentry/internal/api/routes/node"
	apikeyRoutes "github.com/Autumn-27/ScopeSentry/internal/api/routes/apikey"
)

// loggerWithSkipPaths is a custom logging middleware that skips paths with the given prefixes
func loggerWithSkipPaths(skipPrefixes []string) gin.HandlerFunc {
	// Create the default logging middleware
	defaultLogger := gin.Logger()

	return func(c *gin.Context) {
		// Check whether the path should be skipped
		path := c.Request.URL.Path
		for _, prefix := range skipPrefixes {
			if strings.HasPrefix(path, prefix) {
				// If the path should be skipped, run the next middleware without logging
				c.Next()
				return
			}
		}

		// Otherwise run the default logging middleware
		defaultLogger(c)
	}
}

// SetupRouter sets up the router
func SetupRouter() *gin.Engine {
	router := gin.New()

	// Configure the logging middleware to skip logs for /assets and /uploads paths
	router.Use(loggerWithSkipPaths([]string{"/assets", "/uploads", "/images"}))

	// Add the recovery middleware
	router.Use(gin.Recovery())

	// Add global middleware
	router.Use(middleware.I18nMiddleware())

	// Register API routes
	api := router.Group("/api")
	{
		// Register user routes
		user.RegisterUserRoutes(api)

		// Register system routes
		system.RegisterSystemRoutes(api)

		// Register configuration routes
		confRoutes.RegisterConfigurationRoutes(api)

		// Register asset routes
		assets.RegisterAssetsRoutes(api)

		// Register fingerprint routes
		fingerprint.RegisterFingerprintRoutes(api)

		// Register SensitiveRule routes
		sroutes.RegisterSensitiveRuleRoutes(api)

		// Register node routes
		node.RegisterNodeRoutes(api)

		task.RegisterTaskRoutes(api)

		project.RegisterProjectRoutes(api)

		poc.RegisterPocRoutes(api)

		export.RegisterExportRoutes(api)

		plugin.RegisterPluginRoutes(api)

		// Register dictionary routes
		dictRoutes.RegisterDictionaryRoutes(api)

		// Register API Key management routes
		apikeyRoutes.RegisterApiKeyRoutes(api)
	}

	return router
}

// Middleware composition helpers
