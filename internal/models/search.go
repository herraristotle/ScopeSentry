// models-------------------------------------
// @file      : search.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/7 19:27
// -------------------------------------------

package models

// SearchRequest defines the parameter structure of a search request
type SearchRequest struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	// Index search type, corresponding to the key in SearchKey
	Index            string `json:"index"`
	SearchExpression string `json:"search"`
	// Filter filter conditions
	Filter map[string][]interface{} `json:"filter"`
	// FuzzyQuery fuzzy query conditions
	FuzzyQuery map[string]string `json:"-"`
	Sort       map[string]string `json:"sort"`
	Sid        string            `json:"sid"`
}
