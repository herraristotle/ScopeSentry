package asset

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/repositories/assets/asset"
	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service defines the asset service interface
type Service interface {
	GetAssets(ctx *gin.Context, query models.SearchRequest) ([]models.Asset, error)
	GetAssetByIDs(ctx *gin.Context, ids []string) ([]models.Asset, error)
	GetAssetByID(ctx *gin.Context, id string) (*models.Asset, error)
	CreateAsset(ctx *gin.Context, asset *models.Asset) error
	UpdateAsset(ctx *gin.Context, id string, asset *models.Asset) error
	DeleteAsset(ctx *gin.Context, id string) error
	GetScreenshot(ctx *gin.Context, id string) (string, error)
	GetChangeLog(ctx *gin.Context, id string) ([]models.AssetChangeLog, error)
	DeduplicateAssets(ctx *gin.Context, assetType string, filter bson.M, groupFields []string) error
	GetAssetCardData(ctx *gin.Context, query models.SearchRequest) ([]models.Asset, error)
	GetTaskTarget(ctx context.Context, query models.SearchRequest) (string, error)
	GetTaskTargetByIDs(ctx context.Context, ids []string) (string, error)
	BuildSearchQuery(query models.SearchRequest) (map[string]interface{}, error)
	GetHostsByRootDomain(ctx *gin.Context, rootDomain string, pageIndex, pageSize int) ([]string, int64, error)
	GetAssetsByHost(ctx *gin.Context, host string, pageIndex, pageSize int) ([]models.Asset, int64, error)
}

// Global bodyhash cache instance, TTL is 30 seconds (customizable via the NewCache parameter)
var globalBodyHashCache = helper.NewCache(30 * time.Second)

type service struct {
	assetRepo asset.Repository
}

// NewService creates an asset service instance
func NewService() Service {
	return &service{
		assetRepo: asset.NewRepository(),
	}
}

func (s *service) ReplaceBodyQuery(ctx context.Context, query map[string]interface{}) map[string]interface{} {
	return s.replaceBodyRecursively(ctx, query).(map[string]interface{})
}

// convertToInterfaceSlice converts arrays of various types to []interface{}
func convertToInterfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}

	// Use reflection to handle various array types
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		result := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result
	}

	// If it is already []interface{}, return it directly
	if arr, ok := v.([]interface{}); ok {
		return arr
	}

	return nil
}

// replaceBodyRecursively is the internal recursive processing function
func (s *service) replaceBodyRecursively(ctx context.Context, query any) any {
	switch q := query.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for k, v := range q {
			// Recursively process logical operators
			if k == "$and" || k == "$or" || k == "$nor" {
				// Handle the various possible array types
				arr := convertToInterfaceSlice(v)
				if arr != nil {
					newArr := []interface{}{}
					for _, sub := range arr {
						newArr = append(newArr, s.replaceBodyRecursively(ctx, sub))
					}
					newMap[k] = newArr
					continue
				}
			}

			// ✅ Special handling for the body field
			if k == "body" {
				cond := s.convertBodyToBodyhash(ctx, v)
				// Merge the converted conditions into newMap
				// If cond is empty (a negated query with no match), add no condition, meaning no restriction
				for hk, hv := range cond {
					newMap[hk] = hv
				}
				continue
			}

			// Recurse for regular fields
			newMap[k] = s.replaceBodyRecursively(ctx, v)
		}
		return newMap

	case []interface{}:
		newArr := []interface{}{}
		for _, sub := range q {
			newArr = append(newArr, s.replaceBodyRecursively(ctx, sub))
		}
		return newArr

	default:
		return q
	}
}

// convertBodyToBodyhash converts body conditions into bodyhash queries (with global cache and != support)
func (s *service) convertBodyToBodyhash(ctx context.Context, v interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	m, ok := v.(map[string]interface{})
	if !ok {
		return result
	}

	var (
		regexVal   string
		isNegation bool
		cacheKey   string
		bodyHashes []string
	)

	// ✅ body = "xxx" or body: { "$regex": "xxx", "$options": "i" }
	if regex, ok := m["$regex"].(string); ok {
		regexVal = regex
		cacheKey = fmt.Sprintf("eq|%s", regexVal)
		isNegation = false
	} else if not, ok := m["$not"].(map[string]interface{}); ok {
		// ✅ body != "xxx" or body: { "$not": { "$regex": "xxx", "$options": "i" } }
		if regex, ok := not["$regex"].(string); ok {
			regexVal = regex
			isNegation = true
			cacheKey = fmt.Sprintf("ne|%s", regexVal)
		}
	}

	if regexVal == "" {
		// Unrecognized body query format; return an empty result
		return result
	}

	// Check the global cache first
	if hashes, ok := globalBodyHashCache.GetStringSlice(cacheKey); ok {
		bodyHashes = hashes
	} else {
		// Only query the database when not cached
		filter := bson.M{"content": bson.M{"$regex": regexVal, "$options": "i"}}
		opts := options.Find().SetProjection(bson.M{"hash": 1}).SetLimit(1000)

		bodies, err := s.assetRepo.FindBodyWithOptions(ctx, filter, opts)
		if err != nil {
			// On query error, return empty for negated queries (meaning no restriction) and a non-matching condition for positive queries
			if isNegation {
				return map[string]interface{}{}
			}
			// When a positive query fails, return a condition that can never match
			return map[string]interface{}{"bodyhash": bson.M{"$in": []string{}}}
		}

		for _, b := range bodies {
			if b.BodyHash != "" {
				bodyHashes = append(bodyHashes, b.BodyHash)
			}
		}

		// Store in the global cache
		globalBodyHashCache.SetStringSlice(cacheKey, bodyHashes)
	}

	// Build the returned condition
	if len(bodyHashes) == 0 {
		if isNegation {
			// No content matched; a negated query means bodyhash can be any value (including non-existent)
			// Return an empty map, adding no restrictive conditions
			return map[string]interface{}{}
		}
		// No positive content matched; return a condition that can never match (empty array $in)
		return map[string]interface{}{"bodyhash": bson.M{"$in": []string{}}}
	}

	if isNegation {
		result["bodyhash"] = bson.M{"$nin": bodyHashes}
	} else {
		result["bodyhash"] = bson.M{"$in": bodyHashes}
	}

	return result
}

func (s *service) BuildSearchQuery(query models.SearchRequest) (map[string]interface{}, error) {
	query.Index = "asset"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	return s.ReplaceBodyQuery(context.Background(), searchQuery), nil

}

func (s *service) GetAssetByIDs(ctx *gin.Context, ids []string) ([]models.Asset, error) {
	return s.assetRepo.GetAssetByIDs(ctx, ids)
}

// GetAssets gets the asset list
func (s *service) GetAssets(ctx *gin.Context, query models.SearchRequest) ([]models.Asset, error) {
	// Build query conditions
	searchQuery, err := s.BuildSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	// Paginated query
	skip := (query.PageIndex - 1) * query.PageSize
	limit := query.PageSize

	// Build the sort
	sort := bson.D{
		{
			Key: "time", Value: -1,
		},
	}

	// Build the projection
	projection := bson.M{
		"_id":          1,
		"host":         1,
		"url":          1,
		"ip":           1,
		"port":         1,
		"service":      1,
		"type":         1,
		"title":        1,
		"statuscode":   1,
		"rawheaders":   1,
		"technologies": 1,
		"metadata":     1,
		"time":         1,
		"faviconmmh3":  1,
		"tags":         1,
		"bodyhash":     1,
	}

	// Execute the query
	opts := options.Find().
		SetProjection(projection).
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(sort)

	return s.assetRepo.FindWithOptions(ctx.Request.Context(), filter, opts)
}

// GetAssetByID gets an asset by ID
func (s *service) GetAssetByID(ctx *gin.Context, id string) (*models.Asset, error) {
	assetByID, err := s.assetRepo.GetAssetByID(ctx.Request.Context(), id)
	if err != nil {
		return nil, err
	}
	httpBody, err := s.assetRepo.FindBodyByHash(context.Background(), assetByID.ResponseBodyHash)
	if err != nil {
		return nil, err
	}
	assetByID.Body = httpBody.Content
	return assetByID, nil
}

// CreateAsset creates an asset
func (s *service) CreateAsset(ctx *gin.Context, asset *models.Asset) error {
	return s.assetRepo.CreateAsset(ctx.Request.Context(), asset)
}

// UpdateAsset updates an asset
func (s *service) UpdateAsset(ctx *gin.Context, id string, asset *models.Asset) error {
	return s.assetRepo.UpdateAsset(ctx.Request.Context(), id, asset)
}

// DeleteAsset deletes an asset
func (s *service) DeleteAsset(ctx *gin.Context, id string) error {
	return s.assetRepo.DeleteAsset(ctx.Request.Context(), id)
}

// GetScreenshot gets the asset screenshot
func (s *service) GetScreenshot(ctx *gin.Context, id string) (string, error) {
	return s.assetRepo.GetScreenshot(ctx.Request.Context(), id)
}

// GetChangeLog gets the asset change log
func (s *service) GetChangeLog(ctx *gin.Context, id string) ([]models.AssetChangeLog, error) {
	return s.assetRepo.GetChangeLog(ctx.Request.Context(), id)
}

// DeduplicateAssets deduplicates assets
func (s *service) DeduplicateAssets(ctx *gin.Context, assetType string, filter bson.M, groupFields []string) error {
	return s.assetRepo.DeduplicateAssets(ctx.Request.Context(), assetType, filter, groupFields)
}

// GetAssetCardData gets asset card data
func (s *service) GetAssetCardData(ctx *gin.Context, query models.SearchRequest) ([]models.Asset, error) {
	// Build query conditions
	query.Index = "asset"
	searchQuery, err := s.BuildSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	// Paginated query
	skip := (query.PageIndex - 1) * query.PageSize
	limit := query.PageSize

	// Build the sort
	sort := bson.D{
		{
			Key: "time", Value: -1,
		},
	}

	// Build the projection
	projection := bson.M{
		"_id":        1,
		"host":       1,
		"url":        1,
		"port":       1,
		"service":    1,
		"type":       1,
		"title":      1,
		"statuscode": 1,
		"bodyhash":   1,
	}

	// Execute the query
	opts := options.Find().
		SetProjection(projection).
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(sort)

	assets, err := s.assetRepo.FindWithOptions(ctx.Request.Context(), filter, opts)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func (s *service) GetTaskTarget(ctx context.Context, query models.SearchRequest) (string, error) {
	query.Index = "asset"
	searchQuery, err := s.BuildSearchQuery(query)
	if err != nil {
		return "", err
	}
	filter := bson.M(searchQuery)
	projection := bson.M{
		"url":     1,
		"host":    1,
		"port":    1,
		"service": 1,
		"type":    1,
	}

	// Execute the query
	opts := options.Find().
		SetProjection(projection)

	if query.PageSize != 0 {
		opts = opts.SetLimit(int64(query.PageSize))
	}

	assets, err := s.assetRepo.FindWithOptions(ctx, filter, opts)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i, a := range assets {
		if a.Type == "http" {
			builder.WriteString(strings.TrimRightFunc(a.URL, unicode.IsSpace))
		} else {
			if a.Service != "" {
				builder.WriteString(a.Service + "://")
			} else {
				builder.WriteString("http" + "://")
			}
			builder.WriteString(a.Host + ":")
			builder.WriteString(a.Port)
		}
		if i != len(assets)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String(), nil
}

func (s *service) GetTaskTargetByIDs(ctx context.Context, ids []string) (string, error) {
	assets, err := s.assetRepo.GetAssetByIDs(ctx, ids)
	if err != nil {
		return "", err
	}
	targets := ""
	for _, asset := range assets {
		if asset.Type == "http" {
			targets += asset.URL + "\n"
		} else {
			sv := "http"
			if asset.Service != "" {
				sv = asset.Service
			}
			targets += sv + "://" + asset.Host + ":" + asset.Port + "\n"
		}
	}
	return targets, nil
}

// GetHostsByRootDomain queries by rootDomain and returns a deduplicated host list
func (s *service) GetHostsByRootDomain(ctx *gin.Context, rootDomain string, pageIndex, pageSize int) ([]string, int64, error) {
	// Set pagination parameters
	if pageSize <= 0 {
		pageSize = 10 // default size is 10
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	skip := int64((pageIndex - 1) * pageSize)
	limit := int64(pageSize)

	// Use an aggregation pipeline to deduplicate at the database level, returning the truly deduplicated data
	hosts, total, err := s.assetRepo.GetDistinctHostsByRootDomain(ctx.Request.Context(), rootDomain, skip, limit)
	if err != nil {
		return nil, 0, err
	}
	if hosts == nil {
		hosts = []string{}
	}

	return hosts, total, nil
}

// GetAssetsByHost queries the asset list by host
func (s *service) GetAssetsByHost(ctx *gin.Context, host string, pageIndex, pageSize int) ([]models.Asset, int64, error) {
	// Set pagination parameters
	if pageSize <= 0 {
		pageSize = 10 // default size is 10
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	skip := int64((pageIndex - 1) * pageSize)
	limit := int64(pageSize)

	// Query the asset list
	assets, total, err := s.assetRepo.GetAssetsByHost(ctx.Request.Context(), host, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	if assets == nil {
		assets = []models.Asset{}
	}
	return assets, total, nil
}
