package ai

import (
	"testing"
	"time"

	"panoptic/internal/logger"
	"panoptic/internal/vision"
)

// Benchmark AIEnhancedTester operations

func BenchmarkNewAIEnhancedTester(b *testing.B) {
	log := logger.NewLogger(false)
	detector := vision.NewDetector(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAIEnhancedTester(detector, log)
	}
}

func BenchmarkAIEnhancedTester_AnalyzeVisualElements_Empty(b *testing.B) {
	log := logger.NewLogger(false)
	detector := vision.NewDetector(log)
	tester := NewAIEnhancedTester(detector, log)

	elements := []vision.VisualElement{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tester.analyzeVisualElements(elements)
	}
}

func BenchmarkAIEnhancedTester_AnalyzeVisualElements_Small(b *testing.B) {
	log := logger.NewLogger(false)
	detector := vision.NewDetector(log)
	tester := NewAIEnhancedTester(detector, log)

	elements := []vision.VisualElement{
		{Type: "button", X: 10, Y: 20, Width: 100, Height: 40, Confidence: 0.9},
		{Type: "textfield", X: 10, Y: 70, Width: 200, Height: 30, Confidence: 0.85},
		{Type: "link", X: 10, Y: 110, Width: 150, Height: 20, Confidence: 0.88},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tester.analyzeVisualElements(elements)
	}
}

func BenchmarkAIEnhancedTester_AnalyzeVisualElements_Large(b *testing.B) {
	log := logger.NewLogger(false)
	detector := vision.NewDetector(log)
	tester := NewAIEnhancedTester(detector, log)

	// Simulate a large page with many elements
	elements := make([]vision.VisualElement, 1000)
	for i := 0; i < 1000; i++ {
		elemType := []string{"button", "textfield", "link", "image"}[i%4]
		elements[i] = vision.VisualElement{
			Type:       elemType,
			X:          i * 10,
			Y:          (i % 100) * 50,
			Width:      100,
			Height:     40,
			Confidence: 0.75 + float64(i%25)/100,
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tester.analyzeVisualElements(elements)
	}
}

// Benchmark TestGenerator operations

func BenchmarkNewTestGenerator(b *testing.B) {
	log := logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTestGenerator(log)
	}
}

func BenchmarkTestGenerator_GenerateTests_Small(b *testing.B) {
	log := logger.NewLogger(false)
	generator := NewTestGenerator(log)

	elements := []vision.VisualElement{
		{Type: "button", Text: "Submit", X: 10, Y: 20, Width: 100, Height: 40, Confidence: 0.9},
		{Type: "textfield", Text: "Email", X: 10, Y: 70, Width: 200, Height: 30, Confidence: 0.85},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tests, err := generator.GenerateTests(elements)
		if err != nil {
			b.Fatal(err)
		}
		_ = tests
	}
}

func BenchmarkTestGenerator_GenerateTests_Large(b *testing.B) {
	log := logger.NewLogger(false)
	generator := NewTestGenerator(log)

	// Simulate a complex page
	elements := make([]vision.VisualElement, 100)
	for i := 0; i < 100; i++ {
		elemType := []string{"button", "textfield", "link", "image"}[i%4]
		elements[i] = vision.VisualElement{
			Type:       elemType,
			Text:       "Element" + string(rune(i)),
			X:          i * 10,
			Y:          (i % 20) * 50,
			Width:      100,
			Height:     40,
			Confidence: 0.8,
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tests, err := generator.GenerateTests(elements)
		if err != nil {
			b.Fatal(err)
		}
		_ = tests
	}
}

func BenchmarkTestGenerator_FilterElements(b *testing.B) {
	log := logger.NewLogger(false)
	generator := NewTestGenerator(log)

	elements := make([]vision.VisualElement, 1000)
	for i := 0; i < 1000; i++ {
		elements[i] = vision.VisualElement{
			Type:       "button",
			X:          i * 10,
			Y:          i * 10,
			Confidence: float64(i%100) / 100.0, // Varying confidence
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.filterHighConfidenceElements(elements, 0.7)
	}
}

// Benchmark ErrorDetector operations

func BenchmarkNewErrorDetector(b *testing.B) {
	log := logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewErrorDetector(log)
	}
}

func BenchmarkErrorDetector_DetectErrors_NoErrors(b *testing.B) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(log)

	pageState := map[string]interface{}{
		"status": "success",
		"title":  "Home Page",
		"url":    "https://example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errors, err := detector.DetectErrors(pageState)
		if err != nil {
			b.Fatal(err)
		}
		_ = errors
	}
}

func BenchmarkErrorDetector_DetectErrors_WithErrors(b *testing.B) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(log)

	pageState := map[string]interface{}{
		"status":  "error",
		"title":   "404 Not Found",
		"url":     "https://example.com/notfound",
		"console": "Error: Failed to load resource",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errors, err := detector.DetectErrors(pageState)
		if err != nil {
			b.Fatal(err)
		}
		_ = errors
	}
}

func BenchmarkErrorDetector_AnalyzeConsoleErrors(b *testing.B) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(log)

	consoleLogs := []string{
		"INFO: Application started",
		"ERROR: Failed to fetch data from API",
		"WARN: Deprecated function used",
		"ERROR: Network request failed: timeout",
		"INFO: User logged in successfully",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.analyzeConsoleErrors(consoleLogs)
	}
}

func BenchmarkErrorDetector_CategorizeError(b *testing.B) {
	log := logger.NewLogger(false)
	detector := NewErrorDetector(log)

	errorMessages := []string{
		"Network request failed: 404",
		"Element not found: .missing-selector",
		"Authentication failed: invalid credentials",
		"Form validation error: invalid email format",
		"JavaScript error: undefined is not a function",
		"Timeout error: operation took too long",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, msg := range errorMessages {
			_ = detector.categorizeError(msg)
		}
	}
}

// Benchmark test case generation patterns

func BenchmarkGenerateTestCase_Creation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testCase := GeneratedTest{
			ID:          "test-001",
			Name:        "Button Click Test",
			Description: "Verify button click functionality",
			Steps: []TestStep{
				{Action: "navigate", Target: "https://example.com"},
				{Action: "click", Target: ".submit-button"},
				{Action: "wait", Duration: 1},
			},
			ExpectedOutcome: "Button clicked successfully",
			Priority:        "high",
			Category:        "interaction",
			Confidence:      0.85,
		}
		_ = testCase
	}
}

func BenchmarkGenerateMultipleTestCases(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tests := make([]GeneratedTest, 0, 50)
		for j := 0; j < 50; j++ {
			test := GeneratedTest{
				ID:          "test-" + string(rune(j)),
				Name:        "Test Case " + string(rune(j)),
				Description: "Auto-generated test case",
				Steps: []TestStep{
					{Action: "navigate", Target: "https://example.com"},
					{Action: "click", Target: ".button-" + string(rune(j))},
				},
				ExpectedOutcome: "Action completed",
				Priority:        "medium",
				Category:        "generated",
				Confidence:      0.75,
			}
			tests = append(tests, test)
		}
		_ = tests
	}
}

// Benchmark error pattern matching

func BenchmarkErrorPatternMatching_Simple(b *testing.B) {
	patterns := []string{
		"error",
		"fail",
		"exception",
		"timeout",
		"404",
		"500",
	}
	text := "The application encountered an error while processing the request"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pattern := range patterns {
			if len(text) > 0 && len(pattern) > 0 {
				// Simplified pattern matching
				_ = pattern
			}
		}
	}
}

// Benchmark confidence calculation

func BenchmarkConfidenceCalculation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		elementConfidence := 0.85
		contextScore := 0.75
		positionScore := 0.90

		// Weighted average
		totalConfidence := (elementConfidence*0.5 + contextScore*0.3 + positionScore*0.2)
		_ = totalConfidence
	}
}

// Benchmark report generation

func BenchmarkGenerateAIReport_Small(b *testing.B) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	summary := "AI-Enhanced Testing Analysis"
	tests := make([]GeneratedTest, 10)
	errors := make([]DetectedError, 5)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		report := "# " + summary + "\n\n"
		report += "Generated: " + timestamp + "\n\n"
		report += "Total Tests: " + string(rune(len(tests))) + "\n"
		report += "Total Errors: " + string(rune(len(errors))) + "\n"
		_ = report
	}
}
