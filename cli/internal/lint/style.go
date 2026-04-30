package lint

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed style_defaults.yaml
var defaultStyleYAML []byte

// StyleConfig holds the abstract, implementation-independent style rules.
type StyleConfig struct {
	Rules StyleRules `yaml:"rules"`
}

// StyleRules maps abstract rule names to their settings.
type StyleRules struct {
	LineLength               interface{} `yaml:"line_length"`      // false (disabled) or int
	InlineHTML               bool        `yaml:"inline_html"`      // allow inline HTML
	FirstHeadingH1           bool        `yaml:"first_heading_h1"` // require first line to be h1
	HeadingStyle             string      `yaml:"heading_style"`    // "atx" or "setext"
	BlankLinesAroundHeadings bool        `yaml:"blank_lines_around_headings"`
	BlankLinesAroundFences   bool        `yaml:"blank_lines_around_fences"`
	ListMarker               string      `yaml:"list_marker"` // "dash" or "asterisk"
	NoTrailingWhitespace     bool        `yaml:"no_trailing_whitespace"`
	NoConsecutiveBlankLines  bool        `yaml:"no_consecutive_blank_lines"`
	FencedCodeLanguage       bool        `yaml:"fenced_code_language"`
}

// LineLengthLimit returns the max line length, or 0 if disabled.
func (s *StyleRules) LineLengthLimit() int {
	switch v := s.LineLength.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case bool:
		return 0
	default:
		return 0
	}
}

// LoadStyleConfig loads a style.yaml from path, merging it over compiled-in
// defaults. If path is empty, only the defaults are used.
func LoadStyleConfig(path string) (*StyleConfig, error) {
	cfg := &StyleConfig{}
	if err := yaml.Unmarshal(defaultStyleYAML, cfg); err != nil {
		return nil, fmt.Errorf("parse embedded style defaults: %w", err)
	}
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read style config %s: %w", path, err)
	}
	// Unmarshal on top of defaults so unset keys keep their default values.
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse style config %s: %w", path, err)
	}
	return cfg, nil
}
