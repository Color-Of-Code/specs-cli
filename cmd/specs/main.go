// Command specs is the user-scope CLI that operates on host repositories
// using the specs framework. See README.md for an overview.
package main

import (
	"fmt"
	"os"
)

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

type command struct {
	name    string
	summary string
	run     func(args []string) error
}

var commands []command

func init() {
	commands = []command{
		{"version", "print version and exit", cmdVersion},
		{"doctor", "diagnose environment, layout, and version drift", cmdDoctor},
		{"init", "configure an existing host (writes .specs.yaml, optional .vscode tasks)", cmdInit},
		{"bootstrap", "scaffold a new host with .specs.yaml and .specs-tools content", cmdBootstrap},
		{"lint", "run lint checks (--all|--links|--style|--baselines)", cmdLint},
		{"tools", "manage the .specs-tools content layer (update)", cmdTools},
		{"scaffold", "instantiate a template (requirement|feature|component|api|service)", cmdScaffold},
		{"cr", "change-request operations (new, status, drain)", cmdCR},
		{"baseline", "verify or update component baselines (check, update)", cmdBaseline},
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "specs %s — user-scope CLI for the specs framework\n\n", Version)
	fmt.Fprintln(os.Stderr, "Usage: specs <command> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	for _, c := range commands {
		fmt.Fprintf(os.Stderr, "  %-12s %s\n", c.name, c.summary)
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Run 'specs <command> -h' for command-specific help.")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	name := os.Args[1]
	if name == "-h" || name == "--help" || name == "help" {
		usage()
		return
	}
	if name == "-v" || name == "--version" {
		_ = cmdVersion(nil)
		return
	}
	for _, c := range commands {
		if c.name == name {
			if err := c.run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "specs %s: %v\n", name, err)
				if ee, ok := err.(*exitError); ok {
					os.Exit(ee.code)
				}
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "specs: unknown command %q\n\n", name)
	usage()
	os.Exit(2)
}

// exitError lets a command return a specific exit code.
type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

func exitWith(code int, format string, args ...any) error {
	return &exitError{code: code, msg: fmt.Sprintf(format, args...)}
}

func cmdVersion(_ []string) error {
	fmt.Println(Version)
	return nil
}
