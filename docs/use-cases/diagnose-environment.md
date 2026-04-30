# Diagnose the environment

## Summary

Print the engine version, resolved specs root, host repo, framework dir,
framework mode, and any version-drift warnings between the binary and the
framework content.

## Purpose

A first-call diagnostic for any "something's not working" report. It
answers: *Is my engine new enough? Where is it reading from? Which
framework version is materialised?*

## Entry point

`specs doctor`

Or VS Code palette: **Specs: Doctor**.

## Exit point

A human-readable report on stdout. Exit code is non-zero if a hard
incompatibility is detected (e.g. `min_specs_version` exceeds the
running binary). Otherwise zero, with warnings printed to stderr.

## Iteration

When `doctor` flags drift, fix forward: bump the engine, run
[`specs tools update`](update-framework.md), or correct
`.specs.yaml`, then re-run `specs doctor` until clean.
