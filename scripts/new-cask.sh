#!/usr/bin/env bash
#
# new-cask.sh - Generate Homebrew cask from GitHub repository
#
# Usage: ./new-cask.sh <cask-name> <github-repo-url>
# Example: ./new-cask.sh myapp https://github.com/user/myapp

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
    error "Usage: $0 <cask-name> <github-repo-url>\n  Example: $0 myapp https://github.com/user/myapp"
fi

CASK_NAME="$1"
GITHUB_URL="$2"

# Validate cask name (lowercase, alphanumeric, hyphens)
if ! [[ "$CASK_NAME" =~ ^[a-z0-9][a-z0-9-]*$ ]]; then
    error "Cask name must be lowercase, start with alphanumeric, and contain only alphanumeric characters and hyphens"
fi

# Extract owner and repo from GitHub URL
if [[ "$GITHUB_URL" =~ github\.com[:/]([^/]+)/([^/\.]+) ]]; then
    OWNER="${BASH_REMATCH[1]}"
    REPO="${BASH_REMATCH[2]}"
else
    error "Invalid GitHub URL. Expected format: https://github.com/owner/repo or git@github.com:owner/repo"
fi

info "Generating cask for $CASK_NAME from $OWNER/$REPO"

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    error "GitHub CLI (gh) is not installed. Install it with: brew install gh"
fi

# Check if gh is authenticated
if ! gh auth status &> /dev/null; then
    error "GitHub CLI is not authenticated. Run: gh auth login"
fi

# Check if jq is available
if ! command -v jq &> /dev/null; then
    error "jq is not installed. Install it with: brew install jq"
fi

# Fetch repository metadata
info "Fetching repository metadata..."
REPO_DATA=$(gh api "repos/$OWNER/$REPO" 2>/dev/null) || error "Failed to fetch repository metadata. Check if repository exists and is accessible."

DESCRIPTION=$(echo "$REPO_DATA" | jq -r '.description // "A GUI application"')
HOMEPAGE=$(echo "$REPO_DATA" | jq -r '.html_url')
LICENSE=$(echo "$REPO_DATA" | jq -r '.license.spdx_id // empty')

success "Repository: $OWNER/$REPO"
info "Description: $DESCRIPTION"
info "License: $LICENSE"

# Fetch latest release
info "Fetching latest release..."
RELEASE_DATA=$(gh api "repos/$OWNER/$REPO/releases/latest" 2>/dev/null) || error "No releases found. Repository must have at least one release."

VERSION=$(echo "$RELEASE_DATA" | jq -r '.tag_name' | sed 's/^v//')
RELEASE_TAG=$(echo "$RELEASE_DATA" | jq -r '.tag_name')

if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
    error "Could not determine version from latest release"
fi

success "Latest version: $VERSION (tag: $RELEASE_TAG)"

# Find binary asset (tar.gz or .zip)
info "Detecting binary assets..."
ASSETS=$(echo "$RELEASE_DATA" | jq -r '.assets')

# Try to find tar.gz first
ASSET_URL=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("\\.(tar\\.gz|tgz)$"; "i")) | .browser_download_url' | head -n1)
ASSET_TYPE="tarball"
ASSET_NAME=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("\\.(tar\\.gz|tgz)$"; "i")) | .name' | head -n1)

# If no tarball, try .zip
if [ -z "$ASSET_URL" ] || [ "$ASSET_URL" = "null" ]; then
    ASSET_URL=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("\\.zip$"; "i")) | .browser_download_url' | head -n1)
    ASSET_TYPE="zip"
    ASSET_NAME=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("\\.zip$"; "i")) | .name' | head -n1)
fi

# If still nothing found, show available assets and error
if [ -z "$ASSET_URL" ] || [ "$ASSET_URL" = "null" ]; then
    warn "Available assets in latest release:"
    echo "$ASSETS" | jq -r '.[].name' | sed 's/^/  - /'
    error "No suitable binary asset found. Looking for: .tar.gz, .tgz, or .zip"
fi

success "Found $ASSET_TYPE: $ASSET_NAME"

# Download asset and calculate SHA256
info "Downloading asset to calculate SHA256..."
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

ASSET_FILENAME=$(basename "$ASSET_URL")
ASSET_PATH="$TEMP_DIR/$ASSET_FILENAME"
if ! curl -fsSL "$ASSET_URL" -o "$ASSET_PATH" 2>/dev/null; then
    error "Failed to download asset from $ASSET_URL"
fi

SHA256=$(sha256sum "$ASSET_PATH" | awk '{print $1}')
success "SHA256: $SHA256"

# Determine cask path
CASK_DIR="$(dirname "$0")/../Casks"
CASK_PATH="$CASK_DIR/$CASK_NAME.rb"

if [ -f "$CASK_PATH" ]; then
    error "Cask already exists at $CASK_PATH. Remove it first or use a different name."
fi

# Generate cask class name (capitalize first letter, convert hyphens to underscores, camelcase)
# Example: my-app -> MyApp, foo-bar-baz -> FooBarBaz
CLASS_NAME=$(echo "$CASK_NAME" | sed -E 's/(^|-)([a-z])/\U\2/g')

# Generate the cask
info "Generating cask at $CASK_PATH..."

# Build license line if available
LICENSE_LINE=""
if [ -n "$LICENSE" ] && [ "$LICENSE" != "null" ]; then
    LICENSE_LINE="  license \"$LICENSE\""
fi

# Build version line
VERSION_LINE="  version \"$VERSION\""

# Build URL and SHA256 lines
URL_LINE="  url \"$ASSET_URL\""
SHA_LINE="  sha256 \"$SHA256\""

# Generate binary stanza (proper Cask DSL)
# Note: User will need to customize based on actual binary name in archive
BINARY_STANZA="  binary \"$CASK_NAME\""

# Generate test stanza
TEST_STANZA="    # Test that the binary exists and is executable
    assert_predicate bin/\"$CASK_NAME\", :exist?
    assert_predicate bin/\"$CASK_NAME\", :executable?
    
    # Try running with --version
    system bin/\"$CASK_NAME\", \"--version\""

cat > "$CASK_PATH" << EOF
cask "$CASK_NAME" do
$VERSION_LINE
$SHA_LINE
$URL_LINE
  
  name "$CLASS_NAME"
  desc "$DESCRIPTION"
  homepage "$HOMEPAGE"
$LICENSE_LINE

$BINARY_STANZA

  test do
$TEST_STANZA
  end
end
EOF

success "Cask created at $CASK_PATH"

# Display summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}Cask generated successfully!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${BLUE}Package:${NC}     $CASK_NAME"
echo -e "${BLUE}Version:${NC}     $VERSION"
echo -e "${BLUE}Asset Type:${NC}  $ASSET_TYPE"
echo -e "${BLUE}Cask:${NC}        $CASK_PATH"
echo -e "${BLUE}Repository:${NC}  $OWNER/$REPO"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Review and customize the cask: $CASK_PATH"
echo "  2. Adjust the binary name based on actual archive contents"
echo "  3. Extract the archive to find the actual binary name"
echo "  4. Test the cask: brew install --cask $CASK_PATH"
echo "  5. Commit the cask: git add $CASK_PATH && git commit"
echo ""
echo -e "${YELLOW}Note:${NC} The generated cask uses the Homebrew Cask DSL with a 'binary' stanza."
echo "      You MUST customize the binary name to match the actual binary in the archive."
echo "      Extract $ASSET_NAME to find the correct binary path."
echo "      Example: binary \"path/to/actual-binary-name\""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
