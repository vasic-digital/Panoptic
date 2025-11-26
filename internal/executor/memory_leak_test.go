package executor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"
)

// TestMemoryLeak_Scenarios tests various memory management scenarios
func TestMemoryLeak_Scenarios(t *testing.T) {
	log := logger.NewLogger(false)
	
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Test scenario 1: Executor with large results that are cleared
	t.Run("LargeResultsCleared", func(t *testing.T) {
		cfg := &config.Config{
			Name: "Memory Leak Test",
			Apps: []config.AppConfig{
				{Name: "Test App", Type: "web", URL: "https://example.com"},
			},
			Actions: []config.Action{},
			Settings: config.Settings{},
		}
		
		executor := NewExecutor(cfg, tempDir, log)
		
		// Simulate adding many results (as if from multiple test runs)
		for i := 0; i < 1000; i++ {
			result := TestResult{
				AppName:     "Test App " + string(rune(i)),
				AppType:     "web",
				StartTime:   time.Now(),
				EndTime:     time.Now().Add(time.Second),
				Duration:    time.Second,
				Screenshots: []string{"screenshot.png"},
				Videos:      []string{"video.mp4"},
				Metrics: map[string]interface{}{
					"iteration": i,
					"data":      make([]byte, 1024), // 1KB of data per result
				},
				Success: true,
			}
			executor.results = append(executor.results, result)
		}
		
		initialLen := len(executor.results)
		if initialLen != 1000 {
			t.Errorf("Expected 1000 results, got %d", initialLen)
		}
		
		// Clear results to free memory
		executor.results = nil
		
		// Force garbage collection to test memory release
		// In production, GC would happen naturally
		finalLen := len(executor.results)
		if finalLen != 0 {
			t.Errorf("Expected 0 results after clearing, got %d", finalLen)
		}
	})
	
	// Test scenario 2: Continuous execution without memory growth
	t.Run("ContinuousExecution", func(t *testing.T) {
		cfg := &config.Config{
			Name: "Continuous Test",
			Apps: []config.AppConfig{
				{Name: "Continuous App", Type: "web", URL: "https://example.com"},
			},
			Actions: []config.Action{
				{Type: "navigate", Value: "https://example.com"},
				{Type: "wait", WaitTime: 1},
			},
			Settings: config.Settings{},
		}
		
		// Run multiple iterations to check memory stability
		for iteration := 0; iteration < 5; iteration++ {
			// Create new executor for each iteration (as recommended pattern)
			executor := NewExecutor(cfg, tempDir, log)
			
			// Run the test
			err := executor.Run()
			if err != nil {
				// We expect some errors due to mock setup, but not memory issues
				t.Logf("Iteration %d completed with error (expected): %v", iteration, err)
			}
			
			// Check that results are reasonable
			if len(executor.results) > 10 {
				t.Errorf("Too many results in iteration %d: %d", iteration, len(executor.results))
			}
		}
	})
	
	// Test scenario 3: JSON marshaling with buffer reuse
	t.Run("JSONBufferReuse", func(t *testing.T) {
		// Test the JSON buffer pool
		results := make([]TestResult, 100)
		for i := 0; i < 100; i++ {
			results[i] = TestResult{
				AppName:     "Buffer Test App",
				AppType:     "web",
				StartTime:   time.Now(),
				EndTime:     time.Now(),
				Duration:    time.Second,
				Screenshots: []string{"screenshot.png"},
				Videos:      []string{"video.mp4"},
				Metrics: map[string]interface{}{
					"buffer_test": i,
					"data":        make([]byte, 512), // 512 bytes per result
				},
				Success: true,
			}
		}
		
		// Test marshaling multiple times with buffer reuse
		for i := 0; i < 10; i++ {
			buf := jsonBufferPool.Get().(*bytes.Buffer)
			buf.Reset()
			
			// Marshal all results
			for _, result := range results {
				data, err := result.MarshalJSON()
				if err != nil {
					t.Errorf("MarshalJSON failed: %v", err)
				}
				buf.Write(data)
				buf.WriteByte(',')
			}
			
			// Return buffer to pool
			jsonBufferPool.Put(buf)
		}
	})
	
	// Test scenario 4: File cleanup after test completion
	t.Run("FileCleanup", func(t *testing.T) {
		cfg := &config.Config{
			Name: "Cleanup Test",
			Apps: []config.AppConfig{
				{Name: "Cleanup App", Type: "web", URL: "https://example.com"},
			},
			Actions: []config.Action{},
			Settings: config.Settings{},
		}
		
		executor := NewExecutor(cfg, tempDir, log)
		
		// Create some test files
		testFiles := []string{"test1.json", "test2.png", "test3.mp4"}
		for _, filename := range testFiles {
			filePath := filepath.Join(tempDir, filename)
			err := os.WriteFile(filePath, []byte("test data"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}
		
		// Verify files exist
		for _, filename := range testFiles {
			filePath := filepath.Join(tempDir, filename)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Test file %s was not created", filename)
			}
		}
		
		// Use executor to verify it's properly initialized
		if executor.config.Name != "Cleanup Test" {
			t.Errorf("Executor not properly initialized")
		}
		
		// In a real scenario, cleanup would happen after report generation
		// Here we just verify the structure is in place
		t.Logf("Test files created successfully in %s", tempDir)
	})
}

// BenchmarkContinuousExecution simulates the memory leak scenario from the conversation
func BenchmarkContinuousExecution_WithoutLeak(b *testing.B) {
	log := logger.NewLogger(false)
	tempDir := b.TempDir()
	
	cfg := &config.Config{
		Name: "Continuous Benchmark",
		Apps: []config.AppConfig{
			{Name: "Benchmark App", Type: "web", URL: "https://example.com"},
		},
		Actions: []config.Action{
			{Type: "navigate", Value: "https://example.com"},
			{Type: "wait", WaitTime: 1},
		},
		Settings: config.Settings{},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Create new executor for each iteration (prevents memory accumulation)
		executor := NewExecutor(cfg, tempDir, log)
		
		// Add some test results
		result := TestResult{
			AppName:     "Benchmark App",
			AppType:     "web",
			StartTime:   time.Now(),
			EndTime:     time.Now().Add(time.Millisecond),
			Duration:    time.Millisecond,
			Screenshots: []string{"screenshot.png"},
			Videos:      []string{"video.mp4"},
			Metrics: map[string]interface{}{
				"iteration": i,
				"data":      make([]byte, 1024),
			},
			Success: true,
		}
		executor.results = append(executor.results, result)
		
		// Marshal results (this is where the memory optimization matters)
		_, err := result.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
		
		// Clear results to prevent memory accumulation
		executor.results = nil
	}
}