package platforms

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/config"

	"github.com/stretchr/testify/assert"
)

// TestNewMobilePlatform verifies constructor initialization
func TestNewMobilePlatform(t *testing.T) {
	platform := NewMobilePlatform()

	assert.NotNil(t, platform)
	assert.NotNil(t, platform.metrics)

	// Verify all metric slices are initialized
	assert.NotNil(t, platform.metrics["click_actions"])
	assert.NotNil(t, platform.metrics["screenshots_taken"])
	assert.NotNil(t, platform.metrics["fill_actions"])
	assert.NotNil(t, platform.metrics["submit_actions"])
	assert.NotNil(t, platform.metrics["navigate_actions"])
	assert.NotNil(t, platform.metrics["videos_taken"])
	assert.NotNil(t, platform.metrics["mobile_ui_placeholders"])
	assert.NotNil(t, platform.metrics["start_time"])

	// Verify recording state
	assert.False(t, platform.recording)
}

// TestMobilePlatform_Initialize_Android verifies Android platform initialization
func TestMobilePlatform_Initialize_Android(t *testing.T) {
	platform := NewMobilePlatform()

	app := config.AppConfig{
		Platform: "android",
		Device:   "emulator-5554",
		Emulator: true,
		Timeout:  30,
	}

	// This will fail if adb is not installed, which is expected in most test environments
	err := platform.Initialize(app)

	// Either succeeds if adb is available, or fails with expected error
	if err != nil {
		assert.Contains(t, err.Error(), "platform tools not available")
	} else {
		assert.Equal(t, "android", platform.platform)
		assert.Equal(t, "emulator-5554", platform.device)
		assert.True(t, platform.emulator)
		assert.Equal(t, "android", platform.metrics["platform"])
		assert.Equal(t, "emulator-5554", platform.metrics["device"])
		assert.Equal(t, true, platform.metrics["emulator"])
	}
}

// TestMobilePlatform_Initialize_iOS verifies iOS platform initialization
func TestMobilePlatform_Initialize_iOS(t *testing.T) {
	platform := NewMobilePlatform()

	app := config.AppConfig{
		Platform: "ios",
		Device:   "iPhone 14",
		Emulator: true,
		Timeout:  30,
	}

	// This will fail if xcrun is not installed, which is expected on non-macOS systems
	err := platform.Initialize(app)

	// Either succeeds if xcrun is available, or fails with expected error
	if err != nil {
		assert.Contains(t, err.Error(), "platform tools not available")
	} else {
		assert.Equal(t, "ios", platform.platform)
		assert.Equal(t, "iPhone 14", platform.device)
		assert.True(t, platform.emulator)
	}
}

// TestMobilePlatform_Navigate_Android verifies Android navigation
func TestMobilePlatform_Navigate_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.metrics["navigate_actions"] = []string{}

	// This will fail without adb, but verifies the command would be executed
	err := platform.Navigate("https://example.com")

	// Error is expected if adb is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to navigate")
	} else {
		// If successful, verify metrics were updated
		navigateActions := platform.metrics["navigate_actions"].([]string)
		assert.Contains(t, navigateActions, "https://example.com")
	}
}

// TestMobilePlatform_Navigate_iOS verifies iOS navigation
func TestMobilePlatform_Navigate_iOS(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 14"
	platform.emulator = true
	platform.metrics["navigate_actions"] = []string{}

	// This will fail without xcrun, but verifies the command would be executed
	err := platform.Navigate("https://example.com")

	// Error is expected if xcrun is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to navigate")
	} else {
		// If successful, verify metrics were updated
		navigateActions := platform.metrics["navigate_actions"].([]string)
		assert.Contains(t, navigateActions, "https://example.com")
	}
}

// TestMobilePlatform_Click_Android_Coordinates verifies Android coordinate-based clicking
func TestMobilePlatform_Click_Android_Coordinates(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.metrics["click_actions"] = []string{}

	// Test coordinate selector
	err := platform.Click("100,200")

	// Error is expected if adb is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to click")
	} else {
		// If successful, verify metrics were updated
		clickActions := platform.metrics["click_actions"].([]string)
		assert.Contains(t, clickActions, "100,200")
	}
}

// TestMobilePlatform_Click_Android_Center verifies Android center clicking
func TestMobilePlatform_Click_Android_Center(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.metrics["click_actions"] = []string{}

	// Test center selector
	err := platform.Click("center")

	// Error is expected if adb is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to click")
	} else {
		// If successful, verify metrics were updated
		clickActions := platform.metrics["click_actions"].([]string)
		assert.Contains(t, clickActions, "center")
	}
}

// TestMobilePlatform_Click_iOS_Coordinates verifies iOS coordinate-based clicking
func TestMobilePlatform_Click_iOS_Coordinates(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 14"
	platform.emulator = true
	platform.metrics["click_actions"] = []string{}

	// Test coordinate selector
	err := platform.Click("150,300")

	// Error is expected if xcrun is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to click")
	} else {
		// If successful, verify metrics were updated
		clickActions := platform.metrics["click_actions"].([]string)
		assert.Contains(t, clickActions, "150,300")
	}
}

// TestMobilePlatform_Click_iOS_PhysicalDevice verifies iOS physical device returns placeholder
func TestMobilePlatform_Click_iOS_PhysicalDevice(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 13"
	platform.emulator = false // Physical device
	platform.metrics["mobile_ui_placeholders"] = []string{}

	// Physical device should create placeholder
	err := platform.Click("Submit Button")

	// Should succeed but create placeholder
	assert.NoError(t, err)

	// Verify placeholder was created
	placeholders := platform.metrics["mobile_ui_placeholders"].([]string)
	assert.NotEmpty(t, placeholders)

	// Cleanup placeholder file
	if len(placeholders) > 0 {
		os.Remove(placeholders[0])
	}
}

// TestMobilePlatform_Fill_Android verifies Android text input
func TestMobilePlatform_Fill_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.metrics["fill_actions"] = []map[string]string{}

	err := platform.Fill("username", "testuser")

	// Error is expected if adb is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to fill text")
	} else {
		// If successful, verify metrics were updated
		fillActions := platform.metrics["fill_actions"].([]map[string]string)
		assert.Len(t, fillActions, 1)
		assert.Equal(t, "username", fillActions[0]["selector"])
		assert.Equal(t, "testuser", fillActions[0]["value"])
	}
}

// TestMobilePlatform_Submit_Android verifies Android form submission
func TestMobilePlatform_Submit_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.metrics["submit_actions"] = []string{}

	err := platform.Submit("login_form")

	// Error is expected if adb is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to submit")
	} else {
		// If successful, verify metrics were updated
		submitActions := platform.metrics["submit_actions"].([]string)
		assert.Contains(t, submitActions, "login_form")
	}
}

// TestMobilePlatform_Wait verifies wait functionality
func TestMobilePlatform_Wait(t *testing.T) {
	platform := NewMobilePlatform()

	start := time.Now()
	err := platform.Wait(1)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 1*time.Second)
}

// TestMobilePlatform_Screenshot_Android verifies Android screenshot capture
func TestMobilePlatform_Screenshot_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.metrics["screenshots_taken"] = []string{}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "screenshot.png")

	err := platform.Screenshot(filename)

	// Error is expected if adb is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to capture screenshot")
	} else {
		// If successful, verify metrics were updated
		screenshots := platform.metrics["screenshots_taken"].([]string)
		assert.Contains(t, screenshots, filename)
	}
}

// TestMobilePlatform_Screenshot_iOS verifies iOS screenshot capture
func TestMobilePlatform_Screenshot_iOS(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 14"
	platform.emulator = true
	platform.metrics["screenshots_taken"] = []string{}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "screenshot.png")

	err := platform.Screenshot(filename)

	// Error is expected if xcrun is not available
	if err != nil {
		assert.Contains(t, err.Error(), "failed to capture screenshot")
	} else {
		// If successful, verify metrics were updated
		screenshots := platform.metrics["screenshots_taken"].([]string)
		assert.Contains(t, screenshots, filename)
	}
}

// TestMobilePlatform_StartRecording_EmptyFilename verifies validation
func TestMobilePlatform_StartRecording_EmptyFilename(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"

	err := platform.StartRecording("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filename cannot be empty")
}

// TestMobilePlatform_StartRecording_Android verifies Android video recording
func TestMobilePlatform_StartRecording_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.metrics["videos_taken"] = []string{}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "recording.mp4")

	err := platform.StartRecording(filename)

	// Either succeeds or creates placeholder
	if err == nil {
		// Recording started successfully
		assert.True(t, platform.recording)
		assert.Equal(t, filename, platform.metrics["recording_file"])
		assert.Equal(t, "android", platform.metrics["recording_platform"])
		assert.NotNil(t, platform.metrics["recording_started"])

		// Verify metrics
		videos := platform.metrics["videos_taken"].([]string)
		assert.Contains(t, videos, filename)
	} else {
		// Placeholder created (expected if adb not available)
		assert.NoError(t, err)
	}
}

// TestMobilePlatform_StartRecording_iOS verifies iOS video recording
func TestMobilePlatform_StartRecording_iOS(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 14"
	platform.emulator = true
	platform.metrics["videos_taken"] = []string{}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "recording.mp4")

	err := platform.StartRecording(filename)

	// Either succeeds or creates placeholder
	if err == nil {
		// Recording started successfully
		assert.True(t, platform.recording)
		assert.Equal(t, filename, platform.metrics["recording_file"])
		assert.Equal(t, "ios", platform.metrics["recording_platform"])
		assert.Equal(t, "iPhone 14", platform.metrics["recording_device"])

		// Verify metrics
		videos := platform.metrics["videos_taken"].([]string)
		assert.Contains(t, videos, filename)
	} else {
		// Placeholder created (expected if xcrun not available)
		assert.NoError(t, err)
	}
}

// TestMobilePlatform_StartRecording_iOS_PhysicalDevice verifies iOS physical device creates placeholder
func TestMobilePlatform_StartRecording_iOS_PhysicalDevice(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 13"
	platform.emulator = false // Physical device
	platform.metrics["videos_taken"] = []string{}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "recording.mp4")

	err := platform.StartRecording(filename)

	// Should create placeholder for physical device
	assert.NoError(t, err)
	assert.FileExists(t, filename)

	// Verify it's a placeholder file
	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "VIDEO RECORDING PLACEHOLDER")
	assert.Contains(t, string(content), "iOS physical device recording not yet implemented")
}

// TestMobilePlatform_StopRecording_NoRecording verifies error when not recording
func TestMobilePlatform_StopRecording_NoRecording(t *testing.T) {
	platform := NewMobilePlatform()

	err := platform.StopRecording()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recording in progress")
}

// TestMobilePlatform_StopRecording_Android verifies Android recording stop
func TestMobilePlatform_StopRecording_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.recording = true
	platform.metrics["recording_started"] = time.Now()
	platform.metrics["recording_file"] = "recording.mp4"

	err := platform.StopRecording()

	// Should complete without error (even if adb commands fail)
	assert.NoError(t, err)
	assert.False(t, platform.recording)
	assert.NotNil(t, platform.metrics["recording_stopped"])

	// Verify duration was calculated
	assert.NotNil(t, platform.metrics["recording_duration"])
}

// TestMobilePlatform_StopRecording_iOS verifies iOS recording stop
func TestMobilePlatform_StopRecording_iOS(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.emulator = true
	platform.recording = true
	platform.metrics["recording_started"] = time.Now()
	platform.metrics["recording_file"] = "recording.mp4"

	// Wait a moment to ensure measurable duration
	time.Sleep(100 * time.Millisecond)

	err := platform.StopRecording()

	assert.NoError(t, err)
	assert.False(t, platform.recording)
	assert.NotNil(t, platform.metrics["recording_stopped"])

	// Verify duration was calculated and is reasonable
	duration := platform.metrics["recording_duration"].(time.Duration)
	assert.Greater(t, duration, time.Duration(0))
	assert.Less(t, duration, 5*time.Second) // Should be under 5 seconds for test
}

// TestMobilePlatform_GetMetrics verifies metrics collection
func TestMobilePlatform_GetMetrics(t *testing.T) {
	platform := NewMobilePlatform()

	// Simulate some actions
	platform.metrics["click_actions"] = []string{"button1", "button2"}
	platform.metrics["screenshots_taken"] = []string{"screen1.png"}
	platform.metrics["fill_actions"] = []map[string]string{
		{"selector": "username", "value": "test"},
	}

	// Wait a moment to ensure measurable duration
	time.Sleep(100 * time.Millisecond)

	metrics := platform.GetMetrics()

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics["start_time"])
	assert.NotNil(t, metrics["end_time"])
	assert.NotNil(t, metrics["total_duration"])

	// Verify duration is reasonable
	duration := metrics["total_duration"].(time.Duration)
	assert.Greater(t, duration, time.Duration(0))

	// Verify action counts
	assert.Len(t, metrics["click_actions"].([]string), 2)
	assert.Len(t, metrics["screenshots_taken"].([]string), 1)
	assert.Len(t, metrics["fill_actions"].([]map[string]string), 1)
}

// TestMobilePlatform_GetMetrics_InitializesSlices verifies slice initialization
func TestMobilePlatform_GetMetrics_InitializesSlices(t *testing.T) {
	platform := NewMobilePlatform()

	// Remove some slices to test initialization
	delete(platform.metrics, "click_actions")
	delete(platform.metrics, "screenshots_taken")

	metrics := platform.GetMetrics()

	// Verify slices were initialized
	assert.NotNil(t, metrics["click_actions"])
	assert.NotNil(t, metrics["screenshots_taken"])
	assert.NotNil(t, metrics["fill_actions"])
	assert.NotNil(t, metrics["submit_actions"])
	assert.NotNil(t, metrics["navigate_actions"])
}

// TestMobilePlatform_Close verifies cleanup
func TestMobilePlatform_Close(t *testing.T) {
	platform := NewMobilePlatform()

	err := platform.Close()

	assert.NoError(t, err)
}

// TestMobilePlatform_checkPlatformTools_Android verifies Android tool checking
func TestMobilePlatform_checkPlatformTools_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"

	err := platform.checkPlatformTools()

	// Either adb is available or not
	if err != nil {
		assert.Contains(t, err.Error(), "adb not found in PATH")
	} else {
		// adb is installed
		assert.NoError(t, err)
	}
}

// TestMobilePlatform_checkPlatformTools_iOS verifies iOS tool checking
func TestMobilePlatform_checkPlatformTools_iOS(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"

	err := platform.checkPlatformTools()

	// Either xcrun is available or not
	if err != nil {
		assert.Contains(t, err.Error(), "xcrun not found in PATH")
	} else {
		// xcrun is installed (macOS)
		assert.NoError(t, err)
	}
}

// TestMobilePlatform_checkDevice_Android verifies Android device checking
func TestMobilePlatform_checkDevice_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"

	err := platform.checkDevice()

	// Either succeeds or fails based on adb availability
	if err != nil {
		assert.Contains(t, err.Error(), "failed to check Android devices")
	} else {
		// adb devices command succeeded
		assert.NoError(t, err)
	}
}

// TestMobilePlatform_checkDevice_iOS verifies iOS device checking
func TestMobilePlatform_checkDevice_iOS(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.emulator = true

	err := platform.checkDevice()

	// Either succeeds or fails based on xcrun availability
	if err != nil {
		assert.Contains(t, err.Error(), "failed to list iOS simulators")
	} else {
		// xcrun simctl command succeeded
		assert.NoError(t, err)
	}
}

// TestMobilePlatform_createVideoPlaceholder verifies placeholder creation
func TestMobilePlatform_createVideoPlaceholder(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.device = "emulator-5554"
	platform.emulator = true

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "placeholder.mp4")

	err := platform.createVideoPlaceholder(filename, "Test reason")

	assert.NoError(t, err)
	assert.FileExists(t, filename)

	// Verify placeholder content
	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "VIDEO RECORDING PLACEHOLDER")
	assert.Contains(t, string(content), "Mobile Platform - android")
	assert.Contains(t, string(content), "Device: emulator-5554")
	assert.Contains(t, string(content), "Emulator: true")
	assert.Contains(t, string(content), "Test reason")
}

// TestMobilePlatform_createMobileUIPlaceholder verifies UI placeholder creation
func TestMobilePlatform_createMobileUIPlaceholder(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 14"
	platform.emulator = false
	platform.metrics["mobile_ui_placeholders"] = []string{}

	err := platform.createMobileUIPlaceholder("click", "Submit Button", "Test reason")

	assert.NoError(t, err)

	// Verify placeholder was tracked in metrics
	placeholders := platform.metrics["mobile_ui_placeholders"].([]string)
	assert.Len(t, placeholders, 1)

	// Verify placeholder file was created
	assert.FileExists(t, placeholders[0])

	// Verify placeholder content
	content, err := os.ReadFile(placeholders[0])
	assert.NoError(t, err)
	assert.Contains(t, string(content), "MOBILE UI ACTION PLACEHOLDER")
	assert.Contains(t, string(content), "Platform: ios")
	assert.Contains(t, string(content), "Device: iPhone 14")
	assert.Contains(t, string(content), "Emulator: false")
	assert.Contains(t, string(content), "Action: click")
	assert.Contains(t, string(content), "Selector: Submit Button")
	assert.Contains(t, string(content), "Test reason")

	// Cleanup
	os.Remove(placeholders[0])
}

// TestMobilePlatform_Integration_AndroidWorkflow verifies complete Android workflow
func TestMobilePlatform_Integration_AndroidWorkflow(t *testing.T) {
	platform := NewMobilePlatform()

	app := config.AppConfig{
		Platform: "android",
		Device:   "emulator-5554",
		Emulator: true,
		Timeout:  30,
	}

	// Initialize (may fail if adb not available)
	err := platform.Initialize(app)
	if err != nil {
		t.Skip("Skipping Android integration test: adb not available")
	}

	// Test workflow
	platform.Navigate("https://example.com")
	platform.Click("100,200")
	platform.Fill("search", "test query")
	platform.Submit("search_form")
	platform.Wait(1)

	tmpDir := t.TempDir()
	platform.Screenshot(filepath.Join(tmpDir, "test.png"))

	// Get metrics
	metrics := platform.GetMetrics()
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics["total_duration"])

	// Cleanup
	platform.Close()
}

// TestMobilePlatform_Integration_iOSWorkflow verifies complete iOS workflow
func TestMobilePlatform_Integration_iOSWorkflow(t *testing.T) {
	platform := NewMobilePlatform()

	app := config.AppConfig{
		Platform: "ios",
		Device:   "iPhone 14",
		Emulator: true,
		Timeout:  30,
	}

	// Initialize (may fail if xcrun not available)
	err := platform.Initialize(app)
	if err != nil {
		t.Skip("Skipping iOS integration test: xcrun not available")
	}

	// Test workflow
	platform.Navigate("https://example.com")
	platform.Click("center")
	platform.Wait(1)

	tmpDir := t.TempDir()
	platform.Screenshot(filepath.Join(tmpDir, "test.png"))

	// Get metrics
	metrics := platform.GetMetrics()
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics["total_duration"])

	// Cleanup
	platform.Close()
}
