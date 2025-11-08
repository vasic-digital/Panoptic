#!/bin/bash

# Build script for Panoptic

set -e

echo "Building Panoptic..."

# Clean previous builds
if [ -f "panoptic" ]; then
    rm panoptic
fi

# Build for current platform
go build -o panoptic main.go

echo "Build completed. Run './panoptic --help' for usage."

# Example usage
echo ""
echo "Example usage:"
echo "./panoptic run example-config.yaml"