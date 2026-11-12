package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MiniProgram represents a mini-program
type MiniProgram struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`               // Mini-program name
	AppID       string             `bson:"appid" json:"appid"`             // Mini-program AppID
	Type        string             `bson:"type" json:"type"`               // Mini-program type (WeChat/Alipay/Baidu, etc.)
	Version     string             `bson:"version" json:"version"`         // Version number
	Icon        string             `bson:"icon" json:"icon"`               // Mini-program icon URL
	Description string             `bson:"description" json:"description"` // Mini-program description
	Developer   string             `bson:"developer" json:"developer"`     // Developer
	Category    string             `bson:"category" json:"category"`       // Category
	Tags        []string           `bson:"tags" json:"tags"`               // Tags
	Status      string             `bson:"status" json:"status"`           // State (online/offline/under review, etc.)
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`   // Creation time
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`   // Update time
}
