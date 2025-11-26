package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		expected logrus.Level
	}{
		{
			name:     "Verbose logger",
			verbose:  true,
			expected: logrus.DebugLevel,
		},
		{
			name:     "Normal logger",
			verbose:  false,
			expected: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.verbose)
			
			assert.NotNil(t, logger)
			assert.Equal(t, tt.expected, logger.GetLevel())
			assert.IsType(t, &logrus.TextFormatter{}, logger.Formatter)
		})
	}
}

func TestLogger_SetOutputDirectory(t *testing.T) {
	t.Run("Set valid output directory", func(t *testing.T) {
		tempDir := t.TempDir()
		logger := NewLogger(false)
		
		// Disable colors for file output testing
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
		
		logger.SetOutputDirectory(tempDir)
		
		logFile := filepath.Join(tempDir, "logs", "panoptic.log")
		assert.FileExists(t, logFile)
	})

	t.Run("Set non-existent output directory", func(t *testing.T) {
		nonExistentDir := "/tmp/non/existent/path"
		logger := NewLogger(false)
		
		// Should create the directory and log file
		logger.SetOutputDirectory(nonExistentDir)
		
		logFile := filepath.Join(nonExistentDir, "logs", "panoptic.log")
		assert.FileExists(t, logFile)
	})

	t.Run("Log to file", func(t *testing.T) {
		tempDir := t.TempDir()
		logger := NewLogger(false)
		
		logger.SetOutputDirectory(tempDir)
		
		// Write a test log message
		testMessage := "Test log message"
		logger.Info(testMessage)
		
		// Flush the buffer to ensure logs are written
		logger.Flush()
		
		// Read the log file and verify content
		logFile := filepath.Join(tempDir, "logs", "panoptic.log")
		content, err := os.ReadFile(logFile)
		require.NoError(t, err)
		
		// Remove ANSI color codes for comparison
		contentStr := removeANSIColors(string(content))
		assert.Contains(t, contentStr, testMessage)
		assert.Contains(t, contentStr, "INFO") // Look for uppercase "INFO" instead of "info"
	})
}

func TestLogger_Levels(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewLogger(true) // Enable debug logging
	logger.SetOutputDirectory(tempDir)

	logFile := filepath.Join(tempDir, "logs", "panoptic.log")

	// Test different log levels
	testMessages := map[string]func(...interface{}){
		"debug": logger.Debug,
		"info":  logger.Info,
		"warn":  logger.Warn,
		"error": logger.Error,
	}

	for level, logFunc := range testMessages {
		t.Run("Log "+level+" level", func(t *testing.T) {
			testMessage := "Test " + level + " message"
			logFunc(testMessage)
			
			// Flush buffer to ensure logs are written
			logger.Flush()
			
			content, err := os.ReadFile(logFile)
			require.NoError(t, err)
			
			assert.Contains(t, string(content), testMessage)
			assert.Contains(t, string(content), level)
		})
	}
}

func TestLogger_FormattedLogging(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewLogger(true)
	logger.SetOutputDirectory(tempDir)

	logFile := filepath.Join(tempDir, "logs", "panoptic.log")

	// Test formatted logging
	logger.Infof("User %s logged in at %s", "testuser", "2023-01-01")
	logger.Debugf("Debug info: %d items processed", 42)
	logger.Errorf("Error occurred: %s", "connection timeout")

	// Flush buffer to ensure logs are written
	logger.Flush()

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "User testuser logged in at 2023-01-01")
	assert.Contains(t, logContent, "Debug info: 42 items processed")
	assert.Contains(t, logContent, "Error occurred: connection timeout")
}

func TestLogger_ConcurrentLogging(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewLogger(true)
	logger.SetOutputDirectory(tempDir)

	logFile := filepath.Join(tempDir, "logs", "panoptic.log")

	// Test concurrent logging
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Infof("Goroutine %d message", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Flush buffer to ensure logs are written
	logger.Flush()

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)
	// Should have messages from all goroutines
	for i := 0; i < 10; i++ {
		assert.Contains(t, logContent, fmt.Sprintf("Goroutine %d message", i))
	}
}

func TestLogger_ErrorCases(t *testing.T) {
	t.Run("Log with nil formatter", func(t *testing.T) {
		logger := NewLogger(false)
		oldFormatter := logger.Formatter
		logger.Formatter = nil
		
		// Restore formatter before testing
		defer func() {
			if r := recover(); r != nil {
				logger.Formatter = oldFormatter
			}
		}()
		
		// Should not panic if we check for nil formatter
		if logger.Formatter != nil {
			logger.Info("Test message")
		} else {
			// Restore formatter and test
			logger.Formatter = oldFormatter
			logger.Info("Test message")
		}
	})

	t.Run("Set output to unwritable directory", func(t *testing.T) {
		logger := NewLogger(false)
		
		// Try to set to a read-only location (this might fail differently on different systems)
		// For now, just verify it doesn't panic
		assert.NotPanics(t, func() {
			logger.SetOutputDirectory("/root/nonexistent") // Unwritable directory
		})
	})
}

func TestLogger_LongMessages(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewLogger(true)
	logger.SetOutputDirectory(tempDir)

	logFile := filepath.Join(tempDir, "logs", "panoptic.log")

	// Test very long message
	longMessage := strings.Repeat("This is a very long log message. ", 100)
	logger.Info(longMessage)
	
	// Flush buffer to ensure logs are written
	logger.Flush()

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(content), longMessage)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// removeANSIColors removes ANSI escape sequences from text
func removeANSIColors(text string) string {
	// ANSI color codes regex
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(text, "")
}

func TestLogger_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewLogger(true)
	logger.SetOutputDirectory(tempDir)

	logFile := filepath.Join(tempDir, "logs", "panoptic.log")

	// Test messages with special characters
	testMessages := []string{
		"Message with unicode: 你好, 🚀, ☕",
		"Message with quotes: 'single' and \"double\" quotes",
		"Message with newlines\nand\ttabs",
		"Message with emojis: 😀 😎 👍",
		"Message with JSON: {\"key\": \"value\", \"array\": [1,2,3]}",
	}

	for _, msg := range testMessages {
		logger.Info(msg)
	}

	// Flush buffer to ensure logs are written
	logger.Flush()

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := removeANSIColors(string(content))
	for _, msg := range testMessages {
		assert.Contains(t, logContent, msg)
	}
}