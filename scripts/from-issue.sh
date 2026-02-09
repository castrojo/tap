#!/usr/bin/env bash
#
# from-issue.sh - Automate package creation from GitHub issues
#
# Usage: ./from-issue.sh <issue-number> [--create-pr]
# Example: ./from-issue.sh 42
# Example: ./from-issue.sh 42 --create-pr
#
# This script:
# 1. Fetches GitHub issue details
# 2. Extracts repository URL and description
# 3. Derives package name from the repository name
# 4. Detects if it should be a formula (CLI) or cask (GUI)
# 5. Creates a git branch
# 6. Calls the appropriate helper script
# 7. Commits the generated formula/cask
# 8. Optionally creates a PR and comments on the issue

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
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

highlight() {
    echo -e "${CYAN}$1${NC}"
}

section() {
    echo ""
    echo -e "${MAGENTA}━━━ $1 ━━━${NC}"
}

# Validate inputs
if [ $# -lt 1 ]; then
    error "Usage: $0 <issue-number> [--create-pr]\n  Example: $0 42\n  Example: $0 42 --create-pr"
fi

ISSUE_NUMBER="$1"
CREATE_PR=false

# Parse optional flags
shift
while [ $# -gt 0 ]; do
    case "$1" in
        --create-pr)
            CREATE_PR=true
            shift
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

# Validate issue number
if ! [[ "$ISSUE_NUMBER" =~ ^[0-9]+$ ]]; then
    error "Issue number must be a positive integer"
fi

section "Preflight Checks"

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    error "GitHub CLI (gh) is not installed. Install it with: brew install gh"
fi
success "GitHub CLI found"

# Check if gh is authenticated
if ! gh auth status &> /dev/null; then
    error "GitHub CLI is not authenticated. Run: gh auth login"
fi
success "GitHub CLI authenticated"

# Check if jq is available
if ! command -v jq &> /dev/null; then
    error "jq is not installed. Install it with: brew install jq"
fi
success "jq found"

# Check if we're in a git repository
if ! git rev-parse --git-dir &> /dev/null; then
    error "Not in a git repository"
fi
success "Git repository detected"

# Determine the GitHub repository from git remote
REPO_URL=$(git config --get remote.origin.url || echo "")
if [ -z "$REPO_URL" ]; then
    error "Could not determine GitHub repository from git remote"
fi

# Extract owner/repo from git remote URL
if [[ "$REPO_URL" =~ github\.com[:/]([^/]+)/([^/\.]+) ]]; then
    GITHUB_OWNER="${BASH_REMATCH[1]}"
    GITHUB_REPO="${BASH_REMATCH[2]}"
else
    error "Could not parse GitHub owner/repo from remote URL: $REPO_URL"
fi

success "Repository: $GITHUB_OWNER/$GITHUB_REPO"

section "Fetching Issue #$ISSUE_NUMBER"

# Fetch issue details
info "Fetching issue data..."
ISSUE_DATA=$(gh api "repos/$GITHUB_OWNER/$GITHUB_REPO/issues/$ISSUE_NUMBER" 2>/dev/null) || error "Failed to fetch issue #$ISSUE_NUMBER. Check if it exists."

ISSUE_TITLE=$(echo "$ISSUE_DATA" | jq -r '.title')
ISSUE_BODY=$(echo "$ISSUE_DATA" | jq -r '.body // ""')
ISSUE_STATE=$(echo "$ISSUE_DATA" | jq -r '.state')
ISSUE_URL=$(echo "$ISSUE_DATA" | jq -r '.html_url')

success "Issue: $ISSUE_TITLE"
info "State: $ISSUE_STATE"
info "URL: $ISSUE_URL"

if [ "$ISSUE_STATE" = "closed" ]; then
    warn "Issue is already closed. Continuing anyway..."
fi

section "Parsing Issue Template"

# Parse issue body for required fields
# Expected format (from issue template):
# ### Repository or Homepage URL
# https://github.com/owner/repo
#
# ### Description
# A description of the package
#
# NOTE: Package name is derived from the repository name

# Extract Repository URL
REPO_URL=$(echo "$ISSUE_BODY" | sed -n '/###.*\([Rr]epository\|[Uu][Rr][Ll]\|[Hh]omepage\)/,/###/p' | sed '1d;$d' | grep -v '^$' | head -n1 | xargs)
if [ -z "$REPO_URL" ]; then
    error "Could not find 'Repository URL' or 'Homepage URL' in issue body. Please ensure the issue follows the template."
fi
success "Repository URL: $REPO_URL"

# Validate it's a GitHub URL
if ! [[ "$REPO_URL" =~ github\.com ]]; then
    error "Repository URL must be a GitHub URL: $REPO_URL"
fi

# Extract owner and repo from URL to derive package name
if [[ "$REPO_URL" =~ github\.com[:/]([^/]+)/([^/\.]+) ]]; then
    PKG_OWNER="${BASH_REMATCH[1]}"
    PKG_REPO="${BASH_REMATCH[2]}"
else
    error "Invalid GitHub URL format: $REPO_URL"
fi

# Derive package name from repository name (normalize to lowercase, replace underscores with hyphens)
PACKAGE_NAME=$(echo "$PKG_REPO" | tr '[:upper:]' '[:lower:]' | tr '_' '-')
success "Package Name (derived from repo): $PACKAGE_NAME"

# Extract Description (optional, can be empty)
DESCRIPTION=$(echo "$ISSUE_BODY" | sed -n '/###.*[Dd]escription/,/###/p' | sed '1d;$d' | grep -v '^$' | head -n1 | xargs)
if [ -z "$DESCRIPTION" ]; then
    warn "No description found in issue body. Will use repository description."
fi

# Extract Package Type hint (optional)
PACKAGE_TYPE=$(echo "$ISSUE_BODY" | sed -n '/###.*[Pp]ackage [Tt]ype/,/###/p' | sed '1d;$d' | grep -v '^$' | head -n1 | xargs | tr '[:upper:]' '[:lower:]')

section "Detecting Package Type"

# Determine if it should be a formula or cask
# Priority:
# 1. Explicit package type from issue
# 2. Analyze repository metadata (topics, keywords)
# 3. Default to formula

DETECTED_TYPE=""

if [ -n "$PACKAGE_TYPE" ]; then
    case "$PACKAGE_TYPE" in
        formula|cli|command-line)
            DETECTED_TYPE="formula"
            info "Type specified in issue: formula (CLI)"
            ;;
        cask|gui|app|application)
            DETECTED_TYPE="cask"
            info "Type specified in issue: cask (GUI)"
            ;;
        *)
            warn "Unknown package type in issue: $PACKAGE_TYPE. Will auto-detect."
            ;;
    esac
fi

# If not explicitly specified, analyze the repository
if [ -z "$DETECTED_TYPE" ]; then
    info "Auto-detecting package type from repository metadata..."
    
    # Fetch repository metadata (we already have PKG_OWNER and PKG_REPO from earlier)
    PKG_DATA=$(gh api "repos/$PKG_OWNER/$PKG_REPO" 2>/dev/null) || error "Failed to fetch repository metadata for $REPO_URL"
    
    TOPICS=$(echo "$PKG_DATA" | jq -r '.topics[]?' 2>/dev/null || echo "")
    PKG_DESCRIPTION=$(echo "$PKG_DATA" | jq -r '.description // ""')
    
    # Check for GUI/application indicators
    GUI_INDICATORS="gui|desktop|application|app|electron|tauri|qt|gtk|macos-app"
    CLI_INDICATORS="cli|command-line|terminal|shell|tool|utility"
    
    if echo "$TOPICS $PKG_DESCRIPTION" | grep -iEq "$GUI_INDICATORS"; then
        DETECTED_TYPE="cask"
        info "Detected GUI application (cask) based on: topics/description"
    elif echo "$TOPICS $PKG_DESCRIPTION" | grep -iEq "$CLI_INDICATORS"; then
        DETECTED_TYPE="formula"
        info "Detected CLI tool (formula) based on: topics/description"
    else
        # Default to formula (most common case)
        DETECTED_TYPE="formula"
        info "No clear indicators found. Defaulting to: formula"
    fi
fi

success "Package type: $DETECTED_TYPE"

section "Creating Git Branch"

# Normalize package name for branch (lowercase, alphanumeric, hyphens only)
BRANCH_NAME="package-request-${ISSUE_NUMBER}-${PACKAGE_NAME}"
BRANCH_NAME=$(echo "$BRANCH_NAME" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]/-/g')

# Check if branch already exists
if git rev-parse --verify "$BRANCH_NAME" &> /dev/null; then
    warn "Branch $BRANCH_NAME already exists"
    info "Checking out existing branch..."
    git checkout "$BRANCH_NAME" || error "Failed to checkout existing branch"
else
    info "Creating branch: $BRANCH_NAME"
    git checkout -b "$BRANCH_NAME" || error "Failed to create branch"
fi

success "On branch: $BRANCH_NAME"

section "Generating Package"

# Call the appropriate helper script
SCRIPT_DIR="$(dirname "$0")"
TARGET_FILE=""

if [ "$DETECTED_TYPE" = "formula" ]; then
    info "Calling new-formula.sh..."
    "$SCRIPT_DIR/new-formula.sh" "$PACKAGE_NAME" "$REPO_URL" || error "Failed to generate formula"
    TARGET_FILE="Formula/$PACKAGE_NAME.rb"
elif [ "$DETECTED_TYPE" = "cask" ]; then
    info "Calling new-cask.sh..."
    "$SCRIPT_DIR/new-cask.sh" "$PACKAGE_NAME" "$REPO_URL" || error "Failed to generate cask"
    TARGET_FILE="Casks/$PACKAGE_NAME.rb"
else
    error "Unknown package type: $DETECTED_TYPE"
fi

success "Package generated successfully"

section "Committing Changes"

# Stage the generated file
info "Staging $TARGET_FILE..."
git add "$TARGET_FILE" || error "Failed to stage $TARGET_FILE"

# Create commit with reference to issue
COMMIT_MSG="feat: add $PACKAGE_NAME $DETECTED_TYPE (closes #$ISSUE_NUMBER)"
info "Creating commit: $COMMIT_MSG"
git commit -m "$COMMIT_MSG" || error "Failed to commit"

success "Changes committed"

section "Pushing to Remote"

# Push the branch
info "Pushing branch to remote..."
git push -u origin "$BRANCH_NAME" || error "Failed to push branch"

success "Branch pushed to origin/$BRANCH_NAME"

section "Summary"

echo ""
highlight "Package Details:"
echo "  Name:        $PACKAGE_NAME"
echo "  Type:        $DETECTED_TYPE"
echo "  Repository:  $REPO_URL"
echo "  File:        $TARGET_FILE"
echo ""
highlight "Git Details:"
echo "  Branch:      $BRANCH_NAME"
echo "  Commit:      $COMMIT_MSG"
echo ""

# Create PR if requested
if [ "$CREATE_PR" = true ]; then
    section "Creating Pull Request"
    
    PR_TITLE="Add $PACKAGE_NAME $DETECTED_TYPE"
    PR_BODY="## Summary

This PR adds the \`$PACKAGE_NAME\` $DETECTED_TYPE to the tap.

**Package Information:**
- Name: \`$PACKAGE_NAME\`
- Type: $DETECTED_TYPE
- Repository: $REPO_URL
- Source Issue: #$ISSUE_NUMBER

**Generated by:** \`from-issue.sh\`

Closes #$ISSUE_NUMBER"
    
    info "Creating pull request..."
    PR_URL=$(gh pr create --title "$PR_TITLE" --body "$PR_BODY" 2>&1) || error "Failed to create pull request"
    
    success "Pull request created: $PR_URL"
    
    # Comment on the original issue
    info "Commenting on issue #$ISSUE_NUMBER..."
    COMMENT_BODY="✅ Package $DETECTED_TYPE has been generated and a pull request has been created: $PR_URL

The $DETECTED_TYPE will be available once the PR is reviewed and merged."
    
    gh issue comment "$ISSUE_NUMBER" --body "$COMMENT_BODY" || warn "Failed to comment on issue"
    
    echo ""
    highlight "Next Steps:"
    echo "  1. Review the PR: $PR_URL"
    echo "  2. Test the $DETECTED_TYPE locally"
    echo "  3. Merge the PR to publish the package"
else
    echo ""
    highlight "Next Steps:"
    echo "  1. Review the generated $DETECTED_TYPE: $TARGET_FILE"
    echo "  2. Test locally: brew install $([ "$DETECTED_TYPE" = "cask" ] && echo "--cask" || echo "") $TARGET_FILE"
    echo "  3. Create a PR manually: gh pr create --fill"
    echo "  4. Or run with --create-pr flag: $0 $ISSUE_NUMBER --create-pr"
fi

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Automation complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
