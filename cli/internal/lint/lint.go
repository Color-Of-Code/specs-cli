// Package lint ports the .specs-tools/lint/lint.sh checks to Go. It exposes
// modular check functions and a Result type so callers can compose modes
// (--all, --links, --style, --baselines).
package lint

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Color-Of-Code/specs-cli/cli/internal/config"
)

// Result aggregates findings from a lint run.
type Result struct {
	Errors   []string
	Warnings []string
}

// Failed reports whether any error-level issues were recorded.
func (r *Result) Failed() bool { return len(r.Errors) > 0 }

func (r *Result) errf(format string, a ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, a...))
}
func (r *Result) warnf(format string, a ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, a...))
}

// Excluded paths (relative to specs root) that mirror the bash lint script
// and the markdownlint config ignore list.
var excludedPathPrefixes = []string{
	".specs-tools" + string(os.PathSeparator),
	".lint" + string(os.PathSeparator),
	"node_modules" + string(os.PathSeparator),
	".git" + string(os.PathSeparator),
}

func isExcludedRel(rel string) bool {
	rel = filepath.ToSlash(rel)
	for _, p := range excludedPathPrefixes {
		p2 := filepath.ToSlash(p)
		if rel == strings.TrimSuffix(p2, "/") || strings.HasPrefix(rel, p2) {
			return true
		}
	}
	return false
}

// CheckSymlinks finds broken symlinks under specsRoot, mirroring
// `find -xtype l` semantics.
func CheckSymlinks(out io.Writer, specsRoot string, r *Result) {
	fmt.Fprintln(out, "== broken symlinks ==")
	count := 0
	err := filepath.Walk(specsRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// Skip unreadable trees rather than aborting.
			return nil
		}
		rel, _ := filepath.Rel(specsRoot, path)
		if rel == "." {
			return nil
		}
		if isExcludedRel(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if _, err := os.Stat(path); err != nil {
				r.errf("broken symlink: %s", rel)
				count++
			}
		}
		return nil
	})
	if err != nil {
		r.errf("walk specs root: %v", err)
		return
	}
	if count == 0 {
		fmt.Fprintln(out, "ok")
	}
}

// linkRe extracts inline markdown link targets like ](target).
var linkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

// CheckMarkdownLinks verifies that every relative markdown link target
// resolves to an existing file or directory.
func CheckMarkdownLinks(out io.Writer, specsRoot string, r *Result) {
	fmt.Fprintln(out, "== markdown link targets ==")
	count := 0
	err := filepath.Walk(specsRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(specsRoot, path)
		if rel == "." {
			return nil
		}
		if isExcludedRel(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip the change-request template subtree as in the bash version.
		if strings.HasPrefix(filepath.ToSlash(rel), "change-requests/_template/") {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		// Skip symlinks: bash `find -type f` does not match them, and the
		// broken-symlink check is already a separate pass.
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if filepath.Base(path) == "_template.md" {
			return nil
		}
		broken := linksInFile(path)
		baseDir := filepath.Dir(path)
		for _, target := range broken {
			clean := stripFragment(target)
			if clean == "" || strings.HasPrefix(clean, "/") {
				continue
			}
			abs := filepath.Join(baseDir, clean)
			if _, err := os.Stat(abs); err != nil {
				r.errf("broken link: %s -> %s", rel, target)
				count++
			}
		}
		return nil
	})
	if err != nil {
		r.errf("walk specs root: %v", err)
		return
	}
	if count == 0 {
		fmt.Fprintln(out, "ok")
	}
}

// linksInFile returns the targets of every relative inline markdown link
// in path, skipping fenced code blocks. URL-only and anchor-only targets
// are filtered out.
func linksInFile(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var targets []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	fenced := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		for _, m := range linkRe.FindAllStringSubmatch(line, -1) {
			t := strings.TrimSpace(m[1])
			if t == "" {
				continue
			}
			switch {
			case strings.HasPrefix(t, "http://"),
				strings.HasPrefix(t, "https://"),
				strings.HasPrefix(t, "mailto:"),
				strings.HasPrefix(t, "#"),
				strings.HasPrefix(t, "<"):
				continue
			}
			targets = append(targets, t)
		}
	}
	return targets
}

func stripFragment(s string) string {
	if i := strings.IndexAny(s, "#?"); i >= 0 {
		return s[:i]
	}
	return s
}

// CheckMarkdownStyle shells out to markdownlint-cli2 (preferred) or npx
// markdownlint-cli2. It is a no-op (warning only) if neither is available.
func CheckMarkdownStyle(out io.Writer, specsRoot, configPath string, r *Result) {
	fmt.Fprintln(out, "== markdownlint ==")
	if configPath == "" {
		r.warnf("markdownlint config not configured; skipping style check")
		return
	}
	if _, err := os.Stat(configPath); err != nil {
		r.warnf("markdownlint config not found at %s; skipping", configPath)
		return
	}
	var cmd *exec.Cmd
	if path, err := exec.LookPath("markdownlint-cli2"); err == nil {
		cmd = exec.Command(path, "--config", configPath, "**/*.md")
	} else if path, err := exec.LookPath("npx"); err == nil {
		fmt.Fprintln(out, "using npx markdownlint-cli2 (first run downloads the package)")
		cmd = exec.Command(path, "--yes", "markdownlint-cli2", "--config", configPath, "**/*.md")
	} else {
		r.warnf("markdownlint-cli2 and npx not available; skipping style check")
		r.warnf("install with: npm i -g markdownlint-cli2  (or install Node.js for npx)")
		return
	}
	cmd.Dir = specsRoot
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		r.errf("markdownlint: %v", err)
		return
	}
	fmt.Fprintln(out, "ok")
}

// baselineRowRe matches a markdown table row in the Components section of the
// baseline file. The columns are: | Component | Repo | Path | Commit | ... |.
var baselineRowRe = regexp.MustCompile(`^\|\s*\[`)

// CheckBaselines verifies that every component row in the baseline file
// records the latest git commit SHA for the referenced (repo, path).
func CheckBaselines(out io.Writer, cfg *config.Resolved, r *Result) {
	fmt.Fprintln(out, "== component baselines ==")
	if cfg.BaselinesFile == "" {
		r.warnf("no baseline file configured; skipping")
		return
	}
	if _, err := os.Stat(cfg.BaselinesFile); err != nil {
		r.warnf("no baseline file at %s; skipping", cfg.BaselinesFile)
		return
	}
	if len(cfg.Repos) == 0 {
		r.warnf("no repos: map in .specs.yaml; baseline rows will be skipped")
	}

	workspace := filepath.Dir(cfg.HostRoot)

	f, err := os.Open(cfg.BaselinesFile)
	if err != nil {
		r.errf("open baseline file: %v", err)
		return
	}
	defer f.Close()

	checked := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	inSection := false
	for scanner.Scan() {
		line := scanner.Text()
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "## ") && strings.Contains(t, "Components") {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(t, "## ") {
			break
		}
		if !inSection || !baselineRowRe.MatchString(line) {
			continue
		}
		cols := splitTableRow(line)
		if len(cols) < 5 {
			continue
		}
		// Columns (1-indexed to mirror awk -F'|'): 1=Component, 2=Repo, 3=Path, 4=Commit.
		repo := stripBackticks(strings.TrimSpace(cols[2]))
		pathInside := stripBackticks(strings.TrimSpace(cols[3]))
		recorded := stripBackticks(strings.TrimSpace(cols[4]))
		if pathInside == "" || strings.HasPrefix(pathInside, "_") || strings.HasPrefix(pathInside, "(") {
			continue
		}
		repoPath, ok := cfg.Repos[repo]
		if !ok {
			r.warnf("unknown repo %q in baseline table; add it under repos: in .specs.yaml", repo)
			continue
		}
		absRepo := filepath.Join(workspace, repoPath)
		if !isGitRepo(absRepo) {
			r.warnf("repo not checked out: %s (skipping %s)", repoPath, repo)
			continue
		}
		args := []string{"-C", absRepo, "log", "-1", "--format=%H"}
		if pathInside != "/" {
			args = append(args, "--", pathInside)
		}
		cmdOut, err := exec.Command("git", args...).Output()
		actual := strings.TrimSpace(string(cmdOut))
		if err != nil || actual == "" {
			r.warnf("no git history for %s:%s", repo, pathInside)
			continue
		}
		checked++
		if actual != recorded {
			r.errf("baseline stale: %s path=%q\n  recorded: %s\n  actual:   %s",
				repo, pathInside, recorded, actual)
		}
	}
	if !r.Failed() {
		fmt.Fprintf(out, "ok (%d component(s) verified)\n", checked)
	}
}

func splitTableRow(line string) []string {
	// drop the leading and trailing pipe so split returns clean cells
	t := strings.TrimSpace(line)
	t = strings.TrimPrefix(t, "|")
	t = strings.TrimSuffix(t, "|")
	parts := strings.Split(t, "|")
	out := make([]string, 0, len(parts)+1)
	out = append(out, "") // 1-based column indexing to match awk -F'|'
	out = append(out, parts...)
	return out
}

func stripBackticks(s string) string {
	return strings.Trim(s, "` ")
}

func isGitRepo(dir string) bool {
	if st, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		_ = st
		return true
	}
	return false
}
