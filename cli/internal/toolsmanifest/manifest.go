// Package toolsmanifest reads tools-manifest.yaml from a tools_dir to
// support compatibility checks between the host's pinned templates_schema
// and the materialised content layer.
package toolsmanifest

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileName is the canonical name of the manifest at the root of a tools_dir.
const FileName = "tools-manifest.yaml"

// Manifest mirrors the on-disk schema. Unknown fields are tolerated so the
// manifest can grow without breaking older binaries.
type Manifest struct {
	TemplatesSchema int    `yaml:"templates_schema,omitempty"`
	Version         string `yaml:"version,omitempty"`
}

// Load reads <toolsDir>/tools-manifest.yaml. Returns (nil, nil) when the
// file is absent so callers can choose to enforce or ignore.
func Load(toolsDir string) (*Manifest, error) {
	if toolsDir == "" {
		return nil, nil
	}
	p := filepath.Join(toolsDir, FileName)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", p, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	return &m, nil
}

// Compatible reports whether a host's required schema is satisfied by the
// content's manifest. The contract is: the manifest's templates_schema must
// equal the host's required value. (Future: range/min compatibility.)
//
// requiredSchema == 0 means the host did not pin a schema; always
// compatible.
func Compatible(requiredSchema int, m *Manifest) (bool, string) {
	if requiredSchema == 0 {
		return true, ""
	}
	if m == nil {
		return false, fmt.Sprintf("host pins templates_schema=%d but the tools dir has no %s", requiredSchema, FileName)
	}
	if m.TemplatesSchema == 0 {
		return false, fmt.Sprintf("host pins templates_schema=%d but %s does not declare templates_schema", requiredSchema, FileName)
	}
	if m.TemplatesSchema != requiredSchema {
		return false, fmt.Sprintf("templates_schema mismatch: host requires %d, tools dir declares %d", requiredSchema, m.TemplatesSchema)
	}
	return true, ""
}
