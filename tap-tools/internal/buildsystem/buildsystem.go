// Package buildsystem provides build system detection and code generation
// for Homebrew formulas. It detects common build systems (Go, Rust, CMake,
// Meson, etc.) and generates appropriate install blocks.
package buildsystem

import (
	"fmt"
	"strings"
)

// BuildSystem represents a detected build system with methods to generate
// Homebrew formula code blocks.
type BuildSystem interface {
	// Name returns the human-readable name of the build system
	Name() string

	// Detect returns true if this build system is present in the repository
	Detect(files []string) bool

	// GenerateInstallBlock returns Ruby code for the install block
	GenerateInstallBlock(opts InstallOptions) string

	// GenerateDependencies returns formula dependencies needed for building
	GenerateDependencies() []string

	// GenerateTestBlock returns Ruby code for testing the installed formula
	GenerateTestBlock(binaryName string) string
}

// InstallOptions contains information needed to generate install blocks
type InstallOptions struct {
	// BinaryName is the name of the main executable to install
	BinaryName string

	// Prefix is the installation prefix (usually "#{prefix}")
	Prefix string

	// MultipleOutputs indicates if the build produces multiple binaries/libs
	MultipleOutputs bool

	// LDFlags are additional linker flags (for Go builds)
	LDFlags []string
}

// Detect analyzes a list of repository files and returns the detected
// build system, or nil if none is detected.
func Detect(files []string) BuildSystem {
	// Try build systems in order of specificity
	systems := []BuildSystem{
		&GoBuildSystem{},
		&RustBuildSystem{},
		&MesonBuildSystem{},
		&CMakeBuildSystem{},
		&MakefileBuildSystem{},
	}

	for _, sys := range systems {
		if sys.Detect(files) {
			return sys
		}
	}

	return nil
}

// containsFile checks if a filename exists in the list
func containsFile(files []string, target string) bool {
	for _, f := range files {
		if strings.HasSuffix(f, target) {
			return true
		}
	}
	return false
}

// containsAnyFile checks if any of the target files exist in the list
func containsAnyFile(files []string, targets []string) bool {
	for _, target := range targets {
		if containsFile(files, target) {
			return true
		}
	}
	return false
}

// GoBuildSystem represents a Go-based project
type GoBuildSystem struct{}

func (g *GoBuildSystem) Name() string {
	return "Go"
}

func (g *GoBuildSystem) Detect(files []string) bool {
	return containsFile(files, "go.mod") || containsFile(files, "go.sum")
}

func (g *GoBuildSystem) GenerateInstallBlock(opts InstallOptions) string {
	var b strings.Builder

	b.WriteString("def install\n")

	if len(opts.LDFlags) > 0 {
		b.WriteString(fmt.Sprintf("    ldflags = %s\n", formatLDFlags(opts.LDFlags)))
		b.WriteString("    system \"go\", \"build\", *std_go_args(ldflags: ldflags)\n")
	} else {
		b.WriteString("    system \"go\", \"build\", *std_go_args\n")
	}

	if opts.MultipleOutputs {
		b.WriteString("    # TODO: Install additional binaries if present\n")
		b.WriteString("    # bin.install Dir[\"bin/*\"]\n")
	}

	b.WriteString("  end")

	return b.String()
}

func (g *GoBuildSystem) GenerateDependencies() []string {
	return []string{"go"}
}

func (g *GoBuildSystem) GenerateTestBlock(binaryName string) string {
	return fmt.Sprintf("test do\n    system \"#{bin}/%s\", \"--version\"\n  end", binaryName)
}

// formatLDFlags formats a list of linker flags for Go build
func formatLDFlags(flags []string) string {
	quoted := make([]string, len(flags))
	for i, flag := range flags {
		quoted[i] = fmt.Sprintf("\"%s\"", flag)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// RustBuildSystem represents a Rust/Cargo project
type RustBuildSystem struct{}

func (r *RustBuildSystem) Name() string {
	return "Rust"
}

func (r *RustBuildSystem) Detect(files []string) bool {
	return containsFile(files, "Cargo.toml") && containsFile(files, "Cargo.lock")
}

func (r *RustBuildSystem) GenerateInstallBlock(opts InstallOptions) string {
	var b strings.Builder

	b.WriteString("def install\n")
	b.WriteString("    system \"cargo\", \"install\", *std_cargo_args\n")

	if opts.MultipleOutputs {
		b.WriteString("    # TODO: Install additional binaries if present\n")
		b.WriteString("    # bin.install Dir[\"target/release/other-binary\"]\n")
	}

	b.WriteString("  end")

	return b.String()
}

func (r *RustBuildSystem) GenerateDependencies() []string {
	return []string{"rust"}
}

func (r *RustBuildSystem) GenerateTestBlock(binaryName string) string {
	return fmt.Sprintf("test do\n    system \"#{bin}/%s\", \"--version\"\n  end", binaryName)
}

// CMakeBuildSystem represents a CMake-based project
type CMakeBuildSystem struct{}

func (c *CMakeBuildSystem) Name() string {
	return "CMake"
}

func (c *CMakeBuildSystem) Detect(files []string) bool {
	return containsFile(files, "CMakeLists.txt")
}

func (c *CMakeBuildSystem) GenerateInstallBlock(opts InstallOptions) string {
	var b strings.Builder

	b.WriteString("def install\n")
	b.WriteString(fmt.Sprintf("    system \"cmake\", \"-S\", \".\", \"-B\", \"build\", *std_cmake_args\n"))
	b.WriteString("    system \"cmake\", \"--build\", \"build\"\n")
	b.WriteString("    system \"cmake\", \"--install\", \"build\"\n")
	b.WriteString("  end")

	return b.String()
}

func (c *CMakeBuildSystem) GenerateDependencies() []string {
	return []string{"cmake"}
}

func (c *CMakeBuildSystem) GenerateTestBlock(binaryName string) string {
	return fmt.Sprintf("test do\n    system \"#{bin}/%s\", \"--version\"\n  end", binaryName)
}

// MesonBuildSystem represents a Meson-based project
type MesonBuildSystem struct{}

func (m *MesonBuildSystem) Name() string {
	return "Meson"
}

func (m *MesonBuildSystem) Detect(files []string) bool {
	return containsFile(files, "meson.build")
}

func (m *MesonBuildSystem) GenerateInstallBlock(opts InstallOptions) string {
	var b strings.Builder

	b.WriteString("def install\n")
	b.WriteString("    system \"meson\", \"setup\", \"build\", *std_meson_args\n")
	b.WriteString("    system \"meson\", \"compile\", \"-C\", \"build\", \"--verbose\"\n")
	b.WriteString("    system \"meson\", \"install\", \"-C\", \"build\"\n")
	b.WriteString("  end")

	return b.String()
}

func (m *MesonBuildSystem) GenerateDependencies() []string {
	return []string{"meson", "ninja"}
}

func (m *MesonBuildSystem) GenerateTestBlock(binaryName string) string {
	return fmt.Sprintf("test do\n    system \"#{bin}/%s\", \"--version\"\n  end", binaryName)
}

// MakefileBuildSystem represents a traditional Makefile-based project
type MakefileBuildSystem struct{}

func (mk *MakefileBuildSystem) Name() string {
	return "Makefile"
}

func (mk *MakefileBuildSystem) Detect(files []string) bool {
	return containsAnyFile(files, []string{"Makefile", "makefile", "GNUmakefile"})
}

func (mk *MakefileBuildSystem) GenerateInstallBlock(opts InstallOptions) string {
	var b strings.Builder

	b.WriteString("def install\n")
	b.WriteString("    # TODO: Check if configure script exists and run it\n")
	b.WriteString("    # system \"./configure\", \"--prefix=#{prefix}\"\n")
	b.WriteString("    system \"make\", \"install\", \"PREFIX=#{prefix}\"\n")
	b.WriteString("  end")

	return b.String()
}

func (mk *MakefileBuildSystem) GenerateDependencies() []string {
	return []string{}
}

func (mk *MakefileBuildSystem) GenerateTestBlock(binaryName string) string {
	return fmt.Sprintf("test do\n    system \"#{bin}/%s\", \"--version\"\n  end", binaryName)
}
