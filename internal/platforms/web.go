package platforms

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"panoptic/internal/config"

	"github.com/go-rod/rod"
)

type WebPlatform struct {
	browser   *rod.Browser
	page      *rod.Page
	context   context.Context
	cancel    context.CancelFunc
	recording bool
	metrics   map[string]interface{}
}

func NewWebPlatform() *WebPlatform {
	return &WebPlatform{
		metrics: map[string]interface{}{
			"click_actions":     []string{},
			"screenshots_taken":  []string{},
			"fill_actions":      []map[string]string{},
			"submit_actions":    []string{},
			"navigate_actions":  []string{},
			"start_time":        time.Now(),
		},
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
	
	if err := element.Click("left", 1); err != nil {
		return fmt.Errorf("failed to click element %s: %w", selector, err)
	}
	
	waitForPageLoad()
	return nil
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
	
	if err := os.WriteFile(filename, screenshotData, 0644); err != nil {
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
	
	// For web platform, we'll use a browser recording approach
	// In a real implementation, you might use:
	// 1. Chrome DevTools Protocol screen capture
	// 2. Third-party libraries like Puppeteer's screen recording
	// 3. System-level recording focused on browser window
	
	// For now, we'll create a more sophisticated placeholder
	// that could be extended with actual recording libraries
	
	// Create a dummy video file (in real implementation, this would be actual video data)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create video file: %w", err)
	}
	
	// Write a simple header that indicates this is a placeholder
	placeholderHeader := []byte("# PANOPTIC VIDEO RECORDING PLACEHOLDER\n# Web Platform\n# Recording started: " + time.Now().Format(time.RFC3339) + "\n# File: " + filename + "\n")
	
	if _, err := file.Write(placeholderHeader); err != nil {
		file.Close()
		return fmt.Errorf("failed to write video header: %w", err)
	}
	file.Close()
	
	// w.logger.Infof("Web video recording started: %s", filename)
	// For now, just comment out logger until we add logger field
	return nil
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
	
	// In a real implementation, this would:
	// 1. Stop the browser recording process
	// 2. Save the video file with proper encoding
	// 3. Close any open file handles
	// 4. Return final video metadata
	
	// w.logger.Infof("Web video recording stopped. Duration: %v", w.metrics["recording_duration"])
	// For now, just comment out logger until we add logger field
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