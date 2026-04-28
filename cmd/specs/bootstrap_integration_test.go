// Integration tests for `specs bootstrap` across the supported layouts.
//
// These build the binary into a tempdir and run it with --dry-run so no
// network or git submodule work happens; they verify that the command
// emits the expected actions and exit codes for each layout / tools-mode
// combination.
package main_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "specs")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "./")
	cmd.Dir = mustModuleDir(t)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build: %v\n%s", err, stderr.String())
	}
	return bin
}

func mustModuleDir(t *testing.T) string {
	t.Helper()
	// This file lives in cmd/specs; its module dir is the same.
	abs, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func runSpecs(t *testing.T, bin, cwd string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = cwd
	var so, se bytes.Buffer
	cmd.Stdout = &so
	cmd.Stderr = &se
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return so.String(), se.String(), ee.ExitCode()
		}
		t.Fatalf("run: %v", err)
	}
	return so.String(), se.String(), 0
}

func TestBootstrap_LayoutFolder_ToolsManaged(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	out, _, code := runSpecs(t, bin, host, "bootstrap", "--at", "specs", "--layout", "folder", "--tools-mode", "managed", "--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	for _, want := range []string{"would: mkdir -p", "fetch ", "managed cache", "specs init"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s", want, out)
		}
	}
}

func TestBootstrap_LayoutFolder_ToolsSubmodule(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	out, _, code := runSpecs(t, bin, host, "bootstrap", "--at", "specs", "--layout", "folder", "--tools-mode", "submodule", "--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "submodule add") {
		t.Errorf("expected submodule add action\n%s", out)
	}
}

func TestBootstrap_LayoutFolder_ToolsVendor(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	out, _, code := runSpecs(t, bin, host, "bootstrap", "--at", "specs", "--layout", "folder", "--tools-mode", "vendor", "--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "rm -rf") || !strings.Contains(out, "/.git") {
		t.Errorf("expected vendor mode to strip .git\n%s", out)
	}
}

func TestBootstrap_LayoutSubmodule_RequiresURL(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	_, se, code := runSpecs(t, bin, host, "bootstrap", "--at", "specs", "--layout", "submodule", "--dry-run")
	if code == 0 {
		t.Fatalf("expected non-zero exit for missing --specs-url")
	}
	if !strings.Contains(se, "--specs-url") {
		t.Errorf("error should mention --specs-url\n%s", se)
	}
}

func TestBootstrap_LayoutSubmodule_DryRun(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	// Initialise git so the host has a root and bootstrap doesn't try to init.
	if err := exec.Command("git", "-C", host, "init", "-q").Run(); err != nil {
		t.Skip("git not available")
	}
	out, _, code := runSpecs(t, bin, host, "bootstrap",
		"--at", "specs", "--layout", "submodule",
		"--specs-url", "https://example.com/specs.git",
		"--tools-mode", "managed",
		"--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "submodule add https://example.com/specs.git specs") {
		t.Errorf("expected git submodule add line\n%s", out)
	}
}

func TestBootstrap_UnknownLayout(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	_, se, code := runSpecs(t, bin, host, "bootstrap", "--layout", "weird", "--dry-run")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown layout")
	}
	if !strings.Contains(se, "unknown --layout") {
		t.Errorf("expected unknown-layout error: %s", se)
	}
}

func TestBootstrap_UnknownToolsMode(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	_, se, code := runSpecs(t, bin, host, "bootstrap", "--tools-mode", "weird", "--dry-run")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown tools-mode")
	}
	if !strings.Contains(se, "unknown --tools-mode") {
		t.Errorf("expected unknown-tools-mode error: %s", se)
	}
}
