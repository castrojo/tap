# Go Migration Plan: Replace Bash Scripts with Clean Go Code

## Executive Summary

**Goal:** Replace all bash helper scripts with idiomatic Go CLI tools for faster iteration, better testing, and improved maintainability.

**Timeline:** Phased approach, prioritize most-used scripts first

**Benefits:**
- Type safety and compile-time error checking
- Better error handling and user feedback
- Comprehensive unit testing
- Cross-platform compatibility
- Faster execution
- IDE support and better debugging

---

## Current Bash Scripts

### 1. `scripts/new-formula.sh` (Priority: HIGH)
**Purpose:** Generate Homebrew formula from GitHub repository  
**Complexity:** High - API calls, template generation, version detection  
**Lines:** ~300+  
**Dependencies:** `gh`, `jq`, `curl`

**Key Functions:**
- Fetch GitHub repository metadata
- Get latest release information
- Detect build system (Go, Rust, CMake, etc.)
- Generate formula template
- Calculate SHA256 checksums

### 2. `scripts/new-cask.sh` (Priority: HIGH)
**Purpose:** Generate Homebrew cask from GitHub repository  
**Complexity:** High - Asset selection, platform detection, checksum verification  
**Lines:** ~400+  
**Dependencies:** `gh`, `jq`, `curl`, `sha256sum`

**Key Functions:**
- Fetch GitHub repository metadata
- Filter Linux-specific assets (reject macOS/Windows)
- Prioritize tarballs > .deb packages
- Download and calculate SHA256
- Search for upstream checksums
- Verify checksum matches
- Detect desktop files and icons
- Generate cask template with desktop integration
- Automatically append `-linux` suffix

### 3. `scripts/from-issue.sh` (Priority: MEDIUM)
**Purpose:** Process package requests from GitHub issues  
**Complexity:** Medium - Issue parsing, workflow orchestration  
**Lines:** ~200  
**Dependencies:** `gh`, other scripts

**Key Functions:**
- Parse GitHub issue body
- Extract repository URL and description
- Determine package type (formula vs cask)
- Call appropriate generator script
- Create pull request

### 4. `scripts/validate-all.sh` (Priority: LOW)
**Purpose:** Validate all formulas and casks  
**Complexity:** Low - Simple loop with brew commands  
**Lines:** ~50  
**Dependencies:** `brew`

**Key Functions:**
- Find all formula and cask files
- Run `brew audit --strict --online`
- Run `brew style`
- Report errors

### 5. `scripts/update-sha256.sh` (Priority: LOW)
**Purpose:** Update checksums after version changes  
**Complexity:** Medium - Download files, calculate checksums  
**Lines:** ~150  
**Dependencies:** `curl`, `sha256sum`

**Key Functions:**
- Parse formula/cask file
- Extract URL and version
- Download file
- Calculate SHA256
- Update file in place

---

## Proposed Go Architecture

### Project Structure

```
tap-tools/
├── cmd/
│   ├── tap-formula/          # new-formula replacement
│   │   └── main.go
│   ├── tap-cask/             # new-cask replacement
│   │   └── main.go
│   ├── tap-issue/            # from-issue replacement
│   │   └── main.go
│   └── tap-validate/         # validate-all replacement
│       └── main.go
├── internal/
│   ├── github/               # GitHub API client
│   │   ├── client.go
│   │   ├── releases.go
│   │   └── repository.go
│   ├── homebrew/             # Homebrew-specific logic
│   │   ├── formula.go        # Formula generation
│   │   ├── cask.go           # Cask generation
│   │   ├── template.go       # Template rendering
│   │   └── audit.go          # Validation
│   ├── checksum/             # Checksum verification
│   │   ├── download.go
│   │   ├── sha256.go
│   │   └── upstream.go       # Upstream checksum discovery
│   ├── platform/             # Platform detection
│   │   ├── detect.go
│   │   ├── linux.go
│   │   └── priority.go       # Asset priority (tarball > deb)
│   ├── desktop/              # Desktop integration
│   │   ├── desktop_file.go
│   │   ├── icon.go
│   │   └── xdg.go
│   └── buildsystem/          # Build system detection
│       ├── detect.go
│       ├── go.go
│       ├── rust.go
│       ├── cmake.go
│       └── meson.go
├── pkg/
│   └── templates/            # Embedded templates
│       ├── formula.tmpl
│       └── cask.tmpl
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Key Dependencies

```go
// go.mod
module github.com/castrojo/tap-tools

go 1.23

require (
    github.com/google/go-github/v57 v57.0.0  // GitHub API
    github.com/spf13/cobra v1.8.0            // CLI framework
    github.com/spf13/viper v1.18.0           // Configuration
    github.com/charmbracelet/lipgloss v0.9.1 // Pretty output
    github.com/hashicorp/go-version v1.6.0   // Version parsing
    golang.org/x/oauth2 v0.15.0              // OAuth for GitHub
)
```

---

## Phase 1: Foundation (Week 1)

### Goals
- Set up Go project structure
- Implement core GitHub API client
- Implement checksum verification
- Create basic CLI framework

### Deliverables

**1. GitHub Client** (`internal/github/`)
```go
type Client struct {
    gh *github.Client
}

func (c *Client) GetRepository(owner, repo string) (*Repository, error)
func (c *Client) GetLatestRelease(owner, repo string) (*Release, error)
func (c *Client) GetReleaseAssets(owner, repo, tag string) ([]*Asset, error)
```

**2. Checksum Package** (`internal/checksum/`)
```go
func DownloadFile(url string) ([]byte, error)
func CalculateSHA256(data []byte) string
func FindUpstreamChecksum(releaseURL string) (map[string]string, error)
func VerifyChecksum(data []byte, expected string) error
```

**3. Platform Detection** (`internal/platform/`)
```go
type Asset struct {
    Name        string
    URL         string
    Platform    string  // "linux", "macos", "windows"
    Arch        string  // "x86_64", "arm64"
    Format      string  // "tar.gz", "deb", "rpm"
    Priority    int     // 1=tarball, 2=deb, 3=other
}

func DetectPlatform(filename string) (*Asset, error)
func FilterLinuxAssets(assets []*Asset) []*Asset
func SelectBestAsset(assets []*Asset) (*Asset, error)
```

**Testing:**
```bash
go test ./... -v -cover
# Target: >80% coverage
```

---

## Phase 2: Cask Generator (Week 2)

### Goals
- Implement `tap-cask` CLI tool
- Replace `scripts/new-cask.sh` completely
- Add desktop integration detection

### Deliverables

**1. Cask Generator** (`cmd/tap-cask/`)
```go
// Usage:
// tap-cask generate sublime-text https://github.com/sublimehq/sublime_text

func main() {
    rootCmd := &cobra.Command{
        Use:   "tap-cask",
        Short: "Generate Homebrew casks for Linux",
    }
    
    generateCmd := &cobra.Command{
        Use:   "generate [name] [repo-url]",
        Short: "Generate a new cask from GitHub repository",
        Args:  cobra.ExactArgs(2),
        RunE:  generateCask,
    }
    
    rootCmd.AddCommand(generateCmd)
    rootCmd.Execute()
}
```

**2. Cask Template** (`internal/homebrew/cask.go`)
```go
type CaskData struct {
    Name          string
    Token         string  // Always appends -linux
    Version       string
    SHA256        string
    URL           string
    Description   string
    Homepage      string
    License       string
    BinaryPath    string
    BinaryTarget  string
    DesktopFile   *DesktopFileInfo
    Icon          *IconInfo
    XDGPaths      []string
}

func GenerateCask(repo *github.Repository, release *github.Release) (*CaskData, error)
func RenderCask(data *CaskData) (string, error)
```

**3. Desktop Integration** (`internal/desktop/`)
```go
type DesktopFileInfo struct {
    Path         string
    FixedExecPath string
    FixedIconPath string
}

func DetectDesktopFile(tarballContents []string) (*DesktopFileInfo, error)
func DetectIcon(tarballContents []string) (*IconInfo, error)
func GeneratePreflightBlock(desktop *DesktopFileInfo, icon *IconInfo) string
```

**Features:**
- ✅ Automatic `-linux` suffix
- ✅ Linux-only asset detection (reject macOS/Windows)
- ✅ Tarball > .deb priority
- ✅ SHA256 download and verification
- ✅ Upstream checksum discovery
- ✅ Desktop file detection and path fixing
- ✅ Icon detection and installation
- ✅ XDG directory creation in preflight
- ✅ Zap trash for config/cache

**Testing:**
```bash
go test ./internal/homebrew -v
go test ./internal/desktop -v

# Integration test
tap-cask generate sublime-text https://github.com/sublimehq/sublime_text
diff Casks/sublime-text-linux.rb expected/sublime-text-linux.rb
```

---

## Phase 3: Formula Generator (Week 3)

### Goals
- Implement `tap-formula` CLI tool
- Replace `scripts/new-formula.sh`
- Add build system detection

### Deliverables

**1. Formula Generator** (`cmd/tap-formula/`)
```go
// Usage:
// tap-formula generate ripgrep https://github.com/BurntSushi/ripgrep

func main() {
    rootCmd := &cobra.Command{
        Use:   "tap-formula",
        Short: "Generate Homebrew formulas for Linux",
    }
    
    generateCmd := &cobra.Command{
        Use:   "generate [name] [repo-url]",
        Short: "Generate a new formula from GitHub repository",
        Args:  cobra.ExactArgs(2),
        RunE:  generateFormula,
    }
    
    rootCmd.AddCommand(generateCmd)
    rootCmd.Execute()
}
```

**2. Build System Detection** (`internal/buildsystem/`)
```go
type BuildSystem interface {
    Detect(repoFiles []string) bool
    GenerateInstallBlock() string
    GenerateDependencies() []string
    GenerateTestBlock() string
}

type GoBuildSystem struct{}
func (g *GoBuildSystem) Detect(files []string) bool {
    return contains(files, "go.mod") || contains(files, "go.sum")
}

type RustBuildSystem struct{}
type CMakeBuildSystem struct{}
type MesonBuildSystem struct{}

func DetectBuildSystem(repoFiles []string) (BuildSystem, error)
```

**3. Formula Template** (`internal/homebrew/formula.go`)
```go
type FormulaData struct {
    Name         string
    Version      string
    SHA256       string
    URL          string
    Description  string
    Homepage     string
    License      string
    BuildSystem  string
    Dependencies []string
    InstallBlock string
    TestBlock    string
}

func GenerateFormula(repo *github.Repository, release *github.Release) (*FormulaData, error)
func RenderFormula(data *FormulaData) (string, error)
```

**Testing:**
```bash
go test ./internal/buildsystem -v

# Integration tests for each build system
tap-formula generate ripgrep https://github.com/BurntSushi/ripgrep
tap-formula generate jq https://github.com/jqlang/jq
tap-formula generate fd https://github.com/sharkdp/fd
```

---

## Phase 4: Issue Processor (Week 4)

### Goals
- Implement `tap-issue` CLI tool
- Replace `scripts/from-issue.sh`
- Automate full workflow from issue to PR

### Deliverables

**1. Issue Processor** (`cmd/tap-issue/`)
```go
// Usage:
// tap-issue process 42
// tap-issue process 42 --create-pr

type IssueRequest struct {
    Number      int
    Type        string  // "formula" or "cask"
    RepoURL     string
    Description string
}

func ParseIssue(number int) (*IssueRequest, error)
func ProcessIssue(req *IssueRequest) error
func CreatePullRequest(req *IssueRequest, filePath string) error
```

**2. Workflow Orchestration**
```go
func ProcessIssueWorkflow(issueNumber int, createPR bool) error {
    // 1. Parse issue
    req, err := ParseIssue(issueNumber)
    
    // 2. Generate formula/cask
    if req.Type == "cask" {
        err = generateCask(req.RepoURL)
    } else {
        err = generateFormula(req.RepoURL)
    }
    
    // 3. Validate
    err = validatePackage(filePath)
    
    // 4. Create PR (if requested)
    if createPR {
        err = CreatePullRequest(req, filePath)
    }
    
    return nil
}
```

**Testing:**
```bash
# Create test issue
gh issue create --title "Package request: Test App" --body "Repository: https://github.com/test/test"

# Process it
tap-issue process <issue-number> --dry-run
```

---

## Phase 5: Validation & Polish (Week 5)

### Goals
- Implement `tap-validate` CLI tool
- Add comprehensive testing
- Create user documentation
- Performance optimization

### Deliverables

**1. Validator** (`cmd/tap-validate/`)
```go
// Usage:
// tap-validate all
// tap-validate formula/ripgrep.rb
// tap-validate --fix

func ValidateAll() error
func ValidateFile(path string) error
func FixStyle(path string) error
```

**2. Comprehensive Testing**
```bash
# Unit tests (>80% coverage)
go test ./... -v -cover -coverprofile=coverage.out

# Integration tests
make test-integration

# End-to-end tests
make test-e2e
```

**3. Performance Benchmarks**
```bash
# Benchmark cask generation
go test -bench=BenchmarkGenerateCask -benchmem

# Compare with bash script
time bash scripts/new-cask.sh test https://github.com/test/test
time tap-cask generate test https://github.com/test/test
```

**4. Documentation**
```markdown
# docs/TAP_TOOLS.md
- Installation instructions
- Usage examples
- Configuration
- Troubleshooting
```

---

## Migration Strategy

### Gradual Rollout

**Phase 1-2:** 
- Keep bash scripts as fallback
- Add Go binaries to `scripts/` directory
- Update `AGENTS.md` to mention both options

**Phase 3-4:**
- Make Go tools the default
- Add deprecation warnings to bash scripts
- Update CI to use Go tools

**Phase 5:**
- Remove bash scripts
- Update all documentation
- Archive bash scripts in `scripts/legacy/`

### Backward Compatibility

```bash
# Wrapper script: scripts/new-cask.sh
#!/usr/bin/env bash
echo "⚠️  DEPRECATED: This script will be removed in the next release."
echo "Use 'tap-cask generate' instead."
echo ""

# Call Go binary
exec tap-cask generate "$@"
```

---

## Success Metrics

### Performance
- **Cask generation:** <2 seconds (vs ~10 seconds for bash)
- **Formula generation:** <1 second (vs ~5 seconds for bash)
- **Full validation:** <30 seconds (vs ~2 minutes for bash)

### Quality
- **Test coverage:** >80% for all packages
- **Error messages:** Clear, actionable feedback
- **Documentation:** Complete API docs and examples

### User Experience
- **Installation:** Single binary, no dependencies
- **Usage:** Intuitive CLI with help text
- **Output:** Pretty, colored, progress indicators

---

## Implementation Checklist

### Phase 1: Foundation
- [ ] Initialize Go module
- [ ] Set up project structure
- [ ] Implement GitHub client
- [ ] Implement checksum package
- [ ] Implement platform detection
- [ ] Write unit tests (>80% coverage)

### Phase 2: Cask Generator
- [ ] Implement tap-cask CLI
- [ ] Add Linux asset filtering
- [ ] Add format prioritization (tarball > deb)
- [ ] Add desktop file detection
- [ ] Add icon detection
- [ ] Add preflight generation
- [ ] Add -linux suffix enforcement
- [ ] Write integration tests

### Phase 3: Formula Generator
- [ ] Implement tap-formula CLI
- [ ] Add build system detection
- [ ] Add Go build system
- [ ] Add Rust build system
- [ ] Add CMake build system
- [ ] Add Meson build system
- [ ] Write integration tests

### Phase 4: Issue Processor
- [ ] Implement tap-issue CLI
- [ ] Add issue parsing
- [ ] Add workflow orchestration
- [ ] Add PR creation
- [ ] Write integration tests

### Phase 5: Validation & Polish
- [ ] Implement tap-validate CLI
- [ ] Add comprehensive tests
- [ ] Write user documentation
- [ ] Performance optimization
- [ ] Create release binaries

---

## Next Steps

1. **Review this plan** with stakeholders
2. **Set up Go project** in `tap-tools/` subdirectory
3. **Start Phase 1** implementation
4. **Create tracking issue** with checkboxes for each phase
5. **Schedule weekly check-ins** to monitor progress

---

## Questions to Answer

1. Should tap-tools be a separate repository or subdirectory?
2. Do we need Windows/macOS support for the CLI tools themselves?
3. Should we distribute binaries via Homebrew tap?
4. Do we need a plugin system for custom build systems?
5. Should we support non-GitHub repositories (GitLab, Gitea)?

---

## Appendix: Example Usage

### Current (Bash)
```bash
./scripts/new-cask.sh sublime-text https://github.com/sublimehq/sublime_text
```

### Proposed (Go)
```bash
tap-cask generate sublime-text https://github.com/sublimehq/sublime_text

# Output:
🔍 Fetching repository metadata...
✓ Found: Sublime Text (sublimehq/sublime_text)

🔍 Finding latest release...
✓ Version: 4200

🔍 Filtering Linux assets...
✓ Found 3 Linux assets
✓ Selected: sublime_text_build_4200_x64.tar.gz (tarball - Priority 1)

⬇️  Downloading asset...
✓ Downloaded 15.2 MB

🔐 Calculating SHA256...
✓ SHA256: 36f69c551ad18ee46002be4d9c523fe545d93b67fea67beea731e724044b469f

🔍 Searching for upstream checksums...
✗ No upstream checksums found (not an error)

🖼️  Detecting desktop integration...
✓ Found desktop file: sublime_text/sublime_text.desktop
✓ Found icon: sublime_text/Icon/128x128/sublime-text.png

📝 Generating cask...
✓ Created: Casks/sublime-text-linux.rb

✅ Done! Next steps:
   1. Review Casks/sublime-text-linux.rb
   2. Test: brew install --cask castrojo/tap/sublime-text-linux
   3. Commit and push

⏱️  Completed in 1.8s
```
