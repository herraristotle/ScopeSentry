// utils-------------------------------------
// @file      : filetool.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/25 15:54
// -------------------------------------------

package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func EnsureDir(path string) error {
	// Check the path status
	info, err := os.Stat(path)

	// If the path does not exist
	if os.IsNotExist(err) {
		// Automatically create nested directories
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		return nil
	}

	// Path exists but is not a directory
	if err == nil && !info.IsDir() {
		return fmt.Errorf("path already exists but is not a directory: %s", path)
	}

	// Already exists and is a directory
	return nil
}

func WriteFile(filePath string, content []byte) error {
	// os.Create empties an existing file's content, equivalent to overwriting it
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write the content
	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func DownloadFile(url string, savePath string, timeout time.Duration) error {
	// Create the directory
	if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
		return err
	}

	client := &http.Client{
		Timeout: timeout, // Overall timeout (connect + download)
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func DeleteFile(filePath string) {
	// Check whether the file exists
	if filePath == "" {
		return
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	} else if err != nil {
		return
	}

	// File exists, delete it
	err := os.Remove(filePath)
	if err != nil {
		return
	}
}
