package cloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"panoptic/internal/logger"
)

// TestUploadMethods tests each upload provider method
func TestUploadMethods(t *testing.T) {
	log := logger.NewLogger(false)
	
	tests := []struct {
		name     string
		provider  string
	}{
		{"AWS Upload", "aws"},
		{"GCP Upload", "gcp"},
		{"Azure Upload", "azure"},
		{"Local Upload", "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			manager := &CloudManager{
				Logger: *log,
				Config: CloudConfig{
					Provider: tt.provider,
					Bucket:  filepath.Join(tempDir, "storage"),
				},
				Enabled: true,
			}

			// Create test file
			testFile := filepath.Join(tempDir, "test.txt")
			err := os.WriteFile(testFile, []byte("test content"), 0644)
			require.NoError(t, err)

			err = manager.Upload(testFile)
			assert.NoError(t, err, "Upload should succeed")
			assert.Greater(t, len(manager.TestResults), 0, "Should track upload result")

			// Check first result
			result := manager.TestResults[0]
			assert.True(t, result.Success, "Upload should be marked as successful")
			assert.NotEmpty(t, result.Artifacts, "Should have artifacts")
			
			if len(result.Artifacts) > 0 {
				artifact := result.Artifacts[0]
				assert.Equal(t, "video", artifact.Type, "Should have correct artifact type")
				assert.Equal(t, int64(12), artifact.Size, "Should have correct size")
			}
		})
	}
}

// TestUploadToAWS tests AWS-specific upload logic
func TestUploadToAWS(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "aws",
			Bucket:  "test-bucket",
		},
		Enabled: true,
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("aws test content"), 0644)
	require.NoError(t, err)

	err = manager.uploadToAWS(testFile, "test.txt", getFileInfo(testFile))
	assert.NoError(t, err, "AWS upload should succeed")

	assert.Greater(t, len(manager.TestResults), 0, "Should have test results")
	result := manager.TestResults[0]
	assert.True(t, result.Success, "Should be successful")
	assert.Equal(t, "local", result.NodeID, "Should have correct node ID")
}

// TestUploadToGCP tests GCP-specific upload logic
func TestUploadToGCP(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "gcp",
			Bucket:  "test-bucket",
		},
		Enabled: true,
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("gcp test content"), 0644)
	require.NoError(t, err)

	err = manager.uploadToGCP(testFile, "test.txt", getFileInfo(testFile))
	assert.NoError(t, err, "GCP upload should succeed")

	assert.Greater(t, len(manager.TestResults), 0, "Should have test results")
	result := manager.TestResults[0]
	assert.True(t, result.Success, "Should be successful")
}

// TestUploadToAzure tests Azure-specific upload logic
func TestUploadToAzure(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "azure",
			Bucket:  "test-container",
		},
		Enabled: true,
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("azure test content"), 0644)
	require.NoError(t, err)

	err = manager.uploadToAzure(testFile, "test.txt", getFileInfo(testFile))
	assert.NoError(t, err, "Azure upload should succeed")

	assert.Greater(t, len(manager.TestResults), 0, "Should have test results")
	result := manager.TestResults[0]
	assert.True(t, result.Success, "Should be successful")
}

// TestUploadToLocal tests local storage upload logic
func TestUploadToLocal(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	storageDir := filepath.Join(tempDir, "storage")
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "local",
			Bucket:  storageDir,
		},
		Enabled: true,
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("local test content"), 0644)
	require.NoError(t, err)

	err = manager.uploadToLocal(testFile, "2025/11/26/test.txt", getFileInfo(testFile))
	assert.NoError(t, err, "Local upload should succeed")

	// Verify file was actually copied
	destPath := filepath.Join(storageDir, "2025/11/26/test.txt")
	_, err = os.Stat(destPath)
	assert.NoError(t, err, "File should exist in local storage")

	content, err := os.ReadFile(destPath)
	assert.NoError(t, err, "Should be able to read copied file")
	assert.Equal(t, "local test content", string(content), "Content should match")
}

// TestUploadErrorHandling tests error scenarios
func TestUploadErrorHandling(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "local",
			Bucket:  filepath.Join(tempDir, "storage"),
		},
		Enabled: true,
	}

	// Test with nonexistent file
	err := manager.Upload("/nonexistent/file.txt")
	assert.Error(t, err, "Should return error for nonexistent file")
	assert.Contains(t, err.Error(), "file not found", "Error should mention file not found")

	// Test with empty path
	err = manager.Upload("")
	assert.Error(t, err, "Should return error for empty path")
	assert.Contains(t, err.Error(), "cannot be empty", "Error should mention cannot be empty")
}

// Helper function to get file info
func getFileInfo(path string) os.FileInfo {
	info, _ := os.Stat(path)
	return info
}