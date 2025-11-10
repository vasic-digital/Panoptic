# CRUSH.md - Panoptic Development Guide

This guide provides essential information for agents working on the Panoptic automated testing framework.

## Project Overview

Panoptic is a comprehensive Go-based automated testing framework for web, desktop, and mobile applications. It provides UI automation, screenshot capture, and video recording capabilities across multiple platforms.

**Language**: Go 1.21+
**Architecture**: Clean architecture with modular design
**Primary Dependencies**: Cobra (CLI), Viper (Config), Chromedp/Rod (Browser Automation), Logrus (Logging)

## Essential Commands

### Build and Run
```bash
# Build the application
go build -o panoptic main.go
# OR use the build script
./build.sh

# Run with config
./panoptic run config.yaml

# Run help
./panoptic --help
```

### Testing Commands
```bash
# Run all tests with coverage
./scripts/test.sh --coverage

# Run only unit tests
./scripts/test.sh --skip-integration --skip-e2e

# Run with verbose output
./scripts/test.sh -v --coverage

# Manual unit tests
go test ./internal/... ./cmd/...

# Integration tests (requires build tags)
go test -tags=integration ./tests/integration/...

# E2E tests (requires build tags)
go test -tags=e2e ./tests/e2e/...

# Security tests
go test -tags=security ./tests/security/...

# Functional tests
go test -tags=functional ./tests/functional/...
```

### Coverage and Quality
```bash
# Generate detailed coverage report
./scripts/coverage.sh

# Performance testing
./scripts/performance_test.sh --quick

# Monitoring dashboard
./scripts/dashboard.sh

# Generate test data
./scripts/generate_test_data.sh
```

### Manual Quality Checks
```bash
# Format code
gofmt -s -w .

# Vet code
go vet ./...

# Static analysis
staticcheck ./...

# Security scan
gosec ./...

# Run linter (if golangci-lint installed)
golangci-lint run
```

## Project Structure

```
Panoptic/
├── internal/                    # Core application modules
│   ├── config/                 # Configuration management
│   │   ├── config.go           # Config struct and loading
│   │   └── config_test.go      # Config unit tests
│   ├── executor/               # Test execution engine
│   │   ├── executor.go          # Main execution logic
│   │   └── executor_test.go     # Executor unit tests
│   ├── logger/                 # Logging functionality
│   │   ├── logger.go            # Structured logging
│   │   └── logger_test.go       # Logger unit tests
│   └── platforms/              # Platform implementations
│       ├── platform.go          # Platform interface and factory
│       ├── web.go              # Web platform (Chrome/Chromium)
│       ├── desktop.go          # Desktop platform implementation
│       ├── mobile.go           # Mobile platform implementation
│       └── platform_test.go    # Platform unit tests
├── cmd/                        # Command-line interface
│   ├── root.go                 # CLI root command and global flags
│   ├── run.go                  # Run command implementation
│   └── cmd_test.go             # CLI unit tests
├── tests/                      # Test suites (build tags required)
│   ├── functional/             # Functional tests
│   ├── integration/            # Integration tests
│   ├── e2e/                   # End-to-end tests
│   └── security/               # Security tests
├── scripts/                    # Automation scripts
│   ├── test.sh                 # Main test runner
│   ├── coverage.sh             # Coverage analysis
│   ├── performance_test.sh     # Performance testing
│   ├── dashboard.sh            # Monitoring dashboard
│   └── generate_test_data.sh   # Test data generation
├── docs/                       # Documentation
├── .github/workflows/          # CI/CD pipelines
├── main.go                     # Application entry point
├── example-config.yaml         # Example configuration
└── go.mod                      # Go module definition
```

## Code Conventions and Patterns

### Naming Conventions
- **Package names**: Short, lowercase (e.g., `config`, `executor`, `logger`)
- **File names**: `snake_case.go` for source files, `snake_case_test.go` for test files
- **Struct names**: `PascalCase` (e.g., `Config`, `Platform`, `Executor`)
- **Interface names**: `PascalCase` often ending with capability description (e.g., `Platform`)
- **Method names**: `PascalCase` for exported, `camelCase` for unexported
- **Constants**: `UPPER_SNAKE_CASE`
- **Variables**: `camelCase`

### Go Idioms Used
- **Error handling**: Explicit error returns, wrapped errors with context
- **Interface-based design**: Platform interface for extensibility
- **Factory pattern**: PlatformFactory for creating platform instances
- **Configuration with struct tags**: YAML tags for configuration parsing
- **Dependency injection**: Through constructor functions

### Code Style
- Use `gofmt -s` for formatting
- Maximum line length: 120 characters
- Package-level constants before variables
- Interface methods grouped logically
- Exported types have documentation comments
- Error messages are lowercase and don't end with punctuation

## Configuration System

### Config Structure
```go
type Config struct {
    Name     string       `yaml:"name"`
    Output   string       `yaml:"output"`
    Apps     []AppConfig  `yaml:"apps"`
    Actions  []Action     `yaml:"actions"`
    Settings Settings     `yaml:"settings"`
}
```

### Supported Platforms
- **Web**: Chrome/Chromium automation via chromedp/rod
- **Desktop**: Cross-platform desktop application control
- **Mobile**: Android/iOS device and emulator testing

### Action Types
- `navigate`: Navigate to URL
- `click`: Click element by CSS selector
- `fill`: Fill form field
- `submit`: Submit form
- `wait`: Wait for specified duration
- `screenshot`: Capture screenshot
- `record`: Start video recording

### Default Values
- Screenshot format: `png`
- Video format: `mp4`
- Quality: `80`
- Window size: `1920x1080`
- Log level: `info`

## Testing Approach

### Test Categories
1. **Unit Tests**: Test individual functions in isolation
2. **Integration Tests**: Test component interactions
3. **E2E Tests**: Test complete workflows
4. **Functional Tests**: Test functional requirements
5. **Security Tests**: Test security vulnerabilities

### Build Tags
- `integration`: Integration tests
- `e2e`: End-to-end tests
- `functional`: Functional tests
- `security`: Security tests

### Coverage Requirements
- **Minimum**: 75% coverage
- **Target**: 85% for core packages
- **Excellent**: 90%+ for critical components

### Test Patterns
- Table-driven tests for multiple scenarios
- Setup/teardown in each test
- Test both success and failure scenarios
- Error testing with specific error types
- Mock external dependencies

## Platform Implementation Pattern

### Interface Design
```go
type Platform interface {
    Initialize(app config.AppConfig) error
    Navigate(url string) error
    Click(selector string) error
    Fill(selector, value string) error
    Submit(selector string) error
    Wait(duration int) error
    Screenshot(filename string) error
    StartRecording(filename string) error
    StopRecording() error
    GetMetrics() map[string]interface{}
    Close() error
}
```

### Factory Pattern
```go
func (f *PlatformFactory) CreatePlatform(appType string) (Platform, error) {
    switch appType {
    case "web":
        return NewWebPlatform(), nil
    case "desktop":
        return NewDesktopPlatform(), nil
    case "mobile":
        return NewMobilePlatform(), nil
    default:
        return nil, fmt.Errorf("unsupported platform type: %s", appType)
    }
}
```

## Dependencies and External Tools

### Go Dependencies
- `github.com/chromedp/chromedp` - Chrome DevTools Protocol
- `github.com/go-rod/rod` - Browser automation
- `github.com/sirupsen/logrus` - Structured logging
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/stretchr/testify` - Testing utilities
- `gopkg.in/yaml.v3` - YAML parsing

### System Dependencies
- **Chrome/Chromium**: Required for web automation
- **Android SDK**: For mobile testing (optional)
- **Xcode tools**: For iOS testing (optional)
- **ADB**: Android Debug Bridge (for mobile)

### Development Tools
- `gofmt`: Code formatting
- `go vet`: Static analysis
- `staticcheck`: Advanced static analysis
- `gosec`: Security scanning
- `golangci-lint`: Linting (optional)

## Important Gotchas

### Build Tags
- Integration, E2E, and security tests require build tags
- Use `go test -tags=integration` for integration tests
- Build tags allow skipping heavy tests in CI

### Platform Dependencies
- Web automation requires Chrome/Chromium installation
- Mobile testing requires platform-specific tools
- Desktop testing may require system permissions

### Configuration Validation
- Always validate configuration after loading
- Default values are applied in `config.go`
- Invalid configurations should return descriptive errors

### Error Handling
- Wrap errors with context using `fmt.Errorf`
- Use structured error messages
- Log errors at appropriate levels
- Test error paths in unit tests

### Resource Management
- Always call `Close()` on platform instances
- Cleanup temporary files after tests
- Manage browser instances carefully to avoid memory leaks

### Concurrency
- Be careful with concurrent test execution
- Some platform operations may not be thread-safe
- Use proper synchronization when needed

## Development Workflow

### Before Making Changes
1. Run existing tests to ensure clean state
2. Read related code to understand patterns
3. Check if similar functionality exists
4. Plan error handling approach

### Making Changes
1. Add tests for new functionality first
2. Implement changes following existing patterns
3. Run tests frequently during development
4. Update documentation if needed

### Testing Changes
1. Run unit tests: `go test ./internal/... ./cmd/...`
2. Run integration tests: `go test -tags=integration ./tests/integration/...`
3. Check coverage: `./scripts/coverage.sh`
4. Run quality checks: `go vet ./...`, `gosec ./...`

### Final Checks
1. Run full test suite: `./scripts/test.sh --coverage`
2. Format code: `gofmt -s -w .`
3. Check coverage requirements
4. Verify documentation is updated

## Environment Variables

### Configuration
- `PANOPTIC_TEST_TIMEOUT`: Test timeout duration
- `PANOPTIC_TEST_VERBOSE`: Enable verbose test output
- `PANOPTIC_LOG_LEVEL`: Logging level (debug, info, warn, error)

### Browser Configuration
- `PANOPTIC_BROWSER_HEADLESS`: Enable headless browser mode
- `PANOPTIC_BROWSER_WIDTH`: Browser window width
- `PANOPTIC_BROWSER_HEIGHT`: Browser window height

### Mobile Testing
- `ANDROID_SDK_ROOT`: Android SDK path
- `IOS_SIMULATOR_PATH`: iOS simulator path

## Common Issues and Solutions

### Browser Not Available
- Install Chrome/Chromium
- Set `PANOPTIC_BROWSER_HEADLESS=false` for debugging
- Check browser path in system PATH

### Mobile Tools Missing
- Install Android SDK for Android testing
- Install Xcode tools for iOS testing
- Set appropriate environment variables

### Test Timeouts
- Increase `PANOPTIC_TEST_TIMEOUT`
- Check network connectivity for web tests
- Verify system resources are sufficient

### Coverage Below Threshold
- Run `./scripts/coverage.sh` for detailed analysis
- Focus on uncovered critical paths
- Add tests for error conditions

## Performance Considerations

### Resource Usage
- Browser automation can be memory intensive
- Parallel test execution may require more resources
- Video recording consumes significant disk space

### Optimization
- Use headless mode when possible
- Clean up browser instances promptly
- Consider test parallelism for faster execution

This guide should help agents work effectively with the Panoptic codebase while maintaining consistency and quality standards.