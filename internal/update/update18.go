// update-------------------------------------
// @file      : update18.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/11/13 20:36
// -------------------------------------------

package update

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"
	"github.com/Autumn-27/ScopeSentry/internal/utils/random"
	"github.com/schollz/progressbar/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Update18() {
	//collection := mongodb.DB.Collection("asset")
	//batchSize := int64(1000)
	//var lastID primitive.ObjectID
	//ctx := context.Background()
	//// Convert the metadata field of non-http assets to banner
	//for {
	//	filter := bson.M{"type": "other"}
	//	if !lastID.IsZero() {
	//		filter["_id"] = bson.M{"$gt": lastID}
	//	}
	//
	//	opts := options.Find().SetSort(bson.D{{"_id", 1}}).SetLimit(batchSize)
	//	cursor, err := collection.Find(ctx, filter, opts)
	//	if err != nil {
	//		logger.Error(fmt.Sprintf("Find error: %v", err))
	//	}
	//
	//	count := 0
	//	var bulkModels []mongo.WriteModel
	//
	//	for cursor.Next(ctx) {
	//		count++
	//		var doc bson.M
	//		if err := cursor.Decode(&doc); err != nil {
	//			logger.Info(fmt.Sprintf("Decode error: %v", err))
	//			continue
	//		}
	//
	//		id := doc["_id"].(primitive.ObjectID)
	//		lastID = id
	//
	//		// Extract the metadata field (Binary type)
	//		metadataRaw, ok := doc["metadata"].(primitive.Binary)
	//		if !ok {
	//			continue
	//		}
	//
	//		// Convert to ASCII
	//		asciiStr := strconv.QuoteToASCII(string(metadataRaw.Data))
	//
	//		// Construct the bulk update operations
	//		update := mongo.NewUpdateOneModel().
	//			SetFilter(bson.M{"_id": id}).
	//			SetUpdate(bson.M{
	//				"$set": bson.M{
	//					"metadata": asciiStr,
	//				},
	//			})
	//		bulkModels = append(bulkModels, update)
	//	}
	//
	//	cursor.Close(ctx)
	//
	//	if len(bulkModels) > 0 {
	//		// Execute the bulk update
	//		bulkWriteResult, err := collection.BulkWrite(ctx, bulkModels)
	//		if err != nil {
	//			logger.Error(fmt.Sprintf("BulkWrite error: %v", err))
	//		}
	//		logger.Info(fmt.Sprintf("Bulk update: Matched %d, Modified %d\n", bulkWriteResult.MatchedCount, bulkWriteResult.ModifiedCount))
	//	}
	//
	//	if count == 0 {
	//		break // No more data; exit
	//	}
	//}
	// Change the banner field of http assets
	//filter := bson.M{"type": "http"}
	//pipeline := []bson.M{
	//	{
	//		"$set": bson.M{
	//			"banner": "$rawheaders", // Here $rawheaders is resolved as a field reference
	//		},
	//	},
	//	{
	//		"$unset": "rawheaders",
	//	},
	//}
	//
	//result, err := collection.UpdateMany(ctx, filter, pipeline)
	//if err != nil {
	//	logger.Error(fmt.Sprintf("UpdateMany error:", err))
	//}

	//logger.Info(fmt.Sprintf("Matched %d documents and modified %d documents\n", result.MatchedCount, result.ModifiedCount))

	// Asset bodyhash index
	asset := mongodb.DB.Collection("asset")
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"bodyhash", 1}}},
	}
	_, err := asset.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create asset indexes: %v", err))
	}
	// Sensitive information index
	SensitiveBody := mongodb.DB.Collection("SensitiveBody")
	_, err = SensitiveBody.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{"md5", 1}},
	})
	if err != nil {
		fmt.Errorf("failed to create SensitiveBody index: %v", err)
	}
	// Create icon collection indexes
	icon := mongodb.DB.Collection("icon")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"fav3", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = icon.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create icon indexes: %v", err))
	}

	// Create http res body collection indexes
	httpBody := mongodb.DB.Collection("HttpBody")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = httpBody.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create httpBody indexes: %v", err))
	}

	// Create http screenshot indexes
	httpScreenshot := mongodb.DB.Collection("screenshot")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = httpScreenshot.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create httpScreenshot indexes: %v", err))
	}

	ChangeAsset()

	// Create vulDetail indexes
	vulnerabilityDetail := mongodb.DB.Collection("vulnerabilityDetail")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}},
	}
	_, err = vulnerabilityDetail.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create vulnerabilityDetail indexes: %v", err))
	}

	// Migrate vul's req/res to vulnerabilityDetail
	MigrateVulnerabilityDetails()

	// Create IP asset temporary collection indexes
	ipAssetTmp := mongodb.DB.Collection("IPAssetTmp")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"ip", 1}, {"port", 1}, {"domain", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = ipAssetTmp.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create ipAssetTmp indexes: %v", err))
	}

	// Create IP asset collection indexes
	ipAsset := mongodb.DB.Collection("IPAsset")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"time", -1}}},
		{Keys: bson.D{{"ports.server.domain", 1}}},
		{Keys: bson.D{{"ports.port", 1}}},
		{Keys: bson.D{{"ip", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"ports.server.service", 1}}},
		{Keys: bson.D{{"ports.server.webServer", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
		{Keys: bson.D{{"ports.server.technologies", 1}}},
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
	}
	_, err = ipAsset.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create ipAsset indexes: %v", err))
	}

	// Create task-related indexes
	task := mongodb.DB.Collection("task")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"creatTime", -1}}},
		{Keys: bson.D{{"name", 1}}},
		{Keys: bson.D{{"progress", 1}}},
	}
	_, err = task.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create task indexes: %v", err))
	}

	logger.Info(fmt.Sprintf("All updates completed."))

}

func ChangeAsset() {
	ctx := context.Background()
	collection := mongodb.DB.Collection("asset")
	httpBodyCollection := mongodb.DB.Collection("HttpBody")
	ipAssetTmpCollection := mongodb.DB.Collection("IPAssetTmp")

	filter := bson.M{}
	totalCount, _ := collection.CountDocuments(ctx, filter)

	cursor, err := collection.Find(ctx, filter, options.Find().SetBatchSize(1000))
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	bar := progressbar.NewOptions64(
		totalCount,
		progressbar.OptionSetDescription("Processing assets update..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(100*time.Millisecond),
	)

	var bulkModels []mongo.WriteModel
	batchSize := 1000

	ipAssetTmpBatchSize := 100
	ipAssetTmpDocs := make([]interface{}, 0, ipAssetTmpBatchSize)
	insertManyOpts := options.InsertMany().SetOrdered(false)

	handleBulkErr := func(bulkErr *mongo.BulkWriteException) {
		if bulkErr == nil {
			return
		}
		duplicateCount := 0
		var nonDuplicateErrors []mongo.BulkWriteError
		for _, writeErr := range bulkErr.WriteErrors {
			switch writeErr.Code {
			case 11000, 11001, 12582:
				duplicateCount++
			default:
				nonDuplicateErrors = append(nonDuplicateErrors, writeErr)
			}
		}
		if duplicateCount > 0 {
			logger.Warn(fmt.Sprintf("Skipped %d duplicate IPAssetTmp records during batch insert", duplicateCount))
		}
		if len(nonDuplicateErrors) > 0 {
			logger.Warn(fmt.Sprintf("Failed to insert IPAssetTmp batch due to non-duplicate errors: %+v", nonDuplicateErrors))
		}
		if bulkErr.WriteConcernError != nil {
			logger.Warn(fmt.Sprintf("Write concern error when inserting IPAssetTmp batch: %v", bulkErr.WriteConcernError))
		}
	}

	flushIPAssetTmp := func() {
		if len(ipAssetTmpDocs) == 0 {
			return
		}
		_, err := ipAssetTmpCollection.InsertMany(ctx, ipAssetTmpDocs, insertManyOpts)
		if err != nil {
			switch e := err.(type) {
			case mongo.BulkWriteException:
				handleBulkErr(&e)
			case *mongo.BulkWriteException:
				handleBulkErr(e)
			default:
				logger.Error(fmt.Sprintf("Failed to insert IPAssetTmp batch: %v", err))
			}
		}
		ipAssetTmpDocs = ipAssetTmpDocs[:0]
	}

	for cursor.Next(ctx) {
		var asset models.Asset
		if err := cursor.Decode(&asset); err != nil {
			log.Println("decode error:", err)
			bar.Add(1)
			continue
		}

		ipAssetTmpDoc := models.IPAssetTmp{
			IP:           asset.IP,
			Domain:       asset.Host,
			Port:         asset.Port,
			Service:      asset.Service,
			WebServer:    asset.WebServer,
			Technologies: asset.Technologies,
			Project:      asset.Project,
			TaskName:     asset.TaskName,
			RootDomain:   asset.RootDomain,
		}
		ipAssetTmpDocs = append(ipAssetTmpDocs, ipAssetTmpDoc)
		if len(ipAssetTmpDocs) >= ipAssetTmpBatchSize {
			flushIPAssetTmp()
		}

		var hash string

		// Process http-type Body, compute the hash and insert into the HttpBody collection
		if asset.Type == "http" {
			if asset.Body != "" {
				hash = helper.HashXX64String(asset.Body)
				asset.ResponseBodyHash = hash
				// Insert into the HttpBody collection (using upsert to avoid duplicates)
				doc := bson.M{
					"hash":    hash,
					"content": asset.Body,
				}
				_, err := httpBodyCollection.UpdateOne(
					ctx,
					bson.M{"hash": hash},
					bson.M{"$set": doc},
					options.Update().SetUpsert(true),
				)
				if err != nil {
					log.Printf("failed to insert HttpBody for asset %v: %v", asset.ID, err)
				}
			}
		} else {
			// For non-http types
			bar.Add(1)
			continue
		}

		// Process IconContent: decode and write to a file
		if asset.IconContent != "" && asset.FaviconMMH3 != "" {
			iconData, err := base64.StdEncoding.DecodeString(asset.IconContent)
			if err != nil {
				log.Printf("failed to decode IconContent for asset %v: %v", asset.ID, err)
			} else {
				iconPath := filepath.Join(config.GlobalConfig.System.IconDir, fmt.Sprintf("%v.png", asset.FaviconMMH3))
				if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
					log.Printf("failed to write icon file for asset %v: %v", asset.ID, err)
				}
			}
		}

		// Process Screenshot: strip the prefix, decode and write to a file
		if asset.Screenshot != "" {
			// Determine the hash used for the file name: prefer ResponseBodyHash; for http types with an empty ResponseBodyHash, use the computed hash
			screenshotHash := asset.ResponseBodyHash
			if screenshotHash == "" && asset.Type == "http" && hash != "" {
				screenshotHash = hash
			}
			if screenshotHash != "" {
				screenshotBase64 := asset.Screenshot
				// Remove the data:image/png;base64, prefix
				if strings.HasPrefix(screenshotBase64, "data:image/png;base64,") {
					screenshotBase64 = strings.TrimPrefix(screenshotBase64, "data:image/png;base64,")
				}
				screenshotData, err := base64.StdEncoding.DecodeString(screenshotBase64)
				if err != nil {
					log.Printf("failed to decode Screenshot for asset %v: %v", asset.ID, err)
				} else {
					screenshotPath := filepath.Join(config.GlobalConfig.System.ScreenshotDir, fmt.Sprintf("%v.png", screenshotHash))
					if err := os.WriteFile(screenshotPath, screenshotData, 0644); err != nil {
						log.Printf("failed to write screenshot file for asset %v: %v", asset.ID, err)
					}
				}
			}
		}

		// Prepare the bulk update operations
		updateDoc := bson.M{
			"$unset": bson.M{
				"body":        "",
				"iconcontent": "",
				"screenshot":  "",
			},
		}
		// Only update the bodyhash field when the hash is non-empty
		if hash != "" {
			updateDoc["$set"] = bson.M{
				"bodyhash": hash,
			}
		}
		update := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": asset.ID}).
			SetUpdate(updateDoc)
		bulkModels = append(bulkModels, update)

		// When 1000 records have been collected, execute the bulk update
		if len(bulkModels) >= batchSize {
			if len(bulkModels) > 0 {
				_, err := collection.BulkWrite(ctx, bulkModels)
				if err != nil {
					log.Printf("BulkWrite error: %v", err)
				} else {
					logger.Info(fmt.Sprintf("Bulk updated %d assets", len(bulkModels)))
				}
			}
			// Clear the bulk operation list
			bulkModels = []mongo.WriteModel{}
		}

		bar.Add(1)
	}

	// Process the remaining bulk updates
	if len(bulkModels) > 0 {
		_, err := collection.BulkWrite(ctx, bulkModels)
		if err != nil {
			log.Printf("BulkWrite error: %v", err)
		} else {
			logger.Info(fmt.Sprintf("Bulk updated %d assets (final batch)", len(bulkModels)))
		}
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	flushIPAssetTmp()

	logger.Info("Asset update completed.")
	logger.Info("Note: MongoDB disk space has not been reclaimed yet. Use CompactAssetCollection() to reclaim space after data migration.")
	_ = CompactAssetCollection()

}

// MigrateVulnerabilityDetails migrates request/response fields from the vulnerability collection to the vulnerabilityDetail collection
func MigrateVulnerabilityDetails() {
	ctx := context.Background()
	vulnerabilityCollection := mongodb.DB.Collection("vulnerability")
	vulnerabilityDetailCollection := mongodb.DB.Collection("vulnerabilityDetail")

	// Query all vulnerability records
	filter := bson.M{}
	totalCount, err := vulnerabilityCollection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to count vulnerabilities: %v", err))
		return
	}

	if totalCount == 0 {
		logger.Info("No vulnerabilities to migrate.")
		return
	}

	cursor, err := vulnerabilityCollection.Find(ctx, filter, options.Find().SetBatchSize(1000))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to find vulnerabilities: %v", err))
		return
	}
	defer cursor.Close(ctx)

	bar := progressbar.NewOptions64(
		totalCount,
		progressbar.OptionSetDescription("Migrating vulnerability details..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(100*time.Millisecond),
	)

	var bulkDetailModels []mongo.WriteModel
	var bulkUpdateModels []mongo.WriteModel
	batchSize := 1000

	for cursor.Next(ctx) {
		var vul models.Vulnerability
		if err := cursor.Decode(&vul); err != nil {
			log.Println("decode error:", err)
			bar.Add(1)
			continue
		}

		// Skip records without request and response
		if vul.Request == "" && vul.Response == "" {
			bar.Add(1)
			continue
		}

		// If already migrated (hash present), skip
		if vul.Hash != "" {
			bar.Add(1)
			continue
		}

		// Generate a 16-character random string as the hash
		hash := random.GenerateRandomString(16)

		// Prepare the document to insert into vulnerabilityDetail
		detailDoc := bson.M{
			"hash": hash,
			"req":  vul.Request,
			"res":  vul.Response,
		}
		insertDetail := mongo.NewInsertOneModel().SetDocument(detailDoc)
		bulkDetailModels = append(bulkDetailModels, insertDetail)

		// Prepare the document to update in vulnerability
		updateDoc := bson.M{
			"$set": bson.M{
				"hash": hash,
			},
			"$unset": bson.M{
				"request":  "",
				"response": "",
			},
		}
		update := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": vul.ID}).
			SetUpdate(updateDoc)
		bulkUpdateModels = append(bulkUpdateModels, update)

		// When enough records have been collected, execute the bulk operations
		if len(bulkDetailModels) >= batchSize {
			// Bulk insert into vulnerabilityDetail
			if len(bulkDetailModels) > 0 {
				_, err := vulnerabilityDetailCollection.BulkWrite(ctx, bulkDetailModels)
				if err != nil {
					log.Printf("BulkWrite vulnerabilityDetail error: %v", err)
				} else {
					logger.Info(fmt.Sprintf("Bulk inserted %d vulnerability details", len(bulkDetailModels)))
				}
				bulkDetailModels = []mongo.WriteModel{}
			}

			// Bulk update the vulnerability
			if len(bulkUpdateModels) > 0 {
				_, err := vulnerabilityCollection.BulkWrite(ctx, bulkUpdateModels)
				if err != nil {
					log.Printf("BulkWrite vulnerability error: %v", err)
				} else {
					logger.Info(fmt.Sprintf("Bulk updated %d vulnerabilities", len(bulkUpdateModels)))
				}
				bulkUpdateModels = []mongo.WriteModel{}
			}
		}

		bar.Add(1)
	}

	// Process the remaining bulk operations
	if len(bulkDetailModels) > 0 {
		_, err := vulnerabilityDetailCollection.BulkWrite(ctx, bulkDetailModels)
		if err != nil {
			log.Printf("BulkWrite vulnerabilityDetail error (final batch): %v", err)
		} else {
			logger.Info(fmt.Sprintf("Bulk inserted %d vulnerability details (final batch)", len(bulkDetailModels)))
		}
	}

	if len(bulkUpdateModels) > 0 {
		_, err := vulnerabilityCollection.BulkWrite(ctx, bulkUpdateModels)
		if err != nil {
			log.Printf("BulkWrite vulnerability error (final batch): %v", err)
		} else {
			logger.Info(fmt.Sprintf("Bulk updated %d vulnerabilities (final batch)", len(bulkUpdateModels)))
		}
	}

	if err := cursor.Err(); err != nil {
		logger.Error(fmt.Sprintf("Cursor error: %v", err))
		return
	}

	logger.Info("Vulnerability details migration completed.")
	logger.Info("Note: MongoDB disk space has not been reclaimed yet. Use CompactVulnerabilityCollection() to reclaim space after data migration.")
	_ = CompactVulnerabilityCollection()
}

// CompactAssetCollection compacts the asset collection to reclaim disk space
// Note: this operation blocks writes to the collection; it is recommended to run it during off-peak hours
// The compact command requires admin privileges and can only be run on a standalone node or a replica-set secondary
func CompactAssetCollection() error {
	ctx := context.Background()

	logger.Info("Starting compact operation for asset collection...")
	logger.Info("Warning: This operation may block write operations on the collection.")

	// Run the compact command
	// Note: the compact command must be executed at the database level, not the collection level
	cmd := bson.M{
		"compact": "asset",
	}
	result := mongodb.DB.RunCommand(ctx, cmd)

	var resultDoc bson.M
	if err := result.Decode(&resultDoc); err != nil {
		return fmt.Errorf("failed to execute compact command: %w", err)
	}

	logger.Info(fmt.Sprintf("Compact operation completed: %v", resultDoc))
	return nil
}

// CompactVulnerabilityCollection compacts the vulnerability collection to reclaim disk space
// Note: this operation blocks writes to the collection; it is recommended to run it during off-peak hours
// The compact command requires admin privileges and can only be run on a standalone node or a replica-set secondary
func CompactVulnerabilityCollection() error {
	ctx := context.Background()

	logger.Info("Starting compact operation for vulnerability collection...")
	logger.Info("Warning: This operation may block write operations on the collection.")

	// Run the compact command
	// Note: the compact command must be executed at the database level, not the collection level
	cmd := bson.M{
		"compact": "vulnerability",
	}
	result := mongodb.DB.RunCommand(ctx, cmd)

	var resultDoc bson.M
	if err := result.Decode(&resultDoc); err != nil {
		return fmt.Errorf("failed to execute compact command: %w", err)
	}

	logger.Info(fmt.Sprintf("Compact operation completed: %v", resultDoc))
	return nil
}
