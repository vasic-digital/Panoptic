#!/bin/bash

# Build script for Panoptic launcher tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="panoptic-launcher"
SOURCE_DIR="cmd/launcher"
OUTPUT_DIR="bin"
PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64" "windows/arm64")

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to build for current platform
build_current() {
    print_status "Building for current platform..."
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    # Build the binary
    go build -o "$OUTPUT_DIR/$BINARY_NAME" "./$SOURCE_DIR"
    
    if [[ $? -eq 0 ]]; then
        print_success "Built successfully: $OUTPUT_DIR/$BINARY_NAME"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Function to build for all platforms
build_all() {
    print_status "Building for all platforms..."
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    for platform in "${PLATFORMS[@]}"; do
        GOOS=${platform%/*}
        GOARCH=${platform#*/}
        
        print_status "Building for $GOOS/$GOARCH..."
        
        # Set binary name with extension for Windows
        binary_name="$BINARY_NAME"
        if [[ "$GOOS" == "windows" ]]; then
            binary_name="$binary_name.exe"
        fi
        
        # Create platform-specific directory
        platform_dir="$OUTPUT_DIR/$GOOS-$GOARCH"
        mkdir -p "$platform_dir"
        
        # Build the binary
        GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$platform_dir/$binary_name" "./$SOURCE_DIR"
        
        if [[ $? -eq 0 ]]; then
            print_success "Built: $platform_dir/$binary_name"
        else
            print_error "Build failed for $GOOS/$GOARCH"
            exit 1
        fi
    done
    
    print_success "All platforms built successfully"
}

# Function to install the launcher tool
install_tool() {
    print_status "Installing launcher tool..."
    
    # Build for current platform
    build_current
    
    # Copy to a location in PATH (optional)
    if [[ -d "/usr/local/bin" ]]; then
        sudo cp "$OUTPUT_DIR/$BINARY_NAME" "/usr/local/bin/"
        print_success "Installed to /usr/local/bin/$BINARY_NAME"
    else
        print_warning "Could not install to /usr/local/bin. Binary is available in $OUTPUT_DIR/$BINARY_NAME"
        print_warning "You can add $OUTPUT_DIR to your PATH or copy the binary manually"
    fi
}

# Function to test the launcher tool
test_tool() {
    print_status "Testing launcher tool..."
    
    # Build first
    build_current
    
    # Check if icons exist
    if [[ ! -d "Assets/icons" ]]; then
        print_warning "Icons directory not found. Generating icons first..."
        ./scripts/generate_icons.sh
    fi
    
    # Test the tool
    print_status "Testing launcher tool..."
    
    # Test 1: Show info
    print_status "Test 1: Showing launcher info..."
    "./$OUTPUT_DIR/$BINARY_NAME" --info
    
    # Test 2: List icons
    print_status "Test 2: Listing available icons..."
    "./$OUTPUT_DIR/$BINARY_NAME" --list | head -10
    
    # Test 3: Display default icon
    print_status "Test 3: Displaying default icon..."
    "./$OUTPUT_DIR/$BINARY_NAME"
    
    print_success "All tests passed!"
}

# Function to clean build artifacts
clean() {
    print_status "Cleaning build artifacts..."
    
    if [[ -d "$OUTPUT_DIR" ]]; then
        rm -rf "$OUTPUT_DIR"
        print_success "Cleaned $OUTPUT_DIR"
    else
        print_warning "No build artifacts to clean"
    fi
}

# Main function
main() {
    case "${1:-current}" in
        "current")
            build_current
            ;;
        "all")
            build_all
            ;;
        "install")
            install_tool
            ;;
        "test")
            test_tool
            ;;
        "clean")
            clean
            ;;
        "help"|"-h"|"--help")
            echo "Panoptic Launcher Tool Build Script"
            echo ""
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  current  - Build for current platform (default)"
            echo "  all      - Build for all supported platforms"
            echo "  install  - Build and install to system PATH"
            echo "  test     - Build and test the launcher tool"
            echo "  clean    - Clean build artifacts"
            echo "  help     - Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0              # Build for current platform"
            echo "  $0 all          # Build for all platforms"
            echo "  $0 install      # Build and install"
            echo "  $0 test         # Build and test"
            echo "  $0 clean        # Clean build artifacts"
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use '$0 help' to see available commands"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"