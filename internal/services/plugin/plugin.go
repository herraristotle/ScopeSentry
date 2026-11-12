package plugin

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/utils"
	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/plugins"
	schedulerCore "github.com/Autumn-27/ScopeSentry/internal/scheduler"

	"github.com/Autumn-27/ScopeSentry/internal/constants"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/options"
	"github.com/Autumn-27/ScopeSentry/internal/repositories/plugin"
	"github.com/Autumn-27/ScopeSentry/internal/services/node"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Service defines the plugin service interface
type Service interface {
	List(ctx *gin.Context, req *models.PluginListRequest) (*models.PluginListResponse, error)
	ListByModule(ctx *gin.Context, module string) ([]models.Plugin, error)
	GetPluginByHash(ctx context.Context, hash string) (*models.Plugin, error)
	Detail(ctx *gin.Context, req *models.PluginDetailRequest) (*models.Plugin, error)
	Save(ctx *gin.Context, req *models.PluginSaveRequest) error
	Delete(ctx *gin.Context, req *models.PluginDeleteRequest) error
	GetLogs(ctx *gin.Context, req *models.PluginLogRequest) (string, error)
	CleanLogs(ctx *gin.Context, req *models.PluginLogRequest) error
	CleanAllLogs(ctx *gin.Context) error
	Import(ctx *gin.Context, filePath string, key string) error
	Reinstall(ctx *gin.Context, req *models.PluginReinstallRequest) error
	Recheck(ctx *gin.Context, req *models.PluginRecheckRequest) error
	Uninstall(ctx *gin.Context, req *models.PluginUninstallRequest) error
	CheckKey(ctx *gin.Context, req *models.PluginKeyCheckRequest) error
	SearchRemotePlugins(ctx *gin.Context) (map[string]interface{}, error)
	ImportByData(ctx *gin.Context, req *models.PluginImportByDataRequest) error
	UpdateStatus(ctx *gin.Context, req *models.PluginStatusRequest) error
	Run(ctx *gin.Context, req *models.PluginRunRequest) error
	RunTaskEnd(task models.Task) error
}

// service implements the plugin service interface
type service struct {
	repo        plugin.Repository
	nodeService node.Service
}

// NewService creates a new plugin service instance
func NewService() Service {
	return &service{
		repo:        plugin.NewRepository(),
		nodeService: node.NewService(),
	}
}

// List gets the plugin list
func (s *service) List(ctx *gin.Context, req *models.PluginListRequest) (*models.PluginListResponse, error) {
	query := bson.M{}
	if req.Type != "all" {
		if req.Type != "server" {
			query["type"] = bson.M{"$ne": "server"}
		} else {
			query["type"] = req.Type
		}
	}
	if req.Search != "" {
		query["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	opts := mongoOptions.Find().
		SetSkip(int64((req.PageIndex - 1) * req.PageSize)).
		SetLimit(int64(req.PageSize))

	plugins, err := s.repo.FindWithPagination(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find plugins: %w", err)
	}

	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count plugins: %w", err)
	}

	return &models.PluginListResponse{
		List:  plugins,
		Total: total,
	}, nil
}

// ListByModule gets the plugin list by module
func (s *service) ListByModule(ctx *gin.Context, module string) ([]models.Plugin, error) {
	result, err := s.repo.FindByModule(ctx, module)
	if result == nil {
		return []models.Plugin{}, err
	}
	return result, err
}

// GetPluginByHash gets a single plugin by hash
func (s *service) GetPluginByHash(ctx context.Context, hash string) (*models.Plugin, error) {
	return s.repo.FindByHash(ctx, hash)
}

// Detail gets plugin details
func (s *service) Detail(ctx *gin.Context, req *models.PluginDetailRequest) (*models.Plugin, error) {
	id, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %w", err)
	}

	return s.repo.FindByID(ctx, id)
}

// Save saves a plugin
func (s *service) Save(ctx *gin.Context, req *models.PluginSaveRequest) error {
	// Validate the key
	key, err := ioutil.ReadFile("PLUGINKEY")
	if err != nil {
		return fmt.Errorf("failed to read plugin key: %w", err)
	}
	if req.Key != strings.TrimSpace(string(key)) {
		return fmt.Errorf("invalid plugin key")
	}

	if req.ID == "" {
		// Create a new plugin
		plugin := &models.Plugin{
			Name:          req.Name,
			Module:        req.Module,
			Hash:          generatePluginHash(32, req.Type),
			Parameter:     req.Parameter,
			Help:          req.Help,
			Introduction:  req.Introduction,
			Source:        req.Source,
			Version:       req.Version,
			ParameterList: req.ParameterList,
			Status:        true,
			Type:          req.Type,
			IsSystem:      false,
		}
		id, err := s.repo.Create(ctx, plugin)
		if err != nil {
			return fmt.Errorf("failed to create plugin: %w", err)
		}
		if req.Type != "server" {
			go func() {
				msg := models.Message{
					Name:    "all",
					Type:    "install_plugin",
					Content: fmt.Sprintf(`%v`, id),
				}
				err = s.nodeService.RefreshConfig(ctx, msg)
				if err != nil {
					logger.Error("failed to refresh config", zap.Error(err))
				}
			}()
		} else {
			RegisterPlugin(plugin.Source, plugin.Hash)
		}
	} else {
		// Update the existing plugin
		id, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			return fmt.Errorf("invalid id format: %w", err)
		}

		update := bson.M{
			"name":          req.Name,
			"module":        req.Module,
			"parameter":     req.Parameter,
			"help":          req.Help,
			"introduction":  req.Introduction,
			"parameterList": req.ParameterList,
			"source":        req.Source,
			"version":       req.Version,
		}
		err = s.repo.Update(ctx, id, update)
		if err != nil {
			return err
		}
		if req.Type != "server" {
			go func() {
				msg := models.Message{
					Name:    "all",
					Type:    "install_plugin",
					Content: fmt.Sprintf(`%v`, req.ID),
				}
				err = s.nodeService.RefreshConfig(ctx, msg)
				if err != nil {
					logger.Error("failed to refresh config", zap.Error(err))
				}
			}()
		} else {
			// If it is an update, re-register the plugin as well
			RegisterPlugin(req.Source, req.Hash)
		}
	}
	// Server-side plugin: whether inserting or updating, check whether cycle is 1; if it is 1, run only once
	if req.Type == "server" {
		plugins.RunPluginOnce(req.Hash)
	}
	return nil
}

// Delete deletes a plugin
func (s *service) Delete(ctx *gin.Context, req *models.PluginDeleteRequest) error {
	var hashes []string
	for _, item := range req.Data {
		hashes = append(hashes, item.Hash)
		go func() {
			if item.Module != "" {
				msg := models.Message{
					Name:    "all",
					Type:    "delete_plugin",
					Content: fmt.Sprintf(`%v_%v`, item.Hash, item.Module),
				}
				err := s.nodeService.RefreshConfig(ctx, msg)
				if err != nil {
					logger.Error("failed to refresh config", zap.Error(err))
				}
			} else {
				// An empty module means it is a server-side plugin
				// Delete the plugin
				plugins.GlobalPluginManager.DeletePlugin(item.Hash)
				// Remove the scheduled task
				err := schedulerCore.GetGlobalScheduler().RemoveJob(item.Hash)
				if err != nil {
					logger.Error(fmt.Sprintf("failed to remove job: %v", err))
				}
			}
		}()
	}
	return s.repo.DeleteByHash(ctx, hashes)
}

// GetLogs gets plugin logs
func (s *service) GetLogs(ctx *gin.Context, req *models.PluginLogRequest) (string, error) {
	logKey := ""
	if req.Type == "server" {
		logKey = fmt.Sprintf("logs:server_plugins:%v", req.Hash)
	} else {
		logKey = fmt.Sprintf("logs:plugins:%v:%v", req.Module, req.Hash)
	}
	logs, err := s.repo.GetLogs(ctx, logKey)
	if err != nil {
		return "", err
	}
	return logs, nil
}

// CleanLogs cleans up plugin logs
func (s *service) CleanLogs(ctx *gin.Context, req *models.PluginLogRequest) error {
	if req.Hash == "" {
		return fmt.Errorf("module and hash are required")
	}
	var logKey string
	if req.Type == "server" {
		logKey = fmt.Sprintf("logs:server_plugins:%v", req.Hash)
	} else {
		logKey = fmt.Sprintf("logs:plugins:%v:%v", req.Module, req.Hash)
	}
	err := s.repo.CleanLogs(ctx, logKey)
	if err != nil {
		return err
	}
	return nil
}

// CleanAllLogs clears all plugin logs
func (s *service) CleanAllLogs(ctx *gin.Context) error {
	err := s.repo.CleanAllLogs(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Import imports a plugin
func (s *service) Import(ctx *gin.Context, filePath string, reqKey string) error {
	key, err := ioutil.ReadFile("PLUGINKEY")
	if err != nil {
		return fmt.Errorf("failed to read plugin key: %w", err)
	}
	if reqKey != strings.TrimSpace(string(key)) {
		return fmt.Errorf("invalid plugin key")
	}
	// 2. Read the zip file's byte content
	zipData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read zip file: %w", err)
	}

	// 3. Initialize the zip.Reader
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}

	var pluginInfo models.PluginInfo
	var pluginSource string

	// 4. Iterate over the files in the archive
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		baseName := strings.ToLower(filepath.Base(file.Name))

		f, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}
		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("failed to read file content in archive: %w", err)
		}

		switch baseName {
		case "info.json":
			if err := json.Unmarshal(content, &pluginInfo); err != nil {
				return fmt.Errorf("failed to parse info.json: %w", err)
			}
		case "plugin.go":
			pluginSource = string(content)
		}
	}

	// 5. Validate the key fields of info.json
	if pluginInfo.Name == "" {
		return fmt.Errorf("info.json is missing the name field")
	}
	if pluginInfo.Module == "" && pluginInfo.Type != "server" {
		return fmt.Errorf("info.json is missing the module field")
	}

	if pluginInfo.Hash == "" {
		pluginInfo.Hash = generatePluginHash(32, pluginInfo.Type)
	}
	var PLUGINS = make(map[string]bool)
	for _, p := range constants.Plugins {
		PLUGINS[p.Hash] = true
	}
	// 6. Whether the plugin is a system plugin or the module is invalid
	if PLUGINS[pluginInfo.Hash] {
		return fmt.Errorf("plugin already exists")
	}
	var pluginsModules = make(map[string]bool)
	for _, module := range constants.PLUGINSMODULES {
		pluginsModules[module] = true
	}
	if pluginInfo.Type != "server" {
		if !pluginsModules[pluginInfo.Module] {
			return fmt.Errorf("invalid module: %s", pluginInfo.Module)
		}
	}
	// 7. Set the other fields
	pluginInfo.Source = pluginSource
	pluginInfo.IsSystem = false

	// 8. Insert into the database

	// 9. Trigger a configuration refresh

	p := &models.Plugin{
		Name:          pluginInfo.Name,
		Module:        pluginInfo.Module,
		Hash:          pluginInfo.Hash,
		Parameter:     pluginInfo.Parameter,
		Help:          pluginInfo.Help,
		Introduction:  pluginInfo.Introduction,
		ParameterList: pluginInfo.ParameterList,
		Source:        pluginInfo.Source,
		Version:       pluginInfo.Version,
		Type:          pluginInfo.Type,
		IsSystem:      false,
	}
	id, err := s.repo.Create(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}
	go func() {
		if pluginInfo.Type != "server" {
			msg := models.Message{
				Name:    "all",
				Type:    "install_plugin",
				Content: fmt.Sprintf(`%v`, id),
			}
			err = s.nodeService.RefreshConfig(ctx, msg)
			if err != nil {
				logger.Error("failed to refresh config", zap.Error(err))
			}
		}
	}()
	if pluginInfo.Type == "server" {
		RegisterPlugin(pluginInfo.Source, pluginInfo.Hash)
		// Determine whether it should run once
		plugins.RunPluginOnce(pluginInfo.Hash)
	}

	return nil
}

// Reinstall reinstalls a plugin
func (s *service) Reinstall(ctx *gin.Context, req *models.PluginReinstallRequest) error {
	if req.Node == "" {
		req.Node = "all"
	}
	if req.Module == "" {
		// It means this is a server-side plugin
		plg, flag := plugins.GlobalPluginManager.GetPlugin(req.Hash)
		if !flag {
			logger.Error(fmt.Sprintf("<UNK> plugin %s <UNK>", req.Hash))
			return fmt.Errorf("<UNK> plugin %s <UNK>", req.Hash)
		}
		plg.Install()
	} else {
		msg := models.Message{
			Name:    req.Node,
			Type:    "re_install_plugin",
			Content: fmt.Sprintf(`%v_%v`, req.Hash, req.Module),
		}
		err := s.nodeService.RefreshConfig(ctx, msg)
		if err != nil {
			logger.Error("failed to refresh config", zap.Error(err))
		}
	}
	return nil
}

// Recheck rechecks a plugin
func (s *service) Recheck(ctx *gin.Context, req *models.PluginRecheckRequest) error {
	if req.Node == "" {
		req.Node = "all"
	}
	msg := models.Message{
		Name:    req.Node,
		Type:    "re_check_plugin",
		Content: fmt.Sprintf(`%v_%v`, req.Hash, req.Module),
	}
	err := s.nodeService.RefreshConfig(ctx, msg)
	if err != nil {
		logger.Error("failed to refresh config", zap.Error(err))
	}
	return nil
}

// Uninstall uninstalls a plugin
func (s *service) Uninstall(ctx *gin.Context, req *models.PluginUninstallRequest) error {

	if req.Node == "" {
		req.Node = "all"
	}
	msg := models.Message{
		Name:    req.Node,
		Type:    "uninstall_plugin",
		Content: fmt.Sprintf(`%v_%v`, req.Hash, req.Module),
	}
	err := s.nodeService.RefreshConfig(ctx, msg)
	if err != nil {
		logger.Error("failed to refresh config", zap.Error(err))
	}
	return nil
}

// CheckKey checks the plugin key
func (s *service) CheckKey(ctx *gin.Context, req *models.PluginKeyCheckRequest) error {
	key, err := ioutil.ReadFile("PLUGINKEY")
	if err != nil {
		return fmt.Errorf("failed to read plugin key: %w", err)
	}
	if req.Key != strings.TrimSpace(string(key)) {
		return fmt.Errorf("invalid plugin key")
	}
	return nil
}

// RemotePluginResponse is the remote plugin API response struct
type RemotePluginResponse struct {
	Timestamp int64            `json:"timestamp"`
	Status    string           `json:"status"`
	Message   string           `json:"message"`
	Data      RemotePluginData `json:"data"`
}

type RemotePluginData struct {
	Total int64              `json:"total"`
	Data  []RemotePluginItem `json:"data"`
}

type RemotePluginItem struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Module       string `json:"module"`
	PriceStatus  int    `json:"priceStatus"`
	Price        *int   `json:"price"`
	Hash         string `json:"hash"`
	Introduction string `json:"introduction"`
	Version      string `json:"version"`
	CreateTime   string `json:"createTime"`
	Username     string `json:"username"`
	IsInstalled  bool   `json:"isInstalled"` // whether it is installed
	NeedUpdate   bool   `json:"needUpdate"`  // whether an update is needed
}

// SearchRemotePlugins searches remote plugins and compares them with installed plugins
func (s *service) SearchRemotePlugins(ctx *gin.Context) (map[string]interface{}, error) {
	var installedPlugins []models.Plugin
	var remotePlugins RemotePluginResponse
	var installedErr, remoteErr error
	var wg sync.WaitGroup

	// Fetch installed and remote plugins concurrently
	wg.Add(2)

	// Get the installed plugins
	go func() {
		defer wg.Done()
		query := bson.M{}
		opts := mongoOptions.Find()
		installedPlugins, installedErr = s.repo.FindWithPagination(ctx, query, opts)
	}()

	// Get the remote plugins
	go func() {
		defer wg.Done()
		client := &fasthttp.Client{
			ReadTimeout: 30 * time.Second,
		}

		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI("https://api.scope-sentry.top/api/common/plugin/search")
		req.Header.SetMethod("POST")
		req.Header.SetContentType("application/json")

		// Build the request body
		requestBody := map[string]interface{}{
			"name":        "",
			"priceStatus": 2,
			"module":      "All",
			"tags":        "",
			"page":        1,
			"size":        200,
		}
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			remoteErr = fmt.Errorf("failed to marshal request body: %w", err)
			return
		}
		req.SetBody(bodyBytes)

		err = client.Do(req, resp)
		if err != nil {
			remoteErr = fmt.Errorf("failed to request remote api: %w", err)
			return
		}

		if resp.StatusCode() != http.StatusOK {
			remoteErr = fmt.Errorf("remote api returned status code: %d", resp.StatusCode())
			return
		}

		if err := json.Unmarshal(resp.Body(), &remotePlugins); err != nil {
			remoteErr = fmt.Errorf("failed to unmarshal response: %w", err)
			return
		}
	}()

	wg.Wait()

	// Check for errors
	if installedErr != nil {
		return nil, fmt.Errorf("failed to get installed plugins: %w", installedErr)
	}
	if remoteErr != nil {
		return nil, fmt.Errorf("failed to get remote plugins: %w", remoteErr)
	}

	// Build hash and version maps of the installed plugins
	installedHashMap := make(map[string]models.Plugin)
	for _, plugin := range installedPlugins {
		installedHashMap[plugin.Hash] = plugin
	}

	// Add install and update status to the remote plugins
	resultPlugins := make([]RemotePluginItem, 0, len(remotePlugins.Data.Data))
	for _, remotePlugin := range remotePlugins.Data.Data {
		installedPlugin, isInstalled := installedHashMap[remotePlugin.Hash]

		remotePlugin.IsInstalled = isInstalled

		// If installed, check whether the version needs an update
		if isInstalled {
			// Only update when the remote version is non-empty and differs from the installed version
			remoteVersion := strings.TrimSpace(remotePlugin.Version)
			installedVersion := strings.TrimSpace(installedPlugin.Version)

			if remoteVersion != "" && remoteVersion != installedVersion {
				remotePlugin.NeedUpdate = true
			} else {
				remotePlugin.NeedUpdate = false
			}
		} else {
			remotePlugin.NeedUpdate = false
		}

		resultPlugins = append(resultPlugins, remotePlugin)
	}

	return map[string]interface{}{
		"total": remotePlugins.Data.Total,
		"data":  resultPlugins,
	}, nil
}

// ImportByData imports a plugin from POST JSON data
func (s *service) ImportByData(ctx *gin.Context, req *models.PluginImportByDataRequest) error {
	// Validate the key
	key, err := ioutil.ReadFile("PLUGINKEY")
	if err != nil {
		return fmt.Errorf("failed to read plugin key: %w", err)
	}
	if req.Key != strings.TrimSpace(string(key)) {
		return fmt.Errorf("invalid plugin key")
	}

	// Parse the info.json JSON string
	var pluginInfo models.PluginInfo
	if err := json.Unmarshal([]byte(req.JSON), &pluginInfo); err != nil {
		return fmt.Errorf("failed to parse JSON string: %w", err)
	}

	// Validate the key fields
	if pluginInfo.Name == "" {
		return fmt.Errorf("json is missing the name field")
	}
	if pluginInfo.Type != "server" && pluginInfo.Module == "" {
		return fmt.Errorf("json is missing the Module field")
	}

	// If the hash is empty, generate a new hash
	if pluginInfo.Hash == "" {
		pluginInfo.Hash = generatePluginHash(32, pluginInfo.Type)
	}

	// Validate that the module is legal
	if pluginInfo.Type != "server" {
		var pluginsModules = make(map[string]bool)
		for _, module := range constants.PLUGINSMODULES {
			pluginsModules[module] = true
		}
		if !pluginsModules[pluginInfo.Module] {
			return fmt.Errorf("invalid module: %s", pluginInfo.Module)
		}
	}

	var PLUGINS = make(map[string]bool)
	for _, p := range constants.Plugins {
		PLUGINS[p.Hash] = true
	}
	// 6. Whether the plugin is a system plugin or the module is invalid
	if PLUGINS[pluginInfo.Hash] {
		req.IsSystem = true
	}
	// Create the plugin object
	p := &models.Plugin{
		Name:          pluginInfo.Name,
		Module:        pluginInfo.Module,
		Hash:          pluginInfo.Hash,
		Parameter:     pluginInfo.Parameter,
		Help:          pluginInfo.Help,
		Introduction:  pluginInfo.Introduction,
		ParameterList: pluginInfo.ParameterList,
		Source:        req.Source,
		Version:       pluginInfo.Version,
		IsSystem:      req.IsSystem,
		Type:          pluginInfo.Type,
	}

	// Use upsert: update if it exists, insert if it does not
	pluginID, err := s.repo.UpsertByHash(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to upsert plugin: %w", err)
	}

	// If it is not a built-in plugin, RefreshConfig is required
	if !req.IsSystem && pluginInfo.Type != "server" {
		go func() {
			msg := models.Message{
				Name:    "all",
				Type:    "install_plugin",
				Content: fmt.Sprintf(`%v`, pluginID),
			}
			err = s.nodeService.RefreshConfig(ctx, msg)
			if err != nil {
				logger.Error("failed to refresh config", zap.Error(err))
			}
		}()
	}
	if pluginInfo.Type == "server" {
		RegisterPlugin(req.Source, pluginInfo.Hash)
		plugins.RunPluginOnce(pluginInfo.Hash)
	}

	return nil
}

const pluginSalt = "ScopeSentry_"
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generatePluginHash(length int, tp string) string {
	// Initialize the random seed
	rand.Seed(time.Now().UnixNano())

	// Generate a random string
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	randomString := string(b)

	// Concatenate the salt value
	salted := randomString + pluginSalt + tp

	// Compute the MD5
	hash := md5.Sum([]byte(salted))

	// Return the hex string
	return hex.EncodeToString(hash[:])
}

// UpdateStatus updates the plugin status
func (s *service) UpdateStatus(ctx *gin.Context, req *models.PluginStatusRequest) error {

	update := bson.M{
		"status": req.Status,
	}
	err := s.repo.UpdateByHash(ctx, req.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}
	if req.Status {
		plg, flag := plugins.GlobalPluginManager.GetPlugin(req.ID)
		if !flag {
			logger.Error(fmt.Sprintf("failed to get plugin status: %v", req.ID))
			return fmt.Errorf("failed to get plugin: %v", req.ID)
		}
		if plg.Cycle() != "" {
			if plg.Cycle() == "1" {
				plugins.RunPluginOnce(plg.GetPluginId())
			} else {
				err := AddJob(plg.GetPluginId(), plg.Cycle())
				if err != nil {
					return err
				}
			}
		}
	} else {
		flag := utils.HasContext(req.ID)
		if flag {
			utils.CancelContext(req.ID)
		}
		err := schedulerCore.GetGlobalScheduler().RemoveJob(req.ID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to remove job: %v", err))
			return err
		}
	}
	return nil
}

// Run runs the plugin once
func (s *service) Run(ctx *gin.Context, req *models.PluginRunRequest) error {
	// Get the plugin instance from GlobalPluginManager
	plgRunner, flag := plugins.GlobalPluginManager.GetPlugin(req.Hash)
	if !flag {
		// Get the plugin information from the database
		plg, err := s.repo.FindByHash(ctx, req.Hash)
		if err != nil {
			return fmt.Errorf("failed to find plugin: %w", err)
		}
		if plg == nil {
			return fmt.Errorf("plugin not found")
		}

		// Only server-side plugins may be run
		if plg.Type != "server" {
			return fmt.Errorf("only server plugins can be run manually")
		}
		// If the plugin is not loaded, load it
		plugin, err := plugins.LoadPlugin(plg.Source, plg.Hash)
		if err != nil {
			return fmt.Errorf("failed to load plugin: %w", err)
		}
		plugins.GlobalPluginManager.RegisterPlugin(req.Hash, plugin)
		err = plugin.Install()
		if err != nil {
			logger.Error(fmt.Sprintf("failed to install plugin: %v", err))
			return fmt.Errorf("failed to install plugin: %w", err)
		}
		plgRunner = plugin.Clone()
	}

	plgRunner.Log("plugin is running manually")
	_, err := mongodb.UpdateOne("plugins", bson.M{"hash": req.Hash}, bson.M{"$set": bson.M{"lastTime": helper.GetNowTimeString()}})
	if err != nil {
		logger.Error("failed to update plugin", zap.Error(err))
	}
	// Execute the plugin
	go func() {
		err = plgRunner.Execute(options.NewOption(req.Hash))
		if err != nil {
			logger.Error(fmt.Sprintf("plugin execution error: %v", err))
			plgRunner.Log(fmt.Sprintf("plugin execution error: %v", err), "e")
		}
		plgRunner.Log("plugin execution completed")
	}()

	return nil
}

func (s *service) RunTaskEnd(task models.Task) error {
	query := bson.M{"type": "server", "status": true}
	logger.Info(fmt.Sprintf("executing task end plugin: %v", task.Name))
	opts := mongoOptions.Find()

	plgs, err := s.repo.FindWithPagination(context.Background(), query, opts)
	if err != nil {
		return fmt.Errorf("failed to find plugins: %w", err)
	}
	for _, plgTmp := range plgs {
		plg, flag := plugins.GlobalPluginManager.GetPlugin(plgTmp.Hash)
		if flag {
			go func() {
				err := plg.TaskEnd(task, plgTmp.Hash)
				if err != nil {
					logger.Error(fmt.Sprintf("plugin taskend execution error: %v", err))
					return
				}
			}()
		} else {
			logger.Error(fmt.Sprintf("plugin not found: %v", plgTmp.Hash))
		}
	}
	return nil
}
