package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectError    bool
		expectedName   string
		expectedQuality int
		expectedLogLevel string
	}{
		{
			name: "Valid complete config",
			configContent: `
name: "Complete Test Config"
output: "/tmp/test_output"
apps:
  - name: "Web App"
    type: "web"
    url: "https://example.com"
    timeout: 30
  - name: "Desktop App"
    type: "desktop"
    path: "/Applications/Test.app"
    timeout: 60
  - name: "Mobile App"
    type: "mobile"
    platform: "android"
    emulator: true
    device: "emulator-5554"
actions:
  - name: "navigate_home"
    type: "navigate"
    value: "https://example.com"
  - name: "click_button"
    type: "click"
    selector: "#submit"
  - name: "fill_form"
    type: "fill"
    selector: "input[name='username']"
    value: "testuser"
  - name: "wait_load"
    type: "wait"
    wait_time: 3
  - name: "take_screenshot"
    type: "screenshot"
  - name: "record_video"
    type: "record"
    duration: 30
settings:
  screenshot_format: "png"
  video_format: "mp4"
  quality: 90
  headless: true
  window_width: 1920
  window_height: 1080
  enable_metrics: true
  log_level: "debug"
`,
			expectError: false,
			expectedName: "Complete Test Config",
			expectedQuality: 90,
			expectedLogLevel: "debug",
		},
		{
			name: "Minimal valid config",
			configContent: `
name: "Minimal Config"
apps:
  - name: "Test App"
    type: "web"
    url: "https://example.com"
`,
			expectError: false,
			expectedName: "Minimal Config",
		},
		{
			name: "Invalid YAML",
			configContent: `
name: "Invalid Config"
apps:
  - name: "Test App"
    type: "web"
    url: "https://example.com"
  invalid_yaml: [
`,
			expectError: true,
		},
		{
			name: "Empty config",
			configContent: ``,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			if tt.configContent != "" {
				_, err = tmpFile.WriteString(tt.configContent)
				require.NoError(t, err)
			}
			tmpFile.Close()

			config, err := Load(tmpFile.Name())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				if tt.expectedName != "" {
					assert.Equal(t, tt.expectedName, config.Name)
				}
				
				// Test default values (only if not specified in config)
				if config.Settings.ScreenshotFormat == "" {
					assert.Equal(t, "png", config.Settings.ScreenshotFormat)
				}
				if config.Settings.VideoFormat == "" {
					assert.Equal(t, "mp4", config.Settings.VideoFormat)
				}
				if config.Settings.Quality == 0 {
					assert.Equal(t, 80, config.Settings.Quality)
				}
				if config.Settings.WindowWidth == 0 {
					assert.Equal(t, 1920, config.Settings.WindowWidth)
				}
				if config.Settings.WindowHeight == 0 {
					assert.Equal(t, 1080, config.Settings.WindowHeight)
				}
				if config.Settings.LogLevel == "" {
					assert.Equal(t, "info", config.Settings.LogLevel)
				}
				
				// Check specific expected values
				if tt.expectedQuality != 0 {
					assert.Equal(t, tt.expectedQuality, config.Settings.Quality)
				}
				if tt.expectedLogLevel != "" {
					assert.Equal(t, tt.expectedLogLevel, config.Settings.LogLevel)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid web app",
			config: Config{
				Apps: []AppConfig{
					{
						Name: "Test Web App",
						Type: "web",
						URL:  "https://example.com",
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Valid desktop app",
			config: Config{
				Apps: []AppConfig{
					{
						Name: "Test Desktop App",
						Type: "desktop",
						Path: "/Applications/Test.app",
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Valid mobile app",
			config: Config{
				Apps: []AppConfig{
					{
						Name:     "Test Mobile App",
						Type:     "mobile",
						Platform: "android",
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Multiple valid apps",
			config: Config{
				Apps: []AppConfig{
					{
						Name: "Web App",
						Type: "web",
						URL:  "https://example.com",
					},
					{
						Name:     "Mobile App",
						Type:     "mobile",
						Platform: "ios",
					},
				},
			},
			expectErr: false,
		},
		{
			name: "No apps",
			config: Config{
				Apps: []AppConfig{},
			},
			expectErr: true,
			errMsg:    "at least one application must be configured",
		},
		{
			name: "Web app without URL",
			config: Config{
				Apps: []AppConfig{
					{
						Name: "Test App",
						Type: "web",
					},
				},
			},
			expectErr: true,
			errMsg:    "URL is required for web applications",
		},
		{
			name: "Desktop app without path",
			config: Config{
				Apps: []AppConfig{
					{
						Name: "Test App",
						Type: "desktop",
					},
				},
			},
			expectErr: true,
			errMsg:    "path is required for desktop applications",
		},
		{
			name: "Mobile app without platform",
			config: Config{
				Apps: []AppConfig{
					{
						Name: "Test App",
						Type: "mobile",
					},
				},
			},
			expectErr: true,
			errMsg:    "platform is required for mobile applications",
		},
		{
			name: "Unknown app type",
			config: Config{
				Apps: []AppConfig{
					{
						Name: "Test App",
						Type: "unknown",
					},
				},
			},
			expectErr: true,
			errMsg:    "unknown application type: unknown",
		},
		{
			name: "App without name",
			config: Config{
				Apps: []AppConfig{
					{
						Type: "web",
						URL:  "https://example.com",
					},
				},
			},
			expectErr: true,
			errMsg:    "application name is required",
		},
		{
			name: "App without type",
			config: Config{
				Apps: []AppConfig{
					{
						Name: "Test App",
						URL:  "https://example.com",
					},
				},
			},
			expectErr: true,
			errMsg:    "application type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{
		Apps: []AppConfig{
			{
				Name: "Test App",
				Type: "web",
				URL:  "https://example.com",
			},
		},
	}

	err := config.Validate()
	assert.NoError(t, err)

	// Test that defaults are applied properly (Load function applies them, not Validate)
	// So we need to call Load to test defaults
	configContent := `
name: "Test"
apps:
  - name: "Test App"
    type: "web"
    url: "https://example.com"
`

	tmpFile, err := os.CreateTemp("", "test-defaults-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	loadedConfig, err := Load(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "png", loadedConfig.Settings.ScreenshotFormat)
	assert.Equal(t, "mp4", loadedConfig.Settings.VideoFormat)
	assert.Equal(t, 80, loadedConfig.Settings.Quality)
	assert.Equal(t, 1920, loadedConfig.Settings.WindowWidth)
	assert.Equal(t, 1080, loadedConfig.Settings.WindowHeight)
	assert.Equal(t, "info", loadedConfig.Settings.LogLevel)
}

func TestActionGetNavigateURL(t *testing.T) {
	tests := []struct {
		name     string
		action   Action
		expected string
	}{
		{
			name:     "URL field takes precedence",
			action:   Action{URL: "https://url-field.com", Value: "https://value-field.com"},
			expected: "https://url-field.com",
		},
		{
			name:     "Falls back to Value when URL is empty",
			action:   Action{Value: "https://value-field.com"},
			expected: "https://value-field.com",
		},
		{
			name:     "URL field only",
			action:   Action{URL: "https://url-only.com"},
			expected: "https://url-only.com",
		},
		{
			name:     "Both empty returns empty string",
			action:   Action{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action.GetNavigateURL()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPerAppActions(t *testing.T) {
	t.Run("Load config with per-app actions", func(t *testing.T) {
		configContent := `
name: "Per-App Actions Test"
apps:
  - name: "Admin Console"
    type: "web"
    url: "http://localhost:3001"
    actions:
      - name: "navigate_login"
        type: "navigate"
        url: "http://localhost:3001/login"
      - name: "fill_username"
        type: "fill"
        selector: "input[name='username']"
        value: "admin"
  - name: "Web App"
    type: "web"
    url: "http://localhost:3000"
    actions:
      - name: "navigate_home"
        type: "navigate"
        url: "http://localhost:3000"
`
		tmpFile, err := os.CreateTemp("", "test-perapp-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(configContent)
		require.NoError(t, err)
		tmpFile.Close()

		cfg, err := Load(tmpFile.Name())
		assert.NoError(t, err)
		assert.Len(t, cfg.Apps, 2)
		assert.Len(t, cfg.Apps[0].Actions, 2)
		assert.Len(t, cfg.Apps[1].Actions, 1)
		assert.Equal(t, "navigate_login", cfg.Apps[0].Actions[0].Name)
		assert.Equal(t, "http://localhost:3001/login", cfg.Apps[0].Actions[0].URL)
	})

	t.Run("Validate rejects navigate without URL in per-app actions", func(t *testing.T) {
		cfg := Config{
			Apps: []AppConfig{
				{
					Name: "Test App",
					Type: "web",
					URL:  "http://localhost:3000",
					Actions: []Action{
						{Name: "bad_navigate", Type: "navigate"}, // Missing URL and Value
					},
				},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URL or value is required for navigate action")
	})

	t.Run("Validate rejects global navigate without URL", func(t *testing.T) {
		cfg := Config{
			Apps: []AppConfig{
				{Name: "App", Type: "web", URL: "http://localhost:3000"},
			},
			Actions: []Action{
				{Name: "bad_nav", Type: "navigate"}, // Missing URL and Value
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URL or value is required for navigate action")
	})
}

func TestGetActionsForApp(t *testing.T) {
	globalActions := []Action{
		{Name: "global_action", Type: "navigate", URL: "https://global.com"},
	}
	perAppActions := []Action{
		{Name: "app_action", Type: "navigate", URL: "https://app.com"},
	}

	cfg := Config{
		Actions: globalActions,
		Apps: []AppConfig{
			{Name: "AppWithActions", Type: "web", URL: "http://localhost", Actions: perAppActions},
			{Name: "AppWithoutActions", Type: "web", URL: "http://localhost"},
		},
	}

	t.Run("Returns per-app actions when defined", func(t *testing.T) {
		actions := cfg.GetActionsForApp(cfg.Apps[0])
		assert.Len(t, actions, 1)
		assert.Equal(t, "app_action", actions[0].Name)
	})

	t.Run("Falls back to global actions when no per-app actions", func(t *testing.T) {
		actions := cfg.GetActionsForApp(cfg.Apps[1])
		assert.Len(t, actions, 1)
		assert.Equal(t, "global_action", actions[0].Name)
	})
}

func TestLoadConfigWithURLField(t *testing.T) {
	configContent := `
name: "URL Field Test"
apps:
  - name: "Test"
    type: "web"
    url: "http://localhost:3001"
actions:
  - name: "nav_with_url"
    type: "navigate"
    url: "http://localhost:3001/login"
  - name: "nav_with_value"
    type: "navigate"
    value: "http://localhost:3001/dashboard"
`
	tmpFile, err := os.CreateTemp("", "test-url-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:3001/login", cfg.Actions[0].URL)
	assert.Equal(t, "http://localhost:3001/login", cfg.Actions[0].GetNavigateURL())
	assert.Equal(t, "http://localhost:3001/dashboard", cfg.Actions[1].GetNavigateURL())
}

func TestEdgeCases(t *testing.T) {
	t.Run("Non-existent file", func(t *testing.T) {
		_, err := Load("/non/existent/file.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})

	t.Run("Config with special characters", func(t *testing.T) {
		configContent := `
name: "Test & Special @#$%^&*()_+"
output: "/tmp/special path with spaces/test"
apps:
  - name: "App with quotes 'test'"
    type: "web"
    url: "https://example.com?param=value&other=test"
`

		tmpFile, err := os.CreateTemp("", "test-special-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(configContent)
		require.NoError(t, err)
		tmpFile.Close()

		config, err := Load(tmpFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Contains(t, config.Name, "&")
		assert.Contains(t, config.Output, "spaces")
	})
}