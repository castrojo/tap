# Agent Instructions

⚠️ **MANDATORY: LOAD THE PACKAGING SKILL FIRST** ⚠️

**CRITICAL: Before doing ANY package-related work (creating, updating, or debugging packages), you MUST:**

1. **Load the packaging skill:** Read `.github/skills/homebrew-packaging/SKILL.md`
2. **Follow its workflow exactly:** The skill contains the mandatory 6-step process
3. **Complete ALL checkpoints:** Especially validation before every commit

**When to load the skill:**
- ✅ User mentions: "add", "create", "package", "cask", "formula", "update", "fix"
- ✅ User provides a GitHub release URL or repository link
- ✅ User references `Casks/` or `Formula/` directories
- ✅ User asks about tap-tools, tap-cask, or tap-formula
- ✅ User assigns you a GitHub issue about packages
- ✅ **ANY** work involving Homebrew packages

**The skill is the authoritative source. If this file conflicts with the skill, follow the skill.**

---

**⚠️ LINUX ONLY REPOSITORY ⚠️**

**THIS TAP IS LINUX-ONLY. ALL PACKAGES MUST USE LINUX BINARIES.**
- ✓ Use Linux downloads (e.g., `app-linux-x64.tar.gz`, `tool_linux_amd64`)
- ✗ NEVER use macOS downloads (`.dmg`, `.pkg`, `-darwin-`, `-macos-`)
- ✗ NEVER use Windows downloads (`.exe`, `.msi`, `-windows-`)

**⚠️ READ-ONLY FILESYSTEM CONSTRAINT ⚠️**

Target systems use **immutable/read-only root filesystems** (Fedora Silverblue, Universal Blue, etc.).

**CRITICAL: ALL files MUST be installed to user home directory:**
- ✓ `~/.local/share/applications/` - Desktop files (GUI launcher integration)
- ✓ `~/.local/share/icons/` - Application icons
- ✓ `~/.config/` - Configuration files
- ✓ `~/.cache/` - Cache data
- ✗ NEVER install to `/usr/`, `/opt/`, `/etc/` (filesystem is read-only!)

**For GUI applications:**
- MUST install `.desktop` file to `~/.local/share/applications/`
- MUST install icon to `~/.local/share/icons/`
- MUST fix paths in `.desktop` file during `preflight`

See [docs/CASK_CREATION_GUIDE.md](docs/CASK_CREATION_GUIDE.md) for detailed desktop integration patterns.

## Package Format Priority

When packaging software, use this strict priority order:

**1. Tarball (PREFERRED)** - `.tar.gz`, `.tar.xz`, `.tgz`
  - Most portable, works across all Linux distributions
  - Simple extraction and installation
  - No package manager dependencies
  - Example: `app-linux-x64.tar.gz`

**2. Debian Package (SECOND CHOICE)** - `.deb`
  - Use only if no tarball is available
  - Requires extraction via `ar` and `tar`
  - May have distro-specific dependencies
  - Example: `app_amd64.deb`

**3. Other formats** - Only with explicit justification
  - AppImage, snap, flatpak: Case-by-case basis
  - RPM: Generally avoid (requires conversion)

## Checksum Verification (MANDATORY)

**EVERY package MUST have SHA256 verification:**

```ruby
cask "app-name" do
  version "1.0.0"
  sha256 "abc123..."  # REQUIRED - calculated from downloaded file
  
  url "https://example.com/app-linux-x64.tar.gz"
  # ...
end
```

**Verification steps:**
1. Download the file: `curl -LO <url>`
2. Calculate SHA256: `sha256sum <file>`
3. Compare with upstream checksums (if provided)
4. Use the calculated hash in the cask

**NEVER:**
- Skip SHA256 verification
- Use `sha256 :no_check` unless absolutely necessary (requires justification)
- Copy SHA256 from unreliable sources without verification

Read [docs/CASK_CREATION_GUIDE.md](docs/CASK_CREATION_GUIDE.md) for detailed cask creation rules.

---

## Pull Request Policy (MANDATORY)

**CRITICAL: All major features/epics MUST be developed in a pull request, NOT committed directly to main.**

**Why:**
- Enables Gemini Code Assist reviews for quality assurance
- Provides code review feedback before merging
- Creates discussion space for design decisions
- Allows CI to validate changes before they land

**When to Create a Pull Request:**
- ✓ New features or functionality
- ✓ Major refactoring or architectural changes
- ✓ Multi-file changes spanning multiple components
- ✓ Changes affecting critical workflows or patterns
- ✓ Updates to Go CLI tools or core infrastructure
- ✓ Documentation overhauls or new guides

**When Direct Commits are Acceptable:**
- Small bug fixes (single file, < 10 lines)
- Typo corrections in documentation
- Version bumps for existing packages
- CI configuration tweaks (after testing)

**Pull Request Workflow:**

1. **Create Feature Branch:**
   ```bash
   git checkout -b feature/descriptive-name
   # OR for fixes:
   git checkout -b fix/issue-description
   ```

2. **Make Changes and Commit:**
   ```bash
   git add -A
   git commit -m "feat(component): add new feature
   
   Detailed description of changes.
   
   Assisted-by: [Model] via [Tool]"
   ```

3. **Create PR (automatically pushes):**
   ```bash
   # Use gh pr create - it pushes and creates PR in one command
   gh pr create --title "feat(component): add new feature" --body "$(cat <<'EOF'
   ## Summary
   - High-level overview of changes
   - Why this change is needed
   - What problems it solves
   
   ## Changes Made
   - Specific file/component changes
   - New functionality added
   - Tests/documentation updated
   
   ## Testing
   - How changes were validated
   - Test commands run and their results
   EOF
   )"
   ```

4. **Wait for Gemini Code Review:**
   - Gemini Code Assist will automatically review the PR
   - Address any feedback or suggestions
   - Make additional commits to the branch as needed

5. **Merge After Approval:**
   ```bash
   gh pr merge --squash  # Preferred - squashes commits
   # OR
   gh pr merge --merge   # Merge commit
   ```

**Benefits of This Workflow:**
- Gemini provides expert code review feedback
- Catches issues before they reach main branch
- Improves code quality and maintainability
- Creates clear history of feature development
- Enables better collaboration and knowledge sharing

---

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until changes are pushed to remote.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   # For direct commits to main (small fixes, non-workflow files):
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   
   # For feature branches / PRs:
   gh pr create --title "..." --body "..."  # Automatically pushes
   # OR if PR already exists:
   git push  # Push additional commits to existing PR
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until changes are pushed to remote
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- Use `gh pr create` for new PRs (automatically pushes)
- Use `git push` only for: direct main commits OR additional commits to existing PRs
- If push fails, resolve and retry until it succeeds

## Conventional Commits

This repository enforces [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#specification) for all commits and pull request titles.

**Format:**
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only changes
- `style:` - Code style changes (formatting, missing semi-colons, etc)
- `refactor:` - Code change that neither fixes a bug nor adds a feature
- `perf:` - Performance improvement
- `test:` - Adding or updating tests
- `build:` - Changes to build system or dependencies
- `ci:` - Changes to CI configuration files and scripts
- `chore:` - Other changes that don't modify src or test files
- `revert:` - Reverts a previous commit

**Examples:**
```
feat(cask): add sublime-text cask
fix(workflow): add --cask flag for cask audits
docs(cask): add critical cask creation guide
ci(workflow): tap repository before auditing
```

**Breaking Changes:**
Add `!` after type or `BREAKING CHANGE:` in footer:
```
feat(api)!: remove deprecated endpoints

BREAKING CHANGE: The /v1/users endpoint has been removed.
```

## Attribution Requirements

AI agents must disclose what tool and model they are using in the "Assisted-by" commit footer:

```text
Assisted-by: [Model Name] via [Tool Name]
```

**Examples:**

```text
feat(cask): add firefox cask for Linux

Adds Firefox browser cask with proper binary extraction.

Assisted-by: Claude 3.5 Sonnet via GitHub Copilot
```

```text
fix(workflow): correct brew audit command syntax

Homebrew no longer accepts file paths, must use tap names.

Assisted-by: GPT-4 via Cursor
```

```text
docs: update cask creation guide with depends_on rules

Documented why depends_on :linux fails and correct alternatives.

Assisted-by: Claude 3.5 Sonnet via OpenCode
```

## Documentation

All documentation is located in the `docs/` directory:

- **[AGENT_GUIDE.md](docs/AGENT_GUIDE.md)** - Comprehensive guide for agents creating packages
- **[CASK_CREATION_GUIDE.md](docs/CASK_CREATION_GUIDE.md)** - Critical cask rules (READ THIS FIRST before creating casks!)
- **[FORMULA_PATTERNS.md](docs/FORMULA_PATTERNS.md)** - Copy-paste templates for formulas
- **[CASK_PATTERNS.md](docs/CASK_PATTERNS.md)** - Copy-paste templates for casks
- **[TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Common errors and solutions

## Helper Scripts & Go CLI Tools

### Go CLI Tools (RECOMMENDED - Faster & More Reliable)

Located in `tap-tools/`, pre-built binaries available:

- **`tap-cask`** - Generate new cask from GitHub releases (5.5x faster than bash)
  ```bash
  ./tap-tools/tap-cask generate <name> <github-url>
  ```

- **`tap-formula`** - Generate new formula from GitHub releases (4.2x faster than bash)
  ```bash
  ./tap-tools/tap-formula generate <name> <github-url>
  ```

- **`tap-issue`** - Process package requests from GitHub issues
  ```bash
  ./tap-tools/tap-issue process <issue-number>
  ./tap-tools/tap-issue process <issue-number> --create-pr
  ```

- **`tap-validate`** - Validate all formulas and casks (4x faster than bash)
  ```bash
  ./tap-tools/tap-validate all
  ./tap-tools/tap-validate all --fix  # Auto-fix style issues
  ./tap-tools/tap-validate file Formula/jq.rb
  ```

**Features:**
- ✅ Automatic build system detection (Go, Rust, CMake, Meson)
- ✅ Linux-only asset filtering (rejects macOS/Windows automatically)
- ✅ Format prioritization (tarball > deb > other)
- ✅ Desktop integration detection (icons, .desktop files)
- ✅ SHA256 verification and upstream checksum discovery
- ✅ XDG Base Directory Spec compliance
- ✅ **Automatic validation with brew audit and brew style**
- ✅ Pretty terminal output with progress indicators

See [tap-tools/README.md](tap-tools/README.md) for complete documentation.

## Best Practices

### First-Time Setup

**Install git hooks after cloning:**
```bash
./scripts/setup-hooks.sh
```

This installs a pre-commit hook that:
- Automatically validates all Ruby files before commit
- Auto-fixes style issues with `--fix` flag
- Blocks commits that fail validation
- Prevents CI failures from style issues

**The hook is mandatory for all contributors.** If the hook blocks a commit, fix the issues and try again. Do not bypass with `--no-verify` unless absolutely necessary.

### Code Style
- Be surgical with minimal code changes
- Follow existing patterns in the codebase
- Pre-commit hook automatically validates (installed via `./scripts/setup-hooks.sh`)
- Manual validation: `./tap-tools/tap-validate all --fix`
- Keep formulas and casks simple and readable

### Package Naming
Package names are **always derived from the repository name**, never manually specified:
- Repository: `BurntSushi/ripgrep` → Package: `ripgrep`
- Repository: `user/My_Cool_App` → Package: `my-cool-app`

**Linux Cask Naming (Required):**
- ALL casks MUST use `-linux` suffix
- Example: `sublime-text-linux`, `jetbrains-toolbox-linux`
- Prevents collision with official macOS casks
- The `tap-cask` tool automatically appends `-linux` if not present

**Formula naming:**
- No suffix required for formulas (CLI tools)

### Commit Workflow
1. Make changes
2. Run `./tap-tools/tap-validate all` (if needed - generators validate automatically)
3. Commit with conventional commit format
4. Add `Assisted-by:` footer with model and tool
5. Push to remote

### Pull Request Titles
PR titles must also follow conventional commit format:
```
feat(cask): add new application
fix(formula): correct version detection
docs: improve installation instructions
```

## CI/CD

GitHub Actions automatically:
- Runs `brew audit` and `brew style` on all changed formulas/casks
- Labels PRs based on changes
- Updates dependencies via Renovate (every 3 hours)

All checks must pass before merging.
