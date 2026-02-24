package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// newVisionTestRootCmd creates a fresh command tree for vision tests
// to avoid state pollution from other tests.
func newVisionTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "panoptic"}
	root.PersistentFlags().Bool(
		"verbose", false, "enable verbose logging",
	)

	vis := &cobra.Command{
		Use:   "vision",
		Short: "Computer vision element detection from screenshots",
	}

	detect := &cobra.Command{
		Use:   "detect",
		Short: "Detect UI elements in a screenshot",
		RunE:  runVisionDetect,
	}
	detect.Flags().String(
		"screenshot", "",
		"path to the screenshot image file",
	)
	detect.Flags().String(
		"output", "",
		"path to write JSON output (stdout if omitted)",
	)

	report := &cobra.Command{
		Use:   "report",
		Short: "Generate a visual report of detected elements",
		RunE:  runVisionReport,
	}
	report.Flags().String(
		"screenshot", "",
		"path to the screenshot image file",
	)
	report.Flags().String(
		"output", "",
		"output directory for the visual report",
	)

	vis.AddCommand(detect)
	vis.AddCommand(report)
	root.AddCommand(vis)

	return root
}

func TestVisionDetectCmd_NoScreenshot(t *testing.T) {
	cmd := newVisionTestRootCmd()
	cmd.SetArgs([]string{"vision", "detect"})

	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--screenshot flag is required")
}

func TestVisionDetectCmd_InvalidFile(t *testing.T) {
	cmd := newVisionTestRootCmd()
	cmd.SetArgs([]string{
		"vision", "detect",
		"--screenshot", "/nonexistent.png",
	})

	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "screenshot file does not exist")
}

func TestVisionReportCmd_NoScreenshot(t *testing.T) {
	cmd := newVisionTestRootCmd()
	cmd.SetArgs([]string{"vision", "report"})

	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--screenshot flag is required")
}
