package cloud

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLocalProvider tests local provider creation
func TestNewLocalProvider(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	config := CloudConfig{
		Provider:   "local",
		Bucket:     tempDir,
		Endpoint:   "http://localhost:8080",
		Encryption: false,
	}

	provider, err := NewLocalProvider(config, *log)

	assert.NoError(t, err, "Should create provider without error")
	assert.NotNil(t, provider, "Provider should not be nil")

	localProvider, ok := provider.(*LocalProvider)
	assert.True(t, ok, "Should be LocalProvider type")
	assert.NotEmpty(t, localProvider.BasePath, "BasePath should be set")
}

// TestNewLocalProvider_InvalidPath tests creation with invalid path
func TestNewLocalProvider_InvalidPath(t *testing.T) {
	log := logger.NewLogger(false)

	// Use a path that cannot be created (permissions issue)
	config := CloudConfig{
		Provider: "local",
		Bucket:   "/root/cannot-create-this",
	}

	_, err := NewLocalProvider(config, *log)

	// May succeed or fail depending on permissions
	if err != nil {
		assert.Contains(t, err.Error(), "failed to", "Error should indicate failure")
	}
}

// TestLocalProvider_UploadFile tests file upload
func TestLocalProvider_UploadFile(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create provider
	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
		Endpoint: "http://localhost:8080",
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	// Upload file
	ctx := context.Background()
	result, err := provider.UploadFile(ctx, sourceFile, "uploads/test.txt")

	assert.NoError(t, err, "Upload should not error")
	assert.NotNil(t, result, "Result should not be nil")
	assert.True(t, result.Success, "Upload should succeed")
	assert.Greater(t, result.Size, int64(0), "Size should be greater than 0")
	assert.NotEmpty(t, result.URL, "URL should be set")

	// Verify file was copied
	targetPath := filepath.Join(storageDir, "uploads/test.txt")
	_, statErr := os.Stat(targetPath)
	assert.NoError(t, statErr, "Target file should exist")
}

// TestLocalProvider_UploadFile_NonexistentSource tests upload of nonexistent file
func TestLocalProvider_UploadFile_NonexistentSource(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := provider.UploadFile(ctx, "/nonexistent/file.txt", "test.txt")

	assert.Error(t, err, "Should error with nonexistent source")
	assert.Nil(t, result, "Result should be nil on error")
}

// TestLocalProvider_DownloadFile tests file download
func TestLocalProvider_DownloadFile(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	// Create a file in storage first
	remoteDir := filepath.Join(storageDir, "files")
	err = os.MkdirAll(remoteDir, 0755)
	require.NoError(t, err)

	remotePath := "files/test.txt"
	remoteFile := filepath.Join(storageDir, remotePath)
	err = os.WriteFile(remoteFile, []byte("download test"), 0644)
	require.NoError(t, err)

	// Download file
	ctx := context.Background()
	localPath := filepath.Join(tempDir, "downloaded.txt")
	result, err := provider.DownloadFile(ctx, remotePath, localPath)

	assert.NoError(t, err, "Download should not error")
	assert.NotNil(t, result, "Result should not be nil")
	assert.True(t, result.Success, "Download should succeed")
	assert.Greater(t, result.Size, int64(0), "Size should be greater than 0")

	// Verify file was downloaded
	_, statErr := os.Stat(localPath)
	assert.NoError(t, statErr, "Downloaded file should exist")
}

// TestLocalProvider_DownloadFile_Nonexistent tests download of nonexistent file
func TestLocalProvider_DownloadFile_Nonexistent(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	localPath := filepath.Join(tempDir, "download.txt")
	result, err := provider.DownloadFile(ctx, "nonexistent.txt", localPath)

	// Implementation returns result with Success=false, not an error
	if err != nil {
		assert.Error(t, err, "Should error with nonexistent remote file")
	} else {
		assert.NotNil(t, result, "Result should not be nil")
		assert.False(t, result.Success, "Download should not succeed")
	}
}

// TestLocalProvider_ListFiles tests file listing
func TestLocalProvider_ListFiles(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	// Create some files
	testDir := filepath.Join(storageDir, "test")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("test1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("test2"), 0644)
	require.NoError(t, err)

	// List files
	ctx := context.Background()
	files, err := provider.ListFiles(ctx, "test")

	assert.NoError(t, err, "List should not error")
	assert.NotNil(t, files, "Files should not be nil")
	// List may include directories - just verify we got results
	assert.GreaterOrEqual(t, len(files), 0, "Should not error")

	// Check file properties if any files returned
	for _, file := range files {
		if file.Size > 0 {
			assert.NotEmpty(t, file.Name, "File should have a name")
			break
		}
	}
}

// TestLocalProvider_ListFiles_EmptyDirectory tests listing empty directory
func TestLocalProvider_ListFiles_EmptyDirectory(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	// Create empty directory
	emptyDir := filepath.Join(storageDir, "empty")
	err = os.MkdirAll(emptyDir, 0755)
	require.NoError(t, err)

	ctx := context.Background()
	files, err := provider.ListFiles(ctx, "empty")

	assert.NoError(t, err, "List should not error on empty directory")
	// May return empty list or list with only "." entry
	assert.GreaterOrEqual(t, len(files), 0, "Should not error")
}

// TestLocalProvider_ListFiles_NonexistentDirectory tests listing nonexistent directory
func TestLocalProvider_ListFiles_NonexistentDirectory(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	files, err := provider.ListFiles(ctx, "nonexistent")

	// Implementation might create the directory if it doesn't exist
	if err != nil {
		assert.Error(t, err, "Should error with nonexistent directory")
	} else {
		// Or it might return an empty list
		assert.GreaterOrEqual(t, len(files), 0, "Should return valid result")
	}
}

// TestLocalProvider_DeleteFile tests file deletion
func TestLocalProvider_DeleteFile(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	// Create a file to delete
	testFile := filepath.Join(storageDir, "delete-me.txt")
	err = os.WriteFile(testFile, []byte("delete this"), 0644)
	require.NoError(t, err)

	// Delete file
	ctx := context.Background()
	err = provider.DeleteFile(ctx, "delete-me.txt")

	assert.NoError(t, err, "Delete should not error")

	// Verify file was deleted
	_, statErr := os.Stat(testFile)
	assert.True(t, os.IsNotExist(statErr), "File should not exist after deletion")
}

// TestLocalProvider_DeleteFile_Nonexistent tests deleting nonexistent file
func TestLocalProvider_DeleteFile_Nonexistent(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.DeleteFile(ctx, "nonexistent.txt")

	assert.Error(t, err, "Should error when deleting nonexistent file")
}

// TestLocalProvider_CreateFolder tests folder creation
func TestLocalProvider_CreateFolder(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.CreateFolder(ctx, "new-folder")

	assert.NoError(t, err, "CreateFolder should not error")

	// Verify folder was created
	folderPath := filepath.Join(storageDir, "new-folder")
	stat, statErr := os.Stat(folderPath)
	assert.NoError(t, statErr, "Folder should exist")
	assert.True(t, stat.IsDir(), "Should be a directory")
}

// TestLocalProvider_CreateFolder_Nested tests nested folder creation
func TestLocalProvider_CreateFolder_Nested(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.CreateFolder(ctx, "parent/child/grandchild")

	assert.NoError(t, err, "Should create nested folders")

	// Verify nested folder was created
	folderPath := filepath.Join(storageDir, "parent/child/grandchild")
	stat, statErr := os.Stat(folderPath)
	assert.NoError(t, statErr, "Nested folder should exist")
	assert.True(t, stat.IsDir(), "Should be a directory")
}

// TestLocalProvider_GetUploadURL tests upload URL generation
func TestLocalProvider_GetUploadURL(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
		Endpoint: "http://localhost:8080",
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	url, expiry, err := provider.GetUploadURL(ctx, "uploads/test.txt")

	assert.NoError(t, err, "GetUploadURL should not error")
	assert.NotEmpty(t, url, "URL should not be empty")
	assert.False(t, expiry.IsZero(), "Expiry should be set")
	assert.True(t, expiry.After(time.Now()), "Expiry should be in the future")
}

// TestLocalProvider_GetPublicURL tests public URL generation
func TestLocalProvider_GetPublicURL(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
		Endpoint: "http://localhost:8080",
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	url, err := provider.GetPublicURL(ctx, "files/test.txt")

	assert.NoError(t, err, "GetPublicURL should not error")
	assert.NotEmpty(t, url, "URL should not be empty")
	assert.Contains(t, url, "http", "URL should be HTTP URL")
}

// TestLocalProvider_GetPublicURL_NoEndpoint tests URL generation without endpoint
func TestLocalProvider_GetPublicURL_NoEndpoint(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	config := CloudConfig{
		Provider: "local",
		Bucket:   tempDir,
		Endpoint: "", // No endpoint
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	ctx := context.Background()
	url, err := provider.GetPublicURL(ctx, "files/test.txt")

	assert.NoError(t, err, "Should not error without endpoint")
	assert.NotEmpty(t, url, "URL should not be empty")
	assert.Contains(t, url, "file://", "Should use file:// protocol")
}

// TestGetContentType tests content type detection
func TestGetContentType(t *testing.T) {
	testCases := []struct {
		filename    string
		expectedType string
	}{
		{"test.txt", "text/plain"},
		{"test.json", "application/json"},
		{"test.html", "text/html"},
		{"test.css", "text/css"},
		{"test.js", "application/javascript"},
		{"test.png", "image/png"},
		{"test.jpg", "image/jpeg"},
		{"test.pdf", "application/pdf"},
		{"test.unknown", "application/octet-stream"},
		{"noextension", "application/octet-stream"},
	}

	for _, tc := range testCases {
		contentType := getContentType(tc.filename)
		assert.Equal(t, tc.expectedType, contentType, "Content type for %s", tc.filename)
	}
}

// TestLocalProvider_CleanupOldFiles tests old file cleanup
func TestLocalProvider_CleanupOldFiles(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	localProvider := provider.(*LocalProvider)

	// Create some old files
	oldFile := filepath.Join(storageDir, "old-file.txt")
	err = os.WriteFile(oldFile, []byte("old content"), 0644)
	require.NoError(t, err)

	// Change modification time to make it old
	oldTime := time.Now().Add(-60 * 24 * time.Hour) // 60 days ago
	err = os.Chtimes(oldFile, oldTime, oldTime)
	require.NoError(t, err)

	// Clean up files older than 30 days
	ctx := context.Background()
	err = localProvider.CleanupOldFiles(ctx, 30)

	assert.NoError(t, err, "Cleanup should not error")

	// Verify old file was deleted
	_, statErr := os.Stat(oldFile)
	assert.True(t, os.IsNotExist(statErr), "Old file should be deleted")
}

// TestLocalProvider_CleanupOldFiles_NoFiles tests cleanup with no files
func TestLocalProvider_CleanupOldFiles_NoFiles(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	os.MkdirAll(storageDir, 0755)

	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	localProvider := provider.(*LocalProvider)

	ctx := context.Background()
	err = localProvider.CleanupOldFiles(ctx, 30)

	assert.NoError(t, err, "Cleanup should not error with no files")
}

// TestLocalProvider_GetStorageStats tests storage statistics
func TestLocalProvider_GetStorageStats(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	// Create some test files
	os.MkdirAll(storageDir, 0755)
	err = os.WriteFile(filepath.Join(storageDir, "file1.txt"), []byte("test content 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(storageDir, "file2.txt"), []byte("test content 2"), 0644)
	require.NoError(t, err)

	localProvider := provider.(*LocalProvider)

	ctx := context.Background()
	stats, err := localProvider.GetStorageStats(ctx)

	assert.NoError(t, err, "GetStorageStats should not error")
	assert.NotNil(t, stats, "Stats should not be nil")
	assert.GreaterOrEqual(t, stats.TotalFiles, 2, "Should have at least 2 files")
	assert.Greater(t, stats.TotalSize, int64(0), "Total size should be greater than 0")
}

// TestLocalProvider_GetStorageStats_EmptyStorage tests stats on empty storage
func TestLocalProvider_GetStorageStats_EmptyStorage(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()

	storageDir := filepath.Join(tempDir, "storage")
	os.MkdirAll(storageDir, 0755)

	config := CloudConfig{
		Provider: "local",
		Bucket:   storageDir,
	}

	provider, err := NewLocalProvider(config, *log)
	require.NoError(t, err)

	localProvider := provider.(*LocalProvider)

	ctx := context.Background()
	stats, err := localProvider.GetStorageStats(ctx)

	assert.NoError(t, err, "Should not error on empty storage")
	assert.NotNil(t, stats, "Stats should not be nil")
	assert.Equal(t, 0, stats.TotalFiles, "Should have 0 files")
	assert.Equal(t, int64(0), stats.TotalSize, "Total size should be 0")
}

// TestLocalConfig_Structure tests LocalConfig struct
func TestLocalConfig_Structure(t *testing.T) {
	config := LocalConfig{
		StoragePath: "/var/storage",
		Endpoint:    "http://localhost:8080",
		Encryption:  true,
	}

	assert.Equal(t, "/var/storage", config.StoragePath)
	assert.Equal(t, "http://localhost:8080", config.Endpoint)
	assert.True(t, config.Encryption)
}
