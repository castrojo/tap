#!/usr/bin/env bash
#
# new-formula.sh - Generate Homebrew formula from GitHub repository
#
# Usage: ./new-formula.sh <package-name> <github-repo-url>
# Example: ./new-formula.sh myapp https://github.com/user/myapp

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
    error "Usage: $0 <package-name> <github-repo-url>\n  Example: $0 myapp https://github.com/user/myapp"
fi

PACKAGE_NAME="$1"
GITHUB_URL="$2"

# Validate package name (lowercase, alphanumeric, hyphens)
if ! [[ "$PACKAGE_NAME" =~ ^[a-z0-9][a-z0-9-]*$ ]]; then
    error "Package name must be lowercase, start with alphanumeric, and contain only alphanumeric characters and hyphens"
fi

# Extract owner and repo from GitHub URL
if [[ "$GITHUB_URL" =~ github\.com[:/]([^/]+)/([^/\.]+) ]]; then
    OWNER="${BASH_REMATCH[1]}"
    REPO="${BASH_REMATCH[2]}"
else
    error "Invalid GitHub URL. Expected format: https://github.com/owner/repo or git@github.com:owner/repo"
fi

info "Generating formula for $PACKAGE_NAME from $OWNER/$REPO"

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

DESCRIPTION=$(echo "$REPO_DATA" | jq -r '.description // "A command-line tool"')
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

# Find tarball asset or construct URL
info "Detecting tarball URL..."
TARBALL_URL=$(echo "$RELEASE_DATA" | jq -r '.tarball_url')

if [ -z "$TARBALL_URL" ] || [ "$TARBALL_URL" = "null" ]; then
    # Fallback to standard GitHub archive URL
    TARBALL_URL="https://github.com/$OWNER/$REPO/archive/refs/tags/$RELEASE_TAG.tar.gz"
    warn "Using fallback tarball URL: $TARBALL_URL"
else
    success "Tarball URL: $TARBALL_URL"
fi

# Download tarball and calculate SHA256
info "Downloading tarball to calculate SHA256..."
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

TARBALL_PATH="$TEMP_DIR/$PACKAGE_NAME.tar.gz"
if ! curl -fsSL "$TARBALL_URL" -o "$TARBALL_PATH" 2>/dev/null; then
    error "Failed to download tarball from $TARBALL_URL"
fi

SHA256=$(sha256sum "$TARBALL_PATH" | awk '{print $1}')
success "SHA256: $SHA256"

# Determine formula path
FORMULA_DIR="$(dirname "$0")/../Formula"
FORMULA_PATH="$FORMULA_DIR/$PACKAGE_NAME.rb"

if [ -f "$FORMULA_PATH" ]; then
    error "Formula already exists at $FORMULA_PATH. Remove it first or use a different name."
fi

# Generate formula class name (capitalize first letter, convert hyphens to underscores, camelcase)
# Example: my-app -> MyApp, foo-bar-baz -> FooBarBaz
CLASS_NAME=$(echo "$PACKAGE_NAME" | sed -E 's/(^|-)([a-z])/\U\2/g')

# Generate the formula
info "Generating formula at $FORMULA_PATH..."

# Build license line if available
LICENSE_LINE=""
if [ -n "$LICENSE" ] && [ "$LICENSE" != "null" ]; then
    LICENSE_LINE="  license \"$LICENSE\""
fi

cat > "$FORMULA_PATH" << EOF
class $CLASS_NAME < Formula
  desc "$DESCRIPTION"
  homepage "$HOMEPAGE"
  url "$TARBALL_URL"
  sha256 "$SHA256"
$LICENSE_LINE

  def install
    bin.install "$PACKAGE_NAME"
  end

  test do
    # Test that the binary exists and is executable
    assert_predicate bin/"$PACKAGE_NAME", :exist?
    assert_predicate bin/"$PACKAGE_NAME", :executable?

    # Try running with --version or --help
    begin
      output = shell_output("#{bin}/$PACKAGE_NAME --version 2>&1", 0)
      assert_match "$VERSION", output
    rescue
      # If --version doesn't work, try --help
      system "#{bin}/$PACKAGE_NAME", "--help"
    end
  end
end
EOF

success "Formula created at $FORMULA_PATH"

# Display summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}Formula generated successfully!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${BLUE}Package:${NC}     $PACKAGE_NAME"
echo -e "${BLUE}Version:${NC}     $VERSION"
echo -e "${BLUE}Formula:${NC}     $FORMULA_PATH"
echo -e "${BLUE}Repository:${NC}  $OWNER/$REPO"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Review and customize the formula: $FORMULA_PATH"
echo "  2. Install dependencies if needed (add 'depends_on' directives)"
echo "  3. Adjust the install block based on the project's build system"
echo "  4. Test the formula: brew install --build-from-source $FORMULA_PATH"
echo "  5. Commit the formula: git add $FORMULA_PATH && git commit"
echo ""
echo -e "${YELLOW}Note:${NC} The generated formula assumes a simple binary installation."
echo "      You may need to adjust the install block for projects that require"
echo "      compilation, configuration, or have different installation needs."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
