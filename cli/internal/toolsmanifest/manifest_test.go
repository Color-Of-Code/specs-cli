package toolsmanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Missing(t *testing.T) {
	dir := t.TempDir()
	m, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != nil {
		t.Fatalf("expected nil manifest for missing file, got %+v", m)
	}
}

func TestLoad_EmptyDir(t *testing.T) {
	m, err := Load("")
	if err != nil || m != nil {
		t.Fatalf("expected (nil, nil) for empty dir, got %v %v", m, err)
	}
}

func TestLoad_Parse(t *testing.T) {
	dir := t.TempDir()
	contents := "templates_schema: 2\nversion: \"1.4.0\"\n"
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m == nil || m.TemplatesSchema != 2 || m.Version != "1.4.0" {
		t.Fatalf("unexpected manifest: %+v", m)
	}
}

func TestLoad_Malformed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte("templates_schema: [oops\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestCompatible(t *testing.T) {
	tests := []struct {
		name     string
		required int
		manifest *Manifest
		ok       bool
	}{
		{"host pins nothing", 0, nil, true},
		{"host pins, manifest absent", 1, nil, false},
		{"manifest silent", 1, &Manifest{}, false},
		{"match", 2, &Manifest{TemplatesSchema: 2}, true},
		{"mismatch", 2, &Manifest{TemplatesSchema: 3}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, msg := Compatible(tc.required, tc.manifest)
			if ok != tc.ok {
				t.Fatalf("ok=%v want %v (msg=%q)", ok, tc.ok, msg)
			}
			if !ok && msg == "" {
				t.Fatalf("expected non-empty message on incompatibility")
			}
		})
	}
}
