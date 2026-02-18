package platforms

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"panoptic/internal/logger"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ScreencastRecorder captures CDP screencast frames and encodes them into video using ffmpeg.
// If ffmpeg is unavailable, it retains the individual frame PNGs as a fallback.
type ScreencastRecorder struct {
	page       *rod.Page
	logger     logger.Logger
	filename   string
	framesDir  string
	frameCount int
	mu         sync.Mutex
	done       chan struct{}
	recording  bool
	startTime  time.Time
	fps        int
}

// NewScreencastRecorder creates a new screencast recorder for the given page.
func NewScreencastRecorder(page *rod.Page, log logger.Logger) *ScreencastRecorder {
	return &ScreencastRecorder{
		page:   page,
		logger: log,
		fps:    10, // 10 FPS is enough for UI test recordings and keeps resource usage low
	}
}

// Start begins capturing screencast frames from the browser via CDP.
func (r *ScreencastRecorder) Start(filename string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.recording {
		return fmt.Errorf("recording already in progress")
	}

	if r.page == nil {
		return fmt.Errorf("page is nil, cannot start recording")
	}

	r.filename = filename
	r.frameCount = 0
	r.recording = true
	r.startTime = time.Now()
	r.done = make(chan struct{})

	// Create temp directory for frames
	r.framesDir = filename + "_frames"
	if err := os.MkdirAll(r.framesDir, 0755); err != nil {
		r.recording = false
		return fmt.Errorf("failed to create frames directory: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		r.recording = false
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Start CDP screencast - captures rendered frames directly from the compositor
	go r.captureFrames()

	r.logger.Infof("Screencast recording started: %s", filename)
	return nil
}

// captureFrames runs the CDP screencast event loop in a goroutine.
func (r *ScreencastRecorder) captureFrames() {
	// Use page events to capture screencast frames
	go r.page.EachEvent(func(e *proto.PageScreencastFrame) {
		r.mu.Lock()
		if !r.recording {
			r.mu.Unlock()
			return
		}
		count := r.frameCount
		r.frameCount++
		r.mu.Unlock()

		// e.Data is already decoded from base64 by go-rod
		// Save frame as PNG
		framePath := filepath.Join(r.framesDir, fmt.Sprintf("frame_%06d.png", count))
		if err := os.WriteFile(framePath, e.Data, 0644); err != nil {
			r.logger.Warnf("Failed to save frame %d: %v", count, err)
			return
		}

		// Acknowledge the frame so CDP sends the next one
		_ = proto.PageScreencastFrameAck{SessionID: e.SessionID}.Call(r.page)
	})()

	// Start the screencast via CDP
	quality := utils_clampInt(80, 1, 100)
	maxWidth := 1920
	maxHeight := 1080
	everyNth := 3 // Capture every 3rd frame to reduce load (~10fps equivalent)
	err := proto.PageStartScreencast{
		Format:        proto.PageStartScreencastFormatPng,
		Quality:       &quality,
		MaxWidth:      &maxWidth,
		MaxHeight:     &maxHeight,
		EveryNthFrame: &everyNth,
	}.Call(r.page)

	if err != nil {
		r.logger.Warnf("CDP screencast start failed: %v, falling back to screenshot loop", err)
		r.screenshotLoop()
		return
	}

	// Wait until Stop is called
	<-r.done

	// Stop CDP screencast
	_ = proto.PageStopScreencast{}.Call(r.page)
}

// screenshotLoop is a fallback that takes periodic screenshots when CDP screencast is unavailable.
func (r *ScreencastRecorder) screenshotLoop() {
	ticker := time.NewTicker(time.Second / time.Duration(r.fps))
	defer ticker.Stop()

	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			r.mu.Lock()
			if !r.recording {
				r.mu.Unlock()
				return
			}
			count := r.frameCount
			r.frameCount++
			r.mu.Unlock()

			img, err := r.page.Screenshot(false, nil)
			if err != nil {
				continue
			}

			framePath := filepath.Join(r.framesDir, fmt.Sprintf("frame_%06d.png", count))
			_ = os.WriteFile(framePath, img, 0644)
		}
	}
}

// Stop ends the recording and encodes captured frames into a video file.
func (r *ScreencastRecorder) Stop() error {
	r.mu.Lock()
	if !r.recording {
		r.mu.Unlock()
		return fmt.Errorf("no recording in progress")
	}
	r.recording = false
	frameCount := r.frameCount
	r.mu.Unlock()

	// Signal the capture goroutine to stop
	close(r.done)

	// Small delay for final frames to flush
	time.Sleep(200 * time.Millisecond)

	duration := time.Since(r.startTime)
	r.logger.Infof("Screencast stopped: %d frames captured in %v", frameCount, duration)

	if frameCount == 0 {
		r.cleanup()
		return fmt.Errorf("no frames captured during recording")
	}

	// Try to encode with ffmpeg
	if err := r.encodeWithFFmpeg(frameCount, duration); err != nil {
		r.logger.Warnf("ffmpeg encoding failed: %v, keeping individual frames", err)
		// Keep the frames directory as fallback
		return nil
	}

	// Clean up frames directory after successful encoding
	r.cleanup()
	return nil
}

// encodeWithFFmpeg uses ffmpeg to encode captured frames into an MP4 video.
func (r *ScreencastRecorder) encodeWithFFmpeg(frameCount int, duration time.Duration) error {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	// Calculate actual framerate from captured frames and duration
	actualFPS := float64(frameCount) / duration.Seconds()
	if actualFPS < 1 {
		actualFPS = 1
	}
	if actualFPS > 30 {
		actualFPS = 30
	}

	inputPattern := filepath.Join(r.framesDir, "frame_%06d.png")

	// Use resource-limited ffmpeg encoding:
	// -threads 2: limit to 2 threads (stays within 30-40% of 8 cores)
	// -preset ultrafast: minimize CPU usage
	// -crf 28: reasonable quality with lower CPU cost
	args := []string{
		"-y",
		"-framerate", fmt.Sprintf("%.1f", actualFPS),
		"-i", inputPattern,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "28",
		"-threads", "2",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		r.filename,
	}

	cmd := exec.Command(ffmpegPath, args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg encoding failed: %w", err)
	}

	// Verify the output file exists and has content
	info, err := os.Stat(r.filename)
	if err != nil {
		return fmt.Errorf("output video file not found: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("output video file is empty")
	}

	r.logger.Infof("Video encoded: %s (%d bytes, %.1f fps)", r.filename, info.Size(), actualFPS)
	return nil
}

// cleanup removes the temporary frames directory.
func (r *ScreencastRecorder) cleanup() {
	if r.framesDir != "" {
		os.RemoveAll(r.framesDir)
	}
}

// IsRecording returns whether recording is currently active.
func (r *ScreencastRecorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

// FrameCount returns the number of frames captured so far.
func (r *ScreencastRecorder) FrameCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.frameCount
}

// utils_clampInt clamps an integer value between min and max.
func utils_clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
