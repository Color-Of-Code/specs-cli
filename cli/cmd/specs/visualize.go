package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Color-Of-Code/specs-cli/cli/internal/config"
	"github.com/Color-Of-Code/specs-cli/cli/internal/visualize"
)

// cmdVisualize dispatches `specs visualize <subcommand>`.
func cmdVisualize(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs visualize <subcommand>")
		fmt.Fprintln(os.Stderr, "Subcommands: traceability")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "traceability":
		return cmdVisualizeTraceability(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs visualize <traceability> [flags]")
		return nil
	default:
		return exitWith(2, "unknown visualize subcommand %q", args[0])
	}
}

// cmdVisualizeTraceability emits the requirement <-> implementer graph.
//
//	--format dot|mermaid    output format (default: dot)
//	--out <path>            file to write (default: stdout)
func cmdVisualizeTraceability(args []string) error {
	fs := flag.NewFlagSet("visualize traceability", flag.ContinueOnError)
	format := fs.String("format", "dot", "output format: dot | mermaid")
	outPath := fs.String("out", "", "write to file (use - or empty for stdout)")
	fs.StringVar(outPath, "o", "", "shorthand for --out")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs visualize traceability [--format dot|mermaid] [--out <path>|-]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	g, err := visualize.Build(cfg.ModelDir)
	if err != nil {
		return exitWith(1, "%v", err)
	}

	var out *os.File = os.Stdout
	if *outPath != "" && *outPath != "-" {
		f, err := os.Create(*outPath)
		if err != nil {
			return exitWith(1, "create %s: %v", *outPath, err)
		}
		defer f.Close()
		out = f
	}

	switch *format {
	case "dot":
		if err := visualize.WriteDOT(out, g); err != nil {
			return err
		}
	case "mermaid":
		if err := visualize.WriteMermaid(out, g); err != nil {
			return err
		}
	default:
		return exitWith(2, "unknown --format %q (want dot or mermaid)", *format)
	}
	if *outPath != "" && *outPath != "-" {
		fmt.Fprintf(os.Stderr, "wrote %s (%d node(s), %d edge(s))\n", *outPath, len(g.Nodes), len(g.Edges))
	}
	return nil
}
