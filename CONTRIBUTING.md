# Contributing to go-kusto-cli

Thanks for your interest in contributing! Here's how to get started.

## Development Setup

1. **Clone the repo:**

   ```bash
   git clone https://github.com/danielsada/go-kusto-cli.git
   cd go-kusto-cli
   ```

2. **Install dependencies:**

   - Go 1.23+ — https://go.dev/dl/
   - golangci-lint — `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

3. **Verify everything works:**

   ```bash
   make all   # runs lint, vet, test, build
   ```

## Making Changes

1. Create a branch from `main`.
2. Make your changes.
3. Run `make all` and ensure lint, vet, tests, and build all pass.
4. Commit with a clear message describing the change.
5. Open a pull request.

## Code Style

- Follow standard Go conventions (`gofmt`, `goimports`).
- All exported types, functions, and constants must have doc comments.
- Check all returned errors (enforced by `errcheck` linter).
- Keep dependencies minimal — prefer stdlib over external packages.

## Testing

- Write tests for any new or changed functionality.
- Use `net/http/httptest` for HTTP-related tests.
- Use table-driven tests where appropriate.
- Run `make test` to execute the test suite.
- Aim for ≥80% coverage on new code.

## Reporting Issues

Open a GitHub issue with:

- A clear description of the problem or feature request.
- Steps to reproduce (for bugs).
- Expected vs actual behavior.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
