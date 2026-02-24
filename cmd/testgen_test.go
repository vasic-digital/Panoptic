package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// getTestgenRootCmd creates a fresh root command with the
// testgen subcommand tree registered for isolated testing.
func getTestgenRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "panoptic",
		Short: "Automated testing and recording application for multiple platforms",
	}

	cmd.PersistentFlags().StringVar(
		&cfgFile, "config", "",
		"config file (default is $HOME/.panoptic.yaml)",
	)
	cmd.PersistentFlags().String(
		"output", "./output",
		"output directory for screenshots and videos",
	)
	cmd.PersistentFlags().Bool(
		"verbose", false,
		"enable verbose logging",
	)

	_ = viper.BindPFlag(
		"output",
		cmd.PersistentFlags().Lookup("output"),
	)
	_ = viper.BindPFlag(
		"verbose",
		cmd.PersistentFlags().Lookup("verbose"),
	)

	tgCmd := &cobra.Command{
		Use:   "testgen",
		Short: "AI-powered test generation from screenshots",
	}

	genCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate test cases from a screenshot",
		RunE:  runTestgenGenerate,
	}
	genCmd.Flags().String(
		"screenshot", "",
		"path to the screenshot image file",
	)
	genCmd.Flags().String(
		"output", "",
		"path to write JSON output (stdout if omitted)",
	)
	genCmd.Flags().Int(
		"max-tests", 50,
		"maximum number of tests to generate",
	)
	genCmd.Flags().String(
		"app-type", "web",
		"application type (web, desktop, mobile)",
	)

	rptCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate an AI test report from a screenshot",
		RunE:  runTestgenReport,
	}
	rptCmd.Flags().String(
		"screenshot", "",
		"path to the screenshot image file",
	)
	rptCmd.Flags().String(
		"output", "testgen_report.md",
		"output path for the AI test report",
	)

	tgCmd.AddCommand(genCmd)
	tgCmd.AddCommand(rptCmd)
	cmd.AddCommand(tgCmd)

	return cmd
}

func TestTestgenGenerateCmd_NoScreenshot(t *testing.T) {
	viper.Reset()

	cmd := getTestgenRootCmd()
	cmd.SetArgs([]string{"testgen", "generate"})

	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(
		t, err.Error(),
		"--screenshot flag is required",
	)
}

func TestTestgenGenerateCmd_InvalidFile(t *testing.T) {
	viper.Reset()

	cmd := getTestgenRootCmd()
	cmd.SetArgs([]string{
		"testgen", "generate",
		"--screenshot", "/nonexistent/path/image.png",
	})

	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(
		t, err.Error(),
		"screenshot file does not exist",
	)
}

func TestTestgenReportCmd_NoScreenshot(t *testing.T) {
	viper.Reset()

	cmd := getTestgenRootCmd()
	cmd.SetArgs([]string{"testgen", "report"})

	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(
		t, err.Error(),
		"--screenshot flag is required",
	)
}
