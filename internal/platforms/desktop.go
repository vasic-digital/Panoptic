package platforms

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"panoptic/internal/config"
)

type DesktopPlatform struct {
	appPath   string
	process   *os.Process
	metrics   map[string]interface{}
	recording bool
}

func NewDesktopPlatform() *DesktopPlatform {
	return &DesktopPlatform{
		metrics: map[string]interface{}{
			"click_actions":     []string{},
			"screenshots_taken":  []string{},
			"fill_actions":      []map[string]string{},
			"submit_actions":    []string{},
			"navigate_actions":  []string{},
			"start_time":        time.Now(),
		},
	}
}

func (d *DesktopPlatform) Initialize(app config.AppConfig) error {
	d.metrics["start_time"] = time.Now()
	d.appPath = app.Path
	
	// Check if application exists
	if _, err := os.Stat(app.Path); os.IsNotExist(err) {
		return fmt.Errorf("application not found at path: %s", app.Path)
	}
	
	d.metrics["app_path"] = app.Path
	return nil
}

func (d *DesktopPlatform) Navigate(url string) error {
	// Input validation
	if url == "" {
		return fmt.Errorf("url cannot be empty")
	}
	
	// For desktop apps, navigate might mean opening a specific view or menu
	// This is a placeholder implementation
	if navigateActions, ok := d.metrics["navigate_actions"].([]string); ok {
		d.metrics["navigate_actions"] = append(navigateActions, url)
	}
	return nil
}

func (d *DesktopPlatform) Click(selector string) error {
	// Input validation
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}
	
	// Safe slice append
	if clickActions, ok := d.metrics["click_actions"].([]string); ok {
		d.metrics["click_actions"] = append(clickActions, selector)
	}
	
	// This would require platform-specific automation (e.g., AppleScript on macOS, AutoHotkey on Windows)
	// For now, this is a placeholder
	time.Sleep(1 * time.Second) // Simulate click action
	return nil
}

func (d *DesktopPlatform) Fill(selector, value string) error {
	// Input validation
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	
	// Safe slice append
	if fillActions, ok := d.metrics["fill_actions"].([]map[string]string); ok {
		newAction := map[string]string{
			"selector": selector,
			"value":    value,
		}
		d.metrics["fill_actions"] = append(fillActions, newAction)
	}
	
	// Platform-specific form filling implementation
	time.Sleep(500 * time.Millisecond) // Simulate typing
	return nil
}

func (d *DesktopPlatform) Submit(selector string) error {
	// Safe slice append
	if submitActions, ok := d.metrics["submit_actions"].([]string); ok {
		d.metrics["submit_actions"] = append(submitActions, selector)
	}
	
	// Platform-specific form submission
	time.Sleep(1 * time.Second) // Simulate submission
	return nil
}

func (d *DesktopPlatform) Wait(duration int) error {
	time.Sleep(time.Duration(duration) * time.Second)
	return nil
}

func (d *DesktopPlatform) Screenshot(filename string) error {
	// Input validation
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	
	// Platform-specific screenshot implementation
	// macOS: screencapture, Windows: screenshot utilities, Linux: import (ImageMagick)
	var cmd *exec.Cmd
	
	switch {
	case runtime.GOOS == "darwin":
		cmd = exec.Command("screencapture", "-x", filename)
	case runtime.GOOS == "windows":
		// For Windows, you might use PowerShell or other tools
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait('{PRTSC}'); (Get-Clipboard -Format Image).Save('%s')", filename))
	default: // Linux and others
		cmd = exec.Command("import", "-window", "root", filename)
	}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}
	
	// Safe slice append
	if screenshotsTaken, ok := d.metrics["screenshots_taken"].([]string); ok {
		d.metrics["screenshots_taken"] = append(screenshotsTaken, filename)
	}
	
	return nil
}

func (d *DesktopPlatform) StartRecording(filename string) error {
	// Platform-specific screen recording
	// macOS: screencapture -a, Windows: other tools, Linux: ffmpeg
	var cmd *exec.Cmd
	
	switch {
	case runtime.GOOS == "darwin":
		cmd = exec.Command("screencapture", "-v", "-R", "0,0,1920,1080", filename)
	default:
		// Placeholder for other platforms
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create video file: %w", err)
		}
		file.Close()
		return nil
	}
	
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create video directory: %w", err)
	}
	
	if cmd != nil {
		// Start recording in background
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start recording: %w", err)
		}
	}
	
	d.recording = true
	d.metrics["recording_started"] = time.Now()
	d.metrics["recording_file"] = filename
	
	return nil
}

func (d *DesktopPlatform) StopRecording() error {
	if d.recording {
		d.recording = false
		d.metrics["recording_stopped"] = time.Now()
		d.metrics["recording_duration"] = d.metrics["recording_stopped"].(time.Time).Sub(d.metrics["recording_started"].(time.Time))
	}
	return nil
}

func (d *DesktopPlatform) GetMetrics() map[string]interface{} {
	// Initialize slices if not present
	if _, ok := d.metrics["click_actions"]; !ok {
		d.metrics["click_actions"] = []string{}
	}
	if _, ok := d.metrics["screenshots_taken"]; !ok {
		d.metrics["screenshots_taken"] = []string{}
	}
	if _, ok := d.metrics["fill_actions"]; !ok {
		d.metrics["fill_actions"] = []map[string]string{}
	}
	if _, ok := d.metrics["submit_actions"]; !ok {
		d.metrics["submit_actions"] = []string{}
	}
	if _, ok := d.metrics["navigate_actions"]; !ok {
		d.metrics["navigate_actions"] = []string{}
	}
	
	d.metrics["end_time"] = time.Now()
	d.metrics["total_duration"] = d.metrics["end_time"].(time.Time).Sub(d.metrics["start_time"].(time.Time))
	
	return d.metrics
}

func (d *DesktopPlatform) Close() error {
	// Clean up any running processes
	return nil
}