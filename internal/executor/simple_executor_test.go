package executor

import (
	"testing"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewExecutor_Simple(t *testing.T) {
	cfg := &config.Config{
		Name: "Test Config",
		Apps: []config.AppConfig{
			{
				Name: "Test App",
				Type: "web",
				URL:  "https://example.com",
			},
		},
		Actions: []config.Action{
			{
				Name: "test_action",
				Type: "wait",
				WaitTime: 1,
			},
		},
	}

	outputDir := t.TempDir()
	log := logger.NewLogger(false)

	executor := NewExecutor(cfg, outputDir, log)

	assert.NotNil(t, executor)
	assert.Equal(t, cfg, executor.config)
	assert.Equal(t, outputDir, executor.outputDir)
	assert.Equal(t, log, executor.logger)
	assert.NotNil(t, executor.factory)
	assert.Empty(t, executor.results)
}

func TestTestResult_Struct(t *testing.T) {
	result := TestResult{
		AppName:     "Test App",
		AppType:     "web",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(5 * time.Second),
		Duration:   5 * time.Second,
		Success:     true,
		Screenshots: []string{"test.png"},
		Videos:      []string{"test.mp4"},
		Metrics:     map[string]interface{}{"test": "value"},
	}

	assert.Equal(t, "Test App", result.AppName)
	assert.Equal(t, "web", result.AppType)
	assert.Equal(t, true, result.Success)
	assert.Len(t, result.Screenshots, 1)
	assert.Len(t, result.Videos, 1)
	assert.NotNil(t, result.Metrics)
}