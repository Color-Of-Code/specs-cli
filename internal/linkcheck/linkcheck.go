// Package linkcheck verifies bidirectional consistency between requirements
// and the features/components/APIs/services that implement them.
//
// Conventions in the framework:
//
//   - A requirement file (model/requirements/<area>/<NNN>-<slug>.md) has an
//     `## Implemented By` section listing the documents that implement it.
//   - A feature/component/api/service file has a `## Requirements` section
//     listing the requirements it implements.
//
// This package finds asymmetric links: a forward edge that lacks a matching
// reverse edge.
package linkcheck

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Result aggregates findings.
type Result struct {
	Errors   []string
	Warnings []string
}

func (r *Result) Failed() bool { return len(r.Errors) > 0 }

func (r *Result) errf(format string, a ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, a...))
}

func (r *Result) warnf(format string, a ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, a...))
}

// linkRe matches inline markdown link targets.
var linkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

// sectionHeaderRe matches a level-2 heading.
var sectionHeaderRe = regexp.MustCompile(`(?m)^##\s+(.+?)\s*$`)

// CheckBidirectional walks modelDir and reports edges that lack a reverse.
//
//	requirements/.../*.md  ## Implemented By  -> targets
//	features|components|apis|services/.../*.md  ## Requirements  -> targets
//
// For each forward edge A -> B, the reverse expectation is asserted on B.
func CheckBidirectional(out io.Writer, modelDir string, r *Result) {
	fmt.Fprintln(out, "== requirement <-> implementer cross-links ==")
	if _, err := os.Stat(modelDir); err != nil {
		r.warnf("model dir %s missing; skipping link check", modelDir)
		return
	}

	type edge struct {
		from string // absolute path to source file
		to   string // absolute path to target file (resolved)
	}
	var fromReqs, fromImpls []edge

	walkAreas := func(area, section string, sink *[]edge) {
		root := filepath.Join(modelDir, area)
		if _, err := os.Stat(root); err != nil {
			return
		}
		_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".md") || filepath.Base(p) == "_index.md" {
				return nil
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			body := extractSection(string(data), section)
			if body == "" {
				return nil
			}
			for _, target := range linksIn(body) {
				abs := resolveTarget(p, target)
				if abs == "" {
					continue
				}
				*sink = append(*sink, edge{from: p, to: abs})
			}
			return nil
		})
	}

	walkAreas("requirements", "Implemented By", &fromReqs)
	for _, area := range []string{"features", "components", "apis", "services"} {
		walkAreas(area, "Requirements", &fromImpls)
	}

	// Build reverse-lookup sets for O(1) symmetry checks.
	reqHasImpl := make(map[string]map[string]bool) // req -> set of impls
	for _, e := range fromReqs {
		if reqHasImpl[e.from] == nil {
			reqHasImpl[e.from] = map[string]bool{}
		}
		reqHasImpl[e.from][e.to] = true
	}
	implHasReq := make(map[string]map[string]bool)
	for _, e := range fromImpls {
		if implHasReq[e.from] == nil {
			implHasReq[e.from] = map[string]bool{}
		}
		implHasReq[e.from][e.to] = true
	}

	count := 0
	for _, e := range fromReqs {
		// Only enforce reverse edges when target lives inside the model tree
		// and is a known impl-kind (features/components/apis/services).
		if !inImplArea(modelDir, e.to) {
			continue
		}
		if !implHasReq[e.to][e.from] {
			r.errf("missing reverse: %s\n  -> Implemented By links to %s\n  but %s has no `## Requirements` link back to it",
				rel(modelDir, e.from), rel(modelDir, e.to), rel(modelDir, e.to))
			count++
		}
	}
	for _, e := range fromImpls {
		if !inReqArea(modelDir, e.to) {
			continue
		}
		if !reqHasImpl[e.to][e.from] {
			r.errf("missing reverse: %s\n  -> Requirements links to %s\n  but %s has no `## Implemented By` link back to it",
				rel(modelDir, e.from), rel(modelDir, e.to), rel(modelDir, e.to))
			count++
		}
	}
	if count == 0 {
		fmt.Fprintln(out, "ok")
	}
}

// extractSection returns the body between a `## <name>` heading and the
// next `## ` heading (or EOF). Returns "" if not found.
func extractSection(body, name string) string {
	matches := sectionHeaderRe.FindAllStringSubmatchIndex(body, -1)
	for i, m := range matches {
		title := strings.TrimSpace(body[m[2]:m[3]])
		if !strings.EqualFold(title, name) {
			continue
		}
		end := len(body)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		return body[m[1]:end]
	}
	return ""
}

// linksIn returns inline-link targets in body, skipping URLs and anchors.
func linksIn(body string) []string {
	var out []string
	for _, m := range linkRe.FindAllStringSubmatch(body, -1) {
		t := strings.TrimSpace(m[1])
		if t == "" {
			continue
		}
		switch {
		case strings.HasPrefix(t, "http://"),
			strings.HasPrefix(t, "https://"),
			strings.HasPrefix(t, "mailto:"),
			strings.HasPrefix(t, "#"):
			continue
		}
		// strip fragment / query
		if i := strings.IndexAny(t, "#?"); i >= 0 {
			t = t[:i]
		}
		if t == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

// resolveTarget makes target absolute relative to fromFile and verifies it
// exists. Returns "" on failure or when the target isn't a regular .md file
// (directories and _index.md files don't carry per-spec sections to verify).
func resolveTarget(fromFile, target string) string {
	if filepath.IsAbs(target) {
		return ""
	}
	abs := filepath.Clean(filepath.Join(filepath.Dir(fromFile), target))
	st, err := os.Stat(abs)
	if err != nil {
		return ""
	}
	if st.IsDir() {
		return ""
	}
	if !strings.HasSuffix(abs, ".md") {
		return ""
	}
	if filepath.Base(abs) == "_index.md" {
		return ""
	}
	return abs
}

func inReqArea(modelDir, p string) bool {
	rel, err := filepath.Rel(modelDir, p)
	if err != nil {
		return false
	}
	return strings.HasPrefix(filepath.ToSlash(rel), "requirements/")
}

func inImplArea(modelDir, p string) bool {
	rel, err := filepath.Rel(modelDir, p)
	if err != nil {
		return false
	}
	r := filepath.ToSlash(rel)
	return strings.HasPrefix(r, "features/") ||
		strings.HasPrefix(r, "components/") ||
		strings.HasPrefix(r, "apis/") ||
		strings.HasPrefix(r, "services/")
}

func rel(base, p string) string {
	r, err := filepath.Rel(base, p)
	if err != nil {
		return p
	}
	return r
}
