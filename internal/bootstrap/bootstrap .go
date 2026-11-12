package bootstrap

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

// StartupEvent startup event handler function
type StartupEvent func(ctx context.Context) error

var startupEvents []StartupEvent

// RegisterStartupEvent registers a startup event
func RegisterStartupEvent(event StartupEvent) {
	startupEvents = append(startupEvents, event)
}

// ExecuteStartupEvents executes all startup events
func ExecuteStartupEvents() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, event := range startupEvents {
		if err := event(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Register default startup events
func init() {
	// Load configuration
	//RegisterStartupEvent(func(ctx context.Context) error {
	//	return config.Load()
	//})

	//// Initialize logging
	//RegisterStartupEvent(func(ctx context.Context) error {
	//	return logger.Init()
	//})

	// Initialize MongoDB
	//RegisterStartupEvent(func(ctx context.Context) error {
	//	return mongodb.Init()
	//})

	// Initialize Redis
	//RegisterStartupEvent(func(ctx context.Context) error {
	//	return redis.Init()
	//})

	RegisterStartupEvent(func(ctx context.Context) error {
		return Banner()
	})

	//// Initialize database
	RegisterStartupEvent(func(ctx context.Context) error {
		return mongodb.CreateDatabase()
	})

	// Initialize notification API
	RegisterStartupEvent(func(ctx context.Context) error {
		if err := mongodb.FindAll("notification", bson.M{"state": true}, bson.M{"_id": 0, "method": 1, "url": 1, "contentType": 1, "data": 1, "state": 1}, &config.NotificationApi); err != nil {
			return err
		}
		return nil
	})

	// More startup events can be added here
	// For example: initialize cache, load plugins, start scheduled tasks, etc.
	err := ExecuteStartupEvents()
	if err != nil {
		log.Fatalf("Error during startup events: %v", err)
		return
	}
}

func Banner() error {
	banner := "   _____                         _____            _              \n  / ____|                       / ____|          | |             \n | (___   ___ ___  _ __   ___  | (___   ___ _ __ | |_ _ __ _   _ \n  \\___ \\ / __/ _ \\| '_ \\ / _ \\  \\___ \\ / _ \\ '_ \\| __| '__| | | |\n  ____) | (_| (_) | |_) |  __/  ____) |  __/ | | | |_| |  | |_| |\n |_____/ \\___\\___/| .__/ \\___| |_____/ \\___|_| |_|\\__|_|   \\__, |\n                  | |                                       __/ |\n                  |_|                                      |___/ "
	fmt.Println(banner)
	return nil
}
