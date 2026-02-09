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

# Find binary asset - LINUX ONLY - PRIORITY: Tarball > .deb
info "Detecting Linux binary assets (prioritizing tarballs)..."
ASSETS=$(echo "$RELEASE_DATA" | jq -r '.assets')

# FIRST PRIORITY: Linux tarballs (.tar.gz, .tar.xz, .tgz)
info "Searching for Linux tarballs (preferred format)..."
ASSET_URL=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("(linux|amd64|x86_64).*\\.(tar\\.gz|tgz|tar\\.xz)$"; "i")) | .browser_download_url' | head -n1)
ASSET_TYPE="tarball"
ASSET_NAME=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("(linux|amd64|x86_64).*\\.(tar\\.gz|tgz|tar\\.xz)$"; "i")) | .name' | head -n1)

# SECOND PRIORITY: Debian packages (.deb) - only if no tarball found
if [ -z "$ASSET_URL" ] || [ "$ASSET_URL" = "null" ]; then
    info "No tarball found, searching for Debian packages (.deb)..."
    ASSET_URL=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("(linux|amd64|x86_64).*\\.deb$"; "i")) | .browser_download_url' | head -n1)
    ASSET_TYPE="deb"
    ASSET_NAME=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("(linux|amd64|x86_64).*\\.deb$"; "i")) | .name' | head -n1)
    
    if [ -n "$ASSET_URL" ] && [ "$ASSET_URL" != "null" ]; then
        warn "Using .deb package (second choice - prefer tarballs when available)"
    fi
fi

# FALLBACK: Generic tarballs without Linux marker (warn user)
if [ -z "$ASSET_URL" ] || [ "$ASSET_URL" = "null" ]; then
    warn "No Linux-specific tarball found, trying generic tarballs..."
    ASSET_URL=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("\\.(tar\\.gz|tgz|tar\\.xz)$"; "i")) | .browser_download_url' | head -n1)
    ASSET_NAME=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("\\.(tar\\.gz|tgz|tar\\.xz)$"; "i")) | .name' | head -n1)
    ASSET_TYPE="tarball"
    
    if [ -n "$ASSET_URL" ] && [ "$ASSET_URL" != "null" ]; then
        warn "Found generic tarball without Linux marker: $ASSET_NAME"
        warn "⚠️  VERIFY THIS IS A LINUX BINARY BEFORE PROCEEDING!"
    fi
fi

# Check for macOS-specific patterns and REJECT them
MACOS_ASSET=$(echo "$ASSETS" | jq -r '.[] | select(.name | test("(macos|darwin|osx|\\.dmg|\\.pkg)"; "i")) | .name' | head -n1)
if [ -n "$MACOS_ASSET" ] && [ "$MACOS_ASSET" != "null" ]; then
    error "Found macOS asset: $MACOS_ASSET\n  This tap is LINUX ONLY. Do not use macOS downloads.\n  Look for Linux-specific assets (linux, amd64, x86_64)"
fi

# If still nothing found, show available assets and error
if [ -z "$ASSET_URL" ] || [ "$ASSET_URL" = "null" ]; then
    warn "Available assets in latest release:"
    echo "$ASSETS" | jq -r '.[].name' | sed 's/^/  - /'
    error "No suitable Linux binary asset found.\n  This tap is LINUX ONLY.\n  \n  PRIORITY ORDER:\n    1. Tarballs: *linux*.tar.gz, *amd64*.tar.xz, *x86_64*.tgz (PREFERRED)\n    2. Debian: *linux*.deb, *amd64*.deb (second choice)\n  \n  NOT ACCEPTABLE: *macos*, *darwin*, *.dmg, *.pkg, *.exe, *.msi"
fi

success "Found Linux $ASSET_TYPE: $ASSET_NAME"

# Download asset and calculate SHA256 (MANDATORY)
info "Downloading asset to calculate SHA256 (required for all packages)..."
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

ASSET_FILENAME=$(basename "$ASSET_URL")
ASSET_PATH="$TEMP_DIR/$ASSET_FILENAME"
if ! curl -fsSL "$ASSET_URL" -o "$ASSET_PATH" 2>/dev/null; then
    error "Failed to download asset from $ASSET_URL"
fi

SHA256=$(sha256sum "$ASSET_PATH" | awk '{print $1}')
success "SHA256 calculated: $SHA256"

# Try to find and verify against upstream checksums if available
info "Checking for upstream checksums..."
CHECKSUM_PATTERNS=("SHA256SUMS" "checksums.txt" "CHECKSUMS" "${ASSET_NAME}.sha256" "sha256sums.txt")
UPSTREAM_SHA256=""

for pattern in "${CHECKSUM_PATTERNS[@]}"; do
    CHECKSUM_URL=$(echo "$ASSETS" | jq -r ".[] | select(.name | test(\"$pattern\"; \"i\")) | .browser_download_url" | head -n1)
    if [ -n "$CHECKSUM_URL" ] && [ "$CHECKSUM_URL" != "null" ]; then
        info "Found checksum file: $pattern"
        CHECKSUM_FILE="$TEMP_DIR/checksums"
        if curl -fsSL "$CHECKSUM_URL" -o "$CHECKSUM_FILE" 2>/dev/null; then
            # Try to find our asset's checksum in the file
            UPSTREAM_SHA256=$(grep -i "$ASSET_NAME" "$CHECKSUM_FILE" | awk '{print $1}' | head -n1)
            if [ -n "$UPSTREAM_SHA256" ]; then
                info "Found upstream checksum for $ASSET_NAME"
                if [ "$SHA256" = "$UPSTREAM_SHA256" ]; then
                    success "✓ SHA256 verified against upstream checksum!"
                else
                    error "SHA256 MISMATCH!\n  Calculated: $SHA256\n  Upstream:   $UPSTREAM_SHA256\n  This indicates a corrupted download or compromised asset."
                fi
                break
            fi
        fi
    fi
done

if [ -z "$UPSTREAM_SHA256" ]; then
    warn "No upstream checksums found - using calculated SHA256"
    warn "Verify checksum manually if possible against official release notes"
fi

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

# Generate the cask following verified pattern from docs/CASK_CREATION_GUIDE.md
# Correct stanza order: version, sha256, [blank], url, name, desc, homepage, [blank], binary
# Note: NO test blocks (formulas only), NO depends_on :linux (invalid syntax)

cat > "$CASK_PATH" << EOF
cask "$CASK_NAME" do
  version "$VERSION"
  sha256 "$SHA256"

  url "$ASSET_URL"
  name "$CLASS_NAME"
  desc "$DESCRIPTION"
  homepage "$HOMEPAGE"

  binary "$CASK_NAME"
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
echo "  2. Extract the archive to find the actual binary path:"
echo "     tar -tzf $ASSET_NAME (for tar.gz) or unzip -l $ASSET_NAME (for zip)"
echo "  3. Update the binary stanza with the correct path from archive"
echo "     Example: binary \"app/bin/myapp\", target: \"myapp\""
echo "  4. Read docs/CASK_CREATION_GUIDE.md for critical cask rules"
echo "  5. Test the cask: brew install --cask --build-from-source $CASK_PATH"
echo "  6. Validate: brew audit --cask --strict castrojo/tap/$CASK_NAME"
echo "  7. Commit: git add $CASK_PATH && git commit -m \"feat(cask): add $CASK_NAME\""
echo ""
echo -e "${YELLOW}Important:${NC} The binary path is relative to the extracted archive root."
echo "      The generated cask follows the verified template from docs/CASK_CREATION_GUIDE.md"
echo "      - ✓ Correct stanza ordering (no blank lines within groups)"
echo "      - ✓ No test blocks (casks don't support them)"
echo "      - ✓ No depends_on :linux (invalid syntax for casks)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
