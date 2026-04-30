// Package registry implements the user-level framework registry: a name -> source
// mapping stored at ~/.config/specs/frameworks.yaml (or the platform equivalent).
//
// The registry lets users assign short names (e.g. "default", "acme", "local-dev")
// to framework sources and reference them via `--framework <name>` on init and
// bootstrap, instead of pasting URLs.
package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// FileName is the on-disk filename of the registry.
const FileName = "frameworks.yaml"

// Entry is a single framework source. Exactly one of URL or Path must be set.
// Ref is optional; it applies only to URL-based entries (defaults to "main"
// when empty).
type Entry struct {
	URL  string `yaml:"url,omitempty"`
	Ref  string `yaml:"ref,omitempty"`
	Path string `yaml:"path,omitempty"`
}

// Validate returns an error if the entry is malformed.
func (e Entry) Validate() error {
	if e.URL == "" && e.Path == "" {
		return errors.New("entry must set either url or path")
	}
	if e.URL != "" && e.Path != "" {
		return errors.New("entry must not set both url and path")
	}
	if e.Path != "" && e.Ref != "" {
		return errors.New("ref is only valid for url entries")
	}
	return nil
}

// Registry is the in-memory representation of frameworks.yaml.
type Registry struct {
	Frameworks map[string]Entry `yaml:"frameworks,omitempty"`
}

// DefaultPath returns the platform-specific registry path. It does not
// create any directories; callers that want to write should call Save.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(dir, "specs", FileName), nil
}

// Load reads the registry from path. If path is empty, DefaultPath is used.
// A missing file yields an empty registry (not an error).
func Load(path string) (*Registry, error) {
	if path == "" {
		p, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = p
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Registry{Frameworks: map[string]Entry{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var r Registry
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if r.Frameworks == nil {
		r.Frameworks = map[string]Entry{}
	}
	for name, entry := range r.Frameworks {
		if err := entry.Validate(); err != nil {
			return nil, fmt.Errorf("registry entry %q: %w", name, err)
		}
	}
	return &r, nil
}

// Save writes the registry to path, creating parent directories as needed.
// If path is empty, DefaultPath is used.
func (r *Registry) Save(path string) error {
	if path == "" {
		p, err := DefaultPath()
		if err != nil {
			return err
		}
		path = p
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}
	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// Names returns the registered framework names in sorted order.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.Frameworks))
	for n := range r.Frameworks {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Resolve looks up name in the registry. Returns os.ErrNotExist if absent.
func (r *Registry) Resolve(name string) (Entry, error) {
	if name == "" {
		return Entry{}, errors.New("empty framework name")
	}
	e, ok := r.Frameworks[name]
	if !ok {
		return Entry{}, fmt.Errorf("framework %q: %w", name, os.ErrNotExist)
	}
	return e, nil
}

// Add inserts or replaces an entry. Returns an error if the entry is invalid.
func (r *Registry) Add(name string, entry Entry) error {
	if name == "" {
		return errors.New("name is required")
	}
	if err := entry.Validate(); err != nil {
		return err
	}
	if r.Frameworks == nil {
		r.Frameworks = map[string]Entry{}
	}
	r.Frameworks[name] = entry
	return nil
}

// Remove deletes the named entry. Returns os.ErrNotExist if absent.
func (r *Registry) Remove(name string) error {
	if _, ok := r.Frameworks[name]; !ok {
		return fmt.Errorf("framework %q: %w", name, os.ErrNotExist)
	}
	delete(r.Frameworks, name)
	return nil
}
