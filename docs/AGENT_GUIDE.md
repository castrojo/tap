# Agent Guide: Automated Homebrew Packaging

## ⚠️ MANDATORY: READ THE PACKAGING SKILL FIRST ⚠️

**BEFORE using this guide, you MUST read the packaging skill:**

📖 **`.github/skills/homebrew-packaging/SKILL.md`** - Contains the authoritative 6-step workflow

**The skill is mandatory for all packaging work. It includes:**
- Critical constraints (Linux-only, read-only filesystem, XDG paths)
- Step-by-step workflow with checkpoints
- Validation requirements
- All official Homebrew standards

**This guide provides additional context and examples. When in doubt, follow the skill.**

## ⚠️ CRITICAL: Common Errors and How to Avoid Them

**AI Agents: Read [docs/AGENT_BEST_PRACTICES.md](docs/AGENT_BEST_PRACTICES.md) BEFORE creating packages.**

This document catalogs real errors that agents have made and how to prevent them:
- ✓ Common validation errors (regex vs strings, line length, array ordering)
- ✓ Platform-specific issues (Linux vs macOS vs Windows)
- ✓ XDG compliance patterns
- ✓ Pre-commit checklist (prevents 100% of CI failures)
- ✓ Real-world examples from actual PR failures

**Following the best practices guide prevents CI failures and speeds up development.**

---

**⚠️ LINUX ONLY REPOSITORY ⚠️**

**THIS TAP IS LINUX-ONLY. ALL PACKAGES MUST USE LINUX BINARIES.**
- ✓ Use Linux downloads (e.g., `app-linux-x64.tar.gz`, `tool_linux_amd64`)
- ✗ NEVER use macOS downloads (`.dmg`, `.pkg`, `-darwin-`, `-macos-`)
- ✗ NEVER use Windows downloads (`.exe`, `.msi`, `-windows-`, `.zip` for Windows)

This guide enables AI agents to independently package software for Homebrew. Follow these instructions to create formulas (CLI tools) and casks (GUI apps) with consistent quality.

## ⚠️ MUST READ FIRST

**Before creating ANY cask, read [CASK_CREATION_GUIDE.md](./CASK_CREATION_GUIDE.md) - it contains critical, up-to-date information that will prevent common failures.**

**Critical Topics Covered:**
- Why `depends_on :linux` FAILS (and what to use instead)
- Why `test` blocks FAIL in casks (formulas only)
- Correct stanza ordering and spacing
- CI workflow requirements for tapping repositories
- Verified minimal cask template that passes CI

## Table of Contents

1. [Overview](#overview)
   - [Package Naming Convention](#package-naming-convention) **← CRITICAL**
2. [Quick Start](#quick-start)
3. [Workflow Decision Tree](#workflow-decision-tree)
4. [Helper Scripts](#helper-scripts)
5. [Formula Creation Process](#formula-creation-process)
6. [Cask Creation Process](#cask-creation-process) **← Read CASK_CREATION_GUIDE.md first!**
7. [Common Patterns](#common-patterns)
8. [Quality Checks](#quality-checks)
9. [Troubleshooting](#troubleshooting)
10. [Best Practices](#best-practices)

---

## Overview

### Purpose

This guide provides step-by-step instructions for AI agents to:
- Process package requests from GitHub issues
- Create Homebrew formulas for CLI tools
- Create Homebrew casks for GUI applications
- Validate packages against Homebrew quality standards
- Submit pull requests with properly packaged software

### What You Can Do

- **Automate End-to-End**: From issue to merged PR
- **Make Smart Decisions**: Formula vs cask, dependencies, build systems
- **Ensure Quality**: All packages pass `brew audit --strict`
- **Learn from Patterns**: Reusable solutions for common scenarios

### Prerequisites

Before starting, ensure:
- GitHub CLI (`gh`) is installed and authenticated
- `jq` is installed for JSON parsing
- `sha256sum` is installed for checksum verification (MANDATORY)
- Git repository with proper remote configured
- Write access to create branches and PRs

## Package Format Priority and SHA256 Verification

**CRITICAL: Read this section carefully before packaging any software.**

### Package Format Selection (Strict Priority Order)

**PRIORITY 1: Tarball (PREFERRED)**

Look for Linux tarballs in this order:
1. `*-linux-x64.tar.gz` or `*-linux-amd64.tar.gz`
2. `*-x86_64-unknown-linux-gnu.tar.gz`
3. `*-linux.tar.xz` or `*-linux.tgz`
4. Generic `*.tar.gz` (verify it's for Linux)

**Why Tarballs Are Preferred:**
- Works across all Linux distributions
- Simple extraction, no dependencies
- No package manager conflicts
- Predictable directory structure

**PRIORITY 2: Debian Package (SECOND CHOICE)**

Only use if no tarball is available:
1. `*-amd64.deb` or `*_amd64.deb`
2. `*-linux-amd64.deb`

**Note:** Requires extraction using `ar` and `tar`

**PRIORITY 3: Other Formats (Case-by-Case)**

- **AppImage**: Self-contained, can be used directly
- **Snap/Flatpak**: Generally avoid (requires runtime)
- **RPM**: Avoid (use tarball or .deb instead)

### SHA256 Verification (MANDATORY - NO EXCEPTIONS)

**Every package MUST include SHA256 verification. This is non-negotiable.**

#### Verification Workflow:

**Step 1: Download the asset**
```bash
curl -LO https://github.com/user/repo/releases/download/v1.0.0/app-linux-x64.tar.gz
```

**Step 2: Calculate SHA256**
```bash
sha256sum app-linux-x64.tar.gz
# Output: 3a5b8c9def456...  app-linux-x64.tar.gz
```

**Step 3: Verify against upstream (if available)**
```bash
# Download upstream checksums
curl -LO https://github.com/user/repo/releases/download/v1.0.0/SHA256SUMS

# Verify
sha256sum --check SHA256SUMS 2>&1 | grep app-linux-x64.tar.gz
# Expected: app-linux-x64.tar.gz: OK
```

**Common checksum file names:** `SHA256SUMS`, `checksums.txt`, `CHECKSUMS`, `*.sha256`

**Step 4: Use verified SHA256 in cask**
```ruby
cask "app-name" do
  version "1.0.0"
  sha256 "3a5b8c9def456..."  # Verified hash here
  
  url "https://github.com/user/repo/releases/download/v#{version}/app-linux-x64.tar.gz"
  # ...
end
```

#### Verification Checklist:

- [ ] Downloaded actual file
- [ ] Calculated SHA256 using `sha256sum`
- [ ] Checked for upstream checksums
- [ ] Verified match if upstream checksums exist
- [ ] SHA256 is lowercase hex (64 chars)
- [ ] SHA256 is for Linux x86_64 (not ARM/macOS/Windows)

#### When Checksums Don't Match:

**STOP.** This indicates corruption or compromise. Re-download or report to upstream.

#### When No Upstream Checksums:

Still include calculated SHA256. Document in commit: "No upstream checksums available"

### Package Naming Convention

**CRITICAL: Package names are ALWAYS derived from the repository name, not manually specified.**

**Naming Rules:**
1. Use the repository name as-is (e.g., `ripgrep` repo → `ripgrep` package)
2. Convert to lowercase (e.g., `MyApp` → `myapp`)
3. Replace underscores with hyphens (e.g., `my_tool` → `my-tool`)
4. **FOR CASKS ONLY:** Append `-linux` suffix (e.g., `sublime-text` → `sublime-text-linux`)
5. **Never override the repository name** - this ensures consistency with upstream

**Linux Cask Naming (Required):**
- ALL casks MUST use `-linux` suffix
- Prevents collision with official macOS casks in `homebrew-cask`
- Makes Linux-only nature explicit
- Examples: `sublime-text-linux`, `jetbrains-toolbox-linux`
- The `new-cask.sh` script automatically appends `-linux` if not present

**Why This Matters:**
- Ensures package names match what users expect
- Maintains consistency with upstream project naming
- Prevents naming conflicts with macOS casks
- Makes updates predictable (Renovate can track by repo name)

**Examples:**

**Formulas (no suffix):**
- Repository: `BurntSushi/ripgrep` → Formula: `ripgrep`
- Repository: `sharkdp/bat` → Formula: `bat`
- Repository: `user/My_Cool_Tool` → Formula: `my-cool-tool`

**Casks (with -linux suffix):**
- Repository: `sublimehq/sublime_text` → Cask: `sublime-text-linux`
- Repository: `JetBrains/toolbox-app` → Cask: `jetbrains-toolbox-linux`
- Repository: `user/My_Cool_App` → Cask: `my-cool-app-linux`

**Issue Template:**
The issue template only requires:
- Repository URL (package name derived automatically)
- Description

**No "Package Name" field** - the name comes from the repository itself.

---

## Quick Start

### Fastest Path: From Issue to PR

```bash
# Process package request and create PR automatically
./scripts/from-issue.sh <issue-number> --create-pr

# Example: Process issue #42 and create PR
./scripts/from-issue.sh 42 --create-pr
```

This single command:
1. Fetches issue details
2. Determines formula vs cask
3. Creates git branch
4. Generates package file
5. Commits changes
6. Pushes to remote
7. Creates pull request
8. Comments on original issue

### Manual Control

If you need more control over the process:

```bash
# Step 1: Process issue (no PR)
./scripts/from-issue.sh 42

# Step 2: Review generated file
cat Formula/package-name.rb  # or Casks/app-name.rb

# Step 3: Test locally (optional)
brew install --build-from-source Formula/package-name.rb

# Step 4: Create PR manually
gh pr create --fill
```

---

## Workflow Decision Tree

### When to Use Formulas vs Casks

```
Is the software being packaged?
│
├─ CLI tool / command-line utility
│  ├─ Single binary → Formula
│  ├─ Needs compilation → Formula
│  └─ Scripting language (Ruby, Python, etc.) → Formula
│
├─ GUI application
│  ├─ macOS .app bundle → Cask
│  ├─ Electron/Tauri app → Cask
│  └─ Desktop application → Cask
│
└─ Library / Development tool
   ├─ Headers + shared libraries → Formula
   └─ Build tool / SDK → Formula
```

### Detection Heuristics

The `from-issue.sh` script auto-detects package type using:

1. **Explicit declaration** in issue (highest priority)
   - Issue template includes "Package Type" field
   - Values: "formula", "cli", "cask", "gui", "app"

2. **Repository metadata** (topics and description)
   - GUI indicators: `gui`, `desktop`, `application`, `app`, `electron`, `tauri`, `qt`, `gtk`, `macos-app`
   - CLI indicators: `cli`, `command-line`, `terminal`, `shell`, `tool`, `utility`

3. **Default to formula** if unclear

### Override Auto-Detection

If auto-detection fails, manually specify:

```bash
# For formula (CLI)
./scripts/new-formula.sh package-name https://github.com/user/repo

# For cask (GUI)
./scripts/new-cask.sh app-name https://github.com/user/repo
```

---

## Helper Scripts

### 1. from-issue.sh

**Purpose:** Automate package creation from GitHub issues.

**Usage:**
```bash
./scripts/from-issue.sh <issue-number> [--create-pr]
```

**Examples:**
```bash
# Process issue and create branch (manual PR)
./scripts/from-issue.sh 42

# Process issue and auto-create PR
./scripts/from-issue.sh 42 --create-pr
```

**What It Does:**
1. Fetches issue #N from GitHub
2. Parses issue body (repo URL, description)
3. **Derives package name from repository name** (e.g., `ripgrep` repo → `ripgrep` package)
4. Auto-detects formula vs cask based on repository metadata
5. Creates branch `package-request-N-package-name`
6. Calls `new-formula.sh` or `new-cask.sh`
7. Commits with message: `feat: add package-name formula (closes #N)`
8. Pushes to remote
9. (Optional) Creates PR and comments on issue

**Issue Template Format:**

The script expects issues to follow this format:

```markdown
### Repository or Homepage URL
https://github.com/user/my-package

### Description
A brief description of the package

### Package Type (optional)
formula
```

**IMPORTANT:** Package name is automatically derived from the repository name. Do not specify it manually.

**Exit Codes:**
- `0` - Success
- `1` - Error (invalid issue, missing fields, git failure, etc.)

---

### 2. new-formula.sh

**Purpose:** Generate Homebrew formula from GitHub repository.

**Usage:**
```bash
./scripts/new-formula.sh <package-name> <github-repo-url>
```

**Examples:**
```bash
# Create formula for a CLI tool
./scripts/new-formula.sh myapp https://github.com/user/myapp

# Works with SSH URLs too
./scripts/new-formula.sh myapp git@github.com:user/myapp.git
```

**What It Does:**
1. Validates package name (lowercase, alphanumeric, hyphens)
2. Fetches repository metadata (description, license, homepage)
3. Finds latest release and version tag
4. Downloads tarball and calculates SHA256
5. Generates formula class with:
   - Metadata (desc, homepage, url, sha256, license)
   - Install block (customize based on build system)
   - Test block (basic validation)
6. Saves to `Formula/<package-name>.rb`

**Generated Formula Structure:**

```ruby
class MyPackage < Formula
  desc "Description from GitHub"
  homepage "https://github.com/user/my-package"
  url "https://github.com/user/my-package/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "abc123..."
  license "MIT"

  def install
    bin.install "my-package"
  end

  test do
    assert_predicate bin/"my-package", :exist?
    assert_predicate bin/"my-package", :executable?
    
    begin
      output = shell_output("#{bin}/my-package --version 2>&1", 0)
      assert_match "1.0.0", output
    rescue
      system "#{bin}/my-package", "--help"
    end
  end
end
```

**Post-Generation Steps:**

The script outputs next steps:
1. Review and customize the formula
2. Add dependencies if needed (`depends_on` directives)
3. Adjust install block for actual build system
4. Test: `brew install --build-from-source Formula/package-name.rb`
5. Commit changes

**Common Customizations Needed:**

See [Formula Patterns](#common-patterns) for:
- Rust projects (Cargo)
- Go projects
- Python projects
- Projects requiring compilation
- Projects with dependencies

---

### 3. new-cask.sh

**Purpose:** Generate Homebrew cask from GitHub repository.

**Usage:**
```bash
./scripts/new-cask.sh <cask-name> <github-repo-url>
```

**Examples:**
```bash
# Create cask for a GUI app
./scripts/new-cask.sh myapp https://github.com/user/myapp

# Works with SSH URLs too
./scripts/new-cask.sh myapp git@github.com:user/myapp.git
```

**What It Does:**
1. Validates cask name (lowercase, alphanumeric, hyphens)
2. Fetches repository metadata
3. Finds latest release with binary assets (.tar.gz, .tgz, or .zip)
4. Downloads asset and calculates SHA256
5. Generates cask with:
   - Metadata (version, sha256, url, name, desc, homepage, license)
   - Binary stanza (customize based on archive contents)
   - Test block
6. Saves to `Casks/<cask-name>.rb`

**Generated Cask Structure:**

```ruby
cask "my-app" do
  version "1.0.0"
  sha256 "abc123..."
  url "https://github.com/user/my-app/releases/download/v1.0.0/my-app-macos.tar.gz"
  
  name "MyApp"
  desc "Description from GitHub"
  homepage "https://github.com/user/my-app"
  license "MIT"

  binary "my-app"

  test do
    assert_predicate bin/"my-app", :exist?
    assert_predicate bin/"my-app", :executable?
    system bin/"my-app", "--version"
  end
end
```

**Critical Post-Generation Step:**

**You MUST customize the binary name!** The script cannot know the actual binary path inside the archive.

```bash
# Extract the archive to find the binary
tar -tzf /path/to/downloaded/asset.tar.gz  # List contents

# Update the binary stanza to match
binary "actual/path/to/binary-name"
```

**Common Customizations:**

See [Cask Patterns](#common-patterns) for:
- Multiple binaries
- App bundles (.app)
- Artifacts and symlinks
- Architecture-specific downloads

---

### 4. validate-all.sh

**Purpose:** Run quality checks on all formulas and casks.

**Usage:**
```bash
./scripts/validate-all.sh
```

**What It Does:**
1. Finds all formulas in `Formula/`
2. Finds all casks in `Casks/`
3. For each formula:
   - Runs `brew audit --strict --online`
   - Runs `brew style`
4. For each cask:
   - Runs `brew audit --strict --online --cask`
   - Runs `brew style`
5. Reports pass/fail summary

**When to Run:**
- Before creating PR
- After modifying any formula/cask
- In CI/CD pipelines
- Before merging

**Exit Codes:**
- `0` - All validations passed
- `1` - One or more validations failed

**Example Output:**
```
━━━ Validating Formulas ━━━
→ Validating my-package...
✓ my-package passed audit
✓ my-package passed style check

━━━ Validating Casks ━━━
→ Validating my-app...
✓ my-app passed audit
✓ my-app passed style check

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ All validations passed!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  Formulas: 1 passed, 0 failed
  Casks:    1 passed, 0 failed
  Total:    2 passed, 0 failed
```

---

### 5. update-sha256.sh

**Purpose:** Update SHA256 checksums for existing packages.

**Usage:**
```bash
./scripts/update-sha256.sh <package-file>
```

**Examples:**
```bash
# Update formula SHA256
./scripts/update-sha256.sh Formula/my-package.rb

# Update cask SHA256
./scripts/update-sha256.sh Casks/my-app.rb
```

**What It Does:**
1. Parses the package file to extract URL
2. Downloads the file from URL
3. Calculates new SHA256
4. Updates the file with new checksum
5. Reports old vs new SHA256

**When to Use:**
- Upstream release updated without version change
- Manual version bump
- SHA256 mismatch errors
- Renovate PR needs manual SHA256 update

**Example Output:**
```
→ Parsing Formula/my-package.rb...
✓ Found URL: https://github.com/user/my-package/archive/v1.0.0.tar.gz
→ Downloading file...
✓ Downloaded 1.2 MB
→ Calculating SHA256...
✓ Old SHA256: abc123...
✓ New SHA256: def456...
→ Updating Formula/my-package.rb...
✓ SHA256 updated successfully
```

---

## Formula Creation Process

### Step-by-Step Workflow

#### 1. Gather Information

Before creating a formula, collect:
- **Package name**: Lowercase, hyphenated (e.g., `my-tool`)
- **Repository URL**: GitHub repository (must have releases)
- **Description**: One-line summary
- **License**: SPDX identifier (e.g., `MIT`, `Apache-2.0`)
- **Dependencies**: Runtime dependencies (other brew packages)
- **Build system**: Make, CMake, Cargo, Go, etc.

#### 2. Generate Base Formula

```bash
./scripts/new-formula.sh my-tool https://github.com/user/my-tool
```

This creates `Formula/my-tool.rb` with:
- Metadata populated from GitHub
- URL and SHA256 from latest release
- Basic install block (needs customization)
- Basic test block

#### 3. Customize Install Block

The generated install block is a starting point. Customize based on the project's build system.

**Rust (Cargo):**
```ruby
def install
  system "cargo", "install", "--locked", "--root", prefix, "--path", "."
end
```

**Go:**
```ruby
def install
  system "go", "build", "-o", bin/"my-tool", "."
end
```

**Make:**
```ruby
def install
  system "make", "PREFIX=#{prefix}", "install"
end
```

**CMake:**
```ruby
def install
  system "cmake", "-S", ".", "-B", "build", *std_cmake_args
  system "cmake", "--build", "build"
  system "cmake", "--install", "build"
end
```

**Python (setuptools):**
```ruby
def install
  virtualenv_install_with_resources
end
```

**Pre-built Binary:**
```ruby
def install
  bin.install "my-tool"
end
```

See [FORMULA_PATTERNS.md](FORMULA_PATTERNS.md) for detailed examples.

#### 4. Add Dependencies

If the package requires other packages:

```ruby
depends_on "openssl@3"
depends_on "python@3.11"
depends_on "rust" => :build  # Build-time only
```

#### 5. Enhance Test Block

Add meaningful tests that verify the package works:

```ruby
test do
  # Test version output
  assert_match version.to_s, shell_output("#{bin}/my-tool --version")
  
  # Test basic functionality
  (testpath/"test.txt").write("hello")
  assert_equal "HELLO", shell_output("#{bin}/my-tool uppercase test.txt").strip
  
  # Test exit codes
  system bin/"my-tool", "help"
  assert_equal 0, $CHILD_STATUS.exitstatus
end
```

#### 6. Test Locally

```bash
# Install from source
brew install --build-from-source Formula/my-tool.rb

# Run tests
brew test my-tool

# Try using it
my-tool --version
```

#### 7. Validate

```bash
# Run quality checks
./scripts/validate-all.sh

# Or validate individual formula
brew audit --strict --online Formula/my-tool.rb
brew style Formula/my-tool.rb
```

#### 8. Commit and PR

```bash
# Add and commit
git add Formula/my-tool.rb
git commit -m "feat: add my-tool formula"

# Push and create PR
git push -u origin HEAD
gh pr create --fill
```

---

## Cask Creation Process

### Step-by-Step Workflow

#### 1. Gather Information

Before creating a cask:
- **Cask name**: Lowercase, hyphenated (e.g., `my-app`)
- **Repository URL**: GitHub repository with binary releases
- **Binary asset**: Must have `.tar.gz`, `.tgz`, or `.zip` in releases
- **Binary name**: Actual executable name in the archive
- **App bundle**: If macOS app, location of `.app` bundle

#### 2. Generate Base Cask

```bash
./scripts/new-cask.sh my-app https://github.com/user/my-app
```

This creates `Casks/my-app.rb` with:
- Metadata from GitHub
- URL and SHA256 from binary asset
- Placeholder binary stanza (requires customization)

#### 3. Find Binary Path

**Critical step:** Extract the archive to find the actual binary.

```bash
# Download the asset
curl -L -o asset.tar.gz "https://github.com/user/my-app/releases/download/v1.0.0/my-app-macos.tar.gz"

# List contents
tar -tzf asset.tar.gz

# Example output:
# my-app-macos/
# my-app-macos/bin/
# my-app-macos/bin/my-app
# my-app-macos/README.md
```

#### 4. Update Binary Stanza

Based on the archive structure:

```ruby
# Single binary at root
binary "my-app"

# Binary in subdirectory
binary "my-app-macos/bin/my-app"

# Rename on install
binary "my-app-macos/bin/my-app", target: "my-app"

# Multiple binaries
binary "bin/my-app"
binary "bin/my-app-cli"
```

#### 5. Handle macOS App Bundles

If the package is a `.app` bundle:

```ruby
cask "my-app" do
  version "1.0.0"
  sha256 "abc123..."
  url "https://github.com/user/my-app/releases/download/v#{version}/MyApp.zip"
  
  name "MyApp"
  desc "My awesome application"
  homepage "https://github.com/user/my-app"
  
  app "MyApp.app"
  
  # Optional: CLI helper
  binary "#{appdir}/MyApp.app/Contents/MacOS/my-app"
end
```

#### 6. Add Artifacts

For additional files:

```ruby
# Create symlinks
artifact "my-app/share", target: "#{Dir.home}/Library/Application Support/MyApp"

# Install fonts
font "fonts/MyFont.ttf"

# Install plugins
artifact "plugins", target: "#{Dir.home}/.my-app/plugins"
```

#### 7. Test Locally

```bash
# Install cask
brew install --cask Casks/my-app.rb

# Verify binary
which my-app
my-app --version

# Uninstall to test cleanup
brew uninstall --cask my-app
```

#### 8. Validate and Submit

```bash
# Validate
./scripts/validate-all.sh

# Commit and PR
git add Casks/my-app.rb
git commit -m "feat: add my-app cask"
git push -u origin HEAD
gh pr create --fill
```

---

## Common Patterns

### Links to Pattern Documentation

Detailed examples for common scenarios:

- **[FORMULA_PATTERNS.md](FORMULA_PATTERNS.md)** - Formula recipes
  - Rust (Cargo) projects
  - Go projects
  - Python projects with virtualenv
  - CMake projects
  - Makefile projects
  - Projects with dependencies
  - Resource handling
  
- **[CASK_PATTERNS.md](CASK_PATTERNS.md)** - Cask recipes
  - Simple binary casks
  - macOS app bundles
  - Multi-binary casks
  - Architecture-specific downloads
  - Electron/Tauri apps
  - Artifacts and symlinks

- **[TESTING_PATTERNS.md](TESTING_PATTERNS.md)** - Test examples
  - Version checks
  - Functional tests
  - File I/O tests
  - Exit code validation

### Quick Reference: Formula Patterns

**Rust (Cargo):**
```ruby
depends_on "rust" => :build

def install
  system "cargo", "install", "--locked", "--root", prefix, "--path", "."
end
```

**Go:**
```ruby
depends_on "go" => :build

def install
  system "go", "build", "-o", bin/"my-tool", "."
end
```

**Python with dependencies:**
```ruby
depends_on "python@3.11"

resource "requests" do
  url "https://files.pythonhosted.org/..."
  sha256 "abc123..."
end

def install
  virtualenv_install_with_resources
end
```

**Make:**
```ruby
def install
  system "make", "install", "PREFIX=#{prefix}"
end
```

**CMake:**
```ruby
depends_on "cmake" => :build

def install
  system "cmake", "-S", ".", "-B", "build", *std_cmake_args
  system "cmake", "--build", "build"
  system "cmake", "--install", "build"
end
```

### Quick Reference: Cask Patterns

**Simple binary:**
```ruby
binary "my-app"
```

**macOS app:**
```ruby
app "MyApp.app"
```

**Multiple binaries:**
```ruby
binary "bin/my-app"
binary "bin/my-app-cli"
binary "bin/my-app-server"
```

**Architecture-specific:**
```ruby
url "https://example.com/my-app-#{version}-#{arch}.tar.gz"

# Or with conditionals
if Hardware::CPU.intel?
  url "https://example.com/my-app-#{version}-x86_64.tar.gz"
  sha256 "abc123..."
elsif Hardware::CPU.arm?
  url "https://example.com/my-app-#{version}-arm64.tar.gz"
  sha256 "def456..."
end
```

---

## Quality Checks

### Required Validations

Every formula and cask MUST pass:

1. **Brew Audit (Strict + Online)**
   ```bash
   brew audit --strict --online Formula/my-tool.rb
   brew audit --strict --online --cask Casks/my-app.rb
   ```

2. **Brew Style**
   ```bash
   brew style Formula/my-tool.rb
   brew style Casks/my-app.rb
   ```

3. **Local Install Test**
   ```bash
   brew install --build-from-source Formula/my-tool.rb
   brew test my-tool
   ```

4. **Brew Test**
   ```bash
   brew test my-tool
   brew test --cask my-app
   ```

### Automated Validation

Use the validation helper:

```bash
./scripts/validate-all.sh
```

This runs all checks and reports:
- Which packages passed/failed
- Specific errors for failures
- Summary counts

### Common Audit Issues

**Missing license:**
```ruby
# Add SPDX license identifier
license "MIT"
```

**Non-HTTPS URL:**
```ruby
# Change http:// to https://
url "https://github.com/user/repo/archive/v1.0.0.tar.gz"
```

**Missing test:**
```ruby
test do
  # Must have at least basic validation
  system bin/"my-tool", "--version"
end
```

**Incorrect SHA256:**
```bash
# Recalculate
./scripts/update-sha256.sh Formula/my-tool.rb
```

**Trailing whitespace:**
```bash
# Auto-fix
brew style --fix Formula/my-tool.rb
```

### Pre-Commit Checklist

Before committing:

- [ ] Formula/cask file exists in correct directory
- [ ] `brew audit --strict --online` passes
- [ ] `brew style` passes
- [ ] `brew install` succeeds locally
- [ ] `brew test` passes
- [ ] Binary/app is actually usable
- [ ] No sensitive data in formula (API keys, tokens)
- [ ] Version is correct and up-to-date
- [ ] SHA256 matches downloaded file
- [ ] License is valid SPDX identifier

### Pre-PR Checklist

Before creating PR:

- [ ] All commits have meaningful messages
- [ ] Branch name follows convention: `package-request-N-package-name`
- [ ] PR description references issue: `Closes #N`
- [ ] All validations pass
- [ ] Tested on clean system (if possible)
- [ ] Documentation updated (if needed)

---

## Troubleshooting

### Common Issues and Solutions

#### Issue: "No releases found"

**Cause:** Repository has no GitHub releases.

**Solution:**
- Check if repository uses tags without releases
- Ask maintainer to create a release
- Cannot package without releases (Homebrew requirement)

#### Issue: "SHA256 mismatch"

**Cause:** Downloaded file doesn't match recorded checksum.

**Solution:**
```bash
# Recalculate SHA256
./scripts/update-sha256.sh Formula/my-tool.rb

# Or manually
curl -L -o /tmp/file.tar.gz "URL_FROM_FORMULA"
sha256sum /tmp/file.tar.gz
# Update sha256 line in formula
```

#### Issue: "Audit failed: HTTPS required"

**Cause:** Formula uses `http://` URL.

**Solution:**
```ruby
# Change to https://
url "https://github.com/user/repo/archive/v1.0.0.tar.gz"
```

#### Issue: "Binary not found in archive"

**Cause:** Cask references wrong binary path.

**Solution:**
```bash
# Download and extract
curl -L -o /tmp/asset.tar.gz "URL_FROM_CASK"
tar -tzf /tmp/asset.tar.gz | grep -E "(bin/|\.app)"

# Update binary stanza with correct path
binary "correct/path/to/binary"
```

#### Issue: "depends_on 'rust' is required"

**Cause:** Formula tries to use Cargo without declaring dependency.

**Solution:**
```ruby
depends_on "rust" => :build  # Add at top of formula
```

#### Issue: "Python resource missing"

**Cause:** Python package has dependencies not declared.

**Solution:**
```bash
# Use brew's helper to generate resources
brew install poetry2brew  # or equivalent tool
poetry2brew --formula my-tool

# Add generated resource blocks to formula
```

#### Issue: "Test failed"

**Cause:** Test block has incorrect assertions or binary doesn't work.

**Solution:**
```ruby
test do
  # Start simple
  system bin/"my-tool", "--version"
  
  # Add more complex tests once basic test passes
  assert_match version.to_s, shell_output("#{bin}/my-tool --version")
end
```

#### Issue: "Class name mismatch"

**Cause:** Formula class name doesn't match filename.

**Solution:**
```ruby
# File: Formula/my-tool.rb
class MyTool < Formula  # Must match: MyTool from my-tool
  # ...
end

# File: Formula/foo-bar-baz.rb
class FooBarBaz < Formula  # Must match: FooBarBaz from foo-bar-baz
  # ...
end
```

#### Issue: "GitHub CLI not authenticated"

**Cause:** `gh` CLI needs authentication.

**Solution:**
```bash
gh auth login
# Follow prompts to authenticate
```

#### Issue: "Permission denied" when pushing

**Cause:** No write access to repository or branch protected.

**Solution:**
- Verify you have write access
- Check if branch protection rules prevent direct push
- Use PR from fork if needed

#### Issue: "Renovate not updating"

**Cause:** Renovate config incorrect or formula not auto-detectable.

**Solution:**
- Check `renovate.json` includes formulas/casks
- Verify formula has valid URL pattern
- Check Renovate dashboard for errors
- May need manual version updates

---

## Best Practices

### Homebrew Conventions

#### Naming

**CRITICAL: Package names are derived from repository names, not manually specified.**

- **Rule**: Use the repository name (lowercase, underscores→hyphens)
  ```
  Repository: BurntSushi/ripgrep → Package: ripgrep
  Repository: sharkdp/bat → Package: bat  
  Repository: user/My_Tool → Package: my-tool
  ```

- **Never override** the repository name manually
- **Class names**: CamelCase from filename (e.g., `MyTool` from `my-tool.rb`)
- **Formulas/Casks**: Both use lowercase, hyphenated names

**Why This Matters:**
- Ensures consistency with upstream project naming
- Makes Renovate updates predictable (tracks by repo name)
- Prevents naming conflicts
- Matches user expectations

#### Metadata

- **desc**: One-line description, no "A" or "The" prefix
  ```ruby
  desc "Fast and modern tool for..."  # Good
  desc "A tool that does..."          # Bad
  ```

- **homepage**: Official project page
  ```ruby
  homepage "https://example.com"      # Official site
  homepage "https://github.com/..."   # OK if no official site
  ```

- **license**: Valid SPDX identifier
  ```ruby
  license "MIT"                       # Good
  license "Apache-2.0"                # Good
  license "GPL-3.0-or-later"          # Good
  license "Proprietary"               # Only if truly proprietary
  ```

#### URLs

- Always HTTPS when possible
- Use GitHub releases archive for formulas:
  ```ruby
  url "https://github.com/user/repo/archive/refs/tags/v#{version}.tar.gz"
  ```
- Use release assets for casks:
  ```ruby
  url "https://github.com/user/repo/releases/download/v#{version}/app.tar.gz"
  ```

#### Dependencies

- Declare all dependencies:
  ```ruby
  depends_on "openssl@3"              # Runtime dependency
  depends_on "rust" => :build         # Build-only dependency
  depends_on "python@3.11"            # Specific version
  depends_on :macos                   # macOS only
  ```

- Order: runtime deps, then build deps, then OS deps

#### Tests

- Must test that formula actually works
- Start with version check:
  ```ruby
  test do
    assert_match version.to_s, shell_output("#{bin}/my-tool --version")
  end
  ```
- Add functional tests when possible:
  ```ruby
  test do
    # Version
    assert_match version.to_s, shell_output("#{bin}/my-tool --version")
    
    # Functionality
    (testpath/"input.txt").write("hello")
    assert_equal "HELLO", shell_output("#{bin}/my-tool uppercase input.txt").strip
  end
  ```

### Agent-Specific Best Practices

#### Always Validate Before Committing

```bash
# Run validation
./scripts/validate-all.sh

# Only commit if validation passes
if [ $? -eq 0 ]; then
  git add Formula/my-tool.rb
  git commit -m "feat: add my-tool formula"
else
  echo "Validation failed. Fix issues before committing."
fi
```

#### Use Descriptive Commit Messages

```bash
# Good
git commit -m "feat: add my-tool formula"
git commit -m "fix: update my-tool SHA256 for v1.2.0"
git commit -m "chore: update my-app cask to v2.0.0"

# Bad
git commit -m "add package"
git commit -m "fix"
git commit -m "update"
```

#### Reference Issues in Commits

```bash
# Closes issue automatically when merged
git commit -m "feat: add my-tool formula (closes #42)"

# References issue without closing
git commit -m "feat: add my-tool formula (refs #42)"
```

#### Test Locally Before Pushing

```bash
# Install from source
brew install --build-from-source Formula/my-tool.rb

# Run tests
brew test my-tool

# Actually use it
my-tool --version
my-tool --help
```

#### Document Non-Obvious Decisions

Use code comments for unusual patterns:

```ruby
def install
  # Use specific Rust features required by project
  system "cargo", "install", "--locked", "--root", prefix, "--path", ".", 
         "--features", "ssl-vendored"
  
  # Binary is named differently in the source
  bin.install bin/"my-tool-bin" => "my-tool"
end
```

#### Keep Formulas Simple

- Avoid complex logic in install blocks
- Use Homebrew's built-in helpers when available
- Prefer upstream fixes over local patches
- Document any patches with comments

#### Respond to Audit Failures

```bash
# Run audit
brew audit --strict --online Formula/my-tool.rb

# Read the error messages carefully
# Each error has a specific fix
# Common fixes:
# - Add missing license
# - Change http to https
# - Fix class name
# - Add meaningful test
```

---

## Summary

### Quick Command Reference

```bash
# Automate from issue
./scripts/from-issue.sh <issue-number> --create-pr

# Manual formula creation
./scripts/new-formula.sh <name> <url>

# Manual cask creation
./scripts/new-cask.sh <name> <url>

# Validate all packages
./scripts/validate-all.sh

# Update SHA256
./scripts/update-sha256.sh <package-file>

# Test locally
brew install --build-from-source Formula/my-tool.rb
brew test my-tool
brew install --cask Casks/my-app.rb
```

### Decision Flow

```
1. Receive package request → Check issue template
2. Determine type → Formula (CLI) or Cask (GUI)
3. Generate base package → Use helper scripts
4. Customize → Install block, dependencies, tests
5. Validate → Run brew audit and style checks
6. Test locally → Install and use the package
7. Commit → Descriptive message with issue reference
8. Push → Create PR with proper description
9. Respond to feedback → Fix any review comments
10. Merge → Package is live!
```

### Key Principles

1. **Automation First**: Use scripts when available
2. **Validate Always**: Never commit without validation
3. **Test Locally**: Install and use the package yourself
4. **Document Decisions**: Comment non-obvious choices
5. **Follow Conventions**: Homebrew has strong conventions
6. **Be Explicit**: AI agents need clear instructions
7. **Learn from Patterns**: Reference pattern docs for complex cases

---

## Additional Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Cask Cookbook](https://docs.brew.sh/Cask-Cookbook)
- [Formula Patterns](FORMULA_PATTERNS.md) (coming soon)
- [Cask Patterns](CASK_PATTERNS.md) (coming soon)
- [Testing Patterns](TESTING_PATTERNS.md) (coming soon)

---

**Last Updated:** 2025-02-08
**Repository:** https://github.com/castrojo/homebrew-tap
