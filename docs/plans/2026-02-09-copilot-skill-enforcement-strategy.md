# GitHub Copilot Skill Enforcement Strategy

**Date:** 2026-02-09  
**Status:** Planning  
**Problem:** GitHub Copilot is not loading/using the homebrew-packaging skill when assigned package creation tasks

---

## Problem Statement

We've confirmed that GitHub Copilot is **not using** the `.github/skills/homebrew-packaging/SKILL.md` skill when it should be. This skill contains critical workflows for:
- Using tap-tools for package generation
- **MANDATORY validation before commits**
- XDG compliance requirements
- Linux-only constraints

**Impact:** Copilot creates packages that fail CI due to style violations, missing validation, and non-compliance with tap requirements.

---

## Root Cause Analysis

### Why Copilot Isn't Loading the Skill

Based on GitHub's Agent Skills documentation and our repository structure:

#### 1. **Skill Description May Not Trigger Matching**

**Current description:**
```yaml
description: >
  Complete workflow for creating and updating Homebrew packages (casks and 
  formulas) for Linux-only tap targeting read-only filesystem systems. Use 
  this when user requests adding, updating, or fixing packages from GitHub 
  releases.
```

**Problem:** Copilot's skill matching algorithm looks at:
- User prompt: "Add Quarto" or "Package Rancher Desktop"
- Skill description: "...when user requests adding, updating, or fixing packages..."

The matching may not be sensitive enough to trigger on implicit requests like:
- "Add X" (doesn't explicitly say "package")
- "Create a cask for Y" (might not match "adding packages")
- Issue assignment without explicit keywords

#### 2. **Passive Loading Mechanism**

Per GitHub docs:
> "When performing tasks, Copilot will **decide when to use your skills** based on your prompt and the skill's description."

**Key insight:** Skills are **opt-in** and **suggestion-based**, not mandatory. Copilot uses heuristics to decide whether to load a skill, and it can choose **not to load it**.

**Contrast with OpenCode superpowers:**
- OpenCode: Skills are MANDATORY when triggered (enforced by system)
- Copilot: Skills are OPTIONAL suggestions (Copilot decides)

#### 3. **No Explicit Loading Directive**

The current `.github/copilot-instructions.md` (which we now know is actually `AGENTS.md`) doesn't explicitly tell Copilot:
> "BEFORE doing ANY package work, you MUST load and follow the homebrew-packaging skill."

#### 4. **Competing Instructions Create Confusion**

`AGENTS.md` contains comprehensive packaging instructions (573 lines). Copilot might think:
- "I already have detailed instructions in AGENTS.md"
- "I don't need to load the skill"
- "The custom instructions are sufficient"

---

## Enforcement Strategies (Ranked by Effectiveness)

### Strategy 1: **Make Skill Loading Explicit in AGENTS.md** ⭐ **HIGHEST IMPACT**

**Approach:** Add a prominent directive at the top of `AGENTS.md` that **forces Copilot to load the skill**.

**Implementation:**

```markdown
# Agent Instructions

⚠️ **CRITICAL: SKILL LOADING REQUIREMENT** ⚠️

**BEFORE doing ANY package-related work (creating, updating, debugging), you MUST:**
1. Load the `homebrew-packaging` skill from `.github/skills/homebrew-packaging/SKILL.md`
2. Follow the skill's step-by-step workflow exactly
3. Complete ALL mandatory checkpoints (especially validation before commits)

**To load the skill:** Check `.github/skills/homebrew-packaging/SKILL.md` at the start of EVERY packaging session.

**This skill overrides any conflicting guidance in this file.**

---

**⚠️ LINUX ONLY REPOSITORY ⚠️**
...
```

**Why this works:**
- ✅ Explicit directive in always-loaded custom instructions
- ✅ Forces skill loading before any work begins
- ✅ Makes skill the source of truth
- ✅ No code changes required

**Limitations:**
- ❌ Still relies on Copilot following instructions (not 100% enforceable)
- ✅ But makes non-compliance obvious (violates explicit directive)

---

### Strategy 2: **Broaden Skill Description to Catch More Triggers** ⭐⭐ **MEDIUM IMPACT**

**Approach:** Rewrite the skill's `description` field to match a wider variety of user prompts.

**Current:**
```yaml
description: >
  Complete workflow for creating and updating Homebrew packages (casks and 
  formulas) for Linux-only tap targeting read-only filesystem systems. Use 
  this when user requests adding, updating, or fixing packages from GitHub 
  releases.
```

**Improved:**
```yaml
description: >
  MANDATORY workflow for ALL Homebrew work: creating casks, creating formulas,
  adding packages, updating versions, fixing packages, debugging build failures,
  processing GitHub issues, or ANY task involving Casks/ or Formula/ directories.
  Use this WHENEVER the user mentions: package, cask, formula, brew, tap-tools,
  tap-cask, tap-formula, install, or references GitHub release URLs.
```

**Why this works:**
- ✅ Increases trigger keywords dramatically
- ✅ Uses imperative language ("MANDATORY", "ALL", "ANY")
- ✅ Covers edge cases (issue processing, debugging)
- ✅ Mentions directory names (Casks/, Formula/)
- ✅ Lists tool names (tap-cask, tap-formula)

**Limitations:**
- ❌ Still relies on Copilot's matching algorithm
- ❌ GitHub doesn't document what makes a "good" description

---

### Strategy 3: **Create Multiple Specialized Skills** ⭐ **LOW-MEDIUM IMPACT**

**Approach:** Break the single skill into multiple, hyper-specific skills.

**Rationale:** GitHub docs suggest skills should be "specialized tasks". Our current skill is 573 lines and covers multiple workflows.

**Proposed split:**

```
.github/skills/
├── homebrew-create-cask/           # New cask from scratch
│   └── SKILL.md
├── homebrew-create-formula/        # New formula from scratch
│   └── SKILL.md
├── homebrew-update-package/        # Version bumps, SHA256 updates
│   └── SKILL.md
├── homebrew-validate-package/      # Validation-only workflow
│   └── SKILL.md
└── homebrew-debug-ci-failure/      # Fixing failed CI checks
    └── SKILL.md
```

**Each skill has a laser-focused description:**

```yaml
# homebrew-create-cask/SKILL.md
---
name: homebrew-create-cask
description: >
  Create a new Homebrew cask for a Linux GUI application. Use this when the user
  says "add", "create", "package", or provides a GitHub URL ending in /releases
  and mentions "GUI", "app", "application", or ".desktop". ALWAYS load this for
  new cask creation.
---
```

**Why this works:**
- ✅ Each skill has narrow, specific trigger conditions
- ✅ Higher chance of matching user intent
- ✅ Easier for Copilot to decide which skill to load
- ✅ Follows GitHub's "specialized tasks" recommendation

**Limitations:**
- ❌ Maintenance overhead (5 files vs 1)
- ❌ Duplication of common content (XDG rules, naming, etc.)
- ❌ May not fix root issue if Copilot still doesn't load skills

---

### Strategy 4: **Add Skill Invocation to Pre-commit Hook** ⚠️ **TECHNICAL CONTROL**

**Approach:** Make the pre-commit hook explicitly reference the skill.

**Implementation:**

```bash
#!/usr/bin/env bash
# .git/hooks/pre-commit

# ... existing validation ...

echo "ℹ️  Reminder: This validation enforces .github/skills/homebrew-packaging/SKILL.md"
echo "   If this commit was created by an AI agent, ensure it followed the skill workflow."
echo ""

# ... rest of hook ...
```

**Why this works:**
- ✅ Creates visibility into skill compliance
- ✅ Reminds human reviewers to check for skill usage
- ✅ Technical control (hook WILL run)

**Limitations:**
- ❌ Doesn't force Copilot to load the skill
- ✅ But provides feedback loop for detecting non-compliance

---

### Strategy 5: **Create a "Skill Loader" Custom Agent** 💡 **ADVANCED**

**Approach:** Create a custom Copilot agent whose sole job is to load and enforce skills.

**Implementation:**

```markdown
# .github/copilot-agents/skill-loader.yml
name: skill-loader
description: Ensures proper skill loading for all packaging tasks
triggers:
  - "package"
  - "cask"
  - "formula"
  - "tap-tools"
  - "add"
  - "create"
  
instructions: |
  Your ONLY job is to:
  1. Detect if this task is packaging-related
  2. If yes, load .github/skills/homebrew-packaging/SKILL.md
  3. Pass control to the loaded skill
  4. If the task can be completed without the skill, proceed normally
  
  You are a gatekeeper. Always check for skill relevance first.
```

**Why this works:**
- ✅ Dedicated agent for skill loading
- ✅ Can be triggered independently
- ✅ Enforces skill-first workflow

**Limitations:**
- ❌ Requires GitHub Copilot Enterprise (custom agents)
- ❌ Not available in Free/Pro/Team plans
- ❌ Adds complexity

---

## Recommended Implementation Plan

### Phase 1: Immediate (Zero Code Changes) ⭐ **DO THIS NOW**

1. **Update `AGENTS.md`** with explicit skill loading directive (Strategy 1)
2. **Improve skill description** with more trigger keywords (Strategy 2)
3. **Test** by assigning Copilot a new package request

**Effort:** 15 minutes  
**Expected improvement:** 60-80% (Copilot will see explicit directive)

### Phase 2: Short-term (If Phase 1 Insufficient)

1. **Split skill into specialized skills** (Strategy 3)
2. **Each skill** gets laser-focused description
3. **Test** with multiple task types (new cask, version update, CI fix)

**Effort:** 2-3 hours  
**Expected improvement:** 80-90% (more precise skill matching)

### Phase 3: Medium-term (Technical Controls)

1. **Enhance pre-commit hook** with skill reference (Strategy 4)
2. **Add CI check** that validates commits reference the skill in messages
3. **Create dashboard** tracking skill usage (audit logs)

**Effort:** 4-5 hours  
**Expected improvement:** 95% (technical enforcement + visibility)

### Phase 4: Long-term (If Copilot Enterprise Available)

1. **Create skill-loader custom agent** (Strategy 5)
2. **Integrate with GitHub Issues** to auto-assign skill-loader
3. **Monitor effectiveness** via Copilot usage metrics

**Effort:** 8-10 hours  
**Expected improvement:** 99% (dedicated enforcement agent)

---

## Validation Criteria

After implementing each phase, measure success with:

### Quantitative Metrics
- **% of Copilot PRs that pass CI on first try** (Baseline: ~50%, Target: >90%)
- **% of Copilot commits with validation failures** (Baseline: ~30%, Target: <5%)
- **Average time to merge Copilot PRs** (Baseline: 2-3 rounds, Target: 1 round)

### Qualitative Indicators
- ✅ Copilot explicitly mentions the skill in PR descriptions
- ✅ Copilot follows the 6-step workflow (generate → validate → review → test → commit → PR)
- ✅ Copilot runs `tap-validate --fix` before every commit
- ✅ Copilot creates PRs with proper XDG environment variables
- ✅ Copilot uses tap-tools instead of manual cask creation

### Test Scenarios
Create test issues to verify skill loading:

1. **Test 1:** "Add package X" (generic request)
2. **Test 2:** "Create a cask for https://github.com/foo/bar/releases" (explicit)
3. **Test 3:** "Fix the rancher-desktop cask" (update existing)
4. **Test 4:** "Debug why the CI is failing on PR #X" (troubleshooting)
5. **Test 5:** Assign issue with no explicit instructions (implicit)

**Success:** Copilot loads the skill in 5/5 scenarios

---

## Example: Updated AGENTS.md Header

```markdown
# Agent Instructions

⚠️ **MANDATORY SKILL LOADING REQUIREMENT** ⚠️

**CRITICAL: For ANY package-related work, you MUST load the skill FIRST.**

Before creating, updating, or fixing any Homebrew packages (casks or formulas):

1. **Load the skill:** Read `.github/skills/homebrew-packaging/SKILL.md`
2. **Follow it exactly:** The skill contains the mandatory 6-step workflow
3. **Never skip validation:** Step 2 (validation) is MANDATORY before commits

**When to load the skill:**
- ✅ User mentions: "add", "create", "package", "cask", "formula", "update", "fix"
- ✅ User provides GitHub release URL
- ✅ User references Casks/ or Formula/ directories
- ✅ User asks about tap-tools, tap-cask, or tap-formula
- ✅ User assigns you a GitHub issue about packages
- ✅ ANY task involving Homebrew packages

**The skill is the source of truth. If AGENTS.md conflicts with the skill, follow the skill.**

---

**⚠️ LINUX ONLY REPOSITORY ⚠️**

**THIS TAP IS LINUX-ONLY. ALL PACKAGES MUST USE LINUX BINARIES.**
...
```

---

## Example: Improved Skill Description

```yaml
---
name: homebrew-packaging
description: >
  MANDATORY workflow for ALL Homebrew work in this repository. Use this skill
  whenever the user requests: creating packages, adding casks, creating formulas,
  updating versions, fixing packages, debugging CI failures, processing GitHub
  issues about packages, or ANY work involving Casks/ or Formula/ directories.
  Also use when user mentions: tap-tools, tap-cask, tap-formula, brew install,
  package from GitHub releases, XDG compliance, or Linux binaries. This skill
  contains critical constraints (Linux-only, read-only filesystem, XDG paths),
  mandatory validation requirements, and the 6-step workflow that MUST be followed
  for all packaging work. Load this skill BEFORE starting any packaging task.
license: MIT
---
```

---

## Comparison: Skill vs Custom Instructions

| Aspect | Custom Instructions (AGENTS.md) | Skills (.github/skills/) |
|--------|--------------------------------|--------------------------|
| **Loading** | Always loaded (every session) | Loaded when relevant (heuristic) |
| **Size** | Short (context constraints) | Long (loaded on-demand) |
| **Purpose** | Repository overview, general guidance | Detailed step-by-step workflows |
| **Enforcement** | Suggestive ("you should...") | Directive ("you MUST...") |
| **Best for** | Critical constraints, quick reference | Specialized procedures, checklists |

**Current problem:** We're using custom instructions (AGENTS.md) like a skill, but Copilot treats it as general guidance, not mandatory workflow.

**Solution:** Use **both**:
- **AGENTS.md:** "ALWAYS load the skill for packaging work"
- **SKILL.md:** "Here's the exact workflow to follow"

---

## Expected Outcomes

### After Phase 1 (Explicit Directive)
- **Before:** Copilot creates packages without loading skill → CI fails
- **After:** Copilot sees "MUST load skill" → loads skill → follows workflow → CI passes

### After Phase 2 (Specialized Skills)
- **Before:** Single broad skill → vague matching → inconsistent loading
- **After:** Multiple focused skills → precise matching → reliable loading

### After Phase 3 (Technical Controls)
- **Before:** No visibility into skill compliance → hard to debug
- **After:** Pre-commit hook reminds about skill → audit trail → easier debugging

### After Phase 4 (Custom Agent)
- **Before:** Copilot decides whether to load skill
- **After:** skill-loader agent enforces loading → guaranteed compliance

---

## Fallback: If Copilot Still Doesn't Use the Skill

If all strategies fail to make Copilot load the skill, the issue is likely:

### Hypothesis A: GitHub's Skill Implementation Is Immature
**Evidence:**
- Agent Skills are marked as "public preview" in some contexts
- Documentation is sparse on skill matching algorithms
- No debug mode to see why skills weren't loaded

**Mitigation:**
- File GitHub Support ticket asking for skill loading diagnostics
- Request feature: "Explain why skill wasn't loaded for this task"
- Monitor GitHub's Agent Skills changelog for improvements

### Hypothesis B: Skills Work Differently Than Expected
**Evidence:**
- Maybe skills are meant for "specialized tools" not "workflow enforcement"
- Maybe AGENTS.md is the right place for workflows after all
- Maybe we're using skills incorrectly

**Mitigation:**
- Study example skills from anthropics/skills repository
- Compare our skill structure to community examples
- Ask in GitHub Community Discussions: "How to enforce skill usage?"

### Hypothesis C: Copilot Prefers Custom Instructions Over Skills
**Evidence:**
- AGENTS.md (custom instructions) always loaded
- Skills loaded conditionally
- Maybe Copilot prioritizes always-available instructions

**Mitigation:**
- **Consolidate everything into AGENTS.md** (abandon skills)
- Use imperative language throughout ("MUST", "NEVER", "ALWAYS")
- Add explicit checkpoints and validation mandates
- Accept that skills may not be the right tool for our use case

---

## Success Definition

**Primary goal:** Zero CI failures due to Copilot not following packaging requirements.

**Success = All 3 criteria met:**

1. ✅ **Copilot mentions the skill in every packaging PR**  
   Example: "I followed the homebrew-packaging skill workflow..."

2. ✅ **Copilot runs `tap-validate --fix` before every commit**  
   Evidence: Commit timestamps show validation before commit

3. ✅ **95%+ of Copilot packaging PRs pass CI on first attempt**  
   Metric: Track via GitHub Actions success rate

---

## Timeline

| Phase | Tasks | Effort | Timeline |
|-------|-------|--------|----------|
| Phase 1 | Update AGENTS.md, improve skill description | 15 min | Today (2026-02-09) |
| Test Period | Assign 3 test issues to Copilot | 2 hours | Next 2 days |
| Phase 2 | Split into specialized skills (if needed) | 3 hours | Next week |
| Phase 3 | Add technical controls (if needed) | 5 hours | Next 2 weeks |
| Phase 4 | Custom agent (if Enterprise available) | 10 hours | Next month |

---

## Open Questions

1. **Does Copilot have a debug mode that shows skill loading decisions?**
   - No public documentation found
   - May need to contact GitHub Support

2. **What makes a skill description "good" for matching?**
   - GitHub docs don't specify
   - May need experimentation

3. **Can we force skill loading programmatically?**
   - Not via documented APIs
   - Custom agents might be the only way (Enterprise only)

4. **Should we abandon skills and use only AGENTS.md?**
   - Valid option if skills prove unreliable
   - Trade-off: Shorter, always-loaded instructions vs. detailed, conditional loading

5. **Can we track skill usage in Copilot audit logs?**
   - Need to investigate GitHub's agentic audit log events
   - May provide insights into skill loading behavior

---

## Related Documentation

- **GitHub Docs:** [About Agent Skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills)
- **GitHub Docs:** [Repository Custom Instructions](https://docs.github.com/en/copilot/how-tos/configure-custom-instructions/add-repository-instructions)
- **Community:** [anthropics/skills repository](https://github.com/anthropics/skills)
- **Community:** [github/awesome-copilot collection](https://github.com/github/awesome-copilot)
- **Our Docs:** `.github/skills/homebrew-packaging/SKILL.md`
- **Our Docs:** `AGENTS.md`
- **Our Docs:** `docs/brainstorms/2026-02-09-brainstorming-session.md`

---

**Status:** Ready for Phase 1 implementation  
**Next Action:** Update AGENTS.md with explicit skill loading directive  
**Owner:** Repository maintainer  
**Review Date:** 2026-02-11 (after 3 test issues)
