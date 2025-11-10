# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Panoptic is a comprehensive automated testing and recording framework for web, desktop, and mobile applications. It's written in Go and uses a YAML-based configuration system to define test scenarios with advanced features including AI-enhanced testing, cloud storage integration, and enterprise management capabilities.

## Essential Commands

### Building
```bash
# Build the application
go build -o panoptic main.go

# Or use the build script
./build.sh
```

### Running Tests
```bash
# Run all unit tests
go test ./internal/... ./cmd/...

# Run unit tests with verbose output
go test -v ./internal/... ./cmd/...

# Run all tests with coverage
./scripts/test.sh --coverage

# Run tests with verbose output and coverage
./scripts/test.sh -v --coverage

# Run only unit tests (skip integration/e2e)
./scripts/test.sh --skip-integration --skip-e2e

# Run with race detection
./scripts/test.sh --race

# Run integration tests
go test -tags=integration ./tests/integration/...

# Run end-to-end tests
go test -tags=e2e ./tests/e2e/...

# Run functional tests
go test -tags=functional ./tests/functional/...

# Run security tests
go test -tags=security ./tests/security/...

# Generate detailed coverage report
./scripts/coverage.sh
```

### Running the Application
```bash
# Run with a config file
./panoptic run test_config.yaml

# Run with custom output directory
./panoptic run test_config.yaml --output ./my-output

# Run with verbose logging
./panoptic run test_config.yaml --verbose
```

### Performance Testing
```bash
# Quick performance test
./scripts/performance_test.sh --quick

# Stress test
./scripts/performance_test.sh --stress

# Benchmark mode
./scripts/performance_test.sh --benchmark
```

### Monitoring
```bash
# Start real-time monitoring dashboard
./scripts/dashboard.sh
```

## Architecture

### Core Design Pattern

Panoptic follows a clean architecture with clear separation of concerns:

1. **Configuration Layer** (`internal/config`): YAML parsing and validation
2. **Platform Abstraction** (`internal/platforms`): Unified interface for web, desktop, and mobile platforms
3. **Execution Engine** (`internal/executor`): Orchestrates test execution across platforms
4. **Reporting System** (part of executor): Generates HTML and JSON reports

### Platform Implementations

The `Platform` interface (`internal/platforms/platform.go`) defines a contract that all platform implementations must follow:

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

Three implementations exist:
- **WebPlatform** (`web.go`): Uses `go-rod` for Chromium-based browser automation
- **DesktopPlatform** (`desktop.go`): Platform-native UI automation for Windows/macOS/Linux
- **MobilePlatform** (`mobile.go`): Android/iOS device and emulator testing

The `PlatformFactory` creates the appropriate platform based on the app type specified in configuration.

### Executor Flow

The `Executor` (`internal/executor/executor.go`) is the heart of the system:

1. **Initialization**: Creates platform instances, AI components, cloud manager, and enterprise integration
2. **Configuration Processing**: Parses YAML config, validates apps and actions
3. **Test Execution**: For each app, creates appropriate platform and executes actions sequentially
4. **Result Collection**: Captures screenshots, videos, metrics, and error states
5. **Report Generation**: Creates HTML and JSON reports with all test artifacts

### Advanced Features

#### AI-Enhanced Testing (`internal/ai`)
- **TestGenerator** (`testgen.go`): Generates additional test cases based on patterns
- **ErrorDetector** (`errordetector.go`): Intelligent error detection and classification
- **AIEnhancedTester** (`enhanced_tester.go`): Coordinates AI features during test execution

#### Cloud Integration (`internal/cloud`)
- **CloudManager** (`manager.go`): Manages cloud storage providers (AWS S3, GCP, Azure, local)
- **CloudAnalytics**: Performance analytics and CDN integration
- Supports automatic artifact upload, compression, and encryption

#### Enterprise Features (`internal/enterprise`)
- **EnterpriseIntegration** (`integration.go`): Main integration coordinator
- **User/Team Management**: Role-based access control and team collaboration
- **Audit/Compliance**: Comprehensive audit logging and compliance reporting
- **API Management**: RESTful API for programmatic access

#### Computer Vision (`internal/vision`)
- **Detector** (`detector.go`): Visual element detection and validation in screenshots

### Configuration Structure

Test configurations are YAML files with this structure:

```yaml
name: "Test Suite Name"
output: "./output_directory"

apps:
  - name: "App Name"
    type: "web|desktop|mobile"
    url: "https://..."           # For web
    path: "/path/to/app"          # For desktop
    platform: "ios|android|..."   # For mobile

actions:
  - name: "action_identifier"
    type: "navigate|click|fill|submit|wait|screenshot|record"
    selector: "CSS selector or element identifier"
    value: "input value"
    wait_time: 3
    parameters:
      filename: "custom_name.png"

settings:
  screenshot_format: "png|jpg"
  video_format: "mp4|webm"
  quality: 85
  headless: true
  enable_metrics: true
  log_level: "debug|info|warn|error"

  # AI Testing
  ai_testing:
    enable_error_detection: true
    enable_test_generation: true
    enable_vision_analysis: true
    confidence_threshold: 0.8

  # Cloud Integration
  cloud:
    provider: "aws|gcp|azure|local"
    bucket: "bucket-name"
    enable_sync: true

  # Enterprise Features
  enterprise:
    config_path: "enterprise_config.yaml"
```

## Module Dependencies

Key external dependencies:
- **go-rod/rod**: Browser automation (Chromium DevTools Protocol)
- **spf13/cobra**: CLI framework
- **spf13/viper**: Configuration management
- **sirupsen/logrus**: Structured logging
- **stretchr/testify**: Testing assertions

## Development Guidelines

### Adding New Platform Support

1. Implement the `Platform` interface in `internal/platforms/`
2. Add the platform type to `PlatformFactory.CreatePlatform()`
3. Update config validation in `internal/config/config.go`
4. Add tests to `internal/platforms/platform_test.go`

### Adding New Action Types

1. Add action type to `Action` struct in `internal/config/config.go`
2. Implement action handling in `Executor.executeAction()`
3. Ensure all platforms support the action (or handle gracefully)
4. Update documentation and example configs

### Testing Strategy

- **Unit tests**: Test individual components in isolation (internal/*, cmd/*)
- **Integration tests**: Test component interactions with `-tags=integration`
- **E2E tests**: Full workflow tests with `-tags=e2e`
- **Functional tests**: Feature-specific tests with `-tags=functional`
- **Security tests**: Security validation with `-tags=security`

### Output Directory Structure

```
output/
├── screenshots/       # PNG/JPG screenshots
├── videos/           # MP4/WebM recordings
├── logs/             # Execution logs
├── report.html       # HTML test report
└── report.json       # JSON test results
```

## Important Implementation Notes

### Platform-Specific Considerations

**Web Platform**:
- Uses headless Chrome by default (configurable via `settings.headless`)
- CSS selectors for element targeting
- Automatic page load waiting after navigation

**Desktop Platform**:
- Platform detection (Windows/macOS/Linux) for native automation
- Coordinate-based clicking and keyboard input simulation
- Application path must be absolute

**Mobile Platform**:
- Supports both real devices and emulators
- Platform field determines iOS vs Android tooling
- Requires appropriate SDK setup (Android SDK, Xcode)

### Error Handling

All errors should be:
1. Logged via the logger instance
2. Captured in TestResult.Error field
3. Set TestResult.Success to false
4. Not panic - gracefully fail and continue with remaining tests

### Cloud Storage

When cloud settings are configured:
- Artifacts are automatically uploaded after test completion
- Local files are preserved unless cleanup is configured
- Supports multiple providers with unified interface

### Enterprise Integration

When enterprise settings exist:
- Initialize early in Executor.NewExecutor()
- Load from config file or inline YAML settings
- Features are optional and degrade gracefully if not configured
