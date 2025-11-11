package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
)

// Test helper functions

func TestGetStringFromMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{"valid string", map[string]interface{}{"key": "value"}, "key", "value"},
		{"missing key", map[string]interface{}{}, "key", ""},
		{"wrong type", map[string]interface{}{"key": 123}, "key", ""},
		{"nil map", nil, "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringFromMap(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBoolFromMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected bool
	}{
		{"valid bool true", map[string]interface{}{"key": true}, "key", true},
		{"valid bool false", map[string]interface{}{"key": false}, "key", false},
		{"missing key", map[string]interface{}{}, "key", false},
		{"wrong type", map[string]interface{}{"key": "true"}, "key", false},
		{"nil map", nil, "key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolFromMap(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntFromMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected int
	}{
		{"valid int", map[string]interface{}{"key": 42}, "key", 42},
		{"valid int64", map[string]interface{}{"key": int64(42)}, "key", 42},
		{"valid float64", map[string]interface{}{"key": float64(42.5)}, "key", 42},
		{"missing key", map[string]interface{}{}, "key", 0},
		{"wrong type", map[string]interface{}{"key": "42"}, "key", 0},
		{"nil map", nil, "key", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntFromMap(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test executor constructor

func TestNewExecutor(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{
			{Name: "Test App", Type: "web", URL: "https://example.com"},
		},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	assert.NotNil(t, executor)
	assert.Equal(t, cfg, executor.config)
	assert.Equal(t, outputDir, executor.outputDir)
	assert.NotNil(t, executor.logger)
	assert.NotNil(t, executor.factory)
	assert.NotNil(t, executor.results)
	assert.Empty(t, executor.results)
	assert.NotNil(t, executor.aiTester)
	assert.NotNil(t, executor.cloudManager)
	assert.NotNil(t, executor.cloudAnalytics)
	assert.NotNil(t, executor.enterpriseIntegration)
}

func TestNewExecutor_WithCloudConfig(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
		Settings: config.Settings{
			Cloud: map[string]interface{}{
				"provider":    "local",
				"bucket":      "test-bucket",
				"enable_sync": true,
			},
		},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.cloudManager)
}

func TestNewExecutor_WithEnterpriseConfig(t *testing.T) {
	log := logger.NewLogger(false)
	outputDir := t.TempDir()

	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
		Settings: config.Settings{
			Enterprise: map[string]interface{}{
				"enabled": true,
			},
		},
	}

	executor := NewExecutor(cfg, outputDir, log)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.enterpriseIntegration)
}

// Test TestResult struct

func TestTestResult_Creation(t *testing.T) {
	result := TestResult{
		AppName:     "Test App",
		AppType:     "web",
		StartTime:   time.Now(),
		Screenshots: []string{"screenshot1.png"},
		Videos:      []string{"video1.mp4"},
		Metrics:     map[string]interface{}{"clicks": 5},
		Success:     true,
	}

	assert.Equal(t, "Test App", result.AppName)
	assert.Equal(t, "web", result.AppType)
	assert.True(t, result.Success)
	assert.Len(t, result.Screenshots, 1)
	assert.Len(t, result.Videos, 1)
	assert.Equal(t, 5, result.Metrics["clicks"])
}

func TestTestResult_JSONMarshaling(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)

	result := TestResult{
		AppName:     "Test App",
		AppType:     "web",
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    5 * time.Second,
		Screenshots: []string{"test.png"},
		Videos:      []string{"test.mp4"},
		Metrics:     map[string]interface{}{"clicks": 3},
		Success:     true,
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded TestResult
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "Test App", decoded.AppName)
	assert.Equal(t, "web", decoded.AppType)
	assert.True(t, decoded.Success)
}

// Test Run function

func TestExecutor_Run_InvalidConfig(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "", // Invalid: empty name
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)
	err := executor.Run()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestExecutor_Run_ValidConfig(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{
			{Name: "Test App", Type: "web", URL: "https://example.com"},
		},
		Actions: []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)
	err := executor.Run()

	// Should complete without error (though platform init will fail in test env)
	assert.NoError(t, err)
	assert.Len(t, executor.results, 1)
}

// Test executeApp function

func TestExecutor_ExecuteApp_InvalidPlatformType(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	app := config.AppConfig{
		Name: "Test App",
		Type: "invalid_platform",
	}

	result := executor.executeApp(app)

	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "Failed to create platform")
	assert.Equal(t, "Test App", result.AppName)
	assert.Equal(t, "invalid_platform", result.AppType)
}

func TestExecutor_ExecuteApp_EmptyActions(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	app := config.AppConfig{
		Name: "Test App",
		Type: "web",
		URL:  "https://example.com",
	}

	result := executor.executeApp(app)

	// Should fail during initialization (browser not available in test environment)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)
}

// Test executeAction function

func TestExecutor_ExecuteAction_UnknownType(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	action := config.Action{
		Type: "unknown_action",
	}

	err := executor.executeAction(nil, action, config.AppConfig{}, &TestResult{}, new(string))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action type")
}

// Test cloud functions

func TestCalculateSuccessRate(t *testing.T) {
	tests := []struct {
		name     string
		results  []interface{} // Using interface{} since CloudTestResult might not be available
		expected float64
	}{
		{
			name:     "empty results",
			results:  []interface{}{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with empty slice
			result := calculateSuccessRate(nil)
			assert.Equal(t, 0.0, result)
		})
	}
}

// Test enterprise config file creation

func TestExecutor_CreateEnterpriseConfigFile(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	configPath := filepath.Join(outputDir, "enterprise_config.yaml")
	enterpriseConfig := map[string]interface{}{
		"enabled": true,
		"organization": map[string]interface{}{
			"name": "Test Org",
		},
	}

	err := executor.createEnterpriseConfigFile(configPath, enterpriseConfig)

	assert.NoError(t, err)
	assert.FileExists(t, configPath)

	// Verify file content
	content, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "enabled: true")
	assert.Contains(t, string(content), "Test Org")
}

func TestExecutor_CreateEnterpriseConfigFile_NilConfig(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	configPath := filepath.Join(outputDir, "enterprise_config.yaml")

	err := executor.createEnterpriseConfigFile(configPath, nil)

	assert.NoError(t, err)
	assert.FileExists(t, configPath)

	// Verify default config is created
	content, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Default Organization")
}

// Test enterprise status execution

func TestExecutor_ExecuteEnterpriseStatus_NotInitialized(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	action := config.Action{
		Type: "enterprise_status",
	}

	err := executor.executeEnterpriseStatus(config.AppConfig{}, action)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// Test AI functions error handling

func TestExecutor_GenerateAITests_NilTester(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)
	executor.aiTester = nil

	err := executor.generateAITests(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI tester not initialized")
}

func TestExecutor_GenerateSmartErrorDetection_NilTester(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)
	executor.aiTester = nil

	err := executor.generateSmartErrorDetection(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI tester not initialized")
}

func TestExecutor_ExecuteAIEnhancedTesting_NilTester(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)
	executor.aiTester = nil

	err := executor.executeAIEnhancedTesting(nil, config.AppConfig{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AI tester not initialized")
}

// Test cloud functions error handling

func TestExecutor_ExecuteCloudSync_NilManager(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)
	executor.cloudManager = nil

	err := executor.executeCloudSync(config.AppConfig{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloud manager not initialized")
}

func TestExecutor_ExecuteCloudAnalytics_NilAnalytics(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)
	executor.cloudAnalytics = nil

	err := executor.executeCloudAnalytics(config.AppConfig{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloud analytics not initialized")
}

func TestExecutor_ExecuteDistributedCloudTest_NilManager(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)
	executor.cloudManager = nil

	err := executor.executeDistributedCloudTest(config.AppConfig{}, config.Action{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloud manager not initialized")
}

// Test report generation

func TestExecutor_GenerateReport(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	// Add some test results
	executor.results = []TestResult{
		{AppName: "App1", Success: true},
		{AppName: "App2", Success: false},
	}

	reportPath := filepath.Join(outputDir, "report.html")
	err := executor.GenerateReport(reportPath)

	assert.NoError(t, err)
	assert.FileExists(t, reportPath)

	// Verify report content
	content, err := os.ReadFile(reportPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "<!DOCTYPE html>")
	assert.Contains(t, string(content), "Test Report")
	assert.Contains(t, string(content), "Total Tests: 2")
}

func TestExecutor_GenerateReport_EmptyResults(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	reportPath := filepath.Join(outputDir, "report.html")
	err := executor.GenerateReport(reportPath)

	assert.NoError(t, err)
	assert.FileExists(t, reportPath)

	content, err := os.ReadFile(reportPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Total Tests: 0")
}

// Test save enterprise report

func TestExecutor_SaveEnterpriseReport(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	report := map[string]interface{}{
		"status":  "success",
		"message": "Test report",
		"data":    []string{"item1", "item2"},
	}

	reportPath := filepath.Join(outputDir, "enterprise_report.json")
	err := executor.saveEnterpriseReport(report, reportPath)

	assert.NoError(t, err)
	assert.FileExists(t, reportPath)

	// Verify content
	content, err := os.ReadFile(reportPath)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(content, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "success", decoded["status"])
	assert.Equal(t, "Test report", decoded["message"])
}

// Test action parameter extraction (parameters tested in integration test)

// Test action validation

func TestExecutor_ExecuteAction_ClickWithoutSelector(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	action := config.Action{
		Type: "click",
		// No selector or target
	}

	err := executor.executeAction(nil, action, config.AppConfig{}, &TestResult{}, new(string))

	// Should return nil (action ignored) or error depending on implementation
	assert.NoError(t, err) // Current implementation returns nil for empty selector
}

func TestExecutor_ExecuteAction_FillWithoutValue(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	action := config.Action{
		Type:     "fill",
		Selector: "input",
		// No value
	}

	err := executor.executeAction(nil, action, config.AppConfig{}, &TestResult{}, new(string))

	// Should return nil (action ignored) since value is empty
	assert.NoError(t, err)
}

// Test cloud config parsing

func TestExecutor_CloudConfigWithRetentionPolicy(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
		Settings: config.Settings{
			Cloud: map[string]interface{}{
				"provider": "local",
				"bucket":   "test-bucket",
				"retention_policy": map[string]interface{}{
					"enabled":      true,
					"days":         30,
					"max_size_gb":  100,
					"auto_cleanup": true,
				},
			},
		},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.cloudManager)
}

func TestExecutor_CloudConfigWithDistributedNodes(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{},
		Actions: []config.Action{},
		Settings: config.Settings{
			Cloud: map[string]interface{}{
				"provider":           "local",
				"enable_distributed": true,
				"distributed_nodes": []interface{}{
					map[string]interface{}{
						"id":       "node1",
						"name":     "Node 1",
						"location": "us-east",
						"capacity": "high",
						"endpoint": "http://node1.example.com",
						"api_key":  "key123",
						"priority": 1,
					},
				},
			},
		},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.cloudManager)
}

// Integration test

func TestExecutor_Integration_SimpleWorkflow(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Integration Test",
		Apps: []config.AppConfig{
			{Name: "Test App", Type: "web", URL: "https://example.com"},
		},
		Actions: []config.Action{
			{Type: "wait", WaitTime: 1},
		},
		Settings: config.Settings{},
	}
	outputDir := t.TempDir()

	executor := NewExecutor(cfg, outputDir, log)

	err := executor.Run()
	assert.NoError(t, err)

	// Generate report
	reportPath := filepath.Join(outputDir, "report.html")
	err = executor.GenerateReport(reportPath)
	assert.NoError(t, err)
	assert.FileExists(t, reportPath)
}
