# Panoptic Troubleshooting Guide

**Version**: 1.0
**Last Updated**: 2025-11-11
**Target Audience**: Users, Operators, Support Engineers

---

## Table of Contents

1. [General Troubleshooting Workflow](#general-troubleshooting-workflow)
2. [Installation Issues](#installation-issues)
3. [Configuration Issues](#configuration-issues)
4. [Web Platform Issues](#web-platform-issues)
5. [Desktop Platform Issues](#desktop-platform-issues)
6. [Mobile Platform Issues](#mobile-platform-issues)
7. [Cloud Storage Issues](#cloud-storage-issues)
8. [Enterprise Features Issues](#enterprise-features-issues)
9. [Performance Issues](#performance-issues)
10. [Common Error Messages](#common-error-messages)
11. [Diagnostic Commands](#diagnostic-commands)
12. [Getting Help](#getting-help)

---

## General Troubleshooting Workflow

### Step 1: Gather Information

```bash
# Check version
./panoptic --version

# Check system info
uname -a  # Linux/macOS
systeminfo  # Windows

# Check Go version
go version

# Check available disk space
df -h

# Check memory usage
free -h  # Linux
vm_stat  # macOS
```

### Step 2: Review Logs

```bash
# Application logs
tail -100 /opt/panoptic/logs/panoptic.log

# Error logs
tail -100 /opt/panoptic/logs/panoptic-error.log

# System logs
journalctl -u panoptic -n 100  # systemd
docker logs panoptic  # Docker
kubectl logs -f deployment/panoptic -n panoptic  # Kubernetes
```

### Step 3: Verify Configuration

```bash
# Validate configuration syntax
./panoptic validate /path/to/test_config.yaml

# Dry run (no actual execution)
./panoptic run --dry-run /path/to/test_config.yaml

# Verbose mode
./panoptic run -v /path/to/test_config.yaml
```

### Step 4: Check Dependencies

```bash
# Browser (for web automation)
which chromium-browser
chromium-browser --version

# Mobile tools
which adb  # Android
which xcrun  # iOS

# Network connectivity
curl -I https://httpbin.org/get
```

---

## Installation Issues

### Issue: Go Version Too Old

**Symptoms**:
```
go: go.mod requires go >= 1.21, but running go 1.20
```

**Solution**:
```bash
# Install Go 1.22 or later
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
```

### Issue: Build Fails with Missing Dependencies

**Symptoms**:
```
package github.com/go-rod/rod: cannot find package
```

**Solution**:
```bash
# Download dependencies
go mod download

# Tidy up go.mod
go mod tidy

# Rebuild
go build -o panoptic main.go
```

### Issue: Permission Denied

**Symptoms**:
```
-bash: ./panoptic: Permission denied
```

**Solution**:
```bash
# Make executable
chmod +x ./panoptic

# Or run with go
go run main.go run test_config.yaml
```

---

## Configuration Issues

### Issue: YAML Syntax Error

**Symptoms**:
```
ERROR: Failed to load config: yaml: unmarshal errors:
  line 10: cannot unmarshal !!str `invalid` into []config.AppConfig
```

**Solution**:
```bash
# Validate YAML syntax
yamllint test_config.yaml

# Check indentation (must be spaces, not tabs)
cat -A test_config.yaml | grep -E '\t'

# Use example as template
cp examples/test_config.yaml my_config.yaml
```

### Issue: Invalid App Type

**Symptoms**:
```
ERROR: Failed to create platform: unsupported platform type: webapp
```

**Solution**:
```yaml
# Valid app types: "web", "desktop", "mobile"
apps:
  - name: "My App"
    type: "web"  # Not "webapp"
    url: "https://example.com"
```

### Issue: Missing Required Fields

**Symptoms**:
```
ERROR: app 'Test App' is missing required field: url
```

**Solution**:
```yaml
# Web apps require 'url'
apps:
  - name: "Web App"
    type: "web"
    url: "https://example.com"  # Required!

# Desktop apps require 'path'
apps:
  - name: "Desktop App"
    type: "desktop"
    path: "/path/to/app"  # Required!

# Mobile apps require 'platform'
apps:
  - name: "Mobile App"
    type: "mobile"
    platform: "android"  # Required! (android or ios)
```

### Issue: Invalid Action Type

**Symptoms**:
```
ERROR: Unknown action type: 'clickbutton'
```

**Solution**:
```yaml
# Valid action types:
# navigate, click, fill, submit, wait, screenshot, record,
# cloud_sync, cloud_analytics, distributed_test, cloud_cleanup,
# user_create, project_create, etc.

actions:
  - name: "click_button"
    type: "click"  # Not "clickbutton"
    selector: "#submit-btn"
```

---

## Web Platform Issues

### Issue: Browser Not Found

**Symptoms**:
```
ERROR: Failed to initialize web platform: browser not found
```

**Solution**:
```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser chromium-chromedriver

# Fedora/RHEL
sudo dnf install chromium chromium-chromedriver

# macOS
brew install --cask chromium

# Verify installation
which chromium-browser
chromium-browser --version
```

### Issue: Chrome/Chromium Version Mismatch

**Symptoms**:
```
ERROR: ChromeDriver version 120 does not match Chrome version 119
```

**Solution**:
```bash
# Check versions
chromium-browser --version
chromedriver --version

# Update ChromeDriver to match
# Download from: https://chromedriver.chromium.org/downloads

# Or reinstall both
sudo apt-get update
sudo apt-get install --reinstall chromium-browser chromium-chromedriver
```

### Issue: Headless Mode Fails on Linux

**Symptoms**:
```
ERROR: Failed to start browser: display not found
```

**Solution**:
```bash
# Install Xvfb for virtual display
sudo apt-get install xvfb

# Run with Xvfb
xvfb-run -a ./panoptic run test_config.yaml

# Or configure systemd service with Xvfb
```

**Or configure headless mode**:
```yaml
settings:
  headless: true  # Enable headless mode
```

### Issue: Element Not Found

**Symptoms**:
```
ERROR: Action 'click' failed: element not found: #submit-button
```

**Solutions**:

1. **Verify selector syntax**:
```yaml
# CSS Selectors
selector: "#id"           # ID
selector: ".class"        # Class
selector: "div.class"     # Tag + Class
selector: "[name='btn']"  # Attribute
```

2. **Add wait time before action**:
```yaml
actions:
  - name: "wait_for_page"
    type: "wait"
    wait_time: 3  # Wait 3 seconds

  - name: "click_button"
    type: "click"
    selector: "#submit-button"
```

3. **Take screenshot to debug**:
```yaml
actions:
  - name: "debug_screenshot"
    type: "screenshot"
    parameters:
      filename: "debug.png"
```

### Issue: Navigation Timeout

**Symptoms**:
```
ERROR: Navigation failed: timeout waiting for page load
```

**Solutions**:

1. **Increase timeout**:
```yaml
settings:
  page_timeout: 60  # Increase to 60 seconds
```

2. **Check network connectivity**:
```bash
curl -I https://your-target-site.com
```

3. **Check for redirects**:
```bash
curl -L -I https://your-target-site.com
```

---

## Desktop Platform Issues

### Issue: Application Not Found

**Symptoms**:
```
ERROR: Failed to initialize platform: application not found at path: /Applications/Calculator.app
```

**Solution**:
```bash
# Verify application path
ls -la /Applications/Calculator.app

# macOS: Use full .app bundle path
path: "/Applications/Calculator.app"

# Windows: Use .exe path
path: "C:\\Program Files\\MyApp\\app.exe"

# Linux: Use binary path
path: "/usr/bin/myapp"
```

### Issue: Application Won't Launch

**Symptoms**:
```
ERROR: Failed to launch application: exec: permission denied
```

**Solutions**:

1. **Check permissions**:
```bash
# Make executable
chmod +x /path/to/application

# Check ownership
ls -la /path/to/application
```

2. **Check if app is already running**:
```bash
# macOS/Linux
ps aux | grep myapp

# Kill if needed
pkill myapp
```

### Issue: Screenshot Fails

**Symptoms**:
```
ERROR: Failed to capture screenshot: screencapture: not found
```

**Solution**:
```bash
# macOS: screencapture should be available by default
which screencapture

# Linux: Install scrot or import (ImageMagick)
sudo apt-get install scrot
# Or
sudo apt-get install imagemagick

# Windows: No additional tools needed
```

---

## Mobile Platform Issues

### Issue: ADB Not Found (Android)

**Symptoms**:
```
ERROR: Android automation tools not available: adb not found
```

**Solution**:
```bash
# Install Android SDK Platform Tools
# Ubuntu/Debian
sudo apt-get install android-tools-adb android-tools-fastboot

# macOS
brew install android-platform-tools

# Verify
which adb
adb version

# Add to PATH if needed
export PATH=$PATH:$HOME/Android/Sdk/platform-tools
```

### Issue: No Devices Found

**Symptoms**:
```
ERROR: No Android devices available
```

**Solutions**:

1. **List connected devices**:
```bash
adb devices
```

2. **Start ADB server**:
```bash
adb kill-server
adb start-server
adb devices
```

3. **Enable USB debugging** on device:
   - Settings → About Phone → Tap "Build Number" 7 times
   - Settings → Developer Options → Enable "USB Debugging"

4. **For emulators**:
```bash
# List emulators
emulator -list-avds

# Start emulator
emulator -avd <avd_name>

# Wait for device
adb wait-for-device
```

### Issue: iOS Simulator Not Found

**Symptoms**:
```
ERROR: iOS simulation tools not available
```

**Solution**:
```bash
# Verify Xcode installation
xcode-select -p

# Install Xcode Command Line Tools
xcode-select --install

# List available simulators
xcrun simctl list devices

# Boot simulator
xcrun simctl boot "iPhone 15 Pro"
```

### Issue: Device Authorization Failed

**Symptoms**:
```
ERROR: device unauthorized
```

**Solution**:
1. Check device screen for authorization prompt
2. Accept "Allow USB Debugging"
3. Check "Always allow from this computer"
4. Run again:
```bash
adb devices
```

---

## Cloud Storage Issues

### Issue: AWS S3 Access Denied

**Symptoms**:
```
ERROR: Failed to upload to S3: AccessDenied
```

**Solutions**:

1. **Verify credentials**:
```bash
# Check credentials
aws configure list

# Test access
aws s3 ls s3://your-bucket/
```

2. **Verify IAM permissions**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:DeleteObject"
      ],
      "Resource": [
        "arn:aws:s3:::your-bucket",
        "arn:aws:s3:::your-bucket/*"
      ]
    }
  ]
}
```

3. **Check bucket configuration**:
```yaml
settings:
  cloud:
    provider: "aws"
    bucket: "your-bucket"  # Bucket name only, no s3:// prefix
    region: "us-east-1"    # Must match bucket region
```

### Issue: GCP Cloud Storage Failed

**Symptoms**:
```
ERROR: Failed to upload to GCS: permission denied
```

**Solution**:
```bash
# Verify service account
gcloud auth list

# Test access
gsutil ls gs://your-bucket/

# Verify service account has Storage Object Admin role
gcloud projects get-iam-policy YOUR_PROJECT_ID
```

### Issue: Network Timeout

**Symptoms**:
```
ERROR: Failed to upload: context deadline exceeded
```

**Solutions**:

1. **Check network connectivity**:
```bash
curl -I https://s3.amazonaws.com
```

2. **Increase timeout**:
```yaml
settings:
  cloud:
    timeout: 300  # 5 minutes
```

3. **Check proxy settings**:
```bash
echo $HTTP_PROXY
echo $HTTPS_PROXY
```

---

## Enterprise Features Issues

### Issue: Enterprise Config Not Found

**Symptoms**:
```
ERROR: Enterprise configuration file not found: /path/to/enterprise_config.yaml
```

**Solution**:
```yaml
# Create enterprise config
cp examples/enterprise_config.yaml /opt/panoptic/config/

# Update test config
settings:
  enterprise:
    config_path: "/opt/panoptic/config/enterprise_config.yaml"
```

### Issue: License Validation Failed

**Symptoms**:
```
ERROR: License validation failed: license expired
```

**Solution**:
```yaml
# Update license in enterprise_config.yaml
license:
  type: "enterprise"
  max_users: 1000
  expiration_date: "2030-12-31T23:59:59Z"  # Update this
  features:
    - "distributed_testing"
    - "cloud_storage"
    - "audit_logging"
```

### Issue: User Authentication Failed

**Symptoms**:
```
ERROR: Authentication failed: invalid credentials
```

**Solutions**:

1. **Reset user password**:
```bash
# This would be done through enterprise admin interface
# or by directly updating users.json (development only)
```

2. **Check user exists**:
```yaml
# In enterprise action
actions:
  - name: "create_user"
    type: "user_create"
    parameters:
      username: "testuser"
      email: "test@example.com"
      password: "SecureP@ss123"
      role: "admin"
```

### Issue: Audit Log Full

**Symptoms**:
```
WARNING: Audit log size exceeds 100MB
```

**Solution**:
```yaml
# Configure rotation in enterprise_config.yaml
audit:
  max_file_size_mb: 100
  retention_days: 90

# Or manually archive
actions:
  - name: "backup_audit_logs"
    type: "backup_data"
    parameters:
      backup_location: "/backup/audit"
```

---

## Performance Issues

### Issue: Tests Running Slowly

**Symptoms**:
- Tests taking much longer than expected
- High CPU or memory usage

**Solutions**:

1. **Check system resources**:
```bash
# CPU usage
top
htop

# Memory usage
free -h

# Disk I/O
iostat -x 1
```

2. **Optimize configuration**:
```yaml
settings:
  headless: true  # Faster than headed mode
  parallel_tests: 5  # Adjust based on CPU cores
  screenshot_quality: 75  # Reduce from 100
```

3. **Reduce wait times**:
```yaml
actions:
  - name: "wait"
    type: "wait"
    wait_time: 1  # Reduce from 5 if possible
```

4. **Check benchmark results**:
```bash
# Run benchmarks
go test -bench=. ./internal/...

# Look for slow operations
go test -bench=. -cpuprofile=cpu.prof ./internal/executor/
go tool pprof cpu.prof
```

### Issue: Memory Leaks

**Symptoms**:
```
ERROR: Out of memory
```

**Solutions**:

1. **Monitor memory usage**:
```bash
# Watch memory usage
watch -n 1 'ps aux | grep panoptic'
```

2. **Check for resource leaks**:
```bash
# Run with race detector
go test -race ./internal/...
```

3. **Reduce concurrent operations**:
```yaml
settings:
  max_concurrent_browsers: 3  # Reduce from 10
```

4. **Enable garbage collection**:
```bash
# Force GC
export GOGC=50  # More aggressive GC
```

### Issue: Disk Space Full

**Symptoms**:
```
ERROR: Failed to save screenshot: no space left on device
```

**Solutions**:

1. **Check disk usage**:
```bash
df -h
du -sh /opt/panoptic/output/*
```

2. **Clean up old artifacts**:
```bash
# Remove old test results
find /opt/panoptic/output -type f -mtime +7 -delete
```

3. **Configure cleanup**:
```yaml
actions:
  - name: "cleanup_old_data"
    type: "cleanup_data"
    parameters:
      retention_days: 7
```

4. **Use cloud storage**:
```yaml
settings:
  cloud:
    provider: "aws"
    enable_sync: true
    delete_local_after_sync: true  # Remove local files after upload
```

---

## Common Error Messages

### "net::ERR_NAME_NOT_RESOLVED"

**Cause**: DNS resolution failed

**Solutions**:
- Check internet connectivity
- Verify URL is correct
- Check DNS settings
- Try with IP address instead of hostname

### "net::ERR_CONNECTION_REFUSED"

**Cause**: Target server refused connection

**Solutions**:
- Verify server is running
- Check firewall rules
- Verify port is correct
- Check if service is listening: `netstat -tlnp | grep PORT`

### "context deadline exceeded"

**Cause**: Operation timed out

**Solutions**:
- Increase timeout values
- Check network latency
- Verify target is responding
- Check for network issues

### "element not found"

**Cause**: CSS selector didn't match any elements

**Solutions**:
- Verify selector syntax
- Check if page loaded completely
- Add wait time before action
- Take screenshot to debug

### "failed to allocate memory"

**Cause**: Out of memory

**Solutions**:
- Increase available RAM
- Reduce concurrent operations
- Check for memory leaks
- Enable swap space

---

## Diagnostic Commands

### System Information

```bash
# Full system diagnostics
./scripts/diagnostics.sh

# Or manually:

# 1. Check Panoptic version
./panoptic --version

# 2. Check Go version
go version

# 3. Check system resources
free -h
df -h
uptime

# 4. Check browser
chromium-browser --version
which chromium-browser

# 5. Check mobile tools
adb version
xcrun simctl list devices

# 6. Test network
ping -c 3 google.com
curl -I https://httpbin.org/get

# 7. Check logs
tail -50 /opt/panoptic/logs/panoptic.log
```

### Configuration Validation

```bash
# Validate YAML syntax
yamllint test_config.yaml

# Validate Panoptic config
./panoptic validate test_config.yaml

# Dry run
./panoptic run --dry-run test_config.yaml

# Verbose mode
./panoptic run -v test_config.yaml 2>&1 | tee debug.log
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/executor/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/executor/
go tool pprof mem.prof

# Full profiling
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./internal/...
```

### Network Diagnostics

```bash
# Test HTTP connectivity
curl -v https://httpbin.org/get

# Test with proxy
curl -x http://proxy:8080 https://httpbin.org/get

# DNS resolution
nslookup example.com
dig example.com

# Trace route
traceroute example.com

# Check ports
netstat -tlnp
ss -tlnp
```

---

## Getting Help

### Before Requesting Support

1. ✅ Check this troubleshooting guide
2. ✅ Review the logs
3. ✅ Try verbose mode (`-v` flag)
4. ✅ Check GitHub issues for similar problems
5. ✅ Gather diagnostic information

### Information to Include in Support Requests

```bash
# Create support bundle
./scripts/support-bundle.sh

# Or gather manually:

# 1. Version information
./panoptic --version > support-info.txt
go version >> support-info.txt
uname -a >> support-info.txt

# 2. Configuration (sanitize sensitive data!)
cat test_config.yaml >> support-info.txt

# 3. Logs
tail -100 /opt/panoptic/logs/panoptic.log >> support-info.txt
tail -100 /opt/panoptic/logs/panoptic-error.log >> support-info.txt

# 4. Error message
echo "ERROR MESSAGE:" >> support-info.txt
# Paste error here

# 5. Steps to reproduce
echo "STEPS TO REPRODUCE:" >> support-info.txt
# List steps here
```

### Support Channels

- **GitHub Issues**: https://github.com/yourusername/panoptic/issues
- **Documentation**: https://github.com/yourusername/panoptic/docs
- **Community Forum**: https://community.yoursite.com/panoptic
- **Enterprise Support**: support@yourcompany.com
- **Security Issues**: security@yourcompany.com

### Emergency Contacts

**P1 (Critical)**: Production down, data loss, security breach
- Email: emergency@yourcompany.com
- Phone: +1-XXX-XXX-XXXX
- Available: 24/7

**P2 (High)**: Major functionality impaired
- Email: support@yourcompany.com
- Response: Within 4 hours (business hours)

**P3 (Medium)**: Minor functionality issue
- Email: support@yourcompany.com
- Response: Within 1 business day

**P4 (Low)**: General questions, feature requests
- Email: support@yourcompany.com or GitHub Issues
- Response: Within 3 business days

---

## FAQ

### Q: Can I run Panoptic without a browser?

**A**: Yes, for desktop and mobile platforms only. Web platform requires Chrome/Chromium.

### Q: Does Panoptic work offline?

**A**: Partially. Cloud storage and external URLs require internet, but local testing works offline.

### Q: How much disk space do I need?

**A**: Minimum 10GB, recommended 50GB+ for production use (artifacts grow over time).

### Q: Can I run multiple instances simultaneously?

**A**: Yes, but ensure each uses a different output directory and doesn't conflict on ports/resources.

### Q: Is Panoptic compatible with Selenium?

**A**: No, Panoptic uses its own Platform abstraction. Migration tools may be available.

### Q: Can I contribute to Panoptic?

**A**: Yes! See CONTRIBUTING.md for guidelines.

---

**Document Version**: 1.0
**Last Updated**: 2025-11-11
**Next Review**: 2026-02-11

**Found a solution not listed here?** Please contribute by opening a PR to update this guide!
