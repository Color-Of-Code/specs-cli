// Integration tests for `specs init` covering its various framework
// resolution paths. Tests build the binary into a tempdir and run it
// with --dry-run so no network work happens; they verify the command
// emits the expected actions and exit codes.
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

func TestInit_Managed_DryRun(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)
	out, se, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "https://example.com/fw.git@main",
		"--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\nstdout:%s\nstderr:%s", code, out, se)
	}
	for _, want := range []string{"would: fetch", "managed cache", "would: write"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s", want, out)
		}
	}
}

func TestInit_PathSource_NoMaterialisation(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)
	// Local path source: no clone/fetch should occur.
	out, _, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "../specs-framework",
		"--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if strings.Contains(out, "managed cache") || strings.Contains(out, "submodule add") || strings.Contains(out, "git clone") {
		t.Errorf("path source should not materialise framework:\n%s", out)
	}
	if !strings.Contains(out, "would: write") {
		t.Errorf("expected config write line:\n%s", out)
	}
}

func TestInit_PositionalPath(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)
	out, _, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "https://example.com/fw.git@main",
		"--dry-run",
		"specs")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	want := filepath.Join(host, "specs")
	if !strings.Contains(out, want) {
		t.Errorf("expected target path %q in output:\n%s", want, out)
	}
}
