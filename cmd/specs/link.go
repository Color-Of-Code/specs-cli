package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jdehaan/specs-cli/internal/config"
	"github.com/jdehaan/specs-cli/internal/linkcheck"
)

// cmdLink dispatches `specs link <subcommand>`.
func cmdLink(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs link <subcommand>")
		fmt.Fprintln(os.Stderr, "Subcommands: check")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "check":
		return cmdLinkCheck(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs link <check> [flags]")
		return nil
	default:
		return exitWith(2, "unknown link subcommand %q", args[0])
	}
}

// cmdLinkCheck verifies that every Implemented By <-> Requirements pair is
// symmetric (each forward edge has a matching reverse edge).
func cmdLinkCheck(args []string) error {
	fs := flag.NewFlagSet("link check", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprintln(os.Stderr, "Usage: specs link check") }
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	r := &linkcheck.Result{}
	linkcheck.CheckBidirectional(os.Stdout, cfg.ModelDir, r)
	for _, w := range r.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	for _, e := range r.Errors {
		fmt.Fprintln(os.Stderr, "error:", e)
	}
	if r.Failed() {
		return exitWith(1, "link check failed (%d issue(s))", len(r.Errors))
	}
	return nil
}
