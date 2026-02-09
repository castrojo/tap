# Agent Validation Enforcement - Detailed Implementation Plan

**Date:** 2026-02-09  
**Priority:** 🔴 CRITICAL - FUNDAMENTAL ERROR  
**Status:** Comprehensive Implementation Plan - Ready to Execute

---

## 📋 TABLE OF CONTENTS

1. [Problem Statement](#problem-statement)
2. [Root Cause Analysis](#root-cause-analysis)
3. [Solution Architecture](#solution-architecture)
4. [Implementation Details](#implementation-details)
5. [Testing Plan](#testing-plan)
6. [Rollout Strategy](#rollout-strategy)
7. [Success Metrics](#success-metrics)
8. [Maintenance & Monitoring](#maintenance--monitoring)

---

## 🚨 PROBLEM STATEMENT

### The Core Issue

**Agents are committing packages that fail basic validation, resulting in CI failures on every attempt.**

### Evidence

**PR #22 (Copilot/Rancher Desktop) - 5 Architectural + 8 Style Errors:**

**Architectural Errors (FATAL - Would fail at install time):**
1. `binary` with `target:` parameter (Line 44) - NOT supported by Homebrew
2. `Dir.home` usage throughout - Hardcodes username, violates XDG
3. Invalid `depends_on formula:` syntax (Line 14) - NOT valid Homebrew
4. Incorrect artifact target paths - Wrong environment variable usage
5. Manual preflight path construction - Should use ENV.fetch("HOME")

**Style Errors (Auto-correctable):**
1. `Cask/StanzaOrder`: preflight out of order
2. `Cask/StanzaOrder`: binary out of order  
3. `Cask/StanzaOrder`: artifact out of order (x2)
4. `Cask/StanzaGrouping`: missing blank lines (x2)
5. `Style/TrailingCommaInArguments`: missing trailing commas (x2)

### Historical Pattern

| PR | Agent | Errors | Type | Passed? |
|----|-------|--------|------|---------|
| #18 | Unknown | Regex vs string | Style | ❌ No |
| #19 | Human | License stanza | Semantic | ❌ No |
| #22 | Copilot | 5 architectural + 8 style | Fatal | ❌ No |

**Success Rate: 0/3 (0%)**

### Impact

- **Time wasted:** 6-10 minutes per failed PR (vs 1-2 minutes if correct)
- **Developer frustration:** Repeated failures on basic errors
- **Trust erosion:** Agents appear unreliable for packaging work
- **Manual intervention required:** Humans must fix agent mistakes

---

## 🔍 ROOT CAUSE ANALYSIS

### Primary Root Cause

**Agents are NOT using `./tap-tools/tap-cask generate` despite mandatory instructions.**

### Evidence

1. **tap-cask would NEVER generate:**
   - `binary` with `target:` parameter (not in code)
   - `Dir.home` (always uses `ENV.fetch("HOME")`)
   - Invalid `depends_on` syntax (not in templates)

2. **All errors are manual-creation signatures:**
   - Incorrect syntax patterns
   - Hardcoded paths
   - Style violations that tap-validate would catch

3. **Skill was not followed:**
   - Step 1: "ALWAYS use tap-tools" - SKIPPED
   - Step 2: "MANDATORY validation" - SKIPPED

### Contributing Factors

**Factor 1: Skill Not Loaded**
- AGENTS.md says "MANDATORY: LOAD THE PACKAGING SKILL"
- Copilot likely never read `.github/skills/homebrew-packaging/SKILL.md`
- No evidence in PR description of skill usage

**Factor 2: Tools Not Available**
- Copilot runs in GitHub Actions environment
- tap-tools may not be built/available
- Can't use tools even if wanted to

**Factor 3: No Validation Checkpoint**
- Agent can commit without validation
- No automatic validation in PR workflow
- CI is the first time errors are caught (too late)

**Factor 4: Instructions Not Enforced**
- "MANDATORY" interpreted as advisory
- No blocking mechanism
- No immediate consequences for skipping

**Factor 5: No Feedback Loop**
- Agent commits → waits 3+ minutes → CI fails
- By then, agent has moved on
- No real-time feedback

---

## 🏗️ SOLUTION ARCHITECTURE

### Design Philosophy

**"Defense in Depth" - Multiple Overlapping Layers**

Each layer catches different error types and provides redundancy if one layer fails.

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: PREVENTION                                         │
│ Force tap-tools usage, detect manual creation              │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: AUTOMATION                                         │
│ Auto-validate PRs, auto-fix style errors                   │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: SEMANTIC VALIDATION                                │
│ Catch architectural errors (binary target, Dir.home)       │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 4: CI ENFORCEMENT                                     │
│ Final gate: brew audit + brew style                        │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Catches | Auto-Fixes | Blocks |
|-------|---------|------------|--------|
| **1. Prevention** | Manual creation | No | Yes (via detection) |
| **2. Automation** | Style violations | Yes | Yes (required check) |
| **3. Semantic** | Architectural errors | Partial | Yes |
| **4. CI** | Everything | No | Yes |

---

## 🔧 IMPLEMENTATION DETAILS

### Layer 1: Prevention - Force tap-tools Usage

#### 1.1: Update Documentation

**File: AGENTS.md**

Add after line 62 (after AGENT_BEST_PRACTICES.md section):

```markdown
## ⚠️ CRITICAL: PROHIBITED PATTERNS ⚠️

**These patterns indicate manual cask creation and WILL FAIL at install time:**

### ❌ NEVER USE: `binary` with `target:` Parameter

**WRONG:**
```ruby
binary "app", target: "#{Dir.home}/.local/bin/app"
```

**Why it fails:**
- `binary` stanza does NOT support `target:` parameter
- This is NOT valid Homebrew syntax
- Will fail at install time with "unknown keyword: :target"

**CORRECT:**
```ruby
# Option 1: Simple binary (links to $(brew --prefix)/bin)
binary "app"

# Option 2: Use artifact for custom location
artifact "app", target: "#{ENV.fetch("HOME")}/.local/bin/app"
```

### ❌ NEVER USE: `Dir.home`

**WRONG:**
```ruby
target: "#{Dir.home}/.local/bin/app"
"Exec=#{Dir.home}/.local/bin/app"
[
  "#{Dir.home}/.cache/app",
]
```

**Why it fails:**
- Hardcodes the current user's username into paths
- Won't work for other users
- Violates XDG Base Directory Specification

**CORRECT:**
```ruby
target: "#{ENV.fetch("HOME")}/.local/bin/app"
"Exec=#{ENV.fetch("HOME")}/.local/bin/app"
[
  "#{ENV.fetch("XDG_CACHE_HOME", "#{ENV.fetch("HOME")}/.cache")}/app",
]
```

### ❌ NEVER USE: `depends_on formula:`

**WRONG:**
```ruby
depends_on formula: "bash"
depends_on formula: "gtk+3"
```

**Why it fails:**
- `depends_on formula:` is NOT valid Homebrew syntax
- The correct syntax is just `formula "name"` for Formula dependencies
- For casks on Linux, usually don't need dependencies

**CORRECT:**
```ruby
# Usually no depends_on for Linux casks
# System packages should be documented in desc or caveats instead
```

### 🚫 If You See These Patterns: STOP

**If your cask contains ANY of these patterns, you MUST:**

1. **DELETE the file** you created
2. **Use tap-tools instead:**
   ```bash
   ./tap-tools/tap-cask generate <name> <github-url>
   ```
3. **Validate before committing:**
   ```bash
   ./tap-tools/tap-validate file Casks/<name>.rb --fix
   ```

**These patterns are automatic indicators that you:**
- Did NOT use tap-tools (mandatory)
- Created the cask manually (prohibited)
- Will fail CI (guaranteed)

**Use the tools. They exist to prevent these errors.**
```

#### 1.2: Update Packaging Skill

**File: `.github/skills/homebrew-packaging/SKILL.md`**

Update Step 1 (lines 94-122) to be more forceful:

```markdown
### Step 1: Generate Package Using tap-tools (⛔ NO EXCEPTIONS ⛔)

**⚠️ CRITICAL: MANUAL PACKAGE CREATION IS STRICTLY PROHIBITED ⚠️**

**You MUST use tap-tools for ALL package generation. No exceptions.**

**DO NOT:**
- ❌ Create .rb files by hand
- ❌ Copy and modify existing casks
- ❌ Write Ruby code directly
- ❌ Use a text editor to create packages

**DO:**
- ✅ Use `./tap-tools/tap-cask generate` for GUI applications
- ✅ Use `./tap-tools/tap-formula generate` for CLI tools

**Command:**
```bash
# For GUI applications (casks)
./tap-tools/tap-cask generate <name> <github-url>

# For CLI tools (formulas)
./tap-tools/tap-formula generate <name> <github-url>
```

**Example:**
```bash
./tap-tools/tap-cask generate rancher-desktop https://github.com/rancher-sandbox/rancher-desktop
# Creates: Casks/rancher-desktop-linux.rb (note: -linux suffix auto-added)
```

**Why tap-tools are MANDATORY:**

1. **Guarantees valid syntax** - No `binary` with `target:`, no `Dir.home`, no invalid `depends_on`
2. **Automatic validation** - Runs `tap-validate --fix` automatically
3. **Zero CI failures** - 100% success rate when used correctly
4. **XDG compliance** - Always uses `ENV.fetch("HOME")`, never hardcodes paths
5. **Platform detection** - Automatically filters Linux-only assets
6. **Format prioritization** - Tarball > deb > other
7. **SHA256 verification** - Downloads and calculates checksums
8. **Desktop integration** - Detects .desktop files and icons automatically

**What tap-tools PREVENTS:**

- ❌ `binary "app", target: "path"` - INVALID (would fail at install)
- ❌ `Dir.home` usage - INVALID (hardcodes username)
- ❌ `depends_on formula:` - INVALID (wrong syntax)
- ❌ Missing XDG environment variables
- ❌ Incorrect stanza ordering
- ❌ Line length violations
- ❌ Missing trailing commas
- ❌ Wrong path separators

**If you manually create a package, ALL of these errors are guaranteed.**

**CHECKPOINT VERIFICATION:**

Before proceeding to Step 2, you MUST confirm:
- [ ] Used `./tap-tools/tap-cask` OR `./tap-tools/tap-formula`
- [ ] Saw "✓ Successfully generated" message
- [ ] File exists in Casks/ or Formula/ directory
- [ ] Did NOT edit the file manually after generation

**If you cannot check all 4 boxes, you have not completed Step 1 correctly.**
```

### Layer 2: Automation - Auto-Validation Workflow

#### 2.1: Create Agent Validation Workflow

**File: `.github/workflows/agent-validation.yml`**

```yaml
name: Agent Validation & Auto-Fix

# Run on all PR changes to Ruby files
on:
  pull_request:
    types: [opened, synchronize, reopened]
    paths:
      - 'Casks/**/*.rb'
      - 'Formula/**/*.rb'

permissions:
  contents: write  # Needed to push auto-fixes
  pull-requests: write  # Needed to comment on PRs

jobs:
  validate-and-fix:
    name: Validate packages and auto-fix errors
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout PR branch
        uses: actions/checkout@v4
        with:
          ref: ${{ github.head_ref }}
          fetch-depth: 0  # Full history for accurate diffs
          token: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true
          cache-dependency-path: tap-tools/go.sum
      
      - name: Cache tap-tools binaries
        uses: actions/cache@v4
        with:
          path: tap-tools/tap-validate
          key: tap-validate-${{ runner.os }}-${{ hashFiles('tap-tools/cmd/tap-validate/**/*.go', 'tap-tools/internal/**/*.go') }}
          restore-keys: |
            tap-validate-${{ runner.os }}-
      
      - name: Build tap-validate (if not cached)
        run: |
          if [ ! -f tap-tools/tap-validate ]; then
            echo "Building tap-validate..."
            cd tap-tools
            go build -o tap-validate ./cmd/tap-validate
            cd ..
          else
            echo "Using cached tap-validate"
          fi
          chmod +x tap-tools/tap-validate
      
      - name: Get changed Ruby files
        id: changed-files
        uses: tj-actions/changed-files@v45
        with:
          files: |
            Casks/**/*.rb
            Formula/**/*.rb
      
      - name: Validate and auto-fix changed files
        id: validation
        run: |
          set +e  # Don't exit on error, we handle it
          
          if [ -z "${{ steps.changed-files.outputs.all_changed_files }}" ]; then
            echo "status=no_files" >> $GITHUB_OUTPUT
            echo "No Ruby files changed in this PR"
            exit 0
          fi
          
          echo "Changed files:"
          echo "${{ steps.changed-files.outputs.all_changed_files }}"
          echo ""
          
          VALIDATION_FAILED=0
          SEMANTIC_ERRORS=0
          STYLE_FIXED=0
          
          for file in ${{ steps.changed-files.outputs.all_changed_files }}; do
            if [ ! -f "$file" ]; then
              echo "⚠️  File $file was deleted, skipping"
              continue
            fi
            
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "Validating: $file"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            
            # Run validation with --fix flag
            if ./tap-tools/tap-validate file "$file" --fix 2>&1 | tee validation_output.txt; then
              echo "✅ $file passed validation"
              
              # Check if file was modified by --fix
              if ! git diff --quiet "$file"; then
                echo "🔧 Auto-fixed style issues in $file"
                STYLE_FIXED=1
              fi
            else
              echo "❌ $file failed validation"
              VALIDATION_FAILED=1
              
              # Check for semantic errors
              if grep -q "binary.*target:" "$file"; then
                echo "  🚨 SEMANTIC ERROR: binary with target: parameter"
                SEMANTIC_ERRORS=1
              fi
              if grep -q "Dir\.home" "$file"; then
                echo "  🚨 SEMANTIC ERROR: Uses Dir.home instead of ENV.fetch(\"HOME\")"
                SEMANTIC_ERRORS=1
              fi
              if grep -q "depends_on formula:" "$file"; then
                echo "  🚨 SEMANTIC ERROR: Invalid depends_on formula: syntax"
                SEMANTIC_ERRORS=1
              fi
            fi
            echo ""
          done
          
          # Set outputs for later steps
          echo "validation_failed=$VALIDATION_FAILED" >> $GITHUB_OUTPUT
          echo "semantic_errors=$SEMANTIC_ERRORS" >> $GITHUB_OUTPUT
          echo "style_fixed=$STYLE_FIXED" >> $GITHUB_OUTPUT
          
          if [ $VALIDATION_FAILED -eq 1 ]; then
            echo "status=failed" >> $GITHUB_OUTPUT
            exit 1
          else
            echo "status=passed" >> $GITHUB_OUTPUT
            exit 0
          fi
      
      - name: Commit auto-fixes
        if: steps.validation.outputs.style_fixed == '1' && steps.validation.outputs.validation_failed == '0'
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
          
          git add -A
          git commit -m "style: auto-fix validation errors

Auto-fixed by tap-validate --fix

Changes:
$(git diff --cached --stat)

Assisted-by: GitHub Actions Bot"
          
          git push origin HEAD:${{ github.head_ref }}
          
          echo "✅ Auto-fixes committed and pushed"
      
      - name: Comment on PR (validation passed)
        if: steps.validation.outputs.status == 'passed' && steps.validation.outputs.style_fixed == '1'
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `✅ **Validation passed with auto-fixes**
              
              Style issues were automatically fixed and committed.
              
              **Auto-fixed errors:** Style violations (stanza order, trailing commas, etc.)
              
              The PR is ready for CI checks.`
            })
      
      - name: Comment on PR (semantic errors)
        if: steps.validation.outputs.semantic_errors == '1'
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `❌ **CRITICAL: Semantic errors detected**
              
              This cask/formula has **architectural errors that cannot be auto-fixed**.
              
              These errors indicate the package was **manually created** instead of using tap-tools.
              
              **Common semantic errors found:**
              - \`binary\` with \`target:\` parameter (NOT supported by Homebrew)
              - \`Dir.home\` usage (should use \`ENV.fetch("HOME")\`)
              - Invalid \`depends_on formula:\` syntax
              
              **Required fix:**
              
              1. **Delete the manually created file**
              2. **Use tap-tools to generate the package:**
                 \`\`\`bash
                 ./tap-tools/tap-cask generate <name> <github-url>
                 # OR
                 ./tap-tools/tap-formula generate <name> <github-url>
                 \`\`\`
              3. **Validate before committing:**
                 \`\`\`bash
                 ./tap-tools/tap-validate file <path> --fix
                 \`\`\`
              
              **Why this matters:**
              - These errors will cause **install-time failures**
              - CI will fail (guaranteed)
              - Manual creation is **prohibited** by the packaging skill
              
              See: [docs/AGENT_BEST_PRACTICES.md](https://github.com/castrojo/tap/blob/main/docs/AGENT_BEST_PRACTICES.md)`
            })
      
      - name: Fail job if validation failed
        if: steps.validation.outputs.validation_failed == '1'
        run: |
          echo "❌ Validation failed - see errors above"
          echo ""
          echo "Common fixes:"
          echo "  1. Use tap-tools to generate the package (mandatory)"
          echo "  2. Run: ./tap-tools/tap-validate file <path> --fix"
          echo "  3. Never manually create casks/formulas"
          exit 1
```

#### 2.2: Make Validation a Required Check

**Manual step (requires repository admin):**

1. Go to: https://github.com/castrojo/tap/settings/branches
2. Add or edit branch protection rule for `main`
3. Under "Require status checks to pass before merging":
   - Enable: "Require status checks to pass"
   - Add required check: `validate-and-fix`
4. Save changes

**Effect:** PRs cannot be merged until validation passes.

### Layer 3: Semantic Validation - Catch Architectural Errors

#### 3.1: Add Semantic Checks to tap-validate

**File: `tap-tools/internal/validation/semantic.go` (NEW FILE)**

```go
package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// SemanticError represents a semantic/architectural error that cannot be auto-fixed
type SemanticError struct {
	Line        int
	Column      int
	Message     string
	Severity    string // "error" or "warning"
	Code        string // Error code (e.g., "BINARY_WITH_TARGET")
	Suggestion  string
	DocsURL     string
}

// SemanticValidator checks for architectural errors in casks/formulas
type SemanticValidator struct {
	content  string
	lines    []string
	errors   []SemanticError
	warnings []SemanticError
}

// NewSemanticValidator creates a new semantic validator
func NewSemanticValidator(content string) *SemanticValidator {
	return &SemanticValidator{
		content: content,
		lines:   strings.Split(content, "\n"),
		errors:  []SemanticError{},
		warnings: []SemanticError{},
	}
}

// Validate runs all semantic checks
func (sv *SemanticValidator) Validate() ([]SemanticError, []SemanticError) {
	sv.checkBinaryWithTarget()
	sv.checkDirHomeUsage()
	sv.checkInvalidDependsOn()
	sv.checkHardcodedPaths()
	sv.checkMissingXDGVariables()
	
	return sv.errors, sv.warnings
}

// checkBinaryWithTarget detects binary stanzas with target: parameter
func (sv *SemanticValidator) checkBinaryWithTarget() {
	// Regex: binary "app", target: "path"
	re := regexp.MustCompile(`binary\s+"[^"]+",\s+target:`)
	
	for i, line := range sv.lines {
		if re.MatchString(line) {
			sv.errors = append(sv.errors, SemanticError{
				Line:     i + 1,
				Column:   strings.Index(line, "binary"),
				Message:  "binary stanza does not support 'target:' parameter",
				Severity: "error",
				Code:     "BINARY_WITH_TARGET",
				Suggestion: "Use 'artifact' stanza instead: artifact \"app\", target: \"path\"",
				DocsURL:  "https://github.com/castrojo/tap/blob/main/docs/AGENT_BEST_PRACTICES.md#binary-with-target",
			})
		}
	}
}

// checkDirHomeUsage detects Dir.home usage (should use ENV.fetch("HOME"))
func (sv *SemanticValidator) checkDirHomeUsage() {
	re := regexp.MustCompile(`Dir\.home`)
	
	for i, line := range sv.lines {
		if re.MatchString(line) {
			sv.errors = append(sv.errors, SemanticError{
				Line:     i + 1,
				Column:   strings.Index(line, "Dir.home"),
				Message:  "Dir.home hardcodes username; use ENV.fetch(\"HOME\") instead",
				Severity: "error",
				Code:     "DIR_HOME_USAGE",
				Suggestion: "Replace: Dir.home → ENV.fetch(\"HOME\")",
				DocsURL:  "https://github.com/castrojo/tap/blob/main/docs/AGENT_BEST_PRACTICES.md#dir-home-usage",
			})
		}
	}
}

// checkInvalidDependsOn detects invalid depends_on syntax
func (sv *SemanticValidator) checkInvalidDependsOn() {
	re := regexp.MustCompile(`depends_on\s+formula:`)
	
	for i, line := range sv.lines {
		if re.MatchString(line) {
			sv.errors = append(sv.errors, SemanticError{
				Line:     i + 1,
				Column:   strings.Index(line, "depends_on"),
				Message:  "depends_on formula: is invalid syntax",
				Severity: "error",
				Code:     "INVALID_DEPENDS_ON",
				Suggestion: "Remove this line (dependencies usually not needed for Linux casks)",
				DocsURL:  "https://github.com/castrojo/tap/blob/main/docs/AGENT_BEST_PRACTICES.md#depends-on",
			})
		}
	}
}

// checkHardcodedPaths detects hardcoded paths that should use variables
func (sv *SemanticValidator) checkHardcodedPaths() {
	// Check for hardcoded /home/username patterns
	re := regexp.MustCompile(`"/home/[^/]+/`)
	
	for i, line := range sv.lines {
		if re.MatchString(line) {
			sv.warnings = append(sv.warnings, SemanticError{
				Line:     i + 1,
				Column:   re.FindStringIndex(line)[0],
				Message:  "Hardcoded /home/username path detected",
				Severity: "warning",
				Code:     "HARDCODED_PATH",
				Suggestion: "Use ENV.fetch(\"HOME\") instead of hardcoded path",
				DocsURL:  "",
			})
		}
	}
}

// checkMissingXDGVariables warns if XDG variables are not used where expected
func (sv *SemanticValidator) checkMissingXDGVariables() {
	// Check if .local/share, .config, or .cache is used without XDG variables
	patterns := map[string]string{
		`\.local/share`: "XDG_DATA_HOME",
		`\.config`:      "XDG_CONFIG_HOME",
		`\.cache`:       "XDG_CACHE_HOME",
	}
	
	for pattern, varName := range patterns {
		re := regexp.MustCompile(pattern)
		for i, line := range sv.lines {
			// Skip if line already uses the XDG variable
			if strings.Contains(line, varName) {
				continue
			}
			
			if re.MatchString(line) {
				sv.warnings = append(sv.warnings, SemanticError{
					Line:     i + 1,
					Column:   re.FindStringIndex(line)[0],
					Message:  fmt.Sprintf("Consider using ENV.fetch(\"%s\", ...) for XDG compliance", varName),
					Severity: "warning",
					Code:     "MISSING_XDG_VAR",
					Suggestion: fmt.Sprintf("Wrap with: ENV.fetch(\"%s\", \"#{ENV.fetch(\\\"HOME\\\")}/...\")", varName),
					DocsURL:  "https://github.com/castrojo/tap/blob/main/docs/CASK_CREATION_GUIDE.md#xdg-base-directory-spec",
				})
			}
		}
	}
}

// FormatErrors formats semantic errors for display
func FormatErrors(errors []SemanticError, warnings []SemanticError) string {
	var output strings.Builder
	
	if len(errors) > 0 {
		output.WriteString("🚨 SEMANTIC ERRORS (must fix):\n\n")
		for _, err := range errors {
			output.WriteString(fmt.Sprintf("  Line %d: %s\n", err.Line, err.Message))
			if err.Suggestion != "" {
				output.WriteString(fmt.Sprintf("    💡 %s\n", err.Suggestion))
			}
			if err.DocsURL != "" {
				output.WriteString(fmt.Sprintf("    📖 %s\n", err.DocsURL))
			}
			output.WriteString("\n")
		}
	}
	
	if len(warnings) > 0 {
		output.WriteString("⚠️  WARNINGS (should fix):\n\n")
		for _, warn := range warnings {
			output.WriteString(fmt.Sprintf("  Line %d: %s\n", warn.Line, warn.Message))
			if warn.Suggestion != "" {
				output.WriteString(fmt.Sprintf("    💡 %s\n", warn.Suggestion))
			}
			output.WriteString("\n")
		}
	}
	
	return output.String()
}
```

#### 3.2: Integrate Semantic Validation into tap-validate

**File: `tap-tools/cmd/tap-validate/main.go`**

Add semantic validation before RuboCop checks:

```go
// Add after style validation, around line 150
func validateFile(filePath string, fix bool) error {
	// ... existing code ...
	
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Run semantic validation FIRST (before style checks)
	printSection("Semantic Validation")
	validator := validation.NewSemanticValidator(string(content))
	errors, warnings := validator.Validate()
	
	if len(errors) > 0 || len(warnings) > 0 {
		fmt.Println(validation.FormatErrors(errors, warnings))
	}
	
	if len(errors) > 0 {
		printError("Semantic validation failed")
		fmt.Println()
		fmt.Println("These are architectural errors that cannot be auto-fixed.")
		fmt.Println()
		fmt.Println("This usually means the file was manually created instead of using tap-tools.")
		fmt.Println()
		fmt.Println("Required fix:")
		fmt.Println("  1. Delete this file")
		fmt.Println("  2. Use tap-tools to generate:")
		fmt.Println("     ./tap-tools/tap-cask generate <name> <github-url>")
		fmt.Println()
		return fmt.Errorf("semantic validation failed with %d error(s)", len(errors))
	}
	
	if len(warnings) > 0 {
		printWarn(fmt.Sprintf("Found %d warning(s) - consider fixing", len(warnings)))
	} else {
		printSuccess("Semantic validation passed")
	}
	
	// Continue with existing style validation...
	// ... rest of function ...
}
```

### Layer 4: CI Enforcement - Final Safety Net

#### 4.1: Enhance CI Error Messages

**File: `.github/workflows/tests.yml`**

Update the brew style step (around line 40):

```yaml
- name: Run brew style
  id: brew-style
  run: |
    for file in ${{ steps.changed-files.outputs.all_changed_files }}; do
      # Extract formula/cask name from path
      name=$(basename "$file" .rb)
      if [[ "$file" == Casks/* ]]; then
        echo "Checking style for cask $name (from $file)"
        brew style --cask "castrojo/tap/$name"
      elif [[ "$file" == Formula/* ]]; then
        echo "Checking style for formula $name (from $file)"
        brew style "castrojo/tap/$name"
      fi
    done
  continue-on-error: false

- name: Show detailed fix instructions
  if: failure() && steps.brew-style.conclusion == 'failure'
  run: |
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "❌ STYLE VALIDATION FAILED IN CI"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "This should have been caught by the agent-validation workflow."
    echo ""
    echo "📋 TO FIX LOCALLY:"
    echo "  1. Run validation with auto-fix:"
    echo "     ./tap-tools/tap-validate file <filename> --fix"
    echo ""
    echo "  2. Review and commit changes:"
    echo "     git add <filename>"
    echo "     git commit --amend --no-edit"
    echo "     git push --force-with-lease"
    echo ""
    echo "📖 PREVENTION:"
    echo "  - ALWAYS use tap-tools to generate packages"
    echo "  - NEVER create casks/formulas manually"
    echo "  - Run tap-validate before every commit"
    echo ""
    echo "📚 DOCUMENTATION:"
    echo "  - Agent Best Practices: docs/AGENT_BEST_PRACTICES.md"
    echo "  - Packaging Skill: .github/skills/homebrew-packaging/SKILL.md"
    echo ""
```

#### 4.2: Add Semantic Pre-Check to CI

**File: `.github/workflows/tests.yml`**

Add before brew audit (after checkout):

```yaml
- name: Semantic validation pre-check
  run: |
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🔍 Checking for semantic errors..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    SEMANTIC_ERRORS=0
    
    for file in Casks/*.rb Formula/*.rb; do
      [ -f "$file" ] || continue
      
      echo "Checking: $file"
      
      # Check for binary with target
      if grep -q 'binary.*target:' "$file"; then
        echo "  ❌ FATAL: binary with target: parameter (line $(grep -n 'binary.*target:' "$file" | cut -d: -f1))"
        echo "     This will FAIL at install time"
        echo "     Fix: Use artifact stanza instead"
        SEMANTIC_ERRORS=1
      fi
      
      # Check for Dir.home
      if grep -q 'Dir\.home' "$file"; then
        echo "  ❌ ERROR: Uses Dir.home (line $(grep -n 'Dir\.home' "$file" | cut -d: -f1))"
        echo "     Fix: Replace with ENV.fetch(\"HOME\")"
        SEMANTIC_ERRORS=1
      fi
      
      # Check for invalid depends_on
      if grep -q 'depends_on formula:' "$file"; then
        echo "  ❌ ERROR: Invalid depends_on syntax (line $(grep -n 'depends_on formula:' "$file" | cut -d: -f1))"
        echo "     Fix: Remove this line"
        SEMANTIC_ERRORS=1
      fi
      
      if [ $SEMANTIC_ERRORS -eq 0 ]; then
        echo "  ✅ No semantic errors"
      fi
      echo ""
    done
    
    if [ $SEMANTIC_ERRORS -eq 1 ]; then
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      echo "❌ SEMANTIC ERRORS DETECTED"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      echo ""
      echo "These are ARCHITECTURAL errors that indicate manual package creation."
      echo ""
      echo "🚫 REQUIRED ACTION:"
      echo "  1. Delete the manually created file"
      echo "  2. Use tap-tools to generate:"
      echo "     ./tap-tools/tap-cask generate <name> <github-url>"
      echo "  3. Validate before committing:"
      echo "     ./tap-tools/tap-validate file <path> --fix"
      echo ""
      echo "📖 WHY THIS MATTERS:"
      echo "  - These errors will cause INSTALL-TIME FAILURES"
      echo "  - Manual creation is PROHIBITED by packaging skill"
      echo "  - tap-tools generate valid code automatically"
      echo ""
      exit 1
    fi
    
    echo "✅ Semantic validation passed"
```

---

## 🧪 TESTING PLAN

### Test 1: Style Errors (Auto-Fixable)

**Purpose:** Verify auto-fix workflow catches and fixes style violations.

**Setup:**
1. Create test branch: `test/style-errors`
2. Create deliberately broken cask with style issues:
   - Wrong stanza order
   - Missing trailing commas
   - Line too long

**Expected Result:**
- agent-validation workflow runs
- Auto-fixes style issues
- Commits fix automatically
- CI passes

**Commands:**
```bash
git checkout -b test/style-errors
cp Casks/sublime-text-linux.rb Casks/test-style.rb
# Manually break style (reorder stanzas, remove commas)
git add Casks/test-style.rb
git commit -m "test: deliberately broken style"
git push origin test/style-errors
gh pr create --title "Test: Style Auto-Fix" --body "Testing auto-fix workflow"
# Watch: gh pr view --web
```

**Success Criteria:**
- [ ] Validation workflow runs
- [ ] Detects style errors
- [ ] Auto-fixes them
- [ ] Commits fix
- [ ] CI passes

### Test 2: Semantic Errors (Cannot Auto-Fix)

**Purpose:** Verify workflow catches architectural errors and blocks merge.

**Setup:**
1. Create test branch: `test/semantic-errors`
2. Create cask with semantic errors:
   - `binary` with `target:` parameter
   - `Dir.home` usage
   - Invalid `depends_on formula:`

**Expected Result:**
- agent-validation workflow runs
- Detects semantic errors
- Comments on PR with instructions
- Blocks merge (validation fails)

**Commands:**
```bash
git checkout -b test/semantic-errors
cat > Casks/test-semantic.rb <<'EOF'
cask "test-semantic" do
  version "1.0.0"
  sha256 "abc123"
  url "https://example.com/test.tar.gz"
  name "Test"
  desc "Test semantic errors"
  homepage "https://example.com"
  
  depends_on formula: "bash"
  binary "test", target: "#{Dir.home}/.local/bin/test"
end
EOF
git add Casks/test-semantic.rb
git commit -m "test: semantic errors"
git push origin test/semantic-errors
gh pr create --title "Test: Semantic Errors" --body "Testing semantic validation"
```

**Success Criteria:**
- [ ] Validation workflow runs
- [ ] Detects 3 semantic errors (binary target, Dir.home, depends_on)
- [ ] Posts comment with fix instructions
- [ ] Validation check fails
- [ ] Cannot merge PR

### Test 3: Valid Package (Should Pass)

**Purpose:** Verify workflow doesn't break valid packages.

**Setup:**
1. Create test branch: `test/valid-package`
2. Use tap-cask to generate valid package

**Expected Result:**
- agent-validation workflow runs
- No errors found
- Validation passes
- CI passes
- Can merge

**Commands:**
```bash
git checkout -b test/valid-package
./tap-tools/tap-cask generate jq https://github.com/jqlang/jq
./tap-tools/tap-validate file Casks/jq-linux.rb --fix
git add Casks/jq-linux.rb
git commit -m "feat(cask): add jq-linux (test)"
git push origin test/valid-package
gh pr create --title "Test: Valid Package" --body "Testing with tap-cask generated package"
```

**Success Criteria:**
- [ ] Validation workflow runs
- [ ] No errors found
- [ ] Validation passes immediately
- [ ] CI passes
- [ ] PR shows green checkmark

### Test 4: Integration Test (Full Workflow)

**Purpose:** Test complete workflow from issue to merged PR.

**Setup:**
1. Create test issue for new package
2. Have agent process it (or simulate)
3. Verify all checkpoints

**Commands:**
```bash
gh issue create --title "Add bat CLI tool" --body "Repository: https://github.com/sharkdp/bat"
# Then manually or via agent:
./tap-tools/tap-formula generate bat https://github.com/sharkdp/bat
./tap-tools/tap-validate file Formula/bat.rb --fix
git checkout -b feat/add-bat
git add Formula/bat.rb
git commit -m "feat(formula): add bat syntax highlighter"
git push origin feat/add-bat
gh pr create --title "feat(formula): add bat" --body "Closes #<issue-number>"
```

**Success Criteria:**
- [ ] tap-formula generates valid formula
- [ ] tap-validate passes (no fixes needed)
- [ ] Validation workflow passes
- [ ] CI passes on first push
- [ ] PR can be merged

---

## 🚀 ROLLOUT STRATEGY

### Phase 1: Documentation & Preparation (Day 1)

**Duration:** 2 hours

**Tasks:**
1. Update AGENTS.md with PROHIBITED PATTERNS section
2. Update packaging skill with stronger language
3. Update AGENT_BEST_PRACTICES.md with PR #22 example
4. Create semantic validation code
5. Test semantic validation locally
6. Commit and push documentation

**Deliverables:**
- [ ] AGENTS.md updated
- [ ] Packaging skill updated
- [ ] Semantic validation code written
- [ ] Local tests pass

### Phase 2: Workflow Deployment (Day 1-2)

**Duration:** 3 hours

**Tasks:**
1. Create `.github/workflows/agent-validation.yml`
2. Test workflow with test PR (style errors)
3. Test workflow with test PR (semantic errors)
4. Test workflow with valid package
5. Fix any issues found
6. Deploy to production

**Deliverables:**
- [ ] Workflow file created
- [ ] 3 test PRs pass
- [ ] Workflow deployed

### Phase 3: Required Check Setup (Day 2)

**Duration:** 30 minutes

**Tasks:**
1. Make `validate-and-fix` a required status check
2. Test that PR cannot merge without passing
3. Document required check in CONTRIBUTING.md

**Deliverables:**
- [ ] Required check configured
- [ ] Tested and verified
- [ ] Documented

### Phase 4: CI Enhancement (Day 2)

**Duration:** 1 hour

**Tasks:**
1. Add semantic pre-check to tests.yml
2. Enhance error messages
3. Test with deliberately broken cask
4. Deploy updates

**Deliverables:**
- [ ] CI updated
- [ ] Tests pass
- [ ] Error messages helpful

### Phase 5: Monitoring (Week 1)

**Duration:** Ongoing

**Tasks:**
1. Monitor next 5 agent PRs
2. Measure first-push success rate
3. Document any new failure modes
4. Iterate on validation rules
5. Gather feedback

**Deliverables:**
- [ ] Success rate >80% (target: 100%)
- [ ] All PRs use tap-tools
- [ ] Zero semantic errors reach CI

---

## 📊 SUCCESS METRICS

### Primary Metrics

**1. First-Push CI Success Rate**
- **Current:** 0% (0/3 PRs passed)
- **Target:** 100% (10/10 PRs pass)
- **Measure:** `(PRs passing CI on first push) / (Total PRs)` 

**2. Semantic Errors Caught**
- **Target:** 100% caught by validation workflow
- **Measure:** Count semantic errors in validation vs CI

**3. Time to Pass CI**
- **Current:** 6-10 minutes (failed → fixed → passed)
- **Target:** 1-2 minutes (passed immediately)
- **Measure:** Time from PR creation to green CI

**4. tap-tools Usage Rate**
- **Current:** Unknown (likely low)
- **Target:** 100% (all PRs use tap-tools)
- **Measure:** Manual review of PR code

### Secondary Metrics

**5. Auto-Fix Rate**
- **Target:** >90% of style issues auto-fixed
- **Measure:** Count auto-fix commits / total validation runs

**6. False Positive Rate**
- **Target:** <5% (validation fails on valid code)
- **Measure:** Count false positives / total validations

**7. Developer Satisfaction**
- **Target:** Positive feedback on validation workflow
- **Measure:** Qualitative feedback

---

## 🔧 MAINTENANCE & MONITORING

### Daily Checks (Week 1)

**What to monitor:**
- [ ] All agent PRs pass validation
- [ ] No false positives
- [ ] Auto-fix works correctly
- [ ] Error messages are helpful

**Actions if issues found:**
- Adjust validation rules
- Improve error messages
- Update documentation

### Weekly Review (Month 1)

**Metrics to review:**
- First-push success rate trend
- Common error patterns
- tap-tools usage rate
- Time savings

**Actions:**
- Document lessons learned
- Update validation rules for new patterns
- Improve documentation based on feedback

### Monthly Review (Ongoing)

**Strategic review:**
- Is 100% success rate maintained?
- Are agents following instructions?
- Can validation be simplified?
- New Homebrew changes to handle?

**Actions:**
- Update plan based on learnings
- Refine validation rules
- Consider additional automation

---

## 🔗 RELATED DOCUMENTS

- **[AGENTS.md](../../AGENTS.md)** - Main agent instructions (will be updated)
- **[.github/skills/homebrew-packaging/SKILL.md](../../.github/skills/homebrew-packaging/SKILL.md)** - Packaging workflow (will be updated)
- **[AGENT_BEST_PRACTICES.md](../AGENT_BEST_PRACTICES.md)** - Common errors (add PR #22)
- **[2026-02-09-pr-22-catastrophic-failure-analysis.md](../observations/2026-02-09-pr-22-catastrophic-failure-analysis.md)** - Detailed failure analysis
- **[PR #22](https://github.com/castrojo/tap/pull/22)** - Copilot's failed attempt
- **[PR #18](https://github.com/castrojo/tap/pull/18)** - Previous failure (regex)
- **[PR #19](https://github.com/castrojo/tap/pull/19)** - Previous failure (license)

---

## 📝 APPENDIX

### A. Common Semantic Error Patterns

```ruby
# ERROR 1: binary with target
binary "app", target: "#{Dir.home}/.local/bin/app"  # ❌ WRONG
binary "app"  # ✅ CORRECT (or use artifact)

# ERROR 2: Dir.home usage
"#{Dir.home}/.local/share"  # ❌ WRONG
"#{ENV.fetch("HOME")}/.local/share"  # ✅ CORRECT

# ERROR 3: Invalid depends_on
depends_on formula: "bash"  # ❌ WRONG
# (just omit for Linux casks)  # ✅ CORRECT

# ERROR 4: Hardcoded paths
"/home/user/.config/app"  # ❌ WRONG
"#{ENV.fetch("XDG_CONFIG_HOME", "#{ENV.fetch("HOME")}/.config")}/app"  # ✅ CORRECT
```

### B. Validation Workflow Decision Tree

```
PR Created with Ruby file changes
        │
        ▼
agent-validation workflow triggers
        │
        ▼
Build tap-validate (cached if possible)
        │
        ▼
Run semantic validation
        │
        ├─YES─▶ Semantic errors? ──▶ Comment on PR ──▶ FAIL (block merge)
        │
        NO
        │
        ▼
Run style validation with --fix
        │
        ├─YES─▶ Style errors found? ──▶ Auto-fix ──▶ Commit ──▶ Push
        │                                    │
        NO                                   ▼
        │                              Validation passes
        ▼
Validation passes (no changes)
        │
        ▼
CI runs (brew audit + brew style)
        │
        ├─YES─▶ CI fails? ──▶ Something wrong with validation ──▶ Investigate
        │
        NO
        │
        ▼
✅ PR ready to merge
```

### C. Quick Reference Commands

```bash
# Generate package (MANDATORY)
./tap-tools/tap-cask generate <name> <github-url>
./tap-tools/tap-formula generate <name> <github-url>

# Validate before commit (MANDATORY)
./tap-tools/tap-validate file <path> --fix

# Test locally
brew install --cask castrojo/tap/<name>  # For casks
brew install castrojo/tap/<name>         # For formulas

# Fix after CI failure
./tap-tools/tap-validate file <path> --fix
git add <path>
git commit --amend --no-edit
git push --force-with-lease

# Check validation workflow
gh run list --workflow="Agent Validation"
gh run view <run-id> --log
```

---

**Status:** Comprehensive implementation plan complete  
**Total Estimated Time:** 8-10 hours (spread over 2 days)  
**Expected ROI:** 100% CI success rate, 5-8 minutes saved per PR, zero semantic errors  
**Priority:** 🔴 CRITICAL - Blocks all reliable agent automation  
**Next Action:** Begin Phase 1 (Documentation & Preparation)
