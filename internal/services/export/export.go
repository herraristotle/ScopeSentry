package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/logger"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/repositories/export"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service defines the export service interface
type Service interface {
	CreateExportTask(ctx context.Context, task *models.ExportTask) error
	ProcessExportTask(ctx context.Context, run models.ExportRun)
	GetExportFields(ctx context.Context, index string) ([]string, error)
	GetExportRecords(ctx context.Context) ([]models.ExportTask, error)
	DeleteExportFiles(ctx context.Context, fileNames []string) error
	IsValidFileName(fileName string) bool
}

// ExportTask is the export task struct

type service struct {
	repository export.Repository
}

// NewService creates an export service instance
func NewService() Service {
	return &service{
		repository: export.NewRepository(),
	}
}

// CreateExportTask creates an export task
func (s *service) CreateExportTask(ctx context.Context, task *models.ExportTask) error {
	return s.repository.CreateExportTask(ctx, task)
}

// ProcessExportTask processes an export task
func (s *service) ProcessExportTask(ctx context.Context, run models.ExportRun) {

	// Build query options at the service layer
	opts := options.Find().
		SetLimit(int64(run.Quantity))

	// Handle projection at the service layer
	projection := bson.M{}
	for _, field := range run.Field {
		projection[field] = 1
	}
	if run.Index == "asset" {
		projection["type"] = 1
	}
	opts.SetProjection(projection)

	// Get the cursor
	cursor, err := s.repository.GetCollectionCursor(
		ctx,
		run.Index,
		run.Filter,
		opts,
	)
	if err != nil {
		err := s.repository.UpdateExportTaskStatus(ctx, run.FileName, 2, 0)
		if err != nil {
			logger.Error(fmt.Sprintf("Error updating export task status: %s", err.Error()))
			return
		}
		return
	}
	defer cursor.Close(ctx)

	// Export the data
	filePath := filepath.Join(config.GlobalConfig.System.ExeDir, "files", "export", run.FileName)
	var fileSize float64
	if run.FileType == "xlsx" {
		fileSize, err = s.exportToXlsx(cursor, filePath, run.Field, run.Index)
	} else {
		fileSize, err = s.exportToJSON(cursor, filePath)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("Error export: %s", err.Error()))
		err := s.repository.UpdateExportTaskStatus(ctx, run.FileName, 2, 0)
		if err != nil {
			logger.Error(fmt.Sprintf("Error updating export task status: %s", err.Error()))
			return
		}
		return
	}

	// Update the task status
	err = s.repository.UpdateExportTaskStatus(ctx, run.FileName, 1, fileSize)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating export task status: %s", err.Error()))
		return
	}
}

// GetExportFields gets the export fields
func (s *service) GetExportFields(ctx context.Context, index string) ([]string, error) {
	// Get the field list from the configuration
	fields, ok := models.Field[index]
	if !ok {
		return []string{}, fmt.Errorf("not found: %s", index)
	}
	return fields, nil
}

// GetExportRecords gets the export records
func (s *service) GetExportRecords(ctx context.Context) ([]models.ExportTask, error) {
	// Build query options at the service layer
	opts := options.Find().
		SetSort(bson.D{{"create_time", -1}}).
		SetProjection(bson.D{
			{"_id", 0},
			{"id", bson.D{{"$toString", "$_id"}}},
			{"file_name", 1},
			{"end_time", 1},
			{"create_time", 1},
			{"data_type", 1},
			{"state", 1},
			{"file_size", 1},
			{"file_type", 1},
		})

	// Call the repository layer to execute the query
	return s.repository.GetExportRecords(ctx, bson.M{}, opts)
}

// DeleteExportFiles deletes export files
func (s *service) DeleteExportFiles(ctx context.Context, fileNames []string) error {
	// Delete the files
	for _, fileName := range fileNames {
		if s.IsValidFileName(fileName) {
			filePath := filepath.Join(config.GlobalConfig.System.ExeDir, "files", "export", fileName)
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				logger.Error(fmt.Sprintf("failed to delete file: %s, error: %v", fileName, err))
				continue
			}
		}
	}

	// Call the repository layer to delete the database records
	return s.repository.DeleteExportTasks(ctx, fileNames)
}

// IsValidFileName checks whether a file name is valid
func (s *service) IsValidFileName(fileName string) bool {
	parts := strings.Split(fileName, ".")
	if len(parts) != 2 {
		return false
	}

	name := parts[0]
	fileType := parts[1]
	if fileType != "json" && fileType != "xlsx" {
		return false
	}
	matched, err := regexp.MatchString(`^[a-z0-9-]+$`, name)
	if err != nil {
		return false
	}
	return matched
}

// Generate an Excel cell coordinate
func getCellName(col, row int) string {
	colName := ""
	for col > 0 {
		col--
		colName = string(rune('A'+col%26)) + colName
		col /= 26
	}
	if colName == "" {
		colName = "A"
	}
	return fmt.Sprintf("%s%d", colName, row)
}

// sanitizeExcelValue strips characters not allowed by Excel
// Excel XML does not allow control characters (0x00-0x1F, except 0x09 (TAB), 0x0A (LF), 0x0D (CR))
func sanitizeExcelValue(value interface{}) string {
	if value == nil {
		return ""
	}

	str := fmt.Sprintf("%v", value)
	// Filter out invalid control characters, keeping TAB, LF and CR
	result := make([]rune, 0, len(str))
	for _, r := range str {
		// Allowed: printable characters (>=0x20), TAB (0x09), LF (0x0A), CR (0x0D)
		if r >= 0x20 || r == 0x09 || r == 0x0A || r == 0x0D {
			result = append(result, r)
		} else {
			// Replace invalid characters with a space or skip them
			// Here we skip them; use a space if the position needs to be preserved
		}
	}
	return string(result)
}

// Use streaming writes to export to Excel
func (s *service) exportToXlsx(cursor *mongo.Cursor, filePath string, fields []string, index string) (float64, error) {
	// Create a new Excel file
	f := excelize.NewFile()
	defer f.Close()
	start := time.Now()

	// Streaming flush interval (flush every 5000 rows to reduce memory usage)
	const flushInterval = 1000

	if index == "asset" {
		// Separate HTTP and Other fields
		httpFields := []string{}
		otherFields := []string{}
		for _, field := range fields {
			if models.IsHTTPAssetField(field) {
				httpFields = append(httpFields, field)
			}
			if models.IsOtherAssetField(field) {
				otherFields = append(otherFields, field)
			}
		}

		// Create the HTTP worksheet
		httpSheet := "HTTP Data"
		err := f.SetSheetName("Sheet1", httpSheet)
		if err != nil {
			logger.Error(fmt.Sprintf("Error setting sheet name: %s", err.Error()))
			return 0, err
		}

		// Create the Other worksheet
		otherSheet := "Other Data"
		_, err = f.NewSheet(otherSheet)
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating sheet: %s", err.Error()))
			return 0, err
		}

		// Create the stream writer
		httpStreamWriter, err := f.NewStreamWriter(httpSheet)
		if err != nil {
			return 0, fmt.Errorf("failed to create stream writer for HTTP sheet: %w", err)
		}

		otherStreamWriter, err := f.NewStreamWriter(otherSheet)
		if err != nil {
			return 0, fmt.Errorf("failed to create stream writer for Other sheet: %w", err)
		}

		// Write the header - convert to an interface{} slice
		httpHeader := make([]interface{}, len(httpFields))
		for i, field := range httpFields {
			httpHeader[i] = field
		}
		if err := httpStreamWriter.SetRow("A1", httpHeader); err != nil {
			return 0, fmt.Errorf("failed to write HTTP header: %w", err)
		}

		otherHeader := make([]interface{}, len(otherFields))
		for i, field := range otherFields {
			otherHeader[i] = field
		}
		if err := otherStreamWriter.SetRow("A1", otherHeader); err != nil {
			return 0, fmt.Errorf("failed to write Other header: %w", err)
		}

		// Process the data
		// Note: excelize streaming requires row numbers to increase strictly consecutively; no row number may be skipped
		// The header is written to row 1; data starts at row 2
		httpRowIdx := 2  // start at 2 (row 1 is the header)
		otherRowIdx := 2 // start at 2 (row 1 is the header)
		httpCount := 0   // number of HTTP rows actually written
		otherCount := 0  // number of Other rows actually written

		for cursor.Next(context.Background()) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				logger.Error(fmt.Sprintf("Failed to decode document: %v, skipping...", err))
				continue
			}

			if doc["type"] == "http" {
				// Prepare the HTTP data row, stripping invalid characters
				rowData := make([]interface{}, len(httpFields))
				for i, field := range httpFields {
					value := doc[field]
					rowData[i] = sanitizeExcelValue(value)
				}

				// Use streaming write - row numbers must increase strictly consecutively
				cell, err := excelize.CoordinatesToCellName(1, httpRowIdx)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to convert coordinates for HTTP row %d: %v", httpRowIdx, err))
					httpRowIdx++ // increment the row number even on failure to keep continuity
					continue
				}

				if err := httpStreamWriter.SetRow(cell, rowData); err != nil {
					logger.Error(fmt.Sprintf("Failed to write HTTP row %d: %v", httpRowIdx, err))
					// Try writing an empty row to keep row numbers consecutive
					emptyRow := make([]interface{}, len(httpFields))
					for i := range emptyRow {
						emptyRow[i] = ""
					}
					if retryErr := httpStreamWriter.SetRow(cell, emptyRow); retryErr != nil {
						return 0, fmt.Errorf("excelize stream writer corrupted at HTTP row %d: %w", httpRowIdx, retryErr)
					}
				} else {
					httpCount++
				}
				httpRowIdx++ // the row number must be incremented regardless of success or failure
			} else {
				// Prepare the Other data row, stripping invalid characters
				rowData := make([]interface{}, len(otherFields))
				for i, field := range otherFields {
					value := doc[field]
					rowData[i] = sanitizeExcelValue(value)
				}

				// Use streaming write - row numbers must increase strictly consecutively
				cell, err := excelize.CoordinatesToCellName(1, otherRowIdx)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to convert coordinates for Other row %d: %v", otherRowIdx, err))
					otherRowIdx++ // increment the row number even on failure to keep continuity
					continue
				}

				if err := otherStreamWriter.SetRow(cell, rowData); err != nil {
					logger.Error(fmt.Sprintf("Failed to write Other row %d: %v", otherRowIdx, err))
					// Try writing an empty row to keep row numbers consecutive
					emptyRow := make([]interface{}, len(otherFields))
					for i := range emptyRow {
						emptyRow[i] = ""
					}
					if retryErr := otherStreamWriter.SetRow(cell, emptyRow); retryErr != nil {
						return 0, fmt.Errorf("excelize stream writer corrupted at Other row %d: %w", otherRowIdx, retryErr)
					}
				} else {
					otherCount++
				}
				otherRowIdx++ // the row number must be incremented regardless of success or failure
			}

		}

		// Check for cursor errors
		if err := cursor.Err(); err != nil {
			return 0, fmt.Errorf("cursor error: %w", err)
		}

		// Final flush
		if err := httpStreamWriter.Flush(); err != nil {
			return 0, fmt.Errorf("failed to flush HTTP stream: %w", err)
		}
		if err := otherStreamWriter.Flush(); err != nil {
			return 0, fmt.Errorf("failed to flush Other stream: %w", err)
		}

		// Save the file
		if err := f.SaveAs(filePath); err != nil {
			return 0, fmt.Errorf("failed to save excel file: %w", err)
		}

		// Get the file size
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return 0, fmt.Errorf("failed to get file info: %w", err)
		}
		fileSize := float64(fileInfo.Size()) / (1024 * 1024)
		end := time.Now()
		totalRows := httpCount + otherCount
		logger.Info(fmt.Sprintf("Saved %v excel file size: %.2f MB time: %v number: %v (HTTP: %v, Other: %v)", index, fileSize, end.Sub(start), totalRows, httpCount, otherCount))
		return fileSize, nil
	} else {
		// Process other data types
		sheet := "Sheet1"

		// Create the stream writer
		streamWriter, err := f.NewStreamWriter(sheet)
		if err != nil {
			return 0, fmt.Errorf("failed to create stream writer: %w", err)
		}

		// Write the header
		header := make([]interface{}, len(fields))
		for i, field := range fields {
			header[i] = field
		}
		if err := streamWriter.SetRow("A1", header); err != nil {
			return 0, fmt.Errorf("failed to write header: %w", err)
		}

		// Process the data
		row := 2
		rowCount := 0 // number of rows actually written

		for cursor.Next(context.Background()) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				logger.Error(fmt.Sprintf("Failed to decode document: %v, skipping...", err))
				continue
			}

			// Prepare the data row, stripping invalid characters
			rowData := make([]interface{}, len(fields))
			for i, field := range fields {
				value := doc[field]
				rowData[i] = sanitizeExcelValue(value)
			}

			// Use streaming write
			cell, err := excelize.CoordinatesToCellName(1, row)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to convert coordinates for row %d: %v", row, err))
				row++ // increment the row number even on failure to keep continuity
				continue
			}

			if err := streamWriter.SetRow(cell, rowData); err != nil {
				logger.Error(fmt.Sprintf("Failed to write row %d: %v", row, err))
				// Try writing an empty row to keep row numbers consecutive
				emptyRow := make([]interface{}, len(fields))
				for i := range emptyRow {
					emptyRow[i] = ""
				}
				if retryErr := streamWriter.SetRow(cell, emptyRow); retryErr != nil {
					return 0, fmt.Errorf("excelize stream writer corrupted at row %d: %w", row, retryErr)
				}
			} else {
				rowCount++
			}
			row++ // the row number must be incremented regardless of success or failure
		}

		// Check for cursor errors
		if err := cursor.Err(); err != nil {
			return 0, fmt.Errorf("cursor error: %w", err)
		}

		// Final flush
		if err := streamWriter.Flush(); err != nil {
			return 0, fmt.Errorf("failed to flush stream: %w", err)
		}

		// Save the file
		if err := f.SaveAs(filePath); err != nil {
			return 0, fmt.Errorf("failed to save excel file: %w", err)
		}

		// Get the file size
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return 0, fmt.Errorf("failed to get file info: %w", err)
		}
		fileSize := float64(fileInfo.Size()) / (1024 * 1024)
		end := time.Now()
		logger.Info(fmt.Sprintf("Saved %v excel file size: %.2f MB time: %v number: %v", index, fileSize, end.Sub(start), rowCount))
		return fileSize, nil
	}
}

// Use streaming encoding to export to JSON
func (s *service) exportToJSON(cursor *mongo.Cursor, filePath string) (float64, error) {
	start := time.Now()
	file, err := os.Create(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Stream-encode with a JSON encoder directly into the file, avoiding a temporary byte array
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false) // Do not escape HTML characters, for better performance

	// Read data row by row using the cursor
	row := 0
	for cursor.Next(context.Background()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			logger.Error(fmt.Sprintf("Failed to decode document: %v, skipping...", err))
			continue
		}

		// Stream-encode directly into the file, avoiding a temporary byte array
		if err := encoder.Encode(doc); err != nil {
			logger.Error(fmt.Sprintf("Failed to encode document: %v, skipping...", err))
			continue
		}
		row++
	}

	// Check for cursor errors
	if err := cursor.Err(); err != nil {
		return 0, fmt.Errorf("cursor error: %w", err)
	}

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := float64(fileInfo.Size()) / (1024 * 1024)
	end := time.Now()
	logger.Info(fmt.Sprintf("Saved JSON file size: %.2f MB time: %v number: %v", fileSize, end.Sub(start), row))
	return fileSize, nil
}
