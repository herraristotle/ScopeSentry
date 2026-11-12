// worker-------------------------------------
// @file      : iconhandle.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/11/1 18:25
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

func IconHandle(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Task panic recovered:", r)
			// After a panic, restart the task loop if not cancelled
			select {
			case <-ctx.Done():
				fmt.Println("Task cancelled, IconHandle will not restart")
			default:
				go IconHandle(ctx)
			}
		}
	}()

	// Initialize the common repository
	repo := common.NewRepository()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Task cancelled, IconHandle stopped")
			return
		default:
			// Batch query the icon collection, 100 records per query
			limit := int64(100)
			opts := options.Find().SetLimit(limit)

			// Query the data
			results, err := repo.FindMany(ctx, "icon", bson.M{}, opts)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to query icon data: %v\n", err))
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
				fav3, _ := item["fav3"].(string)
				content, _ := item["content"].(string)
				filepath := path.Join(config.GlobalConfig.System.IconDir, fmt.Sprintf("%v.png", fav3))
				decoded, err := base64.StdEncoding.DecodeString(content)
				if err != nil {
					logger.Error(fmt.Sprintf("<UNK> icon base64 decode <UNK>: %v %v\n", err, content))
					continue
				}
				err = utils.WriteFile(filepath, decoded)
				if err != nil {
					logger.Error(fmt.Sprintf("<UNK> icon WriteFile<UNK>: %v\n", err))
					continue
				}

				// Collect _ids for batch deletion
				if id, ok := item["_id"].(primitive.ObjectID); ok {
					ids = append(ids, id)
				}
			}

			// Batch delete the data
			if len(ids) > 0 {
				filter := bson.M{"_id": bson.M{"$in": ids}}
				_, err := repo.DeleteMany(ctx, "icon", filter)
				if err != nil {
					fmt.Printf("Failed to delete icon data: %v\n", err)
				}
			}

			// Brief sleep after each batch to avoid overly frequent database operations
			time.Sleep(3 * time.Second)
		}
	}
}
