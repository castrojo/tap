# Troubleshooting Guide: Homebrew Packaging

This guide helps you diagnose and fix common issues when creating and maintaining Homebrew formulas and casks.

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Formula/Cask Errors](#formulacask-errors)
3. [Download/SHA256 Issues](#downloadsha256-issues)
4. [Build Failures](#build-failures)
5. [Installation Errors](#installation-errors)
6. [Test Failures](#test-failures)
7. [Renovate/Automation Issues](#renovateautomation-issues)
8. [Git/PR Workflow](#gitpr-workflow)

---

## Quick Reference

| Error Type | Common Symptoms | Quick Fix |
|------------|----------------|-----------|
| Style violations | `brew style` failures | Run `brew style --fix` |
| Audit failures | `brew audit --strict` errors | Check formula DSL syntax |
| SHA256 mismatch | `SHA256 mismatch` | Recalculate with `shasum -a 256` |
| Missing license | `missing license` audit error | Add `license` stanza |
| Dead URL | `DownloadError: 404` | Verify URL or use GitHub archive |
| Build failure | `Error: undefined method` | Check Ruby DSL syntax |
| Test failure | `test do` block fails | Test block returns non-zero |
| Renovate stuck | No PRs created | Check `.github/renovate.json` syntax |
| CI failure | Actions fail on PR | Check `brew test-bot` output |

---

## Formula/Cask Errors

### Error: `undefined method 'sha256' for nil:NilClass`

**Symptom:**
```
Error: Invalid formula: /homebrew-tap/Formula/package.rb
undefined method 'sha256' for nil:NilClass
```

**Cause:**
The formula uses `sha256` before defining a `url` or the `url` is malformed.

**Solution:**
```bash
# Edit formula to ensure url comes before sha256
class Package < Formula
  desc "Description"
  homepage "https://..."
  url "https://github.com/user/repo/archive/refs/tags/v1.0.0.tar.gz"  # Must come first
  sha256 "abc123..."  # Must come after url
  license "MIT"
end
```

**Prevention:**
- Always define `url` before `sha256`
- Use `scripts/quickstart.sh` which generates correct order
- Run `brew audit --strict` before committing

---

### Error: `FormulaAudit failed: missing license`

**Symptom:**
```
Error: /homebrew-tap/Formula/package.rb:
  * missing license: add a license stanza to the formula [formula/missing_license]
```

**Cause:**
Formula lacks a `license` stanza (required since Homebrew/core policy change).

**Solution:**
```bash
# Find the license in the upstream repository
gh repo view owner/repo --json licenseInfo

# Add to formula after homepage
class Package < Formula
  desc "Description"
  homepage "https://..."
  url "https://..."
  sha256 "..."
  license "MIT"  # Add this line
  
  def install
    # ...
  end
end

# Verify
brew audit --strict Formula/package.rb
```

**Prevention:**
- Always check `LICENSE` file in upstream repo
- Use `scripts/quickstart.sh` which prompts for license
- Common licenses: `MIT`, `Apache-2.0`, `GPL-3.0`, `BSD-3-Clause`

---

### Error: `style violations`

**Symptom:**
```
Formula/package.rb:12:3: C: Layout/EmptyLineAfterGuardClause: Add empty line after guard clause.
Formula/package.rb:15:81: C: Layout/LineLength: Line is too long. [85/80]
```

**Cause:**
Ruby style violations (indentation, line length, spacing).

**Solution:**
```bash
# Automatic fix (recommended)
brew style --fix Formula/package.rb

# Verify
brew style Formula/package.rb

# For all formulas
brew style --fix
```

**Prevention:**
- Run `brew style --fix` before every commit
- Add to pre-commit hook
- Use consistent 2-space indentation
- Keep lines under 80 characters

---

### Error: `FormulaAudit failed: missing test`

**Symptom:**
```
Error: /homebrew-tap/Formula/package.rb:
  * A `test do` test block should be added [formula/missing_test]
```

**Cause:**
Formula lacks a `test do` block (recommended for all formulas).

**Solution:**
```bash
# Add test block to formula
class Package < Formula
  # ... existing code ...
  
  def install
    bin.install "package"
  end
  
  test do
    # Option 1: Version check (simplest)
    assert_match version.to_s, shell_output("#{bin}/package --version")
    
    # Option 2: Help text
    assert_match "Usage:", shell_output("#{bin}/package --help")
    
    # Option 3: Actual functionality
    system bin/"package", "command", "args"
  end
end

# Verify test works
brew install --build-from-source Formula/package.rb
brew test Formula/package.rb
```

**Prevention:**
- Always add `test do` block
- Test with `brew test` before committing
- Use version check as minimum test
- Reference: [Homebrew Formula Cookbook - Test](https://docs.brew.sh/Formula-Cookbook#add-a-test-to-the-formula)

---

### Error: `Invalid cask: unknown method`

**Symptom:**
```
Error: Invalid cask: /homebrew-tap/Casks/package.rb
Cask/package.rb:7: undefined method 'bin' for #<Cask::DSL>
```

**Cause:**
Using formula DSL methods (`bin`, `lib`, `system`) in cask file.

**Solution:**
```bash
# Casks use artifact stanzas, not install methods
cask "package" do
  version "1.0.0"
  sha256 "abc123..."
  
  url "https://..."
  name "Package"
  desc "Description"
  homepage "https://..."
  
  # Use artifact stanzas (not install methods)
  app "Package.app"                    # For GUI apps
  binary "#{appdir}/Package.app/Contents/MacOS/cli"  # For CLI
  
  # NOT: def install; bin.install "cli"; end
end
```

**Prevention:**
- Formulas = CLI tools, use `def install` with `bin.install`
- Casks = GUI apps, use artifact stanzas (`app`, `binary`)
- Reference: [Cask Cookbook](https://docs.brew.sh/Cask-Cookbook)

---

## Download/SHA256 Issues

### Error: `SHA256 mismatch`

**Symptom:**
```
Error: SHA256 mismatch
Expected: abc123...
  Actual: def456...
```

**Cause:**
- Incorrect SHA256 hash in formula
- Upstream file changed without version change
- Downloaded corrupted file

**Solution:**
```bash
# Method 1: Download and calculate correct hash
curl -L "https://github.com/user/repo/archive/refs/tags/v1.0.0.tar.gz" | shasum -a 256

# Method 2: Use brew to calculate
brew fetch --formula Formula/package.rb 2>&1 | grep "SHA256:"

# Update formula with correct hash
class Package < Formula
  # ...
  sha256 "def456..."  # Use the correct hash from above
end

# Verify
brew fetch --formula Formula/package.rb
```

**Prevention:**
- Always calculate SHA256 from the exact URL
- Don't copy hashes from release notes (may differ)
- Use `scripts/quickstart.sh` which auto-calculates
- Re-verify after any URL change

---

### Error: `DownloadError: 404 Not Found`

**Symptom:**
```
Error: Failed to download resource "package"
DownloadError: 404 Not Found
```

**Cause:**
- URL doesn't exist
- Release was deleted
- GitHub changed URL format
- Private repository

**Solution:**
```bash
# Verify URL exists
curl -IL "https://github.com/user/repo/archive/refs/tags/v1.0.0.tar.gz"

# If 404, try alternate URL formats:

# Format 1: GitHub archive (most reliable)
url "https://github.com/user/repo/archive/refs/tags/v1.0.0.tar.gz"

# Format 2: GitHub release asset
url "https://github.com/user/repo/releases/download/v1.0.0/package-1.0.0.tar.gz"

# Format 3: Tarball from release
gh api repos/user/repo/releases/tags/v1.0.0 | jq -r '.tarball_url'

# Update formula with working URL
brew fetch --formula Formula/package.rb  # Verify it downloads
```

**Prevention:**
- Prefer GitHub archive URLs (most stable)
- Test URL with `curl -IL` before using
- Check release assets with `gh release view`
- Avoid URLs that require authentication

---

### Error: `Trying to download SHA256 sum`

**Symptom:**
```
Error: Empty download: the URL is a SHA256 sum file
```

**Cause:**
The `url` points to a `.sha256` file instead of the actual archive.

**Solution:**
```bash
# Find the actual release files
gh release view v1.0.0 --repo user/repo

# Look for archives (.tar.gz, .zip)
# NOT checksums (.sha256, .sha256sum, .txt)

# Correct URL format
url "https://github.com/user/repo/releases/download/v1.0.0/package-1.0.0.tar.gz"
# NOT: package-1.0.0.tar.gz.sha256
```

**Prevention:**
- Look for `.tar.gz`, `.zip`, `.tgz` files
- Avoid `.sha256`, `.txt`, `.asc` files
- Use `scripts/quickstart.sh` which validates URLs

---

### Error: `Checksum mismatch after Renovate update`

**Symptom:**
```
Renovate PR created but brew fetch fails with SHA256 mismatch
```

**Cause:**
Renovate updated `version` but didn't update `sha256`.

**Solution:**
```bash
# Checkout the Renovate branch
git fetch origin
git checkout renovate/package-1.x

# Recalculate SHA256
VERSION="1.2.0"  # From the PR
URL=$(grep 'url "' Formula/package.rb | cut -d'"' -f2)
curl -L "$URL" | shasum -a 256

# Update formula
# Edit Formula/package.rb with correct sha256

# Commit fix
git add Formula/package.rb
git commit -m "fix: update sha256 for version ${VERSION}"
git push

# CI will re-run
```

**Prevention:**
- Use `${version}` in URL (Renovate can update it)
- Enable Renovate's `digest` update strategy
- Add GitHub Action to auto-calculate SHA256
- Review Renovate PRs before merging

---

## Build Failures

### Error: `No such file or directory - make`

**Symptom:**
```
Error: No such file or directory - make
```

**Cause:**
Formula runs `system "make"` but upstream doesn't use Makefiles.

**Solution:**
```bash
# Check build system in upstream repo
ls  # Look for: Makefile, CMakeLists.txt, meson.build, Cargo.toml, go.mod

# Golang project
def install
  system "go", "build", *std_go_args(ldflags: "-s -w")
end

# Rust project
def install
  system "cargo", "install", *std_cargo_args
end

# CMake project
def install
  system "cmake", "-S", ".", "-B", "build", *std_cmake_args
  system "cmake", "--build", "build"
  system "cmake", "--install", "build"
end

# Pre-built binary (no compilation)
def install
  bin.install "package"
end
```

**Prevention:**
- Check build system before writing `install` method
- Use `scripts/quickstart.sh` which detects build system
- Reference: [Formula Cookbook - Build Systems](https://docs.brew.sh/Formula-Cookbook)

---

### Error: `fatal error: 'dependency.h' file not found`

**Symptom:**
```
compilation terminated.
fatal error: 'openssl/ssl.h' file not found
```

**Cause:**
Missing build-time dependency.

**Solution:**
```bash
# Add depends_on for the missing library
class Package < Formula
  desc "..."
  # ...
  
  depends_on "openssl@3"  # For openssl/ssl.h
  depends_on "libpq"      # For libpq-fe.h
  depends_on "zlib"       # For zlib.h
  
  def install
    system "make", "install"
  end
end

# Test build
brew install --build-from-source Formula/package.rb
```

**Common Dependencies:**
- `openssl@3` - SSL/TLS
- `libpq` - PostgreSQL client
- `zlib` - Compression
- `libyaml` - YAML parsing
- `pkg-config` - Build tool

**Prevention:**
- Check upstream README for dependencies
- Look for `#include` statements in source
- Test with `brew install --build-from-source`

---

### Error: `undefined reference to 'symbol'`

**Symptom:**
```
/usr/bin/ld: undefined reference to `curl_easy_init'
collect2: error: ld returned 1 exit status
```

**Cause:**
Missing linker flags or dependencies.

**Solution:**
```bash
# Add dependency
class Package < Formula
  # ...
  depends_on "curl"
  
  def install
    # Set environment variables for linker
    ENV.append "LDFLAGS", "-L#{Formula["curl"].opt_lib}"
    ENV.append "CPPFLAGS", "-I#{Formula["curl"].opt_include}"
    
    system "make", "install"
  end
end
```

**Prevention:**
- Add all build dependencies
- Check upstream build instructions
- Use `pkg-config` when available

---

## Installation Errors

### Error: `No such file or directory @ rb_sysopen`

**Symptom:**
```
Error: No such file or directory @ rb_sysopen - /usr/local/Cellar/package/1.0.0/bin/package
```

**Cause:**
Formula's `install` method didn't actually install the binary.

**Solution:**
```bash
# Debug: Check what files were created
def install
  system "make"  # Build
  
  # Check what was created
  # Common locations: bin/, target/release/, build/
  
  # Install to Homebrew prefix
  bin.install "package"           # Single binary
  bin.install Dir["bin/*"]        # All files in bin/
  bin.install "target/release/package"  # Rust
end

# Test locally
brew install --debug --build-from-source Formula/package.rb
# After build fails, shell drops you in build directory
# Explore to find where the binary actually is
ls -R
```

**Prevention:**
- Check upstream installation docs
- Verify binary location before `bin.install`
- Test with `brew install --build-from-source`

---

### Error: `Permission denied @ apply2files`

**Symptom:**
```
Error: Permission denied @ apply2files - /usr/local/Cellar/package/1.0.0/bin/package
```

**Cause:**
Binary lacks execute permissions.

**Solution:**
```bash
def install
  bin.install "package"
  chmod 0755, bin/"package"  # Add execute permission
end
```

**Prevention:**
- Usually not needed (Homebrew sets permissions automatically)
- Only add if you see this specific error
- Check if upstream build system creates wrong permissions

---

### Error: `Failed to link package`

**Symptom:**
```
Error: Could not symlink bin/package
Target /usr/local/bin/package already exists.
```

**Cause:**
Another package or manual installation already provides this file.

**Solution:**
```bash
# Option 1: Remove conflict
brew unlink conflicting-package
brew link package

# Option 2: Overwrite
brew link --overwrite package

# Option 3: Force
brew link --overwrite --force package

# Check what provides the file
brew list --formula | xargs -I{} sh -c 'brew list {} | grep -q "bin/package" && echo {}'
```

**Prevention:**
- Check for conflicts before installing
- Document conflicts in formula caveats
- Consider renaming binary if truly conflicts

---

## Test Failures

### Error: `test block failed with exit status 1`

**Symptom:**
```
Testing package
==> /usr/local/Cellar/package/1.0.0/bin/package --version
Error: package: test block failed with exit status 1
```

**Cause:**
- Command in test block returned non-zero exit code
- Binary not found
- Test expects input/files not present

**Solution:**
```bash
# Debug the test
brew install Formula/package.rb
/usr/local/opt/package/bin/package --version  # Run manually

# Common fixes:

# Fix 1: Wrong flag
test do
  assert_match version.to_s, shell_output("#{bin}/package -v")  # Not --version
end

# Fix 2: Command requires arguments
test do
  (testpath/"test.txt").write("hello")
  assert_match "hello", shell_output("#{bin}/package test.txt")
end

# Fix 3: Command outputs to stderr
test do
  assert_match "Usage", shell_output("#{bin}/package --help 2>&1")
end

# Fix 4: Exit code check only
test do
  system bin/"package", "test"  # Just verify it runs
end
```

**Prevention:**
- Test manually before adding to formula
- Use `shell_output` for commands that output text
- Use `system` for commands that just need to succeed
- Handle both stdout and stderr with `2>&1`

---

### Error: `Failure while executing: version`

**Symptom:**
```
Error: Failure while executing: version
```

**Cause:**
Formula uses `version` outside valid context.

**Solution:**
```bash
# Correct usage in url
url "https://example.com/package-#{version}.tar.gz"  # OK

# Correct usage in test
test do
  assert_match version.to_s, shell_output("#{bin}/package --version")  # OK
end

# WRONG: version in string interpolation
url "https://example.com/package-${version}.tar.gz"  # WRONG - use #{}
```

**Prevention:**
- Use `#{version}` not `${version}`
- Use `version.to_s` in assertions
- Only reference `version` in Formula DSL context

---

### Error: `test assertion failed`

**Symptom:**
```
Error: expected "1.0.0" to match /2.0.0/
```

**Cause:**
Test assertion doesn't match actual output.

**Solution:**
```bash
# Debug
brew test Formula/package.rb  # See actual output

# Common issues:

# Issue 1: Version format differs
test do
  # Output is "v1.0.0" not "1.0.0"
  assert_match "v#{version}", shell_output("#{bin}/package --version")
end

# Issue 2: Version in middle of output
test do
  # Output is "package version 1.0.0 build 123"
  output = shell_output("#{bin}/package --version")
  assert_match version.to_s, output
end

# Issue 3: Multiline output
test do
  output = shell_output("#{bin}/package --version")
  assert output.include?(version.to_s), "Version not found in output: #{output}"
end
```

**Prevention:**
- Run actual command to see output format
- Use flexible matching patterns
- Test with multiple versions

---

## Renovate/Automation Issues

### Error: `Renovate not creating PRs`

**Symptom:**
No Renovate PRs appear even though new versions exist.

**Cause:**
- `.github/renovate.json` syntax error
- Formula URL doesn't use `#{version}`
- Renovate can't detect version pattern
- Rate limit reached

**Solution:**
```bash
# Check Renovate config syntax
cat .github/renovate.json | jq .  # Should parse cleanly

# Check Renovate logs
# Go to: https://github.com/castrojo/homebrew-tap/settings/installations
# Click "Configure" on Renovate
# Click "Logs" tab

# Fix formula to use version variable
class Package < Formula
  version "1.0.0"
  url "https://github.com/user/repo/archive/refs/tags/v#{version}.tar.gz"
  # NOT: url "https://github.com/user/repo/archive/refs/tags/v1.0.0.tar.gz"
end

# Force Renovate check
# Go to: https://github.com/castrojo/homebrew-tap
# Settings > GitHub Apps > Renovate > Configure
# Click "Check now"
```

**Prevention:**
- Always use `#{version}` in URL
- Validate `renovate.json` before committing
- Enable Renovate debug logging
- Check Renovate dashboard regularly

---

### Error: `Renovate PR has wrong SHA256`

**Symptom:**
Renovate creates PR but CI fails with SHA256 mismatch.

**Cause:**
Renovate can't auto-calculate SHA256 for GitHub archives.

**Solution:**
```bash
# Option 1: Manual fix (immediate)
# 1. Checkout PR branch
# 2. Recalculate SHA256 (see "SHA256 mismatch" section above)
# 3. Push fix

# Option 2: Add GitHub Action (permanent)
# Create .github/workflows/update-sha256.yml
name: Update SHA256
on:
  pull_request:
    paths:
      - 'Formula/**'
      - 'Casks/**'

jobs:
  update-sha256:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Update SHA256
        run: |
          # Script to recalculate and update SHA256
          # See AGENT_GUIDE.md for full example
```

**Prevention:**
- Use release assets with predictable names (Renovate handles better)
- Add automation to fix SHA256 in PRs
- Monitor Renovate PRs for failures

---

### Error: `Renovate created PR with syntax error`

**Symptom:**
```
Error: Invalid formula: Formula/package.rb
syntax error, unexpected end-of-input
```

**Cause:**
Renovate regex replacement broke Ruby syntax.

**Solution:**
```bash
# Check Renovate config
cat .github/renovate.json

# Ensure regex patterns are correct
{
  "customManagers": [
    {
      "customType": "regex",
      "fileMatch": ["^Formula/.+\\.rb$"],
      "matchStrings": [
        "version \"(?<currentValue>[^\"]+)\"",
        "url \"[^\"]+v(?<currentValue>[^\"]+)\\.tar\\.gz\""
      ]
    }
  ]
}

# Fix the formula manually
git checkout renovate/package-1.x
# Edit Formula/package.rb - fix syntax
git add Formula/package.rb
git commit -m "fix: correct syntax after renovate update"
git push
```

**Prevention:**
- Test Renovate regex patterns carefully
- Use CI to validate syntax before merge
- Consider simpler versioning patterns

---

## Git/PR Workflow

### Error: `remote: Permission to castrojo/homebrew-tap.git denied`

**Symptom:**
```
error: failed to push some refs to 'github.com:castrojo/homebrew-tap.git'
remote: Permission to castrojo/homebrew-tap.git denied to username.
```

**Cause:**
- Wrong GitHub authentication
- SSH key not configured
- HTTPS vs SSH URL mismatch
- Personal access token expired

**Solution:**
```bash
# Check current remote
git remote -v

# If HTTPS - ensure token is valid
gh auth status
gh auth login  # If needed

# If SSH - ensure key is added
ssh -T git@github.com

# Switch remote to HTTPS (recommended with gh CLI)
git remote set-url origin https://github.com/castrojo/homebrew-tap.git

# Or switch to SSH
git remote set-url origin git@github.com:castrojo/homebrew-tap.git
```

**Prevention:**
- Use HTTPS with GitHub CLI (`gh auth login`)
- Keep personal access tokens in secure storage
- Set remote URL during initial setup

---

### Error: `Updates were rejected because the tip of your current branch is behind`

**Symptom:**
```
error: failed to push some refs to 'origin'
hint: Updates were rejected because the tip of your current branch is behind
```

**Cause:**
Remote has commits not in local branch.

**Solution:**
```bash
# Option 1: Rebase (clean history)
git fetch origin
git rebase origin/main

# Fix conflicts if any
# Edit conflicting files
git add .
git rebase --continue

git push

# Option 2: Merge (preserves history)
git fetch origin
git merge origin/main
git push

# Option 3: Force push (dangerous - only for feature branches)
git push --force-with-lease
```

**Prevention:**
- Pull before starting work: `git pull --rebase`
- Create feature branches from updated main
- Never force push to main

---

### Error: `GitHub Actions CI failing on PR`

**Symptom:**
PR shows red X, test-bot workflow fails.

**Cause:**
- Formula fails `brew audit`
- Formula fails `brew install`
- Formula fails `brew test`

**Solution:**
```bash
# View CI logs
gh pr view 123 --web  # Click "Checks" tab

# Reproduce locally
git fetch origin
git checkout pr-branch

# Run same checks as CI
brew audit --strict --online Formula/package.rb
brew style Formula/package.rb
brew install --build-from-source Formula/package.rb
brew test Formula/package.rb

# Fix issues, commit, push
git add Formula/package.rb
git commit -m "fix: resolve CI failures"
git push
```

**Prevention:**
- Run full test suite before pushing
- Use `scripts/quickstart.sh` which runs checks
- Set up pre-push git hook

---

### Error: `Merge conflict in Formula/package.rb`

**Symptom:**
```
CONFLICT (content): Merge conflict in Formula/package.rb
Automatic merge failed; fix conflicts and then commit the result.
```

**Cause:**
Same lines modified in both branches.

**Solution:**
```bash
# View conflict
cat Formula/package.rb

# File shows:
# <<<<<<< HEAD
#   version "1.0.0"
# =======
#   version "2.0.0"
# >>>>>>> main

# Edit file, choose correct version or merge both
# Remove conflict markers (<<<<<<<, =======, >>>>>>>)

# Mark as resolved
git add Formula/package.rb

# If rebasing
git rebase --continue

# If merging
git commit
```

**Prevention:**
- Keep feature branches short-lived
- Pull frequently from main
- Use `${version}` to reduce version conflicts

---

### Error: `CI: test-bot: no changed formulae or casks found`

**Symptom:**
```
Error: No changed formulae or casks found
```

**Cause:**
- PR doesn't modify any formula/cask files
- Changes only to docs or config

**Solution:**
```bash
# This is expected for documentation-only PRs
# CI skips formula testing when no formulas changed

# If formula should have changed, check:
git diff main...HEAD Formula/

# If empty, formula wasn't actually committed
git add Formula/package.rb
git commit --amend
git push --force-with-lease
```

**Prevention:**
- Check `git status` before committing
- Use `git diff` to verify changes
- Run `git log -p` to see actual changes

---

## Additional Resources

### Official Homebrew Documentation

- [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook) - Complete formula reference
- [Cask Cookbook](https://docs.brew.sh/Cask-Cookbook) - Complete cask reference
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae) - What can be packaged
- [How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap) - Tap management
- [Manpage](https://docs.brew.sh/Manpage) - Complete command reference

### This Repository

- [AGENT_GUIDE.md](AGENT_GUIDE.md) - AI agent automation guide
- [FORMULA_PATTERNS.md](FORMULA_PATTERNS.md) - Formula templates
- [CASK_PATTERNS.md](CASK_PATTERNS.md) - Cask templates
- [scripts/quickstart.sh](../scripts/quickstart.sh) - Interactive formula creation

### Community Support

- [Homebrew Discussions](https://github.com/orgs/Homebrew/discussions) - Get help
- [Homebrew Slack](https://brew.sh/community) - Real-time chat

---

## Still Stuck?

If you can't resolve an issue:

1. **Search for similar issues:**
   ```bash
   gh issue list --repo Homebrew/brew --search "your error message"
   ```

2. **Check Homebrew core formulas for examples:**
   ```bash
   brew info similar-package
   brew cat similar-package
   ```

3. **Ask for help:**
   - Open GitHub Discussion: https://github.com/orgs/Homebrew/discussions
   - Include: error message, formula code, steps to reproduce

4. **Debug with verbose output:**
   ```bash
   brew install --debug --verbose --build-from-source Formula/package.rb
   ```

---

*Last updated: February 2026*
