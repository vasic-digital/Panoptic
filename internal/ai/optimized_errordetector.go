package ai

import (
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
	"panoptic/internal/logger"
)

var (
	// Global caches for expensive operations
	errorPatternsCache []ErrorPattern
	errorPatternsOnce  sync.Once
	
	errorIndicatorsCache []string
	errorIndicatorsOnce  sync.Once
)

// initializeOptimizedErrorPatterns sets up error detection patterns with caching
func initializeOptimizedErrorPatterns() []ErrorPattern {
	errorPatternsOnce.Do(func() {
		patterns := []ErrorPattern{
			// Network/Connection Errors
			{
				Name:        "NetworkTimeout",
				Category:    "network",
				Pattern:     mustCompileRegex(`(?i)(timeout|timed out|connection timed|network timeout|request timeout)`),
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
				Pattern:     mustCompileRegex(`(?i)(connection refused|conn refused|cannot connect|unable to connect)`),
				Severity:    "high",
				Confidence:  0.90,
				Description: "Server connection refused",
				Suggestions: []string{
					"Check server status",
					"Verify server is running",
					"Check firewall settings",
					"Verify network connectivity",
				},
				Tags: []string{"network", "connection", "refused"},
			},
			{
				Name:        "SSLCertificateError",
				Category:    "network",
				Pattern:     mustCompileRegex(`(?i)(ssl|tls|certificate|cert.*error|certificate.*expired|certificate.*invalid)`),
				Severity:    "high",
				Confidence:  0.80,
				Description: "SSL/TLS certificate error",
				Suggestions: []string{
					"Update certificate",
					"Check certificate validity",
					"Verify certificate chain",
					"Check system date/time",
				},
				Tags: []string{"network", "ssl", "certificate"},
			},
			
			// HTTP/Protocol Errors
			{
				Name:        "HTTPError",
				Category:    "http",
				Pattern:     mustCompileRegex(`(?i)(http error|status code|4\d\d|5\d\d|bad request|unauthorized|forbidden|not found)`),
				Severity:    "medium",
				Confidence:  0.75,
				Description: "HTTP protocol error",
				Suggestions: []string{
					"Check HTTP status codes",
					"Verify request format",
					"Check authentication",
					"Validate API endpoints",
				},
				Tags: []string{"http", "protocol", "status"},
			},
			
			// Authentication/Authorization Errors
			{
				Name:        "AuthenticationError",
				Category:    "auth",
				Pattern:     mustCompileRegex(`(?i)(auth.*failed|login.*failed|unauthorized|access denied|permission denied|invalid credentials)`),
				Severity:    "high",
				Confidence:  0.85,
				Description: "Authentication or authorization failed",
				Suggestions: []string{
					"Verify credentials",
					"Check user permissions",
					"Ensure valid session",
					"Check token validity",
				},
				Tags: []string{"auth", "login", "permission"},
			},
			
			// Form/Input Validation Errors
			{
				Name:        "ValidationError",
				Category:    "validation",
				Pattern:     mustCompileRegex(`(?i)(validation error|invalid input|field.*required|format.*invalid)`),
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
				Pattern:     mustCompileRegex(`(?i)(required field|field.*required|missing.*field|field.*missing)`),
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
				Pattern:     mustCompileRegex(`(?i)(page.*timeout|load timeout|page not loading|slow page)`),
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
				Pattern:     mustCompileRegex(`(?i)(element.*timeout|element.*not loaded|wait.*timeout)`),
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
				Pattern:     mustCompileRegex(`(?i)(javascript error|script error|js error|runtime error)`),
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
				Pattern:     mustCompileRegex(`(?i)(undefined|not defined|null.*reference|object.*null)`),
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
				Pattern:     mustCompileRegex(`(?i)(database error|sql error|connection.*failed|query.*failed)`),
				Severity:    "high",
				Confidence:  0.85,
				Description: "Database operation failed",
				Suggestions: []string{
					"Check database connection",
					"Verify query syntax",
					"Check database permissions",
					"Review database logs",
				},
				Tags: []string{"database", "sql", "connection"},
			},
			
			// File System Errors
			{
				Name:        "FileNotFoundError",
				Category:    "filesystem",
				Pattern:     mustCompileRegex(`(?i)(file not found|no such file|file.*missing|cannot find file)`),
				Severity:    "medium",
				Confidence:  0.80,
				Description: "File or directory not found",
				Suggestions: []string{
					"Check file path",
					"Verify file exists",
					"Check file permissions",
					"Use absolute paths",
				},
				Tags: []string{"filesystem", "file", "path"},
			},
		}
		
		errorPatternsCache = patterns
	})
	
	return errorPatternsCache
}

// mustCompileRegex is a helper that compiles regex and panics on error
func mustCompileRegex(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	return re
}

// initializeErrorIndicators sets up common error indicators with caching
func initializeErrorIndicators() []string {
	errorIndicatorsOnce.Do(func() {
		indicators := []string{
			"error", "failed", "failure", "exception", "fault",
			"crash", "panic", "abort", "terminate", "unable",
			"cannot", "could not", "not possible", "invalid",
			"incorrect", "wrong", "bad", "malformed",
		}
		
		errorIndicatorsCache = indicators
	})
	
	return errorIndicatorsCache
}

// OptimizedErrorDetector provides a memory-efficient error detector
type OptimizedErrorDetector struct {
	logger     logger.Logger
	enabled    bool
	patterns   []ErrorPattern
	indicators []string
}

// NewOptimizedErrorDetector creates a new optimized error detector
func NewOptimizedErrorDetector(log logger.Logger) *OptimizedErrorDetector {
	detector := &OptimizedErrorDetector{
		logger:     log,
		enabled:    true,
		patterns:   initializeOptimizedErrorPatterns(),
		indicators: initializeErrorIndicators(),
	}
	return detector
}

// IsEnabled returns whether the detector is enabled
func (ed *OptimizedErrorDetector) IsEnabled() bool {
	return ed.enabled
}

// Enable enables the detector
func (ed *OptimizedErrorDetector) Enable() {
	ed.enabled = true
}

// Disable disables the detector
func (ed *OptimizedErrorDetector) Disable() {
	ed.enabled = false
}

// DetectErrors analyzes content for known error patterns
func (ed *OptimizedErrorDetector) DetectErrors(content string) []DetectedError {
	if !ed.enabled {
		return []DetectedError{}
	}

	var detections []DetectedError

	// Check each error pattern
	for _, pattern := range ed.patterns {
		if matches := pattern.Pattern.FindAllString(content, -1); len(matches) > 0 {
			// Calculate confidence based on multiple factors
			confidence := pattern.Confidence
			
			// Boost confidence for explicit error indicators
			for _, indicator := range ed.indicators {
				if strings.Contains(strings.ToLower(content), strings.ToLower(indicator)) {
					confidence = math.Min(confidence+0.1, 1.0)
					break
				}
			}

			// Create a detection for each match
			for _, match := range matches {
				detection := DetectedError{
					Name:        pattern.Name,
					Category:    pattern.Category,
					Message:     match,
					Severity:    pattern.Severity,
					Confidence:  confidence,
					Timestamp:   time.Now(),
					Source:      "content_analysis",
					Position:    ErrorPosition{Type: "content", Value: "content"},
					Suggestions: pattern.Suggestions,
					Tags:        pattern.Tags,
					Context:     map[string]string{"pattern": pattern.Name},
				}

				detections = append(detections, detection)
			}
		}
	}

	return detections
}

// GetErrorPatterns returns the current error patterns
func (ed *OptimizedErrorDetector) GetErrorPatterns() []ErrorPattern {
	return ed.patterns
}

// AddPattern adds a new error pattern
func (ed *OptimizedErrorDetector) AddPattern(pattern ErrorPattern) {
	ed.patterns = append(ed.patterns, pattern)
}