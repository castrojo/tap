# Agent Instructions

**⚠️ LINUX ONLY REPOSITORY ⚠️**

**THIS TAP IS LINUX-ONLY. ALL PACKAGES MUST USE LINUX BINARIES.**
- ✓ Use Linux downloads (e.g., `app-linux-x64.tar.gz`, `tool_linux_amd64`)
- ✗ NEVER use macOS downloads (`.dmg`, `.pkg`, `-darwin-`, `-macos-`)
- ✗ NEVER use Windows downloads (`.exe`, `.msi`, `-windows-`)

Read [docs/CASK_CREATION_GUIDE.md](docs/CASK_CREATION_GUIDE.md) before creating any casks.

---

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
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

## Helper Scripts

Located in `scripts/`:

- `new-formula.sh` - Generate new formula from GitHub releases
- `new-cask.sh` - Generate new cask from GitHub releases
- `from-issue.sh` - Process package requests from GitHub issues
- `validate-all.sh` - Validate all formulas and casks
- `update-sha256.sh` - Update checksums after version changes

## Best Practices

### Code Style
- Be surgical with minimal code changes
- Follow existing patterns in the codebase
- Run `scripts/validate-all.sh` before committing
- Keep formulas and casks simple and readable

### Package Naming
Package names are **always derived from the repository name**, never manually specified:
- Repository: `BurntSushi/ripgrep` → Package: `ripgrep`
- Repository: `user/My_Cool_App` → Package: `my-cool-app`

### Commit Workflow
1. Make changes
2. Run `scripts/validate-all.sh` (if available)
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
