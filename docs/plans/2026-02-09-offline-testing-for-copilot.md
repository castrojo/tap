# Offline Testing Plan for Copilot - Design Document

**Date:** 2026-02-09  
**Status:** Design Phase  
**Purpose:** Enable Copilot to validate packages without network access using GitHub Actions Artifacts

## Problem Statement

Copilot Coding Agent has no network access when creating casks/formulas, which prevents:
- Downloading tarballs to inspect structure
- Verifying SHA256 checksums
- Detecting desktop files and icons for GUI integration
- Confirming binary paths match actual tarball contents

This leads to incomplete casks (missing desktop integration) and potential errors (wrong binary paths).

## Solution: Metadata Artifact Cache

Pre-download and cache package metadata as GitHub Actions artifacts. Copilot downloads the cache at the start of its workflow and validates against cached metadata.

## Architecture

### Component 1: Artifact Structure

```
package-metadata-cache/
├── manifest.json           # Index of all cached packages
├── sublime-text-linux/
│   ├── metadata.json       # Version, URL, SHA256, structure info
│   ├── file-listing.txt    # Complete tar -tzf output
│   ├── desktop-files/      # Extracted .desktop files (if present)
│   │   └── sublime-text.desktop
│   └── icons/              # Extracted icons (if present)
│       └── sublime-text.png
├── quarto-linux/
│   ├── metadata.json
│   ├── file-listing.txt
│   ├── desktop-files/
│   │   └── quarto.desktop
│   └── icons/
│       └── quarto.png
└── cache-info.json         # Cache metadata (generated time, version)
```

**Artifact Naming:** `package-metadata-cache-<YYYY-MM-DD>`  
**Retention:** 90 days (GitHub Actions default)  
**Size Estimate:** ~1-5 MB per package (text files + icons)

### Component 2: Metadata Schema

#### manifest.json
```json
{
  "cache_version": "1.0",
  "generated_at": "2026-02-09T12:34:56Z",
  "packages": {
    "sublime-text-linux": {
      "type": "cask",
      "last_updated": "2026-02-09T12:34:56Z",
      "source_url": "https://www.sublimetext.com/",
      "github_repo": null
    },
    "quarto-linux": {
      "type": "cask",
      "last_updated": "2026-02-09T12:35:10Z",
      "source_url": "https://quarto.org/",
      "github_repo": "quarto-dev/quarto-cli"
    }
  },
  "total_packages": 2
}
```

#### metadata.json (per-package)
```json
{
  "package_name": "quarto-linux",
  "package_type": "cask",
  "version": "1.8.27",
  "download_url": "https://github.com/quarto-dev/quarto-cli/releases/download/v1.8.27/quarto-1.8.27-linux-amd64.tar.gz",
  "sha256": "bdf689b5589789a1f21d89c3b83d78ed02a97914dd702e617294f2cc1ea7387d",
  "size_bytes": 245678912,
  "format": "tar.gz",
  "download_date": "2026-02-09T12:35:10Z",
  
  "structure": {
    "root_directory": "quarto-1.8.27",
    "has_version_in_path": true,
    "version_pattern": "quarto-{VERSION}",
    "total_files": 1234,
    "total_directories": 56,
    
    "binaries": [
      {
        "path": "quarto-1.8.27/bin/quarto",
        "size": 12345678,
        "is_executable": true
      }
    ],
    
    "desktop_files": [
      {
        "path": "quarto-1.8.27/share/applications/quarto.desktop",
        "extracted_to": "desktop-files/quarto.desktop",
        "exec_line": "Exec=/usr/lib/quarto/bin/quarto",
        "icon_line": "Icon=/usr/lib/quarto/share/icons/quarto.png"
      }
    ],
    
    "icons": [
      {
        "path": "quarto-1.8.27/share/icons/hicolor/128x128/apps/quarto.png",
        "extracted_to": "icons/quarto.png",
        "size": "128x128",
        "format": "png"
      }
    ],
    
    "config_files": [
      "quarto-1.8.27/share/quarto/config.yaml"
    ]
  },
  
  "desktop_integration": {
    "is_gui_app": true,
    "has_desktop_file": true,
    "has_icons": true,
    "desktop_file_needs_patching": true,
    "icon_needs_patching": false,
    "recommended_target_name": "quarto"
  },
  
  "validation": {
    "sha256_verified": true,
    "tarball_extracted_successfully": true,
    "all_binaries_executable": true
  }
}
```

#### cache-info.json
```json
{
  "cache_version": "1.0",
  "generated_at": "2026-02-09T12:36:00Z",
  "workflow_run_id": "21813776351",
  "workflow_run_url": "https://github.com/castrojo/tap/actions/runs/21813776351",
  "packages_cached": 2,
  "total_size_bytes": 5242880,
  "generation_duration_seconds": 120
}
```

### Component 3: Workflows

#### Workflow 1: cache-metadata.yml (Pre-cache Generation)

**Triggers:**
- Schedule: Daily at 2:00 AM UTC
- Manual: `workflow_dispatch`
- On merge to main (if cask/formula changed)

**Steps:**
1. Checkout repository
2. Parse all casks and formulas
3. For each package:
   - Download tarball/archive
   - Calculate SHA256 (verify against cask)
   - Extract tarball to temp directory
   - Generate file listing (`tar -tzf`)
   - Find and extract desktop files
   - Find and extract icons
   - Analyze structure (binary paths, version patterns)
   - Create metadata.json
4. Generate manifest.json
5. Generate cache-info.json
6. Upload as artifact: `package-metadata-cache-<date>`
7. Delete old artifacts (keep last 7 days)

**Artifact Upload:**
```yaml
- name: Upload metadata cache
  uses: actions/upload-artifact@v4
  with:
    name: package-metadata-cache-${{ env.DATE }}
    path: metadata-cache/
    retention-days: 90
```

#### Workflow 2: Enhanced tests.yml (CI with Cache Update)

**Current flow:** Download tarball → Audit → Discard

**New flow:** 
1. Download tarball
2. Audit and test
3. **Extract metadata and update cache artifact**
4. Upload updated cache

**Added step:**
```yaml
- name: Update metadata cache
  if: steps.changed-files.outputs.any_changed == 'true'
  run: |
    # Extract metadata from tested tarballs
    ./scripts/extract-metadata.sh "${{ steps.changed-files.outputs.all_changed_files }}"
    
- name: Upload updated cache
  if: steps.changed-files.outputs.any_changed == 'true'
  uses: actions/upload-artifact@v4
  with:
    name: package-metadata-cache-${{ env.DATE }}
    path: metadata-cache/
    retention-days: 90
```

**Benefits:**
- Cache stays fresh automatically
- Renovate updates → CI tests → Cache updated
- No manual cache maintenance

#### Workflow 3: Enhanced Copilot Workflow

**Current flow:** Copilot creates cask → Pushes → CI tests

**New flow:**
1. **Download latest metadata cache**
2. Copilot creates cask
3. **Validate against cached metadata**
4. Push → CI tests with real download

**Added steps:**
```yaml
- name: Download metadata cache
  uses: actions/download-artifact@v4
  with:
    name: package-metadata-cache-latest
    path: .cache/metadata/
    
- name: Validate with cache
  run: |
    ./scripts/validate-with-cache.sh Casks/new-package.rb
```

### Component 4: Validation Scripts

#### scripts/validate-with-cache.sh

**Purpose:** Validate a cask against cached metadata without network

**Usage:**
```bash
./scripts/validate-with-cache.sh Casks/quarto-linux.rb
```

**Checks:**
1. **SHA256 Match**
   - Cask SHA256 matches cached SHA256
   - Error if mismatch (version updated but cache stale?)

2. **Binary Path Validation**
   - Cask binary path exists in cached file listing
   - Warn if version interpolation doesn't match structure
   
3. **Desktop Integration Validation**
   - If cache indicates GUI app, check for desktop file artifact
   - If cache indicates GUI app, check for icon artifact
   - Error if missing desktop integration for GUI app

4. **Structure Consistency**
   - Binary path pattern matches cached version pattern
   - Desktop file paths match cached locations
   - Icon paths match cached locations

**Output:**
```
✓ SHA256 matches cached value
✓ Binary path quarto-1.8.27/bin/quarto exists in tarball
⚠ Desktop integration detected in cache but missing from cask
  Cached desktop file: quarto-1.8.27/share/applications/quarto.desktop
  Cached icon: quarto-1.8.27/share/icons/hicolor/128x128/apps/quarto.png
  
  Suggested additions:
  
  artifact "quarto-#{version}/share/applications/quarto.desktop",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/quarto.desktop"
  artifact "quarto-#{version}/share/icons/hicolor/128x128/apps/quarto.png",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/quarto.png"
  
  preflight do
    xdg_data_home = ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
    FileUtils.mkdir_p "#{xdg_data_home}/applications"
    FileUtils.mkdir_p "#{xdg_data_home}/icons"
  end

Exit code: 1 (validation warnings present)
```

#### scripts/extract-metadata.sh

**Purpose:** Extract metadata from a downloaded tarball

**Usage:**
```bash
./scripts/extract-metadata.sh quarto-1.8.27-linux-amd64.tar.gz quarto-linux
```

**Process:**
1. Extract tarball to temp directory
2. Generate file listing
3. Find binaries (executable files in bin/, usr/bin/, etc.)
4. Find desktop files (*.desktop)
5. Find icons (*.png, *.svg in icons/, pixmaps/, share/)
6. Analyze desktop file (Exec=, Icon= lines)
7. Create metadata.json
8. Copy desktop files and icons to cache
9. Clean up temp directory

#### scripts/cache-lookup.sh

**Purpose:** Query the metadata cache

**Usage:**
```bash
# Get metadata for a package
./scripts/cache-lookup.sh quarto-linux

# Check if package is cached
./scripts/cache-lookup.sh --exists quarto-linux

# List all cached packages
./scripts/cache-lookup.sh --list
```

### Component 5: Integration with tap-tools (Go CLI)

Enhance Go tools to use metadata cache:

#### tap-cask (Enhanced)

```bash
# Current behavior: Generate cask from GitHub releases
./tap-tools/tap-cask generate quarto-linux https://github.com/quarto-dev/quarto-cli

# New behavior: Check cache first
./tap-tools/tap-cask generate quarto-linux https://github.com/quarto-dev/quarto-cli --use-cache

# Cache hit: Use cached metadata for structure
# Cache miss: Download and extract (current behavior)
```

**Cache-aware generation:**
1. Check if package exists in cache
2. If cached:
   - Use cached SHA256
   - Use cached binary paths
   - Include desktop integration if detected
   - Use cached version (or prompt for update)
3. If not cached:
   - Download and extract (current behavior)
   - Optionally add to cache

#### tap-validate (Enhanced)

```bash
# Current behavior: Basic Ruby syntax validation
./tap-tools/tap-validate file Casks/quarto-linux.rb

# New behavior: Validate against cache
./tap-tools/tap-validate file Casks/quarto-linux.rb --with-cache

# Checks:
# - SHA256 matches cache
# - Binary paths exist in cached listing
# - Desktop integration present if GUI app
```

## Implementation Plan

### Phase 1: Core Infrastructure (Week 1)
- [ ] Create metadata schema (JSON structure)
- [ ] Write `extract-metadata.sh` script
- [ ] Write `cache-lookup.sh` script
- [ ] Test manually with 2-3 existing casks

### Phase 2: Cache Generation (Week 2)
- [ ] Create `cache-metadata.yml` workflow
- [ ] Test artifact upload/download
- [ ] Generate cache for all existing packages
- [ ] Verify artifact size and structure

### Phase 3: Validation Integration (Week 3)
- [ ] Write `validate-with-cache.sh` script
- [ ] Add cache download to Copilot workflow
- [ ] Test validation with correct/incorrect casks
- [ ] Document validation output format

### Phase 4: CI Integration (Week 4)
- [ ] Enhance `tests.yml` to update cache
- [ ] Test cache update on Renovate PRs
- [ ] Verify cache stays fresh
- [ ] Add cache status badges to README

### Phase 5: Go Tools Enhancement (Week 5-6)
- [ ] Add cache support to `tap-cask`
- [ ] Add cache support to `tap-validate`
- [ ] Add cache support to `tap-issue`
- [ ] Update documentation

## Testing Strategy

### Unit Tests
- Metadata extraction from various tarball formats
- JSON schema validation
- Cache lookup queries
- Path interpolation logic

### Integration Tests
- Full cache generation workflow
- Cache download in Copilot workflow
- Validation against cached metadata
- Cache update after CI tests

### End-to-End Tests
1. Generate cache for test package
2. Copilot creates cask using cache
3. Validation catches missing desktop integration
4. CI tests real download
5. Cache updated with new version

## Benefits

### For Copilot
- ✅ Validate packages without network access
- ✅ Detect desktop integration requirements
- ✅ Verify binary paths before push
- ✅ Immediate feedback (no wait for CI)

### For Humans
- ✅ Fewer incomplete casks (desktop integration detected)
- ✅ Fewer CI failures (validation catches errors early)
- ✅ Faster reviews (Copilot produces better initial casks)
- ✅ Automatic cache maintenance (no manual updates)

### For Repository
- ✅ Higher quality packages from Copilot
- ✅ Reduced iteration cycles
- ✅ Better documentation (metadata is queryable)
- ✅ Enables future automation (auto-generate docs, etc.)

## Limitations and Trade-offs

### Limitations
- **Cold start:** First package has no cache (must download)
- **Cache staleness:** Cache may lag behind latest release
- **Storage:** Artifacts use GitHub storage quota
- **Complexity:** More moving parts to maintain

### Trade-offs
- **Repo size:** Artifacts don't affect repo size (stored separately)
- **CI time:** Cache generation adds ~2-5 min per package
- **Maintenance:** Cache updates automatically (minimal human work)

### Mitigation
- Cache expires after 90 days (automatic cleanup)
- CI updates cache on every test (stays fresh)
- Fallback to pattern-based validation if cache unavailable
- Document cache structure for troubleshooting

## Rollout Strategy

### Stage 1: Opt-in (2 weeks)
- Cache available, not required
- Copilot can use cache if available
- Fallback to current behavior if unavailable
- Monitor adoption and issues

### Stage 2: Default (2 weeks)
- Cache download becomes default in Copilot workflow
- Validation warnings logged but not blocking
- Gather feedback on false positives/negatives

### Stage 3: Enforcement (4 weeks)
- Validation errors block Copilot completion
- Cache required for package creation
- Full integration with Go tools

## Success Metrics

- **Copilot cask quality:** % with desktop integration when needed
- **CI failure rate:** Reduce failures due to missing desktop files
- **Iteration cycles:** Reduce PR revisions per package
- **Human review time:** Reduce time spent on incomplete casks
- **Cache hit rate:** % of Copilot runs with fresh cache

**Target:** 
- 90%+ desktop integration detection
- 50% reduction in CI failures
- 30% reduction in PR revisions

## Future Enhancements

### Short-term
- SHA256 auto-update bot (integrate with Renovate)
- Cache API for external tools
- Web UI to browse cached metadata

### Long-term
- Machine learning to detect app types
- Automatic cask generation from cache
- Cross-repository cache sharing

## Documentation Deliverables

- [x] This design document
- [x] RENOVATE_GUIDE.md (explains Renovate workflow)
- [ ] CACHE_ARCHITECTURE.md (technical details)
- [ ] scripts/README.md (script usage)
- [ ] Update AGENTS.md (Copilot instructions)
- [ ] Update CASK_CREATION_GUIDE.md (validation workflow)

## Questions for Review

1. Is 90-day artifact retention sufficient?
2. Should cache be versioned (allow rollback)?
3. Should we cache formulas too (not just casks)?
4. Should cache include build instructions for formulas?
5. Should validation warnings be blocking or advisory?

---

**Status:** Design Complete - Ready for Implementation  
**Next Steps:** Review design, approve, begin Phase 1 implementation
