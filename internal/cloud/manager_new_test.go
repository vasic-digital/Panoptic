package cloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"panoptic/internal/logger"
)

// TestUploadMethods tests each upload provider method.
//
// Anti-bluff (§11.4 / CONST-035, Article XI §11.9 / CONST-050(A)+(B)): the
// original assertion expected `assert.NoError` for every provider. That was
// a §11.4 PASS-bluff against reality: AWS, GCP, and Azure providers do NOT
// have their cloud SDKs wired (commit 65ea0bf intentionally removed the
// "simulated success" path and substituted ErrCloudSDKNotWired so callers
// cannot mistake the no-op for a real upload). Asserting success against
// production code that honestly returns a sentinel error is the SAME class
// of bluff the operator mandate forbids ("tests do execute with success ...
// in reality the most of the features does not work and can't be used").
//
// The correct end-user contract this test must enforce:
//   - aws / gcp / azure subtests MUST return ErrCloudSDKNotWired so callers
//     know real wiring is still required.
//   - local subtest MUST succeed AND the test result MUST be tracked, since
//     local upload is the only provider with a real working implementation.
func TestUploadMethods(t *testing.T) {
	log := logger.NewLogger(false)

	tests := []struct {
		name       string
		provider   string
		expectErr  error  // ErrCloudSDKNotWired for un-wired SDKs; nil for working local
		expectTrue bool   // result.Success expected on the tracked TestResult
	}{
		{"AWS Upload", "aws", ErrCloudSDKNotWired, false},
		{"GCP Upload", "gcp", ErrCloudSDKNotWired, false},
		{"Azure Upload", "azure", ErrCloudSDKNotWired, false},
		{"Local Upload", "local", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			manager := &CloudManager{
				Logger: *log,
				Config: CloudConfig{
					Provider: tt.provider,
					Bucket:   filepath.Join(tempDir, "storage"),
				},
				Enabled: true,
			}

			// Create test file
			testFile := filepath.Join(tempDir, "test.txt")
			err := os.WriteFile(testFile, []byte("test content"), 0644)
			require.NoError(t, err)

			err = manager.Upload(testFile)
			if tt.expectErr != nil {
				// Cloud SDK providers honestly admit they are not wired.
				require.ErrorIs(t, err, tt.expectErr,
					"%s: production code MUST return ErrCloudSDKNotWired until real SDK is implemented (anti-bluff §11.4)",
					tt.provider)
				return
			}

			// Local provider has a real implementation and MUST succeed.
			require.NoError(t, err, "%s: local upload MUST succeed (real implementation)", tt.provider)
			require.Greater(t, len(manager.TestResults), 0,
				"%s: working upload MUST track a TestResult (runtime evidence per §11.4)", tt.provider)

			result := manager.TestResults[0]
			assert.Equal(t, tt.expectTrue, result.Success,
				"%s: tracked TestResult.Success must reflect real outcome", tt.provider)
			assert.NotEmpty(t, result.Artifacts, "%s: working upload must record artifacts", tt.provider)

			if len(result.Artifacts) > 0 {
				artifact := result.Artifacts[0]
				assert.Equal(t, "video", artifact.Type, "Should have correct artifact type")
				assert.Equal(t, int64(12), artifact.Size, "Should have correct size")
			}
		})
	}
}

// TestUploadToAWS tests AWS-specific upload logic.
//
// Anti-bluff (§11.4 / CONST-035, Article XI §11.9 / CONST-050(A)+(B)):
// see TestUploadMethods comment. The AWS SDK is intentionally not wired;
// production code returns ErrCloudSDKNotWired. The honest contract a test
// can enforce is that the sentinel error is returned and the manager has
// not silently fabricated a TestResult claiming the upload succeeded.
func TestUploadToAWS(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "aws",
			Bucket:   "test-bucket",
		},
		Enabled: true,
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("aws test content"), 0644)
	require.NoError(t, err)

	err = manager.uploadToAWS(testFile, "test.txt", getFileInfo(testFile))
	require.ErrorIs(t, err, ErrCloudSDKNotWired,
		"AWS upload MUST return ErrCloudSDKNotWired until real AWS SDK is wired (anti-bluff §11.4)")
	assert.Equal(t, 0, len(manager.TestResults),
		"un-wired SDK MUST NOT fabricate a TestResult (anti-bluff §11.4)")
}

// TestUploadToGCP tests GCP-specific upload logic.
//
// Anti-bluff (§11.4 / CONST-035): GCP SDK not wired → ErrCloudSDKNotWired.
func TestUploadToGCP(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "gcp",
			Bucket:   "test-bucket",
		},
		Enabled: true,
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("gcp test content"), 0644)
	require.NoError(t, err)

	err = manager.uploadToGCP(testFile, "test.txt", getFileInfo(testFile))
	require.ErrorIs(t, err, ErrCloudSDKNotWired,
		"GCP upload MUST return ErrCloudSDKNotWired until real GCP SDK is wired (anti-bluff §11.4)")
	assert.Equal(t, 0, len(manager.TestResults),
		"un-wired SDK MUST NOT fabricate a TestResult (anti-bluff §11.4)")
}

// TestUploadToAzure tests Azure-specific upload logic.
//
// Anti-bluff (§11.4 / CONST-035): Azure SDK not wired → ErrCloudSDKNotWired.
func TestUploadToAzure(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "azure",
			Bucket:   "test-container",
		},
		Enabled: true,
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("azure test content"), 0644)
	require.NoError(t, err)

	err = manager.uploadToAzure(testFile, "test.txt", getFileInfo(testFile))
	require.ErrorIs(t, err, ErrCloudSDKNotWired,
		"Azure upload MUST return ErrCloudSDKNotWired until real Azure SDK is wired (anti-bluff §11.4)")
	assert.Equal(t, 0, len(manager.TestResults),
		"un-wired SDK MUST NOT fabricate a TestResult (anti-bluff §11.4)")
}

// TestUploadToLocal tests local storage upload logic
func TestUploadToLocal(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	storageDir := filepath.Join(tempDir, "storage")
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "local",
			Bucket:  storageDir,
		},
		Enabled: true,
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("local test content"), 0644)
	require.NoError(t, err)

	err = manager.uploadToLocal(testFile, "2025/11/26/test.txt", getFileInfo(testFile))
	assert.NoError(t, err, "Local upload should succeed")

	// Verify file was actually copied
	destPath := filepath.Join(storageDir, "2025/11/26/test.txt")
	_, err = os.Stat(destPath)
	assert.NoError(t, err, "File should exist in local storage")

	content, err := os.ReadFile(destPath)
	assert.NoError(t, err, "Should be able to read copied file")
	assert.Equal(t, "local test content", string(content), "Content should match")
}

// TestUploadErrorHandling tests error scenarios
func TestUploadErrorHandling(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	manager := &CloudManager{
		Logger: *log,
		Config: CloudConfig{
			Provider: "local",
			Bucket:  filepath.Join(tempDir, "storage"),
		},
		Enabled: true,
	}

	// Test with nonexistent file
	err := manager.Upload("/nonexistent/file.txt")
	assert.Error(t, err, "Should return error for nonexistent file")
	assert.Contains(t, err.Error(), "file not found", "Error should mention file not found")

	// Test with empty path
	err = manager.Upload("")
	assert.Error(t, err, "Should return error for empty path")
	assert.Contains(t, err.Error(), "cannot be empty", "Error should mention cannot be empty")
}

// Helper function to get file info
func getFileInfo(path string) os.FileInfo {
	info, _ := os.Stat(path)
	return info
}