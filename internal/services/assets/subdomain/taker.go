package subdomain

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/repositories/assets/subdomain"
	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TakerService interface {
	GetSubdomainTakerData(ctx *gin.Context, query models.SearchRequest) (*models.SubdomainTakerResponse, error)
}

type takerService struct {
	repo subdomain.TakerRepository
}

func NewTakerService() TakerService {
	return &takerService{
		repo: subdomain.NewTakerRepository(),
	}
}

func (s *takerService) GetSubdomainTakerData(ctx *gin.Context, query models.SearchRequest) (*models.SubdomainTakerResponse, error) {
	// Build query conditions
	query.Index = "SubdomainTakerResult"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to build search query: %w", err)
	}
	filter := bson.M(searchQuery)

	// Get the total count
	totalCount, err := s.repo.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	if totalCount == 0 {
		return &models.SubdomainTakerResponse{
			List:  []models.SubdomainTakerResult{},
			Total: 0,
		}, nil
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64((query.PageIndex - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize))

	// Query data
	results, err := s.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}

	return &models.SubdomainTakerResponse{
		List:  results,
		Total: totalCount,
	}, nil
}
