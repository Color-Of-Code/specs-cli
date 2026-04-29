// Package tools manages the read-only managed-mode cache of framework
// content under the user's cache dir (os.UserCacheDir() + /specs-toolchain/tools/<ref>/).
//
// The cache is keyed by tools_ref so multiple refs can coexist and several
// host projects on the same machine share a single checkout per ref.
package tools

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CacheRoot returns the base directory for managed tool checkouts.
//
//	Linux:   ${XDG_CACHE_HOME:-~/.cache}/specs-toolchain/tools
//	macOS:   ~/Library/Caches/specs-toolchain/tools
//	Windows: %LocalAppData%\specs-toolchain\tools
func CacheRoot() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "specs-toolchain", "tools"), nil
}

// ManagedPath returns the cache path for a given ref. The ref is sanitised
// so refs with slashes (e.g. "release/v1") create nested dirs safely.
func ManagedPath(ref string) (string, error) {
	root, err := CacheRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, sanitize(ref)), nil
}

// Ensure clones url@ref into the managed cache if not already present.
// Returns the absolute path to the checkout. Idempotent: if the path exists
// it is returned as-is (no fetch). Use Refresh to force-update an existing
// cache entry.
func Ensure(url, ref string) (string, error) {
	if url == "" {
		return "", errors.New("tools_url is required for managed mode")
	}
	if ref == "" {
		ref = "main"
	}
	path, err := ManagedPath(ref)
	if err != nil {
		return "", err
	}
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		return path, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	tmp := path + ".tmp." + strconv.Itoa(os.Getpid())
	_ = os.RemoveAll(tmp)
	if err := clone(url, ref, tmp); err != nil {
		_ = os.RemoveAll(tmp)
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.RemoveAll(tmp)
		return "", fmt.Errorf("rename %s -> %s: %w", tmp, path, err)
	}
	return path, nil
}

// Refresh forces a fresh clone for ref by removing any existing entry and
// calling Ensure. Use sparingly; the cache is content-addressable by ref.
func Refresh(url, ref string) (string, error) {
	if ref == "" {
		ref = "main"
	}
	path, err := ManagedPath(ref)
	if err != nil {
		return "", err
	}
	_ = os.RemoveAll(path)
	return Ensure(url, ref)
}

// clone runs `git clone <url> dst` and, when ref is set, checks it out.
// Branches/tags use --branch --single-branch for speed; commit SHAs and
// other refs fall back to a full clone + checkout.
func clone(url, ref, dst string) error {
	if looksLikeRefName(ref) {
		args := []string{"clone", "--depth", "1", "--branch", ref, "--single-branch", url, dst}
		if err := runGit("", args...); err == nil {
			return nil
		}
		// fall through to full clone if --branch failed (e.g. ref is a SHA)
		_ = os.RemoveAll(dst)
	}
	if err := runGit("", "clone", url, dst); err != nil {
		return err
	}
	if ref != "" {
		if err := runGit(dst, "checkout", ref); err != nil {
			return err
		}
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// sanitize replaces path-unsafe characters in a ref so it can be used as a
// directory name. Slashes are preserved (Git allows hierarchical refs).
func sanitize(ref string) string {
	if ref == "" {
		return "default"
	}
	var b strings.Builder
	for _, r := range ref {
		switch {
		case r == '/' || r == '_' || r == '-' || r == '.':
			b.WriteRune(r)
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// looksLikeRefName reports whether ref is plausibly a branch or tag name
// (i.e. not a 40-char or 7-40 char hex commit SHA). Exact heuristic: all
// hex chars and length 7..40 -> treat as SHA.
func looksLikeRefName(ref string) bool {
	if ref == "" {
		return false
	}
	if l := len(ref); l >= 7 && l <= 40 {
		hex := true
		for _, r := range ref {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				hex = false
				break
			}
		}
		if hex {
			return false
		}
	}
	return true
}
