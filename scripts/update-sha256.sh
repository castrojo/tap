#!/usr/bin/env bash
#
# update-sha256.sh - Update version and SHA256 in existing formulas/casks
#
# Usage: ./update-sha256.sh <package-name> <new-version>
# Example: ./update-sha256.sh myapp 1.2.3

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

info() {
    echo -e "${BLUE}→ $1${NC}"
}

warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Validate inputs
if [ $# -ne 2 ]; then
    error "Usage: $0 <package-name> <new-version>\n  Example: $0 myapp 1.2.3"
fi

PACKAGE_NAME="$1"
NEW_VERSION="$2"

# Validate version format (basic check for semver-like versions)
if ! [[ "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?(-[a-zA-Z0-9._-]+)?(\+[a-zA-Z0-9._-]+)?$ ]]; then
    warn "Version format '$NEW_VERSION' doesn't match typical semver pattern (e.g., 1.2.3)"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Update cancelled"
    fi
fi

# Navigate to repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Check if jq is available
if ! command -v jq &> /dev/null; then
    error "jq is not installed. Install it with: brew install jq"
fi

# Find the package file (formula or cask)
PACKAGE_FILE=""
PACKAGE_TYPE=""

if [ -f "Formula/$PACKAGE_NAME.rb" ]; then
    PACKAGE_FILE="Formula/$PACKAGE_NAME.rb"
    PACKAGE_TYPE="formula"
elif [ -f "Casks/$PACKAGE_NAME.rb" ]; then
    PACKAGE_FILE="Casks/$PACKAGE_NAME.rb"
    PACKAGE_TYPE="cask"
else
    error "Package '$PACKAGE_NAME' not found in Formula/ or Casks/\n  Run 'ls Formula/ Casks/' to see available packages"
fi

success "Found $PACKAGE_TYPE: $PACKAGE_FILE"

# Extract current version
CURRENT_VERSION=$(grep -E '^\s*(version|url)' "$PACKAGE_FILE" | head -n 1 | sed -E 's/.*"([0-9]+\.[0-9]+(\.[0-9]+)?[^"]*)".*/\1/')

if [ -z "$CURRENT_VERSION" ]; then
    error "Could not extract current version from $PACKAGE_FILE"
fi

info "Current version: $CURRENT_VERSION"
info "New version: $NEW_VERSION"

# Check if versions are the same
if [ "$CURRENT_VERSION" = "$NEW_VERSION" ]; then
    warn "New version is the same as current version"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Update cancelled"
    fi
fi

# Extract current URL and SHA256
CURRENT_URL=$(grep -E '^\s*url\s+' "$PACKAGE_FILE" | head -n 1 | sed -E 's/.*url\s+"([^"]+)".*/\1/')

if [ -z "$CURRENT_URL" ]; then
    error "Could not extract URL from $PACKAGE_FILE"
fi

# Replace version in URL (handle both #{version} interpolation and literal versions)
# First try to construct new URL by replacing interpolation or literal version
NEW_URL="$CURRENT_URL"

# Handle Ruby interpolation #{version}
if [[ "$CURRENT_URL" =~ \#\{version\} ]]; then
    # URL uses #{version} interpolation - no need to modify the URL string itself
    # We'll just update the version line in the file
    CONSTRUCTED_URL="${CURRENT_URL//\#\{version\}/$NEW_VERSION}"
    info "URL uses version interpolation: $CURRENT_URL"
    info "Constructed download URL: $CONSTRUCTED_URL"
    NEW_URL="$CURRENT_URL"  # Keep original with interpolation
    DOWNLOAD_URL="$CONSTRUCTED_URL"
else
    # URL has literal version - replace it
    NEW_URL="${CURRENT_URL/$CURRENT_VERSION/$NEW_VERSION}"
    DOWNLOAD_URL="$NEW_URL"
    
    if [ "$NEW_URL" = "$CURRENT_URL" ]; then
        error "Could not substitute version in URL. URL may not contain version string.\n  Current URL: $CURRENT_URL\n  Looking for: $CURRENT_VERSION"
    fi
    
    info "Updated URL: $NEW_URL"
fi

# Verify download URL is accessible
info "Verifying download URL is accessible..."
if ! curl -fsSL --head "$DOWNLOAD_URL" > /dev/null 2>&1; then
    error "Download URL is not accessible: $DOWNLOAD_URL\n  Verify the version exists and the URL pattern is correct"
fi

success "Download URL is accessible"

# Download new tarball/asset and calculate SHA256
info "Downloading asset to calculate SHA256..."
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

ASSET_PATH="$TEMP_DIR/$PACKAGE_NAME"
if ! curl -fsSL "$DOWNLOAD_URL" -o "$ASSET_PATH" 2>/dev/null; then
    error "Failed to download asset from $DOWNLOAD_URL"
fi

NEW_SHA256=$(sha256sum "$ASSET_PATH" | awk '{print $1}')
success "New SHA256: $NEW_SHA256"

# Extract current SHA256 for comparison
CURRENT_SHA256=$(grep -E '^\s*sha256\s+' "$PACKAGE_FILE" | head -n 1 | sed -E 's/.*sha256\s+"([a-f0-9]+)".*/\1/')

if [ -z "$CURRENT_SHA256" ]; then
    error "Could not extract current SHA256 from $PACKAGE_FILE"
fi

info "Current SHA256: $CURRENT_SHA256"

# Check if SHA256 is the same
if [ "$CURRENT_SHA256" = "$NEW_SHA256" ]; then
    warn "New SHA256 is the same as current SHA256 (no changes in asset)"
fi

# Create backup
BACKUP_FILE="$PACKAGE_FILE.backup"
cp "$PACKAGE_FILE" "$BACKUP_FILE"
success "Created backup: $BACKUP_FILE"

# Update version line
# For formulas/casks with version line
if grep -qE '^\s*version\s+"' "$PACKAGE_FILE"; then
    info "Updating version line..."
    sed -i "s|^\(\s*version\s*\)\"$CURRENT_VERSION\"|\\1\"$NEW_VERSION\"|" "$PACKAGE_FILE"
    success "Updated version: $CURRENT_VERSION → $NEW_VERSION"
fi

# Update URL line (only if it contains literal version, not interpolation)
if [[ ! "$CURRENT_URL" =~ \#\{version\} ]]; then
    info "Updating URL line..."
    # Escape special characters for sed
    ESCAPED_CURRENT_URL=$(printf '%s\n' "$CURRENT_URL" | sed 's:[][\/.^$*]:\\&:g')
    ESCAPED_NEW_URL=$(printf '%s\n' "$NEW_URL" | sed 's:[][\\/.$*]:\\&:g')
    sed -i "s|^\(\s*url\s*\)\"$ESCAPED_CURRENT_URL\"|\\1\"$ESCAPED_NEW_URL\"|" "$PACKAGE_FILE"
    success "Updated URL"
fi

# Update SHA256 line
info "Updating SHA256 line..."
sed -i "s|^\(\s*sha256\s*\)\"$CURRENT_SHA256\"|\\1\"$NEW_SHA256\"|" "$PACKAGE_FILE"
success "Updated SHA256"

# Validate Ruby syntax
info "Validating Ruby syntax..."
if ! ruby -c "$PACKAGE_FILE" > /dev/null 2>&1; then
    error "Ruby syntax validation failed. Restoring backup...\n  $(mv "$BACKUP_FILE" "$PACKAGE_FILE" && echo "Backup restored")"
fi

success "Ruby syntax is valid"

# Remove backup after successful update
rm "$BACKUP_FILE"

# Display summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}Update completed successfully!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${BLUE}Package:${NC}     $PACKAGE_NAME ($PACKAGE_TYPE)"
echo -e "${BLUE}File:${NC}        $PACKAGE_FILE"
echo ""
echo -e "${BLUE}Version:${NC}"
echo "  Old: $CURRENT_VERSION"
echo "  New: $NEW_VERSION"
echo ""
echo -e "${BLUE}SHA256:${NC}"
echo "  Old: $CURRENT_SHA256"
echo "  New: $NEW_SHA256"
echo ""
if [[ ! "$CURRENT_URL" =~ \#\{version\} ]]; then
    echo -e "${BLUE}URL:${NC}"
    echo "  Old: $CURRENT_URL"
    echo "  New: $NEW_URL"
    echo ""
fi
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Review the changes: git diff $PACKAGE_FILE"
echo "  2. Test the $PACKAGE_TYPE: brew install --build-from-source $PACKAGE_FILE"
echo "  3. Validate: brew audit --strict $PACKAGE_FILE"
echo "  4. Commit the changes: git add $PACKAGE_FILE && git commit -m \"chore: update $PACKAGE_NAME to $NEW_VERSION\""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
