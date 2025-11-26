package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"
)

// Benchmark report generation with different optimization strategies
func BenchmarkReportGeneration_Original(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	// Add test results
	executor.results = createLargeTestResults(100)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reportPath := filepath.Join(outputDir, "original_report.html")
		
		// Test original fmt.Sprintf implementation
		report := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>Panoptic Test Report</title>
</head>
<body>
	<h1>Test Report</h1>
	<p>Generated: %s</p>
	<p>Total Tests: %d</p>
	<p>Status: Report generation not fully implemented</p>
</body>
</html>`, time.Now().Format(time.RFC3339), len(executor.results))
		
		err := os.WriteFile(reportPath, []byte(report), 0600)
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(reportPath)
	}
}

func BenchmarkReportGeneration_Fast(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	// Add test results
	executor.results = createLargeTestResults(100)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reportPath := filepath.Join(outputDir, "fast_report.html")
		err := FastGenerateReport(reportPath, executor.results)
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(reportPath)
	}
}

func BenchmarkReportGeneration_Fastest(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	// Add test results
	executor.results = createLargeTestResults(100)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reportPath := filepath.Join(outputDir, "fastest_report.html")
		err := FastestGenerateReport(reportPath, executor.results)
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(reportPath)
	}
}

func BenchmarkReportGeneration_Stream(b *testing.B) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	// Add test results
	executor.results = createLargeTestResults(100)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reportPath := filepath.Join(outputDir, "stream_report.html")
		err := StreamGenerateReport(reportPath, executor.results)
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(reportPath)
	}
}

// Test with different dataset sizes
func BenchmarkReportGeneration_Original_Empty(b *testing.B) {
	benchmarkReportGeneration(b, "original", 0)
}

func BenchmarkReportGeneration_Fast_Empty(b *testing.B) {
	benchmarkReportGeneration(b, "fast", 0)
}

func BenchmarkReportGeneration_Fastest_Empty(b *testing.B) {
	benchmarkReportGeneration(b, "fastest", 0)
}

func BenchmarkReportGeneration_Stream_Empty(b *testing.B) {
	benchmarkReportGeneration(b, "stream", 0)
}

func BenchmarkReportGeneration_Original_Small(b *testing.B) {
	benchmarkReportGeneration(b, "original", 10)
}

func BenchmarkReportGeneration_Fast_Small(b *testing.B) {
	benchmarkReportGeneration(b, "fast", 10)
}

func BenchmarkReportGeneration_Fastest_Small(b *testing.B) {
	benchmarkReportGeneration(b, "fastest", 10)
}

func BenchmarkReportGeneration_Stream_Small(b *testing.B) {
	benchmarkReportGeneration(b, "stream", 10)
}

func BenchmarkReportGeneration_Original_Medium(b *testing.B) {
	benchmarkReportGeneration(b, "original", 100)
}

func BenchmarkReportGeneration_Fast_Medium(b *testing.B) {
	benchmarkReportGeneration(b, "fast", 100)
}

func BenchmarkReportGeneration_Fastest_Medium(b *testing.B) {
	benchmarkReportGeneration(b, "fastest", 100)
}

func BenchmarkReportGeneration_Stream_Medium(b *testing.B) {
	benchmarkReportGeneration(b, "stream", 100)
}

func BenchmarkReportGeneration_Original_Large(b *testing.B) {
	benchmarkReportGeneration(b, "original", 1000)
}

func BenchmarkReportGeneration_Fast_Large(b *testing.B) {
	benchmarkReportGeneration(b, "fast", 1000)
}

func BenchmarkReportGeneration_Fastest_Large(b *testing.B) {
	benchmarkReportGeneration(b, "fastest", 1000)
}

func BenchmarkReportGeneration_Stream_Large(b *testing.B) {
	benchmarkReportGeneration(b, "stream", 1000)
}

// Helper functions
func benchmarkReportGeneration(b *testing.B, variant string, resultCount int) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)

	// Add specified number of test results
	executor.results = createLargeTestResults(resultCount)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reportPath := filepath.Join(outputDir, "report.html")
		
		var err error
		switch variant {
		case "original":
			report := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>Panoptic Test Report</title>
</head>
<body>
	<h1>Test Report</h1>
	<p>Generated: %s</p>
	<p>Total Tests: %d</p>
	<p>Status: Report generation not fully implemented</p>
</body>
</html>`, time.Now().Format(time.RFC3339), len(executor.results))
			err = os.WriteFile(reportPath, []byte(report), 0600)
		case "fast":
			err = FastGenerateReport(reportPath, executor.results)
		case "fastest":
			err = FastestGenerateReport(reportPath, executor.results)
		case "stream":
			err = StreamGenerateReport(reportPath, executor.results)
		}
		
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(reportPath)
	}
}

func createLargeTestResults(count int) []TestResult {
	results := make([]TestResult, count)
	for i := 0; i < count; i++ {
		results[i] = TestResult{
			AppName:     "App" + string(rune('A'+i%26)),
			AppType:     []string{"web", "desktop", "mobile"}[i%3],
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Duration:    time.Duration(i) * time.Millisecond,
			Screenshots: []string{"screenshot1.png", "screenshot2.png"},
			Videos:      []string{"video1.mp4"},
			Metrics: map[string]interface{}{
				"requests": 100 + i,
				"errors":   i % 5,
				"duration": float64(i) * 0.1,
			},
			Success: i%2 == 0,
		}
	}
	return results
}

// Memory usage benchmarks
func BenchmarkReportGeneration_Memory_Original(b *testing.B) {
	benchmarkMemoryUsage(b, "original")
}

func BenchmarkReportGeneration_Memory_Fast(b *testing.B) {
	benchmarkMemoryUsage(b, "fast")
}

func BenchmarkReportGeneration_Memory_Fastest(b *testing.B) {
	benchmarkMemoryUsage(b, "fastest")
}

func BenchmarkReportGeneration_Memory_Stream(b *testing.B) {
	benchmarkMemoryUsage(b, "stream")
}

func benchmarkMemoryUsage(b *testing.B, variant string) {
	log := logger.NewLogger(false)
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	outputDir := b.TempDir()
	executor := NewExecutor(cfg, outputDir, log)
	executor.results = createLargeTestResults(100)

	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reportPath := filepath.Join(outputDir, "report.html")
		
		switch variant {
		case "original":
			report := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>Panoptic Test Report</title>
</head>
<body>
	<h1>Test Report</h1>
	<p>Generated: %s</p>
	<p>Total Tests: %d</p>
	<p>Status: Report generation not fully implemented</p>
</body>
</html>`, time.Now().Format(time.RFC3339), len(executor.results))
			os.WriteFile(reportPath, []byte(report), 0600)
		case "fast":
			FastGenerateReport(reportPath, executor.results)
		case "fastest":
			FastestGenerateReport(reportPath, executor.results)
		case "stream":
			StreamGenerateReport(reportPath, executor.results)
		}
		
		os.Remove(reportPath)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
}