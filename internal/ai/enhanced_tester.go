package ai

import (
	"fmt"
	"os"
	"strings"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"
	"panoptic/internal/platforms"
	"panoptic/internal/vision"
)

// AIEnhancedTester provides AI-enhanced testing capabilities
type AIEnhancedTester struct {
	Logger        logger.Logger
	ErrorDetector *ErrorDetector
	TestGenerator *TestGenerator
	VisionDetector *vision.ElementDetector
	enabled       bool
	config        AIConfig
}

// AIConfig contains AI-enhanced testing configuration
type AIConfig struct {
	EnableErrorDetection   bool  `yaml:"enable_error_detection"`
	EnableTestGeneration  bool  `yaml:"enable_test_generation"`
	EnableVisionAnalysis   bool  `yaml:"enable_vision_analysis"`
	AutoGenerateTests      bool  `yaml:"auto_generate_tests"`
	SmartErrorRecovery     bool  `yaml:"smart_error_recovery"`
	AdaptiveTestPriority   bool  `yaml:"adaptive_test_priority"`
	ConfidenceThreshold    float64 `yaml:"confidence_threshold"`
	MaxGeneratedTests      int    `yaml:"max_generated_tests"`
	EnableLearning         bool   `yaml:"enable_learning"`
}

// NewAIEnhancedTester creates a new AI-enhanced tester
func NewAIEnhancedTester(log logger.Logger) *AIEnhancedTester {
	visionDetector := vision.NewElementDetector(log)
	
	return &AIEnhancedTester{
		Logger:         log,
		ErrorDetector:  NewErrorDetector(log),
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
			ConfidenceThreshold:    0.7,
			MaxGeneratedTests:      20,
			EnableLearning:         false,
		},
	}
}

// SetConfig configures AI-enhanced testing settings
func (ait *AIEnhancedTester) SetConfig(config AIConfig) {
	ait.config = config
	ait.enabled = config.EnableErrorDetection || config.EnableTestGeneration || config.EnableVisionAnalysis
}

// ExecuteWithAI executes a test configuration with AI enhancements
func (ait *AIEnhancedTester) ExecuteWithAI(testConfig config.Config, platform platforms.Platform) (AIResult, error) {
	if !ait.enabled {
		return AIResult{}, fmt.Errorf("AI-enhanced testing is disabled")
	}

	ait.Logger.Infof("Starting AI-enhanced testing for %s", testConfig.Name)
	
	result := AIResult{
		OriginalConfig: testConfig,
		StartTime:      time.Now(),
		Errors:         []DetectedError{},
		GeneratedTests: []GeneratedTest{},
		Enhancements:   []TestEnhancement{},
	}

	// Phase 1: Vision Analysis if enabled
	if ait.config.EnableVisionAnalysis {
		ait.Logger.Info("Performing vision analysis...")
		
		// Take a screenshot first for vision analysis
		screenshotPath := fmt.Sprintf("%s/vision_analysis_%d.png", testConfig.Output, time.Now().Unix())
		if err := platform.Screenshot(screenshotPath); err != nil {
			ait.Logger.Warnf("Failed to take screenshot for vision analysis: %v", err)
		} else {
			elements, err := ait.VisionDetector.DetectElements(screenshotPath)
			if err != nil {
				ait.Logger.Warnf("Vision analysis failed: %v", err)
			} else {
				result.VisualElements = elements
				ait.Logger.Infof("Detected %d visual elements", len(elements))
			}
		}
	}

	// Phase 2: AI Test Generation if enabled and elements detected
	if ait.config.EnableTestGeneration && len(result.VisualElements) > 0 {
		ait.Logger.Info("Generating AI-powered tests...")
		generatedTests, err := ait.TestGenerator.GenerateTestsFromElements(result.VisualElements, testConfig.Apps[0].Type)
		if err != nil {
			ait.Logger.Warnf("AI test generation failed: %v", err)
		} else {
			// Filter tests by confidence threshold
			filteredTests := ait.filterTestsByConfidence(generatedTests)
			
			// Limit number of generated tests
			if len(filteredTests) > ait.config.MaxGeneratedTests {
				filteredTests = filteredTests[:ait.config.MaxGeneratedTests]
			}
			
			result.GeneratedTests = filteredTests
			ait.Logger.Infof("Generated %d AI-enhanced tests", len(filteredTests))
		}
	}

	// Phase 3: Execute original and generated tests
	ait.Logger.Info("Executing test suite with AI enhancements...")
	
	// Execute original test actions
	originalResult, originalErrors := ait.executeOriginalTest(testConfig, platform)
	result.OriginalResult = originalResult
	result.Errors = append(result.Errors, originalErrors...)

	// Phase 4: Smart Error Detection
	if ait.config.EnableErrorDetection {
		ait.Logger.Info("Performing smart error detection...")
		messages := ait.collectExecutionMessages(originalResult, result.Errors)
		detectedErrors := ait.ErrorDetector.DetectErrors(messages)
		result.Errors = append(result.Errors, detectedErrors...)
		
		if len(detectedErrors) > 0 {
			ait.Logger.Infof("Detected %d additional errors through AI analysis", len(detectedErrors))
		}
	}

	// Phase 5: Generate Enhancements and Recommendations
	if ait.config.SmartErrorRecovery && len(result.Errors) > 0 {
		enhancements := ait.generateErrorRecoveryEnhancements(result.Errors)
		result.Enhancements = append(result.Enhancements, enhancements...)
	}

	// Phase 6: Adaptive Priority Adjustment
	if ait.config.AdaptiveTestPriority {
		ait.adjustTestPriorities(&result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	
	ait.Logger.Infof("AI-enhanced testing completed in %v", result.Duration)
	
	return result, nil
}

// AIResult contains the results of AI-enhanced testing
type AIResult struct {
	OriginalConfig  config.Config            `json:"original_config"`
	StartTime       time.Time                 `json:"start_time"`
	EndTime         time.Time                 `json:"end_time"`
	Duration        time.Duration             `json:"duration"`
	VisualElements  []vision.ElementInfo      `json:"visual_elements"`
	GeneratedTests  []GeneratedTest           `json:"generated_tests"`
	OriginalResult  interface{}               `json:"original_result"`
	Errors          []DetectedError           `json:"errors"`
	Enhancements    []TestEnhancement         `json:"enhancements"`
	Recommendations []AIRecommendation        `json:"recommendations"`
}

// TestEnhancement represents an AI-generated test enhancement
type TestEnhancement struct {
	Type         string            `json:"type"`         // recovery, optimization, addition
	Description  string            `json:"description"`
	OriginalTest string            `json:"original_test"`
	EnhancedTest string            `json:"enhanced_test"`
	Reasoning    string            `json:"reasoning"`
	Confidence   float64           `json:"confidence"`
	Impact       string            `json:"impact"`
	Parameters   map[string]string `json:"parameters"`
}

// AIRecommendation represents an AI-generated recommendation
type AIRecommendation struct {
	Category    string   `json:"category"`    // test, coverage, performance, security
	Priority    string   `json:"priority"`    // high, medium, low
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ActionItems []string `json:"action_items"`
	Benefit     string   `json:"benefit"`
	Effort      string   `json:"effort"`
}

// executeOriginalTest executes the original test configuration
func (ait *AIEnhancedTester) executeOriginalTest(testConfig config.Config, platform platforms.Platform) (interface{}, []DetectedError) {
	// This would integrate with the existing executor
	// For now, simulate execution and collect results
	ait.Logger.Infof("Executing %d original test actions", len(testConfig.Actions))
	
	var errors []DetectedError
	
	// Simulate executing each action and collecting errors
	for _, action := range testConfig.Actions {
		ait.Logger.Debugf("Executing action: %s", action.Name)
		
		// In real implementation, this would call the actual executor
		// For now, we simulate and collect hypothetical errors
		
		// Add some example error detection based on action type
		if action.Type == "click" && action.Selector == "" {
			errors = append(errors, DetectedError{
				Name:        "MissingSelector",
				Category:    "ui",
				Message:     fmt.Sprintf("Click action '%s' missing selector", action.Name),
				Severity:    "medium",
				Confidence:  0.9,
				Timestamp:   time.Now(),
				Source:      "ai_enhanced_tester",
				Position:    ErrorPosition{Type: "action", Value: action.Name},
				Suggestions: []string{"Add CSS selector", "Use XPath selector", "Specify element ID"},
				Tags:        []string{"click", "selector", "missing"},
			})
		}
		
		if action.Type == "navigate" && action.Value == "" {
			errors = append(errors, DetectedError{
				Name:        "MissingURL",
				Category:    "navigation",
				Message:     fmt.Sprintf("Navigate action '%s' missing URL", action.Name),
				Severity:    "high",
				Confidence:  0.95,
				Timestamp:   time.Now(),
				Source:      "ai_enhanced_tester",
				Position:    ErrorPosition{Type: "action", Value: action.Name},
				Suggestions: []string{"Add target URL", "Verify URL format", "Check URL accessibility"},
				Tags:        []string{"navigate", "url", "missing"},
			})
		}
		
		// Wait between actions
		if action.Type == "wait" {
			time.Sleep(time.Duration(action.WaitTime) * time.Second)
		}
	}
	
	// Return simulated result
	result := map[string]interface{}{
		"actions_executed": len(testConfig.Actions),
		"success_rate":    0.85,
		"execution_time":   time.Now(),
	}
	
	return result, errors
}

// filterTestsByConfidence filters tests based on confidence threshold
func (ait *AIEnhancedTester) filterTestsByConfidence(tests []GeneratedTest) []GeneratedTest {
	var filtered []GeneratedTest
	
	for _, test := range tests {
		if test.Confidence >= ait.config.ConfidenceThreshold {
			filtered = append(filtered, test)
		}
	}
	
	return filtered
}

// collectExecutionMessages collects messages from test execution for error analysis
func (ait *AIEnhancedTester) collectExecutionMessages(result interface{}, errors []DetectedError) []ErrorMessage {
	var messages []ErrorMessage
	
	// Add messages from detected errors
	for _, error := range errors {
		message := ErrorMessage{
			Message:   error.Message,
			Source:    error.Source,
			Timestamp: error.Timestamp,
			Level:     error.Severity,
			Context: map[string]interface{}{
				"category": error.Category,
				"position": error.Position,
			},
		}
		messages = append(messages, message)
	}
	
	// Add simulated execution messages
	if resultMap, ok := result.(map[string]interface{}); ok {
		if successRate, ok := resultMap["success_rate"].(float64); ok && successRate < 1.0 {
			message := ErrorMessage{
				Message:   fmt.Sprintf("Test execution success rate: %.2f%%", successRate*100),
				Source:    "execution_engine",
				Timestamp: time.Now(),
				Level:     "warn",
				Context:   resultMap,
			}
			messages = append(messages, message)
		}
	}
	
	return messages
}

// generateErrorRecoveryEnhancements generates enhancements for error recovery
func (ait *AIEnhancedTester) generateErrorRecoveryEnhancements(errors []DetectedError) []TestEnhancement {
	var enhancements []TestEnhancement
	
	// Group errors by category
	errorCategories := make(map[string][]DetectedError)
	for _, error := range errors {
		errorCategories[error.Category] = append(errorCategories[error.Category], error)
	}
	
	// Generate enhancements for each error category
	for category, categoryErrors := range errorCategories {
		if len(categoryErrors) > 2 { // Only if multiple errors in same category
			enhancement := TestEnhancement{
				Type:        "recovery",
				Description: fmt.Sprintf("AI-generated recovery for %s errors", category),
				Reasoning:   fmt.Sprintf("Detected %d %s-related errors, implementing recovery strategy", len(categoryErrors), category),
				Confidence:  0.75,
				Impact:      "high",
				Parameters:  make(map[string]string),
			}
			
			switch category {
			case "ui":
				enhancement.OriginalTest = "Standard UI interaction"
				enhancement.EnhancedTest = "AI-enhanced UI interaction with element detection"
				enhancement.Parameters["use_vision_detection"] = "true"
				enhancement.Parameters["retry_mechanism"] = "adaptive"
				enhancement.Parameters["element_timeout"] = "increased"
				
			case "network":
				enhancement.OriginalTest = "Direct network calls"
				enhancement.EnhancedTest = "Network calls with retry and fallback"
				enhancement.Parameters["max_retries"] = "3"
				enhancement.Parameters["retry_delay"] = "exponential"
				enhancement.Parameters["fallback_strategy"] = "cached_response"
				
			case "performance":
				enhancement.OriginalTest = "Standard timing"
				enhancement.EnhancedTest = "Adaptive timing with performance monitoring"
				enhancement.Parameters["adaptive_waits"] = "true"
				enhancement.Parameters["performance_thresholds"] = "dynamic"
				enhancement.Parameters["resource_monitoring"] = "enabled"
			}
			
			enhancements = append(enhancements, enhancement)
		}
	}
	
	return enhancements
}

// adjustTestPriorities adjusts test priorities based on AI analysis
func (ait *AIEnhancedTester) adjustTestPriorities(result *AIResult) {
	// Analyze errors to determine priority adjustments
	errorSeverity := make(map[string]int)
	for _, error := range result.Errors {
		errorSeverity[error.Severity]++
	}
	
	// Adjust generated test priorities based on error patterns
	for i := range result.GeneratedTests {
		test := &result.GeneratedTests[i]
		
		// Increase priority for tests that address error-prone areas
		if test.Type == "error_handling" && errorSeverity["high"] > 0 {
			test.Priority = "high"
			if test.Confidence + 0.1 <= 1.0 {
				test.Confidence = test.Confidence + 0.1
			} else {
				test.Confidence = 1.0
			}
		}
		
		// Adjust form tests if validation errors detected
		if test.Type == "form" && errorSeverity["validation"] > 0 {
			test.Priority = "high"
			if test.Confidence + 0.15 <= 1.0 {
				test.Confidence = test.Confidence + 0.15
			} else {
				test.Confidence = 1.0
			}
		}
		
		// Adjust performance tests if performance errors detected
		if test.Type == "performance" && errorSeverity["performance"] > 0 {
			test.Priority = "medium"
			if test.Confidence + 0.1 <= 1.0 {
				test.Confidence = test.Confidence + 0.1
			} else {
				test.Confidence = 1.0
			}
		}
	}
}

// GenerateAIEnhancedReport creates a comprehensive AI-enhanced testing report
func (ait *AIEnhancedTester) GenerateAIEnhancedReport(result AIResult, outputPath string) error {
	ait.Logger.Infof("Generating AI-enhanced testing report")
	
	// Analyze errors for comprehensive report
	errorAnalysis := ait.ErrorDetector.AnalyzeErrors(result.Errors)
	
	// Generate recommendations
	recommendations := ait.generateRecommendations(result, errorAnalysis)
	
	content := fmt.Sprintf(`# AI-Enhanced Testing Report

## Execution Summary
- **Test Name**: %s
- **Start Time**: %s
- **End Time**: %s
- **Duration**: %v
- **Visual Elements Detected**: %d
- **AI Tests Generated**: %d
- **Errors Detected**: %d
- **Enhancements Generated**: %d

## AI Analysis Results

### Vision Analysis
- **Element Types**: %s

### AI Test Generation
- **High Priority Tests**: %d
- **Medium Priority Tests**: %d
- **Low Priority Tests**: %d
- **Average Confidence**: %.2f

### Smart Error Detection
- **Total Errors**: %d
- **Critical Errors**: %d
- **High Severity Errors**: %d
- **Error Categories**: %v

## Test Enhancements

`, result.OriginalConfig.Name, 
	result.StartTime.Format(time.RFC3339),
	result.EndTime.Format(time.RFC3339),
	result.Duration,
	len(result.VisualElements),
	len(result.GeneratedTests),
	len(result.Errors),
	len(result.Enhancements),
	ait.getElementTypeCounts(result.VisualElements),
	ait.countTestsByPriority(result.GeneratedTests, "high"),
	ait.countTestsByPriority(result.GeneratedTests, "medium"),
	ait.countTestsByPriority(result.GeneratedTests, "low"),
	ait.calculateAverageConfidence(result.GeneratedTests),
	errorAnalysis.TotalErrors,
	len(errorAnalysis.CriticalErrors),
	len(errorAnalysis.HighRiskErrors),
	ait.formatErrorCategories(errorAnalysis.ErrorCategories))

	// Add enhancements details
	for i, enhancement := range result.Enhancements {
		content += fmt.Sprintf("### %d. %s Enhancement\n\n", i+1, strings.Title(enhancement.Type))
		content += fmt.Sprintf("- **Type**: %s\n", enhancement.Type)
		content += fmt.Sprintf("- **Description**: %s\n", enhancement.Description)
		content += fmt.Sprintf("- **Original Test**: %s\n", enhancement.OriginalTest)
		content += fmt.Sprintf("- **Enhanced Test**: %s\n", enhancement.EnhancedTest)
		content += fmt.Sprintf("- **Reasoning**: %s\n", enhancement.Reasoning)
		content += fmt.Sprintf("- **Confidence**: %.2f\n", enhancement.Confidence)
		content += fmt.Sprintf("- **Impact**: %s\n", enhancement.Impact)
		
		if len(enhancement.Parameters) > 0 {
			content += "- **Parameters**:\n"
			for key, value := range enhancement.Parameters {
				content += fmt.Sprintf("  - %s: %s\n", key, value)
			}
		}
		content += "\n"
	}
	
	// Add recommendations
	content += "## AI Recommendations\n\n"
	for i, rec := range recommendations {
		content += fmt.Sprintf("### %d. %s\n\n", i+1, rec.Title)
		content += fmt.Sprintf("- **Category**: %s\n", rec.Category)
		content += fmt.Sprintf("- **Priority**: %s\n", rec.Priority)
		content += fmt.Sprintf("- **Description**: %s\n", rec.Description)
		content += fmt.Sprintf("- **Benefit**: %s\n", rec.Benefit)
		content += fmt.Sprintf("- **Effort**: %s\n", rec.Effort)
		
		if len(rec.ActionItems) > 0 {
			content += "- **Action Items**:\n"
			for _, item := range rec.ActionItems {
				content += fmt.Sprintf("  1. %s\n", item)
			}
		}
		content += "\n"
	}
	
	// Write to file
	filename := fmt.Sprintf("%s/ai_enhanced_testing_report.md", outputPath)
	return os.WriteFile(filename, []byte(content), 0644)
}

// Helper methods
func (ait *AIEnhancedTester) getElementTypeCounts(elements []vision.ElementInfo) string {
	typeCounts := make(map[string]int)
	for _, elem := range elements {
		typeCounts[elem.Type]++
	}
	
	var result []string
	for elemType, count := range typeCounts {
		result = append(result, fmt.Sprintf("%s(%d)", elemType, count))
	}
	
	return fmt.Sprintf("[%s]", strings.Join(result, ", "))
}

func (ait *AIEnhancedTester) countTestsByPriority(tests []GeneratedTest, priority string) int {
	count := 0
	for _, test := range tests {
		if test.Priority == priority {
			count++
		}
	}
	return count
}

func (ait *AIEnhancedTester) calculateAverageConfidence(tests []GeneratedTest) float64 {
	if len(tests) == 0 {
		return 0
	}
	
	total := 0.0
	for _, test := range tests {
		total += test.Confidence
	}
	
	return total / float64(len(tests))
}

func (ait *AIEnhancedTester) formatErrorCategories(categories map[string]int) string {
	var result []string
	for category, count := range categories {
		result = append(result, fmt.Sprintf("%s(%d)", category, count))
	}
	return fmt.Sprintf("[%s]", strings.Join(result, ", "))
}

func (ait *AIEnhancedTester) generateRecommendations(result AIResult, analysis ErrorAnalysis) []AIRecommendation {
	var recommendations []AIRecommendation
	
	// Error-based recommendations
	if len(analysis.CriticalErrors) > 0 {
		recommendations = append(recommendations, AIRecommendation{
			Category:    "error",
			Priority:    "high",
			Title:       "Critical Error Resolution",
			Description: fmt.Sprintf("Address %d critical errors detected during testing", len(analysis.CriticalErrors)),
			ActionItems: []string{
				"Review and fix critical error sources",
				"Implement automated error detection",
				"Add error prevention measures",
				"Schedule immediate regression testing",
			},
			Benefit: "Prevents system failures and improves reliability",
			Effort:  "High - Requires immediate attention",
		})
	}
	
	// Test generation recommendations
	if len(result.GeneratedTests) > 0 {
		recommendations = append(recommendations, AIRecommendation{
			Category:    "test",
			Priority:    "medium",
			Title:       "AI-Generated Test Implementation",
			Description: fmt.Sprintf("Implement %d AI-generated tests for comprehensive coverage", len(result.GeneratedTests)),
			ActionItems: []string{
				"Review generated test cases",
				"Integrate high-priority tests first",
				"Customize test parameters",
				"Schedule automated execution",
			},
			Benefit: "Improves test coverage and detects edge cases",
			Effort:  "Medium - Can be implemented incrementally",
		})
	}
	
	// Vision analysis recommendations
	if len(result.VisualElements) > 20 {
		recommendations = append(recommendations, AIRecommendation{
			Category:    "coverage",
			Priority:    "medium",
			Title:       "Visual Element Coverage",
			Description: fmt.Sprintf("Comprehensive visual analysis detected %d UI elements", len(result.VisualElements)),
			ActionItems: []string{
				"Implement element-specific tests",
				"Add accessibility testing",
				"Create responsive design tests",
				"Implement visual regression testing",
			},
			Benefit: "Ensures comprehensive UI testing and accessibility",
			Effort:  "Medium - Requires test case development",
		})
	}
	
	// Enhancement recommendations
	if len(result.Enhancements) > 0 {
		recommendations = append(recommendations, AIRecommendation{
			Category:    "optimization",
			Priority:    "medium",
			Title:       "AI-Enhanced Test Implementation",
			Description: fmt.Sprintf("Implement %d AI-generated test enhancements", len(result.Enhancements)),
			ActionItems: []string{
				"Integrate error recovery mechanisms",
				"Implement adaptive timing strategies",
				"Add intelligent retry logic",
				"Enable smart element detection",
			},
			Benefit: "Improves test reliability and reduces false failures",
			Effort:  "Low to Medium - Depends on complexity",
		})
	}
	
	return recommendations
}