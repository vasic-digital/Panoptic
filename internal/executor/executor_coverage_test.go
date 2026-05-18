package executor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/cloud"
	"panoptic/internal/config"
	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Additional tests for improving coverage of low-coverage functions

func TestExecutor_ExecuteCloudSync_WithFiles(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{
			Cloud: map[string]interface{}{
				"provider":     "aws",
				"bucket":       "test-bucket",
				"enable_sync":  true,
			},
		},
	}
	
	// Create temporary directory with test files
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "test1.txt")
	testFile2 := filepath.Join(tempDir, "test2.txt")
	
	assert.NoError(t, os.WriteFile(testFile1, []byte("test content 1"), 0644))
	assert.NoError(t, os.WriteFile(testFile2, []byte("test content 2"), 0644))
	
	// Create subdirectory with files
	subDir := filepath.Join(tempDir, "subdir")
	assert.NoError(t, os.Mkdir(subDir, 0755))
	subFile := filepath.Join(subDir, "sub.txt")
	assert.NoError(t, os.WriteFile(subFile, []byte("sub content"), 0644))
	
	executor := NewExecutor(cfg, tempDir, log)
	
	// Initialize and set a real cloud manager
	executor.getCloudManager()
	
	// Test executeCloudSync with app config
	app := config.AppConfig{Name: "Test App", Type: "web"}
	err := executor.executeCloudSync(app)
	// Should complete without error even if cloud upload fails (cloud manager is not fully configured)
	assert.NoError(t, err)
}

func TestExecutor_ExecuteCloudSync_ReadDirError(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{
			Cloud: map[string]interface{}{
				"provider":     "aws",
				"bucket":       "test-bucket",
				"enable_sync":  true,
			},
		},
	}
	
	// Use a non-existent directory
	tempDir := "/non/existent/directory"
	executor := NewExecutor(cfg, tempDir, log)
	
	// Initialize cloud manager
	executor.getCloudManager()
	
	// Test executeCloudSync - should handle ReadDir error gracefully
	app := config.AppConfig{Name: "Test App", Type: "web"}
	err := executor.executeCloudSync(app)
	// Should return an error about reading directory
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read output directory")
}

func TestExecutor_ExecuteDistributedCloudTest_WithNodes(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{
			Cloud: map[string]interface{}{
				"provider": "aws",
				"bucket":   "test-bucket",
				"distributed_nodes": []interface{}{
					map[string]interface{}{
						"id":       "node1",
						"name":     "Test Node 1",
						"location": "us-east-1",
						"capacity": "high",
						"endpoint": "https://test1.example.com",
						"api_key":  "test-key-1",
						"priority": 1,
					},
					map[string]interface{}{
						"id":       "node2",
						"name":     "Test Node 2",
						"location": "us-west-2",
						"capacity": "medium",
						"endpoint": "https://test2.example.com",
						"api_key":  "test-key-2",
						"priority": 2,
					},
				},
			},
		},
	}
	
	tempDir := t.TempDir()
	executor := NewExecutor(cfg, tempDir, log)
	
	// Initialize cloud manager
	executor.getCloudManager()
	
	// Test executeDistributedCloudTest
	app := config.AppConfig{Name: "Test App", Type: "web"}
	err := executor.executeDistributedCloudTest(app, config.Action{})
	// Should handle cloud manager not being fully configured
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "distributed test failed")
}

func TestExecutor_ExecuteCloudAnalytics_WithData(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{
			Cloud: map[string]interface{}{
				"provider": "aws",
				"bucket":   "test-bucket",
			},
		},
	}
	
	tempDir := t.TempDir()
	executor := NewExecutor(cfg, tempDir, log)
	
	// Add some test results
	executor.results = []TestResult{
		{
			AppName:   "Test App 1",
			AppType:   "web",
			StartTime: time.Now().Add(-time.Minute),
			EndTime:   time.Now().Add(-time.Second * 50),
			Duration:  10 * time.Second,
			Success:   true,
		},
		{
			AppName:   "Test App 2",
			AppType:   "web",
			StartTime: time.Now().Add(-time.Second * 40),
			EndTime:   time.Now().Add(-time.Second * 20),
			Duration:  20 * time.Second,
			Success:   false,
			Error:     "Test error",
		},
	}
	
	// Initialize cloud analytics
	executor.getCloudAnalytics()

	// Test executeCloudAnalytics
	// Anti-bluff (§11.4 / CONST-035, Article XI §11.9): the original assertion
	// expected an error and matched on "failed to generate analytics" — that
	// string was the OLD stub/TODO path. Commit e5f67a6 ("feat: implement
	// GenerateAnalytics and SaveReport — remove all TODO stubs") wired the
	// real analytics generator + JSON writer, so the live path now succeeds
	// for a populated executor.results slice. Asserting an error against a
	// feature that genuinely works is a PASS-bluff in the OPPOSITE direction
	// (a failing test masking working code). The correct end-user contract:
	// when results are present and a temp output directory is writable, the
	// analytics report MUST be produced. Verify that contract with runtime
	// evidence — file exists on disk, contents are non-empty JSON.
	app := config.AppConfig{Name: "Test App", Type: "web"}
	err := executor.executeCloudAnalytics(app)
	require.NoError(t, err, "executeCloudAnalytics should succeed with populated results + writable outputDir")

	reportPath := filepath.Join(tempDir, "cloud_analytics_report.json")
	info, statErr := os.Stat(reportPath)
	require.NoError(t, statErr, "analytics report file MUST exist (runtime evidence per §11.4)")
	assert.Greater(t, info.Size(), int64(0), "analytics report MUST be non-empty (runtime evidence per §11.4)")
}

func TestExecutor_CalculateSuccessRate_WithData(t *testing.T) {
	// Test with actual data using cloud.CloudTestResult
	results := []cloud.CloudTestResult{
		{Success: true},
		{Success: true},
		{Success: false},
		{Success: true},
		{Success: false},
	}
	
	successRate := calculateSuccessRate(results)
	assert.Equal(t, 60.0, successRate) // 3 out of 5 = 60%
	
	// Test with all success
	results = []cloud.CloudTestResult{
		{Success: true},
		{Success: true},
		{Success: true},
	}
	
	successRate = calculateSuccessRate(results)
	assert.Equal(t, 100.0, successRate) // 100%
	
	// Test with all failures
	results = []cloud.CloudTestResult{
		{Success: false},
		{Success: false},
		{Success: false},
	}
	
	successRate = calculateSuccessRate(results)
	assert.Equal(t, 0.0, successRate) // 0%
}

func TestExecutor_ExecuteAction_MorePlatformActions(t *testing.T) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	
	outputDir := t.TempDir()
	executor := NewExecutor(cfg, outputDir, log)
	
	app := config.AppConfig{Name: "Test", Type: "web"}
	
	// Test navigate action
	action := config.Action{Type: "navigate", Value: "https://example.com"}
	result := TestResult{}
	err := executor.executeAction(nil, action, app, &result, nil)
	// Should fail gracefully without platform
	assert.Error(t, err)
	
	// Test click action
	action = config.Action{Type: "click", Selector: "#button"}
	result = TestResult{}
	err = executor.executeAction(nil, action, app, &result, nil)
	// Should fail gracefully without platform
	assert.Error(t, err)
	
	// Test fill action
	action = config.Action{Type: "fill", Selector: "#input", Value: "test"}
	result = TestResult{}
	err = executor.executeAction(nil, action, app, &result, nil)
	// Should fail gracefully without platform
	assert.Error(t, err)
	
	// Test submit action
	action = config.Action{Type: "submit", Selector: "#form"}
	result = TestResult{}
	err = executor.executeAction(nil, action, app, &result, nil)
	// Should fail gracefully without platform
	assert.Error(t, err)
	
	// Test screenshot action
	action = config.Action{Type: "screenshot"}
	result = TestResult{}
	err = executor.executeAction(nil, action, app, &result, nil)
	// Should fail gracefully without platform
	assert.Error(t, err)
	
	// Test record action
	action = config.Action{Type: "record", Duration: 5}
	result = TestResult{}
	err = executor.executeAction(nil, action, app, &result, nil)
	// Should fail gracefully without platform
	assert.Error(t, err)
	
	// Test vision_click action
	action = config.Action{Type: "vision_click"}
	result = TestResult{}
	err = executor.executeAction(nil, action, app, &result, nil)
	// Should fail gracefully without platform
	assert.Error(t, err)
	
	// Test wait action (should work without platform)
	action = config.Action{Type: "wait", WaitTime: 1}
	result = TestResult{}
	start := time.Now()
	err = executor.executeAction(nil, action, app, &result, nil)
	elapsed := time.Since(start)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, time.Second)
}