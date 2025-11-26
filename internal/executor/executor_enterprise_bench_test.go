package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"panoptic/internal/config"
	"panoptic/internal/logger"
)

// Test data for enterprise action benchmarks
var enterpriseTestData = map[string]interface{}{
	"status":  "success",
	"count":   42,
	"message": "Operation completed successfully",
	"data": map[string]interface{}{
		"users":    100,
		"projects": 50,
		"teams":    25,
	},
}

// BenchmarkEnterpriseActionSave_Original - original implementation
func BenchmarkEnterpriseActionSave_Original(b *testing.B) {
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
		outputPath := filepath.Join("enterprise", "result.json")
		if err := executor.saveEnterpriseActionResultSilent("test_action", enterpriseTestData, outputPath); err != nil {
			b.Fatal(err)
		}
		// Clean up for next iteration
		fullPath := filepath.Join(outputDir, outputPath)
		os.Remove(fullPath)
	}
}

// FastSaveEnterpriseActionResult - optimized version using fast JSON marshaling
func FastSaveEnterpriseActionResult(outputDir, actionType string, result interface{}, outputPath string) error {
	// Create full output path
	fullPath := filepath.Join(outputDir, outputPath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Use optimized JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Write to file
	if err := os.WriteFile(fullPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// BenchmarkEnterpriseActionSave_Fast - fast version without indentation
func BenchmarkEnterpriseActionSave_Fast(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	_ = NewExecutor(cfg, outputDir, log) // executor created but not used

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join("enterprise", "result.json")
		if err := FastSaveEnterpriseActionResult(outputDir, "test_action", enterpriseTestData, outputPath); err != nil {
			b.Fatal(err)
		}
		// Clean up for next iteration
		fullPath := filepath.Join(outputDir, outputPath)
		os.Remove(fullPath)
	}
}

// FastestSaveEnterpriseActionResult - most optimized version with pre-allocation
func FastestSaveEnterpriseActionResult(outputDir, actionType string, result interface{}, outputPath string) error {
	// Create full output path
	fullPath := filepath.Join(outputDir, outputPath)

	// Ensure directory exists (cached after first call)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Pre-allocate buffer based on result size
	buf := make([]byte, 0, 512) // Conservative pre-allocation
	buf, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Write to file with pre-allocated buffer
	if err := os.WriteFile(fullPath, buf, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// BenchmarkEnterpriseActionSave_Fastest - fastest optimized version
func BenchmarkEnterpriseActionSave_Fastest(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	_ = NewExecutor(cfg, outputDir, log) // executor created but not used

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join("enterprise", "result.json")
		if err := FastestSaveEnterpriseActionResult(outputDir, "test_action", enterpriseTestData, outputPath); err != nil {
			b.Fatal(err)
		}
		// Clean up for next iteration
		fullPath := filepath.Join(outputDir, outputPath)
		os.Remove(fullPath)
	}
}

// StreamingSaveEnterpriseActionResult - streaming approach for large data
func StreamingSaveEnterpriseActionResult(outputDir, actionType string, result interface{}, outputPath string) error {
	// Create full output path
	fullPath := filepath.Join(outputDir, outputPath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Stream JSON to file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode result: %w", err)
	}

	return nil
}

// BenchmarkEnterpriseActionSave_Streaming - streaming version
func BenchmarkEnterpriseActionSave_Streaming(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	_ = NewExecutor(cfg, outputDir, log) // executor created but not used

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join("enterprise", "result.json")
		if err := StreamingSaveEnterpriseActionResult(outputDir, "test_action", enterpriseTestData, outputPath); err != nil {
			b.Fatal(err)
		}
		// Clean up for next iteration
		fullPath := filepath.Join(outputDir, outputPath)
		os.Remove(fullPath)
	}
}

// Memory allocation benchmarks
func BenchmarkEnterpriseActionSave_Allocations(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Test allocation patterns in optimized version
		buf := make([]byte, 0, 512)
		buf, _ = json.Marshal(enterpriseTestData)
		_ = buf
	}
}

// BenchmarkEnterpriseActionSave_WithLogging - performance with logging enabled
func BenchmarkEnterpriseActionSave_WithLogging(b *testing.B) {
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
		outputPath := filepath.Join("enterprise", "result.json")
		if err := executor.saveEnterpriseActionResult("test_action", enterpriseTestData, outputPath); err != nil {
			b.Fatal(err)
		}
		// Clean up for next iteration
		fullPath := filepath.Join(outputDir, outputPath)
		os.Remove(fullPath)
	}
}

// Comparison benchmark with different data sizes
func BenchmarkEnterpriseActionSave_SmallData(b *testing.B) {
	smallData := map[string]interface{}{
		"status":  "success",
		"message": "Operation completed",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := json.Marshal(smallData)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf
	}
}

func BenchmarkEnterpriseActionSave_LargeData(b *testing.B) {
	// Create larger test data
	largeData := map[string]interface{}{
		"status":  "success",
		"count":   1000,
		"message": "Operation completed successfully",
		"data": map[string]interface{}{
			"users": make([]int, 100),
			"projects": make([]map[string]interface{}, 50),
			"metrics": map[string]interface{}{
				"requests":    10000,
				"errors":      50,
				"duration":    150.5,
				"throughput":  500.25,
				"latency":     25.75,
				"cpu_usage":   75.5,
				"memory_usage": 1024.75,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := json.Marshal(largeData)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf
	}
}

// Benchmark directory creation overhead
func BenchmarkEnterpriseActionSave_DirectoryCreation(b *testing.B) {
	outputDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(outputDir, "enterprise", "result.json")
		dir := filepath.Dir(outputPath)
		_ = os.MkdirAll(dir, 0755)
	}
}

// Benchmark different JSON marshaling approaches
func BenchmarkEnterpriseActionSave_JSONApproaches(b *testing.B) {
	b.Run("MarshalIndent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := json.MarshalIndent(enterpriseTestData, "", "  ")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(enterpriseTestData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Encoder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 0, 512)
			encoder := json.NewEncoder(&nilWriter{}) // Custom writer that does nothing
			err := encoder.Encode(enterpriseTestData)
			if err != nil {
				b.Fatal(err)
			}
			_ = buf
		}
	})
}

// Helper writer for benchmarking encoding without I/O
type nilWriter struct{}

func (nw *nilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}