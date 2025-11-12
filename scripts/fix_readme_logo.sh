#!/bin/bash

# Fix README logo and title
# This script removes duplicate titles and properly embeds the logo

set -e

echo "🔧 Fixing README logo and title..."

# Read the base64 logo
LOGO_BASE64=$(base64 Assets/Logo.jpeg | tr -d '\n')

# Create temporary file with fixed content
cat > /tmp/readme_fixed.md << 'EOF'
# 🎯 Panoptic

<!-- Panoptic Logo -->
<div align="center">
  <img src="data:image/jpeg;base64,LOGO_PLACEHOLDER" alt="Panoptic Logo" width="200"/>
</div>

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/your-org/panoptic/actions)
[![Coverage](https://img.shields.io/badge/Coverage-78%25-yellow.svg)](docs/COVERAGE_REPORT.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-org/panoptic)](https://goreportcard.com/report/github.com/your-org/panoptic)

**Comprehensive Automated Testing & Recording Framework**

A powerful, multi-platform testing solution for web, desktop, and mobile applications with advanced UI automation, screenshot capture, and video recording capabilities.

</div>

## 📋 Table of Contents

- [Features](#-features)
- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Usage](#-usage)
- [Configuration](#-configuration)
- [Platform Support](#-platform-support)
- [Advanced Features](#-advanced-features)
- [Examples](#-examples)
- [Testing](#-testing)
- [Contributing](#-contributing)
- [License](#-license)

EOF

# Replace the placeholder with actual base64 data
sed "s/LOGO_PLACEHOLDER/$LOGO_BASE64/g" /tmp/readme_fixed.md > /tmp/readme_with_logo.md

# Read the rest of the current README (skip the first 15 lines)
tail -n +16 README.md >> /tmp/readme_with_logo.md

# Replace the README
mv /tmp/readme_with_logo.md README.md

echo "✅ README logo and title fixed successfully!"

# Show first few lines to verify
echo -e "\n📖 First 15 lines of fixed README:"
head -15 README.md