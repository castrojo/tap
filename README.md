# Personal Homebrew Tap

Automated Homebrew tap for Linux packages with intelligent updates and quality gates.

## Features

- **Automated Updates**: Renovate checks every 3 hours, auto-merges patches
- **Quality Gates**: All packages pass `brew audit --strict` and `brew style`
- **Formulas & Casks**: CLI tools and GUI applications
- **Agent-Friendly**: Comprehensive docs for AI-assisted package creation

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

### Request a Package

[Create an issue](../../issues/new/choose) with:
- Package name
- Repository or homepage URL
- Brief description

An agent will research and create the package automatically.

## For Package Maintainers

See [docs/AGENT_GUIDE.md](docs/AGENT_GUIDE.md) for comprehensive packaging instructions.

### Quick Start

```bash
# Create formula from GitHub repository
./scripts/new-formula.sh package-name https://github.com/user/repo

# Create cask from GitHub repository
./scripts/new-cask.sh app-name https://github.com/user/repo

# Process package request from issue
./scripts/from-issue.sh 42

# Validate all packages
./scripts/validate-all.sh
```

### Documentation

- **[AGENT_GUIDE.md](docs/AGENT_GUIDE.md)** - Comprehensive guide for AI-assisted packaging
- **[FORMULA_PATTERNS.md](docs/FORMULA_PATTERNS.md)** - Formula templates for 6 build systems
- **[CASK_PATTERNS.md](docs/CASK_PATTERNS.md)** - Cask templates for 5 installation scenarios
- **[TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Solutions for common issues

## Quality Standards

All packages must:
- Pass `brew audit --strict --online`
- Pass `brew style`
- Include valid SPDX license
- Have working URLs (HTTPS preferred)
- Include meaningful tests

## Update Strategy

- **Patch releases** (1.0.0 → 1.0.1): Auto-merge after 3 hours
- **Minor releases** (1.0.0 → 1.1.0): Auto-merge after 1 day
- **Major releases** (1.0.0 → 2.0.0): Manual review required

## License

MIT
