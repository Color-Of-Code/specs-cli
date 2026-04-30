# Lint and format specifications

## Summary

Auto-format markdown files in place, and run style / link / baseline
lint checks against the specs tree.

## Owner

Any authoring actor — **Author**, **Analyst**, and **Architect** all run these commands while iterating on a CR. Also run by CI on every push. See [../actors.md](../actors.md).

## Purpose

Keep the corpus consistent (table alignment, list markers, blank lines,
trailing whitespace) and catch broken cross-references and stale
component baselines before they reach review.

## Entry point

- Format: `specs format [--check] [--at <path>] [files...]`
- Lint: `specs lint [--all] [--links] [--style] [--baselines]` (no flag
  runs everything).

Or VS Code palette: **Specs: Format**, **Specs: Lint**.

## Exit point

- `specs format`: files rewritten in place; `--check` exits non-zero if
  any file would have changed (CI gate).
- `specs lint`: exit code zero on success; non-zero with a per-file
  diagnostic list otherwise.

## Workflow

1. Run `specs format` to normalise formatting.
2. Run `specs lint --style` for markdown style issues.
3. Run `specs lint --links` to catch broken intra-spec links.
4. Run `specs lint --baselines` if components are tracked.
5. Fix reported issues and re-run.

### Iteration

Repeat 1–5 until clean. In CI, prefer `specs format --check` followed
by `specs lint` so violations fail the build instead of being silently
auto-fixed.
