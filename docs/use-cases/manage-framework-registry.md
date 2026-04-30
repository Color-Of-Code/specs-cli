# Manage the framework registry

## Summary

Maintain the user-level mapping from short framework names (e.g.
`default`, `acme`, `local-dev`) to git URLs or local paths, so
`specs init` can be invoked with `--framework <name>` instead of raw
URLs.

## Actors

Per-machine setup task — performed by each contributor on their own
workstation. Not part of the authoring chain.

## Purpose

Standardise framework selection across teams and machines, and keep
URLs out of every contributor's shell history.

## Entry point

- `specs framework list`
- `specs framework add <name> --url <URL> [--ref <ref>]`
- `specs framework add <name> --path <dir>`
- `specs framework remove <name>`

The registry file lives at the platform's user config directory (Linux:
`~/.config/specs/frameworks.yaml`).

## Exit point

The registry file is created or updated; `specs framework list` shows
the current entries.

## Workflow

1. Add the team's primary framework as `default` so flag-less invocations
   resolve to it.
2. Add additional named entries for forks, vendor frameworks, or local
   working copies (`--path`).
3. Run [`specs init`](setup-host.md) with `--framework <name>` to pick a
   non-default entry.

### Iteration

Re-run `add` to update an entry (it overwrites by name) or `remove`
followed by `add` for a clean replace.
