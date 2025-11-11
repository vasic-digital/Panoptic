package cloud

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/logger"
)

// Benchmark CloudManager operations

func BenchmarkNewCloudManager(b *testing.B) {
	log := logger.NewLogger(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCloudManager(log)
	}
}

func BenchmarkCloudManager_Initialize_Local(b *testing.B) {
	log := logger.NewLogger(false)
	manager := NewCloudManager(log)

	config := CloudConfig{
		Provider:   "local",
		BucketName: "test-bucket",
		Region:     "local",
		Credentials: map[string]string{
			"path": b.TempDir(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := manager.Initialize(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCloudManager_ListFiles(b *testing.B) {
	log := logger.NewLogger(false)
	manager := NewCloudManager(log)

	tmpDir := b.TempDir()
	config := CloudConfig{
		Provider:   "local",
		BucketName: "test-bucket",
		Region:     "local",
		Credentials: map[string]string{
			"path": tmpDir,
		},
	}

	if err := manager.Initialize(config); err != nil {
		b.Fatal(err)
	}

	// Create some test files
	storagePath := filepath.Join(tmpDir, "test-bucket")
	os.MkdirAll(storagePath, 0755)
	for i := 0; i < 10; i++ {
		filename := filepath.Join(storagePath, "test"+string(rune(i))+".txt")
		os.WriteFile(filename, []byte("test data"), 0644)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := manager.ListFiles(ctx, "")
		if err != nil {
			b.Fatal(err)
		}
		_ = files
	}
}

// Benchmark LocalProvider operations

func BenchmarkNewLocalProvider(b *testing.B) {
	log := logger.NewLogger(false)
	storagePath := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewLocalProvider(storagePath, log)
	}
}

func BenchmarkLocalProvider_Upload_Small(b *testing.B) {
	log := logger.NewLogger(false)
	tmpDir := b.TempDir()
	provider := NewLocalProvider(tmpDir, log)

	// Create a small test file
	testFile := filepath.Join(tmpDir, "source.txt")
	data := []byte("This is a small test file")
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destPath := "uploads/file" + string(rune(i)) + ".txt"
		if err := provider.Upload(ctx, testFile, destPath); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLocalProvider_Upload_Large(b *testing.B) {
	log := logger.NewLogger(false)
	tmpDir := b.TempDir()
	provider := NewLocalProvider(tmpDir, log)

	// Create a larger test file (1MB)
	testFile := filepath.Join(tmpDir, "source_large.txt")
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destPath := "uploads/large_file" + string(rune(i)) + ".txt"
		if err := provider.Upload(ctx, testFile, destPath); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLocalProvider_Download(b *testing.B) {
	log := logger.NewLogger(false)
	tmpDir := b.TempDir()
	provider := NewLocalProvider(tmpDir, log)

	// Create and upload a test file
	testData := []byte("Test file content for download benchmark")
	sourcePath := "test/download.txt"
	sourceFile := filepath.Join(tmpDir, sourcePath)
	os.MkdirAll(filepath.Dir(sourceFile), 0755)
	if err := os.WriteFile(sourceFile, testData, 0644); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destPath := filepath.Join(tmpDir, "downloads", "file"+string(rune(i))+".txt")
		if err := provider.Download(ctx, sourcePath, destPath); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLocalProvider_Delete(b *testing.B) {
	log := logger.NewLogger(false)
	tmpDir := b.TempDir()
	provider := NewLocalProvider(tmpDir, log)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create file
		filePath := "test/delete" + string(rune(i)) + ".txt"
		fullPath := filepath.Join(tmpDir, filePath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte("test"), 0644)

		// Delete it
		if err := provider.Delete(ctx, filePath); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark CloudAnalytics operations

func BenchmarkNewCloudAnalytics(b *testing.B) {
	log := logger.NewLogger(false)
	manager := NewCloudManager(log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCloudAnalytics(log, manager)
	}
}

func BenchmarkCloudAnalytics_GenerateReport_Empty(b *testing.B) {
	log := logger.NewLogger(false)
	manager := NewCloudManager(log)
	analytics := NewCloudAnalytics(log, manager)

	var results []interface{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analytics.GenerateAnalytics(results)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCloudAnalytics_GenerateReport_WithData(b *testing.B) {
	log := logger.NewLogger(false)
	manager := NewCloudManager(log)
	analytics := NewCloudAnalytics(log, manager)

	// Create sample test results
	results := make([]interface{}, 50)
	for i := 0; i < 50; i++ {
		results[i] = map[string]interface{}{
			"node_id":  "node-" + string(rune(i)),
			"success":  i%3 != 0,
			"duration": time.Duration(i*100) * time.Millisecond,
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analytics.GenerateAnalytics(results)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark distributed testing structures

func BenchmarkCloudTestResult_Creation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := CloudTestResult{
			NodeID:    "node-001",
			TestID:    "test-001",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Second),
			Duration:  time.Second,
			Success:   true,
			Output:    "Test completed successfully",
			Metrics: map[string]interface{}{
				"cpu":    0.75,
				"memory": 1024,
			},
		}
		_ = result
	}
}

func BenchmarkCloudTestResult_AllocationPattern(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := make([]CloudTestResult, 0, 100)
		for j := 0; j < 100; j++ {
			result := CloudTestResult{
				NodeID:    "node-" + string(rune(j)),
				TestID:    "test-" + string(rune(j)),
				StartTime: time.Now(),
				EndTime:   time.Now(),
				Duration:  time.Millisecond * time.Duration(j),
				Success:   j%2 == 0,
			}
			results = append(results, result)
		}
		_ = results
	}
}

// Benchmark file synchronization patterns

func BenchmarkSyncOperation_FileWalk(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a directory structure with files
	for i := 0; i < 10; i++ {
		dirPath := filepath.Join(tmpDir, "dir"+string(rune(i)))
		os.MkdirAll(dirPath, 0755)
		for j := 0; j < 5; j++ {
			filePath := filepath.Join(dirPath, "file"+string(rune(j))+".txt")
			os.WriteFile(filePath, []byte("test data"), 0644)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var fileCount int
		filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				fileCount++
			}
			return nil
		})
		_ = fileCount
	}
}

// Benchmark cleanup operations

func BenchmarkCleanupOldFiles_Simulation(b *testing.B) {
	now := time.Now()
	retentionDays := 30

	// Simulate file modification times
	files := make([]struct {
		path    string
		modTime time.Time
	}, 100)

	for i := 0; i < 100; i++ {
		daysOld := i % 60 // Mix of files 0-59 days old
		files[i] = struct {
			path    string
			modTime time.Time
		}{
			path:    "file" + string(rune(i)) + ".txt",
			modTime: now.Add(-time.Duration(daysOld) * 24 * time.Hour),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var toDelete []string
		cutoffTime := now.Add(-time.Duration(retentionDays) * 24 * time.Hour)

		for _, file := range files {
			if file.modTime.Before(cutoffTime) {
				toDelete = append(toDelete, file.path)
			}
		}
		_ = toDelete
	}
}
