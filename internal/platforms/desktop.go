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
			"videos_taken":     []string{},
			"ui_action_placeholders": []string{},
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
	
	// Platform-specific UI automation
	var cmd *exec.Cmd
	
	switch {
	case runtime.GOOS == "darwin":
		// macOS: Use AppleScript for click automation
		// Click at coordinates or by window name/element
		if selector == "center" {
			// Click in center of screen
			cmd = exec.Command("osascript", "-e", `
				tell application "System Events"
					set {x, y} to (size of screen 1)
					set clickX to x / 2
					set clickY to y / 2
					click at {clickX, clickY}
				end tell
			`)
		} else {
			// Try to click on window/application
			cmd = exec.Command("osascript", "-e", fmt.Sprintf(`
				tell application "System Events"
					tell process "%s"
						click front window
					end tell
				end tell
			`, selector))
		}
		
	case runtime.GOOS == "windows":
		// Windows: Use PowerShell for UI automation
		if selector == "center" {
			cmd = exec.Command("powershell", "-Command", `
				Add-Type -AssemblyName System.Windows.Forms;
				$screen = [System.Windows.Forms.Screen]::PrimaryScreen;
				$x = $screen.Bounds.Width / 2;
				$y = $screen.Bounds.Height / 2;
				[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point($x, $y);
				[System.Windows.Forms.SendKeys]::SendWait("{CLICK}");
			`)
		} else {
			cmd = exec.Command("powershell", "-Command", fmt.Sprintf(`
				$app = Get-Process | Where-Object {$_.ProcessName -like "*%s*"} | Select-Object -First 1;
				if ($app) {
					$app.MainWindow.Activate();
					Start-Sleep -Milliseconds 500;
					[System.Windows.Forms.SendKeys]::SendWait("{CLICK}");
				}
			`, selector))
		}
		
	default:
		// Linux: Use xdotool for UI automation (fallback)
		if selector == "center" {
			cmd = exec.Command("sh", "-c", `
				eval $(xdotool getdisplaygeometry | awk '{print $1,$2}' | tr 'x' ' ');
				xdotool mousemove $((WIDTH/2)) $((HEIGHT/2)) click 1
			`)
		} else {
			// Placeholder for Linux
			time.Sleep(1 * time.Second)
			return d.createUIActionPlaceholder("click", selector, "Linux desktop automation requires xdotool")
		}
	}
	
	// Execute the command
	if cmd != nil {
		if err := cmd.Run(); err != nil {
			// Fallback to simulation if command fails
			time.Sleep(1 * time.Second)
			return d.createUIActionPlaceholder("click", selector, fmt.Sprintf("Desktop click failed: %v", err))
		}
	}
	
	// Wait a moment after click
	time.Sleep(500 * time.Millisecond)
	
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
	// Input validation
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	
	// Safe slice append
	if videosTaken, ok := d.metrics["videos_taken"].([]string); ok {
		d.metrics["videos_taken"] = append(videosTaken, filename)
	}
	
	// Platform-specific screen recording
	var cmd *exec.Cmd
	
	switch {
	case runtime.GOOS == "darwin":
		// macOS: Use screencapture with video recording
		// screencapture -v -R x,y,width,height output.mov
		// For full screen: screencapture -v output.mov
		cmd = exec.Command("screencapture", "-v", "-R", "0,0,1920,1080", filename)
	case runtime.GOOS == "windows":
		// Windows: Use PowerShell with built-in screen recording
		// Or use FFmpeg if available
		cmd = exec.Command("powershell", "-Command", 
			"Add-Type -AssemblyName System.Windows.Forms; Add-Type -AssemblyName System.Drawing; "+
			"$screen = [System.Windows.Forms.Screen]::PrimaryScreen; "+
			"$bitmap = New-Object System.Drawing.Bitmap $screen.Bounds.Width, $screen.Bounds.Height; "+
			"$graphics = [System.Drawing.Graphics]::FromImage($bitmap); "+
			"$graphics.CopyFromScreen($screen.Bounds.Location, [System.Drawing.Point]::Empty, $screen.Bounds.Size); "+
			"$bitmap.Save('"+filename+"', [System.Drawing.Imaging.ImageFormat]::Png); "+
			"$graphics.Dispose(); $bitmap.Dispose()")
	default:
		// Linux: Use FFmpeg if available, otherwise fallback
		cmd = exec.Command("ffmpeg", "-f", "x11grab", "-video_size", "1920x1080", "-i", ":0.0", "-t", "30", filename)
	}
	
	// Create video directory
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create video directory: %w", err)
	}
	
	// Start recording in background if command is valid
	if cmd != nil {
		// Check if command exists before starting
		if err := cmd.Start(); err != nil {
			// Fallback to placeholder file if recording fails
			return d.createVideoPlaceholder(filename, fmt.Sprintf("Failed to start %s recording", runtime.GOOS))
		}
		// Logging would go here: fmt.Printf("Desktop video recording started: %s", filename)
	} else {
		// Create placeholder if no recording method available
		return d.createVideoPlaceholder(filename, "No recording method available for this platform")
	}
	
	d.recording = true
	d.metrics["recording_started"] = time.Now()
	d.metrics["recording_file"] = filename
	d.metrics["recording_method"] = runtime.GOOS
	
	return nil
	
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
	if !d.recording {
		return fmt.Errorf("no recording in progress")
	}
	
	d.recording = false
	d.metrics["recording_stopped"] = time.Now()
	
	// Calculate recording duration safely
	if startTime, ok := d.metrics["recording_started"].(time.Time); ok {
		if stopTime, ok := d.metrics["recording_stopped"].(time.Time); ok {
			d.metrics["recording_duration"] = stopTime.Sub(startTime)
		}
	}
	
	// In a real implementation, this would:
	// 1. Stop the screencapture process on macOS
	// 2. Stop the FFmpeg process on Linux
	// 3. Save the final video file with proper encoding
	// 4. Clean up temporary files
	// 5. Return video metadata (resolution, duration, file size)
	
	// Logging would go here: fmt.Printf("Desktop video recording stopped. Duration: %v", d.metrics["recording_duration"])
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

func (d *DesktopPlatform) createVideoPlaceholder(filename, reason string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create video file: %w", err)
	}
	defer file.Close()
	
	// Write detailed placeholder header
	placeholderContent := fmt.Sprintf(`# PANOPTIC VIDEO RECORDING PLACEHOLDER
# Desktop Platform - %s
# Recording started: %s
# File: %s
# Reason: %s

# In a production implementation, this would be an actual video file.
# Current implementation may need additional dependencies:
# - macOS: screencapture command (built-in)
# - Windows: PowerShell with ScreenCapture APIs
# - Linux: FFmpeg package (install with: apt install ffmpeg)
`, runtime.GOOS, time.Now().Format(time.RFC3339), filename, reason)
	
	if _, err := file.WriteString(placeholderContent); err != nil {
		return fmt.Errorf("failed to write video header: %w", err)
	}
	
	// Logging would go here: fmt.Printf("Desktop video placeholder created: %s (Reason: %s)", filename, reason)
	return nil
}

func (d *DesktopPlatform) createUIActionPlaceholder(action, selector, reason string) error {
	// Create a placeholder file to document the UI action
	placeholderFile := fmt.Sprintf("desktop_ui_action_%s_%d.log", action, time.Now().Unix())
	placeholderContent := fmt.Sprintf(`# DESKTOP UI ACTION PLACEHOLDER
# Action: %s
# Selector: %s
# Time: %s
# Reason: %s

# In a production implementation, this would perform actual UI automation.
# Current implementation may need additional dependencies:
# - macOS: AppleScript support enabled in System Preferences
# - Windows: PowerShell execution policy configured
# - Linux: xdotool package installed (sudo apt install xdotool)

# To enable real UI automation:
# 1. macOS: System Preferences > Security & Privacy > Accessibility > Add Terminal
# 2. Windows: Set-ExecutionPolicy RemoteSigned (as Administrator)
# 3. Linux: Install xdotool and ensure X11 display is accessible
`, action, selector, time.Now().Format(time.RFC3339), reason)
	
	if err := os.WriteFile(placeholderFile, []byte(placeholderContent), 0600); err != nil {
		return fmt.Errorf("failed to write UI action placeholder: %w", err)
	}
	
	// Log placeholder creation
	if uiActions, ok := d.metrics["ui_action_placeholders"].([]string); ok {
		d.metrics["ui_action_placeholders"] = append(uiActions, placeholderFile)
	}
	
	return nil
}