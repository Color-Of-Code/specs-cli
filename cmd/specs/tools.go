package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdehaan/specs-cli/internal/config"
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
	if cfg.ToolsDir == "" {
		return exitWith(1, "tools_dir not found; run `specs bootstrap --tools-mode submodule`")
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
