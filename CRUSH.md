# CRUSH.md

## Build & Test
- **Build:** `make build`
- **Lint:** `golangci-lint run`
- **Test (single):** `go test -v -run=TestName ./...`

## Code Style
- **Imports:** Grouped (stdlib → third-party → local), sorted alphabetically
- **Formatting:** `gofmt` and `goimports` (run `make fmt`)
- **Naming:** `PascalCase` for types, `snake_case` for variables/functions
- **Error Handling:** Always check errors; wrap with `errors.Wrap`
- **Types:** Prefer named types (e.g., `type ID string`)

## Notes
- Include Cursor rules from `.cursor/rules/`
- Review `.github/copilot-instructions.md` for Copilot guidelines
- All new code must pass `golangci-lint` and `go vet`