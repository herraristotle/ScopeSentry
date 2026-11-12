// Package response -----------------------------
// @file      : data.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/3 12:02
// -------------------------------------------
package response

type DataResponse struct {
	List interface{} `json:"list"` // Use interface{} as a generic field type
}
