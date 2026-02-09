# Infrastructure Documentation

This directory contains documentation about repository infrastructure, tooling, and operational issues that affect AI agents and developers.

## Documents

### [GITHUB_TOKEN_SOLUTION_PLAN.md](GITHUB_TOKEN_SOLUTION_PLAN.md)

**Issue:** AI agents hit GitHub API rate limits when using tap-tools  
**Status:** Planning - Not Yet Implemented  
**Priority:** High (blocks agent automation)

**Problem:**
- Agents don't have `GITHUB_TOKEN` set in their execution environment
- Results in 60 requests/hour limit (unauthenticated) vs 5,000+ with token
- Blocks usage of `tap-issue` command (requires token)
- Causes delays/failures in package generation

**Solution:**
- Verify token availability in GitHub Actions environment
- Document token requirements and setup in AGENTS.md
- Enhance tap-tools with better error messages and rate limit monitoring
- Consider gh CLI fallback for local development

**Impact:**
- Affects: Copilot, GitHub Actions workflows, all agent-based automation
- Evidence: PR #22 (Copilot) encountered rate limiting during Issue #21

---

## Quick Reference

### For AI Agents

**Before using tap-tools, verify token:**
```bash
echo $GITHUB_TOKEN
gh api rate_limit
```

**If missing:**
```bash
# GitHub Actions: Should be automatic, check workflow permissions
# Local: export GITHUB_TOKEN=$(gh auth token)
```

### For Repository Maintainers

**Check rate limit status:**
```bash
gh api rate_limit
```

**Test token in workflow:**
```yaml
- name: Check GITHUB_TOKEN
  run: |
    if [ -z "$GITHUB_TOKEN" ]; then
      echo "❌ GITHUB_TOKEN not set"
      exit 1
    fi
    gh api rate_limit
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### For Local Development

**Setup token once:**
```bash
# Using gh CLI (recommended)
export GITHUB_TOKEN=$(gh auth token)

# Or create personal access token
# 1. Visit: https://github.com/settings/tokens
# 2. Create token with 'repo' scope
# 3. export GITHUB_TOKEN=ghp_your_token_here
```

---

## Related Documentation

- [AGENTS.md](../../AGENTS.md) - Main agent instructions
- [tap-tools/README.md](../../tap-tools/README.md) - tap-tools documentation
- [AGENT_BEST_PRACTICES.md](../AGENT_BEST_PRACTICES.md) - Common errors and prevention

## Issues & PRs

- **PR #22** - Copilot encountered rate limiting (Issue #21 - Rancher Desktop)
- **Run 21831659602** - First documented instance of token issue

## Future Improvements

Planned improvements tracked in GITHUB_TOKEN_SOLUTION_PLAN.md:

1. **Phase 1:** Verify current state (token availability in Copilot/Actions)
2. **Phase 2:** Documentation (AGENTS.md, tap-tools/README.md updates)
3. **Phase 3:** Code improvements (error messages, rate limit monitoring)
4. **Phase 4:** Repository configuration (workflow permissions)
5. **Phase 5:** Monitoring & validation

---

**Last Updated:** 2026-02-09  
**Maintainer:** Repository team  
**Status:** Active - tracking infrastructure issues
