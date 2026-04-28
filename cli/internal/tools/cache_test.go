package tools

import "testing"

func TestSanitize(t *testing.T) {
	cases := map[string]string{
		"":                  "default",
		"main":              "main",
		"release/v1":        "release/v1",
		"feature/foo bar":   "feature/foo_bar",
		"v1.2.3-rc.1":       "v1.2.3-rc.1",
		"users/x@y/branch":  "users/x_y/branch",
		"weird:chars*here?": "weird_chars_here_",
	}
	for in, want := range cases {
		if got := sanitize(in); got != want {
			t.Errorf("sanitize(%q)=%q want %q", in, got, want)
		}
	}
}

func TestLooksLikeRefName(t *testing.T) {
	cases := map[string]bool{
		"":           false,
		"main":       true,
		"release/v1": true,
		"v1.2.3":     true,
		"abc1234":    false, // 7 hex
		"abcdef0123456789abcdef0123456789abcdef01": false, // 40 hex
		"abcde":      true,  // <7
		"deadbeefXY": true,  // not all hex
		"DEADBEEF":   false, // 8 hex
	}
	for in, want := range cases {
		if got := looksLikeRefName(in); got != want {
			t.Errorf("looksLikeRefName(%q)=%v want %v", in, got, want)
		}
	}
}
