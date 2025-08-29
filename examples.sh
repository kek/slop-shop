#!/bin/bash

# Example usage of the repo-context program
# Make sure Ollama is running and the program is built

echo "=== Repository Context to Ollama Examples ==="
echo ""

echo "1. Basic usage - analyze the current repository:"
echo "./slop-shop -prompt \"What is this repository about?\""
echo ""
echo "2. Interactive REPL mode:"
echo "./slop-shop -repl"
echo ""
echo "3. Tools mode - generate and apply changes:"
echo "./slop-shop -tools -prompt \"Add error handling to the main function\""
echo ""
echo "4. Tools mode - enable LLM tool execution:"
echo "./slop-shop -tools -prompt \"Check the current directory and test if Go is working\""
echo ""

echo "5. Code review with specific focus:"
echo "./repo-context -prompt \"Review this code for security vulnerabilities and suggest improvements\""
echo ""

echo "6. Using a different model:"
echo "./repo-context -model llama3.2:3b -prompt \"Explain this code in simple terms\""
echo ""

echo "7. Custom repository path:"
echo "./repo-context -repo /path/to/other/repo -prompt \"Analyze this codebase\""
echo ""

echo "8. Custom Ollama server:"
echo "./repo-context -url http://192.168.1.100:11434 -prompt \"What does this code do?\""
echo ""

echo "9. Adjusting generation parameters:"
echo "./repo-context -temp 0.3 -top-p 0.8 -prompt \"Generate documentation for this code\""
echo ""

echo "10. Custom file exclusions:"
echo "./repo-context -exclude \".git,.jj,node_modules,vendor,*.exe,*.dll,*.so,*.dylib,*.bin,temp,logs\" -prompt \"Analyze this code\""
echo ""

echo "11. Complex analysis prompt:"
echo "./repo-context -prompt \"Analyze this codebase for: 1) Code quality issues, 2) Performance bottlenecks, 3) Security concerns, 4) Best practices violations. Provide specific examples and suggestions for improvement.\""
echo ""

echo "=== Tips ==="
echo "- Make sure Ollama is running: ollama serve"
echo "- Pull the model if needed: ollama pull qwen3-coder"
echo "- For large repositories, use specific exclusion patterns"
echo "- Adjust temperature for more/less creative responses"
echo "- Use -top-p to control response diversity"
