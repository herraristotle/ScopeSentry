package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Subdomain represents a subdomain
type Subdomain struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Host       string             `bson:"host" json:"host"`             // Host name
	Type       string             `bson:"type" json:"type"`             // Type
	Value      []string           `bson:"value" json:"value"`           // Value list
	IP         []string           `bson:"ip" json:"ip"`                 // IP address list
	Time       string             `bson:"time" json:"time"`             // Time
	Tags       []string           `bson:"tags" json:"tags"`             // Tags
	Project    string             `bson:"project" json:"project"`       // Project
	TaskName   string             `bson:"taskName" json:"taskName"`     // Task name
	RootDomain string             `bson:"rootDomain" json:"rootDomain"` // Root domain
}

// SubdomainResponse subdomain response data structure
type SubdomainResponse struct {
	List  []Subdomain `json:"list"`
	Total int64       `json:"total"`
}

type SubdomainTakerResult struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Input    string             `bson:"input" json:"host"`
	Value    string             `bson:"value" json:"value"`
	Cname    string             `bson:"cname" json:"type"`
	Response string             `bson:"response" json:"response"`
	Tags     []string           `bson:"tags" json:"tags"`
}

type SubdomainTakerResponse struct {
	List  []SubdomainTakerResult `json:"list"`
	Total int64                  `json:"total"`
}
