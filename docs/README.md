# Documentation Index

**Last Updated:** 2026-02-09

This directory contains all documentation for the homebrew-tap project, organized by category.

---

## Quick Start

**For AI Agents:** Start here
- **[AGENT_BEST_PRACTICES.md](AGENT_BEST_PRACTICES.md)** - ⚠️ **READ THIS FIRST!** Common errors and how to avoid them
- [AGENT_GUIDE.md](AGENT_GUIDE.md) - Comprehensive workflow for AI agents
- [CASK_CREATION_GUIDE.md](CASK_CREATION_GUIDE.md) - **Critical rules** for creating casks (read first!)

**For Humans:** Start here
- [../README.md](../README.md) - Repository overview and setup
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common errors and solutions

---

## Core Documentation

### Creation Guides

| Document | Purpose | Status |
|----------|---------|--------|
| [AGENT_BEST_PRACTICES.md](AGENT_BEST_PRACTICES.md) | **Common errors and prevention** (READ FIRST!) | ✅ Complete |
| [CASK_CREATION_GUIDE.md](CASK_CREATION_GUIDE.md) | Rules for creating casks (GUI apps) | ✅ Complete |
| [FORMULA_PATTERNS.md](FORMULA_PATTERNS.md) | Copy-paste templates for formulas | ✅ Complete |
| [CASK_PATTERNS.md](CASK_PATTERNS.md) | Copy-paste templates for casks | ✅ Complete |
| [AGENT_GUIDE.md](AGENT_GUIDE.md) | End-to-end workflow for agents | ✅ Complete |

### Process Documentation

| Document | Purpose | Status |
|----------|---------|--------|
| [RENOVATE_GUIDE.md](RENOVATE_GUIDE.md) | Automated version updates | ✅ Complete |
| [GITHUB_ACTIONS_POLICY.md](GITHUB_ACTIONS_POLICY.md) | SHA pinning and security policy | 📋 Ready to Implement |
| [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | Error resolution guide | ✅ Complete |
| [WORKFLOW_IMPROVEMENTS.md](WORKFLOW_IMPROVEMENTS.md) | Lessons from PR #11 | ✅ Complete |

### Tool Documentation

| Document | Purpose | Status |
|----------|---------|--------|
| [../tap-tools/README.md](../tap-tools/README.md) | Go CLI tools documentation | ✅ Complete |
| [GO_MIGRATION_PLAN.md](GO_MIGRATION_PLAN.md) | Bash → Go migration strategy | 🟡 In Progress |

### Infrastructure Documentation

| Document | Purpose | Status |
|----------|---------|--------|
| [infrastructure/README.md](infrastructure/README.md) | Infrastructure issues index | ✅ Complete |
| [infrastructure/GITHUB_TOKEN_SOLUTION_PLAN.md](infrastructure/GITHUB_TOKEN_SOLUTION_PLAN.md) | GitHub API rate limiting solution | 📋 Ready to Implement |

---

## Implementation Plans

**Active plans ready for implementation:**

### Phase 1: Validation Enforcement ✅ COMPLETE

| Plan | Status | Priority | Effort |
|------|--------|----------|--------|
| [LINTING_ENFORCEMENT_PLAN.md](LINTING_ENFORCEMENT_PLAN.md) | ✅ Phase 1 Done | 🔴 Critical | - |

**Completed:**
- ✅ Pre-commit hooks installed
- ✅ Setup script created
- ✅ Documentation updated

### Phase 2: Zero-Error Package Creation ⏳ IN PROGRESS

| Plan | Status | Priority | Effort |
|------|--------|----------|--------|
| [plans/2026-02-09-zero-error-packages-design.md](plans/2026-02-09-zero-error-packages-design.md) | ⏳ Task 1-2 Remaining | 🔴 Critical | ~2 hours |

**Remaining Tasks:**
1. ⏳ Update `tap-cask` to auto-validate after generation (1 hour)
2. ⏳ Update `tap-formula` to auto-validate after generation (1 hour)
3. ✅ Document complete workflow (DONE)

**Goal:** Prevent style violations at generation time, not just commit time.

### CI Optimization 📋 READY TO TEST

| Plan | Status | Priority | Effort |
|------|--------|----------|--------|
| [plans/2026-02-09-ci-optimization-homebrew-setup.md](plans/2026-02-09-ci-optimization-homebrew-setup.md) | 📋 Ready for Experiment | 🟡 Medium | ~2 hours |

**What:** Test simple PATH vs full setup-homebrew action  
**Expected:** 20-30 second speedup per CI run  
**Risk:** Low (easy rollback)

### Phase 3: Smoke Testing 📅 FUTURE

Documented in [plans/2026-02-09-zero-error-packages-design.md](plans/2026-02-09-zero-error-packages-design.md#phase-3-smoke-testing-future)

**What:** CI actually installs packages and tests functionality  
**When:** After Phase 2 achieves zero style failures  
**Effort:** ~8-10 hours

---

## Automation Plans (Design Complete)

**Ready for implementation, lower priority:**

### Renovate SHA256 Automation 🟢 DESIGNED

| Plan | Status | Priority | Effort |
|------|--------|----------|--------|
| [plans/2026-02-09-renovate-sha256-automation.md](plans/2026-02-09-renovate-sha256-automation.md) | 🟢 Design Complete | 🟢 Low | ~4 hours |

**What:** Fully automatic package updates (version + SHA256)  
**Currently:** Renovate updates versions, GitHub Actions updates SHA256  
**Benefit:** Save ~50 minutes/month of manual work

### Offline Testing for Copilot 🟢 DESIGNED

| Plan | Status | Priority | Effort |
|------|--------|----------|--------|
| [plans/2026-02-09-offline-testing-for-copilot.md](plans/2026-02-09-offline-testing-for-copilot.md) | 🟢 Design Complete | 🟢 Low | ~6 hours |

**What:** Pre-cache package metadata for Copilot to use without network  
**Benefit:** Better desktop integration, fewer incomplete casks  
**Note:** Nice to have, not critical

---

## Observations & Analysis

**Learning documents from incidents and experiments:**

| Document | Purpose | Date |
|----------|---------|------|
| [observations/2026-02-09-copilot-pr-22-monitoring.md](observations/2026-02-09-copilot-pr-22-monitoring.md) | Monitoring Copilot PR #22 (Rancher Desktop) | 2026-02-09 |
| [observations/pr-11-failure-analysis.md](observations/pr-11-failure-analysis.md) | Root cause analysis of Rancher Desktop PR | 2026-02-09 |
| [observations/copilot-pr-11-monitoring.md](observations/copilot-pr-11-monitoring.md) | Monitoring Copilot's fix attempt | 2026-02-09 |
| [observations/2026-02-09-brainstorming-session.md](observations/2026-02-09-brainstorming-session.md) | Skills evaluation brainstorm | 2026-02-09 |

---

## Brainstorms

**Exploratory documents (may lead to plans):**

| Document | Purpose | Status |
|----------|---------|--------|
| [brainstorms/2026-02-09-copilot-session-observations.md](brainstorms/2026-02-09-copilot-session-observations.md) | Observations from Copilot testing | Draft |

---

## Summaries

**High-level overviews:**

| Document | Purpose |
|----------|---------|
| [SUMMARY_COPILOT_IMPROVEMENTS.md](SUMMARY_COPILOT_IMPROVEMENTS.md) | Summary of Copilot enhancements |

---

## Historical Plans

**Completed or superseded plans:**

| Document | Status | Note |
|----------|--------|------|
| [plans/2026-02-08-automated-homebrew-tap.md](plans/2026-02-08-automated-homebrew-tap.md) | ✅ Complete | Original automation design |

---

## Priority Matrix

### 🔴 Critical (Do First)

1. **Phase 2: Zero-Error Packages** (2 hours)
   - Update tap-cask/tap-formula to auto-validate
   - Prevents style violations at source
   - Completes defense-in-depth validation

### 🟡 High (Do Soon)

2. **CI Optimization Experiment** (2 hours)
   - Test simple PATH vs full setup
   - Potential 20-30s speedup per run
   - Low risk, high value

### 🟢 Medium (Valuable but Not Urgent)

3. **Renovate SHA256 Automation** (4 hours)
   - Fully automatic updates
   - Saves ~50 min/month

4. **Go Migration Completion** (ongoing)
   - All core tools complete (tap-cask, tap-formula, tap-issue, tap-validate)
   - Legacy bash scripts remain for backward compatibility

### 🔵 Low (Nice to Have)

5. **Offline Testing for Copilot** (6 hours)
   - Better metadata for cask generation
   - Copilot-specific improvement

6. **Phase 3: Smoke Testing** (8-10 hours)
   - After Phase 2 stabilizes
   - Tests actual package installation

---

## Document Status Legend

| Icon | Meaning |
|------|---------|
| ✅ | Complete and stable |
| ⏳ | In progress |
| 📋 | Ready to implement |
| 🟢 | Design complete, awaiting implementation |
| 📅 | Future work (after dependencies) |
| ⚠️ | Needs updating |

---

## Getting Started Paths

### Path 1: I want to create a new cask (GUI app)
1. Read [AGENT_BEST_PRACTICES.md](AGENT_BEST_PRACTICES.md) (5 min) - **Prevents CI failures!**
2. Read [CASK_CREATION_GUIDE.md](CASK_CREATION_GUIDE.md) (10 min)
3. Use `./tap-tools/tap-cask generate <github-url>` (2 min)
4. Validate: `./tap-tools/tap-validate file Casks/app-linux.rb --fix` (1 min)
5. Commit and push

### Path 2: I want to create a new formula (CLI tool)
1. Read [AGENT_BEST_PRACTICES.md](AGENT_BEST_PRACTICES.md) (5 min) - **Prevents CI failures!**
2. Read [FORMULA_PATTERNS.md](FORMULA_PATTERNS.md) (5 min)
3. Use `./tap-tools/tap-formula generate <github-url>` (2 min)
4. Validate: `./tap-tools/tap-validate file Formula/tool.rb --fix` (1 min)
5. Commit and push

### Path 3: I want to understand the workflow
1. Read [AGENT_GUIDE.md](AGENT_GUIDE.md) (15 min)
2. Review [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues
3. Check existing casks/formulas as examples

### Path 4: I want to implement a plan
1. Check **Priority Matrix** above
2. Read the specific plan document
3. Follow implementation steps
4. Update plan with progress

---

## Contributing

When adding new documentation:
1. Add entry to this index
2. Follow existing naming conventions
3. Update "Last Updated" date above
4. Keep status indicators current

---

## Questions?

- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) first
- Review relevant plan document
- Check existing casks/formulas for examples
