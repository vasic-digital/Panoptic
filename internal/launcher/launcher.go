package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Launcher represents a launcher icon manager
type Launcher struct {
	iconDir      string
	currentIcon  string
	platform     string
}

// NewLauncher creates a new launcher icon manager
func NewLauncher(iconDir string) *Launcher {
	return &Launcher{
		iconDir:  iconDir,
		platform: detectPlatform(),
	}
}

// detectPlatform detects the current platform
func detectPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	case "android":
		return "android"
	case "ios":
		return "ios"
	default:
		return "unknown"
	}
}

// SetIcon sets the current launcher icon
func (l *Launcher) SetIcon(iconPath string) error {
	if !filepath.IsAbs(iconPath) {
		iconPath = filepath.Join(l.iconDir, iconPath)
	}
	
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		return fmt.Errorf("icon file not found: %s", iconPath)
	}
	
	l.currentIcon = iconPath
	return nil
}

// GetIconPath returns the current icon path
func (l *Launcher) GetIconPath() string {
	return l.currentIcon
}

// GetPlatformIcon returns the appropriate icon for the current platform
func (l *Launcher) GetPlatformIcon() string {
	switch l.platform {
	case "windows":
		return filepath.Join(l.iconDir, "desktop", "icon.png")
	case "macos":
		return filepath.Join(l.iconDir, "desktop", "large.png")
	case "linux":
		return filepath.Join(l.iconDir, "desktop", "icon.png")
	case "android":
		return filepath.Join(l.iconDir, "android", "xxxhdpi", "icon_xxxhdpi.png")
	case "ios":
		return filepath.Join(l.iconDir, "ios", "appstore", "icon_appstore.png")
	default:
		return filepath.Join(l.iconDir, "desktop", "icon.png")
	}
}

// DisplayIcon displays the launcher icon (platform-specific implementation)
func (l *Launcher) DisplayIcon() error {
	iconPath := l.currentIcon
	if iconPath == "" {
		iconPath = l.GetPlatformIcon()
	}
	
	if iconPath == "" {
		return fmt.Errorf("no icon available to display")
	}
	
	switch l.platform {
	case "windows":
		return l.displayWindowsIcon(iconPath)
	case "macos":
		return l.displayMacOSIcon(iconPath)
	case "linux":
		return l.displayLinuxIcon(iconPath)
	default:
		return fmt.Errorf("icon display not supported on platform: %s", l.platform)
	}
}

// displayWindowsIcon displays the icon on Windows
func (l *Launcher) displayWindowsIcon(iconPath string) error {
	// On Windows, we could use Windows API to display the icon
	// For now, we'll just log that we're displaying it
	fmt.Printf("🎯 Displaying launcher icon: %s\n", iconPath)
	return nil
}

// displayMacOSIcon displays the icon on macOS
func (l *Launcher) displayMacOSIcon(iconPath string) error {
	// On macOS, we could use NSImage to display the icon
	// For now, we'll just log that we're displaying it
	fmt.Printf("🎯 Displaying launcher icon: %s\n", iconPath)
	return nil
}

// displayLinuxIcon displays the icon on Linux
func (l *Launcher) displayLinuxIcon(iconPath string) error {
	// On Linux, we could use GTK or other libraries to display the icon
	// For now, we'll just log that we're displaying it
	fmt.Printf("🎯 Displaying launcher icon: %s\n", iconPath)
	return nil
}

// GetAvailableIcons returns a list of available icons
func (l *Launcher) GetAvailableIcons() ([]string, error) {
	var icons []string
	
	// Check if icon directory exists
	if _, err := os.Stat(l.iconDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("icon directory not found: %s", l.iconDir)
	}
	
	// Walk through the icon directory
	err := filepath.Walk(l.iconDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Check if file is an image
		ext := filepath.Ext(path)
		switch ext {
		case ".png", ".jpg", ".jpeg", ".ico", ".icns":
			relPath, err := filepath.Rel(l.iconDir, path)
			if err != nil {
				return err
			}
			icons = append(icons, relPath)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to scan icon directory: %v", err)
	}
	
	return icons, nil
}

// ShowSplashScreen displays a splash screen with the launcher icon
func (l *Launcher) ShowSplashScreen(splashPath string) error {
	if splashPath == "" {
		// Get default splash screen for platform
		switch l.platform {
		case "android":
			splashPath = filepath.Join(l.iconDir, "..", "splash", "android", "portrait", "xxxhdpi", "splash_xxxhdpi_portrait.png")
		case "ios":
			splashPath = filepath.Join(l.iconDir, "..", "splash", "ios", "iphone", "splash_iphone_portrait.png")
		default:
			splashPath = filepath.Join(l.iconDir, "..", "splash", "android", "portrait", "xxxhdpi", "splash_xxxhdpi_portrait.png")
		}
	}
	
	if !filepath.IsAbs(splashPath) {
		splashPath = filepath.Join(l.iconDir, "..", splashPath)
	}
	
	if _, err := os.Stat(splashPath); os.IsNotExist(err) {
		return fmt.Errorf("splash screen file not found: %s", splashPath)
	}
	
	// Display splash screen (platform-specific implementation)
	switch l.platform {
	case "windows":
		fmt.Printf("🖼️  Displaying splash screen: %s\n", splashPath)
	case "macos":
		fmt.Printf("🖼️  Displaying splash screen: %s\n", splashPath)
	case "linux":
		fmt.Printf("🖼️  Displaying splash screen: %s\n", splashPath)
	default:
		fmt.Printf("🖼️  Displaying splash screen: %s\n", splashPath)
	}
	
	return nil
}

// LauncherInfo contains information about the launcher
type LauncherInfo struct {
	Platform    string   `json:"platform"`
	IconPath    string   `json:"icon_path"`
	Available   []string `json:"available_icons"`
	SplashPath  string   `json:"splash_path,omitempty"`
}

// GetInfo returns launcher information
func (l *Launcher) GetInfo() (*LauncherInfo, error) {
	available, err := l.GetAvailableIcons()
	if err != nil {
		return nil, err
	}
	
	info := &LauncherInfo{
		Platform:  l.platform,
		IconPath:  l.GetPlatformIcon(),
		Available: available,
	}
	
	return info, nil
}