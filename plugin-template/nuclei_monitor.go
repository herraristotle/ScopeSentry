// nuclei_monitor-------------------------------------
// @file      : nuclei_monitor.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/12/20 20:04
// -------------------------------------------

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/options"
	"github.com/Autumn-27/ScopeSentry/internal/utils"
	"github.com/Autumn-27/ScopeSentry/internal/utils/random"
)

// API response structures
type templateResult struct {
	TemplateID string `json:"template_id"`
	Severity   string `json:"severity"`
	Name       string `json:"name"`
	Raw        string `json:"raw"`
}

type apiResponse struct {
	Results []templateResult `json:"results"`
}

func GetName() string {
	return "Nuclei Monitor"
}

// formatSeverityWithColor formats Severity by risk level, adding color markers
func formatSeverityWithColor(severity string) string {
	severityUpper := strings.ToUpper(severity)
	var color string
	var bold bool

	switch severityUpper {
	case "CRITICAL":
		color = "#ff6b6b" // red
		bold = true
	case "HIGH":
		color = "#ff8787" // light red
		bold = true
	case "MEDIUM":
		color = "#ffd43b" // yellow
		bold = false
	case "LOW":
		color = "#51cf66" // green
		bold = false
	case "INFO":
		color = "#74c0fc" // blue
		bold = false
	default:
		// No color by default
		return severity
	}

	if bold {
		return fmt.Sprintf(`<color value="%s" bold>%s</color>`, color, severity)
	}
	return fmt.Sprintf(`<color value="%s">%s</color>`, color, severity)
}

// formatAssetCountWithColor formats by affected asset count; if greater than 0, marks it red
func formatAssetCountWithColor(count int64) string {
	if count > 0 {
		return fmt.Sprintf(`<color value="#ff6b6b" bold>%v</color>`, count)
	}
	return fmt.Sprintf("%v", count)
}

// Cycle run cycle
func Cycle() string {
	return "0 * * * *"
}

func Install() error {
	return nil
}

// CreatTask whether to create tasks automatically
// true: create tasks automatically (warning: high risk)
// false: create tasks manually (recommended)
//
// Important security warnings:
// 1. Some PoCs may include dangerous operations such as data deletion or system damage
// 2. Automatic execution may cause irreversible impact on production systems
// 3. It is recommended to always set this to false and create tasks manually after human review
//
// Security recommendations:
// - Production environments: always set to false
// - Test environments: enable only after careful evaluation
// - New PoC testing: must be verified manually before execution
var CreatTask = false

// SkipScanName scanning of a PoC is skipped when its name contains any string in this list
var SkipScanName = []string{
	"WordPress",
}

func Execute(op options.PluginOption) error {
	value := op.GetStringVariable("nuclei_init")
	if value != "true" {
		// Initialization: import the PoC zip directly
		tempDir := os.TempDir()
		fileName := random.GenerateString(5) + ".zip"
		filePath := filepath.Join(tempDir, fileName)
		op.Log("download ....https://codeload.github.com/projectdiscovery/nuclei-templates")
		err := utils.DownloadFile("https://codeload.github.com/projectdiscovery/nuclei-templates/zip/refs/heads/main", filePath, 10*time.Minute)
		if err != nil {
			op.Log(fmt.Sprintf("download fail: %s", err.Error()), "e")
			return err
		}
		_, err = op.PocService.ImportPoc(context.Background(), filePath)
		if err != nil {
			op.Log(fmt.Sprintf("ImportPoc fail: %s", err.Error()), "e")
			return err
		}
		op.SetStringVariable("nuclei_init", "true")
		op.Log("download nuclei-templates success")
		return nil
	}
	// Monitor PoCs
	ctx := context.Background()
	offset := 0
	limit := 20

	for {
		results, err := fetchPocTemplates(offset, limit)
		if err != nil {
			op.Log(fmt.Sprintf("Failed to fetch POC templates: %s", err.Error()), "e")
			return err
		}

		// If there are no results, exit the loop
		if len(results) == 0 {
			break
		}

		// Iterate over the results
		for _, result := range results {
			result.TemplateID = strings.Replace(result.TemplateID, "-draft", "", 1)
			// Check whether the template_id already exists
			exists, err := op.PocService.TemplateIdExists(ctx, result.TemplateID)
			if err != nil {
				op.Log(fmt.Sprintf("Failed to check TemplateId: %s, error: %s", result.TemplateID, err.Error()), "e")
				continue
			}

			if exists {
				// If it exists, all PoCs have been updated; stop iterating
				op.Log("All PoCs have been updated")
				return nil
			}

			// Insert the PoC to get its poc ID
			pocID, err := op.PocService.AddPoc(context.Background(), &models.PocAddRequest{Content: result.Raw})
			if err != nil {
				op.Log(fmt.Sprintf("AddPoc error: %s", err.Error()), "e")
				return err
			}
			// Give the scanning side time to update the PoC
			time.Sleep(3 * time.Second)

			var pt models.PocTemplate
			err = utils.Unmarshal([]byte(result.Raw), &pt)
			if err != nil {
				return fmt.Errorf("failed to unmarshal poc template: %w", err)
			}
			var searchValueOld interface{}
			var customQuery string
			var scanStatus = "Not created"
			var affectedAssetCount = int64(0)
			if pt.Info.Metadata != nil {
				// First try parsing from fofa-query
				if fofaQueryValue := pt.Info.Metadata["fofa-query"]; fofaQueryValue != nil {
					searchValueOld = fofaQueryValue
					customQuery = helper.ConvertFofaToCustomQuery(fofaQueryValue)
					//op.Log(fmt.Sprintf("Fofa:%v converted search expression: %s", fofaQueryValue, customQuery))
				}
				// If the fofa result is empty, try parsing from shodan-query
				if customQuery == "" {
					if shodanQueryValue := pt.Info.Metadata["shodan-query"]; shodanQueryValue != nil {
						searchValueOld = shodanQueryValue
						customQuery = helper.ConvertShodanToCustomQuery(shodanQueryValue)
						//op.Log(fmt.Sprintf("Shodan:%v converted search expression: %s", shodanQueryValue, customQuery))
					}
				}
				if customQuery != "" {
					// Get the affected count
					searchRequest := models.SearchRequest{
						Index:            "asset",
						SearchExpression: customQuery,
					}
					affectedAssetCount, err = op.AssetCommonService.TotalData(context.Background(), &searchRequest)
					if err != nil {
						return err
					}
					// If there is a search condition and automatic task creation is enabled, create the task
					if CreatTask && affectedAssetCount != 0 {

						// Check whether the PoC is in the whitelist
						skipFlag := false
						for _, v := range SkipScanName {
							if strings.Contains(result.Name, strings.ToLower(v)) {
								skipFlag = true
								scanStatus = "Skipped"
								break
							}
						}
						// Create the scan template
						template := models.ScanTemplate{
							Name:              fmt.Sprintf("nuclei-template-%v-%v", result.Name, helper.GetNowTimeString()),
							VulnerabilityScan: []string{"ed93b8af6b72fe54a60efdb932cf6fbc"}, //nuclei
							VulList:           []string{pocID},
							Parameters: models.Parameters{
								VulnerabilityScan: map[string]string{
									"ed93b8af6b72fe54a60efdb932cf6fbc": "",
								},
							},
						}
						templateID, err := op.TemplateService.Save(context.Background(), "", &template)
						if err != nil {
							op.Log(fmt.Sprintf("Failed to save TemplateId: %s, error: %s", pocID, err.Error()), "e")
							return err
						}

						// Create the task
						if !skipFlag {
							task := models.Task{
								Name:           fmt.Sprintf("[nuclei monitor ]-%v-%v[plugin]", result.Name, helper.GetNowTimeString()),
								AllNode:        true,
								Duplicates:     "None",
								ScheduledTasks: false,
								Template:       templateID,
								TargetSource:   "asset",
								Search:         customQuery,
							}
							_, err := op.TaskCommonService.Insert(context.Background(), &task)
							if err != nil {
								return err
							}
							scanStatus = "Created successfully"
						}
					}
				}
			}
			severityFormatted := formatSeverityWithColor(result.Severity)
			assetCountFormatted := formatAssetCountWithColor(affectedAssetCount)
			op.Log(fmt.Sprintf("New POC found: TemplateID=%s, Name=%s, Severity=%s, OldSearch=%v, SaSSearch=%v, affected assets: %s, task creation status: %v", result.TemplateID, result.Name, severityFormatted, searchValueOld, customQuery, assetCountFormatted, scanStatus))
			notificationMsg := fmt.Sprintf("New POC found: \nTemplateID:%s\nName:%s\nSeverity:%s\nOldSearch:%v\nSaSSearch:%v\nAffected assets: %v\nTask creation status: %v", result.TemplateID, result.Name, result.Severity, searchValueOld, customQuery, affectedAssetCount, scanStatus)
			op.Notification(notificationMsg)
		}

		// If fewer results than limit were returned, this is the last page
		if len(results) < limit {
			break
		}

		// Move to the next page
		offset += limit
	}

	return nil
}

// fetchPocTemplates fetches the POC template list from the API
func fetchPocTemplates(offset, limit int) ([]templateResult, error) {
	// Build the request URL
	reqURL := "https://api.projectdiscovery.io/v2/template/search"
	params := url.Values{}
	params.Set("scope", "public")
	params.Set("offset", strconv.Itoa(offset))
	params.Set("limit", strconv.Itoa(limit))
	params.Set("isSignedIn", "false")

	fullURL := reqURL + "?" + params.Encode()

	// Create the request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set request headers
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://cloud.projectdiscovery.io")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://cloud.projectdiscovery.io/")
	req.Header.Set("sec-ch-ua", `"Chromium";v="136", "Google Chrome";v="136", "Not.A/Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	req.Header.Set("x-team-id", "cus16bdim6vs73cqnh0g")

	// Create the HTTP client and send the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	// Parse the response
	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return apiResp.Results, nil
}
