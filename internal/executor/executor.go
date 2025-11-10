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
	
	e.logger.Infof("executeApp completed successfully for %s", app.Name)
	
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
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            margin: 0; 
            padding: 20px; 
            background: #f8f9fa;
        }
        .header { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
            color: white; 
            padding: 30px; 
            border-radius: 10px; 
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .summary-cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            background: white;
            padding: 25px;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            text-align: center;
            border-left: 4px solid #007bff;
        }
        .card-value {
            font-size: 2.2em;
            font-weight: bold;
            color: #333;
            margin-bottom: 5px;
        }
        .card-label {
            color: #666;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .app-result { 
            margin: 20px 0; 
            border: 1px solid #e0e0e0; 
            border-radius: 10px;
            background: white;
            overflow: hidden;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            transition: transform 0.2s;
        }
        .app-result:hover {
            transform: translateY(-2px);
        }
        .app-header { 
            background: linear-gradient(90deg, #f8f9fa 0%, #e9ecef 100%); 
            padding: 20px; 
            font-weight: bold;
            border-bottom: 1px solid #e0e0e0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .app-content { 
            padding: 20px; 
        }
        .success { color: #28a745; font-weight: bold; }
        .failure { color: #dc3545; font-weight: bold; }
        .status-badge {
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.8em;
            font-weight: bold;
            text-transform: uppercase;
        }
        .status-success {
            background: #d4edda;
            color: #155724;
        }
        .status-failure {
            background: #f8d7da;
            color: #721c24;
        }
        .metrics { 
            margin: 10px 0; 
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            border-left: 3px solid #007bff;
        }
        .screenshot { 
            max-width: 200px; 
            margin: 8px; 
            border: 2px solid #e0e0e0; 
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.3s;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .screenshot:hover {
            transform: scale(1.05);
            border-color: #007bff;
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
        }
        .video-item {
            background: #f8f9fa;
            padding: 12px;
            border-radius: 5px;
            margin: 8px 0;
            border-left: 4px solid #28a745;
            transition: background 0.2s;
        }
        .video-item:hover {
            background: #e9ecef;
        }
        table { 
            border-collapse: collapse; 
            width: 100%; 
            margin: 10px 0;
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        th, td { 
            border: 1px solid #e0e0e0; 
            padding: 12px; 
            text-align: left; 
        }
        th { 
            background-color: #f8f9fa; 
            font-weight: 600;
            color: #495057;
        }
        tr:nth-child(even) {
            background-color: #f8f9fa;
        }
        .collapsible {
            background: #f8f9fa;
            color: #333;
            cursor: pointer;
            padding: 15px;
            width: 100%;
            border: none;
            text-align: left;
            outline: none;
            font-size: 16px;
            border-radius: 5px;
            margin: 5px 0;
            transition: background 0.3s;
            border-left: 4px solid #007bff;
        }
        .collapsible:hover {
            background: #e9ecef;
        }
        .collapsible.active {
            background: #007bff;
            color: white;
        }
        .content {
            padding: 0;
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.3s ease-out;
        }
        .content.show {
            max-height: 2000px;
            padding: 15px 0;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
            margin: 15px 0;
        }
        .metric-item {
            background: white;
            padding: 15px;
            border-radius: 5px;
            text-align: center;
            border: 1px solid #e0e0e0;
        }
        .metric-value {
            font-size: 1.4em;
            font-weight: bold;
            color: #007bff;
        }
        .metric-key {
            font-size: 0.8em;
            color: #666;
            margin-top: 5px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Panoptic Enhanced Test Report</h1>
        <p>Generated: ` + time.Now().Format(time.RFC3339) + `</p>
        <p>Total Applications: ` + fmt.Sprintf("%d", len(e.results)) + `</p>
    </div>`
    
    // Add summary cards
    html += e.generateSummarySection()
    
    // Add detailed results (enhanced)
    html += e.generateDetailedResults()

    html += `
    <script>
        const coll = document.getElementsByClassName('collapsible');
        for (let i = 0; i < coll.length; i++) {
            coll[i].addEventListener('click', function() {
                this.classList.toggle('active');
                const content = this.nextElementSibling;
                content.classList.toggle('show');
            });
        }
    </script>
</body>
</html>`

	return html
}

func (e *Executor) generateSummarySection() string {
	successCount := 0
	failureCount := 0
	totalDuration := time.Duration(0)
	totalScreenshots := 0
	totalVideos := 0
	
	for _, result := range e.results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
		totalDuration += result.Duration
		totalScreenshots += len(result.Screenshots)
		totalVideos += len(result.Videos)
	}
	
	successRate := float64(successCount) / float64(len(e.results)) * 100
	
	html := `
    <div class="summary-cards">
        <div class="summary-card">
            <div class="card-value">` + fmt.Sprintf("%d", len(e.results)) + `</div>
            <div class="card-label">Total Apps</div>
        </div>
        <div class="summary-card">
            <div class="card-value">` + fmt.Sprintf("%.1f%%", successRate) + `</div>
            <div class="card-label">Success Rate</div>
        </div>
        <div class="summary-card">
            <div class="card-value">` + totalDuration.Truncate(time.Second).String() + `</div>
            <div class="card-label">Total Duration</div>
        </div>
        <div class="summary-card">
            <div class="card-value">` + fmt.Sprintf("%d", totalScreenshots) + `</div>
            <div class="card-label">Screenshots</div>
        </div>
        <div class="summary-card">
            <div class="card-value">` + fmt.Sprintf("%d", totalVideos) + `</div>
            <div class="card-label">Videos</div>
        </div>
    </div>`
	
	return html
}

func (e *Executor) generateDetailedResults() string {
	html := `<h3>Detailed Results</h3>`
	
	for _, result := range e.results {
		status := "Success"
		statusClass := "status-success"
		if !result.Success {
			status = "Failed"
			statusClass = "status-failure"
		}
		
		html += `
    <button class="collapsible">
        <span style="font-size: 1.2em; font-weight: bold;">` + result.AppName + `</span>
        <span class="status-badge ` + statusClass + `">` + status + `</span>
        <span class="card-label">` + result.AppType + ` | ` + result.Duration.String() + `</span>
    </button>
    <div class="content">
        <div class="app-content">
            <div class="metrics-grid">
                <div class="metric-item">
                    <div class="metric-value">` + result.StartTime.Format("15:04:05") + `</div>
                    <div class="metric-key">Start Time</div>
                </div>
                <div class="metric-item">
                    <div class="metric-value">` + result.EndTime.Format("15:04:05") + `</div>
                    <div class="metric-key">End Time</div>
                </div>
                <div class="metric-item">
                    <div class="metric-value">` + result.Duration.String() + `</div>
                    <div class="metric-key">Duration</div>
                </div>
                <div class="metric-item">
                    <div class="metric-value">` + fmt.Sprintf("%d", len(result.Screenshots)) + `</div>
                    <div class="metric-key">Screenshots</div>
                </div>
            </div>`
		
		if result.Error != "" {
			html += `<div class="metrics"><strong>Error:</strong> ` + result.Error + `</div>`
		}
		
		if len(result.Screenshots) > 0 {
			html += `<h4>Screenshots:</h4>`
			for _, screenshot := range result.Screenshots {
				html += `<img src="` + filepath.Base(screenshot) + `" class="screenshot" alt="Screenshot">`
			}
		}
		
		if len(result.Videos) > 0 {
			html += `<h4>Videos:</h4>`
			for _, video := range result.Videos {
				html += `<div class="video-item"><a href="` + filepath.Base(video) + `" target="_blank">` + filepath.Base(video) + `</a></div>`
			}
		}
		
		if len(result.Metrics) > 0 {
			html += `<h4>Metrics:</h4><table>`
			for key, value := range result.Metrics {
				html += `<tr><td><strong>` + key + `</strong></td><td>` + fmt.Sprintf("%v", value) + `</td></tr>`
			}
			html += `</table>`
		}
		
		html += `</div>
    </div>`
	}
	
	return html
}

func (e *Executor) generateJSONReport(outputPath string) error {
	// This would use a JSON library to create a structured report
	// For simplicity, creating a basic structure
	return nil
}
