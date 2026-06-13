package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"panoptic/internal/logger"
	"panoptic/internal/recvalidate"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// recvalidateCmd validates a recorded TUI/chat session video: it proves the
// recording shows a real model reply to each prompt, no error/warning text on
// screen, and the intended model selected — using Panoptic's OCR engine
// (ffmpeg + tesseract) and reusing the existing ErrorDetector classifier.
var recvalidateCmd = &cobra.Command{
	Use:   "recvalidate",
	Short: "Validate a recorded chat/TUI session video (real reply, no errors, model selected)",
	RunE:  runRecValidate,
}

func runRecValidate(cmd *cobra.Command, args []string) error {
	video, _ := cmd.Flags().GetString("video")
	if video == "" {
		return fmt.Errorf("--video flag is required")
	}
	if _, err := os.Stat(video); err != nil {
		return fmt.Errorf("video file not found: %s: %w", video, err)
	}

	prompts, _ := cmd.Flags().GetStringArray("prompt")
	model, _ := cmd.Flags().GetString("model")
	fps, _ := cmd.Flags().GetFloat64("fps")
	minReply, _ := cmd.Flags().GetInt("min-reply-chars")
	keepFrames, _ := cmd.Flags().GetBool("keep-frames")
	framesDir, _ := cmd.Flags().GetString("frames-dir")
	extraTokens, _ := cmd.Flags().GetStringArray("error-token")
	chromePatterns, _ := cmd.Flags().GetStringArray("chrome-pattern")
	replyMarkers, _ := cmd.Flags().GetStringArray("reply-marker")
	outPath, _ := cmd.Flags().GetString("json-out")

	log := logger.NewLogger(viper.GetBool("verbose"))
	v := recvalidate.NewValidator(*log)

	rep, err := v.Validate(context.Background(), recvalidate.Options{
		VideoPath:          video,
		ExpectedPrompts:    prompts,
		IntendedModel:      model,
		ExtraErrorTokens:   extraTokens,
		ChromeLinePatterns: chromePatterns,
		ReplyMarkers:       replyMarkers,
		FPS:                fps,
		FrameDir:           framesDir,
		KeepFrames:         keepFrames,
		MinReplyChars:      minReply,
	})
	if err != nil {
		return fmt.Errorf("validation failed to run: %w", err)
	}

	data, mErr := json.MarshalIndent(rep, "", "  ")
	if mErr != nil {
		return fmt.Errorf("failed to marshal report: %w", mErr)
	}
	if outPath != "" {
		if wErr := os.WriteFile(outPath, data, 0o600); wErr != nil {
			return fmt.Errorf("failed to write report: %w", wErr)
		}
		log.Infof("Validation report written to %s", outPath)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}

	// Honest exit semantics: SKIP (tools absent) is exit 0 but clearly marked;
	// a genuine FAIL is a non-zero exit so CI/Challenge harnesses catch it.
	if rep.Skipped {
		log.Infof("SKIP: %s", rep.SkipReason)
		return nil
	}
	if !rep.Pass {
		return fmt.Errorf("recorded-video validation FAILED (%d/%d checks failed)",
			countFailed(rep), len(rep.Checks))
	}
	log.Infof("PASS: recorded-video validation succeeded (%d checks)", len(rep.Checks))
	return nil
}

func countFailed(rep *recvalidate.Report) int {
	n := 0
	for _, c := range rep.Checks {
		if !c.Pass {
			n++
		}
	}
	return n
}

func init() {
	recvalidateCmd.Flags().String("video", "", "path to the recorded session video (mp4/webm)")
	recvalidateCmd.Flags().StringArray("prompt", nil, "expected prompt text (repeatable, ordered)")
	recvalidateCmd.Flags().String("model", "", "intended model name that must appear on screen")
	recvalidateCmd.Flags().Float64("fps", 1.0, "frame sampling rate (frames per video-second)")
	recvalidateCmd.Flags().Int("min-reply-chars", 12, "minimum prose chars after a prompt to count as a real reply")
	recvalidateCmd.Flags().Bool("keep-frames", false, "retain extracted frames as evidence")
	recvalidateCmd.Flags().String("frames-dir", "", "directory to write extracted frames (temp if omitted)")
	recvalidateCmd.Flags().StringArray("error-token", nil, "extra case-insensitive error phrase to flag (repeatable)")
	recvalidateCmd.Flags().StringArray("chrome-pattern", nil, "consumer-supplied case-insensitive regex matching ambient UI chrome lines to exclude from reply prose (repeatable)")
	recvalidateCmd.Flags().StringArray("reply-marker", nil, "assistant-turn prefix marking a model reply, e.g. 'AI:' (repeatable; generic chat defaults when omitted)")
	recvalidateCmd.Flags().String("json-out", "", "path to write the JSON report (stdout if omitted)")

	rootCmd.AddCommand(recvalidateCmd)
}
