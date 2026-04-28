package linkcheck

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

func TestCheckBidirectional_Symmetric(t *testing.T) {
	model := t.TempDir()

	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), `# Foo

## Implemented By

- [feat](../../features/core/foo.md)
`)
	write(t, filepath.Join(model, "features", "core", "foo.md"), `# Foo Feature

## Requirements

- [REQ-001](../../requirements/core/001-foo.md)
`)

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, r)
	if r.Failed() {
		t.Fatalf("expected no errors, got %v\n%s", r.Errors, buf.String())
	}
}

func TestCheckBidirectional_MissingReverse(t *testing.T) {
	model := t.TempDir()

	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), `# Foo

## Implemented By

- [feat](../../features/core/foo.md)
`)
	// Feature file exists but has no `## Requirements` section.
	write(t, filepath.Join(model, "features", "core", "foo.md"), `# Foo Feature

## Notes

(nothing here)
`)

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, r)
	if !r.Failed() {
		t.Fatalf("expected error for missing reverse link\noutput: %s", buf.String())
	}
	joined := strings.Join(r.Errors, "\n")
	if !strings.Contains(joined, "foo.md") {
		t.Errorf("error message should mention the offending file: %s", joined)
	}
}

func TestCheckBidirectional_SkipsIndexAndDirs(t *testing.T) {
	model := t.TempDir()

	// Reference an _index.md (should be ignored as a target).
	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), `# Foo

## Implemented By

- [area](../../features/core/_index.md)
- [dir](../../features/core)
`)
	write(t, filepath.Join(model, "features", "core", "_index.md"), "# Index\n")

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, r)
	if r.Failed() {
		t.Fatalf("_index.md and directory targets should be skipped, got: %v", r.Errors)
	}
}
