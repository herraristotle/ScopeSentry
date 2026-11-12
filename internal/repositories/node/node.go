package node

import (
	"context"
	"strings"

	"github.com/Autumn-27/ScopeSentry/internal/models"

	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/database/redis"

	redisdriver "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	GetAllNodes(ctx context.Context) ([]models.NodeData, error)
	UpdateNodeState(ctx context.Context, nodeID string, state string) error
	RefreshConfig(ctx context.Context, node string, msg string) error
	DeleteNodes(ctx context.Context, names []string) error
	GetNodeLogs(ctx context.Context, name string) (string, error)
	GetNodePlugin(ctx context.Context, nodeName string) ([]models.NodePluginInfo, error)
}

type repository struct {
	redisClient *redisdriver.Client
	collection  *mongo.Collection
}

func NewRepository() Repository {
	return &repository{
		redisClient: redis.Client,
		collection:  mongodb.DB.Collection("plugins"),
	}
}

// GetAllNodes gets all nodes
func (r *repository) GetAllNodes(ctx context.Context) ([]models.NodeData, error) {
	// Get all keys starting with node:
	keys, err := r.redisClient.Keys(ctx, "node:*").Result()
	if err != nil {
		return nil, err
	}

	var nodes []models.NodeData
	for _, key := range keys {
		// Get all fields and values in the hash
		hashData, err := r.redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		name := strings.TrimPrefix(key, "node:")

		// Build the struct, providing default values
		node := models.NodeData{
			Name:          name,
			State:         hashData["state"],
			UpdateTime:    hashData["updateTime"],
			Running:       hashData["running"],
			Finished:      hashData["finished"],
			CPUNum:        hashData["cpuNum"],
			MemNum:        hashData["memNum"],
			MaxTaskNum:    hashData["maxTaskNum"],
			Version:       hashData["version"],
			ModulesConfig: hashData["modulesConfig"],
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// UpdateNodeState updates the node state
func (r *repository) UpdateNodeState(ctx context.Context, nodeID string, state string) error {
	key := "node:" + nodeID
	_, err := r.redisClient.HSet(ctx, key, "state", state).Result()
	return err
}

// RefreshConfig inserts a task at the right end of the specified node's task list
func (r *repository) RefreshConfig(ctx context.Context, node string, msg string) error {
	key := "refresh_config:" + node
	_, err := r.redisClient.RPush(ctx, key, msg).Result()
	return err
}

// DeleteNodes deletes multiple nodes
func (r *repository) DeleteNodes(ctx context.Context, names []string) error {
	for _, name := range names {
		key := "node:" + name
		_, err := r.redisClient.Del(ctx, key).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetNodeLogs gets node logs
func (r *repository) GetNodeLogs(ctx context.Context, name string) (string, error) {
	logKey := "log:" + name
	logs, err := r.redisClient.LRange(ctx, logKey, 0, -1).Result()
	if err != nil {
		return "", err
	}
	return strings.Join(logs, ""), nil
}

// GetNodePlugin gets node plugin information
func (r *repository) GetNodePlugin(ctx context.Context, nodeName string) ([]models.NodePluginInfo, error) {
	// Get all plugin information from MongoDB
	projection := bson.M{
		"_id":    0,
		"name":   1,
		"hash":   1,
		"module": 1,
	}

	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plugins []models.Plugin
	if err := cursor.All(ctx, &plugins); err != nil {
		return nil, err
	}

	// Build the plugin map
	pluginMap := make(map[string]models.Plugin)
	for _, plugin := range plugins {
		pluginMap[plugin.Hash] = plugin
	}

	// Get the node plugin states from Redis
	key := "NodePlg:" + nodeName
	hashData, err := r.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result []models.NodePluginInfo
	for hash, plugin := range pluginMap {
		installValue := hashData[hash+"_install"]
		checkValue := hashData[hash+"_check"]

		// If there is no data in Redis, default to "0"
		if installValue == "" {
			installValue = "0"
		}
		if checkValue == "" {
			checkValue = "0"
		}

		result = append(result, models.NodePluginInfo{
			Name:    plugin.Name,
			Install: installValue,
			Check:   checkValue,
			Hash:    hash,
			Module:  plugin.Module,
		})
	}

	return result, nil
}
