package mongodb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Autumn-27/ScopeSentry/internal/constants"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/utils/random"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func printProgressBar(step, total int, action string) {
	width := 50
	progress := float64(step) / float64(total)
	bar := int(progress * float64(width))

	fmt.Printf("\r[%s%s] %d/%d %s\n",
		strings.Repeat("=", bar),
		strings.Repeat(" ", width-bar),
		step, total, action)
}

func CreateDatabase() error {
	client := Client
	if client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	// Get the list of databases
	databases, err := client.ListDatabaseNames(context.Background(), bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list databases: %v", err)
	}

	db := client.Database(config.GlobalConfig.MongoDB.Database)

	// If the database does not exist, create it
	if !contains(databases, config.GlobalConfig.MongoDB.Database) {
		totalSteps := 13
		currentStep := 0

		// Create the user collection
		collection := db.Collection("user")
		// An operator-supplied password wins, so a redeploy into a fresh
		// database keeps the same login instead of printing a new one that
		// nobody sees (the container log is usually gone by then).
		password := os.Getenv("ADMIN_PASSWORD")
		if password == "" {
			password, err = random.GeneratePassword(16)
			if err != nil {
				password = random.GenerateRandomString(16)
			}
		}

		// Print important information
		separator := strings.Repeat("=", 50)
		fmt.Printf("%s\n", separator)
		fmt.Println("✨✨✨ IMPORTANT NOTICE: Please review the User/Password below ✨✨✨")
		fmt.Println(separator)
		fmt.Printf("🔑 User/Password: ScopeSentry/%s\n", password)
		fmt.Println(separator)
		fmt.Println("✅ Ensure the User/Password is correctly copied!")
		fmt.Println("✅ The initialization password is stored in the file PASSWORD")

		// Save the password to a file
		if err := os.WriteFile(filepath.Join(config.GlobalConfig.System.ExeDir, "PASSWORD"), []byte(password), 0644); err != nil {
			return fmt.Errorf("failed to write password file: %v", err)
		}

		// Encrypt the password
		hashedPassword := helper.Sha256Hex(password)
		// Create the user
		_, err = collection.InsertOne(context.Background(), bson.M{
			"username": "ScopeSentry",
			"password": hashedPassword,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating user")

		// Create the config collection
		configCollection := db.Collection("config")

		// Insert system configuration
		_, err = configCollection.InsertOne(context.Background(), bson.M{
			"name":  "timezone",
			"value": "Asia/Shanghai",
			"type":  "system",
		})
		if err != nil {
			return fmt.Errorf("failed to insert timezone config: %v", err)
		}
		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "ModulesConfig",
			"value": constants.ModulesConfig,
			"type":  "system",
		})

		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "SubfinderApiConfig",
			"value": constants.SubfinderApiConfig,
			"type":  "subfinder",
		})

		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "RadConfig",
			"value": constants.RadConfig,
			"type":  "rad",
		})

		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "SearchApiConfig",
			"value": constants.SearchApiConfig,
			"type":  "searchapi",
		})

		// Insert fingerprint version
		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "FingerVersion",
			"value": "2026-01-20 23:50",
			"type":  "finger",
		})

		currentStep++
		printProgressBar(currentStep, totalSteps, "Setting timezone")

		// Create notification configuration
		_, err = configCollection.InsertOne(context.Background(), bson.M{
			"name":                          "notification",
			"dirScanNotification":           true,
			"portScanNotification":          true,
			"sensitiveNotification":         true,
			"subdomainTakeoverNotification": true,
			"pageMonNotification":           true,
			"subdomainNotification":         true,
			"vulNotification":               true,
			"type":                          "notification",
		})
		if err != nil {
			return fmt.Errorf("failed to insert notification config: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Setting notification config")

		// Create scheduled tasks
		scheduledTasksCollection := db.Collection("ScheduledTasks")
		_, err = scheduledTasksCollection.InsertOne(context.Background(), bson.M{
			"id":      "page_monitoring",
			"name":    "Page Monitoring",
			"hour":    24,
			"node":    []string{},
			"allNode": true,
			"type":    "Page Monitoring",
			"state":   true,
		})
		if err != nil {
			return fmt.Errorf("failed to insert scheduled task: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating scheduled tasks")

		// Create the notification collection
		err = db.CreateCollection(context.Background(), "notification")
		if err != nil {
			return fmt.Errorf("failed to create notification collection: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating notification collection")

		// Create the dictionary collection
		dictionaryCollection := db.Collection("dictionary")
		// Insert the directory scan dictionary
		dirDict := constants.DirDict
		size := float64(len(dirDict)) / (1024 * 1024)
		result, err := dictionaryCollection.InsertOne(context.Background(), bson.M{
			"name":     "default",
			"category": "dir",
			"size":     fmt.Sprintf("%.2f", size),
		})
		if err != nil {
			return fmt.Errorf("failed to insert dir dictionary: %v", err)
		}
		// Store the dictionary content in GridFS
		if err := CreateGridFSFile(db, result.InsertedID.(primitive.ObjectID).Hex(), []byte(dirDict)); err != nil {
			return fmt.Errorf("failed to write dir dictionary: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating dir dictionary")

		// Insert the subdomain dictionary
		domainDict := constants.DomainDict
		size = float64(len(domainDict)) / (1024 * 1024)
		result, err = dictionaryCollection.InsertOne(context.Background(), bson.M{
			"name":     "default",
			"category": "subdomain",
			"size":     fmt.Sprintf("%.2f", size),
		})
		if err != nil {
			return fmt.Errorf("failed to insert subdomain dictionary: %v", err)
		}
		if err := CreateGridFSFile(db, result.InsertedID.(primitive.ObjectID).Hex(), []byte(domainDict)); err != nil {
			return fmt.Errorf("failed to write subdomain dictionary: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating subdomain dictionary")

		// Insert sensitive information rules
		sensitiveCollection := db.Collection("SensitiveRule")
		sensitiveData, _ := constants.GetSensitive()
		if len(sensitiveData) > 0 {
			_, err = sensitiveCollection.InsertMany(context.Background(), sensitiveData)
			if err != nil {
				return fmt.Errorf("failed to insert sensitive rules: %v", err)
			}
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating sensitive rules")

		// Insert default ports
		portCollection := db.Collection("PortDict")
		portData, _ := constants.GetPort()
		_, err = portCollection.InsertMany(context.Background(), portData)
		if err != nil {
			return fmt.Errorf("failed to insert port dictionary: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating port dictionary")

		// Insert POCs
		//pocCollection := db.Collection("PocList")
		//pocData := getPoc()
		//if len(pocData) > 0 {
		//	_, err = pocCollection.InsertMany(context.Background(), pocData)
		//	if err != nil {
		//		return fmt.Errorf("failed to insert POC: %v", err)
		//	}
		//}
		//currentStep++
		//printProgressBar(currentStep, totalSteps, "Creating POC")

		// Insert fingerprint rules
		fingerprintCollection := db.Collection("FingerprintRules")
		fingerprint, _ := constants.GetFingerprintData()
		if len(fingerprint) > 0 {
			_, err = fingerprintCollection.InsertMany(context.Background(), fingerprint)
			if err != nil {
				return fmt.Errorf("failed to insert fingerprint rules: %v", err)
			}
		}

		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating fingerprint rules")

		// Create default plugins
		pluginsCollection := db.Collection("plugins")
		var plgDocs []interface{}
		for _, p := range constants.Plugins {
			plgDocs = append(plgDocs, p)
		}
		_, err = pluginsCollection.InsertMany(context.Background(), plgDocs)
		if err != nil {
			return fmt.Errorf("failed to insert plugins: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating plugins")

		// Create the default scan template
		scanTemplateCollection := db.Collection("ScanTemplates")
		_, err = scanTemplateCollection.InsertOne(context.Background(), constants.ScanTemplateDefault)
		if err != nil {
			return fmt.Errorf("failed to insert scan template: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating scan template")

		// Create indexes
		if err := createIndexes(db); err != nil {
			return fmt.Errorf("failed to create indexes: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating indexes")

		fmt.Println() // Newline
		logger.Info("Project initialization successful")
	} else {
		// Database already exists, check and update necessary configuration
		configCollection := db.Collection("config")
		var result bson.M
		err := configCollection.FindOne(context.Background(), bson.M{"name": "timezone"}).Decode(&result)
		if err != nil {
			return fmt.Errorf("failed to get timezone config: %v", err)
		}

		// Check scheduled tasks
		scheduledTasksCollection := db.Collection("ScheduledTasks")
		var taskResult bson.M
		err = scheduledTasksCollection.FindOne(context.Background(), bson.M{"id": "page_monitoring"}).Decode(&taskResult)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// Create default scheduled tasks
				_, err = scheduledTasksCollection.InsertOne(context.Background(), bson.M{
					"id":    "page_monitoring",
					"name":  "Page Monitoring",
					"hour":  24,
					"type":  "Page Monitoring",
					"state": true,
				})
				if err != nil {
					return fmt.Errorf("failed to insert default scheduled task: %v", err)
				}
			} else {
				return fmt.Errorf("failed to check scheduled task: %v", err)
			}
		}
	}

	return nil
}

func createIndexes(db *mongo.Database) error {
	// Create page monitoring indexes
	pageMonitoring := db.Collection("PageMonitoring")
	_, err := pageMonitoring.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{"url", 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create PageMonitoring index: %v", err)
	}

	// Create page monitoring content indexes
	pageMonitoringBody := db.Collection("PageMonitoringBody")
	_, err = pageMonitoringBody.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{"md5", 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create PageMonitoringBody index: %v", err)
	}

	// Create asset indexes
	asset := db.Collection("asset")
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"time", -1}}},
		{Keys: bson.D{{"url", 1}}},
		{Keys: bson.D{{"ip", 1}}},
		{Keys: bson.D{{"port", 1}}},
		{Keys: bson.D{{"host", 1}, {"port", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"project", 1}, {"time", -1}}},
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
		// 1.8 added indexes
		{Keys: bson.D{{"tags", 1}}},
		{Keys: bson.D{{"technologies", 1}}},
		{Keys: bson.D{{"faviconmmh3", 1}}},
		{Keys: bson.D{{"bodyhash", 1}}},
		// 1.8.1 added indexes
		{Keys: bson.D{{"host", 1}}},

		{Keys: bson.D{{"rootDomain", 1}, {"time", -1}}}, // Optimize performance for queries by rootDomain sorted by time
	}
	_, err = asset.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create asset indexes: %v", err)
	}

	// Create IP asset temporary collection indexes
	ipAssetTmp := db.Collection("IPAssetTmp")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"ip", 1}, {"port", 1}, {"domain", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = ipAssetTmp.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create ipAssetTmp indexes: %v", err)
	}

	// Create IP asset collection indexes
	ipAsset := db.Collection("IPAsset")
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
		return fmt.Errorf("failed to create ipAsset indexes: %v", err)
	}

	// Create icon collection indexes
	icon := db.Collection("icon")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"fav3", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = icon.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create icon indexes: %v", err)
	}

	// Create http res body collection indexes
	httpBody := db.Collection("HttpBody")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = httpBody.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create httpBody indexes: %v", err)
	}

	// Create http screenshot indexes
	httpScreenshot := db.Collection("screenshot")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = httpScreenshot.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create httpBody indexes: %v", err)
	}

	// Create subdomain indexes
	subdomain := db.Collection("subdomain")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
		{Keys: bson.D{{"time", 1}}},
	}
	_, err = subdomain.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create subdomain indexes: %v", err)
	}

	// Create URL scan indexes
	urlScan := db.Collection("UrlScan")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
	}
	_, err = urlScan.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create UrlScan indexes: %v", err)
	}

	// Create sensitive information body indexes
	SensitiveBody := db.Collection("SensitiveBody")
	_, err = SensitiveBody.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{"md5", 1}},
	})
	if err != nil {
		fmt.Errorf("failed to create SensitiveBody index: %v", err)
	}
	// Create crawler indexes
	crawler := db.Collection("crawler")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
	}
	_, err = crawler.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create crawler indexes: %v", err)
	}

	// Create sensitive information result indexes
	sensitiveResult := db.Collection("SensitiveResult")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
	}
	_, err = sensitiveResult.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create SensitiveResult indexes: %v", err)
	}

	// Create directory scan result indexes
	dirScanResult := db.Collection("DirScanResult")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
	}
	_, err = dirScanResult.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create DirScanResult indexes: %v", err)
	}

	// Create vulnerability indexes
	vulnerability := db.Collection("vulnerability")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
		{Keys: bson.D{{"hash", 1}}},
	}
	_, err = vulnerability.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create vulnerability indexes: %v", err)
	}

	// Create vulDetail indexes
	vulnerabilityDetail := db.Collection("vulnerabilityDetail")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}},
	}
	_, err = vulnerabilityDetail.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create vulnerabilityDetail indexes: %v", err)
	}

	// Create root domain indexes
	rootDomain := db.Collection("RootDomain")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"domain", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"time", 1}}},
	}
	_, err = rootDomain.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create RootDomain indexes: %v", err)
	}

	// Create application indexes
	app := db.Collection("app")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"time", 1}}},
		{Keys: bson.D{{"name", 1}}},
	}
	_, err = app.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create app indexes: %v", err)
	}

	// Create MP indexes
	mp := db.Collection("mp")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"time", 1}}},
		{Keys: bson.D{{"name", 1}}},
	}
	_, err = mp.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create mp indexes: %v", err)
	}

	// Create task-related indexes
	task := db.Collection("task")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"creatTime", -1}}},
		{Keys: bson.D{{"name", 1}}},
		{Keys: bson.D{{"progress", 1}}},
	}
	_, err = task.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create task indexes: %v", err)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
