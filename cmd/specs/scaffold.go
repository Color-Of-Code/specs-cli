package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jdehaan/specs-cli/internal/config"
	"github.com/jdehaan/specs-cli/internal/tools"
)

// cmdScaffold instantiates one of the canonical templates from .specs-tools
// into either the model tree or a CR-local working tree.
//
// Usage:
//
//	specs scaffold requirement <area>/<NNN-slug>           -> model/requirements/<area>/<NNN-slug>.md
//	specs scaffold feature     <area>/<slug>               -> model/features/<area>/<slug>.md
//	specs scaffold component   <group>/<slug>              -> model/components/<group>/<slug>.md
//	specs scaffold api         <slug>                      -> model/apis/<slug>.md
//	specs scaffold service     <slug>                      -> model/services/<slug>.md
//	specs scaffold ... --cr <NNN>                          -> change-requests/CR-NNN-*/<kind>s/<...>.md
//
// --title overrides the H1; --force overwrites an existing file; --dry-run
// prints actions without writing.
func cmdScaffold(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs scaffold <kind> [--cr <NNN>] [--title <title>] [--force] [--dry-run] <path>")
		fmt.Fprintln(os.Stderr, "Kinds: requirement, feature, component, api, service")
		return exitWith(2, "missing kind")
	}
	kind := args[0]
	rest := args[1:]

	fs := flag.NewFlagSet("scaffold "+kind, flag.ContinueOnError)
	cr := fs.String("cr", "", "drop the file under change-requests/CR-<NNN>-* instead of model/")
	title := fs.String("title", "", "override the H1 title (default: derived from slug)")
	force := fs.Bool("force", false, "overwrite an existing file")
	dryRun := fs.Bool("dry-run", false, "print actions without writing")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs scaffold "+kind+" [--cr <NNN>] [--title <title>] [--force] [--dry-run] <path>")
		fs.PrintDefaults()
	}
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return exitWith(2, "missing path (place flags before the path)")
	}
	relPath := fs.Arg(0)

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if cfg.SpecsRoot == "" {
		return exitWith(2, "could not determine specs root")
	}
	if cfg.ToolsMode == config.ToolsModeManaged {
		if _, err := tools.Ensure(cfg.ToolsURL, cfg.ToolsRef); err != nil {
			return exitWith(1, "fetch managed tools: %v", err)
		}
	}
	if cfg.ToolsDir == "" {
		return exitWith(1, "tools dir not available; run `specs bootstrap` or set tools_dir/tools_url")
	}

	tplName, dirName, ok := scaffoldKindMap(kind)
	if !ok {
		return exitWith(2, "unknown kind %q (want: requirement|feature|component|api|service)", kind)
	}
	tplPath := filepath.Join(cfg.ToolsDir, "templates", tplName)
	if _, err := os.Stat(tplPath); err != nil {
		return exitWith(1, "template %s not found: %v", tplPath, err)
	}

	relPath = strings.TrimSuffix(relPath, ".md") + ".md"

	var destBase string
	if *cr != "" {
		crDir, err := findCRDir(cfg.ChangeRequestsDir, *cr)
		if err != nil {
			return err
		}
		destBase = filepath.Join(crDir, dirName)
	} else {
		destBase = filepath.Join(cfg.ModelDir, dirName)
	}
	destPath := filepath.Join(destBase, relPath)

	if !*force {
		if _, err := os.Stat(destPath); err == nil {
			return exitWith(1, "%s already exists (use --force)", destPath)
		}
	}

	derivedTitle := deriveTitle(relPath, *title)
	if *dryRun {
		fmt.Printf("would: write %s (title=%q)\n", destPath, derivedTitle)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	if err := writeTemplate(tplPath, destPath, derivedTitle); err != nil {
		return err
	}
	fmt.Println("wrote", destPath)
	return nil
}

func scaffoldKindMap(kind string) (tpl, dir string, ok bool) {
	switch kind {
	case "requirement":
		return "requirement.md", "requirements", true
	case "feature":
		return "feature.md", "features", true
	case "component":
		return "component.md", "components", true
	case "api":
		return "api.md", "apis", true
	case "service":
		return "service.md", "services", true
	}
	return "", "", false
}

// findCRDir returns the absolute path of change-requests/CR-<id>-* (single
// match expected). The id may be supplied with or without leading zeros.
func findCRDir(crRoot, id string) (string, error) {
	id = strings.TrimPrefix(id, "CR-")
	for len(id) < 3 {
		id = "0" + id
	}
	prefix := "CR-" + id + "-"
	entries, err := os.ReadDir(crRoot)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", crRoot, err)
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			return filepath.Join(crRoot, e.Name()), nil
		}
	}
	return "", exitWith(1, "no CR matching %s* in %s", prefix, crRoot)
}

// deriveTitle turns a slug-style path into a Title Case heading. Numeric
// prefixes (e.g. "012-foo-bar") are stripped from the title text.
func deriveTitle(relPath, override string) string {
	if override != "" {
		return override
	}
	base := strings.TrimSuffix(filepath.Base(relPath), ".md")
	// strip leading "NNN-" if present
	if i := strings.IndexByte(base, '-'); i > 0 && allDigits(base[:i]) {
		base = base[i+1:]
	}
	parts := strings.FieldsFunc(base, func(r rune) bool { return r == '-' || r == '_' })
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// writeTemplate copies tplPath to destPath replacing the first H1 with
// "# <title>". The body of the template is preserved verbatim.
func writeTemplate(tplPath, destPath, title string) error {
	src, err := os.ReadFile(tplPath)
	if err != nil {
		return err
	}
	out := replaceFirstH1(string(src), title)
	return os.WriteFile(destPath, []byte(out), 0o644)
}

// replaceFirstH1 replaces the first line that starts with "# " with
// "# <title>". Other content is left untouched. Returns the original
// content if no H1 is found.
func replaceFirstH1(s, title string) string {
	nl := strings.IndexByte(s, '\n')
	if nl < 0 {
		return "# " + title + "\n"
	}
	first := s[:nl]
	if strings.HasPrefix(first, "# ") {
		return "# " + title + s[nl:]
	}
	return s
}

// cmdCR dispatches `specs cr <subcommand>`.
func cmdCR(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs cr <subcommand>")
		fmt.Fprintln(os.Stderr, "Subcommands: new, status, drain")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "new":
		return cmdCRNew(args[1:])
	case "status":
		return cmdCRStatus(args[1:])
	case "drain":
		return cmdCRDrain(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs cr <new|status|drain> [flags]")
		return nil
	default:
		return exitWith(2, "unknown cr subcommand %q", args[0])
	}
}

// cmdCRNew creates change-requests/CR-<NNN>-<slug>/ from the
// templates/change-request/ tree. The CR id is normalised to 3 digits.
func cmdCRNew(args []string) error {
	fs := flag.NewFlagSet("cr new", flag.ContinueOnError)
	id := fs.String("id", "", "CR id (e.g. 4 or 004)")
	slug := fs.String("slug", "", "kebab-case slug (required)")
	title := fs.String("title", "", "human-readable title for the H1 (default: from slug)")
	force := fs.Bool("force", false, "overwrite an existing CR directory")
	dryRun := fs.Bool("dry-run", false, "print actions without writing")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs cr new --id <NNN> --slug <slug> [--title <title>] [--force] [--dry-run]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return exitWith(2, "--id is required")
	}
	if *slug == "" {
		return exitWith(2, "--slug is required")
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if cfg.ToolsMode == config.ToolsModeManaged {
		if _, err := tools.Ensure(cfg.ToolsURL, cfg.ToolsRef); err != nil {
			return exitWith(1, "fetch managed tools: %v", err)
		}
	}
	if cfg.ToolsDir == "" {
		return exitWith(1, "tools dir not available")
	}

	normID := *id
	normID = strings.TrimPrefix(normID, "CR-")
	for len(normID) < 3 {
		normID = "0" + normID
	}
	dirName := "CR-" + normID + "-" + *slug
	destDir := filepath.Join(cfg.ChangeRequestsDir, dirName)
	if !*force {
		if _, err := os.Stat(destDir); err == nil {
			return exitWith(1, "%s already exists (use --force)", destDir)
		}
	}

	srcTree := filepath.Join(cfg.ToolsDir, "templates", "change-request")
	if _, err := os.Stat(srcTree); err != nil {
		return exitWith(1, "template tree %s not found: %v", srcTree, err)
	}

	displayTitle := *title
	if displayTitle == "" {
		displayTitle = deriveTitle(*slug+".md", "")
	}

	if *dryRun {
		fmt.Printf("would: copy tree %s -> %s\n", srcTree, destDir)
		fmt.Printf("would: title=%q id=%s\n", displayTitle, normID)
		return nil
	}
	if err := copyCRTree(srcTree, destDir, normID, displayTitle); err != nil {
		return err
	}
	fmt.Println("wrote", destDir)
	return nil
}

// copyCRTree copies the change-request template tree to dest, performing
// per-file substitutions:
//   - the _index.md H1 is rewritten to "# CR-<id> — <title>"
//   - "CR-XXX" tokens in any text file become "CR-<id>"
func copyCRTree(src, dest, id, title string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		out := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		text := strings.ReplaceAll(string(data), "CR-XXX", "CR-"+id)
		if rel == "_index.md" {
			text = replaceFirstH1(text, "CR-"+id+" — "+title)
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, []byte(text), 0o644)
	})
}

// cmdCRStatus lists every CR directory and reports a one-line summary:
// id, slug, presence of _index.md, count of files in each subtree.
func cmdCRStatus(args []string) error {
	fs := flag.NewFlagSet("cr status", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprintln(os.Stderr, "Usage: specs cr status") }
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(cfg.ChangeRequestsDir)
	if err != nil {
		return err
	}
	fmt.Printf("%-8s %-40s %5s %5s %5s %5s %s\n", "ID", "Slug", "Reqs", "Feats", "Comps", "Arch", "Index")
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "CR-") {
			continue
		}
		dir := filepath.Join(cfg.ChangeRequestsDir, e.Name())
		id, slug := splitCRName(e.Name())
		idx := "-"
		if _, err := os.Stat(filepath.Join(dir, "_index.md")); err == nil {
			idx = "ok"
		}
		fmt.Printf("%-8s %-40s %5d %5d %5d %5d %s\n",
			id, truncate(slug, 40),
			countFiles(filepath.Join(dir, "requirements")),
			countFiles(filepath.Join(dir, "features")),
			countFiles(filepath.Join(dir, "components")),
			countFiles(filepath.Join(dir, "architecture")),
			idx,
		)
	}
	return nil
}

func splitCRName(name string) (id, slug string) {
	// CR-NNN-slug-bits
	rest := strings.TrimPrefix(name, "CR-")
	if i := strings.IndexByte(rest, '-'); i > 0 {
		return "CR-" + rest[:i], rest[i+1:]
	}
	return name, ""
}

func countFiles(dir string) int {
	n := 0
	_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(p, ".md") && filepath.Base(p) != "_index.md" {
			n++
		}
		return nil
	})
	return n
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
