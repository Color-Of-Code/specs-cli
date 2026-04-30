package framework

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/registry"
)

// withRegistry redirects user config lookup to a tempdir and writes the
// supplied registry. Returns the registry's on-disk path.
func withRegistry(t *testing.T, r *registry.Registry) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	// macOS path that os.UserConfigDir reads; setting it harmless on linux.
	t.Setenv("LocalAppData", filepath.Join(home, "AppData", "Local"))
	path, err := registry.DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		if err := r.Save(path); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestResolveSource_RegisteredURLEntry(t *testing.T) {
	withRegistry(t, &registry.Registry{Frameworks: map[string]registry.Entry{
		"acme": {URL: "https://example.com/fw.git", Ref: "v1"},
	}})
	s, err := ResolveSource("acme")
	if err != nil {
		t.Fatalf("ResolveSource: %v", err)
	}
	if s.URL != "https://example.com/fw.git" || s.Ref != "v1" || s.Path != "" {
		t.Errorf("got %+v; want URL+Ref entry", s)
	}
}

func TestResolveSource_RegisteredPathEntry(t *testing.T) {
	withRegistry(t, &registry.Registry{Frameworks: map[string]registry.Entry{
		"local-dev": {Path: "/tmp/fw"},
	}})
	s, err := ResolveSource("local-dev")
	if err != nil {
		t.Fatalf("ResolveSource: %v", err)
	}
	if s.Path != "/tmp/fw" || s.URL != "" || s.Ref != "" {
		t.Errorf("got %+v; want Path entry", s)
	}
}

func TestResolveSource_RefOverride(t *testing.T) {
	withRegistry(t, &registry.Registry{Frameworks: map[string]registry.Entry{
		"acme": {URL: "https://example.com/fw.git", Ref: "v1"},
	}})
	s, err := ResolveSource("acme@v2")
	if err != nil {
		t.Fatalf("ResolveSource: %v", err)
	}
	if s.Ref != "v2" {
		t.Errorf("ref override not applied: %+v", s)
	}
}

func TestResolveSource_RefOverrideOnPathEntryFails(t *testing.T) {
	withRegistry(t, &registry.Registry{Frameworks: map[string]registry.Entry{
		"local-dev": {Path: "/tmp/fw"},
	}})
	if _, err := ResolveSource("local-dev@main"); err == nil {
		t.Error("expected error for @ref on path-based entry")
	}
}

func TestResolveSource_EmptySpecResolvesDefault(t *testing.T) {
	withRegistry(t, &registry.Registry{Frameworks: map[string]registry.Entry{
		"default": {URL: "https://example.com/default.git"},
	}})
	s, err := ResolveSource("")
	if err != nil {
		t.Fatalf("ResolveSource: %v", err)
	}
	if s.URL != "https://example.com/default.git" {
		t.Errorf("default lookup failed: %+v", s)
	}
}

func TestResolveSource_UnknownNameFails(t *testing.T) {
	withRegistry(t, &registry.Registry{})
	_, err := ResolveSource("nope")
	if err == nil {
		t.Fatal("expected error for unknown framework")
	}
	if !strings.Contains(err.Error(), "specs framework add") {
		t.Errorf("expected hint about `specs framework add` in error: %v", err)
	}
}

func TestResolveSource_EmptySpecWithEmptyRegistryFails(t *testing.T) {
	withRegistry(t, &registry.Registry{})
	_, err := ResolveSource("")
	if err == nil {
		t.Fatal("expected error when no default is registered")
	}
	if !strings.Contains(err.Error(), `"default"`) {
		t.Errorf("expected error to mention 'default': %v", err)
	}
}

func TestSource_Validate(t *testing.T) {
	if err := (Source{}).Validate(); err == nil {
		t.Error("empty source should not validate")
	}
	if err := (Source{URL: "u", Path: "p"}).Validate(); err == nil {
		t.Error("url+path should not validate")
	}
	if err := (Source{Path: "p", Ref: "r"}).Validate(); err == nil {
		t.Error("path+ref should not validate")
	}
	if err := (Source{URL: "u", Ref: "r"}).Validate(); err != nil {
		t.Errorf("url+ref should validate: %v", err)
	}
}
