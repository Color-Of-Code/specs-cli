# Initialize specs in an existing repo

## Summary

Add specs configuration (`.specs.yaml`, optional VS Code tasks) to a
repository that already exists and may already contain a `model/` or
`change-requests/` tree.

## Purpose

Onboard an existing project to the specs toolchain non-destructively:
preserve any existing model content, register the chosen framework
source, and make the engine commands runnable from the repo root.

## Entry point

`specs init [--at <path>] [--force] [--with-vscode]
[--framework <name>]
[--framework-url <URL> --framework-ref <ref> | --framework-dir <dir>]`

Or VS Code palette: **Specs: Init host**.

Pre-conditions: the target directory exists; no `.specs.yaml` is present
unless `--force` is passed; `--framework-url` and `--framework-dir` are not
combined.

## Exit point

A `.specs.yaml` written at the target path resolving the framework
source, plus `.vscode/tasks.json` when `--with-vscode` is set. Existing
`model/` and `change-requests/` content is untouched.

## Workflow

1. Pick the framework source: a `--framework <name>` registry entry, an
   explicit `--framework-url`/`--framework-ref`, or `--framework-dir` for a local
   checkout. With no flag the registry's `default` entry wins.
2. Run `specs init` from the repo root.
3. Run [`specs doctor`](diagnose-environment.md) to verify resolved
   paths and version compatibility.
4. Optionally run [`specs vscode init`](configure-vscode.md) later to
   add palette tasks.

### Iteration

Re-run `specs init --force` to overwrite the configuration, or edit
`.specs.yaml` manually for individual key changes (framework mode, repo
mappings, lint config path).
