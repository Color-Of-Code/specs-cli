package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad_PlainFolderLayout creates a fake host repo where specs/ is a
// plain folder (layout B1) with .specs-framework as a plain subfolder, and
// verifies that Load resolves correctly without a .specs.yaml.
func TestLoad_PlainFolderLayout(t *testing.T) {
	dir := t.TempDir()
	host := filepath.Join(dir, "host")
	specs := filepath.Join(host, "specs")
	framework := filepath.Join(specs, ".specs-framework")
	for _, p := range []string{
		filepath.Join(specs, "model"),
		filepath.Join(specs, "change-requests"),
		filepath.Join(framework, "templates"),
	} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Mark host as a git repo.
	if err := os.Mkdir(filepath.Join(host, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(specs)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SpecsRoot != specs {
		t.Errorf("SpecsRoot=%q want %q", cfg.SpecsRoot, specs)
	}
	if cfg.HostRoot != host {
		t.Errorf("HostRoot=%q want %q", cfg.HostRoot, host)
	}
	if cfg.SpecsMode != SpecsModeFolder {
		t.Errorf("SpecsMode=%q want %q", cfg.SpecsMode, SpecsModeFolder)
	}
	if cfg.FrameworkDir != framework {
		t.Errorf("FrameworkDir=%q want %q", cfg.FrameworkDir, framework)
	}
	if cfg.FrameworkMode != FrameworkModeVendor {
		t.Errorf("FrameworkMode=%q want %q (no .git inside framework dir)", cfg.FrameworkMode, FrameworkModeVendor)
	}
}

// TestLoad_RepoRoot covers the case where specs root is the git repo itself
// (layout when --at . was used).
func TestLoad_RepoRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "model"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SpecsMode != SpecsModeRepoRoot {
		t.Errorf("SpecsMode=%q want %q", cfg.SpecsMode, SpecsModeRepoRoot)
	}
}

// TestLoad_WithSpecsYAML round-trips a Save/Load cycle and checks overrides.
func TestLoad_WithSpecsYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	specs := filepath.Join(dir, "specs")
	cfgPath := filepath.Join(specs, FileName)
	in := &File{
		FrameworkDir:    "auto",
		MinSpecsVersion: "1.2.3",
		Repos: map[string]string{
			"redmine": "container/redmine/redmine",
		},
	}
	if err := Save(cfgPath, in); err != nil {
		t.Fatal(err)
	}
	got, err := Load(specs)
	if err != nil {
		t.Fatal(err)
	}
	if got.MinSpecsVersion != "1.2.3" {
		t.Errorf("MinSpecsVersion=%q", got.MinSpecsVersion)
	}
	if got.Repos["redmine"] != "container/redmine/redmine" {
		t.Errorf("Repos[redmine]=%q", got.Repos["redmine"])
	}
	if got.ConfigPath != cfgPath {
		t.Errorf("ConfigPath=%q want %q", got.ConfigPath, cfgPath)
	}
}
