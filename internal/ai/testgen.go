package ai

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"panoptic/internal/logger"
	"panoptic/internal/vision"
)

// TestGenerator creates AI-powered test cases
type TestGenerator struct {
	logger  logger.Logger
	Vision  *vision.ElementDetector
	enabled bool
}

// NewTestGenerator creates a new AI test generator
func NewTestGenerator(log logger.Logger, visionDetector *vision.ElementDetector) *TestGenerator {
	return &TestGenerator{
		logger:  log,
		Vision:  visionDetector,
		enabled: true,
	}
}

// GeneratedTest represents an AI-generated test case
type GeneratedTest struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Steps       []TestStep  `json:"steps"`
	Priority    string      `json:"priority"`
	Confidence  float64     `json:"confidence"`
	Elements    []string     `json:"elements"`
	Duration    int          `json:"estimated_duration"`
}

// TestStep represents a single test step
type TestStep struct {
	Action     string            `json:"action"`
	Target     string            `json:"target"`
	Value      string            `json:"value"`
	Parameters map[string]string `json:"parameters"`
}

// TestAnalysis contains analysis results for test generation
type TestAnalysis struct {
	ElementTypes map[string]int `json:"element_types"`
	TotalElements int          `json:"total_elements"`
	Complexity    string      `json:"complexity"`
	TestCoverage  []string    `json:"test_coverage"`
	RiskLevel     string     `json:"risk_level"`
}

// GenerateTestsFromElements creates AI-powered tests from visual element analysis
func (tg *TestGenerator) GenerateTestsFromElements(elements []vision.ElementInfo, appType string) ([]GeneratedTest, error) {
	if !tg.enabled {
		return []GeneratedTest{}, fmt.Errorf("AI test generation is disabled")
	}

	tg.logger.Infof("Starting AI-powered test generation from %d elements", len(elements))

	// Analyze elements for test generation
	analysis := tg.AnalyzeElements(elements)
	
	// Generate different types of tests
	var tests []GeneratedTest

	// Generate basic interaction tests
	basicTests := tg.generateBasicInteractionTests(elements, analysis)
	tests = append(tests, basicTests...)

	// Generate navigation tests
	navTests := tg.generateNavigationTests(elements, analysis)
	tests = append(tests, navTests...)

	// Generate form tests
	formTests := tg.generateFormTests(elements, analysis)
	tests = append(tests, formTests...)

	// Generate error handling tests
	errorTests := tg.generateErrorHandlingTests(elements, analysis)
	tests = append(tests, errorTests...)

	// Generate accessibility tests
	accessTests := tg.generateAccessibilityTests(elements, analysis)
	tests = append(tests, accessTests...)

	// Generate performance tests
	perfTests := tg.generatePerformanceTests(elements, analysis)
	tests = append(tests, perfTests...)

	// Sort tests by priority and confidence
	tg.sortTestsByPriority(tests)

	tg.logger.Infof("Generated %d AI-powered tests", len(tests))
	return tests, nil
}

// AnalyzeElements performs analysis of elements for test generation
func (tg *TestGenerator) AnalyzeElements(elements []vision.ElementInfo) TestAnalysis {
	analysis := TestAnalysis{
		ElementTypes: make(map[string]int),
		TestCoverage: []string{},
	}

	// Count element types
	for _, elem := range elements {
		analysis.ElementTypes[elem.Type]++
		analysis.TotalElements++
	}

	// Determine complexity
	if analysis.TotalElements > 50 {
		analysis.Complexity = "high"
	} else if analysis.TotalElements > 20 {
		analysis.Complexity = "medium"
	} else {
		analysis.Complexity = "low"
	}

	// Determine test coverage areas
	if _, ok := analysis.ElementTypes["button"]; ok {
		analysis.TestCoverage = append(analysis.TestCoverage, "interaction", "navigation")
	}
	if _, ok := analysis.ElementTypes["textfield"]; ok {
		analysis.TestCoverage = append(analysis.TestCoverage, "form", "input", "validation")
	}
	if _, ok := analysis.ElementTypes["link"]; ok {
		analysis.TestCoverage = append(analysis.TestCoverage, "navigation", "accessibility")
	}
	if _, ok := analysis.ElementTypes["image"]; ok {
		analysis.TestCoverage = append(analysis.TestCoverage, "media", "accessibility")
	}

	// Determine risk level
	if analysis.Complexity == "high" {
		analysis.RiskLevel = "high"
	} else if analysis.Complexity == "medium" {
		analysis.RiskLevel = "medium"
	} else {
		analysis.RiskLevel = "low"
	}

	return analysis
}

// generateBasicInteractionTests creates basic interaction tests
func (tg *TestGenerator) generateBasicInteractionTests(elements []vision.ElementInfo, analysis TestAnalysis) []GeneratedTest {
	var tests []GeneratedTest

	// Generate button click tests
	buttons := tg.filterElementsByType(elements, "button")
	if len(buttons) > 0 {
		test := GeneratedTest{
			Name:        "Basic Button Interaction Test",
			Type:        "interaction",
			Description: "Tests basic button clicking functionality",
			Priority:    "high",
			Confidence:  0.85,
			Elements:    []string{"button"},
			Duration:    len(buttons) * 2, // 2 seconds per button
		}

		for i := range buttons {
			step := TestStep{
				Action: "vision_click",
				Target: "button",
				Parameters: map[string]string{
					"type": "button",
					"index": fmt.Sprintf("%d", i),
				},
			}
			test.Steps = append(test.Steps, step)
		}

		tests = append(tests, test)
	}

	// Generate text field input tests
	textFields := tg.filterElementsByType(elements, "textfield")
	if len(textFields) > 0 {
		test := GeneratedTest{
			Name:        "Basic Text Input Test",
			Type:        "input",
			Description: "Tests basic text field input functionality",
			Priority:    "high",
			Confidence:  0.80,
			Elements:    []string{"textfield"},
			Duration:    len(textFields) * 3, // 3 seconds per text field
		}

		for i := range textFields {
			step1 := TestStep{
				Action: "vision_click",
				Target: "textfield",
				Parameters: map[string]string{
					"type": "textfield",
					"index": fmt.Sprintf("%d", i),
				},
			}
			step2 := TestStep{
				Action: "fill",
				Target: "input",
				Value:  "test_input_" + fmt.Sprintf("%d", i),
			}
			test.Steps = append(test.Steps, step1, step2)
		}

		tests = append(tests, test)
	}

	return tests
}

// generateNavigationTests creates navigation tests
func (tg *TestGenerator) generateNavigationTests(elements []vision.ElementInfo, analysis TestAnalysis) []GeneratedTest {
	var tests []GeneratedTest

	// Generate link navigation tests
	links := tg.filterElementsByType(elements, "link")
	if len(links) > 0 {
		test := GeneratedTest{
			Name:        "Link Navigation Test",
			Type:        "navigation",
			Description: "Tests link clicking and navigation functionality",
			Priority:    "medium",
			Confidence:  0.75,
			Elements:    []string{"link"},
			Duration:    len(links) * 2, // 2 seconds per link
		}

		for i := range links {
			step := TestStep{
				Action: "vision_click",
				Target: "link",
				Parameters: map[string]string{
					"type": "link",
					"index": fmt.Sprintf("%d", i),
				},
			}
			test.Steps = append(test.Steps, step)
		}

		tests = append(tests, test)
	}

	// Generate image interaction tests
	images := tg.filterElementsByType(elements, "image")
	if len(images) > 0 {
		test := GeneratedTest{
			Name:        "Image Interaction Test",
			Type:        "navigation",
			Description: "Tests image clicking and interaction functionality",
			Priority:    "medium",
			Confidence:  0.70,
			Elements:    []string{"image"},
			Duration:    len(images) * 2, // 2 seconds per image
		}

		for i := range images {
			step := TestStep{
				Action: "vision_click",
				Target: "image",
				Parameters: map[string]string{
					"type": "image",
					"index": fmt.Sprintf("%d", i),
				},
			}
			test.Steps = append(test.Steps, step)
		}

		tests = append(tests, test)
	}

	return tests
}

// generateFormTests creates form-related tests
func (tg *TestGenerator) generateFormTests(elements []vision.ElementInfo, analysis TestAnalysis) []GeneratedTest {
	var tests []GeneratedTest

	textFields := tg.filterElementsByType(elements, "textfield")
	buttons := tg.filterElementsByType(elements, "button")

	if len(textFields) > 0 && len(buttons) > 0 {
		// Generate form fill and submit test
		test := GeneratedTest{
			Name:        "Form Fill and Submit Test",
			Type:        "form",
			Description: "Tests form filling and submission functionality",
			Priority:    "high",
			Confidence:  0.82,
			Elements:    []string{"textfield", "button"},
			Duration:    (len(textFields) * 3) + (len(buttons) * 1), // 3s per field, 1s per button
		}

		// Fill text fields
		for i, _ := range textFields {
			step := TestStep{
				Action: "fill",
				Target: "input",
				Value:  fmt.Sprintf("form_test_value_%d", i),
			}
			test.Steps = append(test.Steps, step)
		}

		// Click submit button (if available)
		if len(buttons) > 0 {
			submitStep := TestStep{
				Action: "vision_click",
				Target: "button",
				Parameters: map[string]string{
					"type": "button",
					"text": "submit",
				},
			}
			test.Steps = append(test.Steps, submitStep)
		}

		tests = append(tests, test)

		// Generate form validation test
		validationTest := GeneratedTest{
			Name:        "Form Validation Test",
			Type:        "validation",
			Description: "Tests form validation with empty and invalid inputs",
			Priority:    "medium",
			Confidence:  0.75,
			Elements:    []string{"textfield", "button"},
			Duration:    len(textFields) * 2, // 2 seconds per validation test
		}

		// Test empty field validation
		for i, _ := range textFields {
			step1 := TestStep{
				Action: "vision_click",
				Target: "textfield",
				Parameters: map[string]string{
					"type": "textfield",
					"index": fmt.Sprintf("%d", i),
				},
			}
			step2 := TestStep{
				Action: "fill",
				Target: "input",
				Value:  "",
			}
			step3 := TestStep{
				Action: "wait",
				Parameters: map[string]string{
					"wait_time": "1",
				},
			}
			validationTest.Steps = append(validationTest.Steps, step1, step2, step3)
		}

		tests = append(tests, validationTest)
	}

	return tests
}

// generateErrorHandlingTests creates error handling tests
func (tg *TestGenerator) generateErrorHandlingTests(elements []vision.ElementInfo, analysis TestAnalysis) []GeneratedTest {
	var tests []GeneratedTest

	// Generate invalid input test
	textFields := tg.filterElementsByType(elements, "textfield")
	if len(textFields) > 0 {
		test := GeneratedTest{
			Name:        "Invalid Input Error Handling Test",
			Type:        "error_handling",
			Description: "Tests error handling with invalid inputs",
			Priority:    "medium",
			Confidence:  0.70,
			Elements:    []string{"textfield"},
			Duration:    len(textFields) * 4, // 4 seconds per error test
		}

		for i, _ := range textFields {
			step1 := TestStep{
				Action: "vision_click",
				Target: "textfield",
				Parameters: map[string]string{
					"type": "textfield",
					"index": fmt.Sprintf("%d", i),
				},
			}
			step2 := TestStep{
				Action: "fill",
				Target: "input",
				Value:  "@#$%^&*()_+!~`", // Special characters
			}
			test.Steps = append(test.Steps, step1, step2)
		}

		tests = append(tests, test)
	}

	return tests
}

// generateAccessibilityTests creates accessibility tests
func (tg *TestGenerator) generateAccessibilityTests(elements []vision.ElementInfo, analysis TestAnalysis) []GeneratedTest {
	var tests []GeneratedTest

	// Generate keyboard navigation test
	test := GeneratedTest{
		Name:        "Keyboard Navigation Accessibility Test",
		Type:        "accessibility",
		Description: "Tests keyboard navigation and accessibility",
		Priority:    "medium",
		Confidence:  0.65,
		Elements:    tg.getElementTypes(elements),
		Duration:    10, // 10 seconds for accessibility test
	}

	// Add keyboard navigation steps
	interactiveElements := []string{"button", "textfield", "link"}
	for _, elemType := range interactiveElements {
		step := TestStep{
			Action: "keyboard_test",
			Target: elemType,
			Parameters: map[string]string{
				"keys": "Tab,Enter,Space",
			},
		}
		test.Steps = append(test.Steps, step)
	}

	tests = append(tests, test)
	return tests
}

// generatePerformanceTests creates performance tests
func (tg *TestGenerator) generatePerformanceTests(elements []vision.ElementInfo, analysis TestAnalysis) []GeneratedTest {
	var tests []GeneratedTest

	// Generate rapid interaction test
	test := GeneratedTest{
		Name:        "Rapid Interaction Performance Test",
		Type:        "performance",
		Description: "Tests application performance under rapid interactions",
		Priority:    "low",
		Confidence:  0.60,
		Elements:    tg.getElementTypes(elements),
		Duration:    15, // 15 seconds for performance test
	}

	// Add rapid click steps for buttons
	buttons := tg.filterElementsByType(elements, "button")
	for i, _ := range buttons {
		if i < 5 { // Limit to first 5 buttons for performance test
			step := TestStep{
				Action: "rapid_click",
				Target: "button",
				Parameters: map[string]string{
					"count": "10",
					"interval": "100ms",
				},
			}
			test.Steps = append(test.Steps, step)
		}
	}

	tests = append(tests, test)
	return tests
}

// sortTestsByPriority sorts tests by priority and confidence
func (tg *TestGenerator) sortTestsByPriority(tests []GeneratedTest) {
	sort.Slice(tests, func(i, j int) bool {
		// First sort by priority
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		if priorityOrder[tests[i].Priority] != priorityOrder[tests[j].Priority] {
			return priorityOrder[tests[i].Priority] > priorityOrder[tests[j].Priority]
		}
		// Then sort by confidence
		return tests[i].Confidence > tests[j].Confidence
	})
}

// filterElementsByType filters elements by type
func (tg *TestGenerator) filterElementsByType(elements []vision.ElementInfo, elementType string) []vision.ElementInfo {
	var filtered []vision.ElementInfo
	for _, elem := range elements {
		if elem.Type == elementType {
			filtered = append(filtered, elem)
		}
	}
	return filtered
}

// getElementTypes gets all unique element types
func (tg *TestGenerator) getElementTypes(elements []vision.ElementInfo) []string {
	typeMap := make(map[string]bool)
	var types []string
	
	for _, elem := range elements {
		if !typeMap[elem.Type] {
			typeMap[elem.Type] = true
			types = append(types, elem.Type)
		}
	}
	
	return types
}

// GenerateRandomTests creates random test variations for diversity
func (tg *TestGenerator) GenerateRandomTests(elements []vision.ElementInfo, count int) []GeneratedTest {
	var tests []GeneratedTest
	
	rand.Seed(time.Now().UnixNano())
	
	elementTypes := tg.getElementTypes(elements)
	
	for i := 0; i < count; i++ {
		test := GeneratedTest{
			Name:        fmt.Sprintf("Random Test %d", i+1),
			Type:        "random",
			Description: "AI-generated random test for diversity",
			Priority:    []string{"high", "medium", "low"}[rand.Intn(3)],
			Confidence:  0.5 + (rand.Float64() * 0.4), // 0.5 to 0.9
			Elements:    elementTypes,
			Duration:    rand.Intn(10) + 5, // 5 to 15 seconds
		}
		
		// Add random steps
		stepCount := rand.Intn(5) + 2 // 2 to 6 steps
		for j := 0; j < stepCount; j++ {
			actionTypes := []string{"vision_click", "fill", "wait", "navigate"}
			action := actionTypes[rand.Intn(len(actionTypes))]
			
			step := TestStep{
				Action: action,
				Target: elementTypes[rand.Intn(len(elementTypes))],
				Value:  fmt.Sprintf("random_value_%d", j),
			}
			
			if action == "wait" {
				step.Value = fmt.Sprintf("%d", rand.Intn(3)+1) // 1 to 3 seconds
			}
			
			test.Steps = append(test.Steps, step)
		}
		
		tests = append(tests, test)
	}
	
	return tests
}

// GenerateAITestReport creates a comprehensive AI test generation report
func (tg *TestGenerator) GenerateAITestReport(tests []GeneratedTest, analysis TestAnalysis, outputPath string) error {
	tg.logger.Infof("Generating AI test generation report with %d tests", len(tests))
	
	content := fmt.Sprintf(`# AI-Powered Test Generation Report

## Test Analysis Summary
- **Total Elements Analyzed**: %d
- **Element Types**: %v
- **Application Complexity**: %s
- **Risk Level**: %s
- **Test Coverage Areas**: %v

## Generated Tests (%d)

`, analysis.TotalElements, tg.getElementTypesMap(analysis.ElementTypes), analysis.Complexity, analysis.RiskLevel, analysis.TestCoverage, len(tests))

	// Group tests by type
	testGroups := make(map[string][]GeneratedTest)
	for _, test := range tests {
		testGroups[test.Type] = append(testGroups[test.Type], test)
	}
	
	// Generate sections for each test type
	for testType, testList := range testGroups {
		content += fmt.Sprintf("### %s Tests (%d)\n\n", strings.Title(testType), len(testList))
		
		for i, test := range testList {
			content += fmt.Sprintf("#### %d. %s\n\n", i+1, test.Name)
			content += fmt.Sprintf("- **Type**: %s\n", test.Type)
			content += fmt.Sprintf("- **Priority**: %s\n", test.Priority)
			content += fmt.Sprintf("- **Confidence**: %.2f\n", test.Confidence)
			content += fmt.Sprintf("- **Description**: %s\n", test.Description)
			content += fmt.Sprintf("- **Duration**: %d seconds\n", test.Duration)
			content += fmt.Sprintf("- **Elements**: %v\n", test.Elements)
			
			if len(test.Steps) > 0 {
				content += "- **Steps**:\n"
				for j, step := range test.Steps {
					content += fmt.Sprintf("  %d. %s %s", j+1, step.Action, step.Target)
					if step.Value != "" {
						content += fmt.Sprintf(" with value '%s'", step.Value)
					}
					if len(step.Parameters) > 0 {
						content += fmt.Sprintf(" (params: %v)", step.Parameters)
					}
					content += "\n"
				}
			}
			content += "\n"
		}
	}
	
	// Add recommendations
	content += tg.generateTestRecommendations(analysis, tests)
	
	// Write to file
	filename := fmt.Sprintf("%s/ai_test_generation_report.md", outputPath)
	
	return os.WriteFile(filename, []byte(content), 0600)
}

// getElementTypesMap converts element type map to string representation
func (tg *TestGenerator) getElementTypesMap(elementTypes map[string]int) string {
	var result []string
	for elemType, count := range elementTypes {
		result = append(result, fmt.Sprintf("%s(%d)", elemType, count))
	}
	return fmt.Sprintf("[%s]", strings.Join(result, ", "))
}

// generateTestRecommendations creates AI-generated test recommendations
func (tg *TestGenerator) generateTestRecommendations(analysis TestAnalysis, tests []GeneratedTest) string {
	content := "## AI Recommendations\n\n"
	
	// Priority recommendations
	highPriorityCount := 0
	mediumPriorityCount := 0
	lowPriorityCount := 0
	
	for _, test := range tests {
		switch test.Priority {
		case "high":
			highPriorityCount++
		case "medium":
			mediumPriorityCount++
		case "low":
			lowPriorityCount++
		}
	}
	
	content += "### Test Priority Distribution\n\n"
	content += fmt.Sprintf("- **High Priority**: %d tests (executed first)\n", highPriorityCount)
	content += fmt.Sprintf("- **Medium Priority**: %d tests (executed after high priority)\n", mediumPriorityCount)
	content += fmt.Sprintf("- **Low Priority**: %d tests (executed if time permits)\n\n", lowPriorityCount)
	
	// Coverage recommendations
	if _, hasTextFields := analysis.ElementTypes["textfield"]; hasTextFields {
		if _, hasButtons := analysis.ElementTypes["button"]; hasButtons {
			content += "### Form Testing Recommendations\n\n"
			content += "✅ Form interaction tests are included\n"
			content += "🔍 Consider adding validation tests for form fields\n"
			content += "🚫 Consider adding negative test cases\n\n"
		}
	}
	
	if len(analysis.ElementTypes) > 5 {
		content += "### Complexity Recommendations\n\n"
		content += "⚠️ High application complexity detected\n"
		content += "🔄 Consider breaking tests into smaller test suites\n"
		content += "⏱️ Allocate more time for comprehensive testing\n\n"
	}
	
	// AI improvement suggestions
	content += "### AI Improvement Suggestions\n\n"
	content += "🤖 Enable machine learning for better test accuracy\n"
	content += "📊 Collect test execution data for confidence improvement\n"
	content += "🎯 Implement adaptive test generation based on results\n"
	
	return content
}