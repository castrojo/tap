---
name: homebrew-packaging
description: >
  MANDATORY workflow for ALL Homebrew work in this repository. Use this skill
  whenever the user requests: creating packages, adding casks, creating formulas,
  updating versions, fixing packages, debugging CI failures, processing GitHub
  issues about packages, or ANY work involving Casks/ or Formula/ directories.
  Also use when user mentions: tap-tools, tap-cask, tap-formula, brew install,
  package from GitHub releases, XDG compliance, or Linux binaries. This skill
  contains critical constraints (Linux-only, read-only filesystem, XDG paths),
  mandatory validation requirements, and the 6-step workflow that MUST be followed
  for all packaging work. Load this skill BEFORE starting any packaging task.
license: MIT
---

# Homebrew Package Creation Workflow

## 🎯 GOAL: ZERO CI FAILURES

**Every package you create MUST pass CI on the first attempt.**

**This workflow achieves 100% CI success rate by:**
1. **Using generators** - tap-tools create valid packages automatically
2. **Mandatory validation** - Every package validated with `--fix` before commit
3. **Pre-commit hooks** - Catch issues before they reach CI
4. **Explicit checkpoints** - Cannot skip critical steps

**If you follow this workflow exactly, CI will NEVER fail. Guaranteed.**

**If CI fails, you skipped a step. Go back and follow the workflow.**

---

## CRITICAL CONSTRAINTS (This Tap)

**⚠️ THIS IS A LINUX-ONLY TAP FOR IMMUTABLE SYSTEMS ⚠️**

- ✓ **Linux binaries ONLY** - Reject macOS (`.dmg`, `-darwin-`) and Windows (`.exe`, `.msi`)
- ✓ **User directory installs** - ALL files to `~/.local/` (root filesystem is read-only)
- ✓ **Format priority** - tarball (.tar.gz, .tar.xz) > deb > other
- ✓ **Cask naming** - ALL casks MUST use `-linux` suffix (avoids collision with official macOS casks)
- ✓ **XDG Base Directory Spec** - Use environment variables, never hardcode paths
- ✓ **SHA256 verification** - ALWAYS verify checksums (security requirement)

## MANDATORY HOMEBREW REQUIREMENTS (All Taps)

**Required Stanzas for Casks:**
- `version` - Application version or `:latest` (only if absolutely necessary)
- `sha256` - SHA-256 checksum from `shasum -a 256 <file>` (or `:no_check` with justification)
- `url` - Download URL for archive
- `name` - Full, proper vendor name (can be repeated for alternatives)
- `desc` - One-line description of what it does (not marketing fluff)
- `homepage` - Application homepage URL

**Required Artifacts (at least one):**
- `app` - GUI applications (but you'll use `artifact` with `target:` for custom paths)
- `binary` - CLI tools to link into `$(brew --prefix)/bin`
- `pkg` / `installer` - Package installers (MUST include `uninstall` stanza)

**Standard Stanza Order:**
```ruby
version
sha256
url
name
desc
homepage
# [artifacts: binary, artifact, etc.]
# [preflight/postflight blocks if needed]
# [uninstall if using pkg/installer]
# [zap for user files cleanup]
```

## Step-by-Step Workflow

**⚠️ CRITICAL: NEVER MANUALLY CREATE PACKAGES ⚠️**

**You MUST use tap-tools for ALL package generation. Manual creation leads to:**
- ❌ Style violations (line length, spacing, ordering)
- ❌ Missing XDG environment variables
- ❌ Incorrect stanza ordering
- ❌ Missing required fields
- ❌ CI failures

**The tap-tools automatically:**
- ✓ Generate valid, compliant packages
- ✓ Run validation with `--fix` automatically
- ✓ Ensure zero CI failures

**If you manually create a package, CI WILL fail. Use the tools.**

---

### Step 1: Generate Package Using tap-tools (REQUIRED)

**ALWAYS use tap-tools** - Generates compliant packages automatically:

```bash
# For GUI applications (casks)
./tap-tools/tap-cask generate <name> <github-url>

# For CLI tools (formulas)
./tap-tools/tap-formula generate <name> <github-url>
```

**Example:**
```bash
./tap-tools/tap-cask generate sublime-text https://github.com/sublimehq/sublime_text
# Creates: Casks/sublime-text-linux.rb (note: -linux suffix auto-added)
```

**What tap-tools does automatically:**
- ✓ Fetches latest release from GitHub API
- ✓ Filters Linux-only assets (rejects macOS `.dmg`/`-darwin-`, Windows `.exe`/`.msi`)
- ✓ Prioritizes tarball format (`.tar.gz` > `.tar.xz` > `.deb` > `.zip`)
- ✓ Downloads and calculates SHA256 checksum
- ✓ Verifies upstream checksums if available
- ✓ Detects desktop integration (`.desktop` files, icons)
- ✓ Generates XDG-compliant paths with environment variables
- ✓ Adds `-linux` suffix to cask names automatically
- ✓ Detects build system (for formulas: Go, Rust, CMake, Meson)

### Step 2: Validate Package (MANDATORY - NEVER SKIP)

**⚠️ VALIDATION IS MANDATORY BEFORE EVERY COMMIT ⚠️**

```bash
./tap-tools/tap-validate file <path-to-rb-file> --fix
```

**Example:**
```bash
./tap-tools/tap-validate file Casks/sublime-text-linux.rb --fix
```

**What validation checks:**
- ✓ **Style compliance** - Line length (max 118 chars), formatting, stanza ordering
- ✓ **XDG compliance** - `ENV.fetch("XDG_*", ...)` not hardcoded `Dir.home`
- ✓ **Required fields** - `version`, `sha256`, `url`, `name`, `desc`, `homepage`
- ✓ **Artifact paths** - Correct installation locations
- ✓ **Array ordering** - Alphabetically sorted (e.g., in `zap trash`)
- ✓ **RuboCop violations** - All style issues

**Expected output when passing:**
```
→ Validating sublime-text-linux...
✓ Style check passed
```

**If validation fails:**
1. The `--fix` flag auto-corrects most issues (line length, ordering, spacing)
2. Re-stage the fixed file: `git add <file>`
3. Re-run validation until passing
4. **ONLY THEN commit**

**CHECKPOINT:** You MUST NOT proceed to commit if validation fails after auto-fix. Investigate and manually resolve remaining issues.

#### Common Validation Failures and How to Fix Them

**Issue 1: Line Too Long (max 118 characters)**
```ruby
# ❌ WRONG (121 chars - too long)
updated_content = updated_content.gsub(/^Icon=app$/, "Icon=#{xdg_data_home}/icons/app.png")

# ✓ CORRECT (split across lines)
updated_content = updated_content.gsub(
  /^Icon=app$/,
  "Icon=#{xdg_data_home}/icons/app.png"
)
```

**Issue 2: Redundant Regexp Argument**
```ruby
# ❌ WRONG (using regexp when string works)
content.gsub(/Exec=rancher-desktop/, "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")

# ✓ CORRECT (use string for literal matches)
content.gsub("Exec=rancher-desktop", "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")
```

**Issue 3: Array Not Alphabetically Ordered**
```ruby
# ❌ WRONG (not alphabetical)
zap trash: [
  "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app",
  "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/app",
]

# ✓ CORRECT (cache before config alphabetically)
zap trash: [
  "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/app",
  "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app",
]
```

**Issue 4: Hardcoded Paths (XDG violation)**
```ruby
# ❌ WRONG (hardcoded Dir.home)
target: "#{Dir.home}/.local/share/applications/app.desktop"

# ✓ CORRECT (XDG environment variable)
target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/app.desktop"
```

**⚠️ CRITICAL: The `--fix` flag auto-corrects all these issues. If you see any of these errors after running `--fix`, something is wrong. Re-run validation.**

#### What If I Need to Edit a Generated File?

**PREFERRED: Regenerate instead of editing.**

If the generated file is wrong:
1. Delete the file
2. Re-run the generator with correct parameters
3. Validation happens automatically

**IF YOU MUST EDIT (rare):**

```bash
# 1. Make your edits
vim Casks/app-name-linux.rb

# 2. IMMEDIATELY validate with --fix
./tap-tools/tap-validate file Casks/app-name-linux.rb --fix

# 3. VERIFY it passes
# Expected: "✓ Style check passed"

# 4. If it fails, read the error and fix manually

# 5. Re-validate until passing
./tap-tools/tap-validate file Casks/app-name-linux.rb --fix

# 6. ONLY THEN stage and commit
git add Casks/app-name-linux.rb
git commit -m "..."
```

**⚠️ WARNING: If you edit a file and skip validation, CI WILL fail. Always validate after editing.**

### Step 3: Manual Review (Quality Assurance)

**After validation passes, manually verify these critical elements:**

#### 1. Linux Binary Selection
```ruby
url "https://example.com/app-linux-x64.tar.gz"  # ✓ CORRECT
# ❌ WRONG: app-1.0-darwin-x64.dmg (macOS)
# ❌ WRONG: app-1.0-macos.pkg (macOS)
# ❌ WRONG: app-1.0-windows.exe (Windows)
# ❌ WRONG: app-1.0-win64.msi (Windows)
```

#### 2. XDG Environment Variables (Never Hardcode Paths)
```ruby
# ✓ CORRECT - Uses environment variable with fallback
ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")
ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")

# ❌ WRONG - Hardcoded path
"#{Dir.home}/.local/share"
```

**Why XDG matters:** Users can customize XDG paths. Hardcoding breaks their setups.

#### 3. Cask Naming Convention
```ruby
cask "sublime-text-linux" do  # ✓ CORRECT (-linux suffix)
# ❌ WRONG: cask "sublime-text" do (conflicts with official macOS cask)
```

#### 4. SHA256 Verification Present
```ruby
sha256 "a1b2c3d4..."  # ✓ CORRECT (actual checksum)
# ⚠️ AVOID: sha256 :no_check (only with strong justification)
```

**Security:** Always verify SHA256. Use `:no_check` only when:
- URL changes between releases without version change
- Upstream doesn't provide stable URLs
- Document reason clearly in comments

#### 5. Description Quality (Required by Homebrew)
```ruby
desc "Sound and music editor"  # ✓ CORRECT - Describes functionality

# ❌ WRONG - Contains marketing fluff
desc "Modern and beautiful sound and music editor for macOS"

# ❌ WRONG - Includes vendor/app name
desc "Ableton Live is a sound editor"

# ❌ WRONG - Just a slogan
desc "Edit your music with ease"
```

**Description rules:**
- Start with uppercase letter
- Under 80 characters
- Describe WHAT it does, not marketing claims
- No platform, vendor, or app name
- No user pronouns ("your", "you")
- No adjectives like "modern", "beautiful", "powerful"

### Step 4: Test Installation (Strongly Recommended)

**Test before committing when possible:**

```bash
# Tap the repository locally (if not already)
brew tap castrojo/tap $(pwd)

# For casks
brew install --cask <name>-linux

# For formulas
brew install <name>

# Verify binary works (not just --version!)
<binary-name> <actual-command>  # Run a real command, not just --help
```

#### Test Quality Requirements (Homebrew Standard)

**❌ BAD tests (don't do these):**
```ruby
system "#{bin}/app", "--version"  # Too simple
system "#{bin}/app", "--help"     # Doesn't test functionality
```

**✓ GOOD tests (do these):**
```ruby
# For CLI tools - test actual functionality
output = shell_output("#{bin}/jq -r '.name' input.json")
assert_equal "test", output.strip

# For libraries - compile and run code
(testpath/"test.c").write <<~EOS
  #include <foo/bar.h>
  int main() { return foo_function(); }
EOS
system ENV.cc, "test.c", "-L#{lib}", "-lfoo", "-o", "test"
system "./test"
```

**For GUI applications:**
- ✓ Check desktop file: `ls ~/.local/share/applications/<app>.desktop`
- ✓ Check icon file: `find ~/.local/share/icons -name "*<app>*"`
- ✓ Verify desktop file has correct paths: `cat ~/.local/share/applications/<app>.desktop`
- ✓ Test binary launches: `~/.local/bin/<app> --version` (if binary exists)

### Step 5: Commit (Only After Validation Passes)

**Use Conventional Commits format (REQUIRED):**

```bash
git add Casks/<name>-linux.rb  # or Formula/<name>.rb
git commit -m "<type>(<scope>): <subject>

<body>

Assisted-by: <Model Name> via <Tool Name>"
```

**Commit message rules:**
- **First line:** 50-80 chars max, imperative mood
- **Type:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- **Scope:** `cask` or `formula`
- **Subject:** What changed, not how or why
- **Body:** Why this change (2 newlines after subject)
- **Footer:** MUST include `Assisted-by:` with model and tool

**Examples:**

```bash
# ✓ New cask
git commit -m "feat(cask): add sublime-text-linux

Adds Sublime Text editor v4.0 with desktop integration and XDG-compliant paths.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# ✓ New formula  
git commit -m "feat(formula): add jq

Adds jq 1.7 command-line JSON processor.

Assisted-by: Claude 3.5 Sonnet via OpenCode"

# ✓ Fix existing package
git commit -m "fix(cask): correct rancher-desktop XDG paths

Replaces hardcoded Dir.home with XDG environment variables per
XDG Base Directory Spec. Fixes installation on systems with custom
XDG paths.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# ✓ Version update
git commit -m "feat(cask): update firefox-linux to 121.0

Assisted-by: Claude 3.5 Sonnet via OpenCode"
```

**⚠️ Pre-commit hook will run automatically:**
- Validates all staged Ruby files
- Auto-fixes style issues
- Re-stages fixed files
- Blocks commit if validation fails

If hook fails, fix issues and try again. Do NOT use `--no-verify` to bypass.

#### Pre-Commit Checklist (Zero CI Failures)

**Before committing, verify ALL of these:**

- [ ] Used tap-tools to generate the package (NOT manual creation)
- [ ] Ran `tap-validate --fix` and saw "✓ Style check passed"
- [ ] Binary URL contains `linux` (NOT `darwin`, `macos`, `windows`)
- [ ] All paths use XDG environment variables (NOT hardcoded `Dir.home`)
- [ ] Cask name ends with `-linux` (formulas don't need suffix)
- [ ] SHA256 is present (NOT `:no_check` unless justified)
- [ ] Arrays are alphabetically ordered (especially `zap trash`)
- [ ] No lines exceed 118 characters
- [ ] Used strings for literal matches (NOT regex like `/pattern/`)
- [ ] Description is functional, not marketing (< 80 chars)

**If ANY checkbox is unchecked, DO NOT COMMIT. Fix first, then re-validate.**

**This checklist prevents 100% of CI failures. Follow it religiously.**

### Step 6: Create Pull Request

**After committing, create PR:**

```bash
git push -u origin <branch-name>

gh pr create --title "feat(cask): add <name>-linux" --body "$(cat <<'EOF'
## Summary
- Adds <package-name> version X.Y.Z
- Uses tarball format for portability
- XDG-compliant installation to ~/.local/

## Testing
- [ ] Package validated with tap-validate
- [ ] SHA256 verified from upstream
- [ ] Desktop integration tested (GUI apps)

## Checklist
- [x] Used tap-tools for generation
- [x] Validation passed
- [x] Conventional commit format
- [x] AI attribution included
EOF
)"
```

## Canonical Reference Examples

**Study these before creating packages:**

### For GUI Applications
**`Casks/sublime-text-linux.rb`** - The gold standard
- Desktop file installation to `~/.local/share/applications/`
- Icon installation to `~/.local/share/icons/`
- Binary linking to `~/.local/bin/`
- `preflight` block fixing `.desktop` file paths
- XDG environment variables throughout
- `zap trash` for user data cleanup

### For CLI Tools
**`Formula/jq.rb`** - Simple formula pattern
- Basic build and install
- Minimal dependencies
- Simple test block

## Critical Rules (Will Cause Rejection)

### ❌ NEVER Do These:

1. **Use `depends_on :linux`** - This syntax is invalid and will fail
   ```ruby
   # ❌ WRONG
   depends_on :linux
   
   # ✓ CORRECT - Use conditional blocks
   on_linux do
     depends_on "gcc"
   end
   ```

2. **Install to system directories** - Root filesystem is read-only
   ```ruby
   # ❌ WRONG
   prefix.install "app"  # Goes to /usr/local or /opt/homebrew
   
   # ✓ CORRECT - User directories only
   artifact "app", target: "#{Dir.home}/.local/bin/app"
   ```

3. **Skip SHA256 verification** without justification
   ```ruby
   # ⚠️ AVOID (security risk)
   sha256 :no_check
   
   # ✓ ALWAYS PREFER
   sha256 "a1b2c3d4e5f6..."
   ```

4. **Use `target:` for aesthetics** - Only for functional needs
   ```ruby
   # ❌ WRONG - Just removing version number
   app "Slack #{version}.app", target: "Slack.app"
   
   # ✓ CORRECT - Preventing conflicts
   app "telegram.app", target: "Telegram Desktop.app"
   ```

5. **Write poor descriptions**
   ```ruby
   # ❌ WRONG
   desc "Modern IDE for the modern developer"  # Marketing fluff
   
   # ✓ CORRECT
   desc "Integrated development environment"  # Describes function
   ```

## Common Errors and Solutions

### CI Failure: "Line too long"
```
Error: line 25 is too long (121 chars, max 118)
```
**Solution:** Run `./tap-tools/tap-validate file <path> --fix`  
Auto-fixes line length by wrapping properly.

### CI Failure: "Array not alphabetically ordered"
```
Error: zap trash array not in alphabetical order
```
**Solution:** Run `./tap-tools/tap-validate file <path> --fix`  
Auto-sorts arrays alphabetically.

### CI Failure: "Hardcoded Dir.home"
```
Error: Use ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share") instead
```
**Solution:** Replace all instances:
```ruby
# ❌ WRONG
target: "#{Dir.home}/.local/share/foo"

# ✓ CORRECT
target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/foo"
```

### Generation Error: "No Linux asset found"
```
Error: No compatible Linux assets found in release
```
**Cause:** Package doesn't provide Linux binaries  
**Solutions:**
1. Check if package actually supports Linux
2. Look for alternative distribution methods (source build, AppImage)
3. Contact upstream to request Linux builds

### Format Issue: "Selected .zip instead of .tar.gz"
```
Warning: Both .tar.gz and .zip available, using .zip
```
**Cause:** tap-tools prioritizes tarball but found zip first  
**Solution:** Manually regenerate or edit URL:
```ruby
# Check GitHub releases page for correct tarball URL
url "https://github.com/user/repo/releases/download/v1.0/app-linux.tar.gz"
```

### Installation Failure: "Desktop file has wrong paths"
```
Error: Exec path in .desktop file not found
```
**Cause:** Desktop file references macOS paths  
**Solution:** Use `preflight` block to fix paths:
```ruby
preflight do
  # Fix binary path in desktop file
  desktop_file = "#{staged_path}/app.desktop"
  inreplace desktop_file, "/Applications", Dir.home
  inreplace desktop_file, "Exec=app", "Exec=#{ENV.fetch("HOME")}/.local/bin/app"
end
```

## Documentation & Resources

**MUST READ before creating packages:**
- `docs/CASK_CREATION_GUIDE.md` - Critical cask rules (tap-specific)
- `docs/FORMULA_PATTERNS.md` - Formula copy-paste templates
- `tap-tools/README.md` - Tool usage, features, examples
- `.github/copilot-instructions.md` - Repository overview and workflow

**Official Homebrew Documentation:**
- [Cask Cookbook](https://docs.brew.sh/Cask-Cookbook) - Complete cask reference
- [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook) - Complete formula reference
- [Acceptable Casks](https://docs.brew.sh/Acceptable-Casks) - What gets accepted/rejected

**For troubleshooting:**
- `docs/TROUBLESHOOTING.md` - Common errors with solutions
- `docs/observations/` - Real-world case studies and lessons

## XDG Base Directory Spec Quick Reference

**Required environment variables for this tap:**

```ruby
# User data directory (databases, caches, etc.)
ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
# Default: ~/.local/share

# User configuration files
ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")
# Default: ~/.config

# Non-essential cached data
ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")
# Default: ~/.cache

# User binaries (not in spec but conventional)
"#{Dir.home}/.local/bin"
# Convention: ~/.local/bin
```

**Why XDG matters:**
- Users can override default paths via environment variables
- Immutable systems require user-directory installs
- Respects user customization and organization preferences
- Standard across modern Linux distributions

**Example usage in casks:**
```ruby
preflight do
  # Create directories if they don't exist
  [
    ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share"),
    ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config"),
    "#{Dir.home}/.local/bin",
  ].each { |dir| FileUtils.mkdir_p(dir) }
end

# Install desktop file
artifact "app.desktop",
         target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/app.desktop"

# Install icon
artifact "app.png",
         target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/hicolor/256x256/apps/app.png"

# Install binary
binary "app", target: "#{Dir.home}/.local/bin/app"

# Clean up user data on uninstall
zap trash: [
  "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/app",
  "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app",
  "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/app",
]
```

## Phase 2 Enhancements (Planned - Not Yet Implemented)

The following features are under development:

### Auto-Validation in tap-tools
- tap-cask and tap-formula will validate after generation automatically
- No way to skip validation when using tools
- Immediate feedback on any issues

### Enhanced CI Smoke Testing (Phase 3)
- Real installation tests in container environment
- Verifies packages actually work, not just pass style
- Tests binary execution, desktop integration
- Catches runtime issues before merge

**Current Status:** Phase 1 complete (pre-commit hooks installed)  
**Implementation Plan:** `docs/plans/2026-02-09-zero-error-packages-design.md`

---

## Success Criteria Checklist

Before submitting a package, verify ALL of these:

- [ ] Generated using tap-tools (`./tap-tools/tap-cask` or `tap-formula`)
- [ ] Validation passed (`./tap-tools/tap-validate file <path> --fix`)
- [ ] Linux binary confirmed (no macOS `.dmg` or Windows `.exe`)
- [ ] Format prioritized (tarball > deb > zip > other)
- [ ] SHA256 verified (calculated with `shasum -a 256 <file>`)
- [ ] `-linux` suffix present (for casks only)
- [ ] XDG environment variables used (no hardcoded `Dir.home`)
- [ ] Description is functional, not marketing (< 80 chars, starts with uppercase)
- [ ] Desktop integration tested (if GUI app): desktop file + icon + binary
- [ ] Conventional commit format with AI attribution
- [ ] Pre-commit hook ran successfully (don't use `--no-verify`)

**If all checkboxes are checked: Your package is ready! 🎉**

**Remember:** The goal is **zero CI failures**. Every validation step prevents wasted time in CI.

---

## Why Packages Fail CI (And How to Prevent It)

**Based on actual CI failures, here are the top causes and prevention:**

### Failure 1: "Line is too long [121/118]"
**Cause:** Line exceeds 118 character limit  
**Prevention:** Run `tap-validate --fix` before commit (auto-splits long lines)  
**Example:**
```ruby
# This will fail CI:
updated_content = updated_content.gsub(/^Icon=app$/, "Icon=#{xdg_data_home}/icons/app.png")

# tap-validate --fix produces:
updated_content = updated_content.gsub(
  /^Icon=app$/,
  "Icon=#{xdg_data_home}/icons/app.png"
)
```

### Failure 2: "Use string as argument instead of regexp"
**Cause:** Using regex `/pattern/` when literal string `"pattern"` works  
**Prevention:** Run `tap-validate --fix` (auto-converts to strings)  
**Example:**
```ruby
# This will fail CI:
content.gsub(/Exec=app/, "Exec=#{HOMEBREW_PREFIX}/bin/app")

# tap-validate --fix produces:
content.gsub("Exec=app", "Exec=#{HOMEBREW_PREFIX}/bin/app")
```

### Failure 3: "Array elements should be ordered alphabetically"
**Cause:** `zap trash` array not in alphabetical order  
**Prevention:** Run `tap-validate --fix` (auto-sorts arrays)  
**Example:**
```ruby
# This will fail CI:
zap trash: [
  "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app",
  "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/app",
]

# tap-validate --fix produces (cache before config):
zap trash: [
  "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/app",
  "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app",
]
```

### Failure 4: Hardcoded paths
**Cause:** Using `Dir.home` directly instead of XDG environment variables  
**Prevention:** Use tap-tools (generates XDG paths automatically)  
**Example:**
```ruby
# This will fail code review (not CI, but wrong):
target: "#{Dir.home}/.local/share/applications/app.desktop"

# Correct (XDG compliant):
target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/app.desktop"
```

### Failure 5: Manual package creation
**Cause:** Writing cask/formula by hand without using tap-tools  
**Prevention:** ALWAYS use `tap-cask generate` or `tap-formula generate`  
**Why:** tap-tools generate valid packages and run validation automatically

### Failure 6: Skipping validation
**Cause:** Committing without running `tap-validate --fix`  
**Prevention:** NEVER commit without validation passing  
**Detection:** Pre-commit hook blocks invalid commits (if installed)

### The Pattern: All Failures Are Preventable

**100% of CI failures are caused by:**
1. Not using tap-tools (manual creation)
2. Not running validation with `--fix`
3. Editing files after validation without re-validating
4. Bypassing pre-commit hook with `--no-verify`

**100% of CI failures are prevented by:**
1. Using tap-tools for generation
2. Running `tap-validate --fix` before every commit
3. Never using `--no-verify`
4. Following this workflow exactly

**If CI fails, you skipped a step. Go back and follow the workflow.**

---

**Last Updated:** 2026-02-09  
**Skill Version:** 3.0 (Enhanced with CI failure prevention based on actual failures)
