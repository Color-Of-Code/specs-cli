# Set up a host

## Summary

Create or configure a host repository for use with the specs toolchain in
a single command: `specs init`. It works whether the target directory is
empty, brand-new, or already contains `model/` and `change-requests/`
content.

## Owner

**Project owner** *(role)* — see [../roles.md](../roles.md). Stands up the host repo and chooses managed vs. local mode. Not part of the authoring chain.

## Purpose

Get from "I have a folder" to "I can run `specs lint` and start
authoring" with no manual config or copy-pasting from a sibling repo.
Existing model content is left untouched.

## Entry point

```text
specs init [<path>]
           [--framework <name>[@ref]]
           [--with-model] [--with-vscode]
           [--force] [--dry-run]
```

`<path>` defaults to the current directory and is created if missing.
`--framework` takes a name previously registered with
[`specs framework add`](manage-framework-registry.md): for example
`default`, `acme`, or `acme@v2.1` to override the registered ref.
When `--framework` is omitted the registry's `default` entry is used.

If no entries are registered, `specs init` fails with a hint pointing at
`specs framework add`. URLs and filesystem paths are not accepted on the
`specs init` command line: register them once and refer to them by name.

## Exit point

A directory containing `.specs.yaml` resolving the framework source,
the managed cache populated when the registered entry is URL-based, and
(when requested) `model/`, `change-requests/`, and
`.vscode/tasks.json`. The directory is committable as-is. `--force`
overwrites an existing `.specs.yaml`; otherwise the command refuses.

## Workflow

1. Make sure the framework you want to use is registered (see
   [manage the framework registry](manage-framework-registry.md)). The
   `default` entry is what `specs init` picks when `--framework` is
   omitted.
2. Run `specs init --framework <name>` (or just `specs init` to use
   `default`) with `--dry-run` first to preview the file plan.
3. Re-run without `--dry-run` to materialise the host.
4. Run [`specs doctor`](diagnose-environment.md) to confirm paths,
   versions, and framework resolution.

### Iteration

Re-run with `--force` to rewrite `.specs.yaml`, or edit the file
manually for individual key changes (framework source, repo mappings,
lint config path). Switching from a managed entry to a local checkout
is just a matter of replacing `framework_url` + `framework_ref` with
`framework_dir` in `.specs.yaml`, or registering a new framework name
and re-running `specs init --force --framework <new-name>`.
