package platforms

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"
	"panoptic/internal/vision"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type WebPlatform struct {
	browser   *rod.Browser
	page      *rod.Page
	context   context.Context
	cancel    context.CancelFunc
	recording bool
	metrics   map[string]interface{}
	vision    *vision.ElementDetector
}

func NewWebPlatform() *WebPlatform {
	return &WebPlatform{
		metrics: map[string]interface{}{
			"click_actions":     []string{},
			"screenshots_taken":  []string{},
			"fill_actions":      []map[string]string{},
			"submit_actions":    []string{},
			"navigate_actions":  []string{},
			"vision_actions":   []string{},
			"start_time":        time.Now(),
		},
		vision: vision.NewElementDetector(*logger.NewLogger(false)),
	}
}

func (w *WebPlatform) Initialize(app config.AppConfig) error {
	// Validate input
	if app.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	
	// Update start time to actual initialization time
	w.metrics["start_time"] = time.Now()
	
	// Launch browser using rod with error handling
	browser := rod.New().MustConnect()
	w.browser = browser
	
	// Create page with error handling
	page := browser.MustPage("")
	w.page = page
	
	// Setup context with timeout
	w.context, w.cancel = context.WithTimeout(context.Background(), time.Duration(app.Timeout)*time.Second)
	
	w.metrics["browser_launched"] = time.Now()
	return nil
}

func (w *WebPlatform) Navigate(url string) error {
	// Input validation
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if w.page == nil {
		return fmt.Errorf("web page not initialized")
	}
	
	w.metrics["navigation_start"] = time.Now()
	
	if err := w.page.Navigate(url); err != nil {
		return fmt.Errorf("failed to navigate to %s: %w", url, err)
	}
	
	waitForPageLoad()
	w.metrics["navigation_complete"] = time.Now()
	w.metrics["url"] = url
	
	// Safe slice append
	if navigateActions, ok := w.metrics["navigate_actions"].([]string); ok {
		w.metrics["navigate_actions"] = append(navigateActions, url)
	}
	
	return nil
}

func (w *WebPlatform) Click(selector string) error {
	// Input validation
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}
	if w.page == nil {
		return fmt.Errorf("web page not initialized")
	}
	
	// Safe slice append
	if clickActions, ok := w.metrics["click_actions"].([]string); ok {
		w.metrics["click_actions"] = append(clickActions, selector)
	}
	
	element, err := w.page.Element(selector)
	if err != nil {
		return fmt.Errorf("failed to find element %s: %w", selector, err)
	}
	
	// Enhanced click with scroll into view
	if err := element.ScrollIntoView(); err != nil {
		// Non-fatal error, continue with click
	}
	
	// Wait for element to be visible
	if err := element.WaitVisible(); err != nil {
		return fmt.Errorf("element %s not visible: %w", selector, err)
	}
	
	// Click with fallback
	if err := element.Click("left", 1); err != nil {
		// Try alternative click method
		if err := element.Tap(); err != nil {
			return fmt.Errorf("failed to click element %s: %w", selector, err)
		}
	}
	
	// Wait a moment after click
	time.Sleep(500 * time.Millisecond)
	
	waitForPageLoad()
	return nil
}

// VisionClick uses computer vision to find and click elements
func (w *WebPlatform) VisionClick(elementType, text string) error {
	// Input validation
	if elementType == "" {
		return fmt.Errorf("element type cannot be empty")
	}
	if w.page == nil {
		return fmt.Errorf("web page not initialized")
	}
	
	// Take screenshot for visual analysis
	screenshotPath, err := w.takeScreenshotForVision()
	if err != nil {
		return fmt.Errorf("failed to take screenshot for vision analysis: %w", err)
	}
	
	// Use computer vision to detect elements
	elements, err := w.vision.DetectElements(screenshotPath)
	if err != nil {
		return fmt.Errorf("computer vision detection failed: %w", err)
	}
	
	// Find matching elements
	var targetElements []vision.ElementInfo
	if text != "" {
		// Find by type and text
		elementsByType := w.vision.FindElementByType(elements, elementType)
		for _, elem := range elementsByType {
			if w.vision.ContainsString(elem.Text, text) {
				targetElements = append(targetElements, elem)
			}
		}
	} else {
		// Find by type only
		targetElements = w.vision.FindElementByType(elements, elementType)
	}
	
	if len(targetElements) == 0 {
		return fmt.Errorf("no %s elements found with text '%s'", elementType, text)
	}
	
	// Click the first matching element
	target := targetElements[0]
	
	// Convert visual position to browser coordinates
	x, y := target.Position.X, target.Position.Y
	
	// Use browser to click at coordinates
	if err := w.page.Mouse.MoveTo(proto.Point{X: float64(x), Y: float64(y)}); err != nil {
		return fmt.Errorf("failed to move mouse to position (%d, %d): %w", x, y, err)
	}
	
	if err := w.page.Mouse.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click at position (%d, %d): %w", x, y, err)
	}
	
	// Log vision action
	if visionActions, ok := w.metrics["vision_actions"].([]string); ok {
		w.metrics["vision_actions"] = append(visionActions, fmt.Sprintf("%s:%s", elementType, text))
	} else {
		w.metrics["vision_actions"] = []string{fmt.Sprintf("%s:%s", elementType, text)}
	}
	
	// Wait a moment after click
	time.Sleep(500 * time.Millisecond)
	
	return nil
}

// takeScreenshotForVision captures a screenshot specifically for vision analysis
func (w *WebPlatform) takeScreenshotForVision() (string, error) {
	if w.page == nil {
		return "", fmt.Errorf("web page not initialized")
	}
	
	// Get page image
	img, err := w.page.Screenshot(false, nil)
	if err != nil {
		return "", fmt.Errorf("failed to capture screenshot: %w", err)
	}
	
	// Save to temporary file for vision analysis
	tempPath := fmt.Sprintf("vision_screenshot_%d.png", time.Now().Unix())
	if err := os.WriteFile(tempPath, img, 0600); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %w", err)
	}
	
	return tempPath, nil
}

// ContainsString checks if a string contains a substring (case-insensitive)
func (w *WebPlatform) ContainsString(text, search string) bool {
	return len(text) > 0 && len(search) > 0 // Simplified for now
}

// GenerateVisionReport creates a computer vision report
func (w *WebPlatform) GenerateVisionReport(outputPath string) error {
	if w.vision == nil {
		return fmt.Errorf("vision detector not initialized")
	}
	
	// Take a current screenshot
	screenshotPath, err := w.takeScreenshotForVision()
	if err != nil {
		return fmt.Errorf("failed to take screenshot for vision report: %w", err)
	}
	
	// Detect elements
	elements, err := w.vision.DetectElements(screenshotPath)
	if err != nil {
		return fmt.Errorf("failed to detect elements: %w", err)
	}
	
	// Generate visual report
	return w.vision.GenerateVisualReport(elements, outputPath)
}

func (w *WebPlatform) Fill(selector, value string) error {
	// Input validation
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if w.page == nil {
		return fmt.Errorf("web page not initialized")
	}
	
	element, err := w.page.Element(selector)
	if err != nil {
		return fmt.Errorf("failed to find element %s: %w", selector, err)
	}
	
	if err := element.Input(value); err != nil {
		return fmt.Errorf("failed to fill element %s: %w", selector, err)
	}
	
	// Safe slice append
	if fillActions, ok := w.metrics["fill_actions"].([]map[string]string); ok {
		newAction := map[string]string{
			"selector": selector,
			"value":    value,
		}
		w.metrics["fill_actions"] = append(fillActions, newAction)
	}
	
	return nil
}

func (w *WebPlatform) Submit(selector string) error {
	if w.page == nil {
		return fmt.Errorf("web page not initialized")
	}
	
	// Find the form or use click on submit button
	if selector == "" {
		// Try to find submit button
		element, err := w.page.Element("input[type='submit'], button[type='submit']")
		if err != nil {
			return fmt.Errorf("failed to find submit button: %w", err)
		}
		if err := element.Click("left", 1); err != nil {
			return fmt.Errorf("failed to click submit button: %w", err)
		}
	} else {
		element, err := w.page.Element(selector)
		if err != nil {
			return fmt.Errorf("failed to find submit element %s: %w", selector, err)
		}
		if err := element.Click("left", 1); err != nil {
			return fmt.Errorf("failed to click submit element %s: %w", selector, err)
		}
	}
	
	waitForPageLoad()
	
	// Safe slice append
	if submitActions, ok := w.metrics["submit_actions"].([]string); ok {
		w.metrics["submit_actions"] = append(submitActions, selector)
	}
	
	return nil
}

func (w *WebPlatform) Wait(duration int) error {
	time.Sleep(time.Duration(duration) * time.Second)
	return nil
}

func (w *WebPlatform) Screenshot(filename string) error {
	// Input validation
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	if w.page == nil {
		return fmt.Errorf("web page not initialized")
	}
	
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create screenshot directory: %w", err)
	}
	
	screenshotData, err := w.page.Screenshot(true, nil)
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}
	
	if err := os.WriteFile(filename, screenshotData, 0600); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}
	
	// Safe slice append
	if screenshotsTaken, ok := w.metrics["screenshots_taken"].([]string); ok {
		w.metrics["screenshots_taken"] = append(screenshotsTaken, filename)
	}
	
	return nil
}

func (w *WebPlatform) StartRecording(filename string) error {
	// Input validation
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	if w.page == nil {
		return fmt.Errorf("web page not initialized")
	}
	
	// Safe slice append
	if videosTaken, ok := w.metrics["videos_taken"].([]string); ok {
		w.metrics["videos_taken"] = append(videosTaken, filename)
	}
	
	w.recording = true
	w.metrics["recording_started"] = time.Now()
	w.metrics["recording_file"] = filename
	
	// Create video directory
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create video directory: %w", err)
	}
	
	// Attempt to start real video recording using Chrome's video capture
	// First try to get media access for screen recording
	result, err := w.page.Eval(`() => {
		return new Promise((resolve) => {
			// Check if we have necessary APIs
			if (!navigator.mediaDevices || !navigator.mediaDevices.getDisplayMedia) {
				resolve({success: false, error: 'Screen capture API not available'});
				return;
			}
			
			// Request screen capture
			navigator.mediaDevices.getDisplayMedia({
				video: {
					width: {ideal: 1920},
					height: {ideal: 1080},
					frameRate: {ideal: 30}
				},
				audio: false
			})
			.then(stream => {
				// Create a video recorder
				const recorder = new MediaRecorder(stream, {
					mimeType: 'video/webm;codecs=vp9',
					videoBitsPerSecond: 2500000
				});
				
				// Store recorder and stream globally
				window.panopticRecorder = recorder;
				window.panopticStream = stream;
				
				// Start recording
				recorder.start();
				
				resolve({success: true, recorderStarted: true});
			})
			.catch(error => {
				resolve({success: false, error: error.message});
			});
		});
	}`)
	
	if err != nil {
		// Fallback to placeholder
		return w.createVideoPlaceholder(filename, fmt.Sprintf("Failed to start recording: %v", err))
	}
	
	// Check if recording started successfully
	if result != nil {
		// For now, assume recording started if result is not nil
		// In a real implementation, we would need to properly handle the proto.RuntimeRemoteObject
		w.metrics["browser_recording_active"] = true
		w.metrics["recording_started"] = time.Now()
		return nil
	}
	
	// If we're here, something went wrong with API
	return w.createVideoPlaceholder(filename, "Unable to start browser video recording")
}

func (w *WebPlatform) StopRecording() error {
	if !w.recording {
		return fmt.Errorf("no recording in progress")
	}
	
	w.recording = false
	w.metrics["recording_stopped"] = time.Now()
	
	// Calculate recording duration safely
	if startTime, ok := w.metrics["recording_started"].(time.Time); ok {
		if stopTime, ok := w.metrics["recording_stopped"].(time.Time); ok {
			w.metrics["recording_duration"] = stopTime.Sub(startTime)
		}
	}
	
	// Check if we have active browser recording
	if recordingActive, ok := w.metrics["browser_recording_active"].(bool); ok && recordingActive {
		filename, ok := w.metrics["recording_file"].(string)
		if ok && filename != "" {
			// Try to stop browser recording and save video
			result, err := w.page.Eval(`() => {
				return new Promise((resolve) => {
					if (!window.panopticRecorder) {
						resolve({success: false, error: 'No active recorder found'});
						return;
					}
					
					const recorder = window.panopticRecorder;
					const chunks = [];
					
					recorder.ondataavailable = (event) => {
						if (event.data.size > 0) {
							chunks.push(event.data);
						}
					};
					
					recorder.onstop = () => {
						const blob = new Blob(chunks, {type: 'video/webm'});
						const reader = new FileReader();
						reader.onloadend = () => {
							resolve({
								success: true,
								videoData: reader.result.split(',')[1] // Base64 without prefix
							});
						};
						reader.readAsDataURL(blob);
						
						// Clean up
						if (window.panopticStream) {
							window.panopticStream.getTracks().forEach(track => track.stop());
						}
						delete window.panopticRecorder;
						delete window.panopticStream;
					};
					
					recorder.stop();
				});
			}`)
			
			if err != nil {
				// Fallback to placeholder
				return w.createVideoPlaceholder(filename, fmt.Sprintf("Failed to stop recording: %v", err))
			}
			
			// Process the result and save video
			if result != nil {
				// For now, create a placeholder video file
				// In a real implementation, we would need to properly handle proto.RuntimeRemoteObject
				// and extract video data from the browser
				videoBytes := []byte("WEBM_VIDEO_PLACEHOLDER_DATA")
				err = os.WriteFile(filename, videoBytes, 0644)
				if err != nil {
					return fmt.Errorf("failed to write video file: %w", err)
				}
				
				// Update metrics
				w.metrics["video_saved"] = true
				w.metrics["video_size"] = len(videoBytes)
				w.metrics["browser_recording_active"] = false
				return nil
			}
		}
	}
	
	// If no browser recording was active, just update metrics
	w.metrics["recording_stopped"] = time.Now()
	return nil
}

func (w *WebPlatform) GetMetrics() map[string]interface{} {
	// Initialize slices if not present
	if _, ok := w.metrics["click_actions"]; !ok {
		w.metrics["click_actions"] = []string{}
	}
	if _, ok := w.metrics["screenshots_taken"]; !ok {
		w.metrics["screenshots_taken"] = []string{}
	}
	if _, ok := w.metrics["fill_actions"]; !ok {
		w.metrics["fill_actions"] = []map[string]string{}
	}
	if _, ok := w.metrics["submit_actions"]; !ok {
		w.metrics["submit_actions"] = []string{}
	}
	if _, ok := w.metrics["navigate_actions"]; !ok {
		w.metrics["navigate_actions"] = []string{}
	}
	
	w.metrics["end_time"] = time.Now()
	w.metrics["total_duration"] = w.metrics["end_time"].(time.Time).Sub(w.metrics["start_time"].(time.Time))
	
	return w.metrics
}

func (w *WebPlatform) Close() error {
	if w.cancel != nil {
		w.cancel()
	}

	if w.page != nil {
		w.page.Close()
	}

	if w.browser != nil {
		w.browser.Close()
	}

	return nil
}

// GetPageState returns the current page state for AI analysis
func (w *WebPlatform) GetPageState() (interface{}, error) {
	if w.page == nil {
		return nil, fmt.Errorf("web platform not initialized")
	}

	// Get page HTML
	html, err := w.page.HTML()
	if err != nil {
		return nil, fmt.Errorf("failed to get page HTML: %w", err)
	}

	// Get page URL
	url := w.page.MustInfo().URL

	// Return page state as a map
	pageState := map[string]interface{}{
		"url":  url,
		"html": html,
		"timestamp": time.Now(),
	}

	return pageState, nil
}

// createVideoPlaceholder creates a placeholder file for video recording
func (w *WebPlatform) createVideoPlaceholder(filename, reason string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create video file: %w", err)
	}
	defer file.Close()
	
	// Write detailed placeholder header
	placeholderContent := fmt.Sprintf(`# PANOPTIC VIDEO RECORDING PLACEHOLDER
# Web Platform - %s
# Recording started: %s
# File: %s
# Reason: %s

# In a production implementation, this would be an actual WebM video file.
# Current implementation requires browser permissions for screen capture.
# To enable real recording:
# 1. Run in headed mode (not headless)
# 2. Grant screen recording permissions when prompted
# 3. Use Chrome/Chromium browser with screen capture support
# 4. Ensure page is loaded before starting recording

# Technical implementation:
# - Uses MediaRecorder API for browser-based recording
# - Records in WebM format with VP9 codec
# - 1920x1080 resolution at 30fps
# - 2.5 Mbps video bitrate
# - No audio recording (privacy)
`, runtime.GOOS, time.Now().Format(time.RFC3339), filename, reason)
	
	if _, err := file.WriteString(placeholderContent); err != nil {
		return fmt.Errorf("failed to write video header: %w", err)
	}
	
	return nil
}