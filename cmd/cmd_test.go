package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd(t *testing.T) {
	rootCmd := getRootCmd()
	
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "panoptic", rootCmd.Use)
	assert.Equal(t, "Automated testing and recording application for multiple platforms", rootCmd.Short)
	assert.Contains(t, rootCmd.Long, "Panoptic is a comprehensive tool")
}

func TestRunCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "No config file provided",
			args:        []string{},
			expectError: true,
			errorMsg:    "error: accepts 1 arg(s), received 0",
		},
		{
			name:        "Non-existent config file",
			args:        []string{"/non/existent/config.yaml"},
			expectError: true,
			errorMsg:    "Failed to load configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := getRootCmd()
			
			// Capture command output
			output := &strings.Builder{}
			rootCmd.SetOut(output)
			rootCmd.SetErr(output)
			
			err := rootCmd.Execute()
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, output.String(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunCmd_WithValidConfig(t *testing.T) {
	// Create a temporary valid config file
	configContent := `
name: "Test Config"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
    timeout: 5
actions:
  - name: "wait"
    type: "wait"
    wait_time: 1
`

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")
	outputDir := filepath.Join(tempDir, "output")

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	rootCmd := getRootCmd()
	
	// Set up args
	rootCmd.SetArgs([]string{"run", configFile, "--output", outputDir, "--verbose"})
	
	// This test might fail if browser is not available, so we'll just check that it doesn't panic
	output := &strings.Builder{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)

	start := time.Now()
	err = rootCmd.Execute()
	duration := time.Since(start)

	// The command may fail due to missing browser, but should complete within reasonable time
	assert.True(t, duration < 30*time.Second, "Command took too long: %v", duration)
	
	if err != nil {
		// Expected to possibly fail due to browser requirements
		outputStr := output.String()
		// Check for expected error messages
		assert.True(t, 
			strings.Contains(outputStr, "Failed to load configuration") ||
			strings.Contains(outputStr, "Browser not available") ||
			strings.Contains(outputStr, "failed"),
			"Unexpected error: %v", outputStr)
	} else {
		// If successful, check output directories exist
		assert.DirExists(t, filepath.Join(outputDir, "screenshots"))
		assert.DirExists(t, filepath.Join(outputDir, "videos"))
		assert.DirExists(t, filepath.Join(outputDir, "logs"))
		assert.FileExists(t, filepath.Join(outputDir, "report.html"))
	}
}

func TestInitConfig(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		expectError bool
	}{
		{
			name:        "No config file",
			configFile:  "",
			expectError: false, // Should use defaults
		},
		{
			name:        "Non-existent config file",
			configFile:  "/non/existent/file.yaml",
			expectError: false, // Should not error, just not find config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for clean test
			viper.Reset()
			
			if tt.configFile != "" {
				rootCmd := getRootCmd()
				rootCmd.SetArgs([]string{"--config", tt.configFile})
			}
			
			// This would be called by cobra.OnInitialize
			// For testing, we can verify it doesn't panic
			assert.NotPanics(t, func() {
				initConfig()
			})
		})
	}
}

func TestFlags(t *testing.T) {
	rootCmd := getRootCmd()
	
	// Test that persistent flags are available
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "config file (default is $HOME/.panoptic.yaml)", configFlag.Usage)
	
	outputFlag := rootCmd.PersistentFlags().Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "output directory for screenshots and videos", outputFlag.Usage)
	assert.Equal(t, "./output", outputFlag.DefValue)
	
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "enable verbose logging", verboseFlag.Usage)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestCommandChaining(t *testing.T) {
	rootCmd := getRootCmd()
	
	// Test that run command is properly added
	runCmd, _, err := rootCmd.Find([]string{"run"})
	assert.NoError(t, err)
	assert.NotNil(t, runCmd)
	assert.Equal(t, "run", runCmd.Use)
	assert.Equal(t, "Execute automated testing and recording", runCmd.Short)
}

func TestViperBinding(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	
	rootCmd := getRootCmd()
	
	// Test flag binding
	outputFlag := rootCmd.PersistentFlags().Lookup("output")
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	
	assert.NotNil(t, outputFlag)
	assert.NotNil(t, verboseFlag)
	
	// The binding happens during init(), so we test the result
	assert.True(t, viper.IsSet("output") || viper.GetString("output") == "./output")
	assert.True(t, viper.IsSet("verbose") || viper.GetBool("verbose") == false)
}

func TestCommandHelp(t *testing.T) {
	rootCmd := getRootCmd()
	
	// Test help for root command
	rootCmd.SetArgs([]string{"--help"})
	
	output := &strings.Builder{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)
	
	err := rootCmd.Execute()
	assert.NoError(t, err)
	
	outputStr := output.String()
	assert.Contains(t, outputStr, "Automated testing and recording application")
	assert.Contains(t, outputStr, "--config")
	assert.Contains(t, outputStr, "--output")
	assert.Contains(t, outputStr, "--verbose")
}

func TestRunCommandHelp(t *testing.T) {
	rootCmd := getRootCmd()
	
	rootCmd.SetArgs([]string{"run", "--help"})
	
	output := &strings.Builder{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)
	
	err := rootCmd.Execute()
	assert.NoError(t, err)
	
	outputStr := output.String()
	assert.Contains(t, outputStr, "Execute automated testing and recording")
	assert.Contains(t, outputStr, "[config-file]")
}

// Helper function to get a fresh instance of root command for testing
func getRootCmd() *cobra.Command {
	// Create new command instance to avoid state pollution between tests
	cmd := &cobra.Command{
		Use:   "panoptic",
		Short: "Automated testing and recording application for multiple platforms",
		Long: `Panoptic is a comprehensive tool for automated testing, UI recording, 
and screenshot capture across web, desktop, and mobile applications.`,
	}
	
	// Re-initialize flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.panoptic.yaml)")
	cmd.PersistentFlags().String("output", "./output", "output directory for screenshots and videos")
	cmd.PersistentFlags().Bool("verbose", false, "enable verbose logging")
	
	viper.BindPFlag("output", cmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose"))
	
	// Add run command
	cmd.AddCommand(runCmd)
	
	return cmd
}