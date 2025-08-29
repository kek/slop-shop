package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kek/slop-shop/ollama"
	"github.com/kek/slop-shop/repo"
	"github.com/kek/slop-shop/tools"
)

// MockOllamaServer creates a test server that mimics Ollama API responses
func MockOllamaServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/generate" {
			// Simulate streaming response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// Send multiple chunks to simulate streaming
			chunks := []string{
				`{"response":"Hello","done":false}`,
				`{"response":" from","done":false}`,
				`{"response":" mock","done":false}`,
				`{"response":" Ollama","done":true}`,
			}
			
			for _, chunk := range chunks {
				fmt.Fprintf(w, "%s\n", chunk)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(10 * time.Millisecond) // Simulate network delay
			}
		} else {
			http.NotFound(w, r)
		}
	}))
}

func TestOllamaIntegration(t *testing.T) {
	server := MockOllamaServer()
	defer server.Close()

	t.Run("Ollama API Call", func(t *testing.T) {
		var receivedChunks []string
		chunkCallback := func(chunk string) {
			receivedChunks = append(receivedChunks, chunk)
		}

		response, err := ollama.SendToOllamaWithCallback(
			server.URL,
			"test-model",
			"Test prompt",
			"",
			0.7,
			0.9,
			false,
			chunkCallback,
		)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		expectedResponse := "Hello from mock Ollama"
		if response != expectedResponse {
			t.Errorf("Expected response %q, got %q", expectedResponse, response)
		}

		expectedChunks := []string{"Hello", " from", " mock", " Ollama"}
		if len(receivedChunks) != len(expectedChunks) {
			t.Errorf("Expected %d chunks, got %d", len(expectedChunks), len(receivedChunks))
		}

		for i, expected := range expectedChunks {
			if i < len(receivedChunks) && receivedChunks[i] != expected {
				t.Errorf("Chunk %d: expected %q, got %q", i, expected, receivedChunks[i])
			}
		}
	})

	t.Run("Ollama API with Tools", func(t *testing.T) {
		var receivedChunks []string
		chunkCallback := func(chunk string) {
			receivedChunks = append(receivedChunks, chunk)
		}

		response, err := ollama.SendToOllamaWithCallback(
			server.URL,
			"test-model",
			"Test prompt with tools",
			"",
			0.7,
			0.9,
			true, // Enable tools
			chunkCallback,
		)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should still get the same response even with tools enabled
		expectedResponse := "Hello from mock Ollama"
		if response != expectedResponse {
			t.Errorf("Expected response %q, got %q", expectedResponse, response)
		}
	})
}

func TestRepositoryScanningIntegration(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "slop-shop-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files of different types
	testFiles := map[string]string{
		"main.go":        "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
		"README.md":      "# Test Project\n\nThis is a test project.",
		"config.json":    `{"name": "test", "version": "1.0.0"}`,
		"script.sh":      "#!/bin/bash\necho 'Hello World'",
		"Makefile":       "build:\n\tgo build -o app main.go",
		"test.txt":       "This is a plain text file",
		"subdir/file.go": "package subdir\n\nfunc Helper() {}",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	t.Run("Repository Reading", func(t *testing.T) {
		files, err := repo.ReadRepository(tempDir, []string{})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		expectedFileCount := len(testFiles)
		if len(files) != expectedFileCount {
			t.Errorf("Expected %d files, got %d", expectedFileCount, len(files))
		}

		// Check that all expected files are present
		foundFiles := make(map[string]bool)
		for _, file := range files {
			foundFiles[file.Path] = true
		}

		for expectedFile := range testFiles {
			if !foundFiles[expectedFile] {
				t.Errorf("Expected file %s not found", expectedFile)
			}
		}
	})

	t.Run("Repository Reading with Exclusions", func(t *testing.T) {
		excludePatterns := []string{"test.txt", "subdir"}
		files, err := repo.ReadRepository(tempDir, excludePatterns)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should exclude test.txt and subdir/file.go
		expectedFileCount := len(testFiles) - 2
		if len(files) != expectedFileCount {
			t.Errorf("Expected %d files after exclusions, got %d", expectedFileCount, len(files))
		}

		// Verify excluded files are not present
		for _, file := range files {
			if file.Path == "test.txt" || strings.HasPrefix(file.Path, "subdir/") {
				t.Errorf("Excluded file %s should not be present", file.Path)
			}
		}
	})

	t.Run("Context Creation", func(t *testing.T) {
		files, err := repo.ReadRepository(tempDir, []string{})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		context := repo.CreateContext(files)
		if context == "" {
			t.Error("Expected non-empty context")
		}

		// Check that context contains expected file information
		expectedFiles := []string{"main.go", "README.md", "config.json", "script.sh", "Makefile"}
		for _, expectedFile := range expectedFiles {
			if !strings.Contains(context, expectedFile) {
				t.Errorf("Context should contain %s", expectedFile)
			}
		}
	})
}

func TestToolExecutionIntegration(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "slop-shop-tools-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Tool Execution System", func(t *testing.T) {
		// Test the ExecuteTools function with a mock response containing tool calls
		mockResponse := `Here's what I'll do:

RUN_COMMAND: echo "Hello from tool execution test"
READ_FILE: README.md
LIST_DIR: .

Let me execute these tools.`

		result := tools.ExecuteTools(mockResponse, tempDir)
		
		// Verify that the tool execution system processes the response
		if result == "" {
			t.Error("Expected non-empty result from ExecuteTools")
		}

		// Check that the result contains expected tool execution output
		if !strings.Contains(result, "Tool Execution Results") {
			t.Error("Expected 'Tool Execution Results' in output")
		}
	})

	t.Run("Tool Execution with File Operations", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := "This is test content"
		if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Test tool execution with file operations
		mockResponse := `I need to read a file:

READ_FILE: test.txt`

		result := tools.ExecuteTools(mockResponse, tempDir)
		
		// Verify that file reading was attempted
		if !strings.Contains(result, "Tool Execution Results") {
			t.Error("Expected tool execution results")
		}
	})
}

func TestStreamingResponseHandling(t *testing.T) {
	server := MockOllamaServer()
	defer server.Close()

	t.Run("Streaming Response Processing", func(t *testing.T) {
		var receivedChunks []string
		var chunkCount int
		
		chunkCallback := func(chunk string) {
			receivedChunks = append(receivedChunks, chunk)
			chunkCount++
		}

		response, err := ollama.SendToOllamaWithCallback(
			server.URL,
			"test-model",
			"Test streaming",
			"",
			0.7,
			0.9,
			false,
			chunkCallback,
		)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify we received multiple chunks
		if chunkCount < 2 {
			t.Errorf("Expected multiple chunks, got %d", chunkCount)
		}

		// Verify final response is complete
		expectedResponse := "Hello from mock Ollama"
		if response != expectedResponse {
			t.Errorf("Expected final response %q, got %q", expectedResponse, response)
		}

		// Verify chunks were received in order
		expectedChunks := []string{"Hello", " from", " mock", " Ollama"}
		if len(receivedChunks) != len(expectedChunks) {
			t.Errorf("Expected %d chunks, got %d", len(expectedChunks), len(receivedChunks))
		}

		for i, expected := range expectedChunks {
			if i < len(receivedChunks) && receivedChunks[i] != expected {
				t.Errorf("Chunk %d: expected %q, got %q", i, expected, receivedChunks[i])
			}
		}
	})
}

func TestEndToEndIntegration(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "slop-shop-e2e-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"main.go":   "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
		"README.md": "# Test Project\n\nThis is a test project.",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Mock Ollama server
	server := MockOllamaServer()
	defer server.Close()

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	t.Run("End-to-End Batch Mode", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run batch mode
		runBatch("Test prompt", "", server.URL, "test-model", 0.7, 0.9, false, tempDir)

		// Restore stdout and read output
		w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)

		// Verify output contains expected elements
		outputStr := string(output)
		if !strings.Contains(outputStr, "Starting with empty context") {
			t.Error("Expected 'Starting with empty context' message")
		}
		if !strings.Contains(outputStr, "Hello from mock Ollama") {
			t.Error("Expected mock Ollama response")
		}
	})

	t.Run("End-to-End with Repository Context", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run batch mode with repository context
		runBatch("Test prompt", "test context", server.URL, "test-model", 0.7, 0.9, false, tempDir)

		// Restore stdout and read output
		w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)

		// Verify output contains expected elements
		outputStr := string(output)
		// Check for repository scanning output (the exact message may vary)
		if !strings.Contains(outputStr, "files") && !strings.Contains(outputStr, "Repository") {
			t.Error("Expected repository scanning output")
		}
		if !strings.Contains(outputStr, "Hello from mock Ollama") {
			t.Error("Expected mock Ollama response")
		}
	})
}
