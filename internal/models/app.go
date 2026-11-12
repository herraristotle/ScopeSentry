package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// App represents an application
type App struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`                 // Application name
	PackageName string             `bson:"package_name" json:"package_name"` // Package name
	Version     string             `bson:"version" json:"version"`           // Version number
	Platform    string             `bson:"platform" json:"platform"`         // Platform (android/ios)
	Size        int64              `bson:"size" json:"size"`                 // Application size (bytes)
	Icon        string             `bson:"icon" json:"icon"`                 // Application icon URL
	Description string             `bson:"description" json:"description"`   // Application description
	Developer   string             `bson:"developer" json:"developer"`       // Developer
	Category    string             `bson:"category" json:"category"`         // Application category
	Tags        []string           `bson:"tags" json:"tags"`                 // Tags
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`     // Creation time
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`     // Update time
}

// AppResult application result
type AppResult struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Icp         string             `bson:"icp" json:"icp"`
	Company     string             `bson:"company" json:"company"`
	Project     string             `bson:"project" json:"project"`
	Time        string             `bson:"time" json:"time"`
	Tags        []string           `bson:"tags" json:"tags"`
	Category    string             `bson:"category" json:"category"`
	Description string             `bson:"description" json:"description"`
	BundleID    string             `bson:"bundleID" json:"bundleID"`
	Apk         string             `bson:"apk" json:"apk"`
	URL         string             `bson:"url" json:"url"`
}

// AppResponse application response
type AppResponse struct {
	List  []AppResult `json:"list"`
	Total int64       `json:"total"`
}
