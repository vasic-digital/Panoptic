package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"panoptic/internal/ai"
	"panoptic/internal/logger"
	"panoptic/internal/vision"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// generatedTestOutput is the JSON output struct for
// generated tests.
type generatedTestOutput struct {
	Name       string           `json:"name"`
	Category   string           `json:"category"`
	Priority   string           `json:"priority"`
	Confidence float64          `json:"confidence"`
	Steps      []testStepOutput `json:"steps"`
}

// testStepOutput is the JSON output struct for a single
// test step.
type testStepOutput struct {
	Action     string            `json:"action"`
	Target     string            `json:"target"`
	Value      string            `json:"value,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

var testgenCmd = &cobra.Command{
	Use:   "testgen",
	Short: "AI-powered test generation from screenshots",
}

var testgenGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate test cases from a screenshot",
	RunE:  runTestgenGenerate,
}

var testgenReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate an AI test report from a screenshot",
	RunE:  runTestgenReport,
}

func runTestgenGenerate(
	cmd *cobra.Command, args []string,
) error {
	screenshot, _ := cmd.Flags().GetString("screenshot")
	if screenshot == "" {
		return fmt.Errorf("--screenshot flag is required")
	}

	if _, err := os.Stat(screenshot); os.IsNotExist(err) {
		return fmt.Errorf(
			"screenshot file does not exist: %s",
			screenshot,
		)
	}

	appType, _ := cmd.Flags().GetString("app-type")
	maxTests, _ := cmd.Flags().GetInt("max-tests")

	log := logger.NewLogger(viper.GetBool("verbose"))
	detector := vision.NewElementDetector(*log)
	tg := ai.NewTestGenerator(*log, detector)

	elements, err := detector.DetectElements(screenshot)
	if err != nil {
		return fmt.Errorf(
			"failed to detect elements: %w", err,
		)
	}

	tests, err := tg.GenerateTestsFromElements(
		elements, appType,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to generate tests: %w", err,
		)
	}

	if maxTests > 0 && len(tests) > maxTests {
		tests = tests[:maxTests]
	}

	output := make(
		[]generatedTestOutput, 0, len(tests),
	)
	for _, t := range tests {
		steps := make(
			[]testStepOutput, 0, len(t.Steps),
		)
		for _, s := range t.Steps {
			steps = append(steps, testStepOutput{
				Action:     s.Action,
				Target:     s.Target,
				Value:      s.Value,
				Parameters: s.Parameters,
			})
		}
		output = append(output, generatedTestOutput{
			Name:       t.Name,
			Category:   t.Type,
			Priority:   t.Priority,
			Confidence: t.Confidence,
			Steps:      steps,
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf(
			"failed to marshal JSON: %w", err,
		)
	}

	outputPath, _ := cmd.Flags().GetString("output")
	if outputPath != "" {
		if writeErr := os.WriteFile(
			outputPath, data, 0600,
		); writeErr != nil {
			return fmt.Errorf(
				"failed to write output file: %w",
				writeErr,
			)
		}
		log.Infof("Output written to %s", outputPath)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func runTestgenReport(
	cmd *cobra.Command, args []string,
) error {
	screenshot, _ := cmd.Flags().GetString("screenshot")
	if screenshot == "" {
		return fmt.Errorf("--screenshot flag is required")
	}

	if _, err := os.Stat(screenshot); os.IsNotExist(err) {
		return fmt.Errorf(
			"screenshot file does not exist: %s",
			screenshot,
		)
	}

	outputPath, _ := cmd.Flags().GetString("output")

	log := logger.NewLogger(viper.GetBool("verbose"))
	detector := vision.NewElementDetector(*log)
	tg := ai.NewTestGenerator(*log, detector)

	elements, err := detector.DetectElements(screenshot)
	if err != nil {
		return fmt.Errorf(
			"failed to detect elements: %w", err,
		)
	}

	tests, err := tg.GenerateTestsFromElements(
		elements, "web",
	)
	if err != nil {
		return fmt.Errorf(
			"failed to generate tests: %w", err,
		)
	}

	analysis := tg.AnalyzeElements(elements)

	if err := tg.GenerateAITestReport(
		tests, analysis, outputPath,
	); err != nil {
		return fmt.Errorf(
			"failed to generate report: %w", err,
		)
	}

	log.Infof("AI test report generated at %s", outputPath)
	return nil
}

func init() {
	testgenGenerateCmd.Flags().String(
		"screenshot", "",
		"path to the screenshot image file",
	)
	testgenGenerateCmd.Flags().String(
		"output", "",
		"path to write JSON output (stdout if omitted)",
	)
	testgenGenerateCmd.Flags().Int(
		"max-tests", 50,
		"maximum number of tests to generate",
	)
	testgenGenerateCmd.Flags().String(
		"app-type", "web",
		"application type (web, desktop, mobile)",
	)

	testgenReportCmd.Flags().String(
		"screenshot", "",
		"path to the screenshot image file",
	)
	testgenReportCmd.Flags().String(
		"output", "testgen_report.md",
		"output path for the AI test report",
	)

	testgenCmd.AddCommand(testgenGenerateCmd)
	testgenCmd.AddCommand(testgenReportCmd)
	rootCmd.AddCommand(testgenCmd)
}
