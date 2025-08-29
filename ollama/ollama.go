package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Request represents the request structure for Ollama API
type Request struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Stream  bool    `json:"stream"`
	Options Options `json:"options,omitempty"`
}

// Options represents additional options for Ollama
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
}

// Response represents the response from Ollama API
type Response struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// SendToOllamaWithCallback sends the request to Ollama API with streaming support and optional callback
func SendToOllamaWithCallback(url, model, prompt, context string, temperature, topP float64, toolsEnabled bool, chunkCallback func(string)) (string, error) {
	// Combine context and prompt
	fullPrompt := context + "\n\nUser Question: " + prompt

	if toolsEnabled {
		fullPrompt = addToolInstructions(fullPrompt)
	}

	// Prepare the request
	request := Request{
		Model:  model,
		Prompt: fullPrompt,
		Stream: true, // Enable streaming
		Options: Options{
			Temperature: temperature,
			TopP:        topP,
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	// Send HTTP request
	resp, err := http.Post(url+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Handle streaming response
	var fullResponse strings.Builder
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("error reading streaming response: %v", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse each JSON line
		var ollamaResp Response
		if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
			continue // Skip malformed lines
		}

		// Collect the response chunk
		if ollamaResp.Response != "" {
			fullResponse.WriteString(ollamaResp.Response)

			// If callback is provided, stream the chunk in real-time
			if chunkCallback != nil {
				chunkCallback(ollamaResp.Response)
			}
		}

		// Check if response is complete
		if ollamaResp.Done {
			break
		}
	}

	return fullResponse.String(), nil
}

// addToolInstructions adds tool execution instructions to the prompt
func addToolInstructions(prompt string) string {
	toolInstructions := `

AVAILABLE TOOLS:
You can use the following tools by including them in your response:

1. RUN_COMMAND: Execute a shell command
   Format: RUN_COMMAND: <command>
   Example: RUN_COMMAND: ls -la
   Example: RUN_COMMAND: go build -o test
   Example: RUN_COMMAND: git status

2. READ_FILE: Read the contents of a file
   Format: READ_FILE: <filepath>
   Example: READ_FILE: main.go
   Example: READ_FILE: README.md

3. LIST_DIR: List contents of a directory
   Format: LIST_DIR: <directory>
   Example: LIST_DIR: .
   Example: LIST_DIR: src/

4. TEST_COMMAND: Test if a command works
   Format: TEST_COMMAND: <command>
   Example: TEST_COMMAND: go version
   Example: TEST_COMMAND: python3 --version

5. SEARCH_FILES: Search for text in files
   Format: SEARCH_FILES: <pattern> <directory>
   Example: SEARCH_FILES: "func main" .
   Example: SEARCH_FILES: "import" src/

6. GENERATE_DIFF: Generate a unified diff for suggested changes
   Format: GENERATE_DIFF: <description of changes>
   Example: GENERATE_DIFF: Add error handling to main function
   Example: GENERATE_DIFF: Update README with new features

7. APPLY_DIFF: Apply a unified diff to the repository
   Format: APPLY_DIFF: <unified diff content>
   Example: APPLY_DIFF: --- a/file.txt\n+++ b/file.txt\n@@ -1,3 +1,4 @@\n line1\n+new line\n line2\n line3

8. CREATE_FILE: Create a new file with specified content
   Format: CREATE_FILE: <filepath>
   <content>
   END_FILE
   
   Example: CREATE_FILE: newfile.txt
   This is the content of the new file
   END_FILE
   
   Example: CREATE_FILE: docs/README.md
   # Documentation
   
   This is a new documentation file.
   END_FILE

CRITICAL INSTRUCTIONS FOR TOOL USAGE:
- You MUST use these tools to accomplish the user's request
- Do NOT just describe what you would do - actually DO it using the tools
- Start by examining the current state using READ_FILE, LIST_DIR, or SEARCH_FILES
- Then use GENERATE_DIFF to create the necessary changes
- Finally use APPLY_DIFF to implement those changes
- Each tool call must be on a separate line with the exact format shown above
- Do NOT mix tool calls with other output
- You can use multiple tools in one response, but each tool call should be on a separate line
- After using tools, you can analyze the results and provide insights or suggestions

WORKFLOW FOR FILE MODIFICATIONS:
1. First, examine the current files using READ_FILE or SEARCH_FILES
2. Use GENERATE_DIFF to create the changes needed
3. Use APPLY_DIFF to implement those changes
4. Verify the changes worked as expected

User request: ` + prompt

	return prompt + toolInstructions
}
