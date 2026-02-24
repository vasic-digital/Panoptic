package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// getRecordRootCmd creates a fresh root command with the record
// subcommands attached, avoiding shared state between tests.
func getRecordRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "panoptic",
		Short: "Automated testing and recording application",
	}
	root.PersistentFlags().Bool(
		"verbose", false, "enable verbose logging",
	)

	rec := &cobra.Command{
		Use:   "record",
		Short: "Record browser sessions as video",
	}

	start := &cobra.Command{
		Use:   "start",
		Short: "Start recording a browser session",
		RunE:  runRecordStart,
	}
	start.Flags().String("url", "", "URL to record (required)")
	_ = start.MarkFlagRequired("url")
	start.Flags().String(
		"output", "recording_output",
		"output directory for the recording",
	)
	start.Flags().Int("fps", 10, "target frames per second")
	start.Flags().Int("max-width", 1920, "max viewport width")
	start.Flags().Int("max-height", 1080, "max viewport height")
	start.Flags().Bool("headless", true, "headless mode")

	stop := &cobra.Command{
		Use:   "stop",
		Short: "Stop an active recording session",
		RunE:  runRecordStop,
	}
	stop.Flags().String(
		"session", "",
		"session ID returned by record start (required)",
	)
	_ = stop.MarkFlagRequired("session")

	rec.AddCommand(start)
	rec.AddCommand(stop)
	root.AddCommand(rec)
	return root
}

func TestRecordStartCmd_NoURL(t *testing.T) {
	root := getRecordRootCmd()
	root.SetArgs([]string{"record", "start"})

	out := &strings.Builder{}
	root.SetOut(out)
	root.SetErr(out)

	err := root.Execute()
	assert.Error(t, err)

	combined := out.String() + err.Error()
	assert.True(t,
		strings.Contains(combined, "required") ||
			strings.Contains(combined, "url"),
		"expected error about missing --url flag, got: %s",
		combined,
	)
}

func TestRecordStopCmd_NoSession(t *testing.T) {
	root := getRecordRootCmd()
	root.SetArgs([]string{"record", "stop"})

	out := &strings.Builder{}
	root.SetOut(out)
	root.SetErr(out)

	err := root.Execute()
	assert.Error(t, err)

	combined := out.String() + err.Error()
	assert.True(t,
		strings.Contains(combined, "required") ||
			strings.Contains(combined, "session"),
		"expected error about missing --session flag, got: %s",
		combined,
	)
}
