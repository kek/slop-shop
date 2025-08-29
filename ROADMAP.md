# Slop Shop Development Roadmap

## Completed

### Fix Test Suite

- [x] Move `repl_test.go` tests to the `tui` package where `REPLModel` is defined
- [x] Create proper test structure for main package tests
- [x] Implement missing helper functions:
  - [x] `analyzeFileTypes()` function
  - [x] `formatBytes()` function
- [x] Fix all undefined type references (`REPLModel`, `tickMsg`, `FileInfo`)
- [x] Ensure all tests pass with `go test -v`

### Clean Up Dependencies

- [x] Remove unused dependencies:
  - [x] `github.com/containerd/console`
  - [x] `github.com/muesli/reflow`
  - [x] `golang.org/x/term`
- [x] Run `go mod tidy` to clean up
- [x] Verify no functionality is broken

### Add Integration Tests

- [x] Create integration tests for Ollama API calls
- [x] Add tests for tool execution system
- [x] Test repository scanning with various file types
- [x] Add tests for streaming response handling
- [x] Create mock Ollama server for testing

## To Do

### Fix Error Handling

- [x] Fix empty error handler in `main.go:93` (Ollama streaming errors) - **FIXED**: Now shows clear error message
- [x] Fix empty error handler in `tui/tui.go:188` (Ollama response errors) - **FIXED**: Now shows error in conversation history
- [x] Add user-visible error messages instead of silent failures - **FIXED**: Both batch and REPL modes show errors
- [x] Improve error display in REPL mode when Ollama fails - **FIXED**: Errors appear in conversation history

### Add Missing Tool Types

- [ ] Add `WRITE_FILE` tool (currently only has `CREATE_FILE`)
- [ ] Add `DELETE_FILE` tool
- [ ] Add `MOVE_FILE` tool
- [ ] Add `COPY_FILE` tool
