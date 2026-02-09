# Copilot Instructions for homebrew-tap

## ⚠️ MANDATORY: LOAD THE PACKAGING SKILL FIRST ⚠️

**BEFORE doing ANY package work, you MUST load the homebrew-packaging skill:**

1. **Read the skill:** `.github/skills/homebrew-packaging/SKILL.md`
2. **Follow its 6-step workflow exactly**
3. **Complete ALL mandatory checkpoints** (especially validation)

**The skill contains:**
- Critical constraints (Linux-only, read-only filesystem, XDG paths)
- Mandatory 6-step workflow (generate → validate → review → test → commit → PR)
- All packaging requirements and rules

**Also read:** `AGENTS.md` in the repository root for additional agent-specific guidance.

**The skill and AGENTS.md are your authoritative sources. This file is a quick reference only.**

---

## 🚨 MANDATORY WORKFLOW - FOLLOW EVERY STEP

**THIS IS NOT OPTIONAL. CI WILL FAIL IF YOU SKIP ANY STEP.**

The generators (`tap-cask`, `tap-formula`) automatically validate the generated code. You DO NOT need to run validation manually unless you edit a file after generation.

---

## STEP 1: Environment Setup (RUN ONCE PER SESSION)

**Before doing ANYTHING else, verify your environment:**

```bash
# Check if pre-commit hook is installed
[ -f .git/hooks/pre-commit ] || ./scripts/setup-hooks.sh
```

**Expected output:**
```
✓ Pre-commit hook installed
```

**What this does:**
- Installs a pre-commit hook that validates Ruby files before commit
- The hook automatically runs `tap-validate --fix` on staged Ruby files
- Blocks commits that fail validation
- You only need to do this once per session

---

## STEP 2: Creating a New Package

### Option A: From a GitHub URL (Most Common)

**FOR CASKS (GUI Applications):**

```bash
# Generate the cask (validation happens automatically)
./tap-tools/tap-cask generate https://github.com/user/repo

# ✅ WAIT FOR: "✓ Validation passed" or "✓ Validation passed (style issues auto-fixed)"
# ❌ IF YOU SEE: "✗ Validation failed" - STOP! Read the error and fix it first.

# Stage the generated file
git add Casks/<name>-linux.rb

# Commit (pre-commit hook will validate again automatically)
git commit -m "feat(cask): add <name>-linux v<version>

<one-line description>

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# Push to remote
git push
```

**FOR FORMULAS (CLI Tools):**

```bash
# Generate the formula (validation happens automatically)
./tap-tools/tap-formula generate https://github.com/user/repo

# ✅ WAIT FOR: "✓ Validation passed"
# ❌ IF YOU SEE: "✗ Validation failed" - STOP! Read the error and fix it first.

# Stage the generated file
git add Formula/<name>.rb

# Commit (pre-commit hook will validate again automatically)
git commit -m "feat(formula): add <name> v<version>

<one-line description>

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# Push to remote
git push
```

### Option B: From a GitHub Issue

```bash
# Process the issue (automatically calls tap-cask or tap-formula)
./tap-tools/tap-issue process <issue-number>

# ✅ WAIT FOR: "✓ Validation passed"
# ❌ IF YOU SEE: "✗ Validation failed" - STOP! Read the error and fix it first.

# The tool will tell you what file was created. Stage it:
git add Casks/<name>-linux.rb  # or Formula/<name>.rb

# Commit
git commit -m "feat(cask): add <name>-linux v<version>

Fixes #<issue-number>

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# Push
git push
```

---

## STEP 3: Verification Checklist

**After running the generator, verify these outputs:**

- [ ] You saw: `✓ Validation passed` or `✓ Validation passed (style issues auto-fixed)`
- [ ] The generator did NOT show: `✗ Validation failed`
- [ ] The file was created in `Casks/` or `Formula/`
- [ ] The filename ends with `-linux.rb` (for casks only)

**If ANY checkbox is unchecked, DO NOT COMMIT. Fix the issues first.**

---

## STEP 4: What If Something Goes Wrong?

### Scenario 1: Generator says "Validation failed"

**The generator already tried to auto-fix. This is rare and indicates a serious issue.**

```bash
# Look at the error message. Common issues:
# - File not found (check the URL)
# - No Linux assets found (package doesn't support Linux)
# - Checksum mismatch (try again, might be transient)

# If the file was created despite the error, try manual validation:
./tap-tools/tap-validate file Casks/<name>-linux.rb --fix

# Check the output:
# ✅ "✓ Style check passed" - You can now commit
# ❌ "✗ Validation failed" - Read the error and fix manually
```

### Scenario 2: Pre-commit hook blocks your commit

**This means the file has validation issues. The hook already tried to auto-fix it.**

```bash
# The hook modified the file. You need to re-stage it:
git add Casks/<name>-linux.rb

# Now try committing again:
git commit -m "..."

# The hook will run again and should pass this time
```

### Scenario 3: You want to edit a generated file

**Generally, don't do this. Re-generate instead. But if you must:**

```bash
# Edit the file
vim Casks/<name>-linux.rb

# Run validation manually:
./tap-tools/tap-validate file Casks/<name>-linux.rb --fix

# ✅ Wait for: "✓ Style check passed"
# ❌ If failed: Read the error, fix, and validate again

# Stage and commit:
git add Casks/<name>-linux.rb
git commit -m "..."
```

### Scenario 4: CI fails after pushing

**This should NEVER happen if you followed the workflow. But if it does:**

1. Look at the GitHub Actions logs
2. Identify the error (usually style issues)
3. Fix locally:
   ```bash
   ./tap-tools/tap-validate file <file> --fix
   git add <file>
   git commit -m "style: fix validation issues"
   git push
   ```

---

## STEP 5: Understanding This Repository

### Key Constraints (MEMORIZE THESE)

**This is a Linux-only tap:**
- ✅ Use Linux binaries ONLY (`linux`, `amd64`, `x86_64`, `.tar.gz`, `.deb`)
- ❌ NEVER use macOS binaries (`.dmg`, `.pkg`, `darwin`, `macos`)
- ❌ NEVER use Windows binaries (`.exe`, `.msi`, `windows`, `win64`)

**This tap targets read-only filesystems:**
- ✅ Install to `~/.local/share/`, `~/.local/bin/`, `~/.config/`, `~/.cache/`
- ❌ NEVER install to `/usr/`, `/opt/`, `/etc/` (filesystem is read-only!)

**Naming requirements:**
- ✅ Casks MUST end with `-linux` (e.g., `sublime-text-linux`)
- ✅ Formulas do NOT need `-linux` suffix

### Critical Rules (BREAKING THESE CAUSES CI FAILURES)

**NEVER do these things:**
- ❌ `depends_on :linux` (causes "wrong number of arguments" error)
- ❌ `test do ... end` blocks in casks (causes "wrong number of arguments" error)
- ❌ Hardcoded paths like `"#{Dir.home}/.local/share"` (use XDG variables instead)
- ❌ `--no-verify` flag when committing (bypasses validation)
- ❌ Manual cask writing (always use generators)

**ALWAYS do these things:**
- ✅ Use generators (`tap-cask`, `tap-formula`) to create packages
- ✅ Wait for "✓ Validation passed" before committing
- ✅ Use XDG environment variables: `ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")`
- ✅ Include SHA256 verification (generators do this automatically)

---

## STEP 6: Tools Reference

### tap-cask - Generate casks for GUI applications

```bash
./tap-tools/tap-cask generate https://github.com/user/repo
./tap-tools/tap-cask generate user/repo
./tap-tools/tap-cask generate https://github.com/user/repo --name custom-name
```

**What it does automatically:**
- ✅ Fetches latest release from GitHub
- ✅ Selects Linux-only assets (rejects macOS/Windows)
- ✅ Prioritizes formats: tarball > deb > other
- ✅ Downloads and calculates SHA256
- ✅ Detects binaries, desktop files, and icons
- ✅ Generates cask with proper XDG paths
- ✅ **Validates with --fix automatically**
- ✅ Ensures `-linux` suffix in cask name

**Output file:** `Casks/<name>-linux.rb`

### tap-formula - Generate formulas for CLI tools

```bash
./tap-tools/tap-formula generate https://github.com/user/repo
./tap-tools/tap-formula generate user/repo
```

**What it does automatically:**
- ✅ Fetches latest release from GitHub
- ✅ Selects Linux-only assets
- ✅ Detects build system (Go, Rust, CMake, Meson, Make)
- ✅ Generates proper build instructions
- ✅ **Validates with --fix automatically**

**Output file:** `Formula/<name>.rb`

### tap-issue - Process package requests from issues

```bash
./tap-tools/tap-issue process <issue-number>
./tap-tools/tap-issue process <issue-number> --create-pr
```

**What it does automatically:**
- ✅ Reads the GitHub issue
- ✅ Determines if it's a cask or formula request
- ✅ Calls `tap-cask` or `tap-formula`
- ✅ **Validation is handled by the generators**
- ✅ Optionally creates a PR with `--create-pr`

### tap-validate - Manual validation (rarely needed)

```bash
# Validate a single file
./tap-tools/tap-validate file Casks/<name>-linux.rb --fix

# Validate all files in repository
./tap-tools/tap-validate all --fix
```

**When to use:**
- When generator validation fails (to debug)
- After manually editing a file (not recommended)
- To verify CI will pass before pushing

**What it does:**
- ✅ Runs `brew style` with RuboCop
- ✅ Auto-fixes most issues with `--fix`
- ✅ Returns clear success/failure messages

---

## STEP 7: Commit Format (MANDATORY)

Use **Conventional Commits** with AI attribution:

```
<type>(<scope>): <description>

<optional body>

Assisted-by: <Model> via <Tool>
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `ci`

**Examples:**

```
feat(cask): add rancher-desktop-linux v1.22.0

Adds Rancher Desktop with desktop integration and XDG paths.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot
```

```
feat(formula): add jq v1.7.1

JSON command-line processor.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot
```

```
fix(cask): correct sublime-text-linux binary path

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot
```

---

## STEP 8: Pull Requests (For Major Work)

**All major features MUST use pull requests, NOT direct commits to main.**

```bash
# Create feature branch
git checkout -b feat/add-new-package

# Generate and commit package (follow STEP 2)
./tap-tools/tap-cask generate https://github.com/user/repo
git add Casks/<name>-linux.rb
git commit -m "feat(cask): add <name>-linux v<version>

<description>

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# Push and create PR
git push -u origin feat/add-new-package

gh pr create --title "feat(cask): add <name>-linux" --body "$(cat <<'EOF'
## Summary
- Adds <name> v<version> as Linux cask
- Includes desktop integration with XDG paths
- Passes validation and style checks

## Testing
- Generated with tap-cask
- Validated with tap-validate
- Pre-commit hook passed

Fixes #<issue-number>
EOF
)"
```

---

## STEP 9: Important Reminders

### Things You Should NEVER Do

1. ❌ **Skip validation** - Generators validate automatically, don't skip the step
2. ❌ **Edit generated files** - Regenerate instead
3. ❌ **Use `--no-verify`** - This bypasses pre-commit hooks
4. ❌ **Write casks manually** - Always use `tap-cask generate`
5. ❌ **Commit without seeing "✓ Validation passed"**
6. ❌ **Use hardcoded paths** - Generators use XDG variables automatically
7. ❌ **Add `depends_on :linux`** - This is a Linux-only tap, it's implicit

### Things That Guarantee Success

1. ✅ **Use generators** - They handle validation automatically
2. ✅ **Wait for "✓ Validation passed"** - Don't commit before seeing this
3. ✅ **Let pre-commit hook run** - Don't bypass it
4. ✅ **Follow the exact command sequence** - Copy-paste from STEP 2
5. ✅ **Trust the tools** - They're designed to prevent CI failures
6. ✅ **Read error messages** - They tell you exactly what's wrong

---

## STEP 10: Why This Workflow Exists

### The Problem It Solves

Before this workflow:
- ❌ Agents wrote casks manually → style issues → CI failures
- ❌ Validation was optional → agents skipped it → CI failures  
- ❌ No pre-commit hooks → issues caught too late → wasted time

After this workflow:
- ✅ Generators create valid code automatically
- ✅ Validation is built into generators → impossible to skip
- ✅ Pre-commit hooks catch issues before CI → zero CI failures
- ✅ Fast feedback loop (2-3 seconds vs 60 seconds)

### The Design Philosophy

**"Make the right thing easy, and the wrong thing hard."**

- ✅ Right thing (using generators): Easy, fast, always works
- ❌ Wrong thing (manual writing): Hard, slow, requires knowledge

**"Shift validation left."**

- Catch issues at creation time (generators), not at commit time (CI)
- Every layer validates (generator → pre-commit hook → CI)
- Defense-in-depth approach

---

## STEP 11: Getting Help

### If Something Is Unclear

1. Read the error message carefully - they're designed to be actionable
2. Check `docs/TROUBLESHOOTING.md` for common issues
3. Check `docs/CASK_CREATION_GUIDE.md` for cask-specific rules
4. Check `docs/FORMULA_PATTERNS.md` for formula examples

### If You Encounter a Bug

1. Save the command you ran
2. Save the complete error output
3. Save the generated file (if any)
4. File an issue with all three

---

## Quick Reference: Complete Workflow

```bash
# 1. Setup (once per session)
[ -f .git/hooks/pre-commit ] || ./scripts/setup-hooks.sh

# 2. Generate cask
./tap-tools/tap-cask generate https://github.com/user/repo
# ✅ Wait for: "✓ Validation passed"

# 3. Stage
git add Casks/<name>-linux.rb

# 4. Commit
git commit -m "feat(cask): add <name>-linux v<version>

<description>

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot"

# 5. Push
git push
```

**That's it. If you follow these 5 steps, CI will pass every time.**

---

## Document Version

**Version:** 2.0  
**Last Updated:** February 9, 2026  
**Changes:** Complete rewrite with mandatory checklist workflow and automatic validation
