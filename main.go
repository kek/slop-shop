package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/kek/slop-shop/ollama"
	"github.com/kek/slop-shop/repo"
	"github.com/kek/slop-shop/styles"
	"github.com/kek/slop-shop/tools"
	"github.com/kek/slop-shop/tui"
)

func main() {
	// Parse command line flags
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

	flag.Parse()

	// Set global debug flag
	tui.SetGlobalDebug(*debugMode)

	if *prompt == "" && !*replMode {
		log.Fatal("Error: -prompt flag is required unless using -repl mode")
	}

	// Parse exclude patterns
	excludeList := strings.Split(*excludePatterns, ",")
	for i, pattern := range excludeList {
		excludeList[i] = strings.TrimSpace(pattern)
	}

	// Read repository contents (unless empty context is requested)
	var context string
	if *emptyContext {
		context = ""
	} else {
		files, err := repo.ReadRepository(*repoPath, excludeList)
		if err != nil {
			log.Fatalf("Error reading repository: %v", err)
		}

		// Create context from repository contents
		context = repo.CreateContext(files)
	}

	// Handle chat mode or batch mode
	if *replMode {
		tui.StartChat(*ollamaURL, *model, context, *temperature, *topP, *toolsEnabled, *debugMode)
	} else {
		runBatch(*prompt, context, *ollamaURL, *model, *temperature, *topP, *toolsEnabled, *repoPath)
	}
}

// runBatch handles the single-prompt mode without Bubble Tea
func runBatch(prompt, context, ollamaURL, model string, temperature, topP float64, toolsEnabled bool, repoPath string) {
	fmt.Println(styles.TitleStyle.Render("🚀 Slop Shop - AI-Powered Code Analysis"))
	fmt.Println(styles.InfoStyle.Render(fmt.Sprintf("Reading repository at: %s", repoPath)))
	fmt.Println(styles.InfoStyle.Render(fmt.Sprintf("Using model: %s", model)))
	fmt.Println(styles.InfoStyle.Render(fmt.Sprintf("Prompt: %s", prompt)))
	fmt.Println(styles.InfoStyle.Render(fmt.Sprintf("Ollama URL: %s", ollamaURL)))

	if context != "" {
		fmt.Println(styles.SuccessStyle.Render(fmt.Sprintf("Found %d files", strings.Count(context, "File:"))))
		fmt.Println(styles.InfoStyle.Render(fmt.Sprintf("Total context size: %d characters", len(context))))
	} else {
		fmt.Println(styles.InfoStyle.Render("Starting with empty context (no repository files loaded)"))
	}

	fmt.Print(styles.PromptStyle.Render("🤖 "))

	// Channel for streaming response chunks
	streamChannel := make(chan string, 100)
	var response strings.Builder

	go func() {
		_, err := ollama.SendToOllamaWithCallback(ollamaURL, model, prompt, context, temperature, topP, toolsEnabled, func(chunk string) {
			streamChannel <- chunk
		})
		if err != nil {
			// Send error message to channel instead of silently failing
			streamChannel <- fmt.Sprintf("\n❌ Error: %v\n", err)
		}
		close(streamChannel)
	}()

	for chunk := range streamChannel {
		fmt.Print(chunk)
		response.WriteString(chunk)
	}

	fmt.Println()

	if toolsEnabled {
		tools.ExecuteTools(response.String(), repoPath)
	}
}