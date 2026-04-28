package visualize

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sampleModel(t *testing.T) string {
	model := t.TempDir()
	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), `# Foo Requirement

## Implemented By

- [feat](../../features/core/foo.md)
- [comp](../../components/core/comp.md)
`)
	write(t, filepath.Join(model, "features", "core", "foo.md"), `# Foo Feature

## Requirements

- [REQ-001](../../requirements/core/001-foo.md)
`)
	write(t, filepath.Join(model, "components", "core", "comp.md"), `# Comp

## Requirements

- [REQ-001](../../requirements/core/001-foo.md)
`)
	return model
}

func TestBuild(t *testing.T) {
	g, err := Build(sampleModel(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) != 3 {
		t.Errorf("want 3 nodes, got %d", len(g.Nodes))
	}
	// Edges include forward and reverse, deduped/sorted; expect at least 4.
	if len(g.Edges) < 4 {
		t.Errorf("want >=4 edges, got %d: %+v", len(g.Edges), g.Edges)
	}

	kinds := map[string]int{}
	for _, n := range g.Nodes {
		kinds[n.Kind]++
	}
	for _, want := range []string{"requirement", "feature", "component"} {
		if kinds[want] != 1 {
			t.Errorf("expected one %s node, got %d", want, kinds[want])
		}
	}

	// Labels use the H1 text.
	for _, n := range g.Nodes {
		if n.Label == "" || n.Label == filepath.Base(n.Path) {
			t.Errorf("node %s has empty/basename label %q", n.Path, n.Label)
		}
	}
}

func TestBuild_MissingDir(t *testing.T) {
	if _, err := Build(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing model dir")
	}
}

func TestWriteDOT(t *testing.T) {
	g, err := Build(sampleModel(t))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteDOT(&buf, g); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"digraph", "rankdir=LR", "subgraph cluster_requirement", "subgraph cluster_feature", "->"} {
		if !strings.Contains(out, want) {
			t.Errorf("DOT output missing %q\n---\n%s", want, out)
		}
	}
}

func TestWriteMermaid(t *testing.T) {
	g, err := Build(sampleModel(t))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteMermaid(&buf, g); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"flowchart", "-->"} {
		if !strings.Contains(out, want) {
			t.Errorf("Mermaid output missing %q\n---\n%s", want, out)
		}
	}
	// Requirement uses [[..]] subroutine shape; component uses [(..)] cylinder.
	if !strings.Contains(out, "[[") || !strings.Contains(out, "[(") {
		t.Errorf("Mermaid output missing per-kind shapes\n---\n%s", out)
	}
}

func TestKindFor(t *testing.T) {
	cases := map[string]string{
		"requirements/core/001-foo.md": "requirement",
		"features/core/foo.md":         "feature",
		"components/core/foo.md":       "component",
		"apis/core/foo.md":             "api",
		"services/foo.md":              "service",
		"baselines/repo-baseline.md":   "",
		"glossary.md":                  "",
	}
	for in, want := range cases {
		if got := kindFor(in); got != want {
			t.Errorf("kindFor(%q)=%q want %q", in, got, want)
		}
	}
}
