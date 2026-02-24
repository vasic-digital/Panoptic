package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// newErrorsTestRootCmd creates a fresh command tree for
// errors tests to avoid state pollution from other tests.
func newErrorsTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "panoptic"}
	root.PersistentFlags().Bool(
		"verbose", false, "enable verbose logging",
	)

	errs := &cobra.Command{
		Use:   "errors",
		Short: "Error detection and analysis commands",
	}

	analyze := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze log input for errors",
		RunE:  runErrorsAnalyze,
	}
	analyze.Flags().String(
		"input", "",
		"input file path or \"-\" for stdin",
	)
	analyze.Flags().String(
		"output", "",
		"path to write JSON output (stdout if omitted)",
	)

	errs.AddCommand(analyze)
	root.AddCommand(errs)

	return root
}

func TestErrorsAnalyzeCmd_NoInput(t *testing.T) {
	cmd := newErrorsTestRootCmd()
	cmd.SetArgs([]string{"errors", "analyze"})

	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(
		t, err.Error(), "--input flag is required",
	)
}

func TestErrorsAnalyzeCmd_InvalidFile(t *testing.T) {
	cmd := newErrorsTestRootCmd()
	cmd.SetArgs([]string{
		"errors", "analyze",
		"--input", "/nonexistent/path/to/file.log",
	})

	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(
		t, err.Error(), "input file does not exist",
	)
}
