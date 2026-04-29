package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Color-Of-Code/specs-toolchain/cli/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/cli/internal/lint"
)

// cmdBaseline dispatches `specs baseline <subcommand>`.
func cmdBaseline(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs baseline <subcommand>")
		fmt.Fprintln(os.Stderr, "Subcommands: check, update")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "check":
		return cmdBaselineCheck(args[1:])
	case "update":
		return cmdBaselineUpdate(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs baseline <check|update> [flags]")
		return nil
	default:
		return exitWith(2, "unknown baseline subcommand %q", args[0])
	}
}

// cmdBaselineCheck is a focused alias for `specs lint --baselines`.
func cmdBaselineCheck(args []string) error {
	fs := flag.NewFlagSet("baseline check", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprintln(os.Stderr, "Usage: specs baseline check") }
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	r := &lint.Result{}
	lint.CheckBaselines(os.Stdout, cfg, r)
	for _, w := range r.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	for _, e := range r.Errors {
		fmt.Fprintln(os.Stderr, "error:", e)
	}
	if r.Failed() {
		return exitWith(1, "baseline check failed")
	}
	return nil
}

// baselineRowRe matches a Components-table row whose first cell is a link.
var baselineRowRe = regexp.MustCompile(`^\|\s*\[`)

// cmdBaselineUpdate rewrites the Components table in the baseline file:
// for every link-rowed entry, it queries git for the actual last-touching
// commit and replaces the recorded SHA. Other lines are preserved.
//
// Flags:
//
//	--only <component-substring>   restrict the update to rows whose first
//	                               cell contains the substring (e.g. a slug)
//	--dry-run                      print proposed changes only
func cmdBaselineUpdate(args []string) error {
	fs := flag.NewFlagSet("baseline update", flag.ContinueOnError)
	only := fs.String("only", "", "only update rows whose first cell contains this substring")
	dryRun := fs.Bool("dry-run", false, "print the proposed table without writing")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs baseline update [--only <substr>] [--dry-run]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if cfg.BaselinesFile == "" {
		return exitWith(1, "no baselines file configured")
	}
	data, err := os.ReadFile(cfg.BaselinesFile)
	if err != nil {
		return exitWith(1, "read %s: %v", cfg.BaselinesFile, err)
	}

	workspace := filepath.Dir(cfg.HostRoot)
	lines := strings.Split(string(data), "\n")

	updated := 0
	skipped := 0
	inSection := false
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "## ") && strings.Contains(t, "Components") {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(t, "## ") {
			inSection = false
			continue
		}
		if !inSection || !baselineRowRe.MatchString(line) {
			continue
		}
		newLine, action, err := rewriteBaselineRow(line, cfg.Repos, workspace, *only)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			skipped++
			continue
		}
		switch action {
		case "updated":
			if newLine != line {
				if *dryRun {
					fmt.Println("- " + line)
					fmt.Println("+ " + newLine)
				}
				lines[i] = newLine
				updated++
			}
		case "skip-filter", "skip-placeholder":
			skipped++
		}
	}

	if *dryRun {
		fmt.Printf("would update %d row(s); skipped %d\n", updated, skipped)
		return nil
	}
	if updated == 0 {
		fmt.Println("no changes")
		return nil
	}
	if err := os.WriteFile(cfg.BaselinesFile, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	fmt.Printf("updated %d row(s) in %s\n", updated, cfg.BaselinesFile)
	return nil
}

// rewriteBaselineRow returns line with the SHA cell replaced by the actual
// last-touching commit for (repo, path). action is one of "updated",
// "skip-filter", "skip-placeholder".
func rewriteBaselineRow(line string, repos map[string]string, workspace, only string) (string, string, error) {
	// Preserve any leading/trailing whitespace in cells; we operate on the
	// inner content between pipes.
	t := strings.TrimRight(line, "\n\r")
	if !strings.HasPrefix(strings.TrimSpace(t), "|") {
		return line, "skip-placeholder", nil
	}
	// Split by | preserving cells; the very first and last empty entries
	// come from the leading/trailing pipes.
	cells := strings.Split(t, "|")
	if len(cells) < 6 {
		return line, "skip-placeholder", nil
	}
	component := strings.TrimSpace(cells[1])
	repoCell := strings.TrimSpace(cells[2])
	pathCell := strings.TrimSpace(cells[3])
	repo := strings.Trim(repoCell, "` ")
	pathInside := strings.Trim(pathCell, "` ")

	if pathInside == "" || strings.HasPrefix(pathInside, "_") || strings.HasPrefix(pathInside, "(") {
		return line, "skip-placeholder", nil
	}
	if only != "" && !strings.Contains(component, only) {
		return line, "skip-filter", nil
	}
	repoPath, ok := repos[repo]
	if !ok {
		return line, "", fmt.Errorf("unknown repo %q (add to repos: in .specs.yaml); component=%q", repo, component)
	}
	absRepo := filepath.Join(workspace, repoPath)
	if _, err := os.Stat(filepath.Join(absRepo, ".git")); err != nil {
		return line, "", fmt.Errorf("repo not checked out at %s: %v", absRepo, err)
	}
	gitArgs := []string{"-C", absRepo, "log", "-1", "--format=%H"}
	if pathInside != "/" {
		gitArgs = append(gitArgs, "--", pathInside)
	}
	out, err := exec.Command("git", gitArgs...).Output()
	if err != nil {
		return line, "", fmt.Errorf("git log failed for %s:%s: %v", repo, pathInside, err)
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return line, "", fmt.Errorf("no git history for %s:%s", repo, pathInside)
	}

	// Replace cell 4 (the recorded SHA). Preserve surrounding whitespace.
	old := cells[4]
	newCell := replacePreservingPadding(old, sha)
	cells[4] = newCell
	return strings.Join(cells, "|"), "updated", nil
}

// replacePreservingPadding replaces the non-whitespace content of cell with
// "`<value>`" while keeping the leading and trailing whitespace of cell.
func replacePreservingPadding(cell, value string) string {
	leading := ""
	for i := 0; i < len(cell) && (cell[i] == ' ' || cell[i] == '\t'); i++ {
		leading += string(cell[i])
	}
	trailing := ""
	for i := len(cell) - 1; i >= 0 && (cell[i] == ' ' || cell[i] == '\t'); i-- {
		trailing = string(cell[i]) + trailing
	}
	return leading + "`" + value + "`" + trailing
}

// cmdCRDrain interactively migrates files from a CR's local working tree
// into their canonical homes under model/. Each file is shown with its
// proposed destination; the user accepts (y), edits the destination (e),
// or skips (n). Accepted moves use `git mv` so the rename is tracked.
//
// Usage:
//
//	specs cr drain --id <NNN> [--yes] [--dry-run]
//
// --yes accepts every default destination without prompting.
func cmdCRDrain(args []string) error {
	fs := flag.NewFlagSet("cr drain", flag.ContinueOnError)
	id := fs.String("id", "", "CR id (required)")
	yes := fs.Bool("yes", false, "accept every default destination")
	dryRun := fs.Bool("dry-run", false, "print proposed moves only")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs cr drain --id <NNN> [--yes] [--dry-run]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return exitWith(2, "--id is required")
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	crDir, err := findCRDir(cfg.ChangeRequestsDir, *id)
	if err != nil {
		return err
	}

	type move struct{ src, dst string }
	var moves []move

	// CR-local subtrees that map 1:1 onto model/.
	for _, area := range []string{"requirements", "features", "components", "architecture"} {
		srcRoot := filepath.Join(crDir, area)
		dstRoot := filepath.Join(cfg.ModelDir, area)
		_ = filepath.Walk(srcRoot, func(p string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".md") {
				return nil
			}
			rel, err := filepath.Rel(srcRoot, p)
			if err != nil {
				return nil
			}
			moves = append(moves, move{src: p, dst: filepath.Join(dstRoot, rel)})
			return nil
		})
	}

	if len(moves) == 0 {
		fmt.Println("no files to drain")
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	accepted := 0
	for _, m := range moves {
		fmt.Println("")
		relSrc, _ := filepath.Rel(cfg.SpecsRoot, m.src)
		relDst, _ := filepath.Rel(cfg.SpecsRoot, m.dst)
		fmt.Printf("  %s\n  -> %s\n", relSrc, relDst)
		dst := m.dst
		if !*yes {
			fmt.Print("  [y]es / [n]o / [e]dit dest: ")
			ans, _ := reader.ReadString('\n')
			ans = strings.TrimSpace(ans)
			switch ans {
			case "n", "no":
				continue
			case "e", "edit":
				fmt.Print("  new destination (relative to specs root): ")
				newRel, _ := reader.ReadString('\n')
				newRel = strings.TrimSpace(newRel)
				if newRel == "" {
					continue
				}
				dst = filepath.Join(cfg.SpecsRoot, newRel)
			}
		}
		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("  destination exists; skipping\n")
			continue
		}
		if *dryRun {
			fmt.Printf("  would: git mv %s %s\n", m.src, dst)
			accepted++
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "  error: mkdir %s: %v\n", filepath.Dir(dst), err)
			continue
		}
		gitDir := cfg.HostRoot
		if gitDir == "" {
			gitDir = cfg.SpecsRoot
		}
		cmd := exec.Command("git", "-C", gitDir, "mv", m.src, dst)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Fall back to a plain rename if git mv refuses (untracked file).
			if rerr := os.Rename(m.src, dst); rerr != nil {
				fmt.Fprintf(os.Stderr, "  error: %v / %v\n", err, rerr)
				continue
			}
			fmt.Println("  moved (untracked, plain rename)")
		} else {
			fmt.Println("  moved")
		}
		accepted++
	}
	fmt.Printf("\ndrained %d file(s); review with `git status` and commit.\n", accepted)
	return nil
}
