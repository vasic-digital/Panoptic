package main

import (
	"flag"
	"fmt"
	"os"

	"panoptic/internal/launcher"
)

func main() {
	var (
		iconDir   = flag.String("icons", "Assets/icons", "Directory containing icons")
		iconFile  = flag.String("icon", "", "Specific icon file to display")
		splash    = flag.String("splash", "", "Splash screen file to display")
		list      = flag.Bool("list", false, "List available icons")
		info      = flag.Bool("info", false, "Show launcher information")
		platform  = flag.String("platform", "", "Override platform detection")
	)
	
	flag.Parse()
	
	// Check if icon directory exists
	if _, err := os.Stat(*iconDir); os.IsNotExist(err) {
		fmt.Printf("❌ Error: Icon directory not found: %s\n", *iconDir)
		fmt.Println("💡 Tip: Run './scripts/generate_icons.sh' to generate icons first")
		os.Exit(1)
	}
	
	// Create launcher instance
	lnchr := launcher.NewLauncher(*iconDir)
	
	// Override platform if specified
	if *platform != "" {
		fmt.Printf("📱 Using platform override: %s\n", *platform)
		// Note: In a real implementation, you'd set the platform on the launcher
	}
	
	// Handle different commands
	switch {
	case *list:
		icons, err := lnchr.GetAvailableIcons()
		if err != nil {
			fmt.Printf("❌ Error getting available icons: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("📁 Available icons in %s:\n", *iconDir)
		for i, icon := range icons {
			fmt.Printf("   %d. %s\n", i+1, icon)
		}
		
	case *info:
		info, err := lnchr.GetInfo()
		if err != nil {
			fmt.Printf("❌ Error getting launcher info: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("🎯 Launcher Information:\n")
		fmt.Printf("   Platform: %s\n", info.Platform)
		fmt.Printf("   Default Icon: %s\n", info.IconPath)
		fmt.Printf("   Available Icons: %d\n", len(info.Available))
		
	case *splash != "":
		err := lnchr.ShowSplashScreen(*splash)
		if err != nil {
			fmt.Printf("❌ Error displaying splash screen: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Splash screen displayed successfully\n")
		
	case *iconFile != "":
		err := lnchr.SetIcon(*iconFile)
		if err != nil {
			fmt.Printf("❌ Error setting icon: %v\n", err)
			os.Exit(1)
		}
		
		err = lnchr.DisplayIcon()
		if err != nil {
			fmt.Printf("❌ Error displaying icon: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ Icon displayed successfully: %s\n", *iconFile)
		
	default:
		// Default action: display the platform-specific icon
		iconPath := lnchr.GetPlatformIcon()
		if iconPath == "" {
			fmt.Printf("❌ Error: No default icon available for platform\n")
			os.Exit(1)
		}
		
		err := lnchr.DisplayIcon()
		if err != nil {
			fmt.Printf("❌ Error displaying icon: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ Default icon displayed successfully: %s\n", iconPath)
		
		// Also show splash screen
		err = lnchr.ShowSplashScreen("")
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not display splash screen: %v\n", err)
		} else {
			fmt.Printf("✅ Splash screen displayed successfully\n")
		}
	}
}

func init() {
	// Set usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "🎯 Panoptic Launcher Icon Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                    # Display default icon\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --list            # List available icons\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --icon web/icon.png  # Display specific icon\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --splash splash/android/portrait/xxxhdpi/splash_xxxhdpi_portrait.png\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --info             # Show launcher information\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --icons ./custom_icons  # Use custom icon directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "💡 Tip: Run './scripts/generate_icons.sh' to generate icons first\n")
	}
}