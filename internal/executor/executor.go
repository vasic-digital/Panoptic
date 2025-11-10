package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	testGen           *ai.TestGenerator
	errorDet          *ai.ErrorDetector
	aiTester          *ai.AIEnhancedTester
	cloudManager      *cloud.CloudManager
	cloudAnalytics    *cloud.CloudAnalytics
	enterpriseIntegration *enterprise.EnterpriseIntegration
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

func NewExecutor(cfg *config.Config, outputDir string, log *logger.Logger) *Executor {
	cloudManager := cloud.NewCloudManager(*log)
	cloudAnalytics := cloud.NewCloudAnalytics(*log, cloudManager)
	enterpriseIntegration := enterprise.NewEnterpriseIntegration(*log)
	
	executor := &Executor{
		config:               cfg,
		outputDir:            outputDir,
		logger:               log,
		factory:              platforms.NewPlatformFactory(),
		results:              make([]TestResult, 0),
		aiTester:             ai.NewAIEnhancedTester(*log),
		cloudManager:         cloudManager,
		cloudAnalytics:       cloudAnalytics,
		enterpriseIntegration: enterpriseIntegration,
	}
	
	// Initialize enterprise manager early if enterprise settings exist
	if cfg.Settings.Enterprise != nil {
		// Load enterprise configuration from file or use inline config
		enterpriseConfigPath := ""
		if enterprisePath, ok := cfg.Settings.Enterprise["config_path"].(string); ok {
			enterpriseConfigPath = enterprisePath
		} else {
			// Create temporary config file from inline settings
			enterpriseConfigPath = filepath.Join(outputDir, "enterprise_config.yaml")
			if err := executor.createEnterpriseConfigFile(enterpriseConfigPath, cfg.Settings.Enterprise); err != nil {
				executor.logger.Warnf("Failed to create enterprise config file: %v", err)
			}
		}
		
		if err := executor.enterpriseIntegration.Initialize(enterpriseConfigPath); err != nil {
			executor.logger.Warnf("Failed to initialize enterprise integration: %v", err)
		}
	}
	
	// Configure cloud manager early if cloud settings exist
	if cfg.Settings.Cloud != nil {
		// Convert map to CloudConfig
		cloudConfig := cloud.CloudConfig{
			Provider: getStringFromMap(cfg.Settings.Cloud, "provider"),
			Bucket:    getStringFromMap(cfg.Settings.Cloud, "bucket"),
			Region:    getStringFromMap(cfg.Settings.Cloud, "region"),
			AccessKey: getStringFromMap(cfg.Settings.Cloud, "access_key"),
			SecretKey: getStringFromMap(cfg.Settings.Cloud, "secret_key"),
			Endpoint:  getStringFromMap(cfg.Settings.Cloud, "endpoint"),
			EnableSync:        getBoolFromMap(cfg.Settings.Cloud, "enable_sync"),
			SyncInterval:      getIntFromMap(cfg.Settings.Cloud, "sync_interval"),
			EnableCDN:        getBoolFromMap(cfg.Settings.Cloud, "enable_cdn"),
			CDNEndpoint:       getStringFromMap(cfg.Settings.Cloud, "cdn_endpoint"),
			Compression:      getBoolFromMap(cfg.Settings.Cloud, "compression"),
			Encryption:       getBoolFromMap(cfg.Settings.Cloud, "encryption"),
			EnableDistributed: getBoolFromMap(cfg.Settings.Cloud, "enable_distributed"),
		}
		
		// Handle retention policy
		if retentionMap, ok := cfg.Settings.Cloud["retention_policy"].(map[string]interface{}); ok {
			cloudConfig.RetentionPolicy = cloud.RetentionPolicy{
				Enabled:     getBoolFromMap(retentionMap, "enabled"),
				Days:        getIntFromMap(retentionMap, "days"),
				MaxSizeGB:   getIntFromMap(retentionMap, "max_size_gb"),
				AutoCleanup: getBoolFromMap(retentionMap, "auto_cleanup"),
			}
		}
		
		// Handle backup locations
		if backupLocations, ok := cfg.Settings.Cloud["backup_locations"].([]interface{}); ok {
			for _, location := range backupLocations {
				if locationStr, ok := location.(string); ok {
					cloudConfig.BackupLocations = append(cloudConfig.BackupLocations, locationStr)
				}
			}
		}
		
		// Handle distributed nodes
		if nodesInterface, ok := cfg.Settings.Cloud["distributed_nodes"].([]interface{}); ok {
			for _, node := range nodesInterface {
				if nodeMap, ok := node.(map[string]interface{}); ok {
					node := cloud.DistributedNode{
						ID:       getStringFromMap(nodeMap, "id"),
						Name:     getStringFromMap(nodeMap, "name"),
						Location: getStringFromMap(nodeMap, "location"),
						Capacity: getStringFromMap(nodeMap, "capacity"),
						Endpoint: getStringFromMap(nodeMap, "endpoint"),
						APIKey:   getStringFromMap(nodeMap, "api_key"),
						Priority: getIntFromMap(nodeMap, "priority"),
					}
					cloudConfig.DistributedNodes = append(cloudConfig.DistributedNodes, node)
				}
			}
		}
		
		cloudManager.Configure(cloudConfig)
	}
	
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
		return platform.Wait(waitTime)
		
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
		
	case "enterprise_status":
		// Get enterprise status
		return e.executeEnterpriseStatus(app, action)
		
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
}

// executeEnterpriseStatus executes enterprise status check

// executeEnterpriseStatus executes enterprise status check
func (e *Executor) executeEnterpriseStatus(app config.AppConfig, action config.Action) error {
	e.logger.Info("Checking enterprise status...")
	
	if !e.enterpriseIntegration.Initialized {
		return fmt.Errorf("enterprise integration is not initialized")
	}
	
	// Execute status check
	result, err := e.enterpriseIntegration.ExecuteEnterpriseAction(context.Background(), "enterprise_status", action.Parameters)
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

func calculateSuccessRate(results []cloud.CloudTestResult) float64 {
			"retention_days": 30,
			"locations":      []string{"./enterprise_backup"},
			"compression":    true,
			"encryption":     true,
		},
		"compliance": map[string]interface{}{
			"enabled":           true,
			"standards":         []string{"GDPR", "SOC2"},
			"data_retention":     365,
			"audit_retention":    1825,
			"data_encryption":    true,
			"audit_encryption":   true,
			"require_approval":   false,
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

// saveEnterpriseReport saves an enterprise report to JSON file
func (e *Executor) saveEnterpriseReport(report interface{}, filePath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal enterprise report: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

