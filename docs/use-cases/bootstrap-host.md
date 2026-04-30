# Bootstrap a new specs host

## Summary

Create a brand-new repository (or sub-tree) ready to author specs in: a
`.specs.yaml` configuration, the chosen framework-mode wiring, and optional
`model/` and `.vscode/tasks.json` scaffolding — in a single command.

## Purpose

Get from "empty directory" to "I can run `specs lint` and start writing
requirements" without manual setup or copy-pasting from a sibling repo.

## Entry point

`specs bootstrap [--at <path>] [--layout folder|submodule]
[--framework-mode managed|submodule|folder|vendor]
[--framework <name> | --framework-url <URL> --framework-ref <ref>]
[--specs-url <URL>] [--specs-ref <ref>]
[--with-model] [--with-vscode] [--dry-run]`

Or VS Code palette: **Specs: Bootstrap host**.

Pre-conditions: the engine binary is installed and on `PATH`; the target
directory is empty or new; for `--layout submodule` a remote `--specs-url`
is reachable.

## Exit point

A directory containing at minimum `.specs.yaml`, framework content
materialised according to `--framework-mode`, and (when requested) a `model/`
seed and `.vscode/tasks.json`. The directory is committable as-is.

## Workflow

1. Decide the **layout**: a self-contained folder (default) or a git
   submodule embedded in a host repo.
2. Decide the **framework mode**: `managed` (engine-cached, recommended),
   `submodule`, `folder` (local checkout), or `vendor` (snapshot copy).
3. Choose the **framework source**: a registered name (`--framework`), an
   explicit URL (`--framework-url` + `--framework-ref`), or fall back to the
   registry's `default` entry.
4. Run `specs bootstrap` with `--dry-run` first to preview the file plan.
5. Re-run without `--dry-run` to materialise the host.
6. Run [`specs doctor`](diagnose-environment.md) to confirm paths,
   versions, and framework-mode resolution.

### Iteration

If `doctor` reports drift or a wrong source, edit `.specs.yaml` directly
or re-run `bootstrap` against a clean directory; switching framework mode
later only requires editing `.specs.yaml`.
