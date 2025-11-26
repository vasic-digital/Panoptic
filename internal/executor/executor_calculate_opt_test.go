package executor

import (
	"testing"
	"time"

	"panoptic/internal/cloud"
)

// Benchmark optimized calculateSuccessRate functions

func BenchmarkCalculateSuccessRate_Original(b *testing.B) {
	results := make([]cloud.CloudTestResult, 1000)
	for i := 0; i < 1000; i++ {
		results[i] = cloud.CloudTestResult{
			Success:   i%3 != 0, // ~66% success rate
			TestID:    "test-id",
			NodeID:    "node-id",
			NodeName:  "test-node",
			Location:  "local",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Second),
			Duration:  time.Second,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = calculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_Fast(b *testing.B) {
	results := make([]cloud.CloudTestResult, 1000)
	for i := 0; i < 1000; i++ {
		results[i] = cloud.CloudTestResult{
			Success:   i%3 != 0, // ~66% success rate
			TestID:    "test-id",
			NodeID:    "node-id",
			NodeName:  "test-node",
			Location:  "local",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Second),
			Duration:  time.Second,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FastCalculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_SIMD(b *testing.B) {
	results := make([]cloud.CloudTestResult, 1000)
	for i := 0; i < 1000; i++ {
		results[i] = cloud.CloudTestResult{
			Success:   i%3 != 0, // ~66% success rate
			TestID:    "test-id",
			NodeID:    "node-id",
			NodeName:  "test-node",
			Location:  "local",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Second),
			Duration:  time.Second,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = SIMDCalculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_Small_Original(b *testing.B) {
	results := []cloud.CloudTestResult{
		{Success: true},
		{Success: true},
		{Success: false},
		{Success: true},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = calculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_Small_Fast(b *testing.B) {
	results := []cloud.CloudTestResult{
		{Success: true},
		{Success: true},
		{Success: false},
		{Success: true},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FastCalculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_Empty_Original(b *testing.B) {
	results := []cloud.CloudTestResult{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = calculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_Empty_Fast(b *testing.B) {
	results := []cloud.CloudTestResult{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FastCalculateSuccessRate(results)
	}
}

// Test with larger datasets
func BenchmarkCalculateSuccessRate_10000_Original(b *testing.B) {
	results := make([]cloud.CloudTestResult, 10000)
	for i := 0; i < 10000; i++ {
		results[i] = cloud.CloudTestResult{
			Success:   i%7 != 0, // ~86% success rate
			TestID:    "test-id",
			NodeID:    "node-id",
			NodeName:  "test-node",
			Location:  "local",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Second),
			Duration:  time.Second,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = calculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_10000_Fast(b *testing.B) {
	results := make([]cloud.CloudTestResult, 10000)
	for i := 0; i < 10000; i++ {
		results[i] = cloud.CloudTestResult{
			Success:   i%7 != 0, // ~86% success rate
			TestID:    "test-id",
			NodeID:    "node-id",
			NodeName:  "test-node",
			Location:  "local",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Second),
			Duration:  time.Second,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FastCalculateSuccessRate(results)
	}
}

func BenchmarkCalculateSuccessRate_10000_SIMD(b *testing.B) {
	results := make([]cloud.CloudTestResult, 10000)
	for i := 0; i < 10000; i++ {
		results[i] = cloud.CloudTestResult{
			Success:   i%7 != 0, // ~86% success rate
			TestID:    "test-id",
			NodeID:    "node-id",
			NodeName:  "test-node",
			Location:  "local",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Second),
			Duration:  time.Second,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = SIMDCalculateSuccessRate(results)
	}
}