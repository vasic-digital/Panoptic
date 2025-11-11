package platforms

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"panoptic/internal/config"
)

// Test NewWebPlatform constructor
func TestNewWebPlatform(t *testing.T) {
	platform := NewWebPlatform()

	assert.NotNil(t, platform)
	assert.NotNil(t, platform.metrics)
	assert.NotNil(t, platform.vision)

	// Verify metrics are initialized
	assert.Contains(t, platform.metrics, "start_time")
	assert.Contains(t, platform.metrics, "click_actions")
	assert.Contains(t, platform.metrics, "screenshots_taken")
	assert.Contains(t, platform.metrics, "fill_actions")
	assert.Contains(t, platform.metrics, "submit_actions")
	assert.Contains(t, platform.metrics, "navigate_actions")
	assert.Contains(t, platform.metrics, "vision_actions")

	// Verify slices are initialized as empty
	assert.Equal(t, []string{}, platform.metrics["click_actions"])
	assert.Equal(t, []string{}, platform.metrics["screenshots_taken"])
	assert.Equal(t, []map[string]string{}, platform.metrics["fill_actions"])
	assert.Equal(t, []string{}, platform.metrics["submit_actions"])
	assert.Equal(t, []string{}, platform.metrics["navigate_actions"])
	assert.Equal(t, []string{}, platform.metrics["vision_actions"])
}

// Test Initialize with invalid timeout
func TestWebPlatform_Initialize_InvalidTimeout(t *testing.T) {
	platform := NewWebPlatform()

	tests := []struct {
		name    string
		timeout int
		wantErr bool
	}{
		{"zero timeout", 0, true},
		{"negative timeout", -1, true},
		{"valid timeout", 30, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := config.AppConfig{
				Name:    "Test App",
				Type:    "web",
				URL:     "https://example.com",
				Timeout: tt.timeout,
			}

			err := platform.Initialize(app)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout must be greater than 0")
			} else {
				// May fail if browser isn't available, but shouldn't be timeout error
				if err != nil {
					assert.NotContains(t, err.Error(), "timeout must be greater than 0")
				}
			}
		})
	}
}

// Test Navigate input validation
func TestWebPlatform_Navigate_Validation(t *testing.T) {
	platform := NewWebPlatform()

	tests := []struct {
		name        string
		url         string
		initPage    bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty URL",
			url:         "",
			initPage:    true,
			wantErr:     true,
			errContains: "URL cannot be empty",
		},
		{
			name:        "nil page",
			url:         "https://example.com",
			initPage:    false,
			wantErr:     true,
			errContains: "web page not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.initPage {
				platform.page = nil
			}

			err := platform.Navigate(tt.url)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				// Actual navigation may fail without browser
				if err != nil {
					assert.NotContains(t, err.Error(), tt.errContains)
				}
			}
		})
	}
}

// Test Click input validation
func TestWebPlatform_Click_Validation(t *testing.T) {
	platform := NewWebPlatform()

	tests := []struct {
		name        string
		selector    string
		initPage    bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty selector",
			selector:    "",
			initPage:    true,
			wantErr:     true,
			errContains: "selector cannot be empty",
		},
		{
			name:        "nil page",
			selector:    "button.test",
			initPage:    false,
			wantErr:     true,
			errContains: "web page not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.initPage {
				platform.page = nil
			}

			err := platform.Click(tt.selector)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// Test VisionClick input validation
func TestWebPlatform_VisionClick_Validation(t *testing.T) {
	platform := NewWebPlatform()

	tests := []struct {
		name        string
		elementType string
		text        string
		initPage    bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty element type",
			elementType: "",
			text:        "Click Me",
			initPage:    true,
			wantErr:     true,
			errContains: "element type cannot be empty",
		},
		{
			name:        "nil page",
			elementType: "button",
			text:        "Submit",
			initPage:    false,
			wantErr:     true,
			errContains: "web page not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.initPage {
				platform.page = nil
			}

			err := platform.VisionClick(tt.elementType, tt.text)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// Test Fill input validation
func TestWebPlatform_Fill_Validation(t *testing.T) {
	platform := NewWebPlatform()

	tests := []struct {
		name        string
		selector    string
		value       string
		initPage    bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty selector",
			selector:    "",
			value:       "test",
			initPage:    true,
			wantErr:     true,
			errContains: "selector cannot be empty",
		},
		{
			name:        "empty value",
			selector:    "input.test",
			value:       "",
			initPage:    true,
			wantErr:     true,
			errContains: "value cannot be empty",
		},
		{
			name:        "nil page",
			selector:    "input.test",
			value:       "test",
			initPage:    false,
			wantErr:     true,
			errContains: "web page not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.initPage {
				platform.page = nil
			}

			err := platform.Fill(tt.selector, tt.value)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// Test Submit with nil page
func TestWebPlatform_Submit_NilPage(t *testing.T) {
	platform := NewWebPlatform()
	platform.page = nil

	err := platform.Submit("form.test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "web page not initialized")
}

// Test Screenshot input validation
func TestWebPlatform_Screenshot_Validation(t *testing.T) {
	platform := NewWebPlatform()

	tests := []struct {
		name        string
		filename    string
		initPage    bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty filename",
			filename:    "",
			initPage:    true,
			wantErr:     true,
			errContains: "filename cannot be empty",
		},
		{
			name:        "nil page",
			filename:    "/tmp/test.png",
			initPage:    false,
			wantErr:     true,
			errContains: "web page not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.initPage {
				platform.page = nil
			}

			err := platform.Screenshot(tt.filename)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// Test StartRecording input validation
func TestWebPlatform_StartRecording_Validation(t *testing.T) {
	platform := NewWebPlatform()

	tests := []struct {
		name        string
		filename    string
		initPage    bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty filename",
			filename:    "",
			initPage:    true,
			wantErr:     true,
			errContains: "filename cannot be empty",
		},
		{
			name:        "nil page",
			filename:    "/tmp/test.mp4",
			initPage:    false,
			wantErr:     true,
			errContains: "web page not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.initPage {
				platform.page = nil
			}

			err := platform.StartRecording(tt.filename)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// Test StopRecording without active recording
func TestWebPlatform_StopRecording_NoRecording(t *testing.T) {
	platform := NewWebPlatform()
	platform.recording = false

	err := platform.StopRecording()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recording in progress")
}

// Test StopRecording with active recording
func TestWebPlatform_StopRecording_WithRecording(t *testing.T) {
	platform := NewWebPlatform()
	platform.recording = true
	platform.metrics["recording_started"] = time.Now()

	err := platform.StopRecording()

	assert.NoError(t, err)
	assert.False(t, platform.recording)
	assert.Contains(t, platform.metrics, "recording_stopped")
	assert.Contains(t, platform.metrics, "recording_duration")
}

// Test Wait function
func TestWebPlatform_Wait(t *testing.T) {
	platform := NewWebPlatform()

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

// Test GetMetrics
func TestWebPlatform_GetMetrics(t *testing.T) {
	platform := NewWebPlatform()

	// Simulate some actions
	platform.metrics["click_actions"] = []string{"button1", "button2"}
	platform.metrics["screenshots_taken"] = []string{"screen1.png"}
	platform.metrics["navigate_actions"] = []string{"https://example.com"}

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

	// Verify slices are present (even if empty)
	assert.NotNil(t, metrics["click_actions"])
	assert.NotNil(t, metrics["screenshots_taken"])
	assert.NotNil(t, metrics["fill_actions"])
	assert.NotNil(t, metrics["submit_actions"])
	assert.NotNil(t, metrics["navigate_actions"])
}

// Test GetMetrics with missing slices
func TestWebPlatform_GetMetrics_MissingSlices(t *testing.T) {
	platform := NewWebPlatform()

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
func TestWebPlatform_Close(t *testing.T) {
	platform := NewWebPlatform()

	// Close should not error even if browser/page are nil
	err := platform.Close()
	assert.NoError(t, err)
}

// Test GetPageState with nil page
func TestWebPlatform_GetPageState_NilPage(t *testing.T) {
	platform := NewWebPlatform()
	platform.page = nil

	state, err := platform.GetPageState()

	assert.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "web platform not initialized")
}

// Test takeScreenshotForVision with nil page
func TestWebPlatform_TakeScreenshotForVision_NilPage(t *testing.T) {
	platform := NewWebPlatform()
	platform.page = nil

	path, err := platform.takeScreenshotForVision()

	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "web page not initialized")
}

// Test GenerateVisionReport with nil vision detector
func TestWebPlatform_GenerateVisionReport_NilVision(t *testing.T) {
	platform := NewWebPlatform()
	platform.vision = nil

	err := platform.GenerateVisionReport("/tmp/report.html")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vision detector not initialized")
}

// Test ContainsString helper
func TestWebPlatform_ContainsString(t *testing.T) {
	platform := NewWebPlatform()

	tests := []struct {
		name   string
		text   string
		search string
		want   bool
	}{
		{"both non-empty", "Hello World", "World", true},
		{"empty text", "", "test", false},
		{"empty search", "test", "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := platform.ContainsString(tt.text, tt.search)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Test metrics tracking for actions
func TestWebPlatform_MetricsTracking(t *testing.T) {
	platform := NewWebPlatform()

	// Test navigate action tracking
	t.Run("Navigate metrics tracking", func(t *testing.T) {
		// Simulate navigation (without actual page)
		initialActions := platform.metrics["navigate_actions"].([]string)
		assert.Equal(t, 0, len(initialActions))

		// Note: Can't actually call Navigate without a page,
		// but we can verify the metrics structure is correct
	})

	// Test click action tracking
	t.Run("Click metrics tracking", func(t *testing.T) {
		initialActions := platform.metrics["click_actions"].([]string)
		assert.Equal(t, 0, len(initialActions))
	})

	// Test fill action tracking
	t.Run("Fill metrics tracking", func(t *testing.T) {
		initialActions := platform.metrics["fill_actions"].([]map[string]string)
		assert.Equal(t, 0, len(initialActions))
	})

	// Test submit action tracking
	t.Run("Submit metrics tracking", func(t *testing.T) {
		initialActions := platform.metrics["submit_actions"].([]string)
		assert.Equal(t, 0, len(initialActions))
	})
}

// Test recording file creation
func TestWebPlatform_StartRecording_FileCreation(t *testing.T) {
	platform := NewWebPlatform()

	// StartRecording requires page to be non-nil
	// Test file path validation
	err := platform.StartRecording("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filename cannot be empty")
}

// Test screenshot directory creation
func TestWebPlatform_Screenshot_DirectoryCreation(t *testing.T) {
	platform := NewWebPlatform()

	tmpDir := t.TempDir()
	screenshotPath := filepath.Join(tmpDir, "screenshots", "nested", "test.png")

	// Screenshot requires page to be non-nil
	err := platform.Screenshot(screenshotPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "web page not initialized")
}

// Test metrics duration calculation
func TestWebPlatform_MetricsDuration(t *testing.T) {
	platform := NewWebPlatform()

	startTime := time.Now().Add(-5 * time.Second)
	platform.metrics["start_time"] = startTime

	time.Sleep(100 * time.Millisecond)

	metrics := platform.GetMetrics()

	assert.Contains(t, metrics, "total_duration")
	duration := metrics["total_duration"].(time.Duration)
	assert.Greater(t, duration, 100*time.Millisecond)
}

// Test recording duration calculation
func TestWebPlatform_RecordingDuration(t *testing.T) {
	platform := NewWebPlatform()
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

// Test vision actions tracking
func TestWebPlatform_VisionActionsTracking(t *testing.T) {
	platform := NewWebPlatform()

	// Verify vision_actions is initialized
	assert.Contains(t, platform.metrics, "vision_actions")
	visionActions := platform.metrics["vision_actions"].([]string)
	assert.Equal(t, 0, len(visionActions))
}

// Test metrics initialization on construction
func TestWebPlatform_MetricsInitialization(t *testing.T) {
	platform := NewWebPlatform()

	// All action slices should be initialized
	requiredMetrics := []string{
		"click_actions",
		"screenshots_taken",
		"fill_actions",
		"submit_actions",
		"navigate_actions",
		"vision_actions",
		"start_time",
	}

	for _, metric := range requiredMetrics {
		assert.Contains(t, platform.metrics, metric, "Missing metric: %s", metric)
	}

	// start_time should be a time.Time
	assert.IsType(t, time.Time{}, platform.metrics["start_time"])
}

// Test vision detector initialization
func TestWebPlatform_VisionDetectorInit(t *testing.T) {
	platform := NewWebPlatform()

	assert.NotNil(t, platform.vision, "Vision detector should be initialized")
}

// Test multiple Close calls (should be idempotent)
func TestWebPlatform_Close_Idempotent(t *testing.T) {
	platform := NewWebPlatform()

	err1 := platform.Close()
	assert.NoError(t, err1)

	err2 := platform.Close()
	assert.NoError(t, err2, "Multiple Close() calls should not error")
}

// Test context cancellation
func TestWebPlatform_ContextCancellation(t *testing.T) {
	platform := NewWebPlatform()

	// After Close, context should be cancelled
	err := platform.Close()
	assert.NoError(t, err)

	// Verify cancel was called (context should be done if it was set)
	if platform.context != nil {
		select {
		case <-platform.context.Done():
			// Expected - context was cancelled
		case <-time.After(100 * time.Millisecond):
			t.Error("Context was not cancelled after Close()")
		}
	}
}
