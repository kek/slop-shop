package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kek/slop-shop/ollama"
	"github.com/kek/slop-shop/styles"
)

// DiffChange represents a single file change from a diff
type DiffChange struct {
	FilePath string
	Hunks    []DiffHunk
}

// DiffHunk represents a section of changes in a file
type DiffHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []DiffLine
}

// DiffLine represents a single line change
type DiffLine struct {
	Type    string // " ", "-", "+"
	Content string
	LineNum int
}

// ExecuteTools executes tools found in the LLM response
func ExecuteTools(response, repoPath string) string {
	fmt.Println(styles.HeaderStyle.Render("\n🔧 Tool Execution"))
	fmt.Println(styles.SeparatorStyle.Render("================================================"))

	var results strings.Builder
	results.WriteString("Tool Execution Results:\n")
	results.WriteString("=====================\n\n")

	lines := strings.Split(response, "\n")
	toolCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Execute RUN_COMMAND
		if strings.HasPrefix(line, "RUN_COMMAND:") {
			toolCount++
			command := strings.TrimSpace(strings.TrimPrefix(line, "RUN_COMMAND:"))
			fmt.Printf(styles.ToolStyle.Render("🔧 [%d] RUN_COMMAND detected: %s\n"), toolCount, command)
			fmt.Print(styles.InfoStyle.Render("   📍 Working directory: " + repoPath + "\n"))
			fmt.Print(styles.InfoStyle.Render("   ⏳ Executing...\n"))

			result := executeCommand(command, repoPath)

			fmt.Print(styles.SuccessStyle.Render("   ✅ Completed\n"))
			results.WriteString(fmt.Sprintf("RUN_COMMAND: %s\n", command))
			results.WriteString(result)
			results.WriteString("\n")
		}

		// Execute READ_FILE
		if strings.HasPrefix(line, "READ_FILE:") {
			toolCount++
			filePath := strings.TrimSpace(strings.TrimPrefix(line, "READ_FILE:"))
			fmt.Printf(styles.ToolStyle.Render("📖 [%d] READ_FILE detected: %s\n"), toolCount, filePath)
			fmt.Print(styles.InfoStyle.Render("   📍 Repository: " + repoPath + "\n"))
			fmt.Print(styles.InfoStyle.Render("   ⏳ Reading...\n"))

			result := readFileContent(filePath, repoPath)

			fmt.Print(styles.SuccessStyle.Render("   ✅ Completed\n"))
			results.WriteString(fmt.Sprintf("READ_FILE: %s\n", filePath))
			results.WriteString(result)
			results.WriteString("\n")
		}

		// Execute LIST_DIR
		if strings.HasPrefix(line, "LIST_DIR:") {
			toolCount++
			dir := strings.TrimSpace(strings.TrimPrefix(line, "LIST_DIR:"))
			fmt.Printf("📁 [%d] LIST_DIR detected: %s\n", toolCount, dir)
			fmt.Printf("   📍 Repository: %s\n", repoPath)
			fmt.Printf("   ⏳ Scanning...\n")

			result := listDirectory(dir, repoPath)

			fmt.Printf("   ✅ Completed\n")
			results.WriteString(fmt.Sprintf("LIST_DIR: %s\n", dir))
			results.WriteString(result)
			results.WriteString("\n")
		}

		// Execute TEST_COMMAND
		if strings.HasPrefix(line, "TEST_COMMAND:") {
			toolCount++
			command := strings.TrimSpace(strings.TrimPrefix(line, "TEST_COMMAND:"))
			fmt.Printf("🧪 [%d] TEST_COMMAND detected: %s\n", toolCount, command)
			fmt.Printf("   📍 Working directory: %s\n", repoPath)
			fmt.Printf("   ⏳ Testing...\n")

			result := testCommand(command, repoPath)

			fmt.Printf("   ✅ Completed\n")
			results.WriteString(fmt.Sprintf("TEST_COMMAND: %s\n", command))
			results.WriteString(result)
			results.WriteString("\n")
		}

		// Execute SEARCH_FILES
		if strings.HasPrefix(line, "SEARCH_FILES:") {
			toolCount++
			parts := strings.SplitN(strings.TrimPrefix(line, "SEARCH_FILES:"), " ", 2)
			if len(parts) == 2 {
				pattern := strings.TrimSpace(parts[0])
				directory := strings.TrimSpace(parts[1])
				fmt.Printf("🔍 [%d] SEARCH_FILES detected: pattern='%s' in '%s'\n", toolCount, pattern, directory)
				fmt.Printf("   📍 Repository: %s\n", repoPath)
				fmt.Printf("   ⏳ Searching...\n")

				result := searchFiles(pattern, directory, repoPath)

				fmt.Printf("   ✅ Completed\n")
				results.WriteString(fmt.Sprintf("SEARCH_FILES: %s in %s\n", pattern, directory))
				results.WriteString(result)
				results.WriteString("\n")
			}
		}

		// Execute GENERATE_DIFF
		if strings.HasPrefix(line, "GENERATE_DIFF:") {
			toolCount++
			description := strings.TrimSpace(strings.TrimPrefix(line, "GENERATE_DIFF:"))
			fmt.Printf("📝 [%d] GENERATE_DIFF detected: %s\n", toolCount, description)
			fmt.Printf("   📍 Repository: %s\n", repoPath)
			fmt.Printf("   ⏳ Generating diff...\n")

			result := generateDiff(description, repoPath)

			fmt.Printf("   ✅ Completed\n")
			results.WriteString(fmt.Sprintf("GENERATE_DIFF: %s\n", description))
			results.WriteString(result)
			results.WriteString("\n")
		}

		// Execute APPLY_DIFF
		if strings.HasPrefix(line, "APPLY_DIFF:") {
			toolCount++
			diffContent := strings.TrimSpace(strings.TrimPrefix(line, "APPLY_DIFF:"))
			fmt.Printf("🔧 [%d] APPLY_DIFF detected\n", toolCount)
			fmt.Printf("   📍 Repository: %s\n", repoPath)
			fmt.Printf("   ⏳ Applying diff...\n")

			result := applyDiffTool(diffContent, repoPath)

			fmt.Printf("   ✅ Completed\n")
			results.WriteString("APPLY_DIFF: Applied\n")
			results.WriteString(result)
			results.WriteString("\n")
		}

		// Execute CREATE_FILE
		if strings.HasPrefix(line, "CREATE_FILE:") {
			toolCount++
			filePath := strings.TrimSpace(strings.TrimPrefix(line, "CREATE_FILE:"))
			fmt.Printf("📝 [%d] CREATE_FILE detected: %s\n", toolCount, filePath)
			fmt.Printf("   📍 Repository: %s\n", repoPath)
			fmt.Printf("   ⏳ Creating file...\n")

			// Collect content until END_FILE
			var contentLines []string
			for i := toolCount; i < len(lines); i++ {
				if strings.TrimSpace(lines[i]) == "END_FILE" {
					break
				}
				contentLines = append(contentLines, lines[i])
			}
			content := strings.Join(contentLines, "\n")

			result := createFile(filePath, content, repoPath)

			fmt.Printf("   ✅ Completed\n")
			results.WriteString(fmt.Sprintf("CREATE_FILE: %s\n", filePath))
			results.WriteString(result)
			results.WriteString("\n")
		}
	}

	if toolCount == 0 {
		fmt.Println(styles.InfoStyle.Render("ℹ️  No tools detected in LLM response"))
	} else {
		fmt.Printf(styles.SuccessStyle.Render("🎯 Total tools executed: %d\n"), toolCount)
	}

	fmt.Println(styles.SeparatorStyle.Render("================================================"))

	return results.String()
}

// generateDiff generates a unified diff based on a description
func generateDiff(description, repoPath string) string {
	// Use the LLM to generate an actual diff
	diffPrompt := fmt.Sprintf("Based on this description: '%s', generate a unified diff that implements the requested changes. "+
		"Only output the unified diff format, no explanations. The diff should be in the format:\n"+
		"--- a/filename\n"+
		"+++ b/filename\n"+
		"@@ -line,count +line,count @@\n"+
		" unchanged line\n"+
		"-removed line\n"+
		"+added line\n\n"+
		"Description: %s", description, description)

	// Send to Ollama to generate the diff
	fmt.Printf("   🤖 Generating diff with LLM...\n")
	var response strings.Builder
	_, err := ollama.SendToOllamaWithCallback("http://localhost:11434", "qwen3-coder", diffPrompt, "", 0.3, 0.8, true, func(chunk string) {
		response.WriteString(chunk)
	})
	if err != nil {
		return fmt.Sprintf("Error generating diff: %v", err)
	}

	// Check if the response contains a valid diff format
	if strings.Contains(response.String(), "--- a/") && strings.Contains(response.String(), "+++ b/") {
		return fmt.Sprintf("Generated diff:\n\n%s", response.String())
	} else {
		return fmt.Sprintf("LLM response (may not be valid diff format):\n\n%s\n\nNote: This may not be a valid unified diff. "+
			"You can copy the content above and use APPLY_DIFF if it looks correct.", response.String())
	}
}

// applyDiffTool applies a unified diff using the existing diff logic
func applyDiffTool(diffContent, repoPath string) string {
	if err := applyDiff(diffContent, repoPath); err != nil {
		return fmt.Sprintf("Error applying diff: %v", err)
	}
	return "Diff applied successfully to the repository"
}

// executeCommand executes a shell command
func executeCommand(command, repoPath string) string {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error executing command: %v\nOutput: %s", err, string(output))
	}

	return fmt.Sprintf("Command executed successfully:\n%s", string(output))
}

// readFileContent reads the contents of a file
func readFileContent(filePath, repoPath string) string {
	fullPath := filePath
	if !strings.HasPrefix(filePath, "/") {
		fullPath = filepath.Join(repoPath, filePath)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}

	return fmt.Sprintf("File contents:\n%s", string(content))
}

// listDirectory lists the contents of a directory
func listDirectory(dir, repoPath string) string {
	fullPath := dir
	if !strings.HasPrefix(dir, "/") {
		fullPath = filepath.Join(repoPath, dir)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return fmt.Sprintf("Error reading directory: %v", err)
	}

	var result strings.Builder
	result.WriteString("Directory contents:\n")
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileType := "f"
		if info.IsDir() {
			fileType = "d"
		}

		result.WriteString(fmt.Sprintf("%s %8d %s\n", fileType, info.Size(), entry.Name()))
	}

	return result.String()
}

// testCommand tests if a command works
func testCommand(command, repoPath string) string {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output))
	}

	return fmt.Sprintf("Command works successfully:\n%s", string(output))
}

// searchFiles searches for text patterns in files
func searchFiles(pattern, directory, repoPath string) string {
	fullPath := directory
	if !strings.HasPrefix(directory, "/") {
		fullPath = filepath.Join(repoPath, directory)
	}

	var results strings.Builder
	results.WriteString("Search results:\n")

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Skip binary files
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if !isTextFile(content) {
			return nil
		}

		// Simple text search
		if strings.Contains(string(content), pattern) {
			relPath, _ := filepath.Rel(repoPath, path)
			results.WriteString(fmt.Sprintf("Found in: %s\n", relPath))
		}

		return nil
	})

	if err != nil {
		return fmt.Sprintf("Error searching files: %v", err)
	}

	return results.String()
}

// createFile creates a new file with the specified content
func createFile(filePath, content, repoPath string) string {
	fullPath := filePath
	if !strings.HasPrefix(filePath, "/") {
		fullPath = filepath.Join(repoPath, filePath)
	}

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}

	// Create the file with content
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Sprintf("Error creating file: %v", err)
	}

	return fmt.Sprintf("File created successfully: %s", filePath)
}

// applyDiff applies a unified diff to the repository
func applyDiff(diffOutput, repoPath string) error {
	// Parse the diff output to extract file changes
	changes, err := parseDiff(diffOutput)
	if err != nil {
		return fmt.Errorf("failed to parse diff: %v", err)
	}

	// Apply each change
	for _, change := range changes {
		if err := applyFileChange(change, repoPath); err != nil {
			return fmt.Errorf("failed to apply change to %s: %v", change.FilePath, err)
		}
	}

	return nil
}

// parseDiff parses a unified diff output
func parseDiff(diffOutput string) ([]DiffChange, error) {
	var changes []DiffChange
	lines := strings.Split(diffOutput, "\n")

	var currentChange *DiffChange
	var currentHunk *DiffHunk

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// File header
		if strings.HasPrefix(line, "--- a/") {
			if currentChange != nil {
				if currentHunk != nil {
					currentChange.Hunks = append(currentChange.Hunks, *currentHunk)
				}
				changes = append(changes, *currentChange)
			}

			filePath := strings.TrimPrefix(line, "--- a/")
			currentChange = &DiffChange{FilePath: filePath}
			currentHunk = nil
			continue
		}

		if strings.HasPrefix(line, "+++ b/") {
			// Verify file path matches
			filePath := strings.TrimPrefix(line, "+++ b/")
			if currentChange != nil && currentChange.FilePath != filePath {
				return nil, fmt.Errorf("mismatched file paths in diff: %s vs %s", currentChange.FilePath, filePath)
			}
			continue
		}

		// Hunk header
		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil && currentChange != nil {
				currentChange.Hunks = append(currentChange.Hunks, *currentHunk)
			}

			// Parse hunk header: @@ -oldStart,oldCount +newStart,newCount @@
			parts := strings.Split(line, " ")
			if len(parts) < 3 {
				continue
			}

			oldPart := strings.TrimPrefix(parts[1], "-")
			newPart := strings.TrimPrefix(parts[2], "+")

			oldStart, oldCount := parseRange(oldPart)
			newStart, newCount := parseRange(newPart)

			currentHunk = &DiffHunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
			}
			continue
		}

		// Content lines
		if currentHunk != nil {
			lineType := " "
			content := line

			if strings.HasPrefix(line, "+") {
				lineType = "+"
				content = strings.TrimPrefix(line, "+")
			} else if strings.HasPrefix(line, "-") {
				lineType = "-"
				content = strings.TrimPrefix(line, "-")
			}

			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:    lineType,
				Content: content,
			})
		}
	}

	// Add the last change and hunk
	if currentChange != nil {
		if currentHunk != nil {
			currentChange.Hunks = append(currentChange.Hunks, *currentHunk)
		}
		changes = append(changes, *currentChange)
	}

	return changes, nil
}

// parseRange parses a range like "10,5" into start and count
func parseRange(rangeStr string) (start, count int) {
	parts := strings.Split(rangeStr, ",")
	if len(parts) >= 1 {
		start, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		count, _ = strconv.Atoi(parts[1])
	} else {
		count = 1
	}
	return start, count
}

// applyFileChange applies changes to a single file
func applyFileChange(change DiffChange, repoPath string) error {
	filePath := filepath.Join(repoPath, change.FilePath)

	// Read current file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Apply changes in reverse order to maintain line numbers
	for i := len(change.Hunks) - 1; i >= 0; i-- {
		hunk := change.Hunks[i]
		lines = applyHunk(lines, hunk)
	}

	// Write modified content back to file
	newContent := strings.Join(lines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("Applied changes to: %s\n", change.FilePath)
	return nil
}

// applyHunk applies a single hunk to the file lines
func applyHunk(lines []string, hunk DiffHunk) []string {
	// Convert to 0-based indexing
	start := hunk.OldStart - 1

	// Remove old lines
	if hunk.OldCount > 0 {
		end := start + hunk.OldCount
		if end > len(lines) {
			end = len(lines)
		}
		lines = append(lines[:start], lines[end:]...)
	}

	// Insert new lines
	newLines := make([]string, 0, len(hunk.Lines))
	for _, line := range hunk.Lines {
		if line.Type == "+" || line.Type == " " {
			newLines = append(newLines, line.Content)
		}
	}

	if len(newLines) > 0 {
		lines = append(lines[:start], append(newLines, lines[start:]...)...)
	}

	return lines
}

func isTextFile(content []byte) bool {
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
