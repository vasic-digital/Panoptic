package launcher

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLauncher tests creating a new launcher
func TestNewLauncher(t *testing.T) {
	iconDir := "/test/icons"
	launcher := NewLauncher(iconDir)

	assert.NotNil(t, launcher, "Launcher should not be nil")
	assert.Equal(t, iconDir, launcher.iconDir, "Icon directory should be set correctly")
	assert.Equal(t, detectPlatform(), launcher.platform, "Platform should be detected correctly")
}

// TestDetectPlatform tests platform detection
func TestDetectPlatform(t *testing.T) {
	platform := detectPlatform()
	expectedPlatform := runtime.GOOS
	
	switch expectedPlatform {
	case "windows":
		assert.Equal(t, "windows", platform, "Should detect Windows platform")
	case "darwin":
		assert.Equal(t, "macos", platform, "Should detect macOS platform")
	case "linux":
		assert.Equal(t, "linux", platform, "Should detect Linux platform")
	case "android":
		assert.Equal(t, "android", platform, "Should detect Android platform")
	case "ios":
		assert.Equal(t, "ios", platform, "Should detect iOS platform")
	default:
		assert.Equal(t, "unknown", platform, "Should detect unknown platform for unsupported OS")
	}
}

// TestSetIcon tests setting launcher icon
func TestSetIcon(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	// Test setting icon with absolute path
	iconPath := filepath.Join(tempDir, "test.png")
	err := os.WriteFile(iconPath, []byte("fake png data"), 0644)
	require.NoError(t, err, "Should create test icon file")

	err = launcher.SetIcon(iconPath)
	assert.NoError(t, err, "Should set icon successfully")
	assert.Equal(t, iconPath, launcher.GetIconPath(), "Should store correct icon path")

	// Test setting icon with relative path
	relativeIcon := "relative.png"
	relativePath := filepath.Join(tempDir, relativeIcon)
	err = os.WriteFile(relativePath, []byte("fake png data"), 0644)
	require.NoError(t, err, "Should create relative test icon file")

	err = launcher.SetIcon(relativeIcon)
	assert.NoError(t, err, "Should set relative icon successfully")
	assert.Equal(t, relativePath, launcher.GetIconPath(), "Should store full path for relative icon")
}

// TestSetIcon_NonExistentFile tests setting non-existent icon
func TestSetIcon_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	nonExistentPath := filepath.Join(tempDir, "nonexistent.png")
	err := launcher.SetIcon(nonExistentPath)
	assert.Error(t, err, "Should return error for non-existent file")
	assert.Contains(t, err.Error(), "icon file not found", "Error should mention file not found")
}

// TestGetPlatformIcon tests getting platform-specific icon path
func TestGetPlatformIcon(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	platformIconPath := launcher.GetPlatformIcon()
	
	// Should return a path based on the platform
	assert.NotEmpty(t, platformIconPath, "Should return non-empty platform icon path")
	assert.Contains(t, platformIconPath, tempDir, "Should include icon directory in path")
	
	// Check path structure based on platform
	switch launcher.platform {
	case "windows", "linux":
		assert.Contains(t, platformIconPath, filepath.Join("desktop", "icon.png"), "Should use desktop icon for Windows/Linux")
	case "macos":
		assert.Contains(t, platformIconPath, filepath.Join("desktop", "large.png"), "Should use large icon for macOS")
	case "android":
		assert.Contains(t, platformIconPath, filepath.Join("android", "xxxhdpi", "icon_xxxhdpi.png"), "Should use Android icon path")
	case "ios":
		assert.Contains(t, platformIconPath, filepath.Join("ios", "appstore", "icon_appstore.png"), "Should use iOS appstore icon")
	default:
		assert.Contains(t, platformIconPath, filepath.Join("desktop", "icon.png"), "Should use default desktop icon for unknown platform")
	}
}

// TestDisplayIcon tests displaying launcher icon
func TestDisplayIcon(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	// Test with no icon set (should use platform icon)
	err := launcher.DisplayIcon()
	assert.NoError(t, err, "Should display platform icon successfully")

	// Test with custom icon set
	iconPath := filepath.Join(tempDir, "custom.png")
	err = os.WriteFile(iconPath, []byte("fake png data"), 0644)
	require.NoError(t, err, "Should create custom icon")

	err = launcher.SetIcon(iconPath)
	require.NoError(t, err, "Should set custom icon")

	err = launcher.DisplayIcon()
	assert.NoError(t, err, "Should display custom icon successfully")
}

// TestDisplayIcon_UnsupportedPlatform tests display on unsupported platform
func TestDisplayIcon_UnsupportedPlatform(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)
	launcher.platform = "unsupported" // Force unsupported platform

	err := launcher.DisplayIcon()
	assert.Error(t, err, "Should return error for unsupported platform")
	assert.Contains(t, err.Error(), "icon display not supported", "Error should mention not supported")
}

// TestGetAvailableIcons tests getting list of available icons
func TestGetAvailableIcons(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	// Create test icon files
	icons := []string{"icon1.png", "icon2.jpg", "icon3.jpeg", "icon4.ico", "icon5.icns"}
	for _, icon := range icons {
		err := os.WriteFile(filepath.Join(tempDir, icon), []byte("fake icon data"), 0644)
		require.NoError(t, err, "Should create test icon file")
	}

	// Create non-icon files
	nonIcons := []string{"text.txt", "data.json", "script.sh"}
	for _, file := range nonIcons {
		err := os.WriteFile(filepath.Join(tempDir, file), []byte("fake data"), 0644)
		require.NoError(t, err, "Should create test non-icon file")
	}

	available, err := launcher.GetAvailableIcons()
	assert.NoError(t, err, "Should get available icons successfully")
	assert.Len(t, available, len(icons), "Should return only icon files")

	// Verify all returned files are icons
	for _, icon := range available {
		ext := filepath.Ext(icon)
		assert.Contains(t, []string{".png", ".jpg", ".jpeg", ".ico", ".icns"}, ext, "Should only return image files")
	}
}

// TestGetAvailableIcons_NonExistentDirectory tests with non-existent directory
func TestGetAvailableIcons_NonExistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(filepath.Join(tempDir, "nonexistent"))

	_, err := launcher.GetAvailableIcons()
	assert.Error(t, err, "Should return error for non-existent directory")
	assert.Contains(t, err.Error(), "icon directory not found", "Error should mention directory not found")
}

// TestGetAvailableIcons_EmptyDirectory tests with empty directory
func TestGetAvailableIcons_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	available, err := launcher.GetAvailableIcons()
	assert.NoError(t, err, "Should not return error for empty directory")
	assert.Empty(t, available, "Should return empty list for empty directory")
}

// TestShowSplashScreen tests showing splash screen
func TestShowSplashScreen(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	// Create splash directory structure
	splashDir := filepath.Join(tempDir, "..", "splash")
	switch launcher.platform {
	case "android":
		splashDir = filepath.Join(splashDir, "android", "portrait", "xxxhdpi")
	case "ios":
		splashDir = filepath.Join(splashDir, "ios", "iphone")
	default:
		splashDir = filepath.Join(splashDir, "android", "portrait", "xxxhdpi")
	}
	err := os.MkdirAll(splashDir, 0755)
	require.NoError(t, err, "Should create splash directory")

	// Create splash file with the correct name
	var splashFile string
	switch launcher.platform {
	case "android", "unknown":
		splashFile = filepath.Join(splashDir, "splash_xxxhdpi_portrait.png")
	case "ios":
		splashFile = filepath.Join(splashDir, "splash_iphone_portrait.png")
	default:
		splashFile = filepath.Join(splashDir, "splash_xxxhdpi_portrait.png")
	}
	err = os.WriteFile(splashFile, []byte("fake splash data"), 0644)
	require.NoError(t, err, "Should create splash file")

	err = launcher.ShowSplashScreen("")
	assert.NoError(t, err, "Should show default splash screen successfully")
}

// TestShowSplashScreen_CustomPath tests showing custom splash screen
func TestShowSplashScreen_CustomPath(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	// Create custom splash file
	customSplash := filepath.Join(tempDir, "custom_splash.png")
	err := os.WriteFile(customSplash, []byte("fake custom splash data"), 0644)
	require.NoError(t, err, "Should create custom splash file")

	err = launcher.ShowSplashScreen(customSplash)
	assert.NoError(t, err, "Should show custom splash screen successfully")
}

// TestShowSplashScreen_NonExistentFile tests showing non-existent splash
func TestShowSplashScreen_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	nonExistentSplash := filepath.Join(tempDir, "nonexistent_splash.png")
	err := launcher.ShowSplashScreen(nonExistentSplash)
	assert.Error(t, err, "Should return error for non-existent splash file")
	assert.Contains(t, err.Error(), "splash screen file not found", "Error should mention file not found")
}

// TestGetInfo tests getting launcher information
func TestGetInfo(t *testing.T) {
	tempDir := t.TempDir()
	launcher := NewLauncher(tempDir)

	// Create test icon
	iconPath := filepath.Join(tempDir, "test.png")
	err := os.WriteFile(iconPath, []byte("fake png data"), 0644)
	require.NoError(t, err, "Should create test icon")

	err = launcher.SetIcon(iconPath)
	require.NoError(t, err, "Should set test icon")

	info, err := launcher.GetInfo()
	assert.NoError(t, err, "Should get launcher info successfully")
	assert.NotNil(t, info, "Info should not be nil")

	assert.Equal(t, launcher.platform, info.Platform, "Should have correct platform")
	assert.Equal(t, launcher.GetPlatformIcon(), info.IconPath, "Should have correct icon path")
	assert.NotEmpty(t, info.Available, "Should have available icons")
	assert.Contains(t, info.Available, "test.png", "Should include test icon in available icons")
}

// TestLauncherInfo_Structure tests LauncherInfo structure
func TestLauncherInfo_Structure(t *testing.T) {
	info := &LauncherInfo{
		Platform:   "test_platform",
		IconPath:   "/test/icon.png",
		Available:  []string{"icon1.png", "icon2.jpg"},
		SplashPath: "/test/splash.png",
	}

	assert.Equal(t, "test_platform", info.Platform, "Should store platform correctly")
	assert.Equal(t, "/test/icon.png", info.IconPath, "Should store icon path correctly")
	assert.Equal(t, []string{"icon1.png", "icon2.jpg"}, info.Available, "Should store available icons correctly")
	assert.Equal(t, "/test/splash.png", info.SplashPath, "Should store splash path correctly")
}