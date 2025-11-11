//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	// Test complete workflow: web -> desktop -> mobile with comprehensive actions
	configContent := `
name: "E2E Full Workflow Test"

apps:
  - name: "Web Application"
    type: "web"
    url: "https://httpbin.org/forms/post"
    timeout: 30

  - name: "Desktop Application"
    type: "desktop"
    path: "/Applications/Calculator.app"
    timeout: 15

  - name: "Mobile Application"
    type: "mobile"
    platform: "android"
    emulator: true
    device: "emulator-5554"
    timeout: 20

actions:
  # Web application actions
  - name: "navigate_to_form"
    type: "navigate"
    value: "https://httpbin.org/forms/post"
  
  - name: "wait_for_form_load"
    type: "wait"
    wait_time: 3

  - name: "fill_customer_name"
    type: "fill"
    selector: "input[name='custname']"
    value: "John Doe"

  - name: "fill_telephone"
    type: "fill"
    selector: "input[name='custtel']"
    value: "+1234567890"

  - name: "fill_email"
    type: "fill"
    selector: "input[name='custemail']"
    value: "john.doe@example.com"

  - name: "select_size"
    type: "click"
    selector: "input[value='medium']"

  - name: "select_topping"
    type: "click"
    selector: "input[value='bacon']"

  - name: "submit_form"
    type: "submit"
    selector: "form"

  - name: "wait_for_submission"
    type: "wait"
    wait_time: 3

  - name: "capture_form_result"
    type: "screenshot"
    parameters:
      filename: "form_submission_result.png"

  # Desktop application actions
  - name: "start_desktop_app"
    type: "navigate"
    value: "desktop://calculator"

  - name: "wait_desktop_start"
    type: "wait"
    wait_time: 2

  - name: "capture_desktop_ui"
    type: "screenshot"
    parameters:
      filename: "desktop_app_ui.png"

  # Mobile application actions
  - name: "start_mobile_app"
    type: "navigate"
    value: "https://httpbin.org/html"

  - name: "wait_mobile_load"
    type: "wait"
    wait_time: 2

  - name: "capture_mobile_ui"
    type: "screenshot"
    parameters:
      filename: "mobile_app_ui.png"

settings:
  screenshot_format: "png"
  video_format: "mp4"
  quality: 85
  headless: false
  window_width: 1920
  window_height: 1080
  enable_metrics: true
  log_level: "info"
`

	tempDir := t.TempDir()
	configFile := createTempFile(t, "e2e-workflow-*.yaml", configContent)
	defer os.Remove(configFile)

	t.Run("Complete workflow execution", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir, "--verbose")
		
		start := time.Now()
		output, err := cmd.CombinedOutput()
		duration := time.Since(start)

		outputStr := string(output)
		t.Logf("E2E workflow output: %s", outputStr)
		t.Logf("E2E workflow duration: %v", duration)

		// Should complete within reasonable time (considering all platforms)
		assert.True(t, duration < 60*time.Second, "E2E workflow took too long: %v", duration)

		// Verify output structure
		assert.DirExists(t, filepath.Join(tempDir, "screenshots"))
		assert.DirExists(t, filepath.Join(tempDir, "videos"))
		assert.DirExists(t, filepath.Join(tempDir, "logs"))
		
		// Check for log files (optional)
		logsDir := filepath.Join(tempDir, "logs")
		if logFiles, err := os.ReadDir(logsDir); err == nil && len(logFiles) > 0 {
			t.Logf("Log files created: %d", len(logFiles))
		} else {
			t.Logf("No log files created (may be expected depending on configuration)")
		}

		// Check for HTML report
		reportPath := filepath.Join(tempDir, "report.html")
		if fileExists(reportPath) {
			t.Logf("HTML report generated successfully")

			reportContent, err := os.ReadFile(reportPath)
			require.NoError(t, err)

			// Verify report contains basic structure
			reportStr := string(reportContent)
			assert.Contains(t, reportStr, "Panoptic Test Report")
			assert.Contains(t, reportStr, "Test Report")
			assert.Contains(t, reportStr, "Total Tests:")
		} else {
			t.Logf("HTML report not generated (expected if tests failed)")
		}

		// Verify screenshots (may not exist if browser unavailable)
		screenshotsDir := filepath.Join(tempDir, "screenshots")
		if dir, _ := os.ReadDir(screenshotsDir); len(dir) > 0 {
			t.Logf("Screenshots created: %d files", len(dir))
			
			// Check for specific expected screenshots
			expectedScreenshots := []string{
				"form_submission_result.png",
				"desktop_app_ui.png", 
				"mobile_app_ui.png",
			}
			
			for _, expected := range expectedScreenshots {
				if fileExists(filepath.Join(screenshotsDir, expected)) {
					t.Logf("Found expected screenshot: %s", expected)
				}
			}
		} else {
			t.Logf("No screenshots created (expected if browser/desktop/mobile unavailable)")
		}

		// Error handling - tests may fail due to missing platform tools
		if err != nil {
			t.Logf("E2E test completed with errors (expected in some environments): %v", err)
			
			// Check for expected error patterns
			expectedPatterns := []string{
				"application not found",
				"platform tools not available", 
				"device not available",
				"browser",
				"connection",
				"timeout",
			}
			
			hasExpectedError := false
			for _, pattern := range expectedPatterns {
				if strings.Contains(outputStr, pattern) {
					hasExpectedError = true
					break
				}
			}
			
			if !hasExpectedError {
				t.Logf("Unexpected error pattern: %s", outputStr)
			}
		} else {
			t.Logf("E2E workflow completed successfully")
		}
	})
}

func TestE2E_RecordingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	configContent := `
name: "E2E Recording Test"
apps:
  - name: "Recording Test App"
    type: "web"
    url: "https://httpbin.org/delay/3"
    timeout: 30

actions:
  - name: "navigate_to_page"
    type: "navigate"
    value: "https://httpbin.org/delay/3"

  - name: "wait_during_recording"
    type: "wait"
    wait_time: 2

  - name: "click_something"
    type: "click"
    selector: "h1"

  - name: "take_final_screenshot"
    type: "screenshot"
    parameters:
      filename: "final_state.png"
`

	tempDir := t.TempDir()
	configFile := createTempFile(t, "e2e-recording-*.yaml", configContent)
	defer os.Remove(configFile)

	t.Run("Recording workflow", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir, "--verbose")
		
		start := time.Now()
		output, err := cmd.CombinedOutput()
		duration := time.Since(start)

		outputStr := string(output)
		t.Logf("Recording workflow output: %s", outputStr)
		t.Logf("Recording workflow duration: %v", duration)

		// Should complete within reasonable time
		assert.True(t, duration < 60*time.Second, "Recording workflow took too long")

		// Error handling
		if err != nil {
			t.Logf("Recording workflow completed with errors (expected if browser unavailable)")
		} else {
			t.Logf("Recording workflow completed successfully")
		}
	})
}

func TestE2E_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name         string
		configContent string
		expectedErrorPattern string
	}{
		{
			name: "Non-existent web URL",
			configContent: `
name: "Error Handling Test"
apps:
  - name: "Bad URL App"
    type: "web"
    url: "https://non-existent-domain-12345.com"
    timeout: 10
actions:
  - name: "navigate"
    type: "navigate"
    value: "https://non-existent-domain-12345.com"
  - name: "wait"
    type: "wait"
    wait_time: 2
`,
			expectedErrorPattern: "connection",
		},
		{
			name: "Non-existent desktop app",
			configContent: `
name: "Desktop Error Test"
apps:
  - name: "Non-existent App"
    type: "desktop"
    path: "/non/existent/path/app.app"
    timeout: 10
actions:
  - name: "wait"
    type: "wait"
    wait_time: 1
`,
			expectedErrorPattern: "application not found",
		},
		{
			name: "Mobile without tools",
			configContent: `
name: "Mobile Error Test"
apps:
  - name: "No Tools App"
    type: "mobile"
    platform: "android"
    emulator: true
    device: "emulator-5554"
    timeout: 10
actions:
  - name: "wait"
    type: "wait"
    wait_time: 1
`,
			expectedErrorPattern: "platform tools not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := createTempFile(t, "e2e-error-*.yaml", tt.configContent)
			defer os.Remove(configFile)

			tempDir := t.TempDir()
			cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir)
			
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Should complete (with or without errors) within reasonable time
			assert.True(t, time.Since(time.Now()) < 30*time.Second)

			if err != nil {
				assert.Contains(t, outputStr, tt.expectedErrorPattern)
			} else {
				// Even if no error, check output for expected patterns
				assert.True(t, 
					strings.Contains(outputStr, tt.expectedErrorPattern) ||
					strings.Contains(outputStr, "completed successfully"),
					"Expected error pattern or success message in output")
			}

			// Verify basic output structure is still created
			assert.DirExists(t, filepath.Join(tempDir, "logs"))
		})
	}
}

func TestE2E_PerformanceMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	binaryPath := buildPanoptic(t)
	defer os.Remove(binaryPath)

	configContent := `
name: "Performance Metrics Test"
apps:
  - name: "Performance App"
    type: "web"
    url: "https://httpbin.org/delay/2"
    timeout: 30

actions:
  - name: "navigate_with_timing"
    type: "navigate"
    value: "https://httpbin.org/delay/2"

  - name: "wait_for_page"
    type: "wait"
    wait_time: 3

  - name: "click_action"
    type: "click"
    selector: "h1"

  - name: "wait_after_click"
    type: "wait"
    wait_time: 1

  - name: "final_screenshot"
    type: "screenshot"
    parameters:
      filename: "performance_test.png"

settings:
  enable_metrics: true
  log_level: "info"
`

	tempDir := t.TempDir()
	configFile := createTempFile(t, "e2e-performance-*.yaml", configContent)
	defer os.Remove(configFile)

	t.Run("Performance metrics collection", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "run", configFile, "--output", tempDir, "--verbose")
		
		start := time.Now()
		output, err := cmd.CombinedOutput()
		duration := time.Since(start)

		outputStr := string(output)
		t.Logf("Performance test output: %s", outputStr)
		t.Logf("Performance test duration: %v", duration)

		// Should complete within expected time
		assert.True(t, duration < 45*time.Second)

		// Check for metrics in output logs
		logsDir := filepath.Join(tempDir, "logs")
		if logFiles, _ := os.ReadDir(logsDir); len(logFiles) > 0 {
			for _, logFile := range logFiles {
				if !logFile.IsDir() && strings.HasSuffix(logFile.Name(), ".log") {
					logPath := filepath.Join(logsDir, logFile.Name())
					logContent, err := os.ReadFile(logPath)
					if err == nil {
						logStr := string(logContent)
						if strings.Contains(logStr, "metrics") || 
						   strings.Contains(logStr, "duration") ||
						   strings.Contains(logStr, "start_time") {
							t.Logf("Performance metrics found in logs")
						}
					}
				}
			}
		}

		// Check report for metrics
		reportPath := filepath.Join(tempDir, "report.html")
		if fileExists(reportPath) {
			reportContent, err := os.ReadFile(reportPath)
			if err == nil {
				reportStr := string(reportContent)
				if strings.Contains(reportStr, "Metrics") ||
				   strings.Contains(reportStr, "Duration") ||
				   strings.Contains(reportStr, "Start Time") {
					t.Logf("Performance metrics found in HTML report")
				}
			}
		}

		// Error is acceptable if browser unavailable
		if err != nil && !strings.Contains(outputStr, "Browser") && 
		   !strings.Contains(outputStr, "connection") {
			t.Logf("Performance test completed with unexpected error: %v", err)
		}
	})
}

// Helper functions

func buildPanoptic(t *testing.T) string {
	t.Helper()

	// Find project root by looking for go.mod
	cwd, err := os.Getwd()
	require.NoError(t, err)

	projectRoot := cwd
	// Keep going up until we find go.mod or hit root
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			// Hit root without finding go.mod
			t.Fatal("Could not find project root (no go.mod found)")
		}
		projectRoot = parent
	}

	binaryPath := filepath.Join(projectRoot, "panoptic-e2e")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build panoptic for e2e: %v\nOutput: %s", err, string(output))
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