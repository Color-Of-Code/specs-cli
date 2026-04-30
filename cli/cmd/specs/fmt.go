package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Color-Of-Code/specs-toolchain/cli/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/cli/internal/lint"
)

func cmdFormat(args []string) error {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	check := fs.Bool("check", false, "check formatting without writing (exit 1 if changes needed)")
	at := fs.String("at", "", "path to format (default: specs root)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs fmt [--check] [--at <path>] [files...]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Format markdown files (normalize whitespace, align tables, ensure LF endings).")
		fmt.Fprintln(os.Stderr, "Without arguments, formats all .md files under the specs root.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Collect files to format.
	var files []string
	if fs.NArg() > 0 {
		// Explicit file arguments.
		for _, f := range fs.Args() {
			abs, err := filepath.Abs(f)
			if err != nil {
				return err
			}
			files = append(files, abs)
		}
	} else {
		// Walk the specs root (or --at path).
		root := *at
		if root == "" {
			cfg, err := config.Load("")
			if err != nil {
				return err
			}
			root = cfg.SpecsRoot
		}
		abs, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		err = filepath.Walk(abs, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			rel, _ := filepath.Rel(abs, path)
			if isExcludedPath(rel) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("walk: %w", err)
		}
	}

	unformatted := 0
	for _, f := range files {
		if *check {
			data, err := os.ReadFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s: %v\n", f, err)
				continue
			}
			formatted := lint.Format(data)
			if string(data) != string(formatted) {
				fmt.Println(f)
				unformatted++
			}
		} else {
			changed, err := lint.FormatFileInPlace(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s: %v\n", f, err)
				continue
			}
			if changed {
				fmt.Println(f)
			}
		}
	}

	if *check && unformatted > 0 {
		return exitWith(1, "%d file(s) need formatting", unformatted)
	}
	return nil
}

// isExcludedPath checks common excluded directories.
func isExcludedPath(rel string) bool {
	rel = filepath.ToSlash(rel)
	excluded := []string{
		".specs-framework/", ".specs-tools/", ".git/",
		"node_modules/", ".lint/",
	}
	for _, ex := range excluded {
		if rel == strings.TrimSuffix(ex, "/") || strings.HasPrefix(rel, ex) {
			return true
		}
	}
	return false
}
