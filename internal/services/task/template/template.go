package template

import (
	"context"
	"fmt"

	"github.com/Autumn-27/ScopeSentry/internal/models"

	"github.com/Autumn-27/ScopeSentry/internal/repositories/task/template"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service defines the template service interface
type Service interface {
	List(ctx context.Context, pageIndex, pageSize int, query string) (*models.TemplateList, error)
	Detail(ctx *gin.Context, id string) (*models.ScanTemplate, error)
	Save(ctx context.Context, id string, result *models.ScanTemplate) (string, error)
	Delete(ctx *gin.Context, ids []string) error
}

type service struct {
	repo template.Repository
}

// NewService creates a template service instance
func NewService() Service {
	return &service{
		repo: template.NewRepository(),
	}
}

// List retrieves the template list
func (s *service) List(ctx context.Context, pageIndex, pageSize int, query string) (*models.TemplateList, error) {
	filter := bson.M{
		"name": bson.M{"$regex": query},
	}

	// Set query options
	opts := options.Find().
		SetProjection(bson.M{
			"_id":  1,
			"name": 1,
		}).
		SetSkip(int64((pageIndex - 1) * pageSize)).
		SetLimit(int64(pageSize)).SetSort(bson.M{"_id": -1})

	// Get total count
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count templates: %w", err)
	}

	// Get the list
	list, err := s.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find templates: %w", err)
	}

	return &models.TemplateList{
		List:  list,
		Total: total,
	}, nil
}

// Detail retrieves template details
func (s *service) Detail(ctx *gin.Context, id string) (*models.ScanTemplate, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %w", err)
	}

	filter := bson.M{"_id": objID}

	result, err := s.repo.FindOne(ctx, filter, options.FindOne())
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("template not found")
	}

	return result, nil
}

// Save saves a template
func (s *service) Save(ctx context.Context, id string, result *models.ScanTemplate) (string, error) {
	if id == "" {
		// Insert a new template
		insertId, err := s.repo.InsertOne(ctx, result)
		if err != nil {
			return "", fmt.Errorf("failed to insert template: %w", err)
		}
		return insertId, err
	}

	// Update the existing template
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", fmt.Errorf("invalid id format: %w", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": result}

	err = s.repo.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", fmt.Errorf("failed to update template: %w", err)
	}

	return "", nil
}

// Delete deletes templates
func (s *service) Delete(ctx *gin.Context, ids []string) error {
	var objIDs []primitive.ObjectID
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid id format: %w", err)
		}
		objIDs = append(objIDs, objID)
	}

	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	err := s.repo.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete templates: %w", err)
	}

	return nil
}
