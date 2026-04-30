package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/cache"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	fwsrc "github.com/Color-Of-Code/specs-toolchain/engine/internal/framework"
)

// cmdInit configures a host repository for use with the specs toolchain.
//
// It is git-init-like: idempotent, creates the target directory if missing,
// resolves the framework source, and writes .specs.yaml.
//
// Usage:
//
//	specs init [<path>] [--framework <source>]
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
// Remote sources are fetched into the user cache (managed mode); the host
// commits only `.specs.yaml`. Local paths are recorded in `framework_dir`
// and left untouched, so the user can keep the framework as a plain folder,
// a git submodule, or a vendored snapshot — whichever fits the host.
func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	frameworkSpec := fs.String("framework", "", "framework source: registry name[@ref], git URL[@ref], or local path (default: registry's \"default\" entry)")
	withModel := fs.Bool("with-model", false, "create empty model/ and change-requests/ skeletons")
	withVSCode := fs.Bool("with-vscode", false, "write .vscode/tasks.json")
	force := fs.Bool("force", false, "overwrite an existing .specs.yaml")
	dryRun := fs.Bool("dry-run", false, "print actions without performing them")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs init [<path>] [--framework <source>] [--with-model] [--with-vscode] [--force] [--dry-run]")
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

	src, err := fwsrc.ResolveSource(*frameworkSpec)
	if err != nil {
		return exitWith(2, "resolve framework: %v", err)
	}

	if err := ensureDir(specsRoot, *dryRun); err != nil {
		return err
	}

	f := &config.File{MinSpecsVersion: Version}
	if src.Path != "" {
		// Local source: record framework_dir, do not materialise anything.
		f.FrameworkDir = src.Path
	} else {
		// Remote source: fetch into the managed cache, record url+ref.
		ref := src.Ref
		if ref == "" {
			ref = "main"
		}
		if *dryRun {
			fmt.Printf("would: fetch %s@%s into managed cache\n", src.URL, ref)
		} else {
			path, err := cache.Ensure(src.URL, ref)
			if err != nil {
				return exitWith(1, "fetch managed framework: %v", err)
			}
			fmt.Printf("managed framework cached at %s\n", path)
		}
		f.FrameworkURL = src.URL
		f.FrameworkRef = ref
	}

	if err := saveConfig(cfgPath, f, *dryRun); err != nil {
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
