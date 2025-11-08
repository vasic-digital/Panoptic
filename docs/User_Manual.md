# Panoptic User Manual

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [Supported Platforms](#supported-platforms)
6. [Actions Reference](#actions-reference)
7. [Examples](#examples)
8. [Command Line Interface](#command-line-interface)
9. [Troubleshooting](#troubleshooting)
10. [Advanced Usage](#advanced-usage)

---

## Introduction

Panoptic is a comprehensive automated testing and recording framework that enables you to test applications across multiple platforms - web, desktop, and mobile. It can capture screenshots, record videos, fill forms, click elements, and generate detailed reports of your testing sessions.

### Key Features

- **Multi-Platform Support**: Test web applications, desktop applications, and mobile apps/emulators
- **Automated Interaction**: Navigate, click, fill forms, and submit data automatically
- **Media Capture**: Take screenshots and record video of test sessions
- **Flexible Configuration**: YAML-based configuration for complex test scenarios
- **Detailed Reporting**: Generate comprehensive HTML reports with metrics
- **Logging**: Extensive logging of all actions and system metrics

---

## Installation

### Prerequisites

- Go 1.21 or higher
- Platform-specific tools (see [Platform Requirements](#platform-requirements))

### Basic Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd Panoptic
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build the application**:
   ```bash
   go build -o panoptic main.go
   ```

4. **Verify installation**:
   ```bash
   ./panoptic --help
   ```

### Platform Requirements

#### Web Applications
- Chrome or Chromium browser
- Optional: Headless mode support

#### Desktop Applications
- **macOS**: Built-in system utilities
- **Windows**: PowerShell access
- **Linux**: ImageMagick and X11 tools

#### Mobile Applications
- **Android**:
  - Android SDK
  - ADB (Android Debug Bridge)
- **iOS**:
  - Xcode command-line tools
  - iOS Simulator

---

## Quick Start

### 1. Create Your First Test

Create a simple configuration file `test.yaml`:

```yaml
name: "My First Test"
apps:
  - name: "Example Website"
    type: "web"
    url: "https://example.com"
actions:
  - name: "navigate_to_site"
    type: "navigate"
    value: "https://example.com"
  - name: "wait_for_load"
    type: "wait"
    wait_time: 3
  - name: "take_screenshot"
    type: "screenshot"
```

### 2. Run the Test

```bash
./panoptic run test.yaml
```

### 3. View Results

After execution, check the `./output` directory:
- `screenshots/`: Captured images
- `videos/`: Recorded videos  
- `logs/`: Execution logs
- `report.html`: Interactive test report

---

## Configuration

### Basic Structure

```yaml
name: "Test Suite Name"
output: "./output_directory"    # Optional: override default output

apps:                          # Applications to test
  - name: "App Name"
    type: "web|desktop|mobile"
    # Platform-specific fields...

actions:                        # Actions to perform
  - name: "action_name"
    type: "action_type"
    # Action-specific fields...

settings:                        # Global settings
  screenshot_format: "png|jpg"
  video_format: "mp4|webm"
  quality: 80                     # 1-100
  headless: false                  # For web apps
  window_width: 1920
  window_height: 1080
  enable_metrics: true
  log_level: "debug|info|warn|error"
```

### Application Configuration

#### Web Application
```yaml
- name: "Web App"
  type: "web"
  url: "https://example.com"
  timeout: 30                     # Optional: timeout in seconds
```

#### Desktop Application
```yaml
- name: "Desktop App"
  type: "desktop"
  path: "/Applications/Calculator.app"  # macOS
  # path: "C:\\Windows\\System32\\notepad.exe"  # Windows
  # path: "/usr/bin/gedit"                   # Linux
  timeout: 30
  environment:                     # Optional: environment variables
    VAR1: "value1"
    VAR2: "value2"
```

#### Mobile Application
```yaml
- name: "Mobile App"
  type: "mobile"
  platform: "android|ios"
  emulator: true                  # Use emulator/simulator
  device: "emulator-5554"        # Device ID
  timeout: 30
```

### Settings Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `screenshot_format` | string | "png" | Image format for screenshots |
| `video_format` | string | "mp4" | Video format for recordings |
| `quality` | int | 80 | Media quality (1-100) |
| `headless` | boolean | false | Run browser in headless mode |
| `window_width` | int | 1920 | Browser window width |
| `window_height` | int | 1080 | Browser window height |
| `enable_metrics` | boolean | true | Collect performance metrics |
| `log_level` | string | "info" | Logging verbosity |

---

## Supported Platforms

### Web Applications
- **Browser**: Chrome/Chromium
- **Selectors**: CSS selectors
- **Features**: Navigation, clicking, form filling, screenshots, recording
- **Headless Mode**: Supported

### Desktop Applications
- **macOS**: Uses `screencapture`, AppleScript, system utilities
- **Windows**: Uses PowerShell, Windows API
- **Linux**: Uses X11 tools, ImageMagick
- **Features**: Application launching, screenshots, basic automation
- **Limitations**: Platform-specific, requires accessible applications

### Mobile Applications
- **Android**: Uses ADB, screenrecord, screencap
- **iOS**: Uses Xcode simctl, simulator tools
- **Features**: Device control, screenshots, screen recording
- **Requirements**: Platform tools installed and configured

---

## Actions Reference

### Navigation Actions

#### Navigate
Navigate to a URL or open a specific application view.

```yaml
- name: "navigate_to_page"
  type: "navigate"
  value: "https://example.com"    # URL to navigate to
```

### Interaction Actions

#### Click
Click on an element using CSS selector or coordinates.

```yaml
- name: "click_button"
  type: "click"
  selector: "#submit-button"        # CSS selector
  # target: "button_id"            # Alternative to selector
```

#### Fill
Fill form fields with specified values.

```yaml
- name: "fill_username"
  type: "fill"
  selector: "input[name='username']"
  value: "testuser"
```

#### Submit
Submit forms or trigger submit actions.

```yaml
- name: "submit_form"
  type: "submit"
  selector: "form.login"         # Form selector (optional)
```

### Wait Actions

#### Wait
Wait for specified duration.

```yaml
- name: "wait_for_load"
  type: "wait"
  wait_time: 3                    # Seconds to wait
```

### Media Capture Actions

#### Screenshot
Capture screenshot of current state.

```yaml
- name: "capture_page"
  type: "screenshot"
  parameters:
    filename: "custom_name.png"   # Optional: custom filename
```

#### Record
Start video recording for specified duration.

```yaml
- name: "record_session"
  type: "record"
  duration: 30                    # Recording duration in seconds
  parameters:
    filename: "session.mp4"        # Optional: custom filename
```

---

## Examples

### Web Application Testing

```yaml
name: "Web Login Test"
apps:
  - name: "Login Page"
    type: "web"
    url: "https://demo.testfire.net/bank/login.aspx"

actions:
  - name: "navigate_to_login"
    type: "navigate"
    value: "https://demo.testfire.net/bank/login.aspx"

  - name: "wait_for_page"
    type: "wait"
    wait_time: 2

  - name: "fill_username"
    type: "fill"
    selector: "input[name='uid']"
    value: "admin"

  - name: "fill_password"
    type: "fill"
    selector: "input[name='passw']"
    value: "admin"

  - name: "submit_login"
    type: "submit"
    selector: "input[type='submit']"

  - name: "wait_for_dashboard"
    type: "wait"
    wait_time: 3

  - name: "capture_dashboard"
    type: "screenshot"
    parameters:
      filename: "dashboard_after_login.png"
```

### Multi-Platform Testing

```yaml
name: "Cross-Platform Test"
output: "./multi-platform-results"

apps:
  - name: "Web Version"
    type: "web"
    url: "https://example.com"

  - name: "Desktop Version"
    type: "desktop"
    path: "/Applications/MyApp.app"

  - name: "Mobile Version"
    type: "mobile"
    platform: "android"
    emulator: true
    device: "emulator-5554"

actions:
  - name: "test_web_flow"
    type: "screenshot"
    parameters:
      filename: "web_initial.png"

  - name: "test_desktop_flow"
    type: "screenshot"
    parameters:
      filename: "desktop_initial.png"

  - name: "test_mobile_flow"
    type: "screenshot"
    parameters:
      filename: "mobile_initial.png"

settings:
  screenshot_format: "png"
  quality: 90
  enable_metrics: true
```

### Recording Workflow

```yaml
name: "User Session Recording"
apps:
  - name: "Web Application"
    type: "web"
    url: "https://example.com"

actions:
  - name: "start_recording"
    type: "record"
    duration: 60
    parameters:
      filename: "user_journey.mp4"

  - name: "navigate_home"
    type: "navigate"
    value: "https://example.com"

  - name: "wait_load"
    type: "wait"
    wait_time: 3

  - name: "click_products"
    type: "click"
    selector: "#products-link"

  - name: "wait_products"
    type: "wait"
    wait_time: 2

  - name: "capture_products_page"
    type: "screenshot"
    parameters:
      filename: "products_page.png"

settings:
  video_format: "mp4"
  quality: 85
  window_width: 1920
  window_height: 1080
```

---

## Command Line Interface

### Global Options

```bash
./panoptic [global-options] <command> [command-options]
```

#### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--config` | `-c` | Configuration file path | `~/.panoptic.yaml` |
| `--output` | `-o` | Output directory | `./output` |
| `--verbose` | `-v` | Enable verbose logging | `false` |
| `--help` | `-h` | Show help | |

### Commands

#### run
Execute automated testing and recording.

```bash
./panoptic run [config-file] [options]
```

**Arguments:**
- `config-file`: Path to YAML configuration file (required)

**Example:**
```bash
# Basic usage
./panoptic run test.yaml

# With custom output directory
./panoptic run test.yaml --output ./my-results

# With verbose logging
./panoptic run test.yaml --verbose

# Combined options
./panoptic run test.yaml --output ./results --verbose
```

#### help
Show help information.

```bash
./panoptic help [command]

# Show help for run command
./panoptic help run
```

### Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Platform/dependency error |

---

## Troubleshooting

### Common Issues

#### Browser Not Found
**Error**: `Browser not available`

**Solution**: 
- Install Chrome or Chromium
- For headless mode, ensure Chrome supports it
- Check browser PATH

#### Configuration File Not Found
**Error**: `Failed to load configuration: no such file`

**Solution**:
- Verify file path is correct
- Check file permissions
- Use absolute path if needed

#### Mobile Tools Not Available
**Error**: `platform tools not available`

**Solution**:
- Install Android SDK for Android testing
- Install Xcode tools for iOS testing
- Verify ADB/Xcode tools in PATH

#### Permission Denied
**Error**: `Permission denied` when creating output

**Solution**:
- Check directory permissions
- Use different output directory
- Run with appropriate permissions

### Debug Mode

Enable verbose logging for detailed troubleshooting:

```bash
./panoptic run test.yaml --verbose
```

### Log Analysis

Check logs in `output/logs/panoptic.log`:

```bash
# View real-time logs
tail -f output/logs/panoptic.log

# Search for errors
grep "ERROR" output/logs/panoptic.log

# Search for specific app
grep "Web App" output/logs/panoptic.log
```

### Performance Issues

#### Slow Execution
- Reduce wait times in configuration
- Use headless mode for web tests
- Close unnecessary applications

#### Memory Usage
- Use shorter recording durations
- Reduce concurrent applications
- Monitor system resources

---

## Advanced Usage

### Custom Selectors

#### CSS Selectors
```yaml
# ID selector
selector: "#submit-button"

# Class selector
selector: ".login-form"

# Attribute selector
selector: "input[name='username']"

# Complex selector
selector: "form.login > div.input-group > input[type='text']"
```

#### XPath (Limited Support)
```yaml
# XPath expressions
selector: "//button[contains(text(), 'Submit')]"
```

### Environment Variables

Use environment variables in configurations:

```yaml
apps:
  - name: "Test App"
    type: "web"
    url: "${TEST_URL}/login"     # Environment variable substitution

actions:
  - name: "fill_user"
    type: "fill"
    selector: "#username"
    value: "${TEST_USER}"        # Environment variable substitution
```

Set variables before running:
```bash
export TEST_URL="https://example.com"
export TEST_USER="testuser"
./panoptic run test.yaml
```

### Conditional Logic

Use multiple configurations for different scenarios:

```bash
# Development environment
./panoptic run config-dev.yaml --output ./dev-results

# Production environment  
./panoptic run config-prod.yaml --output ./prod-results

# Mobile testing only
./panoptic run config-mobile.yaml
```

### Batch Testing

Create shell scripts for batch execution:

```bash
#!/bin/bash
# test-all.sh

TESTS=("config-web.yaml" "config-desktop.yaml" "config-mobile.yaml")
RESULTS_BASE="./batch-results"

for test in "${TESTS[@]}"; do
    echo "Running $test..."
    ./panoptic run "$test" --output "$RESULTS_BASE/$(basename $test .yaml)"
done
```

### Integration with CI/CD

#### GitHub Actions
```yaml
name: Panoptic Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y chromium-browser
    - name: Run Panoptic
      run: |
        go build -o panoptic main.go
        ./panoptic run test-config.yaml --verbose
    - name: Upload results
      uses: actions/upload-artifact@v2
      with:
        name: test-results
        path: output/
```

#### Jenkins Pipeline
```groovy
pipeline {
    agent any
    stages {
        stage('Test') {
            steps {
                sh 'go build -o panoptic main.go'
                sh './panoptic run test-config.yaml --output ./jenkins-results'
                archiveArtifacts artifacts: 'jenkins-results/**', allowEmptyArchive: true
            }
        }
    }
}
```

### Custom Reports

Extend HTML reporting by modifying templates or generating custom formats:

```bash
# Generate JSON report for integration
./panoptic run test.yaml
python generate-custom-report.py output/report.html output/custom-report.json
```

### Performance Optimization

#### Parallel Execution
Run multiple test suites simultaneously:

```bash
# Run different configs in parallel
./panoptic run web-tests.yaml &
./panoptic run desktop-tests.yaml &
./panoptic run mobile-tests.yaml &
wait  # Wait for all to complete
```

#### Resource Management
Monitor system resources during testing:

```bash
# Monitor memory usage
watch -n 1 'ps aux | grep panoptic'

# Monitor disk usage
du -sh output/
```

---

## Tips and Best Practices

### Configuration Best Practices

1. **Use Descriptive Names**: Clear names for apps and actions
2. **Organize Actions**: Group related actions together
3. **Set Appropriate Timeouts**: Balance reliability and speed
4. **Use Environment Variables**: Separate configuration from secrets

### Testing Best Practices

1. **Start Simple**: Begin with basic navigation and screenshots
2. **Incremental Complexity**: Add actions gradually
3. **Test Locally First**: Verify before CI/CD integration
4. **Monitor Logs**: Use verbose logging for debugging

### Performance Tips

1. **Headless Mode**: Faster for web testing
2. **Optimize Wait Times**: Minimize unnecessary waits
3. **Resource Limits**: Don't overload system resources
4. **Cleanup**: Remove old test results regularly

### Security Considerations

1. **No Credentials in Config**: Use environment variables
2. **Secure Test Data**: Use dummy/test data
3. **Network Isolation**: Test in sandboxed environment
4. **Permission Management**: Run with minimum required permissions

---

## Support and Contributing

### Getting Help

- **Documentation**: Check this manual and inline help (`./panoptic --help`)
- **Issues**: Report bugs at the project repository
- **Community**: Join discussions and ask questions

### Contributing

1. **Fork Repository**: Create your own fork
2. **Create Branch**: Develop your feature or fix
3. **Add Tests**: Ensure test coverage
4. **Submit Pull Request**: With clear description

### Development Setup

```bash
# Clone your fork
git clone <your-fork-url>
cd Panoptic

# Install dependencies
go mod tidy

# Run tests
./scripts/test.sh --coverage

# Build for development
go build -o panoptic main.go
```

---

## Glossary

| Term | Definition |
|-------|------------|
| **Platform** | Application type (web, desktop, mobile) |
| **Action** | Single operation performed during testing |
| **Selector** | CSS selector or XPath for element identification |
| **Headless Mode** | Browser operation without visible UI |
| **Emulator** | Software simulation of mobile device |
| **Configuration** | YAML file defining test parameters |
| **Report** | HTML/JSON output summarizing test results |

---

*This manual covers Panoptic version 1.0.0. For the latest updates and additional information, visit the project documentation.*