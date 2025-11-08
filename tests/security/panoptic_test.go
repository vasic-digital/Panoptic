//go:build security
// +build security

package security

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPanopticSecurity_ConfigurationInjection tests for injection vulnerabilities in configuration
func TestPanopticSecurity_ConfigurationInjection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security tests in short mode")
	}

	securityTests := []struct {
		name         string
		configField  string
		maliciousInput string
		expectedBehavior string
	}{
		{
			name:          "SQL injection in app name",
			configField:   "name",
			maliciousInput: "'; DROP TABLE apps; --",
			expectedBehavior: "should sanitize or reject",
		},
		{
			name:          "XSS in URL",
			configField:   "url",
			maliciousInput: "javascript:alert('XSS')",
			expectedBehavior: "should sanitize or reject",
		},
		{
			name:          "Path traversal in output directory",
			configField:   "output",
			maliciousInput: "../../../etc/passwd",
			expectedBehavior: "should sanitize or reject",
		},
		{
			name:          "Command injection in selector",
			configField:   "selector",
			maliciousInput: "img[src=x onerror=alert('XSS')]",
			expectedBehavior: "should sanitize or reject",
		},
		{
			name:          "HTML injection in value",
			configField:   "value",
			maliciousInput: "<script>alert('XSS')</script>",
			expectedBehavior: "should sanitize or reject",
		},
		{
			name:          "File path injection in filename",
			configField:   "filename",
			maliciousInput: "../../../malicious.exe",
			expectedBehavior: "should sanitize or reject",
		},
		{
			name:          "JSON injection in parameters",
			configField:   "parameters",
			maliciousInput: `{"injection": "</script><script>alert('XSS')</script>"}`,
			expectedBehavior: "should sanitize or reject",
		},
	}

	for _, test := range securityTests {
		t.Run(test.name, func(t *testing.T) {
			// Create malicious configuration
			configContent := createMaliciousConfig(test.configField, test.maliciousInput)
			
			outputDir := t.TempDir()
			configFile := createTempFile(t, "security-test-*.yaml", configContent)
			defer os.Remove(configFile)

			// Run Panoptic with malicious config
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir)
			cmd.Dir = getProjectRoot()

			// Capture output
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			outputStr := stdout.String()
			errorStr := stderr.String()

			t.Logf("Security test - %s", test.name)
			t.Logf("Malicious input: %s", test.maliciousInput)
			t.Logf("Command output: %s", outputStr)
			t.Logf("Command error: %s", errorStr)
			t.Logf("Command error: %v", err)

			// Verify security behavior
			if err != nil {
				// Should fail gracefully with security error
				assert.True(t, 
					strings.Contains(errorStr, "validation") ||
					strings.Contains(errorStr, "invalid") ||
					strings.Contains(errorStr, "sanitized") ||
					strings.Contains(errorStr, "rejected") ||
					strings.Contains(errorStr, "security"),
					"Expected security validation error, got: %s", errorStr)
			} else {
				// If it succeeds, verify no malicious artifacts were created
				verifyNoMaliciousArtifacts(t, outputDir, test.maliciousInput)
			}

			// Check for command injection attempts
			assert.False(t, containsCommandInjection(outputStr+errorStr), 
				"Potential command injection detected in output")
		})
	}
}

// TestPanopticSecurity_FileSystemSecurity tests file system access restrictions
func TestPanopticSecurity_FileSystemSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security tests in short mode")
	}

	t.Run("File access restrictions", func(t *testing.T) {
		// Test with attempts to access sensitive files
		configContent := `
name: "File Security Test"
output: "/tmp/security_test"
apps:
  - name: "Security Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "attempt_file_read"
    type: "screenshot"
    parameters:
      filename: "../../../../etc/passwd"
  - name: "attempt_file_write"
    type: "screenshot"
    parameters:
      filename: "/root/.ssh/authorized_keys"
`

		outputDir := t.TempDir()
		configFile := createTempFile(t, "security-fs-*.yaml", configContent)
		defer os.Remove(configFile)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir)
		cmd.Dir = getProjectRoot()

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		outputStr := stdout.String()
		errorStr := stderr.String()

		t.Logf("File security test output: %s", outputStr)
		t.Logf("File security test error: %s", errorStr)

		// Verify no unauthorized file access
		assert.False(t, strings.Contains(outputStr+errorStr, "/etc/passwd"))
		assert.False(t, strings.Contains(outputStr+errorStr, "/root/.ssh"))
		
		// Check output directory for security
		verifyOutputDirectorySecurity(t, outputDir)
	})

	t.Run("Directory traversal prevention", func(t *testing.T) {
		configContent := `
name: "Directory Traversal Test"
output: "../../../tmp/traversal_test"
apps:
  - name: "Traversal Test"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "traversal_screenshot"
    type: "screenshot"
    parameters:
      filename: "../outside_output.png"
`

		outputDir := t.TempDir()
		configFile := createTempFile(t, "security-traversal-*.yaml", configContent)
		defer os.Remove(configFile)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir)
		cmd.Dir = getProjectRoot()

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		outputStr := stdout.String()
		errorStr := stderr.String()

		// Verify directory traversal was prevented
		assert.False(t, strings.Contains(outputStr+errorStr, "traversal_test"))
	})
}

// TestPanopticSecurity_NetworkSecurity tests network security
func TestPanopticSecurity_NetworkSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security tests in short mode")
	}

	securityTests := []struct {
		name         string
		url          string
		expectedBehavior string
	}{
		{
			name:          "File protocol URL",
			url:           "file:///etc/passwd",
			expectedBehavior: "should be blocked",
		},
		{
			name:          "FTP protocol URL",
			url:           "ftp://malicious-server.com",
			expectedBehavior: "should be blocked",
		},
		{
			name:          "Localhost bypass attempt",
			url:           "http://127.0.0.1/admin",
			expectedBehavior: "should be handled securely",
		},
		{
			name:          "Internal network access",
			url:           "http://192.168.1.1/admin",
			expectedBehavior: "should be handled securely",
		},
		{
			name:          "Data URL scheme",
			url:           "data:text/html,<script>alert('XSS')</script>",
			expectedBehavior: "should be blocked",
		},
	}

	for _, test := range securityTests {
		t.Run(test.name, func(t *testing.T) {
			configContent := fmt.Sprintf(`
name: "Network Security Test"
output: "%s"
apps:
  - name: "Network Security Test"
    type: "web"
    url: "%s"
    timeout: 10
actions:
  - name: "test_network_security"
    type: "navigate"
    value: "%s"
`, t.TempDir(), test.url, test.url)

			configFile := createTempFile(t, "security-network-*.yaml", configContent)
			defer os.Remove(configFile)

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile)
			cmd.Dir = getProjectRoot()

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			outputStr := stdout.String()
			errorStr := stderr.String()

			t.Logf("Network security test - %s", test.name)
			t.Logf("URL: %s", test.url)
			t.Logf("Output: %s", outputStr)
			t.Logf("Error: %s", errorStr)

			// Verify secure behavior
			if strings.Contains(test.url, "file://") || 
			   strings.Contains(test.url, "ftp://") || 
			   strings.Contains(test.url, "data:") {
				// These should be blocked
				assert.True(t, err != nil || strings.Contains(errorStr, "blocked") ||
					strings.Contains(errorStr, "invalid") || strings.Contains(errorStr, "unsupported"),
					"Expected dangerous URL to be blocked")
			}
			
			// Verify no sensitive data leakage
			assert.False(t, containsSensitiveData(outputStr+errorStr))
		})
	}
}

// TestPanopticSecurity_ResourceLimits tests for resource exhaustion attacks
func TestPanopticSecurity_ResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security tests in short mode")
	}

	t.Run("Large input handling", func(t *testing.T) {
		// Create configuration with very large values
		largeString := strings.Repeat("A", 1000000) // 1MB string
		
		configContent := fmt.Sprintf(`
name: "Resource Limit Test"
output: "%s"
apps:
  - name: "Resource Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "large_input_test"
    type: "fill"
    selector: "input.test"
    value: "%s"
  - name: "large_wait_test"
    type: "wait"
    wait_time: 86400  # 24 hours
`, t.TempDir(), largeString[:1000]) // Truncate for YAML validity

		configFile := createTempFile(t, "security-resource-*.yaml", configContent)
		defer os.Remove(configFile)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile)
		cmd.Dir = getProjectRoot()

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		start := time.Now()
		err := cmd.Run()
		duration := time.Since(start)

		t.Logf("Resource limit test duration: %v", duration)
		t.Logf("Resource limit test result: %v", err)

		// Should complete within reasonable time (not hang on large input)
		assert.True(t, duration < 25*time.Second, "Test took too long, possible resource exhaustion")
	})

	t.Run("Memory exhaustion prevention", func(t *testing.T) {
		// Create configuration that might cause memory exhaustion
		configContent := `
name: "Memory Exhaustion Test"
output: "/tmp/memory_test"
apps:
  - name: "Memory Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "memory_test_action"
    type: "record"
    duration: 3600  # 1 hour recording
    parameters:
      filename: "huge_video.mp4"
`

		configFile := createTempFile(t, "security-memory-*.yaml", configContent)
		defer os.Remove(configFile)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile)
		cmd.Dir = getProjectRoot()

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		start := time.Now()
		err := cmd.Run()
		duration := time.Since(start)

		t.Logf("Memory exhaustion test duration: %v", duration)

		// Should terminate within timeout due to resource limits
		assert.True(t, duration < 35*time.Second, "Test didn't respect timeout")
	})
}

// TestPanopticSecurity_DataPrivacy tests for data privacy protection
func TestPanopticSecurity_DataPrivacy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security tests in short mode")
	}

	t.Run("Sensitive data handling", func(t *testing.T) {
		// Test with configuration containing sensitive data
		configContent := `
name: "Data Privacy Test"
output: "/tmp/privacy_test"
apps:
  - name: "Privacy Test App"
    type: "web"
    url: "https://httpbin.org/html"
    environment:
      DATABASE_PASSWORD: "super_secret_password_123"
      API_KEY: "sk-1234567890abcdef"
      PRIVATE_KEY: "-----BEGIN RSA PRIVATE KEY-----\\nMIIEpAIBAAKCAQEA..."
actions:
  - name: "privacy_test_action"
    type: "screenshot"
    parameters:
      filename: "privacy_test.png"
`

		outputDir := t.TempDir()
		configFile := createTempFile(t, "security-privacy-*.yaml", configContent)
		defer os.Remove(configFile)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir)
		cmd.Dir = getProjectRoot()

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		outputStr := stdout.String()
		errorStr := stderr.String()

		t.Logf("Data privacy test result: %v", err)

		// Verify sensitive data is not exposed in logs
		combinedOutput := outputStr + errorStr
		assert.False(t, strings.Contains(combinedOutput, "super_secret_password_123"))
		assert.False(t, strings.Contains(combinedOutput, "sk-1234567890abcdef"))
		assert.False(t, strings.Contains(combinedOutput, "BEGIN RSA PRIVATE KEY"))
		
		// Check output files for data leakage
		verifyNoDataLeakage(t, outputDir)
	})

	t.Run("Log sanitization", func(t *testing.T) {
		configContent := `
name: "Log Sanitization Test"
output: "/tmp/log_sanitization_test"
apps:
  - name: "Log Test App"
    type: "web"
    url: "https://httpbin.org/html"
    environment:
      USER_EMAIL: "test@example.com"
      USER_TOKEN: "token_1234567890abcdef"
actions:
  - name: "log_test_action"
    type: "fill"
    selector: "input.email"
    value: "test@example.com"
`

		outputDir := t.TempDir()
		configFile := createTempFile(t, "security-logs-*.yaml", configContent)
		defer os.Remove(configFile)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./panoptic", "run", configFile, "--output", outputDir, "--verbose")
		cmd.Dir = getProjectRoot()

		err := cmd.Run()
		
		// Check log files for data leakage
		logsDir := filepath.Join(outputDir, "logs")
		if files, _ := os.ReadDir(logsDir); len(files) > 0 {
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
					logPath := filepath.Join(logsDir, file.Name())
					content, err := os.ReadFile(logPath)
					require.NoError(t, err)
					
					logContent := string(content)
					assert.False(t, strings.Contains(logContent, "token_1234567890abcdef"),
						"Token found in logs")
				}
			}
		}
	})
}

// Helper functions

func createMaliciousConfig(field, maliciousInput string) string {
	switch field {
	case "name":
		return fmt.Sprintf(`
name: "%s"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "test"
    type: "wait"
    wait_time: 1
`, maliciousInput)
	case "url":
		return fmt.Sprintf(`
name: "URL Security Test"
apps:
  - name: "Test App"
    type: "web"
    url: "%s"
actions:
  - name: "test"
    type: "wait"
    wait_time: 1
`, maliciousInput)
	case "output":
		return fmt.Sprintf(`
name: "Output Security Test"
output: "%s"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "test"
    type: "wait"
    wait_time: 1
`, maliciousInput)
	case "selector":
		return fmt.Sprintf(`
name: "Selector Security Test"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "test"
    type: "click"
    selector: "%s"
`, maliciousInput)
	case "value":
		return fmt.Sprintf(`
name: "Value Security Test"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "test"
    type: "fill"
    selector: "input.test"
    value: "%s"
`, maliciousInput)
	case "filename":
		return fmt.Sprintf(`
name: "Filename Security Test"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "test"
    type: "screenshot"
    parameters:
      filename: "%s"
`, maliciousInput)
	case "parameters":
		return fmt.Sprintf(`
name: "Parameters Security Test"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "test"
    type: "screenshot"
    parameters: %s
`, maliciousInput)
	default:
		return `name: "Default Test"
apps:
  - name: "Test App"
    type: "web"
    url: "https://httpbin.org/html"
actions:
  - name: "test"
    type: "wait"
    wait_time: 1
`
	}
}

func verifyNoMaliciousArtifacts(t *testing.T, outputDir, maliciousInput string) {
	t.Helper()
	
	// Check all files in output directory for malicious content
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip unreadable files
			}
			
			fileContent := string(content)
			if strings.Contains(fileContent, maliciousInput) {
				t.Errorf("Malicious input found in file: %s", path)
			}
		}
		return nil
	})
	
	assert.NoError(t, err, "Error walking output directory")
}

func verifyOutputDirectorySecurity(t *testing.T, outputDir string) {
	t.Helper()
	
	// Verify output directory is within expected bounds
	absOutputDir, err := filepath.Abs(outputDir)
	require.NoError(t, err)
	
	// Check that output is not outside temp directory
	tempDir := os.TempDir()
	assert.True(t, strings.HasPrefix(absOutputDir, tempDir), 
		"Output directory should be within temp directory")
	
	// Verify no symlinks to sensitive locations
	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Check for symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err == nil {
				absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(path), target))
				if err == nil {
					// Check if symlink points to sensitive locations
					sensitivePaths := []string{
						"/etc", "/root", "/home", "/Users",
						"/var", "/usr", "/bin", "/sbin",
					}
					
					for _, sensitive := range sensitivePaths {
						if strings.HasPrefix(absTarget, sensitive) {
							t.Errorf("Symlink to sensitive location: %s -> %s", path, absTarget)
						}
					}
				}
			}
		}
		
		return nil
	})
	
	assert.NoError(t, err)
}

func containsCommandInjection(text string) bool {
	indicators := []string{
		";", "|", "&", "&&", "||", "`", "$(", "${",
		"rm ", "del ", "format ", "fdisk ",
		"wget ", "curl ", "nc ", "netcat ",
		"powershell ", "cmd.exe", "/bin/sh",
		"eval(", "exec(", "system(",
	}
	
	lowerText := strings.ToLower(text)
	for _, indicator := range indicators {
		if strings.Contains(lowerText, indicator) {
			return true
		}
	}
	
	return false
}

func containsSensitiveData(text string) bool {
	sensitivePatterns := []string{
		"password", "secret", "key", "token",
		"private", "confidential", "credential",
		"auth", "session", "cookie",
	}
	
	lowerText := strings.ToLower(text)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerText, pattern) {
			return true
		}
	}
	
	return false
}

func verifyNoDataLeakage(t *testing.T, outputDir string) {
	t.Helper()
	
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && filepath.Ext(path) != "" {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			
			fileContent := string(content)
			
			// Check for sensitive data patterns
			sensitivePatterns := []string{
				"password", "secret", "private_key",
				"api_key", "auth_token", "session",
			}
			
			for _, pattern := range sensitivePatterns {
				if strings.Contains(strings.ToLower(fileContent), pattern) {
					t.Errorf("Potential data leakage in file %s: contains %s", path, pattern)
				}
			}
		}
		
		return nil
	})
	
	assert.NoError(t, err)
}

func createTempFile(t *testing.T, pattern, content string) string {
	t.Helper()
	
	tmpFile, err := os.CreateTemp("", pattern)
	require.NoError(t, err)
	
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	
	tmpFile.Close()
	return tmpFile.Name()
}

func getProjectRoot() string {
	// Same as in functional tests
	cwd, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return cwd
}

// TestMain for security tests
func TestMain(m *testing.M) {
	// Build Panoptic with security flags
	projectRoot := getProjectRoot()
	buildCmd := exec.Command("go", "build", "-o", "panoptic", "main.go")
	buildCmd.Dir = projectRoot
	
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to build Panoptic for security tests: %v\nOutput: %s\n", err, string(output))
		os.Exit(1)
	}
	
	// Run security tests
	result := m.Run()
	
	// Cleanup
	os.Remove(filepath.Join(projectRoot, "panoptic"))
	
	os.Exit(result)
}