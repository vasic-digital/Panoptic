#!/bin/bash

# Panoptic Icon and Splash Screen Generator
# This script generates launcher icons and splash screens from the main logo
# and updates the README with the app logo.

set -e

# Configuration
ASSETS_DIR="Assets"
LOGO_FILE="$ASSETS_DIR/Logo.jpeg"
ICONS_DIR="Assets/icons"
SPLASH_DIR="Assets/splash"
README_FILE="README.md"

# Icon sizes for different platforms
ICON_SIZES="
android_ldpi:36x36
android_mdpi:48x48
android_hdpi:72x72
android_xhdpi:96x96
android_xxhdpi:144x144
android_xxxhdpi:192x192
ios_iphone:60x60
ios_ipad:76x76
ios_ipad_retina:152x152
ios_iphone_retina:120x120
ios_appstore:1024x1024
favicon:16x16
web_icon:32x32
desktop_icon:256x256
desktop_large:512x512
"

# Splash screen sizes
SPLASH_SIZES="
android_portrait_ldpi:200x320
android_portrait_mdpi:320x480
android_portrait_hdpi:480x800
android_portrait_xhdpi:720x1280
android_portrait_xxhdpi:1080x1920
android_portrait_xxxhdpi:1440x2560
android_landscape_ldpi:320x200
android_landscape_mdpi:480x320
android_landscape_hdpi:800x480
android_landscape_xhdpi:1280x720
android_landscape_xxhdpi:1920x1080
android_landscape_xxxhdpi:2560x1440
ios_iphone_portrait:375x667
ios_iphone_plus_portrait:414x736
ios_iphone_x_portrait:375x812
ios_ipad_portrait:768x1024
ios_ipad_landscape:1024x768
"

# Colors for README badge
README_BADGE_COLOR="#0066cc"

# Function to print colored output
print_status() {
    echo -e "\033[1;34m[INFO]\033[0m $1"
}

print_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1" >&2
}

# Function to check if logo file exists
check_logo_file() {
    if [[ ! -f "$LOGO_FILE" ]]; then
        print_error "Logo file not found: $LOGO_FILE"
        exit 1
    fi
}

# Function to create directories
create_directories() {
    print_status "Creating directories for icons and splash screens..."
    
    mkdir -p "$ICONS_DIR"/{android,ios,web,desktop}
    mkdir -p "$SPLASH_DIR"/{android,ios}
    
    # Create subdirectories for Android and iOS
    mkdir -p "$ICONS_DIR"/android/{ldpi,mdpi,hdpi,xhdpi,xxhdpi,xxxhdpi}
    mkdir -p "$ICONS_DIR"/ios/{iphone,ipad,appstore}
    mkdir -p "$SPLASH_DIR"/android/{portrait,landscape}/{ldpi,mdpi,hdpi,xhdpi,xxhdpi,xxxhdpi}
    mkdir -p "$SPLASH_DIR"/ios/{iphone,ipad}
    
    print_success "Directories created successfully"
}

# Function to make background transparent
make_transparent() {
    local input_file="$1"
    local output_file="$2"
    
    print_status "Making background transparent for $input_file..."
    
    # First convert to PNG and try to remove white background
    convert "$input_file" \
        -alpha on \
        -channel rgba \
        -fuzz 15% \
        -fill none \
        -floodfill +0+0 white \
        "$output_file" 2>/dev/null || {
        # If white removal fails, try black background
        convert "$input_file" \
            -alpha on \
            -channel rgba \
            -fuzz 15% \
            -fill none \
            -floodfill +0+0 black \
            "$output_file" 2>/dev/null || {
            # If both fail, just convert to PNG with transparency
            convert "$input_file" \
                -background none \
                -alpha remove \
                "$output_file"
        }
    }
    
    print_status "Background transparency applied"
}

# Function to get size from the ICON_SIZES string
get_icon_size() {
    local key="$1"
    echo "$ICON_SIZES" | grep "^$key:" | cut -d: -f2
}

# Function to get size from the SPLASH_SIZES string  
get_splash_size() {
    local key="$1"
    echo "$SPLASH_SIZES" | grep "^$key:" | cut -d: -f2
}

# Function to generate icons
generate_icons() {
    print_status "Generating launcher icons..."
    
    local temp_logo="/tmp/panoptic_logo_transparent.png"
    make_transparent "$LOGO_FILE" "$temp_logo"
    
    echo "$ICON_SIZES" | while IFS=: read -r key size; do
        [[ -z "$key" || -z "$size" ]] && continue
        
        local output_path=""
        
        # Determine output path based on platform
        if [[ $key == android_* ]]; then
            local density="${key#android_}"
            output_path="$ICONS_DIR/android/$density/icon_${density}.png"
        elif [[ $key == ios_* ]]; then
            local device="${key#ios_}"
            case $device in
                "iphone") output_path="$ICONS_DIR/ios/iphone/icon_iphone.png" ;;
                "ipad") output_path="$ICONS_DIR/ios/ipad/icon_ipad.png" ;;
                "ipad_retina") output_path="$ICONS_DIR/ios/ipad/icon_ipad_retina.png" ;;
                "iphone_retina") output_path="$ICONS_DIR/ios/iphone/icon_iphone_retina.png" ;;
                "appstore") output_path="$ICONS_DIR/ios/appstore/icon_appstore.png" ;;
            esac
        elif [[ $key == favicon ]]; then
            output_path="$ICONS_DIR/web/favicon.ico"
            # Generate ICO file for favicon
            convert "$temp_logo" -resize "$size" -background none -alpha remove "$output_path"
            print_status "Generated: $output_path ($size)"
            continue
        elif [[ $key == web_* ]]; then
            output_path="$ICONS_DIR/web/${key#web_}.png"
        elif [[ $key == desktop_* ]]; then
            output_path="$ICONS_DIR/desktop/${key#desktop_}.png"
        fi
        
        # Generate the icon
        if [[ -n "$output_path" ]]; then
            convert "$temp_logo" \
                -resize "$size" \
                -background none \
                -alpha remove \
                -quality 100 \
                "$output_path"
            
            print_status "Generated: $output_path ($size)"
        fi
    done
    
    # Clean up temp file
    rm -f "$temp_logo"
    
    print_success "All icons generated successfully"
}

# Function to generate splash screens
generate_splash_screens() {
    print_status "Generating splash screens..."
    
    local temp_logo="/tmp/panoptic_logo_transparent.png"
    make_transparent "$LOGO_FILE" "$temp_logo"
    
    echo "$SPLASH_SIZES" | while IFS=: read -r key size; do
        [[ -z "$key" || -z "$size" ]] && continue
        
        local output_path=""
        
        # Determine output path based on platform and orientation
        if [[ $key == android_* ]]; then
            local orientation="${key#android_}"
            local density="${orientation#*_}"
            orientation="${orientation%_*}"
            
            output_path="$SPLASH_DIR/android/$orientation/$density/splash_${density}_${orientation}.png"
        elif [[ $key == ios_* ]]; then
            local device="${key#ios_}"
            local orientation=""
            
            # Handle different iOS device naming
            if [[ $device == *"iphone_plus"* ]]; then
                device="iphone_plus"
                orientation="portrait"
            elif [[ $device == *"iphone_x"* ]]; then
                device="iphone_x"
                orientation="portrait"
            elif [[ $device == *_* ]]; then
                orientation="${device#*_}"
                device="${device%_*}"
            fi
            
            # Create the device directory if it doesn't exist
            mkdir -p "$SPLASH_DIR/ios/$device"
            
            if [[ -n "$orientation" ]]; then
                output_path="$SPLASH_DIR/ios/$device/splash_${device}_${orientation}.png"
            else
                output_path="$SPLASH_DIR/ios/$device/splash_${device}.png"
            fi
        fi
        
        # Create splash screen with logo centered on transparent background
        if [[ -n "$output_path" ]]; then
            local width="${size%x*}"
            local height="${size#*x}"
            
            # Create a canvas with the specified dimensions and center the logo
            convert -size "$size" xc:none \
                "$temp_logo" -resize "80x80" \
                -gravity center -composite \
                "$output_path"
            
            print_status "Generated: $output_path ($size)"
        fi
    done
    
    # Clean up temp file
    rm -f "$temp_logo"
    
    print_success "All splash screens generated successfully"
}

# Function to generate README badge
generate_readme_badge() {
    print_status "Generating README badge..."
    
    local badge_icon="$ICONS_DIR/web/icon.png"
    
    if [[ ! -f "$badge_icon" ]]; then
        print_error "Web icon not found for README badge: $badge_icon"
        return 1
    fi
    
    # Create a base64 encoded version of the icon for README
    local base64_icon=$(base64 -i "$badge_icon")
    
    # Generate the README with logo
    if [[ -f "$README_FILE" ]]; then
        # Create backup
        cp "$README_FILE" "$README_FILE.backup"
        
        # Check if logo already exists
        if grep -q "Panoptic Logo" "$README_FILE"; then
            print_status "Logo already exists in README, skipping..."
        else
            # Create a temporary file with the logo section
            local temp_logo_file="/tmp/readme_logo_section.txt"
            cat > "$temp_logo_file" << EOF

<!-- Panoptic Logo -->
<div align="center">
  <img src="data:image/png;base64,$base64_icon" alt="Panoptic Logo" width="120" style="margin-bottom: 20px;">
</div>

EOF
            
            # Insert logo after the first line (title) using awk
            awk 'NR==1 {print; print; while ((getline line < "'"$temp_logo_file"'") > 0) print line; close("'$temp_logo_file'")} NR>1' "$README_FILE" > "$README_FILE.tmp" && mv "$README_FILE.tmp" "$README_FILE"
            
            # Clean up temp file
            rm -f "$temp_logo_file"
            
            print_success "README updated with logo"
        fi
    else
        print_error "README file not found: $README_FILE"
        return 1
    fi
}

# Function to create icon manifest
create_icon_manifest() {
    print_status "Creating icon manifest..."
    
    local manifest_file="$ICONS_DIR/manifest.json"
    
    cat > "$manifest_file" << EOF
{
  "generated": "$(date)",
  "source_logo": "$LOGO_FILE",
  "icons": {
    "android": {
EOF
    
    # Add Android icons
    local first=true
    echo "$ICON_SIZES" | grep "^android_" | while IFS=: read -r key size; do
        [[ -z "$key" || -z "$size" ]] && continue
        
        local density="${key#android_}"
        if [[ "$first" == true ]]; then
            first=false
        else
            echo "," >> "$manifest_file"
        fi
        cat >> "$manifest_file" << EOF
      "$density": {
        "size": "$size",
        "file": "android/$density/icon_${density}.png"
      }
EOF
    done
    
    cat >> "$manifest_file" << EOF
    },
    "ios": {
EOF
    
    # Add iOS icons
    first=true
    echo "$ICON_SIZES" | grep "^ios_" | while IFS=: read -r key size; do
        [[ -z "$key" || -z "$size" ]] && continue
        
        local device="${key#ios_}"
        if [[ "$first" == true ]]; then
            first=false
        else
            echo "," >> "$manifest_file"
        fi
        cat >> "$manifest_file" << EOF
      "$device": {
        "size": "$size",
        "file": "ios/$device/icon_${device}.png"
      }
EOF
    done
    
    cat >> "$manifest_file" << EOF
    },
    "web": {
      "favicon": {
        "size": "$(get_icon_size favicon)",
        "file": "web/favicon.ico"
      },
      "icon": {
        "size": "$(get_icon_size web_icon)",
        "file": "web/web_icon.png"
      }
    },
    "desktop": {
      "standard": {
        "size": "$(get_icon_size desktop_icon)",
        "file": "desktop/desktop_icon.png"
      },
      "large": {
        "size": "$(get_icon_size desktop_large)",
        "file": "desktop/desktop_large.png"
      }
    }
  },
  "splash_screens": {
    "android": {
EOF
    
    # Add Android splash screens
    first=true
    echo "$SPLASH_SIZES" | grep "^android_" | while IFS=: read -r key size; do
        [[ -z "$key" || -z "$size" ]] && continue
        
        local orientation="${key#android_}"
        local density="${orientation#*_}"
        orientation="${orientation%_*}"
        
        if [[ "$first" == true ]]; then
            first=false
        else
            echo "," >> "$manifest_file"
        fi
        
        cat >> "$manifest_file" << EOF
      "${orientation}_${density}": {
        "size": "$size",
        "file": "android/$orientation/$density/splash_${density}_${orientation}.png"
      }
EOF
    done
    
    cat >> "$manifest_file" << EOF
    },
    "ios": {
EOF
    
    # Add iOS splash screens
    first=true
    echo "$SPLASH_SIZES" | grep "^ios_" | while IFS=: read -r key size; do
        [[ -z "$key" || -z "$size" ]] && continue
        
        local device="${key#ios_}"
        local orientation="${device#*_}"
        device="${device%_*}"
        
        if [[ "$first" == true ]]; then
            first=false
        else
            echo "," >> "$manifest_file"
        fi
        
        cat >> "$manifest_file" << EOF
      "${device}_${orientation}": {
        "size": "$size",
        "file": "ios/$device/splash_${device}_${orientation}.png"
      }
EOF
    done
    
    cat >> "$manifest_file" << EOF
    }
  }
}
EOF
    
    print_success "Icon manifest created: $manifest_file"
}

# Function to create usage documentation
create_usage_docs() {
    print_status "Creating usage documentation..."
    
    local docs_file="$ICONS_DIR/README.md"
    
    cat > "$docs_file" << EOF
# Panoptic Icons and Splash Screens

This directory contains automatically generated launcher icons and splash screens for the Panoptic application.

## Generated Files

### Icons
- **Android**: Various densities (ldpi, mdpi, hdpi, xhdpi, xxhdpi, xxxhdpi)
- **iOS**: iPhone, iPad, and App Store icons
- **Web**: Favicon and web icons
- **Desktop**: Standard and large desktop icons

### Splash Screens
- **Android**: Portrait and landscape splash screens for various densities
- **iOS**: iPhone and iPad splash screens

## Usage

### Android
Copy the appropriate density icons from \`android/\` to your Android project's \`res/mipmap-{density}/\` directories.

### iOS
Copy the icons from \`ios/\` to your Xcode project's asset catalog.

### Web
Use the files in \`web/\` for your web application:
- \`favicon.ico\` for the browser tab icon
- \`web_icon.png\` for web app icons

### Desktop
Use the files in \`desktop/\` for desktop applications.

## Regeneration

To regenerate all icons and splash screens:

\`\`\`bash
./scripts/generate_icons.sh
\`\`\`

This will:
1. Process the logo from \`Assets/Logo.jpeg\`
2. Make the background transparent
3. Generate all required icons and splash screens
4. Update the main README with the app logo
5. Create an icon manifest

## Source Logo

The source logo file is located at \`../Logo.jpeg\`.

---
*Generated on $(date)*
EOF
    
    print_success "Usage documentation created: $docs_file"
}

# Main function
main() {
    print_status "Starting Panoptic icon and splash screen generation..."
    
    # Check requirements
    check_logo_file
    
    # Create directories
    create_directories
    
    # Generate icons
    generate_icons
    
    # Generate splash screens
    generate_splash_screens
    
    # Update README
    generate_readme_badge
    
    # Create icon manifest
    create_icon_manifest
    
    # Create usage documentation
    create_usage_docs
    
    print_success "Icon and splash screen generation completed successfully!"
    print_status "Generated files are located in:"
    print_status "  - Icons: $ICONS_DIR/"
    print_status "  - Splash screens: $SPLASH_DIR/"
    print_status "  - README updated with logo"
}

# Run main function
main "$@"