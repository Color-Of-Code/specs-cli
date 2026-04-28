package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdehaan/specs-cli/cli/internal/config"
	"github.com/jdehaan/specs-cli/cli/internal/tools"
)

// cmdTools dispatches subcommands managing the .specs-tools content layer.
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

	switch cfg.ToolsMode {
	case config.ToolsModeManaged:
		return updateManaged(cfg, *to)
	}

	if cfg.ToolsDir == "" {
		return exitWith(1, "tools_dir not found; run `specs bootstrap` (managed) or set tools_dir (dev)")
	}

	switch cfg.ToolsMode {
	case config.ToolsModeSubmodule, config.ToolsModeFolder:
		if err := runGit(cfg.ToolsDir, "fetch", "--tags"); err != nil {
			return err
		}
		if *to != "" {
			if err := runGit(cfg.ToolsDir, "checkout", *to); err != nil {
				return err
			}
		} else {
			// pull on current branch; if detached, this is a no-op-ish error
			// that we report but do not fail on.
			_ = runGit(cfg.ToolsDir, "pull", "--ff-only")
		}
		if cfg.ToolsMode == config.ToolsModeSubmodule && cfg.HostRoot != "" {
			rel, _ := filepath.Rel(cfg.HostRoot, cfg.ToolsDir)
			_ = runGit(cfg.HostRoot, "add", rel)
			fmt.Println("staged submodule pointer in host; remember to commit.")
		}
		return nil
	case config.ToolsModeVendor:
		return exitWith(2, "tools_mode=vendor: re-run `specs bootstrap --tools-mode vendor --tools-ref <ref>` to refresh")
	case config.ToolsModeMissing:
		return exitWith(1, "tools_dir is missing on disk; run `specs bootstrap`")
	default:
		return exitWith(1, "unknown tools_mode %q", cfg.ToolsMode)
	}
}

// updateManaged fetches the requested ref into the user cache and rewrites
// tools_ref in .specs.yaml so subsequent invocations resolve to it.
func updateManaged(cfg *config.Resolved, to string) error {
	ref := to
	if ref == "" {
		ref = cfg.ToolsRef
	}
	if ref == "" {
		ref = "main"
	}
	path, err := tools.Ensure(cfg.ToolsURL, ref)
	if err != nil {
		return exitWith(1, "fetch %s@%s: %v", cfg.ToolsURL, ref, err)
	}
	fmt.Printf("managed tools cached at %s\n", path)

	// Rewrite tools_ref in .specs.yaml only when the caller pinned a new ref.
	if to != "" && to != cfg.ToolsRef && cfg.ConfigPath != "" && cfg.Source != nil {
		newFile := *cfg.Source
		newFile.ToolsRef = to
		if err := config.Save(cfg.ConfigPath, &newFile); err != nil {
			return exitWith(1, "write %s: %v", cfg.ConfigPath, err)
		}
		fmt.Printf("updated %s: tools_ref=%s\n", cfg.ConfigPath, to)
	}
	return nil
}
