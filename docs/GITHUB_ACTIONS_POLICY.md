# GitHub Actions Policy

**Last Updated:** 2026-02-09  
**Status:** ✅ Active Policy

## Overview

This document defines policies for GitHub Actions usage in this repository to ensure security, stability, and maintainability.

---

## Policy 1: SHA Pinning (MANDATORY)

### Rule

**ALL GitHub Actions MUST be pinned to full commit SHAs**, not tags or branch names.

### Rationale

**Security:**
- Tags and branches can be moved to point at malicious code
- SHA pins are immutable and cannot be changed after commit
- Prevents supply chain attacks via compromised actions

**Stability:**
- Prevents unexpected breaking changes from action updates
- Ensures reproducible builds
- CI behavior is deterministic

**Automation:**
- Renovate can automatically update SHA pins
- Renovate includes tag information in PR descriptions
- We get security updates without manual tracking

### Examples

**❌ WRONG - Tag pinning:**
```yaml
- uses: actions/checkout@v4
- uses: tj-actions/changed-files@v46
```

**❌ WRONG - Branch pinning:**
```yaml
- uses: Homebrew/actions/setup-homebrew@master
```

**✅ CORRECT - SHA pinning with comment:**
```yaml
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
- uses: tj-actions/changed-files@20576b4b9ed46d41e2d45a2256e5e2316dde6834 # v46.0.0
- uses: Homebrew/actions/setup-homebrew@1ccc07ccd8045a63f6a5c51c1b51f4e3a1d072a9 # master
```

**Format:** `<owner>/<repo>@<full-sha> # <tag-or-version>`

### Implementation

**1. Update all existing workflows:**
```bash
# Find current usage
grep -r "uses:" .github/workflows/

# For each action, find the SHA for the current tag/branch
# Example for actions/checkout@v4:
git ls-remote https://github.com/actions/checkout v4
# Copy the SHA and update workflow
```

**2. Add Renovate configuration:**

See `.github/renovate.json5` - GitHub Actions updates are enabled.

**3. Verify pinning in CI:**

Add check to prevent unpinned actions from being merged (future):
```yaml
# .github/workflows/validate-actions.yml
name: Validate Actions
on: [pull_request]
jobs:
  check-pinning:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<sha>
      - name: Check all actions are SHA pinned
        run: |
          # Check for non-SHA references
          if grep -r "uses:.*@v[0-9]" .github/workflows/; then
            echo "ERROR: Found tag-pinned actions"
            exit 1
          fi
```

### Exceptions

**None.** All actions must be SHA pinned without exception.

If an action doesn't have releases, pin to a specific commit SHA on the main branch and add a comment explaining why.

---

## Policy 2: Action Source Restrictions

### Rule

**Only use GitHub Actions from trusted sources:**

**✅ Allowed sources:**
1. **GitHub Official** - `actions/*` (actions/checkout, actions/cache, etc.)
2. **Verified Publishers** - Actions with blue checkmark on Marketplace
3. **Well-known projects** - Homebrew, major open source projects (case-by-case)
4. **Our own actions** - Actions in this repository

**⚠️ Requires justification:**
- Actions from individual developers (not organizations)
- Actions with < 100 stars on GitHub
- Actions without recent maintenance activity

**❌ Not allowed:**
- Obfuscated or minified action code
- Actions without source code available
- Actions from unknown/suspicious sources

### Verification Process

Before adding a new action:
1. Check the action's source repository
2. Review recent commits and contributors
3. Check for security issues or CVEs
4. Verify the action is actively maintained
5. Document the decision in PR description

---

## Policy 3: Automated Updates via Renovate

### Rule

**Renovate automatically updates all GitHub Actions** by creating PRs with SHA updates.

### Configuration

See `.github/renovate.json5`:

```json5
{
  "github-actions": {
    "enabled": true,
    "pinDigests": true
  },
  "packageRules": [
    {
      "matchManagers": ["github-actions"],
      "matchUpdateTypes": ["patch", "minor"],
      "automerge": true,
      "minimumReleaseAge": "3 days"
    },
    {
      "matchManagers": ["github-actions"],
      "matchUpdateTypes": ["major"],
      "automerge": false,
      "labels": ["major-update", "needs-review"]
    }
  ]
}
```

### Update Strategy

**Patch/Minor updates:**
- ✅ Auto-merge after 3 days
- CI must pass
- No manual review required

**Major updates:**
- ⚠️ Requires manual review
- May include breaking changes
- Check release notes before merging

### Monitoring

- Review Renovate PRs regularly
- Check GitHub Security Advisories
- Subscribe to action repository releases (optional)

---

## Policy 4: Workflow Permissions

### Rule

**Use least-privilege permissions** for all workflows.

### Default Permissions

Set at repository level (Settings → Actions → Workflow permissions):
- ✅ Read repository contents
- ❌ Write access (must be explicitly granted per job)

### Per-Job Permissions

Always specify minimal permissions needed:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: read        # Read repo contents
      pull-requests: write  # Comment on PRs (if needed)
    steps:
      # ...
```

### Common Permission Sets

**Read-only (testing, validation):**
```yaml
permissions:
  contents: read
```

**PR commenting:**
```yaml
permissions:
  contents: read
  pull-requests: write
```

**Automated commits (Renovate, SHA256 updates):**
```yaml
permissions:
  contents: write
  pull-requests: write
```

---

## Policy 5: Secret Management

### Rule

**Never hardcode secrets** in workflows. Always use GitHub Secrets.

### Usage

```yaml
steps:
  - name: Use secret
    env:
      API_TOKEN: ${{ secrets.API_TOKEN }}
    run: |
      # Use $API_TOKEN environment variable
```

### Available Secrets

Document required secrets in workflow comments:

```yaml
# Required secrets:
#   - GITHUB_TOKEN (automatically provided)
#   - API_TOKEN (manual setup required)
```

---

## Policy 6: Workflow Documentation

### Rule

**All workflows must be documented** with clear comments.

### Required Documentation

```yaml
name: Descriptive Name

# Purpose: What this workflow does
# Triggers: When it runs
# Permissions: What access it needs
# Secrets: What secrets it requires

on:
  push:
    branches: [main]

jobs:
  job-name:
    # Description of what this job does
    runs-on: ubuntu-latest
    steps:
      - name: Clear step description
        # Comment if the step does something non-obvious
        run: |
          echo "Commands"
```

---

## Migration Plan: Current → SHA Pinned

### Current State (2026-02-09)

**Unpinned actions in use:**
```yaml
actions/checkout@v4                          # tests.yml, cask-sha256-update.yml
actions/github-script@v7                     # cask-sha256-update.yml
Homebrew/actions/setup-homebrew@master       # tests.yml
ruby/setup-ruby@v1                           # cask-sha256-update.yml
tj-actions/changed-files@v46                 # tests.yml
```

### Migration Steps

#### Step 1: Get Current SHAs

```bash
# actions/checkout@v4
gh api repos/actions/checkout/git/ref/tags/v4.1.1 --jq '.object.sha'
# → b4ffde65f46336ab88eb53be808477a3936bae11

# actions/github-script@v7
gh api repos/actions/github-script/git/ref/tags/v7.0.1 --jq '.object.sha'
# → 60a0d83039c74a4aee543508d2ffcb1c3799cdea

# Homebrew/actions/setup-homebrew@master
gh api repos/Homebrew/actions/git/ref/heads/master --jq '.object.sha'
# → (get latest commit SHA)

# ruby/setup-ruby@v1
gh api repos/ruby/setup-ruby/git/ref/tags/v1.195.0 --jq '.object.sha'
# → 8a8f61f0001d09fd8fb7f12301c061e29b182ce6

# tj-actions/changed-files@v46
gh api repos/tj-actions/changed-files/git/ref/tags/v46.0.2 --jq '.object.sha'
# → 20576b4b9ed46d41e2d45a2256e5e2316dde6834
```

#### Step 2: Update Workflows

Update all workflow files to use SHA pins:

**tests.yml:**
```yaml
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
- uses: Homebrew/actions/setup-homebrew@<sha> # master
- uses: tj-actions/changed-files@20576b4b9ed46d41e2d45a2256e5e2316dde6834 # v46.0.2
```

**cask-sha256-update.yml:**
```yaml
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
- uses: ruby/setup-ruby@8a8f61f0001d09fd8fb7f12301c061e29b182ce6 # v1.195.0
- uses: actions/github-script@60a0d83039c74a4aee543508d2ffcb1c3799cdea # v7.0.1
```

#### Step 3: Update Renovate Configuration

Add GitHub Actions manager to `.github/renovate.json5`:

```json5
{
  // Enable GitHub Actions updates with SHA pinning
  "github-actions": {
    "enabled": true,
    "pinDigests": true
  },
  
  "packageRules": [
    // ... existing rules ...
    
    {
      // Auto-merge GitHub Actions patch/minor updates
      "matchManagers": ["github-actions"],
      "matchUpdateTypes": ["patch", "minor"],
      "automerge": true,
      "automergeType": "pr",
      "automergeStrategy": "squash",
      "minimumReleaseAge": "3 days",
      "schedule": ["at any time"]
    },
    {
      // Manual review for major GitHub Actions updates
      "matchManagers": ["github-actions"],
      "matchUpdateTypes": ["major"],
      "automerge": false,
      "labels": ["major-update", "needs-review"]
    },
    {
      // Group all GitHub Actions updates together
      "matchManagers": ["github-actions"],
      "groupName": "GitHub Actions",
      "groupSlug": "github-actions",
      "commitMessageTopic": "GitHub Actions"
    }
  ]
}
```

#### Step 4: Document and Commit

```bash
git add .github/workflows/*.yml .github/renovate.json5 docs/GITHUB_ACTIONS_POLICY.md
git commit -m "security(ci): pin all GitHub Actions to commit SHAs

Implements SHA pinning policy for all GitHub Actions:
- Pins actions/checkout@v4 → SHA
- Pins tj-actions/changed-files@v46 → SHA
- Pins ruby/setup-ruby@v1 → SHA
- Pins actions/github-script@v7 → SHA
- Pins Homebrew/actions/setup-homebrew@master → SHA

Benefits:
- Prevents supply chain attacks (immutable references)
- Ensures reproducible builds
- Enables automated updates via Renovate

Renovate will automatically create PRs to update SHAs when new
versions are released.

See docs/GITHUB_ACTIONS_POLICY.md for complete policy.

Assisted-by: Claude 3.5 Sonnet via OpenCode"
```

---

## Testing

### Verify SHA Pinning

```bash
# Check for any non-SHA references
grep -rn "uses:.*@v[0-9]" .github/workflows/
grep -rn "uses:.*@main" .github/workflows/
grep -rn "uses:.*@master" .github/workflows/

# Should return no results (except in comments/docs)
```

### Test Renovate

After merging:
1. Wait for Renovate to scan repository
2. Check for "Update GitHub Actions" PRs
3. Verify PRs include both old and new SHAs
4. Confirm auto-merge works for patch/minor

---

## Maintenance

### Regular Review

**Quarterly (every 3 months):**
- Review all GitHub Actions in use
- Check for deprecated actions
- Verify Renovate is creating update PRs
- Review any stuck/failed updates

### Security Incidents

If an action has a security vulnerability:
1. Check if we're affected (version/SHA)
2. Update immediately if vulnerable
3. Document incident in PR
4. Consider alternative actions

---

## References

- [GitHub Actions Security Best Practices](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)
- [Renovate GitHub Actions Manager](https://docs.renovatebot.com/modules/manager/github-actions/)
- [Pinning Actions to SHA](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions)

---

## Status

- [ ] Policy documented ✅
- [ ] Current workflows audited
- [ ] SHAs collected
- [ ] Workflows updated
- [ ] Renovate configured
- [ ] Changes committed
- [ ] Renovate tested

**Target completion:** Immediate (next session)
