package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"gopkg.in/yaml.v3"
	
	"panoptic/internal/config"
	"panoptic/internal/logger"
	"panoptic/internal/platforms"
	"panoptic/internal/vision"
)

var (
	// Global pools for object reuse
	visionDetectorPool = sync.Pool{
		New: func() interface{} {
			return vision.NewElementDetector(logger.Logger{}) // Will be properly initialized later
		},
	}
	
	optimizedAIConfigPool = sync.Pool{
		New: func() interface{} {
			return &AIConfig{}
		},
	}
)

// OptimizedAIEnhancedTester provides a memory-efficient AI-enhanced tester
type OptimizedAIEnhancedTester struct {
	Logger         logger.Logger
	ErrorDetector  *OptimizedErrorDetector
	TestGenerator  *TestGenerator
	VisionDetector *vision.ElementDetector
	enabled        bool
	config         AIConfig
}

// NewOptimizedAIEnhancedTester creates a new optimized AI-enhanced tester
func NewOptimizedAIEnhancedTester(log logger.Logger) *OptimizedAIEnhancedTester {
	// Get vision detector from pool or create new
	visionDetector := visionDetectorPool.Get().(*vision.ElementDetector)
	// Reinitialize the detector with proper logger
	*visionDetector = *vision.NewElementDetector(log)
	
	return &OptimizedAIEnhancedTester{
		Logger:         log,
		ErrorDetector:  NewOptimizedErrorDetector(log),
		TestGenerator:  NewTestGenerator(log, visionDetector),
		VisionDetector: visionDetector,
		enabled:        true,
		config: AIConfig{
			EnableErrorDetection:   true,
			EnableTestGeneration:  true,
			EnableVisionAnalysis:   true,
			AutoGenerateTests:      false,
			SmartErrorRecovery:     true,
			AdaptiveTestPriority:   true,
			ConfidenceThreshold:    0.8,
			MaxGeneratedTests:      10,
			EnableLearning:         false,
		},
	}
}

// Release returns objects to pool for reuse
func (t *OptimizedAIEnhancedTester) Release() {
	if t.VisionDetector != nil {
		visionDetectorPool.Put(t.VisionDetector)
		t.VisionDetector = nil
	}
}

// SetConfig efficiently sets AI configuration
func (t *OptimizedAIEnhancedTester) SetConfig(config AIConfig) {
	t.config = config
}

// GetConfig returns current AI configuration
func (t *OptimizedAIEnhancedTester) GetConfig() AIConfig {
	return t.config
}

// Enable enables AI features
func (t *OptimizedAIEnhancedTester) Enable() {
	t.enabled = true
}

// Disable disables AI features
func (t *OptimizedAIEnhancedTester) Disable() {
	t.enabled = false
}

// IsEnabled returns whether AI features are enabled
func (t *OptimizedAIEnhancedTester) IsEnabled() bool {
	return t.enabled
}

// GenerateTests generates test cases from page state
func (t *OptimizedAIEnhancedTester) GenerateTests(pageState interface{}) ([]interface{}, error) {
	t.Logger.Debug("Generating AI-powered tests from page state...")

	// Convert pageState to a map for analysis
	pageData, ok := pageState.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid page state format")
	}

	// Extract actionable elements from page
	tests := make([]interface{}, 0)
	
	// Generate tests for clickable elements
	if elements, exists := pageData["elements"]; exists {
		if elemList, ok := elements.([]map[string]interface{}); ok {
			for i, elem := range elemList {
				// Generate test for buttons
				if elemType, ok := elem["type"].(string); ok && elemType == "button" {
					if selector, ok := elem["selector"].(string); ok {
						test := map[string]interface{}{
							"name":        fmt.Sprintf("AI_Generated_Button_Click_%d", i+1),
							"type":        "click",
							"selector":    selector,
							"description": fmt.Sprintf("AI-generated test for clicking button: %s", selector),
							"confidence":  0.85,
							"auto_generated": true,
						}
						tests = append(tests, test)
					}
				}
				
				// Generate test for input fields
				if elemType, ok := elem["type"].(string); ok && elemType == "input" {
					if selector, ok := elem["selector"].(string); ok {
						test := map[string]interface{}{
							"name":        fmt.Sprintf("AI_Generated_Input_Fill_%d", i+1),
							"type":        "fill",
							"selector":    selector,
							"value":       "test_input_value",
							"description": fmt.Sprintf("AI-generated test for filling input: %s", selector),
							"confidence":  0.80,
							"auto_generated": true,
						}
						tests = append(tests, test)
					}
				}
			}
		}
	}
	
	// Generate navigation test if URL is available
	if url, ok := pageData["url"].(string); ok && url != "" {
		navTest := map[string]interface{}{
			"name":        "AI_Generated_Navigation_Test",
			"type":        "navigate",
			"url":         url,
			"description": "AI-generated navigation test",
			"confidence":  0.90,
			"auto_generated": true,
		}
		tests = append([]interface{}{navTest}, tests...)
	}
	
	// Generate screenshot test for documentation
	screenshotTest := map[string]interface{}{
		"name":        "AI_Generated_Screenshot_Documentation",
		"type":        "screenshot",
		"description": "AI-generated screenshot for page documentation",
		"confidence":  0.95,
		"auto_generated": true,
	}
	tests = append(tests, screenshotTest)

	t.Logger.Infof("Generated %d AI test cases", len(tests))
	return tests, nil
}

// SaveTests saves generated tests to a file
func (t *OptimizedAIEnhancedTester) SaveTests(tests []interface{}, path string) error {
	t.Logger.Debugf("Saving %d tests to %s...", len(tests), path)

	// Create a structured test configuration
	testConfig := map[string]interface{}{
		"name":        "AI Generated Tests",
		"description": "Automatically generated test cases",
		"generated_at": time.Now().Format(time.RFC3339),
		"ai_version":  "1.0.0",
		"apps": []map[string]interface{}{
			{
				"name": "AI Generated Test App",
				"type": "web",
				"url":  "https://example.com",
			},
		},
		"actions": tests,
		"settings": map[string]interface{}{
			"screenshot_format": "png",
			"video_format":     "mp4",
			"quality":          85,
			"headless":         true,
		},
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to YAML file
	data, err := yaml.Marshal(testConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal test configuration: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write tests file: %w", err)
	}

	t.Logger.Infof("Successfully saved %d tests to %s", len(tests), path)
	return nil
}

// DetectErrors detects errors in page state
func (t *OptimizedAIEnhancedTester) DetectErrors(pageState interface{}) ([]interface{}, error) {
	t.Logger.Debug("Detecting errors in page state...")

	// Convert pageState to a map for analysis
	pageData, ok := pageState.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid page state format")
	}

	errors := make([]interface{}, 0)

	// Check for common error indicators in page content
	if content, exists := pageData["content"].(string); exists {
		// Use the optimized error detector
		detectedErrors := t.ErrorDetector.DetectErrors(content)
		for _, err := range detectedErrors {
			errors = append(errors, map[string]interface{}{
				"type":        err.Name,
				"category":    err.Category,
				"severity":    err.Severity,
				"confidence":  err.Confidence,
				"description": err.Message,
				"suggestions": err.Suggestions,
				"tags":        err.Tags,
				"detected_at": err.Timestamp.Format(time.RFC3339),
			})
		}
	}

	// Check for broken links or missing resources
	if resources, exists := pageData["resources"].([]map[string]interface{}); exists {
		for _, resource := range resources {
			if status, ok := resource["status"].(float64); ok && status >= 400 {
				errors = append(errors, map[string]interface{}{
					"type":        "ResourceError",
					"category":    "network",
					"severity":    "medium",
					"confidence":  0.90,
					"description": fmt.Sprintf("Resource failed to load: %s", resource["url"]),
					"suggestions": []string{
						"Check resource URL",
						"Verify resource exists",
						"Check network connectivity",
					},
					"tags": []string{"resource", "network", "error"},
					"detected_at": time.Now().Format(time.RFC3339),
				})
			}
		}
	}

	t.Logger.Infof("Detected %d errors in page state", len(errors))
	return errors, nil
}

// SaveErrorReport saves error analysis to a file
func (t *OptimizedAIEnhancedTester) SaveErrorReport(errors []interface{}, path string) error {
	t.Logger.Debugf("Saving error report with %d errors to %s...", len(errors), path)

	// Create error report structure
	report := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"total_errors":  len(errors),
		"ai_version":   "1.0.0",
		"errors":       errors,
		"summary": map[string]interface{}{
			"by_category": t.categorizeErrors(errors),
			"by_severity": t.severitySummary(errors),
		},
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to JSON file
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal error report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write error report: %w", err)
	}

	t.Logger.Infof("Successfully saved error report with %d errors to %s", len(errors), path)
	return nil
}

// ExecuteEnhancedTesting executes comprehensive AI-enhanced testing
func (t *OptimizedAIEnhancedTester) ExecuteEnhancedTesting(platform interface{}, actions interface{}) (interface{}, error) {
	t.Logger.Info("Starting AI-enhanced testing...")

	// Convert platform to WebPlatform for testing
	webPlatform, ok := platform.(*platforms.WebPlatform)
	if !ok {
		return nil, fmt.Errorf("platform must be WebPlatform")
	}

	// Convert actions to config.Action slice
	actionList, ok := actions.([]config.Action)
	if !ok {
		return nil, fmt.Errorf("actions must be []config.Action")
	}

	results := make(map[string]interface{})
	successCount := 0
	totalCount := len(actionList)

	// Execute each action with AI analysis
	for i, action := range actionList {
		t.Logger.Debugf("Executing action %d/%d: %s", i+1, totalCount, action.Type)

		// Execute the action
		var err error
		switch action.Type {
		case "click":
			err = webPlatform.Click(action.Selector)
		case "fill":
			err = webPlatform.Fill(action.Selector, action.Value)
		case "navigate":
			err = webPlatform.Navigate(action.Value)
		case "screenshot":
			if filename, ok := action.Parameters["filename"].(string); ok {
				err = webPlatform.Screenshot(filename)
			} else {
				err = fmt.Errorf("filename parameter required for screenshot")
			}
		default:
			err = fmt.Errorf("unsupported action type: %s", action.Type)
		}

		// Analyze results with AI
		pageState, stateErr := webPlatform.GetPageState()
		if stateErr != nil {
			t.Logger.Warnf("Failed to get page state for action %d: %v", i+1, stateErr)
		}

		errors := []interface{}{}
		if pageState != nil {
			detectedErrors, detectErr := t.DetectErrors(pageState)
			if detectErr != nil {
				t.Logger.Warnf("Failed to detect errors for action %d: %v", i+1, detectErr)
			} else {
				errors = detectedErrors
			}
		}

		// Store result
		result := map[string]interface{}{
			"action_index": i + 1,
			"action_type":  action.Type,
			"success":      err == nil,
			"error":        nil,
			"detected_errors": len(errors),
			"errors":         errors,
			"timestamp":      time.Now().Format(time.RFC3339),
		}
		
		if err != nil {
			result["error"] = err.Error()
		}

		results[fmt.Sprintf("action_%d", i+1)] = result

		if err == nil {
			successCount++
		}
	}

	// Create summary
	summary := map[string]interface{}{
		"total_actions": totalCount,
		"successful":    successCount,
		"failed":        totalCount - successCount,
		"success_rate":  float64(successCount) / float64(totalCount) * 100,
		"ai_enabled":    t.enabled,
	}

	t.Logger.Infof("AI-enhanced testing completed: %d/%d actions successful", successCount, totalCount)

	return map[string]interface{}{
		"summary": summary,
		"results": results,
	}, nil
}

// SaveTestingReport saves AI-enhanced testing report to a file
func (t *OptimizedAIEnhancedTester) SaveTestingReport(results interface{}, path string) error {
	t.Logger.Debugf("Saving AI-enhanced testing report to %s...", path)

	// Create report structure
	report := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"ai_version":   "1.0.0",
		"report_type":  "ai_enhanced_testing",
		"results":      results,
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to JSON file
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal testing report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write testing report: %w", err)
	}

	t.Logger.Infof("Successfully saved AI-enhanced testing report to %s", path)
	return nil
}

// Helper methods

// categorizeErrors groups errors by category
func (t *OptimizedAIEnhancedTester) categorizeErrors(errors []interface{}) map[string]int {
	categories := make(map[string]int)
	
	for _, err := range errors {
		if errMap, ok := err.(map[string]interface{}); ok {
			if category, ok := errMap["category"].(string); ok {
				categories[category]++
			}
		}
	}
	
	return categories
}

// severitySummary groups errors by severity
func (t *OptimizedAIEnhancedTester) severitySummary(errors []interface{}) map[string]int {
	severities := make(map[string]int)
	
	for _, err := range errors {
		if errMap, ok := err.(map[string]interface{}); ok {
			if severity, ok := errMap["severity"].(string); ok {
				severities[severity]++
			}
		}
	}
	
	return severities
}