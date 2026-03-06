# AGENT DEVELOPMENT GUIDELINES FOR SINGEROS

This document contains essential information for AI agents working with the SingerOS codebase.

## BUILD/LINT/TEST COMMANDS

### Build Commands
- `go build -o ./bundles/singeros ./cmd/singeros/main.go` - Build the main SingerOS binary
- `go install ./cmd/singeros/main.go` - Install the SingerOS binary
- `make docker-build` - Build Docker image (tag: registry.yygu.cn/insmtx/SingerOS:latest)
- `make docker-run` - Run the Docker image locally

### Test Commands
- `go test ./...` - Run all tests in the project
- `go test -v ./...` - Run all tests with verbose output
- `go test ./pkg/path/to/package` - Run tests for a specific package
- `go test -run ^TestFunctionName$ ./pkg/path` - Run a specific test function
- `go test -race ./...` - Run all tests with race condition detection
- `go test -cover ./...` - Run tests and display coverage information

### Alternative Test Commands (from CONTRIBUTING.md)
- `make test` - Run all tests (as referenced in documentation)
- `make test-cover` - Run tests with coverage (as referenced in documentation)

### Lint Commands
- `go fmt ./...` - Format all Go code
- `go vet ./...` - Vet all Go code for common mistakes
- `golint ./...` - Lint all Go code (install via `go install golang.org/x/lint/golint@latest`)
- `gofmt -s -w .` - Simplify code and write changes (as per the existing Makefile)
- `staticcheck ./...` - Comprehensive Go static analysis (if installed)

## CODE STYLE GUIDELINES

### Import Organization
- Group imports with blank lines between standard library, third-party, and project-specific packages
- Use semantic import aliases only when they prevent naming conflicts
- Organize in three groups: stdlib, third-party, internal packages
```
import (
	"fmt"
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	
	"github.com/insmtx/SingerOS/internal/config"
)
```

### Formatting Conventions
- Use tabs for indentation, not spaces (as verified from existing Go files)
- Execute `go fmt ./...` before committing
- Keep lines under 120 characters where possible
- Use `gofmt -s` for simplification of code

### Naming Conventions
- Use CamelCase for exported functions/types (`GetUser`, `UserService`)
- Use camelCase for unexported/internal functions/types (`getUser`, `userService`)
- Use clear, descriptive names; prefer clarity over brevity
- Use consistent names for similar concepts across packages
- Variables related to the system should reference SingerOS concepts

### Types and Interfaces
- Define interfaces close to their first usage
- Keep interfaces small, typically one or a few methods
- Name interface types with "-er" suffix when applicable (e.g., `Runner`, `Handler`)
- Use concrete types explicitly in function signatures when interface is not needed
- Prefer returning pointers for structs when passing to functions if they will be modified

### Error Handling
- Handle errors explicitly; don't ignore them
- Use specific error types when appropriate with wrapped errors
- Follow the pattern: "if err != nil { return err }"
- Use `errors.New()` for simple static strings
- Use `fmt.Errorf()` with `%w` verb for wrapping errors with more context
- Log errors contextually when appropriate

### Additional Guidelines
- All public functions must have GoDoc comments
- Comments should be in English and explain why rather than what
- Maintain consistent logging format throughout the application
- Use context.Context appropriately for cancellation and request-scoped values
- Follow dependency injection patterns rather than global variables
- Use Cobra for command-line interface implementations as shown in main.go files

## PROJECT STRUCTURE

- `/backend` - Main Go application code
- `/backend/cmd` - Entry points for different SingerOS services
- `/internal` - Private internal code that should not be imported by other projects
- `/pkg` - Public libraries that can be used by other applications
- `/docs` - Documentation files
- `/deployments/build/Dockerfile` - Container build configuration

## CONTRIBUTION NOTES

- See CONTRIBUTING.md for commit message style guidance
- Make sure all tests pass (`go test ./...`) before submitting changes
- Follow Go's idiomatic patterns and standard practices
- When implementing, consider how components fit into the broader microservices architecture described in ARCHITECTURE.md