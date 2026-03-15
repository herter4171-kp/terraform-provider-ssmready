#!/bin/bash
set -e

# ansible-ssm installer script
# Installs the ansible-ssm CLI to ~/.local/bin

INSTALL_DIR="$HOME/.local/bin"
MAN_DIR="$HOME/.local/share/man/man1"
BINARY_NAME="ansible-ssm"
MAN_PAGE="ansible-ssm.1"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

cat << EOF
Installing ansible-ssm CLI...

EOF

# Create install directories if they don't exist
if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating $INSTALL_DIR..."
    mkdir -p "$INSTALL_DIR"
fi

if [ ! -d "$MAN_DIR" ]; then
    echo "Creating $MAN_DIR..."
    mkdir -p "$MAN_DIR"
fi

# Check if binary exists in current directory
if [ ! -f "./$BINARY_NAME" ]; then
    echo -e "${RED}Error: $BINARY_NAME binary not found in current directory${NC}"
    echo "Please run 'make build-cli' first"
    exit 1
fi

# Copy binary to install directory
echo "Copying $BINARY_NAME to $INSTALL_DIR..."
cp "./$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Copy man page if it exists
if [ -f "./$MAN_PAGE" ]; then
    echo "Installing man page to $MAN_DIR..."
    cp "./$MAN_PAGE" "$MAN_DIR/"
    chmod 644 "$MAN_DIR/$MAN_PAGE"
fi

echo ""
echo -e "${GREEN}[OK]${NC} ansible-ssm installed to $INSTALL_DIR/"
echo -e "${GREEN}[OK]${NC} Man page installed to $MAN_DIR/"
echo ""

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}[WARNING]${NC} $INSTALL_DIR is not in your PATH"
    echo ""
    echo "Add this line to your shell configuration file:"
    echo ""
    
    # Detect shell
    if [ -n "$ZSH_VERSION" ]; then
        cat << 'EOF'
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
  source ~/.zshrc
EOF
    elif [ -n "$BASH_VERSION" ]; then
        cat << 'EOF'
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bash_profile
  source ~/.bash_profile
EOF
    else
        cat << 'EOF'
  export PATH="$HOME/.local/bin:$PATH"

Add this to your shell's configuration file (~/.zshrc, ~/.bash_profile, etc.)
EOF
    fi
    echo ""
else
    echo -e "${GREEN}[OK]${NC} $INSTALL_DIR is already in your PATH"
    echo ""
fi

# Check if man can find the page
if command -v manpath &> /dev/null; then
    MANPATH_OUTPUT=$(manpath 2>/dev/null || echo "")
    if [[ ":$MANPATH_OUTPUT:" != *":$HOME/.local/share/man:"* ]]; then
        echo -e "${BLUE}[INFO]${NC} To use the man page, you may need to add to your shell config:"
        echo '  export MANPATH="$HOME/.local/share/man:$MANPATH"'
        echo ""
    fi
fi

cat << EOF
You can now run:
  ansible-ssm --help
  man ansible-ssm
EOF
