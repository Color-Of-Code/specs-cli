package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/cache"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	fwsrc "github.com/Color-Of-Code/specs-toolchain/engine/internal/framework"
)

// cmdInit configures a host repository for use with the specs toolchain.
//
// It is git-init-like: idempotent, creates the target directory if missing,
// resolves the framework source, materialises framework content according to
// the chosen mode, and writes .specs.yaml.
//
// Usage:
//
//	specs init [<path>] [--framework <source>] [--framework-mode <mode>]
//	           [--with-model] [--with-vscode] [--force] [--dry-run]
//
// `<path>` defaults to the current directory.
//
// `--framework <source>` accepts:
//   - a registered name           (e.g. "default", "acme")
//   - a name with ref override    (e.g. "acme@v2.1")
//   - a remote git URL            (e.g. "https://github.com/foo/bar.git[@ref]",
//     "git@github.com:foo/bar.git[@ref]")
//   - a local path                (e.g. "./fw", "../specs-framework", "/abs/dir")
//
// `--framework-mode` is one of: managed (default), submodule, folder, vendor.
// Path-based sources implicitly use the existing checkout (mode is ignored).
func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	frameworkSpec := fs.String("framework", "", "framework source: registry name[@ref], git URL[@ref], or local path (default: registry's \"default\" entry)")
	frameworkMode := fs.String("framework-mode", "managed", "how .specs-framework is materialised: managed|submodule|folder|vendor (ignored for local paths)")
	withModel := fs.Bool("with-model", false, "create empty model/ and change-requests/ skeletons")
	withVSCode := fs.Bool("with-vscode", false, "write .vscode/tasks.json")
	force := fs.Bool("force", false, "overwrite an existing .specs.yaml")
	dryRun := fs.Bool("dry-run", false, "print actions without performing them")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs init [<path>] [--framework <source>] [--framework-mode managed|submodule|folder|vendor] [--with-model] [--with-vscode] [--force] [--dry-run]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Positional <path> (default: cwd).
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	specsRoot := cwd
	switch fs.NArg() {
	case 0:
		// already cwd
	case 1:
		p := fs.Arg(0)
		if filepath.IsAbs(p) {
			specsRoot = p
		} else {
			specsRoot = filepath.Join(cwd, p)
		}
	default:
		return exitWith(2, "too many positional arguments; expected at most one <path>")
	}

	// Refuse to overwrite an existing .specs.yaml without --force.
	cfgPath := filepath.Join(specsRoot, config.FileName)
	if _, err := os.Stat(cfgPath); err == nil && !*force {
		return exitWith(1, "%s already exists (use --force to overwrite)", cfgPath)
	}

	// Resolve framework source.
	src, err := fwsrc.ResolveSource(*frameworkSpec)
	if err != nil {
		return exitWith(2, "resolve framework: %v", err)
	}

	// Path-based sources skip materialisation entirely.
	if src.Path != "" {
		if err := ensureDir(specsRoot, *dryRun); err != nil {
			return err
		}
		if err := writeSpecsConfig(cfgPath, src, *dryRun); err != nil {
			return err
		}
		return finalizeInit(specsRoot, *withModel, *withVSCode, *dryRun)
	}

	// Validate mode.
	switch *frameworkMode {
	case "managed", "submodule", "folder", "vendor":
	default:
		return exitWith(2, "unknown --framework-mode %q", *frameworkMode)
	}

	if err := ensureDir(specsRoot, *dryRun); err != nil {
		return err
	}

	if err := materialiseFramework(specsRoot, src, *frameworkMode, *dryRun); err != nil {
		return err
	}

	if err := writeSpecsConfigForMode(cfgPath, src, *frameworkMode, *dryRun); err != nil {
		return err
	}

	return finalizeInit(specsRoot, *withModel, *withVSCode, *dryRun)
}

// ensureDir creates the specs root if it does not exist.
func ensureDir(dir string, dryRun bool) error {
	if _, err := os.Stat(dir); err == nil {
		return nil
	}
	return runOrLog(dryRun, "mkdir -p "+dir, func() error { return os.MkdirAll(dir, 0o755) })
}

// materialiseFramework brings the .specs-framework directory into existence
// according to the requested mode for a remote source.
func materialiseFramework(specsRoot string, src fwsrc.Source, mode string, dryRun bool) error {
	frameworkDir := filepath.Join(specsRoot, ".specs-framework")
	ref := src.Ref
	if ref == "" {
		ref = "main"
	}
	switch mode {
	case "managed":
		if dryRun {
			fmt.Printf("would: fetch %s@%s into managed cache\n", src.URL, ref)
			return nil
		}
		path, err := cache.Ensure(src.URL, ref)
		if err != nil {
			return exitWith(1, "fetch managed framework: %v", err)
		}
		fmt.Printf("managed framework cached at %s\n", path)
		return nil
	case "submodule":
		hostGitRoot := findGitRoot(specsRoot)
		if hostGitRoot == "" {
			if err := runOrLog(dryRun, "git init "+specsRoot, func() error {
				return runGit(specsRoot, "init")
			}); err != nil {
				return err
			}
			hostGitRoot = specsRoot
		}
		rel, err := filepath.Rel(hostGitRoot, frameworkDir)
		if err != nil {
			return fmt.Errorf("compute submodule path: %w", err)
		}
		gitArgs := []string{"submodule", "add", "-b", ref, src.URL, rel}
		return runOrLog(dryRun, fmt.Sprintf("git -C %s %v", hostGitRoot, gitArgs), func() error {
			return runGit(hostGitRoot, gitArgs...)
		})
	case "folder":
		gitArgs := []string{"clone", "--branch", ref, src.URL, frameworkDir}
		return runOrLog(dryRun, fmt.Sprintf("git %v", gitArgs), func() error {
			return runGit("", gitArgs...)
		})
	case "vendor":
		gitArgs := []string{"clone", "--depth", "1", "--branch", ref, src.URL, frameworkDir}
		return runOrLog(dryRun, fmt.Sprintf("git %v && rm -rf %s/.git", gitArgs, frameworkDir), func() error {
			if err := runGit("", gitArgs...); err != nil {
				return err
			}
			return os.RemoveAll(filepath.Join(frameworkDir, ".git"))
		})
	}
	return exitWith(2, "unknown --framework-mode %q", mode)
}

// writeSpecsConfigForMode writes .specs.yaml when the source is remote.
// `managed` keeps framework_url/framework_ref; the other modes record
// framework_dir relative to the specs root.
func writeSpecsConfigForMode(cfgPath string, src fwsrc.Source, mode string, dryRun bool) error {
	f := &config.File{MinSpecsVersion: Version}
	if mode == "managed" {
		f.FrameworkURL = src.URL
		f.FrameworkRef = src.Ref
		if f.FrameworkRef == "" {
			f.FrameworkRef = "main"
		}
	} else {
		f.FrameworkDir = ".specs-framework"
	}
	return saveConfig(cfgPath, f, dryRun)
}

// writeSpecsConfig writes .specs.yaml when the source is path-based.
func writeSpecsConfig(cfgPath string, src fwsrc.Source, dryRun bool) error {
	f := &config.File{
		MinSpecsVersion: Version,
		FrameworkDir:    src.Path,
	}
	return saveConfig(cfgPath, f, dryRun)
}

func saveConfig(cfgPath string, f *config.File, dryRun bool) error {
	if f.Repos == nil {
		f.Repos = map[string]string{}
	}
	if dryRun {
		fmt.Printf("would: write %s\n", cfgPath)
		return nil
	}
	if err := config.Save(cfgPath, f); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", cfgPath)
	return nil
}

// finalizeInit writes optional skeletons after the config is in place.
func finalizeInit(specsRoot string, withModel, withVSCode, dryRun bool) error {
	if withModel {
		for _, sub := range []string{"model", "change-requests"} {
			p := filepath.Join(specsRoot, sub)
			if err := runOrLog(dryRun, "mkdir -p "+p, func() error { return os.MkdirAll(p, 0o755) }); err != nil {
				return err
			}
		}
	}
	if withVSCode {
		if dryRun {
			fmt.Printf("would: write %s/.vscode/tasks.json\n", specsRoot)
			return nil
		}
		if err := writeVSCodeTasks(specsRoot); err != nil {
			return err
		}
		fmt.Println("wrote .vscode/tasks.json")
	}
	return nil
}

// runOrLog executes fn unless dryRun is set, in which case it just prints
// the label.
func runOrLog(dryRun bool, label string, fn func() error) error {
	if dryRun {
		fmt.Println("would:", label)
		return nil
	}
	fmt.Println("$", label)
	return fn()
}

// runGit invokes git with the given args, optionally chdir'd to dir.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// findGitRoot walks upward from start until a directory containing .git is
// found. Returns "" when no git root exists in any ancestor.
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
