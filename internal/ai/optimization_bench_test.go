package ai

import (
	"testing"
	
	"panoptic/internal/logger"
)

// Benchmark original vs optimized implementations

func BenchmarkOriginalAIEnhancedTester(b *testing.B) {
	log := logger.NewLogger(false)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tester := NewAIEnhancedTester(*log)
		_ = tester
	}
}

func BenchmarkOptimizedAIEnhancedTester(b *testing.B) {
	log := logger.NewLogger(false)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tester := NewOptimizedAIEnhancedTester(*log)
		tester.Release() // Clean up for fairness
	}
}

func BenchmarkOriginalErrorDetector(b *testing.B) {
	log := logger.NewLogger(false)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		detector := NewErrorDetector(*log)
		_ = detector
	}
}

func BenchmarkOptimizedErrorDetector(b *testing.B) {
	log := logger.NewLogger(false)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		detector := NewOptimizedErrorDetector(*log)
		_ = detector
	}
}

// Benchmark concurrent access patterns
func BenchmarkOptimizedErrorDetector_Concurrent(b *testing.B) {
	log := logger.NewLogger(false)
	detector := NewOptimizedErrorDetector(*log)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = detector.IsEnabled()
		}
	})
}