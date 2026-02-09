# Renovate SHA256 Automation Plan

**Date:** 2026-02-09  
**Status:** Design Phase  
**Purpose:** Enable fully automatic package updates via Renovate

## Problem Statement

Currently, Renovate creates PRs with updated versions but **wrong SHA256 checksums**, requiring humans to:
1. Download the new tarball
2. Calculate SHA256 manually
3. Update the cask/formula
4. Wait for CI to pass

This defeats the purpose of automation and requires ~5 minutes per update × ~10 updates/month = ~50 minutes/month.

## Research: Renovate's Built-in Capabilities

### Renovate's `homebrew` Manager

According to [Renovate documentation](https://docs.renovatebot.com/modules/manager/homebrew/):

> "When a new version is available, Renovate:
> 1. Downloads the new tarball from GitHub or NPM registry
> 2. Calculates the SHA256 checksum
> 3. Updates both the `url` and `sha256` fields in the Formula file"

**Limitations:**
- ✅ **Formulas:** Full support with automatic SHA256
- ❌ **Casks:** Not supported (file pattern: `/^Formula/[^/]+[.]rb$/` only)
- ❌ **Custom URLs:** Only GitHub releases/archives and NPM registry

### Our Current Setup

We use `regexManagers` in `.github/renovate.json5`:
- ✅ Works for both `Formula/` and `Casks/`
- ❌ Only does text replacement (no SHA256 calculation)
- ❌ Cannot download files or execute code

## Solution: Hybrid Approach

Use **both** Renovate's built-in manager and custom workflows:

### For Formulas (Formula/*.rb)
Use Renovate's built-in `homebrew` manager → **Fully automatic** ✅

### For Casks (Casks/*.rb)
Add custom GitHub Actions workflow → **Fully automatic** ✅

## Implementation

### Part 1: Update Renovate Configuration

Modify `.github/renovate.json5` to use built-in manager for formulas:

```json5
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  
  "schedule": ["every 3 hours"],
  "semanticCommits": "enabled",
  "commitMessagePrefix": "chore(deps):",
  "platformAutomerge": true,
  
  // Enable built-in homebrew manager for formulas
  "homebrew": {
    "enabled": true
  },
  
  // Still use regex manager for casks (not supported by built-in manager)
  "regexManagers": [
    {
      // Match Homebrew Cask files
      // Built-in manager doesn't support casks, so use regex
      "fileMatch": ["^Casks/.+\\.rb$"],
      "matchStrings": [
        "version\\s+[\"'](?<currentValue>[^\"']+)[\"']\\s*\\n\\s*sha256\\s+[\"'](?<currentDigest>[^\"']+)[\"']\\s*\\n\\s*url\\s+[\"']https://github\\.com/(?<depName>[^/]+/[^/]+)/.*[\"']"
      ],
      "datasourceTemplate": "github-releases",
      "extractVersionTemplate": "^v?(?<version>.+)$"
    }
  ],
  
  "packageRules": [
    {
      // Auto-merge formulas (SHA256 handled by built-in manager)
      "matchFileNames": ["Formula/**"],
      "matchUpdateTypes": ["patch"],
      "automerge": true,
      "minimumReleaseAge": "3 hours"
    },
    {
      "matchFileNames": ["Formula/**"],
      "matchUpdateTypes": ["minor"],
      "automerge": true,
      "minimumReleaseAge": "1 day"
    },
    {
      // Casks need workflow to update SHA256 first
      "matchFileNames": ["Casks/**"],
      "matchUpdateTypes": ["patch"],
      "automerge": true,
      "minimumReleaseAge": "3 hours",
      "labels": ["cask-update"]
    },
    {
      "matchFileNames": ["Casks/**"],
      "matchUpdateTypes": ["minor"],
      "automerge": true,
      "minimumReleaseAge": "1 day",
      "labels": ["cask-update"]
    },
    {
      // Major updates always need review
      "matchUpdateTypes": ["major"],
      "automerge": false,
      "labels": ["major-update", "needs-review"]
    }
  ]
}
```

### Part 2: Add Cask SHA256 Auto-Update Workflow

Create `.github/workflows/cask-sha256-update.yml`:

```yaml
name: Update SHA256 for Cask Updates

on:
  pull_request:
    types: [opened, synchronize]
    paths:
      - 'Casks/**'

jobs:
  update-sha256:
    # Only run on Renovate PRs for casks
    if: |
      github.actor == 'renovate[bot]' &&
      contains(github.event.pull_request.labels.*.name, 'cask-update')
    
    runs-on: ubuntu-latest
    
    permissions:
      contents: write
      pull-requests: write
    
    steps:
      - name: Checkout PR branch
        uses: actions/checkout@v4
        with:
          ref: ${{ github.head_ref }}
          token: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Set up Ruby
        uses: ruby/setup-ruby@v1
        with:
          ruby-version: '3.2'
      
      - name: Detect changed casks
        id: changes
        run: |
          git fetch origin main
          CHANGED_CASKS=$(git diff --name-only origin/main...HEAD | grep '^Casks/.*\.rb$' || true)
          echo "casks<<EOF" >> $GITHUB_OUTPUT
          echo "$CHANGED_CASKS" >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
      
      - name: Update SHA256 for changed casks
        if: steps.changes.outputs.casks != ''
        run: |
          for cask_file in ${{ steps.changes.outputs.casks }}; do
            echo "Processing $cask_file..."
            
            # Extract version and URL from cask
            VERSION=$(ruby -e "
              content = File.read('$cask_file')
              if content =~ /version\s+[\"']([^\"']+)[\"']/
                puts \$1
              end
            ")
            
            URL_TEMPLATE=$(ruby -e "
              content = File.read('$cask_file')
              if content =~ /url\s+[\"']([^\"']+)[\"']/
                puts \$1
              end
            ")
            
            # Replace #{version} placeholder with actual version
            URL=$(echo "$URL_TEMPLATE" | sed "s/#{version}/$VERSION/g")
            
            echo "Version: $VERSION"
            echo "URL: $URL"
            
            # Download and calculate SHA256
            echo "Downloading tarball..."
            curl -sSL "$URL" -o /tmp/download.tar.gz
            
            NEW_SHA256=$(sha256sum /tmp/download.tar.gz | awk '{print $1}')
            echo "Calculated SHA256: $NEW_SHA256"
            
            # Update SHA256 in cask file
            ruby -i -pe "
              if /sha256\s+[\"'][a-f0-9]{64}[\"']/
                gsub(/[a-f0-9]{64}/, '$NEW_SHA256')
              end
            " "$cask_file"
            
            echo "✅ Updated $cask_file with new SHA256"
            
            rm /tmp/download.tar.gz
          done
      
      - name: Commit SHA256 updates
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          
          if git diff --quiet; then
            echo "No SHA256 changes needed"
            exit 0
          fi
          
          git add -A
          git commit -m "chore(cask): update SHA256 checksums

Automatically calculated SHA256 for cask version updates.

Co-authored-by: renovate[bot] <29139614+renovate[bot]@users.noreply.github.com>"
          
          git push
      
      - name: Comment on PR
        uses: actions/github-script@v7
        with:
          script: |
            const changedCasks = `${{ steps.changes.outputs.casks }}`.trim();
            if (!changedCasks) return;
            
            const caskList = changedCasks.split('\n').map(c => `- \`${c}\``).join('\n');
            
            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body: `✅ **SHA256 checksums automatically updated**

Updated casks:
${caskList}

CI will now verify the checksums and test installation.`
            });
```

## How It Works End-to-End

### Formulas (Fully Automatic)
```
1. New version released (e.g., jq 1.7.0 → 1.7.1)
   ↓
2. Renovate detects update (built-in manager)
   ↓
3. Renovate downloads tarball, calculates SHA256
   ↓
4. Renovate creates PR with version AND SHA256 updated ✅
   ↓
5. CI tests pass
   ↓
6. Auto-merge after 3 hours (patch) or 1 day (minor) ✅
```

**Human intervention:** NONE ✅

### Casks (Fully Automatic with Workflow)
```
1. New version released (e.g., Quarto 1.8.27 → 1.8.28)
   ↓
2. Renovate detects update (regex manager)
   ↓
3. Renovate creates PR with version updated, SHA256 wrong ⚠️
   ↓
4. SHA256 workflow triggers (detects renovate[bot] + cask-update label)
   ↓
5. Workflow downloads tarball, calculates SHA256
   ↓
6. Workflow commits SHA256 update to PR ✅
   ↓
7. CI tests pass
   ↓
8. Auto-merge after 3 hours (patch) or 1 day (minor) ✅
```

**Human intervention:** NONE ✅

### Major Updates (Manual Review)
```
1. Major version released (e.g., Quarto 1.8.27 → 2.0.0)
   ↓
2. Renovate/workflow update version and SHA256
   ↓
3. CI tests pass
   ↓
4. Human reviews for breaking changes ⚠️
   - Check tarball structure
   - Verify binary paths
   - Test desktop integration
   ↓
5. Manual merge after approval ✅
```

**Human intervention:** Review and approve (~5 min)

## Benefits

**Time savings:**
- Current: ~5 min per update × ~10 updates/month = **50 min/month**
- With automation: ~0 min for patch/minor, ~5 min for major = **~5-10 min/month**
- **Savings: 80-90% reduction in manual work**

**Quality improvements:**
- ✅ SHA256 always correct (calculated, not manual)
- ✅ Immediate updates (no waiting for humans)
- ✅ Consistent process (no human error)
- ✅ Focus human time on high-value work (major updates)

## Security Considerations

**Is automatic SHA256 calculation safe?**

✅ **YES** because:

1. **Verification still happens**
   - CI validates SHA256 against download
   - Homebrew verifies on user installation
   - Git history shows all changes

2. **Limited scope**
   - Only runs on Renovate PRs (trusted bot)
   - Only updates SHA256 (nothing else)
   - Uses URLs already in cask (not arbitrary)

3. **Transparency**
   - Bot comments explaining changes
   - Git commit shows who/what updated SHA256
   - CI logs show download and verification

4. **Same security as built-in manager**
   - Formulas use Renovate's built-in SHA256 calculation
   - Casks use same approach via workflow
   - Both are equally secure

**This is standard practice:** Many Homebrew taps use similar automation.

## Testing Plan

### Phase 1: Test with Formulas (Week 1)
1. Update `renovate.json5` to enable built-in manager
2. Wait for Renovate to detect update
3. Verify SHA256 is automatically updated
4. Verify CI passes and auto-merge works

### Phase 2: Test with Casks (Week 2)
1. Deploy cask SHA256 workflow
2. Trigger test by manually creating a cask update PR as renovate[bot]
3. Verify workflow downloads, calculates, commits SHA256
4. Verify CI passes and auto-merge works

### Phase 3: Production Testing (Week 3)
1. Monitor real Renovate PRs (both formulas and casks)
2. Verify no human intervention needed
3. Measure time savings
4. Document any issues

### Phase 4: Validation (Week 4)
1. Review 5+ automatic merges
2. Verify SHA256 checksums are correct
3. Test installations work
4. Sign off on automation

## Rollout Plan

### Week 1: Formulas Only
- Update Renovate config for formulas
- Monitor formula updates
- Validate built-in manager works correctly

### Week 2: Add Cask Workflow
- Deploy SHA256 workflow for casks
- Test with one cask manually
- Monitor first real cask update

### Week 3: Full Automation
- Enable auto-merge for both formulas and casks
- Monitor all updates
- Document any issues

### Week 4: Documentation
- Update RENOVATE_GUIDE.md
- Add troubleshooting section
- Document success metrics

## Success Metrics

**Baseline (Current):**
- 100% updates require human intervention
- ~5 min per update
- ~50 min/month total maintenance time

**Target (After Automation):**
- 80% updates fully automatic (patch/minor)
- 20% require human review (major)
- ~10 min/month total maintenance time

**Measurements:**
- Track Renovate PR count
- Track auto-merge success rate
- Track time spent on manual reviews
- Track SHA256 accuracy (0 errors expected)

## Documentation Updates Needed

- [x] This plan document
- [ ] Update `.github/renovate.json5`
- [ ] Create `.github/workflows/cask-sha256-update.yml`
- [ ] Update `docs/RENOVATE_GUIDE.md` (remove manual SHA256 steps)
- [ ] Add troubleshooting section for automation
- [ ] Update `AGENTS.md` with automation details

## Questions for Review

1. ✅ **Should we use built-in manager for formulas?** YES
2. ✅ **Should we add SHA256 workflow for casks?** YES
3. **Should major updates auto-merge?** NO (needs review)
4. **What release age for auto-merge?** 3 hours (patch), 1 day (minor)
5. **Should we test on a separate branch first?** YES

## Next Steps

1. Review and approve this plan
2. Update Renovate configuration
3. Create SHA256 workflow
4. Test with one formula update
5. Test with one cask update
6. Enable full automation
7. Monitor and document results

---

**Status:** Design Complete - Ready for Implementation  
**Owner:** Repository maintainer  
**Timeline:** 4 weeks from approval  
**Risk:** Low (can disable if issues found)
