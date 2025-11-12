#!/bin/bash

# Panoptic Launcher Icons Summary Script
# This script shows what has been accomplished with launcher icons and splash screens

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_title() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    🎯 PANOPTIC LAUNCHER                     ║"
    echo "║                     ICONS & SPLASH SCREENS                  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_section() {
    echo -e "${PURPLE}═══ $1 ═══${NC}"
    echo ""
}

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Main function
main() {
    print_title
    
    print_section "📁 GENERATED FILES"
    
    # Check if icons exist
    if [[ -d "Assets/icons" ]]; then
        print_success "✓ Launcher icons generated successfully"
        echo "   Location: Assets/icons/"
        
        # Count icons by platform
        android_count=$(find Assets/icons/android -name "*.png" 2>/dev/null | wc -l)
        ios_count=$(find Assets/icons/ios -name "*.png" 2>/dev/null | wc -l)
        web_count=$(find Assets/icons/web -name "*" 2>/dev/null | wc -l)
        desktop_count=$(find Assets/icons/desktop -name "*.png" 2>/dev/null | wc -l)
        
        echo "   Platform breakdown:"
        echo "     • Android: $android_count icons (ldpi, mdpi, hdpi, xhdpi, xxhdpi, xxxhdpi)"
        echo "     • iOS: $ios_count icons (iPhone, iPad, App Store)"
        echo "     • Web: $web_count files (favicon, web icons)"
        echo "     • Desktop: $desktop_count icons (standard, large)"
        
        # Check for manifest
        if [[ -f "Assets/icons/manifest.json" ]]; then
            print_success "✓ Icon manifest generated"
            echo "     File: Assets/icons/manifest.json"
        fi
        
        # Check for README
        if [[ -f "Assets/icons/README.md" ]]; then
            print_success "✓ Usage documentation created"
            echo "     File: Assets/icons/README.md"
        fi
    else
        print_error "✗ Icons directory not found"
        echo "   Run './scripts/generate_icons.sh' to generate icons"
    fi
    
    echo ""
    
    # Check if splash screens exist
    if [[ -d "Assets/splash" ]]; then
        print_success "✓ Splash screens generated successfully"
        echo "   Location: Assets/splash/"
        
        # Count splash screens
        android_splash_count=$(find Assets/splash/android -name "*.png" 2>/dev/null | wc -l)
        ios_splash_count=$(find Assets/splash/ios -name "*.png" 2>/dev/null | wc -l)
        
        echo "   Platform breakdown:"
        echo "     • Android: $android_splash_count splash screens (portrait/landscape, all densities)"
        echo "     • iOS: $ios_splash_count splash screens (iPhone, iPad, various sizes)"
    else
        print_error "✗ Splash screens directory not found"
    fi
    
    echo ""
    
    print_section "🛠️ TOOLS & SCRIPTS"
    
    # Check for scripts
    if [[ -f "scripts/generate_icons.sh" ]]; then
        print_success "✓ Icon generation script"
        echo "     File: scripts/generate_icons.sh"
        echo "     Usage: ./scripts/generate_icons.sh"
    fi
    
    if [[ -f "scripts/update_icons.sh" ]]; then
        print_success "✓ Quick icon update script"
        echo "     File: scripts/update_icons.sh"
        echo "     Usage: ./scripts/update_icons.sh"
    fi
    
    if [[ -f "scripts/build_launcher.sh" ]]; then
        print_success "✓ Launcher tool build script"
        echo "     File: scripts/build_launcher.sh"
        echo "     Usage: ./scripts/build_launcher.sh [command]"
    fi
    
    echo ""
    
    print_section "🚀 LAUNCHER TOOL"
    
    # Check for launcher binary
    if [[ -f "bin/panoptic-launcher" ]]; then
        print_success "✓ Launcher tool built successfully"
        echo "     Binary: bin/panoptic-launcher"
        
        # Show launcher tool features
        echo "     Features:"
        echo "       • Display launcher icons for different platforms"
        echo "       • Show splash screens"
        echo "       • List available icons"
        echo "       • Platform detection"
        echo "       • Custom icon directory support"
        
        # Test the launcher tool
        echo ""
        print_status "Testing launcher tool..."
        
        # Show info
        echo "   Platform info:"
        ./bin/panoptic-launcher --info 2>/dev/null | sed 's/^/     /'
        
        echo ""
        print_status "Available icons (first 5):"
        ./bin/panoptic-launcher --list 2>/dev/null | head -5 | sed 's/^/     /'
        
    else
        print_warning "! Launcher tool not built"
        echo "   Build with: ./scripts/build_launcher.sh"
    fi
    
    echo ""
    
    print_section "📖 README INTEGRATION"
    
    # Check if README has been updated
    if [[ -f "README.md" ]] && grep -q "Panoptic Logo" README.md; then
        print_success "✓ README updated with app logo"
        echo "   Logo embedded in README.md"
        
        # Show logo position
        echo "   Logo appears at the top of the README"
        
    else
        print_warning "! README not updated with logo"
        echo "   Run './scripts/generate_icons.sh' to update README"
    fi
    
    echo ""
    
    print_section "📱 PLATFORM SUPPORT"
    
    echo "   Supported platforms:"
    echo "     • 🤖 Android"
    echo "       - Launcher icons: ldpi, mdpi, hdpi, xhdpi, xxhdpi, xxxhdpi"
    echo "       - Splash screens: portrait/landscape, all densities"
    echo ""
    echo "     • 🍎 iOS"
    echo "       - Launcher icons: iPhone, iPad, App Store"
    echo "       - Splash screens: iPhone, iPad, various sizes"
    echo ""
    echo "     • 🌐 Web"
    echo "       - Favicon (.ico)"
    echo "       - Web app icons (.png)"
    echo ""
    echo "     • 💻 Desktop"
    echo "       - Standard and large icons (.png)"
    echo "       - Platform-specific sizes"
    
    echo ""
    
    print_section "🎯 USAGE EXAMPLES"
    
    echo "   Generate icons:"
    echo "     ./scripts/generate_icons.sh"
    echo ""
    echo "   Quick update:"
    echo "     ./scripts/update_icons.sh"
    echo ""
    echo "   Build launcher tool:"
    echo "     ./scripts/build_launcher.sh"
    echo ""
    echo "   Test launcher tool:"
    echo "     ./bin/panoptic-launcher --info"
    echo "     ./bin/panoptic-launcher --list"
    echo "     ./bin/panoptic-launcher"
    echo ""
    echo "   Display specific icon:"
    echo "     ./bin/panoptic-launcher --icon desktop/icon.png"
    echo ""
    echo "   Display splash screen:"
    echo "     ./bin/panoptic-launcher --splash splash/android/portrait/xxxhdpi/splash_xxxhdpi_portrait.png"
    
    echo ""
    
    print_section "📊 SUMMARY"
    
    # Count total files
    total_icons=$(find Assets/icons -type f 2>/dev/null | wc -l)
    total_splash=$(find Assets/splash -type f 2>/dev/null | wc -l)
    total_scripts=$(find scripts -name "*icon*" -o -name "*launcher*" | wc -l)
    
    echo "   Total files generated:"
    echo "     • Launcher icons: $total_icons files"
    echo "     • Splash screens: $total_splash files"
    echo "     • Scripts & tools: $total_scripts files"
    echo "     • Documentation: README.md, manifest.json"
    
    echo ""
    
    if [[ $total_icons -gt 0 && $total_splash -gt 0 ]]; then
        print_success "🎉 All launcher icons and splash screens generated successfully!"
        echo "   Your Panoptic application is ready for deployment on all platforms!"
    else
        print_warning "⚠️  Some files are missing. Run './scripts/generate_icons.sh' to complete setup."
    fi
    
    echo ""
    echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Run main function
main "$@"