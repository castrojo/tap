# Agent Validation Enforcement - Critical Fix Plan

**Date:** 2026-02-09  
**Priority:** 🔴 CRITICAL - FUNDAMENTAL ERROR  
**Status:** Design Complete - Ready for Implementation

---

## 🚨 THE PROBLEM (CRITICAL)

**Agents continue to commit packages with BASIC linting errors that should NEVER reach CI.**

### Evidence from PR #22 (Copilot - Rancher Desktop)

**CI Failed with 8 style violations:**
```
Cask/StanzaOrder: `preflight` stanza out of order
Cask/StanzaOrder: `binary` stanza out of order
Cask/StanzaOrder: `artifact` stanza out of order (x2)
Cask/StanzaGrouping: stanzas should have no lines between them (x2)
Style/TrailingCommaInArguments: Put a comma after the last parameter (x2)
```

**ALL 8 ERRORS WERE AUTO-CORRECTABLE** ✅

### Why This Is CRITICAL

1. **These are NOT complex errors** - Basic style violations
2. **ALL are auto-fixable** - `tap-validate --fix` would catch them
3. **Instructions exist** - AGENTS.md, packaging skill, best practices doc ALL say to validate
4. **Pre-commit hook exists** - Would have blocked this commit locally
5. **This is the 3rd attempt** - PR #18, PR #19, now PR #22 - ALL failed CI

**Root Cause:** Agents are NOT following the mandatory validation step before committing.

---

## 🔍 ROOT CAUSE ANALYSIS

### What SHOULD Happen (Per Instructions)

From `.github/skills/homebrew-packaging/SKILL.md` Step 2:

```markdown
### Step 2: Validate Package (MANDATORY - NEVER SKIP)

⚠️ VALIDATION IS MANDATORY BEFORE EVERY COMMIT ⚠️

./tap-tools/tap-validate file <path-to-rb-file> --fix
```

### What ACTUALLY Happened (PR #22)

**Copilot's actions:**
1. ✅ Deleted old file
2. ✅ Created new cask with good structure
3. ❌ **SKIPPED VALIDATION** before committing
4. ❌ Committed code with 8 style violations
5. ❌ CI failed

**The agent committed without running `tap-validate --fix`.**

### Why Agents Skip Validation

**Hypothesis 1: Instructions Not Strong Enough**
- Current: "MANDATORY - NEVER SKIP"
- Reality: Agents interpret this as advisory, not blocking
- **Fix:** Make validation actually mandatory (can't proceed without it)

**Hypothesis 2: No Immediate Feedback**
- Validation only fails in CI (2-3 minutes later)
- Agent doesn't know they made an error until CI runs
- **Fix:** Validate BEFORE agent can commit

**Hypothesis 3: Pre-commit Hook Doesn't Run in CI**
- Pre-commit hook is local-only
- Copilot runs in GitHub Actions environment
- **Fix:** Add CI-side validation that blocks merge

**Hypothesis 4: Agents Don't Have Tools**
- Copilot environment may not have `tap-validate` built
- Can't run validation even if they wanted to
- **Fix:** Ensure tools are available in agent environments

**Hypothesis 5: Manual Cask Creation**
- Copilot created cask manually (didn't use `tap-cask generate`)
- Manual creation = higher error rate
- **Fix:** FORCE agents to use tap-tools

---

## 📊 CURRENT STATE ANALYSIS

### What Exists Today

**✅ We Have (Tools):**
1. `tap-validate` - Validates and auto-fixes style issues
2. `tap-cask generate` - Creates valid casks automatically
3. `tap-formula generate` - Creates valid formulas automatically
4. Pre-commit hook - Blocks local commits with errors

**✅ We Have (Documentation):**
1. AGENTS.md - Says "MANDATORY: LOAD THE PACKAGING SKILL"
2. `.github/skills/homebrew-packaging/SKILL.md` - 6-step workflow with validation
3. `docs/AGENT_BEST_PRACTICES.md` - Common errors and prevention
4. Multiple warnings about validation being required

**❌ What's Missing (Enforcement):**
1. **No way to FORCE agents to validate** - They can skip it
2. **No validation in agent environments** - Can't run even if they try
3. **No blocking gate** - Can commit invalid code and find out later in CI
4. **No feedback loop** - Agent doesn't know they failed until minutes later

### Gap Analysis

| Component | Current State | Desired State | Gap |
|-----------|---------------|---------------|-----|
| **Tools** | ✅ Complete | Run in agent env | Not built/available |
| **Documentation** | ✅ Complete | Actually followed | Not enforced |
| **Validation** | ⚠️ Manual | Automatic/blocking | Not mandatory |
| **Feedback** | ❌ Post-commit (CI) | Pre-commit | Too late |
| **Enforcement** | ❌ None | Cannot skip | No mechanism |

---

## 🎯 SOLUTION: MULTI-LAYERED VALIDATION ENFORCEMENT

**Goal:** Make it IMPOSSIBLE for agents to commit invalid code.

**Strategy:** Defense in depth - multiple checkpoints, each one blocking.

### Layer 1: Generate with tap-tools (Prevention)

**FORCE agents to use tap-tools** - Don't allow manual creation.

**Implementation:**

1. **Strengthen skill language:**
   ```markdown
   ### Step 1: Generate Package Using tap-tools (REQUIRED - NO EXCEPTIONS)
   
   ⛔ MANUAL PACKAGE CREATION IS PROHIBITED ⛔
   
   You MUST use tap-tools. Manual creation WILL fail CI.
   
   DO NOT create cask/formula files by hand.
   DO NOT copy and modify existing files.
   DO NOT write Ruby code directly.
   
   USE THE TOOLS:
   ./tap-tools/tap-cask generate <name> <github-url>
   ./tap-tools/tap-formula generate <name> <github-url>
   ```

2. **Add detection in CI:**
   ```yaml
   - name: Detect manual package creation
     run: |
       # Check if package was created manually (heuristics)
       # - Missing XDG environment variables
       # - Incorrect stanza order
       # - Line length violations
       # Exit 1 and tell agent to use tap-tools
   ```

**Why this helps:** tap-tools generate valid code automatically.

### Layer 2: Validate in Agent Environment (Automatic)

**Make validation automatic** - Don't rely on agents remembering to run it.

**Implementation:**

**Option A: GitHub Actions Workflow (Agent-Side)**

Add to `.github/workflows/agent-validation.yml`:

```yaml
name: Agent Validation

on:
  pull_request:
    types: [opened, synchronize]
    paths:
      - 'Casks/**'
      - 'Formula/**'

jobs:
  pre-validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Build tap-validate
        run: |
          cd tap-tools
          go build -o tap-validate ./cmd/tap-validate
      
      - name: Validate changed files (auto-fix)
        run: |
          # Get changed Ruby files
          FILES=$(git diff --name-only origin/main...HEAD | grep -E '\.(rb)$' || true)
          
          if [ -z "$FILES" ]; then
            echo "No Ruby files changed"
            exit 0
          fi
          
          echo "Validating and auto-fixing files..."
          FAILED=0
          
          for FILE in $FILES; do
            if [ -f "$FILE" ]; then
              echo "Validating: $FILE"
              if ./tap-tools/tap-validate file "$FILE" --fix; then
                echo "✓ $FILE validated"
              else
                echo "✗ $FILE FAILED VALIDATION"
                FAILED=1
              fi
            fi
          done
          
          if [ $FAILED -eq 1 ]; then
            echo ""
            echo "❌ VALIDATION FAILED"
            echo ""
            echo "Some files have validation errors that could not be auto-fixed."
            echo "Run locally: ./tap-tools/tap-validate file <filename> --fix"
            exit 1
          fi
          
      - name: Auto-commit fixes (if any)
        if: success()
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          
          if git diff --quiet; then
            echo "No changes to commit"
          else
            git add -A
            git commit -m "style: auto-fix validation errors

Auto-applied by tap-validate --fix

Assisted-by: GitHub Actions"
            git push
          fi
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Why this helps:** Runs immediately when PR is created, auto-fixes errors.

**Option B: Required Status Check (Blocking)**

Make the validation workflow a **required check** before merge:

1. Go to: Repository Settings → Branches → Branch protection rules
2. Add rule for `main` branch
3. Require status check: `pre-validate`
4. **Cannot merge until validation passes**

**Why this helps:** Physically blocks merge until validation succeeds.

### Layer 3: CI Validation (Enforcement)

**Keep existing CI checks** - Final safety net.

Current `.github/workflows/tests.yml` already runs:
- `brew audit --cask` (for casks)
- `brew style` (style checking)

**Enhancement:** Make error messages more helpful.

```yaml
- name: Run brew style
  run: |
    brew style --cask --display-cop-names Casks/*
  continue-on-error: false
  
- name: Show fix instructions (on failure)
  if: failure()
  run: |
    echo "❌ Style validation failed in CI"
    echo ""
    echo "To fix locally before pushing:"
    echo "  ./tap-tools/tap-validate file <filename> --fix"
    echo "  git add <filename>"
    echo "  git commit --amend --no-edit"
    echo "  git push --force-with-lease"
    echo ""
    echo "These errors should have been caught by tap-validate."
    echo "Always run validation before committing."
```

### Layer 4: Documentation Updates (Guidance)

**Make validation steps impossible to miss.**

**Update AGENTS.md:**

```markdown
## ⚠️ CRITICAL CHECKPOINT: VALIDATION ⚠️

**BEFORE EVERY COMMIT, YOU MUST RUN:**

./tap-tools/tap-validate file <path> --fix

**THIS IS NOT OPTIONAL. THIS IS NOT ADVISORY. THIS IS MANDATORY.**

**If you commit without validation:**
- ❌ CI WILL fail
- ❌ PR will be blocked
- ❌ You will waste time fixing errors
- ❌ You will have to redo the commit

**Expected output:**
```
→ Validating sublime-text-linux...
✓ Style check passed
```

**If validation fails, it means you skipped tap-tools.**
**Go back to Step 1 and use the generators.**
```

**Update `.github/skills/homebrew-packaging/SKILL.md`:**

```markdown
### Step 2: Validate Package (⛔ CANNOT SKIP ⛔)

**CHECKPOINT: You CANNOT proceed to Step 3 without completing validation.**

Run validation with auto-fix:
  ./tap-tools/tap-validate file <path> --fix

Expected output:
  ✓ Style check passed

If validation fails:
  - You likely created the package manually (prohibited)
  - Go back to Step 1 and use tap-tools
  - Do NOT manually fix validation errors
  - Regenerate with tap-tools instead

**VERIFICATION REQUIRED:**
Before proceeding, confirm:
  [ ] Ran tap-validate with --fix flag
  [ ] Saw "Style check passed" message
  [ ] No error output from validation

If you cannot check all 3 boxes, STOP and run validation again.
```

---

## 📋 IMPLEMENTATION PLAN

### Phase 1: Immediate Fixes (15 minutes)

**Goal:** Update documentation to be absolutely clear.

1. **Update AGENTS.md** - Add "CRITICAL CHECKPOINT" section
2. **Update packaging skill** - Add "CANNOT SKIP" language
3. **Update AGENT_BEST_PRACTICES.md** - Add PR #22 as example
4. **Commit and push** - Documentation improvements

**Deliverables:**
- Stronger language in all docs
- Clear checkpoints agents cannot skip
- PR #22 documented as failure case

### Phase 2: Agent-Side Validation (30 minutes)

**Goal:** Auto-validate and auto-fix in PRs.

1. **Create `.github/workflows/agent-validation.yml`**
2. **Test with a deliberately broken cask**
3. **Verify auto-fix works**
4. **Make it a required status check**

**Deliverables:**
- Working validation workflow
- Auto-fixes style issues
- Blocks merge if non-fixable errors

### Phase 3: Tool Availability (30 minutes)

**Goal:** Ensure agents have tap-tools available.

1. **Add build step to agent-validation workflow**
2. **Cache built binaries** for speed
3. **Verify Copilot can access tools**

**Deliverables:**
- tap-validate available in all agent environments
- Fast builds (cached)
- Verified with test PR

### Phase 4: Monitoring & Iteration (Ongoing)

**Goal:** Track success rate and iterate.

1. **Monitor next 5 agent PRs**
2. **Measure first-push CI success rate**
3. **Document any new failure modes**
4. **Iterate on enforcement**

**Success Metrics:**
- 100% first-push CI success rate
- Zero style violations in PRs
- Agents consistently use tap-tools

---

## 🎯 SUCCESS CRITERIA

**For this solution to be considered successful:**

1. **✅ Zero CI failures from style violations** (next 10 PRs)
2. **✅ All agents use tap-tools** (no manual creation)
3. **✅ Validation runs automatically** (agent doesn't need to remember)
4. **✅ Fast feedback** (errors caught in <1 minute, not 3+ minutes in CI)
5. **✅ Auto-fix works** (8/8 errors from PR #22 would be auto-fixed)

---

## 🔮 EXPECTED OUTCOMES

### Before (Current State - PR #22)

```
Agent creates cask manually
  → No validation run
  → Commits with 8 style violations
  → Push to GitHub
  → CI runs (3 minutes later)
  → CI fails with style errors
  → Agent or human must fix
  → Push again
  → CI runs again
  → Finally passes

Total time: 6-10 minutes
Success rate: 0% first-push
```

### After (With This Solution)

```
Agent uses tap-tools/tap-cask generate
  → Valid cask created automatically
  → Agent commits
  → Push to GitHub
  → Validation workflow runs (30 seconds)
  → Auto-fixes any style issues
  → Auto-commits fixes
  → CI runs
  → CI passes immediately

Total time: 1-2 minutes
Success rate: 100% first-push
```

---

## 🚧 RISKS & MITIGATIONS

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Agent bypasses validation** | Medium | Required status check blocks merge |
| **Auto-fix breaks code** | Low | tap-validate only fixes style, not logic |
| **Workflow too slow** | Low | Cache Go builds, runs in <30s |
| **Agent environment differs** | Medium | Test in Copilot-like environment |
| **False positives** | Low | tap-validate well-tested on existing casks |

---

## 📝 NEXT STEPS

1. **Review this plan** - Confirm approach is sound
2. **Implement Phase 1** (15 min) - Documentation updates
3. **Implement Phase 2** (30 min) - Agent validation workflow
4. **Test with broken cask** - Verify auto-fix works
5. **Monitor PR #22** - Will next attempt pass?
6. **Track next 5 agent PRs** - Measure success rate

---

## 🔗 RELATED DOCUMENTS

- [AGENTS.md](../../AGENTS.md) - Main agent instructions
- [.github/skills/homebrew-packaging/SKILL.md](../../.github/skills/homebrew-packaging/SKILL.md) - Packaging workflow
- [AGENT_BEST_PRACTICES.md](../AGENT_BEST_PRACTICES.md) - Common errors (add PR #22)
- [PR #22](https://github.com/castrojo/tap/pull/22) - Latest failure (8 style violations)
- [PR #18](https://github.com/castrojo/tap/pull/18) - Previous failure (regex error)
- [PR #19](https://github.com/castrojo/tap/pull/19) - Previous failure (license stanza)

---

**Status:** Plan complete, awaiting implementation approval  
**Priority:** 🔴 CRITICAL - Affects ALL agent automation  
**Estimated Time:** 1-2 hours total  
**Expected ROI:** 100% first-push CI success rate, 5-8 minutes saved per PR
