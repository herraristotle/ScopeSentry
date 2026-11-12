// Package models -----------------------------
// @file      : plugin.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/4/27 17:20
// -------------------------------------------
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Plugin struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Module        string             `json:"module"`
	Name          string             `json:"name"`
	Hash          string             `json:"hash"`
	Parameter     string             `json:"parameter"`
	Help          string             `json:"help"`
	Status        bool               `json:"status"`
	Type          string             `bson:"type" json:"type"`
	LastTime      string             `bson:"lastTime" json:"lastTime"`
	NextTime      string             `bson:"nextTime" json:"nextTime"`
	Introduction  string             `json:"introduction"`
	ParameterList string             `json:"parameterList"`
	IsSystem      bool               `json:"isSystem"`
	Version       string             `json:"version"`
	Source        string             `json:"source"`
}

// PluginListRequest plugin list request
type PluginListRequest struct {
	PageIndex int    `json:"pageIndex" binding:"required,min=1"`
	PageSize  int    `json:"pageSize" binding:"required,min=1"`
	Type      string `json:"type"`
	Search    string `json:"search"`
}

// PluginListResponse plugin list response
type PluginListResponse struct {
	List  []Plugin `json:"list"`
	Total int64    `json:"total"`
}

// PluginDetailRequest plugin detail request
type PluginDetailRequest struct {
	ID string `json:"id" binding:"required"`
}

// PluginSaveRequest plugin save request
type PluginSaveRequest struct {
	ID            string `json:"id"`
	Name          string `json:"name" binding:"required"`
	Module        string `json:"module"`
	Hash          string `json:"hash"`
	Parameter     string `json:"parameter"`
	Help          string `json:"help"`
	Introduction  string `json:"introduction"`
	ParameterList string `json:"parameterList"`
	Source        string `json:"source"`
	Version       string `json:"version"`
	IsSystem      bool   `json:"isSystem"`
	Type          string `json:"type"`
	Key           string `json:"key" binding:"required"`
}

// PluginDeleteRequest plugin deletion request
type PluginDeleteRequest struct {
	Data []struct {
		Hash   string `json:"hash" binding:"required"`
		Module string `json:"module" binding:"required"`
	} `json:"data" binding:"required"`
}

// PluginLogRequest plugin log request
type PluginLogRequest struct {
	Module string `json:"module"`
	Hash   string `json:"hash" binding:"required"`
	Type   string `json:"type"`
}

// PluginReinstallRequest plugin reinstall request
type PluginReinstallRequest struct {
	Node   string `json:"node"`
	Hash   string `json:"hash"`
	Module string `json:"module"`
}

// PluginRecheckRequest plugin recheck request
type PluginRecheckRequest struct {
	Node   string `json:"node" binding:"required"`
	Hash   string `json:"hash" binding:"required"`
	Module string `json:"module" binding:"required"`
}

// PluginUninstallRequest plugin uninstall request
type PluginUninstallRequest struct {
	Node   string `json:"node" binding:"required"`
	Hash   string `json:"hash" binding:"required"`
	Module string `json:"module" binding:"required"`
}

// PluginKeyCheckRequest plugin key check request
type PluginKeyCheckRequest struct {
	Key string `json:"key" binding:"required"`
}

// PluginImportByDataRequest plugin import-by-POST-JSON-data request
type PluginImportByDataRequest struct {
	JSON     string `json:"json" `    // Complete serialized JSON string of info.json
	Source   string `json:"source" `  // Complete source code content of plugin.go
	IsSystem bool   `json:"isSystem"` // Whether it is a built-in plugin
	Key      string `json:"key"`      // Plugin key
}

// PluginStatusRequest plugin status update request
type PluginStatusRequest struct {
	ID     string `json:"id"`     // Plugin ID
	Status bool   `json:"status"` // Plugin status
}

// PluginRunRequest plugin run request
type PluginRunRequest struct {
	Hash string `json:"hash" binding:"required"` // Plugin hash
}

type PluginInfo struct {
	Help          string `json:"help"`
	Parameter     string `json:"parameter"`
	Module        string `json:"module"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Introduction  string `json:"introduction"`
	Hash          string `json:"hash"`
	ParameterList string `json:"parameterList"`
	Type          string `json:"type"`
	// The following fields are added dynamically at runtime and are optional
	Source   string `json:"source,omitempty"`
	IsSystem bool   `json:"isSystem,omitempty"`
}

type NodePluginInfo struct {
	Name    string `json:"name"`
	Install string `json:"install"`
	Check   string `json:"check"`
	Hash    string `json:"hash"`
	Module  string `json:"module"`
}
