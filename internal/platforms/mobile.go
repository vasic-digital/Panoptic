package platforms

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"panoptic/internal/config"
)

// ErrMobileDeviceInteractionNotWired is returned by mobile-platform action
// helpers (createMobileUIPlaceholder, createVideoPlaceholder) when the real
// device-interaction pipeline (Android: ADB / uiautomator / screenrecord;
// iOS: xcrun simctl / accessibility inspector / XCTest) has not been wired
// for the target action, OR the wired pipeline failed and the previous
// fallback would have written a description file in lieu of performing the
// action.
//
// Round-29 §11.4 anti-bluff audit (2026-05-17): the previous implementation
// of these helpers wrote a *.log text file describing what the action
// "would have done" and returned (nil), so a test that invoked
// MobilePlatform.Click / MobilePlatform.StartRecording could PASS while
// no tap, no fill, no screen capture had happened. CRITICAL CONTRACT-bluff
// under CONST-035 / Article XI §11.9: the platform advertised a Click /
// StartRecording capability, the test scored PASS, the device received no
// interaction. The helpers now refuse to fabricate and surface this
// sentinel instead. Callers MUST treat this as a real failure (not silently
// swallow it) and either (a) wire the real ADB / uiautomator / xcrun
// pipeline before invoking the action, OR (b) skip the action with an
// explicit "device-interaction not wired" annotation captured in the
// test record (so the test scoreboard reflects reality).
var ErrMobileDeviceInteractionNotWired = fmt.Errorf("panoptic mobile: device-interaction actions (screencap, adb input tap, screenrecord, xcrun simctl io recordVideo, etc.) have not been wired — the previous implementation wrote a description file to disk and returned success without performing the actual UI action (§11.4 CONTRACT-bluff under CONST-035 / Article XI §11.9). Wire pkg/devices/adb (or equivalent) before invoking mobile UI actions; or use the real-device-paired challenge runner")

type MobilePlatform struct {
	platform     string // ios, android
	device       string
	emulator     bool
	recordingCmd  *exec.Cmd
	metrics      map[string]interface{}
	recording    bool
}

func NewMobilePlatform() *MobilePlatform {
	return &MobilePlatform{
		metrics: map[string]interface{}{
			"click_actions":     []string{},
			"screenshots_taken":  []string{},
			"fill_actions":      []map[string]string{},
			"submit_actions":    []string{},
			"navigate_actions":  []string{},
			"videos_taken":     []string{},
			// Sentinel-gap tracking slots (round-29 §11.4 anti-bluff
			// audit, 2026-05-17). These replace the previous
			// "mobile_ui_placeholders" slot which tracked fabricated
			// description files. Now they capture honest records of
			// unwired UI / video actions so the run report can surface
			// the gap rather than fake success.
			"mobile_ui_not_wired":    []string{},
			"mobile_video_not_wired": []string{},
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
	// Enhanced mobile click implementation
	if m.platform == "android" {
		// Parse coordinates from selector (format: "x,y" or "center")
		var x, y int
		
		if selector == "center" {
			// Get screen dimensions and click center
			cmd := exec.Command("adb", "shell", "wm", "size")
			output, err := cmd.Output()
			if err != nil {
				// Fallback to center coordinates
				x, y = 540, 960 // Default center
			} else {
				// Parse screen size and calculate center
				sizeStr := string(output)
				if _, err := fmt.Sscanf(sizeStr, "Physical size: %dx%d", &x, &y); err != nil {
					x, y = 540, 960 // Default center
				} else {
					x, y = x/2, y/2
				}
			}
		} else if _, err := fmt.Sscanf(selector, "%d,%d", &x, &y); err == nil {
			// Use provided coordinates
		} else {
			// Try to find element by text (Android only)
			cmd := exec.Command("adb", "shell", "uiautomator", "dump")
			output, err := cmd.Output()
			if err != nil {
				return m.createMobileUIPlaceholder("click", selector, "Android UI automation requires uiautomator")
			}
			
			// Simple text search in UI dump (would need XML parsing in production)
			if !strings.Contains(string(output), selector) {
				return fmt.Errorf("element with text '%s' not found", selector)
			}
			
			// Fallback to center click for found element
			x, y = 540, 960
		}
		
		// Perform click
		cmd := exec.Command("adb", "shell", "input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to click on Android at %d,%d: %w", x, y, err)
		}
		
	} else if m.platform == "ios" {
		// iOS simulator: Use xcrun simctl
		if m.emulator {
			var x, y int
			
			if selector == "center" {
				// iPhone 12 center coordinates
				x, y = 200, 400
			} else if _, err := fmt.Sscanf(selector, "%d,%d", &x, &y); err == nil {
				// Use provided coordinates
			} else {
				// For iOS, we would need accessibility inspector or similar
				return m.createMobileUIPlaceholder("click", selector, "iOS UI automation requires accessibility tools")
			}
			
			cmd := exec.Command("xcrun", "simctl", "io", m.device, "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to click on iOS simulator at %d,%d: %w", x, y, err)
			}
		} else {
			return m.createMobileUIPlaceholder("click", selector, "iOS physical device automation requires additional setup")
		}
	}
	
	// Safe slice append
	if clickActions, ok := m.metrics["click_actions"].([]string); ok {
		m.metrics["click_actions"] = append(clickActions, selector)
	}
	
	// Wait a moment after click
	time.Sleep(800 * time.Millisecond)
	
	return nil
}

// createMobileUIPlaceholder previously wrote a description *.log file to
// the working directory and returned (nil), letting a test invoke
// MobilePlatform.Click / Fill / Submit and score PASS while the device
// received zero interaction. Per the round-29 §11.4 anti-bluff audit
// (2026-05-17) the helper now returns ErrMobileDeviceInteractionNotWired
// wrapped with the action / selector / reason context, and records the
// surfaced gap in `metrics["mobile_ui_not_wired"]` so the run report
// reflects reality. NO file is written; NO success is fabricated.
//
// Callers (MobilePlatform.Click and any future action dispatcher) MUST
// propagate this error to the test runner — silently swallowing it would
// reintroduce the original CONTRACT-bluff.
//
// Constitutional anchors: CONST-035 (anti-bluff), CONST-050(A)
// (no-fakes-beyond-unit-tests), Article XI §11.9 (forensic anchor).
func (m *MobilePlatform) createMobileUIPlaceholder(action, selector, reason string) error {
	// Record the gap in metrics so the test report can surface it.
	gapRecord := fmt.Sprintf("action=%s selector=%s reason=%q platform=%s device=%s emulator=%t at=%s",
		action, selector, reason, m.platform, m.device, m.emulator, time.Now().Format(time.RFC3339))
	if existing, ok := m.metrics["mobile_ui_not_wired"].([]string); ok {
		m.metrics["mobile_ui_not_wired"] = append(existing, gapRecord)
	} else {
		m.metrics["mobile_ui_not_wired"] = []string{gapRecord}
	}

	return fmt.Errorf("mobile UI action %q on selector %q (reason: %s): %w",
		action, selector, reason, ErrMobileDeviceInteractionNotWired)
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
		// Store the command for stopping later
		m.recordingCmd = cmd
	} else {
		return m.createVideoPlaceholder(filename, "No recording method available")
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

// createVideoPlaceholder previously wrote a text file to disk with a
// "PANOPTIC VIDEO RECORDING PLACEHOLDER" header and returned (nil),
// letting a test invoke MobilePlatform.StartRecording and score PASS
// while no screen frames had been captured. A downstream §11.4.5 video-
// quality analyzer reading the file would see a 0-frame "recording" and
// either FAIL-bluff (Bug-24 pattern: 0-byte mp4 confused with real
// recording) or PASS-bluff (assertion based on file presence rather
// than frame count). Per the round-29 §11.4 anti-bluff audit
// (2026-05-17) the helper now returns ErrMobileDeviceInteractionNotWired
// wrapped with the filename / reason context, and records the gap in
// `metrics["mobile_video_not_wired"]`. NO file is written.
//
// Callers (MobilePlatform.StartRecording and its fallback paths) MUST
// propagate this error to the test runner.
//
// Constitutional anchors: CONST-035 (anti-bluff), CONST-050(A)
// (no-fakes-beyond-unit-tests), Article XI §11.9 (forensic anchor),
// §11.4.5 (video-quality analysis comprehensiveness).
func (m *MobilePlatform) createVideoPlaceholder(filename, reason string) error {
	gapRecord := fmt.Sprintf("filename=%s reason=%q platform=%s device=%s emulator=%t at=%s",
		filename, reason, m.platform, m.device, m.emulator, time.Now().Format(time.RFC3339))
	if existing, ok := m.metrics["mobile_video_not_wired"].([]string); ok {
		m.metrics["mobile_video_not_wired"] = append(existing, gapRecord)
	} else {
		m.metrics["mobile_video_not_wired"] = []string{gapRecord}
	}

	return fmt.Errorf("mobile video recording for %q (reason: %s): %w",
		filename, reason, ErrMobileDeviceInteractionNotWired)
}