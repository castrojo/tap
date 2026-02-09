#!/bin/bash
# Setup git hooks for the repository
# Run this script after cloning to install pre-commit validation

set -e

echo "Installing git hooks..."

# Check if running from repository root
if [ ! -d ".git" ]; then
    echo "Error: Must run from repository root"
    exit 1
fi

# Check if tap-validate exists
if [ ! -f "./tap-tools/tap-validate" ]; then
    echo "Warning: tap-validate not found!"
    echo "Building tap-tools..."
    cd tap-tools
    go build -o tap-validate ./cmd/tap-validate
    cd ..
fi

# Install pre-commit hook
if [ -f ".git/hooks/pre-commit" ]; then
    echo "Backing up existing pre-commit hook to pre-commit.backup"
    mv .git/hooks/pre-commit .git/hooks/pre-commit.backup
fi

cp scripts/git-hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

echo "✓ Pre-commit hook installed"
echo ""
echo "The hook will:"
echo "  - Validate all Ruby files before commit"
echo "  - Auto-fix style issues with --fix flag"
echo "  - Block commits that fail validation"
echo ""
echo "To bypass (not recommended): git commit --no-verify"
