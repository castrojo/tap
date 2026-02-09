# Zero-Error Package Creation Strategy

**Date:** 2026-02-09  
**Status:** Design  
**Problem:** Packages failing CI after creation due to style violations and missing validation steps  
**Goal:** Ensure packages pass CI + install successfully + actually work on first try

## Background

### Current Problem
PR #11 (Rancher Desktop) failed CI with:
- 2 style violations (line too long, array ordering)
- 1 XDG compliance issue (hardcoded `Dir.home`)

**Root Cause:** Validation not run before commit, despite having `.github/copilot-instructions.md`

### Success Criteria
1. **Passes CI** - Style and audit checks pass
2. **Installs successfully** - `brew install` completes without errors
3. **Actually works** - Binary executes, GUI apps launch, desktop integration functions

## Solution: Hybrid Approach (Mandatory Validation + Smoke Testing)

This strategy combines two approaches:
- **Approach 1:** Mandatory validation gates at every entry point
- **Approach 3:** Real installation smoke tests in CI

**Why Hybrid:**
- Mandatory validation solves style failures at the source (prevents bad commits)
- Smoke testing validates "actually works" criteria (catches runtime issues)
- Defense-in-depth: Multiple layers catch different classes of errors

## Architecture

### Three-Phase Implementation

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Local Validation (COMPLETED)                       │
│ ✅ Pre-commit hook blocks invalid commits                   │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Tool-Level Validation (NEXT)                       │
│ • tap-tools auto-validate after generation                  │
│ • validate-and-commit.sh helper script                      │
│ • Enhanced copilot instructions with checklist              │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Smoke Testing (FUTURE)                             │
│ • CI installs package in container                          │
│ • Tests binary execution, desktop integration               │
│ • Validates "actually works" success criteria               │
└─────────────────────────────────────────────────────────────┘
```

### Validation Entry Points

Every path to creating a package must include validation:

```
┌──────────────────┐
│  Package Source  │
└────────┬─────────┘
         │
    ┌────┴────────────────────┐
    │                         │
┌───▼────────┐        ┌──────▼─────────┐
│ tap-tools  │        │ Manual editing │
│ (tap-cask, │        │ by human/agent │
│ tap-formula│        │                │
└───┬────────┘        └──────┬─────────┘
    │                        │
    │ Auto-validate          │ Pre-commit
    │ after generation       │ hook validates
    │                        │
    └────┬───────────────────┘
         │
    ┌────▼────────────────┐
    │ Validated Package   │
    │ (ready to commit)   │
    └─────────────────────┘
```

## Phase 1: Local Validation (COMPLETED)

### Implemented Components

1. **Pre-commit Hook** (`scripts/git-hooks/pre-commit`)
   - Validates all staged `.rb` files automatically
   - Runs `tap-validate --fix` to auto-correct style issues
   - Re-stages fixed files
   - Blocks commit if validation fails

2. **Setup Script** (`scripts/setup-hooks.sh`)
   - Installs pre-commit hook to `.git/hooks/`
   - Makes hook executable
   - Adds to first-time setup instructions

3. **Documentation Updates**
   - `.github/copilot-instructions.md` - Added validation section
   - `README.md` - Added setup instructions

### What It Solves
- ✅ Prevents style violations from being committed locally
- ✅ Auto-fixes most common issues (line length, formatting)
- ✅ Works for both human and AI contributors

### What It Doesn't Solve
- ❌ Doesn't help if hook not installed (new contributors)
- ❌ Can be bypassed with `--no-verify`
- ❌ Doesn't validate "actually works" - only style compliance
- ❌ Doesn't help with tap-tools generated files (fixed in Phase 2)

## Phase 2: Tool-Level Validation (NEXT)

### Components to Implement

#### 2.1 Update tap-tools (tap-cask, tap-formula)

**Goal:** Auto-validate generated packages before returning control to user.

**Changes needed:**
- After writing `.rb` file, run `tap-validate file <path> --fix`
- Report validation results to user
- Exit with error if validation fails (shouldn't happen with --fix)

**Example output:**
```bash
$ ./tap-tools/tap-cask generate rancher-desktop https://github.com/rancher-sandbox/rancher-desktop

✓ Downloaded release metadata from GitHub
✓ Selected asset: rancher-desktop-1.14.0-linux-x64.tar.gz (tarball)
✓ Downloaded and verified SHA256
✓ Generated Casks/rancher-desktop-linux.rb
✓ Running validation...
  ⚠ Fixed 2 style issues (line length, array ordering)
✓ Package ready to commit

Next steps:
  git add Casks/rancher-desktop-linux.rb
  git commit -m "feat(cask): add rancher-desktop-linux"
```

**Benefits:**
- Validation happens automatically at generation time
- User sees fix results immediately
- No way to skip validation (it's built into the tool)

#### 2.2 Create validate-and-commit.sh Helper Script

**Purpose:** Combine validation + commit in one command for AI agents.

**Location:** `scripts/validate-and-commit.sh`

**Usage:**
```bash
./scripts/validate-and-commit.sh Casks/app-linux.rb "feat(cask): add app-linux"
```

**Behavior:**
1. Run `tap-validate file <path> --fix`
2. If validation passes:
   - `git add <path>`
   - `git commit -m "<message>"`
3. If validation fails:
   - Report errors
   - Exit with error code

**Benefits:**
- Single command for agents to use (less chance of skipping validation)
- Consistent workflow across manual and automated contributions
- Clear success/failure feedback

#### 2.3 Enhance Copilot Instructions

**File:** `.github/copilot-instructions.md`

**Changes needed:**
1. Add **mandatory validation checklist** to top of file:
   ```markdown
   ## MANDATORY VALIDATION CHECKLIST (Before Every Commit)
   
   - [ ] Run `./tap-tools/tap-validate file <path> --fix`
   - [ ] Verify all style issues are resolved
   - [ ] Check XDG environment variable usage (not hardcoded Dir.home)
   - [ ] Commit with validation passing
   ```

2. Update workflow section to emphasize validation:
   ```markdown
   ## Workflow
   
   1. Generate package with tap-tools (preferred) or manual creation
   2. **VALIDATE:** `./tap-tools/tap-validate file <path> --fix`
   3. Test installation (if possible)
   4. Commit with conventional format
   5. Create PR
   ```

3. Add examples of validation output (what success looks like)

**Benefits:**
- Checklist format is hard to ignore
- Clear "before every commit" language
- Examples show what to expect

### What Phase 2 Solves
- ✅ Validation at every entry point (tap-tools + manual)
- ✅ Impossible to skip validation when using tap-tools
- ✅ Clear workflow for AI agents (validate-and-commit.sh)
- ✅ Better instructions for Copilot

### What Phase 2 Doesn't Solve
- ❌ Still doesn't validate "actually works"
- ❌ Can't catch runtime issues (bad URLs, missing dependencies)
- ❌ Doesn't test installation in real environment

## Phase 3: Smoke Testing (FUTURE)

### Goal
Validate that packages actually work, not just that they pass style checks.

### Architecture

#### 3.1 CI Job: test-installation

**Workflow file:** `.github/workflows/test-installation.yml`

**Triggers:**
- On PR to changed Formula/Cask files
- On demand via workflow_dispatch

**Matrix strategy:**
- Ubuntu 22.04 (baseline)
- Fedora 40 (rpm-based, test with universal-blue target)

**Job steps:**

```yaml
name: Test Package Installation

on:
  pull_request:
    paths:
      - 'Casks/**'
      - 'Formula/**'

jobs:
  test-install:
    strategy:
      matrix:
        os: [ubuntu-22.04]
        # Future: Add fedora-40
    
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Homebrew
        run: |
          /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
          echo "/home/linuxbrew/.linuxbrew/bin" >> $GITHUB_PATH
      
      - name: Tap this repository
        run: brew tap castrojo/tap "${{ github.workspace }}"
      
      - name: Detect changed packages
        id: changed
        run: |
          # Get changed .rb files
          CHANGED_FILES=$(git diff --name-only origin/main...HEAD | grep -E '\.(rb)$' || true)
          echo "files=$CHANGED_FILES" >> $GITHUB_OUTPUT
      
      - name: Test installation
        run: |
          for file in ${{ steps.changed.outputs.files }}; do
            if [[ "$file" == Casks/* ]]; then
              PACKAGE=$(basename "$file" .rb)
              echo "Testing cask: $PACKAGE"
              brew install --cask "$PACKAGE"
              
              # Smoke tests for casks
              ./scripts/test-cask.sh "$PACKAGE"
            elif [[ "$file" == Formula/* ]]; then
              PACKAGE=$(basename "$file" .rb)
              echo "Testing formula: $PACKAGE"
              brew install "$PACKAGE"
              
              # Smoke tests for formulas
              ./scripts/test-formula.sh "$PACKAGE"
            fi
          done
```

#### 3.2 Smoke Test Scripts

**For Formulas:** `scripts/test-formula.sh`

```bash
#!/bin/bash
# Test that a formula actually works

FORMULA="$1"

echo "Testing formula: $FORMULA"

# Check binary exists
if ! command -v "$FORMULA" &> /dev/null; then
  echo "❌ Binary '$FORMULA' not found in PATH"
  exit 1
fi

# Check binary executes
if ! "$FORMULA" --version &> /dev/null; then
  echo "❌ Binary '$FORMULA' failed to execute"
  exit 1
fi

echo "✓ Formula $FORMULA works"
```

**For Casks:** `scripts/test-cask.sh`

```bash
#!/bin/bash
# Test that a cask actually works

CASK="$1"

echo "Testing cask: $CASK"

# Get installation info
INSTALL_DIR=$(brew --prefix)/Caskroom/"$CASK"

# Check installation directory exists
if [ ! -d "$INSTALL_DIR" ]; then
  echo "❌ Installation directory not found: $INSTALL_DIR"
  exit 1
fi

# Check for desktop file (GUI apps)
DESKTOP_FILE="$HOME/.local/share/applications/${CASK}.desktop"
if [ -f "$DESKTOP_FILE" ]; then
  echo "✓ Desktop file exists: $DESKTOP_FILE"
  
  # Validate desktop file
  if command -v desktop-file-validate &> /dev/null; then
    desktop-file-validate "$DESKTOP_FILE"
  fi
fi

# Check for icon (GUI apps)
ICON_DIR="$HOME/.local/share/icons"
if [ -d "$ICON_DIR" ]; then
  ICON_COUNT=$(find "$ICON_DIR" -name "*${CASK}*" | wc -l)
  if [ "$ICON_COUNT" -gt 0 ]; then
    echo "✓ Found $ICON_COUNT icon(s)"
  fi
fi

# Try to find and execute binary
# This is heuristic - may need per-package customization
BINARY=$(find "$INSTALL_DIR" -type f -executable | head -n1)
if [ -n "$BINARY" ]; then
  echo "✓ Found executable: $BINARY"
  
  # Try --version (many apps support this)
  if "$BINARY" --version &> /dev/null; then
    echo "✓ Binary executes successfully"
  else
    echo "⚠ Binary found but --version failed (may be GUI-only)"
  fi
fi

echo "✓ Cask $CASK smoke test passed"
```

#### 3.3 Test Configuration

**Per-package test overrides:** `test-configs/`

Some packages need custom smoke tests:

```yaml
# test-configs/sublime-text-linux.yml
name: sublime-text-linux
type: cask
smoke_tests:
  - check: desktop_file
    path: ~/.local/share/applications/sublime-text.desktop
  - check: icon
    pattern: sublime-text
  - check: binary_exists
    path: ~/.local/bin/sublime-text
  - check: command_succeeds
    command: sublime-text --version
```

### What Phase 3 Solves
- ✅ Validates "actually works" success criteria
- ✅ Catches runtime issues (bad URLs, broken binaries)
- ✅ Tests desktop integration for GUI apps
- ✅ Provides confidence before merge

### Limitations
- ❌ Slow (container spin-up + brew install takes 2-5 minutes)
- ❌ May have false failures (network issues, upstream changes)
- ❌ GUI apps hard to test (need headless mode or Xvfb)
- ❌ Custom tests needed for some packages

## Implementation Plan

### Phase 2 (Next Steps)

#### Task 1: Update tap-cask
**File:** `tap-tools/cmd/tap-cask/generate.go`
**Change:** After writing cask file, run validation
**Estimated effort:** 1 hour

#### Task 2: Update tap-formula
**File:** `tap-tools/cmd/tap-formula/generate.go`
**Change:** Same as tap-cask
**Estimated effort:** 1 hour

#### Task 3: Create validate-and-commit.sh
**File:** `scripts/validate-and-commit.sh`
**Change:** New script combining validation + commit
**Estimated effort:** 30 minutes

#### Task 4: Enhance copilot-instructions.md
**File:** `.github/copilot-instructions.md`
**Change:** Add mandatory checklist, update workflow section
**Estimated effort:** 30 minutes

#### Task 5: Test Phase 2
**Test cases:**
- Generate new cask with tap-cask (should auto-validate)
- Generate new formula with tap-formula (should auto-validate)
- Use validate-and-commit.sh script (should validate + commit)
- Verify Copilot reads updated instructions

**Total effort for Phase 2:** ~3-4 hours

### Phase 3 (Future)

#### Task 1: Create test-installation.yml workflow
**Estimated effort:** 2 hours

#### Task 2: Create test-formula.sh and test-cask.sh
**Estimated effort:** 2 hours

#### Task 3: Test on real packages
**Packages to test:**
- jq (simple formula)
- sublime-text-linux (GUI cask)
- quarto-cli-linux (complex cask with multiple binaries)

**Estimated effort:** 2 hours

#### Task 4: Add per-package test configs as needed
**Estimated effort:** Ongoing (1-2 hours per complex package)

**Total effort for Phase 3:** ~8-10 hours

## Success Metrics

### Phase 2 Success Criteria
- [ ] tap-cask auto-validates after generation
- [ ] tap-formula auto-validates after generation
- [ ] validate-and-commit.sh works for manual contributions
- [ ] Copilot instructions include mandatory checklist
- [ ] Zero style failures in next 5 PRs

### Phase 3 Success Criteria
- [ ] CI tests installation on Ubuntu
- [ ] Smoke tests catch at least one real issue
- [ ] Test suite runs in < 5 minutes
- [ ] Zero "doesn't install" issues reported by users

## Alternatives Considered

### Alternative: Smart CI with Auto-Fix PR
**Why rejected:** Fixes symptoms, not causes. Still allows bad commits to be created.

### Alternative: Require manual validation in PR template
**Why rejected:** Easy to ignore checklists. Need enforcement, not suggestions.

### Alternative: Block merges with branch protection rules
**Why rejected:** Already have CI checks. Doesn't prevent bad commits, just bad merges.

## Open Questions

1. **Should Phase 2 validation be optional?**
   - Pro: Faster for experts who know what they're doing
   - Con: Creates bypass path that defeats the purpose
   - **Decision:** No opt-out. Fast enough with --fix that it's not burdensome.

2. **Should smoke tests run on every commit or just PR?**
   - Pro (every commit): Catch issues immediately
   - Con (every commit): Slow, expensive
   - **Decision:** PR only. Style checks run on every commit.

3. **How to test GUI apps without display?**
   - Option A: Use Xvfb (virtual framebuffer)
   - Option B: Test with `--help` or `--version` flags
   - Option C: Skip GUI execution test, just check files exist
   - **Decision:** Start with Option C, add Option B for packages that support it

4. **Should we block merges on smoke test failures?**
   - Pro: Guarantees only working packages merge
   - Con: False failures would block valid packages
   - **Decision:** Start as non-blocking, make blocking once stable

## Related Documentation

- `docs/LINTING_ENFORCEMENT_PLAN.md` - Original multi-layer strategy
- `docs/WORKFLOW_IMPROVEMENTS.md` - Lessons from PR #11
- `docs/CASK_CREATION_GUIDE.md` - Rules that validation enforces
- `.github/copilot-instructions.md` - AI agent instructions

## Next Steps

1. **Implement Phase 2 tasks 1-4**
2. **Test with PR #11** - Use validate-and-commit.sh to fix Rancher Desktop
3. **Monitor next 5 PRs** - Track style failure rate
4. **Plan Phase 3** - If Phase 2 achieves zero style failures, proceed with smoke testing

---

**Document Status:** Ready for implementation  
**Approved by:** Design review (2026-02-09)  
**Implementation Start:** Ready to begin Phase 2
