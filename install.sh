#!/usr/bin/env bash
# insta-assist installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/robotButler/instassist/master/install.sh | bash
#
# Options (via environment variables):
#   PREFIX     - Installation prefix (default: $HOME/.local)
#   BINDIR     - Binary directory (default: $PREFIX/bin)
#
# Examples:
#   # Install to ~/.local/bin (default, no sudo needed)
#   curl -fsSL https://raw.githubusercontent.com/robotButler/instassist/master/install.sh | bash
#
#   # Install system-wide to /usr/local/bin
#   curl -fsSL https://raw.githubusercontent.com/robotButler/instassist/master/install.sh | PREFIX=/usr/local sudo bash

set -euo pipefail

REPO_URL="https://github.com/robotButler/instassist.git"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-${PREFIX}/bin}"
TARGET="${BINDIR}/inst"

echo "==> insta-assist installer"
echo ""

# Check for Go
if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go toolchain not found. Please install Go first." >&2
  echo "Visit https://go.dev/dl/ for installation instructions." >&2
  exit 1
fi

# Check for git
if ! command -v git >/dev/null 2>&1; then
  echo "Error: git not found. Please install git first." >&2
  exit 1
fi

# Create temp directory and cleanup on exit
BUILD_DIR="$(mktemp -d)"
cleanup() { rm -rf "$BUILD_DIR"; }
trap cleanup EXIT

echo "==> Cloning repository..."
git clone --depth 1 "$REPO_URL" "$BUILD_DIR/instassist" 2>/dev/null

echo "==> Building inst..."
cd "$BUILD_DIR/instassist"
GO_INSTALL_DIR="${BUILD_DIR}/bin"
GOBIN="${GO_INSTALL_DIR}" go install ./cmd/inst

BIN_SRC="${GO_INSTALL_DIR}/inst"
if [ ! -f "$BIN_SRC" ]; then
  echo "Build failed: binary not found at ${BIN_SRC}" >&2
  exit 1
fi

echo "==> Installing to ${TARGET}..."
mkdir -p "${BINDIR}"
cp "${BIN_SRC}" "${TARGET}"
chmod +x "${TARGET}"

echo ""
echo "✅ Installed inst to ${TARGET}"
echo ""

# Check if BINDIR is in PATH
if [[ ":$PATH:" != *":${BINDIR}:"* ]]; then
  echo "Note: Add ${BINDIR} to your PATH:"
  echo ""
  echo "  # Add to your ~/.bashrc or ~/.zshrc:"
  echo "  export PATH=\"${BINDIR}:\$PATH\""
  echo ""
fi

echo "Run 'inst' to get started!"
