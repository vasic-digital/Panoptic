package executor

import (
	"encoding/json"
	"testing"
	"time"
)

// Create test data for benchmarks
var (
	smallTestResult = TestResult{
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
	
	// Create a large test result with more data
	largeScreenshots = make([]string, 50)
	largeVideos      = make([]string, 10)
	largeMetrics     = make(map[string]interface{})
)

func init() {
	// Initialize large test data
	for i := 0; i < 50; i++ {
		largeScreenshots[i] = "screenshot_long_path_name_" + string(rune(i)) + ".png"
	}
	for i := 0; i < 10; i++ {
		largeVideos[i] = "video_long_path_name_" + string(rune(i)) + ".mp4"
	}
	for i := 0; i < 20; i++ {
		largeMetrics["metric_key_"+string(rune(i))] = "metric_value_with_long_content_" + string(rune(i))
	}
}

var largeTestResult = TestResult{
	AppName:     "Large Test Application with Very Long Name",
	AppType:     "web",
	StartTime:   time.Now(),
	EndTime:     time.Now(),
	Duration:    5 * time.Minute,
	Screenshots: largeScreenshots,
	Videos:      largeVideos,
	Metrics:     largeMetrics,
	Success:     true,
	Error:       "This is a detailed error message with lots of context and information about what went wrong during the test execution",
}

// Benchmark standard library JSON marshaling with indentation
func BenchmarkStandardLibrary_MarshalIndent(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.MarshalIndent(&smallTestResult, "", "  ")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark standard library JSON marshaling without indentation
func BenchmarkStandardLibrary_Marshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(&smallTestResult)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark custom optimized MarshalJSON
func BenchmarkCustom_MarshalJSON(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := smallTestResult.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark large data with standard library
func BenchmarkStandardLibrary_Marshal_Large(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(&largeTestResult)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark large data with custom implementation
func BenchmarkCustom_MarshalJSON_Large(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := largeTestResult.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark memory allocation patterns
func BenchmarkCustom_MarshalJSON_NoAlloc(b *testing.B) {
	// Test with pre-allocated buffer to see allocation reduction
	buf := make([]byte, 0, 2048)
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := smallTestResult.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
		// Copy to pre-allocated buffer to simulate zero-copy usage
		if len(result) > cap(buf) {
			buf = make([]byte, 0, len(result)*2)
		}
		buf = append(buf[:0], result...)
		_ = buf
	}
}

// Test different data type scenarios
func BenchmarkCustom_MarshalJSON_WithComplexMetrics(b *testing.B) {
	complexResult := smallTestResult
	complexResult.Metrics = map[string]interface{}{
		"string_array":    []string{"a", "b", "c", "d", "e"},
		"nested_map":      map[string]string{"key1": "value1", "key2": "value2"},
		"mixed_types":     []interface{}{"string", 42, true, 3.14, time.Now()},
		"large_string":     "This is a very long string with lots of content to test how the marshaling handles large text fields efficiently",
		"unicode_content": "Testing unicode: ñáéíóú 中文 العربية русский 한국어 日本語",
		"special_chars":   "Testing \"quotes\" and \\slashes\\ and \nnewlines\tand\ttabs",
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := complexResult.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}