package cloud

import (
	"testing"

	"panoptic/internal/logger"
)

// Benchmark CloudManager operations

func BenchmarkNewCloudManager(b *testing.B) {
	log := *logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCloudManager(log)
	}
}

func BenchmarkCloudManager_Creation(b *testing.B) {
	log := *logger.NewLogger(false)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewCloudManager(log)
		_ = manager
	}
}

func BenchmarkCloudConfig_Creation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config := CloudConfig{
			Provider:   "local",
			Bucket:     "test-bucket",
			Region:     "local",
			EnableSync: true,
		}
		_ = config
	}
}

// Benchmark CloudAnalytics operations

func BenchmarkNewCloudAnalytics(b *testing.B) {
	log := *logger.NewLogger(false)
	manager := NewCloudManager(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCloudAnalytics(log, manager)
	}
}
