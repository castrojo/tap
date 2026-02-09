#!/bin/bash
# Setup git hooks for the repository
# Run this script after cloning to install pre-commit validation

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🔧 Installing git hooks..."

# Check if running from repository root
if [ ! -d ".git" ]; then
    echo -e "${RED}✗ Error: Must run from repository root${NC}"
    exit 1
fi

# Check if tap-validate exists, build if needed
if [ ! -f "./tap-tools/tap-validate" ]; then
    echo -e "${YELLOW}⚠ tap-validate not found. Building tap-tools...${NC}"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}✗ Error: Go is not installed${NC}"
        echo "Please install Go and try again"
        exit 1
    fi
    
    # Build tap-validate
    cd tap-tools
    echo "Building tap-validate..."
    if go build -o tap-validate ./cmd/tap-validate; then
        echo -e "${GREEN}✓ tap-validate built successfully${NC}"
    else
        echo -e "${RED}✗ Failed to build tap-validate${NC}"
        exit 1
    fi
    cd ..
fi

# Verify tap-validate works
if ! ./tap-tools/tap-validate --help &> /dev/null; then
    echo -e "${RED}✗ Error: tap-validate is not working correctly${NC}"
    exit 1
fi

# Install pre-commit hook
if [ -f ".git/hooks/pre-commit" ]; then
    # Check if it's already our hook
    if grep -q "Pre-commit hook: Validate Ruby files" .git/hooks/pre-commit; then
        echo -e "${GREEN}✓ Pre-commit hook already installed${NC}"
    else
        echo -e "${YELLOW}⚠ Backing up existing pre-commit hook to pre-commit.backup${NC}"
        mv .git/hooks/pre-commit .git/hooks/pre-commit.backup
        cp scripts/git-hooks/pre-commit .git/hooks/pre-commit
        chmod +x .git/hooks/pre-commit
        echo -e "${GREEN}✓ Pre-commit hook installed${NC}"
    fi
else
    cp scripts/git-hooks/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo -e "${GREEN}✓ Pre-commit hook installed${NC}"
fi

# Verify hook is executable
if [ ! -x ".git/hooks/pre-commit" ]; then
    echo -e "${RED}✗ Error: Pre-commit hook is not executable${NC}"
    chmod +x .git/hooks/pre-commit
fi

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  ✓ Git hooks installed successfully                      ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "The pre-commit hook will:"
echo "  • Automatically validate all Ruby files before commit"
echo "  • Auto-fix style issues using tap-validate --fix"
echo "  • Block commits that fail validation"
echo "  • Re-stage files if they were modified by --fix"
echo ""
echo -e "${YELLOW}Note:${NC} You can bypass the hook with: ${RED}git commit --no-verify${NC}"
echo -e "      ${YELLOW}(This is NOT RECOMMENDED and may cause CI failures)${NC}"
echo ""
echo -e "${GREEN}You're all set! Start creating packages with:${NC}"
echo "  ./tap-tools/tap-cask generate https://github.com/user/repo"
echo "  ./tap-tools/tap-formula generate https://github.com/user/repo"
echo ""
