package root_domain

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository defines the root domain repository interface
type Repository interface {
	CountDocuments(ctx context.Context, query bson.M) (int64, error)
	FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.RootDomain, error)
}

type repository struct {
	collection *mongo.Collection
}

// NewRepository creates a root domain repository instance
func NewRepository() Repository {
	return &repository{
		collection: mongodb.DB.Collection("RootDomain"),
	}
}

// CountDocuments counts the number of documents
func (r *repository) CountDocuments(ctx context.Context, query bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, query)
}

// FindWithPagination queries documents with pagination
func (r *repository) FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.RootDomain, error) {
	// Execute the query
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode the results
	var results []models.RootDomain
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}
