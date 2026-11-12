// plugin-------------------------------------
// @file      : utils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/12/20 18:42
// -------------------------------------------

package plugin

import (
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/plugins"
	"go.uber.org/zap"
)

func RegisterPlugin(source, hash string) {
	loadPlugin, err := plugins.LoadPlugin(source, hash)
	if err != nil {
		logger.Error("failed to load plugin", zap.Error(err))
		return
	}
	plugins.GlobalPluginManager.RegisterPlugin(hash, loadPlugin)
	err = loadPlugin.Install()
	if err != nil {
		logger.Error("failed to install plugin", zap.Error(err))
		return
	}
	cycle := loadPlugin.Cycle()
	// Scheduled task is not empty
	if cycle != "" {
		// If it is "1", run only once
		if cycle == "1" {
			return
		}
		err := AddJob(hash, loadPlugin.Cycle())
		if err != nil {
			return
		}
	}
}
