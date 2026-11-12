package subdomain

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository defines the subdomain repository interface
type Repository interface {
	BuildSearchQuery(query map[string]interface{}) (bson.M, error)
	CountDocuments(ctx context.Context, query bson.M) (int64, error)
	FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.Subdomain, error)
}

type repository struct {
	collection *mongo.Collection
}

// NewRepository creates a subdomain repository instance
func NewRepository() Repository {
	return &repository{
		collection: mongodb.DB.Collection("subdomain"),
	}
}

// BuildSearchQuery builds search query conditions
func (r *repository) BuildSearchQuery(query map[string]interface{}) (bson.M, error) {
	if len(query) == 0 {
		return bson.M{"_id": bson.M{"$exists": true}}, nil
	}

	// Build query conditions
	searchQuery := bson.M{}
	for key, value := range query {
		searchQuery[key] = value
	}

	return searchQuery, nil
}

// CountDocuments counts the number of documents
func (r *repository) CountDocuments(ctx context.Context, query bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, query)
}

// FindWithPagination queries documents with pagination
func (r *repository) FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.Subdomain, error) {
	// Execute the query
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode the results
	var results []models.Subdomain
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}
