// plugins-------------------------------------
// @file      : plugins.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/12/3 20:33
// -------------------------------------------

package plugins

import (
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/options"
	"github.com/Autumn-27/ScopeSentry/internal/symbols"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"go.mongodb.org/mongo-driver/bson"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
	"reflect"
	"sync"
)

func init() {
	GlobalPluginManager = NewPluginManager()
	InstallPlugin()
}

func LoadPlugin(code string, plgHash string) (interfaces.Plugin, error) {
	// Initialize the yaegi interpreter
	interp := interp.New(interp.Options{})
	// Load the standard library and symbols
	err := interp.Use(stdlib.Symbols)
	if err != nil {
		return nil, err
	}
	err = interp.Use(symbols.Symbols)
	if err != nil {
		return nil, err
	}
	err = interp.Use(map[string]map[string]reflect.Value{
		"os/exec": stdlib.Symbols["os/exec"],
	})
	err = interp.Use(map[string]map[string]reflect.Value{
		"gopkg.in/yaml.v3": {
			"Marshal":   reflect.ValueOf(yaml.Marshal),
			"Unmarshal": reflect.ValueOf(yaml.Unmarshal),
		},
	})
	if err != nil {
		return nil, err
	}
	_, err = interp.Eval(code)
	if err != nil {
		return nil, err
	}
	// Get Execute
	v, err := interp.Eval("plugin.Execute")
	if err != nil {
		return nil, err
	}
	executeFunc := v.Interface().(func(op options.PluginOption) error)

	// Get TaskEnd
	v, err = interp.Eval("plugin.TaskEnd")
	taskEndFunc := func(task models.Task, plgHash string) error { return nil }
	if err != nil {
	} else {
		taskEndFunc = v.Interface().(func(task models.Task, plgHash string) error)
	}

	// Get Cycle
	v, err = interp.Eval("plugin.Cycle")
	if err != nil {
		return nil, err
	}
	cycleFunc := v.Interface().(func() string)

	// Get Cycle
	v, err = interp.Eval("plugin.Install")
	if err != nil {
		return nil, err
	}
	installFunc := v.Interface().(func() error)

	// Get Cycle
	v, err = interp.Eval("plugin.GetName")
	if err != nil {
		return nil, err
	}
	getNameFunc := v.Interface().(func() string)

	plg := NewPlugin(plgHash, installFunc, executeFunc, taskEndFunc, getNameFunc, cycleFunc)

	return plg, err
}

type PluginManager struct {
	plugins map[string]interfaces.Plugin // Store plugins, indexed by module and name
	mu      sync.RWMutex
}

var GlobalPluginManager *PluginManager

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]interfaces.Plugin),
	}
}

func (pm *PluginManager) RegisterPlugin(id string, plugin interfaces.Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.plugins, id)
	pm.plugins[id] = plugin
}

func (pm *PluginManager) GetPlugin(id string) (interfaces.Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, ok := pm.plugins[id]
	if !ok || plugin == nil {
		return nil, false
	}

	return plugin.Clone(), true
}

func (pm *PluginManager) DeletePlugin(id string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.plugins, id)
}

type PluginInfo struct {
	Hash   string `bson:"hash"`
	Source string `bson:"source"`
	Model  string `bson:"model"`
	Status bool   `bson:"status"`
}

func InstallPlugin() {
	var result []PluginInfo
	opts := mongoOptions.Find().
		SetProjection(bson.M{
			"module": 1,
			"hash":   1,
			"source": 1,
			"status": 1,
			"_id":    0,
		})
	err := mongodb.FindMany(
		"plugins",
		bson.M{
			"type": "server",
		},
		&result,
		opts,
	)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	for _, plugin := range result {
		loadPlugin, err := LoadPlugin(plugin.Source, plugin.Hash)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		GlobalPluginManager.RegisterPlugin(plugin.Hash, loadPlugin)
		err = loadPlugin.Install()
		if err != nil {
			logger.Error(err.Error())
			return
		}
		// Determine at startup whether it needs to run once
		if plugin.Status == true {
			RunPluginOnce(plugin.Hash)
		}
	}
}
