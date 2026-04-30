package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/lint"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/tools"
)

func cmdLint(args []string) error {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	all := fs.Bool("all", false, "run every check (default)")
	links := fs.Bool("links", false, "check broken symlinks and markdown link targets")
	style := fs.Bool("style", false, "run markdown style checks")
	baselines := fs.Bool("baselines", false, "verify component baseline SHAs")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs lint [--all|--links|--style|--baselines]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*links && !*style && !*baselines {
		*all = true
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if cfg.SpecsRoot == "" {
		return exitWith(2, "could not determine specs root; run from within a specs repo or pass via .specs.yaml")
	}

	// Managed mode: fetch into the user cache on first use.
	if cfg.FrameworkMode == config.FrameworkModeManaged {
		if _, err := tools.Ensure(cfg.FrameworkURL, cfg.FrameworkRef); err != nil {
			return exitWith(1, "fetch managed framework: %v", err)
		}
	}

	r := &lint.Result{}
	if *all || *links {
		lint.CheckSymlinks(os.Stdout, cfg.SpecsRoot, r)
		lint.CheckMarkdownLinks(os.Stdout, cfg.SpecsRoot, r)
	}
	if *all || *style {
		lint.CheckMarkdownStyle(os.Stdout, cfg.SpecsRoot, cfg.StyleConfig, r)
	}
	if *all || *baselines {
		lint.CheckBaselines(os.Stdout, cfg, r)
	}

	for _, w := range r.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	for _, e := range r.Errors {
		fmt.Fprintln(os.Stderr, "error:", e)
	}
	if r.Failed() {
		fmt.Println("== FAIL ==")
		return exitWith(1, "lint failed")
	}
	fmt.Println("== PASS ==")
	return nil
}
