// update-------------------------------------
// @file      : update21.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2026/11/19 10:00
// -------------------------------------------

package update

import (
	"context"
	"fmt"

	"github.com/Autumn-27/ScopeSentry/internal/constants"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// Update21 registers the FOFA and Shodan asset-search plugins in the catalog
// and seeds the credential document they read. Existing installs never rerun
// initdb, so without this the two plugins would be invisible in the UI.
func Update21() {
	ctx := context.Background()
	upsert := mongoOptions.Update().SetUpsert(true)

	pluginsCollection := mongodb.DB.Collection("plugins")
	for _, plg := range constants.Plugins {
		if plg.Module != "TargetHandler" {
			continue
		}
		// Match on hash: the two new entries are the only TargetHandler plugins
		// in the catalog, and a re-run must not duplicate them.
		_, err := pluginsCollection.UpdateOne(ctx,
			bson.M{"hash": plg.Hash},
			bson.M{"$setOnInsert": plg},
			upsert,
		)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to register asset-search plugin %s: %v", plg.Name, err))
		}
	}

	configCollection := mongodb.DB.Collection("config")
	_, err := configCollection.UpdateOne(ctx,
		bson.M{"name": "SearchApiConfig"},
		bson.M{"$setOnInsert": bson.M{
			"name":  "SearchApiConfig",
			"value": constants.SearchApiConfig,
			"type":  "searchapi",
		}},
		upsert,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to seed SearchApiConfig: %v", err))
	}
}
