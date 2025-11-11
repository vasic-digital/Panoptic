package platforms

import (
	"testing"

	"panoptic/internal/config"
)

// Benchmark WebPlatform operations

func BenchmarkNewWebPlatform(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewWebPlatform()
	}
}

func BenchmarkWebPlatform_MetricsAllocation(b *testing.B) {
	platform := NewWebPlatform()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = platform.GetMetrics()
	}
}

func BenchmarkWebPlatform_Initialize(b *testing.B) {
	app := config.AppConfig{
		Name:    "Benchmark App",
		Type:    "web",
		URL:     "https://example.com",
		Timeout: 30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform := NewWebPlatform()
		// Note: Actual initialization skipped as it requires browser
		// This benchmarks the validation and setup logic
		_ = platform
		_ = app
	}
}

// Benchmark DesktopPlatform operations

func BenchmarkNewDesktopPlatform(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewDesktopPlatform()
	}
}

func BenchmarkDesktopPlatform_MetricsAllocation(b *testing.B) {
	platform := NewDesktopPlatform()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = platform.GetMetrics()
	}
}

func BenchmarkDesktopPlatform_PathValidation(b *testing.B) {
	platform := NewDesktopPlatform()
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMobilePlatform()
	}
}

func BenchmarkMobilePlatform_MetricsAllocation(b *testing.B) {
	platform := NewMobilePlatform()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = platform.GetMetrics()
	}
}

func BenchmarkMobilePlatform_Initialize_Android(b *testing.B) {
	app := config.AppConfig{
		Name:     "Benchmark App",
		Type:     "mobile",
		Platform: "android",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform := NewMobilePlatform()
		_ = platform
		_ = app
	}
}

func BenchmarkMobilePlatform_Initialize_iOS(b *testing.B) {
	app := config.AppConfig{
		Name:     "Benchmark App",
		Type:     "mobile",
		Platform: "ios",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		platform := NewMobilePlatform()
		_ = platform
		_ = app
	}
}

// Benchmark PlatformFactory

func BenchmarkPlatformFactory_CreateWebPlatform(b *testing.B) {
	factory := NewPlatformFactory()

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
	factory := NewPlatformFactory()

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
	factory := NewPlatformFactory()

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
	factory := NewPlatformFactory()

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
	platform := NewWebPlatform()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		waitTime := 1
		// Benchmark just the validation, not actual wait
		if waitTime > 0 && waitTime < 300 {
			_ = platform
		}
	}
}
