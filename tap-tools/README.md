# Tap Tools - Go-based Homebrew Package Generators

Go CLI tools to replace bash scripts for generating Homebrew formulas and casks for Linux.

## Status

**Phase 1: Foundation** ✅ COMPLETE
- Go module initialized
- GitHub API client implemented
- Checksum verification package
- Linux platform detection (Linux-only tap)
- Unit tests with 62% coverage

## Project Structure

```
tap-tools/
├── cmd/                    # CLI applications
│   ├── tap-formula/       # Formula generator (planned)
│   ├── tap-cask/          # Cask generator (planned)
│   ├── tap-issue/         # Issue processor (planned)
│   └── tap-validate/      # Validator (planned)
├── internal/
│   ├── github/            # ✅ GitHub API client
│   ├── checksum/          # ✅ SHA256 verification
│   ├── platform/          # ✅ Linux format detection
│   ├── homebrew/          # Formula/cask generation (planned)
│   ├── desktop/           # Desktop integration (planned)
│   └── buildsystem/       # Build system detection (planned)
├── pkg/
│   └── templates/         # Embedded templates (planned)
├── go.mod
├── go.sum
└── README.md
```

## Implemented Features (Phase 1)

### GitHub Client (`internal/github/`)
- Parse repository URLs (owner/repo format)
- Fetch repository metadata
- Get latest release and all releases
- Extract release assets
- OAuth token support via `GITHUB_TOKEN`

### Checksum Package (`internal/checksum/`)
- Download files from URLs
- Calculate SHA256 checksums
- Parse upstream checksum files (sha256sums.txt, etc.)
- Verify checksums against upstream

### Platform Detection (`internal/platform/`)
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

**Current Coverage:** 62% (Target: >80% for Phase 1)

## Dependencies

- `github.com/google/go-github/v60` - GitHub API client
- `github.com/spf13/cobra` - CLI framework (ready for use)
- `github.com/charmbracelet/lipgloss` - Pretty terminal output (ready for use)
- `github.com/hashicorp/go-version` - Version parsing (ready for use)
- `golang.org/x/oauth2` - OAuth support

## Next Steps (Phase 2)

See [../docs/GO_MIGRATION_PLAN.md](../docs/GO_MIGRATION_PLAN.md) for full plan.

**Phase 2: Cask Generator** (planned)
- [ ] Implement `tap-cask` CLI command
- [ ] Add Linux asset filtering
- [ ] Add desktop file detection
- [ ] Add icon detection
- [ ] Generate cask templates
- [ ] Integration tests

## Development

```bash
# Build all commands (when implemented)
go build ./cmd/...

# Run a command (when implemented)
go run ./cmd/tap-cask generate sublime-text https://github.com/sublimehq/sublime_text

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
