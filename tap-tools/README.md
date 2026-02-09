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

## Project Structure

```
tap-tools/
├── cmd/                    # CLI applications
│   ├── tap-formula/       # Formula generator (planned)
│   ├── tap-cask/          # ✅ Cask generator
│   ├── tap-issue/         # Issue processor (planned)
│   └── tap-validate/      # Validator (planned)
├── internal/
│   ├── github/            # ✅ GitHub API client
│   ├── checksum/          # ✅ SHA256 verification
│   ├── platform/          # ✅ Linux format detection
│   ├── homebrew/          # ✅ Cask generation
│   ├── desktop/           # ✅ Desktop integration
│   └── buildsystem/       # Build system detection (planned)
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

#### Homebrew Cask Generation (`internal/homebrew/`)
- Generate cask templates from release data
- Automatic `-linux` suffix enforcement
- XDG Base Directory Spec compliance
- Binary extraction from tarballs and .deb files
- Zap trash for config/cache cleanup

**Usage Example:**
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
- Overall: ~70%
- `internal/platform`: 95.7%
- `internal/desktop`: 93.2%
- `internal/homebrew`: 91.7%
- `internal/checksum`: 35.2%
- `internal/github`: 30.9%

## Dependencies

- `github.com/google/go-github/v60` - GitHub API client
- `github.com/spf13/cobra` - CLI framework (ready for use)
- `github.com/charmbracelet/lipgloss` - Pretty terminal output (ready for use)
- `github.com/hashicorp/go-version` - Version parsing (ready for use)
- `golang.org/x/oauth2` - OAuth support

## Next Steps (Phase 3)

See [../docs/GO_MIGRATION_PLAN.md](../docs/GO_MIGRATION_PLAN.md) for full plan.

**Phase 3: Formula Generator** (in progress)
- [ ] Implement build system detection package (`internal/buildsystem/`)
  - [ ] Go build system
  - [ ] Rust build system
  - [ ] CMake build system
  - [ ] Meson build system
- [ ] Implement formula generation (`internal/homebrew/formula.go`)
- [ ] Implement `tap-formula` CLI command
- [ ] Write integration tests

## Development

```bash
# Build all commands
go build ./cmd/...

# Run tap-cask
./tap-cask generate sublime-text https://github.com/sublimehq/sublime_text

# Or run directly
go run ./cmd/tap-cask generate app-name https://github.com/user/repo

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
