package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/tools"
)

// cmdTools dispatches subcommands managing the .specs-framework content layer.
func cmdTools(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs tools <subcommand>")
		fmt.Fprintln(os.Stderr, "Subcommands: update")
		return exitWith(2, "missing subcommand")
	}
	sub := args[0]
	switch sub {
	case "update":
		return cmdToolsUpdate(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs tools <update> [flags]")
		return nil
	default:
		return exitWith(2, "unknown tools subcommand %q", sub)
	}
}

// cmdToolsUpdate updates the content layer in place.
//
//	submodule: git fetch + checkout, then host-side git add
//	folder:    git pull (or checkout <ref>)
//	vendor:    re-clone tarball-style at the requested ref
func cmdToolsUpdate(args []string) error {
	fs := flag.NewFlagSet("tools update", flag.ContinueOnError)
	to := fs.String("to", "", "tag/branch/commit to check out (empty = pull current branch / default branch)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs tools update [--to <ref>]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	switch cfg.FrameworkMode {
	case config.FrameworkModeManaged:
		return updateManaged(cfg, *to)
	}

	if cfg.FrameworkDir == "" {
		return exitWith(1, "framework_dir not found; run `specs bootstrap` (managed) or set framework_dir (dev)")
	}

	switch cfg.FrameworkMode {
	case config.FrameworkModeSubmodule, config.FrameworkModeFolder:
		if err := runGit(cfg.FrameworkDir, "fetch", "--tags"); err != nil {
			return err
		}
		if *to != "" {
			if err := runGit(cfg.FrameworkDir, "checkout", *to); err != nil {
				return err
			}
		} else {
			// pull on current branch; if detached, this is a no-op-ish error
			// that we report but do not fail on.
			_ = runGit(cfg.FrameworkDir, "pull", "--ff-only")
		}
		if cfg.FrameworkMode == config.FrameworkModeSubmodule && cfg.HostRoot != "" {
			rel, _ := filepath.Rel(cfg.HostRoot, cfg.FrameworkDir)
			_ = runGit(cfg.HostRoot, "add", rel)
			fmt.Println("staged submodule pointer in host; remember to commit.")
		}
		return nil
	case config.FrameworkModeVendor:
		return exitWith(2, "framework_mode=vendor: re-run `specs bootstrap --framework-mode vendor --framework-ref <ref>` to refresh")
	case config.FrameworkModeMissing:
		return exitWith(1, "framework_dir is missing on disk; run `specs bootstrap`")
	default:
		return exitWith(1, "unknown framework_mode %q", cfg.FrameworkMode)
	}
}

// updateManaged fetches the requested ref into the user cache and rewrites
// framework_ref in .specs.yaml so subsequent invocations resolve to it.
func updateManaged(cfg *config.Resolved, to string) error {
	ref := to
	if ref == "" {
		ref = cfg.FrameworkRef
	}
	if ref == "" {
		ref = "main"
	}
	path, err := tools.Ensure(cfg.FrameworkURL, ref)
	if err != nil {
		return exitWith(1, "fetch %s@%s: %v", cfg.FrameworkURL, ref, err)
	}
	fmt.Printf("managed framework cached at %s\n", path)

	// Rewrite framework_ref in .specs.yaml only when the caller pinned a new ref.
	if to != "" && to != cfg.FrameworkRef && cfg.ConfigPath != "" && cfg.Source != nil {
		newFile := *cfg.Source
		newFile.FrameworkRef = to
		if err := config.Save(cfg.ConfigPath, &newFile); err != nil {
			return exitWith(1, "write %s: %v", cfg.ConfigPath, err)
		}
		fmt.Printf("updated %s: framework_ref=%s\n", cfg.ConfigPath, to)
	}
	return nil
}
