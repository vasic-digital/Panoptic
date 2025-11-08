package cmd

import (
	"os"
	"path/filepath"

	"panoptic/internal/config"
	"panoptic/internal/executor"
	"panoptic/internal/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run [config-file]",
	Short: "Execute automated testing and recording",
	Long: `Run the automated testing and recording process based on the provided configuration.
The configuration file should define the applications to test and the actions to perform.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := args[0]
		
		// Initialize logger
		log := logger.NewLogger(viper.GetBool("verbose"))
		log.Info("Starting Panoptic execution")
		
		// Load configuration
		cfg, err := config.Load(configFile)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		
		// Set output directory
		outputDir := viper.GetString("output")
		if cfg.Output != "" {
			outputDir = cfg.Output
		}
		
		// Ensure output directory exists
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
		
		// Create subdirectories
		screenshotsDir := filepath.Join(outputDir, "screenshots")
		videosDir := filepath.Join(outputDir, "videos")
		logsDir := filepath.Join(outputDir, "logs")
		
		for _, dir := range []string{screenshotsDir, videosDir, logsDir} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create subdirectory %s: %v", dir, err)
			}
		}
		
		log.Infof("Output directory: %s", outputDir)
		
		// Execute the configuration
		exec := executor.NewExecutor(cfg, outputDir, log)
		if err := exec.Run(); err != nil {
			log.Fatalf("Execution failed: %v", err)
		}
		
		// Generate report
		reportPath := filepath.Join(outputDir, "report.html")
		if err := exec.GenerateReport(reportPath); err != nil {
			log.Errorf("Failed to generate report: %v", err)
		} else {
			log.Infof("Report generated: %s", reportPath)
		}
		
		log.Info("Execution completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}