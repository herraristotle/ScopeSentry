package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CrawlerTask represents a crawler task
type CrawlerResult struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Url        string             `bson:"url" json:"url"`
	Method     string             `bson:"method" json:"method"`
	Body       string             `bson:"body" json:"body"`
	Project    string             `bson:"project" json:"project"`
	ResBody    string             `bson:"-" json:"-"` // This field is not serialized
	TaskName   string             `bson:"taskName" json:"taskName"`
	ResultId   string             `bson:"resultId" json:"resultId"`
	RootDomain string             `bson:"rootDomain" json:"rootDomain"`
	Time       string             `bson:"time" json:"time"`
	Tags       []string           `bson:"tags" json:"tags"`
}
