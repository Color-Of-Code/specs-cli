// Integration tests for the core specs commands (version, init, doctor,
// format, lint, framework). These build the binary into a tempdir and
// invoke it against tempdir hosts with isolated HOME / XDG_CONFIG_HOME
// so the tests never touch the user's real registry or git config.
package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runSpecsEnv is like runSpecs but allows the caller to override env vars
// (HOME, XDG_CONFIG_HOME) to keep tests hermetic.
func runSpecsEnv(t *testing.T, bin, cwd string, env []string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), env...)
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

// isolatedEnv returns env vars that point HOME and XDG_CONFIG_HOME at a
// fresh tempdir so commands that read user config don't leak across runs
// or pick up the developer's real setup.
func isolatedEnv(t *testing.T) []string {
	t.Helper()
	home := t.TempDir()
	return []string{
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
		"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
	}
}

func TestVersion_Smoke(t *testing.T) {
	bin := buildBinary(t)
	out, _, code := runSpecs(t, bin, t.TempDir(), "version")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Errorf("expected non-empty version output")
	}
}

func TestInit_WritesConfig_FrameworkDir(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	out, se, code := runSpecsEnv(t, bin, host, env, "init", "--framework-dir", "../specs-framework")
	if code != 0 {
		t.Fatalf("exit %d\nstdout:%s\nstderr:%s", code, out, se)
	}
	cfgPath := filepath.Join(host, ".specs.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read .specs.yaml: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "framework_dir:") || !strings.Contains(body, "../specs-framework") {
		t.Errorf("expected framework_dir line in .specs.yaml:\n%s", body)
	}
}

func TestInit_FailsWithoutForce(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	if _, _, code := runSpecsEnv(t, bin, host, env, "init", "--framework-dir", "x"); code != 0 {
		t.Fatalf("first init unexpectedly failed (exit %d)", code)
	}
	_, se, code := runSpecsEnv(t, bin, host, env, "init", "--framework-dir", "x")
	if code == 0 {
		t.Fatalf("expected non-zero exit when .specs.yaml exists without --force")
	}
	if !strings.Contains(se, "already exists") {
		t.Errorf("expected 'already exists' error, got: %s", se)
	}
}

func TestInit_RejectsConflictingFlags(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	_, se, code := runSpecsEnv(t, bin, host, env,
		"init", "--framework-url", "https://example.com/x.git", "--framework-dir", "y")
	if code == 0 {
		t.Fatalf("expected non-zero exit for conflicting flags")
	}
	if !strings.Contains(se, "mutually exclusive") {
		t.Errorf("expected mutual-exclusion error, got: %s", se)
	}
}

func TestDoctor_JSON_AfterInit(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	if _, se, code := runSpecsEnv(t, bin, host, env, "init", "--framework-dir", "../specs-framework"); code != 0 {
		t.Fatalf("init failed: exit %d\n%s", code, se)
	}
	out, se, code := runSpecsEnv(t, bin, host, env, "doctor", "--json")
	if code != 0 {
		t.Fatalf("doctor exit %d\nstdout:%s\nstderr:%s", code, out, se)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("doctor --json output not valid JSON: %v\n%s", err, out)
	}
	for _, key := range []string{"version", "specs_root", "framework_dir", "framework_mode"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("doctor JSON missing key %q\nfull output:\n%s", key, out)
		}
	}
}

func TestFormat_Check_CleanFile(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	// A trivially-clean markdown file: lone heading + one paragraph.
	if err := os.WriteFile(filepath.Join(host, "doc.md"),
		[]byte("# Title\n\nBody paragraph.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, se, code := runSpecsEnv(t, bin, host, env, "format", "--check", "doc.md")
	if code != 0 {
		t.Fatalf("expected exit 0 for clean file; got %d\nstderr: %s", code, se)
	}
}

func TestFormat_Check_DirtyFileFails(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	// Trailing whitespace on the heading should require formatting.
	if err := os.WriteFile(filepath.Join(host, "dirty.md"),
		[]byte("# Title   \n\nBody.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, code := runSpecsEnv(t, bin, host, env, "format", "--check", "dirty.md")
	if code == 0 {
		t.Fatalf("expected non-zero exit for dirty file")
	}
}

func TestFrameworkList_EmptyRegistry(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	out, se, code := runSpecsEnv(t, bin, host, env, "framework", "list")
	if code != 0 {
		t.Fatalf("exit %d\nstderr: %s", code, se)
	}
	if !strings.Contains(out, "No frameworks registered") {
		t.Errorf("expected 'No frameworks registered' message, got:\n%s", out)
	}
}

func TestFrameworkAddListRemove_Roundtrip(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	if _, se, code := runSpecsEnv(t, bin, host, env,
		"framework", "add", "demo",
		"--url", "https://example.com/specs-framework.git",
		"--ref", "v1.2.3"); code != 0 {
		t.Fatalf("add: exit %d\n%s", code, se)
	}

	out, _, code := runSpecsEnv(t, bin, host, env, "framework", "list")
	if code != 0 {
		t.Fatalf("list exit %d", code)
	}
	if !strings.Contains(out, "demo") || !strings.Contains(out, "v1.2.3") {
		t.Errorf("expected demo/v1.2.3 in list output:\n%s", out)
	}

	if _, se, code := runSpecsEnv(t, bin, host, env, "framework", "remove", "demo"); code != 0 {
		t.Fatalf("remove: exit %d\n%s", code, se)
	}

	out, _, code = runSpecsEnv(t, bin, host, env, "framework", "list")
	if code != 0 {
		t.Fatalf("list-after-remove exit %d", code)
	}
	if strings.Contains(out, "demo") {
		t.Errorf("removed entry still present:\n%s", out)
	}
}

func TestFrameworkUpdate_RequiresConfig(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)

	// No .specs.yaml in the host; update should fail with a clear error.
	_, se, code := runSpecsEnv(t, bin, host, env, "framework", "update")
	if code == 0 {
		t.Fatalf("expected non-zero exit when no .specs.yaml is present")
	}
	if se == "" {
		t.Errorf("expected an error message on stderr")
	}
}
