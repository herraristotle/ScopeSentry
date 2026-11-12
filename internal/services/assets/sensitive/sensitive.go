package sensitive

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/repositories/assets/sensitive"
	"go.mongodb.org/mongo-driver/bson"
)

type Service interface {
	GetSensitiveInfo(ctx context.Context, query models.SearchRequest) ([]models.Sensitive, error)
	GetSensitiveInfoNumber(ctx context.Context, query models.SearchRequest) (int64, int64, error)
	GetBodyByID(ctx *gin.Context, id string) (string, error)
	GetSIDStatistics(ctx *gin.Context, query models.SearchRequest) ([]models.SensitiveSIDStat, error)
	GetMatchInfo(ctx *gin.Context, req models.SearchRequest) (*models.MatchListResponse, error)
}

var ErrContentNotFound = errors.New("content not found")

type service struct {
	repo sensitive.Repository
}

func NewService() Service {
	return &service{
		repo: sensitive.NewRepository(),
	}
}

func (s *service) GetSensitiveInfo(ctx context.Context, query models.SearchRequest) ([]models.Sensitive, error) {
	query.Index = "SensitiveResult"
	// Build query conditions
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)
	projection := bson.M{
		"_id":    1,
		"url":    1,
		"time":   1,
		"sid":    1,
		"match":  1,
		"color":  1,
		"md5":    1,
		"tags":   1,
		"status": 1,
	}
	opts := options.Find().
		SetProjection(projection).
		SetSkip(int64((query.PageIndex - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize)).
		SetSort(bson.D{{"time", -1}})

	result, err := s.repo.FindWithPagination(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//func (s *service) GetSensitiveInfo(ctx context.Context, query models.SearchRequest) ([]models.SensitiveInfo, error) {
//	query.Index = "SensitiveResult"
//	// Build query conditions
//	searchQuery, err := helper.GetSearchQuery(query)
//	if err != nil {
//		return nil, err
//	}
//	filter := bson.M(searchQuery)
//
//	pipeline := buildAggregatePipeline(filter, query.PageIndex, query.PageSize)
//	result, err := s.repo.Aggregate(ctx, pipeline)
//	if err != nil {
//		return nil, err
//	}
//	return result, nil
//}

func buildAggregatePipeline(query bson.M, pageIndex, pageSize int) mongo.Pipeline {
	skip := (pageIndex - 1) * pageSize

	pipeline := mongo.Pipeline{
		// 1. Match search conditions (filter data as early as possible)
		{{"$match", query}},

		// 2. Exact projection: take only the fields needed for the later group stage, reducing memory pressure
		{{"$project", bson.M{
			"url":    1,
			"time":   1,
			"sid":    1,
			"match":  1,
			"color":  1,
			"md5":    1,
			"tags":   1,
			"status": 1,
		}}},

		// 3. Sort by time (or another field)
		{{"$sort", bson.D{{"time", -1}}}}, // Avoid "$_id" as the default sort field

		// 4. Group: aggregate all records belonging to each URL
		{{"$group", bson.D{
			{"_id", "$url"},
			{"time", bson.D{{"$first", "$time"}}}, // Latest time
			{"url", bson.D{{"$first", "$url"}}},
			{"body_id", bson.D{{"$last", bson.D{{"$toString", "$md5"}}}}},
			{"children", bson.D{{"$push", bson.M{
				"id":     bson.D{{"$toString", "$_id"}},
				"name":   "$sid",
				"color":  "$color",
				"match":  "$match",
				"time":   "$time",
				"tags":   "$tags",
				"status": "$status",
			}}}},
		}}},

		// 5. Sort again (by each group's time descending)
		{{"$sort", bson.D{{"time", -1}}}},

		// 6. Paginate (note: a large skip degrades performance; consider cursor-based pagination using last_id)
		{{"$skip", skip}},
		{{"$limit", pageSize}},

		// 7. Final output fields
		{{"$project", bson.M{
			"id":       "$_id",
			"url":      1,
			"time":     1,
			"body_id":  1,
			"children": 1,
		}}},
	}

	return pipeline
}

func (s *service) GetSensitiveInfoNumber(ctx context.Context, query models.SearchRequest) (int64, int64, error) {
	query.Index = "SensitiveResult"
	// Build query conditions
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return 0, 0, err
	}
	filter := bson.M(searchQuery)
	totalCount, err := s.repo.CountDocuments(ctx, filter)
	if err != nil {
		return 0, 0, err
	}
	return totalCount, totalCount, err
	//pipeline := mongo.Pipeline{
	//	{{"$match", filter}},
	//	{{"$group", bson.D{{"_id", "$url"}}}},
	//	{{"$count", "total"}},
	//}
	//rawResults, err := s.repo.AggregateRaw(ctx, pipeline)
	//if err != nil {
	//	return totalCount, 0, err
	//}
	//var urlTotal int64
	//// Parse results
	//if len(rawResults) > 0 {
	//	switch val := rawResults[0]["total"].(type) {
	//	case int32:
	//		urlTotal = int64(val)
	//	case int64:
	//		urlTotal = val
	//	default:
	//		return totalCount, 0, fmt.Errorf("unexpected type for count: %T", val)
	//	}
	//}
	//return totalCount, urlTotal, nil
}

func (s *service) GetBodyByID(ctx *gin.Context, id string) (string, error) {
	content, err := s.repo.FindBodyByMD5(ctx, id)
	if err != nil {
		return "", err
	}
	if content == "" {
		return "", ErrContentNotFound
	}
	return content, nil
}

func (s *service) GetSIDStatistics(ctx *gin.Context, query models.SearchRequest) ([]models.SensitiveSIDStat, error) {
	query.Index = "SensitiveResult"
	if query.Filter != nil {
		if _, ok := query.Filter["sname"]; ok {
			delete(query.Filter, "sname")
		}
	}
	// Build query conditions
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$sid",
			"count": bson.M{"$sum": 1},
			"color": bson.M{"$first": "$color"},
		}}},
		{{Key: "$sort", Value: bson.M{
			"count": -1, // Sort by count first
			"_id":   1,  // when count is equal, sort by _id to ensure a stable order
		}}},
		{{Key: "$project", Value: bson.M{
			"name":  "$_id",
			"count": 1,
			"color": 1,
			"_id":   0,
		}}},
	}

	result, err := s.repo.AggregateSIDStat(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []models.SensitiveSIDStat{}, nil
	}

	return result, nil
}

func (s *service) GetMatchInfo(ctx *gin.Context, query models.SearchRequest) (*models.MatchListResponse, error) {
	query.Index = "SensitiveResult"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)
	filter["sid"] = query.Sid

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$project", Value: bson.M{"match": 1}}},
		{{Key: "$unwind", Value: "$match"}},
		{
			{Key: "$group", Value: bson.M{
				"_id":            nil,
				"unique_matches": bson.M{"$addToSet": "$match"},
			}},
		},
	}

	matches, err := s.repo.AggregateMatchInfo(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	return &models.MatchListResponse{List: matches}, nil
}
