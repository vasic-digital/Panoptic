package executor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/cloud"
	"panoptic/internal/config"
	"panoptic/internal/logger"
)

// Benchmark helper functions

func BenchmarkGetStringFromMap(b *testing.B) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getStringFromMap(m, "key1")
	}
}

func BenchmarkGetBoolFromMap(b *testing.B) {
	m := map[string]interface{}{
		"enabled": true,
		"disabled": false,
		"flag": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getBoolFromMap(m, "enabled")
	}
}

func BenchmarkGetIntFromMap(b *testing.B) {
	m := map[string]interface{}{
		"count": 42,
		"size": int64(1024),
		"ratio": float64(0.5),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getIntFromMap(m, "count")
	}
}

// Benchmark executor creation

func BenchmarkNewExecutor(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Benchmark Test",
		Apps: []config.AppConfig{
			{Name: "Test App", Type: "web", URL: "https://example.com"},
		},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewExecutor(cfg, outputDir, log)
	}
}

func BenchmarkNewExecutor_WithCloudConfig(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Benchmark Test",
		Apps: []config.AppConfig{
			{Name: "Test App", Type: "web", URL: "https://example.com"},
		},
		Actions: []config.Action{},
		Settings: config.Settings{
			Cloud: map[string]interface{}{
				"provider": "local",
				"bucket":   "test-bucket",
				"region":   "us-east-1",
			},
		},
	}
	outputDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewExecutor(cfg, outputDir, log)
	}
}

func BenchmarkNewExecutor_WithEnterpriseConfig(b *testing.B) {
	log := logger.NewLogger(false)

	// Create temporary enterprise config
	tmpDir := b.TempDir()
	enterpriseConfigPath := filepath.Join(tmpDir, "enterprise_config.yaml")
	enterpriseConfig := `enabled: true
organization:
  name: "Benchmark Org"
  id: "bench-org"
license:
  type: "enterprise"
  max_users: 100
  expiration_date: "2030-12-31T23:59:59Z"
storage:
  data_path: "` + filepath.Join(tmpDir, "data") + `"
`
	if err := os.WriteFile(enterpriseConfigPath, []byte(enterpriseConfig), 0644); err != nil {
		b.Fatal(err)
	}

	cfg := &config.Config{
		Name: "Benchmark Test",
		Apps: []config.AppConfig{
			{Name: "Test App", Type: "web", URL: "https://example.com"},
		},
		Actions: []config.Action{},
		Settings: config.Settings{
			Enterprise: map[string]interface{}{
				"config_path": enterpriseConfigPath,
			},
		},
	}
	outputDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewExecutor(cfg, outputDir, log)
	}
}

// Benchmark TestResult operations

func BenchmarkTestResult_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TestResult{
			AppName:     "Test App",
			AppType:     "web",
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Duration:    time.Second,
			Screenshots: []string{"screenshot1.png", "screenshot2.png"},
			Videos:      []string{"video1.mp4"},
			Metrics: map[string]interface{}{
				"requests": 100,
				"errors":   2,
				"duration": 1.5,
			},
			Success: true,
		}
	}
}

func BenchmarkTestResult_JSONMarshaling(b *testing.B) {
	result := TestResult{
		AppName:     "Test App",
		AppType:     "web",
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		Duration:    time.Second,
		Screenshots: []string{"screenshot1.png", "screenshot2.png"},
		Videos:      []string{"video1.mp4"},
		Metrics: map[string]interface{}{
			"requests": 100,
			"errors":   2,
			"duration": 1.5,
		},
		Success: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Use the production optimized MarshalJSON method
		data, err := result.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
		// Simulate usage to prevent compiler optimizations
		if len(data) == 0 {
			b.Fatal("empty data")
		}
	}
}

// Benchmark configuration validation

func BenchmarkExecutor_ConfigValidation(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name: "Test",
		Apps: []config.AppConfig{
			{Name: "App1", Type: "web", URL: "https://example.com"},
			{Name: "App2", Type: "desktop", Path: "/usr/bin/app"},
			{Name: "App3", Type: "mobile", Platform: "ios"},
		},
		Actions: []config.Action{
			{Type: "navigate", Value: "https://example.com"},
			{Type: "click", Selector: ".button"},
			{Type: "fill", Selector: "#input", Value: "test"},
			{Type: "wait", WaitTime: 1},
		},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Validate config by checking apps and actions
		for _, app := range cfg.Apps {
			_ = app.Name
			_ = app.Type
		}
		for _, action := range cfg.Actions {
			_ = action.Type
		}
		_ = executor
	}
}

// Benchmark calculateSuccessRate

func BenchmarkCalculateSuccessRate_Empty(b *testing.B) {
	results := []cloud.CloudTestResult{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_Small(b *testing.B) {
	results := []cloud.CloudTestResult{
		{Success: true},
		{Success: true},
		{Success: false},
		{Success: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_Large(b *testing.B) {
	results := make([]cloud.CloudTestResult, 1000)
	for i := 0; i < 1000; i++ {
		results[i] = cloud.CloudTestResult{Success: i%3 != 0} // ~66% success rate
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSuccessRate(results)
	}
}

// Benchmark report generation

func BenchmarkExecutor_GenerateReport_EmptyResults(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reportPath := filepath.Join(outputDir, "report.html")
		if err := executor.GenerateReport(reportPath); err != nil {
			b.Fatal(err)
		}
		// Clean up for next iteration
		os.Remove(reportPath)
	}
}

func BenchmarkExecutor_GenerateReport_WithResults(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	// Add some test results
	executor.results = []TestResult{
		{
			AppName:     "App1",
			AppType:     "web",
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Duration:    time.Second,
			Screenshots: []string{"screenshot1.png"},
			Videos:      []string{"video1.mp4"},
			Metrics:     map[string]interface{}{"requests": 100},
			Success:     true,
		},
		{
			AppName:     "App2",
			AppType:     "desktop",
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Duration:    time.Second * 2,
			Screenshots: []string{"screenshot2.png"},
			Videos:      []string{},
			Metrics:     map[string]interface{}{"clicks": 50},
			Success:     false,
			Error:       "Test error",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reportPath := filepath.Join(outputDir, "report.html")
		if err := executor.GenerateReport(reportPath); err != nil {
			b.Fatal(err)
		}
		// Clean up for next iteration
		os.Remove(reportPath)
	}
}

// Benchmark enterprise operations

func BenchmarkExecutor_SaveEnterpriseActionResult(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	result := map[string]interface{}{
		"status":  "success",
		"count":   42,
		"message": "Operation completed successfully",
		"data": map[string]interface{}{
			"users":    100,
			"projects": 50,
			"teams":    25,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join("enterprise", "result.json")
		if err := executor.saveEnterpriseActionResult("test_action", result, outputPath); err != nil {
			b.Fatal(err)
		}
		// Clean up for next iteration
		fullPath := filepath.Join(outputDir, outputPath)
		os.Remove(fullPath)
	}
}

// Benchmark memory allocation patterns

func BenchmarkExecutor_ResultsAllocation_Small(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := make([]TestResult, 0, 10)
		for j := 0; j < 10; j++ {
			results = append(results, TestResult{
				AppName: "App",
				Success: true,
			})
		}
		_ = results
	}
}

func BenchmarkExecutor_ResultsAllocation_Large(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := make([]TestResult, 0, 100)
		for j := 0; j < 100; j++ {
			results = append(results, TestResult{
				AppName: "App",
				Success: true,
			})
		}
		_ = results
	}
}

func BenchmarkExecutor_MetricsMapCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		metrics := make(map[string]interface{})
		metrics["requests"] = 100
		metrics["errors"] = 5
		metrics["duration"] = 1.5
		metrics["memory"] = 1024
		metrics["cpu"] = 0.75
		_ = metrics
	}
}
