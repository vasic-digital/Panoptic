package platforms

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"panoptic/internal/config"
)

// Test NewDesktopPlatform constructor
func TestNewDesktopPlatform(t *testing.T) {
	platform := NewDesktopPlatform()

	assert.NotNil(t, platform)
	assert.NotNil(t, platform.metrics)

	// Verify metrics are initialized
	assert.Contains(t, platform.metrics, "start_time")
	assert.Contains(t, platform.metrics, "click_actions")
	assert.Contains(t, platform.metrics, "screenshots_taken")
	assert.Contains(t, platform.metrics, "fill_actions")
	assert.Contains(t, platform.metrics, "submit_actions")
	assert.Contains(t, platform.metrics, "navigate_actions")
	assert.Contains(t, platform.metrics, "videos_taken")
	assert.Contains(t, platform.metrics, "ui_action_placeholders")

	// Verify slices are initialized as empty
	assert.Equal(t, []string{}, platform.metrics["click_actions"])
	assert.Equal(t, []string{}, platform.metrics["screenshots_taken"])
	assert.Equal(t, []map[string]string{}, platform.metrics["fill_actions"])
	assert.Equal(t, []string{}, platform.metrics["submit_actions"])
	assert.Equal(t, []string{}, platform.metrics["navigate_actions"])
	assert.Equal(t, []string{}, platform.metrics["videos_taken"])
	assert.Equal(t, []string{}, platform.metrics["ui_action_placeholders"])
}

// Test Initialize with valid application path
func TestDesktopPlatform_Initialize_ValidPath(t *testing.T) {
	platform := NewDesktopPlatform()

	// Create a temporary file to simulate an app
	tmpFile := filepath.Join(t.TempDir(), "test_app")
	err := os.WriteFile(tmpFile, []byte("test"), 0755)
	assert.NoError(t, err)

	app := config.AppConfig{
		Name:    "Test App",
		Type:    "desktop",
		Path:    tmpFile,
		Timeout: 30,
	}

	err = platform.Initialize(app)

	assert.NoError(t, err)
	assert.Equal(t, tmpFile, platform.appPath)
	assert.Contains(t, platform.metrics, "app_path")
	assert.Equal(t, tmpFile, platform.metrics["app_path"])
}

// Test Initialize with non-existent application path
func TestDesktopPlatform_Initialize_NonExistentPath(t *testing.T) {
	platform := NewDesktopPlatform()

	app := config.AppConfig{
		Name:    "Non-existent App",
		Type:    "desktop",
		Path:    "/non/existent/path/to/app",
		Timeout: 30,
	}

	err := platform.Initialize(app)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application not found at path")
}

// Test Navigate input validation
func TestDesktopPlatform_Navigate_Validation(t *testing.T) {
	platform := NewDesktopPlatform()

	tests := []struct {
		name        string
		url         string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty URL",
			url:         "",
			wantErr:     true,
			errContains: "url cannot be empty",
		},
		{
			name:    "valid URL",
			url:     "test://navigation",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := platform.Navigate(tt.url)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Navigate metrics tracking
func TestDesktopPlatform_Navigate_MetricsTracking(t *testing.T) {
	platform := NewDesktopPlatform()

	initialActions := platform.metrics["navigate_actions"].([]string)
	assert.Equal(t, 0, len(initialActions))

	err := platform.Navigate("test://view1")
	assert.NoError(t, err)

	actions := platform.metrics["navigate_actions"].([]string)
	assert.Equal(t, 1, len(actions))
	assert.Equal(t, "test://view1", actions[0])

	// Navigate again
	err = platform.Navigate("test://view2")
	assert.NoError(t, err)

	actions = platform.metrics["navigate_actions"].([]string)
	assert.Equal(t, 2, len(actions))
	assert.Equal(t, "test://view2", actions[1])
}

// Test Click input validation
func TestDesktopPlatform_Click_Validation(t *testing.T) {
	platform := NewDesktopPlatform()

	tests := []struct {
		name        string
		selector    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty selector",
			selector:    "",
			wantErr:     true,
			errContains: "selector cannot be empty",
		},
		{
			name:     "valid selector",
			selector: "button.test",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := platform.Click(tt.selector)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				// May succeed or fail depending on platform tools
				// Just verify it doesn't panic
			}
		})
	}
}

// Test Click metrics tracking
func TestDesktopPlatform_Click_MetricsTracking(t *testing.T) {
	platform := NewDesktopPlatform()

	initialActions := platform.metrics["click_actions"].([]string)
	assert.Equal(t, 0, len(initialActions))

	// Click will attempt to execute but may fail - that's okay
	// We're testing metrics tracking
	_ = platform.Click("button1")

	actions := platform.metrics["click_actions"].([]string)
	assert.Equal(t, 1, len(actions))
	assert.Equal(t, "button1", actions[0])
}

// Test Fill input validation
func TestDesktopPlatform_Fill_Validation(t *testing.T) {
	platform := NewDesktopPlatform()

	tests := []struct {
		name        string
		selector    string
		value       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty selector",
			selector:    "",
			value:       "test",
			wantErr:     true,
			errContains: "selector cannot be empty",
		},
		{
			name:        "empty value",
			selector:    "input.test",
			value:       "",
			wantErr:     true,
			errContains: "value cannot be empty",
		},
		{
			name:     "valid input",
			selector: "input.test",
			value:    "test value",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := platform.Fill(tt.selector, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Fill metrics tracking
func TestDesktopPlatform_Fill_MetricsTracking(t *testing.T) {
	platform := NewDesktopPlatform()

	initialActions := platform.metrics["fill_actions"].([]map[string]string)
	assert.Equal(t, 0, len(initialActions))

	err := platform.Fill("input1", "value1")
	assert.NoError(t, err)

	actions := platform.metrics["fill_actions"].([]map[string]string)
	assert.Equal(t, 1, len(actions))
	assert.Equal(t, "input1", actions[0]["selector"])
	assert.Equal(t, "value1", actions[0]["value"])

	// Fill again
	err = platform.Fill("input2", "value2")
	assert.NoError(t, err)

	actions = platform.metrics["fill_actions"].([]map[string]string)
	assert.Equal(t, 2, len(actions))
	assert.Equal(t, "input2", actions[1]["selector"])
	assert.Equal(t, "value2", actions[1]["value"])
}

// Test Submit
func TestDesktopPlatform_Submit(t *testing.T) {
	platform := NewDesktopPlatform()

	initialActions := platform.metrics["submit_actions"].([]string)
	assert.Equal(t, 0, len(initialActions))

	err := platform.Submit("form.test")
	assert.NoError(t, err)

	actions := platform.metrics["submit_actions"].([]string)
	assert.Equal(t, 1, len(actions))
	assert.Equal(t, "form.test", actions[0])
}

// Test Wait function
func TestDesktopPlatform_Wait(t *testing.T) {
	platform := NewDesktopPlatform()

	tests := []struct {
		name     string
		duration int
	}{
		{"zero duration", 0},
		{"one second", 1},
		{"negative duration", -1}, // Should handle gracefully
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := platform.Wait(tt.duration)
			elapsed := time.Since(start)

			assert.NoError(t, err)

			// For positive durations, verify timing
			if tt.duration > 0 {
				expectedDuration := time.Duration(tt.duration) * time.Second
				assert.GreaterOrEqual(t, elapsed, expectedDuration)
			}
		})
	}
}

// Test Screenshot input validation
func TestDesktopPlatform_Screenshot_Validation(t *testing.T) {
	platform := NewDesktopPlatform()

	tests := []struct {
		name        string
		filename    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty filename",
			filename:    "",
			wantErr:     true,
			errContains: "filename cannot be empty",
		},
		{
			name:     "valid filename",
			filename: filepath.Join(t.TempDir(), "screenshot.png"),
			wantErr:  false, // May fail if tool not available, but not due to validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := platform.Screenshot(tt.filename)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				// Screenshot may fail if screencapture not available
				// Just verify it doesn't panic on validation
			}
		})
	}
}

// Test StartRecording input validation
func TestDesktopPlatform_StartRecording_Validation(t *testing.T) {
	platform := NewDesktopPlatform()

	tests := []struct {
		name        string
		filename    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty filename",
			filename:    "",
			wantErr:     true,
			errContains: "filename cannot be empty",
		},
		{
			name:     "valid filename",
			filename: filepath.Join(t.TempDir(), "recording.mp4"),
			wantErr:  false, // May create placeholder if tool not available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := platform.StartRecording(tt.filename)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				// Recording may create placeholder or fail gracefully
				// Just verify validation works
			}
		})
	}
}

// Test StopRecording without active recording
func TestDesktopPlatform_StopRecording_NoRecording(t *testing.T) {
	platform := NewDesktopPlatform()
	platform.recording = false

	err := platform.StopRecording()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recording in progress")
}

// Test StopRecording with active recording
func TestDesktopPlatform_StopRecording_WithRecording(t *testing.T) {
	platform := NewDesktopPlatform()
	platform.recording = true
	platform.metrics["recording_started"] = time.Now()

	err := platform.StopRecording()

	assert.NoError(t, err)
	assert.False(t, platform.recording)
	assert.Contains(t, platform.metrics, "recording_stopped")
	assert.Contains(t, platform.metrics, "recording_duration")
}

// Test GetMetrics
func TestDesktopPlatform_GetMetrics(t *testing.T) {
	platform := NewDesktopPlatform()

	// Simulate some actions
	platform.metrics["click_actions"] = []string{"button1", "button2"}
	platform.metrics["screenshots_taken"] = []string{"screen1.png"}
	platform.metrics["navigate_actions"] = []string{"test://view1"}

	metrics := platform.GetMetrics()

	assert.NotNil(t, metrics)

	// Check required fields
	assert.Contains(t, metrics, "start_time")
	assert.Contains(t, metrics, "end_time")
	assert.Contains(t, metrics, "total_duration")
	assert.Contains(t, metrics, "click_actions")
	assert.Contains(t, metrics, "screenshots_taken")
	assert.Contains(t, metrics, "fill_actions")
	assert.Contains(t, metrics, "submit_actions")
	assert.Contains(t, metrics, "navigate_actions")

	// Verify types
	assert.IsType(t, time.Time{}, metrics["start_time"])
	assert.IsType(t, time.Time{}, metrics["end_time"])
	assert.IsType(t, time.Duration(0), metrics["total_duration"])

	// Verify slices are present
	assert.NotNil(t, metrics["click_actions"])
	assert.NotNil(t, metrics["screenshots_taken"])
	assert.NotNil(t, metrics["fill_actions"])
	assert.NotNil(t, metrics["submit_actions"])
	assert.NotNil(t, metrics["navigate_actions"])
}

// Test GetMetrics with missing slices
func TestDesktopPlatform_GetMetrics_MissingSlices(t *testing.T) {
	platform := NewDesktopPlatform()

	// Remove some metrics to test initialization
	delete(platform.metrics, "click_actions")
	delete(platform.metrics, "screenshots_taken")

	metrics := platform.GetMetrics()

	// Should initialize missing slices
	assert.Contains(t, metrics, "click_actions")
	assert.Contains(t, metrics, "screenshots_taken")
	assert.Contains(t, metrics, "fill_actions")
	assert.Contains(t, metrics, "submit_actions")
	assert.Contains(t, metrics, "navigate_actions")

	// Verify they're empty slices, not nil
	assert.Equal(t, []string{}, metrics["click_actions"])
	assert.Equal(t, []string{}, metrics["screenshots_taken"])
}

// Test Close
func TestDesktopPlatform_Close(t *testing.T) {
	platform := NewDesktopPlatform()

	// Close should not error
	err := platform.Close()
	assert.NoError(t, err)
}

// Test createVideoPlaceholder
func TestDesktopPlatform_CreateVideoPlaceholder(t *testing.T) {
	platform := NewDesktopPlatform()

	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "test_video.mp4")

	err := platform.createVideoPlaceholder(videoPath, "Test reason")

	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(videoPath)
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(videoPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "PANOPTIC VIDEO RECORDING PLACEHOLDER")
	assert.Contains(t, string(content), "Desktop Platform")
	assert.Contains(t, string(content), "Test reason")
	assert.Contains(t, string(content), runtime.GOOS)
}

// Test createUIActionPlaceholder
func TestDesktopPlatform_CreateUIActionPlaceholder(t *testing.T) {
	platform := NewDesktopPlatform()

	// Clean up any files created during test
	defer func() {
		files, _ := filepath.Glob("desktop_ui_action_*.log")
		for _, f := range files {
			os.Remove(f)
		}
	}()

	err := platform.createUIActionPlaceholder("click", "button.test", "Test reason")

	assert.NoError(t, err)

	// Verify placeholder was tracked
	placeholders := platform.metrics["ui_action_placeholders"].([]string)
	assert.Greater(t, len(placeholders), 0)

	// Verify file was created
	placeholderFile := placeholders[len(placeholders)-1]
	_, err = os.Stat(placeholderFile)
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(placeholderFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "DESKTOP UI ACTION PLACEHOLDER")
	assert.Contains(t, string(content), "Action: click")
	assert.Contains(t, string(content), "Selector: button.test")
	assert.Contains(t, string(content), "Test reason")
}

// Test metrics initialization on construction
func TestDesktopPlatform_MetricsInitialization(t *testing.T) {
	platform := NewDesktopPlatform()

	// All action slices should be initialized
	requiredMetrics := []string{
		"click_actions",
		"screenshots_taken",
		"fill_actions",
		"submit_actions",
		"navigate_actions",
		"videos_taken",
		"ui_action_placeholders",
		"start_time",
	}

	for _, metric := range requiredMetrics {
		assert.Contains(t, platform.metrics, metric, "Missing metric: %s", metric)
	}

	// start_time should be a time.Time
	assert.IsType(t, time.Time{}, platform.metrics["start_time"])
}

// Test multiple Close calls (should be idempotent)
func TestDesktopPlatform_Close_Idempotent(t *testing.T) {
	platform := NewDesktopPlatform()

	err1 := platform.Close()
	assert.NoError(t, err1)

	err2 := platform.Close()
	assert.NoError(t, err2, "Multiple Close() calls should not error")
}

// Test metrics duration calculation
func TestDesktopPlatform_MetricsDuration(t *testing.T) {
	platform := NewDesktopPlatform()

	startTime := time.Now().Add(-5 * time.Second)
	platform.metrics["start_time"] = startTime

	time.Sleep(100 * time.Millisecond)

	metrics := platform.GetMetrics()

	assert.Contains(t, metrics, "total_duration")
	duration := metrics["total_duration"].(time.Duration)
	assert.Greater(t, duration, 5*time.Second)
}

// Test recording duration calculation
func TestDesktopPlatform_RecordingDuration(t *testing.T) {
	platform := NewDesktopPlatform()
	platform.recording = true

	startTime := time.Now().Add(-2 * time.Second)
	platform.metrics["recording_started"] = startTime

	time.Sleep(100 * time.Millisecond)

	err := platform.StopRecording()
	assert.NoError(t, err)

	assert.Contains(t, platform.metrics, "recording_duration")
	duration := platform.metrics["recording_duration"].(time.Duration)
	assert.Greater(t, duration, 2*time.Second)
}

// Test StartRecording creates directory
func TestDesktopPlatform_StartRecording_CreatesDirectory(t *testing.T) {
	platform := NewDesktopPlatform()

	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "videos", "nested", "recording.mp4")

	err := platform.StartRecording(videoPath)

	// Should succeed or create placeholder
	// Just verify directory was created
	if err == nil {
		_, dirErr := os.Stat(filepath.Dir(videoPath))
		assert.NoError(t, dirErr, "Video directory should be created")
	}
}

// Test recording state tracking
func TestDesktopPlatform_RecordingState(t *testing.T) {
	platform := NewDesktopPlatform()

	assert.False(t, platform.recording, "Should not be recording initially")

	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "test.mp4")

	_ = platform.StartRecording(videoPath)

	// Recording flag should be set
	assert.True(t, platform.recording, "Should be recording after StartRecording")

	err := platform.StopRecording()
	assert.NoError(t, err)

	assert.False(t, platform.recording, "Should not be recording after StopRecording")
}

// Test platform detection
func TestDesktopPlatform_PlatformDetection(t *testing.T) {
	// Just verify runtime.GOOS returns expected values
	validPlatforms := []string{"darwin", "windows", "linux", "freebsd", "openbsd"}
	assert.Contains(t, validPlatforms, runtime.GOOS)
}

// Test Initialize sets app_path metric
func TestDesktopPlatform_Initialize_SetsMetrics(t *testing.T) {
	platform := NewDesktopPlatform()

	tmpFile := filepath.Join(t.TempDir(), "test_app")
	err := os.WriteFile(tmpFile, []byte("test"), 0755)
	assert.NoError(t, err)

	app := config.AppConfig{
		Path: tmpFile,
	}

	err = platform.Initialize(app)
	assert.NoError(t, err)

	// Verify metrics were set
	assert.Equal(t, tmpFile, platform.metrics["app_path"])
	assert.IsType(t, time.Time{}, platform.metrics["start_time"])
}

// Test videos_taken tracking
func TestDesktopPlatform_VideosTakenTracking(t *testing.T) {
	platform := NewDesktopPlatform()

	initialVideos := platform.metrics["videos_taken"].([]string)
	assert.Equal(t, 0, len(initialVideos))

	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "video1.mp4")

	_ = platform.StartRecording(videoPath)

	videos := platform.metrics["videos_taken"].([]string)
	assert.Equal(t, 1, len(videos))
	assert.Equal(t, videoPath, videos[0])
}
