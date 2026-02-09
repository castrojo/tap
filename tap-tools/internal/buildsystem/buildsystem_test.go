package buildsystem

import (
	"strings"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string // Build system name, or "" for nil
	}{
		{
			name:     "Go project with go.mod",
			files:    []string{"main.go", "go.mod", "README.md"},
			expected: "Go",
		},
		{
			name:     "Go project with go.sum",
			files:    []string{"main.go", "go.sum"},
			expected: "Go",
		},
		{
			name:     "Rust project",
			files:    []string{"src/main.rs", "Cargo.toml", "Cargo.lock"},
			expected: "Rust",
		},
		{
			name:     "CMake project",
			files:    []string{"src/main.c", "CMakeLists.txt"},
			expected: "CMake",
		},
		{
			name:     "Meson project",
			files:    []string{"src/main.c", "meson.build"},
			expected: "Meson",
		},
		{
			name:     "Makefile project",
			files:    []string{"main.c", "Makefile"},
			expected: "Makefile",
		},
		{
			name:     "No build system",
			files:    []string{"README.md", "LICENSE"},
			expected: "",
		},
		{
			name:     "Go takes priority over Makefile",
			files:    []string{"main.go", "go.mod", "Makefile"},
			expected: "Go",
		},
		{
			name:     "Rust takes priority over Makefile",
			files:    []string{"src/main.rs", "Cargo.toml", "Cargo.lock", "Makefile"},
			expected: "Rust",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Detect(tt.files)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected no build system, got %s", result.Name())
				}
			} else {
				if result == nil {
					t.Errorf("Expected %s, got nil", tt.expected)
				} else if result.Name() != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result.Name())
				}
			}
		})
	}
}

func TestGoBuildSystem(t *testing.T) {
	bs := &GoBuildSystem{}

	t.Run("Name", func(t *testing.T) {
		if bs.Name() != "Go" {
			t.Errorf("Expected name 'Go', got %s", bs.Name())
		}
	})

	t.Run("Detect with go.mod", func(t *testing.T) {
		files := []string{"main.go", "go.mod"}
		if !bs.Detect(files) {
			t.Error("Expected to detect Go project with go.mod")
		}
	})

	t.Run("Detect with go.sum", func(t *testing.T) {
		files := []string{"main.go", "go.sum"}
		if !bs.Detect(files) {
			t.Error("Expected to detect Go project with go.sum")
		}
	})

	t.Run("Detect without Go files", func(t *testing.T) {
		files := []string{"main.c", "Makefile"}
		if bs.Detect(files) {
			t.Error("Should not detect Go project without go.mod or go.sum")
		}
	})

	t.Run("GenerateInstallBlock simple", func(t *testing.T) {
		opts := InstallOptions{
			BinaryName: "myapp",
			Prefix:     "#{prefix}",
		}
		result := bs.GenerateInstallBlock(opts)

		if !strings.Contains(result, "def install") {
			t.Error("Install block should contain 'def install'")
		}
		if !strings.Contains(result, "system \"go\", \"build\", *std_go_args") {
			t.Error("Install block should contain Go build command")
		}
	})

	t.Run("GenerateInstallBlock with ldflags", func(t *testing.T) {
		opts := InstallOptions{
			BinaryName: "myapp",
			Prefix:     "#{prefix}",
			LDFlags:    []string{"-X main.version=1.0.0", "-s", "-w"},
		}
		result := bs.GenerateInstallBlock(opts)

		if !strings.Contains(result, "ldflags") {
			t.Error("Install block should contain ldflags")
		}
		if !strings.Contains(result, "-X main.version=1.0.0") {
			t.Error("Install block should contain version ldflags")
		}
	})

	t.Run("GenerateDependencies", func(t *testing.T) {
		deps := bs.GenerateDependencies()
		if len(deps) != 1 || deps[0] != "go" {
			t.Errorf("Expected dependencies [\"go\"], got %v", deps)
		}
	})

	t.Run("GenerateTestBlock", func(t *testing.T) {
		result := bs.GenerateTestBlock("myapp")

		if !strings.Contains(result, "test do") {
			t.Error("Test block should contain 'test do'")
		}
		if !strings.Contains(result, "myapp") {
			t.Error("Test block should contain binary name")
		}
		if !strings.Contains(result, "--version") {
			t.Error("Test block should test --version flag")
		}
	})
}

func TestRustBuildSystem(t *testing.T) {
	bs := &RustBuildSystem{}

	t.Run("Name", func(t *testing.T) {
		if bs.Name() != "Rust" {
			t.Errorf("Expected name 'Rust', got %s", bs.Name())
		}
	})

	t.Run("Detect with both Cargo files", func(t *testing.T) {
		files := []string{"src/main.rs", "Cargo.toml", "Cargo.lock"}
		if !bs.Detect(files) {
			t.Error("Expected to detect Rust project")
		}
	})

	t.Run("Detect with only Cargo.toml", func(t *testing.T) {
		files := []string{"src/main.rs", "Cargo.toml"}
		if bs.Detect(files) {
			t.Error("Should require both Cargo.toml and Cargo.lock")
		}
	})

	t.Run("GenerateInstallBlock", func(t *testing.T) {
		opts := InstallOptions{
			BinaryName: "myapp",
		}
		result := bs.GenerateInstallBlock(opts)

		if !strings.Contains(result, "def install") {
			t.Error("Install block should contain 'def install'")
		}
		if !strings.Contains(result, "cargo") {
			t.Error("Install block should contain cargo command")
		}
		if !strings.Contains(result, "*std_cargo_args") {
			t.Error("Install block should use std_cargo_args")
		}
	})

	t.Run("GenerateDependencies", func(t *testing.T) {
		deps := bs.GenerateDependencies()
		if len(deps) != 1 || deps[0] != "rust" {
			t.Errorf("Expected dependencies [\"rust\"], got %v", deps)
		}
	})
}

func TestCMakeBuildSystem(t *testing.T) {
	bs := &CMakeBuildSystem{}

	t.Run("Name", func(t *testing.T) {
		if bs.Name() != "CMake" {
			t.Errorf("Expected name 'CMake', got %s", bs.Name())
		}
	})

	t.Run("Detect", func(t *testing.T) {
		files := []string{"src/main.c", "CMakeLists.txt"}
		if !bs.Detect(files) {
			t.Error("Expected to detect CMake project")
		}
	})

	t.Run("GenerateInstallBlock", func(t *testing.T) {
		opts := InstallOptions{
			BinaryName: "myapp",
		}
		result := bs.GenerateInstallBlock(opts)

		if !strings.Contains(result, "cmake") {
			t.Error("Install block should contain cmake command")
		}
		if !strings.Contains(result, "*std_cmake_args") {
			t.Error("Install block should use std_cmake_args")
		}
		if !strings.Contains(result, "--build") {
			t.Error("Install block should contain build command")
		}
		if !strings.Contains(result, "--install") {
			t.Error("Install block should contain install command")
		}
	})

	t.Run("GenerateDependencies", func(t *testing.T) {
		deps := bs.GenerateDependencies()
		if len(deps) != 1 || deps[0] != "cmake" {
			t.Errorf("Expected dependencies [\"cmake\"], got %v", deps)
		}
	})
}

func TestMesonBuildSystem(t *testing.T) {
	bs := &MesonBuildSystem{}

	t.Run("Name", func(t *testing.T) {
		if bs.Name() != "Meson" {
			t.Errorf("Expected name 'Meson', got %s", bs.Name())
		}
	})

	t.Run("Detect", func(t *testing.T) {
		files := []string{"src/main.c", "meson.build"}
		if !bs.Detect(files) {
			t.Error("Expected to detect Meson project")
		}
	})

	t.Run("GenerateInstallBlock", func(t *testing.T) {
		opts := InstallOptions{
			BinaryName: "myapp",
		}
		result := bs.GenerateInstallBlock(opts)

		if !strings.Contains(result, "meson") {
			t.Error("Install block should contain meson command")
		}
		if !strings.Contains(result, "*std_meson_args") {
			t.Error("Install block should use std_meson_args")
		}
		if !strings.Contains(result, "compile") {
			t.Error("Install block should contain compile command")
		}
	})

	t.Run("GenerateDependencies", func(t *testing.T) {
		deps := bs.GenerateDependencies()
		if len(deps) != 2 || deps[0] != "meson" || deps[1] != "ninja" {
			t.Errorf("Expected dependencies [\"meson\", \"ninja\"], got %v", deps)
		}
	})
}

func TestMakefileBuildSystem(t *testing.T) {
	bs := &MakefileBuildSystem{}

	t.Run("Name", func(t *testing.T) {
		if bs.Name() != "Makefile" {
			t.Errorf("Expected name 'Makefile', got %s", bs.Name())
		}
	})

	t.Run("Detect Makefile", func(t *testing.T) {
		files := []string{"main.c", "Makefile"}
		if !bs.Detect(files) {
			t.Error("Expected to detect Makefile project")
		}
	})

	t.Run("Detect makefile lowercase", func(t *testing.T) {
		files := []string{"main.c", "makefile"}
		if !bs.Detect(files) {
			t.Error("Expected to detect makefile project")
		}
	})

	t.Run("Detect GNUmakefile", func(t *testing.T) {
		files := []string{"main.c", "GNUmakefile"}
		if !bs.Detect(files) {
			t.Error("Expected to detect GNUmakefile project")
		}
	})

	t.Run("GenerateInstallBlock", func(t *testing.T) {
		opts := InstallOptions{
			BinaryName: "myapp",
		}
		result := bs.GenerateInstallBlock(opts)

		if !strings.Contains(result, "make") {
			t.Error("Install block should contain make command")
		}
		if !strings.Contains(result, "PREFIX=#{prefix}") {
			t.Error("Install block should set PREFIX")
		}
	})

	t.Run("GenerateDependencies", func(t *testing.T) {
		deps := bs.GenerateDependencies()
		if len(deps) != 0 {
			t.Errorf("Expected no dependencies, got %v", deps)
		}
	})
}

func TestContainsFile(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		target   string
		expected bool
	}{
		{
			name:     "Exact match",
			files:    []string{"main.go", "go.mod", "README.md"},
			target:   "go.mod",
			expected: true,
		},
		{
			name:     "Suffix match",
			files:    []string{"src/main.go", "go.mod", "README.md"},
			target:   "main.go",
			expected: true,
		},
		{
			name:     "Not found",
			files:    []string{"main.go", "README.md"},
			target:   "go.mod",
			expected: false,
		},
		{
			name:     "Empty list",
			files:    []string{},
			target:   "go.mod",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsFile(tt.files, tt.target)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContainsAnyFile(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		targets  []string
		expected bool
	}{
		{
			name:     "First target matches",
			files:    []string{"Makefile", "main.c"},
			targets:  []string{"Makefile", "makefile"},
			expected: true,
		},
		{
			name:     "Second target matches",
			files:    []string{"makefile", "main.c"},
			targets:  []string{"Makefile", "makefile"},
			expected: true,
		},
		{
			name:     "No matches",
			files:    []string{"main.c", "README.md"},
			targets:  []string{"Makefile", "makefile"},
			expected: false,
		},
		{
			name:     "Empty targets",
			files:    []string{"Makefile"},
			targets:  []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAnyFile(tt.files, tt.targets)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
