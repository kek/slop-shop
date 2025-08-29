package tui

import (
	"fmt"
	
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kek/slop-shop/ollama"
	"github.com/kek/slop-shop/styles"
)

// REPLModel represents the Bubble Tea model for the REPL
type REPLModel struct {
	input               string
	history             []string
	historyIndex        int
	context             string
	ollamaURL           string
	model               string
	temperature         float64
	topP                float64
	toolsEnabled        bool
	debugEnabled        bool
	conversationHistory []string
	showHelp            bool
	showHistory         bool
	showContext         bool
	quitting            bool
	processing          bool
	spinnerFrame        int
	responseBuffer      strings.Builder
	responseComplete    bool
	streamChannel       chan string // Channel for streaming response chunks
}

// REPLMsg represents messages for the REPL
type REPLMsg interface{}

type inputMsg string

// Custom message types for the REPL
type tickMsg time.Time
type inputSubmittedMsg struct {
	input string
}
type processingCompleteMsg struct{}
type ollamaResponseMsg struct {
	response string
	err      error
}
type ollamaRequestMsg struct {
	input string
}
type ollamaStreamMsg struct {
	chunk string
}
type ollamaDoneMsg struct{}

// StartChat starts an interactive chat session with the repository context
func StartChat(url, model, context string, temperature, topP float64, toolsEnabled, debugEnabled bool) {
	logToFile("Starting REPL...")

	// Create the REPL model
	m := &REPLModel{
		context:             context,
		ollamaURL:           url,
		model:               model,
		temperature:         temperature,
		topP:                topP,
		toolsEnabled:        toolsEnabled,
		debugEnabled:        debugEnabled,
		history:             make([]string, 0),
		historyIndex:        -1,
		conversationHistory: make([]string, 0),
		processing:          false,
		spinnerFrame:        0,
		responseBuffer:      strings.Builder{},
		responseComplete:    false,
		streamChannel:       make(chan string, 100), // Buffer for streaming chunks
	}

	logToFile("Model created, starting Bubble Tea program...")

	// Create and run the Bubble Tea program
	logToFile("About to create program...")
	p := tea.NewProgram(m) // Removed tea.WithAltScreen() to fix display issues
	logToFile("Program created, running...")

	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			logToFile(fmt.Sprintf("Panic recovered: %v", r))
		}
	}()

	logToFile("About to call p.Run()...")
	if _, err := p.Run(); err != nil {
		logToFile(fmt.Sprintf("Error running REPL: %v", err))
	}
	logToFile("REPL finished.")
}

// Init initializes the REPL model
func (m *REPLModel) Init() tea.Cmd {
	logToFile("Init() called")
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m *REPLModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logToFile(fmt.Sprintf("Update() called with message type: %T", msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		logToFile(fmt.Sprintf("Key pressed: '%s' (type: %T)", key, msg))

		switch key {
		case "ctrl+c":
			logToFile("Ctrl+C detected, quitting...")
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.input != "" {
				logToFile(fmt.Sprintf("Enter pressed with input: '%s'", m.input))
				return m, m.submitInput()
			}
		case "up":
			logToFile("Up arrow pressed")
			return m, m.navigateHistory(-1)
		case "down":
			logToFile("Down arrow pressed")
			return m, m.navigateHistory(1)
		case "f1":
			logToFile("F1 pressed, toggling help")
			m.showHelp = !m.showHelp
		case "f2":
			logToFile("F2 pressed, toggling history")
			m.showHistory = !m.showHistory
		case "f3":
			logToFile("F3 pressed, toggling context")
			m.showContext = !m.showContext
		case "f4":
			logToFile("F4 pressed, clearing conversation")
			m.conversationHistory = nil
		case "f5":
			logToFile("F5 pressed, clearing context")
			m.context = ""
			m.conversationHistory = append(m.conversationHistory, "System: Local context cleared. Note: Ollama internal context persists - restart Ollama for complete reset.")
		case "f10":
			logToFile("F10 pressed, quitting...")
			m.quitting = true
			return m, tea.Quit
		case "esc":
			logToFile("Escape pressed, hiding panels")
			m.showHelp = false
			m.showHistory = false
			m.showContext = false
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				logToFile("Backspace pressed, input length now: " + fmt.Sprint(len(m.input)))
			}
		case "space":
			m.input += " "
			logToFile("Space pressed, input length now: " + fmt.Sprint(len(m.input)))
		default:
			// Handle regular character input (including space)
			if len(key) == 1 {
				// Check if it's a printable character (including space)
				r := rune(key[0])
				if r >= 32 && r <= 126 {
					m.input += key
					logToFile(fmt.Sprintf("Character '%s' added, input length now: %d", key, len(m.input)))
				} else {
					logToFile(fmt.Sprintf("Non-printable character ignored: '%s' (rune: %d)", key, r))
				}
			} else {
				logToFile(fmt.Sprintf("Multi-character key ignored: '%s'", key))
			}
		}
	case ollamaResponseMsg:
		if msg.err != nil {
			// Add error message to conversation history
			m.conversationHistory = append(m.conversationHistory, fmt.Sprintf("User: %s", m.input))
			m.conversationHistory = append(m.conversationHistory, fmt.Sprintf("❌ Error: %v", msg.err))
			if len(m.conversationHistory) > 20 {
				m.conversationHistory = m.conversationHistory[len(m.conversationHistory)-20:]
			}
		} else {
			m.conversationHistory = append(m.conversationHistory, fmt.Sprintf("User: %s", m.input))
			m.conversationHistory = append(m.conversationHistory, msg.response)
			if len(m.conversationHistory) > 20 {
				m.conversationHistory = m.conversationHistory[len(m.conversationHistory)-20:]
			}
		}
		m.input = ""
	case inputSubmittedMsg:
		// Input was submitted, add to conversation history
		m.conversationHistory = append(m.conversationHistory, fmt.Sprintf("User: %s", msg.input))
		// Keep processing = true so spinner shows until we get a response
	case ollamaRequestMsg:
		// Actually call Ollama and keep processing true until response arrives
		input := msg.input

		// Add user input to conversation history immediately
		m.conversationHistory = append(m.conversationHistory, fmt.Sprintf("User: %s", input))
		if len(m.conversationHistory) > 20 {
			m.conversationHistory = m.conversationHistory[len(m.conversationHistory)-20:]
		}

		// Start building the current response
		m.conversationHistory = append(m.conversationHistory, "")

		// Keep processing = true so spinner continues
		// The spinner will keep spinning until we get a real response

		// Call Ollama in a goroutine and stream response chunks in real-time
		go func() {
			// Clear the response buffer for new response
			m.responseBuffer.Reset()

			// Stream response chunks to the buffer and send updates to main thread
			_, err := ollama.SendToOllamaWithCallback(m.ollamaURL, m.model, input, m.context, m.temperature, m.topP, m.toolsEnabled, func(chunk string) {
				// Send chunk to main thread for real-time display via channel
				select {
				case m.streamChannel <- chunk:
					// Chunk sent successfully
				default:
					// Channel buffer full, skip this chunk
				}
			})

			if err != nil {
				logToFile(fmt.Sprintf("Ollama error: %v", err))
				// Add error to conversation history
				m.conversationHistory[len(m.conversationHistory)-1] += fmt.Sprintf("Error: %v", err)
			}

			// Stop processing and spinner
			m.processing = false
			m.responseComplete = true
		}()

		return m, nil
	case processingCompleteMsg:
		// Processing is complete, stop the spinner
		m.processing = false
	case ollamaStreamMsg:
		// Handle streaming response chunks
		// NOTE: This case is not currently used - all streaming goes through streamChannel
		// Keeping this for potential future use but not processing chunks here to avoid duplication
		if m.processing {
			// Log that we received a direct stream message (for debugging)
			logToFile(fmt.Sprintf("Received direct stream message (not used): '%s'", msg.chunk))
			// Don't append here to avoid duplicate processing
		}
	case tickMsg:
		// Update spinner frame
		if m.processing {
			m.spinnerFrame = (m.spinnerFrame + 1) % 10 // Fixed: use 10 for all spinner characters
			logToFile(fmt.Sprintf("Tick: processing=true, spinnerFrame=%d", m.spinnerFrame))

			// Check for streaming chunks while processing
			select {
			case chunk := <-m.streamChannel:
				// Got a chunk, append it to the current response
				logToFile(fmt.Sprintf("Received chunk: '%s'", chunk))

				// Ensure we have a valid conversation history index
				if len(m.conversationHistory) > 0 {
					// For JSON responses, don't break them up - just append
					m.conversationHistory[len(m.conversationHistory)-1] += chunk
				} else {
					// Fallback: create a new response entry if conversation history is empty
					logToFile("Warning: conversation history empty, creating new response entry")
					m.conversationHistory = append(m.conversationHistory, chunk)
				}
			default:
				// No chunk available, continue
			}
		} else {
			logToFile(fmt.Sprintf("Tick: processing=false, spinnerFrame=%d", m.spinnerFrame))
		}
		// Return a new tick command to keep the animation going
		return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return m, nil
}

// View renders the REPL interface
func (m *REPLModel) View() string {
	logToFile("View() called")

	if m.quitting {
		return "Goodbye! 👋\n"
	}

	var s strings.Builder

	// Title
	s.WriteString("🚀 Slop Shop - AI-Powered Code Analysis\n")
	s.WriteString("Repository context loaded. Type your questions about the codebase.\n")
	s.WriteString("Use ↑/↓ arrows to navigate command history, F1-F4 for shortcuts, Ctrl+C to quit.\n\n")

	// Show help if requested
	if m.showHelp {
		s.WriteString("Keyboard Shortcuts:\n")
		s.WriteString("  F1       - Toggle this help message\n")
		s.WriteString("  F2       - Toggle command history display\n")
		s.WriteString("  F3       - Toggle repository context info\n")
		s.WriteString("  F4       - Clear conversation history\n")
		s.WriteString("  F5       - Clear local context (Ollama internal context persists)\n")
		s.WriteString("  F10      - Exit the REPL\n")
		if m.debugEnabled {
			s.WriteString("  Debug logging: ENABLED\n")
		}
		s.WriteString("  ↑/↓      - Navigate command history\n")
		s.WriteString("  Esc      - Hide all panels\n")
		s.WriteString("  Ctrl+C   - Force quit\n")
		s.WriteString("\n")
	}

	// Show history if requested
	if m.showHistory {
		if len(m.history) == 0 {
			s.WriteString("No commands in history yet.\n")
		} else {
			s.WriteString("Command History:\n")
			for i, cmd := range m.history {
				s.WriteString(fmt.Sprintf("  %d: %s\n", i+1, cmd))
			}
			s.WriteString(fmt.Sprintf("Total: %d commands\n", len(m.history)))
		}
		s.WriteString("\n")
	}

	// Show context if requested
	if m.showContext {
		s.WriteString("Repository Context:\n")
		s.WriteString(fmt.Sprintf("Loaded: %d characters\n", len(m.context)))

		// Show file type breakdown
		if m.context != "" {
			// We need to recreate the file list to show the breakdown
			// For now, just show the context size info
			s.WriteString(fmt.Sprintf("Context size: %d bytes\n", len(m.context)))
		}
		s.WriteString("\n")
	}

	// Conversation history
	if len(m.conversationHistory) > 0 {
		s.WriteString("Recent conversation:\n")
		start := 0
		if len(m.conversationHistory) > 6 {
			start = len(m.conversationHistory) - 6
		}
		for _, exchange := range m.conversationHistory[start:] {
			if strings.HasPrefix(exchange, "User: ") {
				s.WriteString(styles.UserStyle.Render(exchange) + "\n")
			} else if !strings.HasPrefix(exchange, "User: ") && !strings.HasPrefix(exchange, "System: ") {
				// This is an assistant response (no prefix)
				response := exchange

				// Don't wrap JSON responses - they should stay intact
				if strings.Contains(response, "{") && strings.Contains(response, "}") {
					// This looks like JSON, don't wrap it
					s.WriteString(styles.AssistantStyle.Render(response) + "\n")
				} else {
					// Process markdown responses to preserve line breaks
					// First try splitting by actual newline characters
					lines := strings.Split(response, "\n")
					if len(lines) == 1 {
						// No actual newlines, try literal \n characters
						lines := strings.Split(response, "\\n")
						if len(lines) == 1 {
							// No line breaks at all, handle as before
							if len(response) > 80 {
								wrapped := wrapText(response, 80)
								s.WriteString(styles.AssistantStyle.Render(wrapped) + "\n")
							} else {
								s.WriteString(styles.AssistantStyle.Render(response) + "\n")
							}
						} else {
							// Found literal \n characters, render each line
							for _, line := range lines {
								if strings.TrimSpace(line) != "" {
									// Apply word wrapping to each line
									if len(line) > 80 {
										wrapped := wrapText(line, 80)
										s.WriteString(styles.AssistantStyle.Render(wrapped) + "\n")
									} else {
										s.WriteString(styles.AssistantStyle.Render(line) + "\n")
									}
								}
							}
						}
					} else {
						// Found actual newlines, render each line
						for _, line := range lines {
							if strings.TrimSpace(line) != "" {
								// Apply word wrapping to each line
								if len(line) > 80 {
									wrapped := wrapText(line, 80)
									s.WriteString(styles.AssistantStyle.Render(wrapped) + "\n")
								} else {
									s.WriteString(styles.AssistantStyle.Render(line) + "\n")
								}
							}
						}
					}
				}
			} else {
				s.WriteString(exchange + "\n")
			}
		}
		s.WriteString("\n")
	}

	// Input prompt
	if m.processing {
		// Show rotating spinner when processing
		spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinnerChar := spinnerChars[m.spinnerFrame%len(spinnerChars)]
		s.WriteString(spinnerChar)
		s.WriteString(" ")
		logToFile(fmt.Sprintf("View: processing=true, spinnerFrame=%d, spinnerChar='%s'", m.spinnerFrame, spinnerChar))
	} else {
		// Show robot emoji when idle
		s.WriteString("🤖 ")
		logToFile(fmt.Sprintf("View: processing=false, input='%s'", m.input))
	}
	s.WriteString(m.input)
	s.WriteString("█")

	return s.String()
}

// submitInput processes the current input
func (m *REPLModel) submitInput() tea.Cmd {
	input := strings.TrimSpace(m.input)
	if input == "" {
		return nil
	}

	// Add to history
	if len(m.history) == 0 || input != m.history[len(m.history)-1] {
		m.history = append(m.history, input)
	}
	m.historyIndex = len(m.history)

	// Clear input immediately and set processing state
	m.input = ""
	m.processing = true

	// Send request to Ollama
	return func() tea.Msg {
		return ollamaRequestMsg{input: input}
	}
}

// navigateHistory moves through command history
func (m *REPLModel) navigateHistory(direction int) tea.Cmd {
	return func() tea.Msg {
		if direction < 0 && m.historyIndex > 0 {
			m.historyIndex--
			m.input = m.history[m.historyIndex]
		} else if direction > 0 && m.historyIndex < len(m.history)-1 {
			m.historyIndex++
			m.input = m.history[m.historyIndex]
		} else if direction > 0 && m.historyIndex == len(m.history)-1 {
			m.historyIndex++
			m.input = ""
		}
		return nil
	}
}

// toggleHelp shows/hides help
func (m *REPLModel) toggleHelp() tea.Cmd {
	return func() tea.Msg {
		m.showHelp = !m.showHelp
		return nil
	}
}

// toggleHistory shows/hides command history
func (m *REPLModel) toggleHistory() tea.Cmd {
	return func() tea.Msg {
		m.showHistory = !m.showHistory
		return nil
	}
}

// toggleContext shows/hides repository context info
func (m *REPLModel) toggleContext() tea.Cmd {
	return func() tea.Msg {
		m.showContext = !m.showContext
		return nil
	}
}

// Global debug flag
var globalDebugEnabled bool

// SetGlobalDebug sets the global debug flag
func SetGlobalDebug(enabled bool) {
	globalDebugEnabled = enabled
}

// logToFile writes debug information to a log file only if debug is enabled
func logToFile(message string) {
	if !globalDebugEnabled {
		return
	}

	f, err := os.OpenFile("repl_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("15:04:05.000")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
	f.WriteString(logMessage)
}

// wrapText wraps text to a specified width, breaking at word boundaries
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine != "" {
				currentLine += " " + word
			} else {
				currentLine = word
			}
		} else {
			if currentLine != "" {
				result.WriteString(currentLine + "\n")
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		result.WriteString(currentLine)
	}

	return result.String()
}

// buildREPLPrompt builds a prompt that includes conversation history
func buildREPLPrompt(context, currentInput string, history []string) string {
	var buf strings.Builder

	// Add repository context
	buf.WriteString("Repository Context:\n")
	buf.WriteString(context)
	buf.WriteString("\n\n")

	// Add conversation history if any
	if len(history) > 0 {
		buf.WriteString("Previous conversation:\n")
		for _, exchange := range history {
			buf.WriteString(exchange)
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// Add current user input
	buf.WriteString("Current question: ")
	buf.WriteString(currentInput)

	return buf.String()
}

// showREPLHelp displays available REPL commands
func showREPLHelp() {
	fmt.Println(styles.HeaderStyle.Render("Available commands:"))
	fmt.Println(styles.InfoStyle.Render("  help     - Show this help message"))
	fmt.Println(styles.InfoStyle.Render("  clear    - Clear conversation history"))
	fmt.Println(styles.InfoStyle.Render("  context  - Show repository context info"))
	fmt.Println(styles.InfoStyle.Render("  history  - Show command history"))
	fmt.Println(styles.InfoStyle.Render("  quit     - Exit the REPL"))
	fmt.Println(styles.InfoStyle.Render("  exit     - Exit the REPL"))
	fmt.Println(styles.InfoStyle.Render("  q        - Exit the REPL"))
	fmt.Println("")
	fmt.Println(styles.InfoStyle.Render("Just type your questions about the codebase!"))
	fmt.Println(styles.InfoStyle.Render("Use ↑/↓ arrows to navigate command history."))
}

// showCommandHistory displays the command history
func showCommandHistory(history []string) {
	if len(history) == 0 {
		fmt.Println(styles.InfoStyle.Render("No commands in history yet."))
		return
	}

	fmt.Println(styles.HeaderStyle.Render("Command History:"))
	for i, cmd := range history {
		fmt.Print(styles.InfoStyle.Render(fmt.Sprintf("  %d: %s\n", i+1, cmd)))
	}
	fmt.Println(styles.InfoStyle.Render(fmt.Sprintf("Total: %d commands", len(history))))
}

// getContextManagementInfo returns information about context management options
func getContextManagementInfo() string {
	return `Context Management Information:

⚠️  IMPORTANT: Ollama maintains internal conversation context that cannot be cleared via API.

Current F5 behavior:
- Reloads repository files (clears local context)
- Ollama internal context persists

To completely clear Ollama's internal context:
1. Restart Ollama service: 'ollama serve' or restart the Ollama process
2. Use a new model instance
3. Restart your terminal session

Alternative approaches:
- Use F4 to clear conversation history (local only)
- Start with -empty-context flag for no repository context
- Restart the entire application`
}
