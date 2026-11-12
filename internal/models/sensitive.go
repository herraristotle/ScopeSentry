package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SensitiveInfo represents a sensitive information record
type SensitiveChild struct {
	ID     primitive.ObjectID `bson:"_id" json:"id"`
	Match  []string           `bson:"match" json:"match"`
	Time   string             `bson:"time" json:"time"`
	Url    string             `bson:"url" json:"url"`
	Tags   []string           `bson:"tags" json:"tags"`
	Status int                `bson:"status" json:"status"`
	Name   string             `bson:"name" json:"name"`
	Color  string             `bson:"color" json:"color"`
}

type Sensitive struct {
	ID     primitive.ObjectID `bson:"_id" json:"id"`
	Match  []string           `bson:"match" json:"match"`
	Time   string             `bson:"time" json:"time"`
	Url    string             `bson:"url" json:"url"`
	Tags   []string           `bson:"tags" json:"tags"`
	Status int                `bson:"status" json:"status"`
	Name   string             `bson:"sid" json:"name"`
	Color  string             `bson:"color" json:"color"`
	BodyId string             `bson:"md5" json:"body_id"`
}

type SensitiveInfo struct {
	ID       string           `bson:"_id" json:"id"`
	Time     string           `bson:"time" json:"time"`
	Url      string           `bson:"url" json:"url"`
	BodyID   string           `bson:"body_id" json:"body_id"`
	Children []SensitiveChild `bson:"children" json:"children"`
}

// SensitiveRule represents a sensitive information detection rule
type SensitiveRule struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`               // Rule name
	Type        string             `bson:"type" json:"type"`               // Rule type
	Pattern     string             `bson:"pattern" json:"pattern"`         // Match pattern (regular expression)
	Description string             `bson:"description" json:"description"` // Rule description
	Level       string             `bson:"level" json:"level"`             // Sensitivity level (high/medium/low)
	State       string             `bson:"state" json:"state"`             // Rule state (enabled/disabled)
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`   // Created at
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`   // Updated at
}

// SensitiveRuleItem is used by frontend rule management (maps the Python SensitiveRule collection)
type SensitiveRuleItem struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `bson:"name" json:"name"`
	Regular string             `bson:"regular" json:"regular"`
	Color   string             `bson:"color" json:"color"`
	State   bool               `bson:"state" json:"state"`
}

type SensitiveRuleListItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Regular string `json:"regular"`
	Color   string `json:"color"`
	State   bool   `json:"state"`
}

type GetSensitiveBodyRequest struct {
	ID string `json:"id" binding:"required" example:"d41d8cd98f00b204e9800998ecf8427e"`
}

type GetSensitiveBodyResponse struct {
	Body string `json:"body" example:"<html>content</html>"`
}

type SensitiveSIDStat struct {
	Name  string `json:"name" bson:"name"`
	Count int    `json:"count" bson:"count"`
	Color string `json:"color" bson:"color"`
}

type GetSensitiveMatchInfoRequest struct {
	Sid string `json:"sid" binding:"required" example:"1234"`
}

type MatchListResponse struct {
	List []string `json:"list"`
}
