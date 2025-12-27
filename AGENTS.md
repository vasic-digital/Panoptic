# AGENTS.md

This file contains essential information for AI agents working on the Panoptic project. It includes build commands, testing procedures, code style guidelines, and conventions.

## Build Commands

### Building the Application
```bash
# Build the application
go build -o panoptic main.go

# Alternative using build script
./build.sh

# Build for specific architecture
GOOS=linux GOARCH=amd64 go build -o panoptic-linux-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o panoptic-darwin-arm64 main.go
GOOS=windows GOARCH=amd64 go build -o panoptic-windows-amd64.exe main.go
```

### Linting and Formatting
```bash
# Format all Go files
gofmt -s -w .

# Check formatting without applying changes
gofmt -s -l .

# Run golangci-lint
golangci-lint run

# Run specific linters
golangci-lint run -E govet -E staticcheck -E errcheck

# Fix issues automatically (if supported by linter)
golangci-lint run --fix
```

## Testing Commands

### Running All Tests
```bash
# Run all tests using the test script
./scripts/test.sh

# Verbose output
./scripts/test.sh -v

# With coverage
./scripts/test.sh --coverage

# With race detector
./scripts/test.sh --race

# Skip integration/e2e tests
./scripts/test.sh --skip-integration --skip-e2e
```

### Running Specific Tests
```bash
# Run all unit tests
go test ./internal/... ./cmd/...

# Run tests in verbose mode
go test -v ./internal/... ./cmd/...

# Run tests for specific package
go test ./internal/platforms/...
go test ./internal/executor/...

# Run single test file
go test ./internal/platforms/platform_test.go

# Run specific test function
go test -v ./internal/platforms -run TestPlatformFactory_CreatePlatform

# Run tests with coverage
go test -coverprofile=coverage.out ./internal/...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Test Categories
```bash
# Integration tests (require -tags=integration)
go test -tags=integration ./tests/integration/...

# End-to-end tests (require -tags=e2e)
go test -tags=e2e ./tests/e2e/...

# Functional tests (require -tags=functional)
go test -tags=functional ./tests/functional/...

# Security tests (require -tags=security)
go test -tags=security ./tests/security/...

# Benchmark tests
go test -bench=. -benchmem ./internal/...

# Benchmark specific function
go test -bench=BenchmarkExecuteActions -benchmem ./internal/executor/...
```

### Coverage Analysis
```bash
# Run detailed coverage analysis
./scripts/coverage.sh

# Show coverage by function
go tool cover -func=coverage.out

# Show coverage by file
go tool cover -func=coverage.out | grep -v "total:"
```

## Code Style Guidelines

### Imports Organization
Group imports in the following order:
1. Standard library imports
2. Third-party imports
3. Local project imports

Example:
```go
import (
    "context"
    "fmt"
    "time"

    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"

    "panoptic/internal/config"
    "panoptic/internal/logger"
)
```

### Naming Conventions
- **Variables**: Use camelCase
- **Constants**: Use PascalCase or ALL_CAPS with underscores
- **Types**: Use PascalCase
- **Interfaces**: Use PascalCase, typically ending with 'er' when appropriate (e.g., `Platform`, `Executor`)
- **Files**: Use snake_case for test files (e.g., `platform_test.go`)

### Error Handling
Always check errors and provide context:
```go
// Good
result, err := someOperation()
if err != nil {
    return fmt.Errorf("failed to perform operation: %w", err)
}

// Better with logging
if err := initializeComponent(); err != nil {
    e.logger.Errorf("Failed to initialize component: %v", err)
    return fmt.Errorf("initialization failed: %w", err)
}
```

### Struct Tags
Use consistent struct tags for YAML/JSON serialization:
```go
type Config struct {
    Name     string       `yaml:"name" json:"name"`
    Output   string       `yaml:"output" json:"output"`
    Apps     []AppConfig  `yaml:"apps" json:"apps"`
    Actions  []Action     `yaml:"actions" json:"actions"`
    Settings Settings     `yaml:"settings" json:"settings"`
}
```

### Logging
Use structured logging with logrus:
```go
import "github.com/sirupsen/logrus"

type Service struct {
    logger *logrus.Logger
}

func (s *Service) Process() error {
    s.logger.Info("Starting processing")
    s.logger.WithFields(logrus.Fields{
        "count": len(items),
        "type":  "batch",
    }).Debug("Processing items")
    
    if err := doWork(); err != nil {
        s.logger.WithError(err).Error("Failed to process")
        return err
    }
    
    s.logger.Info("Processing completed")
    return nil
}
```

### Testing Patterns
- Use testify for assertions (`assert`, `require`)
- Use table-driven tests for multiple test cases
- Mock external dependencies
- Clean up resources in tests

Example test structure:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        expectError bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
        },
        {
            name:        "empty input",
            input:       "",
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Performance Considerations
- Use `sync.Pool` for frequent allocations
- Pre-allocate slices/maps when size is known
- Avoid unnecessary string conversions
- Use `strings.Builder` for string concatenation in loops

### Concurrency
- Use `sync.Once` for lazy initialization
- Protect shared resources with mutexes
- Use context for cancellation/timeouts
- Always clean up goroutines

Example:
```go
type Service struct {
    mu       sync.RWMutex
    data     map[string]interface{}
    initOnce sync.Once
}

func (s *Service) Get(key string) (interface{}, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    val, ok := s.data[key]
    if !ok {
        return nil, fmt.Errorf("key not found: %s", key)
    }
    return val, nil
}
```

## Project Structure

### Key Directories
- `cmd/`: Command-line interface
- `internal/`: Internal packages (not for external use)
  - `config/`: Configuration parsing and validation
  - `platforms/`: Platform implementations (web, desktop, mobile)
  - `executor/`: Test execution engine
  - `ai/`: AI-enhanced testing features
  - `cloud/`: Cloud storage integration
  - `enterprise/`: Enterprise features
  - `vision/`: Computer vision capabilities
  - `logger/`: Logging utilities
- `tests/`: Test suites
  - `integration/`: Integration tests
  - `e2e/`: End-to-end tests
  - `functional/`: Functional tests
  - `security/`: Security tests
- `scripts/`: Build and utility scripts

### Configuration Files
- YAML configuration files in root directory (e.g., `test_config.yaml`)
- Use `example-config.yaml` as template
- Configuration supports AI testing, cloud integration, and enterprise features

## Development Workflow

### Before Committing
1. Run tests: `./scripts/test.sh`
2. Check formatting: `gofmt -s -l .`
3. Run linting: `golangci-lint run`
4. Verify dependencies: `go mod tidy`

### Creating New Features
1. Add tests for new functionality
2. Update configuration structures if needed
3. Document new features in appropriate README files
4. Ensure backward compatibility

### Adding New Platform Support
1. Implement the `Platform` interface in `internal/platforms/`
2. Add platform type to `PlatformFactory.CreatePlatform()`
3. Update config validation in `internal/config/config.go`
4. Add comprehensive tests

### Adding New Actions
1. Add action type to `Action` struct in `internal/config/config.go`
2. Implement action handling in `Executor.executeAction()`
3. Ensure all platforms support the action (or handle gracefully)
4. Update documentation and example configs

## CI/CD Pipeline

The project uses GitHub Actions with the following jobs:
- `lint`: Formatting and linting checks
- `test-unit`: Unit tests with coverage
- `test-integration`: Integration tests
- `test-e2e`: End-to-end tests
- `build`: Build verification
- `docker-build`: Docker image building

## Important Notes

- Minimum Go version: 1.22
- Use Go modules for dependency management
- All exported functions must have GoDoc comments
- Error messages should be user-friendly and actionable
- Tests must be deterministic and not rely on external services
- Performance-critical code should include benchmarks
- Security-sensitive operations must be properly validated

## Running the Application

```bash
# Basic usage
./panoptic run config.yaml

# With custom output directory
./panoptic run config.yaml --output ./results

# With verbose logging
./panoptic run config.yaml --verbose

# Show help
./panoptic --help
./panoptic run --help
```