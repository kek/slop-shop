package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileInfo represents information about a file in the repository
type FileInfo struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Size    int64  `json:"size"`
}

// ReadRepository walks through the repository and reads all relevant files
func ReadRepository(repoPath string, excludePatterns []string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file should be excluded
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		if ShouldExclude(relPath, excludePatterns) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Warning: Could not read file %s: %v\n", path, err)
			return nil
		}

		// Check if file is text-based (simple heuristic)
		if IsTextFile(content) {
			files = append(files, FileInfo{
				Path:    relPath,
				Content: string(content),
				Size:    info.Size(),
			})
		}

		return nil
	})

	return files, err
}

// ShouldExclude checks if a file path matches any exclude pattern
func ShouldExclude(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		// Simple pattern matching
		if strings.Contains(pattern, "*") {
			// Basic glob-like matching
			if strings.HasSuffix(pattern, "*") {
				prefix := strings.TrimSuffix(pattern, "*")
				if strings.HasPrefix(path, prefix) {
					return true
				}
			}
		} else if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// IsTextFile checks if file content appears to be text-based
func IsTextFile(content []byte) bool {
	// Check first 1024 bytes for null bytes
	checkSize := len(content)
	if checkSize > 1024 {
		checkSize = 1024
	}

	for i := 0; i < checkSize; i++ {
		if content[i] == 0 {
			return false
		}
	}
	return true
}

// CreateContext creates a formatted context string from repository files
func CreateContext(files []FileInfo) string {
	var buf strings.Builder

	buf.WriteString("Repository Contents:\n")
	buf.WriteString("===================\n\n")

	for _, file := range files {
		buf.WriteString(fmt.Sprintf("File: %s (Size: %d bytes)\n", file.Path, file.Size))
		buf.WriteString(strings.Repeat("-", 50) + "\n")
		buf.WriteString(file.Content)
		buf.WriteString("\n\n")
	}

	return buf.String()
}
