package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"panoptic/internal/logger"
	"panoptic/internal/vision"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// detectedElementOutput is the JSON output struct for detected elements.
type detectedElementOutput struct {
	Type       string  `json:"type"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Confidence float64 `json:"confidence"`
	Text       string  `json:"text,omitempty"`
	Selector   string  `json:"selector"`
}

var visionCmd = &cobra.Command{
	Use:   "vision",
	Short: "Computer vision element detection from screenshots",
}

var visionDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect UI elements in a screenshot",
	RunE:  runVisionDetect,
}

var visionReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a visual report of detected elements",
	RunE:  runVisionReport,
}

func runVisionDetect(cmd *cobra.Command, args []string) error {
	screenshot, _ := cmd.Flags().GetString("screenshot")
	if screenshot == "" {
		return fmt.Errorf("--screenshot flag is required")
	}

	if _, err := os.Stat(screenshot); os.IsNotExist(err) {
		return fmt.Errorf(
			"screenshot file does not exist: %s", screenshot,
		)
	}

	log := logger.NewLogger(viper.GetBool("verbose"))
	detector := vision.NewElementDetector(*log)

	elements, err := detector.DetectElements(screenshot)
	if err != nil {
		return fmt.Errorf("failed to detect elements: %w", err)
	}

	output := make([]detectedElementOutput, 0, len(elements))
	for _, elem := range elements {
		output = append(output, detectedElementOutput{
			Type:       elem.Type,
			X:          elem.Position.X,
			Y:          elem.Position.Y,
			Width:      elem.Size.Width,
			Height:     elem.Size.Height,
			Confidence: elem.Confidence,
			Text:       elem.Text,
			Selector:   elem.Selector,
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	outputPath, _ := cmd.Flags().GetString("output")
	if outputPath != "" {
		if err := os.WriteFile(outputPath, data, 0600); err != nil {
			return fmt.Errorf(
				"failed to write output file: %w", err,
			)
		}
		log.Infof("Output written to %s", outputPath)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func runVisionReport(cmd *cobra.Command, args []string) error {
	screenshot, _ := cmd.Flags().GetString("screenshot")
	if screenshot == "" {
		return fmt.Errorf("--screenshot flag is required")
	}

	if _, err := os.Stat(screenshot); os.IsNotExist(err) {
		return fmt.Errorf(
			"screenshot file does not exist: %s", screenshot,
		)
	}

	log := logger.NewLogger(viper.GetBool("verbose"))
	detector := vision.NewElementDetector(*log)

	elements, err := detector.DetectElements(screenshot)
	if err != nil {
		return fmt.Errorf("failed to detect elements: %w", err)
	}

	outputDir, _ := cmd.Flags().GetString("output")
	if outputDir == "" {
		outputDir = "."
	}

	if err := detector.GenerateVisualReport(
		elements, outputDir,
	); err != nil {
		return fmt.Errorf(
			"failed to generate visual report: %w", err,
		)
	}

	log.Infof("Visual report generated in %s", outputDir)
	return nil
}

func init() {
	visionDetectCmd.Flags().String(
		"screenshot", "",
		"path to the screenshot image file",
	)
	visionDetectCmd.Flags().String(
		"output", "",
		"path to write JSON output (stdout if omitted)",
	)

	visionReportCmd.Flags().String(
		"screenshot", "",
		"path to the screenshot image file",
	)
	visionReportCmd.Flags().String(
		"output", "",
		"output directory for the visual report",
	)

	visionCmd.AddCommand(visionDetectCmd)
	visionCmd.AddCommand(visionReportCmd)
	rootCmd.AddCommand(visionCmd)
}
