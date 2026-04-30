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
Remote sources are fetched into the user cache (managed mode); local
paths are recorded in `framework_dir` and left untouched, so the host
can hold the framework as a plain folder, a git submodule, or a
vendored snapshot — whichever fits.

## Exit point

A directory containing `.specs.yaml` resolving the framework source,
the managed cache populated when the source is remote, and (when
requested) `model/`, `change-requests/`, and `.vscode/tasks.json`.
The directory is committable as-is. `--force` overwrites an existing
`.specs.yaml`; otherwise the command refuses.

## Workflow

1. Pick the framework source and pass it as `--framework <source>`
   (or rely on the registry's `default` entry).
2. Run `specs init` with `--dry-run` first to preview the file plan.
3. Re-run without `--dry-run` to materialise the host.
4. Run [`specs doctor`](diagnose-environment.md) to confirm paths,
   versions, and framework resolution.

### Iteration

Re-run with `--force` to rewrite `.specs.yaml`, or edit the file
manually for individual key changes (framework source, repo mappings,
lint config path). Switching from managed to a local checkout is just
a matter of replacing `framework_url` + `framework_ref` with
`framework_dir` in `.specs.yaml`.
