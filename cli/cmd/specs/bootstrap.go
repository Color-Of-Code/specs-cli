package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Color-Of-Code/specs-toolchain/cli/internal/tools"
)

// cmdBootstrap scaffolds a new host with .specs.yaml pointing at the
// framework content. The default mode is `managed`: content is fetched
// once into the user's cache dir and shared across host projects.
//
//	--at .                 makes the cwd itself the specs root
//	--layout folder        creates specs/ as a plain folder (default)
//	--layout submodule     register specs/ as a git submodule of the host repo;
//	                       requires --specs-url
//
//	--tools-mode managed   (default) fetch into the user cache, hide it
//	--tools-mode submodule add .specs-framework as a submodule of the host
//	--tools-mode folder    clone .specs-framework next to specs root
//	--tools-mode vendor    snapshot .specs-framework (no .git)
func cmdBootstrap(args []string) error {
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	at := fs.String("at", "", "path to the specs root (created if missing); use '.' for repo root")
	layout := fs.String("layout", "folder", "how specs/ is materialised: submodule|folder")
	specsURL := fs.String("specs-url", "", "git URL of the host's specs repo (required for --layout submodule)")
	specsRef := fs.String("specs-ref", "", "branch/tag for --layout submodule (optional)")
	toolsMode := fs.String("tools-mode", "managed", "how .specs-framework is materialised: managed|submodule|folder|vendor")
	toolsURL := fs.String("tools-url", "https://github.com/Color-Of-Code/specs-framework.git", "git URL of specs-framework content repo")
	toolsRef := fs.String("tools-ref", "main", "tag/branch/commit for content")
	withModel := fs.Bool("with-model", false, "create empty model/ and change-requests/ skeletons")
	withVSCode := fs.Bool("with-vscode", false, "write .vscode/tasks.json")
	dryRun := fs.Bool("dry-run", false, "print actions without performing them")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs bootstrap [--at <path>] [--layout submodule|folder] [--specs-url URL] [--specs-ref REF] [--tools-mode managed|submodule|folder|vendor] [--tools-url URL] [--tools-ref REF] [--with-model] [--with-vscode] [--dry-run]")
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
		if *specsURL == "" {
			return exitWith(2, "--layout submodule requires --specs-url <git-url-of-specs-repo>")
		}
		hostGitRoot := findGitRoot(cwd)
		if hostGitRoot == "" {
			if err := runOrLog(*dryRun, "git init "+cwd, func() error {
				return runGit(cwd, "init")
			}); err != nil {
				return err
			}
			hostGitRoot = cwd
		}
		rel, err := filepath.Rel(hostGitRoot, specsRoot)
		if err != nil {
			return fmt.Errorf("compute submodule path: %w", err)
		}
		gitArgs := []string{"submodule", "add"}
		if *specsRef != "" {
			gitArgs = append(gitArgs, "-b", *specsRef)
		}
		gitArgs = append(gitArgs, *specsURL, rel)
		if err := runOrLog(*dryRun, fmt.Sprintf("git -C %s %v", hostGitRoot, gitArgs), func() error {
			return runGit(hostGitRoot, gitArgs...)
		}); err != nil {
			return err
		}
	default:
		return exitWith(2, "unknown --layout %q", *layout)
	}

	// Materialise .specs-framework content.
	toolsDir := filepath.Join(specsRoot, ".specs-framework")
	switch *toolsMode {
	case "managed":
		if *dryRun {
			fmt.Printf("would: fetch %s@%s into managed cache\n", *toolsURL, *toolsRef)
		} else {
			path, err := tools.Ensure(*toolsURL, *toolsRef)
			if err != nil {
				return exitWith(1, "fetch managed tools: %v", err)
			}
			fmt.Printf("managed tools cached at %s\n", path)
		}
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
	if *toolsMode == "managed" {
		initArgs = append(initArgs, "--tools-url", *toolsURL, "--tools-ref", *toolsRef)
	}
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
