# Linting Enforcement Strategy
## Preventing CI Failures Through Early Validation

**Problem:** Style/linting failures happening at CI stage (too late)  
**Goal:** Catch and fix style issues BEFORE code reaches CI  
**Status:** 🔴 CRITICAL - Implement immediately  

---

## Current Situation Analysis

### Where Linting Happens Now
1. ❌ **CI Stage (GitHub Actions)** - TOO LATE
   - Runs `brew style` on PR
   - Failures block merge
   - Wastes time and creates noise

2. ⚠️ **Manual (Optional)** - INCONSISTENT
   - `./tap-tools/tap-validate --fix` available but not enforced
   - Developers can forget to run it
   - AI agents don't always use it

3. ❌ **tap-tools Generation** - PARTIAL
   - `tap-cask` and `tap-formula` generate valid code
   - BUT: Manual edits after generation bypass validation
   - No validation run after generation completes

### Recent Failures
- **PR #11 (Rancher Desktop):** Line too long + array ordering
  - Root cause: No validation before commit
  - Impact: CI failure, requires rework

---

## The Linting Enforcement Ladder

### Level 1: Generator Auto-Validation (IMMEDIATE)
**What:** Make tap-tools auto-run validation after generation

**Implementation:** Modify tap-cask/tap-formula to:
```go
// After generating file
fmt.Println("✓ Generated:", outputPath)
fmt.Println("\n🔍 Running validation...")

cmd := exec.Command("./tap-tools/tap-validate", "file", outputPath, "--fix")
output, err := cmd.CombinedOutput()
fmt.Println(string(output))

if err != nil {
    fmt.Println("⚠️  Validation failed - please review and fix manually")
} else {
    fmt.Println("✅ Validation passed")
}
```

**Benefits:**
- Zero-effort validation for generated packages
- Catches issues immediately
- Users see validation results right away

**Effort:** Low (1 hour)

---

### Level 2: Git Pre-Commit Hook (CRITICAL)
**What:** Automatically validate Ruby files before each commit

**Implementation:** Create `.git/hooks/pre-commit`

```bash
#!/bin/bash
# Pre-commit hook: Validate Ruby files before commit

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔍 Running pre-commit validation..."

# Check if tap-validate exists
if [ ! -f "./tap-tools/tap-validate" ]; then
    echo -e "${RED}✗ tap-validate not found!${NC}"
    echo "Build it with: cd tap-tools && go build"
    exit 1
fi

# Get staged Ruby files
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(rb)$' || true)

if [ -z "$STAGED_FILES" ]; then
    echo -e "${GREEN}✓ No Ruby files to validate${NC}"
    exit 0
fi

VALIDATION_FAILED=0

for FILE in $STAGED_FILES; do
    if [ -f "$FILE" ]; then
        echo ""
        echo "Validating: $FILE"
        
        # Run validation with --fix
        if ./tap-tools/tap-validate file "$FILE" --fix; then
            echo -e "${GREEN}✓ $FILE passed validation${NC}"
            
            # Re-stage file if it was auto-fixed
            git add "$FILE"
        else
            echo -e "${RED}✗ $FILE failed validation${NC}"
            VALIDATION_FAILED=1
        fi
    fi
done

if [ $VALIDATION_FAILED -eq 1 ]; then
    echo ""
    echo -e "${RED}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║  COMMIT BLOCKED: Style validation failed                  ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Some files failed validation. Please fix the issues and try again."
    echo ""
    echo "To fix automatically:"
    echo "  ./tap-tools/tap-validate file <filename> --fix"
    echo ""
    echo "To bypass this hook (NOT RECOMMENDED):"
    echo "  git commit --no-verify"
    echo ""
    exit 1
fi

echo ""
echo -e "${GREEN}✅ All validations passed!${NC}"
exit 0
```

**Distribution Strategy:**

Since `.git/hooks/` is not versioned, we need to:

1. **Create template:** `scripts/git-hooks/pre-commit`
2. **Setup script:** `scripts/setup-hooks.sh`
   ```bash
   #!/bin/bash
   echo "Installing git hooks..."
   cp scripts/git-hooks/pre-commit .git/hooks/pre-commit
   chmod +x .git/hooks/pre-commit
   echo "✓ Pre-commit hook installed"
   ```
3. **Auto-setup on clone:** Add to README
4. **CI verification:** Check hook exists in CI

**Benefits:**
- Impossible to commit invalid Ruby files
- Auto-fix applied automatically
- Files re-staged after fix
- Clear error messages
- Bypassable in emergencies (--no-verify)

**Effort:** Medium (2-3 hours)

---

### Level 3: GitHub Actions Status Check (ENHANCED)
**What:** Make style check a separate, prominent status

**Current:** Combined with audit in single "test" job

**Proposed:** Split into separate jobs for visibility

```yaml
name: Tests

on:
  pull_request:
    paths:
      - 'Formula/**'
      - 'Casks/**'

jobs:
  style:
    name: Style Check (RuboCop)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: Homebrew/actions/setup-homebrew@master
      
      - name: Tap repository
        run: |
          mkdir -p $(brew --repository)/Library/Taps/castrojo
          ln -s $GITHUB_WORKSPACE $(brew --repository)/Library/Taps/castrojo/homebrew-tap
      
      - name: Get changed files
        id: changed-files
        uses: tj-actions/changed-files@v46
        with:
          files: |
            Formula/**
            Casks/**
      
      - name: Style check (with auto-fix suggestion)
        if: steps.changed-files.outputs.any_changed == 'true'
        run: |
          FAILED=0
          for file in ${{ steps.changed-files.outputs.all_changed_files }}; do
            name=$(basename "$file" .rb)
            if [[ "$file" == Casks/* ]]; then
              echo "::group::Style check: $name"
              if ! brew style --cask "castrojo/tap/$name"; then
                echo "::error file=$file::Style check failed. Run: ./tap-tools/tap-validate file $file --fix"
                FAILED=1
              fi
              echo "::endgroup::"
            elif [[ "$file" == Formula/* ]]; then
              echo "::group::Style check: $name"
              if ! brew style "castrojo/tap/$name"; then
                echo "::error file=$file::Style check failed. Run: ./tap-tools/tap-validate file $file --fix"
                FAILED=1
              fi
              echo "::endgroup::"
            fi
          done
          
          if [ $FAILED -eq 1 ]; then
            echo ""
            echo "❌ Style check failed!"
            echo ""
            echo "To fix locally:"
            echo "  gh pr checkout ${{ github.event.pull_request.number }}"
            echo "  ./tap-tools/tap-validate all --fix"
            echo "  git commit -am 'style: fix RuboCop violations'"
            echo "  git push"
            exit 1
          fi

  audit:
    name: Brew Audit
    runs-on: ubuntu-latest
    steps:
      # ... audit steps (separate from style)
```

**Benefits:**
- Clear separation of concerns
- Style failures are prominent
- Includes fix instructions in error
- Can be required status check

**Effort:** Low (1 hour)

---

### Level 4: Copilot Instructions Enhancement (IMMEDIATE)
**What:** Make validation absolutely mandatory in documentation

**File:** `.github/copilot-instructions.md`

**Add prominent section at top:**

```markdown
# ⚠️ CRITICAL: VALIDATION IS MANDATORY

**BEFORE EVERY COMMIT, YOU MUST RUN VALIDATION:**

\`\`\`bash
# For single file (REQUIRED after creating/editing)
./tap-tools/tap-validate file Casks/your-cask-linux.rb --fix

# For all files
./tap-tools/tap-validate all --fix
\`\`\`

**❌ NEVER commit without validation passing**
**❌ NEVER skip this step - it prevents CI failures**
**❌ NEVER use --no-verify to bypass pre-commit hooks**

**Expected output:**
\`\`\`
✓ Style check passed
\`\`\`

If validation fails:
1. Review the error message
2. The --fix flag auto-corrects most issues
3. Re-run validation until passing
4. Only then commit
```

**Update existing validation section:**

```markdown
## Step-by-Step: Creating a Package

1. Generate package
   \`\`\`bash
   ./tap-tools/tap-cask generate https://github.com/user/repo
   \`\`\`

2. **[MANDATORY] Validate immediately**
   \`\`\`bash
   ./tap-tools/tap-validate file Casks/package-name-linux.rb --fix
   \`\`\`
   
3. Review generated code
   
4. **[MANDATORY] If you made any edits, validate again**
   \`\`\`bash
   ./tap-tools/tap-validate file Casks/package-name-linux.rb --fix
   \`\`\`

5. Only commit if validation passes

6. Push to remote
```

**Add validation checklist:**

```markdown
## Pre-Commit Checklist (MANDATORY)

Before EVERY commit, verify:

- [ ] Generated/edited Ruby files
- [ ] Ran `./tap-tools/tap-validate file <file> --fix`
- [ ] Validation output shows `✓ Style check passed`
- [ ] Reviewed any auto-fixes applied
- [ ] File is re-staged if auto-fixed
- [ ] No `--no-verify` used to bypass hooks

**If ANY checkbox is unchecked, DO NOT COMMIT.**
```

**Effort:** Low (30 minutes)

---

### Level 5: Validation Helper Script (CONVENIENCE)
**What:** Wrapper script that makes validation foolproof

**File:** `scripts/validate-and-commit.sh`

```bash
#!/bin/bash
# Validate-and-commit helper
# Usage: ./scripts/validate-and-commit.sh "commit message"

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 \"commit message\""
    exit 1
fi

COMMIT_MSG="$1"

echo "🔍 Validating all changed files..."

# Get all changed/staged Ruby files
CHANGED_FILES=$(git diff --name-only --diff-filter=ACM HEAD | grep -E '\.(rb)$' || true)
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(rb)$' || true)
ALL_FILES=$(echo "$CHANGED_FILES" "$STAGED_FILES" | tr ' ' '\n' | sort -u)

if [ -z "$ALL_FILES" ]; then
    echo "No Ruby files changed"
    git commit -m "$COMMIT_MSG"
    exit 0
fi

VALIDATION_FAILED=0

for FILE in $ALL_FILES; do
    if [ -f "$FILE" ]; then
        echo ""
        echo "Validating: $FILE"
        
        if ./tap-tools/tap-validate file "$FILE" --fix; then
            echo "✓ $FILE passed"
            git add "$FILE"  # Re-stage after auto-fix
        else
            echo "✗ $FILE FAILED"
            VALIDATION_FAILED=1
        fi
    fi
done

if [ $VALIDATION_FAILED -eq 1 ]; then
    echo ""
    echo "❌ Validation failed - commit aborted"
    exit 1
fi

echo ""
echo "✅ All validations passed - committing..."
git commit -m "$COMMIT_MSG"
```

**Usage:**
```bash
# Instead of: git commit -m "message"
./scripts/validate-and-commit.sh "feat(cask): add new package"
```

**Effort:** Low (30 minutes)

---

### Level 6: Editor Integration (FUTURE)
**What:** Real-time linting in editors

**VS Code:** Add `.vscode/settings.json`
```json
{
  "ruby.rubocop.executePath": "brew",
  "ruby.rubocop.onSave": true,
  "ruby.format": "rubocop"
}
```

**Vim/Neovim:** Add to README
```vim
" Using ALE
let g:ale_linters = {'ruby': ['rubocop']}
let g:ale_fixers = {'ruby': ['rubocop']}
let g:ale_ruby_rubocop_executable = 'brew'
let g:ale_fix_on_save = 1
```

**Effort:** Low (documentation only)

---

## Implementation Priority

### 🔴 Phase 1: IMMEDIATE (COMPLETED 2026-02-09)
1. ✅ **Update Copilot instructions** - Add mandatory validation section (30 min)
2. ✅ **Create pre-commit hook template** - `scripts/git-hooks/pre-commit` (1 hour)
3. ✅ **Create setup script** - `scripts/setup-hooks.sh` (15 min)
4. ✅ **Install locally** - Test the hook works (15 min)
5. ✅ **Update README** - Add setup instructions (15 min)

**Status:** ✅ COMPLETE - All items already existed in repo, documentation updated

### 🟡 Phase 2: SHORT-TERM (COMPLETED 2026-02-09)
1. ✅ **Modify tap-tools** - Auto-validate after generation - ALREADY IMPLEMENTED in Go
   - tap-cask: Lines 286-300 call `validate.ValidateFile()` with auto-fix
   - tap-formula: Lines 300-317 call `validate.ValidateFile()` with auto-fix
2. ✅ **Split CI workflow** - Separate style and audit jobs (1 hour) - PR #17 created
3. ❌ **Create validation helper** - SKIPPED (no shell scripts per repo policy)
4. ✅ **Update all documentation** - Ensure consistency
   - Updated AGENTS.md with first-time setup section
   - Codified use of `gh` CLI for all GitHub operations
   - Updated PR workflow to use `gh pr create`

**Status:** ✅ COMPLETE - All critical work done, validation helper skipped per policy

### 🟢 Phase 3: FUTURE (Not Started)
1. ⏳ **Editor integration docs** - Add linting setup guides
2. ⏳ **Metrics tracking** - Monitor validation compliance
3. ⏳ **Pre-push validation** - Catch issues before push

**Status:** 🟢 DEFERRED - Monitor effectiveness of current implementation first

---

## Success Metrics

Track these to measure effectiveness:

1. **CI Pass Rate**
   - Baseline: ~70% (estimated from PR #11 failure)
   - Target: 95%
   - Measure: GitHub Actions success rate

2. **Style Failure Rate**
   - Baseline: 100% (PR #11 failed on style)
   - Target: <5%
   - Measure: Count of style-only CI failures

3. **Pre-Commit Hook Usage**
   - Target: 100% of commits pass hook
   - Measure: Hook execution logs

4. **Time to Fix**
   - Baseline: Hours (wait for CI, fix, re-push)
   - Target: Seconds (auto-fix on commit)
   - Measure: Time from code complete to merge

---

## Testing Plan

### Test Pre-Commit Hook
```bash
# 1. Install hook
./scripts/setup-hooks.sh

# 2. Create intentionally bad file
cat > Casks/test-bad.rb <<'EOF'
cask "test-bad" do
  version "1.0.0"
  sha256 "abc123"
  url "https://example.com/test.tar.gz"
  name "Test"
  desc "This line is way too long and will definitely exceed the maximum allowed line length of 118 characters for RuboCop"
  homepage "https://example.com"
  binary "test"
end
EOF

# 3. Try to commit (should fail)
git add Casks/test-bad.rb
git commit -m "test: bad file" 
# Expected: Hook blocks commit, shows error

# 4. Fix with validation
./tap-tools/tap-validate file Casks/test-bad.rb --fix

# 5. Commit should now succeed
git commit -m "test: fixed file"
# Expected: Commit succeeds

# 6. Cleanup
git reset --soft HEAD~1
git restore --staged Casks/test-bad.rb
rm Casks/test-bad.rb
```

---

## Summary (Updated 2026-02-09)

### ✅ Completed Work

**Level 1: Generator Auto-Validation**
- Already implemented in Go tools (tap-cask and tap-formula)
- Both tools call `validate.ValidateFile()` after generation with auto-fix
- No changes needed - functionality already exists

**Level 2: Git Pre-Commit Hook**
- Hook already exists at `scripts/git-hooks/pre-commit`
- Setup script already exists at `scripts/setup-hooks.sh`
- Hook calls tap-validate Go tool (no shell script duplication)
- Documentation updated in AGENTS.md and README.md

**Level 3: Split CI Workflow**
- Created PR #17 to split style and audit into separate jobs
- Added GitHub annotations for file-specific errors
- Added helpful fix instructions using tap-validate
- Jobs run in parallel for faster feedback

**Documentation Updates**
- Added "First-Time Setup" section to AGENTS.md
- Codified use of `gh` CLI for all GitHub operations
- Updated PR workflow to use `gh pr create` (auto-pushes)
- Clarified when to use `git` vs `gh` commands

### 📊 Expected Results

With Phases 1-2 complete, we should see:
- ✅ Pre-commit hook blocks invalid Ruby files locally
- ✅ Generators auto-validate and fix style issues
- ✅ CI provides clear, actionable error messages
- ✅ Style failures are prominent as separate status check
- 🎯 Target: >95% CI pass rate (up from ~70% baseline)
- 🎯 Target: <5% style-only failures

### 🚀 Next Steps

1. **Monitor effectiveness** - Track CI pass rate over next 2 weeks
2. **Wait for PR #17 to merge** - Split CI workflow changes
3. **Collect feedback** - Identify any gaps in validation coverage
4. **Phase 3 (optional)** - Add editor integration docs if needed

---

## Next Steps

**Immediate actions:**
1. ✅ Create pre-commit hook files - Already existed
2. ✅ Update `.github/copilot-instructions.md` - Already complete
3. ✅ Test locally - Verified working
4. ✅ Commit and push changes - All pushed to main
5. ✅ Split CI workflow - PR #17 created

**Follow-up:**
1. Monitor next PR to verify effectiveness
2. Adjust based on feedback
3. Implement Phase 2 enhancements

---

## Conclusion

The linting enforcement ladder provides multiple defensive layers:

1. **Generator** - Auto-validate on creation
2. **Pre-commit hook** - Block invalid commits locally
3. **CI** - Final safety net with clear error messages
4. **Documentation** - Make process obvious for humans and AI

With all layers in place, style failures should become **extremely rare** instead of the current "too often" situation.

**Status:** ⏳ READY TO IMPLEMENT
