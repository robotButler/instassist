#!/bin/bash

# Simple installation script for instassist
# Run with: bash install.sh

set -e

BINARY_NAME="instassist"
INSTALL_PATH="${INSTALL_PATH:-/usr/local/bin}"
SCHEMA_PATH="${SCHEMA_PATH:-/usr/local/share/$BINARY_NAME}"

echo "🔨 Building $BINARY_NAME..."
go build -o "$BINARY_NAME" .

echo "📦 Installing binary to $INSTALL_PATH..."
sudo mkdir -p "$INSTALL_PATH"
sudo cp "$BINARY_NAME" "$INSTALL_PATH/"
sudo chmod +x "$INSTALL_PATH/$BINARY_NAME"

echo "📄 Installing schema to $SCHEMA_PATH..."
sudo mkdir -p "$SCHEMA_PATH"
sudo cp options.schema.json "$SCHEMA_PATH/"

echo ""
echo "✅ Installation complete!"
echo ""
echo "🚀 Run with: $BINARY_NAME"
echo "📚 See README.md for usage instructions"
echo ""
echo "To uninstall, run:"
echo "  sudo rm $INSTALL_PATH/$BINARY_NAME"
echo "  sudo rm -rf $SCHEMA_PATH"
