# Slop Shop

This Go program reads the contents of a repository and sends them as context to an Ollama model (default: qwen3:latest) along with a user-provided prompt.

## Features

- Walks through repository directories recursively
- Excludes binary files, git files, and other non-text files
- Configurable exclusion patterns
- Sends repository contents as context to Ollama API
- Supports custom prompts and model parameters

## Prerequisites

- Go 1.25.0 or later
- Ollama running locally with qwen3:latest model available
- The qwen3:latest model should be pulled: `ollama pull qwen3:latest`

## Installation

1. Clone or download this repository
2. Build the program:
   ```bash
   go build -o slop-shop
   ```

## Usage

### Basic Usage

```bash
# Single prompt mode
./slop-shop -prompt "Analyze this codebase and suggest improvements"

# Interactive REPL mode
./slop-shop -repl
```

### Advanced Usage

```bash
./slop-shop \
  -prompt "Find potential security vulnerabilities in this code" \
  -model qwen3:latest \
  -repo /path/to/your/repo \
  -url http://localhost:11434 \
  -temp 0.5 \
  -top-p 0.8 \
  -exclude ".git,.jj,node_modules,vendor,*.exe,*.dll,*.so,*.dylib,*.bin"
```

### Command Line Flags

| Flag             | Description                                           | Default                                                             | Required                     |
| ---------------- | ----------------------------------------------------- | ------------------------------------------------------------------- | ---------------------------- |
| `-prompt`        | The prompt to send to the model                       | -                                                                   | **Yes** unless using `-repl` |
| `-repl`          | Start interactive REPL mode                           | false                                                               | No                           |
| `-tools`         | Enable tool execution for LLM                         | false                                                               | No                           |
| `-model`         | Ollama model to use                                   | qwen3:latest                                                        | No                           |
| `-repo`          | Path to repository                                    | . (current directory)                                               | No                           |
| `-url`           | Ollama API URL                                        | http://localhost:11434                                              | No                           |
| `-temp`          | Temperature for generation                            | 0.7                                                                 | No                           |
| `-top-p`         | Top-p for generation                                  | 0.9                                                                 | No                           |
| `-exclude`       | Comma-separated patterns to exclude                   | .git,.jj,node_modules,vendor,_.exe,_.dll,_.so,_.dylib,\*.bin,.crush | No                           |
| `-empty-context` | Start with empty context (no repository files loaded) | false                                                               | No                           |
| `-debug`         | Enable debug logging to file                          | false                                                               | No                           |

## How It Works

1. **Repository Scanning**: The program recursively walks through the specified repository directory
2. **File Filtering**: Excludes binary files, git files, and other non-text files based on patterns
3. **Context Creation**: Combines all text files into a formatted context string
4. **API Communication**: Sends the context and prompt to Ollama via HTTP API
5. **Response Display**: Shows the model's response

## Modes of Operation

### Single Prompt Mode

Send a single prompt and get a response. Use the `-prompt` flag.

### Interactive REPL Mode

Start an interactive session where you can ask multiple questions about the codebase. Use the `-repl` flag.

**REPL Shortcuts:**

- `F1` - Toggle help message
- `F2` - Toggle command history
- `F3` - Toggle repository context info
- `F4` - Clear conversation history
- `F5` - Clear local context
- `F10` - Exit the REPL
- `Ctrl+C` - Force quit

**REPL Features:**

- Maintains conversation history for context
- Automatic context management to prevent overflow
- Interactive prompt for continuous code analysis

### Tools Mode

Enable the LLM to execute tools for gathering information, testing, and file modifications. Use the `-tools` flag.

**Available Tools:**

- **RUN_COMMAND**: Execute shell commands
- **READ_FILE**: Read file contents
- **LIST_DIR**: List directory contents
- **TEST_COMMAND**: Test if commands work
- **SEARCH_FILES**: Search for text patterns in files
- **GENERATE_DIFF**: Generate unified diffs for suggested changes
- **APPLY_DIFF**: Apply unified diffs to repository files
- **CREATE_FILE**: Create a new file with specified content

**Example:**

```bash
./slop-shop -tools -prompt "Add error handling to the main function"
```

**Tool Usage in REPL:**

```bash
./slop-shop -repl -tools
```

Then in the REPL, you can use tools like:

```
LIST_DIR: .
TEST_COMMAND: go version
RUN_COMMAND: ls -la
GENERATE_DIFF: Add error handling to main function
```

## Example Output

```
Reading repository at: .
Using model: qwen3:latest
Prompt: Analyze this codebase and suggest improvements
Ollama URL: http://localhost:11434
Found 3 files
Total context size: 1250 characters

=== Ollama Response ===
Based on the repository contents, here are my suggestions for improvements...
```

## Excluding Files

The program automatically excludes common non-text files and directories. You can customize exclusions using the `-exclude` flag:

```bash
-exclude ".git,.jj,node_modules,vendor,*.exe,*.dll,*.so,*.dylib,*.bin,temp,logs"
```

## Troubleshooting

### Ollama Not Running

Make sure Ollama is running:

```bash
ollama serve
```

### Model Not Found

Pull the required model:

```bash
ollama pull qwen3:latest
```

### Permission Issues

Ensure the program has read access to the repository directory.

### Large Repositories

For very large repositories, consider using more specific exclusion patterns to reduce context size.

## License

This project is open source and available under the MIT License.
