package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinksInFile_FiltersURLsAndAnchors(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "doc.md")
	body := strings.Join([]string{
		"# Doc",
		"",
		"See [external](https://example.com) and [anchor](#section).",
		"Also [neighbour](./neighbour.md) and [parent](../up.md#frag).",
		"",
		"```markdown",
		"[ignored](should-not-match.md)",
		"```",
		"",
		"And [mailto](mailto:a@b.c).",
	}, "\n")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := linksInFile(p)
	want := []string{"./neighbour.md", "../up.md#frag"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("link[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestSplitTableRow_Indices(t *testing.T) {
	row := "| Component | repo-name | path/inside | abc123 | reads |"
	cols := splitTableRow(row)
	// 1=Component, 2=Repo, 3=Path, 4=Commit, 5=Reads.
	if strings.TrimSpace(cols[2]) != "repo-name" {
		t.Errorf("repo col=%q", cols[2])
	}
	if strings.TrimSpace(cols[3]) != "path/inside" {
		t.Errorf("path col=%q", cols[3])
	}
	if strings.TrimSpace(cols[4]) != "abc123" {
		t.Errorf("commit col=%q", cols[4])
	}
}

func TestStripFragment(t *testing.T) {
	cases := map[string]string{
		"a/b.md":         "a/b.md",
		"a/b.md#section": "a/b.md",
		"a/b.md?x=1":     "a/b.md",
		"a/b.md#s?x":     "a/b.md",
	}
	for in, want := range cases {
		if got := stripFragment(in); got != want {
			t.Errorf("stripFragment(%q)=%q want %q", in, got, want)
		}
	}
}
