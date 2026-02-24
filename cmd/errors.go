package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"panoptic/internal/ai"
	"panoptic/internal/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// errorAnalysisOutput is the JSON output struct for error analysis.
type errorAnalysisOutput struct {
	TotalErrors     int                    `json:"total_errors"`
	Categories      map[string]int         `json:"categories"`
	Severity        map[string]int         `json:"severity"`
	Recommendations []recommendationOutput `json:"recommendations"`
}

// recommendationOutput is the JSON output struct for a recommendation.
type recommendationOutput struct {
	Type     string `json:"type"`
	Priority string `json:"priority"`
	Message  string `json:"message"`
	Impact   string `json:"impact"`
}

var errorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Error detection and analysis commands",
}

var errorsAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze log input for errors",
	RunE:  runErrorsAnalyze,
}

func runErrorsAnalyze(
	cmd *cobra.Command, args []string,
) error {
	input, _ := cmd.Flags().GetString("input")
	if input == "" {
		return fmt.Errorf("--input flag is required")
	}

	log := logger.NewLogger(viper.GetBool("verbose"))
	ed := ai.NewErrorDetector(*log)

	text, err := readInput(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	lines := strings.Split(
		strings.TrimSpace(text), "\n",
	)

	messages := make([]ai.ErrorMessage, 0, len(lines))
	now := time.Now()
	for _, line := range lines {
		if line == "" {
			continue
		}
		messages = append(messages, ai.ErrorMessage{
			Message:   line,
			Source:    "log",
			Timestamp: now,
			Context:   map[string]interface{}{},
			Level:     "error",
		})
	}

	detected := ed.DetectErrors(messages)
	analysis := ed.AnalyzeErrors(detected)

	recs := make(
		[]recommendationOutput, 0,
		len(analysis.Recommendations),
	)
	for _, r := range analysis.Recommendations {
		recs = append(recs, recommendationOutput{
			Type:     r.Type,
			Priority: r.Priority,
			Message:  r.Description,
			Impact:   r.Impact,
		})
	}

	out := errorAnalysisOutput{
		TotalErrors:     analysis.TotalErrors,
		Categories:      analysis.ErrorCategories,
		Severity:        analysis.SeverityLevels,
		Recommendations: recs,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf(
			"failed to marshal JSON: %w", err,
		)
	}

	outputPath, _ := cmd.Flags().GetString("output")
	if outputPath != "" {
		if err := os.WriteFile(
			outputPath, data, 0600,
		); err != nil {
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

// readInput reads text from a file path or stdin when
// path is "-".
func readInput(path string) (string, error) {
	if path == "-" {
		return readAll(os.Stdin)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"input file does not exist: %s", path,
		)
	}

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf(
			"failed to open input file: %w", err,
		)
	}
	defer f.Close()

	return readAll(f)
}

// readAll reads the full contents of a reader.
func readAll(r io.Reader) (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf(
			"failed to read input: %w", err,
		)
	}
	return sb.String(), nil
}

func init() {
	errorsAnalyzeCmd.Flags().String(
		"input", "",
		"input file path or \"-\" for stdin",
	)
	errorsAnalyzeCmd.Flags().String(
		"output", "",
		"path to write JSON output (stdout if omitted)",
	)

	errorsCmd.AddCommand(errorsAnalyzeCmd)
	rootCmd.AddCommand(errorsCmd)
}
