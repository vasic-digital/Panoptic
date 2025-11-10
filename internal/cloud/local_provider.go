package cloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"panoptic/internal/logger"
)

// LocalProvider implements CloudProvider for local file system
type LocalProvider struct {
	Config LocalConfig
	BasePath string
	Logger  logger.Logger
}

// LocalConfig contains local storage configuration
type LocalConfig struct {
	StoragePath string `yaml:"storage_path"`
	Endpoint   string `yaml:"endpoint"`
	Encryption bool   `yaml:"encryption"`
}

// NewLocalProvider creates a new local provider
func NewLocalProvider(config CloudConfig, log logger.Logger) (CloudProvider, error) {
	localConfig := LocalConfig{
		StoragePath: config.Bucket,    // Use bucket as storage path
		Endpoint:   config.Endpoint,   // Use endpoint as base URL
		Encryption: config.Encryption, // Note: local encryption would require additional implementation
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(localConfig.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(localConfig.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	provider := &LocalProvider{
		Config:   localConfig,
		BasePath: absPath,
		Logger:    log,
	}

	log.Infof("Local provider initialized with storage path: %s", absPath)
	return provider, nil
}

// UploadFile "uploads" a file to local storage (copy operation)
func (lp *LocalProvider) UploadFile(ctx context.Context, localPath, remotePath string) (*UploadResult, error) {
	startTime := time.Now()
	
	lp.Logger.Debugf("Copying file to local storage: %s -> %s", localPath, remotePath)

	// Create target directory
	targetPath := filepath.Join(lp.BasePath, remotePath)
	targetDir := filepath.Dir(targetPath)
	
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Open source file
	sourceFile, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create target file
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create target file: %w", err)
	}
	defer targetFile.Close()

	// Copy file content
	size, err := targetFile.ReadFrom(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	duration := time.Since(startTime)

	// Generate URL (based on endpoint configuration)
	var url string
	if lp.Config.Endpoint != "" {
		url = fmt.Sprintf("%s/%s", strings.TrimSuffix(lp.Config.Endpoint, "/"), remotePath)
	} else {
		url = fmt.Sprintf("file://%s", targetPath)
	}

	lp.Logger.Infof("Successfully copied to local storage: %s (%s, %d bytes)", 
		remotePath, duration.String(), size)

	return &UploadResult{
		Success:    true,
		URL:        url,
		Size:       size,
		Duration:   duration.String(),
		RemotePath: remotePath,
		ETag:       fmt.Sprintf("local_%d", time.Now().UnixNano()),
	}, nil
}

// DownloadFile "downloads" a file from local storage (copy operation)
func (lp *LocalProvider) DownloadFile(ctx context.Context, remotePath, localPath string) (*DownloadResult, error) {
	startTime := time.Now()
	
	lp.Logger.Debugf("Copying file from local storage: %s -> %s", remotePath, localPath)

	// Create target directory
	targetDir := filepath.Dir(localPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Source path in local storage
	sourcePath := filepath.Join(lp.BasePath, remotePath)

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return &DownloadResult{
			Success:    false,
			RemotePath: remotePath,
			LocalPath:  localPath,
			Duration:   time.Since(startTime).String(),
		}, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get source file info
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get source file info: %w", err)
	}

	// Create target file
	targetFile, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create target file: %w", err)
	}
	defer targetFile.Close()

	// Copy file content
	size, err := targetFile.ReadFrom(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	duration := time.Since(startTime)
	etag := fmt.Sprintf("local_%d", sourceInfo.ModTime().UnixNano())

	lp.Logger.Infof("Successfully copied from local storage: %s (%s, %d bytes)", 
		remotePath, duration.String(), size)

	return &DownloadResult{
		Success:    true,
		LocalPath:  localPath,
		Size:       size,
		ETag:       etag,
		Duration:   duration.String(),
		RemotePath: remotePath,
	}, nil
}

// ListFiles lists files in local storage
func (lp *LocalProvider) ListFiles(ctx context.Context, remotePath string) ([]*CloudFile, error) {
	lp.Logger.Debugf("Listing files in local storage: %s", remotePath)

	var files []*CloudFile
	searchPath := filepath.Join(lp.BasePath, remotePath)

	// Walk through directory
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory itself
		if path == lp.BasePath {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(lp.BasePath, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Convert to forward slashes for cloud compatibility
		cloudPath := filepath.ToSlash(relPath)

		// Determine if it's a folder
		isFolder := info.IsDir()

		var size int64
		if !isFolder {
			size = info.Size()
		}

		// Generate URL
		var url string
		if lp.Config.Endpoint != "" {
			url = fmt.Sprintf("%s/%s", strings.TrimSuffix(lp.Config.Endpoint, "/"), cloudPath)
		} else {
			url = fmt.Sprintf("file://%s", path)
		}

		files = append(files, &CloudFile{
			Name:         info.Name(),
			Path:         cloudPath,
			Size:         size,
			LastModified: info.ModTime(),
			ETag:         fmt.Sprintf("local_%d", info.ModTime().UnixNano()),
			ContentType:  getContentType(path),
			IsFolder:     isFolder,
			URL:          url,
		})

		// Don't recurse into subdirectories if we're listing a specific folder
		if remotePath != "" && info.IsDir() && filepath.Dir(path) == searchPath {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to walk local storage directory: %w", err)
	}

	// If path doesn't exist, return empty list
	if os.IsNotExist(err) {
		return []*CloudFile{}, nil
	}

	lp.Logger.Debugf("Listed %d files from local storage: %s", len(files), remotePath)
	return files, nil
}

// DeleteFile deletes a file from local storage
func (lp *LocalProvider) DeleteFile(ctx context.Context, remotePath string) error {
	lp.Logger.Debugf("Deleting file from local storage: %s", remotePath)

	// Full path in local storage
	targetPath := filepath.Join(lp.BasePath, remotePath)

	err := os.Remove(targetPath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	lp.Logger.Infof("Successfully deleted from local storage: %s", remotePath)
	return nil
}

// CreateFolder creates a folder in local storage
func (lp *LocalProvider) CreateFolder(ctx context.Context, remotePath string) error {
	lp.Logger.Debugf("Creating folder in local storage: %s", remotePath)

	// Full path in local storage
	targetPath := filepath.Join(lp.BasePath, remotePath)

	err := os.MkdirAll(targetPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	lp.Logger.Infof("Successfully created folder in local storage: %s", remotePath)
	return nil
}

// GetUploadURL generates a "upload URL" for local storage (placeholder)
func (lp *LocalProvider) GetUploadURL(ctx context.Context, remotePath string) (string, time.Time, error) {
	lp.Logger.Debugf("Generating upload URL for local storage: %s", remotePath)

	// For local storage, return a file:// URL as placeholder
	fullPath := filepath.Join(lp.BasePath, remotePath)
	url := fmt.Sprintf("file://%s", fullPath)
	expiresAt := time.Now().Add(1 * time.Hour)

	lp.Logger.Debugf("Generated upload URL for local storage: %s (expires: %s)", remotePath, expiresAt)
	return url, expiresAt, nil
}

// GetPublicURL gets a public URL for a file in local storage
func (lp *LocalProvider) GetPublicURL(ctx context.Context, remotePath string) (string, error) {
	// Generate URL based on endpoint configuration
	if lp.Config.Endpoint != "" {
		url := fmt.Sprintf("%s/%s", strings.TrimSuffix(lp.Config.Endpoint, "/"), remotePath)
		return url, nil
	}

	// Fallback to file:// URL
	fullPath := filepath.Join(lp.BasePath, remotePath)
	url := fmt.Sprintf("file://%s", fullPath)
	return url, nil
}

// getContentType determines content type based on file extension
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	contentTypes := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain",
		".pdf":  "application/pdf",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}

	return "application/octet-stream"
}

// CleanupOldFiles removes files older than specified days
func (lp *LocalProvider) CleanupOldFiles(ctx context.Context, days int) error {
	lp.Logger.Infof("Cleaning up local storage files older than %d days", days)

	cutoffTime := time.Now().AddDate(0, 0, -days)
	deletedCount := 0
	totalSizeDeleted := int64(0)

	// Walk through storage directory
	err := filepath.Walk(lp.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory itself
		if path == lp.BasePath {
			return nil
		}

		// Only process files, not directories
		if info.IsDir() {
			return nil
		}

		// Check if file is older than cutoff time
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(path); err != nil {
				lp.Logger.Errorf("Failed to delete old file %s: %v", path, err)
				return nil
			}

			deletedCount++
			totalSizeDeleted += info.Size()
			lp.Logger.Debugf("Deleted old file: %s (%d bytes)", path, info.Size())
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk local storage directory for cleanup: %w", err)
	}

	lp.Logger.Infof("Local storage cleanup completed: %d files deleted, %d bytes freed", 
		deletedCount, totalSizeDeleted)
	return nil
}

// GetStorageStats returns storage statistics for local storage
func (lp *LocalProvider) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	lp.Logger.Info("Calculating local storage statistics")

	stats := &StorageStats{
		FileTypeCounts: make(map[string]int),
	}

	var totalSize int64
	var largestSize int64
	var oldestTime time.Time = time.Now()
	var newestTime time.Time = time.Time{}

	// Walk through storage directory
	err := filepath.Walk(lp.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory itself
		if path == lp.BasePath {
			return nil
		}

		// Only process files, not directories
		if info.IsDir() {
			return nil
		}

		stats.TotalFiles++
		totalSize += info.Size()

		// Count file types
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = "no_extension"
		}
		stats.FileTypeCounts[ext]++

		// Track largest file
		if info.Size() > largestSize {
			largestSize = info.Size()
			stats.LargestFile = filepath.Base(path)
		}

		// Track oldest and newest files
		if info.ModTime().Before(oldestTime) {
			oldestTime = info.ModTime()
			stats.OldestFile = filepath.Base(path)
		}
		if info.ModTime().After(newestTime) {
			newestTime = info.ModTime()
			stats.NewestFile = filepath.Base(path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk local storage directory: %w", err)
	}

	stats.TotalSize = totalSize
	stats.TotalSizeGB = float64(totalSize) / (1024 * 1024 * 1024)

	if stats.TotalFiles > 0 {
		stats.AverageFileSize = float64(totalSize) / float64(stats.TotalFiles)
	}

	return stats, nil
}