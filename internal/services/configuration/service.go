// configuration-------------------------------------
// @file      : service.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/10/29 21:20
// -------------------------------------------

package configuration

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	assetCommon "github.com/Autumn-27/ScopeSentry/internal/services/assets/common"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	nservice "github.com/Autumn-27/ScopeSentry/internal/services/node"

	"github.com/Autumn-27/ScopeSentry/internal/repositories/common"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type Service struct {
	repo         common.Repository
	nodeService  nservice.Service
	dedupService assetCommon.DedupService
}

func NewService() *Service {
	return &Service{
		repo:         common.NewRepository(),
		nodeService:  nservice.NewService(),
		dedupService: assetCommon.NewDedupService(),
	}
}

const (
	collConfig       = "config"
	collNotification = "notification"
)

// GetSubfinderContent reads SubfinderApiConfig
func (s *Service) GetSubfinderContent(ctx *gin.Context) (string, error) {
	doc, err := s.repo.FindOne(ctx.Request.Context(), collConfig, bson.M{"name": "SubfinderApiConfig"}, bson.M{"_id": 0})
	if err != nil {
		return "", err
	}
	if v, ok := doc["value"].(string); ok {
		return v, nil
	}
	return "", nil
}

// SaveSubfinderContent saves SubfinderApiConfig and notifies nodes to refresh
func (s *Service) SaveSubfinderContent(ctx *gin.Context, content string) error {
	if err := s.repo.Upsert(ctx.Request.Context(), collConfig, bson.M{"name": "SubfinderApiConfig"}, bson.M{"name": "SubfinderApiConfig", "value": content}); err != nil {
		return err
	}
	// Notify all nodes to refresh the subfinder config
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "subfinder"})
	return nil
}

// GetRadContent reads RadConfig
func (s *Service) GetRadContent(ctx *gin.Context) (string, error) {
	doc, err := s.repo.FindOne(ctx.Request.Context(), collConfig, bson.M{"name": "RadConfig"}, bson.M{"_id": 0})
	if err != nil {
		return "", err
	}
	if v, ok := doc["value"].(string); ok {
		return v, nil
	}
	return "", nil
}

// SaveRadContent saves RadConfig and notifies nodes to refresh
func (s *Service) SaveRadContent(ctx *gin.Context, content string) error {
	if err := s.repo.Upsert(ctx.Request.Context(), collConfig, bson.M{"name": "RadConfig"}, bson.M{"name": "RadConfig", "value": content}); err != nil {
		return err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "rad"})
	return nil
}

// GetSystemData returns all key-value pairs with type=system
func (s *Service) GetSystemData(ctx *gin.Context) (map[string]interface{}, error) {
	list, err := s.repo.FindMany(ctx.Request.Context(), collConfig, bson.M{"type": "system"}, nil)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{})
	for _, m := range list {
		name, _ := m["name"].(string)
		if name == "" {
			continue
		}
		out[name] = m["value"]
	}
	return out, nil
}

// SaveSystemData saves the system config and notifies nodes
func (s *Service) SaveSystemData(ctx *gin.Context, kv map[string]interface{}) error {
	for k, v := range kv {
		doc := bson.M{"type": "system", "name": k, "value": v}
		if err := s.repo.Upsert(ctx.Request.Context(), collConfig, bson.M{"type": "system", "name": k}, doc); err != nil {
			return err
		}
	}
	timezone, _ := kv["timezone"].(string)
	modulesConfig := fmt.Sprintf("%v", kv["ModulesConfig"]) // Compatible with string or other types
	msg := timezone + "[*]" + modulesConfig
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "system", Content: msg})
	return nil
}

// GetDeduplicationConfig queries the deduplication config and the next run time (if any)
func (s *Service) GetDeduplicationConfig(ctx *gin.Context) (map[string]interface{}, error) {
	doc, err := s.repo.FindOne(ctx.Request.Context(), collConfig, bson.M{"name": "deduplication"}, bson.M{"_id": 0})
	if err != nil {
		// If it does not exist, return an empty map
		return map[string]interface{}{}, nil
	}
	return doc, nil
}

// SaveDeduplicationConfig saves the deduplication config; if an immediate run is needed it is passed through to the node or scheduler (here it only saves and refreshes)
func (s *Service) SaveDeduplicationConfig(ctx *gin.Context, cfg models.DepConfig) error {
	cfg.Name = "deduplication"
	runNow := cfg.RunNow
	cfg.RunNow = false
	if err := s.repo.Upsert(ctx.Request.Context(), collConfig, bson.M{"name": "deduplication"}, cfg); err != nil {
		return err
	}
	// This can be extended as needed to call an internal scheduler or dispatch messages; the current version only saves
	if runNow {
		go func() {
			err := s.dedupService.DoAssetDeduplication()
			if err != nil {
				logger.Error(err.Error())
				return
			}
		}()
	}
	return nil
}

// GetNotificationList gets the notification list
func (s *Service) GetNotificationList(ctx *gin.Context) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.repo.Find(ctx.Request.Context(), collNotification, bson.M{}, nil, &notifications)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// AddNotification adds a notification and refreshes nodes
func (s *Service) AddNotification(ctx *gin.Context, data models.Notification) error {
	_, err := s.repo.InsertOne(ctx.Request.Context(), collNotification, data)
	if err != nil {
		return err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "notification"})
	return nil
}

// UpdateNotification updates a notification and refreshes nodes
func (s *Service) UpdateNotification(ctx *gin.Context, data models.UpdateNotification) error {
	if data.ID == "" {
		return fmt.Errorf("missing id")
	}
	objID, err := primitive.ObjectIDFromHex(data.ID)
	if err != nil {
		return fmt.Errorf("invalid ObjectID: %v", err)
	}
	if err := s.repo.UpdateOne(ctx.Request.Context(), collNotification, bson.M{"_id": objID}, bson.M{"$set": data}); err != nil {
		return err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "notification"})
	if err := mongodb.FindAll("notification", bson.M{"state": true}, bson.M{"_id": 0, "method": 1, "url": 1, "contentType": 1, "data": 1, "state": 1}, &config.NotificationApi); err != nil {
		return err
	}
	return nil
}

// DeleteNotifications batch-deletes notifications and refreshes nodes
func (s *Service) DeleteNotifications(ctx *gin.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	// Convert the string to ObjectID
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid ObjectID: %s, error: %w", id, err)
		}
		objectIDs = append(objectIDs, objID)
	}
	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	if _, err := s.repo.DeleteMany(ctx.Request.Context(), collNotification, filter); err != nil {
		return err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "notification"})
	return nil
}

// GetNotificationConfig reads the notification config
func (s *Service) GetNotificationConfig(ctx *gin.Context) (map[string]interface{}, error) {
	doc, err := s.repo.FindOne(ctx.Request.Context(), collConfig, bson.M{"name": "notification"}, bson.M{})
	if err != nil {
		return map[string]interface{}{}, nil
	}
	delete(doc, "_id")
	delete(doc, "type")
	delete(doc, "name")
	return doc, nil
}

// UpdateNotificationConfig updates the notification config and refreshes nodes
func (s *Service) UpdateNotificationConfig(ctx *gin.Context, data map[string]interface{}) error {
	if data == nil {
		return nil
	}
	if err := s.repo.UpdateOne(ctx.Request.Context(), collConfig, bson.M{"name": "notification"}, bson.M{"$set": data}); err != nil {
		return err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "notification"})
	return nil
}
