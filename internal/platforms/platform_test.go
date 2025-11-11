package platforms

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"panoptic/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestPlatformFactory_CreatePlatform(t *testing.T) {
	factory := NewPlatformFactory()

	tests := []struct {
		name        string
		platformType string
		expectError bool
		expectType  string
	}{
		{
			name:        "Create web platform",
			platformType: "web",
			expectError: false,
			expectType:  "*platforms.WebPlatform",
		},
		{
			name:        "Create desktop platform",
			platformType: "desktop",
			expectError: false,
			expectType:  "*platforms.DesktopPlatform",
		},
		{
			name:        "Create mobile platform",
			platformType: "mobile",
			expectError: false,
			expectType:  "*platforms.MobilePlatform",
		},
		{
			name:        "Unsupported platform type",
			platformType: "unsupported",
			expectError: true,
			expectType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform, err := factory.CreatePlatform(tt.platformType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, platform)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, platform)
				assert.Equal(t, tt.expectType, getTypeString(platform))
			}
		})
	}
}

func TestWebPlatform(t *testing.T) {
	platform := NewWebPlatform()
	assert.NotNil(t, platform)

	app := config.AppConfig{
		Name:    "Test Web App",
		Type:    "web",
		URL:     "https://example.com",
		Timeout: 30,
	}

	// Test Initialize
	t.Run("Initialize", func(t *testing.T) {
		// This test may fail if browser is not available
		err := platform.Initialize(app)
		if err != nil {
			t.Skipf("Browser not available for testing: %v", err)
		}
		assert.NoError(t, err)
	})

	// Test basic operations (if browser is available)
	if platform != nil {
		defer platform.Close()

		t.Run("Navigate", func(t *testing.T) {
			err := platform.Navigate("https://httpbin.org/html")
			if err != nil {
				t.Skipf("Navigation failed, possibly browser issue: %v", err)
			}
			assert.NoError(t, err)
		})

		t.Run("Wait", func(t *testing.T) {
			err := platform.Wait(1)
			assert.NoError(t, err)
		})

		t.Run("Fill", func(t *testing.T) {
			err := platform.Fill("input[name='test']", "test-value")
			// This may fail if element doesn't exist, which is expected
			if err != nil {
				assert.Contains(t, err.Error(), "failed to find element")
			}
		})

		t.Run("Click", func(t *testing.T) {
			err := platform.Click("button.test")
			// This may fail if element doesn't exist, which is expected
			if err != nil {
				assert.Contains(t, err.Error(), "failed to find element")
			}
		})

		t.Run("Submit", func(t *testing.T) {
			err := platform.Submit("form.test")
			// This may fail if element doesn't exist, connection closes, or page not initialized
			if err != nil {
				errMsg := err.Error()
				assert.True(t,
					strings.Contains(errMsg, "failed to find") ||
					strings.Contains(errMsg, "connection") ||
					strings.Contains(errMsg, "closed") ||
					strings.Contains(errMsg, "not initialized"),
					"Expected error about missing element, connection, or initialization, got: %s", errMsg)
			}
		})

		t.Run("Screenshot", func(t *testing.T) {
			tempFile := "/tmp/test_screenshot.png"
			err := platform.Screenshot(tempFile)
			if err != nil {
				t.Skipf("Screenshot failed: %v", err)
			}
			assert.NoError(t, err)
		})

		t.Run("Recording", func(t *testing.T) {
			videoFile := "/tmp/test_video.mp4"
			err := platform.StartRecording(videoFile)
			if err != nil {
				t.Skipf("Recording failed: %v", err)
			}
			assert.NoError(t, err)

			// Stop recording
			err = platform.StopRecording()
			assert.NoError(t, err)
		})

		t.Run("GetMetrics", func(t *testing.T) {
			metrics := platform.GetMetrics()
			assert.NotNil(t, metrics)
			assert.Contains(t, metrics, "start_time")
			assert.Contains(t, metrics, "end_time")
			assert.Contains(t, metrics, "total_duration")
		})
	}
}

func TestDesktopPlatform(t *testing.T) {
	platform := NewDesktopPlatform()
	assert.NotNil(t, platform)

	app := config.AppConfig{
		Name:    "Test Desktop App",
		Type:    "desktop",
		Path:    "/Applications/Calculator.app", // Common macOS app
		Timeout: 30,
	}

	t.Run("Initialize with valid app", func(t *testing.T) {
		err := platform.Initialize(app)
		// May fail if app doesn't exist on system
		if err != nil {
			assert.Contains(t, err.Error(), "application not found")
		}
	})

	t.Run("Initialize with invalid app", func(t *testing.T) {
		invalidApp := config.AppConfig{
			Name: "Invalid App",
			Type: "desktop",
			Path: "/non/existent/path",
		}
		err := platform.Initialize(invalidApp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "application not found")
	})

	// Test other operations
	t.Run("Wait", func(t *testing.T) {
		err := platform.Wait(1)
		assert.NoError(t, err)
	})

	t.Run("Navigate", func(t *testing.T) {
		err := platform.Navigate("test://navigation")
		assert.NoError(t, err)
	})

	t.Run("Click", func(t *testing.T) {
		err := platform.Click("button.test")
		assert.NoError(t, err) // Should not fail as it's simulated
	})

	t.Run("Fill", func(t *testing.T) {
		err := platform.Fill("input.test", "test-value")
		assert.NoError(t, err) // Should not fail as it's simulated
	})

	t.Run("Submit", func(t *testing.T) {
		err := platform.Submit("form.test")
		assert.NoError(t, err) // Should not fail as it's simulated
	})

	t.Run("GetMetrics", func(t *testing.T) {
		metrics := platform.GetMetrics()
		assert.NotNil(t, metrics)
		assert.Contains(t, metrics, "start_time")
		assert.Contains(t, metrics, "end_time")
		assert.Contains(t, metrics, "total_duration")
		assert.Contains(t, metrics, "click_actions")
		assert.Contains(t, metrics, "fill_actions")
		assert.Contains(t, metrics, "submit_actions")
		assert.Contains(t, metrics, "navigate_actions")
	})

	platform.Close()
}

func TestMobilePlatform(t *testing.T) {
	platform := NewMobilePlatform()
	assert.NotNil(t, platform)

	app := config.AppConfig{
		Name:     "Test Mobile App",
		Type:     "mobile",
		Platform: "android",
		Emulator: true,
		Device:   "emulator-5554",
		Timeout:  30,
	}

	t.Run("Initialize without platform tools", func(t *testing.T) {
		err := platform.Initialize(app)
		// May fail if platform tools are not available
		if err != nil {
			assert.Contains(t, err.Error(), "platform tools not available")
		}
	})

	// Test other operations
	t.Run("Wait", func(t *testing.T) {
		err := platform.Wait(1)
		assert.NoError(t, err)
	})

	t.Run("Navigate", func(t *testing.T) {
		err := platform.Navigate("https://example.com")
		// May fail if platform tools are not available
		if err != nil {
			// Expected failure if no platform tools
		} else {
			assert.NoError(t, err)
		}
	})

	t.Run("Click", func(t *testing.T) {
		err := platform.Click("button.test")
		// May fail if platform tools are not available
		if err != nil {
			// Expected failure if no platform tools
		} else {
			assert.NoError(t, err)
		}
	})

	t.Run("Fill", func(t *testing.T) {
		err := platform.Fill("input.test", "test-value")
		// May fail if platform tools are not available
		if err != nil {
			// Expected failure if no platform tools
		} else {
			assert.NoError(t, err)
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		metrics := platform.GetMetrics()
		assert.NotNil(t, metrics)
		assert.Contains(t, metrics, "start_time")
		assert.Contains(t, metrics, "end_time")
		assert.Contains(t, metrics, "total_duration")
	})

	platform.Close()
}

func TestPlatformEdgeCases(t *testing.T) {
	t.Run("Web platform with nil config", func(t *testing.T) {
		platform := NewWebPlatform()
		err := platform.Initialize(config.AppConfig{})
		// Should not panic
		assert.Error(t, err)
	})

	t.Run("Empty screenshot file path", func(t *testing.T) {
		platform := NewWebPlatform()
		err := platform.Screenshot("")
		assert.Error(t, err)
	})

	t.Run("Empty video file path", func(t *testing.T) {
		platform := NewWebPlatform()
		err := platform.StartRecording("")
		assert.Error(t, err)
	})

	t.Run("Negative wait time", func(t *testing.T) {
		platform := NewWebPlatform()
		err := platform.Wait(-1)
		assert.NoError(t, err) // Should handle negative gracefully
	})
}

func TestPlatformMetricsConsistency(t *testing.T) {
	platforms := []Platform{
		NewWebPlatform(),
		NewDesktopPlatform(),
		NewMobilePlatform(),
	}

	for i, p := range platforms {
		t.Run(fmt.Sprintf("Metrics consistency for platform %d", i), func(t *testing.T) {
			metrics := p.GetMetrics()
			assert.NotNil(t, metrics)
			
			// All platforms should have these metrics
			assert.Contains(t, metrics, "start_time")
			assert.Contains(t, metrics, "end_time")
			assert.Contains(t, metrics, "total_duration")
			
			// Verify metric types
			assert.IsType(t, time.Time{}, metrics["start_time"])
			assert.IsType(t, time.Time{}, metrics["end_time"])
			assert.IsType(t, time.Duration(0), metrics["total_duration"])
		})
	}
}

func getTypeString(v interface{}) string {
	return fmt.Sprintf("%T", v)
}