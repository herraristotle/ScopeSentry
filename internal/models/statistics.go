// models-------------------------------------
// @file      : STATISTICS.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/9 21:16
// -------------------------------------------

package models

// StatisticsData asset statistics data
type StatisticsData struct {
	AssetCount         int64 `json:"assetCount"`
	SubdomainCount     int64 `json:"subdomainCount"`
	SensitiveCount     int64 `json:"sensitiveCount"`
	UrlCount           int64 `json:"urlCount"`
	VulnerabilityCount int64 `json:"vulnerabilityCount"`
}

// PortStatisticsRequest port statistics request
type PortStatisticsRequest struct {
	Filter map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
}

// PortStatisticsResponse port statistics response
type PortStatisticsResponse struct {
	Port []PortItem `json:"Port"`
}

type PortItem struct {
	Value  string `bson:"_id" json:"value" example:"80"`
	Number int64  `bson:"number" json:"number" example:"100"`
}

// TitleStatisticsRequest title statistics request
type TitleStatisticsRequest struct {
	Filter map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
}

// TitleStatisticsResponse title statistics response
type TitleStatisticsResponse struct {
	Title []TitleItem `json:"Title"`
}

type TitleItem struct {
	Value  string `json:"value" example:"Example Title"`
	Number int64  `json:"number" example:"50"`
}

// TypeStatisticsRequest service type statistics request
type TypeStatisticsRequest struct {
	Filter map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
}

// TypeStatisticsResponse service type statistics response
type TypeStatisticsResponse struct {
	Service []TypeItem `json:"Service"`
}

type TypeItem struct {
	Value  string `bson:"_id" json:"value" example:"nginx"`
	Number int64  `bson:"number" json:"number" example:"200"`
}

// IconStatisticsRequest icon statistics request
type IconStatisticsRequest struct {
	Filter   map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
	Page     int                    `json:"page" binding:"required" example:"1"`
	PageSize int                    `json:"page_size" binding:"required" example:"10"`
}

// IconStatisticsResponse icon statistics response
type IconStatisticsResponse struct {
	Icon []IconItem `json:"Icon"`
}

type IconItem struct {
	Value    string `bson:"iconcontent" json:"value" example:"base64_icon_content"`
	Number   int64  `bson:"number" json:"number" example:"30"`
	IconHash string `bson:"_id" json:"icon_hash" example:"abc123"`
}

// AppStatisticsRequest application statistics request
type AppStatisticsRequest struct {
	Filter map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
}

// AppStatisticsResponse application statistics response
type AppStatisticsResponse struct {
	Product []AppItem `json:"Product"`
}

type AppItem struct {
	Value  string `bson:"_id" json:"value" example:"WordPress"`
	Number int64  `bson:"number" json:"number" example:"150"`
}

// APP application mapping table
var APP = map[string]string{
	// Application mapping relations need to be added here
}
