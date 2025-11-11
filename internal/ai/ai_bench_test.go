package ai

import (
	"testing"

	"panoptic/internal/logger"
)

// Benchmark AIEnhancedTester operations

func BenchmarkNewAIEnhancedTester(b *testing.B) {
	log := *logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAIEnhancedTester(log)
	}
}

func BenchmarkAIEnhancedTester_Creation(b *testing.B) {
	log := *logger.NewLogger(false)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tester := NewAIEnhancedTester(log)
		_ = tester
	}
}

// Benchmark ErrorDetector operations

func BenchmarkNewErrorDetector(b *testing.B) {
	log := *logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewErrorDetector(log)
	}
}

func BenchmarkErrorDetector_Creation(b *testing.B) {
	log := *logger.NewLogger(false)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector := NewErrorDetector(log)
		_ = detector
	}
}
