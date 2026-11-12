// plugins-------------------------------------
// @file      : context_manager.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/12/20 20:15
// -------------------------------------------

package utils

import (
	"context"
	"sync"
)

// contextInfo stores a context and its corresponding cancel function
type contextInfo struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	// pluginContexts global variable storing all plugin contexts
	pluginContexts = make(map[string]*contextInfo)
	// contextMutex mutex protecting pluginContexts
	contextMutex sync.RWMutex
)

// GetContext returns the context for a plugin ID, creating a new one if it does not exist
func GetContext(id string) context.Context {
	contextMutex.Lock()
	defer contextMutex.Unlock()

	// If it already exists, return it directly
	if info, exists := pluginContexts[id]; exists && info != nil {
		return info.ctx
	}

	// If it does not exist, create a new context
	ctx, cancel := context.WithCancel(context.Background())
	pluginContexts[id] = &contextInfo{
		ctx:    ctx,
		cancel: cancel,
	}

	return ctx
}

// CancelContext cancels and deletes the context for a plugin ID
func CancelContext(id string) {
	contextMutex.Lock()
	defer contextMutex.Unlock()

	// If it exists, cancel it first and then delete it
	if info, exists := pluginContexts[id]; exists && info != nil {
		info.cancel()
		delete(pluginContexts, id)
	}
}

// HasContext checks whether a plugin already has a context
func HasContext(id string) bool {
	contextMutex.RLock()
	defer contextMutex.RUnlock()

	_, exists := pluginContexts[id]
	return exists
}
