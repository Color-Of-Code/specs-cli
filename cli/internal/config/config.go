// Package config loads and resolves the host-side .specs.yaml configuration
// and provides layout auto-detection for the specs/ root and tools_dir.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Color-Of-Code/specs-toolchain/cli/internal/tools"
	"gopkg.in/yaml.v3"
)

// FileName is the canonical name of the host-side config file.
const FileName = ".specs.yaml"

const (
	defaultToolsDirName = ".specs-framework"
	legacyToolsDirName  = ".specs-tools"
)

// SpecsMode describes how the specs root is materialised in the host.
type SpecsMode string

const (
	SpecsModeRepoRoot   SpecsMode = "repo-root"  // specs root == git repo root
	SpecsModeFolder     SpecsMode = "folder"     // plain subdirectory of host repo
	SpecsModeSubmodule  SpecsMode = "submodule"  // git submodule of host repo
	SpecsModeStandalone SpecsMode = "standalone" // not in a git repo
)

// ToolsMode describes how the framework content is materialised.
type ToolsMode string

const (
	ToolsModeManaged   ToolsMode = "managed" // CLI-managed read-only checkout in the user cache dir
	ToolsModeSubmodule ToolsMode = "submodule"
	ToolsModeFolder    ToolsMode = "folder" // plain folder; may or may not be a git working tree
	ToolsModeVendor    ToolsMode = "vendor" // vendored snapshot (no .git)
	ToolsModeMissing   ToolsMode = "missing"
)

// File is the on-disk schema for .specs.yaml. Unknown fields are tolerated
// so newer hosts can opt-in to features without breaking older binaries.
type File struct {
	SpecsRoot          string            `yaml:"specs_root,omitempty"`
	ToolsDir           string            `yaml:"tools_dir,omitempty"`
	ToolsURL           string            `yaml:"tools_url,omitempty"`
	ToolsRef           string            `yaml:"tools_ref,omitempty"`
	ChangeRequestsDir  string            `yaml:"change_requests_dir,omitempty"`
	ModelDir           string            `yaml:"model_dir,omitempty"`
	BaselinesFile      string            `yaml:"baselines_file,omitempty"`
	MarkdownlintConfig string            `yaml:"markdownlint_config,omitempty"`
	MinSpecsVersion    string            `yaml:"min_specs_version,omitempty"`
	TemplatesSchema    int               `yaml:"templates_schema,omitempty"`
	Repos              map[string]string `yaml:"repos,omitempty"`
}

// Resolved is a fully-resolved configuration with absolute paths and
// detected layout modes.
type Resolved struct {
	ConfigPath         string // empty when no .specs.yaml was found
	SpecsRoot          string // absolute path; always set
	HostRoot           string // git repo root; equal to SpecsRoot if mode is repo-root or standalone
	SpecsMode          SpecsMode
	ToolsDir           string // absolute path; may be empty if missing
	ToolsMode          ToolsMode
	ToolsURL           string // managed mode: upstream git URL
	ToolsRef           string // managed mode: pinned tag/branch/commit
	ChangeRequestsDir  string // absolute path
	ModelDir           string // absolute path
	BaselinesFile      string // absolute path; may not exist
	MarkdownlintConfig string // absolute path; may not exist
	MinSpecsVersion    string
	TemplatesSchema    int
	Repos              map[string]string // logical name -> path relative to HostRoot's parent (workspace)
	Source             *File             // raw file (nil if no .specs.yaml)
}

// Load discovers and parses .specs.yaml starting from start (or CWD if empty).
// Returns a Resolved configuration with absolute paths and detected layout.
// If no .specs.yaml is found, fallback heuristics select a SpecsRoot from
// the current directory or its closest ancestor containing model/ or
// change-requests/. Layout detection is run regardless.
func Load(start string) (*Resolved, error) {
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getwd: %w", err)
		}
		start = cwd
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}

	cfgPath := findUp(abs, FileName)
	r := &Resolved{}

	var f File
	if cfgPath != "" {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", cfgPath, err)
		}
		if err := yaml.Unmarshal(data, &f); err != nil {
			return nil, fmt.Errorf("parse %s: %w", cfgPath, err)
		}
		r.ConfigPath = cfgPath
		r.Source = &f
		r.SpecsRoot = filepath.Dir(cfgPath)
		if f.SpecsRoot != "" {
			r.SpecsRoot = absRelTo(filepath.Dir(cfgPath), f.SpecsRoot)
		}
	} else {
		// Heuristic fallback: look upward for a dir with model/ or change-requests/
		root := findUpDirContaining(abs, "model", "change-requests")
		if root == "" {
			root = abs
		}
		r.SpecsRoot = root
	}

	// Detect host root via git.
	r.HostRoot = gitRepoRoot(r.SpecsRoot)
	if r.HostRoot == "" {
		r.HostRoot = r.SpecsRoot
	}
	r.SpecsMode = detectSpecsMode(r.HostRoot, r.SpecsRoot)

	// Resolve tools location. Managed mode wins when tools_url is set: the
	// content lives in the user cache dir, and tools_dir (if any) is ignored.
	r.ToolsURL = f.ToolsURL
	r.ToolsRef = f.ToolsRef
	if f.ToolsURL != "" {
		cachePath, err := tools.ManagedPath(f.ToolsRef)
		if err != nil {
			return nil, fmt.Errorf("resolve managed cache path: %w", err)
		}
		r.ToolsDir = cachePath
		r.ToolsMode = ToolsModeManaged
	} else {
		tools := f.ToolsDir
		if tools == "" {
			tools = "auto"
		}
		resolvedTools, mode := resolveToolsDir(tools, r.SpecsRoot, r.HostRoot)
		r.ToolsDir = resolvedTools
		r.ToolsMode = mode
	}

	// Resolve other dirs/files (relative paths anchored to SpecsRoot).
	r.ChangeRequestsDir = absOr(r.SpecsRoot, f.ChangeRequestsDir, "change-requests")
	r.ModelDir = absOr(r.SpecsRoot, f.ModelDir, "model")
	r.BaselinesFile = absOr(r.SpecsRoot, f.BaselinesFile, filepath.Join("model", "baselines", "repo-baseline.md"))

	if f.MarkdownlintConfig != "" {
		r.MarkdownlintConfig = absRelTo(r.SpecsRoot, f.MarkdownlintConfig)
	} else if r.ToolsDir != "" {
		r.MarkdownlintConfig = filepath.Join(r.ToolsDir, "lint", ".markdownlint-cli2.jsonc")
	}

	r.MinSpecsVersion = f.MinSpecsVersion
	r.TemplatesSchema = f.TemplatesSchema
	r.Repos = f.Repos
	return r, nil
}

// findUp walks parents of dir looking for a file with the given name.
// Returns the absolute path to the file or empty string if not found.
func findUp(dir, name string) string {
	for {
		p := filepath.Join(dir, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// findUpDirContaining walks parents of dir looking for a directory containing
// any of the given child names. Returns absolute path or empty string.
func findUpDirContaining(dir string, children ...string) string {
	for {
		for _, c := range children {
			if st, err := os.Stat(filepath.Join(dir, c)); err == nil && st.IsDir() {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// gitRepoRoot returns the absolute path to the .git working tree root for
// dir, or empty string when dir is not inside a git repo. It honours the
// gitfile pointer used by submodules and worktrees by walking parents and
// returning the first ancestor that contains a .git entry of any kind.
func gitRepoRoot(dir string) string {
	d := dir
	for {
		if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		d = parent
	}
}

// detectSpecsMode determines how specsRoot relates to hostRoot.
func detectSpecsMode(hostRoot, specsRoot string) SpecsMode {
	if hostRoot == "" {
		return SpecsModeStandalone
	}
	if hostRoot == specsRoot {
		return SpecsModeRepoRoot
	}
	// If specsRoot itself contains a .git entry distinct from hostRoot,
	// it's a submodule (or worktree).
	if st, err := os.Stat(filepath.Join(specsRoot, ".git")); err == nil {
		_ = st
		// Check the host's .gitmodules for an entry whose path matches.
		if isSubmodule(hostRoot, specsRoot) {
			return SpecsModeSubmodule
		}
		// .git inside but not a declared submodule: treat as nested repo
		// (still effectively a submodule from CLI perspective).
		return SpecsModeSubmodule
	}
	return SpecsModeFolder
}

// isSubmodule returns true when child is registered as a submodule in the
// host's .gitmodules. Pure text scan; tolerant of malformed files.
func isSubmodule(hostRoot, child string) bool {
	data, err := os.ReadFile(filepath.Join(hostRoot, ".gitmodules"))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(hostRoot, child)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	for _, line := range splitLines(string(data)) {
		t := trimSpace(line)
		if hasPrefix(t, "path") {
			// path = <value>
			eq := indexByte(t, '=')
			if eq < 0 {
				continue
			}
			val := trimSpace(t[eq+1:])
			if val == rel {
				return true
			}
		}
	}
	return false
}

// resolveToolsDir resolves the tools_dir setting to an absolute path and
// detects the content mode. Recognised values for raw:
//   - "auto": try <specsRoot>/.specs-framework, then <specsRoot>/.specs-tools,
//     then the same two names under <hostRoot>.
//   - absolute or relative path: anchored to specsRoot.
func resolveToolsDir(raw, specsRoot, hostRoot string) (string, ToolsMode) {
	candidates := []string{}
	if raw == "" || raw == "auto" {
		candidates = append(candidates,
			filepath.Join(specsRoot, defaultToolsDirName),
			filepath.Join(specsRoot, legacyToolsDirName),
		)
		if hostRoot != "" && hostRoot != specsRoot {
			candidates = append(candidates,
				filepath.Join(hostRoot, defaultToolsDirName),
				filepath.Join(hostRoot, legacyToolsDirName),
			)
		}
	} else {
		candidates = append(candidates, absRelTo(specsRoot, raw))
	}
	for _, p := range candidates {
		st, err := os.Stat(p)
		if err != nil || !st.IsDir() {
			continue
		}
		return p, detectToolsMode(p, hostRoot)
	}
	return "", ToolsModeMissing
}

func detectToolsMode(toolsDir, hostRoot string) ToolsMode {
	// .git file or dir present?
	gitPath := filepath.Join(toolsDir, ".git")
	if st, err := os.Stat(gitPath); err == nil {
		_ = st
		if hostRoot != "" && isSubmodule(hostRoot, toolsDir) {
			return ToolsModeSubmodule
		}
		return ToolsModeFolder
	}
	return ToolsModeVendor
}

// absRelTo returns p resolved to absolute, anchored on base if relative.
func absRelTo(base, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(base, p))
}

// absOr returns absRelTo(base, p) when p is non-empty, otherwise
// absRelTo(base, fallback).
func absOr(base, p, fallback string) string {
	if p == "" {
		p = fallback
	}
	return absRelTo(base, p)
}

// Save writes f to path in canonical YAML form.
func Save(path string, f *File) error {
	if f == nil {
		return errors.New("nil config")
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// --- tiny string helpers (avoid pulling in strings to keep this file
// importable from very small contexts; revisit if usage grows). ---

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start <= len(s) {
		out = append(out, s[start:])
	}
	return out
}

func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
