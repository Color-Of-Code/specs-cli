package main

import "testing"

func TestRewriteBaselineRow_PreservesPadding(t *testing.T) {
	// Cell 4 (the SHA) has surrounding whitespace that should survive.
	line := "| comp | repo | / | `oldsha` | 2025-01-01 |"
	repos := map[string]string{"repo": "."}
	// We can't actually run `git log` in unit tests; use a path-style that
	// triggers the placeholder skip path so the function returns without
	// invoking git.
	skip := "| comp | repo | _placeholder_ | `oldsha` | 2025-01-01 |"
	got, action, err := rewriteBaselineRow(skip, repos, t.TempDir(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "skip-placeholder" {
		t.Fatalf("expected skip-placeholder, got %q", action)
	}
	if got != skip {
		t.Fatalf("placeholder rows must be returned unchanged\nin:  %q\nout: %q", skip, got)
	}

	// Filter mismatch -> skip-filter, line unchanged.
	got, action, err = rewriteBaselineRow(line, repos, t.TempDir(), "other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "skip-filter" {
		t.Fatalf("expected skip-filter, got %q", action)
	}
	if got != line {
		t.Fatalf("filtered rows must be returned unchanged")
	}

	// Unknown repo -> error.
	_, _, err = rewriteBaselineRow(line, map[string]string{}, t.TempDir(), "")
	if err == nil {
		t.Fatal("expected error for unknown repo")
	}

	// Non-table line -> skip-placeholder.
	got, action, _ = rewriteBaselineRow("not a table line", repos, t.TempDir(), "")
	if action != "skip-placeholder" || got != "not a table line" {
		t.Fatalf("non-table lines must skip; got action=%q out=%q", action, got)
	}
}

func TestReplacePreservingPadding(t *testing.T) {
	cases := []struct {
		in    string
		value string
		want  string
	}{
		{" `oldsha` ", "newsha", " `newsha` "},
		{"   `x`   ", "y", "   `y`   "},
		{"`x`", "y", "`y`"},
		{"  ", "y", "  `y`  "},
	}
	for _, tc := range cases {
		got := replacePreservingPadding(tc.in, tc.value)
		if got != tc.want {
			t.Errorf("replacePreservingPadding(%q,%q)=%q want %q", tc.in, tc.value, got, tc.want)
		}
	}
}
