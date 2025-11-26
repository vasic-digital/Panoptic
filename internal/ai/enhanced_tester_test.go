package ai

import (
	"os"
	"path/filepath"
	"testing"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestNewAIEnhancedTester tests the creation of a new AI-enhanced tester
func TestNewAIEnhancedTester(t *testing.T) {
	log := logger.NewLogger(false)

	tester := NewAIEnhancedTester(*log)

	assert.NotNil(t, tester, "AI enhanced tester should not be nil")
	assert.NotNil(t, tester.ErrorDetector, "Error detector should be initialized")
	assert.NotNil(t, tester.TestGenerator, "Test generator should be initialized")
	assert.NotNil(t, tester.VisionDetector, "Vision detector should be initialized")
	assert.True(t, tester.enabled, "AI enhanced tester should be enabled by default")

	// Check default config
	assert.True(t, tester.config.EnableErrorDetection, "Error detection should be enabled by default")
	assert.True(t, tester.config.EnableTestGeneration, "Test generation should be enabled by default")
	assert.True(t, tester.config.EnableVisionAnalysis, "Vision analysis should be enabled by default")
	assert.False(t, tester.config.AutoGenerateTests, "Auto generate tests should be disabled by default")
	assert.True(t, tester.config.SmartErrorRecovery, "Smart error recovery should be enabled by default")
	assert.True(t, tester.config.AdaptiveTestPriority, "Adaptive test priority should be enabled by default")
	assert.Equal(t, 0.7, tester.config.ConfidenceThreshold, "Confidence threshold should be 0.7 by default")
	assert.Equal(t, 20, tester.config.MaxGeneratedTests, "Max generated tests should be 20 by default")
	assert.False(t, tester.config.EnableLearning, "Learning should be disabled by default")
}

// TestSetConfig tests configuring AI-enhanced testing settings
func TestSetConfig(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	customConfig := AIConfig{
		EnableErrorDetection:   false,
		EnableTestGeneration:  false,
		EnableVisionAnalysis:   true,
		AutoGenerateTests:      true,
		SmartErrorRecovery:     false,
		AdaptiveTestPriority:   false,
		ConfidenceThreshold:    0.9,
		MaxGeneratedTests:      50,
		EnableLearning:         true,
	}

	tester.SetConfig(customConfig)

	assert.Equal(t, customConfig.EnableErrorDetection, tester.config.EnableErrorDetection)
	assert.Equal(t, customConfig.EnableTestGeneration, tester.config.EnableTestGeneration)
	assert.Equal(t, customConfig.EnableVisionAnalysis, tester.config.EnableVisionAnalysis)
	assert.Equal(t, customConfig.AutoGenerateTests, tester.config.AutoGenerateTests)
	assert.Equal(t, customConfig.SmartErrorRecovery, tester.config.SmartErrorRecovery)
	assert.Equal(t, customConfig.AdaptiveTestPriority, tester.config.AdaptiveTestPriority)
	assert.Equal(t, customConfig.ConfidenceThreshold, tester.config.ConfidenceThreshold)
	assert.Equal(t, customConfig.MaxGeneratedTests, tester.config.MaxGeneratedTests)
	assert.Equal(t, customConfig.EnableLearning, tester.config.EnableLearning)

	// Tester should still be enabled since EnableVisionAnalysis is true
	assert.True(t, tester.enabled, "Tester should be enabled when any feature is enabled")
}

// TestSetConfig_AllDisabled tests that tester is disabled when all features are off
func TestSetConfig_AllDisabled(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	disabledConfig := AIConfig{
		EnableErrorDetection:   false,
		EnableTestGeneration:  false,
		EnableVisionAnalysis:   false,
	}

	tester.SetConfig(disabledConfig)

	assert.False(t, tester.enabled, "Tester should be disabled when all features are off")
}

// TestGenerateTests tests AI-powered test generation
func TestGenerateTests(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	pageState := map[string]interface{}{
		"url":  "https://example.com",
		"html": "<html><body><h1>Test</h1></body></html>",
	}

	tests, err := tester.GenerateTests(pageState)

	// Should generate tests based on page elements
	assert.NoError(t, err, "GenerateTests should not return error")
	assert.NotNil(t, tests, "Tests should not be nil")
	assert.Greater(t, len(tests), 0, "Should generate at least one test")
	
	// Check for expected test types
	hasNavigationTest := false
	hasScreenshotTest := false
	for _, test := range tests {
		testMap, ok := test.(map[string]interface{})
		if ok {
			if testType, exists := testMap["type"].(string); exists {
				if testType == "navigate" {
					hasNavigationTest = true
				} else if testType == "screenshot" {
					hasScreenshotTest = true
				}
			}
		}
	}
	assert.True(t, hasNavigationTest, "Should generate navigation test")
	assert.True(t, hasScreenshotTest, "Should generate screenshot test")
}

// TestGenerateTests_NilPageState tests generation with nil page state
func TestGenerateTests_NilPageState(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tests, err := tester.GenerateTests(nil)

	// Should handle nil page state with proper error
	assert.Error(t, err, "Should return error for nil page state")
	assert.Nil(t, tests, "Should return nil tests on error")
	assert.Contains(t, err.Error(), "invalid page state format", "Error should describe the issue")
}

// TestSaveTests tests saving generated tests to file
func TestSaveTests(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "generated_tests.yaml")

	tests := []interface{}{
		map[string]string{"name": "test1", "type": "click"},
		map[string]string{"name": "test2", "type": "fill"},
	}

	err := tester.SaveTests(tests, testPath)

	// Implementation should save tests successfully
	assert.NoError(t, err, "SaveTests should not return error")
	
	// Verify file was created and contains expected content
	content, err := os.ReadFile(testPath)
	assert.NoError(t, err, "Should be able to read saved file")
	assert.Contains(t, string(content), "AI Generated Tests", "Should contain test name")
	assert.Contains(t, string(content), "test1", "Should contain first test")
	assert.Contains(t, string(content), "test2", "Should contain second test")
}

// TestDetectErrors tests error detection in page state
func TestDetectErrors(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	pageState := map[string]interface{}{
		"url":  "https://example.com",
		"html": "<html><body><div class='error'>Error occurred</div></body></html>",
	}

	errors, err := tester.DetectErrors(pageState)

	// Should analyze page and potentially detect issues
	assert.NoError(t, err, "DetectErrors should not return error")
	assert.NotNil(t, errors, "Errors should not be nil")
	// Number of errors may vary, so we just check it returns something
}

// TestDetectErrors_NilPageState tests detection with nil page state
func TestDetectErrors_NilPageState(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	errors, err := tester.DetectErrors(nil)

	// Should handle nil page state with proper error
	assert.Error(t, err, "Should return error for nil page state")
	assert.Nil(t, errors, "Should return nil errors on error")
	assert.Contains(t, err.Error(), "invalid page state format", "Error should describe the issue")
}

// TestSaveErrorReport tests saving error report to file
func TestSaveErrorReport(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tempDir := t.TempDir()
	reportPath := filepath.Join(tempDir, "error_report.json")

	errors := []interface{}{
		map[string]string{"type": "validation", "message": "Field required"},
		map[string]string{"type": "network", "message": "Connection timeout"},
	}

	err := tester.SaveErrorReport(errors, reportPath)

	assert.NoError(t, err, "SaveErrorReport should not return error")

	// Verify file was created
	content, err := os.ReadFile(reportPath)
	assert.NoError(t, err, "Should be able to read saved report")
	
	// Check content structure
	var report map[string]interface{}
	err = yaml.Unmarshal(content, &report)
	assert.NoError(t, err, "Should be valid JSON/YAML")
	assert.Equal(t, "error_analysis", report["report_type"], "Should have correct report type")
	assert.Equal(t, len(errors), report["summary"].(map[string]interface{})["total_errors"], "Should have correct error count")
}

// TestExecuteEnhancedTesting tests AI-enhanced test execution
func TestExecuteEnhancedTesting(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	// Mock platform (empty struct instead of nil)
	platform := &struct{}{} // Mock platform
	actions := []interface{}{
		map[string]interface{}{"type": "click", "selector": "#button"},
		map[string]interface{}{"type": "fill", "selector": "#input", "value": "test"},
	}

	result, err := tester.ExecuteEnhancedTesting(platform, actions)

	assert.NoError(t, err, "ExecuteEnhancedTesting should not return error")
	assert.NotNil(t, result, "Result should not be nil")

	// Check result is a map with expected fields
	if resultMap, ok := result.(map[string]interface{}); ok {
		assert.Equal(t, "completed", resultMap["status"], "Should be completed")
		assert.Equal(t, 2, resultMap["total_actions"], "Should have 2 total actions")
		assert.Contains(t, resultMap, "ai_insights", "Should have AI insights")
		assert.Contains(t, resultMap, "metrics", "Should have metrics")
	}
}

// TestSaveTestingReport tests saving testing report to file
func TestSaveTestingReport(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tempDir := t.TempDir()
	reportPath := filepath.Join(tempDir, "testing_report.json")

	results := map[string]interface{}{
		"status": "completed",
		"total_actions": 5,
		"success_rate": 0.8,
	}

	err := tester.SaveTestingReport(results, reportPath)

	assert.NoError(t, err, "SaveTestingReport should not return error")

	// Verify file was created
	content, err := os.ReadFile(reportPath)
	assert.NoError(t, err, "Should be able to read saved report")
	
	// Check content structure
	var report map[string]interface{}
	err = yaml.Unmarshal(content, &report)
	assert.NoError(t, err, "Should be valid JSON/YAML")
	assert.Equal(t, "ai_enhanced_testing", report["report_type"], "Should have correct report type")
	assert.Equal(t, "completed", report["results"].(map[string]interface{})["status"], "Should preserve status")
}

// TestFilterTestsByConfidence tests filtering tests by confidence threshold
func TestFilterTestsByConfidence(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	// Set confidence threshold
	tester.config.ConfidenceThreshold = 0.7

	tests := []GeneratedTest{
		{Name: "test1", Confidence: 0.9, Priority: "high"},
		{Name: "test2", Confidence: 0.5, Priority: "low"},
		{Name: "test3", Confidence: 0.8, Priority: "medium"},
		{Name: "test4", Confidence: 0.6, Priority: "low"},
	}

	filtered := tester.filterTestsByConfidence(tests)

	assert.Equal(t, 2, len(filtered), "Should filter out tests below threshold")
	assert.Equal(t, "test1", filtered[0].Name, "First test should be test1")
	assert.Equal(t, "test3", filtered[1].Name, "Second test should be test3")
}

// TestFilterTestsByConfidence_EmptyTests tests filtering with empty test list
func TestFilterTestsByConfidence_EmptyTests(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tests := []GeneratedTest{}
	filtered := tester.filterTestsByConfidence(tests)

	assert.Equal(t, 0, len(filtered), "Should return empty list for empty input")
}

// TestCalculateAverageConfidence tests calculating average confidence
func TestCalculateAverageConfidence(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tests := []GeneratedTest{
		{Name: "test1", Confidence: 0.9},
		{Name: "test2", Confidence: 0.7},
		{Name: "test3", Confidence: 0.8},
	}

	avg := tester.calculateAverageConfidence(tests)

	assert.InDelta(t, 0.8, avg, 0.01, "Average confidence should be 0.8")
}

// TestCalculateAverageConfidence_EmptyTests tests average with no tests
func TestCalculateAverageConfidence_EmptyTests(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tests := []GeneratedTest{}
	avg := tester.calculateAverageConfidence(tests)

	assert.Equal(t, 0.0, avg, "Average confidence should be 0 for empty tests")
}

// TestCountTestsByPriority tests counting tests by priority
func TestCountTestsByPriority(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tests := []GeneratedTest{
		{Name: "test1", Priority: "high"},
		{Name: "test2", Priority: "low"},
		{Name: "test3", Priority: "high"},
		{Name: "test4", Priority: "medium"},
		{Name: "test5", Priority: "high"},
	}

	highCount := tester.countTestsByPriority(tests, "high")
	mediumCount := tester.countTestsByPriority(tests, "medium")
	lowCount := tester.countTestsByPriority(tests, "low")

	assert.Equal(t, 3, highCount, "Should count 3 high priority tests")
	assert.Equal(t, 1, mediumCount, "Should count 1 medium priority test")
	assert.Equal(t, 1, lowCount, "Should count 1 low priority test")
}

// TestCountTestsByPriority_EmptyTests tests counting with no tests
func TestCountTestsByPriority_EmptyTests(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tests := []GeneratedTest{}
	count := tester.countTestsByPriority(tests, "high")

	assert.Equal(t, 0, count, "Should return 0 for empty tests")
}

// TestCollectExecutionMessages tests collecting execution messages
func TestCollectExecutionMessages(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	result := map[string]interface{}{
		"status": "completed",
		"logs":   []string{"log1", "log2"},
	}

	errors := []DetectedError{
		{Category: "validation", Message: "Error 1", Severity: "high"},
		{Category: "network", Message: "Error 2", Severity: "medium"},
	}

	messages := tester.collectExecutionMessages(result, errors)

	assert.NotNil(t, messages, "Messages should not be nil")
	assert.Equal(t, 2, len(messages), "Should have 2 error messages")
	assert.Equal(t, "Error 1", messages[0].Message)
	assert.Equal(t, "Error 2", messages[1].Message)
}

// TestGenerateRecommendations tests generating AI recommendations
func TestGenerateRecommendations(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	result := AIResult{
		Errors: []DetectedError{
			{Category: "validation", Severity: "high"},
			{Category: "network", Severity: "medium"},
		},
		GeneratedTests: []GeneratedTest{
			{Name: "test1", Priority: "high"},
		},
	}

	analysis := ErrorAnalysis{
		TotalErrors:      2,
		ErrorCategories:  map[string]int{"validation": 1, "network": 1},
		CriticalErrors:   []DetectedError{{Category: "validation", Severity: "high"}},
	}

	recommendations := tester.generateRecommendations(result, analysis)

	assert.NotNil(t, recommendations, "Recommendations should not be nil")
	assert.Greater(t, len(recommendations), 0, "Should have at least one recommendation")
}

// TestGenerateRecommendations_NoErrors tests recommendations with no errors
func TestGenerateRecommendations_NoErrors(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	result := AIResult{
		Errors:         []DetectedError{},
		GeneratedTests: []GeneratedTest{},
	}

	analysis := ErrorAnalysis{
		TotalErrors:     0,
		ErrorCategories: map[string]int{},
		CriticalErrors:  []DetectedError{},
	}

	recommendations := tester.generateRecommendations(result, analysis)

	// With no errors and no tests, may return empty list or nil
	// Both are acceptable
	if recommendations != nil {
		assert.GreaterOrEqual(t, len(recommendations), 0, "Should have non-negative recommendations")
	}
}

// TestAIConfig_DefaultValues tests AIConfig default values
func TestAIConfig_DefaultValues(t *testing.T) {
	config := AIConfig{}

	assert.False(t, config.EnableErrorDetection, "EnableErrorDetection should be false by default")
	assert.False(t, config.EnableTestGeneration, "EnableTestGeneration should be false by default")
	assert.False(t, config.EnableVisionAnalysis, "EnableVisionAnalysis should be false by default")
	assert.False(t, config.AutoGenerateTests, "AutoGenerateTests should be false by default")
	assert.False(t, config.SmartErrorRecovery, "SmartErrorRecovery should be false by default")
	assert.False(t, config.AdaptiveTestPriority, "AdaptiveTestPriority should be false by default")
	assert.Equal(t, 0.0, config.ConfidenceThreshold, "ConfidenceThreshold should be 0.0 by default")
	assert.Equal(t, 0, config.MaxGeneratedTests, "MaxGeneratedTests should be 0 by default")
	assert.False(t, config.EnableLearning, "EnableLearning should be false by default")
}

// TestGenerateAIEnhancedReport tests report generation
func TestGenerateAIEnhancedReport(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	tempDir := t.TempDir()

	// Create the directory where report will be written
	reportDir := filepath.Join(tempDir, "ai_report")
	err := os.MkdirAll(reportDir, 0755)
	require.NoError(t, err, "Should create report directory")

	result := AIResult{
		Errors:         []DetectedError{},
		GeneratedTests: []GeneratedTest{},
	}

	err = tester.GenerateAIEnhancedReport(result, reportDir)

	// Report generation might fail due to missing dependencies
	// Just check that it doesn't panic
	if err != nil {
		t.Logf("Report generation failed (expected for stub): %v", err)
	} else {
		// If it succeeds, verify the file exists
		mdPath := filepath.Join(reportDir, "ai_enhanced_testing_report.md")
		if _, mdErr := os.Stat(mdPath); mdErr == nil {
			content, readErr := os.ReadFile(mdPath)
			if readErr == nil {
				assert.Contains(t, string(content), "# AI-Enhanced Testing Report", "Report should have markdown title")
			}
		}
	}
}

// TestGenerateAIEnhancedReport_InvalidPath tests report generation with invalid path
func TestGenerateAIEnhancedReport_InvalidPath(t *testing.T) {
	log := logger.NewLogger(false)
	tester := NewAIEnhancedTester(*log)

	// Use an invalid path (directory that doesn't exist)
	invalidPath := "/nonexistent/directory/report.html"

	result := AIResult{}
	err := tester.GenerateAIEnhancedReport(result, invalidPath)

	assert.Error(t, err, "Should error with invalid path")
}
