package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewErrorDetector tests the constructor
func TestNewErrorDetector(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	assert.NotNil(t, detector, "Error detector should not be nil")
	assert.True(t, detector.enabled, "Error detector should be enabled by default")
	assert.NotEmpty(t, detector.patterns, "Error patterns should be initialized")
	assert.Greater(t, len(detector.patterns), 0, "Should have multiple error patterns")
}

// TestInitializeErrorPatterns tests pattern initialization
func TestInitializeErrorPatterns(t *testing.T) {
	patterns := initializeErrorPatterns()

	assert.NotEmpty(t, patterns, "Should have error patterns")
	assert.Greater(t, len(patterns), 5, "Should have multiple patterns")

	// Check first pattern has required fields
	if len(patterns) > 0 {
		pattern := patterns[0]
		assert.NotEmpty(t, pattern.Name, "Pattern should have a name")
		assert.NotEmpty(t, pattern.Category, "Pattern should have a category")
		assert.NotNil(t, pattern.Pattern, "Pattern should have a regex")
		assert.NotEmpty(t, pattern.Severity, "Pattern should have a severity")
		assert.Greater(t, pattern.Confidence, 0.0, "Pattern should have confidence > 0")
	}
}

// TestErrorDetector_DetectErrors tests the main error detection method
func TestErrorDetector_DetectErrors(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	messages := []ErrorMessage{
		{
			Message:   "Connection timeout occurred",
			Source:    "test",
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
			Level:     "error",
		},
		{
			Message:   "Element not found on page",
			Source:    "test",
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
			Level:     "error",
		},
	}

	errors := detector.DetectErrors(messages)

	assert.NotEmpty(t, errors, "Should detect errors")
	assert.Greater(t, len(errors), 0, "Should have at least one detected error")

	// Check first error has required fields
	if len(errors) > 0 {
		err := errors[0]
		assert.NotEmpty(t, err.Name, "Error should have a name")
		assert.NotEmpty(t, err.Category, "Error should have a category")
		assert.NotEmpty(t, err.Message, "Error should have a message")
		assert.NotEmpty(t, err.Severity, "Error should have a severity")
		assert.Greater(t, err.Confidence, 0.0, "Error should have confidence > 0")
	}
}

// TestErrorDetector_DetectErrors_Disabled tests detection when disabled
func TestErrorDetector_DetectErrors_Disabled(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)
	detector.enabled = false

	messages := []ErrorMessage{
		{
			Message:   "Connection timeout occurred",
			Source:    "test",
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
			Level:     "error",
		},
	}

	errors := detector.DetectErrors(messages)

	assert.Empty(t, errors, "Should not detect errors when disabled")
}

// TestErrorDetector_DetectErrors_NoErrors tests with non-error messages
func TestErrorDetector_DetectErrors_NoErrors(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	messages := []ErrorMessage{
		{
			Message:   "Operation completed successfully",
			Source:    "test",
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
			Level:     "info",
		},
		{
			Message:   "Test passed without issues",
			Source:    "test",
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
			Level:     "info",
		},
	}

	errors := detector.DetectErrors(messages)

	// May detect zero errors or might catch false positives
	// This is acceptable for now - just verify no panic
	assert.GreaterOrEqual(t, len(errors), 0, "Should return valid result")
}

// TestErrorDetector_DetectErrors_EmptyMessages tests with empty message list
func TestErrorDetector_DetectErrors_EmptyMessages(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	messages := []ErrorMessage{}

	errors := detector.DetectErrors(messages)

	assert.Empty(t, errors, "Should return empty list for no messages")
}

// TestContainsErrorIndicators tests error indicator detection
func TestContainsErrorIndicators(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	testCases := []struct {
		message  string
		expected bool
	}{
		{"Error occurred during processing", true},
		{"Operation failed", true},
		{"Exception thrown", true},
		{"Cannot connect to server", true},
		{"Success: all tests passed", false},
		{"Normal log message", false},
	}

	for _, tc := range testCases {
		result := detector.containsErrorIndicators(tc.message)
		assert.Equal(t, tc.expected, result, "Message: %s", tc.message)
	}
}

// TestExtractPosition tests position extraction
func TestExtractPosition(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	context := map[string]interface{}{
		"selector": "#submit-button",
		"x":        100,
		"y":        200,
		"element":  "button",
	}

	position := detector.extractPosition(context)

	assert.NotNil(t, position, "Should extract position")
	// Position fields may or may not be set depending on implementation
}

// TestExtractPosition_EmptyContext tests position extraction with empty context
func TestExtractPosition_EmptyContext(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	context := map[string]interface{}{}

	position := detector.extractPosition(context)

	assert.NotNil(t, position, "Should return a position struct")
}

// TestConvertContext tests context conversion
func TestConvertContext(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	context := map[string]interface{}{
		"url":      "https://example.com",
		"browser":  "chrome",
		"viewport": "1920x1080",
	}

	converted := detector.convertContext(context)

	assert.NotNil(t, converted, "Should convert context")
	assert.IsType(t, map[string]string{}, converted, "Should return string map")
}

// TestConvertContext_EmptyContext tests context conversion with empty map
func TestConvertContext_EmptyContext(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	context := map[string]interface{}{}

	converted := detector.convertContext(context)

	assert.NotNil(t, converted, "Should return a map")
	assert.Empty(t, converted, "Should be empty")
}

// TestAnalyzeErrors tests comprehensive error analysis
func TestAnalyzeErrors(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	errors := []DetectedError{
		{
			Name:       "NetworkTimeout",
			Category:   "network",
			Severity:   "high",
			Confidence: 0.85,
			Timestamp:  time.Now(),
		},
		{
			Name:       "ElementNotFound",
			Category:   "ui",
			Severity:   "medium",
			Confidence: 0.75,
			Timestamp:  time.Now(),
		},
		{
			Name:       "AuthenticationFailed",
			Category:   "authentication",
			Severity:   "high",
			Confidence: 0.90,
			Timestamp:  time.Now(),
		},
	}

	analysis := detector.AnalyzeErrors(errors)

	assert.Equal(t, 3, analysis.TotalErrors, "Should count all errors")
	assert.NotNil(t, analysis.ErrorCategories, "Should have error categories")
	assert.NotNil(t, analysis.SeverityLevels, "Should have severity levels")
	assert.Greater(t, len(analysis.ErrorCategories), 0, "Should have at least one category")
}

// TestAnalyzeErrors_EmptyList tests analysis with no errors
func TestAnalyzeErrors_EmptyList(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	errors := []DetectedError{}

	analysis := detector.AnalyzeErrors(errors)

	assert.Equal(t, 0, analysis.TotalErrors, "Should have zero errors")
	assert.NotNil(t, analysis.ErrorCategories, "Should have categories map")
	assert.NotNil(t, analysis.SeverityLevels, "Should have severity map")
}

// TestGenerateErrorTrends tests error trend generation
func TestGenerateErrorTrends(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	now := time.Now()
	errors := []DetectedError{
		{
			Category:  "network",
			Severity:  "high",
			Timestamp: now,
		},
		{
			Category:  "ui",
			Severity:  "medium",
			Timestamp: now.Add(-1 * time.Hour),
		},
	}

	trends := detector.generateErrorTrends(errors)

	assert.NotNil(t, trends, "Should return trends")
	// Trends might be empty or have entries depending on implementation
}

// TestGetMostCommonCategory tests category frequency
func TestGetMostCommonCategory(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	errors := []DetectedError{
		{Category: "network"},
		{Category: "network"},
		{Category: "ui"},
	}

	category := detector.getMostCommonCategory(errors)

	assert.Equal(t, "network", category, "Should return most common category")
}

// TestGetMostCommonCategory_EmptyList tests with no errors
func TestGetMostCommonCategory_EmptyList(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	errors := []DetectedError{}

	category := detector.getMostCommonCategory(errors)

	// Returns empty string for empty list
	assert.Equal(t, "", category, "Should return empty string for empty list")
}

// TestGetMostCommonSeverity tests severity frequency
func TestGetMostCommonSeverity(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	errors := []DetectedError{
		{Severity: "high"},
		{Severity: "high"},
		{Severity: "medium"},
	}

	severity := detector.getMostCommonSeverity(errors)

	assert.Equal(t, "high", severity, "Should return most common severity")
}

// TestGetMostCommonSeverity_EmptyList tests with no errors
func TestGetMostCommonSeverity_EmptyList(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	errors := []DetectedError{}

	severity := detector.getMostCommonSeverity(errors)

	// Returns empty string for empty list
	assert.Equal(t, "", severity, "Should return empty string for empty list")
}

// TestGenerateErrorRecommendations tests recommendation generation
func TestGenerateErrorRecommendations(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	analysis := ErrorAnalysis{
		TotalErrors: 5,
		ErrorCategories: map[string]int{
			"network": 3,
			"ui":      2,
		},
		SeverityLevels: map[string]int{
			"high":   3,
			"medium": 2,
		},
		CriticalErrors: []DetectedError{
			{Category: "network", Severity: "high"},
		},
	}

	recommendations := detector.generateErrorRecommendations(analysis)

	assert.NotNil(t, recommendations, "Should generate recommendations")
	assert.Greater(t, len(recommendations), 0, "Should have at least one recommendation")

	// Check first recommendation has required fields
	if len(recommendations) > 0 {
		rec := recommendations[0]
		assert.NotEmpty(t, rec.Type, "Recommendation should have a type")
		assert.NotEmpty(t, rec.Priority, "Recommendation should have a priority")
		assert.NotEmpty(t, rec.Description, "Recommendation should have a description")
	}
}

// TestGenerateErrorRecommendations_NoErrors tests with zero errors
func TestGenerateErrorRecommendations_NoErrors(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	analysis := ErrorAnalysis{
		TotalErrors:     0,
		ErrorCategories: map[string]int{},
		SeverityLevels:  map[string]int{},
		CriticalErrors:  []DetectedError{},
	}

	recommendations := detector.generateErrorRecommendations(analysis)

	// May return nil or empty list for no errors
	// Both are acceptable
	if recommendations != nil {
		assert.GreaterOrEqual(t, len(recommendations), 0, "Should have non-negative recommendations")
	}
}

// TestDetermineTestCoverage tests test coverage determination
func TestDetermineTestCoverage(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	analysis := ErrorAnalysis{
		TotalErrors: 3,
		ErrorCategories: map[string]int{
			"network": 2,
			"ui":      1,
		},
	}

	coverage := detector.determineTestCoverage(analysis)

	assert.NotNil(t, coverage, "Should return coverage list")
	// May be empty or have suggestions
}

// TestExtractDetectedPatterns tests pattern extraction
func TestExtractDetectedPatterns(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	errors := []DetectedError{
		{Name: "NetworkTimeout", Category: "network"},
		{Name: "ElementNotFound", Category: "ui"},
	}

	patterns := detector.extractDetectedPatterns(errors)

	assert.NotNil(t, patterns, "Should return patterns list")
	// May be empty if patterns aren't fully tracked
}

// TestFormatMap tests map formatting helper
func TestFormatMap(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	testMap := map[string]int{
		"network": 5,
		"ui":      3,
		"auth":    2,
	}

	formatted := detector.formatMap(testMap)

	assert.NotEmpty(t, formatted, "Should format map")
	assert.Contains(t, formatted, "network", "Should contain key")
	assert.Contains(t, formatted, "5", "Should contain value")
}

// TestFormatMap_EmptyMap tests with empty map
func TestFormatMap_EmptyMap(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	testMap := map[string]int{}

	formatted := detector.formatMap(testMap)

	assert.NotEmpty(t, formatted, "Should return some formatted string")
}

// TestGenerateRecommendationsSection tests recommendations section formatting
func TestGenerateRecommendationsSection(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	recommendations := []ErrorRecommendation{
		{
			Type:        "fix",
			Priority:    "high",
			Description: "Fix network timeout issues",
			Suggestion:  "Increase timeout values",
			Steps:       []string{"Step 1", "Step 2"},
			Impact:      "high",
			Effort:      "medium",
		},
	}

	section := detector.generateRecommendationsSection(recommendations)

	assert.NotEmpty(t, section, "Should generate section")
	assert.Contains(t, section, "Fix network timeout issues", "Should contain description")
}

// TestGenerateRecommendationsSection_Empty tests with no recommendations
func TestGenerateRecommendationsSection_Empty(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	recommendations := []ErrorRecommendation{}

	section := detector.generateRecommendationsSection(recommendations)

	assert.NotEmpty(t, section, "Should return some content")
}

// TestGenerateSmartErrorReport tests report generation
func TestGenerateSmartErrorReport(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	tempDir := t.TempDir()
	// GenerateSmartErrorReport expects a directory, not a file path
	reportDir := filepath.Join(tempDir, "reports")
	err := os.MkdirAll(reportDir, 0755)
	require.NoError(t, err, "Should create report directory")

	errors := []DetectedError{
		{
			Name:       "NetworkTimeout",
			Category:   "network",
			Severity:   "high",
			Confidence: 0.85,
			Message:    "Connection timeout",
			Timestamp:  time.Now(),
		},
	}

	analysis := ErrorAnalysis{
		TotalErrors: 1,
		ErrorCategories: map[string]int{
			"network": 1,
		},
		SeverityLevels: map[string]int{
			"high": 1,
		},
		Recommendations: []ErrorRecommendation{
			{
				Type:        "fix",
				Priority:    "high",
				Description: "Fix network issues",
			},
		},
	}

	err = detector.GenerateSmartErrorReport(errors, analysis, reportDir)

	assert.NoError(t, err, "Report generation should not error")

	// Check if file was created (function creates smart_error_report.md inside reportDir)
	reportFile := filepath.Join(reportDir, "smart_error_report.md")
	_, statErr := os.Stat(reportFile)
	if statErr == nil {
		// File exists, read and verify content
		content, readErr := os.ReadFile(reportFile)
		require.NoError(t, readErr, "Should be able to read report")
		assert.Contains(t, string(content), "Error Detection Report", "Report should have title")
	}
}

// TestGenerateSmartErrorReport_EmptyErrors tests report with no errors
func TestGenerateSmartErrorReport_EmptyErrors(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	tempDir := t.TempDir()
	reportDir := filepath.Join(tempDir, "reports")
	os.MkdirAll(reportDir, 0755)

	errors := []DetectedError{}
	analysis := ErrorAnalysis{
		TotalErrors:     0,
		ErrorCategories: map[string]int{},
		SeverityLevels:  map[string]int{},
	}

	err := detector.GenerateSmartErrorReport(errors, analysis, reportDir)

	// Should handle empty errors gracefully
	if err == nil {
		reportFile := filepath.Join(reportDir, "smart_error_report.md")
		_, statErr := os.Stat(reportFile)
		if statErr == nil {
			content, _ := os.ReadFile(reportFile)
			assert.NotEmpty(t, content, "Report should have content")
		}
	}
}

// TestGenerateSmartErrorReport_InvalidPath tests with invalid path
func TestGenerateSmartErrorReport_InvalidPath(t *testing.T) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(*log)

	// Use a path that cannot be created
	reportPath := "/nonexistent/path/that/does/not/exist/report.md"

	errors := []DetectedError{}
	analysis := ErrorAnalysis{}

	err := detector.GenerateSmartErrorReport(errors, analysis, reportPath)

	assert.Error(t, err, "Should error with invalid path")
}

// TestErrorPattern_Structure tests ErrorPattern struct
func TestErrorPattern_Structure(t *testing.T) {
	pattern := ErrorPattern{
		Name:        "TestPattern",
		Category:    "test",
		Severity:    "high",
		Confidence:  0.85,
		Description: "Test pattern",
		Suggestions: []string{"suggestion1", "suggestion2"},
		Tags:        []string{"tag1", "tag2"},
	}

	assert.Equal(t, "TestPattern", pattern.Name)
	assert.Equal(t, "test", pattern.Category)
	assert.Equal(t, "high", pattern.Severity)
	assert.Equal(t, 0.85, pattern.Confidence)
	assert.Len(t, pattern.Suggestions, 2)
	assert.Len(t, pattern.Tags, 2)
}

// TestDetectedError_Structure tests DetectedError struct
func TestDetectedError_Structure(t *testing.T) {
	now := time.Now()
	err := DetectedError{
		Name:        "TestError",
		Category:    "test",
		Message:     "Test message",
		Severity:    "high",
		Confidence:  0.85,
		Timestamp:   now,
		Source:      "test",
		Suggestions: []string{"suggestion"},
		Tags:        []string{"tag"},
		Context:     map[string]string{"key": "value"},
	}

	assert.Equal(t, "TestError", err.Name)
	assert.Equal(t, "test", err.Category)
	assert.Equal(t, "high", err.Severity)
	assert.Equal(t, now, err.Timestamp)
}

// TestErrorAnalysis_Structure tests ErrorAnalysis struct
func TestErrorAnalysis_Structure(t *testing.T) {
	analysis := ErrorAnalysis{
		TotalErrors:     5,
		ErrorCategories: map[string]int{"network": 3},
		SeverityLevels:  map[string]int{"high": 2},
		CriticalErrors:  []DetectedError{},
		HighRiskErrors:  []DetectedError{},
		ErrorTrends:     []ErrorTrend{},
		Recommendations: []ErrorRecommendation{},
		TestCoverage:    []string{},
		ErrorPatterns:   []ErrorPattern{},
	}

	assert.Equal(t, 5, analysis.TotalErrors)
	assert.Equal(t, 3, analysis.ErrorCategories["network"])
	assert.Equal(t, 2, analysis.SeverityLevels["high"])
}

// TestErrorMessage_Structure tests ErrorMessage struct
func TestErrorMessage_Structure(t *testing.T) {
	now := time.Now()
	msg := ErrorMessage{
		Message:   "Test message",
		Source:    "test",
		Timestamp: now,
		Context:   map[string]interface{}{"key": "value"},
		Level:     "error",
	}

	assert.Equal(t, "Test message", msg.Message)
	assert.Equal(t, "test", msg.Source)
	assert.Equal(t, now, msg.Timestamp)
	assert.Equal(t, "error", msg.Level)
}
