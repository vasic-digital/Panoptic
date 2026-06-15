package enterprise

import (
	"fmt"
	"os"
	"testing"
)

// TestMain isolates the enterprise unit tests from the source tree.
//
// HXC-090: EnterpriseManager.loadJSON/saveJSON resolve filenames relative to
// em.StoragePath via filepath.Join(em.StoragePath, filename). Many test
// managers are constructed as struct literals that leave StoragePath unset
// (""), so a save resolves to a bare relative path (e.g. "users.json"), which
// is written into the test process's current working directory — the package
// source directory — overwriting the committed seed JSON files
// (api_keys/audit/projects/roles/subscriptions/teams/users.json) on every run.
// That mutates version-controlled files (a CONST-053 violation) and leaves the
// working tree perpetually dirty.
//
// Fix (test-only, lowest-risk): run the whole package's tests from a throwaway
// temp directory. Empty-StoragePath managers then write their relative-path
// JSON into the temp CWD, never into the source tree. The committed seed JSONs
// remain untouched (production still loads them via Initialize with a real
// StoragePath). No production code and no assertion is changed. Tests that need
// real storage already point StoragePath at their own t.TempDir()/"/tmp", and
// all file paths in this package's tests are absolute (temp-dir based), so the
// chdir is safe.
func TestMain(m *testing.M) {
	origWD, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: cannot determine working directory: %v\n", err)
		os.Exit(1)
	}

	tmpDir, err := os.MkdirTemp("", "enterprise_test_cwd_")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: cannot create temp dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.Chdir(tmpDir); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: cannot chdir to temp dir: %v\n", err)
		_ = os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	code := m.Run()

	// Restore CWD before cleanup so the temp dir is not the process CWD when removed.
	_ = os.Chdir(origWD)
	_ = os.RemoveAll(tmpDir)

	os.Exit(code)
}
