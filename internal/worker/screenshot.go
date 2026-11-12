// worker-------------------------------------
// @file      : screenshot.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/11/2 17:58
// -------------------------------------------

package worker

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/utils"
	"path"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/repositories/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ScreenshotHandle(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Task panic recovered:", r)
			// After a panic, restart the task loop if not cancelled
			select {
			case <-ctx.Done():
				fmt.Println("Task cancelled, ScreenshotHandle will not restart")
			default:
				go ScreenshotHandle(ctx)
			}
		}
	}()

	// Initialize the common repository
	repo := common.NewRepository()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Task cancelled, ScreenshotHandle stopped")
			return
		default:

			// Batch query the collection, 100 records per query
			limit := int64(100)
			opts := options.Find().SetLimit(limit)

			// Query the data
			results, err := repo.FindMany(ctx, "screenshot", bson.M{}, opts)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to query screenshot data: %v\n", err))
				time.Sleep(10 * time.Second) // Wait 10 seconds after an error before retrying
				continue
			}

			// If there is no data, processing is complete
			if len(results) == 0 {
				time.Sleep(60 * time.Second) // Wait a while before checking again
				continue
			}

			// Collect the _ids to delete
			var ids []primitive.ObjectID

			// Print and collect the data
			for _, item := range results {
				// Extract the fav3 and content fields
				hash, _ := item["hash"].(string)
				content, _ := item["content"].(string)
				// Collect _ids for batch deletion
				if id, ok := item["_id"].(primitive.ObjectID); ok {
					ids = append(ids, id)
				}
				filepath := path.Join(config.GlobalConfig.System.ScreenshotDir, fmt.Sprintf("%v.png", hash))
				decoded, err := base64.StdEncoding.DecodeString(content)
				if err != nil {
					logger.Error(fmt.Sprintf("<UNK> screenshot base64 decode <UNK>: %v %v\n", err, content))
					continue
				}
				err = utils.WriteFile(filepath, decoded)
				if err != nil {
					logger.Error(fmt.Sprintf("<UNK> screenshot WriteFile<UNK>: %v\n", err))
					continue
				}

			}

			// Batch delete the data
			if len(ids) > 0 {
				filter := bson.M{"_id": bson.M{"$in": ids}}
				_, err := repo.DeleteMany(ctx, "screenshot", filter)
				if err != nil {
					fmt.Printf("Failed to delete screenshot data: %v\n", err)
				}
			}

			// Brief sleep after each batch to avoid overly frequent database operations
			time.Sleep(3 * time.Second)
		}
	}
}
