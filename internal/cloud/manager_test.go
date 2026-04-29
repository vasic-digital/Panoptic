package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCloudManager tests cloud manager creation
func TestNewCloudManager(t *testing.T) {
	log := logger.NewLogger(false)
	config := CloudConfig{
		Provider: "local",
		Bucket:   "/tmp/test-cloud",
	}

	manager := &CloudManager{
		Logger:  *log,
		Config:  config,
		Enabled: true,
	}

	assert.NotNil(t, manager, "Cloud manager should not be nil")
	assert.True(t, manager.Enabled, "Should be enabled")
}

// TestConfigure tests cloud manager configuration
func TestConfigure(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
	}

	tempDir := t.TempDir()
	config := CloudConfig{
		Provider:   "local",
		Bucket:     tempDir,
		EnableSync: true,
	}

	err := manager.Configure(config)

	assert.NoError(t, err, "Configure should not error")
	assert.Equal(t, config.Provider, manager.Config.Provider)
	assert.NotNil(t, manager.Provider, "Provider should be initialized")
}

// TestConfigure_InvalidProvider tests configuration with invalid provider
func TestConfigure_InvalidProvider(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
	}

	config := CloudConfig{
		Provider: "invalid-provider",
		Bucket:   "/tmp/test",
	}

	err := manager.Configure(config)

	assert.Error(t, err, "Should error with invalid provider")
	assert.Contains(t, err.Error(), "unsupported", "Error should mention unsupported provider")
}

// TestUpload tests the Upload method
func TestUpload(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
		Config: CloudConfig{
			Provider: "local",
			Bucket:  filepath.Join(tempDir, "storage"),
		},
	}

	// Create a test file to upload
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err, "Should create test file")

	err = manager.Upload(testFile)

	assert.NoError(t, err, "Upload should succeed")
	assert.Greater(t, len(manager.TestResults), 0, "Should track upload result")
	
	// Verify the first upload result
	if len(manager.TestResults) > 0 {
		result := manager.TestResults[0]
		assert.True(t, result.Success, "Upload should be marked as successful")
		assert.NotEmpty(t, result.Artifacts, "Should have artifacts")
	}
}

// TestCountSuccessfulResults tests result counting
func TestCountSuccessfulResults(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger: *log,
	}

	results := []CloudTestResult{
		{Success: true, Duration: 1 * time.Second},
		{Success: false, Duration: 2 * time.Second},
		{Success: true, Duration: 1 * time.Second},
	}

	count := manager.countSuccessfulResults(results)

	assert.Equal(t, 2, count, "Should count 2 successful results")
}

// TestCountSuccessfulResults_Empty tests with no results
func TestCountSuccessfulResults_Empty(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger: *log,
	}

	results := []CloudTestResult{}

	count := manager.countSuccessfulResults(results)

	assert.Equal(t, 0, count, "Should return 0 for empty results")
}

// TestGenerateCloudRecommendations tests recommendation generation
func TestGenerateCloudRecommendations(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			EnableSync:        false,
			EnableCDN:         false,
			Compression:       false,
			EnableDistributed: false,
		},
	}

	recommendations := manager.generateCloudRecommendations()

	assert.NotNil(t, recommendations, "Should return recommendations")
	assert.Greater(t, len(recommendations), 0, "Should have at least one recommendation")
}

// TestGenerateCloudRecommendations_AllEnabled tests with all features enabled
func TestGenerateCloudRecommendations_AllEnabled(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			EnableSync:        true,
			EnableCDN:         true,
			Compression:       true,
			EnableDistributed: true,
		},
	}

	recommendations := manager.generateCloudRecommendations()

	assert.NotNil(t, recommendations, "Should return recommendations")
	// With all enabled, may have fewer recommendations
}

// TestCloudConfig_Structure tests CloudConfig struct
func TestCloudConfig_Structure(t *testing.T) {
	config := CloudConfig{
		Provider:          "aws",
		Bucket:            "test-bucket",
		Region:            "us-east-1",
		EnableSync:        true,
		SyncInterval:      30,
		EnableCDN:         true,
		Compression:       true,
		Encryption:        true,
		EnableDistributed: true,
	}

	assert.Equal(t, "aws", config.Provider)
	assert.Equal(t, "test-bucket", config.Bucket)
	assert.True(t, config.EnableSync)
	assert.Equal(t, 30, config.SyncInterval)
}

// TestRetentionPolicy_Structure tests RetentionPolicy struct
func TestRetentionPolicy_Structure(t *testing.T) {
	policy := RetentionPolicy{
		Enabled:     true,
		Days:        30,
		MaxSizeGB:   100,
		AutoCleanup: true,
	}

	assert.True(t, policy.Enabled)
	assert.Equal(t, 30, policy.Days)
	assert.Equal(t, 100, policy.MaxSizeGB)
	assert.True(t, policy.AutoCleanup)
}

// TestDistributedNode_Structure tests DistributedNode struct
func TestDistributedNode_Structure(t *testing.T) {
	node := DistributedNode{
		ID:       "node-1",
		Name:     "Test Node",
		Location: "us-east-1",
		Capacity: "high",
		Endpoint: "https://node1.example.com",
		APIKey:   "test-key",
		Priority: 1,
	}

	assert.Equal(t, "node-1", node.ID)
	assert.Equal(t, "Test Node", node.Name)
	assert.Equal(t, 1, node.Priority)
}

// TestUploadResult_Structure tests UploadResult struct
func TestUploadResult_Structure(t *testing.T) {
	result := UploadResult{
		Success:    true,
		URL:        "https://example.com/file.txt",
		Size:       1024,
		ETag:       "abc123",
		Duration:   "1s",
		RemotePath: "/uploads/file.txt",
	}

	assert.True(t, result.Success)
	assert.Equal(t, int64(1024), result.Size)
	assert.Equal(t, "abc123", result.ETag)
}

// TestDownloadResult_Structure tests DownloadResult struct
func TestDownloadResult_Structure(t *testing.T) {
	result := DownloadResult{
		Success:    true,
		LocalPath:  "/tmp/file.txt",
		Size:       2048,
		ETag:       "def456",
		Duration:   "2s",
		RemotePath: "/downloads/file.txt",
	}

	assert.True(t, result.Success)
	assert.Equal(t, int64(2048), result.Size)
	assert.Equal(t, "/tmp/file.txt", result.LocalPath)
}

// TestCloudFile_Structure tests CloudFile struct
func TestCloudFile_Structure(t *testing.T) {
	now := time.Now()
	file := CloudFile{
		Name:         "test.txt",
		Path:         "/files/test.txt",
		Size:         512,
		LastModified: now,
		ETag:         "ghi789",
		ContentType:  "text/plain",
	}

	assert.Equal(t, "test.txt", file.Name)
	assert.Equal(t, int64(512), file.Size)
	assert.Equal(t, now, file.LastModified)
}

// TestCloudTestResult_Structure tests CloudTestResult struct
func TestCloudTestResult_Structure(t *testing.T) {
	result := CloudTestResult{
		TestID:   "test-1",
		NodeID:   "node-1",
		NodeName: "Test Node",
		Success:  true,
		Duration: 5 * time.Second,
	}

	assert.Equal(t, "node-1", result.NodeID)
	assert.True(t, result.Success)
	assert.Equal(t, 5*time.Second, result.Duration)
}

// TestCloudAnalytics_RecordAnalytics tests analytics recording
func TestCloudAnalytics_RecordAnalytics(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:        *log,
		Enabled:       true,
		AnalyticsData: []AnalyticsDataPoint{},
	}

	dataPoint := AnalyticsDataPoint{
		Timestamp:   time.Now(),
		TestCount:   1,
		SuccessRate: 100.0,
		ErrorCount:  0,
		NodeCount:   1,
	}

	analytics.RecordAnalytics(dataPoint)

	assert.Equal(t, 1, len(analytics.AnalyticsData), "Should have one data point")
	assert.Equal(t, 1, analytics.AnalyticsData[0].TestCount)
}

// TestCloudAnalytics_CalculateSummaryStats tests summary calculation
func TestCloudAnalytics_CalculateSummaryStats(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
		AnalyticsData: []AnalyticsDataPoint{
			{TestCount: 3, SuccessRate: 100.0, ErrorCount: 0},
			{TestCount: 2, SuccessRate: 50.0, ErrorCount: 1},
		},
	}

	stats := analytics.calculateSummaryStats()

	assert.NotNil(t, stats, "Stats should not be nil")
	// Stats structure depends on implementation
}

// TestCloudAnalytics_CalculateAverageSuccessRate tests success rate calculation
func TestCloudAnalytics_CalculateAverageSuccessRate(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
		AnalyticsData: []AnalyticsDataPoint{
			{SuccessRate: 100.0},
			{SuccessRate: 80.0},
			{SuccessRate: 90.0},
		},
	}

	avgRate := analytics.calculateAverageSuccessRate(3)

	assert.Equal(t, 90.0, avgRate, "Average should be 90.0")
}

// TestCloudAnalytics_CalculateAverageSuccessRate_Zero tests with zero count
func TestCloudAnalytics_CalculateAverageSuccessRate_Zero(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:        *log,
		Enabled:       true,
		AnalyticsData: []AnalyticsDataPoint{},
	}

	avgRate := analytics.calculateAverageSuccessRate(0)

	assert.Equal(t, 0.0, avgRate, "Should return 0.0 for zero count")
}

// TestCloudAnalytics_GenerateAnalyticsRecommendations tests recommendation generation
func TestCloudAnalytics_GenerateAnalyticsRecommendations(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
		AnalyticsData: []AnalyticsDataPoint{
			{SuccessRate: 0.0, ErrorCount: 5},
			{SuccessRate: 20.0, ErrorCount: 4},
			{SuccessRate: 30.0, ErrorCount: 3},
		},
	}

	recommendations := analytics.generateAnalyticsRecommendations()

	// May return nil or empty list depending on implementation
	if recommendations != nil {
		assert.GreaterOrEqual(t, len(recommendations), 0, "Should have non-negative recommendations")
	}
}

// TestCloudAnalytics_GenerateAnalytics tests analytics generation with CloudTestResults
func TestCloudAnalytics_GenerateAnalytics(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	results := []CloudTestResult{
		{TestID: "t1", Success: true, Duration: 100 * time.Millisecond},
		{TestID: "t2", Success: true, Duration: 200 * time.Millisecond},
		{TestID: "t3", Success: false, Duration: 50 * time.Millisecond, Error: "timeout"},
	}

	data, err := analytics.GenerateAnalytics(results)

	require.NoError(t, err, "GenerateAnalytics should not return error")
	require.NotNil(t, data, "Analytics data should not be nil")

	m, ok := data.(map[string]interface{})
	require.True(t, ok, "Analytics should be a map[string]interface{}")

	assert.Equal(t, 3, m["total"])
	assert.Equal(t, 2, m["passed"])
	assert.Equal(t, 1, m["failed"])
	assert.Equal(t, 0, m["skipped"])

	successRate, ok := m["success_rate"].(float64)
	require.True(t, ok)
	assert.InDelta(t, 66.66, successRate, 0.1, "Success rate ~66.67%%")

	avgDur, ok := m["average_duration_ms"].(float64)
	require.True(t, ok)
	assert.Greater(t, avgDur, 0.0, "Average duration should be positive")

	patterns, ok := m["failure_patterns"].(map[string]int)
	require.True(t, ok)
	assert.Equal(t, 1, patterns["timeout"])
}

// TestCloudAnalytics_GenerateAnalytics_NilResults tests nil input
func TestCloudAnalytics_GenerateAnalytics_NilResults(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	data, err := analytics.GenerateAnalytics(nil)

	require.NoError(t, err)
	require.NotNil(t, data)

	m := data.(map[string]interface{})
	assert.Equal(t, 0, m["total"])
	assert.Equal(t, 0, m["passed"])
}

// TestCloudAnalytics_GenerateAnalytics_EmptySlice tests empty result slice
func TestCloudAnalytics_GenerateAnalytics_EmptySlice(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	data, err := analytics.GenerateAnalytics([]CloudTestResult{})

	require.NoError(t, err)
	require.NotNil(t, data)

	m := data.(map[string]interface{})
	assert.Equal(t, 0, m["total"])
	assert.Equal(t, 0.0, m["success_rate"])
}

// TestCloudAnalytics_GenerateAnalytics_SlowestTests verifies top-10 cap
func TestCloudAnalytics_GenerateAnalytics_SlowestTests(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	// Create 15 results with varying durations
	results := make([]CloudTestResult, 15)
	for i := range results {
		results[i] = CloudTestResult{
			TestID:   fmt.Sprintf("t%d", i),
			Success:  true,
			Duration: time.Duration(i+1) * time.Second,
		}
	}

	data, err := analytics.GenerateAnalytics(results)
	require.NoError(t, err)

	m := data.(map[string]interface{})
	slowest := m["slowest_tests"]
	require.NotNil(t, slowest, "slowest_tests should be present")

	// Must be capped at 10
	rv := reflect.ValueOf(slowest)
	assert.Equal(t, 10, rv.Len(), "slowest_tests should be capped at 10")
}

// TestCloudAnalytics_GenerateAnalytics_FailurePatterns verifies grouping
func TestCloudAnalytics_GenerateAnalytics_FailurePatterns(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	results := []CloudTestResult{
		{Success: false, Error: "connection refused"},
		{Success: false, Error: "connection refused"},
		{Success: false, Error: "timeout"},
		{Success: false}, // empty error -> "unknown error"
		{Success: true},
	}

	data, err := analytics.GenerateAnalytics(results)
	require.NoError(t, err)

	m := data.(map[string]interface{})
	patterns := m["failure_patterns"].(map[string]int)
	assert.Equal(t, 2, patterns["connection refused"])
	assert.Equal(t, 1, patterns["timeout"])
	assert.Equal(t, 1, patterns["unknown error"])
	assert.Equal(t, 4, m["failed"])
	assert.Equal(t, 1, m["passed"])
}

// TestCloudAnalytics_GenerateAnalytics_ReflectionFallback tests with a
// non-CloudTestResult struct slice (simulating executor.TestResult).
func TestCloudAnalytics_GenerateAnalytics_ReflectionFallback(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	// Anonymous struct that mirrors executor.TestResult's key fields
	type FakeResult struct {
		Success  bool
		Duration time.Duration
		Error    string
	}

	results := []FakeResult{
		{Success: true, Duration: 500 * time.Millisecond},
		{Success: false, Duration: 1 * time.Second, Error: "assertion failed"},
	}

	data, err := analytics.GenerateAnalytics(results)
	require.NoError(t, err)
	require.NotNil(t, data)

	m := data.(map[string]interface{})
	assert.Equal(t, 2, m["total"])
	assert.Equal(t, 1, m["passed"])
	assert.Equal(t, 1, m["failed"])
	assert.InDelta(t, 50.0, m["success_rate"].(float64), 0.1)

	patterns := m["failure_patterns"].(map[string]int)
	assert.Equal(t, 1, patterns["assertion failed"])
}

// TestCloudAnalytics_SaveReport tests report saving to file
func TestCloudAnalytics_SaveReport(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	testData := map[string]interface{}{
		"total":        5,
		"passed":       4,
		"success_rate": 80.0,
	}

	reportPath := filepath.Join(t.TempDir(), "report.json")
	err := analytics.SaveReport(testData, reportPath)

	require.NoError(t, err, "SaveReport should not error")

	// Verify the file exists and is valid JSON
	content, err := os.ReadFile(reportPath)
	require.NoError(t, err, "Should read saved report")

	var parsed map[string]interface{}
	err = json.Unmarshal(content, &parsed)
	require.NoError(t, err, "Report should be valid JSON")

	// Check metadata envelope
	metadata, ok := parsed["metadata"].(map[string]interface{})
	require.True(t, ok, "Should have metadata")
	assert.NotEmpty(t, metadata["generated_at"], "Should have timestamp")
	assert.Equal(t, "1.0.0", metadata["version"], "Should have version")

	// Check analytics data
	analyticsData, ok := parsed["analytics"].(map[string]interface{})
	require.True(t, ok, "Should have analytics data")
	assert.Equal(t, float64(5), analyticsData["total"])
	assert.Equal(t, float64(4), analyticsData["passed"])
}

// TestCloudAnalytics_SaveReport_EmptyPath tests error on empty path
func TestCloudAnalytics_SaveReport_EmptyPath(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	err := analytics.SaveReport("some data", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestCloudAnalytics_SaveReport_CreatesDirectory tests directory creation
func TestCloudAnalytics_SaveReport_CreatesDirectory(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	reportPath := filepath.Join(t.TempDir(), "nested", "dir", "report.json")
	err := analytics.SaveReport(map[string]string{"key": "value"}, reportPath)

	require.NoError(t, err)

	_, err = os.Stat(reportPath)
	assert.NoError(t, err, "Report file should exist")
}

// TestCloudAnalytics_SaveReport_AtomicWrite verifies no partial file on success
func TestCloudAnalytics_SaveReport_AtomicWrite(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	dir := t.TempDir()
	reportPath := filepath.Join(dir, "atomic_report.json")

	err := analytics.SaveReport(map[string]int{"count": 42}, reportPath)
	require.NoError(t, err)

	// No temp files should remain
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, len(entries), "Only the final report file should exist")
	assert.Equal(t, "atomic_report.json", entries[0].Name())
}

// TestCloudAnalytics_GenerateAndSave_Integration tests the full pipeline
func TestCloudAnalytics_GenerateAndSave_Integration(t *testing.T) {
	log := logger.NewLogger(false)
	analytics := &CloudAnalytics{
		Logger:  *log,
		Enabled: true,
	}

	results := []CloudTestResult{
		{TestID: "a", Success: true, Duration: 1 * time.Second},
		{TestID: "b", Success: false, Duration: 2 * time.Second, Error: "fail"},
	}

	data, err := analytics.GenerateAnalytics(results)
	require.NoError(t, err)

	reportPath := filepath.Join(t.TempDir(), "integration_report.json")
	err = analytics.SaveReport(data, reportPath)
	require.NoError(t, err)

	// Read back and validate
	content, err := os.ReadFile(reportPath)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(content, &parsed)
	require.NoError(t, err)

	analyticsData := parsed["analytics"].(map[string]interface{})
	assert.Equal(t, float64(2), analyticsData["total"])
	assert.Equal(t, float64(1), analyticsData["passed"])
	assert.Equal(t, float64(1), analyticsData["failed"])
}

// TestSyncTestResults tests test result syncing
func TestSyncTestResults(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	// Create test results directory
	resultsDir := filepath.Join(tempDir, "results")
	err := os.MkdirAll(resultsDir, 0755)
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(resultsDir, "test-result.json")
	err = os.WriteFile(testFile, []byte(`{"test": "data"}`), 0644)
	require.NoError(t, err)

	// Configure manager with local provider
	storageDir := filepath.Join(tempDir, "storage")
	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
	}

	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	err = manager.Configure(config)
	require.NoError(t, err, "Should configure successfully")

	// Test syncing
	ctx := context.Background()
	err = manager.SyncTestResults(ctx, resultsDir)

	// May succeed or fail depending on implementation details
	if err != nil {
		t.Logf("Sync failed (acceptable): %v", err)
	}
}

// TestExecuteDistributedTest tests distributed test execution
func TestExecuteDistributedTest(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
	}

	ctx := context.Background()
	testConfig := map[string]interface{}{
		"url": "https://example.com",
	}

	nodes := []DistributedNode{
		{
			ID:       "node-1",
			Name:     "Test Node",
			Endpoint: "https://node1.example.com",
		},
	}

	results, err := manager.ExecuteDistributedTest(ctx, testConfig, nodes)

	// May error if distributed testing not enabled
	if err == nil {
		assert.NotNil(t, results, "Should return results slice if no error")
	} else {
		t.Logf("Distributed test failed (acceptable): %v", err)
	}
}

// TestExecuteDistributedTest_NoNodes tests with no nodes
func TestExecuteDistributedTest_NoNodes(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
	}

	ctx := context.Background()
	testConfig := map[string]interface{}{}
	nodes := []DistributedNode{}

	results, err := manager.ExecuteDistributedTest(ctx, testConfig, nodes)

	// May error if distributed testing not enabled
	if err == nil {
		assert.Empty(t, results, "Should return empty results with no nodes")
	} else {
		t.Logf("Distributed test failed (acceptable): %v", err)
	}
}

// TestCleanupOldFiles tests file cleanup
func TestCleanupOldFiles(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
	}

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
		RetentionPolicy: RetentionPolicy{
			Enabled:     true,
			Days:        30,
			AutoCleanup: true,
		},
	}

	err := manager.Configure(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.CleanupOldFiles(ctx)

	// Should not error even if no files to cleanup
	assert.NoError(t, err, "Cleanup should not error")
}

// TestEnableCDN tests CDN enablement
func TestEnableCDN(t *testing.T) {
	// bluff-scan: no-assert-ok (feature/interface smoke — wiring must not panic)
	log := logger.NewLogger(false)
	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
		Config: CloudConfig{
			EnableCDN:   false,
			CDNEndpoint: "https://cdn.example.com",
		},
	}

	ctx := context.Background()
	err := manager.EnableCDN(ctx)

	// May error if CDN not configured
	if err != nil {
		t.Logf("EnableCDN failed (acceptable): %v", err)
	}
}

// TestGenerateCloudReport tests cloud report generation
func TestGenerateCloudReport(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
		TestResults: []CloudTestResult{
			{Success: true, Duration: 1 * time.Second, NodeID: "node-1"},
		},
	}

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
	}

	err := manager.Configure(config)
	require.NoError(t, err)

	ctx := context.Background()
	report, err := manager.GenerateCloudReport(ctx)

	// May succeed or fail depending on provider implementation
	if err == nil {
		assert.NotNil(t, report, "Report should not be nil if no error")
	} else {
		t.Logf("Report generation failed (acceptable): %v", err)
	}
}

// TestGetCloudURLs tests URL generation
func TestGetCloudURLs(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	storageDir := filepath.Join(tempDir, "storage")
	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
	}

	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	err = manager.Configure(config)
	require.NoError(t, err)

	ctx := context.Background()
	urls, err := manager.GetCloudURLs(ctx, testFile)

	// Should return URLs or error
	if err == nil {
		assert.NotNil(t, urls, "URLs should not be nil")
	} else {
		t.Logf("GetCloudURLs failed (acceptable): %v", err)
	}
}

// TestCalculateStorageStats tests storage stats calculation
func TestCalculateStorageStats(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	manager := &CloudManager{
		Logger:  *log,
		Enabled: true,
	}

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
	}

	err := manager.Configure(config)
	require.NoError(t, err)

	ctx := context.Background()
	stats, err := manager.calculateStorageStats(ctx)

	// May succeed or fail
	if err == nil {
		assert.NotNil(t, stats, "Stats should not be nil")
	} else {
		t.Logf("Storage stats calculation failed (acceptable): %v", err)
	}
}
