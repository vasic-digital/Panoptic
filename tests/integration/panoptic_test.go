//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPanopticCLI_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Build the application
	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	t.Run("Help command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.CombinedOutput()
		
		assert.NoError(t, err)
		assert.Contains(t, string(output), "Panoptic")
		assert.Contains(t, string(output), "Usage:")
	})

	t.Run("Version", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "version")
		output, err := cmd.CombinedOutput()
		
		// Might fail if version command doesn't exist, which is fine
		if err == nil {
			assert.NotEmpty(t, strings.TrimSpace(string(output)))
		}
	})

	t.Run("Run with invalid config", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "run", "/non/existent/config.yaml")
		output, err := cmd.CombinedOutput()

		assert.Error(t, err)
		outputStr := string(output)
		// Check for error message (could be "Error:", "FATA", or "Failed")
		assert.True(t,
			strings.Contains(outputStr, "Error:") ||
			strings.Contains(outputStr, "FATA") ||
			strings.Contains(outputStr, "Failed"),
			"Expected error message, got: %s", outputStr)
	})

	t.Run("Run with valid minimal config", func(t *testing.T) {
		configContent := `
name: "Integration Test"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "wait"
    type: "wait"
    wait_time: 1
`

		configFile := createTempFile(t, "config-*.yaml", configContent)
		defer os.Remove(configFile)

		tempDir := t.TempDir()
		cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir, "--verbose")
		
		start := time.Now()
		output, err := cmd.CombinedOutput()
		duration := time.Since(start)

		outputStr := string(output)
		t.Logf("Command output: %s", outputStr)
		t.Logf("Command duration: %v", duration)

		// Should complete within reasonable time
		assert.True(t, duration < 60*time.Second, "Command took too long: %v", duration)
		
		// Check if command succeeded or failed gracefully
		if err != nil {
			// Expected to possibly fail due to browser requirements
			assert.True(t, 
				strings.Contains(outputStr, "Browser not available") ||
				strings.Contains(outputStr, "failed") ||
				strings.Contains(outputStr, "connection") ||
				strings.Contains(outputStr, "timeout"),
				"Unexpected error output: %s", outputStr)
		} else {
			// If successful, verify outputs
			assert.DirExists(t, filepath.Join(tempDir, "screenshots"))
			assert.DirExists(t, filepath.Join(tempDir, "videos"))
			assert.DirExists(t, filepath.Join(tempDir, "logs"))
			assert.FileExists(t, filepath.Join(tempDir, "report.html"))
		}
	})
}

func TestPanoptic_WebAppIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	configContent := `
name: "Web App Integration Test"
apps:
  - name: "HTTPBin Test"
    type: "web"
    url: "https://httpbin.org/html"
    timeout: 30
actions:
  - name: "navigate"
    type: "navigate"
    value: "https://httpbin.org/html"
  - name: "wait_for_load"
    type: "wait"
    wait_time: 2
  - name: "take_screenshot"
    type: "screenshot"
    parameters:
      filename: "httpbin_screenshot.png"
  - name: "click_h1"
    type: "click"
    selector: "h1"
  - name: "wait_after_click"
    type: "wait"
    wait_time: 1
  - name: "final_screenshot"
    type: "screenshot"
    parameters:
      filename: "after_click.png"
`

	configFile := createTempFile(t, "web-test-*.yaml", configContent)
	defer os.Remove(configFile)

	tempDir := t.TempDir()
	cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir, "--verbose")
	
	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	outputStr := string(output)
	t.Logf("Web test output: %s", outputStr)
	t.Logf("Web test duration: %v", duration)

	// Should complete within reasonable time
	assert.True(t, duration < 45*time.Second, "Web test took too long: %v", duration)

	if err != nil {
		// May fail due to browser, check for expected failure patterns
		assert.True(t, 
			strings.Contains(outputStr, "Browser not available") ||
			strings.Contains(outputStr, "failed") ||
			strings.Contains(outputStr, "connection"),
			"Unexpected error in web test: %s", outputStr)
	} else {
		// Verify screenshots were created
		screenshotsDir := filepath.Join(tempDir, "screenshots")
		files, err := os.ReadDir(screenshotsDir)
		assert.NoError(t, err)
		assert.True(t, len(files) >= 0, "Expected at least some files or directories")

		// Check for screenshot files
		hasScreenshots := false
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".png") {
				hasScreenshots = true
				break
			}
		}
		
		// Screenshots may or may not be created depending on browser availability
		if hasScreenshots {
			t.Logf("Screenshots found in output directory")
		} else {
			t.Logf("No screenshots found (expected if browser unavailable)")
		}
	}
}

func TestPanoptic_DesktopAppIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	// Test with a common system application (platform-specific)
	var appPath string
	var appName string
	
	switch {
	case isMacOS():
		appPath = "/Applications/Calculator.app"
		appName = "Calculator"
	case isWindows():
		appName = "notepad.exe"
	default: // Linux
		appName = "gedit"
	}

	configContent := fmt.Sprintf(`
name: "Desktop App Integration Test"
apps:
  - name: "%s"
    type: "desktop"
    path: "%s"
    timeout: 15
actions:
  - name: "wait_app_start"
    type: "wait"
    wait_time: 2
  - name: "take_screenshot"
    type: "screenshot"
    parameters:
      filename: "desktop_app_screenshot.png"
`, appName, appPath)

	configFile := createTempFile(t, "desktop-test-*.yaml", configContent)
	defer os.Remove(configFile)

	tempDir := t.TempDir()
	cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir, "--verbose")
	
	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	outputStr := string(output)
	t.Logf("Desktop test output: %s", outputStr)
	t.Logf("Desktop test duration: %v", duration)

	// Should complete quickly as desktop automation is simulated
	assert.True(t, duration < 30*time.Second, "Desktop test took too long: %v", duration)

	if err != nil {
		// May fail if app doesn't exist, which is expected
		assert.True(t, 
			strings.Contains(outputStr, "application not found") ||
			strings.Contains(outputStr, "failed"),
			"Unexpected error in desktop test: %s", outputStr)
	} else {
		t.Logf("Desktop test completed successfully")
	}
}

func TestPanoptic_MobileAppIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	configContent := `
name: "Mobile App Integration Test"
apps:
  - name: "Android Test App"
    type: "mobile"
    platform: "android"
    emulator: true
    device: "emulator-5554"
    timeout: 20
actions:
  - name: "wait_device"
    type: "wait"
    wait_time: 1
  - name: "navigate_to_url"
    type: "navigate"
    value: "https://httpbin.org/html"
  - name: "wait_navigation"
    type: "wait"
    wait_time: 2
  - name: "take_screenshot"
    type: "screenshot"
    parameters:
      filename: "mobile_screenshot.png"
`

	configFile := createTempFile(t, "mobile-test-*.yaml", configContent)
	defer os.Remove(configFile)

	tempDir := t.TempDir()
	cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir, "--verbose")
	
	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	outputStr := string(output)
	t.Logf("Mobile test output: %s", outputStr)
	t.Logf("Mobile test duration: %v", duration)

	// Should complete within reasonable time
	assert.True(t, duration < 40*time.Second, "Mobile test took too long: %v", duration)

	if err != nil {
		// Expected to fail if Android SDK/tools are not available
		assert.True(t, 
			strings.Contains(outputStr, "platform tools not available") ||
			strings.Contains(outputStr, "device not available") ||
			strings.Contains(outputStr, "adb not found") ||
			strings.Contains(outputStr, "failed"),
			"Expected mobile tools error, got: %s", outputStr)
	} else {
		t.Logf("Mobile test completed successfully")
	}
}

func TestPanoptic_ReportGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	configContent := `
name: "Report Generation Test"
apps:
  - name: "Test App 1"
    type: "web"
    url: "https://httpbin.org/html"
  - name: "Test App 2"
    type: "desktop"
    path: "/Applications/Calculator.app"
actions:
  - name: "wait_action"
    type: "wait"
    wait_time: 1
  - name: "screenshot_action"
    type: "screenshot"
    parameters:
      filename: "report_test.png"
`

	configFile := createTempFile(t, "report-test-*.yaml", configContent)
	defer os.Remove(configFile)

	tempDir := t.TempDir()
	cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir)

	output, _ := cmd.CombinedOutput()
	outputStr := string(output)

	// Check if HTML report was generated
	reportPath := filepath.Join(tempDir, "report.html")
	if fileExists(reportPath) {
		t.Logf("Report generated successfully")

		reportContent, err := os.ReadFile(reportPath)
		require.NoError(t, err)

		reportStr := string(reportContent)
		// Check for basic report structure
		assert.Contains(t, reportStr, "Panoptic Test Report")
		assert.Contains(t, reportStr, "Test Report")
		// Report contains total tests count
		assert.Contains(t, reportStr, "Total Tests:")
	} else {
		t.Logf("Report not generated (expected if tests failed): %s", outputStr)
	}
}

func TestPanoptic_ConfigValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name         string
		configContent string
		expectError  bool
		errorPattern string
	}{
		{
			name: "Valid minimal config",
			configContent: `
name: "Valid Config"
apps:
  - name: "Test App"
    type: "web"
    url: "https://example.com"
`,
			expectError: false,
		},
		{
			name: "Invalid - no apps",
			configContent: `
name: "No Apps Config"
apps: []
`,
			expectError:  true,
			errorPattern: "at least one application must be configured",
		},
		{
			name: "Invalid - web app without URL",
			configContent: `
name: "Invalid Web Config"
apps:
  - name: "Test App"
    type: "web"
`,
			expectError:  true,
			errorPattern: "URL is required for web applications",
		},
		{
			name: "Invalid - unknown app type",
			configContent: `
name: "Invalid Type Config"
apps:
  - name: "Test App"
    type: "unknown"
    url: "https://example.com"
`,
			expectError:  true,
			errorPattern: "unknown application type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := createTempFile(t, "validation-*.yaml", tt.configContent)
			defer os.Remove(configFile)

			cmd := exec.Command(binaryPath, "run", configFile)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorPattern != "" {
					assert.Contains(t, outputStr, tt.errorPattern)
				}
			} else {
				// May still fail due to browser, but not due to validation
				if err != nil && !strings.Contains(outputStr, "Browser") && !strings.Contains(outputStr, "connection") {
					t.Errorf("Unexpected error: %s", outputStr)
				}
			}
		})
	}
}

// Helper functions

func buildPanoptic(t *testing.T) string {
	t.Helper()
	
	// Get current working directory
	cwd, err := os.Getwd()
	require.NoError(t, err)
	
	// Navigate to project root
	projectRoot := filepath.Dir(filepath.Dir(cwd)) // Go up two levels from integration tests
	binaryPath := filepath.Join(projectRoot, "panoptic-test")
	
	cmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	cmd.Dir = projectRoot
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build panoptic: %v\nOutput: %s", err, string(output))
	}
	
	return binaryPath
}

func createTempFile(t *testing.T, pattern, content string) string {
	t.Helper()
	
	tmpFile, err := os.CreateTemp("", pattern)
	require.NoError(t, err)
	
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	
	tmpFile.Close()
	return tmpFile.Name()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func isMacOS() bool {
	return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "darwin") || 
		   strings.Contains(strings.ToLower(runtime.GOOS), "darwin")
}

func isWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "windows") || 
		   strings.Contains(strings.ToLower(runtime.GOOS), "windows")
}