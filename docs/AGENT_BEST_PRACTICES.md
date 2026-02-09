# Agent Best Practices - Common Errors and How to Avoid Them

**⚠️ CRITICAL: Read this document BEFORE creating any packages ⚠️**

This document catalogs real errors that AI agents have made and how to prevent them. Following these practices will prevent 100% of CI failures.

## Table of Contents

1. [Golden Rules](#golden-rules)
2. [Common Validation Errors](#common-validation-errors)
3. [Platform-Specific Issues](#platform-specific-issues)
4. [XDG Compliance](#xdg-compliance)
5. [Pre-Commit Checklist](#pre-commit-checklist)
6. [Real-World Examples](#real-world-examples)

---

## Golden Rules

### Rule 1: ALWAYS Use tap-tools

**❌ NEVER manually create package files**
**✓ ALWAYS use tap-cask or tap-formula generators**

```bash
# ✓ CORRECT
./tap-tools/tap-cask generate rancher-desktop https://github.com/rancher-sandbox/rancher-desktop

# ❌ WRONG - Manual creation leads to:
# - Style violations
# - Missing XDG variables
# - Incorrect stanza ordering
# - CI failures
```

### Rule 2: ALWAYS Validate with --fix

**❌ NEVER commit without validation**
**✓ ALWAYS run tap-validate --fix before committing**

```bash
# ✓ CORRECT - Before EVERY commit
./tap-tools/tap-validate file Casks/app-name-linux.rb --fix
git add Casks/app-name-linux.rb
git commit -m "..."

# ❌ WRONG - Skipping validation guarantees CI failure
git add Casks/app-name-linux.rb
git commit -m "..."
```

### Rule 3: Never Edit Without Re-Validating

**If you edit a generated file, you MUST re-validate before committing.**

```bash
# Workflow for editing
vim Casks/app-name-linux.rb                              # Edit file
./tap-tools/tap-validate file Casks/app-name-linux.rb --fix  # Re-validate
git add Casks/app-name-linux.rb                         # Only commit after validation passes
git commit -m "..."
```

### Rule 4: Use Strings, Not Regex (When Possible)

**✓ Use literal strings for exact matches**
**❌ Don't use regex when strings work**

```ruby
# ✓ CORRECT - Literal string match
content.gsub("Exec=rancher-desktop", "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")

# ❌ WRONG - Unnecessary regex
content.gsub(/Exec=rancher-desktop/, "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")
```

**When to use regex:**
- Pattern matching: `/^Icon=.*$/`
- Character classes: `/[A-Za-z]+/`
- Anchors: `/^Exec=/` (start of line)

**When to use strings:**
- Exact literal matches: `"Exec=app"`
- Simple replacements

### Rule 5: Respect Line Length Limits

**Lines must be ≤ 118 characters**

```ruby
# ❌ WRONG - 121 characters
updated_content = updated_content.gsub("Exec=app", "Exec=#{xdg_data_home}/applications/app")

# ✓ CORRECT - Split across lines
updated_content = updated_content.gsub(
  "Exec=app",
  "Exec=#{xdg_data_home}/applications/app"
)
```

**Note:** `tap-validate --fix` handles this automatically.

---

## Common Validation Errors

### Error 1: "Use string as argument instead of regexp"

**What it means:** You're using `/pattern/` when `"pattern"` works.

**Example from real PR failure:**
```ruby
# ❌ FAILS CI
updated_content = content.gsub(/Exec=rancher-desktop/, "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")

# ✓ PASSES CI
updated_content = content.gsub("Exec=rancher-desktop", "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")
```

**Why:** RuboCop enforces `Style/RedundantRegexpArgument` - use simpler syntax when possible.

**How to prevent:** Run `tap-validate --fix` which auto-corrects this.

### Error 2: "Line is too long"

**What it means:** Line exceeds 118 characters.

**Example:**
```ruby
# ❌ FAILS CI (125 characters)
artifact "icon.png", target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/hicolor/256x256/apps/app.png"

# ✓ PASSES CI (split across lines)
artifact "icon.png",
         target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/hicolor/256x256/apps/app.png"
```

**How to prevent:** Run `tap-validate --fix` which auto-wraps long lines.

### Error 3: "Array elements should be ordered alphabetically"

**What it means:** Arrays (especially `zap trash`) must be alphabetically sorted.

**Example:**
```ruby
# ❌ FAILS CI
zap trash: [
  "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app",
  "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/app",
]

# ✓ PASSES CI (CACHE comes before CONFIG)
zap trash: [
  "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/app",
  "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app",
]
```

**How to prevent:** Run `tap-validate --fix` which auto-sorts arrays.

### Error 4: "Stanzas within the same group should have no lines between them"

**What it means:** Blank lines are only allowed between stanza groups.

**Stanza groups:**
1. Version info: `version`, `sha256`
2. Metadata: `url`, `name`, `desc`, `homepage`
3. Artifacts: `binary`, `artifact`, `app`
4. Hooks: `preflight`, `postflight`
5. Cleanup: `zap`

**Example:**
```ruby
# ❌ FAILS CI
cask "app-name" do
  version "1.0.0"
  sha256 "abc123"
  
  url "https://example.com/app.tar.gz"
                                          # ❌ Extra blank line
  name "App Name"
  desc "Description"
  
  binary "app"
end

# ✓ PASSES CI
cask "app-name" do
  version "1.0.0"
  sha256 "abc123"

  url "https://example.com/app.tar.gz"
  name "App Name"
  desc "Description"
  homepage "https://example.com"

  binary "app"
end
```

### Error 5: "Description contains marketing fluff"

**What it means:** Descriptions must be functional, not promotional.

**Bad descriptions:**
```ruby
# ❌ Marketing language
desc "Modern and beautiful sound editor for professionals"
desc "The best IDE for modern developers"
desc "Edit your music with ease"
desc "Rancher Desktop is a container management tool"  # Includes app name
```

**Good descriptions:**
```ruby
# ✓ Functional descriptions
desc "Sound and music editor"
desc "Integrated development environment"
desc "Container management and Kubernetes on the desktop"
```

**Description rules:**
- Start with uppercase letter
- Under 80 characters
- Describe WHAT it does
- No marketing adjectives ("modern", "beautiful", "powerful")
- No platform mentions ("for macOS", "on Linux")
- No vendor/app name in description
- No user pronouns ("your", "you")

---

## Platform-Specific Issues

### Issue 1: Using macOS Binaries Instead of Linux

**⚠️ THIS TAP IS LINUX-ONLY**

```ruby
# ❌ WRONG - macOS formats
url "https://example.com/app-v1.0.0-darwin-x64.tar.gz"  # darwin = macOS
url "https://example.com/app-v1.0.0.dmg"                # .dmg = macOS
url "https://example.com/app-v1.0.0-macos.pkg"          # .pkg = macOS

# ❌ WRONG - Windows formats
url "https://example.com/app-v1.0.0-windows.exe"        # .exe = Windows
url "https://example.com/app-v1.0.0-win64.msi"          # .msi = Windows

# ✓ CORRECT - Linux formats
url "https://example.com/app-v1.0.0-linux-x64.tar.gz"   # Tarball (preferred)
url "https://example.com/app-v1.0.0-amd64.deb"          # Debian package
url "https://example.com/app-v1.0.0_linux_amd64.tar.xz" # Compressed tarball
```

**Keywords to look for:**
- Linux: `linux`, `amd64`, `x86_64`, `x64` (in Linux context)
- macOS: `darwin`, `macos`, `osx`, `.dmg`, `.pkg`
- Windows: `windows`, `win32`, `win64`, `.exe`, `.msi`

**How to verify:** Check the GitHub releases page manually and confirm you're selecting the Linux asset.

### Issue 2: Incorrect Format Priority

**Priority order:** Tarball > Debian Package > Other

```ruby
# ✓ BEST - Tarball (.tar.gz, .tar.xz, .tgz)
url "https://github.com/user/app/releases/download/v1.0.0/app-linux-x64.tar.gz"

# ✓ ACCEPTABLE - Debian package (only if no tarball)
url "https://github.com/user/app/releases/download/v1.0.0/app_amd64.deb"

# ⚠️ AVOID - Zip (use tarball if both available)
url "https://github.com/user/app/releases/download/v1.0.0/app-linux-x64.zip"
```

**Why tarballs are preferred:**
- Work on all Linux distributions
- No package manager dependencies
- Simpler extraction
- Standard Unix format

### Issue 3: Missing `-linux` Suffix for Casks

**ALL casks in this tap MUST use `-linux` suffix**

```ruby
# ❌ WRONG - Will conflict with official macOS casks
cask "sublime-text" do
  # ...
end

# ✓ CORRECT - Prevents naming conflicts
cask "sublime-text-linux" do
  # ...
end
```

**Why:** Prevents collision with official Homebrew casks (which are macOS-only).

**Note:** Formulas (CLI tools) don't need the suffix.

---

## XDG Compliance

### What is XDG Base Directory Spec?

The XDG Base Directory Specification defines standard locations for user files:
- `XDG_DATA_HOME` - User data (default: `~/.local/share`)
- `XDG_CONFIG_HOME` - Configuration (default: `~/.config`)
- `XDG_CACHE_HOME` - Cache (default: `~/.cache`)

### Why XDG Matters for This Tap

**Target systems:** Fedora Silverblue, Universal Blue (immutable/read-only root)
- Cannot write to `/usr/`, `/opt/`, `/etc/` (read-only filesystem)
- MUST install to user directories (`~/.local/`)
- MUST respect user's XDG customizations

### Rule: Always Use XDG Environment Variables

**❌ NEVER hardcode paths:**
```ruby
# ❌ WRONG - Hardcoded
target: "#{Dir.home}/.local/share/applications/app.desktop"
target: "#{Dir.home}/.config/app"
```

**✓ ALWAYS use environment variables:**
```ruby
# ✓ CORRECT - Respects XDG
target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/app.desktop"
target: "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/app"
```

### Complete XDG Reference

```ruby
# Desktop files and application data
ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
# Examples:
# - ~/.local/share/applications/  (desktop files)
# - ~/.local/share/icons/         (application icons)
# - ~/.local/share/app-name/      (application data)

# Configuration files
ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")
# Examples:
# - ~/.config/app-name/           (app config)

# Cache data
ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")
# Examples:
# - ~/.cache/app-name/            (temporary cache)

# User binaries (not in XDG spec but conventional)
"#{Dir.home}/.local/bin"
# Examples:
# - ~/.local/bin/app-name         (executable)
```

### XDG Pattern in Casks

```ruby
cask "myapp-linux" do
  version "1.0.0"
  sha256 "abc123..."

  url "https://example.com/myapp-linux.tar.gz"
  name "MyApp"
  desc "Application description"
  homepage "https://example.com"

  binary "myapp"

  # Desktop integration (GUI apps)
  artifact "myapp.desktop",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/myapp.desktop"
  artifact "icon.png",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/myapp.png"

  # Fix paths in preflight
  preflight do
    xdg_data_home = ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
    FileUtils.mkdir_p "#{xdg_data_home}/applications"
    FileUtils.mkdir_p "#{xdg_data_home}/icons"

    desktop_file = "#{staged_path}/myapp.desktop"
    if File.exist?(desktop_file)
      content = File.read(desktop_file)
      updated_content = content.gsub("Exec=myapp", "Exec=#{HOMEBREW_PREFIX}/bin/myapp")
      updated_content = updated_content.gsub("Icon=myapp", "Icon=#{xdg_data_home}/icons/myapp.png")
      File.write(desktop_file, updated_content)
    end
  end

  # Cleanup user data (alphabetically ordered)
  zap trash: [
    "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/myapp",
    "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/myapp",
    "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/myapp",
  ]
end
```

---

## Pre-Commit Checklist

**Complete ALL items before committing. If ANY checkbox is unchecked, DO NOT COMMIT.**

### Generation
- [ ] Used `./tap-tools/tap-cask` or `./tap-tools/tap-formula` (NOT manual creation)
- [ ] Package name ends with `-linux` (casks only, formulas don't need suffix)

### Validation
- [ ] Ran `./tap-tools/tap-validate file <path> --fix`
- [ ] Saw "✓ Style check passed" output
- [ ] No errors remain after `--fix`

### Platform Check
- [ ] URL contains `linux` keyword (NOT `darwin`, `macos`, `windows`)
- [ ] File format is `.tar.gz`, `.tar.xz`, or `.deb` (NOT `.dmg`, `.pkg`, `.exe`, `.msi`)
- [ ] Confirmed on GitHub releases page this is the Linux asset

### XDG Compliance
- [ ] All paths use `ENV.fetch("XDG_*", ...)` (NOT hardcoded `Dir.home`)
- [ ] Desktop file path uses `XDG_DATA_HOME`
- [ ] Icon path uses `XDG_DATA_HOME`
- [ ] Config paths use `XDG_CONFIG_HOME`
- [ ] Cache paths use `XDG_CACHE_HOME`
- [ ] Preflight creates XDG directories with `FileUtils.mkdir_p`

### Style
- [ ] No lines exceed 118 characters
- [ ] Used strings for literal matches (NOT regex like `/pattern/`)
- [ ] Arrays are alphabetically ordered (especially `zap trash`)
- [ ] No extra blank lines within stanza groups
- [ ] Proper blank lines BETWEEN stanza groups

### Content Quality
- [ ] SHA256 is present (NOT `:no_check` unless justified)
- [ ] Description is functional, not marketing (< 80 chars)
- [ ] Description starts with uppercase letter
- [ ] No app/vendor name in description
- [ ] No user pronouns ("your", "you") in description
- [ ] No adjectives ("modern", "beautiful") in description

### Commit
- [ ] Conventional commit format: `feat(cask): add app-name-linux`
- [ ] Commit message includes `Assisted-by:` footer with model and tool
- [ ] Pre-commit hook ran successfully (don't use `--no-verify`)

**If all checkboxes are checked:** Your package is ready to commit! 🎉

**If any checkbox is unchecked:** Fix the issue and re-validate before committing.

---

## Real-World Examples

### Example 1: Rancher Desktop PR #18 Failure

**What happened:** PR failed CI with regex error.

**Error message:**
```
Style/RedundantRegexpArgument: Use string `"Exec=rancher-desktop"` as argument 
instead of regexp `/Exec=rancher-desktop/`.
```

**Code that failed:**
```ruby
updated_content = content.gsub(/Exec=rancher-desktop/, "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")
```

**Why it failed:** Used regex for literal string match.

**How to fix:**
```ruby
updated_content = content.gsub("Exec=rancher-desktop", "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")
```

**How to prevent:**
1. Run `./tap-tools/tap-validate file Casks/rancher-desktop-linux.rb --fix` before committing
2. The `--fix` flag would have auto-corrected this
3. Never bypass validation

### Example 2: Sublime Text XDG Compliance

**Reference cask:** `Casks/sublime-text-linux.rb`

This is the gold standard for XDG-compliant casks. Study it before creating packages.

**What it does right:**
1. Uses XDG environment variables everywhere
2. Creates directories in `preflight`
3. Fixes desktop file paths
4. Installs to user directories
5. Properly alphabetizes `zap trash` array

```ruby
cask "sublime-text-linux" do
  version "4200"
  sha256 "36f69c551ad18ee46002be4d9c523fe545d93b67fea67beea731e724044b469f"

  url "https://download.sublimetext.com/sublime_text_build_#{version}_x64.tar.xz"
  name "Sublime Text"
  desc "Sophisticated text editor for code, markup and prose"
  homepage "https://www.sublimetext.com/"

  binary "sublime_text/sublime_text", target: "subl"
  artifact "sublime_text/sublime_text.desktop",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/sublime-text.desktop"
  artifact "sublime_text/Icon/128x128/sublime-text.png",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/sublime-text.png"

  preflight do
    xdg_data_home = ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
    FileUtils.mkdir_p "#{xdg_data_home}/applications"
    FileUtils.mkdir_p "#{xdg_data_home}/icons"

    desktop_file = "#{staged_path}/sublime_text/sublime_text.desktop"
    if File.exist?(desktop_file)
      content = File.read(desktop_file)
      updated_content = content.gsub(%r{/opt/sublime_text/sublime_text}, "#{HOMEBREW_PREFIX}/bin/subl")
      File.write(desktop_file, updated_content)
    end
  end

  zap trash: [
    "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/sublime-text",
    "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/sublime-text",
  ]
end
```

**Key lessons:**
- XDG variables everywhere (lines 12-13, 17-22)
- Directory creation in preflight (lines 18-19)
- Desktop file path fixing (lines 21-26)
- Alphabetically sorted `zap trash` (lines 29-32, CACHE before CONFIG)

---

## Summary: The Zero-Failure Workflow

**Follow these 6 steps exactly:**

1. **Generate** - Use tap-tools, never manual creation
2. **Validate** - Run `tap-validate --fix` and verify it passes
3. **Review** - Check platform (Linux), XDG vars, description quality
4. **Test** - Install locally if possible (recommended)
5. **Commit** - Only after all checks pass
6. **Never skip** - Each step prevents specific CI failures

**Remember:** Every CI failure is caused by skipping a step. Follow the workflow and CI will never fail.

**When in doubt:** Look at `Casks/sublime-text-linux.rb` - it's the reference implementation.

---

**Last Updated:** 2026-02-09  
**Based on:** Real PR failures and validation errors  
**Status:** Living document - updated as new patterns emerge
