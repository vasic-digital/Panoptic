package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateComprehensiveReport_EmptyResults(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.html")

	err := GenerateComprehensiveReport(outputPath, []TestResult{})
	assert.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	html := string(data)
	assert.Contains(t, html, "Panoptic Test Report")
	assert.Contains(t, html, `<div class="value">0</div>`)
	assert.Contains(t, html, "Total Apps")
}

func TestGenerateComprehensiveReport_WithResults(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.html")

	// Create a test screenshot file
	screenshotDir := filepath.Join(tmpDir, "screenshots")
	err := os.MkdirAll(screenshotDir, 0755)
	require.NoError(t, err)
	screenshotPath := filepath.Join(screenshotDir, "test.png")
	err = os.WriteFile(screenshotPath, []byte("fake png"), 0644)
	require.NoError(t, err)

	results := []TestResult{
		{
			AppName:     "Admin Console",
			AppType:     "web",
			StartTime:   time.Now().Add(-5 * time.Second),
			EndTime:     time.Now(),
			Duration:    5 * time.Second,
			Success:     true,
			Screenshots: []string{screenshotPath},
			Videos:      []string{},
			Metrics:     map[string]interface{}{"url": "http://localhost:3001"},
		},
		{
			AppName:   "Web App",
			AppType:   "web",
			StartTime: time.Now().Add(-3 * time.Second),
			EndTime:   time.Now(),
			Duration:  3 * time.Second,
			Success:   false,
			Error:     "Login form not found",
			Metrics:   map[string]interface{}{},
		},
	}

	err = GenerateComprehensiveReport(outputPath, results)
	assert.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	html := string(data)
	assert.Contains(t, html, "Admin Console")
	assert.Contains(t, html, "Web App")
	assert.Contains(t, html, "PASSED")
	assert.Contains(t, html, "FAILED")
	assert.Contains(t, html, "Login form not found")
	assert.Contains(t, html, "test.png")
	// Should have both pass and fail stats
	assert.True(t, strings.Contains(html, `class="stat pass"`))
	assert.True(t, strings.Contains(html, `class="stat fail"`))
}

func TestGenerateComprehensiveReport_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "nested", "deep", "report.html")

	err := GenerateComprehensiveReport(outputPath, []TestResult{})
	assert.NoError(t, err)

	_, err = os.Stat(outputPath)
	assert.NoError(t, err)
}

func TestGenerateComprehensiveReport_HTMLEscaping(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.html")

	results := []TestResult{
		{
			AppName:   "App <script>alert('xss')</script>",
			AppType:   "web",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  time.Second,
			Success:   false,
			Error:     "Error with <html> & \"quotes\"",
			Metrics:   map[string]interface{}{},
		},
	}

	err := GenerateComprehensiveReport(outputPath, results)
	assert.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	html := string(data)
	// Should be escaped, not raw HTML
	assert.NotContains(t, html, "<script>alert")
	assert.Contains(t, html, "&lt;script&gt;")
	assert.Contains(t, html, "&amp;")
}

func TestGenerateComprehensiveReport_WithVideoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.html")

	// Create a test video file
	videoDir := filepath.Join(tmpDir, "videos")
	err := os.MkdirAll(videoDir, 0755)
	require.NoError(t, err)
	videoPath := filepath.Join(videoDir, "session.mp4")
	err = os.WriteFile(videoPath, []byte("fake video"), 0644)
	require.NoError(t, err)

	results := []TestResult{
		{
			AppName:   "Test App",
			AppType:   "web",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  time.Second,
			Success:   true,
			Videos:    []string{videoPath},
			Metrics:   map[string]interface{}{},
		},
	}

	err = GenerateComprehensiveReport(outputPath, results)
	assert.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	html := string(data)
	assert.Contains(t, html, "<video controls")
	assert.Contains(t, html, "session.mp4")
	assert.Contains(t, html, "download")
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"milliseconds", 500 * time.Millisecond, "500ms"},
		{"one second", time.Second, "1.0s"},
		{"seconds", 5500 * time.Millisecond, "5.5s"},
		{"one minute", 60 * time.Second, "1m 0s"},
		{"minutes and seconds", 90 * time.Second, "1m 30s"},
		{"multiple minutes", 185 * time.Second, "3m 5s"},
		{"zero", 0, "0ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateComprehensiveReport_LargeResultSet(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.html")

	// Create 20 test results
	var results []TestResult
	for i := 0; i < 20; i++ {
		results = append(results, TestResult{
			AppName:   "App " + string(rune('A'+i%26)),
			AppType:   "web",
			StartTime: time.Now().Add(-time.Duration(i) * time.Second),
			EndTime:   time.Now(),
			Duration:  time.Duration(i+1) * time.Second,
			Success:   i%3 != 0,
			Metrics:   map[string]interface{}{},
		})
	}

	err := GenerateComprehensiveReport(outputPath, results)
	assert.NoError(t, err)

	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.True(t, info.Size() > 0)
}
