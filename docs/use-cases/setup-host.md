# Set up a host

## Summary

Create or configure a host repository for use with the specs toolchain in
a single command: `specs init`. It works whether the target directory is
empty, brand-new, or already contains `model/` and `change-requests/`
content.

## Actors

One-off setup task — performed by whoever stands up the host repo. Not
part of the authoring chain.

## Purpose

Get from "I have a folder" to "I can run `specs lint` and start
authoring" with no manual config or copy-pasting from a sibling repo.
Existing model content is left untouched.

## Entry point

```text
specs init [<path>]
           [--framework <source>]
           [--framework-mode managed|submodule|folder|vendor]
           [--with-model] [--with-vscode]
           [--force] [--dry-run]
```

`<path>` defaults to the current directory and is created if missing.
`--framework` accepts:

- a registered name: `default`, `acme`, `local-dev`
- a name with ref override: `acme@v2.1`
- a remote git URL: `https://github.com/foo/bar.git[@ref]`,
  `git@host:owner/repo.git[@ref]`
- a local path: `./fw`, `../specs-framework`, `/abs/dir`

When `--framework` is omitted the registry's `default` entry is used.
Local paths skip framework materialisation; `--framework-mode` only
applies to remote sources.

## Exit point

A directory containing `.specs.yaml` resolving the framework source,
plus framework content materialised according to `--framework-mode`,
and (when requested) `model/`, `change-requests/`, and
`.vscode/tasks.json`. The directory is committable as-is. `--force`
overwrites an existing `.specs.yaml`; otherwise the command refuses.

## Workflow

1. Pick the framework source and pass it as `--framework <source>`
   (or rely on the registry's `default` entry).
2. Pick the framework mode: `managed` (engine-cached, recommended),
   `submodule`, `folder`, or `vendor`.
3. Run `specs init` with `--dry-run` first to preview the file plan.
4. Re-run without `--dry-run` to materialise the host.
5. Run [`specs doctor`](diagnose-environment.md) to confirm paths,
   versions, and framework-mode resolution.

### Iteration

Re-run with `--force` to rewrite `.specs.yaml`, or edit the file
manually for individual key changes (framework mode, repo mappings,
lint config path). Switching framework mode later only requires
editing `.specs.yaml`.
