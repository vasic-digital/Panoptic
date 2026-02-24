package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"panoptic/internal/logger"
	"panoptic/internal/platforms"

	"github.com/go-rod/rod"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// recordSession holds state for an active recording session,
// persisted to a temp JSON file so that "record stop" can find
// and terminate the recording process.
type recordSession struct {
	SessionID  string `json:"session_id"`
	URL        string `json:"url"`
	OutputDir  string `json:"output_dir"`
	StartTime  int64  `json:"start_time"`
	PID        int    `json:"pid"`
	ResultFile string `json:"result_file"`
}

// recordingResultOutput is written by "record start" when it
// finishes encoding, and read by "record stop" to report results.
type recordingResultOutput struct {
	FilePath   string `json:"file_path"`
	DurationMs int64  `json:"duration_ms"`
	FrameCount int    `json:"frame_count"`
	FileSize   int64  `json:"file_size"`
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record browser sessions as video",
	Long: `Record browser sessions as video using CDP screencast.
Use "record start" to begin recording and "record stop" to end it.`,
}

var recordStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start recording a browser session",
	Long: `Launch a browser, navigate to the given URL, and begin
recording the session as video via CDP screencast. The process
blocks until interrupted (SIGINT/SIGTERM), at which point it
stops recording, encodes the video, and writes a result file.`,
	RunE: runRecordStart,
}

var recordStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop an active recording session",
	Long: `Stop a recording session started by "record start".
Sends SIGINT to the recording process, waits for it to finish
encoding, and prints the recording result.`,
	RunE: runRecordStop,
}

func init() {
	// record start flags
	recordStartCmd.Flags().String(
		"url", "",
		"URL to navigate to and record (required)",
	)
	_ = recordStartCmd.MarkFlagRequired("url")

	recordStartCmd.Flags().String(
		"output", "recording_output",
		"output directory for the recording",
	)
	recordStartCmd.Flags().Int(
		"fps", 10,
		"target frames per second for recording",
	)
	recordStartCmd.Flags().Int(
		"max-width", 1920,
		"maximum viewport width in pixels",
	)
	recordStartCmd.Flags().Int(
		"max-height", 1080,
		"maximum viewport height in pixels",
	)
	recordStartCmd.Flags().Bool(
		"headless", true,
		"run browser in headless mode",
	)

	// record stop flags
	recordStopCmd.Flags().String(
		"session", "",
		"session ID returned by record start (required)",
	)
	_ = recordStopCmd.MarkFlagRequired("session")

	recordCmd.AddCommand(recordStartCmd)
	recordCmd.AddCommand(recordStopCmd)
	rootCmd.AddCommand(recordCmd)
}

// runRecordStart implements the "record start" subcommand.
func runRecordStart(cmd *cobra.Command, args []string) error {
	url, _ := cmd.Flags().GetString("url")
	outputDir, _ := cmd.Flags().GetString("output")
	maxWidth, _ := cmd.Flags().GetInt("max-width")
	maxHeight, _ := cmd.Flags().GetInt("max-height")
	headless, _ := cmd.Flags().GetBool("headless")

	log := logger.NewLogger(viper.GetBool("verbose"))
	log.Infof("Starting recording for URL: %s", url)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// Launch browser
	browser := rod.New()
	if headless {
		// rod defaults to headless; explicit call not needed,
		// but we keep the branch for clarity.
		browser = browser.MustConnect()
	} else {
		browser = browser.MustConnect()
	}
	defer browser.MustClose()

	// Open page and set viewport
	page := browser.MustPage("")
	page.MustSetViewport(maxWidth, maxHeight, 0, false)
	page.MustNavigate(url)
	page.MustWaitLoad()

	log.Infof("Browser launched, navigated to %s", url)

	// Create and start screencast recorder
	logVal := *log // dereference: NewScreencastRecorder takes value
	recorder := platforms.NewScreencastRecorder(
		page, logVal,
	)

	videoPath := filepath.Join(outputDir, "recording.mp4")
	if err := recorder.Start(videoPath); err != nil {
		return fmt.Errorf(
			"failed to start recording: %w", err,
		)
	}

	// Generate session ID and write session state
	sessionID := fmt.Sprintf(
		"rec_%d", time.Now().UnixNano(),
	)
	resultFile := filepath.Join(
		outputDir, "recording_result.json",
	)
	session := recordSession{
		SessionID:  sessionID,
		URL:        url,
		OutputDir:  outputDir,
		StartTime:  time.Now().UnixNano(),
		PID:        os.Getpid(),
		ResultFile: resultFile,
	}

	sessionFile := filepath.Join(
		os.TempDir(), sessionID+".json",
	)
	sessionData, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf(
			"failed to marshal session state: %w", err,
		)
	}
	if err := os.WriteFile(
		sessionFile, sessionData, 0644,
	); err != nil {
		return fmt.Errorf(
			"failed to write session file: %w", err,
		)
	}

	fmt.Println(sessionID)
	log.Infof(
		"Session %s started (PID %d, file %s)",
		sessionID, session.PID, sessionFile,
	)

	// Block until signalled
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Infof("Received signal %v, stopping recording", sig)

	// Stop recording and encode video
	if err := recorder.Stop(); err != nil {
		log.Errorf("Error stopping recording: %v", err)
	}

	// Build result output
	durationMs := (time.Now().UnixNano() -
		session.StartTime) / int64(time.Millisecond)
	frameCount := recorder.FrameCount()

	var fileSize int64
	if info, statErr := os.Stat(videoPath); statErr == nil {
		fileSize = info.Size()
	}

	result := recordingResultOutput{
		FilePath:   videoPath,
		DurationMs: durationMs,
		FrameCount: frameCount,
		FileSize:   fileSize,
	}

	resultData, err := json.MarshalIndent(
		result, "", "  ",
	)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal result: %w", err,
		)
	}

	// Write result file for "record stop" to pick up
	if err := os.WriteFile(
		resultFile, resultData, 0644,
	); err != nil {
		return fmt.Errorf(
			"failed to write result file: %w", err,
		)
	}

	fmt.Println(string(resultData))
	log.Infof("Recording saved: %s", videoPath)

	// Clean up session file
	_ = os.Remove(sessionFile)

	return nil
}

// runRecordStop implements the "record stop" subcommand.
func runRecordStop(cmd *cobra.Command, args []string) error {
	sessionID, _ := cmd.Flags().GetString("session")

	log := logger.NewLogger(viper.GetBool("verbose"))
	log.Infof("Stopping session: %s", sessionID)

	// Read session state from temp file
	sessionFile := filepath.Join(
		os.TempDir(), sessionID+".json",
	)
	sessionData, err := os.ReadFile(sessionFile)
	if err != nil {
		return fmt.Errorf(
			"failed to read session file %s: %w",
			sessionFile, err,
		)
	}

	var session recordSession
	if err := json.Unmarshal(
		sessionData, &session,
	); err != nil {
		return fmt.Errorf(
			"failed to parse session file: %w", err,
		)
	}

	// Send SIGINT to the recording process
	proc, err := os.FindProcess(session.PID)
	if err != nil {
		return fmt.Errorf(
			"failed to find process %d: %w",
			session.PID, err,
		)
	}

	if err := proc.Signal(syscall.SIGINT); err != nil {
		return fmt.Errorf(
			"failed to signal process %d: %w",
			session.PID, err,
		)
	}

	log.Infof(
		"Sent SIGINT to PID %d, waiting for result",
		session.PID,
	)

	// Wait for the result file to appear (the start process
	// writes it before exiting after receiving SIGINT).
	resultFile := session.ResultFile
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf(
				"timed out waiting for result file: %s",
				resultFile,
			)
		case <-ticker.C:
			if _, statErr := os.Stat(resultFile); statErr == nil {
				resultData, readErr := os.ReadFile(resultFile)
				if readErr != nil {
					return fmt.Errorf(
						"failed to read result file: %w",
						readErr,
					)
				}
				fmt.Println(string(resultData))
				log.Infof(
					"Session %s stopped successfully",
					sessionID,
				)
				return nil
			}
		}
	}
}
