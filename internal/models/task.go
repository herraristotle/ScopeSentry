// Package models -----------------------------
// @file      : task.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/4 22:04
// -------------------------------------------
package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Task is the task model
type Task struct {
	ID             primitive.ObjectID       `bson:"_id,omitempty" json:"id"`
	TaskID         string                   `bson:"id"` // Associated task ID
	Name           string                   `bson:"name" json:"name"`
	Node           []string                 `bson:"node" json:"node"`
	AllNode        bool                     `bson:"allNode" json:"allNode"`
	ScheduledTasks bool                     `bson:"scheduledTasks" json:"scheduledTasks"`
	Target         string                   `bson:"target" json:"target"`
	Ignore         string                   `bson:"ignore" json:"ignore"`
	Template       string                   `bson:"template" json:"template"`
	Project        []string                 `bson:"project" json:"project"`
	Search         string                   `bson:"search" json:"search"`
	CycleType      string                   `bson:"cycleType" json:"cycleType"`
	Hour           int                      `bson:"hour" json:"hour"`
	Minute         int                      `bson:"minute" json:"minute"`
	Day            int                      `bson:"day" json:"day"`
	Week           int                      `bson:"week" json:"week"`
	Status         int                      `bson:"status" json:"status"`
	Progress       int                      `bson:"progress" json:"progress"`
	CreatTime      string                   `bson:"creatTime" json:"creatTime"`
	EndTime        string                   `bson:"endTime" json:"endTime"`
	TaskNum        int                      `bson:"taskNum" json:"taskNum"`
	Duplicates     string                   `bson:"duplicates" json:"duplicates"`
	Filter         map[string][]interface{} `bson:"filter" json:"filter"`
	TargetNumber   int                      `bson:"targetNumber" json:"targetNumber"`
	TargetIds      []string                 `bson:"targetIds" json:"targetIds"`
	TargetTp       string                   `bson:"targetTp" json:"targetTp"`
	TargetSource   string                   `bson:"targetSource" json:"targetSource"`
	BindProject    string                   `bson:"bindProject" json:"bindProject"`
	Type           string                   `bson:"type" json:"type"`
	IsStart        bool                     `bson:"isStart" json:"isStart"`
	LastTime       string                   `bson:"lastTime"  json:"lastTime"`
	NextTime       string                   `bson:"nextTime" json:"nextTime"`
	Cycle          string                   `bson:"-" json:"cycle"` // Cycle description (computed field)
}

// TaskProgress is the task progress model
type TaskProgress struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TargetHandler       int                `bson:"targetHandler" json:"targetHandler"`             // Target processing
	SubdomainScan       int                `bson:"subdomainScan" json:"subdomainScan"`             // Subdomain scanning
	SubdomainSecurity   int                `bson:"subdomainSecurity" json:"subdomainSecurity"`     // Subdomain security
	PortScanPreparation int                `bson:"portScanPreparation" json:"portScanPreparation"` // Port scan preparation
	PortScan            int                `bson:"portScan" json:"portScan"`                       // Port scanning
	PortFingerprint     int                `bson:"portFingerprint" json:"portFingerprint"`         // Port fingerprinting
	AssetMapping        int                `bson:"assetMapping" json:"assetMapping"`               // Asset mapping
	AssetHandle         int                `bson:"assetHandle" json:"assetHandle"`                 // Asset handling
	URLScan             int                `bson:"urlScan" json:"urlScan"`                         // URL scanning
	WebCrawler          int                `bson:"webCrawler" json:"webCrawler"`                   // Web crawler
	URLSecurity         int                `bson:"urlSecurity" json:"urlSecurity"`                 // URL security
	DirScan             int                `bson:"dirScan" json:"dirScan"`                         // Directory scanning
	VulnerabilityScan   int                `bson:"vulnerabilityScan" json:"vulnerabilityScan"`     // Vulnerability scanning
	All                 int                `bson:"all" json:"all"`                                 // Total progress
	Target              string             `bson:"target" json:"target"`                           // Target
	Node                string             `bson:"node" json:"node"`                               // Node
	Children            []TaskProgress     `bson:"children" json:"children"`                       // Subtask progress
}

// TaskProgressDetail is the task progress detail (used for API responses)
type TaskProgressDetail struct {
	ID                  primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	TargetHandler       []string             `json:"TargetHandler"`       // [start, end]
	SubdomainScan       []string             `json:"SubdomainScan"`       // [start, end]
	SubdomainSecurity   []string             `json:"SubdomainSecurity"`   // [start, end]
	PortScanPreparation []string             `json:"PortScanPreparation"` // [start, end]
	PortScan            []string             `json:"PortScan"`            // [start, end]
	PortFingerprint     []string             `json:"PortFingerprint"`     // [start, end]
	AssetMapping        []string             `json:"AssetMapping"`        // [start, end]
	AssetHandle         []string             `json:"AssetHandle"`         // [start, end]
	URLScan             []string             `json:"URLScan"`             // [start, end]
	WebCrawler          []string             `json:"WebCrawler"`          // [start, end]
	URLSecurity         []string             `json:"URLSecurity"`         // [start, end]
	DirScan             []string             `json:"DirScan"`             // [start, end]
	VulnerabilityScan   []string             `json:"VulnerabilityScan"`   // [start, end]
	All                 []string             `json:"All"`                 // [start, end]
	Target              string               `json:"target"`
	Node                string               `json:"node"`
	Children            []TaskProgressDetail `json:"children,omitempty"`
}

// ListRequest is the task list paginated query request
type ListRequest struct {
	Search    string `json:"search" binding:"omitempty" example:"test"`
	PageIndex int    `json:"pageIndex" binding:"required,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"required,min=1" example:"10"`
}

// ScheduledTask is the scheduled task model
type ScheduledTask struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	TaskID         string             `bson:"id" json:"id"`                         // Associated task ID
	Name           string             `bson:"name" json:"name"`                     // Task name
	Type           string             `bson:"type" json:"type"`                     // Task type
	LastTime       string             `bson:"lastTime" json:"lastTime"`             // Last execution time
	NextTime       string             `bson:"nextTime" json:"nextTime"`             // Next execution time
	State          bool               `bson:"scheduledTasks" json:"state"`          // Whether enabled
	Node           []string           `bson:"node" json:"node"`                     // Node list
	Cycle          string             `bson:"-" json:"cycle"`                       // Cycle description (computed field)
	AllNode        bool               `bson:"allNode" json:"allNode"`               // Whether all nodes
	RunnerID       string             `bson:"runner_id" json:"runner_id"`           // Runner ID
	Project        []string           `bson:"project" json:"project"`               // Project list
	TargetSource   string             `bson:"targetSource" json:"targetSource"`     // Target source
	Day            int                `bson:"day" json:"day"`                       // Days
	Minute         int                `bson:"minute" json:"minute"`                 // Minutes
	Hour           int                `bson:"hour" json:"hour"`                     // Hours
	Search         string             `bson:"search" json:"search"`                 // Search condition
	CycleType      string             `bson:"cycleType" json:"cycleType"`           // Cycle type
	ScheduledTasks bool               `bson:"scheduledTasks" json:"scheduledTasks"` // Whether it is a scheduled task
	Target         string             `bson:"target" json:"target"`                 // Target
	Ignore         string             `bson:"ignore" json:"ignore"`                 // Ignore rules
	Template       string             `bson:"template" json:"template"`             // Template
	Duplicates     string             `bson:"duplicates" json:"duplicates"`         // Deduplication rule
	Week           int                `bson:"week" json:"week"`                     // Weekday
}

// ScheduledTaskListRequest is the scheduled task list request
type ScheduledTaskListRequest struct {
	Search    string `json:"search" binding:"omitempty" example:"test"`
	PageIndex int    `json:"pageIndex" binding:"omitempty,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"omitempty,min=1" example:"10"`
}

// ScheduledTaskDetailRequest is the scheduled task detail request
type ScheduledTaskDetailRequest struct {
	ID string `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
}

// ScheduledTaskDeleteRequest is the scheduled task delete request
type ScheduledTaskDeleteRequest struct {
	IDs []string `json:"ids" binding:"required" example:"[\"507f1f77bcf86cd799439011\"]"`
}

// ScheduledTaskAddRequest is the scheduled task add request
type ScheduledTaskAddRequest struct {
	Name           string   `json:"name" binding:"required" example:"scheduled scan task"`
	Type           string   `json:"type" binding:"omitempty" example:"scan"`
	Node           []string `json:"node" binding:"required" example:"[\"node1\", \"node2\"]"`
	AllNode        bool     `json:"allNode" binding:"omitempty" example:"true"`
	Project        []string `json:"project" binding:"omitempty" example:"[\"project1\"]"`
	TargetSource   string   `json:"targetSource" binding:"omitempty" example:"general"`
	Day            int      `json:"day" binding:"omitempty" example:"1"`
	Minute         int      `json:"minute" binding:"omitempty" example:"0"`
	Hour           int      `json:"hour" binding:"omitempty" example:"1"`
	Search         string   `json:"search" binding:"omitempty" example:""`
	CycleType      string   `json:"cycleType" binding:"omitempty" example:"nhours"`
	ScheduledTasks bool     `json:"scheduledTasks" binding:"omitempty" example:"true"`
	Target         string   `json:"target" binding:"omitempty" example:"example.com"`
	Ignore         string   `json:"ignore" binding:"omitempty" example:""`
	Template       string   `json:"template" binding:"omitempty" example:"default"`
	Duplicates     string   `json:"duplicates" binding:"omitempty" example:"true"`
	Week           int      `json:"week" binding:"omitempty" example:"1"`
}

// ScheduledTaskUpdateRequest is the scheduled task update request
type ScheduledTaskUpdateRequest struct {
	ID             string   `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
	Name           string   `json:"name" binding:"omitempty" example:"scheduled scan task"`
	Type           string   `json:"type" binding:"omitempty" example:"scan"`
	Node           []string `json:"node" binding:"omitempty" example:"[\"node1\", \"node2\"]"`
	AllNode        bool     `json:"allNode" binding:"omitempty" example:"true"`
	Project        []string `json:"project" binding:"omitempty" example:"[\"project1\"]"`
	TargetSource   string   `json:"targetSource" binding:"omitempty" example:"general"`
	Day            int      `json:"day" binding:"omitempty" example:"1"`
	Minute         int      `json:"minute" binding:"omitempty" example:"0"`
	Hour           int      `json:"hour" binding:"omitempty" example:"1"`
	Search         string   `json:"search" binding:"omitempty" example:""`
	CycleType      string   `json:"cycleType" binding:"omitempty" example:"nhours"`
	ScheduledTasks bool     `json:"scheduledTasks" binding:"omitempty" example:"true"`
	Target         string   `json:"target" binding:"omitempty" example:"example.com"`
	Ignore         string   `json:"ignore" binding:"omitempty" example:""`
	Template       string   `json:"template" binding:"omitempty" example:"default"`
	Duplicates     string   `json:"duplicates" binding:"omitempty" example:"true"`
	Week           int      `json:"week" binding:"omitempty" example:"1"`
}

// PageMonitoringTask is the page monitoring task model
type PageMonitoringTask struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	URL     string             `bson:"url" json:"url"`
	Hash    []string           `bson:"hash" json:"hash"`
	MD5     string             `bson:"md5" json:"md5"`
	State   int                `bson:"state" json:"state"`
	Project string             `bson:"project" json:"project"`
	Time    string             `bson:"time" json:"time"`
}

// PageMonitoringListRequest is the page monitoring list request
type PageMonitoringListRequest struct {
	Search    string `json:"search" binding:"omitempty" example:"example.com"`
	PageIndex int    `json:"pageIndex" binding:"omitempty,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"omitempty,min=1" example:"10"`
}

// PageMonitoringAddRequest is the page monitoring add request
type PageMonitoringAddRequest struct {
	URL string `json:"url" binding:"required" example:"https://example.com"`
}

// PageMonitoringDeleteRequest is the page monitoring delete request
type PageMonitoringDeleteRequest struct {
	IDs []string `json:"ids" binding:"required" example:"[\"507f1f77bcf86cd799439011\"]"`
}

// PageMonitoringUpdateRequest is the page monitoring update request
type PageMonitoringUpdateRequest struct {
	Hour    int      `json:"hour" binding:"omitempty" example:"24"`
	Node    []string `json:"node" binding:"omitempty" example:"[\"node1\"]"`
	AllNode bool     `json:"allNode" binding:"omitempty" example:"true"`
	State   bool     `json:"state" binding:"omitempty" example:"true"`
}
