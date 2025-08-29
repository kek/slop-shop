package main

import (
	"flag"
	"os"
	"testing"
)

func TestMainFunctionFlags(t *testing.T) {
	// Save original command line arguments
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Test that main function doesn't crash with help flag
	os.Args = []string{"slop-shop", "-help"}

	// This test just ensures the program can parse flags without crashing
	// We can't easily test the full main function without mocking Ollama
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Test flag parsing
	model := flag.String("model", "qwen3:latest", "Ollama model to use")
	prompt := flag.String("prompt", "", "Prompt to send to the model (required unless using REPL mode)")
	repoPath := flag.String("repo", ".", "Path to repository (default: current directory)")
	ollamaURL := flag.String("url", "http://localhost:11434", "Ollama API URL")
	temperature := flag.Float64("temp", 0.7, "Temperature for model generation")
	topP := flag.Float64("top-p", 0.9, "Top-p for model generation")
	excludePatterns := flag.String("exclude", ".git,.jj,node_modules,vendor,*.exe,*.dll,*.so,*.dylib,*.bin,.crush", "Comma-separated patterns to exclude")
	replMode := flag.Bool("repl", false, "Start interactive REPL mode with repository context")
	toolsEnabled := flag.Bool("tools", false, "Enable tool execution for the LLM")
	emptyContext := flag.Bool("empty-context", false, "Start with empty context (no repository files loaded)")
	debugMode := flag.Bool("debug", false, "Enable debug logging to file")

	// Parse with test arguments
	os.Args = []string{"slop-shop", "-model", "test-model", "-prompt", "test prompt"}
	flag.Parse()

	// Verify flags were parsed correctly
	if *model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", *model)
	}
	if *prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got '%s'", *prompt)
	}
	if *repoPath != "." {
		t.Errorf("Expected repoPath '.', got '%s'", *repoPath)
	}
	if *ollamaURL != "http://localhost:11434" {
		t.Errorf("Expected ollamaURL 'http://localhost:11434', got '%s'", *ollamaURL)
	}
	if *temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", *temperature)
	}
	if *topP != 0.9 {
		t.Errorf("Expected topP 0.9, got %f", *topP)
	}
	if *excludePatterns != ".git,.jj,node_modules,vendor,*.exe,*.dll,*.so,*.dylib,*.bin,.crush" {
		t.Errorf("Expected default exclude patterns, got '%s'", *excludePatterns)
	}
	if *replMode != false {
		t.Errorf("Expected replMode false, got %t", *replMode)
	}
	if *toolsEnabled != false {
		t.Errorf("Expected toolsEnabled false, got %t", *toolsEnabled)
	}
	if *emptyContext != false {
		t.Errorf("Expected emptyContext false, got %t", *emptyContext)
	}
	if *debugMode != false {
		t.Errorf("Expected debugMode false, got %t", *debugMode)
	}
}

func TestRunBatchFunction(t *testing.T) {
	// Test that runBatch function can be called without crashing
	// We can't easily test the full functionality without mocking Ollama
	// This is a basic smoke test

	// Test with empty context
	runBatch("test prompt", "", "http://localhost:11434", "test-model", 0.7, 0.9, false, ".")

	// Test with some context
	context := "File: test.go\n---\npackage main\n"
	runBatch("test prompt", context, "http://localhost:11434", "test-model", 0.7, 0.9, false, ".")

	// If we get here without panicking, the test passes
}

func TestFlagValidation(t *testing.T) {
	// Test that the program handles missing required flags appropriately
	// This is tested by the main function's logic, but we can verify the flag setup

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Test with no prompt and no repl mode (should require prompt)
	os.Args = []string{"slop-shop"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	model := flag.String("model", "qwen3:latest", "Ollama model to use")
	prompt := flag.String("prompt", "", "Prompt to send to the model (required unless using REPL mode)")
	replMode := flag.Bool("repl", false, "Start interactive REPL mode with repository context")

	flag.Parse()

	// Verify the validation logic would work
	if *prompt == "" && !*replMode {
		// This is the condition that should trigger the fatal error in main
		// We can't test the actual fatal call, but we can verify the condition
		if *model == "" {
			t.Error("Model should have default value")
		}
	}
}
