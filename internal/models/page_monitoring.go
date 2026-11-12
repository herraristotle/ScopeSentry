package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// PageMonitoring page monitoring model
type PageMonitoring struct {
	ID        string    `json:"id" bson:"_id"`
	URL       string    `json:"url" bson:"url"`
	Content   []string  `json:"content" bson:"content"`
	Hash      []string  `json:"hash" bson:"hash"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// PageMonitoringBody page monitoring content model
type PageMonitoringBody struct {
	ID      string   `json:"id" bson:"_id"`
	MD5     string   `json:"md5" bson:"md5"`
	Content []string `json:"diff" bson:"content"`
}

// PageMonitoringHistoryRequest history request
type PageMonitoringHistoryRequest struct {
	ID string `json:"id"`
}

// PageMonitoringContentRequest content request
type PageMonitoringContentRequest struct {
	ID   string `json:"id"`
	Flag string `json:"flag"` // 1: get the previous response, 2: get the current response
}

// PageMonitoringDiffRequest diff comparison request
type PageMonitoringDiffRequest struct {
	ID string `json:"id"`
}

// PageMonitoringResultRequest page monitoring result request
type PageMonitoringResultRequest struct {
	PageIndex int                    `json:"pageIndex" binding:"required,min=1" example:"1"`
	PageSize  int                    `json:"pageSize" binding:"required,min=1" example:"10"`
	Query     map[string]interface{} `json:"query" binding:"required" example:"{\"url\":\"example.com\"}"`
}

// PageMonitoringResult page monitoring result
type PageMonitoringResult struct {
	URL        string             `json:"url" bson:"url"`
	Hash       []string           `json:"hash" bson:"hash"`
	MD5        string             `json:"md5" bson:"md5"`
	Time       string             `json:"time" bson:"time"`
	StatusCode []int              `json:"statusCode" bson:"statusCode"`
	Similarity float64            `json:"similarity" bson:"similarity"`
	Tags       []string           `json:"tags" bson:"tags"`
	ID         primitive.ObjectID `bson:"_id" json:"id"`
}

// PageMonitoringResultResponse page monitoring result response
type PageMonitoringResultResponse struct {
	List  []PageMonitoringResult `json:"list"`
	Total int64                  `json:"total"`
}
