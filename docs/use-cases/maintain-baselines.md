# Maintain component baselines

## Summary

Verify or refresh the SHAs recorded for each tracked component in the
baselines table, so reviewers see exactly which upstream commit a
component spec was written against.

## Owner

**Architect** *(actor)* — see [../actors.md](../actors.md). Owns components and the baselines table that pins each of them to an upstream commit.

## Purpose

Detect when an upstream `repos:` entry has moved on without the spec
catching up, and provide a one-shot way to bring the table in sync.

## Entry point

- Verify: `specs lint --baselines`
- Update: `specs baseline update [--only <substr>] [--dry-run]`

Or VS Code palette: **Specs: Update baselines**.

Pre-conditions: `repos:` mappings in `.specs.yaml` resolve to local git
checkouts; the baselines file (default `model/baselines.md` or the
configured `baselines_file`) exists.

## Exit point

After **update**: the components table contains current `git log`
SHAs. After **lint --baselines**: zero exit when the table is in sync,
non-zero with a per-component diff otherwise.

## Workflow

1. Run `specs lint --baselines` to detect drift.
2. If drift is reported, run `specs baseline update --dry-run` to
   preview rewrites.
3. Re-run without `--dry-run` (optionally narrowing with `--only`).
4. Review the resulting diff in the baselines file; commit alongside
   any related spec changes.

### Iteration

Repeat per CR or per release. The `--only <substr>` flag scopes
updates to a subset of components when only some are in scope.
