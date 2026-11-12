package models

// DeleteRequest deletion request model
type DeleteRequest struct {
	IDs           []string `json:"ids" binding:"required"`
	Index         string   `json:"index"`
	DeleateAssets bool     `json:"delA"`
}

// TagRequest tag request model
type TagRequest struct {
	Type string `json:"tp" binding:"required"`
	ID   string `json:"id" binding:"required"`
	Tag  string `json:"tag" binding:"required"`
}

// StatusRequest status update request model
type StatusRequest struct {
	Type   string `json:"tp" binding:"required"`
	ID     string `json:"id" binding:"required"`
	Status int    `json:"status" binding:"required"`
}

// TotalRequest total count request model
type TotalRequest struct {
	Index string `json:"index" binding:"required"`
}

// Document document model
type Document struct {
	ID     string   `bson:"_id,omitempty" json:"id,omitempty"`
	Tags   []string `bson:"tags,omitempty" json:"tags,omitempty"`
	Status bool     `bson:"status,omitempty" json:"status,omitempty"`
}

type IdRequest struct {
	ID string `json:"id" `
}
