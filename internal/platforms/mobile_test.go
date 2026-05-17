package platforms

import (
	"errors"
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
	// Sentinel-gap tracking slots replaced the legacy
	// `mobile_ui_placeholders` slot in round-29 §11.4 audit.
	assert.NotNil(t, platform.metrics["mobile_ui_not_wired"])
	assert.NotNil(t, platform.metrics["mobile_video_not_wired"])
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

// TestMobilePlatform_Click_iOS_PhysicalDevice verifies iOS physical
// device clicks surface ErrMobileDeviceInteractionNotWired rather than
// silently fabricating a placeholder file (round-29 §11.4 anti-bluff
// audit). Before the audit Click() returned (nil) for the physical-
// device fallback path, letting the test runner score PASS while the
// device received zero interaction.
func TestMobilePlatform_Click_iOS_PhysicalDevice(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 13"
	platform.emulator = false // Physical device

	err := platform.Click("Submit Button")

	// Sentinel surfaced — Click MUST propagate the gap, not swallow it.
	require := assert.New(t)
	require.Error(err, "Click on iOS physical device MUST surface the device-interaction-not-wired sentinel")
	require.True(errors.Is(err, ErrMobileDeviceInteractionNotWired),
		"surfaced error MUST be ErrMobileDeviceInteractionNotWired (got %v)", err)

	// Gap MUST be recorded in metrics for the run report.
	notWired, ok := platform.metrics["mobile_ui_not_wired"].([]string)
	require.True(ok)
	require.NotEmpty(notWired)
	require.Contains(notWired[0], "selector=Submit Button")
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

// TestMobilePlatform_StartRecording_Android verifies Android video
// recording — either the real ADB screenrecord pipeline starts AND
// recording state is set, OR (when adb is unavailable in the test
// environment) the sentinel-default surfaces
// ErrMobileDeviceInteractionNotWired. NEVER a silent placeholder
// success per round-29 §11.4 anti-bluff audit.
func TestMobilePlatform_StartRecording_Android(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.metrics["videos_taken"] = []string{}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "recording.mp4")

	err := platform.StartRecording(filename)

	if err == nil {
		// Real recording started successfully (adb available).
		assert.True(t, platform.recording)
		assert.Equal(t, filename, platform.metrics["recording_file"])
		assert.Equal(t, "android", platform.metrics["recording_platform"])
		assert.NotNil(t, platform.metrics["recording_started"])

		videos := platform.metrics["videos_taken"].([]string)
		assert.Contains(t, videos, filename)
	} else {
		// Sentinel surfaced — adb unavailable AND fallback refused to
		// fabricate a placeholder file. The error MUST be the wired
		// sentinel and the file MUST NOT exist on disk.
		assert.True(t, errors.Is(err, ErrMobileDeviceInteractionNotWired),
			"non-real-recording path MUST surface ErrMobileDeviceInteractionNotWired (got %v)", err)
		assert.NoFileExists(t, filename)
	}
}

// TestMobilePlatform_StartRecording_iOS — same anti-bluff contract as
// TestMobilePlatform_StartRecording_Android, dispatched against the
// xcrun simctl pipeline.
func TestMobilePlatform_StartRecording_iOS(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 14"
	platform.emulator = true
	platform.metrics["videos_taken"] = []string{}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "recording.mp4")

	err := platform.StartRecording(filename)

	if err == nil {
		assert.True(t, platform.recording)
		assert.Equal(t, filename, platform.metrics["recording_file"])
		assert.Equal(t, "ios", platform.metrics["recording_platform"])
		assert.Equal(t, "iPhone 14", platform.metrics["recording_device"])

		videos := platform.metrics["videos_taken"].([]string)
		assert.Contains(t, videos, filename)
	} else {
		assert.True(t, errors.Is(err, ErrMobileDeviceInteractionNotWired),
			"non-real-recording path MUST surface ErrMobileDeviceInteractionNotWired (got %v)", err)
		assert.NoFileExists(t, filename)
	}
}

// TestMobilePlatform_StartRecording_iOS_PhysicalDevice verifies iOS
// physical device recording surfaces ErrMobileDeviceInteractionNotWired
// rather than fabricating a placeholder file. Before round-29 the
// fallback wrote a "VIDEO RECORDING PLACEHOLDER" text file and returned
// (nil), letting a downstream §11.4.5 video-quality analyzer FAIL-bluff
// on a 0-frame "recording" (Bug-24 pattern) or PASS-bluff on a presence-
// only assertion.
func TestMobilePlatform_StartRecording_iOS_PhysicalDevice(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 13"
	platform.emulator = false // Physical device
	platform.metrics["videos_taken"] = []string{}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "recording.mp4")

	err := platform.StartRecording(filename)

	require := assert.New(t)
	require.Error(err, "iOS physical device recording MUST surface a sentinel error, not silently fabricate a placeholder")
	require.True(errors.Is(err, ErrMobileDeviceInteractionNotWired),
		"surfaced error MUST be ErrMobileDeviceInteractionNotWired (got %v)", err)
	require.NoFileExists(filename, "iOS physical device path MUST NOT write a description file in lieu of a real recording")

	notWired, ok := platform.metrics["mobile_video_not_wired"].([]string)
	require.True(ok)
	require.NotEmpty(notWired)
	require.Contains(notWired[0], "iOS physical device recording not yet implemented")
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

// TestMobilePlatform_createVideoPlaceholder is the round-29 §11.4
// anti-bluff regression test for the sentinel-default of
// createVideoPlaceholder. Before the round-29 audit the function wrote
// a description text file to disk and returned (nil), letting a test
// invoke MobilePlatform.StartRecording and score PASS while no frames
// had been captured (Bug-24 0-byte-mp4 PASS-bluff pattern). The
// helper MUST now refuse to fabricate evidence: NO file is written,
// ErrMobileDeviceInteractionNotWired is surfaced, and the gap is
// recorded in `metrics["mobile_video_not_wired"]` for the run report.
//
// Constitutional anchors: CONST-035 (anti-bluff), CONST-050(A)
// (no-fakes-beyond-unit-tests), Article XI §11.9 (forensic anchor),
// §11.4.5 (video-quality analysis comprehensiveness).
func TestMobilePlatform_createVideoPlaceholder(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "android"
	platform.device = "emulator-5554"
	platform.emulator = true

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "placeholder.mp4")

	err := platform.createVideoPlaceholder(filename, "Test reason")

	// Honest contract: helper MUST surface the gap, NOT write a file.
	require := assert.New(t)
	require.Error(err, "createVideoPlaceholder MUST surface a gap error, not silently pretend a recording was captured")
	require.True(errors.Is(err, ErrMobileDeviceInteractionNotWired),
		"surfaced error MUST be ErrMobileDeviceInteractionNotWired (got %v)", err)
	require.NoFileExists(filename, "createVideoPlaceholder MUST NOT write a description file in lieu of a real recording")

	// Verify the gap was recorded in metrics for the run report.
	notWired, ok := platform.metrics["mobile_video_not_wired"].([]string)
	require.True(ok, "metrics[\"mobile_video_not_wired\"] MUST be populated to surface the gap in the run report")
	require.Len(notWired, 1)
	require.Contains(notWired[0], filename)
	require.Contains(notWired[0], "Test reason")
}

// TestMobilePlatform_createMobileUIPlaceholder is the round-29 §11.4
// anti-bluff regression test for the sentinel-default of
// createMobileUIPlaceholder. Before the audit the function wrote a
// *.log text file to the working directory and returned (nil), letting
// MobilePlatform.Click / Fill / Submit score PASS while the device
// received zero interaction (CONTRACT-bluff). The helper MUST now
// refuse to fabricate success: NO file is written,
// ErrMobileDeviceInteractionNotWired is surfaced, and the gap is
// recorded in `metrics["mobile_ui_not_wired"]`.
//
// Constitutional anchors: CONST-035, CONST-050(A), Article XI §11.9.
func TestMobilePlatform_createMobileUIPlaceholder(t *testing.T) {
	platform := NewMobilePlatform()
	platform.platform = "ios"
	platform.device = "iPhone 14"
	platform.emulator = false

	err := platform.createMobileUIPlaceholder("click", "Submit Button", "Test reason")

	require := assert.New(t)
	require.Error(err, "createMobileUIPlaceholder MUST surface a gap error, not silently pretend the action succeeded")
	require.True(errors.Is(err, ErrMobileDeviceInteractionNotWired),
		"surfaced error MUST be ErrMobileDeviceInteractionNotWired (got %v)", err)

	// Verify the gap was recorded in metrics for the run report.
	notWired, ok := platform.metrics["mobile_ui_not_wired"].([]string)
	require.True(ok, "metrics[\"mobile_ui_not_wired\"] MUST be populated to surface the gap in the run report")
	require.Len(notWired, 1)
	require.Contains(notWired[0], "action=click")
	require.Contains(notWired[0], "selector=Submit Button")
	require.Contains(notWired[0], "Test reason")

	// The legacy "mobile_ui_placeholders" metric MUST NOT exist anymore —
	// it would imply description-file fabrication still happens.
	_, legacy := platform.metrics["mobile_ui_placeholders"]
	require.False(legacy, "legacy fabricated-placeholder metric MUST NOT be populated by the sentinel-default helper")
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
		t.Skip("Skipping Android integration test: adb not available")  // SKIP-OK: #integration-mode-only
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
		t.Skip("Skipping iOS integration test: xcrun not available")  // SKIP-OK: #integration-mode-only
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
