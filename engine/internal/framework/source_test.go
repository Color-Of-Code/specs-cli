package framework

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveSource_RemoteURLs(t *testing.T) {
	cases := []struct {
		spec    string
		wantURL string
		wantRef string
	}{
		{"https://example.com/foo.git", "https://example.com/foo.git", ""},
		{"https://example.com/foo.git@v1.0", "https://example.com/foo.git", "v1.0"},
		{"git@github.com:owner/repo.git", "git@github.com:owner/repo.git", ""},
		{"git@github.com:owner/repo.git@main", "git@github.com:owner/repo.git", "main"},
		{"ssh://git@example.com/x.git", "ssh://git@example.com/x.git", ""},
	}
	for _, tc := range cases {
		t.Run(tc.spec, func(t *testing.T) {
			s, err := ResolveSource(tc.spec)
			if err != nil {
				t.Fatalf("ResolveSource(%q): %v", tc.spec, err)
			}
			if s.URL != tc.wantURL || s.Ref != tc.wantRef || s.Path != "" {
				t.Errorf("got %+v; want URL=%q Ref=%q", s, tc.wantURL, tc.wantRef)
			}
		})
	}
}

func TestResolveSource_LocalPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("path normalisation differs on windows")
	}
	cases := []string{"./foo", "../bar", "/tmp"}
	for _, spec := range cases {
		t.Run(spec, func(t *testing.T) {
			s, err := ResolveSource(spec)
			if err != nil {
				t.Fatalf("ResolveSource(%q): %v", spec, err)
			}
			if s.URL != "" || s.Ref != "" {
				t.Errorf("path spec resolved as URL: %+v", s)
			}
			if !filepath.IsAbs(s.Path) {
				t.Errorf("path not absolute: %q", s.Path)
			}
		})
	}
}

func TestResolveSource_PathRejectsRef(t *testing.T) {
	// The "@ref" suffix is never split off a path-style spec, so when a
	// caller passes "/tmp@v1" they get a literal path.
	s, err := ResolveSource("/tmp@v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Ref != "" {
		t.Errorf("ref unexpectedly extracted from path: %+v", s)
	}
	if !strings.HasSuffix(s.Path, "@v1") {
		t.Errorf("path lost @v1 suffix: %q", s.Path)
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
