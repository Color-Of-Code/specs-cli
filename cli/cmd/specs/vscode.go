package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Color-Of-Code/specs-toolchain/cli/internal/config"
)

// cmdVSCode dispatches `specs vscode <subcommand>`.
func cmdVSCode(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs vscode <subcommand>")
		fmt.Fprintln(os.Stderr, "Subcommands: init")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "init":
		return cmdVSCodeInit(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs vscode <init> [flags]")
		return nil
	default:
		return exitWith(2, "unknown vscode subcommand %q", args[0])
	}
}

// cmdVSCodeInit writes .vscode/tasks.json with every Specs task. By
// default, an existing tasks.json is preserved and the new file lands at
// tasks.specs.json. --force overwrites tasks.json directly.
func cmdVSCodeInit(args []string) error {
	fs := flag.NewFlagSet("vscode init", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite an existing .vscode/tasks.json")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs vscode init [--force]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if err := writeVSCodeTasksAt(cfg.HostRoot, *force); err != nil {
		return err
	}
	if *force {
		fmt.Println("wrote .vscode/tasks.json (forced)")
	} else {
		fmt.Println("wrote .vscode/tasks.json (or tasks.specs.json if one already existed)")
	}
	return nil
}
