# Copilot Instructions for homebrew-tap

## Repository Overview

This is a **Linux-only** Homebrew tap for packages unavailable or incompatible with the official Homebrew repositories. The tap targets **immutable/read-only filesystem distributions** (Fedora Silverblue, Universal Blue) where system directories like `/usr/`, `/opt/`, and `/etc/` are read-only.

**Key Constraints:**
- ALL packages MUST use Linux binaries only (never macOS or Windows)
- ALL file installations MUST go to user home directory (`~/.local/`, `~/.config/`, `~/.cache/`)
- NEVER install to system directories (`/usr/`, `/opt/`, `/etc/`)

## Repository Structure

```
homebrew-tap/
├── .github/
│   ├── workflows/          # CI/CD automation
│   ├── ISSUE_TEMPLATE/     # Package request template
│   ├── renovate.json5      # Automatic version updates
│   └── labeler.yml         # PR labeling
├── Casks/                  # GUI applications (*.rb)
├── Formula/                # CLI tools (*.rb)
├── tap-tools/              # Go CLI tools (PREFERRED)
│   ├── tap-cask           # Generate casks from GitHub releases
│   ├── tap-formula        # Generate formulas from GitHub releases
│   ├── tap-issue          # Process package requests from issues
│   └── tap-validate       # Validate all packages
├── scripts/                # Legacy bash scripts (deprecated)
├── docs/
│   ├── AGENT_GUIDE.md            # Comprehensive agent workflow
│   ├── CASK_CREATION_GUIDE.md    # CRITICAL: Read before creating casks
│   ├── FORMULA_PATTERNS.md       # Copy-paste formula templates
│   ├── CASK_PATTERNS.md          # Copy-paste cask templates
│   └── TROUBLESHOOTING.md        # Common errors and solutions
└── AGENTS.md              # This file (agent instructions)
```

## Critical Documentation (READ FIRST)

**BEFORE creating any package, you MUST read:**
1. `docs/CASK_CREATION_GUIDE.md` - Contains critical rules that prevent CI failures
2. `AGENTS.md` - Contains Linux-only requirements, XDG paths, and workflow

**Key points from CASK_CREATION_GUIDE.md:**
- ❌ NEVER use `depends_on :linux` (causes errors)
- ❌ NEVER use `test do ... end` blocks in casks
- ✅ ALWAYS use XDG environment variables for paths
- ✅ ALWAYS include SHA256 verification
- ✅ ALWAYS use `-linux` suffix for cask names

## Build, Test, and Validation Workflow

### 1. Creating Packages (ALWAYS use tap-tools)

**For GUI applications (casks):**
```bash
./tap-tools/tap-cask generate https://github.com/user/repo
```

**For CLI tools (formulas):**
```bash
./tap-tools/tap-formula generate https://github.com/user/repo
```

**From GitHub issues:**
```bash
./tap-tools/tap-issue process <issue-number>
./tap-tools/tap-issue process <issue-number> --create-pr
```

**Benefits of tap-tools:**
- Automatically selects Linux-only assets (rejects macOS/Windows)
- Prioritizes formats: tarball > deb > other
- Calculates and verifies SHA256 checksums
- Detects desktop integration needs
- 4-5x faster than bash scripts
- Ensures XDG compliance

### 2. Validation (ALWAYS run before committing)

```bash
# Validate specific file
./tap-tools/tap-validate file Casks/app-name-linux.rb
./tap-tools/tap-validate file Casks/app-name-linux.rb --fix  # Auto-fix style

# Validate all packages
./tap-tools/tap-validate all
./tap-tools/tap-validate all --fix
```

**Expected output:**
- ✓ Style check passed
- Audit check may show path errors (expected - requires tapping first)

### 3. Testing Installation (Optional but recommended)

```bash
# Install cask
brew install --cask castrojo/tap/app-name-linux

# Install formula
brew install castrojo/tap/tool-name

# Uninstall
brew uninstall --cask app-name-linux  # or tool-name
```

### 4. Commit Format (MANDATORY)

Use **Conventional Commits** with AI attribution:

```
<type>[optional scope]: <description>

[optional body]

Assisted-by: <Model> via <Tool>
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `ci`

**Examples:**
```
feat(cask): add rancher-desktop-linux v1.15.0

Adds Rancher Desktop with desktop integration and XDG paths.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot
```

```
fix(formula): correct jq binary installation path

Assisted-by: GPT-4 via GitHub Copilot
```

### 5. Pull Request Workflow (REQUIRED for major work)

**ALL major features/epics MUST use pull requests, NOT direct commits to main.**

```bash
# Create feature branch
git checkout -b feature/add-rancher-desktop

# Make changes, validate, commit
./tap-tools/tap-cask generate https://github.com/rancher-sandbox/rancher-desktop
./tap-tools/tap-validate file Casks/rancher-desktop-linux.rb --fix
git add Casks/rancher-desktop-linux.rb
git commit -m "feat(cask): add rancher-desktop-linux v1.15.0

Adds Rancher Desktop with desktop integration and XDG paths.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# Push and create PR
git push -u origin feature/add-rancher-desktop
gh pr create --title "feat(cask): add rancher-desktop-linux" --body "$(cat <<'EOF'
## Summary
- Adds Rancher Desktop v1.15.0 as Linux cask
- Uses .deb package format (no tarball available)
- Includes desktop integration with XDG paths
- Passes style validation

## Testing
- Validated with tap-validate --fix
- Style check passed
EOF
)"
```

**Why Pull Requests:**
- Enables Gemini Code Assist automatic reviews
- Creates discussion space for complex changes
- Allows CI validation before merge
- Documents feature development history

## Package Format Priority

When selecting downloads, follow this **strict priority order:**

1. **Tarball (PREFERRED)** - `.tar.gz`, `.tar.xz`, `.tgz`
   - Most portable across distributions
   - Simple extraction
   - Example: `app-linux-x64.tar.gz`

2. **Debian Package (SECOND CHOICE)** - `.deb`
   - Use only if no tarball available
   - Requires extraction via `ar` and `tar`
   - Example: `app_amd64.deb`

3. **Other formats** - Only with justification
   - AppImage, snap, flatpak: Case-by-case
   - RPM: Generally avoid

## XDG Base Directory Specification (CRITICAL)

**ALWAYS use XDG environment variables, NEVER hardcoded paths:**

```ruby
# ✅ CORRECT
target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/app.desktop"

# ❌ WRONG
target: "#{Dir.home}/.local/share/applications/app.desktop"
```

**Standard XDG directories:**
- `$XDG_DATA_HOME` (default: `~/.local/share`) - Application data, desktop files, icons
- `$XDG_CONFIG_HOME` (default: `~/.config`) - Configuration files
- `$XDG_CACHE_HOME` (default: `~/.cache`) - Cache data
- `$XDG_STATE_HOME` (default: `~/.local/state`) - State data, logs

## Desktop Integration (GUI Applications)

**ALL GUI applications MUST install:**
1. Desktop file (`.desktop`) to `$XDG_DATA_HOME/applications/`
2. Icon to `$XDG_DATA_HOME/icons/`

**Example pattern (see sublime-text-linux.rb for reference):**

```ruby
cask "app-name-linux" do
  version "1.0.0"
  sha256 "abc123..."

  url "https://example.com/app-linux-x64.tar.gz"
  name "App Name"
  desc "One-line description"
  homepage "https://example.com/"

  binary "app/bin/app"
  artifact "app/app.desktop",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/app.desktop"
  artifact "app/icon.png",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/app.png"

  preflight do
    xdg_data_home = ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
    FileUtils.mkdir_p "#{xdg_data_home}/applications"
    FileUtils.mkdir_p "#{xdg_data_home}/icons"

    desktop_file = "#{staged_path}/app/app.desktop"
    if File.exist?(desktop_file)
      content = File.read(desktop_file)
      updated_content = content.gsub(%r{/opt/app/app}, "#{HOMEBREW_PREFIX}/bin/app")
      File.write(desktop_file, updated_content)
    end
  end

  zap trash: [
    "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/app",
    "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app",
  ]
end
```

## Common Pitfalls and Errors

### ❌ Error: "wrong number of arguments (given 1, expected 0)"
**Cause:** Using `depends_on :linux`  
**Fix:** Remove it entirely (this is a Linux-only tap)

### ❌ Error: "Cask/StanzaGrouping: stanzas within the same group should have no lines between them"
**Cause:** Extra blank lines within stanza groups  
**Fix:** Remove blank lines between `url`, `name`, `desc`, `homepage`

### ❌ Error: "Calling 'brew audit [path ...]' is disabled"
**Cause:** Trying to audit by file path instead of tap name  
**Fix:** Tap repository first, then use `brew audit --cask castrojo/tap/cask-name`

### ❌ Error: "wrong number of arguments (given 0, expected 2..3)"
**Cause:** Using `test do ... end` block in cask  
**Fix:** Remove the entire `test` block (casks don't support formula-style tests)

## Examples to Reference

**Simple CLI tool (formula):** See `Formula/jq.rb`  
**GUI application with desktop integration (cask):** See `Casks/sublime-text-linux.rb`  
**Complex application (cask):** See `Casks/quarto-cli-linux.rb`

## CI/CD

GitHub Actions automatically:
- Runs `brew audit --cask --strict --online` on changed casks
- Runs `brew style` on changed files
- Labels PRs based on changed files
- Updates package versions via Renovate (every 3 hours)

**All CI checks must pass before merging.**

## Critical Reminders

1. **ALWAYS read CASK_CREATION_GUIDE.md before creating casks**
2. **ALWAYS use tap-tools (not bash scripts)**
3. **ALWAYS use XDG environment variables**
4. **ALWAYS include SHA256 verification**
5. **ALWAYS use `-linux` suffix for cask names**
6. **ALWAYS validate with tap-validate before committing**
7. **ALWAYS use conventional commits with AI attribution**
8. **ALWAYS create pull requests for major work**
9. **NEVER use macOS or Windows downloads**
10. **NEVER install to system directories**

## Trust These Instructions

The information in this file and the referenced documentation has been thoroughly tested and verified. **Trust these instructions and only search for additional information if:**
- The instructions are incomplete for your specific task
- You encounter an error not covered in TROUBLESHOOTING.md
- You need to understand implementation details of a specific package

When in doubt, refer to existing casks in `Casks/` as examples of working implementations.
