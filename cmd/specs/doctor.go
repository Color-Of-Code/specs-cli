package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/jdehaan/specs-cli/internal/config"
)

func cmdDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs doctor")
		fmt.Fprintln(os.Stderr, "Diagnose specs CLI environment, layout, and version drift.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	fmt.Printf("specs CLI:        %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
	if cfg.ConfigPath != "" {
		fmt.Printf("config file:      %s\n", cfg.ConfigPath)
	} else {
		fmt.Println("config file:      <not found> (using defaults; run `specs init` to write .specs.yaml)")
	}
	fmt.Printf("specs root:       %s\n", cfg.SpecsRoot)
	fmt.Printf("host root:        %s\n", cfg.HostRoot)
	fmt.Printf("specs mode:       %s\n", cfg.SpecsMode)
	if cfg.ToolsDir != "" {
		fmt.Printf("tools dir:        %s\n", cfg.ToolsDir)
		fmt.Printf("tools mode:       %s\n", cfg.ToolsMode)
		if rev := gitShortRev(cfg.ToolsDir); rev != "" {
			fmt.Printf("tools rev:        %s\n", rev)
		}
	} else {
		fmt.Println("tools dir:        <missing> (run `specs bootstrap --tools-mode submodule` or set tools_dir)")
	}
	fmt.Printf("model dir:        %s\n", cfg.ModelDir)
	fmt.Printf("change-requests:  %s\n", cfg.ChangeRequestsDir)
	fmt.Printf("baselines file:   %s%s\n", cfg.BaselinesFile, existsSuffix(cfg.BaselinesFile))
	fmt.Printf("markdownlint:     %s%s\n", cfg.MarkdownlintConfig, existsSuffix(cfg.MarkdownlintConfig))
	if cfg.MinSpecsVersion != "" {
		fmt.Printf("min_specs_version: %s\n", cfg.MinSpecsVersion)
	}
	fmt.Printf("repos configured: %d\n", len(cfg.Repos))

	fmt.Println("")
	fmt.Println("External tools:")
	reportTool("git", true)
	reportTool("markdownlint-cli2", false)
	reportTool("npx", false)
	reportTool("dot", false)

	if cfg.MinSpecsVersion != "" && Version != "dev" && Version < cfg.MinSpecsVersion {
		return exitWith(1, "installed CLI %s is older than min_specs_version %s", Version, cfg.MinSpecsVersion)
	}
	return nil
}

func existsSuffix(p string) string {
	if p == "" {
		return ""
	}
	if _, err := os.Stat(p); err != nil {
		return "  (missing)"
	}
	return ""
}

func reportTool(name string, required bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		mark := "optional"
		if required {
			mark = "REQUIRED"
		}
		fmt.Printf("  %-20s not found  [%s]\n", name, mark)
		return
	}
	fmt.Printf("  %-20s %s\n", name, path)
}

func gitShortRev(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(bytesTrim(out, "\n\r "))
}

// bytesTrim trims any of the bytes in cutset from both ends of b.
func bytesTrim(b []byte, cutset string) []byte {
	i, j := 0, len(b)
	for i < j && containsByte(cutset, b[i]) {
		i++
	}
	for j > i && containsByte(cutset, b[j-1]) {
		j--
	}
	return b[i:j]
}

func containsByte(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

// joinPath is a small helper used by other commands.
func joinPath(parts ...string) string { return filepath.Join(parts...) }
