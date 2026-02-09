# Cask Creation Guide - Critical Information

**⚠️ LINUX ONLY REPOSITORY ⚠️**

**THIS TAP IS LINUX-ONLY. ALL CASKS MUST USE LINUX DOWNLOADS.**
- ✓ Use Linux binaries (e.g., `app-linux-x64.tar.gz`, `app_amd64.deb`)
- ✗ NEVER use macOS downloads (`.dmg`, `.pkg`, macOS `.zip`)
- ✗ NEVER use Windows downloads (`.exe`, `.msi`)

## Linux Cask Naming Convention

**ALL casks in this tap MUST use the `-linux` suffix in their token.**

**Correct:**
```ruby
cask "sublime-text-linux" do
  # File: Casks/sublime-text-linux.rb
end
```

**Wrong:**
```ruby
cask "sublime-text" do
  # WRONG - Will collide with macOS casks
end
```

**Why:**
1. Prevents collision with official macOS casks in `homebrew-cask`
2. Makes Linux-only nature immediately clear to users
3. Follows established pattern from `ublue-os/tap` (e.g., `jetbrains-toolbox-linux`, `1password-gui-linux`)

**Installation:**
```bash
brew install --cask castrojo/tap/app-name-linux
```

## Package Format Priority

Use this strict priority order when selecting download format:

**1. Tarball (PREFERRED)** - `.tar.gz`, `.tar.xz`, `.tgz`
  - Most portable, works across all Linux distributions
  - Simple extraction and installation
  - No package manager dependencies
  - Example: `app-linux-x64.tar.gz`

**2. Debian Package (SECOND CHOICE)** - `.deb`
  - Use only if no tarball is available
  - Requires extraction via `ar` and `tar`
  - May have distro-specific dependencies
  - Example: `app_amd64.deb`

**3. Other formats** - Only with explicit justification
  - AppImage, snap, flatpak: Case-by-case basis
  - RPM: Generally avoid (requires conversion)

## Checksum Verification (MANDATORY)

**EVERY cask MUST include SHA256 verification. NO EXCEPTIONS.**

### Step-by-Step Verification Process:

```bash
# 1. Download the file
curl -LO https://example.com/app-linux-x64.tar.gz

# 2. Calculate SHA256
sha256sum app-linux-x64.tar.gz
# Output: abc123def456... app-linux-x64.tar.gz

# 3. If upstream provides checksums, verify against them
curl -LO https://example.com/SHA256SUMS
grep app-linux-x64.tar.gz SHA256SUMS
# Compare with your calculated hash

# 4. Use the verified hash in your cask
```

### Cask Example with SHA256:

```ruby
cask "app-name" do
  version "1.0.0"
  sha256 "abc123def456..."  # MANDATORY - calculated from actual download
  
  url "https://example.com/app-linux-x64.tar.gz"
  name "App Name"
  desc "Description"
  homepage "https://example.com"
  
  binary "app-name"
end
```

### SHA256 Rules:

**DO:**
- ✓ Download the actual file and calculate its SHA256
- ✓ Verify against upstream checksums if available
- ✓ Use lowercase hexadecimal (64 characters)
- ✓ Include SHA256 on the line immediately after `version`

**DON'T:**
- ✗ Skip SHA256 verification
- ✗ Use `sha256 :no_check` (only acceptable with documented justification)
- ✗ Copy SHA256 from untrusted sources
- ✗ Use checksums for different architectures/platforms

**Last Updated:** 2026-02-09  
**Homebrew Version:** Current (2026)  
**Status:** TESTED AND VERIFIED

This guide documents critical, up-to-date information for creating Homebrew casks that will pass CI on the first run. All information here has been verified against actual Homebrew behavior and documentation.

## 🚨 Critical Rules - Read These First

### 1. DO NOT Use `depends_on :linux`

**WRONG:**
```ruby
depends_on :linux
```

**ERROR:** `wrong number of arguments (given 1, expected 0)`

**WHY:** The `depends_on` stanza for casks does NOT accept `:linux` as a symbol argument. This syntax is for formulas only.

**CORRECT OPTIONS:**

#### Option A: For Linux-Only Tap (Recommended)
If your entire tap is Linux-specific (THIS TAP), **omit `depends_on` entirely**:

```ruby
cask "my-app" do
  version "1.0.0"
  # ... no depends_on needed - this tap is Linux-only
  binary "my-app"
end
```

**IMPORTANT:** Always verify you're downloading the **Linux** version of the binary, not macOS or Windows.

#### Option B: For Platform-Specific Requirements
If you need platform requirements, use the hash syntax:

```ruby
# macOS only (not applicable for Linux taps)
depends_on macos: ">= :big_sur"

# Architecture requirements (valid for Linux)
depends_on arch: :x86_64
depends_on arch: :arm64
depends_on arch: [:x86_64, :arm64]  # Either architecture
```

#### Option C: For Formula Dependencies
If you need Homebrew formulas installed first:

```ruby
# Depends on a Homebrew formula
depends_on formula: "unar"

# Multiple formula dependencies
depends_on formula: ["unar", "jq"]
```

#### Option D: For Cask Dependencies
If you need other casks installed first:

```ruby
# Depends on another cask
depends_on cask: "macfuse"

# Multiple cask dependencies
depends_on cask: ["docker", "virtualbox"]
```

**Reference:** [Homebrew Cask Cookbook - depends_on](https://docs.brew.sh/Cask-Cookbook#stanza-depends_on)

### 2. DO NOT Include `test` Blocks in Casks

**WRONG:**
```ruby
cask "my-app" do
  binary "my-app"
  
  test do
    assert_predicate bin/"my-app", :exist?
    system bin/"my-app", "--version"
  end
end
```

**ERROR:** `wrong number of arguments (given 0, expected 2..3)`

**WHY:** The `test` block with `assert_predicate` and similar methods is for **formulas only**, not casks. Casks use a different testing approach.

**CORRECT:**
```ruby
cask "my-app" do
  binary "my-app"
  # No test block needed - CI will verify installation
end
```

**Note:** Test blocks are optional for casks. If you need verification, Homebrew's audit and installation process will handle it automatically.

### 3. Follow Strict Stanza Ordering and Spacing

**WRONG:**
```ruby
cask "my-app" do
  version "1.0.0"
  sha256 "abc123"
  
  url "https://example.com/app.tar.gz"
                                         # ❌ Extra blank line
  name "My App"
  desc "Description"
  homepage "https://example.com"
  
  binary "my-app"
end
```

**ERROR:** `Cask/StanzaGrouping: stanzas within the same group should have no lines between them`

**CORRECT:**
```ruby
cask "my-app" do
  version "1.0.0"
  sha256 "abc123"

  url "https://example.com/app.tar.gz"
  name "My App"
  desc "Description"
  homepage "https://example.com"

  binary "my-app"
end
```

**Stanza Groups (with blank lines BETWEEN groups only):**

```ruby
# Group 1: Version info (blank line AFTER)
version "..."
sha256 "..."

# Group 2: Metadata (blank line AFTER)
url "..."
name "..."
desc "..."
homepage "..."

# Group 3: Artifacts (no blank line needed before end)
binary "..."
```

**Complete Ordering Reference:**
```
cask "name" do
  version
  sha256

  url
  name
  desc
  homepage

  binary / app / pkg / installer
end
```

## Verified Minimal Cask Template

This template has been tested and passes all CI checks:

```ruby
cask "app-name" do
  version "1.0.0"
  sha256 "ACTUAL_SHA256_HASH_HERE"

  url "https://github.com/user/repo/releases/download/v#{version}/app-linux-x64.tar.xz"
  name "App Display Name"
  desc "One-line description of what the app does"
  homepage "https://github.com/user/repo"

  binary "path/to/binary", target: "command-name"
end
```

**⚠️ CRITICAL:** The URL MUST point to a Linux binary. Common patterns:
- `app-linux-x64.tar.gz` ✓
- `app-x86_64-unknown-linux-gnu.tar.gz` ✓
- `app_amd64.deb` ✓
- `app-macos.dmg` ✗ WRONG
- `app-darwin-x64.zip` ✗ WRONG

### Binary Path Rules

The `binary` stanza path is **relative to the extracted archive root**:

```ruby
# If archive structure is:
# app-1.0.0/
#   bin/
#     myapp
binary "bin/myapp", target: "myapp"

# If archive structure is:
# sublime_text/
#   sublime_text
binary "sublime_text/sublime_text", target: "subl"

# If binary is at root:
# myapp
binary "myapp"
```

## CI Workflow Requirements (GitHub Actions)

### Critical: Tap the Repository Before Auditing

**WRONG:**
```yaml
- name: Run brew audit
  run: |
    brew audit --cask --strict --online Casks/my-cask.rb  # ❌ Path not allowed
```

**ERROR:** `Calling 'brew audit [path ...]' is disabled! Use 'brew audit [name ...]' instead.`

**CORRECT:**
```yaml
- name: Set up Homebrew
  uses: Homebrew/actions/setup-homebrew@master

- name: Tap repository
  run: |
    mkdir -p $(brew --repository)/Library/Taps/USER
    ln -s $GITHUB_WORKSPACE $(brew --repository)/Library/Taps/USER/homebrew-tap

- name: Run brew audit
  run: |
    brew audit --cask --strict --online "USER/tap/cask-name"
```

**Why:** Homebrew changed their API to disallow file paths. You must:
1. Tap the repository first (symlink it into Homebrew's tap directory)
2. Reference casks by `tap-name/cask-name`, not by file path

## Common Errors and Solutions

### Error: "Cask '...' is unavailable: No Cask with this name exists"

**Cause:** Trying to audit a cask before tapping the repository.

**Solution:** Add the "Tap repository" step to your CI workflow (see above).

### Error: "wrong number of arguments (given 1, expected 0)"

**Cause:** Using `depends_on :linux` (invalid syntax for casks).

**Solution:** Either remove `depends_on` entirely, or use `depends_on arch: :x86_64` for architecture requirements.

### Error: "Cask/StanzaGrouping: stanzas within the same group should have no lines between them"

**Cause:** Extra blank lines within stanza groups.

**Solution:** Remove blank lines between `url`, `name`, `desc`, `homepage`. Keep blank lines only between major groups (version info, metadata, artifacts).

### Error: "wrong number of arguments (given 0, expected 2..3)"

**Cause:** Using `assert_predicate` or other formula test methods in a cask.

**Solution:** Remove the entire `test` block from the cask.

## Verification Checklist

Before submitting a cask, verify:

- [ ] No `depends_on :linux` (omit entirely or use valid syntax)
- [ ] No `test do ... end` block
- [ ] Proper stanza ordering (version/sha256, then url/name/desc/homepage, then artifacts)
- [ ] No blank lines within stanza groups
- [ ] One blank line between stanza groups
- [ ] Binary paths are relative to extracted archive root
- [ ] CI workflow taps the repository before auditing
- [ ] CI workflow uses `brew audit --cask --strict --online "tap-name/cask-name"` (not file paths)

## Example: Complete Working Cask

This is a real, tested cask that passes all CI checks:

```ruby
cask "sublime-text" do
  version "4200"
  sha256 "36f69c551ad18ee46002be4d9c523fe545d93b67fea67beea731e724044b469f"

  url "https://download.sublimetext.com/sublime_text_build_#{version}_x64.tar.xz"
  name "Sublime Text"
  desc "Sophisticated text editor for code, markup and prose"
  homepage "https://www.sublimetext.com/"

  binary "sublime_text/sublime_text", target: "subl"
end
```

**Key Points:**
- No `depends_on` (Linux-only tap)
- No `test` block
- Proper spacing: blank line after sha256, blank line after homepage
- Binary path matches archive structure

## Resources

- [Official Homebrew Cask Cookbook](https://docs.brew.sh/Cask-Cookbook) - Always check this for latest updates
- [Cask Stanza Reference](https://docs.brew.sh/Cask-Cookbook#stanzas)
- [Example Casks in Official Tap](https://github.com/Homebrew/homebrew-cask/tree/HEAD/Casks)

---

**Last Verified:** 2026-02-09  
**Homebrew API Version:** Current (2026)  
**Status:** All patterns tested and passing CI
