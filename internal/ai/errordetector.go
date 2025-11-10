package ai

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"panoptic/internal/logger"
)

// ErrorDetector provides intelligent error detection and analysis
type ErrorDetector struct {
	logger  logger.Logger
	enabled bool
	patterns []ErrorPattern
}

// NewErrorDetector creates a new smart error detector
func NewErrorDetector(log logger.Logger) *ErrorDetector {
	detector := &ErrorDetector{
		logger:  log,
		enabled: true,
		patterns: initializeErrorPatterns(),
	}
	return detector
}

// ErrorPattern represents an error detection pattern
type ErrorPattern struct {
	Name        string            `json:"name"`
	Category    string            `json:"category"`
	Pattern     *regexp.Regexp    `json:"pattern"`
	Severity    string            `json:"severity"`
	Confidence  float64           `json:"confidence"`
	Description string            `json:"description"`
	Suggestions []string          `json:"suggestions"`
	Tags        []string          `json:"tags"`
}

// DetectedError represents a detected error
type DetectedError struct {
	Name        string            `json:"name"`
	Category    string            `json:"category"`
	Message     string            `json:"message"`
	Severity    string            `json:"severity"`
	Confidence  float64           `json:"confidence"`
	Timestamp   time.Time         `json:"timestamp"`
	Source      string            `json:"source"`
	Position    ErrorPosition      `json:"position"`
	Suggestions []string          `json:"suggestions"`
	Tags        []string          `json:"tags"`
	Context     map[string]string `json:"context"`
}

// ErrorPosition represents the position where error occurred
type ErrorPosition struct {
	Type     string `json:"type"`     // selector, coordinates, etc.
	Value    string `json:"value"`    // selector value or coordinates
	X        int    `json:"x"`        // X coordinate if applicable
	Y        int    `json:"y"`        // Y coordinate if applicable
	Element  string `json:"element"`  // Element type if known
}

// ErrorAnalysis contains comprehensive error analysis
type ErrorAnalysis struct {
	TotalErrors       int                         `json:"total_errors"`
	ErrorCategories   map[string]int               `json:"error_categories"`
	SeverityLevels    map[string]int               `json:"severity_levels"`
	CriticalErrors   []DetectedError             `json:"critical_errors"`
	HighRiskErrors   []DetectedError             `json:"high_risk_errors"`
	ErrorTrends      []ErrorTrend                `json:"error_trends"`
	Recommendations   []ErrorRecommendation       `json:"recommendations"`
	TestCoverage     []string                    `json:"test_coverage"`
	ErrorPatterns    []ErrorPattern              `json:"detected_patterns"`
}

// ErrorTrend represents error trend over time
type ErrorTrend struct {
	Timestamp   time.Time `json:"timestamp"`
	ErrorCount  int        `json:"error_count"`
	Category    string     `json:"category"`
	Severity    string     `json:"severity"`
}

// ErrorRecommendation represents AI-generated error fix recommendations
type ErrorRecommendation struct {
	Type         string   `json:"type"`         // fix, improve, prevent
	Priority     string   `json:"priority"`     // high, medium, low
	Description  string   `json:"description"`
	Suggestion   string   `json:"suggestion"`
	Steps        []string `json:"steps"`
	Impact       string   `json:"impact"`
	Effort       string   `json:"effort"`
	Tags         []string `json:"tags"`
}

// initializeErrorPatterns sets up error detection patterns
func initializeErrorPatterns() []ErrorPattern {
	patterns := []ErrorPattern{
		// Network/Connection Errors
		{
			Name:        "NetworkTimeout",
			Category:    "network",
			Pattern:     regexp.MustCompile(`(?i)(timeout|timed out|connection timed|network timeout|request timeout)`),
			Severity:    "high",
			Confidence:  0.85,
			Description: "Network connection or request timeout",
			Suggestions: []string{
				"Increase timeout duration",
				"Check network connectivity",
				"Implement retry mechanism",
				"Verify endpoint availability",
			},
			Tags: []string{"network", "timeout", "connection"},
		},
		{
			Name:        "ConnectionRefused",
			Category:    "network",
			Pattern:     regexp.MustCompile(`(?i)(connection refused|conn refused|cannot connect|unable to connect)`),
			Severity:    "high",
			Confidence:  0.90,
			Description: "Server connection refused",
			Suggestions: []string{
				"Verify server is running",
				"Check firewall settings",
				"Verify correct port and address",
				"Ensure server is accepting connections",
			},
			Tags: []string{"network", "connection", "refused"},
		},
		
		// UI/Element Errors
		{
			Name:        "ElementNotFound",
			Category:    "ui",
			Pattern:     regexp.MustCompile(`(?i)(element not found|no such element|unable to locate element|element.*not found)`),
			Severity:    "medium",
			Confidence:  0.75,
			Description: "UI element not found on page",
			Suggestions: []string{
				"Wait for element to load",
				"Check element selector accuracy",
				"Verify element is present in DOM",
				"Use alternative locator strategy",
			},
			Tags: []string{"ui", "element", "locator"},
		},
		{
			Name:        "ElementNotClickable",
			Category:    "ui",
			Pattern:     regexp.MustCompile(`(?i)(element not clickable|not clickable|element.*obscured|element.*covered)`),
			Severity:    "medium",
			Confidence:  0.70,
			Description: "Element cannot be clicked",
			Suggestions: []string{
				"Scroll element into view",
				"Wait for element to be visible",
				"Check if element is enabled",
				"Use JavaScript click as fallback",
			},
			Tags: []string{"ui", "click", "visibility"},
		},
		
		// Authentication Errors
		{
			Name:        "AuthenticationFailed",
			Category:    "authentication",
			Pattern:     regexp.MustCompile(`(?i)(unauthorized|authentication failed|login failed|invalid credentials|access denied)`),
			Severity:    "high",
			Confidence:  0.85,
			Description: "Authentication or authorization failed",
			Suggestions: []string{
				"Verify username and password",
				"Check authentication service status",
				"Verify user permissions",
				"Check session token validity",
			},
			Tags: []string{"auth", "login", "security"},
		},
		{
			Name:        "SessionExpired",
			Category:    "authentication",
			Pattern:     regexp.MustCompile(`(?i)(session expired|session timeout|invalid session|token expired)`),
			Severity:    "medium",
			Confidence:  0.80,
			Description: "User session has expired",
			Suggestions: []string{
				"Implement session refresh",
				"Show session expiration warning",
				"Auto-redirect to login page",
				"Extend session timeout",
			},
			Tags: []string{"auth", "session", "timeout"},
		},
		
		// Form/Validation Errors
		{
			Name:        "ValidationError",
			Category:    "validation",
			Pattern:     regexp.MustCompile(`(?i)(validation error|invalid input|field.*required|format.*invalid)`),
			Severity:    "low",
			Confidence:  0.65,
			Description: "Form validation failed",
			Suggestions: []string{
				"Provide valid input format",
				"Fill required fields",
				"Check field constraints",
				"Display clear error messages",
			},
			Tags: []string{"form", "validation", "input"},
		},
		{
			Name:        "RequiredFieldMissing",
			Category:    "validation",
			Pattern:     regexp.MustCompile(`(?i)(required field|field.*required|missing.*field|field.*missing)`),
			Severity:    "low",
			Confidence:  0.70,
			Description: "Required form field is missing",
			Suggestions: []string{
				"Fill all required fields",
				"Mark required fields clearly",
				"Add field validation indicators",
				"Provide helpful field labels",
			},
			Tags: []string{"form", "validation", "required"},
		},
		
		// Performance/Timeout Errors
		{
			Name:        "PageLoadTimeout",
			Category:    "performance",
			Pattern:     regexp.MustCompile(`(?i)(page.*timeout|load timeout|page not loading|slow page)`),
			Severity:    "high",
			Confidence:  0.75,
			Description: "Page failed to load within timeout",
			Suggestions: []string{
				"Increase page load timeout",
				"Check page size and complexity",
				"Optimize page resources",
				"Check server response time",
			},
			Tags: []string{"performance", "timeout", "load"},
		},
		{
			Name:        "ElementLoadTimeout",
			Category:    "performance",
			Pattern:     regexp.MustCompile(`(?i)(element.*timeout|element.*not loaded|wait.*timeout)`),
			Severity:    "medium",
			Confidence:  0.70,
			Description: "Element failed to load within timeout",
			Suggestions: []string{
				"Increase element wait timeout",
				"Check element dependencies",
				"Verify element exists in page source",
				"Use explicit wait conditions",
			},
			Tags: []string{"performance", "element", "timeout"},
		},
		
		// JavaScript/Runtime Errors
		{
			Name:        "JavaScriptError",
			Category:    "javascript",
			Pattern:     regexp.MustCompile(`(?i)(javascript error|script error|js error|runtime error)`),
			Severity:    "high",
			Confidence:  0.80,
			Description: "JavaScript runtime error occurred",
			Suggestions: []string{
				"Check browser console for details",
				"Verify script syntax",
				"Debug JavaScript code",
				"Check for undefined variables",
			},
			Tags: []string{"javascript", "runtime", "error"},
		},
		{
			Name:        "UndefinedReference",
			Category:    "javascript",
			Pattern:     regexp.MustCompile(`(?i)(undefined|not defined|null.*reference|object.*null)`),
			Severity:    "medium",
			Confidence:  0.75,
			Description: "JavaScript undefined or null reference",
			Suggestions: []string{
				"Check variable declarations",
				"Add null/undefined checks",
				"Initialize variables properly",
				"Debug object references",
			},
			Tags: []string{"javascript", "undefined", "reference"},
		},
		
		// Database/System Errors
		{
			Name:        "DatabaseError",
			Category:    "database",
			Pattern:     regexp.MustCompile(`(?i)(database error|sql error|connection.*failed|query.*failed)`),
			Severity:    "high",
			Confidence:  0.85,
			Description: "Database operation failed",
			Suggestions: []string{
				"Check database connection",
				"Verify SQL query syntax",
				"Check database permissions",
				"Review database logs",
			},
			Tags: []string{"database", "sql", "connection"},
		},
		{
			Name:        "FileNotFoundError",
			Category:    "filesystem",
			Pattern:     regexp.MustCompile(`(?i)(file not found|no such file|path.*not found|file.*missing)`),
			Severity:    "medium",
			Confidence:  0.80,
			Description: "File or directory not found",
			Suggestions: []string{
				"Verify file path exists",
				"Check file permissions",
				"Ensure file is not deleted",
				"Use absolute file paths",
			},
			Tags: []string{"filesystem", "file", "path"},
		},
	}
	
	return patterns
}

// DetectErrors analyzes messages and detects errors
func (ed *ErrorDetector) DetectErrors(messages []ErrorMessage) []DetectedError {
	if !ed.enabled {
		return []DetectedError{}
	}

	var detectedErrors []DetectedError

	for _, msg := range messages {
		errors := ed.analyzeMessage(msg)
		detectedErrors = append(detectedErrors, errors...)
	}

	ed.logger.Infof("Detected %d errors from %d messages", len(detectedErrors), len(messages))
	return detectedErrors
}

// ErrorMessage represents a message to be analyzed for errors
type ErrorMessage struct {
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
	Level     string                 `json:"level"`
}

// analyzeMessage analyzes a single message for errors
func (ed *ErrorDetector) analyzeMessage(msg ErrorMessage) []DetectedError {
	var detectedErrors []DetectedError

	// Check against each error pattern
	for _, pattern := range ed.patterns {
		matches := pattern.Pattern.FindAllString(msg.Message, -1)
		if len(matches) > 0 {
			error := DetectedError{
				Name:        pattern.Name,
				Category:    pattern.Category,
				Message:     msg.Message,
				Severity:    pattern.Severity,
				Confidence:  pattern.Confidence,
				Timestamp:   msg.Timestamp,
				Source:      msg.Source,
				Suggestions: pattern.Suggestions,
				Tags:        pattern.Tags,
				Context:     ed.convertContext(msg.Context),
			}
			
			// Extract position if available
			error.Position = ed.extractPosition(msg.Context)
			
			detectedErrors = append(detectedErrors, error)
		}
	}

	// Additional pattern matching for common error indicators
	if ed.containsErrorIndicators(msg.Message) && len(detectedErrors) == 0 {
		error := DetectedError{
			Name:        "UnknownError",
			Category:    "general",
			Message:     msg.Message,
			Severity:    "medium",
			Confidence:  0.50,
			Timestamp:   msg.Timestamp,
			Source:      msg.Source,
			Suggestions: []string{
				"Review error message for details",
				"Check application logs",
				"Verify application state",
				"Contact support if issue persists",
			},
			Tags:        []string{"unknown", "general"},
			Context:     ed.convertContext(msg.Context),
		}
		
		error.Position = ed.extractPosition(msg.Context)
		detectedErrors = append(detectedErrors, error)
	}

	return detectedErrors
}

// containsErrorIndicators checks for common error indicators
func (ed *ErrorDetector) containsErrorIndicators(message string) bool {
	indicators := []string{
		"error", "failed", "failure", "exception", "fault",
		"crash", "panic", "abort", "terminate", "unable",
		"cannot", "could not", "not possible", "invalid",
		"incorrect", "wrong", "bad", "malformed",
	}
	
	messageLower := strings.ToLower(message)
	for _, indicator := range indicators {
		if strings.Contains(messageLower, indicator) {
			return true
		}
	}
	return false
}

// extractPosition extracts error position from context
func (ed *ErrorDetector) extractPosition(context map[string]interface{}) ErrorPosition {
	position := ErrorPosition{
		Type: "unknown",
	}
	
	if selector, ok := context["selector"].(string); ok {
		position.Type = "selector"
		position.Value = selector
	}
	
	if x, ok := context["x"].(int); ok {
		position.X = x
	}
	
	if y, ok := context["y"].(int); ok {
		position.Y = y
	}
	
	if element, ok := context["element"].(string); ok {
		position.Element = element
	}
	
	if position.Type == "unknown" && position.X > 0 && position.Y > 0 {
		position.Type = "coordinates"
		position.Value = fmt.Sprintf("(%d, %d)", position.X, position.Y)
	}
	
	return position
}

// convertContext converts interface context to string context
func (ed *ErrorDetector) convertContext(context map[string]interface{}) map[string]string {
	stringContext := make(map[string]string)
	
	for key, value := range context {
		if str, ok := value.(string); ok {
			stringContext[key] = str
		} else {
			stringContext[key] = fmt.Sprintf("%v", value)
		}
	}
	
	return stringContext
}

// AnalyzeErrors performs comprehensive error analysis
func (ed *ErrorDetector) AnalyzeErrors(errors []DetectedError) ErrorAnalysis {
	analysis := ErrorAnalysis{
		TotalErrors:     len(errors),
		ErrorCategories:  make(map[string]int),
		SeverityLevels:   make(map[string]int),
		CriticalErrors:   []DetectedError{},
		HighRiskErrors:   []DetectedError{},
		ErrorTrends:      []ErrorTrend{},
		Recommendations:   []ErrorRecommendation{},
		TestCoverage:     []string{},
		ErrorPatterns:    []ErrorPattern{},
	}

	// Analyze error categories and severity
	for _, error := range errors {
		analysis.ErrorCategories[error.Category]++
		analysis.SeverityLevels[error.Severity]++
		
		// Categorize critical and high-risk errors
		if error.Severity == "critical" {
			analysis.CriticalErrors = append(analysis.CriticalErrors, error)
		} else if error.Severity == "high" {
			analysis.HighRiskErrors = append(analysis.HighRiskErrors, error)
		}
	}

	// Generate error trends
	analysis.ErrorTrends = ed.generateErrorTrends(errors)
	
	// Generate recommendations
	analysis.Recommendations = ed.generateErrorRecommendations(analysis)
	
	// Determine test coverage gaps
	analysis.TestCoverage = ed.determineTestCoverage(analysis)
	
	// Extract detected patterns
	analysis.ErrorPatterns = ed.extractDetectedPatterns(errors)

	return analysis
}

// generateErrorTrends analyzes error trends over time
func (ed *ErrorDetector) generateErrorTrends(errors []DetectedError) []ErrorTrend {
	var trends []ErrorTrend
	
	// Group errors by time windows (e.g., hourly)
	timeWindows := make(map[string][]DetectedError)
	
	for _, error := range errors {
		timeKey := error.Timestamp.Format("2006-01-02-15") // Hourly grouping
		timeWindows[timeKey] = append(timeWindows[timeKey], error)
	}
	
	// Create trend data
	for timeKey, windowErrors := range timeWindows {
		if timestamp, err := time.Parse("2006-01-02-15", timeKey); err == nil {
			// Get most common category and severity for this window
			category := ed.getMostCommonCategory(windowErrors)
			severity := ed.getMostCommonSeverity(windowErrors)
			
			trend := ErrorTrend{
				Timestamp:  timestamp,
				ErrorCount: len(windowErrors),
				Category:   category,
				Severity:   severity,
			}
			trends = append(trends, trend)
		}
	}
	
	return trends
}

// getMostCommonCategory finds the most common error category
func (ed *ErrorDetector) getMostCommonCategory(errors []DetectedError) string {
	categoryCount := make(map[string]int)
	
	for _, error := range errors {
		categoryCount[error.Category]++
	}
	
	var mostCommon string
	maxCount := 0
	
	for category, count := range categoryCount {
		if count > maxCount {
			maxCount = count
			mostCommon = category
		}
	}
	
	return mostCommon
}

// getMostCommonSeverity finds the most common error severity
func (ed *ErrorDetector) getMostCommonSeverity(errors []DetectedError) string {
	severityCount := make(map[string]int)
	
	for _, error := range errors {
		severityCount[error.Severity]++
	}
	
	var mostCommon string
	maxCount := 0
	
	for severity, count := range severityCount {
		if count > maxCount {
			maxCount = count
			mostCommon = severity
		}
	}
	
	return mostCommon
}

// generateErrorRecommendations creates AI-powered error fix recommendations
func (ed *ErrorDetector) generateErrorRecommendations(analysis ErrorAnalysis) []ErrorRecommendation {
	var recommendations []ErrorRecommendation
	
	// Critical error recommendations
	if len(analysis.CriticalErrors) > 0 {
		rec := ErrorRecommendation{
			Type:        "fix",
			Priority:    "high",
			Description: "Critical errors detected requiring immediate attention",
			Suggestion:  "Address critical errors immediately to prevent system failure",
			Steps: []string{
				"Review critical error details",
				"Implement emergency fixes",
				"Add monitoring for critical issues",
				"Schedule immediate testing",
			},
			Impact:   "Prevents system failures and data loss",
			Effort:   "High - Requires immediate resources",
			Tags:     []string{"critical", "urgent", "fix"},
		}
		recommendations = append(recommendations, rec)
	}
	
	// High severity error recommendations
	if len(analysis.HighRiskErrors) > 0 {
		rec := ErrorRecommendation{
			Type:        "fix",
			Priority:    "high",
			Description: "High severity errors require prompt resolution",
			Suggestion:  "Prioritize high severity error fixes in next release",
			Steps: []string{
				"Analyze high severity error patterns",
				"Implement targeted fixes",
				"Add automated error detection",
				"Schedule regression testing",
			},
			Impact:   "Reduces system instability and user impact",
			Effort:   "Medium - Can be addressed in next sprint",
			Tags:     []string{"high", "priority", "fix"},
		}
		recommendations = append(recommendations, rec)
	}
	
	// Category-specific recommendations
	for category, count := range analysis.ErrorCategories {
		if count > 5 { // More than 5 errors in a category
			rec := ErrorRecommendation{
				Type:        "improve",
				Priority:    "medium",
				Description: fmt.Sprintf("High number of %s errors detected", category),
				Suggestion:  fmt.Sprintf("Investigate and fix root causes of %s errors", category),
				Steps: []string{
					fmt.Sprintf("Analyze %s error patterns", category),
					fmt.Sprintf("Review %s-related code", category),
					"Implement comprehensive testing",
					"Add error prevention measures",
				},
				Impact:   fmt.Sprintf("Reduces %s-related errors and improves reliability", category),
				Effort:   "Medium - Requires focused investigation",
				Tags:     []string{category, "improve", "reliability"},
			}
			recommendations = append(recommendations, rec)
		}
	}
	
	return recommendations
}

// determineTestCoverage identifies test coverage gaps based on errors
func (ed *ErrorDetector) determineTestCoverage(analysis ErrorAnalysis) []string {
	var coverage []string
	
	// Check if specific error types suggest missing test coverage
	if analysis.ErrorCategories["ui"] > 0 {
		coverage = append(coverage, "UI automation tests")
		coverage = append(coverage, "Element locator testing")
	}
	
	if analysis.ErrorCategories["network"] > 0 {
		coverage = append(coverage, "Network connectivity tests")
		coverage = append(coverage, "API endpoint tests")
	}
	
	if analysis.ErrorCategories["authentication"] > 0 {
		coverage = append(coverage, "Authentication flow tests")
		coverage = append(coverage, "Session management tests")
	}
	
	if analysis.ErrorCategories["validation"] > 0 {
		coverage = append(coverage, "Form validation tests")
		coverage = append(coverage, "Input boundary testing")
	}
	
	if analysis.ErrorCategories["performance"] > 0 {
		coverage = append(coverage, "Performance tests")
		coverage = append(coverage, "Load testing")
	}
	
	return coverage
}

// extractDetectedPatterns extracts patterns from detected errors
func (ed *ErrorDetector) extractDetectedPatterns(errors []DetectedError) []ErrorPattern {
	patternCounts := make(map[string]int)
	
	for _, error := range errors {
		patternCounts[error.Name]++
	}
	
	var detectedPatterns []ErrorPattern
	
	// Include patterns that were detected
	for _, pattern := range ed.patterns {
		if count, exists := patternCounts[pattern.Name]; exists {
			if count > 0 {
				detectedPatterns = append(detectedPatterns, pattern)
			}
		}
	}
	
	return detectedPatterns
}

// GenerateSmartErrorReport creates a comprehensive smart error analysis report
func (ed *ErrorDetector) GenerateSmartErrorReport(errors []DetectedError, analysis ErrorAnalysis, outputPath string) error {
	ed.logger.Infof("Generating smart error report with %d errors", len(errors))
	
	content := fmt.Sprintf(`# Smart Error Detection Report

## Error Analysis Summary
- **Total Errors Detected**: %d
- **Error Categories**: %v
- **Severity Levels**: %v
- **Critical Errors**: %d
- **High Risk Errors**: %d

## Error Category Distribution

`, analysis.TotalErrors, ed.formatMap(analysis.ErrorCategories), ed.formatMap(analysis.SeverityLevels), len(analysis.CriticalErrors), len(analysis.HighRiskErrors))

	// Add category breakdown
	for category, count := range analysis.ErrorCategories {
		content += fmt.Sprintf("### %s Errors (%d)\n\n", strings.Title(category), count)
		
		// Add errors in this category
		categoryErrors := []DetectedError{}
		for _, error := range errors {
			if error.Category == category {
				categoryErrors = append(categoryErrors, error)
			}
		}
		
		for i, error := range categoryErrors {
			content += fmt.Sprintf("#### %d. %s\n\n", i+1, error.Name)
			content += fmt.Sprintf("- **Message**: %s\n", error.Message)
			content += fmt.Sprintf("- **Severity**: %s\n", error.Severity)
			content += fmt.Sprintf("- **Confidence**: %.2f\n", error.Confidence)
			content += fmt.Sprintf("- **Source**: %s\n", error.Source)
			content += fmt.Sprintf("- **Timestamp**: %s\n", error.Timestamp.Format(time.RFC3339))
			content += fmt.Sprintf("- **Position**: %s (%s)\n", error.Position.Type, error.Position.Value)
			
			if len(error.Suggestions) > 0 {
				content += "- **Suggestions**:\n"
				for _, suggestion := range error.Suggestions {
					content += fmt.Sprintf("  - %s\n", suggestion)
				}
			}
			
			if len(error.Tags) > 0 {
				content += fmt.Sprintf("- **Tags**: %v\n", error.Tags)
			}
			
			content += "\n"
		}
	}
	
	// Add critical errors section
	if len(analysis.CriticalErrors) > 0 {
		content += fmt.Sprintf("## Critical Errors (%d)\n\n", len(analysis.CriticalErrors))
		for i, error := range analysis.CriticalErrors {
			content += fmt.Sprintf("### %d. %s\n\n", i+1, error.Name)
			content += fmt.Sprintf("- **Message**: %s\n", error.Message)
			content += fmt.Sprintf("- **Source**: %s\n", error.Source)
			content += fmt.Sprintf("- **Timestamp**: %s\n", error.Timestamp.Format(time.RFC3339))
			content += fmt.Sprintf("- **Suggestions**: %v\n", error.Suggestions)
			content += "\n"
		}
	}
	
	// Add recommendations
	content += ed.generateRecommendationsSection(analysis.Recommendations)
	
	// Add test coverage
	if len(analysis.TestCoverage) > 0 {
		content += "## Test Coverage Gaps\n\n"
		for _, coverage := range analysis.TestCoverage {
			content += fmt.Sprintf("- **Recommended**: %s\n", coverage)
		}
		content += "\n"
	}
	
	// Write to file
	filename := fmt.Sprintf("%s/smart_error_report.md", outputPath)
	
	return os.WriteFile(filename, []byte(content), 0644)
}

// formatMap converts map to string representation
func (ed *ErrorDetector) formatMap(m map[string]int) string {
	var result []string
	for key, value := range m {
		result = append(result, fmt.Sprintf("%s(%d)", key, value))
	}
	return fmt.Sprintf("[%s]", strings.Join(result, ", "))
}

// generateRecommendationsSection creates recommendations section for report
func (ed *ErrorDetector) generateRecommendationsSection(recommendations []ErrorRecommendation) string {
	content := "## AI-Generated Recommendations\n\n"
	
	// Group by priority
	highPriority := []ErrorRecommendation{}
	mediumPriority := []ErrorRecommendation{}
	lowPriority := []ErrorRecommendation{}
	
	for _, rec := range recommendations {
		switch rec.Priority {
		case "high":
			highPriority = append(highPriority, rec)
		case "medium":
			mediumPriority = append(mediumPriority, rec)
		case "low":
			lowPriority = append(lowPriority, rec)
		}
	}
	
	// High priority recommendations
	if len(highPriority) > 0 {
		content += "### High Priority Recommendations\n\n"
		for i, rec := range highPriority {
			content += fmt.Sprintf("#### %d. %s\n\n", i+1, rec.Type)
			content += fmt.Sprintf("- **Priority**: %s\n", rec.Priority)
			content += fmt.Sprintf("- **Description**: %s\n", rec.Description)
			content += fmt.Sprintf("- **Suggestion**: %s\n", rec.Suggestion)
			content += fmt.Sprintf("- **Impact**: %s\n", rec.Impact)
			content += fmt.Sprintf("- **Effort**: %s\n", rec.Effort)
			
			if len(rec.Steps) > 0 {
				content += "- **Steps**:\n"
				for _, step := range rec.Steps {
					content += fmt.Sprintf("  1. %s\n", step)
				}
			}
			
			content += "\n"
		}
	}
	
	// Medium priority recommendations
	if len(mediumPriority) > 0 {
		content += "### Medium Priority Recommendations\n\n"
		for i, rec := range mediumPriority {
			content += fmt.Sprintf("#### %d. %s\n\n", i+1, rec.Type)
			content += fmt.Sprintf("- **Priority**: %s\n", rec.Priority)
			content += fmt.Sprintf("- **Description**: %s\n", rec.Description)
			content += fmt.Sprintf("- **Suggestion**: %s\n", rec.Suggestion)
			content += fmt.Sprintf("- **Impact**: %s\n", rec.Impact)
			content += fmt.Sprintf("- **Effort**: %s\n", rec.Effort)
			content += "\n"
		}
	}
	
	// Low priority recommendations
	if len(lowPriority) > 0 {
		content += "### Low Priority Recommendations\n\n"
		for i, rec := range lowPriority {
			content += fmt.Sprintf("#### %d. %s\n\n", i+1, rec.Type)
			content += fmt.Sprintf("- **Priority**: %s\n", rec.Priority)
			content += fmt.Sprintf("- **Description**: %s\n", rec.Description)
			content += fmt.Sprintf("- **Suggestion**: %s\n", rec.Suggestion)
			content += fmt.Sprintf("- **Impact**: %s\n", rec.Impact)
			content += fmt.Sprintf("- **Effort**: %s\n", rec.Effort)
			content += "\n"
		}
	}
	
	return content
}