# Implementation Task List

**Generated:** 2026-02-09  
**Purpose:** Actionable task list for all planned work  
**Status:** Ready for execution

This document consolidates all implementation tasks from various plans into a single, prioritized checklist.

---

## 🔴 Priority 1: Critical (Do Immediately)

### Task 1.1: Phase 2 - Auto-Validation in tap-tools

**From:** [plans/2026-02-09-zero-error-packages-design.md](plans/2026-02-09-zero-error-packages-design.md)  
**Effort:** 2 hours  
**Status:** ⏳ TODO

**Subtasks:**
- [ ] Update `tap-cask` to run validation after generation
  - [ ] Call `tap-validate` internally after writing file
  - [ ] Show verbose output (all fixes applied)
  - [ ] Exit with error if validation fails
  - [ ] Auto-rewrite file with fixes
- [ ] Update `tap-formula` with same validation logic
  - [ ] Call `tap-validate` internally after writing file
  - [ ] Show verbose output
  - [ ] Exit with error if validation fails
  - [ ] Auto-rewrite file with fixes
- [ ] Test both tools generate valid, commit-ready files
- [ ] Update tool README with new behavior

**Acceptance Criteria:**
- ✅ Generated files always pass validation
- ✅ Tools exit with error if validation cannot fix issues
- ✅ Verbose output shows all validation steps
- ✅ No `--skip-validation` flag exists (mandatory)

---

### Task 1.2: GitHub Actions SHA Pinning

**From:** [GITHUB_ACTIONS_POLICY.md](GITHUB_ACTIONS_POLICY.md)  
**Effort:** 1 hour  
**Status:** ⏳ TODO

**Subtasks:**
- [ ] Get current SHAs for all actions
  ```bash
  # actions/checkout@v4
  gh api repos/actions/checkout/git/ref/tags/v4.1.1 --jq '.object.sha'
  
  # actions/github-script@v7
  gh api repos/actions/github-script/git/ref/tags/v7.0.1 --jq '.object.sha'
  
  # Homebrew/actions/setup-homebrew@master
  gh api repos/Homebrew/actions/git/ref/heads/master --jq '.object.sha'
  
  # ruby/setup-ruby@v1
  gh api repos/ruby/setup-ruby/git/ref/tags/v1.195.0 --jq '.object.sha'
  
  # tj-actions/changed-files@v46
  gh api repos/tj-actions/changed-files/git/ref/tags/v46.0.2 --jq '.object.sha'
  ```
- [ ] Update `.github/workflows/tests.yml` with SHA pins
- [ ] Update `.github/workflows/cask-sha256-update.yml` with SHA pins
- [ ] Verify renovate.json5 has GitHub Actions config (✅ already done)
- [ ] Test workflows still pass with SHA pins
- [ ] Commit with security policy message

**Acceptance Criteria:**
- ✅ All actions pinned to commit SHAs
- ✅ Comment shows original tag (e.g., `@<sha> # v4.1.1`)
- ✅ Workflows pass with SHA pins
- ✅ Renovate configured to update actions

---

### Task 1.3: Bash Script Deprecation

**From:** [GO_MIGRATION_PLAN.md](GO_MIGRATION_PLAN.md)  
**Effort:** 2-3 hours  
**Status:** ⏳ TODO

**Subtasks:**
- [ ] **Step 1: Audit** - List all bash scripts and their status
  ```bash
  ls -la scripts/
  # For each script, document: superseded by X, or keep with justification
  ```
- [ ] **Step 2: Verify feature parity**
  - [ ] Compare `new-formula.sh` vs `tap-formula` functionality
  - [ ] Compare `new-cask.sh` vs `tap-cask` functionality
  - [ ] Compare `from-issue.sh` vs `tap-issue` functionality
  - [ ] Compare `validate-all.sh` vs `tap-validate all` functionality
  - [ ] Compare `update-sha256.sh` vs existing Go/renovate functionality
  - [ ] Document any missing features
- [ ] **Step 3: Port missing features** (if any found)
  - [ ] Add missing functionality to Go tools
  - [ ] Test feature parity
- [ ] **Step 4: Update documentation**
  - [ ] Remove bash script references from README.md
  - [ ] Update docs/AGENT_GUIDE.md (already uses Go tools ✅)
  - [ ] Update .github/copilot-instructions.md (already uses Go tools ✅)
  - [ ] Search for `scripts/` references: `grep -r "scripts/" .github/ docs/ *.md`
  - [ ] Update any workflow files that call scripts
- [ ] **Step 5: Remove bash scripts**
  - [ ] Delete `scripts/new-formula.sh`
  - [ ] Delete `scripts/new-cask.sh`
  - [ ] Delete `scripts/from-issue.sh`
  - [ ] Delete `scripts/validate-all.sh`
  - [ ] Delete `scripts/update-sha256.sh`
  - [ ] Keep `scripts/setup-hooks.sh` (one-time setup)
  - [ ] Keep `scripts/git-hooks/` (git hooks)
- [ ] **Step 6: Commit removal**
  - [ ] Single commit removing all deprecated scripts
  - [ ] Include "BREAKING CHANGE" in commit message
  - [ ] Document migration path in commit body

**Acceptance Criteria:**
- ✅ All bash scripts removed except setup/hooks
- ✅ All functionality preserved in Go tools
- ✅ Documentation updated to remove script references
- ✅ No references to removed scripts in code/docs

---

## 🟡 Priority 2: High (Do Soon)

### Task 2.1: CI Optimization Experiment

**From:** [plans/2026-02-09-ci-optimization-homebrew-setup.md](plans/2026-02-09-ci-optimization-homebrew-setup.md)  
**Effort:** 2 hours  
**Status:** ⏳ TODO

**Subtasks:**
- [ ] Create experiment branch `experiment/ci-homebrew-optimization`
- [ ] **Baseline measurement** (Approach 1 - current)
  - [ ] Trigger workflow 3 times
  - [ ] Record total time and setup time for each run
  - [ ] Calculate average
- [ ] **Optimized measurement** (Approach 2 - simple PATH)
  - [ ] Modify tests.yml to use simple PATH addition
  - [ ] Change runner to ubuntu-24.04 (forward-looking)
  - [ ] Trigger workflow 3 times
  - [ ] Record total time and setup time for each run
  - [ ] Calculate average
- [ ] **Analysis**
  - [ ] Create comparison table
  - [ ] Calculate time savings
  - [ ] Verify no functionality regression
  - [ ] Document findings in observations/
- [ ] **Decision**
  - [ ] If savings ≥ 30 seconds: adopt Approach 2 immediately (aggressive)
  - [ ] If savings < 30 seconds: keep Approach 1, document why
- [ ] **Cleanup**
  - [ ] Merge winning approach to main (or revert)
  - [ ] Delete experiment branch
  - [ ] Update plan with results

**Acceptance Criteria:**
- ✅ Measured time savings documented
- ✅ Decision documented with data
- ✅ If adopted: tests.yml uses simple PATH on ubuntu-24.04
- ✅ If rejected: documented why optimization not worth it

---

## 🟢 Priority 3: Medium (Valuable but Not Urgent)

### Task 3.1: Investigate tap-cask Metadata Enhancement

**From:** User feedback on offline testing plan  
**Effort:** 2 hours  
**Status:** ⏳ TODO (Investigation phase)

**Purpose:** Determine if we need elaborate offline testing infrastructure, or if enhancing tap-cask metadata output is sufficient.

**Subtasks:**
- [ ] **Review current tap-cask output**
  - [ ] What metadata is currently shown?
  - [ ] What's missing that agents might need?
- [ ] **Identify gaps**
  - [ ] Has Copilot created incomplete casks due to missing metadata?
  - [ ] What desktop integration info is missing?
  - [ ] What binary path info is missing?
- [ ] **Design enhanced output**
  - [ ] Proposal: Add `--verbose` flag to show more metadata
  - [ ] Proposal: Generate `.metadata.json` file alongside `.rb` file
  - [ ] Alternative: Improve in-file comments with structure info
- [ ] **Prototype solution**
  - [ ] Implement simplest approach
  - [ ] Test with a few casks
  - [ ] Evaluate if this solves the problem
- [ ] **Decision**
  - [ ] If enhanced metadata sufficient: close offline testing plan
  - [ ] If not sufficient: revisit offline testing plan later
  - [ ] Document decision and rationale

**Acceptance Criteria:**
- ✅ Decision documented: enhance metadata vs full offline testing
- ✅ If metadata enhancement: implemented and tested
- ✅ If offline testing needed: plan updated with timeline

---

### Task 3.2: Renovate SHA256 Automation (Optional)

**From:** [plans/2026-02-09-renovate-sha256-automation.md](plans/2026-02-09-renovate-sha256-automation.md)  
**Effort:** 4 hours  
**Status:** 📋 Deferred (current automation works)

**Note:** Current SHA256 automation via GitHub Actions is working well. This plan proposes switching formulas to Renovate's built-in manager. Defer until pain point emerges.

**When to implement:**
- Current workflow starts failing frequently
- Manual intervention required > 1x per month
- Desire for fully hands-off updates

---

## 🔵 Priority 4: Low (Future Work)

### Task 4.1: Phase 3 Smoke Testing

**From:** [plans/2026-02-09-zero-error-packages-design.md](plans/2026-02-09-zero-error-packages-design.md)  
**Effort:** 8-10 hours  
**Status:** 📅 Future (blocked by Phase 2 completion)

**Trigger:** After Phase 2 achieves zero style failures for 5-10 consecutive PRs

**Requirements (Decided):**
- Tests are BLOCKING (prevent merge on failure)
- Retry strategy to limit flakes (3 attempts, exponential backoff)
- ubuntu-24.04 only (sufficient for our use case)
- Acceptable failure rate: ≤ 2%

**Subtasks:**
- [ ] Create `.github/workflows/test-installation.yml`
- [ ] Create `scripts/test-formula.sh` smoke test script
- [ ] Create `scripts/test-cask.sh` smoke test script
- [ ] Implement retry strategy (use `nick-invision/retry@v2`)
- [ ] Test on 3-5 existing packages
- [ ] Monitor failure rate for 1 week
- [ ] Make blocking if failure rate < 2%

**Acceptance Criteria:**
- ✅ Workflow installs packages successfully
- ✅ Smoke tests catch at least one real issue
- ✅ False positive rate < 2%
- ✅ Tests block merges when they fail

---

### Task 4.2: Offline Testing Infrastructure (Optional)

**From:** [plans/2026-02-09-offline-testing-for-copilot.md](plans/2026-02-09-offline-testing-for-copilot.md)  
**Effort:** 6 hours  
**Status:** 🔵 On Hold (investigate Task 3.1 first)

**Note:** Elaborate infrastructure for caching package metadata. Only implement if Task 3.1 (enhanced metadata) proves insufficient.

---

## Progress Tracking

### Completion Summary

| Priority | Total Tasks | Completed | Remaining | % Complete |
|----------|-------------|-----------|-----------|------------|
| 🔴 P1    | 3           | 0         | 3         | 0%         |
| 🟡 P2    | 1           | 0         | 1         | 0%         |
| 🟢 P3    | 2           | 0         | 2         | 0%         |
| 🔵 P4    | 2           | 0         | 2         | 0%         |
| **Total**| **8**       | **0**     | **8**     | **0%**     |

### Sprint Plan (Aggressive Execution)

**Week 1 Focus: Critical Tasks**
- Day 1: Task 1.1 (tap-tools auto-validation) - 2 hours
- Day 1: Task 1.2 (SHA pinning) - 1 hour
- Day 2-3: Task 1.3 (bash script deprecation) - 2-3 hours

**Week 2 Focus: High Priority**
- Day 1-2: Task 2.1 (CI optimization) - 2 hours

**Week 3+: Medium/Low Priority**
- As time permits: Task 3.1, 3.2
- Future: Task 4.1, 4.2 (blocked or deferred)

### Update Frequency

Update this document:
- ✅ After completing each task
- ✅ When new tasks are identified
- ✅ When priorities change

---

## Quick Reference: Next Actions

**Right now (next session):**
1. ✅ Complete Task 1.1: tap-tools auto-validation
   - Start with tap-cask, then tap-formula
   - Test thoroughly
   - Ship immediately if tests pass (aggressive)

2. ✅ Complete Task 1.2: SHA pinning
   - Quick task, security improvement
   - No dependencies

**Then:**
3. ✅ Complete Task 1.3: Bash script deprecation
   - Requires verification step
   - Document thoroughly
   - Ship aggressively

**After P1 complete:**
4. ✅ Run Task 2.1: CI optimization experiment
   - Potential 30+ second savings
   - Ship immediately if successful

---

## Notes

**Aggressive shipping strategy:**
- Ship tasks immediately if tests pass
- Document thoroughly in commits
- No PR required for low-risk changes
- Add more testing later if issues emerge

**Documentation strategy:**
- All work is documented in plans/
- Observations/ captures lessons learned
- This task list consolidates everything

**Review cadence:**
- Update task list after each completion
- Review priorities weekly
- Adjust based on pain points

---

**Last Updated:** 2026-02-09  
**Next Review:** After completing Priority 1 tasks
