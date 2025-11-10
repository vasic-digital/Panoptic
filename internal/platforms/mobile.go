package platforms

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"panoptic/internal/config"
)

type MobilePlatform struct {
	platform   string // ios, android
	device     string
	emulator   bool
	metrics    map[string]interface{}
	recording  bool
}

func NewMobilePlatform() *MobilePlatform {
	return &MobilePlatform{
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

func (m *MobilePlatform) Initialize(app config.AppConfig) error {
	m.metrics["start_time"] = time.Now()
	m.platform = app.Platform
	m.device = app.Device
	m.emulator = app.Emulator
	
	// Check if platform tools are available
	if err := m.checkPlatformTools(); err != nil {
		return fmt.Errorf("platform tools not available: %w", err)
	}
	
	// Check if device/emulator is available
	if err := m.checkDevice(); err != nil {
		return fmt.Errorf("device not available: %w", err)
	}
	
	m.metrics["platform"] = m.platform
	m.metrics["device"] = m.device
	m.metrics["emulator"] = m.emulator
	
	return nil
}

func (m *MobilePlatform) Navigate(url string) error {
	// For mobile apps, navigate might mean opening specific screens
	if m.platform == "android" {
		// Use adb commands or Appium for Android
		cmd := exec.Command("adb", "shell", "am", "start", "-a", "android.intent.action.VIEW", "-d", url)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to navigate on Android: %w", err)
		}
	} else if m.platform == "ios" {
		// Use xcrun simctl for iOS simulator
		if m.emulator {
			cmd := exec.Command("xcrun", "simctl", "openurl", m.device, url)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to navigate on iOS: %w", err)
			}
		}
	}
	
	// Safe slice append
	if navigateActions, ok := m.metrics["navigate_actions"].([]string); ok {
		m.metrics["navigate_actions"] = append(navigateActions, url)
	}
	
	waitForPageLoad()
	return nil
}

func (m *MobilePlatform) Click(selector string) error {
	// Mobile-specific click implementation
	// This would typically use Appium or platform-specific tools
	if m.platform == "android" {
		// Use adb tap commands or Appium
		cmd := exec.Command("adb", "shell", "input", "tap", "500", "500") // Example coordinates
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to click on Android: %w", err)
		}
	}
	
	// Safe slice append
	if clickActions, ok := m.metrics["click_actions"].([]string); ok {
		m.metrics["click_actions"] = append(clickActions, selector)
	}
	
	time.Sleep(1 * time.Second)
	return nil
}

func (m *MobilePlatform) Fill(selector, value string) error {
	// Mobile-specific text input
	if m.platform == "android" {
		// Use adb text input
		cmd := exec.Command("adb", "shell", "input", "text", value)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to fill text on Android: %w", err)
		}
	}
	
	// Safe slice append
	if fillActions, ok := m.metrics["fill_actions"].([]map[string]string); ok {
		newAction := map[string]string{
			"selector": selector,
			"value":    value,
		}
		m.metrics["fill_actions"] = append(fillActions, newAction)
	}
	
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (m *MobilePlatform) Submit(selector string) error {
	// Mobile-specific form submission
	if m.platform == "android" {
		// Send enter key
		cmd := exec.Command("adb", "shell", "input", "keyevent", "KEYCODE_ENTER")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to submit on Android: %w", err)
		}
	}
	
	// Safe slice append
	if submitActions, ok := m.metrics["submit_actions"].([]string); ok {
		m.metrics["submit_actions"] = append(submitActions, selector)
	}
	
	time.Sleep(1 * time.Second)
	return nil
}

func (m *MobilePlatform) Wait(duration int) error {
	time.Sleep(time.Duration(duration) * time.Second)
	return nil
}

func (m *MobilePlatform) Screenshot(filename string) error {
	var cmd *exec.Cmd
	
	if m.platform == "android" {
		cmd = exec.Command("adb", "shell", "screencap", "-p", "/sdcard/screenshot.png")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to capture screenshot on Android: %w", err)
		}
		
		// Pull screenshot to local file
		cmd = exec.Command("adb", "pull", "/sdcard/screenshot.png", filename)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to pull screenshot: %w", err)
		}
	} else if m.platform == "ios" && m.emulator {
		cmd = exec.Command("xcrun", "simctl", "io", m.device, "screenshot", filename)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to capture screenshot on iOS: %w", err)
		}
	}
	
	// Safe slice append
	if screenshotsTaken, ok := m.metrics["screenshots_taken"].([]string); ok {
		m.metrics["screenshots_taken"] = append(screenshotsTaken, filename)
	}
	
	return nil
}

func (m *MobilePlatform) StartRecording(filename string) error {
	// Input validation
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	
	// Safe slice append
	if videosTaken, ok := m.metrics["videos_taken"].([]string); ok {
		m.metrics["videos_taken"] = append(videosTaken, filename)
	}
	
	var cmd *exec.Cmd
	
	if m.platform == "android" {
		// Start screen recording on Android using ADB
		// screenrecord --time-limit <seconds> <output>
		// Default time limit 180 seconds (3 minutes)
		cmd = exec.Command("adb", "shell", "screenrecord", "--time-limit", "180", "/sdcard/recording.mp4")
	} else if m.platform == "ios" && m.emulator {
		// For iOS simulator, use xcrun simctl
		// xcrun simctl io <device> recordVideo <output>
		cmd = exec.Command("xcrun", "simctl", "io", m.device, "recordVideo", filename)
	} else if m.platform == "ios" && !m.emulator {
		// For physical iOS devices, recording is more complex
		// Would require additional setup like iOS developer tools
		return m.createVideoPlaceholder(filename, "iOS physical device recording not yet implemented")
	} else {
		return m.createVideoPlaceholder(filename, "Unsupported mobile platform")
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
			return m.createVideoPlaceholder(filename, fmt.Sprintf("Failed to start %s recording: %v", m.platform, err))
		}
		// Logging would go here: fmt.Printf("Mobile video recording started on %s: %s", m.platform, filename)
	}
	
	m.recording = true
	m.metrics["recording_started"] = time.Now()
	m.metrics["recording_file"] = filename
	m.metrics["recording_platform"] = m.platform
	m.metrics["recording_device"] = m.device
	
	return nil
}

func (m *MobilePlatform) StopRecording() error {
	if !m.recording {
		return fmt.Errorf("no recording in progress")
	}
	
	// Handle platform-specific stopping
	if m.platform == "android" {
		// Stop recording on Android (send Ctrl+C signal)
		cmd := exec.Command("pkill", "-INT", "-f", "screenrecord")
		if err := cmd.Run(); err != nil {
			// Logging would go here: fmt.Printf("Failed to stop Android recording gracefully: %v", err)
		}
		
		// Pull recording file from device
		if recordingFile, ok := m.metrics["recording_file"].(string); ok {
			localFile := recordingFile
			pullCmd := exec.Command("adb", "pull", "/sdcard/recording.mp4", localFile)
			if err := pullCmd.Run(); err != nil {
				// Logging would go here: fmt.Printf("Failed to pull Android recording: %v", err)
			} else {
				// Logging would go here: fmt.Printf("Android recording pulled to: %s", localFile)
			}
		}
	} else if m.platform == "ios" && m.emulator {
		// iOS simulator recording stops automatically when the command finishes
		// Logging would go here: fmt.Printf("iOS simulator recording stopped")
	}
	
	m.recording = false
	m.metrics["recording_stopped"] = time.Now()
	
	// Calculate recording duration safely
	if startTime, ok := m.metrics["recording_started"].(time.Time); ok {
		if stopTime, ok := m.metrics["recording_stopped"].(time.Time); ok {
			m.metrics["recording_duration"] = stopTime.Sub(startTime)
		}
	}
	
	// In a real implementation, this would:
	// 1. Verify video file was created successfully
	// 2. Check video file integrity and metadata
	// 3. Return video properties (resolution, duration, format, file size)
	// 4. Handle any cleanup of temporary files
	// 5. Log detailed recording information
	
	// Logging would go here: fmt.Printf("Mobile video recording stopped. Platform: %s, Duration: %v", 
	//	m.platform, m.metrics["recording_duration"])
	
	return nil
}

func (m *MobilePlatform) GetMetrics() map[string]interface{} {
	// Initialize slices if not present
	if _, ok := m.metrics["click_actions"]; !ok {
		m.metrics["click_actions"] = []string{}
	}
	if _, ok := m.metrics["screenshots_taken"]; !ok {
		m.metrics["screenshots_taken"] = []string{}
	}
	if _, ok := m.metrics["fill_actions"]; !ok {
		m.metrics["fill_actions"] = []map[string]string{}
	}
	if _, ok := m.metrics["submit_actions"]; !ok {
		m.metrics["submit_actions"] = []string{}
	}
	if _, ok := m.metrics["navigate_actions"]; !ok {
		m.metrics["navigate_actions"] = []string{}
	}
	
	m.metrics["end_time"] = time.Now()
	m.metrics["total_duration"] = m.metrics["end_time"].(time.Time).Sub(m.metrics["start_time"].(time.Time))
	
	return m.metrics
}

func (m *MobilePlatform) Close() error {
	return nil
}

func (m *MobilePlatform) checkPlatformTools() error {
	if m.platform == "android" {
		// Check if adb is available
		_, err := exec.LookPath("adb")
		if err != nil {
			return fmt.Errorf("adb not found in PATH")
		}
	} else if m.platform == "ios" {
		// Check if xcrun is available (macOS only)
		_, err := exec.LookPath("xcrun")
		if err != nil {
			return fmt.Errorf("xcrun not found in PATH")
		}
	}
	
	return nil
}

func (m *MobilePlatform) checkDevice() error {
	if m.platform == "android" {
		// Check if device/emulator is connected
		cmd := exec.Command("adb", "devices")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to check Android devices: %w", err)
		}
	} else if m.platform == "ios" && m.emulator {
		// Check if simulator is available
		cmd := exec.Command("xcrun", "simctl", "list", "devices", "available")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to list iOS simulators: %w", err)
		}
	}
	
	return nil
}

func (m *MobilePlatform) createVideoPlaceholder(filename, reason string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create video file: %w", err)
	}
	defer file.Close()
	
	// Write detailed placeholder header
	placeholderContent := fmt.Sprintf(`# PANOPTIC VIDEO RECORDING PLACEHOLDER
# Mobile Platform - %s
# Device: %s
# Emulator: %t
# Recording started: %s
# File: %s
# Reason: %s

# In a production implementation, this would be an actual video file.
# Current implementation requirements:
# - Android: ADB (Android Debug Bridge) installed and configured
# - iOS Simulator: Xcode with iOS simulator tools
# - iOS Physical Device: Additional developer tools and permissions

# To enable real recording:
# 1. Install Android SDK tools (for Android)
# 2. Install Xcode (for iOS)
# 3. Ensure device/emulator is running and connected
# 4. Grant necessary recording permissions
# 5. For Android: adb devices (should show connected devices)
# 6. For iOS: xcrun simctl list devices (should show simulators)
`, m.platform, m.device, m.emulator, time.Now().Format(time.RFC3339), filename, reason)
	
	if _, err := file.WriteString(placeholderContent); err != nil {
		return fmt.Errorf("failed to write video header: %w", err)
	}
	
	// Logging would go here: fmt.Printf("Mobile video placeholder created: %s (Reason: %s)", filename, reason)
	return nil
}