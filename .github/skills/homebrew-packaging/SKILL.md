---
name: homebrew-packaging
description: >
  Complete workflow for creating and updating Homebrew packages (casks and 
  formulas) for Linux-only tap targeting read-only filesystem systems. Use 
  this when user requests adding, updating, or fixing packages from GitHub 
  releases.
license: MIT
---

# Homebrew Package Creation Workflow

**CRITICAL CONSTRAINTS:**
- ✓ Linux binaries ONLY (no macOS/Windows)
- ✓ Install to `~/.local/` (read-only root filesystem)
- ✓ Format priority: tarball > deb > other
- ✓ All casks MUST use `-linux` suffix

## Workflow Steps

### Step 1: Generate Package Using tap-tools

**ALWAYS use tap-tools for package generation** (fastest, most reliable):

```bash
# For GUI applications (casks)
./tap-tools/tap-cask generate <name> <github-url>

# For CLI tools (formulas)
./tap-tools/tap-formula generate <name> <github-url>
```

**Example:**
```bash
./tap-tools/tap-cask generate sublime-text https://github.com/sublimehq/sublime_text
# Creates: Casks/sublime-text-linux.rb
```

**What tap-tools does automatically:**
- ✓ Fetches latest release from GitHub
- ✓ Filters Linux-only assets (rejects macOS/Windows)
- ✓ Prioritizes tarball format
- ✓ Downloads and verifies SHA256
- ✓ Detects desktop integration (icons, .desktop files)
- ✓ Generates XDG-compliant installation paths

### Step 2: Validate Package (MANDATORY)

**NEVER commit without validation:**

```bash
./tap-tools/tap-validate file <path-to-rb-file> --fix
```

**Example:**
```bash
./tap-tools/tap-validate file Casks/sublime-text-linux.rb --fix
```

**What validation checks:**
- ✓ Style compliance (line length, formatting, ordering)
- ✓ XDG environment variable usage (not hardcoded paths)
- ✓ Required fields present (version, sha256, url)

**CHECKPOINT:** Validation MUST pass before proceeding. If errors remain after `--fix`, investigate and resolve manually.

### Step 3: Review Generated Package

**Verify these critical elements:**

1. **Linux asset selected:**
   ```ruby
   url "https://example.com/app-linux-x64.tar.gz"  # ✓ Good
   # NOT: app-darwin.dmg or app-windows.exe
   ```

2. **XDG environment variables used:**
   ```ruby
   ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")  # ✓ Good
   # NOT: "#{Dir.home}/.local/share" (hardcoded)
   ```

3. **Cask has `-linux` suffix:**
   ```ruby
   cask "sublime-text-linux" do  # ✓ Good
   # NOT: "sublime-text" (collides with macOS cask)
   ```

4. **SHA256 present:**
   ```ruby
   sha256 "abc123..."  # ✓ Required
   # NOT: sha256 :no_check (only with justification)
   ```

### Step 4: Test Installation (Recommended)

**If possible, test locally:**

```bash
# For casks
brew install --cask <name>-linux

# For formulas
brew install <name>

# Verify binary works
<binary-name> --version
```

**For GUI apps:**
- Check desktop file exists: `~/.local/share/applications/`
- Check icon exists: `~/.local/share/icons/`
- Try launching from application menu

### Step 5: Commit With Validation Passing

**Use conventional commit format:**

```bash
git add Casks/<name>-linux.rb  # or Formula/<name>.rb
git commit -m "feat(cask): add <name>-linux

<Brief description of the package>

Assisted-by: [Model Name] via [Tool Name]"
```

**Examples:**
```bash
# New cask
git commit -m "feat(cask): add sublime-text-linux

Adds Sublime Text editor with desktop integration.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# New formula
git commit -m "feat(formula): add jq

Adds jq command-line JSON processor.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# Fix existing package
git commit -m "fix(cask): correct rancher-desktop XDG paths

Replaces hardcoded Dir.home with XDG environment variables.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"
```

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

## Canonical Examples

**For GUI applications, reference:**
- `Casks/sublime-text-linux.rb` - Desktop integration, XDG paths

**For CLI tools, reference:**
- `Formula/jq.rb` - Simple formula pattern

## Common Issues and Solutions

### Issue: CI fails with "line too long"
**Solution:** Run `./tap-tools/tap-validate file <path> --fix` - auto-fixes line length

### Issue: CI fails with "array not alphabetically ordered"
**Solution:** Run `./tap-tools/tap-validate file <path> --fix` - auto-fixes ordering

### Issue: "Dir.home is hardcoded"
**Solution:** Replace with:
```ruby
ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")
ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")
```

### Issue: "No Linux asset found"
**Solution:** Check GitHub releases - package may not provide Linux binaries

### Issue: "Selected .zip instead of .tar.gz"
**Solution:** Manually specify preferred asset or regenerate with tap-tools (it prioritizes tarballs)

## Documentation References

**Before creating packages, read:**
- `docs/CASK_CREATION_GUIDE.md` - Critical cask rules
- `docs/FORMULA_PATTERNS.md` - Formula templates
- `tap-tools/README.md` - Tool usage and features

**For troubleshooting:**
- `docs/TROUBLESHOOTING.md` - Common errors and solutions

## Phase 2 Enhancements (Coming Soon)

The following features are planned:

- **Auto-validation in tap-tools:** Tools will validate after generation automatically
- **validate-and-commit.sh script:** One-command validation + commit workflow
- **Enhanced CI smoke testing:** Real installation tests to verify packages work

Check `docs/plans/2026-02-09-zero-error-packages-design.md` for implementation status.

---

**Remember:** The goal is zero CI failures. Always validate before committing!
