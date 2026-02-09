# Renovate Guide - Automated Dependency Updates

**Date:** 2026-02-09  
**Purpose:** Explain how Renovate works in this tap and what humans need to do

## What is Renovate?

Renovate is a bot that automatically:
- Monitors GitHub releases for packages in your casks/formulas
- Creates PRs when new versions are released
- Updates version numbers in the cask/formula files
- Runs every 3 hours (configured in `.github/renovate.json5`)

## How Renovate Works with This Tap

### Current Renovate Configuration

The regex matcher in `.github/renovate.json5` looks for this pattern in casks:

```ruby
# Renovate scans for this exact pattern:
version "1.8.27"                    # ← Extracts current version
sha256 "bdf689b558..."              # ← Extracts current SHA256
url "https://github.com/quarto-dev/quarto-cli/releases/..." # ← Extracts repo name
```

**What Renovate CAN do automatically:**
- ✅ Detect new versions from GitHub releases
- ✅ Update the `version` line
- ✅ Create a PR with the version bump

**What Renovate CANNOT do automatically:**
- ❌ Download the new tarball
- ❌ Calculate the new SHA256 hash
- ❌ Update the `sha256` line
- ❌ Verify the tarball structure hasn't changed

## Example: What a Renovate PR Looks Like

### Scenario: Quarto releases version 1.9.0

**Current cask (v1.8.27):**
```ruby
cask "quarto-linux" do
  version "1.8.27"
  sha256 "bdf689b5589789a1f21d89c3b83d78ed02a97914dd702e617294f2cc1ea7387d"

  url "https://github.com/quarto-dev/quarto-cli/releases/download/v#{version}/quarto-#{version}-linux-amd64.tar.gz"
  name "Quarto"
  desc "Open-source scientific and technical publishing system built on Pandoc"
  homepage "https://quarto.org/"

  binary "quarto-#{version}/bin/quarto"
end
```

**Renovate PR (v1.9.0) - INCOMPLETE:**
```ruby
cask "quarto-linux" do
  version "1.9.0"                    # ✅ Renovate updates this
  sha256 "bdf689b5589789..."         # ❌ WRONG - Still has old hash!

  url "https://github.com/quarto-dev/quarto-cli/releases/download/v#{version}/quarto-#{version}-linux-amd64.tar.gz"
  name "Quarto"
  desc "Open-source scientific and technical publishing system built on Pandoc"
  homepage "https://quarto.org/"

  binary "quarto-#{version}/bin/quarto"
end
```

**Problem:** The SHA256 is still for version 1.8.27, but the URL now points to 1.9.0. This will fail `brew install`!

## What Humans Need to Do

When Renovate creates a PR, humans must:

### Step 1: Review the Renovate PR

```bash
# View the PR
gh pr view <PR-NUMBER>

# Check what changed
gh pr diff <PR-NUMBER>
```

**You'll see:**
- Version bumped (e.g., `1.8.27` → `1.9.0`)
- SHA256 is WRONG (still has old hash)

### Step 2: Download and Calculate New SHA256

```bash
# Download the new version
curl -LO https://github.com/quarto-dev/quarto-cli/releases/download/v1.9.0/quarto-1.9.0-linux-amd64.tar.gz

# Calculate SHA256
sha256sum quarto-1.9.0-linux-amd64.tar.gz
# Output: abc123def456... quarto-1.9.0-linux-amd64.tar.gz

# Optional: Check upstream checksums if available
curl -sL https://github.com/quarto-dev/quarto-cli/releases/download/v1.9.0/SHA256SUMS | grep linux-amd64
```

### Step 3: Update the SHA256 in the PR

**Option A: Edit directly on GitHub**
1. Go to the PR
2. Click "Files changed"
3. Click the "..." menu on the file → "Edit file"
4. Update the SHA256 line
5. Commit directly to the PR branch

**Option B: Edit locally and push**
```bash
# Check out the PR branch
gh pr checkout <PR-NUMBER>

# Edit the cask
nano Casks/quarto-linux.rb
# Update sha256 line with new hash

# Commit and push
git add Casks/quarto-linux.rb
git commit -m "fix(cask): update SHA256 for quarto-linux v1.9.0"
git push
```

### Step 4: Verify Tarball Structure (Important!)

Check if the tarball structure changed:

```bash
# Extract and inspect
tar -tzf quarto-1.9.0-linux-amd64.tar.gz | head -20

# Look for these issues:
# - Did the root directory name change? (quarto-1.8.27 → quarto-1.9.0 is OK)
# - Did binary paths move? (e.g., bin/quarto → usr/bin/quarto)
# - Did desktop files move or get renamed?
# - Are there new icons or desktop files?
```

**If structure changed:**
```ruby
# Example: Binary moved from bin/ to usr/bin/
binary "quarto-#{version}/usr/bin/quarto"  # Update path

# Example: Desktop file path changed
artifact "quarto-#{version}/share/applications/quarto.desktop",  # Update path
```

### Step 5: Wait for CI and Merge

After you update the SHA256:
- CI will run and download the actual tarball
- CI will verify the SHA256 matches
- CI will run `brew audit` and `brew style`
- If CI passes, the PR can be merged

**With metadata caching (proposed):**
- CI will also update the metadata artifact cache
- Next Copilot run will have fresh metadata for this version

## When to Auto-Merge vs Manual Review

**Configured in `.github/renovate.json5`:**

### Auto-Merge (After Human Updates SHA256):
- **Patch releases** (1.8.27 → 1.8.28) - Auto-merge after 3 hours
- **Minor releases** (1.8.27 → 1.9.0) - Auto-merge after 1 day

**Requirements for auto-merge:**
- SHA256 must be updated by human
- CI checks must pass
- Minimum release age passed (security delay)

### Manual Review Required:
- **Major releases** (1.8.27 → 2.0.0) - Labeled `major-update`, `needs-review`
- Likely to have breaking changes or structural changes
- Should inspect tarball structure carefully

## Current Limitations and Workarounds

### Limitation 1: SHA256 Cannot Be Automated

**Why:** Homebrew's integrity model requires humans to verify downloads.

**Workaround Options:**

**Option A: Accept manual SHA256 updates** (Current approach)
- Humans update SHA256 for each Renovate PR
- ~5 minutes per PR
- Most secure

**Option B: Add SHA256 automation script** (Proposed)
```bash
# scripts/update-renovate-sha256.sh
# - Runs in CI on Renovate PRs
# - Downloads tarball
- Calculates SHA256
# - Pushes commit back to PR branch
# - Humans still review, but saves manual work
```

**Option C: Enhanced Renovate config with custom manager**
- Write custom Renovate manager that fetches SHA256
- Complex to implement
- May not work for all package formats

**Recommendation:** Start with Option A (manual), implement Option B later if volume is high.

### Limitation 2: Cannot Detect Structural Changes

**Why:** Renovate only does text replacement, doesn't inspect tarballs.

**Impact:** If a package changes its directory structure, Renovate won't know.

**Example:**
```ruby
# Version 1.8.27 structure:
quarto-1.8.27/
  bin/quarto

# Version 2.0.0 structure changed:
quarto-2.0.0/
  usr/bin/quarto  # ← Binary moved!
```

**Workaround:**
- CI will fail with "binary not found"
- Human reviews error, updates binary path
- This is why major versions need manual review

**With metadata caching (proposed):**
- Cache will have old structure
- Validation script will warn about path mismatch
- CI updates cache with new structure after merge

## Frequency and Volume

**Current configuration:**
- Runs every 3 hours
- Max 2 PRs per hour
- Max 5 concurrent PRs

**Expected volume:**
- ~2-5 packages in tap currently
- ~1-2 updates per month per package
- ~5-10 Renovate PRs per month total

**Time investment:**
- ~5 minutes per patch/minor update (update SHA256)
- ~15 minutes per major update (check structure, test)

## Future Improvements

### Proposed: Metadata Cache Integration

When implemented:
1. Renovate creates PR with version bump
2. Human updates SHA256
3. CI downloads new tarball
4. **CI extracts metadata and updates artifact cache**
5. CI runs tests
6. Merge
7. Next Copilot run has fresh metadata

**Benefits:**
- Copilot can validate against latest package structures
- Cache stays in sync with actual packages
- No manual cache maintenance needed

### Proposed: SHA256 Auto-Update Bot

A GitHub Action that:
1. Triggers on Renovate PRs
2. Downloads the new tarball
3. Calculates SHA256
4. Commits update to PR branch
5. Comments with tarball structure info

**Human workflow becomes:**
1. Renovate creates PR
2. Bot adds SHA256 commit automatically
3. Human reviews bot's structure analysis
4. Approve and merge (or fix if structure changed)

**Time savings:** ~4 minutes per PR

## Troubleshooting

### Renovate PR Failed CI

**Error:** `SHA256 mismatch`
- **Cause:** SHA256 not updated yet
- **Fix:** Update SHA256 per Step 2-3 above

**Error:** `binary not found at path quarto-1.9.0/bin/quarto`
- **Cause:** Tarball structure changed
- **Fix:** Extract tarball, find new binary path, update cask

**Error:** `Cask not found`
- **Cause:** Renovate regex didn't match cask format
- **Fix:** Check cask follows standard format (version/sha256/url order)

### Renovate Not Creating PRs

**Check:**
```bash
# View Renovate config
cat .github/renovate.json5

# Check if package matches regex
# Must have:
# - version line before sha256
# - sha256 line before url  
# - GitHub URL in url line
```

**Debug:**
- Check Renovate logs: https://app.renovatebot.com/dashboard
- Verify package has GitHub releases
- Ensure version string is simple (no complex version schemes)

## Summary

**What Renovate does:** Monitors releases, creates PRs with version bumps

**What humans do:** 
1. Calculate and update SHA256 (~2 min)
2. Verify tarball structure (~2 min)
3. Review and merge (~1 min)

**Total time per update:** ~5 minutes

**Frequency:** ~5-10 PRs per month

**With proposed improvements:** ~1 minute per update (bot handles SHA256)

---

**See Also:**
- [Renovate Configuration](.github/renovate.json5)
- [CASK_CREATION_GUIDE.md](CASK_CREATION_GUIDE.md) - SHA256 verification
- [CI Workflow](.github/workflows/tests.yml)
