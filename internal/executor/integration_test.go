package executor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ConfigToExecutor validates the full pipeline from config loading
// through executor creation and report generation.
func TestIntegration_ConfigToExecutor(t *testing.T) {
	configContent := `
name: "Integration Test"
output: "./test_output"
apps:
  - name: "Admin Console"
    type: "web"
    url: "http://localhost:3001"
    timeout: 30
    actions:
      - name: "nav_login"
        type: "navigate"
        url: "http://localhost:3001/login"
      - name: "fill_user"
        type: "fill"
        selector: "input[name='username']"
        value: "admin"
      - name: "click_submit"
        type: "click"
        selector: "button[type='submit']"
      - name: "wait_load"
        type: "wait"
        wait_time: 2
      - name: "capture"
        type: "screenshot"
        parameters:
          filename: "dashboard.png"
  - name: "Web App"
    type: "web"
    url: "http://localhost:3000"
    timeout: 30
    actions:
      - name: "nav_login"
        type: "navigate"
        url: "http://localhost:3000/login"
settings:
  headless: true
  quality: 90
  screenshot_format: "png"
  video_format: "mp4"
`
	tmpFile, err := os.CreateTemp("", "integration-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	cfg, err := config.Load(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, "Integration Test", cfg.Name)
	assert.Len(t, cfg.Apps, 2)

	// Validate config
	err = cfg.Validate()
	assert.NoError(t, err)

	// Verify per-app actions
	adminActions := cfg.GetActionsForApp(cfg.Apps[0])
	assert.Len(t, adminActions, 5)
	assert.Equal(t, "navigate", adminActions[0].Type)
	assert.Equal(t, "http://localhost:3001/login", adminActions[0].GetNavigateURL())

	webActions := cfg.GetActionsForApp(cfg.Apps[1])
	assert.Len(t, webActions, 1)
	assert.Equal(t, "http://localhost:3000/login", webActions[0].GetNavigateURL())

	// Create executor
	tmpDir := t.TempDir()
	log := logger.NewLogger(false)
	executor := NewExecutor(cfg, tmpDir, log)
	assert.NotNil(t, executor)

	// Simulate test results
	executor.results = []TestResult{
		{
			AppName:     "Admin Console",
			AppType:     "web",
			StartTime:   time.Now().Add(-10 * time.Second),
			EndTime:     time.Now(),
			Duration:    10 * time.Second,
			Success:     true,
			Screenshots: []string{},
			Videos:      []string{},
			Metrics:     map[string]interface{}{"url": "http://localhost:3001"},
		},
		{
			AppName:   "Web App",
			AppType:   "web",
			StartTime: time.Now().Add(-5 * time.Second),
			EndTime:   time.Now(),
			Duration:  5 * time.Second,
			Success:   false,
			Error:     "Connection refused: http://localhost:3000",
			Metrics:   map[string]interface{}{"url": "http://localhost:3000"},
		},
	}

	// Generate report
	reportPath := filepath.Join(tmpDir, "report.html")
	err = executor.GenerateReport(reportPath)
	assert.NoError(t, err)

	// Verify report contents
	data, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	html := string(data)

	assert.Contains(t, html, "Panoptic Test Report")
	assert.Contains(t, html, "Admin Console")
	assert.Contains(t, html, "Web App")
	assert.Contains(t, html, "PASSED")
	assert.Contains(t, html, "FAILED")
	assert.Contains(t, html, "Connection refused")
	assert.Contains(t, html, `class="stat pass"`)
	assert.Contains(t, html, `class="stat fail"`)
}

// TestIntegration_PerAppVsGlobalActions verifies per-app actions override globals.
func TestIntegration_PerAppVsGlobalActions(t *testing.T) {
	configContent := `
name: "Per-App Actions Test"
apps:
  - name: "App With Own Actions"
    type: "web"
    url: "http://localhost:3001"
    actions:
      - name: "app_specific"
        type: "navigate"
        url: "http://localhost:3001/specific"
  - name: "App Without Own Actions"
    type: "web"
    url: "http://localhost:3000"
actions:
  - name: "global_nav"
    type: "navigate"
    url: "http://global.example.com"
`
	tmpFile, err := os.CreateTemp("", "perapp-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := config.Load(tmpFile.Name())
	require.NoError(t, err)

	// App with per-app actions should use its own
	actions1 := cfg.GetActionsForApp(cfg.Apps[0])
	assert.Len(t, actions1, 1)
	assert.Equal(t, "app_specific", actions1[0].Name)
	assert.Equal(t, "http://localhost:3001/specific", actions1[0].GetNavigateURL())

	// App without per-app actions should fall back to global
	actions2 := cfg.GetActionsForApp(cfg.Apps[1])
	assert.Len(t, actions2, 1)
	assert.Equal(t, "global_nav", actions2[0].Name)
	assert.Equal(t, "http://global.example.com", actions2[0].GetNavigateURL())
}

// TestIntegration_AllActionTypesValidation verifies all 27 action types are accepted.
func TestIntegration_AllActionTypesValidation(t *testing.T) {
	actionTypes := []string{
		"navigate", "click", "fill", "submit", "wait", "screenshot", "record",
		"vision_click", "vision_report",
		"ai_test_generation", "smart_error_detection", "ai_enhanced_testing",
		"cloud_sync", "cloud_analytics", "distributed_test", "cloud_cleanup",
		"user_create", "user_authenticate", "project_create", "team_create",
		"api_key_create", "audit_report", "compliance_check", "license_info",
		"enterprise_status", "backup_data", "cleanup_data",
	}

	log := logger.NewLogger(false)
	tmpDir := t.TempDir()

	for _, actionType := range actionTypes {
		t.Run(actionType, func(t *testing.T) {
			cfg := &config.Config{
				Name: "Type Test",
				Apps: []config.AppConfig{
					{Name: "Test App", Type: "web", URL: "http://localhost:3000"},
				},
			}
			executor := NewExecutor(cfg, tmpDir, log)
			assert.NotNil(t, executor, "Executor should be created for action type: %s", actionType)
		})
	}
}

// TestIntegration_NavigateURLPrecedence verifies URL field takes precedence over Value.
func TestIntegration_NavigateURLPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		action   config.Action
		expected string
	}{
		{
			name:     "url takes precedence",
			action:   config.Action{URL: "http://url.com", Value: "http://value.com"},
			expected: "http://url.com",
		},
		{
			name:     "falls back to value",
			action:   config.Action{Value: "http://value.com"},
			expected: "http://value.com",
		},
		{
			name:     "url only",
			action:   config.Action{URL: "http://url.com"},
			expected: "http://url.com",
		},
		{
			name:     "both empty",
			action:   config.Action{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.action.GetNavigateURL())
		})
	}
}

// TestIntegration_ReportWithAllMediaTypes tests report with screenshots, videos, and errors.
func TestIntegration_ReportWithAllMediaTypes(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "full_report.html")

	// Create test media files
	screenshotDir := filepath.Join(tmpDir, "screenshots")
	videoDir := filepath.Join(tmpDir, "videos")
	require.NoError(t, os.MkdirAll(screenshotDir, 0755))
	require.NoError(t, os.MkdirAll(videoDir, 0755))

	screenshot1 := filepath.Join(screenshotDir, "login.png")
	screenshot2 := filepath.Join(screenshotDir, "dashboard.png")
	video1 := filepath.Join(videoDir, "session.mp4")
	require.NoError(t, os.WriteFile(screenshot1, []byte("fake png 1"), 0644))
	require.NoError(t, os.WriteFile(screenshot2, []byte("fake png 2"), 0644))
	require.NoError(t, os.WriteFile(video1, []byte("fake video"), 0644))

	results := []TestResult{
		{
			AppName:     "Admin Console",
			AppType:     "web",
			StartTime:   time.Now().Add(-30 * time.Second),
			EndTime:     time.Now(),
			Duration:    30 * time.Second,
			Success:     true,
			Screenshots: []string{screenshot1, screenshot2},
			Videos:      []string{video1},
			Metrics: map[string]interface{}{
				"url":              "http://localhost:3001",
				"actions_executed": 12,
				"screenshots":     2,
			},
		},
		{
			AppName:   "Web App",
			AppType:   "web",
			StartTime: time.Now().Add(-15 * time.Second),
			EndTime:   time.Now(),
			Duration:  15 * time.Second,
			Success:   false,
			Error:     "Element not found: #login-form",
			Metrics:   map[string]interface{}{"url": "http://localhost:3000"},
		},
		{
			AppName:   "Mobile App",
			AppType:   "mobile",
			StartTime: time.Now().Add(-10 * time.Second),
			EndTime:   time.Now(),
			Duration:  10 * time.Second,
			Success:   true,
			Metrics:   map[string]interface{}{"platform": "android"},
		},
	}

	err := GenerateComprehensiveReport(outputPath, results)
	assert.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	html := string(data)

	// Verify all apps appear
	assert.Contains(t, html, "Admin Console")
	assert.Contains(t, html, "Web App")
	assert.Contains(t, html, "Mobile App")

	// Verify pass/fail statuses
	assert.Contains(t, html, "PASSED")
	assert.Contains(t, html, "FAILED")

	// Verify media references
	assert.Contains(t, html, "login.png")
	assert.Contains(t, html, "dashboard.png")
	assert.Contains(t, html, "session.mp4")
	assert.Contains(t, html, "<video controls")
	assert.Contains(t, html, "download")

	// Verify error message
	assert.Contains(t, html, "Element not found: #login-form")

	// Verify summary stats
	assert.Contains(t, html, `<div class="value">3</div>`) // 3 total apps
}

// TestIntegration_ComprehensiveConfigLoad validates loading the comprehensive YAML config.
func TestIntegration_ComprehensiveConfigLoad(t *testing.T) {
	configContent := `
name: "Comprehensive Validation"
output: "./reports/qa/panoptic/validation"
apps:
  - name: "Admin Console"
    type: "web"
    url: "http://localhost:3001"
    timeout: 120
    actions:
      - name: "navigate_login"
        type: "navigate"
        url: "http://localhost:3001/login"
      - name: "fill_username"
        type: "fill"
        selector: "input[name='username']"
        value: "admin"
      - name: "fill_password"
        type: "fill"
        selector: "input[name='password']"
        value: "admin"
      - name: "click_login"
        type: "click"
        selector: "button[type='submit']"
      - name: "wait_dashboard"
        type: "wait"
        wait_time: 3
      - name: "screenshot_dash"
        type: "screenshot"
        parameters:
          filename: "dashboard.png"
      - name: "record_session"
        type: "record"
        duration: 60
        parameters:
          filename: "session.mp4"
      - name: "vision_check"
        type: "vision_report"
        parameters:
          output: "vision.json"
      - name: "ai_tests"
        type: "ai_test_generation"
        parameters:
          output: "ai_tests.json"
      - name: "error_detect"
        type: "smart_error_detection"
        parameters:
          output: "errors.json"
      - name: "enhanced_ai"
        type: "ai_enhanced_testing"
        parameters:
          output: "enhanced.json"
  - name: "Web App"
    type: "web"
    url: "http://localhost:3000"
    timeout: 120
    actions:
      - name: "navigate_login"
        type: "navigate"
        url: "http://localhost:3000/login"
actions:
  - name: "enterprise_user"
    type: "user_create"
    parameters:
      username: "test_user"
  - name: "enterprise_auth"
    type: "user_authenticate"
    parameters:
      username: "test_user"
  - name: "enterprise_team"
    type: "team_create"
    parameters:
      name: "QA Team"
  - name: "enterprise_project"
    type: "project_create"
    parameters:
      name: "QA Project"
  - name: "enterprise_api_key"
    type: "api_key_create"
    parameters:
      name: "QA Key"
  - name: "audit"
    type: "audit_report"
    parameters:
      output: "audit.json"
  - name: "compliance"
    type: "compliance_check"
    parameters:
      output: "compliance.json"
  - name: "license"
    type: "license_info"
    parameters:
      output: "license.json"
  - name: "status"
    type: "enterprise_status"
    parameters:
      output: "status.json"
  - name: "backup"
    type: "backup_data"
    parameters:
      output: "backup.json"
  - name: "cleanup"
    type: "cleanup_data"
    parameters:
      output: "cleanup.json"
  - name: "cloud_sync"
    type: "cloud_sync"
    parameters:
      output: "sync.json"
  - name: "cloud_analytics"
    type: "cloud_analytics"
    parameters:
      output: "analytics.json"
  - name: "cloud_distributed"
    type: "distributed_test"
    parameters:
      output: "distributed.json"
  - name: "cloud_cleanup"
    type: "cloud_cleanup"
    parameters:
      output: "cloud_cleanup.json"
settings:
  headless: true
  window_width: 1920
  window_height: 1080
  screenshot_format: "png"
  video_format: "mp4"
  quality: 90
  enable_metrics: true
  log_level: "info"
`

	tmpFile, err := os.CreateTemp("", "comprehensive-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := config.Load(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "Comprehensive Validation", cfg.Name)
	assert.Len(t, cfg.Apps, 2)

	// Admin app has 11 per-app actions
	adminActions := cfg.GetActionsForApp(cfg.Apps[0])
	assert.Len(t, adminActions, 11)

	// Web app has 1 per-app action
	webActions := cfg.GetActionsForApp(cfg.Apps[1])
	assert.Len(t, webActions, 1)

	// Global actions: 15 enterprise + cloud actions
	assert.Len(t, cfg.Actions, 15)

	// Validate
	err = cfg.Validate()
	assert.NoError(t, err)

	// Verify settings
	assert.True(t, cfg.Settings.Headless)
	assert.Equal(t, 1920, cfg.Settings.WindowWidth)
	assert.Equal(t, 1080, cfg.Settings.WindowHeight)
	assert.Equal(t, "png", cfg.Settings.ScreenshotFormat)
	assert.Equal(t, "mp4", cfg.Settings.VideoFormat)
	assert.Equal(t, 90, cfg.Settings.Quality)
	assert.Equal(t, "info", cfg.Settings.LogLevel)
}

// TestIntegration_ExecutorWithAllActionTypes tests that the executor can
// handle configs with all action types without crashing.
func TestIntegration_ExecutorWithAllActionTypes(t *testing.T) {
	log := logger.NewLogger(false)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Name: "All Actions",
		Apps: []config.AppConfig{
			{
				Name:    "Test App",
				Type:    "web",
				URL:     "http://localhost:3000",
				Timeout: 30,
				Actions: []config.Action{
					{Name: "nav", Type: "navigate", URL: "http://localhost:3000"},
					{Name: "wait", Type: "wait", WaitTime: 1},
					{Name: "screenshot", Type: "screenshot"},
				},
			},
		},
		Actions: []config.Action{
			{Name: "user_create", Type: "user_create"},
			{Name: "user_auth", Type: "user_authenticate"},
			{Name: "team_create", Type: "team_create"},
			{Name: "project_create", Type: "project_create"},
			{Name: "api_key", Type: "api_key_create"},
			{Name: "audit", Type: "audit_report"},
			{Name: "compliance", Type: "compliance_check"},
			{Name: "license", Type: "license_info"},
			{Name: "status", Type: "enterprise_status"},
			{Name: "backup", Type: "backup_data"},
			{Name: "cleanup", Type: "cleanup_data"},
		},
	}

	executor := NewExecutor(cfg, tmpDir, log)
	assert.NotNil(t, executor)

	// Verify per-app actions are used for the app
	appActions := cfg.GetActionsForApp(cfg.Apps[0])
	assert.Len(t, appActions, 3)
	assert.Equal(t, "nav", appActions[0].Name)

	// Verify global actions exist for fallback
	assert.Len(t, cfg.Actions, 11)
}

// TestIntegration_ReportDurationFormatting tests all duration format cases.
func TestIntegration_ReportDurationFormatting(t *testing.T) {
	tmpDir := t.TempDir()

	results := []TestResult{
		{
			AppName:   "Fast Test",
			AppType:   "web",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  100 * time.Millisecond,
			Success:   true,
			Metrics:   map[string]interface{}{},
		},
		{
			AppName:   "Medium Test",
			AppType:   "web",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  5500 * time.Millisecond,
			Success:   true,
			Metrics:   map[string]interface{}{},
		},
		{
			AppName:   "Long Test",
			AppType:   "web",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  185 * time.Second,
			Success:   false,
			Error:     "Timeout after 3 minutes",
			Metrics:   map[string]interface{}{},
		},
	}

	outputPath := filepath.Join(tmpDir, "duration_report.html")
	err := GenerateComprehensiveReport(outputPath, results)
	assert.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	html := string(data)

	// Verify duration formatting in report
	assert.Contains(t, html, "100ms")   // Milliseconds format
	assert.Contains(t, html, "5.5s")    // Seconds format
	assert.Contains(t, html, "3m 5s")   // Minutes format
	assert.Contains(t, html, "Timeout after 3 minutes")
}
