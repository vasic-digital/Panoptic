package recvalidate

import "os"

// osTempDir returns the host temp directory (indirection kept tiny so tests
// could override the temp location if ever needed).
func osTempDir() string { return os.TempDir() }
