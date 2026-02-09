# Personal Homebrew Tap for Linux

Automated Homebrew tap with intelligent package generation, quality gates, and AI agent support.

## Features

- **Fast Package Generation**: Go CLI tools generate packages from GitHub releases (4-5x faster than bash)
- **Automated Updates**: Renovate checks every 3 hours with automatic SHA256 verification
- **Quality Gates**: All packages pass `brew audit --strict` and `brew style`
- **Formulas & Casks**: CLI tools and GUI applications with desktop integration
- **AI Agent Support**: Comprehensive documentation for Copilot and other coding agents
- **XDG Compliant**: All installations respect user home directory structure

## Installation

```bash
brew tap castrojo/tap
```

## Usage

### Install a Package
```bash
brew install package-name           # Formula (CLI tool)
brew install --cask app-name        # Cask (GUI app)
```

### Available Packages

#### GUI Applications (Casks)

**Sublime Text** - Sophisticated text editor for code, markup and prose
```bash
brew install --cask castrojo/tap/sublime-text-linux
```
Launch from command line: `subl` | Desktop launcher: Available in application menu

**Quarto** - Open-source scientific and technical publishing system
```bash
brew install --cask castrojo/tap/quarto-linux
```
Launch from command line: `quarto`

#### CLI Tools (Formulas)
*No formulas yet - contributions welcome!*

### Request a Package

[Create an issue](../../issues/new/choose) with the repository or homepage URL. An AI agent will automatically:
1. Research the package and select the appropriate Linux binary
2. Generate the formula/cask with proper XDG paths
3. Create a pull request for review

Package names are automatically derived from repository names (e.g., `user/my-app` → `my-app-linux` for casks).

## For Package Maintainers & Contributors

### First-Time Setup (Required)

**After cloning the repository, install git hooks:**

```bash
./scripts/setup-hooks.sh
```

This installs a pre-commit hook that:
- ✅ Automatically validates all Ruby files before commit
- ✅ Auto-fixes style issues with RuboCop
- ✅ Prevents committing invalid code that would fail CI
- ✅ Saves time by catching issues locally

**The hook is mandatory for all contributors.**

### Quick Start with Go CLI Tools (Recommended)

**Generate from GitHub releases:**
```bash
# GUI application (cask)
./tap-tools/tap-cask generate https://github.com/user/repo

# CLI tool (formula)
./tap-tools/tap-formula generate https://github.com/user/repo

# From GitHub issue
./tap-tools/tap-issue process 42
./tap-tools/tap-issue process 42 --create-pr

# Validate packages
./tap-tools/tap-validate all
./tap-tools/tap-validate all --fix  # Auto-fix style issues
./tap-tools/tap-validate file Casks/app-name-linux.rb
```

**Features:**
- ✅ Automatically selects Linux-only assets (rejects macOS/Windows)
- ✅ Prioritizes formats: tarball > deb > other
- ✅ Calculates and verifies SHA256 checksums
- ✅ Detects desktop integration needs (.desktop files, icons)
- ✅ Ensures XDG Base Directory Spec compliance

See [tap-tools/README.md](tap-tools/README.md) for detailed documentation.

### Documentation

**For AI Agents & Copilot:**
- **[.github/copilot-instructions.md](.github/copilot-instructions.md)** - Comprehensive agent instructions (auto-loaded by GitHub Copilot)
- **[AGENTS.md](AGENTS.md)** - Repository-specific agent guidance and constraints

**For Developers:**
- **[docs/AGENT_GUIDE.md](docs/AGENT_GUIDE.md)** - Comprehensive packaging guide
- **[docs/CASK_CREATION_GUIDE.md](docs/CASK_CREATION_GUIDE.md)** - **CRITICAL** - Read before creating casks (prevents CI failures)
- **[docs/FORMULA_PATTERNS.md](docs/FORMULA_PATTERNS.md)** - Copy-paste templates for 6 build systems
- **[docs/CASK_PATTERNS.md](docs/CASK_PATTERNS.md)** - Copy-paste templates for 5 installation scenarios
- **[docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Common errors and solutions
- **[tap-tools/README.md](tap-tools/README.md)** - Go CLI tools documentation

## Critical Requirements

### Linux-Only & Read-Only Filesystem

**All packages MUST:**
- ✅ Use Linux binaries ONLY (never macOS `.dmg`/`.pkg` or Windows `.exe`/`.msi`)
- ✅ Install to user home directory (NEVER to `/usr/`, `/opt/`)
- ✅ Use XDG environment variables (`$XDG_DATA_HOME`, `$XDG_CONFIG_HOME`, `$XDG_CACHE_HOME`)
- ✅ Include desktop integration for GUI apps (`.desktop` file + icon)
- ✅ Include SHA256 verification (MANDATORY)
- ✅ Use `-linux` suffix for cask names (e.g., `app-name-linux`)

**Why:** Target systems use read-only root filesystems (Fedora Silverblue, Universal Blue) where `/usr/` and `/opt/`, are read-only.

### Package Format Priority

1. **Tarball (PREFERRED)** - `.tar.gz`, `.tar.xz`, `.tgz` - Most portable
2. **Debian Package** - `.deb` - Use only if no tarball available
3. **Other formats** - AppImage, snap, flatpak - Case-by-case basis

## Quality Standards & CI

All packages must pass:
- ✅ `brew audit --cask --strict --online` or `brew audit --strict --online`
- ✅ `brew style` (RuboCop linting)
- ✅ SHA256 verification
- ✅ Valid SPDX license identifier
- ✅ Working HTTPS URLs
- ✅ XDG Base Directory Spec compliance

**Automated Checks:**
- GitHub Actions run on every PR
- Renovate updates packages every 3 hours
- Auto-merge policy: patches (3h), minors (1 day), majors (manual review)

## Contributing

**For major features/changes:**
1. Create a feature branch
2. Make changes and validate with `tap-validate`
3. Create pull request (enables Gemini Code Assist review)
4. Wait for CI checks and code review
5. Merge after approval

**For AI agents:** Read `.github/copilot-instructions.md` and `AGENTS.md` before starting work.

**Commit format:** Use [Conventional Commits](https://www.conventionalcommits.org/) with AI attribution:
```
feat(cask): add app-name-linux v1.0.0

Brief description of changes.

Assisted-by: <Model> via <Tool>
```

## Update Strategy

- **Patch releases** (1.0.0 → 1.0.1): Auto-merge after 3 hours
- **Minor releases** (1.0.0 → 1.1.0): Auto-merge after 3 days
- **Major releases** (1.0.0 → 2.0.0): Manual review required

## License

Apache 2.0
