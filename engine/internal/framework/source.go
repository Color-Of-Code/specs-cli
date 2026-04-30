package framework

import (
	"errors"
	"fmt"
	"os"
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

// ResolveSource looks up a framework source in the user-level registry.
//
// `spec` is one of:
//   - ""              -> the registry's "default" entry
//   - "name"          -> the registry's entry for that name
//   - "name@ref"      -> the registry's entry, with a ref override
//
// URLs and filesystem paths are not accepted here: register them once with
// `specs framework add` and refer to them by name. This keeps `specs init`
// invocations short and consistent across machines.
func ResolveSource(spec string) (Source, error) {
	name, refOverride := splitRef(strings.TrimSpace(spec))
	if name == "" {
		name = "default"
	}

	reg, err := registry.Load("")
	if err != nil {
		return Source{}, err
	}
	entry, err := reg.Resolve(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			path, _ := registry.DefaultPath()
			return Source{}, fmt.Errorf("framework %q is not registered (see %s); add it with `specs framework add %s --url <git-url> [--ref <ref>]` or `specs framework add %s --path <dir>`", name, path, name, name)
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

// splitRef splits "<name>@<ref>" into (name, ref). A spec with no '@'
// yields (spec, "").
func splitRef(spec string) (name, ref string) {
	at := strings.LastIndex(spec, "@")
	if at <= 0 || at == len(spec)-1 {
		return spec, ""
	}
	return spec[:at], spec[at+1:]
}
