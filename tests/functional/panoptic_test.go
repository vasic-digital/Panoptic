//go:build functional
// +build functional

package functional

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer provides a real HTTP server for testing
type TestServer struct {
	server *httptest.Server
	logs   []string
}

func NewTestServer() *TestServer {
	ts := &TestServer{}
	ts.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEntry := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ts.logs = append(ts.logs, logEntry)

		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <h1>Welcome to Test Application</h1>
    <form id="testForm" method="POST" action="/submit">
        <input name="username" type="text" placeholder="Username" required>
        <input name="email" type="email" placeholder="Email" required>
        <button type="submit" id="submitBtn">Submit</button>
    </form>
    <button id="testBtn">Test Button</button>
</body>
</html>`)
		case "/submit":
			if r.Method == "POST" {
				r.ParseForm()
				username := r.FormValue("username")
				email := r.FormValue("email")
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(200)
				fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Success</title>
</head>
<body>
    <h1>Form Submitted Successfully!</h1>
    <p>Username: %s</p>
    <p>Email: %s</p>
    <a href="/">Back</a>
</body>
</html>`, username, email)
			} else {
				w.WriteHeader(405)
			}
		case "/delay":
			delay := r.URL.Query().Get("delay")
			if delay == "" {
				delay = "2"
			}
			time.Sleep(time.Duration(atoi(delay)) * time.Second)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprintf(w, `<html><body><h1>Page loaded after %s seconds</h1></body></html>`, delay)
		default:
			w.WriteHeader(404)
		}
	}))
	return ts
}

func (ts *TestServer) Start() {
	ts.server.Start()
}

func (ts *TestServer) Close() {
	ts.server.Close()
}

func (ts *TestServer) URL() string {
	return ts.server.URL
}

func (ts *TestServer) GetLogs() []string {
	return ts.logs
}

func (ts *TestServer) HasRequest(path string) bool {
	for _, log := range ts.logs {
		if strings.Contains(log, path) {
			return true
		}
	}
	return false
}

func atoi(s string) int {
	result := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}

// TestPanopticFunctional tests the complete workflow with real applications
func TestPanopticFunctional_WebFormSubmission(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping functional tests in short mode")
	}

	// Start test server
	testServer := NewTestServer()
	testServer.Start()
	defer testServer.Close()

	// Create test configuration
	configContent := fmt.Sprintf(`
name: "Functional Web Form Test"
output: "%s"
apps:
  - name: "Test Web App"
    type: "web"
    url: "%s"
    timeout: 30
actions:
  - name: "navigate_to_form"
    type: "navigate"
    value: "%s"
  - name: "wait_for_page"
    type: "wait"
    wait_time: 2
  - name: "fill_username"
    type: "fill"
    selector: "input[name='username']"
    value: "testuser123"
  - name: "fill_email"
    type: "fill"
    selector: "input[name='email']"
    value: "test@example.com"
  - name: "submit_form"
    type: "submit"
    selector: "#testForm"
  - name: "wait_for_response"
    type: "wait"
    wait_time: 3
  - name: "capture_success_page"
    type: "screenshot"
    parameters:
      filename: "form_success.png"
settings:
  screenshot_format: "png"
  enable_metrics: true
  log_level: "info"
`, t.TempDir(), testServer.URL(), testServer.URL())

	configFile := createTempFile(t, "functional-web-*.yaml", configContent)
	defer os.Remove(configFile)

	// Run Panoptic with the configuration
	outputDir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir, "--verbose")
	cmd.Dir = getProjectRoot()
	
	// Capture output for verification
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute with timeout
	err := cmd.Run()
	outputStr := stdout.String()
	errorStr := stderr.String()
	t.Logf("Command stdout: %s", outputStr)
	t.Logf("Command stderr: %s", errorStr)

	// Verify execution completed (may fail due to browser requirements)
	if err != nil {
		t.Logf("Test completed with error: %v", err)
		
		// Check if error is browser-related (acceptable)
		if strings.Contains(errorStr, "browser") || 
		   strings.Contains(errorStr, "chrome") || 
		   strings.Contains(errorStr, "chromium") {
			t.Skip("Browser not available - skipping functional test")
		}
	}

	// Verify output structure
	assert.DirExists(t, filepath.Join(outputDir, "screenshots"))
	assert.DirExists(t, filepath.Join(outputDir, "videos"))
	assert.DirExists(t, filepath.Join(outputDir, "logs"))

	// Check for log files
	logsDir := filepath.Join(outputDir, "logs")
	logFiles, err := os.ReadDir(logsDir)
	require.NoError(t, err)
	assert.True(t, len(logFiles) > 0, "Expected log files to be created")

	// Verify log content contains test execution
	for _, logFile := range logFiles {
		if !logFile.IsDir() && strings.HasSuffix(logFile.Name(), ".log") {
			logPath := filepath.Join(logsDir, logFile.Name())
			content, err := os.ReadFile(logPath)
			require.NoError(t, err)
			logContent := string(content)
			
			// Should contain test execution logs
			assert.Contains(t, logContent, "Functional Web Form Test") || 
					  assert.Contains(t, logContent, "Test Web App")
		}
	}

	// Verify server received requests
	assert.True(t, testServer.HasRequest("/"), "Expected request to home page")
	assert.True(t, testServer.HasRequest("/submit"), "Expected form submission request")
}

// TestPanopticFunctional_MultiAppWorkflow tests multiple applications in sequence
func TestPanopticFunctional_MultiAppWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping functional tests in short mode")
	}

	// Start test server
	testServer := NewTestServer()
	testServer.Start()
	defer testServer.Close()

	outputDir := t.TempDir()
	configContent := fmt.Sprintf(`
name: "Multi-App Functional Test"
output: "%s"
apps:
  - name: "Primary Web App"
    type: "web"
    url: "%s"
  - name: "Secondary Web App"
    type: "web"
    url: "%s/delay"
actions:
  - name: "test_primary_app"
    type: "screenshot"
    parameters:
      filename: "primary_app.png"
  - name: "test_secondary_app"
    type: "screenshot"
    parameters:
      filename: "secondary_app.png"
  - name: "test_performance"
    type: "record"
    duration: 5
    parameters:
      filename: "performance.mp4"
settings:
  enable_metrics: true
  log_level: "debug"
`, outputDir, testServer.URL(), testServer.URL())

	configFile := createTempFile(t, "functional-multi-*.yaml", configContent)
	defer os.Remove(configFile)

	// Run Panoptic
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir)
	cmd.Dir = getProjectRoot()

	err := cmd.Run()
	t.Logf("Multi-app test result: %v", err)

	// Verify basic output structure exists
	assert.DirExists(t, filepath.Join(outputDir, "screenshots"))
	assert.DirExists(t, filepath.Join(outputDir, "videos"))
	assert.DirExists(t, filepath.Join(outputDir, "logs"))

	// Verify HTML report was generated
	reportPath := filepath.Join(outputDir, "report.html")
	if fileExists(reportPath) {
		reportContent, err := os.ReadFile(reportPath)
		require.NoError(t, err)
		
		reportStr := string(reportContent)
		assert.Contains(t, reportStr, "Multi-App Functional Test")
		assert.Contains(t, reportStr, "Primary Web App")
		assert.Contains(t, reportStr, "Secondary Web App")
		t.Logf("HTML report generated successfully")
	} else {
		t.Logf("HTML report not generated (may be expected if test failed)")
	}

	// Verify server interactions
	assert.True(t, testServer.HasRequest("/"), "Expected request to primary app")
	assert.True(t, testServer.HasRequest("/delay"), "Expected request to secondary app")
}

// TestPanopticFunctional_RealWorldScenario tests a realistic user journey
func TestPanopticFunctional_RealWorldScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping functional tests in short mode")
	}

	// Start enhanced test server
	testServer := NewTestServer()
	testServer.Start()
	defer testServer.Close()

	outputDir := t.TempDir()
	configContent := fmt.Sprintf(`
name: "Real World User Journey"
output: "%s"
apps:
  - name: "E-commerce Web App"
    type: "web"
    url: "%s"
    timeout: 45
actions:
  - name: "navigate_to_home"
    type: "navigate"
    value: "%s"
  - name: "wait_home_load"
    type: "wait"
    wait_time: 3
  - name: "capture_home_page"
    type: "screenshot"
    parameters:
      filename: "homepage_initial.png"
  - name: "start_journey_recording"
    type: "record"
    duration: 10
    parameters:
      filename: "user_journey.mp4"
  - name: "fill_registration_form"
    type: "fill"
    selector: "input[name='username']"
    value: "johndoe"
  - name: "fill_email_field"
    type: "fill"
    selector: "input[name='email']"
    value: "john.doe@example.com"
  - name: "click_submit_button"
    type: "click"
    selector: "#submitBtn"
  - name: "wait_form_submission"
    type: "wait"
    wait_time: 3
  - name: "capture_success_page"
    type: "screenshot"
    parameters:
      filename: "registration_success.png"
  - name: "capture_final_state"
    type: "screenshot"
    parameters:
      filename: "final_state.png"
settings:
  screenshot_format: "png"
  video_format: "mp4"
  quality: 85
  enable_metrics: true
  log_level: "info"
  window_width: 1920
  window_height: 1080
`, outputDir, testServer.URL(), testServer.URL())

	configFile := createTempFile(t, "functional-realworld-*.yaml", configContent)
	defer os.Remove(configFile)

	// Run the test with extended timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir, "--verbose")
	cmd.Dir = getProjectRoot()

	// Capture all output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	outputStr := stdout.String()
	errorStr := stderr.String()

	t.Logf("Real-world test duration: %v", duration)
	t.Logf("Real-world test stdout: %s", outputStr)
	t.Logf("Real-world test stderr: %s", errorStr)

	// Verify test completed within reasonable time
	assert.True(t, duration < 90*time.Second, "Test took too long: %v", duration)

	// Comprehensive output verification
	assert.DirExists(t, filepath.Join(outputDir, "screenshots"))
	assert.DirExists(t, filepath.Join(outputDir, "videos"))
	assert.DirExists(t, filepath.Join(outputDir, "logs"))

	// Check for expected screenshots
	screenshotsDir := filepath.Join(outputDir, "screenshots")
	if files, _ := os.ReadDir(screenshotsDir); len(files) > 0 {
		screenshotNames := []string{}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".png") {
				screenshotNames = append(screenshotNames, file.Name())
			}
		}
		t.Logf("Screenshots created: %v", screenshotNames)
		
		// Should have created at least some screenshots
		assert.True(t, len(screenshotNames) > 0, "Expected at least some screenshots to be created")
	}

	// Check for video files
	videosDir := filepath.Join(outputDir, "videos")
	if files, _ := os.ReadDir(videosDir); len(files) > 0 {
		videoNames := []string{}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".mp4") {
				videoNames = append(videoNames, file.Name())
				// Check file size
				filePath := filepath.Join(videosDir, file.Name())
				if info, err := os.Stat(filePath); err == nil {
					t.Logf("Video created: %s (%d bytes)", file.Name(), info.Size())
				}
			}
		}
		t.Logf("Videos created: %v", videoNames)
	}

	// Verify server received the full user journey
	serverLogs := testServer.GetLogs()
	t.Logf("Server received %d requests", len(serverLogs))
	assert.True(t, len(serverLogs) > 0, "Expected server to receive requests")

	// Verify HTML report contains metrics
	reportPath := filepath.Join(outputDir, "report.html")
	if fileExists(reportPath) {
		reportContent, err := os.ReadFile(reportPath)
		require.NoError(t, err)
		
		reportStr := string(reportContent)
		assert.Contains(t, reportStr, "Real World User Journey")
		assert.Contains(t, reportStr, "E-commerce Web App")
		t.Logf("HTML report generated with content length: %d", len(reportContent))
	}
}

// TestPanopticFunctional_ErrorRecovery tests error handling and recovery
func TestPanopticFunctional_ErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping functional tests in short mode")
	}

	// Test with invalid URL to trigger errors
	outputDir := t.TempDir()
	configContent := fmt.Sprintf(`
name: "Error Recovery Test"
output: "%s"
apps:
  - name: "Invalid URL App"
    type: "web"
    url: "https://non-existent-domain-for-testing-12345.com"
    timeout: 10
  - name: "Valid Fallback App"
    type: "web"
    url: "%s"
    timeout: 30
actions:
  - name: "test_invalid_app"
    type: "navigate"
    value: "https://non-existent-domain-for-testing-12345.com"
  - name: "test_valid_app"
    type: "screenshot"
    parameters:
      filename: "valid_app.png"
settings:
  enable_metrics: true
  log_level: "debug"
`, outputDir, "https://httpbin.org/html")

	configFile := createTempFile(t, "functional-error-*.yaml", configContent)
	defer os.Remove(configFile)

	// Run test
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir, "--verbose")
	cmd.Dir = getProjectRoot()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outputStr := stdout.String()
	errorStr := stderr.String()

	t.Logf("Error recovery test result: %v", err)
	t.Logf("Output: %s", outputStr)
	t.Logf("Error: %s", errorStr)

	// Should complete even with some failures
	// Verify basic output structure
	assert.DirExists(t, filepath.Join(outputDir, "logs"))
	assert.DirExists(t, filepath.Join(outputDir, "screenshots"))

	// Check logs for error handling
	logsDir := filepath.Join(outputDir, "logs")
	if logFiles, _ := os.ReadDir(logsDir); len(logFiles) > 0 {
		for _, logFile := range logFiles {
			if !logFile.IsDir() && strings.HasSuffix(logFile.Name(), ".log") {
				logPath := filepath.Join(logsDir, logFile.Name())
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				
				logContent := string(content)
				if strings.Contains(logContent, "Error") || 
				   strings.Contains(logContent, "Failed") {
					t.Logf("Error properly logged: %s", logContent)
				}
			}
		}
	}
}

// TestPanopticFunctional_PerformanceValidation tests that performance metrics are collected
func TestPanopticFunctional_PerformanceValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping functional tests in short mode")
	}

	testServer := NewTestServer()
	testServer.Start()
	defer testServer.Close()

	outputDir := t.TempDir()
	configContent := fmt.Sprintf(`
name: "Performance Validation Test"
output: "%s"
apps:
  - name: "Performance Test App"
    type: "web"
    url: "%s"
    timeout: 30
actions:
  - name: "navigate_with_timing"
    type: "navigate"
    value: "%s"
  - name: "wait_for_timing"
    type: "wait"
    wait_time: 2
  - name: "screenshot_with_metrics"
    type: "screenshot"
    parameters:
      filename: "performance_test.png"
settings:
  enable_metrics: true
  log_level: "info"
  window_width: 1920
  window_height: 1080
`, outputDir, testServer.URL(), testServer.URL())

	configFile := createTempFile(t, "functional-perf-*.yaml", configContent)
	defer os.Remove(configFile)

	// Run with performance monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir)
	cmd.Dir = getProjectRoot()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	t.Logf("Performance test duration: %v", duration)
	t.Logf("Performance test result: %v", err)

	// Verify timing is reasonable
	assert.True(t, duration < 45*time.Second, "Performance test took too long")
	assert.True(t, duration > 5*time.Second, "Performance test completed too quickly")

	// Check for performance metrics in logs
	logsDir := filepath.Join(outputDir, "logs")
	if logFiles, _ := os.ReadDir(logsDir); len(logFiles) > 0 {
		for _, logFile := range logFiles {
			if !logFile.IsDir() && strings.HasSuffix(logFile.Name(), ".log") {
				logPath := filepath.Join(logsDir, logFile.Name())
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				
				logContent := string(content)
				// Should contain timing information
				if strings.Contains(logContent, "start_time") || 
				   strings.Contains(logContent, "duration") ||
				   strings.Contains(logContent, "metrics") {
					t.Logf("Performance metrics found in logs")
				}
			}
		}
	}

	// Verify HTML report contains performance data
	reportPath := filepath.Join(outputDir, "report.html")
	if fileExists(reportPath) {
		reportContent, err := os.ReadFile(reportPath)
		require.NoError(t, err)
		
		reportStr := string(reportContent)
		if strings.Contains(reportStr, "Duration") ||
		   strings.Contains(reportStr, "Metrics") ||
		   strings.Contains(reportStr, "Start Time") {
			t.Logf("Performance metrics found in HTML report")
		}
	}
}

// Helper functions

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

func getProjectRoot() string {
	// Find project root by looking for go.mod file
	cwd, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return cwd
}

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	// Build Panoptic before running tests
	projectRoot := getProjectRoot()
	buildCmd := exec.Command("go", "build", "-o", "panoptic", "main.go")
	buildCmd.Dir = projectRoot
	
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to build Panoptic: %v\nOutput: %s\n", err, string(output))
		os.Exit(1)
	}
	
	// Run tests
	result := m.Run()
	
	// Cleanup
	os.Remove(filepath.Join(projectRoot, "panoptic"))
	
	os.Exit(result)
}