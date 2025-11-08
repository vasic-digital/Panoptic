package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"
	"panoptic/internal/platforms"
)

type Executor struct {
	config     *config.Config
	outputDir  string
	logger     *logger.Logger
	factory    *platforms.PlatformFactory
	results    []TestResult
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
	return &Executor{
		config:    cfg,
		outputDir: outputDir,
		logger:    log,
		factory:   platforms.NewPlatformFactory(),
		results:   make([]TestResult, 0),
	}
}

func (e *Executor) Run() error {
	e.logger.Info("Starting execution")
	e.logger.SetOutputDirectory(e.outputDir)
	
	// Validate configuration
	if err := e.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Execute tests for each application
	for _, app := range e.config.Apps {
		e.logger.Infof("Processing application: %s (%s)", app.Name, app.Type)
		
		result := e.executeApp(app)
		e.results = append(e.results, result)
		
		if result.Success {
			e.logger.Infof("Successfully completed app: %s", app.Name)
		} else {
			e.logger.Errorf("Failed app: %s - %s", app.Name, result.Error)
		}
	}
	
	e.logger.Info("Execution completed")
	return nil
}

func (e *Executor) executeApp(app config.AppConfig) TestResult {
	result := TestResult{
		AppName:     app.Name,
		AppType:     app.Type,
		StartTime:   time.Now(),
		Screenshots: make([]string, 0),
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
		
		if err := e.executeAction(platform, action, app.Name, &result, &currentRecordingFile); err != nil {
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
	
	return result
}

func (e *Executor) executeAction(platform platforms.Platform, action config.Action, appName string, result *TestResult, recordingFile *string) error {
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
		filename := filepath.Join(e.outputDir, "screenshots", fmt.Sprintf("%s_%s_%d.png", appName, action.Name, time.Now().Unix()))
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
		
		filename := filepath.Join(e.outputDir, "videos", fmt.Sprintf("%s_%s_%d.mp4", appName, action.Name, time.Now().Unix()))
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
		
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
	
	return nil
}

func (e *Executor) GenerateReport(outputPath string) error {
	// Generate HTML report
	html := e.generateHTMLReport()
	
	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}
	
	// Also generate JSON report for programmatic access
	jsonPath := filepath.Join(filepath.Dir(outputPath), "report.json")
	if err := e.generateJSONReport(jsonPath); err != nil {
		e.logger.Errorf("Failed to generate JSON report: %v", err)
	}
	
	return nil
}

func (e *Executor) generateHTMLReport() string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Panoptic Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .app-result { margin: 20px 0; border: 1px solid #ddd; border-radius: 5px; }
        .app-header { background: #e8e8e8; padding: 15px; font-weight: bold; }
        .app-content { padding: 15px; }
        .success { color: green; }
        .failure { color: red; }
        .metrics { margin: 10px 0; }
        .screenshot { max-width: 200px; margin: 5px; border: 1px solid #ccc; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Panoptic Test Report</h1>
        <p>Generated: ` + time.Now().Format(time.RFC3339) + `</p>
        <p>Total Applications: ` + fmt.Sprintf("%d", len(e.results)) + `</p>
    </div>
`

	for _, result := range e.results {
		status := "success"
		statusClass := "success"
		if !result.Success {
			status = "failed"
			statusClass = "failure"
		}
		
		html += `
    <div class="app-result">
        <div class="app-header">
            ` + result.AppName + ` (` + result.AppType + `) - <span class="` + statusClass + `">` + status + `</span>
        </div>
        <div class="app-content">
            <p><strong>Duration:</strong> ` + result.Duration.String() + `</p>
            <p><strong>Start Time:</strong> ` + result.StartTime.Format(time.RFC3339) + `</p>
            <p><strong>End Time:</strong> ` + result.EndTime.Format(time.RFC3339) + `</p>`
		
		if result.Error != "" {
			html += `<p><strong>Error:</strong> ` + result.Error + `</p>`
		}
		
		if len(result.Screenshots) > 0 {
			html += `<h4>Screenshots:</h4>`
			for _, screenshot := range result.Screenshots {
				html += `<img src="` + filepath.Base(screenshot) + `" class="screenshot" alt="Screenshot">`
			}
		}
		
		if len(result.Videos) > 0 {
			html += `<h4>Videos:</h4><ul>`
			for _, video := range result.Videos {
				html += `<li><a href="` + filepath.Base(video) + `">` + filepath.Base(video) + `</a></li>`
			}
			html += `</ul>`
		}
		
		if len(result.Metrics) > 0 {
			html += `<h4>Metrics:</h4><table>`
			for key, value := range result.Metrics {
				html += `<tr><td><strong>` + key + `</strong></td><td>` + fmt.Sprintf("%v", value) + `</td></tr>`
			}
			html += `</table>`
		}
		
		html += `</div></div>`
	}

	html += `
</body>
</html>`

	return html
}

func (e *Executor) generateJSONReport(outputPath string) error {
	// This would use a JSON library to create a structured report
	// For simplicity, creating a basic structure
	return nil
}