# Update the framework content layer

## Summary

Refresh `.specs-framework` content (templates, process, skills, agents,
lint config) to a newer ref of the configured framework source.

## Actors

Maintenance task — performed by whoever maintains the host repo. Not
part of the authoring chain.

## Purpose

Pick up template fixes, new lint rules, or process updates published by
the framework maintainer without re-running the full setup.

## Entry point

`specs framework update [--to <ref>]`

Or VS Code palette: **Specs: Framework: Update cache**.

Pre-conditions: the host is in `managed` framework mode (or a writable
checkout); network access to the framework source URL.

## Exit point

`framework_ref` in `.specs.yaml` is rewritten to the new ref and the
content is re-fetched into the cache. The host repo's only diff is
`.specs.yaml`.

## Iteration

After updating, run [`specs doctor`](diagnose-environment.md) and
[`specs lint`](lint-and-format.md) to surface any new style rules or
template-schema changes; address findings before merging the bump.
