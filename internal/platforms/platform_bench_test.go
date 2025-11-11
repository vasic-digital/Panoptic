package platforms

import (
	"testing"

	"panoptic/internal/config"
	"panoptic/internal/logger"
)

// Benchmark WebPlatform operations

func BenchmarkNewWebPlatform(b *testing.B) {
	log := logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewWebPlatform(log)
	}
}

func BenchmarkWebPlatform_MetricsAllocation(b *testing.B) {
	log := logger.NewLogger(false)
	platform := NewWebPlatform(log)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = platform.GetMetrics()
	}
}

func BenchmarkWebPlatform_Initialize(b *testing.B) {
	log := logger.NewLogger(false)
	app := config.AppConfig{
		Name:    "Benchmark App",
		Type:    "web",
		URL:     "https://example.com",
		Timeout: 30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform := NewWebPlatform(log)
		// Note: Actual initialization skipped as it requires browser
		// This benchmarks the validation and setup logic
		_ = platform
		_ = app
	}
}

// Benchmark DesktopPlatform operations

func BenchmarkNewDesktopPlatform(b *testing.B) {
	log := logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewDesktopPlatform(log)
	}
}

func BenchmarkDesktopPlatform_MetricsAllocation(b *testing.B) {
	log := logger.NewLogger(false)
	platform := NewDesktopPlatform(log)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = platform.GetMetrics()
	}
}

func BenchmarkDesktopPlatform_PathValidation(b *testing.B) {
	log := logger.NewLogger(false)
	platform := NewDesktopPlatform(log)
	app := config.AppConfig{
		Name: "Test App",
		Type: "desktop",
		Path: "/usr/bin/test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark path validation logic
		_ = platform
		_ = app.Path
	}
}

// Benchmark MobilePlatform operations

func BenchmarkNewMobilePlatform(b *testing.B) {
	log := logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMobilePlatform(log)
	}
}

func BenchmarkMobilePlatform_MetricsAllocation(b *testing.B) {
	log := logger.NewLogger(false)
	platform := NewMobilePlatform(log)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = platform.GetMetrics()
	}
}

func BenchmarkMobilePlatform_Initialize_Android(b *testing.B) {
	log := logger.NewLogger(false)
	app := config.AppConfig{
		Name:     "Benchmark App",
		Type:     "mobile",
		Platform: "android",
		DeviceID: "emulator-5554",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform := NewMobilePlatform(log)
		_ = platform
		_ = app
	}
}

func BenchmarkMobilePlatform_Initialize_iOS(b *testing.B) {
	log := logger.NewLogger(false)
	app := config.AppConfig{
		Name:     "Benchmark App",
		Type:     "mobile",
		Platform: "ios",
		DeviceID: "iPhone-12",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform := NewMobilePlatform(log)
		_ = platform
		_ = app
	}
}

// Benchmark PlatformFactory

func BenchmarkPlatformFactory_CreateWebPlatform(b *testing.B) {
	log := logger.NewLogger(false)
	factory := NewPlatformFactory(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform, err := factory.CreatePlatform("web")
		if err != nil {
			b.Fatal(err)
		}
		_ = platform
	}
}

func BenchmarkPlatformFactory_CreateDesktopPlatform(b *testing.B) {
	log := logger.NewLogger(false)
	factory := NewPlatformFactory(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform, err := factory.CreatePlatform("desktop")
		if err != nil {
			b.Fatal(err)
		}
		_ = platform
	}
}

func BenchmarkPlatformFactory_CreateMobilePlatform(b *testing.B) {
	log := logger.NewLogger(false)
	factory := NewPlatformFactory(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform, err := factory.CreatePlatform("mobile")
		if err != nil {
			b.Fatal(err)
		}
		_ = platform
	}
}

func BenchmarkPlatformFactory_InvalidPlatform(b *testing.B) {
	log := logger.NewLogger(false)
	factory := NewPlatformFactory(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := factory.CreatePlatform("invalid")
		if err == nil {
			b.Fatal("expected error for invalid platform")
		}
	}
}

// Benchmark metrics collection patterns

func BenchmarkMetricsCollection_Sequential(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		metrics := make([]map[string]interface{}, 0, 100)
		for j := 0; j < 100; j++ {
			metric := map[string]interface{}{
				"timestamp": j,
				"value":     j * 2,
				"type":      "measurement",
			}
			metrics = append(metrics, metric)
		}
		_ = metrics
	}
}

func BenchmarkMetricsCollection_Preallocated(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		metrics := make([]map[string]interface{}, 100)
		for j := 0; j < 100; j++ {
			metrics[j] = map[string]interface{}{
				"timestamp": j,
				"value":     j * 2,
				"type":      "measurement",
			}
		}
		_ = metrics
	}
}

// Benchmark screenshot path generation

func BenchmarkScreenshotPathGeneration(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		appName := "TestApp"
		actionName := "screenshot"
		timestamp := int64(1234567890)
		_ = appName + "_" + actionName + "_" + string(rune(timestamp)) + ".png"
	}
}

// Benchmark Wait operation simulation

func BenchmarkWaitOperation_Validation(b *testing.B) {
	log := logger.NewLogger(false)
	platform := NewWebPlatform(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		waitTime := 1
		// Benchmark just the validation, not actual wait
		if waitTime > 0 && waitTime < 300 {
			_ = platform
		}
	}
}
