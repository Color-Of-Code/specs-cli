# Seed a brand-new framework

## Summary

Pre-create the minimal directory skeleton of a fresh framework
(`templates/`, `process/`, `skills/`, `agents/`) in an empty
directory, ready for the caller to author content and publish to a git
remote.

## Actors

Advanced setup task — performed once by an organisation that needs a
bespoke framework. Not part of the authoring chain in a host project.

## Purpose

Support organisations that need a bespoke framework rather than forking
the default one. This is an advanced, low-level operation: the seeded
tree is **not** managed by the toolchain after creation.

## Entry point

`specs framework seed --out <dir>`

Pre-conditions: `<dir>` does not exist or is empty.

## Exit point

The output directory contains the empty skeleton. The caller is
responsible for `git init`, pushing to a remote, and registering the
result via [`specs framework add`](manage-framework-registry.md).

## Workflow

1. Run `specs framework seed --out /path/to/my-framework`.
2. Author templates, process docs, skills, and agents.
3. `git init`, commit, push to a remote.
4. Register: `specs framework add my-org --url <remote> --ref main`.
5. Use it: [`specs init --framework my-org`](setup-host.md).

### Iteration

Once published, evolve the framework like any other repository.
Consumers pick up changes via
[`specs framework update`](update-framework.md).
