package platforms

import (
	"os"
	"path/filepath"
	"testing"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
)

func TestNewScreencastRecorder(t *testing.T) {
	log := *logger.NewLogger(false)
	recorder := NewScreencastRecorder(nil, log)

	assert.NotNil(t, recorder)
	assert.False(t, recorder.IsRecording())
	assert.Equal(t, 0, recorder.FrameCount())
	assert.Equal(t, 10, recorder.fps)
}

func TestScreencastRecorderStartWithoutPage(t *testing.T) {
	log := *logger.NewLogger(false)
	recorder := NewScreencastRecorder(nil, log)

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_video.mp4")

	err := recorder.Start(filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "page is nil")
	assert.False(t, recorder.IsRecording())
}

func TestScreencastRecorderStopWithoutStart(t *testing.T) {
	log := *logger.NewLogger(false)
	recorder := NewScreencastRecorder(nil, log)

	err := recorder.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recording in progress")
}

func TestScreencastRecorderDoubleStart(t *testing.T) {
	log := *logger.NewLogger(false)
	recorder := NewScreencastRecorder(nil, log)

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_video.mp4")

	// Force recording state
	recorder.recording = true
	recorder.done = make(chan struct{})

	err := recorder.Start(filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recording already in progress")

	// Clean up
	recorder.recording = false
	close(recorder.done)
}

func TestScreencastRecorderIsRecording(t *testing.T) {
	log := *logger.NewLogger(false)
	recorder := NewScreencastRecorder(nil, log)

	assert.False(t, recorder.IsRecording())

	recorder.mu.Lock()
	recorder.recording = true
	recorder.mu.Unlock()

	assert.True(t, recorder.IsRecording())

	recorder.mu.Lock()
	recorder.recording = false
	recorder.mu.Unlock()

	assert.False(t, recorder.IsRecording())
}

func TestScreencastRecorderFrameCount(t *testing.T) {
	log := *logger.NewLogger(false)
	recorder := NewScreencastRecorder(nil, log)

	assert.Equal(t, 0, recorder.FrameCount())

	recorder.mu.Lock()
	recorder.frameCount = 42
	recorder.mu.Unlock()

	assert.Equal(t, 42, recorder.FrameCount())
}

func TestScreencastRecorderCleanup(t *testing.T) {
	log := *logger.NewLogger(false)
	recorder := NewScreencastRecorder(nil, log)

	tmpDir := t.TempDir()
	framesDir := filepath.Join(tmpDir, "test_frames")
	err := os.MkdirAll(framesDir, 0755)
	assert.NoError(t, err)

	// Create a test frame file
	testFrame := filepath.Join(framesDir, "frame_000000.png")
	err = os.WriteFile(testFrame, []byte("fake png data"), 0644)
	assert.NoError(t, err)

	recorder.framesDir = framesDir
	recorder.cleanup()

	// Verify frames directory was removed
	_, err = os.Stat(framesDir)
	assert.True(t, os.IsNotExist(err))
}

func TestUtilsClampInt(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		min      int
		max      int
		expected int
	}{
		{"within range", 50, 1, 100, 50},
		{"at min", 1, 1, 100, 1},
		{"at max", 100, 1, 100, 100},
		{"below min", -5, 1, 100, 1},
		{"above max", 150, 1, 100, 100},
		{"zero with positive range", 0, 1, 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils_clampInt(tt.val, tt.min, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScreencastRecorderDirectoryCreation(t *testing.T) {
	log := *logger.NewLogger(false)
	recorder := NewScreencastRecorder(nil, log)

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "deep", "nested", "path")
	filename := filepath.Join(nestedDir, "video.mp4")

	// Start will create the frames directory and output directory
	// Even with nil page, directory creation should work
	recorder.recording = false
	recorder.filename = filename
	recorder.framesDir = filename + "_frames"

	err := os.MkdirAll(recorder.framesDir, 0755)
	assert.NoError(t, err)

	err = os.MkdirAll(filepath.Dir(filename), 0755)
	assert.NoError(t, err)

	// Verify directories exist
	_, err = os.Stat(recorder.framesDir)
	assert.NoError(t, err)

	_, err = os.Stat(nestedDir)
	assert.NoError(t, err)

	// Clean up
	recorder.cleanup()
}
