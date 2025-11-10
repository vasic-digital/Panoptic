package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name     string       `yaml:"name"`
	Output   string       `yaml:"output"`
	Apps     []AppConfig  `yaml:"apps"`
	Actions  []Action     `yaml:"actions"`
	Settings Settings     `yaml:"settings"`
}

type AppConfig struct {
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"` // web, desktop, mobile
	URL         string            `yaml:"url"`
	Path        string            `yaml:"path"`
	Platform    string            `yaml:"platform"` // ios, android, windows, macos, linux
	Emulator    bool              `yaml:"emulator"`
	Device      string            `yaml:"device"`
	Timeout     int               `yaml:"timeout"`
	Environment map[string]string `yaml:"environment"`
}

type Action struct {
	Name        string                 `yaml:"name"`
	Type        string                 `yaml:"type"` // navigate, click, fill, submit, wait, screenshot, record
	Target      string                 `yaml:"target"`
	Value       string                 `yaml:"value"`
	Selector    string                 `yaml:"selector"`
	WaitTime    int                    `yaml:"wait_time"`
	Parameters  map[string]interface{} `yaml:"parameters"`
	Screenshot  bool                   `yaml:"screenshot"`
	Record      bool                   `yaml:"record"`
	Duration    int                    `yaml:"duration"`
}

type Settings struct {
	ScreenshotFormat string `yaml:"screenshot_format"` // png, jpg
	VideoFormat      string `yaml:"video_format"`      // mp4, webm
	Quality          int    `yaml:"quality"`           // 1-100
	Headless         bool   `yaml:"headless"`
	WindowWidth      int    `yaml:"window_width"`
	WindowHeight     int    `yaml:"window_height"`
	MobileDevice     string `yaml:"mobile_device"`
	EnableMetrics    bool   `yaml:"enable_metrics"`
	LogLevel         string `yaml:"log_level"`
	
	// AI-Enhanced Testing Settings
	AITesting        *AITestingSettings `yaml:"ai_testing,omitempty"`
}

type AITestingSettings struct {
	EnableErrorDetection   bool    `yaml:"enable_error_detection"`
	EnableTestGeneration  bool    `yaml:"enable_test_generation"`
	EnableVisionAnalysis   bool    `yaml:"enable_vision_analysis"`
	AutoGenerateTests      bool    `yaml:"auto_generate_tests"`
	SmartErrorRecovery     bool    `yaml:"smart_error_recovery"`
	AdaptiveTestPriority   bool    `yaml:"adaptive_test_priority"`
	ConfidenceThreshold    float64 `yaml:"confidence_threshold"`
	MaxGeneratedTests      int     `yaml:"max_generated_tests"`
	EnableLearning         bool    `yaml:"enable_learning"`
}

func Load(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Settings.ScreenshotFormat == "" {
		config.Settings.ScreenshotFormat = "png"
	}
	if config.Settings.VideoFormat == "" {
		config.Settings.VideoFormat = "mp4"
	}
	if config.Settings.Quality == 0 {
		config.Settings.Quality = 80
	}
	if config.Settings.WindowWidth == 0 {
		config.Settings.WindowWidth = 1920
	}
	if config.Settings.WindowHeight == 0 {
		config.Settings.WindowHeight = 1080
	}
	if config.Settings.LogLevel == "" {
		config.Settings.LogLevel = "info"
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if len(c.Apps) == 0 {
		return fmt.Errorf("at least one application must be configured")
	}

	for _, app := range c.Apps {
		if app.Name == "" {
			return fmt.Errorf("application name is required")
		}
		if app.Type == "" {
			return fmt.Errorf("application type is required")
		}
		
		switch app.Type {
		case "web":
			if app.URL == "" {
				return fmt.Errorf("URL is required for web applications")
			}
		case "desktop":
			if app.Path == "" {
				return fmt.Errorf("path is required for desktop applications")
			}
		case "mobile":
			if app.Platform == "" {
				return fmt.Errorf("platform is required for mobile applications")
			}
		default:
			return fmt.Errorf("unknown application type: %s", app.Type)
		}
	}

	return nil
}