# Automated Homebrew Tap Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a fully automated personal Homebrew tap for Linux packages with Renovate-driven updates, comprehensive agent documentation, and quality gates.

**Architecture:** Standard Homebrew tap structure (Formula/ + Casks/) with GitHub Actions for CI/CD, Renovate for version updates, helper scripts for package creation, and comprehensive documentation enabling agents to package software independently.

**Tech Stack:** Ruby (Homebrew DSL), Bash (helper scripts), GitHub Actions, Renovate Bot, GitHub CLI

---

## Task 1: Initialize Repository Structure

**Files:**
- Create: `.gitignore`
- Create: `README.md`
- Create: `LICENSE`
- Create: `.editorconfig`

**Step 1: Create .gitignore**

```bash
cat > .gitignore << 'EOF'
# macOS
.DS_Store

# Editor files
*.swp
*.swo
*~
.vscode/
.idea/

# Homebrew
*.bottle*.tar.gz
EOF
```

**Step 2: Create LICENSE (MIT)**

```bash
cat > LICENSE << 'EOF'
MIT License

Copyright (c) 2026 [Your Name]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF
```

**Step 3: Create .editorconfig**

```bash
cat > .editorconfig << 'EOF'
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.rb]
indent_style = space
indent_size = 2

[*.{yml,yaml}]
indent_style = space
indent_size = 2

[*.sh]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false
EOF
```

**Step 4: Create README.md**

```bash
cat > README.md << 'EOF'
# Personal Homebrew Tap

Automated Homebrew tap for Linux packages with intelligent updates and quality gates.

## Features

- 🤖 **Automated Updates**: Renovate checks every 3 hours, auto-merges patches
- ✅ **Quality Gates**: All packages pass `brew audit --strict` and `brew style`
- 📦 **Formulas & Casks**: CLI tools and GUI applications
- 🤝 **Agent-Friendly**: Comprehensive docs for AI-assisted package creation

## Installation

```bash
brew tap [username]/tap
```

## Usage

### Install a Package
```bash
brew install package-name           # Formula (CLI tool)
brew install --cask app-name        # Cask (GUI app)
```

### Request a Package

[Create an issue](../../issues/new/choose) with:
- Package name
- Repository or homepage URL
- Brief description

An agent will research and create the package automatically.

## For Package Maintainers

See [docs/AGENT_GUIDE.md](docs/AGENT_GUIDE.md) for comprehensive packaging instructions.

### Quick Start

```bash
# Create formula from GitHub repository
./scripts/new-formula.sh package-name https://github.com/user/repo

# Create cask from GitHub repository
./scripts/new-cask.sh app-name https://github.com/user/repo

# Process package request from issue
./scripts/from-issue.sh 42

# Validate all packages
./scripts/validate-all.sh
```

## Quality Standards

All packages must:
- Pass `brew audit --strict --online`
- Pass `brew style`
- Include valid SPDX license
- Have working URLs (HTTPS preferred)
- Include meaningful tests

## Update Strategy

- **Patch releases** (1.0.0 → 1.0.1): Auto-merge after 3 hours
- **Minor releases** (1.0.0 → 1.1.0): Auto-merge after 1 day
- **Major releases** (1.0.0 → 2.0.0): Manual review required

## License

MIT
EOF
```

**Step 5: Create directory structure**

```bash
mkdir -p Formula Casks docs scripts .github/workflows .github/ISSUE_TEMPLATE
```

**Step 6: Initialize git and commit**

```bash
git init
git add .gitignore LICENSE .editorconfig README.md
git commit -m "feat: initialize repository structure"
```

---

## Task 2: Create GitHub Issue Templates

**Files:**
- Create: `.github/ISSUE_TEMPLATE/config.yml`
- Create: `.github/ISSUE_TEMPLATE/01-package-request.yml`
- Create: `.github/ISSUE_TEMPLATE/02-bug-report.yml`

**Step 1: Create config.yml**

```bash
cat > .github/ISSUE_TEMPLATE/config.yml << 'EOF'
blank_issues_enabled: false
contact_links:
  - name: Homebrew Documentation
    url: https://docs.brew.sh
    about: Official Homebrew documentation
EOF
```

**Step 2: Create package request template**

```bash
cat > .github/ISSUE_TEMPLATE/01-package-request.yml << 'EOF'
name: 📦 Package Request
description: Request a new formula or cask
title: "Package: "
labels: ["package-request"]
assignees: []

body:
  - type: markdown
    attributes:
      value: |
        ## Package Request
        Provide basic information. An agent will research and create the formula/cask.

  - type: input
    id: package-name
    attributes:
      label: Package Name
      description: What should this package be called?
      placeholder: "e.g., jq, ripgrep, visual-studio-code"
    validations:
      required: true

  - type: input
    id: repository-url
    attributes:
      label: Repository or Homepage URL
      description: GitHub repository or official website
      placeholder: "https://github.com/user/project"
    validations:
      required: true

  - type: textarea
    id: description
    attributes:
      label: Description
      description: What does this package do?
      placeholder: "A command-line JSON processor..."
    validations:
      required: true

  - type: checkboxes
    id: confirmation
    attributes:
      label: Confirmation
      options:
        - label: This package is open source and freely distributable
          required: true
EOF
```

**Step 3: Create bug report template**

```bash
cat > .github/ISSUE_TEMPLATE/02-bug-report.yml << 'EOF'
name: 🐛 Bug Report
description: Report a problem with an existing package
title: "[Bug] "
labels: ["bug"]
assignees: []

body:
  - type: input
    id: package-name
    attributes:
      label: Package Name
      placeholder: "e.g., jq"
    validations:
      required: true

  - type: textarea
    id: problem
    attributes:
      label: What's wrong?
      placeholder: "Installation fails with error..."
    validations:
      required: true

  - type: textarea
    id: error-output
    attributes:
      label: Error Output (if applicable)
      render: shell
    validations:
      required: false
EOF
```

**Step 4: Commit issue templates**

```bash
git add .github/ISSUE_TEMPLATE/
git commit -m "feat: add GitHub issue templates"
```

---

## Task 3: Create GitHub Actions Workflows

**Files:**
- Create: `.github/workflows/tests.yml`
- Create: `.github/workflows/label.yml`

**Step 1: Create tests workflow**

```bash
cat > .github/workflows/tests.yml << 'EOF'
name: Tests

on:
  pull_request:
    paths:
      - 'Formula/**'
      - 'Casks/**'
  push:
    branches:
      - main
    paths:
      - 'Formula/**'
      - 'Casks/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Homebrew
        id: set-up-homebrew
        uses: Homebrew/actions/setup-homebrew@master

      - name: Cache Homebrew Bundler RubyGems
        uses: actions/cache@v4
        with:
          path: ${{ steps.set-up-homebrew.outputs.gems-path }}
          key: ${{ runner.os }}-rubygems-${{ steps.set-up-homebrew.outputs.gems-hash }}
          restore-keys: ${{ runner.os }}-rubygems-

      - name: Install Homebrew Bundler RubyGems
        run: brew install-bundler-gems

      - name: Get changed files
        id: changed-files
        uses: tj-actions/changed-files@v44
        with:
          files: |
            Formula/**/*.rb
            Casks/**/*.rb

      - name: Run brew audit on changed formulas
        if: steps.changed-files.outputs.any_changed == 'true'
        run: |
          for file in ${{ steps.changed-files.outputs.all_changed_files }}; do
            echo "Auditing $file"
            brew audit --strict --online "$file"
          done

      - name: Run brew style on changed files
        if: steps.changed-files.outputs.any_changed == 'true'
        run: |
          for file in ${{ steps.changed-files.outputs.all_changed_files }}; do
            echo "Checking style for $file"
            brew style "$file"
          done
EOF
```

**Step 2: Create label workflow**

```bash
cat > .github/workflows/label.yml << 'EOF'
name: Label PRs

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  label:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/labeler@v5
        with:
          configuration-path: .github/labeler.yml
          repo-token: ${{ secrets.GITHUB_TOKEN }}
EOF
```

**Step 3: Create labeler configuration**

```bash
cat > .github/labeler.yml << 'EOF'
formula:
  - changed-files:
    - any-glob-to-any-file: 'Formula/**'

cask:
  - changed-files:
    - any-glob-to-any-file: 'Casks/**'

documentation:
  - changed-files:
    - any-glob-to-any-file: 'docs/**'
    - any-glob-to-any-file: '*.md'

automation:
  - changed-files:
    - any-glob-to-any-file: '.github/**'
    - any-glob-to-any-file: 'scripts/**'
EOF
```

**Step 4: Commit workflows**

```bash
git add .github/workflows/ .github/labeler.yml
git commit -m "feat: add GitHub Actions workflows for testing and labeling"
```

---

## Task 4: Configure Renovate

**Files:**
- Create: `.github/renovate.json5`

**Step 1: Create Renovate configuration**

```bash
cat > .github/renovate.json5 << 'EOF'
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["config:base"],
  "schedule": ["every 3 hours"],
  "labels": ["dependencies"],
  "separateMinorPatch": true,
  "prConcurrentLimit": 5,
  "prHourlyLimit": 2,
  "packageRules": [
    {
      "description": "Auto-merge patch releases after 3 hours",
      "matchUpdateTypes": ["patch"],
      "matchDatasources": ["github-releases", "github-tags"],
      "automerge": true,
      "automergeType": "pr",
      "minimumReleaseAge": "3 hours",
      "labels": ["automerge", "patch-update"]
    },
    {
      "description": "Auto-merge minor releases after 1 day",
      "matchUpdateTypes": ["minor"],
      "matchDatasources": ["github-releases", "github-tags"],
      "automerge": true,
      "automergeType": "pr",
      "minimumReleaseAge": "1 day",
      "labels": ["automerge", "minor-update"]
    },
    {
      "description": "Major releases need manual review",
      "matchUpdateTypes": ["major"],
      "automerge": false,
      "labels": ["major-update", "needs-review"]
    },
    {
      "description": "Non-GitHub sources need manual review",
      "matchManagers": ["regex"],
      "automerge": false,
      "labels": ["non-github-source", "needs-review"]
    }
  ],
  "regexManagers": [
    {
      "description": "Detect versions in Formula files",
      "fileMatch": ["^Formula/.+\\.rb$"],
      "matchStrings": [
        "version \"(?<currentValue>.*?)\"\\n\\s+sha256 \"(?<currentDigest>.*?)\"\\n\\s+url \"(?<depName>https://github\\.com/[^/]+/[^/]+)/releases/download/v?(?<currentValue>[^/]+)/",
        "url \"(?<depName>https://github\\.com/[^/]+/[^/]+)/archive/v?(?<currentValue>.*?)\\.tar\\.gz\""
      ],
      "datasourceTemplate": "github-releases"
    },
    {
      "description": "Detect versions in Cask files",
      "fileMatch": ["^Casks/.+\\.rb$"],
      "matchStrings": [
        "version \"(?<currentValue>.*?)\"\\n\\s+sha256 \"(?<currentDigest>.*?)\"\\n\\s+url \"(?<depName>https://github\\.com/[^/]+/[^/]+)/releases/download/v?(?<currentValue>[^/]+)/",
        "url \"(?<depName>https://github\\.com/[^/]+/[^/]+)/releases/download/v?(?<currentValue>[^/]+)/"
      ],
      "datasourceTemplate": "github-releases"
    }
  ]
}
EOF
```

**Step 2: Commit Renovate configuration**

```bash
git add .github/renovate.json5
git commit -m "feat: configure Renovate for automated updates"
```

---

## Task 5: Create Helper Scripts - Part 1 (new-formula.sh)

**Files:**
- Create: `scripts/new-formula.sh`

**Step 1: Create new-formula.sh script**

```bash
cat > scripts/new-formula.sh << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

# Usage: ./scripts/new-formula.sh <name> <github-repo-url>
# Example: ./scripts/new-formula.sh jq https://github.com/jqlang/jq

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <name> <github-repo-url>"
  echo "Example: $0 jq https://github.com/jqlang/jq"
  exit 1
fi

NAME="$1"
REPO_URL="$2"

# Extract owner/repo from URL
if [[ "$REPO_URL" =~ github\.com/([^/]+)/([^/]+) ]]; then
  OWNER="${BASH_REMATCH[1]}"
  REPO="${BASH_REMATCH[2]}"
else
  echo "Error: Invalid GitHub URL format"
  exit 1
fi

echo "→ Fetching latest release for $OWNER/$REPO..."

# Get latest release info
RELEASE_JSON=$(gh api "repos/$OWNER/$REPO/releases/latest" 2>/dev/null || echo "{}")

if [ "$RELEASE_JSON" = "{}" ]; then
  echo "Error: Could not fetch release information. Is this a public repository with releases?"
  exit 1
fi

VERSION=$(echo "$RELEASE_JSON" | jq -r '.tag_name' | sed 's/^v//')
echo "→ Latest version: $VERSION"

# Get repository info for license and description
REPO_JSON=$(gh api "repos/$OWNER/$REPO")
LICENSE=$(echo "$REPO_JSON" | jq -r '.license.spdx_id // "NOASSERTION"')
DESC=$(echo "$REPO_JSON" | jq -r '.description // "No description"')

echo "→ License: $LICENSE"
echo "→ Description: $DESC"

# Find tarball in assets (prefer linux-x86_64 or linux-amd64)
ASSETS=$(echo "$RELEASE_JSON" | jq -r '.assets[] | select(.name | test("linux.*(x86_64|amd64|x64).*\\.(tar\\.gz|tar\\.xz|tgz)$")) | .browser_download_url' | head -1)

if [ -z "$ASSETS" ]; then
  # Fallback to source tarball
  ASSETS="https://github.com/$OWNER/$REPO/archive/v$VERSION.tar.gz"
  echo "→ No Linux binary found, using source tarball"
fi

echo "→ Download URL: $ASSETS"

# Download and calculate SHA256
echo "→ Downloading to calculate SHA256..."
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

curl -sSL "$ASSETS" -o "$TEMP_FILE"
SHA256=$(sha256sum "$TEMP_FILE" | awk '{print $1}')
echo "→ SHA256: $SHA256"

# Determine class name (capitalize first letter, handle hyphens)
CLASS_NAME=$(echo "$NAME" | sed -E 's/(^|-)([a-z])/\U\2/g' | sed 's/-//g')

# Create formula file
FORMULA_FILE="$REPO_ROOT/Formula/$NAME.rb"

cat > "$FORMULA_FILE" << FORMULA
class $CLASS_NAME < Formula
  desc "$DESC"
  homepage "https://github.com/$OWNER/$REPO"
  url "$ASSETS"
  sha256 "$SHA256"
  license "$LICENSE"

  def install
    # TODO: Adjust installation based on package contents
    # Common patterns:
    # bin.install "binary-name"
    # lib.install Dir["lib/*"]
    # For source builds: system "./configure", "--prefix=#{prefix}"
    #                    system "make", "install"
  end

  test do
    # TODO: Add meaningful test
    # system "#{bin}/$NAME", "--version"
  end
end
FORMULA

echo "✓ Created $FORMULA_FILE"
echo ""
echo "Next steps:"
echo "1. Edit $FORMULA_FILE to complete the install block"
echo "2. Test locally: HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source $NAME"
echo "3. Run audit: brew audit --strict --online $NAME"
echo "4. Commit and push"
EOF

chmod +x scripts/new-formula.sh
```

**Step 2: Commit new-formula.sh**

```bash
git add scripts/new-formula.sh
git commit -m "feat: add new-formula.sh helper script"
```

---

## Task 6: Create Helper Scripts - Part 2 (new-cask.sh)

**Files:**
- Create: `scripts/new-cask.sh`

**Step 1: Create new-cask.sh script**

```bash
cat > scripts/new-cask.sh << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

# Usage: ./scripts/new-cask.sh <name> <github-repo-url>
# Example: ./scripts/new-cask.sh visual-studio-code https://github.com/microsoft/vscode

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <name> <github-repo-url>"
  echo "Example: $0 visual-studio-code https://github.com/microsoft/vscode"
  exit 1
fi

NAME="$1"
REPO_URL="$2"

# Extract owner/repo from URL
if [[ "$REPO_URL" =~ github\.com/([^/]+)/([^/]+) ]]; then
  OWNER="${BASH_REMATCH[1]}"
  REPO="${BASH_REMATCH[2]}"
else
  echo "Error: Invalid GitHub URL format"
  exit 1
fi

echo "→ Fetching latest release for $OWNER/$REPO..."

# Get latest release info
RELEASE_JSON=$(gh api "repos/$OWNER/$REPO/releases/latest" 2>/dev/null || echo "{}")

if [ "$RELEASE_JSON" = "{}" ]; then
  echo "Error: Could not fetch release information. Is this a public repository with releases?"
  exit 1
fi

VERSION=$(echo "$RELEASE_JSON" | jq -r '.tag_name' | sed 's/^v//')
echo "→ Latest version: $VERSION"

# Get repository info for license and description
REPO_JSON=$(gh api "repos/$OWNER/$REPO")
LICENSE=$(echo "$REPO_JSON" | jq -r '.license.spdx_id // "NOASSERTION"')
DESC=$(echo "$REPO_JSON" | jq -r '.description // "No description"')
DISPLAY_NAME=$(echo "$REPO_JSON" | jq -r '.name')

echo "→ License: $LICENSE"
echo "→ Description: $DESC"

# Find Linux installer (.deb, .AppImage, or tarball)
ASSET_URL=$(echo "$RELEASE_JSON" | jq -r '.assets[] | select(.name | test("linux.*(amd64|x86_64).*\\.(deb|AppImage|tar\\.gz)$")) | .browser_download_url' | head -1)

if [ -z "$ASSET_URL" ]; then
  echo "Error: No Linux installer found in releases"
  exit 1
fi

echo "→ Download URL: $ASSET_URL"

# Detect installer type
if [[ "$ASSET_URL" =~ \.deb$ ]]; then
  INSTALLER_TYPE="deb"
elif [[ "$ASSET_URL" =~ \.AppImage$ ]]; then
  INSTALLER_TYPE="appimage"
else
  INSTALLER_TYPE="tarball"
fi

echo "→ Detected installer type: $INSTALLER_TYPE"

# Download and calculate SHA256
echo "→ Downloading to calculate SHA256..."
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

curl -sSL "$ASSET_URL" -o "$TEMP_FILE"
SHA256=$(sha256sum "$TEMP_FILE" | awk '{print $1}')
echo "→ SHA256: $SHA256"

# Create cask file
CASK_FILE="$REPO_ROOT/Casks/$NAME.rb"

# Generate cask based on installer type
if [ "$INSTALLER_TYPE" = "deb" ]; then
cat > "$CASK_FILE" << CASK
cask "$NAME" do
  version "$VERSION"
  sha256 "$SHA256"

  url "$ASSET_URL"
  name "$DISPLAY_NAME"
  desc "$DESC"
  homepage "https://github.com/$OWNER/$REPO"

  depends_on formula: "dpkg"

  installer script: {
    executable: "dpkg-deb",
    args:       ["-x", staged_path.join("$(basename "$ASSET_URL")"), staged_path.join("extracted")],
  }

  # TODO: Adjust binary path based on actual .deb contents
  binary "extracted/usr/bin/$NAME"

  zap trash: [
    "~/.config/$NAME",
    "~/.local/share/$NAME",
  ]
end
CASK

elif [ "$INSTALLER_TYPE" = "appimage" ]; then
cat > "$CASK_FILE" << CASK
cask "$NAME" do
  version "$VERSION"
  sha256 "$SHA256"

  url "$ASSET_URL"
  name "$DISPLAY_NAME"
  desc "$DESC"
  homepage "https://github.com/$OWNER/$REPO"

  binary "$(basename "$ASSET_URL")", target: "$NAME"

  postflight do
    set_permissions staged_path.join("$(basename "$ASSET_URL")"), "0755"
  end

  zap trash: "~/.config/$NAME"
end
CASK

else
cat > "$CASK_FILE" << CASK
cask "$NAME" do
  version "$VERSION"
  sha256 "$SHA256"

  url "$ASSET_URL"
  name "$DISPLAY_NAME"
  desc "$DESC"
  homepage "https://github.com/$OWNER/$REPO"

  # TODO: Adjust binary name and path based on tarball contents
  binary "$NAME"

  postflight do
    set_permissions staged_path.join("$NAME"), "0755"
  end

  zap trash: "~/.config/$NAME"
end
CASK
fi

echo "✓ Created $CASK_FILE"
echo ""
echo "Next steps:"
echo "1. Edit $CASK_FILE to adjust paths if needed"
echo "2. Test locally: brew install --cask $NAME"
echo "3. Run audit: brew audit --strict --online --cask $NAME"
echo "4. Commit and push"
EOF

chmod +x scripts/new-cask.sh
```

**Step 2: Commit new-cask.sh**

```bash
git add scripts/new-cask.sh
git commit -m "feat: add new-cask.sh helper script"
```

---

## Task 7: Create Helper Scripts - Part 3 (validate-all.sh and update-sha256.sh)

**Files:**
- Create: `scripts/validate-all.sh`
- Create: `scripts/update-sha256.sh`

**Step 1: Create validate-all.sh**

```bash
cat > scripts/validate-all.sh << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🔍 Validating all formulas and casks..."
echo ""

FAILED=0

# Check formulas
if [ -d "$REPO_ROOT/Formula" ] && [ "$(ls -A "$REPO_ROOT/Formula"/*.rb 2>/dev/null)" ]; then
  echo "→ Auditing formulas..."
  for formula in "$REPO_ROOT/Formula"/*.rb; do
    FORMULA_NAME=$(basename "$formula" .rb)
    echo "  Checking $FORMULA_NAME..."
    
    if ! brew audit --strict --online "$formula"; then
      echo "  ✗ $FORMULA_NAME failed audit"
      FAILED=$((FAILED + 1))
    fi
    
    if ! brew style "$formula"; then
      echo "  ✗ $FORMULA_NAME failed style check"
      FAILED=$((FAILED + 1))
    fi
  done
else
  echo "→ No formulas to check"
fi

echo ""

# Check casks
if [ -d "$REPO_ROOT/Casks" ] && [ "$(ls -A "$REPO_ROOT/Casks"/*.rb 2>/dev/null)" ]; then
  echo "→ Auditing casks..."
  for cask in "$REPO_ROOT/Casks"/*.rb; do
    CASK_NAME=$(basename "$cask" .rb)
    echo "  Checking $CASK_NAME..."
    
    if ! brew audit --strict --online --cask "$cask"; then
      echo "  ✗ $CASK_NAME failed audit"
      FAILED=$((FAILED + 1))
    fi
    
    if ! brew style "$cask"; then
      echo "  ✗ $CASK_NAME failed style check"
      FAILED=$((FAILED + 1))
    fi
  done
else
  echo "→ No casks to check"
fi

echo ""

if [ $FAILED -eq 0 ]; then
  echo "✓ All checks passed!"
  exit 0
else
  echo "✗ $FAILED check(s) failed"
  exit 1
fi
EOF

chmod +x scripts/validate-all.sh
```

**Step 2: Create update-sha256.sh**

```bash
cat > scripts/update-sha256.sh << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

# Usage: ./scripts/update-sha256.sh <formula-or-cask-name> <new-version>
# Example: ./scripts/update-sha256.sh jq 1.7.1

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <formula-or-cask-name> <new-version>"
  echo "Example: $0 jq 1.7.1"
  exit 1
fi

NAME="$1"
NEW_VERSION="$2"

# Find the file (check Formula first, then Casks)
FILE=""
if [ -f "$REPO_ROOT/Formula/$NAME.rb" ]; then
  FILE="$REPO_ROOT/Formula/$NAME.rb"
  TYPE="formula"
elif [ -f "$REPO_ROOT/Casks/$NAME.rb" ]; then
  FILE="$REPO_ROOT/Casks/$NAME.rb"
  TYPE="cask"
else
  echo "Error: Could not find $NAME.rb in Formula/ or Casks/"
  exit 1
fi

echo "→ Found $TYPE: $FILE"

# Extract current URL pattern
URL_LINE=$(grep -E '^\s*url\s+"' "$FILE" || true)
if [ -z "$URL_LINE" ]; then
  echo "Error: Could not find URL in $FILE"
  exit 1
fi

# Extract URL and replace version
CURRENT_VERSION=$(grep -E '^\s*version\s+"' "$FILE" | sed -E 's/.*version\s+"([^"]+)".*/\1/')
NEW_URL=$(echo "$URL_LINE" | sed -E 's/.*url\s+"([^"]+)".*/\1/' | sed "s/$CURRENT_VERSION/$NEW_VERSION/g")

echo "→ Current version: $CURRENT_VERSION"
echo "→ New version: $NEW_VERSION"
echo "→ New URL: $NEW_URL"

# Download and calculate SHA256
echo "→ Downloading to calculate SHA256..."
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

if ! curl -sSL "$NEW_URL" -o "$TEMP_FILE"; then
  echo "Error: Failed to download from $NEW_URL"
  exit 1
fi

NEW_SHA256=$(sha256sum "$TEMP_FILE" | awk '{print $1}')
echo "→ New SHA256: $NEW_SHA256"

# Create backup
cp "$FILE" "$FILE.bak"

# Update version
sed -i "s/version \"$CURRENT_VERSION\"/version \"$NEW_VERSION\"/" "$FILE"

# Update SHA256
sed -i "s/sha256 \"[^\"]*\"/sha256 \"$NEW_SHA256\"/" "$FILE"

echo "✓ Updated $FILE"
echo ""
echo "Next steps:"
echo "1. Review changes: git diff $FILE"
echo "2. Test: brew install $([[ "$TYPE" == "cask" ]] && echo "--cask ") $NAME"
echo "3. Audit: brew audit --strict --online $([[ "$TYPE" == "cask" ]] && echo "--cask ") $NAME"
echo "4. Commit: git add $FILE && git commit -m \"chore: update $NAME to $NEW_VERSION\""
echo ""
echo "Backup saved to: $FILE.bak"
EOF

chmod +x scripts/update-sha256.sh
```

**Step 3: Commit helper scripts**

```bash
git add scripts/validate-all.sh scripts/update-sha256.sh
git commit -m "feat: add validate-all.sh and update-sha256.sh helper scripts"
```

---

## Task 8: Create Helper Scripts - Part 4 (from-issue.sh)

**Files:**
- Create: `scripts/from-issue.sh`

**Step 1: Create from-issue.sh script**

```bash
cat > scripts/from-issue.sh << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

# Usage: ./scripts/from-issue.sh <issue-number>
# Example: ./scripts/from-issue.sh 42

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ $# -ne 1 ]; then
  echo "Usage: $0 <issue-number>"
  echo "Example: $0 42"
  exit 1
fi

ISSUE_NUM="$1"

echo "→ Fetching issue #$ISSUE_NUM..."

# Get issue data
ISSUE_JSON=$(gh issue view "$ISSUE_NUM" --json title,body,labels)

TITLE=$(echo "$ISSUE_JSON" | jq -r '.title')
BODY=$(echo "$ISSUE_JSON" | jq -r '.body')
LABELS=$(echo "$ISSUE_JSON" | jq -r '.labels[].name' | tr '\n' ' ')

echo "→ Title: $TITLE"

# Check if it's a package request
if [[ ! "$LABELS" =~ "package-request" ]]; then
  echo "Error: Issue #$ISSUE_NUM is not a package request"
  exit 1
fi

# Extract package name from body
PACKAGE_NAME=$(echo "$BODY" | grep -A1 "Package Name" | tail -1 | sed 's/^[[:space:]]*//')
REPO_URL=$(echo "$BODY" | grep -A1 "Repository or Homepage URL" | tail -1 | sed 's/^[[:space:]]*//')
DESCRIPTION=$(echo "$BODY" | grep -A1 "Description" | tail -1 | sed 's/^[[:space:]]*//')

if [ -z "$PACKAGE_NAME" ] || [ -z "$REPO_URL" ]; then
  echo "Error: Could not extract package name or repository URL from issue"
  exit 1
fi

echo "→ Package: $PACKAGE_NAME"
echo "→ Repository: $REPO_URL"
echo "→ Description: $DESCRIPTION"

# Validate it's a GitHub URL
if [[ ! "$REPO_URL" =~ github\.com ]]; then
  echo "Error: Only GitHub repositories are supported by this automation"
  exit 1
fi

# Get latest release to determine type
OWNER_REPO=$(echo "$REPO_URL" | sed -E 's|.*github\.com/([^/]+/[^/]+).*|\1|')
echo "→ Checking releases for $OWNER_REPO..."

RELEASE_JSON=$(gh api "repos/$OWNER_REPO/releases/latest" 2>/dev/null || echo "{}")

if [ "$RELEASE_JSON" = "{}" ]; then
  echo "Error: No releases found. Cannot auto-detect package type."
  exit 1
fi

# Detect package type based on assets
ASSETS=$(echo "$RELEASE_JSON" | jq -r '.assets[].name')
TYPE="formula"  # Default

if echo "$ASSETS" | grep -qiE '\.(deb|AppImage)$'; then
  TYPE="cask"
  echo "→ Detected GUI application (found .deb or .AppImage)"
elif echo "$ASSETS" | grep -qiE '(app|gui|desktop)'; then
  TYPE="cask"
  echo "→ Detected GUI application (found GUI-related keywords)"
else
  echo "→ Detected CLI tool"
fi

# Create branch
BRANCH="package/$PACKAGE_NAME"
echo "→ Creating branch: $BRANCH"
git checkout -b "$BRANCH" 2>/dev/null || git checkout "$BRANCH"

# Run appropriate script
if [ "$TYPE" = "formula" ]; then
  echo "→ Creating formula..."
  "$SCRIPT_DIR/new-formula.sh" "$PACKAGE_NAME" "$REPO_URL"
  FILE="Formula/$PACKAGE_NAME.rb"
else
  echo "→ Creating cask..."
  "$SCRIPT_DIR/new-cask.sh" "$PACKAGE_NAME" "$REPO_URL"
  FILE="Casks/$PACKAGE_NAME.rb"
fi

# Commit
git add "$FILE"
git commit -m "feat: add $PACKAGE_NAME

Closes #$ISSUE_NUM

Package: $PACKAGE_NAME
Repository: $REPO_URL
Description: $DESCRIPTION
Type: $TYPE"

echo ""
echo "✓ Created $FILE and committed"
echo ""
echo "Next steps:"
echo "1. Review and edit $FILE if needed"
echo "2. Test installation locally"
echo "3. Push branch: git push -u origin $BRANCH"
echo "4. Create PR: gh pr create --fill --base main"
EOF

chmod +x scripts/from-issue.sh
```

**Step 2: Commit from-issue.sh**

```bash
git add scripts/from-issue.sh
git commit -m "feat: add from-issue.sh for automated package creation from issues"
```

---

## Task 9: Create Documentation - Part 1 (AGENT_GUIDE.md)

**Files:**
- Create: `docs/AGENT_GUIDE.md`

**Step 1: Create AGENT_GUIDE.md**

```bash
cat > docs/AGENT_GUIDE.md << 'EOF'
# Agent Guide: Packaging Software for Homebrew

This guide helps AI agents (like GitHub Copilot) package software for this Homebrew tap.

## Quick Start

**Most common workflow:**

```bash
# User creates GitHub issue with package request
# Agent processes it with one command:
./scripts/from-issue.sh <issue-number>

# Review generated file, push, and create PR
```

## Decision Tree

### 1. Formula or Cask?

- **Formula** → CLI tools, libraries, servers, development tools
- **Cask** → GUI applications with graphical interfaces

### 2. What's the source?

- **GitHub Release (tarball)** → Most common, use Pattern 1 or 2
- **GitHub Release (.deb)** → Use Pattern 3 (formula) or Pattern 1 (cask for GUI)
- **GitHub Release (AppImage)** → Use Pattern 2 (cask)
- **Direct URL** → Use appropriate pattern, add manual livecheck

## Formula Workflow

### Step 1: Generate Template

```bash
./scripts/new-formula.sh package-name https://github.com/user/repo
```

This auto-generates:
- Version from latest release
- SHA256 checksum
- License from repository
- Basic structure

### Step 2: Complete the Install Block

See [FORMULA_PATTERNS.md](FORMULA_PATTERNS.md) for copy-paste templates.

**Simple binary:**
```ruby
def install
  bin.install "binary-name"
end
```

**Multiple files:**
```ruby
def install
  bin.install "bin/tool"
  lib.install Dir["lib/*"]
  man1.install Dir["man/*.1"]
end
```

**From .deb:**
```ruby
def install
  system "ar", "x", cached_download
  system "tar", "xf", "data.tar.xz"
  bin.install "usr/bin/tool"
end
```

### Step 3: Add Meaningful Test

```ruby
test do
  system "#{bin}/tool", "--version"
  # or
  assert_match "expected", shell_output("#{bin}/tool --help")
end
```

### Step 4: Validate

```bash
# Check syntax and standards
brew audit --strict --online Formula/package-name.rb

# Check Ruby style
brew style Formula/package-name.rb

# Test installation
HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source package-name

# Run test
brew test package-name
```

### Step 5: Commit and PR

```bash
git add Formula/package-name.rb
git commit -m "feat: add package-name formula"
git push -u origin branch-name
gh pr create --fill --base main
```

## Cask Workflow

### Step 1: Generate Template

```bash
./scripts/new-cask.sh app-name https://github.com/user/repo
```

### Step 2: Verify Binary Paths

For .deb packages, extract and check:
```bash
ar x package.deb
tar xf data.tar.xz
tree usr/  # Check where binaries are
```

Update cask:
```ruby
binary "extracted/usr/bin/actual-binary-name"
```

### Step 3: Add Cleanup (zap)

```ruby
zap trash: [
  "~/.config/app-name",
  "~/.local/share/app-name",
]
```

### Step 4: Validate

```bash
brew audit --strict --online --cask Casks/app-name.rb
brew style Casks/app-name.rb
brew install --cask app-name
```

### Step 5: Commit and PR

```bash
git add Casks/app-name.rb
git commit -m "feat: add app-name cask"
git push -u origin branch-name
gh pr create --fill --base main
```

## Common Patterns

### Adding Livecheck

```ruby
livecheck do
  url :stable
  strategy :github_latest
end
```

### Adding Dependencies

```ruby
depends_on "openssl@3"
depends_on "pcre2"
```

### Multi-Architecture Support (Cask)

```ruby
arch arm: "arm64", intel: "x86_64"

version "1.0.0"
sha256 arm:   "abc123...",
       intel: "def456..."

url "https://example.com/app-#{version}-#{arch}.tar.gz"
```

## Troubleshooting

**"Formula doesn't install"**
- Check binary paths with `tar -tzf downloaded-file.tar.gz`
- Verify binary name matches what's in archive

**"audit failed: missing license"**
- Check GitHub API: `gh api repos/user/repo | jq .license.spdx_id`
- Add to formula: `license "MIT"`

**"SHA256 mismatch"**
- Recalculate: `shasum -a 256 file.tar.gz`
- Update in formula: `sha256 "new-hash"`

**"Renovate not detecting updates"**
- Ensure livecheck block present
- Check URL pattern matches releases

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for more details.

## Resources

- [FORMULA_PATTERNS.md](FORMULA_PATTERNS.md) - Copy-paste formula templates
- [CASK_PATTERNS.md](CASK_PATTERNS.md) - Copy-paste cask templates
- [DEB_CONVERSION.md](DEB_CONVERSION.md) - Extract from .deb packages
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Cask Cookbook](https://docs.brew.sh/Cask-Cookbook)
EOF
```

**Step 2: Commit AGENT_GUIDE.md**

```bash
git add docs/AGENT_GUIDE.md
git commit -m "docs: add AGENT_GUIDE.md"
```

---

## Task 10: Create Documentation - Part 2 (FORMULA_PATTERNS.md)

**Files:**
- Create: `docs/FORMULA_PATTERNS.md`

**Step 1: Create FORMULA_PATTERNS.md**

```bash
cat > docs/FORMULA_PATTERNS.md << 'EOF'
# Formula Patterns for Linux

Copy-paste templates for common formula types. Replace placeholders with actual values.

## Pattern 1: Simple Binary from GitHub Release

**Use when:** Pre-built Linux binary in GitHub releases (.tar.gz with single binary)

```ruby
class ExampleCli < Formula
  desc "Short one-line description"
  homepage "https://github.com/user/project"
  url "https://github.com/user/project/releases/download/v1.0.0/project-1.0.0-linux-x86_64.tar.gz"
  sha256 "abc123..."
  license "MIT"

  def install
    bin.install "binary-name"
  end

  test do
    system "#{bin}/binary-name", "--version"
  end
end
```

**Placeholders:**
- `ExampleCli` → PascalCase class name (e.g., `Jq`, `Ripgrep`)
- `Short one-line description` → What the tool does
- `user/project` → GitHub owner/repo
- `v1.0.0` → Latest version tag
- `project-1.0.0-linux-x86_64.tar.gz` → Actual asset name
- `abc123...` → SHA256 checksum
- `MIT` → SPDX license identifier
- `binary-name` → Name of executable in tarball

---

## Pattern 2: Tarball with Multiple Files

**Use when:** Release contains multiple binaries, libraries, man pages, or other files

```ruby
class ExampleTool < Formula
  desc "Tool with multiple components"
  homepage "https://github.com/user/project"
  url "https://github.com/user/project/releases/download/v1.0.0/project-1.0.0-linux-x86_64.tar.gz"
  sha256 "abc123..."
  license "Apache-2.0"

  def install
    bin.install "bin/tool1", "bin/tool2"
    lib.install Dir["lib/*"]
    man1.install Dir["man/*.1"]
    share.install "share/project"
  end

  test do
    assert_match "version 1.0.0", shell_output("#{bin}/tool1 --version")
  end
end
```

**Installation paths:**
- `bin.install` → Executable binaries (available in $PATH)
- `lib.install` → Shared libraries
- `man1.install` → Man pages (section 1)
- `share.install` → Application data, configs, docs

**Find files in tarball:**
```bash
tar -tzf downloaded.tar.gz
```

---

## Pattern 3: Extraction from .deb Package

**Use when:** Software only distributed as .deb (Debian package)

```ruby
class ExampleFromDeb < Formula
  desc "Tool packaged as .deb"
  homepage "https://example.com"
  url "https://example.com/releases/tool_1.0.0_amd64.deb"
  sha256 "abc123..."
  license "GPL-3.0"

  def install
    # Extract .deb package
    system "ar", "x", cached_download
    system "tar", "xf", "data.tar.xz"
    
    # Install files from usr/* structure
    bin.install "usr/bin/tool"
    
    # Optional: Install libraries and shared data
    lib.install Dir["usr/lib/*"] if File.directory?("usr/lib")
    share.install Dir["usr/share/tool"] if File.directory?("usr/share/tool")
  end

  test do
    system "#{bin}/tool", "--help"
  end
end
```

**Inspect .deb contents:**
```bash
ar x package.deb
tar -tzf data.tar.xz
```

**Common paths in .deb:**
- `usr/bin/` → Binaries
- `usr/lib/` → Libraries
- `usr/share/` → Application data

---

## Pattern 4: Tool with Dependencies

**Use when:** Package requires other Homebrew packages to function

```ruby
class ExampleWithDeps < Formula
  desc "Tool that requires other packages"
  homepage "https://github.com/user/project"
  url "https://github.com/user/project/releases/download/v1.0.0/project-1.0.0.tar.gz"
  sha256 "abc123..."
  license "MIT"

  depends_on "openssl@3"
  depends_on "pcre2"

  def install
    system "./configure", "--prefix=#{prefix}",
                          "--with-openssl=#{Formula["openssl@3"].opt_prefix}"
    system "make", "install"
  end

  test do
    system "#{bin}/tool", "--version"
  end
end
```

**Common dependencies:**
- `openssl@3` → TLS/SSL support
- `pcre2` → Regular expressions
- `sqlite` → Database
- `zlib` → Compression

**Build dependencies** (only needed during installation):
```ruby
depends_on "cmake" => :build
depends_on "pkg-config" => :build
```

---

## Pattern 5: Auto-Update with Livecheck

**Use when:** You want Renovate to automatically detect new versions

```ruby
class ExampleAutoUpdate < Formula
  desc "Tool with automatic version detection"
  homepage "https://github.com/user/project"
  url "https://github.com/user/project/releases/download/v1.0.0/project-1.0.0-linux-x86_64.tar.gz"
  sha256 "abc123..."
  license "MIT"

  livecheck do
    url :stable
    strategy :github_latest
  end

  def install
    bin.install "binary"
  end

  test do
    system "#{bin}/binary", "--version"
  end
end
```

**Livecheck strategies:**
- `:github_latest` → Use latest GitHub release
- `:github_releases` → Check all releases (if :latest doesn't work)
- Custom regex for non-GitHub URLs

**Custom livecheck example:**
```ruby
livecheck do
  url "https://example.com/downloads/"
  regex(/href=.*?v?(\d+(?:\.\d+)+)\.tar\.gz/i)
end
```

---

## Quick Reference

### Standard Methods

- `bin.install` → Install to bin/
- `lib.install` → Install to lib/
- `man1.install` → Install man pages (section 1)
- `share.install` → Install shared data
- `system` → Run shell command
- `prefix` → Installation prefix path
- `cached_download` → Downloaded file path

### Test Helpers

- `system "#{bin}/tool", "arg"` → Run command
- `shell_output("cmd")` → Capture command output
- `assert_match "text", output` → Check output contains text

### Finding Information

**License:**
```bash
gh api repos/user/repo | jq -r .license.spdx_id
```

**Latest version:**
```bash
gh api repos/user/repo/releases/latest | jq -r .tag_name
```

**SHA256:**
```bash
shasum -a 256 file.tar.gz
```
EOF
```

**Step 2: Commit FORMULA_PATTERNS.md**

```bash
git add docs/FORMULA_PATTERNS.md
git commit -m "docs: add FORMULA_PATTERNS.md"
```

---

## Task 11: Create Documentation - Part 3 (CASK_PATTERNS.md)

**Files:**
- Create: `docs/CASK_PATTERNS.md`

**Step 1: Create CASK_PATTERNS.md**

```bash
cat > docs/CASK_PATTERNS.md << 'EOF'
# Cask Patterns for Linux

Copy-paste templates for Linux GUI applications. Replace placeholders with actual values.

## Pattern 1: Linux .deb GUI Application

**Use when:** GUI app distributed as .deb package

```ruby
cask "example-app" do
  version "1.0.0"
  sha256 "abc123..."

  url "https://example.com/releases/app_#{version}_amd64.deb"
  name "Example App"
  desc "GUI application description"
  homepage "https://example.com"

  depends_on formula: "dpkg"

  installer script: {
    executable: "dpkg-deb",
    args:       ["-x", staged_path.join("app_#{version}_amd64.deb"), staged_path.join("extracted")],
  }

  binary "extracted/usr/bin/example-app"

  zap trash: [
    "~/.config/example-app",
    "~/.local/share/example-app",
  ]
end
```

**How to find binary path:**
```bash
ar x app.deb
tar -tzf data.tar.xz | grep bin/
```

**Placeholders:**
- `example-app` → Cask token (lowercase, hyphens)
- `1.0.0` → Version number
- `abc123...` → SHA256 checksum
- `Example App` → Display name
- `GUI application description` → What the app does
- `extracted/usr/bin/example-app` → Actual binary path in .deb

---

## Pattern 2: AppImage

**Use when:** GUI app distributed as AppImage file

```ruby
cask "example-appimage" do
  version "1.0.0"
  sha256 "abc123..."

  url "https://github.com/user/project/releases/download/v#{version}/Example-#{version}.AppImage"
  name "Example App"
  desc "Linux AppImage application"
  homepage "https://github.com/user/project"

  binary "Example-#{version}.AppImage", target: "example-app"

  postflight do
    set_permissions staged_path.join("Example-#{version}.AppImage"), "0755"
  end

  zap trash: "~/.config/example-app"
end
```

**Notes:**
- `target:` renames the binary in $PATH
- `postflight` makes AppImage executable
- AppImages are self-contained, no extraction needed

---

## Pattern 3: Tarball with Binary (GUI)

**Use when:** GUI app distributed as .tar.gz with executable

```ruby
cask "example-tarball-gui" do
  version "1.0.0"
  sha256 "abc123..."

  url "https://example.com/releases/example-#{version}-linux-x64.tar.gz"
  name "Example App"
  desc "GUI application from tarball"
  homepage "https://example.com"

  binary "example-app"

  postflight do
    set_permissions staged_path.join("example-app"), "0755"
  end

  zap trash: [
    "~/.config/example-app",
    "~/.local/share/example-app",
  ]
end
```

**Finding binary in tarball:**
```bash
tar -tzf app.tar.gz
```

---

## Pattern 4: Multi-Architecture Support

**Use when:** App provides separate builds for ARM64 and x86_64

```ruby
cask "example-multiarch" do
  arch arm: "arm64", intel: "x86_64"

  version "1.0.0"
  sha256 arm:   "abc123...",
         intel: "def456..."

  url "https://example.com/releases/app-#{version}-linux-#{arch}.tar.gz"
  name "Example App"
  desc "Linux application with ARM and x86_64 support"
  homepage "https://example.com"

  binary "example-app"

  postflight do
    set_permissions staged_path.join("example-app"), "0755"
  end

  zap trash: "~/.config/example-app"
end
```

**Notes:**
- `arch arm: "arm64", intel: "x86_64"` defines architecture-specific values
- Need separate SHA256 for each architecture
- `#{arch}` substitutes the architecture in URLs

**Calculate both SHA256s:**
```bash
shasum -a 256 app-linux-arm64.tar.gz
shasum -a 256 app-linux-x86_64.tar.gz
```

---

## Pattern 5: GUI App with CLI Tools

**Use when:** GUI application includes additional command-line tools

```ruby
cask "example-with-cli" do
  version "1.0.0"
  sha256 "abc123..."

  url "https://example.com/releases/example_#{version}_amd64.deb"
  name "Example App"
  desc "GUI app that includes CLI tools"
  homepage "https://example.com"

  depends_on formula: "dpkg"

  installer script: {
    executable: "dpkg-deb",
    args:       ["-x", staged_path.join("example_#{version}_amd64.deb"), staged_path.join("extracted")],
  }

  binary "extracted/usr/bin/example-app"
  binary "extracted/usr/bin/example-cli"
  binary "extracted/usr/bin/example-tool"

  zap trash: [
    "~/.config/example-app",
    "~/.local/share/example-app",
  ]
end
```

**Multiple binaries:**
- Add multiple `binary` lines
- Each becomes available in $PATH

---

## Common Cask Elements

### Required Fields

```ruby
version "1.0.0"           # Version number
sha256 "abc123..."        # Checksum
url "https://..."         # Download URL
name "Display Name"       # Human-readable name
desc "Description"        # What the app does
homepage "https://..."    # Official website
```

### Optional Fields

```ruby
livecheck do              # Auto-update detection
  url :stable
  strategy :github_latest
end

depends_on formula: "dpkg"  # Formula dependency

conflicts_with cask: "other-cask"  # Cannot coexist
```

### Installation

```ruby
binary "path/to/binary"              # Install binary to $PATH
binary "long/path", target: "short"  # Rename binary

app "App.app"                        # macOS only
```

### Cleanup

```ruby
zap trash: [
  "~/.config/app",
  "~/.local/share/app",
  "~/.cache/app",
]
```

### Lifecycle Hooks

```ruby
preflight do
  # Runs before installation
end

postflight do
  # Runs after installation
  set_permissions staged_path.join("file"), "0755"
end

uninstall quit: "com.example.app"

zap delete: ["~/file"]
```

---

## Quick Reference

### Finding Information

**SHA256:**
```bash
shasum -a 256 file.deb
```

**Inspect .deb:**
```bash
ar x file.deb
tar -tzf data.tar.xz
```

**Test cask syntax:**
```bash
brew audit --strict --online --cask Casks/app.rb
brew style Casks/app.rb
```

**Test installation:**
```bash
brew install --cask app-name
```

### Common Paths (Linux)

**Config:** `~/.config/app-name`
**Data:** `~/.local/share/app-name`
**Cache:** `~/.cache/app-name`

### Naming Conventions

**Cask token:** lowercase, hyphens, no spaces
- `visual-studio-code` ✓
- `Visual_Studio_Code` ✗
- `vscode` ✓ (abbreviations OK)

**Binary target:** Short, memorable command
- `target: "code"` for VS Code ✓
- `target: "vsc"` ✓
- `target: "visual-studio-code"` (too long) ✗
EOF
```

**Step 2: Commit CASK_PATTERNS.md**

```bash
git add docs/CASK_PATTERNS.md
git commit -m "docs: add CASK_PATTERNS.md"
```

---

## Task 12: Create Documentation - Part 4 (DEB_CONVERSION.md and TROUBLESHOOTING.md)

**Files:**
- Create: `docs/DEB_CONVERSION.md`
- Create: `docs/TROUBLESHOOTING.md`

**Step 1: Create DEB_CONVERSION.md**

```bash
cat > docs/DEB_CONVERSION.md << 'EOF'
# Converting .deb Packages to Homebrew

Guide for extracting and repackaging Debian (.deb) packages as Homebrew formulas or casks.

## Understanding .deb Packages

A .deb package is an archive containing:
- **control.tar.xz** → Package metadata
- **data.tar.xz** → Actual files to install

The data archive uses standard Unix filesystem layout:
- `usr/bin/` → Executable binaries
- `usr/lib/` → Shared libraries
- `usr/share/` → Application data, icons, docs
- `etc/` → Configuration files

## Inspecting a .deb Package

### Step 1: Download the .deb

```bash
curl -L -O https://example.com/package.deb
```

### Step 2: Extract the archive

```bash
ar x package.deb
```

This creates:
- `control.tar.xz`
- `data.tar.xz`

### Step 3: List contents

```bash
tar -tzf data.tar.xz
```

Output example:
```
./
./usr/
./usr/bin/
./usr/bin/tool
./usr/lib/
./usr/lib/libtool.so
./usr/share/
./usr/share/doc/
```

### Step 4: Extract to inspect

```bash
mkdir extracted
tar -xzf data.tar.xz -C extracted
tree extracted/
```

## Formula from .deb (CLI Tool)

Use when the .deb contains a command-line tool.

### Template

```ruby
class ToolFromDeb < Formula
  desc "CLI tool from .deb package"
  homepage "https://example.com"
  url "https://example.com/releases/tool_1.0.0_amd64.deb"
  sha256 "abc123..."
  license "MIT"

  def install
    # Extract .deb
    system "ar", "x", cached_download
    system "tar", "xf", "data.tar.xz"
    
    # Install binaries
    bin.install "usr/bin/tool"
    
    # Optional: Install libraries
    lib.install Dir["usr/lib/lib*.so*"] if File.directory?("usr/lib")
    
    # Optional: Install man pages
    man1.install Dir["usr/share/man/man1/*.1"] if File.directory?("usr/share/man/man1")
    
    # Optional: Install shared data
    share.install "usr/share/tool" if File.directory?("usr/share/tool")
  end

  test do
    system "#{bin}/tool", "--version"
  end
end
```

### Common Adjustments

**Multiple binaries:**
```ruby
bin.install "usr/bin/tool1", "usr/bin/tool2"
```

**Libraries with subdirectories:**
```ruby
lib.install Dir["usr/lib/tool/*"]
```

**Config files:**
```ruby
etc.install "etc/tool/config.yml"
```

## Cask from .deb (GUI App)

Use when the .deb contains a graphical application.

### Template

```ruby
cask "app-from-deb" do
  version "1.0.0"
  sha256 "abc123..."

  url "https://example.com/releases/app_#{version}_amd64.deb"
  name "Application Name"
  desc "GUI application from .deb"
  homepage "https://example.com"

  depends_on formula: "dpkg"

  installer script: {
    executable: "dpkg-deb",
    args:       ["-x", staged_path.join("app_#{version}_amd64.deb"), staged_path.join("extracted")],
  }

  # Main application binary
  binary "extracted/usr/bin/app"
  
  # Optional: Additional CLI tools
  binary "extracted/usr/bin/app-cli"

  zap trash: [
    "~/.config/app",
    "~/.local/share/app",
  ]
end
```

### Finding Cleanup Paths

**Check where app stores data:**
```bash
strings extracted/usr/bin/app | grep -E "(\.config|\.local|HOME)"
```

**Common patterns:**
- Config: `~/.config/app-name`
- Data: `~/.local/share/app-name`
- Cache: `~/.cache/app-name`

## Handling Dependencies

.deb packages declare dependencies in the control file.

### View dependencies

```bash
tar -xf control.tar.xz
cat control
```

Look for `Depends:` line:
```
Depends: libc6, libssl3, libpcre2-8-0
```

### Map to Homebrew

Common mappings:
- `libssl3` → `openssl@3`
- `libpcre2-8-0` → `pcre2`
- `libsqlite3-0` → `sqlite`
- `zlib1g` → Usually provided by system

### Add to formula

```ruby
depends_on "openssl@3"
depends_on "pcre2"
```

## When NOT to Convert .deb

**Skip .deb conversion if:**

1. **Available as tarball** → Use tarball instead (simpler)
2. **System packages required** → Dependencies not in Homebrew
3. **Kernel modules needed** → Can't package kernel modules
4. **Complex post-install scripts** → .deb does heavy system integration

**Alternatives:**
- Ask upstream for tarball releases
- Build from source if available
- Package as snap or flatpak instead

## Troubleshooting

### Binary not executable

**Problem:** Installed binary doesn't run

**Solution:** Check permissions in formula
```ruby
def install
  bin.install "usr/bin/tool"
  chmod 0755, bin/"tool"
end
```

### Missing libraries

**Problem:** Binary fails with "library not found"

**Solution 1:** Install library files
```ruby
lib.install Dir["usr/lib/lib*.so*"]
```

**Solution 2:** Add dependency
```ruby
depends_on "library-name"
```

### Data files not found

**Problem:** App runs but can't find assets

**Solution:** Install to share/
```ruby
share.install "usr/share/app"
```

Update binary wrapper if needed:
```ruby
(bin/"app").write <<~EOS
  #!/bin/bash
  export APP_DATA="#{share}/app"
  exec "#{libexec}/bin/app-binary" "$@"
EOS
```

## Complete Example

Real-world example: Hypothetical "megatool" .deb package

```ruby
class Megatool < Formula
  desc "Command-line client for MEGA cloud storage"
  homepage "https://megatools.example.com"
  url "https://megatools.example.com/releases/megatool_1.11.0_amd64.deb"
  sha256 "1234567890abcdef..."
  license "GPL-2.0"

  depends_on "openssl@3"

  def install
    # Extract .deb package
    system "ar", "x", cached_download
    system "tar", "xf", "data.tar.xz"
    
    # Install binaries
    bin.install "usr/bin/megatool"
    bin.install "usr/bin/mega-dl"
    bin.install "usr/bin/mega-ls"
    
    # Install man pages
    man1.install Dir["usr/share/man/man1/*.1.gz"]
    
    # Install documentation
    doc.install Dir["usr/share/doc/megatool/*"]
  end

  test do
    system "#{bin}/megatool", "--version"
  end
end
```

## Resources

- [Debian Binary Package Format](https://www.debian.org/doc/debian-policy/ch-binary.html)
- [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
EOF
```

**Step 2: Create TROUBLESHOOTING.md**

```bash
cat > docs/TROUBLESHOOTING.md << 'EOF'
# Troubleshooting Guide

Common issues when creating or maintaining Homebrew formulas and casks.

## brew audit Failures

### Missing Required Fields

**Error:**
```
Error: Formula missing required field: desc
```

**Fix:** Add the missing field
```ruby
desc "Short one-line description"
```

**Required fields:**
- `desc` → Description
- `homepage` → Official website
- `url` → Download URL
- `sha256` → Checksum
- `license` → SPDX identifier

---

### Invalid License

**Error:**
```
Error: Formula has invalid license: "Apache 2.0"
```

**Fix:** Use SPDX identifier
```ruby
license "Apache-2.0"  # Not "Apache 2.0"
```

**Common SPDX licenses:**
- `MIT`
- `Apache-2.0`
- `GPL-3.0`, `GPL-2.0`
- `BSD-3-Clause`, `BSD-2-Clause`
- `ISC`
- `MPL-2.0`

**Find SPDX identifier:**
```bash
gh api repos/user/repo | jq -r .license.spdx_id
```

---

### URL Issues

**Error:**
```
Error: URL does not use HTTPS
```

**Fix:** Use `https://` instead of `http://`
```ruby
url "https://example.com/file.tar.gz"  # Not http://
```

**Error:**
```
Error: URL returns 404
```

**Fix:** Verify URL is accessible
```bash
curl -I https://example.com/file.tar.gz
```

---

### SHA256 Mismatch

**Error:**
```
Error: SHA256 mismatch
Expected: abc123...
  Actual: def456...
```

**Fix:** Recalculate SHA256
```bash
curl -L https://example.com/file.tar.gz | shasum -a 256
```

Update formula:
```ruby
sha256 "def456..."  # Use actual hash
```

---

## brew style Failures

### Indentation Issues

**Error:**
```
Error: Incorrect indentation (expected 2 spaces)
```

**Fix:** Use 2-space indentation
```ruby
class Example < Formula
  def install  # 2 spaces
    bin.install "binary"  # 4 spaces
  end
end
```

**Auto-fix:**
```bash
brew style --fix Formula/example.rb
```

---

### Trailing Whitespace

**Error:**
```
Error: Line has trailing whitespace
```

**Fix:** Remove spaces at end of lines

**Auto-fix:** Configure editor or run
```bash
sed -i 's/[[:space:]]*$//' Formula/example.rb
```

---

## Installation Failures

### Binary Not Found

**Problem:**
```
Error: No such file or directory - /path/to/binary
```

**Debug:** Check tarball contents
```bash
tar -tzf /path/to/cached/download.tar.gz
```

**Fix:** Use correct path from tarball
```ruby
# If binary is in bin/ subdirectory:
bin.install "bin/tool"  # Not just "tool"
```

---

### Binary Not Executable

**Problem:** Binary installs but won't run

**Fix:** Explicitly set permissions
```ruby
def install
  bin.install "tool"
  chmod 0755, bin/"tool"
end
```

For casks:
```ruby
postflight do
  set_permissions staged_path.join("tool"), "0755"
end
```

---

### Missing Dependencies

**Problem:**
```
Error: Library not loaded: libssl.3.dylib
```

**Fix:** Add dependency
```ruby
depends_on "openssl@3"
```

**Find required libraries:**
```bash
ldd /path/to/binary  # Linux
```

---

### File Paths Wrong After Install

**Problem:** App can't find data files

**Fix:** Install to correct location
```ruby
share.install "data"
```

**For .deb packages:** Check extraction
```bash
ar x package.deb
tar -tzf data.tar.xz | grep share
```

---

## Renovate Issues

### Not Detecting Updates

**Problem:** Renovate doesn't create PRs for new versions

**Fix 1:** Add livecheck block
```ruby
livecheck do
  url :stable
  strategy :github_latest
end
```

**Fix 2:** Check GitHub URL format
```ruby
# Renovate can parse:
url "https://github.com/user/repo/releases/download/v1.0.0/file.tar.gz"

# Renovate cannot parse:
url "https://cdn.example.com/file-v1.0.0.tar.gz"  # Add manual livecheck
```

---

### Wrong Version Detected

**Problem:** Renovate detects wrong version (e.g., `nightly`, `beta`)

**Fix:** Add version filter to livecheck
```ruby
livecheck do
  url :stable
  regex(/^v?(\d+\.\d+\.\d+)$/i)  # Only stable versions
end
```

---

### Auto-merge Not Working

**Problem:** PR created but doesn't auto-merge

**Check:**
1. Is it a patch release? (Only patches auto-merge by default)
2. Do CI tests pass?
3. Is there a conflict with main branch?

**Fix:** Manually merge or check Renovate config:
```json5
{
  "packageRules": [
    {
      "matchUpdateTypes": ["patch"],
      "automerge": true
    }
  ]
}
```

---

## Testing Issues

### Test Fails

**Problem:**
```
Error: Test failed: exit status 1
```

**Debug:** Run test manually
```bash
brew test package-name --verbose
```

**Common fixes:**

**Binary doesn't support --version:**
```ruby
# Instead of:
system "#{bin}/tool", "--version"

# Use:
system "#{bin}/tool", "--help"
# or just check it exists:
assert_predicate bin/"tool", :exist?
```

**Binary requires arguments:**
```ruby
# Create temp file for testing
(testpath/"test.txt").write "hello"
assert_match "hello", shell_output("#{bin}/tool test.txt")
```

---

## Script Issues

### from-issue.sh Can't Extract Info

**Problem:**
```
Error: Could not extract package name from issue
```

**Fix:** Ensure issue uses the template format

**Expected format:**
```
### Package Name
tool-name

### Repository or Homepage URL
https://github.com/user/repo

### Description
What the tool does
```

---

### new-formula.sh Fails to Download

**Problem:**
```
Error: Failed to download from URL
```

**Fix 1:** Check release has Linux binaries
```bash
gh api repos/user/repo/releases/latest | jq '.assets[].name'
```

**Fix 2:** Manually specify URL
```bash
./scripts/new-formula.sh tool-name https://github.com/user/repo
# Then edit Formula/tool-name.rb to fix URL
```

---

## General Debugging

### Enable Verbose Output

```bash
brew install --verbose package-name
brew audit --verbose package-name
```

### Check Homebrew Logs

```bash
cat "$(brew --cache)/Logs/package-name.log"
```

### Test in Clean Environment

```bash
# Uninstall and reinstall
brew uninstall package-name
brew install --build-from-source package-name
```

### Validate Locally Before PR

```bash
# Run full validation
./scripts/validate-all.sh

# Or individually:
brew audit --strict --online Formula/package.rb
brew style Formula/package.rb
brew install --build-from-source package
brew test package
```

---

## Getting Help

1. **Check official docs:** https://docs.brew.sh
2. **Search existing formulas:** https://github.com/Homebrew/homebrew-core
3. **Homebrew discussions:** https://github.com/orgs/Homebrew/discussions
4. **This repo's issues:** Create an issue with `bug` label

---

## Quick Command Reference

```bash
# Audit
brew audit --strict --online Formula/package.rb
brew audit --strict --online --cask Casks/package.rb

# Style
brew style Formula/package.rb
brew style --fix Formula/package.rb

# Install
brew install --build-from-source package
brew install --cask package

# Test
brew test package

# Info
brew info package

# Uninstall
brew uninstall package
brew uninstall --cask package

# Calculate SHA256
shasum -a 256 file.tar.gz
curl -sL URL | shasum -a 256

# Extract .deb
ar x package.deb
tar -tzf data.tar.xz

# Check tarball contents
tar -tzf file.tar.gz
```
EOF
```

**Step 3: Commit documentation**

```bash
git add docs/DEB_CONVERSION.md docs/TROUBLESHOOTING.md
git commit -m "docs: add DEB_CONVERSION.md and TROUBLESHOOTING.md"
```

---

## Task 13: Final Setup and Testing

**Files:**
- Update: `README.md` (replace placeholder with actual username)
- Verify all files

**Step 1: Review repository structure**

```bash
tree -L 2 -a
```

Expected output:
```
.
├── .editorconfig
├── .git/
├── .github/
│   ├── ISSUE_TEMPLATE/
│   ├── labeler.yml
│   ├── renovate.json5
│   └── workflows/
├── .gitignore
├── Casks/
├── Formula/
├── LICENSE
├── README.md
├── docs/
│   ├── AGENT_GUIDE.md
│   ├── CASK_PATTERNS.md
│   ├── DEB_CONVERSION.md
│   ├── FORMULA_PATTERNS.md
│   ├── TROUBLESHOOTING.md
│   └── plans/
└── scripts/
    ├── from-issue.sh
    ├── new-cask.sh
    ├── new-formula.sh
    ├── update-sha256.sh
    └── validate-all.sh
```

**Step 2: Verify script permissions**

```bash
ls -l scripts/*.sh
```

All should be executable (`-rwxr-xr-x`).

**Step 3: Test validate-all.sh on empty repo**

```bash
./scripts/validate-all.sh
```

Expected output:
```
→ No formulas to check
→ No casks to check
✓ All checks passed!
```

**Step 4: Final commit**

```bash
git add -A
git status  # Review all files
git commit -m "chore: complete automated homebrew tap setup"
```

**Step 5: Create GitHub repository**

```bash
# Replace USERNAME with your GitHub username
gh repo create homebrew-tap --public --source=. --remote=origin

# Push all commits
git push -u origin main
```

**Step 6: Enable Renovate**

Visit: https://github.com/apps/renovate
1. Click "Configure"
2. Select your repository
3. Grant access to homebrew-tap repository

Renovate will automatically detect `.github/renovate.json5`.

**Step 7: Test issue template**

1. Visit: `https://github.com/USERNAME/homebrew-tap/issues/new/choose`
2. Verify templates appear
3. Create test issue (don't submit)

**Step 8: Document repository URL**

Update `README.md` with your actual username:
```bash
sed -i 's/\[username\]/YOUR-GITHUB-USERNAME/g' README.md
git add README.md
git commit -m "docs: update README with actual GitHub username"
git push
```

---

## Verification Checklist

After completing all tasks:

- [ ] Git repository initialized and pushed to GitHub
- [ ] All directories created (Formula/, Casks/, docs/, scripts/, .github/)
- [ ] All helper scripts executable
- [ ] Issue templates visible on GitHub
- [ ] GitHub Actions workflows present in repo
- [ ] Renovate installed and configured
- [ ] Documentation complete (5 markdown files in docs/)
- [ ] README.md updated with actual username
- [ ] validate-all.sh runs successfully
- [ ] Repository is public and accessible

---

## Next Steps After Implementation

1. **Create first package** to test the workflow:
   ```bash
   # Create an issue or run directly:
   ./scripts/new-formula.sh jq https://github.com/jqlang/jq
   ```

2. **Test CI/CD**:
   - Create PR with first package
   - Verify GitHub Actions run
   - Check audit and style pass

3. **Monitor Renovate**:
   - Wait 3 hours for first Renovate scan
   - Verify PRs are created for updates
   - Check auto-merge works for patches

4. **Share with agents**:
   - Reference docs/AGENT_GUIDE.md when asking Copilot to package software
   - Use from-issue.sh workflow for new packages

5. **Iterate and improve**:
   - Add more patterns to documentation as you encounter them
   - Update scripts based on real usage
   - Refine Renovate config based on noise/accuracy

---

## Success Criteria

The tap is complete when:

✓ Empty repository validates without errors
✓ Scripts generate working formulas/casks
✓ GitHub Actions run on PRs
✓ Renovate detects and updates versions
✓ Documentation enables agents to work independently
✓ First package installs successfully via `brew install`
