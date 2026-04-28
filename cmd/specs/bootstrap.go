package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// cmdBootstrap scaffolds a new host with .specs-tools and .specs.yaml.
//
// Layouts and modes:
//
//	--layout submodule   adds specs/ as a submodule of the cwd
//	--layout folder      creates specs/ as a plain folder (default)
//	--at .               makes the cwd itself the specs root
//
//	--tools-mode submodule  adds .specs-tools as a submodule (default)
//	--tools-mode folder     clones .specs-tools and keeps the .git working tree
//	--tools-mode vendor     downloads a tarball at --tools-ref into .specs-tools
//
// This command intentionally does the minimum needed to land a working host;
// further scaffolding (model/, change-requests/) is gated behind --with-model.
func cmdBootstrap(args []string) error {
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	at := fs.String("at", "", "path to the specs root (created if missing); use '.' for repo root")
	layout := fs.String("layout", "folder", "how specs/ is materialised: submodule|folder")
	toolsMode := fs.String("tools-mode", "submodule", "how .specs-tools is materialised: submodule|folder|vendor")
	toolsURL := fs.String("tools-url", "https://github.com/jdehaan/specs-tools.git", "git URL of specs-tools content repo")
	toolsRef := fs.String("tools-ref", "", "tag/branch/commit for content (empty = default branch)")
	withModel := fs.Bool("with-model", false, "create empty model/ and change-requests/ skeletons")
	withVSCode := fs.Bool("with-vscode", false, "write .vscode/tasks.json")
	dryRun := fs.Bool("dry-run", false, "print actions without performing them")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs bootstrap [--at <path>] [--layout submodule|folder] [--tools-mode submodule|folder|vendor] [--tools-url URL] [--tools-ref REF] [--with-model] [--with-vscode] [--dry-run]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Resolve specs root path.
	var specsRoot string
	switch *at {
	case "":
		specsRoot = filepath.Join(cwd, "specs")
	case ".":
		specsRoot = cwd
	default:
		if filepath.IsAbs(*at) {
			specsRoot = *at
		} else {
			specsRoot = filepath.Join(cwd, *at)
		}
	}

	switch *layout {
	case "folder":
		if err := runOrLog(*dryRun, "mkdir -p", func() error { return os.MkdirAll(specsRoot, 0o755) }); err != nil {
			return err
		}
	case "submodule":
		// Add specs root as a submodule of cwd. URL not collected here (operator
		// adds the upstream specs repo separately or via --tools-url-style fork).
		return exitWith(2, "--layout submodule requires the host to register the submodule explicitly first; see docs")
	default:
		return exitWith(2, "unknown --layout %q", *layout)
	}

	// Materialise .specs-tools content.
	toolsDir := filepath.Join(specsRoot, ".specs-tools")
	switch *toolsMode {
	case "submodule":
		// Submodule must be added in the host repo's git root, not below specsRoot
		// when specsRoot is itself just a folder. Auto-detect the git root from cwd.
		hostGitRoot := findGitRoot(specsRoot)
		if hostGitRoot == "" {
			// initialise a git repo in cwd if none exists, so we can register a submodule
			if err := runOrLog(*dryRun, "git init "+cwd, func() error {
				return runGit(cwd, "init")
			}); err != nil {
				return err
			}
			hostGitRoot = cwd
		}
		rel, _ := filepath.Rel(hostGitRoot, toolsDir)
		gitArgs := []string{"submodule", "add"}
		if *toolsRef != "" {
			gitArgs = append(gitArgs, "-b", *toolsRef)
		}
		gitArgs = append(gitArgs, *toolsURL, rel)
		if err := runOrLog(*dryRun, fmt.Sprintf("git -C %s %v", hostGitRoot, gitArgs), func() error {
			return runGit(hostGitRoot, gitArgs...)
		}); err != nil {
			return err
		}
	case "folder":
		gitArgs := []string{"clone"}
		if *toolsRef != "" {
			gitArgs = append(gitArgs, "--branch", *toolsRef)
		}
		gitArgs = append(gitArgs, *toolsURL, toolsDir)
		if err := runOrLog(*dryRun, fmt.Sprintf("git %v", gitArgs), func() error {
			return runGit("", gitArgs...)
		}); err != nil {
			return err
		}
	case "vendor":
		// Vendor: same as folder, but strip the .git directory afterwards.
		gitArgs := []string{"clone", "--depth", "1"}
		if *toolsRef != "" {
			gitArgs = append(gitArgs, "--branch", *toolsRef)
		}
		gitArgs = append(gitArgs, *toolsURL, toolsDir)
		if err := runOrLog(*dryRun, fmt.Sprintf("git %v && rm -rf %s/.git", gitArgs, toolsDir), func() error {
			if err := runGit("", gitArgs...); err != nil {
				return err
			}
			return os.RemoveAll(filepath.Join(toolsDir, ".git"))
		}); err != nil {
			return err
		}
	default:
		return exitWith(2, "unknown --tools-mode %q", *toolsMode)
	}

	if *withModel {
		for _, sub := range []string{"model", "change-requests"} {
			p := filepath.Join(specsRoot, sub)
			if err := runOrLog(*dryRun, "mkdir -p "+p, func() error { return os.MkdirAll(p, 0o755) }); err != nil {
				return err
			}
		}
	}

	// Run init logic to write .specs.yaml in the new specs root.
	initArgs := []string{"--at", specsRoot, "--force"}
	if *withVSCode {
		initArgs = append(initArgs, "--with-vscode")
	}
	if *dryRun {
		fmt.Printf("would: specs init %v\n", initArgs)
		return nil
	}
	return cmdInit(initArgs)
}

func runOrLog(dryRun bool, label string, fn func() error) error {
	if dryRun {
		fmt.Println("would:", label)
		return nil
	}
	fmt.Println("$", label)
	return fn()
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findGitRoot(start string) string {
	d := start
	for {
		if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		d = parent
	}
}
