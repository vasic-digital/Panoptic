#!/bin/bash

# Quick Icon Update Script
# This is a simplified version for quick updates

set -e

echo "🔄 Updating Panoptic icons and splash screens..."
echo "This will regenerate all icons from the main logo."

# Check if we're in the right directory
if [[ ! -f "scripts/generate_icons.sh" ]]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# Run the main icon generation script
echo "📦 Running icon generation script..."
./scripts/generate_icons.sh

echo "✅ Icon update completed successfully!"
echo ""
echo "📁 Generated files:"
echo "   - Icons: Assets/icons/"
echo "   - Splash screens: Assets/splash/"
echo "   - README updated with logo"
echo ""
echo "💡 Tip: You can view the icon manifest at Assets/icons/manifest.json"
echo "💡 Tip: Usage documentation is available at Assets/icons/README.md"