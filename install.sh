#!/usr/bin/env bash
# AxiomLang (axlc) Automated Universal Installer for Linux (Ubuntu / Debian / Fedora / Arch)
set -e

echo "================================================================================"
echo "  AxiomLang Compiler & Toolchain (axlc) Installer for Linux"
echo "================================================================================"

INSTALL_DIR="/usr/local/bin"
SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1. Check dependencies
if ! command -v go &> /dev/null; then
    echo " [!] Go toolchain not found. Attempting to install golang..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y golang-go
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y golang
    elif command -v pacman &> /dev/null; then
        sudo pacman -S --noconfirm go
    else
        echo " [FATAL] Please install Go (1.20+) manually and run this script again."
        exit 1
    fi
fi

# 2. Build axlc compiler binary
echo " [*] Compiling axlc binary..."
cd "$SOURCE_DIR"
go build -ldflags="-s -w" -o axlc cmd/axlc/main.go

# 3. Install to system PATH
echo " [*] Installing binary to $INSTALL_DIR/axlc (requires sudo)..."
sudo cp axlc "$INSTALL_DIR/axlc"
sudo chmod +x "$INSTALL_DIR/axlc"

# 4. Register MIME type & syntax file associations for .axl and .axlui
if [ -d "/usr/share/mime/packages" ]; then
    echo " [*] Registering .axl and .axlui file associations..."
    sudo tee /usr/share/mime/packages/axiomlang.xml > /dev/null << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
  <mime-type type="text/x-axiom">
    <comment>AxiomLang Source File</comment>
    <glob pattern="*.axl"/>
  </mime-type>
  <mime-type type="text/x-axlui">
    <comment>AxiomLang Visual UI Specification</comment>
    <glob pattern="*.axlui"/>
  </mime-type>
</mime-info>
EOF
    sudo update-mime-database /usr/share/mime > /dev/null 2>&1 || true
fi

echo "================================================================================"
echo "  [SUCCESS] AxiomLang (axlc) successfully installed!"
echo "  Try running:"
echo "    axlc -run examples/01_hello_world.axl"
echo "    axlc -o my_app examples/09_axlui_gui_studio.axl"
echo "================================================================================"
