# Tap Tools - Go-based Homebrew Package Generators

Go CLI tools to replace bash scripts for generating Homebrew formulas and casks for Linux.

## Status

**Phase 1: Foundation** ✅ COMPLETE
- Go module initialized
- GitHub API client implemented
- Checksum verification package
- Linux platform detection (Linux-only tap)
- Unit tests with 62% coverage

**Phase 2: Cask Generator** ✅ COMPLETE
- `tap-cask` CLI tool fully implemented
- Linux-only asset filtering (rejects macOS/Windows)
- Tarball > .deb format prioritization
- Desktop file and icon detection
- SHA256 download and verification
- Upstream checksum discovery
- XDG directory structure generation
- Automatic `-linux` suffix enforcement
- Unit tests with >90% coverage for core packages

**Phase 3: Formula Generator** ✅ COMPLETE
- `tap-formula` CLI tool fully implemented
- Build system detection (Go, Rust, CMake, Meson, Makefile)
- Automatic install block generation
- Support for pre-built binaries and source builds
- Formula template generation
- Unit tests with >89% coverage for buildsystem and formula packages

**Phase 4: Issue Processor** ✅ COMPLETE
- `tap-issue` CLI tool fully implemented
- GitHub issue parsing and metadata extraction
- Automatic package type detection (formula vs cask)
- Workflow orchestration (calls tap-formula or tap-cask)
- Git branch creation and commit automation
- PR creation and issue commenting with `--create-pr` flag
- Dry-run mode for previewing actions
- Unit tests for issue parsing

## Project Structure

```
tap-tools/
├── cmd/                    # CLI applications
│   ├── tap-formula/       # ✅ Formula generator
│   ├── tap-cask/          # ✅ Cask generator
│   ├── tap-issue/         # ✅ Issue processor
│   └── tap-validate/      # Validator (planned)
├── internal/
│   ├── github/            # ✅ GitHub API client
│   ├── checksum/          # ✅ SHA256 verification
│   ├── platform/          # ✅ Linux format detection
│   ├── homebrew/          # ✅ Formula & Cask generation
│   ├── desktop/           # ✅ Desktop integration
│   ├── buildsystem/       # ✅ Build system detection
│   └── issues/            # ✅ Issue parsing & PR creation
├── pkg/
│   └── templates/         # Embedded templates (planned)
├── go.mod
├── go.sum
└── README.md
```

## Implemented Features

### Phase 1: Foundation

#### GitHub Client (`internal/github/`)
- Parse repository URLs (owner/repo format)
- Fetch repository metadata
- Get latest release and all releases
- Extract release assets
- OAuth token support via `GITHUB_TOKEN`

#### Checksum Package (`internal/checksum/`)
- Download files from URLs
- Calculate SHA256 checksums
- Parse upstream checksum files (sha256sums.txt, etc.)
- Verify checksums against upstream

#### Platform Detection (`internal/platform/`)
- **Linux-only focus** - rejects macOS and Windows
- Detect platform from filenames
- Detect architecture (x86_64, amd64, arm64)
- Detect package formats:
  - ✅ Priority 1: Tarballs (`.tar.gz`, `.tar.xz`, `.tgz`)
  - ✅ Priority 2: Debian packages (`.deb`)
  - ✅ Priority 3: RPM, AppImage
- Filter and select best Linux assets
- Package name normalization
- Enforce `-linux` suffix for casks

### Phase 2: Cask Generator

#### `tap-cask` CLI (`cmd/tap-cask/`)
- Generate casks from GitHub repository URLs
- Pretty colored terminal output
- Detailed progress reporting

#### Desktop Integration (`internal/desktop/`)
- Detect .desktop files in extracted archives
- Detect icons (PNG, SVG)
- Fix paths in .desktop files for XDG directories
- Generate preflight blocks for directory creation

#### Homebrew Package Generation (`internal/homebrew/`)
- Generate cask templates from release data
- Generate formula templates with build system detection
- Automatic `-linux` suffix enforcement for casks
- XDG Base Directory Spec compliance
- Binary extraction from tarballs and .deb files
- Zap trash for config/cache cleanup

### Phase 3: Formula Generator

#### Build System Detection (`internal/buildsystem/`)
- Detect build systems from repository files
- Supported build systems:
  - Go (go.mod, go.sum)
  - Rust (Cargo.toml, Cargo.lock)
  - CMake (CMakeLists.txt)
  - Meson (meson.build)
  - Makefile (Makefile, makefile, GNUmakefile)
- Generate appropriate install blocks with Homebrew helpers
- Automatic dependency detection
- Test block generation

#### `tap-formula` CLI (`cmd/tap-formula/`)
- Generate formulas from GitHub repository URLs
- Automatic build system detection and install block generation
- Support for pre-built binaries and source builds
- Pretty colored terminal output
- Flags:
  - `--from-source`: Force building from source
  - `--name`: Override package name
  - `--binary`: Specify binary name
  - `--output`: Custom output path

### Phase 4: Issue Processor

#### Issues Package (`internal/issues/`)
- Parse GitHub issues for package requests
- Extract repository URL from issue body
- Extract package description (optional)
- Detect package type (formula vs cask) from keywords
- Derive package name from repository URL
- Create pull requests
- Comment on issues

#### `tap-issue` CLI (`cmd/tap-issue/`)
- Process GitHub issues to create packages automatically
- Workflow:
  1. Fetch and parse GitHub issue
  2. Extract repository URL and metadata
  3. Detect package type (formula or cask)
  4. Create git branch: `package-request-<issue>-<name>`
  5. Call appropriate generator (tap-formula or tap-cask)
  6. Commit changes with conventional commit format
  7. Push to remote
  8. Optionally create PR and comment on issue
- Flags:
  - `--create-pr`: Create pull request after generation
  - `--dry-run`: Preview actions without executing
  - `--owner`: GitHub repository owner (auto-detected)
  - `--repo`: GitHub repository name (auto-detected)

**Usage Examples:**

**Formula Generator:**
```bash
./tap-formula generate https://github.com/BurntSushi/ripgrep

# Output:
# 🔍 Parsing repository URL...
# ✓ Repository: BurntSushi/ripgrep
# 🔍 Fetching repository metadata...
# ✓ Found: Recursively search directories for a regex pattern
# 🔍 Finding latest release...
# ✓ Version: 14.0.0
# 🔍 Analyzing release assets...
# ✓ Selected: ripgrep-14.0.0-x86_64-unknown-linux-musl.tar.gz
# ⬇️  Downloading asset...
# 🔐 Calculating SHA256...
# 📝 Generating formula...
# ✅ Created: Formula/ripgrep.rb
```

**Cask Generator:**
```bash
./tap-cask generate sublime-text https://github.com/sublimehq/sublime_text

# Output:
# 🔍 Fetching repository metadata...
# ✓ Found: Sublime Text
# 🔍 Finding latest release...
# ✓ Version: 4200
# 🔍 Filtering Linux assets...
# ✓ Selected: sublime_text_build_4200_x64.tar.gz (tarball - Priority 1)
# ⬇️  Downloading asset...
# 🔐 Calculating SHA256...
# 🖼️  Detecting desktop integration...
# 📝 Generating cask...
# ✅ Created: Casks/sublime-text-linux.rb
```

**Issue Processor:**
```bash
# Preview what would happen (dry-run)
./tap-issue process 42 --dry-run

# Process issue and create package
./tap-issue process 42

# Process issue, create package, and open PR
./tap-issue process 42 --create-pr

# Output:
# ━━━ Preflight Checks ━━━
# ✓ GitHub token found
# ✓ Git repository detected
# ✓ Repository: castrojo/homebrew-tap
#
# ━━━ Fetching Issue #42 ━━━
# → Fetching issue data...
# ✓ Issue: Add ripgrep CLI tool
# → State: open
# → URL: https://github.com/castrojo/homebrew-tap/issues/42
#
# ━━━ Package Detection ━━━
# ✓ Repository URL: https://github.com/BurntSushi/ripgrep
# ✓ Package Name: ripgrep
# ✓ Package Type: formula
#
# ━━━ Creating Git Branch ━━━
# → Creating branch: package-request-42-ripgrep
# ✓ On branch: package-request-42-ripgrep
#
# ━━━ Generating Package ━━━
# → Generating formula...
# ✓ Package generated successfully
#
# ━━━ Committing Changes ━━━
# → Staging Formula/ripgrep.rb...
# → Creating commit: feat: add ripgrep formula (closes #42)
# ✓ Changes committed
#
# ━━━ Pushing to Remote ━━━
# → Pushing branch to remote...
# ✓ Branch pushed to origin/package-request-42-ripgrep
#
# ━━━ Summary ━━━
# Package Details:
#   Name:        ripgrep
#   Type:        formula
#   Repository:  https://github.com/BurntSushi/ripgrep
#   File:        Formula/ripgrep.rb
#
# Git Details:
#   Branch:      package-request-42-ripgrep
#   Commit:      feat: add ripgrep formula (closes #42)
```

## Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Current Coverage:** 
- Overall: ~80%
- `internal/platform`: 95.7%
- `internal/desktop`: 93.2%
- `internal/homebrew`: 92.3%
- `internal/buildsystem`: 89.6%
- `internal/checksum`: 35.2%
- `internal/github`: 30.9%

## Dependencies

- `github.com/google/go-github/v60` - GitHub API client
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/lipgloss` - Pretty terminal output
- `golang.org/x/oauth2` - OAuth support

## Next Steps (Phase 5)

See [../docs/GO_MIGRATION_PLAN.md](../docs/GO_MIGRATION_PLAN.md) for full plan.

**Phase 5: Validation & Polish** (planned)
- [ ] Implement `tap-validate` CLI command
- [ ] Add brew audit integration
- [ ] Validate formula/cask syntax
- [ ] Check for common issues
- [ ] Write integration tests for full workflows

## Development

```bash
# Build all commands
go build ./cmd/...

# Run tap-formula
./tap-formula generate https://github.com/user/tool

# Run tap-cask
./tap-cask generate https://github.com/user/app

# Run tap-issue (requires GITHUB_TOKEN)
export GITHUB_TOKEN=ghp_...
./tap-issue process 42
./tap-issue process 42 --dry-run
./tap-issue process 42 --create-pr

# Or run directly
go run ./cmd/tap-formula generate https://github.com/user/repo
go run ./cmd/tap-cask generate https://github.com/user/repo
go run ./cmd/tap-issue process 42

# Format code
go fmt ./...

# Lint (if golangci-lint installed)
golangci-lint run
```

## Design Principles

1. **Linux-only** - This is a Linux-specific tap, reject all macOS/Windows packages
2. **Format priority** - Prefer tarballs > .deb > other formats
3. **SHA256 mandatory** - Every package must have verified checksum
4. **XDG compliance** - All installations to user home directory
5. **Type safety** - Leverage Go's type system for correctness
6. **Testability** - Comprehensive unit tests for all packages

## Contributing

This is part of the castrojo/homebrew-tap repository migration from bash to Go.
See the main repository README for contribution guidelines.
