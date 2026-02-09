# CI Optimization: Homebrew Setup Strategy

**Date:** 2026-02-09  
**Status:** Planning  
**Problem:** Current CI uses `Homebrew/actions/setup-homebrew@master` which updates taps unnecessarily  
**Goal:** Determine fastest, most reliable Homebrew setup for CI validation

## Background

### Discovery

Ubuntu runners (ubuntu-latest) come with **Homebrew pre-installed**:
- Location: `/home/linuxbrew/.linuxbrew`
- Version: 5.0.12+ (maintained by GitHub)
- **NOT in PATH by default** - requires `eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"`

Source: [GitHub Actions Runner Images - Ubuntu 22.04](https://github.com/actions/runner-images/blob/main/images/ubuntu/Ubuntu2204-Readme.md)

### Current Implementation

Our `tests.yml` workflow currently uses:
```yaml
- name: Set up Homebrew
  id: set-up-homebrew
  uses: Homebrew/actions/setup-homebrew@master
```

**What this action does:**
1. Checks if `brew` is in PATH
2. If not found, installs Homebrew (unnecessary - already installed)
3. Adds Homebrew to PATH via `eval "$(brew shellenv)"`
4. Fetches and updates homebrew-core tap (~15-20 seconds)
5. Fetches and updates homebrew-cask tap (~10-15 seconds)
6. Sets up git authentication
7. Configures Homebrew environment variables

**Total time:** ~40-50 seconds

### What We Actually Need

Our CI workflow only runs:
- `brew audit --cask/--strict --online` - Style and best practice checks
- `brew style --cask` - RuboCop style validation

**Neither requires:**
- ❌ Updated homebrew-core tap (we're testing our own tap)
- ❌ Updated homebrew-cask tap (we're testing our own casks)
- ❌ Formula installation from core (validation only)

**We only require:**
- ✅ `brew` command available in PATH
- ✅ Our tap linked/tapped correctly
- ✅ Ability to run audit and style commands

## Comparison: Two Approaches

### Approach 1: Full Setup (Current)

**Implementation:**
```yaml
- name: Set up Homebrew
  uses: Homebrew/actions/setup-homebrew@master

- name: Tap repository
  run: |
    mkdir -p $(brew --repository)/Library/Taps/castrojo
    ln -s $GITHUB_WORKSPACE $(brew --repository)/Library/Taps/castrojo/homebrew-tap
```

**Characteristics:**
- Time: ~40-50 seconds
- Maintained by: Homebrew team
- Updates: core + cask taps
- Complexity: Low (handled by action)

### Approach 2: Simple PATH (Optimized)

**Implementation:**
```yaml
- name: Add Homebrew to PATH
  run: echo "/home/linuxbrew/.linuxbrew/bin" >> $GITHUB_PATH

- name: Tap repository
  run: |
    mkdir -p $(brew --repository)/Library/Taps/castrojo
    ln -s $GITHUB_WORKSPACE $(brew --repository)/Library/Taps/castrojo/homebrew-tap
```

**Characteristics:**
- Time: ~5-10 seconds (estimated)
- Maintained by: Us
- Updates: None (uses pre-installed version)
- Complexity: Low (explicit and simple)

## Testing Plan

### Objective

Measure actual performance difference and verify functionality.

### Test Methodology

**Create experimental branch with both approaches:**

1. **Baseline measurement** (Approach 1 - Current)
   - Run tests.yml workflow 3 times
   - Record total CI time and setup time
   - Verify all tests pass

2. **Optimized measurement** (Approach 2 - Simple PATH)
   - Modify tests.yml to use simple PATH addition
   - Run tests.yml workflow 3 times
   - Record total CI time and setup time
   - Verify all tests pass identically

3. **Comparison analysis**
   - Calculate average time savings
   - Verify no functionality regression
   - Document any differences in behavior

### Test Cases

#### Test Case 1: Style Check (RuboCop)

**File:** Create a test cask with intentional style issue

```ruby
cask "test-style-linux" do
  version "1.0.0"
  sha256 "abc123def456"
  
  url "https://example.com/test.tar.gz"
  name "Test Style"
  desc "This line is intentionally too long to exceed the maximum allowed line length and trigger a RuboCop style violation"
  homepage "https://example.com"
  
  binary "test"
end
```

**Expected:** Both approaches should detect the style violation

#### Test Case 2: Audit Check (Homebrew Best Practices)

**File:** Use existing cask (e.g., `sublime-text-linux.rb`)

**Expected:** Both approaches should pass audit checks

#### Test Case 3: Multiple Changed Files

**File:** Modify 2-3 casks simultaneously

**Expected:** Both approaches should validate all changed files

### Implementation Steps

#### Step 1: Create Experimental Branch

```bash
git checkout -b experiment/ci-homebrew-optimization
```

#### Step 2: Baseline Test (Approach 1)

```bash
# Push current tests.yml to trigger CI
git push -u origin experiment/ci-homebrew-optimization

# Record workflow run times from GitHub Actions UI:
# - Total workflow duration
# - "Set up Homebrew" step duration
# - "Run brew audit" step duration
# - "Run brew style" step duration
```

**Data to collect (3 runs):**
- Run 1: Total time ___s, Setup time ___s
- Run 2: Total time ___s, Setup time ___s  
- Run 3: Total time ___s, Setup time ___s
- Average: Total time ___s, Setup time ___s

#### Step 3: Create Test Cask

```bash
# Create test cask to trigger workflow
cat > Casks/test-optimization-linux.rb <<'EOF'
cask "test-optimization-linux" do
  version "1.0.0"
  sha256 "abc123def456abc123def456abc123def456abc123def456abc123def456abcd"
  
  url "https://github.com/BurntSushi/ripgrep/releases/download/#{version}/ripgrep-#{version}-x86_64-unknown-linux-musl.tar.gz"
  name "Test Optimization"
  desc "Test cask for CI optimization experiment"
  homepage "https://github.com/castrojo/tap"
  
  binary "rg", target: "test-rg"
end
EOF

git add Casks/test-optimization-linux.rb
git commit -m "test: add test cask for CI optimization"
git push
```

**Verify:** Workflow runs and tests pass/fail as expected

#### Step 4: Switch to Approach 2 (Simple PATH)

```bash
# Modify .github/workflows/tests.yml
# Replace "Set up Homebrew" step with simple PATH addition
```

**New tests.yml:**
```yaml
name: Tests

on:
  push:
    branches:
      - main
    paths:
      - 'Formula/**'
      - 'Casks/**'
  pull_request:
    paths:
      - 'Formula/**'
      - 'Casks/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Add Homebrew to PATH
        run: echo "/home/linuxbrew/.linuxbrew/bin" >> $GITHUB_PATH

      - name: Verify Homebrew is available
        run: |
          echo "Homebrew version:"
          brew --version
          echo "Homebrew prefix:"
          brew --prefix

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

      - name: Run brew audit
        if: steps.changed-files.outputs.any_changed == 'true'
        run: |
          for file in ${{ steps.changed-files.outputs.all_changed_files }}; do
            name=$(basename "$file" .rb)
            if [[ "$file" == Casks/* ]]; then
              echo "Auditing cask $name (from $file)"
              brew audit --cask --strict --online "castrojo/tap/$name"
            elif [[ "$file" == Formula/* ]]; then
              echo "Auditing formula $name (from $file)"
              brew audit --strict --online "castrojo/tap/$name"
            fi
          done

      - name: Run brew style
        if: steps.changed-files.outputs.any_changed == 'true'
        run: |
          for file in ${{ steps.changed-files.outputs.all_changed_files }}; do
            name=$(basename "$file" .rb)
            if [[ "$file" == Casks/* ]]; then
              echo "Checking style for cask $name (from $file)"
              brew style --cask "castrojo/tap/$name"
            elif [[ "$file" == Formula/* ]]; then
              echo "Checking style for formula $name (from $file)"
              brew style "castrojo/tap/$name"
            fi
          done
```

```bash
git add .github/workflows/tests.yml
git commit -m "test(ci): use simple PATH instead of setup-homebrew action"
git push
```

#### Step 5: Optimized Measurement (3 runs)

```bash
# Trigger workflow 3 times with dummy commits
echo "# Test run 1" >> Casks/test-optimization-linux.rb
git commit -am "test: run 1"
git push

# Wait for workflow to complete, record times

echo "# Test run 2" >> Casks/test-optimization-linux.rb  
git commit -am "test: run 2"
git push

# Wait for workflow to complete, record times

echo "# Test run 3" >> Casks/test-optimization-linux.rb
git commit -am "test: run 3"
git push

# Wait for workflow to complete, record times
```

**Data to collect (3 runs):**
- Run 1: Total time ___s, Setup time ___s
- Run 2: Total time ___s, Setup time ___s
- Run 3: Total time ___s, Setup time ___s
- Average: Total time ___s, Setup time ___s

#### Step 6: Analyze Results

Create comparison table:

| Metric | Approach 1 (Full Setup) | Approach 2 (Simple PATH) | Improvement |
|--------|-------------------------|--------------------------|-------------|
| Setup Time | ___s | ___s | ___% |
| Total Time | ___s | ___s | ___% |
| Tests Pass | ✅/❌ | ✅/❌ | Same? |

#### Step 7: Document Findings

Create `docs/observations/2026-02-09-ci-homebrew-optimization.md` with:
- Raw data from all test runs
- Screenshots of workflow timings
- Any errors or warnings encountered
- Comparison analysis
- Recommendation

#### Step 8: Cleanup

```bash
# Remove test cask
git rm Casks/test-optimization-linux.rb
git commit -m "test: remove optimization test cask"

# If Approach 2 is better, keep the changes
git push

# If Approach 1 is better, revert
git checkout main .github/workflows/tests.yml
git commit -m "test: revert to setup-homebrew action"
git push

# Merge to main or close branch based on results
```

## Success Criteria

### Approach 2 is Superior If:

1. **Time savings ≥ 20 seconds** per workflow run
2. **No functionality regression** - all tests pass identically
3. **No warnings or errors** introduced
4. **Maintainable** - simple and clear code

### Approach 1 Remains Better If:

1. **Time savings < 10 seconds** - not worth maintenance burden
2. **Functionality breaks** - audit/style commands fail
3. **Warnings appear** - about missing taps or outdated versions
4. **Complexity increases** - requires workarounds or fixes

## Risk Assessment

### Risks of Approach 2 (Simple PATH)

1. **Homebrew version drift**
   - Risk: GitHub may update pre-installed version with breaking changes
   - Mitigation: Monitor runner image changelogs
   - Severity: Low (Homebrew maintains backward compatibility)

2. **Missing environment variables**
   - Risk: Some Homebrew features may require specific env vars
   - Mitigation: Add them if needed (we only use audit/style)
   - Severity: Low (can add env vars easily)

3. **Future CI needs**
   - Risk: Phase 3 smoke testing may need full setup
   - Mitigation: Use Approach 1 for smoke test workflow
   - Severity: Low (separate workflows can use different approaches)

### Risks of Approach 1 (Current)

1. **Wasted CI time**
   - Risk: Slower CI = longer feedback loops
   - Impact: ~20-30 seconds per run × multiple runs per day
   - Severity: Low-Medium (accumulated time waste)

2. **Tap update failures**
   - Risk: homebrew-core/cask updates can fail transiently
   - Impact: CI fails due to upstream issues, not our code
   - Severity: Low (rare but happens)

## Decision Framework

```
┌─────────────────────────────────────────────────┐
│ Run Experiments (Step 1-5)                      │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│ Time savings ≥ 20s AND no regressions?          │
└─────────────────┬───────────────────────────────┘
                  │
         ┌────────┴────────┐
         │                 │
        YES               NO
         │                 │
         ▼                 ▼
┌────────────────┐  ┌─────────────────┐
│ Adopt          │  │ Keep            │
│ Approach 2     │  │ Approach 1      │
│ (Simple PATH)  │  │ (setup-homebrew)│
└────────────────┘  └─────────────────┘
         │                 │
         ▼                 ▼
┌────────────────┐  ┌─────────────────┐
│ Document in    │  │ Document why    │
│ tests.yml and  │  │ full setup is   │
│ this plan      │  │ necessary       │
└────────────────┘  └─────────────────┘
```

## Phase 3 Implications

**Important:** This optimization is specifically for **validation workflows** (audit/style).

**Phase 3 smoke testing** (actual package installation) will likely need:
- Full Homebrew setup via `setup-homebrew@master`
- Updated core tap (for dependencies)
- Proper environment variables

**Strategy:**
- Use Approach 2 (simple PATH) for `tests.yml` (validation)
- Use Approach 1 (full setup) for `test-installation.yml` (smoke tests)
- Each workflow optimized for its specific needs

## Expected Outcomes

### If Approach 2 Succeeds (Expected)

**Benefits:**
- ✅ **20-30 second CI speedup** per workflow run
- ✅ **Simpler, more explicit** setup (easier to debug)
- ✅ **No dependency on action maintenance** (Homebrew team)
- ✅ **Fewer failure points** (no tap updates to fail)

**Implementation:**
- Update tests.yml to use simple PATH
- Document decision in this plan
- Monitor for any issues over next 10 CI runs

### If Approach 1 Remains (Unlikely)

**Reasons might include:**
- Approach 2 causes audit/style failures
- GitHub changes runner image Homebrew setup
- Missing required environment variables
- Community best practice strongly favors setup-homebrew

**Implementation:**
- Keep tests.yml unchanged
- Document why full setup is necessary
- Revisit quarterly in case situation changes

## Timeline

| Task | Estimated Time | Status |
|------|----------------|--------|
| Create experiment branch | 5 min | ⏳ TODO |
| Baseline measurement (3 runs) | 30 min | ⏳ TODO |
| Modify tests.yml | 10 min | ⏳ TODO |
| Optimized measurement (3 runs) | 30 min | ⏳ TODO |
| Analyze and document results | 20 min | ⏳ TODO |
| Update plan with findings | 10 min | ⏳ TODO |
| Cleanup and merge | 10 min | ⏳ TODO |

**Total:** ~2 hours

## Next Steps

1. **Create experiment branch** - `experiment/ci-homebrew-optimization`
2. **Run baseline tests** - Collect Approach 1 timing data
3. **Implement Approach 2** - Modify tests.yml
4. **Run optimized tests** - Collect Approach 2 timing data
5. **Analyze results** - Create observations document
6. **Make decision** - Adopt best approach
7. **Update this plan** - Document findings and decision

---

**Document Status:** Ready for experimentation  
**Approval:** Not required (non-production testing)  
**Implementation:** Can start immediately
