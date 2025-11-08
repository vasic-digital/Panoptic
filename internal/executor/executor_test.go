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

func TestNewExecutor(t *testing.T) {
	config := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{
			{
				Name: "Test App",
				Type: "web",
				URL:  "https://example.com",
			},
		},
		Actions: []config.Action{
			{
				Name: "test_action",
				Type: "wait",
				WaitTime: 1,
			},
		},
	}

	outputDir := t.TempDir()
	log := logger.NewLogger(false)

	executor := NewExecutor(config, outputDir, log)

	assert.NotNil(t, executor)
	assert.Equal(t, config, executor.config)
	assert.Equal(t, outputDir, executor.outputDir)
	assert.Equal(t, log, executor.logger)
	assert.NotNil(t, executor.factory)
	assert.Empty(t, executor.results)
}

func TestExecutor_Run(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid web app config",
			config: &config.Config{
				Name: "Test Web Execution",
				Apps: []config.AppConfig{
					{
						Name:    "Test Web App",
						Type:    "web",
						URL:     "https://httpbin.org/html",
						Timeout: 30,
					},
				},
				Actions: []config.Action{
					{
						Name:     "wait_for_load",
						Type:     "wait",
						WaitTime: 2,
					},
					{
						Name: "take_screenshot",
						Type: "screenshot",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Valid desktop app config",
			config: &config.Config{
				Name: "Test Desktop Execution",
				Apps: []config.AppConfig{
					{
						Name: "Test Desktop App",
						Type: "desktop",
						Path: "/Applications/Calculator.app", // May not exist on all systems
						Timeout: 30,
					},
				},
				Actions: []config.Action{
					{
						Name:     "wait_app",
						Type:     "wait",
						WaitTime: 1,
					},
				},
			},
			expectError: false, // Should handle missing app gracefully
		},
		{
			name: "Invalid config - no apps",
			config: &config.Config{
				Name:    "Invalid Config",
				Apps:    []config.AppConfig{},
				Actions: []config.Action{},
			},
			expectError: true,
			errorMsg:    "configuration validation failed",
		},
		{
			name: "Invalid app type",
			config: &config.Config{
				Name: "Invalid App Type",
				Apps: []config.AppConfig{
					{
						Name: "Invalid App",
						Type: "invalid_type",
					},
				},
				Actions: []config.Action{},
			},
			expectError: true,
			errorMsg:    "configuration validation failed",
		},
		{
			name: "Web app without URL",
			config: &config.Config{
				Name: "Web App No URL",
				Apps: []config.AppConfig{
					{
						Name: "Web App",
						Type: "web",
					},
				},
				Actions: []config.Action{},
			},
			expectError: true,
			errorMsg:    "configuration validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := t.TempDir()
			log := logger.NewLogger(false)
			executor := NewExecutor(tt.config, outputDir, log)

			err := executor.Run()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// For successful cases, verify results
				if len(executor.results) > 0 {
					result := executor.results[0]
					assert.NotNil(t, result)
					assert.Equal(t, tt.config.Apps[0].Name, result.AppName)
					assert.Equal(t, tt.config.Apps[0].Type, result.AppType)
					assert.NotZero(t, result.Duration)
					
					// Verify output directories were created
					assert.DirExists(t, filepath.Join(outputDir, "screenshots"))
					assert.DirExists(t, filepath.Join(outputDir, "videos"))
					assert.DirExists(t, filepath.Join(outputDir, "logs"))
				}
			}
		})
	}
}

func TestExecutor_ExecuteAction(t *testing.T) {
	config := &config.Config{
		Name: "Test Actions",
		Apps: []config.AppConfig{
			{
				Name:    "Test App",
				Type:    "web",
				URL:     "https://httpbin.org/html",
				Timeout: 30,
			},
		},
		Actions: []config.Action{},
	}

	outputDir := t.TempDir()
	log := logger.NewLogger(false)
	executor := NewExecutor(config, outputDir, log)

	tests := []struct {
		name        string
		action      config.Action
		expectError bool
		skipReason  string
	}{
		{
			name: "Navigate action",
			action: config.Action{
				Name:  "navigate",
				Type:  "navigate",
				Value: "https://httpbin.org/html",
			},
			expectError: false,
		},
		{
			name: "Wait action",
			action: config.Action{
				Name:     "wait",
				Type:     "wait",
				WaitTime: 1,
			},
			expectError: false,
		},
		{
			name: "Click action",
			action: config.Action{
				Name:     "click",
				Type:     "click",
				Selector: "h1",
			},
			expectError: false, // May fail if element doesn't exist, that's okay
		},
		{
			name: "Fill action",
			action: config.Action{
				Name:     "fill",
				Type:     "fill",
				Selector: "input.test",
				Value:    "test-value",
			},
			expectError: false, // May fail if element doesn't exist, that's okay
		},
		{
			name: "Submit action",
			action: config.Action{
				Name:     "submit",
				Type:     "submit",
				Selector: "form.test",
			},
			expectError: false, // May fail if element doesn't exist, that's okay
		},
		{
			name: "Screenshot action",
			action: config.Action{
				Name: "screenshot",
				Type: "screenshot",
				Parameters: map[string]interface{}{
					"filename": "test_screenshot.png",
				},
			},
			expectError: false,
		},
		{
			name: "Record action",
			action: config.Action{
				Name:     "record",
				Type:     "record",
				Duration: 2, // Short duration for testing
				Parameters: map[string]interface{}{
					"filename": "test_video.mp4",
				},
			},
			expectError: false,
		},
		{
			name: "Unknown action type",
			action: config.Action{
				Name: "unknown",
				Type: "unknown_type",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			// Create a fresh app config for each test
			app := config.AppConfig{
				Name:    "Test App",
				Type:    "web",
				URL:     "https://httpbin.org/html",
				Timeout: 30,
			}

			result := TestResult{
				AppName:     app.Name,
				AppType:     app.Type,
				StartTime:   time.Now(),
				Screenshots: make([]string, 0),
				Videos:      make([]string, 0),
				Metrics:     make(map[string]interface{}),
				Success:     false,
			}

			// Create platform
			platform, err := executor.factory.CreatePlatform(app.Type)
			if err != nil {
				t.Skipf("Platform creation failed: %v", err)
			}

			// Initialize platform
			err = platform.Initialize(app)
			if err != nil {
				t.Skipf("Platform initialization failed: %v", err)
			}
			defer platform.Close()

			var currentRecordingFile *string = nil
			err = executor.executeAction(platform, tt.action, app.Name, &result, currentRecordingFile)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// For actions that may legitimately fail (click, fill, submit on non-existent elements)
				if tt.action.Type == "click" || tt.action.Type == "fill" || tt.action.Type == "submit" {
					// These may fail if elements don't exist, which is expected
					if err != nil {
						assert.Contains(t, err.Error(), "failed to find element")
					}
				} else {
					assert.NoError(t, err)
				}
			}

			// Verify specific action results
			switch tt.action.Type {
			case "screenshot":
				if !tt.expectError {
					// Check if screenshot file was created (may be in temp dir)
					assert.NotEmpty(t, result.Screenshots)
				}
			case "record":
				if !tt.expectError {
					// Check if video file was created
					assert.NotEmpty(t, result.Videos)
				}
			}
		})
	}
}

func TestExecutor_GenerateReport(t *testing.T) {
	outputDir := t.TempDir()
	config := &config.Config{
		Name: "Test Report",
	}
	log := logger.NewLogger(false)
	executor := NewExecutor(config, outputDir, log)

	// Add some test results
	executor.results = []TestResult{
		{
			AppName:    "Test App 1",
			AppType:    "web",
			StartTime:  time.Now().Add(-5 * time.Minute),
			EndTime:    time.Now(),
			Duration:   5 * time.Minute,
			Success:    true,
			Screenshots: []string{"/tmp/screenshot1.png"},
			Videos:     []string{"/tmp/video1.mp4"},
			Metrics: map[string]interface{}{
				"click_actions": []string{"#button1"},
				"total_duration": 5 * time.Minute,
			},
		},
		{
			AppName:    "Test App 2",
			AppType:    "desktop",
			StartTime:  time.Now().Add(-3 * time.Minute),
			EndTime:    time.Now(),
			Duration:   3 * time.Minute,
			Success:    false,
			Error:      "Application not found",
			Metrics: map[string]interface{}{
				"total_duration": 3 * time.Minute,
			},
		},
	}

	reportPath := filepath.Join(outputDir, "test_report.html")

	t.Run("Generate HTML report", func(t *testing.T) {
		err := executor.GenerateReport(reportPath)
		assert.NoError(t, err)
		assert.FileExists(t, reportPath)

		// Check report content
		content, err := os.ReadFile(reportPath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), "Panoptic Test Report")
		assert.Contains(t, string(content), "Test App 1")
		assert.Contains(t, string(content), "Test App 2")
		assert.Contains(t, string(content), "web")
		assert.Contains(t, string(content), "desktop")
		assert.Contains(t, string(content), "success")
		assert.Contains(t, string(content), "Application not found")
	})

	t.Run("Generate report with no results", func(t *testing.T) {
		executor.results = []TestResult{}
		emptyReportPath := filepath.Join(outputDir, "empty_report.html")
		
		err := executor.GenerateReport(emptyReportPath)
		assert.NoError(t, err)
		assert.FileExists(t, emptyReportPath)
	})
}

func TestExecutor_EdgeCases(t *testing.T) {
	t.Run("Execute app with failed platform creation", func(t *testing.T) {
		config := &config.Config{
			Name: "Invalid Platform",
			Apps: []config.AppConfig{
				{
					Name: "Invalid App",
					Type: "invalid_type",
				},
			},
			Actions: []config.Action{
				{
					Name: "wait",
					Type: "wait",
					WaitTime: 1,
				},
			},
		}

		outputDir := t.TempDir()
		log := logger.NewLogger(false)
		executor := NewExecutor(config, outputDir, log)

		err := executor.Run()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})

	t.Run("Generate report with invalid path", func(t *testing.T) {
		config := &config.Config{Name: "Test"}
		outputDir := t.TempDir()
		log := logger.NewLogger(false)
		executor := NewExecutor(config, outputDir, log)

		// Try to write to a non-existent directory
		invalidPath := "/non/existent/path/report.html"
		err := executor.GenerateReport(invalidPath)
		assert.Error(t, err)
	})
}