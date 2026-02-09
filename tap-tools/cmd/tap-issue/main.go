package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/castrojo/tap-tools/internal/issues"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Styles
var (
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	warnStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	sectionStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

func printSuccess(msg string) {
	fmt.Println(successStyle.Render("✓ " + msg))
}

func printError(msg string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render("Error: "+msg))
}

func printInfo(msg string) {
	fmt.Println(infoStyle.Render("→ " + msg))
}

func printWarn(msg string) {
	fmt.Println(warnStyle.Render("⚠ " + msg))
}

func printSection(msg string) {
	fmt.Println()
	fmt.Println(sectionStyle.Render("━━━ " + msg + " ━━━"))
}

func printHighlight(msg string) {
	fmt.Println(highlightStyle.Render(msg))
}

var (
	createPR bool
	dryRun   bool
	owner    string
	repo     string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tap-issue",
		Short: "Process GitHub issues to create Homebrew packages",
		Long: `Automates package creation from GitHub issues by:
1. Parsing issue for repository URL and metadata
2. Detecting package type (formula vs cask)
3. Generating the appropriate package
4. Creating git branch and commit
5. Optionally creating PR and commenting on issue`,
	}

	processCmd := &cobra.Command{
		Use:   "process <issue-number>",
		Short: "Process a GitHub issue and create package",
		Args:  cobra.ExactArgs(1),
		RunE:  runProcess,
	}

	processCmd.Flags().BoolVar(&createPR, "create-pr", false, "Create pull request after generating package")
	processCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Parse issue and show plan without creating anything")
	processCmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (auto-detected from git remote if not specified)")
	processCmd.Flags().StringVar(&repo, "repo", "", "GitHub repository name (auto-detected from git remote if not specified)")

	rootCmd.AddCommand(processCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runProcess(cmd *cobra.Command, args []string) error {
	issueNumber, err := strconv.Atoi(args[0])
	if err != nil {
		printError("Issue number must be a positive integer")
		return err
	}

	// Preflight checks
	printSection("Preflight Checks")

	// Check for GitHub token
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		printError("GITHUB_TOKEN environment variable not set")
		return fmt.Errorf("GITHUB_TOKEN required")
	}
	printSuccess("GitHub token found")

	// Check if we're in a git repository
	if !isGitRepo() {
		printError("Not in a git repository")
		return fmt.Errorf("must be run from git repository")
	}
	printSuccess("Git repository detected")

	// Auto-detect owner/repo from git remote if not specified
	if owner == "" || repo == "" {
		detectedOwner, detectedRepo, err := getGitHubRepo()
		if err != nil {
			printError("Could not determine GitHub repository from git remote")
			return err
		}
		owner = detectedOwner
		repo = detectedRepo
	}
	printSuccess(fmt.Sprintf("Repository: %s/%s", owner, repo))

	// Fetch and parse issue
	printSection(fmt.Sprintf("Fetching Issue #%d", issueNumber))

	client := issues.NewClient()

	printInfo("Fetching issue data...")
	req, err := client.GetIssue(owner, repo, issueNumber)
	if err != nil {
		printError(fmt.Sprintf("Failed to fetch issue: %v", err))
		return err
	}

	printSuccess(fmt.Sprintf("Issue: %s", req.Title))
	printInfo(fmt.Sprintf("State: %s", req.State))
	printInfo(fmt.Sprintf("URL: %s", req.URL))

	if req.State == "closed" {
		printWarn("Issue is already closed. Continuing anyway...")
	}

	printSection("Package Detection")

	printSuccess(fmt.Sprintf("Repository URL: %s", req.RepoURL))
	printSuccess(fmt.Sprintf("Package Name: %s", req.PackageName))
	printSuccess(fmt.Sprintf("Package Type: %s", req.PackageType))

	if req.Description != "" {
		printInfo(fmt.Sprintf("Description: %s", req.Description))
	}

	// Dry run - show plan and exit
	if dryRun {
		printSection("Dry Run - Plan")
		fmt.Println()
		printHighlight("Would execute:")
		fmt.Printf("  1. Create branch: package-request-%d-%s\n", issueNumber, req.PackageName)

		var targetFile string
		if req.PackageType == issues.PackageTypeCask {
			targetFile = fmt.Sprintf("Casks/%s.rb", req.PackageName)
			fmt.Printf("  2. Generate cask: %s\n", targetFile)
		} else {
			targetFile = fmt.Sprintf("Formula/%s.rb", req.PackageName)
			fmt.Printf("  2. Generate formula: %s\n", targetFile)
		}

		fmt.Printf("  3. Commit: feat: add %s %s (closes #%d)\n", req.PackageName, req.PackageType, issueNumber)
		fmt.Printf("  4. Push to origin\n")

		if createPR {
			fmt.Printf("  5. Create pull request\n")
			fmt.Printf("  6. Comment on issue #%d\n", issueNumber)
		}

		return nil
	}

	// Create git branch
	printSection("Creating Git Branch")

	branchName := fmt.Sprintf("package-request-%d-%s", issueNumber, req.PackageName)
	branchName = normalizeBranchName(branchName)

	if branchExists(branchName) {
		printWarn(fmt.Sprintf("Branch %s already exists", branchName))
		printInfo("Checking out existing branch...")
		if err := runCommand("git", "checkout", branchName); err != nil {
			printError("Failed to checkout existing branch")
			return err
		}
	} else {
		printInfo(fmt.Sprintf("Creating branch: %s", branchName))
		if err := runCommand("git", "checkout", "-b", branchName); err != nil {
			printError("Failed to create branch")
			return err
		}
	}
	printSuccess(fmt.Sprintf("On branch: %s", branchName))

	// Generate package
	printSection("Generating Package")

	var targetFile string
	if req.PackageType == issues.PackageTypeCask {
		printInfo("Generating cask...")
		targetFile = fmt.Sprintf("Casks/%s.rb", req.PackageName)

		// Run tap-cask generate
		caskCmd := exec.Command("./tap-cask", "generate", req.RepoURL)
		caskCmd.Dir = filepath.Join(mustGetWorkingDir(), "tap-tools")
		caskCmd.Stdout = os.Stdout
		caskCmd.Stderr = os.Stderr

		if err := caskCmd.Run(); err != nil {
			printError(fmt.Sprintf("Failed to generate cask: %v", err))
			return err
		}
	} else {
		printInfo("Generating formula...")
		targetFile = fmt.Sprintf("Formula/%s.rb", req.PackageName)

		// Run tap-formula generate
		formulaCmd := exec.Command("./tap-formula", "generate", req.RepoURL)
		formulaCmd.Dir = filepath.Join(mustGetWorkingDir(), "tap-tools")
		formulaCmd.Stdout = os.Stdout
		formulaCmd.Stderr = os.Stderr

		if err := formulaCmd.Run(); err != nil {
			printError(fmt.Sprintf("Failed to generate formula: %v", err))
			return err
		}
	}
	printSuccess("Package generated successfully")

	// Commit changes
	printSection("Committing Changes")

	printInfo(fmt.Sprintf("Staging %s...", targetFile))
	if err := runCommand("git", "add", targetFile); err != nil {
		printError(fmt.Sprintf("Failed to stage %s", targetFile))
		return err
	}

	commitMsg := fmt.Sprintf("feat: add %s %s (closes #%d)\n\nAssisted-by: Claude 3.5 Sonnet via OpenCode",
		req.PackageName, req.PackageType, issueNumber)
	printInfo(fmt.Sprintf("Creating commit: feat: add %s %s (closes #%d)", req.PackageName, req.PackageType, issueNumber))

	if err := runCommand("git", "commit", "-m", commitMsg); err != nil {
		printError("Failed to commit")
		return err
	}
	printSuccess("Changes committed")

	// Push to remote
	printSection("Pushing to Remote")

	printInfo("Pushing branch to remote...")
	if err := runCommand("git", "push", "-u", "origin", branchName); err != nil {
		printError("Failed to push branch")
		return err
	}
	printSuccess(fmt.Sprintf("Branch pushed to origin/%s", branchName))

	// Summary
	printSection("Summary")
	fmt.Println()
	printHighlight("Package Details:")
	fmt.Printf("  Name:        %s\n", req.PackageName)
	fmt.Printf("  Type:        %s\n", req.PackageType)
	fmt.Printf("  Repository:  %s\n", req.RepoURL)
	fmt.Printf("  File:        %s\n", targetFile)
	fmt.Println()
	printHighlight("Git Details:")
	fmt.Printf("  Branch:      %s\n", branchName)
	fmt.Printf("  Commit:      feat: add %s %s (closes #%d)\n", req.PackageName, req.PackageType, issueNumber)
	fmt.Println()

	// Create PR if requested
	if createPR {
		printSection("Creating Pull Request")

		prTitle := fmt.Sprintf("feat(%s): add %s", req.PackageType, req.PackageName)
		prBody := fmt.Sprintf(`## Summary

This PR adds the `+"`%s`"+` %s to the tap.

**Package Information:**
- Name: `+"`%s`"+`
- Type: %s
- Repository: %s
- Source Issue: #%d

**Generated by:** `+"`tap-issue`"+`

Closes #%d`, req.PackageName, req.PackageType, req.PackageName, req.PackageType, req.RepoURL, issueNumber, issueNumber)

		printInfo("Creating pull request...")
		// Get default branch (typically "main")
		prURL, err := client.CreatePullRequest(owner, repo, branchName, "main", prTitle, prBody)
		if err != nil {
			printError(fmt.Sprintf("Failed to create PR: %v", err))
			return err
		}
		printSuccess(fmt.Sprintf("Pull request created: %s", prURL))

		// Comment on issue
		printInfo(fmt.Sprintf("Commenting on issue #%d...", issueNumber))
		commentBody := fmt.Sprintf("✅ Package %s has been generated and a pull request has been created: %s\n\nThe %s will be available once the PR is reviewed and merged.",
			req.PackageType, prURL, req.PackageType)

		if err := client.CommentOnIssue(owner, repo, issueNumber, commentBody); err != nil {
			printWarn("Failed to comment on issue")
		}

		fmt.Println()
		printHighlight("Next Steps:")
		fmt.Printf("  1. Review the PR: %s\n", prURL)
		fmt.Printf("  2. Test the %s locally\n", req.PackageType)
		fmt.Printf("  3. Merge the PR to publish the package\n")
	} else {
		fmt.Println()
		printHighlight("Next Steps:")
		fmt.Printf("  1. Review the generated %s: %s\n", req.PackageType, targetFile)

		caskFlag := ""
		if req.PackageType == issues.PackageTypeCask {
			caskFlag = "--cask "
		}
		fmt.Printf("  2. Test locally: brew install %s%s\n", caskFlag, req.PackageName)
		fmt.Printf("  3. Create a PR manually: gh pr create --fill\n")
		fmt.Printf("  4. Or run with --create-pr flag: tap-issue process %d --create-pr\n", issueNumber)
	}

	fmt.Println()
	fmt.Println(successStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(successStyle.Render("✓ Automation complete!"))
	fmt.Println(successStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	return nil
}

// Helper functions

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func getGitHubRepo() (string, string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	repoURL := strings.TrimSpace(string(output))

	// Parse owner/repo from git remote URL
	// Handles both https://github.com/owner/repo.git and git@github.com:owner/repo.git
	var owner, repo string

	if strings.Contains(repoURL, "github.com") {
		// Remove .git suffix if present
		repoURL = strings.TrimSuffix(repoURL, ".git")

		// Extract owner/repo
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 2 {
			repo = parts[len(parts)-1]
			owner = strings.TrimPrefix(parts[len(parts)-2], ":")
		}
	}

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("could not parse GitHub owner/repo from remote URL: %s", repoURL)
	}

	return owner, repo, nil
}

func branchExists(branchName string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branchName)
	return cmd.Run() == nil
}

func normalizeBranchName(name string) string {
	name = strings.ToLower(name)
	// Replace any non-alphanumeric characters (except hyphens) with hyphens
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		} else {
			result.WriteRune('-')
		}
	}
	return result.String()
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func mustGetWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}
