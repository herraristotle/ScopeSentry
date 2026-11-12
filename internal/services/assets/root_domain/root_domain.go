package root_domain

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/bootstrap"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/repositories/assets/root_domain"
	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

// Service defines the root domain service interface
type Service interface {
	GetRootDomainData(ctx *gin.Context, query models.SearchRequest) (*models.RootDomainResponse, error)
	GetTaskTarget(ctx context.Context, query models.SearchRequest) (string, error)
	GetTaskTargetByIDs(ctx context.Context, ids []string) (string, error)
}

type service struct {
	rootDomainRepo root_domain.Repository
}

// NewService creates a root domain service instance
func NewService() Service {
	return &service{
		rootDomainRepo: root_domain.NewRepository(),
	}
}

func (s *service) GetTaskTargetByIDs(ctx context.Context, ids []string) (string, error) {
	if len(ids) == 0 {
		return "", nil
	}

	// Convert the string ID to ObjectID
	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return "", err
		}
		objectIDs = append(objectIDs, objectID)
	}

	// Build query conditions
	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	projection := bson.M{
		"domain": 1,
	}

	// Execute the query
	opts := options.Find().
		SetProjection(projection)

	rootDomains, err := s.rootDomainRepo.FindWithPagination(ctx, filter, opts)
	if err != nil {
		return "", err
	}

	// Join hosts into a string
	var builder strings.Builder
	for i, subDomain := range rootDomains {
		builder.WriteString(subDomain.Domain)
		if i != len(rootDomains)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String(), nil
}

// GetRootDomainData gets root domain data
func (s *service) GetRootDomainData(ctx *gin.Context, query models.SearchRequest) (*models.RootDomainResponse, error) {
	query.Index = "RootDomain"
	// Build query conditions
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	// Get the total count
	totalCount, err := s.rootDomainRepo.CountDocuments(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	// Build the projection
	projection := bson.M{
		"_id":     1,
		"domain":  1,
		"icp":     1,
		"company": 1,
		"project": 1,
		"time":    1,
		"tags":    1,
	}

	// Build the query options
	opts := options.Find().
		SetProjection(projection).
		SetSkip(int64((query.PageIndex - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize)).
		SetSort(bson.D{{"time", -1}})

	// Get the paginated data
	rootDomains, err := s.rootDomainRepo.FindWithPagination(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}

	// Handle project name mapping
	for i := range rootDomains {
		if projectName, exists := bootstrap.ProjectList[rootDomains[i].Project]; exists {
			rootDomains[i].Project = projectName
		} else {
			rootDomains[i].Project = ""
		}
	}

	return &models.RootDomainResponse{
		List:  rootDomains,
		Total: totalCount,
	}, nil
}

func (s *service) GetTaskTarget(ctx context.Context, query models.SearchRequest) (string, error) {
	query.Index = "RootDomain"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return "", err
	}
	filter := bson.M(searchQuery)
	projection := bson.M{
		"domain": 1,
	}

	// Execute the query
	opts := options.Find().
		SetProjection(projection)

	if query.PageSize != 0 {
		opts = opts.SetLimit(int64(query.PageSize))
	}

	rootDomains, err := s.rootDomainRepo.FindWithPagination(ctx, filter, opts)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i, rootDomain := range rootDomains {
		builder.WriteString(rootDomain.Domain)
		if i != len(rootDomains)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String(), nil
}
