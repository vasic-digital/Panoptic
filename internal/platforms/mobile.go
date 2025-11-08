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
		metrics: make(map[string]interface{}),
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
	
	m.metrics["navigate_actions"] = append(m.metrics["navigate_actions"].([]string), url)
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
	
	m.metrics["click_actions"] = append(m.metrics["click_actions"].([]string), selector)
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
	
	m.metrics["fill_actions"] = append(m.metrics["fill_actions"].([]map[string]string), map[string]string{
		"selector": selector,
		"value":    value,
	})
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
	
	m.metrics["submit_actions"] = append(m.metrics["submit_actions"].([]string), selector)
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
	
	m.metrics["screenshots_taken"] = append(m.metrics["screenshots_taken"].([]string), filename)
	return nil
}

func (m *MobilePlatform) StartRecording(filename string) error {
	var cmd *exec.Cmd
	
	if m.platform == "android" {
		// Start screen recording on Android
		cmd = exec.Command("adb", "shell", "screenrecord", "--time-limit", "180", "/sdcard/recording.mp4")
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start recording on Android: %w", err)
		}
	} else if m.platform == "ios" && m.emulator {
		// For iOS simulator
		cmd = exec.Command("xcrun", "simctl", "io", m.device, "recordVideo", filename)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start recording on iOS: %w", err)
		}
	}
	
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create video directory: %w", err)
	}
	
	m.recording = true
	m.metrics["recording_started"] = time.Now()
	m.metrics["recording_file"] = filename
	
	return nil
}

func (m *MobilePlatform) StopRecording() error {
	if m.recording {
		if m.platform == "android" {
			// Stop recording on Android (Ctrl+C)
			cmd := exec.Command("pkill", "-f", "screenrecord")
			cmd.Run()
			
			// Pull recording file
			if recordingFile, ok := m.metrics["recording_file"].(string); ok {
				localFile := recordingFile
				cmd = exec.Command("adb", "pull", "/sdcard/recording.mp4", localFile)
				cmd.Run()
			}
		}
		
		m.recording = false
		m.metrics["recording_stopped"] = time.Now()
		m.metrics["recording_duration"] = m.metrics["recording_stopped"].(time.Time).Sub(m.metrics["recording_started"].(time.Time))
	}
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