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

// registerURL adds a URL-based framework entry to the registry that
// `env` resolves so subsequent `specs init` calls can refer to it by name.
func registerURL(t *testing.T, bin, host string, env []string, name, url, ref string) {
	t.Helper()
	args := []string{"framework", "add", name, "--url", url}
	if ref != "" {
		args = append(args, "--ref", ref)
	}
	if _, se, code := runSpecsEnv(t, bin, host, env, args...); code != 0 {
		t.Fatalf("framework add %s: exit %d\n%s", name, code, se)
	}
}

// registerPath adds a path-based framework entry.
func registerPath(t *testing.T, bin, host string, env []string, name, path string) {
	t.Helper()
	if _, se, code := runSpecsEnv(t, bin, host, env, "framework", "add", name, "--path", path); code != 0 {
		t.Fatalf("framework add %s: exit %d\n%s", name, code, se)
	}
}

func TestInit_Managed_DryRun(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)
	registerURL(t, bin, host, env, "demo", "https://example.com/fw.git", "main")

	out, se, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "demo",
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

func TestInit_RefOverride_DryRun(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)
	registerURL(t, bin, host, env, "demo", "https://example.com/fw.git", "main")

	out, _, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "demo@v9.9.9",
		"--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "v9.9.9") {
		t.Errorf("expected ref override v9.9.9 in output:\n%s", out)
	}
}

func TestInit_PathEntry_NoMaterialisation(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)
	registerPath(t, bin, host, env, "local-dev", "/tmp/specs-framework")

	out, _, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "local-dev",
		"--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if strings.Contains(out, "managed cache") || strings.Contains(out, "git clone") {
		t.Errorf("path entry should not materialise framework:\n%s", out)
	}
	if !strings.Contains(out, "would: write") {
		t.Errorf("expected config write line:\n%s", out)
	}
}

func TestInit_UnknownName_Fails(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	_, se, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "nope",
		"--dry-run")
	if code == 0 {
		t.Fatal("expected non-zero exit when framework is not registered")
	}
	if !strings.Contains(se, "specs framework add") {
		t.Errorf("expected hint about `specs framework add`: %s", se)
	}
}

func TestInit_NoFrameworkFlag_NoDefault_Fails(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	_, se, code := runSpecsEnv(t, bin, host, env, "init", "--dry-run")
	if code == 0 {
		t.Fatal("expected non-zero exit when no framework registered")
	}
	if !strings.Contains(se, `"default"`) {
		t.Errorf("expected error to mention the missing default entry: %s", se)
	}
}

func TestInit_PositionalPath(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)
	registerURL(t, bin, host, env, "demo", "https://example.com/fw.git", "main")

	out, _, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "demo",
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
