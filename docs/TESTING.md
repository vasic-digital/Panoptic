# Panoptic Test Suite Documentation

This document provides comprehensive information about the Panoptic test suite, including unit tests, integration tests, end-to-end tests, and test coverage reporting.

## Overview

The Panoptic test suite is designed to ensure the reliability and correctness of the automated testing framework across all supported platforms (web, desktop, mobile). The test suite consists of multiple layers:

- **Unit Tests**: Test individual functions and components in isolation
- **Integration Tests**: Test interactions between components
- **End-to-End Tests**: Test complete workflows from CLI execution to report generation

## Test Structure

```
tests/
├── integration/          # Integration tests
│   └── panoptic_test.go  # Platform integration tests
├── e2e/                  # End-to-end tests
│   └── panoptic_test.go  # Full workflow tests
internal/
├── config/
│   ├── config.go        # Configuration management
│   └── config_test.go   # Unit tests for config
├── executor/
│   ├── executor.go       # Test execution engine
│   └── executor_test.go  # Unit tests for executor
├── logger/
│   ├── logger.go         # Logging functionality
│   └── logger_test.go    # Unit tests for logger
├── platforms/
│   ├── platform.go      # Platform interface
│   ├── web.go          # Web platform implementation
│   ├── desktop.go      # Desktop platform implementation
│   ├── mobile.go       # Mobile platform implementation
│   └── platform_test.go # Unit tests for platforms
cmd/
├── root.go             # CLI root command
├── run.go              # Run command implementation
└── cmd_test.go         # CLI tests
scripts/
├── test.sh             # Test runner script
└── coverage.sh         # Coverage analysis script
```

## Running Tests

### Quick Start

Run all tests with coverage:
```bash
./scripts/test.sh --coverage
```

Run only unit tests:
```bash
./scripts/test.sh --skip-integration --skip-e2e
```

Run with verbose output:
```bash
./scripts/test.sh -v --coverage
```

### Manual Test Execution

#### Unit Tests
```bash
# Run all unit tests
go test ./internal/... ./cmd/...

# Run specific package tests
go test ./internal/config
go test ./internal/executor
go test ./internal/platforms
go test ./internal/logger
go test ./cmd

# Run with coverage
go test -coverprofile=coverage.out ./...
```

#### Integration Tests
```bash
# Run integration tests (requires build tags)
go test -tags=integration ./tests/integration/...

# Run with coverage
go test -tags=integration -coverprofile=integration_coverage.out ./tests/integration/...
```

#### End-to-End Tests
```bash
# Run e2e tests (requires build tags and external dependencies)
go test -tags=e2e ./tests/e2e/...

# Run with coverage
go test -tags=e2e -coverprofile=e2e_coverage.out ./tests/e2e/...
```

## Test Coverage

### Coverage Requirements

- **Minimum Coverage**: 75% for all packages
- **Target Coverage**: 85% for core packages
- **Excellent Coverage**: 90%+ for critical components

### Coverage Analysis

Generate detailed coverage report:
```bash
./scripts/coverage.sh
```

Generate HTML coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Coverage by package:
```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out
```

## Test Categories

### Unit Tests

Unit tests focus on individual components:

#### Configuration Tests (`internal/config/config_test.go`)
- Test YAML parsing and validation
- Test default value application
- Test error handling for invalid configurations
- Test edge cases (empty configs, special characters)

#### Platform Tests (`internal/platforms/platform_test.go`)
- Test platform factory creation
- Test platform initialization
- Test individual actions (navigate, click, fill, submit, wait)
- Test screenshot and recording functionality
- Test metrics collection
- Test error handling for unavailable platforms

#### Executor Tests (`internal/executor/executor_test.go`)
- Test executor initialization
- Test action execution across different platforms
- Test report generation
- Test error handling and recovery
- Test output directory structure

#### Logger Tests (`internal/logger/logger_test.go`)
- Test logger creation with different verbosity levels
- Test file output and directory creation
- Test concurrent logging
- Test formatted logging and special characters
- Test long messages and error cases

#### CLI Tests (`cmd/cmd_test.go`)
- Test command registration and flag handling
- Test configuration file processing
- Test command-line argument parsing
- Test help system
- Test error scenarios

### Integration Tests

Integration tests (`tests/integration/panoptic_test.go`) test component interactions:

#### Platform Integration
- Web browser automation integration
- Desktop application control integration
- Mobile device/emulator integration
- Platform tool availability checks

#### CLI Integration
- Command-line interface functionality
- Configuration file processing
- Output directory creation
- Report generation workflow

#### Workflow Integration
- End-to-end test execution
- Multi-platform testing scenarios
- Error handling and recovery
- Metrics collection and reporting

### End-to-End Tests

E2E tests (`tests/e2e/panoptic_test.go`) test complete user workflows:

#### Full Workflow Test
- Complete test execution from config to report
- Multi-platform testing (web, desktop, mobile)
- Comprehensive action sequences
- Output verification

#### Recording Workflow Test
- Video recording functionality
- Screenshot capture during recording
- File generation and validation
- Performance metrics during recording

#### Error Handling Test
- Graceful failure handling
- Invalid configuration handling
- Platform unavailability scenarios
- Partial failure recovery

#### Performance Metrics Test
- Timing and performance measurement
- Metrics collection accuracy
- Resource usage monitoring
- Report generation validation

## Test Data and Fixtures

### Configuration Files

Test configurations are generated programmatically in tests:

```yaml
name: "Test Config"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "test_action"
    type: "wait"
    wait_time: 1
```

### Test Applications

Tests use publicly available test services:
- **HTTPBin** (`https://httpbin.org`) - HTTP testing service
- **System Applications** - Calculator, Notepad for desktop testing
- **Emulated Mobile** - Android/iOS simulators when available

## Test Environment Setup

### Prerequisites

#### For Unit Tests
- Go 1.21+
- Go testing dependencies (testify)

#### For Integration Tests
- Chrome/Chromium (for web testing)
- System applications (Calculator, Notepad)
- ADB (for Android testing)
- Xcode tools (for iOS testing)

#### For E2E Tests
- All integration test prerequisites
- Sufficient system resources for browser automation
- Network access for web testing

### Environment Variables

```bash
# Test configuration
export PANOPTIC_TEST_TIMEOUT=30s
export PANOPTIC_TEST_VERBOSE=true

# Browser configuration
export PANOPTIC_BROWSER_HEADLESS=true
export PANOPTIC_BROWSER_WIDTH=1920
export PANOPTIC_BROWSER_HEIGHT=1080

# Mobile testing
export ANDROID_SDK_ROOT=/path/to/android/sdk
export IOS_SIMULATOR_PATH=/path/to/simulator
```

## Test Scripts

### `scripts/test.sh`

Main test runner script with options:

```bash
Usage: ./scripts/test.sh [options]

Options:
  -v, --verbose         Enable verbose output
  --skip-integration    Skip integration tests
  --skip-e2e           Skip end-to-end tests
  --coverage           Generate coverage profile
  --race               Enable race detector
  -h, --help           Show help message
```

### `scripts/coverage.sh`

Coverage analysis script with reporting:

- Generates coverage profiles
- Creates HTML coverage reports
- Analyzes coverage by package
- Identifies low-coverage areas
- Tracks coverage trends
- Provides improvement recommendations

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
    - name: Run tests
      run: ./scripts/test.sh --coverage
    - name: Upload coverage
      uses: codecov/codecov-action@v1
```

### CI Requirements

- Run full test suite on every push/PR
- Maintain minimum coverage threshold
- Generate and upload coverage reports
- Fail on test failures or coverage regression

## Test Best Practices

### Writing Tests

1. **Descriptive Names**: Use clear, descriptive test names
2. **Table-Driven Tests**: Use table-driven tests for multiple scenarios
3. **Setup/Teardown**: Proper setup and teardown in each test
4. **Error Testing**: Test both success and failure scenarios
5. **Edge Cases**: Test boundary conditions and edge cases

### Test Organization

1. **Package Structure**: Mirror source code structure
2. **Test Files**: Name test files as `*_test.go`
3. **Build Tags**: Use build tags for integration/e2e tests
4. **Test Data**: Generate test data programmatically
5. **Cleanup**: Ensure proper cleanup in all tests

### Coverage Guidelines

1. **Critical Paths**: Focus on critical execution paths
2. **Error Handling**: Test all error conditions
3. **Edge Cases**: Cover boundary conditions
4. **Integration Points**: Test component interactions
5. **Platform Differences**: Test all supported platforms

## Troubleshooting

### Common Issues

#### Browser Not Available
```bash
Error: Browser not available
Solution: Install Chrome/Chromium or set PANOPTIC_BROWSER_HEADLESS=false
```

#### Mobile Tools Missing
```bash
Error: platform tools not available
Solution: Install Android SDK or iOS development tools
```

#### Coverage Generation Failed
```bash
Error: Coverage file not generated
Solution: Check test execution and ensure all tests pass
```

### Debug Mode

Enable debug mode for troubleshooting:

```bash
# Verbose test output
./scripts/test.sh -v

# Debug logging
export PANOPTIC_LOG_LEVEL=debug

# Race detection
./scripts/test.sh --race
```

### Test Isolation

Ensure tests are isolated:

```bash
# Clean test environment
go clean -testcache

# Run specific test
go test -run TestFunctionName ./package

# Run tests in parallel
go test -parallel 4 ./package
```

## Performance Considerations

### Test Execution Time

- Unit tests: < 5 seconds
- Integration tests: < 30 seconds  
- E2E tests: < 2 minutes

### Resource Usage

- Memory: Tests may use up to 1GB during browser automation
- CPU: Consider using parallel execution with limits
- Disk: Test outputs may consume several hundred MB

### Optimization

```bash
# Parallel test execution
go test -parallel $(nproc) ./...

# Skip heavy tests in CI
./scripts/test.sh --skip-e2e

# Use race detector selectively
go test -race ./critical/packages/...
```

## Contributing Tests

### Adding New Tests

1. **Unit Tests**: Add alongside source code
2. **Integration Tests**: Add to `tests/integration/`
3. **E2E Tests**: Add to `tests/e2e/`
4. **Update Documentation**: Update this file and test data
5. **Verify Coverage**: Ensure minimum coverage is maintained

### Test Review Checklist

- [ ] Tests cover all new functionality
- [ ] Tests include both success and failure scenarios
- [ ] Tests are properly isolated and clean up after themselves
- [ ] Test names are descriptive and follow conventions
- [ ] Coverage requirements are met
- [ ] Documentation is updated

This comprehensive test suite ensures the Panoptic application remains reliable, maintainable, and performs correctly across all supported platforms and use cases.