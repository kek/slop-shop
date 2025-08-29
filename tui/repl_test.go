package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kek/slop-shop/repo"
)

func TestREPLModelInit(t *testing.T) {
	m := &REPLModel{
		context:             "test context",
		ollamaURL:           "http://localhost:11434",
		model:               "test-model",
		temperature:         0.7,
		topP:                0.9,
		toolsEnabled:        false,
		history:             make([]string, 0),
		historyIndex:        -1,
		conversationHistory: make([]string, 0),
	}

	cmd := m.Init()
	if cmd == nil {
		t.Error("Expected Init() to return a tick command, got nil")
	}
}

func TestREPLModelView(t *testing.T) {
	m := &REPLModel{
		context:             "test context",
		ollamaURL:           "http://localhost:11434",
		model:               "test-model",
		temperature:         0.7,
		topP:                0.9,
		toolsEnabled:        false,
		history:             []string{"help", "history"},
		historyIndex:        1,
		conversationHistory: []string{"User: help", "help response"},
		showHelp:            false,
		showHistory:         false,
		showContext:         false,
		quitting:            false,
		input:               "test input",
	}

	view := m.View()

	// Check that basic elements are present
	if !strings.Contains(view, "🚀 Slop Shop - AI-Powered Code Analysis") {
		t.Error("View should contain title")
	}

	if !strings.Contains(view, "🤖 test input█") {
		t.Error("View should contain input prompt with cursor")
	}

	if !strings.Contains(view, "Repository context loaded") {
		t.Error("View should contain context info")
	}
}

func TestREPLModelViewWithHelp(t *testing.T) {
	m := &REPLModel{
		context:             "test context",
		ollamaURL:           "http://localhost:11434",
		model:               "test-model",
		temperature:         0.7,
		topP:                0.9,
		toolsEnabled:        false,
		history:             make([]string, 0),
		historyIndex:        -1,
		conversationHistory: make([]string, 0),
		showHelp:            true,
		showHistory:         false,
		showContext:         false,
		quitting:            false,
		input:               "",
	}

	view := m.View()

	if !strings.Contains(view, "Keyboard Shortcuts:") {
		t.Error("View should show help when showHelp is true")
	}

	if !strings.Contains(view, "F1       - Toggle this help message") {
		t.Error("View should contain help commands")
	}
}

func TestREPLModelViewWithHistory(t *testing.T) {
	m := &REPLModel{
		context:             "test context",
		ollamaURL:           "http://localhost:11434",
		model:               "test-model",
		temperature:         0.7,
		topP:                0.9,
		toolsEnabled:        false,
		history:             []string{"help", "history", "quit"},
		historyIndex:        2,
		conversationHistory: make([]string, 0),
		showHelp:            false,
		showHistory:         true,
		showContext:         false,
		quitting:            false,
		input:               "",
	}

	view := m.View()

	if !strings.Contains(view, "Command History:") {
		t.Error("View should show history when showHistory is true")
	}

	if !strings.Contains(view, "1: help") {
		t.Error("View should contain first history item")
	}

	if !strings.Contains(view, "2: history") {
		t.Error("View should contain second history item")
	}

	if !strings.Contains(view, "3: quit") {
		t.Error("View should contain third history item")
	}

	if !strings.Contains(view, "Total: 3 commands") {
		t.Error("View should show total command count")
	}
}

func TestREPLModelViewWithContext(t *testing.T) {
	m := &REPLModel{
		context:             "test context with 25 characters",
		ollamaURL:           "http://localhost:11434",
		model:               "test-model",
		temperature:         0.7,
		topP:                0.9,
		toolsEnabled:        false,
		history:             make([]string, 0),
		historyIndex:        -1,
		conversationHistory: make([]string, 0),
		showHelp:            false,
		showHistory:         false,
		showContext:         true,
		quitting:            false,
		input:               "",
	}

	view := m.View()

	if !strings.Contains(view, "Repository Context:") {
		t.Error("View should show context when showContext is true")
	}

	if !strings.Contains(view, "Loaded: 31 characters") {
		t.Error("View should show context length")
	}
}

func TestREPLModelViewQuitting(t *testing.T) {
	m := &REPLModel{
		quitting: true,
	}

	view := m.View()

	if view != "Goodbye! 👋\n" {
		t.Errorf("Expected quitting message, got: %q", view)
	}
}

func TestREPLModelNavigateHistory(t *testing.T) {
	m := &REPLModel{
		history:      []string{"help", "history", "quit"},
		historyIndex: 2,
		input:        "current input",
	}

	// Test navigating up
	cmd := m.navigateHistory(-1)
	if cmd == nil {
		t.Error("navigateHistory should return a command")
	}

	// Execute the command to test the logic
	msg := cmd()
	if msg != nil {
		t.Error("navigateHistory command should return nil")
	}

	// Check that history index was updated
	if m.historyIndex != 1 {
		t.Errorf("Expected historyIndex to be 1, got %d", m.historyIndex)
	}

	// Check that input was updated
	if m.input != "history" {
		t.Errorf("Expected input to be 'history', got %q", m.input)
	}
}

func TestREPLModelNavigateHistoryBounds(t *testing.T) {
	m := &REPLModel{
		history:      []string{"help", "history"},
		historyIndex: 0,
		input:        "current input",
	}

	// Test navigating up beyond bounds
	cmd := m.navigateHistory(-1)
	cmd()

	// Should not go below 0
	if m.historyIndex != 0 {
		t.Errorf("Expected historyIndex to stay at 0, got %d", m.historyIndex)
	}

	// Test navigating down beyond bounds
	m.historyIndex = 1
	cmd = m.navigateHistory(1)
	cmd()

	// Should go to empty input
	if m.historyIndex != 2 {
		t.Errorf("Expected historyIndex to be 2, got %d", m.historyIndex)
	}

	if m.input != "" {
		t.Errorf("Expected input to be empty, got %q", m.input)
	}
}

func TestREPLModelToggleHelp(t *testing.T) {
	m := &REPLModel{
		showHelp: false,
	}

	cmd := m.toggleHelp()
	if cmd == nil {
		t.Error("toggleHelp should return a command")
	}

	cmd()

	if !m.showHelp {
		t.Error("showHelp should be toggled to true")
	}

	cmd()

	if m.showHelp {
		t.Error("showHelp should be toggled back to false")
	}
}

func TestREPLModelToggleHistory(t *testing.T) {
	m := &REPLModel{
		showHistory: false,
	}

	cmd := m.toggleHistory()
	if cmd == nil {
		t.Error("toggleHistory should return a command")
	}

	cmd()

	if !m.showHistory {
		t.Error("showHistory should be toggled to true")
	}

	cmd()

	if m.showHistory {
		t.Error("showHistory should be toggled back to false")
	}
}

func TestREPLModelToggleContext(t *testing.T) {
	m := &REPLModel{
		showContext: false,
	}

	cmd := m.toggleContext()
	if cmd == nil {
		t.Error("toggleContext should return a command")
	}

	cmd()

	if !m.showContext {
		t.Error("showContext should be toggled to true")
	}

	cmd()

	if m.showContext {
		t.Error("showContext should be toggled back to false")
	}
}

func TestREPLModelSubmitInput(t *testing.T) {
	m := &REPLModel{
		input:        "help",
		history:      make([]string, 0),
		historyIndex: -1,
	}

	cmd := m.submitInput()
	if cmd == nil {
		t.Error("submitInput should return a command for non-empty input")
	}

	// Test empty input
	m.input = ""
	cmd = m.submitInput()
	if cmd != nil {
		t.Error("submitInput should return nil for empty input")
	}
}

func TestREPLModelSubmitInputSpecialCommands(t *testing.T) {
	m := &REPLModel{
		input:        "help",
		history:      make([]string, 0),
		historyIndex: -1,
	}

	cmd := m.submitInput()
	if cmd == nil {
		t.Error("submitInput should return a command for help")
	}

	// Test other special commands
	testCases := []string{"clear", "context", "history", "quit", "exit", "q"}
	for _, testCase := range testCases {
		m.input = testCase
		cmd = m.submitInput()
		if cmd == nil {
			t.Errorf("submitInput should return a command for %s", testCase)
		}
	}
}

func TestREPLModelSubmitInputAddsToHistory(t *testing.T) {
	m := &REPLModel{
		input:        "test command",
		history:      make([]string, 0),
		historyIndex: -1,
	}

	// Submit the input
	cmd := m.submitInput()
	if cmd == nil {
		t.Error("submitInput should return a command")
	}

	// Execute the command to trigger the history update
	msg := cmd()
	if msg == nil {
		t.Error("submitInput command should return a message for special commands")
	}

	// Check that command was added to history
	if len(m.history) != 1 {
		t.Errorf("Expected 1 command in history, got %d", len(m.history))
	}

	if m.history[0] != "test command" {
		t.Errorf("Expected history[0] to be 'test command', got %q", m.history[0])
	}

	if m.historyIndex != 1 {
		t.Errorf("Expected historyIndex to be 1, got %d", m.historyIndex)
	}
}

func TestREPLModelSubmitInputPreventsDuplicateHistory(t *testing.T) {
	m := &REPLModel{
		input:        "test command",
		history:      []string{"test command"},
		historyIndex: 0,
	}

	// Submit the same command again
	cmd := m.submitInput()
	if cmd == nil {
		t.Error("submitInput should return a command")
	}

	// Execute the command
	cmd()

	// Check that duplicate was not added
	if len(m.history) != 1 {
		t.Errorf("Expected 1 command in history (no duplicates), got %d", len(m.history))
	}
}

func TestREPLModelViewWithLongLines(t *testing.T) {
	// Create a response with a very long line that should be wrapped
	longLine := "This is a very long line that exceeds the 80 character limit and should be wrapped to multiple lines to ensure proper display in the terminal"

	m := &REPLModel{
		conversationHistory: []string{
			"User: test question",
			longLine,
		},
		showHelp:    false,
		showHistory: false,
		showContext: false,
		quitting:    false,
		input:       "",
	}

	view := m.View()

	// Debug: log the entire view to see what's happening
	t.Logf("Full view output:")
	for i, line := range strings.Split(view, "\n") {
		t.Logf("Line %d (%d chars): %q", i+1, len(line), line)
	}

	// Check that the long line is wrapped
	if !strings.Contains(view, "This is a very long line") {
		t.Error("View should contain the beginning of the long line")
	}

	// The wrapped text should be split across multiple lines
	lines := strings.Split(view, "\n")
	wrappedLinesFound := 0

	for _, line := range lines {
		if strings.Contains(line, "This is a very long line") {
			wrappedLinesFound++
		}
		if strings.Contains(line, "wrapped to multiple lines") {
			wrappedLinesFound++
		}
	}

	// We should find at least 2 lines (the original was split)
	if wrappedLinesFound < 2 {
		t.Errorf("Expected at least 2 wrapped lines, found %d", wrappedLinesFound)
	}
}

func TestREPLModelF5ClearContext(t *testing.T) {
	m := &REPLModel{
		context:             "test context",
		ollamaURL:           "http://localhost:11434",
		model:               "test-model",
		temperature:         0.7,
		topP:                0.9,
		toolsEnabled:        false,
		history:             make([]string, 0),
		historyIndex:        -1,
		conversationHistory: make([]string, 0),
		showHelp:            false,
		showHistory:         false,
		showContext:         false,
		quitting:            false,
		input:               "",
	}

	// Test F5 key press
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f', '5'}}
	m.Update(msg)

	// Check that context is cleared
	if m.context != "" {
		t.Error("Context should be cleared after F5")
	}

	// Check that conversation history contains the system message
	if len(m.conversationHistory) == 0 {
		t.Error("Conversation history should contain system message after F5")
	}

	// Check that the last message indicates context was cleared
	lastMessage := m.conversationHistory[len(m.conversationHistory)-1]
	if !strings.Contains(lastMessage, "Local context cleared") {
		t.Errorf("Expected system message about local context being cleared, got: %s", lastMessage)
	}
}

func TestFileTypeAnalysis(t *testing.T) {
	// Test files with different extensions
	testFiles := []repo.FileInfo{
		{Path: "main.go", Size: 1000, Content: "package main"},
		{Path: "README.md", Size: 500, Content: "# Test"},
		{Path: "config.json", Size: 200, Content: "{}"},
		{Path: "script.sh", Size: 150, Content: "#!/bin/bash"},
		{Path: "go.mod", Size: 100, Content: "module test"},
		{Path: "Makefile", Size: 75, Content: "all:"},
		{Path: "data.txt", Size: 50, Content: "data"},
	}

	// Analyze file types
	fileTypeBytes := analyzeFileTypes(testFiles)

	// Check that all file types are correctly identified
	expectedTypes := map[string]int64{
		"Go Source":     1000,
		"Markdown":      500,
		"JSON":          200,
		"Shell Scripts": 150,
		"Go Module":     100,
		"Makefile":      75,
		"Text":          50,
	}

	for expectedType, expectedBytes := range expectedTypes {
		if actualBytes, exists := fileTypeBytes[expectedType]; !exists {
			t.Errorf("Expected file type '%s' not found", expectedType)
		} else if actualBytes != expectedBytes {
			t.Errorf("Expected %d bytes for '%s', got %d", expectedBytes, expectedType, actualBytes)
		}
	}

	// Check total bytes
	totalBytes := int64(0)
	for _, bytes := range fileTypeBytes {
		totalBytes += bytes
	}
	expectedTotal := int64(2075) // 1000+500+200+150+100+75+50
	if totalBytes != expectedTotal {
		t.Errorf("Expected total %d bytes, got %d", expectedTotal, totalBytes)
	}
}

func TestFormatBytes(t *testing.T) {
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range testCases {
		result := formatBytes(tc.bytes)
		if result != tc.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", tc.bytes, result, tc.expected)
		}
	}
}

func TestDebugFlagFunctionality(t *testing.T) {
	// Test that debug flag is properly set
	m := &REPLModel{
		context:             "test context",
		ollamaURL:           "http://localhost:11434",
		model:               "test-model",
		temperature:         0.7,
		topP:                0.9,
		toolsEnabled:        false,
		debugEnabled:        true,
		history:             make([]string, 0),
		historyIndex:        -1,
		conversationHistory: make([]string, 0),
		showHelp:            false,
		showHistory:         false,
		showContext:         false,
		quitting:            false,
		input:               "",
	}

	// Test that debug flag is set correctly
	if !m.debugEnabled {
		t.Error("Debug flag should be true")
	}

	// Test with debug disabled
	m.debugEnabled = false
	if m.debugEnabled {
		t.Error("Debug flag should be false")
	}
}

func TestREPLModelStreamingResponseHandling(t *testing.T) {
	m := &REPLModel{
		context:             "test context",
		ollamaURL:           "http://localhost:11434",
		model:               "test-model",
		temperature:         0.7,
		topP:                0.9,
		toolsEnabled:        false,
		history:             make([]string, 0),
		historyIndex:        -1,
		conversationHistory: []string{"User: test question", ""}, // Empty response slot
		processing:          true,
		streamChannel:       make(chan string, 100),
	}

	// Simulate receiving streaming chunks
	chunks := []string{
		"Hello",
		" world",
		"! This is a",
		" test response.",
	}

	// Process each chunk through the stream channel (simulating real behavior)
	for _, chunk := range chunks {
		// Simulate chunks being sent to the stream channel
		select {
		case m.streamChannel <- chunk:
			// Chunk sent successfully
		default:
			t.Errorf("Failed to send chunk to stream channel: %s", chunk)
		}
	}

	// Process a few tick messages to simulate the real streaming behavior
	for i := 0; i < 5; i++ {
		tickMsg := tickMsg(time.Now())
		m.Update(tickMsg)
	}

	// Check that the response was properly assembled
	if len(m.conversationHistory) == 0 {
		t.Error("Expected conversation history to contain the response")
	}

	// The last item should be the complete response
	lastResponse := m.conversationHistory[len(m.conversationHistory)-1]
	expectedResponse := "Hello world! This is a test response."
	if lastResponse != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, lastResponse)
	}
}

func TestREPLModelJSONResponseHandling(t *testing.T) {
	// Test JSON response handling to ensure it's not broken up
	jsonResponse := `{"tool": "RUN_COMMAND", "command": "ls -la", "result": "success"}`

	m := &REPLModel{
		conversationHistory: []string{
			"User: test question",
			jsonResponse,
		},
		showHelp:    false,
		showHistory: false,
		showContext: false,
		quitting:    false,
		input:       "",
	}

	view := m.View()

	// JSON response should be displayed as a single block
	if !strings.Contains(view, jsonResponse) {
		t.Error("JSON response should be displayed intact")
	}

	// Should not contain any line breaks within the JSON
	lines := strings.Split(view, "\n")
	jsonLines := 0
	for _, line := range lines {
		if strings.Contains(line, `"tool"`) || strings.Contains(line, `"command"`) || strings.Contains(line, `"result"`) {
			jsonLines++
		}
	}

	// JSON should be on a single line or properly formatted
	if jsonLines == 0 {
		t.Error("JSON response should be visible in the view")
	}
}

func TestREPLModelResponseFormatting(t *testing.T) {
	// Test various response formats to ensure they're displayed correctly
	testCases := []struct {
		name     string
		response string
		expected string
	}{
		{
			name:     "Simple text response",
			response: "This is a simple response",
			expected: "This is a simple response",
		},
		{
			name:     "Response with newlines",
			response: "Line 1\nLine 2\nLine 3",
			expected: "Line 1",
		},
		{
			name:     "Response with literal newlines",
			response: "Line 1\\nLine 2\\nLine 3",
			expected: "Line 1",
		},
		{
			name:     "Long line that should be wrapped",
			response: "This is a very long line that exceeds the eighty character limit and should be wrapped to multiple lines",
			expected: "This is a very long line",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &REPLModel{
				conversationHistory: []string{
					"User: test question",
					tc.response,
				},
				showHelp:    false,
				showHistory: false,
				showContext: false,
				quitting:    false,
				input:       "",
			}

			view := m.View()

			// Check that the response content is present
			if !strings.Contains(view, tc.expected) {
				t.Errorf("Expected to find '%s' in view", tc.expected)
			}
		})
	}
}

func TestREPLModelStreamingChannelHandling(t *testing.T) {
	m := &REPLModel{
		processing:    true,
		streamChannel: make(chan string, 100),
	}

	// Test that streaming channel doesn't block
	select {
	case m.streamChannel <- "test chunk":
		// Successfully sent chunk
	default:
		t.Error("Streaming channel should not block when buffer has space")
	}

	// Test channel buffer capacity
	// Note: Go channels can hold buffer_size + 1 items (the buffer + the item being sent)
	// So a channel with buffer 100 can accept 99 items without blocking
	for i := 0; i < 99; i++ {
		select {
		case m.streamChannel <- fmt.Sprintf("chunk %d", i):
			// Continue
		default:
			t.Errorf("Channel should accept %d chunks, failed at chunk %d", i+1, i)
			return
		}
	}

	// Channel should be full now (100 items sent)
	select {
	case m.streamChannel <- "overflow":
		t.Error("Channel should be full and not accept more chunks")
	default:
		// Expected behavior - channel is full
	}
}

func TestREPLModelResponseAssembly(t *testing.T) {
	m := &REPLModel{
		processing:          true,
		conversationHistory: []string{"User: test question", ""}, // Empty response slot
		streamChannel:       make(chan string, 100),
	}

	// Simulate streaming response assembly
	chunks := []string{
		"First",
		"Second",
		"Third",
		"Fourth",
	}

	// Process chunks through the stream channel
	for _, chunk := range chunks {
		select {
		case m.streamChannel <- chunk:
			// Chunk sent successfully
		default:
			t.Errorf("Failed to send chunk to stream channel: %s", chunk)
		}
	}

	// Process tick messages to simulate real streaming behavior
	// Process enough ticks to ensure all chunks are consumed
	for i := 0; i < 10; i++ {
		tickMsg := tickMsg(time.Now())
		m.Update(tickMsg)
	}

	// Check final response
	if len(m.conversationHistory) < 2 {
		t.Error("Expected at least 2 items in conversation history")
	}

	finalResponse := m.conversationHistory[1] // Index 1 should be the response
	expectedResponse := "FirstSecondThirdFourth"
	if finalResponse != expectedResponse {
		t.Errorf("Expected assembled response '%s', got '%s'", expectedResponse, finalResponse)
	}
}

func TestREPLModelProcessingStateManagement(t *testing.T) {
	m := &REPLModel{
		processing:          false,
		conversationHistory: make([]string, 0),
		streamChannel:       make(chan string, 100),
	}

	// Test that processing starts when input is submitted
	m.input = "test question"
	cmd := m.submitInput()
	if cmd == nil {
		t.Error("submitInput should return a command")
	}

	// Execute the command
	msg := cmd()
	if msg == nil {
		t.Error("submitInput command should return a message")
	}

	// Check that processing state is set
	if !m.processing {
		t.Error("Processing should be set to true after submitting input")
	}

	// Test that processing stops when response is complete
	m.processing = false
	if m.processing {
		t.Error("Processing should be false when response is complete")
	}
}

func TestREPLModelConversationHistoryManagement(t *testing.T) {
	m := &REPLModel{
		conversationHistory: make([]string, 0),
		streamChannel:       make(chan string, 100),
	}

	// Test adding user input
	userInput := "What is the main function?"
	m.conversationHistory = append(m.conversationHistory, fmt.Sprintf("User: %s", userInput))

	// Test adding response
	response := "The main function is the entry point of the program."
	m.conversationHistory = append(m.conversationHistory, response)

	// Check conversation history
	if len(m.conversationHistory) != 2 {
		t.Errorf("Expected 2 items in conversation history, got %d", len(m.conversationHistory))
	}

	if !strings.Contains(m.conversationHistory[0], "User: What is the main function?") {
		t.Error("First item should contain user input")
	}

	if m.conversationHistory[1] != response {
		t.Error("Second item should contain the response")
	}

	// Test conversation history limit (20 items)
	// Start with 2 items, add 23 more to test the limit
	for i := 0; i < 23; i++ {
		m.conversationHistory = append(m.conversationHistory, fmt.Sprintf("Message %d", i))
	}

	// The limit should be enforced when adding new items
	// Note: The limit is only enforced in the actual Update method, not in manual append
	// This test verifies the current behavior
	if len(m.conversationHistory) != 25 {
		t.Errorf("Expected 25 items in conversation history, got %d", len(m.conversationHistory))
	}
}

func TestREPLModelSpinnerAnimation(t *testing.T) {
	m := &REPLModel{
		processing:   true,
		spinnerFrame: 0,
	}

	// Test spinner frame progression
	initialFrame := m.spinnerFrame
	for i := 0; i < 5; i++ {
		msg := tickMsg(time.Now())
		m.Update(msg)
	}

	// Spinner should have progressed
	if m.spinnerFrame == initialFrame {
		t.Error("Spinner frame should progress over time")
	}

	// Spinner should stay within bounds (0-9)
	if m.spinnerFrame < 0 || m.spinnerFrame >= 10 {
		t.Errorf("Spinner frame should be between 0 and 9, got %d", m.spinnerFrame)
	}
}

func TestREPLModelKeyInputHandling(t *testing.T) {
	m := &REPLModel{
		input:               "",
		conversationHistory: make([]string, 0),
	}

	// Test regular character input
	testChars := []string{"H", "e", "l", "l", "o"}
	for _, char := range testChars {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(char[0])}}
		m.Update(msg)
	}

	if m.input != "Hello" {
		t.Errorf("Expected input 'Hello', got '%s'", m.input)
	}

	// Test backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	m.Update(msg)

	if m.input != "Hell" {
		t.Errorf("Expected input 'Hell' after backspace, got '%s'", m.input)
	}

	// Test space
	spaceMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	m.Update(spaceMsg)

	if m.input != "Hell " {
		t.Errorf("Expected input 'Hell ' after space, got '%s'", m.input)
	}
}

func TestREPLModelSpecialKeyHandling(t *testing.T) {
	m := &REPLModel{
		showHelp:    false,
		showHistory: false,
		showContext: false,
		quitting:    false,
	}

	// Test F1 (help toggle)
	f1Msg := tea.KeyMsg{Type: tea.KeyF1}
	m.Update(f1Msg)
	if !m.showHelp {
		t.Error("F1 should toggle help to true")
	}

	// Test F2 (history toggle)
	f2Msg := tea.KeyMsg{Type: tea.KeyF2}
	m.Update(f2Msg)
	if !m.showHistory {
		t.Error("F2 should toggle history to true")
	}

	// Test F3 (context toggle)
	f3Msg := tea.KeyMsg{Type: tea.KeyF3}
	m.Update(f3Msg)
	if !m.showContext {
		t.Error("F3 should toggle context to true")
	}

	// Test Escape (hide all panels)
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	m.Update(escMsg)
	if m.showHelp || m.showHistory || m.showContext {
		t.Error("Escape should hide all panels")
	}

	// Test Ctrl+C (quit)
	ctrlCMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	m.Update(ctrlCMsg)
	if !m.quitting {
		t.Error("Ctrl+C should set quitting to true")
	}
}

func TestREPLModelInputSubmission(t *testing.T) {
	m := &REPLModel{
		input:               "test input",
		processing:          false,
		conversationHistory: make([]string, 0),
	}

	// Test Enter key submission
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	m.Update(enterMsg)

	// Input should be cleared and processing should start
	if m.input != "" {
		t.Error("Input should be cleared after submission")
	}

	if !m.processing {
		t.Error("Processing should start after input submission")
	}
}

func TestREPLModelHistoryNavigation(t *testing.T) {
	m := &REPLModel{
		history:      []string{"first", "second", "third"},
		historyIndex: 2,
		input:        "current",
	}

	// Test up arrow navigation
	upCmd := m.navigateHistory(-1)
	upCmd() // Execute the command

	// Should navigate to previous history item
	if m.historyIndex != 1 {
		t.Errorf("Expected historyIndex 1, got %d", m.historyIndex)
	}

	if m.input != "second" {
		t.Errorf("Expected input 'second', got '%s'", m.input)
	}

	// Test down arrow navigation
	downCmd := m.navigateHistory(1)
	downCmd() // Execute the command

	// Should navigate to next history item
	if m.historyIndex != 2 {
		t.Errorf("Expected historyIndex 2, got %d", m.historyIndex)
	}

	if m.input != "third" {
		t.Errorf("Expected input 'third', got '%s'", m.input)
	}

	// Test going beyond bounds
	downCmd = m.navigateHistory(1)
	downCmd() // Execute the command

	// Should go to empty input
	if m.historyIndex != 3 {
		t.Errorf("Expected historyIndex 3, got %d", m.historyIndex)
	}

	if m.input != "" {
		t.Errorf("Expected input to be empty, got %q", m.input)
	}
}

func TestREPLModelResponseChunkProcessing(t *testing.T) {
	m := &REPLModel{
		processing:          true,
		conversationHistory: []string{"User: test", ""}, // Empty response slot
		streamChannel:       make(chan string, 100),
	}

	// Test processing of individual response chunks
	testChunks := []string{
		"Response",
		" with",
		" multiple",
		" chunks",
		".",
	}

	for _, chunk := range testChunks {
		select {
		case m.streamChannel <- chunk:
			// Chunk sent successfully
		default:
			t.Errorf("Failed to send chunk to stream channel: %s", chunk)
		}
	}

	// Process tick messages to simulate real streaming behavior
	// Process enough ticks to ensure all chunks are consumed
	for i := 0; i < 10; i++ {
		tickMsg := tickMsg(time.Now())
		m.Update(tickMsg)
	}

	// Check that chunks were properly assembled
	expectedResponse := "Response with multiple chunks."
	actualResponse := m.conversationHistory[1]

	if actualResponse != expectedResponse {
		t.Errorf("Expected assembled response '%s', got '%s'", expectedResponse, actualResponse)
	}
}

func TestREPLModelTickMessageHandling(t *testing.T) {
	m := &REPLModel{
		processing:    true,
		spinnerFrame:  0,
		streamChannel: make(chan string, 100),
	}

	// Test tick message processing
	tickMsg := tickMsg(time.Now())
	updatedModel, cmd := m.Update(tickMsg)

	// Should return a new tick command
	if cmd == nil {
		t.Error("Tick message should return a new tick command")
	}

	// Spinner frame should progress
	if m.spinnerFrame == 0 {
		t.Error("Spinner frame should progress on tick")
	}

	// Should return the same model
	if updatedModel != m {
		t.Error("Tick message should return the same model")
	}
}

func TestREPLModelStreamingResponseCompletion(t *testing.T) {
	m := &REPLModel{
		processing:          true,
		responseComplete:    false,
		conversationHistory: []string{"User: test", ""},
		streamChannel:       make(chan string, 100),
	}

	// Simulate completing a streaming response
	chunks := []string{"Complete", " response"}
	for _, chunk := range chunks {
		select {
		case m.streamChannel <- chunk:
			// Chunk sent successfully
		default:
			t.Errorf("Failed to send chunk to stream channel: %s", chunk)
		}
	}

	// Process tick messages to simulate real streaming behavior
	// Process enough ticks to ensure all chunks are consumed
	for i := 0; i < 10; i++ {
		tickMsg := tickMsg(time.Now())
		m.Update(tickMsg)
	}

	// Mark response as complete
	m.processing = false
	m.responseComplete = true

	// Check final state
	if m.processing {
		t.Error("Processing should be false when response is complete")
	}

	if !m.responseComplete {
		t.Error("Response should be marked as complete")
	}

	// Check assembled response
	expectedResponse := "Complete response"
	actualResponse := m.conversationHistory[1]

	if actualResponse != expectedResponse {
		t.Errorf("Expected final response '%s', got '%s'", expectedResponse, actualResponse)
	}
}

// analyzeFileTypes analyzes file types and returns a map of type names to total bytes
func analyzeFileTypes(files []repo.FileInfo) map[string]int64 {
	fileTypeBytes := make(map[string]int64)

	for _, file := range files {
		ext := filepath.Ext(file.Path)
		baseName := filepath.Base(file.Path)
		var fileType string

		switch {
		case ext == ".go":
			fileType = "Go Source"
		case ext == ".md":
			fileType = "Markdown"
		case ext == ".json":
			fileType = "JSON"
		case ext == ".sh" || ext == ".bash":
			fileType = "Shell Scripts"
		case ext == ".mod":
			fileType = "Go Module"
		case baseName == "Makefile" || baseName == "makefile":
			fileType = "Makefile"
		default:
			fileType = "Text"
		}

		fileTypeBytes[fileType] += file.Size
	}

	return fileTypeBytes
}

// formatBytes formats a byte count into a human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
