# Cask Patterns

Complete, copy-paste ready cask templates for Linux installations. Each pattern includes detection rules, installation strategies, and troubleshooting guidance.

**⚠️ LINUX ONLY:** These patterns are for Linux systems only. All examples use Linux binaries.

**⚠️ CASK NAMING:** ALL casks MUST use `-linux` suffix (e.g., `app-name-linux`). This prevents collision with macOS casks and makes the Linux-only nature explicit.

## Package Format Priority

**When selecting which asset to package, follow this strict priority:**

1. **Tarball** (`.tar.gz`, `.tar.xz`, `.tgz`) - PREFERRED
   - Most portable across distributions
   - Simple extraction
   - Examples: `app-linux-x64.tar.gz`, `app-x86_64-unknown-linux-gnu.tar.gz`

2. **Debian Package** (`.deb`) - SECOND CHOICE
   - Only if no tarball available
   - Requires `ar` and `tar` extraction
   - Examples: `app_amd64.deb`, `app-linux-amd64.deb`

3. **Other formats** - Case-by-case
   - AppImage: Self-contained, can use directly
   - Snap/Flatpak: Generally avoid
   - RPM: Avoid (use tarball or .deb)

## SHA256 Verification (MANDATORY)

**Every cask MUST include SHA256 verification:**

```bash
# Calculate SHA256 for any download
curl -LO https://example.com/app-linux-x64.tar.gz
sha256sum app-linux-x64.tar.gz

# Verify against upstream if available
curl -LO https://example.com/SHA256SUMS
sha256sum --check SHA256SUMS
```

**NEVER skip SHA256 verification.** Use `sha256 :no_check` only with explicit justification.

## Table of Contents

1. [Simple Binary](#simple-binary)
2. [Multi-Binary Applications](#multi-binary-applications)
3. [Binary with Resources](#binary-with-resources)
4. [Tarball with Subdirectories](#tarball-with-subdirectories)
5. [Wrapper Scripts](#wrapper-scripts)

---

## Simple Binary

### When to Use This Pattern

**Detection Rules:**
- Archive contains a single executable at root or in `bin/` directory
- No additional resources, libraries, or configuration files
- Binary is self-contained and standalone
- Typical tarball structure: `app-version/binary` or `binary` at root

**Typical Projects:** Simple CLI tools distributed as GUI apps, single-file utilities, minimal TUI applications

### Complete Cask Template

```ruby
cask "app-name-linux" do
  version "1.0.0"
  sha256 "SHA256_HASH_HERE"
  
  url "https://github.com/USERNAME/PROJECT/releases/download/v#{version}/app-name-linux-x86_64.tar.gz"
  
  name "App Name"
  desc "Brief description of what this application does"
  homepage "https://github.com/USERNAME/PROJECT"
  license "MIT"
  
  # Specify architecture if binary is arch-specific
  # depends_on arch: :x86_64
  
  # If archive has subdirectory: "app-name-version/bin/app-name"
  binary "bin/app-name"
  
  # If binary is at root of archive
  # binary "app-name"
  
  # If binary needs to be renamed during installation
  # binary "app-name-bin", target: "app-name"

  test do
    # Verify binary exists and is executable
    assert_predicate bin/"app-name", :exist?
    assert_predicate bin/"app-name", :executable?
    
    # Test version output
    system bin/"app-name", "--version"
    
    # More thorough version check
    # output = shell_output("#{bin}/app-name --version")
    # assert_match version.to_s, output
  end
end
```

### Common Variations

**Binary at Root of Archive:**
```ruby
# Archive structure: app-name (directly in archive)
binary "app-name"
```

**Binary in Subdirectory:**
```ruby
# Archive structure: app-name-1.0.0/bin/app-name
binary "app-name-#{version}/bin/app-name", target: "app-name"

# Or with fixed path
binary "bin/app-name"
```

**Rename Binary on Install:**
```ruby
# Archive has: app-name-linux
# Install as: app-name
binary "app-name-linux", target: "app-name"
```

**Architecture-Specific Downloads:**
```ruby
on_intel do
  url "https://github.com/user/project/releases/download/v#{version}/app-linux-x86_64.tar.gz"
  sha256 "INTEL_SHA256"
end

on_arm do
  url "https://github.com/user/project/releases/download/v#{version}/app-linux-aarch64.tar.gz"
  sha256 "ARM_SHA256"
end

binary "app"
```

### Installation Strategies

**Direct Binary Installation:**
- Use when binary is ready to run without modification
- Homebrew creates symlink in `~/.linuxbrew/bin/`
- Binary permissions preserved automatically

**With Target Rename:**
- Use when upstream uses verbose naming (e.g., `app-x86_64-linux`)
- Provides cleaner command name for users
- Syntax: `binary "source-name", target: "dest-name"`

### Test Patterns

**Basic Existence Check:**
```ruby
test do
  assert_predicate bin/"app-name", :exist?
  assert_predicate bin/"app-name", :executable?
end
```

**Version Verification:**
```ruby
test do
  output = shell_output("#{bin}/app-name --version")
  assert_match version.to_s, output
end
```

**Help Command Test:**
```ruby
test do
  system bin/"app-name", "--help"
  assert_equal 0, $CHILD_STATUS.exitstatus
end
```

### Troubleshooting

**Problem:** Binary not found after extraction
**Solution:** List archive contents to find actual path:
```bash
tar -tzf app-name.tar.gz | grep -E "bin/|app-name"
# Update binary stanza with correct path
```

**Problem:** "Permission denied" when running binary
**Solution:** Archive may not preserve execute permissions. Use `chmod`:
```ruby
def install
  # Fix permissions before installing
  chmod 0755, "app-name"
  bin.install "app-name"
end

# But prefer the binary stanza when possible (it handles this automatically)
binary "app-name"  # This is better
```

**Problem:** Binary runs but shows wrong version
**Solution:** Upstream may not update version strings. Verify:
```bash
./app-name --version  # Check actual output
# Adjust test to match actual behavior
```

**Problem:** "cannot execute binary file: Exec format error"
**Solution:** Wrong architecture. Verify binary with:
```bash
file app-name
# Should show: ELF 64-bit LSB executable, x86-64
# Add architecture constraint to cask
```

---

## Multi-Binary Applications

### When to Use This Pattern

**Detection Rules:**
- Archive contains multiple executable files
- Application provides main binary plus helper tools
- CLI tool with server/daemon/client architecture
- Suite of related utilities bundled together

**Typical Projects:** Developer tools with multiple commands, client-server applications, toolchain bundles

### Complete Cask Template

```ruby
cask "app-name" do
  version "1.0.0"
  sha256 "SHA256_HASH_HERE"
  
  url "https://github.com/USERNAME/PROJECT/releases/download/v#{version}/app-name-linux.tar.gz"
  
  name "App Name"
  desc "Application suite with multiple utilities"
  homepage "https://github.com/USERNAME/PROJECT"
  license "Apache-2.0"

  depends_on :linux

  # Main application binary
  binary "bin/app-name"
  
  # Additional utilities
  binary "bin/app-name-cli"
  binary "bin/app-name-server"
  binary "bin/app-name-daemon"
  
  # Helper tools (optional)
  # binary "bin/app-name-utils"
  # binary "tools/helper"
  
  # If binaries need renaming for clarity
  # binary "bin/app-server", target: "app-name-server"
  # binary "bin/app-client", target: "app-name-client"

  test do
    # Test main binary
    assert_predicate bin/"app-name", :exist?
    system bin/"app-name", "--version"
    
    # Test additional binaries
    assert_predicate bin/"app-name-cli", :exist?
    system bin/"app-name-cli", "--help"
    
    assert_predicate bin/"app-name-server", :exist?
    system bin/"app-name-server", "--version"
    
    # Verify all binaries are executable
    %w[app-name app-name-cli app-name-server app-name-daemon].each do |cmd|
      assert_predicate bin/cmd, :executable?
    end
  end
end
```

### Common Variations

**Binaries in Different Subdirectories:**
```ruby
# Main binary in bin/
binary "bin/app-name"

# Utilities in tools/
binary "tools/app-name-tool"
binary "tools/converter"

# Scripts in scripts/
binary "scripts/setup-helper"
```

**Selective Binary Installation:**
```ruby
# Only install specific binaries, not all executables in archive
binary "bin/app-name"
binary "bin/app-cli"
# Deliberately skip: bin/internal-tool, bin/test-helper
```

**Wildcard Installation (Use Cautiously):**
```ruby
# Install all binaries from bin/ directory
# WARNING: Only use if you control what's in the archive
Dir["bin/*"].each do |binary_file|
  binary binary_file
end
```

**Namespace Binaries:**
```ruby
# Prevent command name conflicts by prefixing
binary "cli", target: "app-name-cli"
binary "server", target: "app-name-server"
binary "daemon", target: "app-name-daemon"
```

### Installation Strategies

**Hierarchical Structure:**
- Main binary: Most commonly used command
- CLI tools: Additional utilities for specific tasks
- Daemons/servers: Background processes
- Helpers: Internal tools (consider whether to install)

**Naming Conventions:**
- Keep upstream names when unambiguous
- Add prefixes to prevent conflicts (`tool` → `app-name-tool`)
- Use descriptive names for clarity

### Test Patterns

**Test All Binaries:**
```ruby
test do
  %w[app-name app-cli app-server].each do |cmd|
    assert_predicate bin/cmd, :exist?
    assert_predicate bin/cmd, :executable?
    system bin/cmd, "--version"
  end
end
```

**Test Main Binary Thoroughly, Others Minimally:**
```ruby
test do
  # Thorough test of main binary
  output = shell_output("#{bin}/app-name --version")
  assert_match version.to_s, output
  
  # Quick checks for others
  assert_predicate bin/"app-cli", :exist?
  assert_predicate bin/"app-server", :exist?
end
```

**Functional Relationship Test:**
```ruby
test do
  # Test that binaries work together
  # Start server (if it has a test/dry-run mode)
  server_output = shell_output("#{bin}/app-server --test-mode", 0)
  assert_match "Server ready", server_output
  
  # Use client to connect
  client_output = shell_output("#{bin}/app-cli status")
  assert_match "Connected", client_output
end
```

### Troubleshooting

**Problem:** Too many binaries, unclear which to install
**Solution:** Check upstream documentation. Install:
- Public-facing commands: YES
- Internal helpers: NO
- Optional tools: Case-by-case

**Problem:** Binary name conflicts with other packages
**Solution:** Use target rename:
```ruby
binary "server", target: "app-name-server"  # Prevents conflict with generic "server" command
```

**Problem:** Binaries depend on each other's paths
**Solution:** Use wrapper script (see [Wrapper Scripts](#wrapper-scripts) pattern)

---

## Binary with Resources

### When to Use This Pattern

**Detection Rules:**
- Archive contains binary plus additional files
- Application needs data files, configuration, libraries, or assets
- Presence of `lib/`, `share/`, `etc/`, or `resources/` directories
- Binary references external files at runtime

**Typical Projects:** GUI applications, games, tools with plugins, applications with embedded resources

### Complete Cask Template

```ruby
cask "app-name" do
  version "1.0.0"
  sha256 "SHA256_HASH_HERE"
  
  url "https://github.com/USERNAME/PROJECT/releases/download/v#{version}/app-name-linux-bundle.tar.gz"
  
  name "App Name"
  desc "Application with bundled resources and libraries"
  homepage "https://github.com/USERNAME/PROJECT"
  license "GPL-3.0"

  depends_on :linux
  
  # Install to libexec to keep resources together
  # This installs entire directory structure
  artifact "app-name-#{version}", target: "#{HOMEBREW_PREFIX}/libexec/app-name"
  
  # Create wrapper script that sets up environment
  # The binary needs to know where to find resources
  binary "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name"
  
  # Alternative: Copy resources and binary separately
  # artifact "share", target: "#{HOMEBREW_PREFIX}/share/app-name"
  # artifact "lib", target: "#{HOMEBREW_PREFIX}/lib/app-name"
  # binary "bin/app-name"

  test do
    assert_predicate bin/"app-name", :exist?
    assert_predicate bin/"app-name", :executable?
    
    # Verify resources are installed
    assert_predicate Pathname("#{HOMEBREW_PREFIX}/libexec/app-name/share"), :directory?
    
    # Test that binary can find its resources
    system bin/"app-name", "--version"
  end
end
```

### Common Variations

**Install Resources to Standard Locations:**
```ruby
# Follow FHS (Filesystem Hierarchy Standard)
artifact "share/app-name", target: "#{HOMEBREW_PREFIX}/share/app-name"
artifact "lib/app-name", target: "#{HOMEBREW_PREFIX}/lib/app-name"
binary "bin/app-name"

# Install config templates
artifact "etc/app-name", target: "#{HOMEBREW_PREFIX}/etc/app-name"
```

**Bundle Everything in libexec:**
```ruby
# Keep entire application isolated
artifact ".", target: "#{HOMEBREW_PREFIX}/libexec/app-name"

# Symlink binary to bin
binary "#{HOMEBREW_PREFIX}/libexec/app-name/app-name"
```

**Selective Resource Installation:**
```ruby
# Only install necessary resources
artifact "resources/required", target: "#{HOMEBREW_PREFIX}/share/app-name/resources"
artifact "themes", target: "#{HOMEBREW_PREFIX}/share/app-name/themes"
binary "app-name"

# Skip: docs/, examples/, tests/
```

**Install Shared Libraries:**
```ruby
# Application includes .so files
artifact "lib", target: "#{HOMEBREW_PREFIX}/lib/app-name"
binary "bin/app-name"

# May need wrapper to set LD_LIBRARY_PATH (see Wrapper Scripts pattern)
```

### Installation Strategies

**libexec Strategy (Recommended for Complex Apps):**
- Install entire application to `libexec/app-name/`
- Keeps all files together
- Binary knows where to find resources (relative paths)
- Clean namespace isolation

**Standard Locations Strategy:**
- Binary in `bin/`
- Resources in `share/app-name/`
- Libraries in `lib/app-name/`
- Config in `etc/app-name/`
- Follows Unix conventions
- May require binary to be patched or wrapped

**Hybrid Strategy:**
- Binary in `bin/` (for easy access)
- Resources in `libexec/app-name/` (for isolation)
- Use wrapper script to bridge the gap

### Test Patterns

**Verify Resource Installation:**
```ruby
test do
  # Check binary
  assert_predicate bin/"app-name", :exist?
  
  # Check resources
  assert_predicate Pathname("#{HOMEBREW_PREFIX}/share/app-name"), :directory?
  assert_predicate Pathname("#{HOMEBREW_PREFIX}/share/app-name/resources/data.json"), :exist?
  
  # Test binary can find resources
  output = shell_output("#{bin}/app-name --list-resources")
  assert_match "data.json", output
end
```

**Test Library Loading:**
```ruby
test do
  # Check that shared libraries can be found
  system "ldd", bin/"app-name"
  
  # Or test that binary runs (will fail if libs missing)
  system bin/"app-name", "--version"
end
```

### Troubleshooting

**Problem:** Binary can't find resources after installation
**Solution:** Binary may expect resources at hardcoded path. Options:
1. Use wrapper script to set environment variables
2. Patch binary to use different path (advanced)
3. Install to location binary expects (use artifact target)

**Problem:** "error while loading shared libraries: libapp.so.1: cannot open shared object file"
**Solution:** Binary can't find bundled libraries. Create wrapper:
```ruby
# See Wrapper Scripts pattern for full example
binary "bin/app-name"
# Need to set LD_LIBRARY_PATH
```

**Problem:** Application creates files relative to binary location
**Solution:** Keep everything together:
```ruby
artifact ".", target: "#{HOMEBREW_PREFIX}/libexec/app-name"
binary "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name"
```

**Problem:** Too much disk space used by bundled resources
**Solution:** Review archive contents, only install what's needed:
```bash
tar -tzf app.tar.gz | less  # Review contents
# Selectively install with multiple artifact stanzas
```

---

## Tarball with Subdirectories

### When to Use This Pattern

**Detection Rules:**
- Archive extracts to versioned directory: `app-name-1.0.0/...`
- Complex directory structure: `bin/`, `lib/`, `share/`, `doc/`, etc.
- Multiple levels of nesting
- Need to navigate into subdirectory before installing

**Typical Projects:** Traditional Unix-style software distributions, autotools packages, source releases adapted as binary releases

### Complete Cask Template

```ruby
cask "app-name" do
  version "1.0.0"
  sha256 "SHA256_HASH_HERE"
  
  url "https://github.com/USERNAME/PROJECT/releases/download/v#{version}/app-name-#{version}-linux-x86_64.tar.gz"
  
  name "App Name"
  desc "Application with traditional Unix directory structure"
  homepage "https://github.com/USERNAME/PROJECT"
  license "BSD-3-Clause"

  depends_on :linux

  # Approach 1: Reference files with full path including version
  binary "app-name-#{version}/bin/app-name"
  artifact "app-name-#{version}/share", target: "#{HOMEBREW_PREFIX}/share/app-name"
  artifact "app-name-#{version}/lib", target: "#{HOMEBREW_PREFIX}/lib/app-name"
  
  # Approach 2: Use container directive to change context
  # container type: :naked
  # This extracts archive contents without subdirectory
  
  test do
    assert_predicate bin/"app-name", :exist?
    system bin/"app-name", "--version"
  end
end
```

### Common Variations

**Using Container Directive:**
```ruby
# When archive structure is: app-name-version/ containing bin/, lib/, etc.
container type: :naked

# Now references are relative to inner directory
binary "bin/app-name"
artifact "share", target: "#{HOMEBREW_PREFIX}/share/app-name"
```

**Version-Agnostic Paths (Use Globbing):**
```ruby
# Avoid hardcoding version in paths
# This is more maintainable when version changes

# Option 1: Use staged_path
def install
  # staged_path is the root of extracted archive
  (staged_path/"app-name-#{version}/bin").each_child do |f|
    binary f if f.executable?
  end
end

# Option 2: Use variable
prefix = "app-name-#{version}"
binary "#{prefix}/bin/app-name"
artifact "#{prefix}/share", target: "#{HOMEBREW_PREFIX}/share/app-name"
```

**Multiple Subdirectories:**
```ruby
prefix = "app-name-#{version}"

# Binaries from bin/
binary "#{prefix}/bin/app-name"
binary "#{prefix}/bin/app-name-helper"

# Resources from multiple locations
artifact "#{prefix}/share/app-name", target: "#{HOMEBREW_PREFIX}/share/app-name"
artifact "#{prefix}/lib/app-name", target: "#{HOMEBREW_PREFIX}/lib/app-name"
artifact "#{prefix}/etc/app-name.conf", target: "#{HOMEBREW_PREFIX}/etc/app-name/app-name.conf"

# Documentation (optional)
artifact "#{prefix}/doc", target: "#{HOMEBREW_PREFIX}/share/doc/app-name"
```

**Nested Binaries:**
```ruby
# Binary is deeply nested: app-name-1.0.0/linux/x86_64/bin/app-name
prefix = "app-name-#{version}"
binary "#{prefix}/linux/x86_64/bin/app-name"
```

**Install Entire Subtree:**
```ruby
# Install entire versioned directory to libexec
artifact "app-name-#{version}", target: "#{HOMEBREW_PREFIX}/libexec/app-name"

# Then symlink binary
binary "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name"
```

### Installation Strategies

**Path Construction:**
- Use `version` variable in paths: `app-name-#{version}/bin/app-name`
- Create prefix variable for DRY: `prefix = "app-name-#{version}"`
- Use `container type: :naked` to strip top-level directory

**Resource Organization:**
- Map upstream structure to Homebrew locations
- `bin/` → directly to `bin/`
- `share/` → to `#{HOMEBREW_PREFIX}/share/app-name/`
- `lib/` → to `#{HOMEBREW_PREFIX}/lib/app-name/`
- `etc/` → to `#{HOMEBREW_PREFIX}/etc/app-name/`

**Simplification Techniques:**
- Use artifact to move entire directory trees
- Reference binaries with full path including subdirs
- Consider using custom install method for complex cases

### Test Patterns

**Standard Tests with Subdirectories:**
```ruby
test do
  assert_predicate bin/"app-name", :exist?
  
  # Verify resources in target location
  assert_predicate Pathname("#{HOMEBREW_PREFIX}/share/app-name"), :directory?
  
  system bin/"app-name", "--version"
end
```

**Verify Complex Structure:**
```ruby
test do
  # Check binary
  assert_predicate bin/"app-name", :exist?
  
  # Check multiple resource locations
  %w[share/app-name lib/app-name etc/app-name].each do |dir|
    assert_predicate Pathname("#{HOMEBREW_PREFIX}/#{dir}"), :directory?
  end
  
  # Functional test
  system bin/"app-name", "--help"
end
```

### Troubleshooting

**Problem:** "No such file or directory" when installing
**Solution:** Version in path doesn't match actual archive structure. Check:
```bash
tar -tzf app-name.tar.gz | head -20
# Verify actual directory name (might be "app-name-v1.0.0" not "app-name-1.0.0")
```

**Problem:** Archive extracts to unexpected directory name
**Solution:** Explicitly specify directory name or use container:
```ruby
# If archive creates "app-name-v1.0.0" directory:
binary "app-name-v#{version}/bin/app-name"

# Or use container to strip it:
container type: :naked
binary "bin/app-name"
```

**Problem:** Too many paths to update when version changes
**Solution:** Use variable:
```ruby
prefix = "app-name-#{version}"

binary "#{prefix}/bin/app-name"
artifact "#{prefix}/share", target: "#{HOMEBREW_PREFIX}/share/app-name"
artifact "#{prefix}/lib", target: "#{HOMEBREW_PREFIX}/lib/app-name"
# Only need to update prefix definition
```

**Problem:** Archive contains multiple versions or architectures
**Solution:** Navigate to correct subdirectory:
```ruby
arch = Hardware::CPU.intel? ? "x86_64" : "aarch64"
binary "app-name-#{version}/linux-#{arch}/bin/app-name"
```

---

## Wrapper Scripts

### When to Use This Pattern

**Detection Rules:**
- Binary needs environment variables set before execution
- Application requires `LD_LIBRARY_PATH` for bundled libraries
- Binary expects specific working directory
- Need to set `PATH`, `HOME`, or other variables
- Complex startup requirements

**Typical Projects:** Applications with bundled dependencies, tools requiring specific environments, JVM applications, proprietary software with bundled libraries

### Complete Cask Template

```ruby
cask "app-name" do
  version "1.0.0"
  sha256 "SHA256_HASH_HERE"
  
  url "https://github.com/USERNAME/PROJECT/releases/download/v#{version}/app-name-bundle.tar.gz"
  
  name "App Name"
  desc "Application requiring environment setup"
  homepage "https://github.com/USERNAME/PROJECT"
  license "MIT"

  depends_on :linux
  
  # Install everything to libexec
  artifact "app-name", target: "#{HOMEBREW_PREFIX}/libexec/app-name"
  
  # Create wrapper script in bin/
  # This script sets up environment and executes the real binary
  def install
    (bin/"app-name").write <<~BASH
      #!/bin/bash
      # Wrapper script for app-name
      
      # Set library path for bundled libraries
      export LD_LIBRARY_PATH="#{HOMEBREW_PREFIX}/libexec/app-name/lib:$LD_LIBRARY_PATH"
      
      # Set application home directory
      export APP_NAME_HOME="#{HOMEBREW_PREFIX}/libexec/app-name"
      
      # Set data directory
      export APP_NAME_DATA="$HOME/.local/share/app-name"
      
      # Ensure data directory exists
      mkdir -p "$APP_NAME_DATA"
      
      # Execute the real binary, passing all arguments
      exec "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name" "$@"
    BASH
    
    # Make wrapper executable
    chmod 0755, bin/"app-name"
  end

  test do
    # Test wrapper script exists and is executable
    assert_predicate bin/"app-name", :exist?
    assert_predicate bin/"app-name", :executable?
    
    # Test that wrapper works
    system bin/"app-name", "--version"
    
    # Verify environment is set (if binary reports it)
    # output = shell_output("#{bin}/app-name --print-env")
    # assert_match "APP_NAME_HOME", output
  end
end
```

### Common Variations

**Simple Library Path Wrapper:**
```ruby
def install
  artifact "app-name", target: "#{HOMEBREW_PREFIX}/libexec/app-name"
  
  (bin/"app-name").write <<~BASH
    #!/bin/bash
    export LD_LIBRARY_PATH="#{HOMEBREW_PREFIX}/libexec/app-name/lib:$LD_LIBRARY_PATH"
    exec "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name" "$@"
  BASH
  
  chmod 0755, bin/"app-name"
end
```

**Java/JVM Application Wrapper:**
```ruby
def install
  artifact "app-name", target: "#{HOMEBREW_PREFIX}/libexec/app-name"
  
  (bin/"app-name").write <<~BASH
    #!/bin/bash
    # Java application wrapper
    
    export JAVA_HOME="${JAVA_HOME:-/usr/lib/jvm/default-java}"
    export APP_HOME="#{HOMEBREW_PREFIX}/libexec/app-name"
    
    # Set classpath
    export CLASSPATH="$APP_HOME/lib/*:$CLASSPATH"
    
    # Set Java options
    JAVA_OPTS="${JAVA_OPTS:--Xmx2g -Xms512m}"
    
    # Execute via java
    exec "$JAVA_HOME/bin/java" $JAVA_OPTS -jar "$APP_HOME/lib/app-name.jar" "$@"
  BASH
  
  chmod 0755, bin/"app-name"
end
```

**Multiple Wrappers for Multiple Binaries:**
```ruby
def install
  artifact "app-name", target: "#{HOMEBREW_PREFIX}/libexec/app-name"
  
  # Wrapper for main binary
  (bin/"app-name").write <<~BASH
    #!/bin/bash
    export LD_LIBRARY_PATH="#{HOMEBREW_PREFIX}/libexec/app-name/lib:$LD_LIBRARY_PATH"
    exec "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name" "$@"
  BASH
  chmod 0755, bin/"app-name"
  
  # Wrapper for CLI tool
  (bin/"app-name-cli").write <<~BASH
    #!/bin/bash
    export LD_LIBRARY_PATH="#{HOMEBREW_PREFIX}/libexec/app-name/lib:$LD_LIBRARY_PATH"
    exec "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-cli" "$@"
  BASH
  chmod 0755, bin/"app-name-cli"
end
```

**Wrapper with Conditional Logic:**
```ruby
def install
  artifact "app-name", target: "#{HOMEBREW_PREFIX}/libexec/app-name"
  
  (bin/"app-name").write <<~BASH
    #!/bin/bash
    
    # Detect architecture
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then
      BINARY="app-name-x64"
    elif [ "$ARCH" = "aarch64" ]; then
      BINARY="app-name-arm64"
    else
      echo "Unsupported architecture: $ARCH" >&2
      exit 1
    fi
    
    export LD_LIBRARY_PATH="#{HOMEBREW_PREFIX}/libexec/app-name/lib:$LD_LIBRARY_PATH"
    exec "#{HOMEBREW_PREFIX}/libexec/app-name/bin/$BINARY" "$@"
  BASH
  
  chmod 0755, bin/"app-name"
end
```

**Wrapper with Config File Setup:**
```ruby
def install
  artifact "app-name", target: "#{HOMEBREW_PREFIX}/libexec/app-name"
  
  (bin/"app-name").write <<~BASH
    #!/bin/bash
    
    CONFIG_DIR="$HOME/.config/app-name"
    CONFIG_FILE="$CONFIG_DIR/config.yaml"
    
    # Create config directory if it doesn't exist
    if [ ! -d "$CONFIG_DIR" ]; then
      mkdir -p "$CONFIG_DIR"
      # Copy default config
      cp "#{HOMEBREW_PREFIX}/libexec/app-name/share/config.default.yaml" "$CONFIG_FILE"
      echo "Created default config at $CONFIG_FILE"
    fi
    
    export APP_NAME_CONFIG="$CONFIG_FILE"
    export LD_LIBRARY_PATH="#{HOMEBREW_PREFIX}/libexec/app-name/lib:$LD_LIBRARY_PATH"
    
    exec "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name" "$@"
  BASH
  
  chmod 0755, bin/"app-name"
end
```

### Installation Strategies

**When to Use Wrappers:**
- Binary needs `LD_LIBRARY_PATH` for bundled .so files
- Application expects specific environment variables
- Need to perform setup before binary runs
- Binary has hardcoded paths that need working directory
- Complex multi-component applications

**Wrapper Best Practices:**
- Use `exec` to replace shell process (cleaner process tree)
- Quote `"$@"` to preserve argument spacing
- Set environment, don't modify global system
- Make wrapper executable with `chmod 0755`
- Add comments explaining what wrapper does

**Alternative: Homebrew's write_env_script:**
```ruby
# Homebrew provides a helper for simple cases
# However, custom scripts offer more flexibility

# Simple example with write_env_script:
def install
  libexec.install Dir["*"]
  bin.write_env_script libexec/"bin/app-name",
    LD_LIBRARY_PATH: "#{libexec}/lib:$LD_LIBRARY_PATH",
    APP_HOME:        libexec
end
```

### Test Patterns

**Test Wrapper Execution:**
```ruby
test do
  # Verify wrapper exists
  assert_predicate bin/"app-name", :exist?
  assert_predicate bin/"app-name", :executable?
  
  # Test wrapper executes successfully
  system bin/"app-name", "--version"
end
```

**Test Environment Variable Setting:**
```ruby
test do
  # If binary can report its environment
  output = shell_output("#{bin}/app-name --print-env")
  assert_match "LD_LIBRARY_PATH", output
  assert_match "libexec/app-name/lib", output
end
```

**Test Binary Can Find Libraries:**
```ruby
test do
  # Test that wrapped binary doesn't fail with library errors
  # If binary runs without "cannot open shared object file" errors,
  # the wrapper is setting LD_LIBRARY_PATH correctly
  system bin/"app-name", "--version"
  assert_equal 0, $CHILD_STATUS.exitstatus
end
```

### Troubleshooting

**Problem:** Wrapper script doesn't execute, "Permission denied"
**Solution:** Script not marked executable:
```ruby
chmod 0755, bin/"app-name"  # Must call this after writing script
```

**Problem:** Binary still can't find libraries
**Solution:** Check library path is correct and libraries exist:
```bash
# Debug: print what wrapper sets
cat ~/.linuxbrew/bin/app-name  # View wrapper contents

# Check libraries exist
ls ~/.linuxbrew/libexec/app-name/lib/

# Test manually
LD_LIBRARY_PATH=~/.linuxbrew/libexec/app-name/lib ~/.linuxbrew/libexec/app-name/bin/app-name --version
```

**Problem:** Wrapper breaks when binary path contains spaces
**Solution:** Quote all paths:
```bash
exec "#{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name" "$@"
# Not: exec #{HOMEBREW_PREFIX}/libexec/app-name/bin/app-name "$@"
```

**Problem:** Arguments with spaces don't work
**Solution:** Use `"$@"` not `$*`:
```bash
exec "binary" "$@"  # Correct: preserves argument boundaries
# Not: exec "binary" $@  # Wrong: splits on spaces
```

**Problem:** Wrapper script is too complex and hard to maintain
**Solution:** Consider if upstream can fix this. If binary requires extensive environment setup, it may be better packaged differently or requested as an upstream improvement.

---

## Quick Reference

| Pattern | Use When | Key Stanzas |
|---------|----------|-------------|
| **Simple Binary** | Single executable, no resources | `binary "app-name"` |
| **Multi-Binary** | Multiple executables in archive | Multiple `binary` stanzas |
| **Binary with Resources** | Executables + data/lib/share files | `artifact` + `binary` |
| **Tarball Subdirectories** | Complex directory structure | `binary "prefix-#{version}/bin/app"` |
| **Wrapper Scripts** | Needs environment setup | Custom `install` method |

## Common Installation Paths

| Type | Homebrew Location | Example |
|------|------------------|---------|
| Binaries | `~/.linuxbrew/bin/` | Symlinks to actual binaries |
| Resources | `~/.linuxbrew/share/app-name/` | Data, themes, assets |
| Libraries | `~/.linuxbrew/lib/app-name/` | Shared objects (.so) |
| Config | `~/.linuxbrew/etc/app-name/` | Configuration files |
| Isolated apps | `~/.linuxbrew/libexec/app-name/` | Complete app bundles |

## Testing Checklist

Before submitting a cask:

- [ ] Binary installs to `~/.linuxbrew/bin/` and is executable
- [ ] Binary runs: `app-name --version` (or `--help`)
- [ ] Resources are in expected locations (if applicable)
- [ ] No hardcoded paths in scripts (use `#{HOMEBREW_PREFIX}`)
- [ ] Test block passes: `brew test --cask app-name`
- [ ] Audit passes: `brew audit --strict --online --cask app-name`
- [ ] Style check passes: `brew style Casks/app-name.rb`

## Getting Help

- Check existing casks: `brew cat --cask <similar-app>`
- Homebrew Cask Cookbook: https://docs.brew.sh/Cask-Cookbook
- Ask in discussions: https://github.com/orgs/Homebrew/discussions

## Platform Notes

**Linux-Specific Considerations:**

- **No .app bundles:** Use `binary` and `artifact`, not `app`
- **Shared libraries:** May need `LD_LIBRARY_PATH` in wrapper
- **Architecture:** Explicitly test x86_64 vs aarch64 if arch-specific
- **Permissions:** Archives may not preserve execute bits
- **Dependencies:** System libraries may differ across distros

**Supported Archive Formats:**
- `.tar.gz`, `.tgz` (most common)
- `.tar.xz`
- `.tar.bz2`
- `.zip`
- AppImage (can be used directly as binary)

---

**Last Updated:** 2025-02-08  
**Repository:** https://github.com/castrojo/homebrew-tap
