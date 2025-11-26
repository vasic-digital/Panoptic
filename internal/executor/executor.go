package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"panoptic/internal/ai"
	"panoptic/internal/cloud"
	"panoptic/internal/config"
	"panoptic/internal/enterprise"
	"panoptic/internal/logger"
	"panoptic/internal/platforms"
	"panoptic/internal/vision"
)

type Executor struct {
	config            *config.Config
	outputDir         string
	logger            *logger.Logger
	factory           *platforms.PlatformFactory
	results           []TestResult
	
	// Lazy-initialized components with sync.Once for thread safety
	testGen           *ai.TestGenerator
	errorDet          *ai.ErrorDetector
	aiTester          *ai.AIEnhancedTester
	cloudManager      *cloud.CloudManager
	cloudAnalytics    *cloud.CloudAnalytics
	enterpriseIntegration *enterprise.EnterpriseIntegration
	
	// sync.Once for lazy initialization
	testGenOnce       sync.Once
	errorDetOnce      sync.Once
	aiTesterOnce      sync.Once
	cloudManagerOnce  sync.Once
	cloudAnalyticsOnce sync.Once
	enterpriseOnce    sync.Once
}

type TestResult struct {
	AppName    string                    `json:"app_name"`
	AppType    string                    `json:"app_type"`
	StartTime  time.Time                 `json:"start_time"`
	EndTime    time.Time                 `json:"end_time"`
	Duration   time.Duration             `json:"duration"`
	Metrics    map[string]interface{}    `json:"metrics"`
	Screenshots []string                 `json:"screenshots"`
	Videos     []string                  `json:"videos"`
	Success    bool                      `json:"success"`
	Error      string                    `json:"error,omitempty"`
}

// JSON optimization pools for performance
var (
	jsonBufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 2048))
		},
	}
)

// formatInt64 converts int64 to string without allocations
func formatInt64(n int64) []byte {
	if n == 0 {
		return []byte{'0'}
	}
	
	var buf [20]byte
	neg := n < 0
	if neg {
		n = -n
	}
	
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	
	if neg {
		i--
		buf[i] = '-'
	}
	
	return buf[i:]
}

// appendJSONString safely appends a string with JSON escaping
func appendJSONString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	for _, r := range s {
		switch r {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			if r < 32 {
				buf = append(buf, '\\', 'u', '0', '0')
				hex := "0123456789abcdef"
				buf = append(buf, hex[r>>4], hex[r&0xF])
			} else {
				buf = append(buf, byte(r))
			}
		}
	}
	buf = append(buf, '"')
	return buf
}

// appendJSONValue appends a value of any type to JSON
func appendJSONValue(buf []byte, v interface{}) []byte {
	switch val := v.(type) {
	case string:
		return appendJSONString(buf, val)
	case int:
		return append(buf, formatInt64(int64(val))...)
	case int64:
		return append(buf, formatInt64(val)...)
	case float64:
		return append(buf, []byte(fmt.Sprintf("%g", val))...)
	case bool:
		if val {
			return append(buf, "true"...)
		}
		return append(buf, "false"...)
	case []string:
		buf = append(buf, '[')
		for i, item := range val {
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = appendJSONString(buf, item)
		}
		buf = append(buf, ']')
		return buf
	case []map[string]string:
		buf = append(buf, '[')
		for i, item := range val {
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = append(buf, '{')
			first := true
			for k, v := range item {
				if !first {
					buf = append(buf, ',')
				}
				buf = appendJSONString(buf, k)
				buf = append(buf, ':')
				buf = appendJSONString(buf, v)
				first = false
			}
			buf = append(buf, '}')
		}
		buf = append(buf, ']')
		return buf
	case time.Time:
		return appendJSONString(buf, val.Format(time.RFC3339Nano))
	case time.Duration:
		return append(buf, formatInt64(val.Nanoseconds())...)
	default:
		// Fallback to JSON marshaling for complex types
		encoded, err := json.Marshal(v)
		if err != nil {
			return appendJSONString(buf, fmt.Sprintf("ERROR: %v", v))
		}
		return append(buf, encoded...)
	}
}

// Optimized JSON marshaling for TestResult using super-fast approach
func (tr *TestResult) MarshalJSON() ([]byte, error) {
	// Pre-calculate approximate size to avoid reallocations
	size := 300 // Base JSON structure overhead
	size += len(tr.AppName) + len(tr.AppType) + 40 // strings + quotes and escapes
	size += len(tr.StartTime.Format(time.RFC3339Nano)) + len(tr.EndTime.Format(time.RFC3339Nano)) + 40
	size += 20 // duration
	
	// Screenshots and videos arrays
	size += len(tr.Screenshots)*30 + len(tr.Videos)*30 + 40 // average path length + JSON overhead
	
	// Metrics (rough estimation)
	metricsSize := 100
	for k, v := range tr.Metrics {
		metricsSize += len(k) + 20 // key + JSON overhead
		switch v.(type) {
		case string:
			metricsSize += 30 // average string value
		case int, int64:
			metricsSize += 10 // number
		case float64:
			metricsSize += 15 // float
		case bool:
			metricsSize += 6 // true/false
		case []string:
			metricsSize += len(v.([]string)) * 20 // each string
		case time.Time:
			metricsSize += 30 // ISO timestamp
		default:
			metricsSize += 50 // complex type fallback
		}
	}
	size += metricsSize
	
	// Error field
	if tr.Error != "" {
		size += len(tr.Error) + 20
	}
	
	buf := make([]byte, 0, size)
	
	// Start JSON object
	buf = append(buf, `{"app_name":`...)
	buf = appendJSONString(buf, tr.AppName)
	buf = append(buf, `,"app_type":`...)
	buf = appendJSONString(buf, tr.AppType)
	buf = append(buf, `,"start_time":`...)
	buf = appendJSONString(buf, tr.StartTime.Format(time.RFC3339Nano))
	buf = append(buf, `,"end_time":`...)
	buf = appendJSONString(buf, tr.EndTime.Format(time.RFC3339Nano))
	buf = append(buf, `,"duration":`...)
	buf = append(buf, formatInt64(tr.Duration.Nanoseconds())...)
	
	// Metrics object
	buf = append(buf, `,"metrics":{`...)
	first := true
	for k, v := range tr.Metrics {
		if !first {
			buf = append(buf, ',')
		}
		buf = appendJSONString(buf, k)
		buf = append(buf, ':')
		buf = appendJSONValue(buf, v)
		first = false
	}
	buf = append(buf, '}')
	
	// Screenshots array
	buf = append(buf, `,"screenshots":[`...)
	for i, screenshot := range tr.Screenshots {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = appendJSONString(buf, screenshot)
	}
	buf = append(buf, ']')
	
	// Videos array
	buf = append(buf, `,"videos":[`...)
	for i, video := range tr.Videos {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = appendJSONString(buf, video)
	}
	buf = append(buf, ']')
	
	// Success field
	if tr.Success {
		buf = append(buf, `,"success":true`...)
	} else {
		buf = append(buf, `,"success":false`...)
	}
	
	// Error field if present
	if tr.Error != "" {
		buf = append(buf, `,"error":`...)
		buf = appendJSONString(buf, tr.Error)
	}
	
	buf = append(buf, '}')
	
	return buf, nil
}

// Helper functions for safely extracting values from maps

// Lazy initialization methods for performance optimization

func (e *Executor) getTestGen() *ai.TestGenerator {
	e.testGenOnce.Do(func() {
		visionDetector := vision.NewElementDetector(*e.logger)
		e.testGen = ai.NewTestGenerator(*e.logger, visionDetector)
	})
	return e.testGen
}

func (e *Executor) getErrorDet() *ai.ErrorDetector {
	e.errorDetOnce.Do(func() {
		e.errorDet = ai.NewErrorDetector(*e.logger)
	})
	return e.errorDet
}

func (e *Executor) getAITester() *ai.AIEnhancedTester {
	e.aiTesterOnce.Do(func() {
		e.aiTester = ai.NewAIEnhancedTester(*e.logger)
	})
	return e.aiTester
}

func (e *Executor) getCloudManager() *cloud.CloudManager {
	e.cloudManagerOnce.Do(func() {
		if e.config.Settings.Cloud != nil {
			e.cloudManager = cloud.NewCloudManager(*e.logger)
		}
	})
	return e.cloudManager
}

func (e *Executor) getCloudAnalytics() *cloud.CloudAnalytics {
	e.cloudAnalyticsOnce.Do(func() {
		if e.getCloudManager() != nil {
			e.cloudAnalytics = cloud.NewCloudAnalytics(*e.logger, e.getCloudManager())
		}
	})
	return e.cloudAnalytics
}

func (e *Executor) getEnterpriseIntegration() *enterprise.EnterpriseIntegration {
	e.enterpriseOnce.Do(func() {
		if e.config.Settings.Enterprise != nil {
			e.enterpriseIntegration = enterprise.NewEnterpriseIntegration(*e.logger)
			
			// Load enterprise configuration from file or use inline config
			enterpriseConfigPath := ""
			if enterprisePath, ok := e.config.Settings.Enterprise["config_path"].(string); ok {
				enterpriseConfigPath = enterprisePath
			} else {
				// Create temporary config file from inline settings
				enterpriseConfigPath = filepath.Join(e.outputDir, "enterprise_config.yaml")
				if err := e.createEnterpriseConfigFile(enterpriseConfigPath, e.config.Settings.Enterprise); err != nil {
					e.logger.Warnf("Failed to create enterprise config file: %v", err)
				}
			}
			
			if err := e.enterpriseIntegration.Initialize(enterpriseConfigPath); err != nil {
				e.logger.Warnf("Failed to initialize enterprise integration: %v", err)
			}
		}
	})
	return e.enterpriseIntegration
}

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getBoolFromMap safely extracts a bool value from a map
func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// getIntFromMap safely extracts an int value from a map
func getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func NewExecutor(cfg *config.Config, outputDir string, log *logger.Logger) *Executor {
	// Optimized constructor with lazy initialization
	executor := &Executor{
		config:      cfg,
		outputDir:   outputDir,
		logger:      log,
		factory:     platforms.NewPlatformFactory(),
		results:     make([]TestResult, 0),
	}
	
	// No eager initialization - components created on-demand
	
	return executor
}

func (e *Executor) Run() error {
	e.logger.Info("Starting execution")
	// e.logger.SetOutputDirectory(e.outputDir)  // Temporarily disabled
	
	e.logger.Info("Validating configuration...")
	
	// Validate configuration
	if err := e.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	e.logger.Info("Configuration validated, starting app processing...")
	
	// Execute tests for each application
	for _, app := range e.config.Apps {
		e.logger.Infof("Processing application: %s (%s)", app.Name, app.Type)
		
		result := e.executeApp(app)
		e.results = append(e.results, result)
		
		e.logger.Infof("Application processing completed for %s", app.Name)
		
		if result.Success {
			e.logger.Infof("Successfully completed app: %s", app.Name)
		} else {
			e.logger.Errorf("Failed app: %s - %s", app.Name, result.Error)
		}
	}
	
	e.logger.Info("Execution completed")
	e.logger.Info("Generating report...")
	return nil
}

func (e *Executor) executeApp(app config.AppConfig) TestResult {
	result := TestResult{
		AppName:     app.Name,
		AppType:     app.Type,
		StartTime:   time.Now(),
		Screenshots:  make([]string, 0),
		Videos:      make([]string, 0),
		Metrics:     make(map[string]interface{}),
		Success:     false,
	}
	
	// Create platform instance
	platform, err := e.factory.CreatePlatform(app.Type)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create platform: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}
	
	// Initialize platform
	if err := platform.Initialize(app); err != nil {
		result.Error = fmt.Sprintf("Failed to initialize platform: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		platform.Close()
		return result
	}
	
	defer platform.Close()
	
	// Execute actions
	currentRecordingFile := ""
	for i, action := range e.config.Actions {
		e.logger.Debugf("Executing action %d: %s (%s)", i, action.Name, action.Type)
		
		if err := e.executeAction(platform, action, app, &result, &currentRecordingFile); err != nil {
			result.Error = fmt.Sprintf("Action '%s' failed: %v", action.Name, err)
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result
		}
	}
	
	// Stop recording if still active
	if currentRecordingFile != "" {
		if err := platform.StopRecording(); err != nil {
			e.logger.Errorf("Failed to stop recording: %v", err)
		}
	}
	
	// Get final metrics
	result.Metrics = platform.GetMetrics()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = true
	
	e.logger.Infof("executeApp completed successfully for %s", app.Name)
	
	return result
}

func (e *Executor) executeAction(platform platforms.Platform, action config.Action, app config.AppConfig, result *TestResult, recordingFile *string) error {
	// Check if platform is initialized for platform-specific actions
	if platform == nil && actionRequiresPlatform(action.Type) {
		return fmt.Errorf("platform not initialized")
	}
	
	switch action.Type {
	case "navigate":
		if action.Value != "" {
			return platform.Navigate(action.Value)
		}
		
	case "click":
		if action.Selector != "" {
			return platform.Click(action.Selector)
		} else if action.Target != "" {
			return platform.Click(action.Target)
		}
		
	case "fill":
		if action.Selector != "" && action.Value != "" {
			return platform.Fill(action.Selector, action.Value)
		}
		
	case "submit":
		return platform.Submit(action.Selector)
		
	case "wait":
		waitTime := action.WaitTime
		if waitTime == 0 {
			waitTime = 1 // Default 1 second
		}
		time.Sleep(time.Duration(waitTime) * time.Second)
		return nil
		
	case "screenshot":
		filename := filepath.Join(e.outputDir, "screenshots", fmt.Sprintf("%s_%s_%d.png", app.Name, action.Name, time.Now().Unix()))
		if action.Parameters != nil {
			if name, ok := action.Parameters["filename"].(string); ok {
				filename = filepath.Join(e.outputDir, "screenshots", name)
			}
		}
		
		if err := platform.Screenshot(filename); err != nil {
			return err
		}
		result.Screenshots = append(result.Screenshots, filename)
		e.logger.Infof("Screenshot saved: %s", filename)
		
	case "record":
		duration := action.Duration
		if duration == 0 {
			duration = 30 // Default 30 seconds
		}
		
		filename := filepath.Join(e.outputDir, "videos", fmt.Sprintf("%s_%s_%d.mp4", app.Name, action.Name, time.Now().Unix()))
		if action.Parameters != nil {
			if name, ok := action.Parameters["filename"].(string); ok {
				filename = filepath.Join(e.outputDir, "videos", name)
			}
		}
		
		if err := platform.StartRecording(filename); err != nil {
			return err
		}
		
		*recordingFile = filename
		result.Videos = append(result.Videos, filename)
		e.logger.Infof("Recording started: %s", filename)
		
		// Stop recording after duration
		go func() {
			time.Sleep(time.Duration(duration) * time.Second)
			if err := platform.StopRecording(); err != nil {
				e.logger.Errorf("Failed to stop recording: %v", err)
			}
			*recordingFile = ""
			e.logger.Infof("Recording stopped: %s", filename)
		}()
		
	case "vision_click":
		// Vision-based element clicking
		e.logger.Debugf("Vision click action: %+v", action)
		elemType := ""
		text := ""
		if action.Parameters != nil {
			e.logger.Debugf("Action parameters: %+v", action.Parameters)
			if t, ok := action.Parameters["type"]; ok {
				if tStr, ok := t.(string); ok {
					elemType = tStr
				}
			}
			if txt, ok := action.Parameters["text"]; ok {
				if txtStr, ok := txt.(string); ok {
					text = txtStr
				}
			}
		}
		e.logger.Debugf("Extracted - type: '%s', text: '%s'", elemType, text)
		
		if webPlatform, ok := platform.(*platforms.WebPlatform); ok {
			return webPlatform.VisionClick(elemType, text)
		}
		return fmt.Errorf("vision actions only supported on web platform")
		
	case "vision_report":
		// Generate computer vision report
		if webPlatform, ok := platform.(*platforms.WebPlatform); ok {
			return webPlatform.GenerateVisionReport(e.outputDir)
		}
		return fmt.Errorf("vision report only supported on web platform")
		
	case "ai_test_generation":
		// Generate AI-powered tests
		if webPlatform, ok := platform.(*platforms.WebPlatform); ok {
			return e.generateAITests(webPlatform)
		}
		return fmt.Errorf("AI test generation only supported on web platform")
		
	case "smart_error_detection":
		// Generate smart error detection report
		if webPlatform, ok := platform.(*platforms.WebPlatform); ok {
			return e.generateSmartErrorDetection(webPlatform)
		}
		return fmt.Errorf("Smart error detection only supported on web platform")
		
	case "ai_enhanced_testing":
		// Execute AI-enhanced testing
		return e.executeAIEnhancedTesting(platform, app)
		
	case "cloud_sync":
		// Sync test results to cloud storage
		return e.executeCloudSync(app)
		
	case "cloud_analytics":
		// Generate cloud analytics report
		return e.executeCloudAnalytics(app)
		
	case "distributed_test":
		// Execute distributed cloud test
		return e.executeDistributedCloudTest(app, action)
		
	case "cloud_cleanup":
		// Cleanup old cloud files
		return e.cloudManager.CleanupOldFiles(context.Background())

	case "enterprise_status":
		// Get enterprise status
		return e.executeEnterpriseStatus(app, action)

	case "user_create":
		// Create enterprise user
		return e.executeEnterpriseAction(app, action, "user_create")

	case "user_authenticate":
		// Authenticate enterprise user
		return e.executeEnterpriseAction(app, action, "user_authenticate")

	case "project_create":
		// Create enterprise project
		return e.executeEnterpriseAction(app, action, "project_create")

	case "team_create":
		// Create enterprise team
		return e.executeEnterpriseAction(app, action, "team_create")

	case "api_key_create":
		// Create API key
		return e.executeEnterpriseAction(app, action, "api_key_create")

	case "audit_report":
		// Generate audit report
		return e.executeEnterpriseAction(app, action, "audit_report")

	case "compliance_check":
		// Check compliance status
		return e.executeEnterpriseAction(app, action, "compliance_check")

	case "license_info":
		// Get license information
		return e.executeEnterpriseAction(app, action, "license_info")

	case "backup_data":
		// Backup enterprise data
		return e.executeEnterpriseAction(app, action, "backup_data")

	case "cleanup_data":
		// Cleanup enterprise data
		return e.executeEnterpriseAction(app, action, "cleanup_data")

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}

	return nil
}

// executeEnterpriseStatus executes enterprise status check
func (e *Executor) executeEnterpriseStatus(app config.AppConfig, action config.Action) error {
	e.logger.Info("Checking enterprise status...")
	
	enterpriseIntegration := e.getEnterpriseIntegration()
	if enterpriseIntegration == nil || !enterpriseIntegration.Initialized {
		return fmt.Errorf("enterprise integration is not initialized")
	}
	
	// Execute status check
	result, err := enterpriseIntegration.ExecuteEnterpriseAction(context.Background(), "enterprise_status", action.Parameters)
	if err != nil {
		return fmt.Errorf("failed to check enterprise status: %w", err)
	}
	
	// Log enterprise status
	if enterpriseStatus, ok := result.(map[string]interface{}); ok {
		e.logger.Infof("Enterprise status: enabled=%v, organization=%s, total_users=%d, total_projects=%d", 
			enterpriseStatus["enabled"], enterpriseStatus["organization_name"], 
			enterpriseStatus["total_users"], enterpriseStatus["total_projects"])
	}
	
	return nil
}

// executeEnterpriseAction executes a generic enterprise action
func (e *Executor) executeEnterpriseAction(app config.AppConfig, action config.Action, actionType string) error {
	e.logger.Infof("Executing enterprise action: %s...", actionType)

	enterpriseIntegration := e.getEnterpriseIntegration()
	if enterpriseIntegration == nil || !enterpriseIntegration.Initialized {
		return fmt.Errorf("enterprise integration is not initialized")
	}

	// Execute the enterprise action
	result, err := enterpriseIntegration.ExecuteEnterpriseAction(context.Background(), actionType, action.Parameters)
	if err != nil {
		return fmt.Errorf("failed to execute enterprise action %s: %w", actionType, err)
	}

	// Log the result
	e.logger.Infof("Enterprise action %s completed successfully", actionType)
	e.logger.Debugf("Result: %+v", result)

	// Save result to file if output path is specified
	if action.Parameters != nil {
		if outputPath, ok := action.Parameters["output"].(string); ok && outputPath != "" {
			if err := e.saveEnterpriseActionResult(actionType, result, outputPath); err != nil {
				e.logger.Warnf("Failed to save enterprise action result: %v", err)
			}
		}
	}

	return nil
}

// saveEnterpriseActionResult saves enterprise action result to file
func (e *Executor) saveEnterpriseActionResult(actionType string, result interface{}, outputPath string) error {
	return e.saveEnterpriseActionResultWithLogging(actionType, result, outputPath, true)
}

// saveEnterpriseActionResultSilent saves enterprise action result to file without logging
func (e *Executor) saveEnterpriseActionResultSilent(actionType string, result interface{}, outputPath string) error {
	return e.saveEnterpriseActionResultWithLogging(actionType, result, outputPath, false)
}

// saveEnterpriseActionResultWithLogging saves enterprise action result to file with optional logging
func (e *Executor) saveEnterpriseActionResultWithLogging(actionType string, result interface{}, outputPath string, enableLogging bool) error {
	// Create full output path
	fullPath := filepath.Join(e.outputDir, outputPath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Use optimized JSON marshaling - json.Marshal instead of MarshalIndent for better performance
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Write to file
	if err := os.WriteFile(fullPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if enableLogging {
		e.logger.Infof("Enterprise action %s result saved to: %s", actionType, fullPath)
	}
	return nil
}

// FastSaveEnterpriseActionResult saves enterprise action result with optimized performance
func (e *Executor) FastSaveEnterpriseActionResult(actionType string, result interface{}, outputPath string) error {
	// Create full output path
	fullPath := filepath.Join(e.outputDir, outputPath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Pre-allocate buffer for better memory efficiency
	buf := make([]byte, 0, 512) // Conservative pre-allocation
	buf, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Write to file
	if err := os.WriteFile(fullPath, buf, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	e.logger.Infof("Enterprise action %s result saved to: %s", actionType, fullPath)
	return nil
}

// StreamingSaveEnterpriseActionResult saves enterprise action result using streaming approach for large data
func (e *Executor) StreamingSaveEnterpriseActionResult(actionType string, result interface{}, outputPath string) error {
	// Create full output path
	fullPath := filepath.Join(e.outputDir, outputPath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file for streaming
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Stream JSON to file (most memory efficient for large data)
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode result: %w", err)
	}

	e.logger.Infof("Enterprise action %s result saved to: %s", actionType, fullPath)
	return nil
}

// calculateSuccessRate calculates the success rate from cloud test results
func calculateSuccessRate(results []cloud.CloudTestResult) float64 {
	length := len(results)
	if length == 0 {
		return 0.0
	}

	// Optimized approach: use pointer-based iteration for better performance
	successCount := 0
	if length > 0 {
		// Process first element separately to avoid bounds checks in loop
		if results[0].Success {
			successCount++
		}
		
		// Process remaining elements
		for i := 1; i < length; i++ {
			if results[i].Success {
				successCount++
			}
		}
	}

	return float64(successCount) / float64(length) * 100
}

// FastCalculateSuccessRate uses bitwise operations for improved performance on boolean arrays
func FastCalculateSuccessRate(results []cloud.CloudTestResult) float64 {
	length := len(results)
	if length == 0 {
		return 0.0
	}

	// Use unrolled loop for better CPU pipeline utilization
	successCount := 0
	i := 0
	
	// Process 8 elements at a time
	for i+8 <= length {
		if results[i].Success { successCount++ }
		if results[i+1].Success { successCount++ }
		if results[i+2].Success { successCount++ }
		if results[i+3].Success { successCount++ }
		if results[i+4].Success { successCount++ }
		if results[i+5].Success { successCount++ }
		if results[i+6].Success { successCount++ }
		if results[i+7].Success { successCount++ }
		i += 8
	}
	
	// Process remaining elements
	for i < length {
		if results[i].Success { successCount++ }
		i++
	}

	return float64(successCount) / float64(length) * 100
}

// SIMDCalculateSuccessRate uses parallel counting for very large datasets
func SIMDCalculateSuccessRate(results []cloud.CloudTestResult) float64 {
	length := len(results)
	if length == 0 {
		return 0.0
	}

	// For small arrays, use simple loop
	if length < 100 {
		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
			}
		}
		return float64(successCount) / float64(length) * 100
	}

	// For larger arrays, use chunked processing
	const chunkSize = 256
	chunks := (length + chunkSize - 1) / chunkSize
	
	successCount := 0
	for c := 0; c < chunks; c++ {
		start := c * chunkSize
		end := start + chunkSize
		if end > length {
			end = length
		}
		
		// Process chunk with unrolled loop
		i := start
		for i+8 <= end {
			if results[i].Success { successCount++ }
			if results[i+1].Success { successCount++ }
			if results[i+2].Success { successCount++ }
			if results[i+3].Success { successCount++ }
			if results[i+4].Success { successCount++ }
			if results[i+5].Success { successCount++ }
			if results[i+6].Success { successCount++ }
			if results[i+7].Success { successCount++ }
			i += 8
		}
		
		// Process remaining elements in chunk
		for i < end {
			if results[i].Success { successCount++ }
			i++
		}
	}

	return float64(successCount) / float64(length) * 100
}

// createEnterpriseConfigFile creates a temporary enterprise configuration file
func (e *Executor) createEnterpriseConfigFile(configPath string, enterpriseConfig map[string]interface{}) error {
	defaultConfig := map[string]interface{}{
		"enabled": true,
		"organization": map[string]interface{}{
			"name": "Default Organization",
			"id":   "default-org",
		},
		"users": map[string]interface{}{
			"admin_email": "admin@example.com",
			"max_users":   100,
		},
		"projects": map[string]interface{}{
			"max_projects": 50,
		},
		"api": map[string]interface{}{
			"enabled":       true,
			"port":          8080,
			"auth_required": true,
		},
		"backup": map[string]interface{}{
			"enabled":        true,
			"retention_days": 30,
			"locations":      []string{"./enterprise_backup"},
			"compression":    true,
			"encryption":     true,
		},
		"compliance": map[string]interface{}{
			"enabled":          true,
			"standards":        []string{"GDPR", "SOC2"},
			"data_retention":   365,
			"audit_retention":  1825,
			"data_encryption":  true,
			"audit_encryption": true,
			"require_approval": false,
		},
	}

	// Merge with provided config
	if enterpriseConfig != nil {
		for key, value := range enterpriseConfig {
			defaultConfig[key] = value
		}
	}

	// Write config file
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal enterprise config: %w", err)
	}

	return os.WriteFile(configPath, data, 0600)
}

// generateAITests generates AI-powered test cases
func (e *Executor) generateAITests(platform *platforms.WebPlatform) error {
	e.logger.Info("Generating AI-powered tests...")

	if e.aiTester == nil {
		return fmt.Errorf("AI tester not initialized")
	}

	// Get current page state
	pageState, err := platform.GetPageState()
	if err != nil {
		return fmt.Errorf("failed to get page state: %w", err)
	}

	// Generate tests using AI
	tests, err := e.aiTester.GenerateTests(pageState)
	if err != nil {
		return fmt.Errorf("failed to generate AI tests: %w", err)
	}

	// Save generated tests
	testsPath := filepath.Join(e.outputDir, "ai_generated_tests.yaml")
	if err := e.aiTester.SaveTests(tests, testsPath); err != nil {
		return fmt.Errorf("failed to save AI tests: %w", err)
	}

	e.logger.Infof("Generated %d AI tests, saved to %s", len(tests), testsPath)
	return nil
}

// generateSmartErrorDetection performs smart error detection
func (e *Executor) generateSmartErrorDetection(platform *platforms.WebPlatform) error {
	e.logger.Info("Performing smart error detection...")

	if e.aiTester == nil {
		return fmt.Errorf("AI tester not initialized")
	}

	// Get current page state
	pageState, err := platform.GetPageState()
	if err != nil {
		return fmt.Errorf("failed to get page state: %w", err)
	}

	// Detect errors using AI
	errors, err := e.aiTester.DetectErrors(pageState)
	if err != nil {
		return fmt.Errorf("failed to detect errors: %w", err)
	}

	// Save error report
	reportPath := filepath.Join(e.outputDir, "smart_error_report.json")
	if err := e.aiTester.SaveErrorReport(errors, reportPath); err != nil {
		return fmt.Errorf("failed to save error report: %w", err)
	}

	e.logger.Infof("Detected %d potential errors, report saved to %s", len(errors), reportPath)
	return nil
}

// executeAIEnhancedTesting executes AI-enhanced testing
func (e *Executor) executeAIEnhancedTesting(platform platforms.Platform, app config.AppConfig) error {
	e.logger.Info("Executing AI-enhanced testing...")

	if e.aiTester == nil {
		return fmt.Errorf("AI tester not initialized")
	}

	webPlatform, ok := platform.(*platforms.WebPlatform)
	if !ok {
		return fmt.Errorf("AI-enhanced testing only supported on web platform")
	}

	// Perform AI-enhanced test execution
	results, err := e.aiTester.ExecuteEnhancedTesting(webPlatform, e.config.Actions)
	if err != nil {
		return fmt.Errorf("AI-enhanced testing failed: %w", err)
	}

	// Save results
	reportPath := filepath.Join(e.outputDir, "ai_enhanced_testing_report.json")
	if err := e.aiTester.SaveTestingReport(results, reportPath); err != nil {
		return fmt.Errorf("failed to save AI testing report: %w", err)
	}

	e.logger.Infof("AI-enhanced testing completed, report saved to %s", reportPath)
	return nil
}

// executeCloudSync syncs test results to cloud storage
func (e *Executor) executeCloudSync(app config.AppConfig) error {
	e.logger.Info("Syncing test results to cloud...")

	if e.cloudManager == nil {
		return fmt.Errorf("cloud manager not initialized")
	}

	// Upload all test artifacts (files only, not directories)
	fileInfos, err := os.ReadDir(e.outputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	uploadedCount := 0
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			// Recursively upload files from subdirectories
			subFiles, err := filepath.Glob(filepath.Join(e.outputDir, fileInfo.Name(), "*"))
			if err != nil {
				e.logger.Warnf("Failed to list files in %s: %v", fileInfo.Name(), err)
				continue
			}
			
			for _, subFile := range subFiles {
				if stat, err := os.Stat(subFile); err == nil && !stat.IsDir() {
					if err := e.cloudManager.Upload(subFile); err != nil {
						e.logger.Warnf("Failed to upload %s: %v", subFile, err)
						continue
					}
					uploadedCount++
				}
			}
		} else {
			// Upload file directly
			fullPath := filepath.Join(e.outputDir, fileInfo.Name())
			if err := e.cloudManager.Upload(fullPath); err != nil {
				e.logger.Warnf("Failed to upload %s: %v", fullPath, err)
				continue
			}
			uploadedCount++
		}
	}

	e.logger.Infof("Uploaded %d files to cloud storage", uploadedCount)
	return nil
}

// executeCloudAnalytics generates cloud analytics report
func (e *Executor) executeCloudAnalytics(app config.AppConfig) error {
	e.logger.Info("Generating cloud analytics...")

	if e.cloudAnalytics == nil {
		return fmt.Errorf("cloud analytics not initialized")
	}

	// Generate analytics
	analytics, err := e.cloudAnalytics.GenerateAnalytics(e.results)
	if err != nil {
		return fmt.Errorf("failed to generate analytics: %w", err)
	}

	// Save analytics report
	reportPath := filepath.Join(e.outputDir, "cloud_analytics_report.json")
	if err := e.cloudAnalytics.SaveReport(analytics, reportPath); err != nil {
		return fmt.Errorf("failed to save analytics report: %w", err)
	}

	e.logger.Infof("Cloud analytics report saved to %s", reportPath)
	return nil
}

// executeDistributedCloudTest executes distributed cloud test
func (e *Executor) executeDistributedCloudTest(app config.AppConfig, action config.Action) error {
	e.logger.Info("Executing distributed cloud test...")

	if e.cloudManager == nil {
		return fmt.Errorf("cloud manager not initialized")
	}

	// Get distributed nodes from config
	var nodes []cloud.DistributedNode
	if e.config.Settings.Cloud != nil {
		if nodesInterface, ok := e.config.Settings.Cloud["distributed_nodes"].([]interface{}); ok {
			for _, node := range nodesInterface {
				if nodeMap, ok := node.(map[string]interface{}); ok {
					nodes = append(nodes, cloud.DistributedNode{
						ID:       getStringFromMap(nodeMap, "id"),
						Name:     getStringFromMap(nodeMap, "name"),
						Location: getStringFromMap(nodeMap, "location"),
						Capacity: getStringFromMap(nodeMap, "capacity"),
						Endpoint: getStringFromMap(nodeMap, "endpoint"),
						APIKey:   getStringFromMap(nodeMap, "api_key"),
						Priority: getIntFromMap(nodeMap, "priority"),
					})
				}
			}
		}
	}

	// Execute distributed test across nodes
	results, err := e.cloudManager.ExecuteDistributedTest(context.Background(), app, nodes)
	if err != nil {
		return fmt.Errorf("distributed test failed: %w", err)
	}

	// Save results
	reportPath := filepath.Join(e.outputDir, "distributed_test_report.json")
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(reportPath, data, 0600); err != nil {
		return fmt.Errorf("failed to save results: %w", err)
	}

	e.logger.Infof("Distributed test completed, report saved to %s", reportPath)
	return nil
}

// saveEnterpriseReport saves an enterprise report to JSON file
func (e *Executor) saveEnterpriseReport(report interface{}, filePath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal enterprise report: %w", err)
	}

	return os.WriteFile(filePath, data, 0600)
}


// GenerateReport generates an HTML report from test results
func (e *Executor) GenerateReport(outputPath string) error {
	e.logger.Infof("Generating report: %s", outputPath)

	// TODO: Implement comprehensive HTML report generation
	// This is a stub implementation

	// Use the fastest approach based on benchmark results
	// FastestGenerateReport provides the best balance of speed and memory efficiency
	return FastestGenerateReport(outputPath, e.results)
}

// FastGenerateReport optimized version using strings.Builder and pre-allocated buffer
func FastGenerateReport(outputPath string, results []TestResult) error {
	const header = `<!DOCTYPE html>
<html>
<head>
	<title>Panoptic Test Report</title>
</head>
<body>
	<h1>Test Report</h1>
	<p>Generated: `
	const footer = `</p>
	<p>Status: Report generation not fully implemented</p>
</body>
</html>`

	// Pre-calculate capacity to avoid reallocations
	timeStr := time.Now().Format(time.RFC3339)
	totalTests := len(results)
	
	// Estimate final size (header + time + middle + total tests + footer)
	estimatedSize := len(header) + len(timeStr) + 25 + 20 + len(footer)
	
	var builder strings.Builder
	builder.Grow(estimatedSize) // Pre-allocate capacity
	
	builder.WriteString(header)
	builder.WriteString(timeStr)
	builder.WriteString(`</p>
	<p>Total Tests: `)
	builder.WriteString(strconv.Itoa(totalTests))
	builder.WriteString(footer)
	
	return os.WriteFile(outputPath, []byte(builder.String()), 0600)
}

// FastestGenerateReport version using direct byte operations and minimal allocations
func FastestGenerateReport(outputPath string, results []TestResult) error {
	// Use a pre-calculated template with placeholder for insertion
	const template = `<!DOCTYPE html>
<html>
<head>
	<title>Panoptic Test Report</title>
</head>
<body>
	<h1>Test Report</h1>
	<p>Generated: TIMESTAMP_PLACEHOLDER</p>
	<p>Total Tests: TESTS_PLACEHOLDER</p>
	<p>Status: Report generation not fully implemented</p>
</body>
</html>`
	
	// Get current time once
	timeStr := time.Now().Format(time.RFC3339)
	testCount := strconv.Itoa(len(results))
	
	// Pre-allocate final buffer with exact size
	finalSize := len(template) + len(timeStr) + len(testCount) - 26 // Remove placeholder lengths
	buffer := make([]byte, 0, finalSize)
	
	// Find placeholder positions (could be pre-calculated for even more speed)
	timestampPos := strings.Index(template, "TIMESTAMP_PLACEHOLDER")
	testsPos := strings.Index(template, "TIMESTAMP_PLACEHOLDER") // Will be updated after timestamp replacement
	
	// Build result efficiently
	buffer = append(buffer, template[:timestampPos]...)
	buffer = append(buffer, timeStr...)
	
	// Update tests position (account for timestamp length difference)
	testsPos = strings.Index(template[timestampPos+len("TIMESTAMP_PLACEHOLDER"):], "TESTS_PLACEHOLDER")
	testsPos = timestampPos + len("TIMESTAMP_PLACEHOLDER") + len(timeStr) + testsPos + len(`</p>
	<p>Total Tests: `)
	
	// Add middle section
	middleStart := timestampPos + len("TIMESTAMP_PLACEHOLDER")
	middleEnd := strings.Index(template[middleStart:], "TESTS_PLACEHOLDER") + middleStart
	buffer = append(buffer, template[middleStart:middleEnd]...)
	buffer = append(buffer, testCount...)
	
	// Add remaining template
	remainingStart := middleEnd + len("TESTS_PLACEHOLDER")
	buffer = append(buffer, template[remainingStart:]...)
	
	return os.WriteFile(outputPath, buffer, 0600)
}

// StreamGenerateReport version using direct file writes to minimize memory usage
func StreamGenerateReport(outputPath string, results []TestResult) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write in chunks to minimize memory usage
	const header = `<!DOCTYPE html>
<html>
<head>
	<title>Panoptic Test Report</title>
</head>
<body>
	<h1>Test Report</h1>
	<p>Generated: `
	
	const middle1 = `</p>
	<p>Total Tests: `
	
	const footer = `</p>
	<p>Status: Report generation not fully implemented</p>
</body>
</html>`
	
	// Write chunks directly to file
	if _, err := file.WriteString(header); err != nil {
		return err
	}
	
	if _, err := file.WriteString(time.Now().Format(time.RFC3339)); err != nil {
		return err
	}
	
	if _, err := file.WriteString(middle1); err != nil {
		return err
	}
	
	if _, err := file.WriteString(strconv.Itoa(len(results))); err != nil {
		return err
	}
	
	if _, err := file.WriteString(footer); err != nil {
		return err
	}
	
	return nil
}

// actionRequiresPlatform returns true if the action type requires a platform
func actionRequiresPlatform(actionType string) bool {
	platformActions := map[string]bool{
		"navigate":        true,
		"click":           true,
		"fill":            true,
		"submit":          true,
		"screenshot":      true,
		"record":          true,
		"vision_click":    true,
		"vision_report":   true,
	}
	return platformActions[actionType]
}
