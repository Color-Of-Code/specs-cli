package framework

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/registry"
)

// Source is a fully-resolved framework source: either a remote git URL
// (with optional Ref) or a local Path. Exactly one of URL/Path is set.
type Source struct {
	URL  string
	Ref  string
	Path string
}

// Validate returns an error if the source is malformed.
func (s Source) Validate() error {
	if s.URL == "" && s.Path == "" {
		return errors.New("source must set url or path")
	}
	if s.URL != "" && s.Path != "" {
		return errors.New("source must not set both url and path")
	}
	if s.Path != "" && s.Ref != "" {
		return errors.New("ref is only valid for url sources")
	}
	return nil
}

// ResolveSource turns a single user-provided string into a Source.
//
// Dispatch rules:
//   - empty spec      -> registry "default" entry (or hard-coded fallback)
//   - "scheme://..."  -> remote git URL (http(s)/ssh/git/file)
//   - "user@host:..." -> remote git URL (scp-style)
//   - starts with "/", "./", "../", "~" or names an existing directory
//     on disk                                  -> local path
//   - anything else                            -> registry name lookup
//
// A trailing "@ref" is parsed off URL specs and registry names; it overrides
// the registry entry's Ref. It is not allowed on path specs.
func ResolveSource(spec string) (Source, error) {
	if strings.TrimSpace(spec) == "" {
		return resolveRegistryName("default", "", true)
	}

	// Path-style specs are recognised first so a user-provided path that
	// happens to contain '@' (rare on disk but possible) is not split.
	if isPathSpec(spec) {
		path, err := expandHome(spec)
		if err != nil {
			return Source{}, err
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return Source{}, err
		}
		return Source{Path: abs}, nil
	}

	body, ref := splitRef(spec)

	if isRemoteURL(body) {
		return Source{URL: body, Ref: ref}, nil
	}

	// Treat as registry name. Allow @ref to override the registry's ref.
	return resolveRegistryName(body, ref, false)
}

func resolveRegistryName(name, refOverride string, allowFallback bool) (Source, error) {
	reg, err := registry.Load("")
	if err != nil {
		return Source{}, err
	}
	entry, err := reg.Resolve(name)
	if err != nil {
		if allowFallback && errors.Is(err, os.ErrNotExist) {
			return Source{
				URL: "https://github.com/Color-Of-Code/specs-framework.git",
				Ref: "main",
			}, nil
		}
		return Source{}, err
	}
	src := Source{URL: entry.URL, Ref: entry.Ref, Path: entry.Path}
	if refOverride != "" {
		if src.Path != "" {
			return Source{}, fmt.Errorf("framework %q is path-based; @ref is not allowed", name)
		}
		src.Ref = refOverride
	}
	return src, nil
}

func isPathSpec(spec string) bool {
	if spec == "." || spec == ".." {
		return true
	}
	if strings.HasPrefix(spec, "/") ||
		strings.HasPrefix(spec, "./") ||
		strings.HasPrefix(spec, "../") ||
		strings.HasPrefix(spec, "~") {
		return true
	}
	// Existing directory on disk (relative to cwd) is also a path.
	if info, err := os.Stat(spec); err == nil && info.IsDir() {
		return true
	}
	return false
}

func isRemoteURL(spec string) bool {
	for _, prefix := range []string{"https://", "http://", "ssh://", "git://", "git+ssh://", "file://"} {
		if strings.HasPrefix(spec, prefix) {
			return true
		}
	}
	// scp-style: user@host:path (no scheme, contains '@' before ':')
	if at := strings.Index(spec, "@"); at > 0 {
		if colon := strings.Index(spec[at:], ":"); colon > 0 {
			return true
		}
	}
	return false
}

// splitRef splits "<body>@<ref>" into (body, ref). The split is only applied
// when the '@' appears after the last '/' so URLs like "git@host:owner/repo"
// (which have an '@' early in the path) are not corrupted.
func splitRef(spec string) (body, ref string) {
	slash := strings.LastIndex(spec, "/")
	at := strings.LastIndex(spec, "@")
	if at > slash && at < len(spec)-1 {
		return spec[:at], spec[at+1:]
	}
	return spec, ""
}

func expandHome(p string) (string, error) {
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}
