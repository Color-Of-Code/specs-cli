# Configure VS Code integration

## Summary

Write `.vscode/tasks.json` so every specs engine command is reachable
from the VS Code task runner with no manual JSON authoring.

## Owner

**Project owner** *(role)* when seeding tasks for the team, or **Any user** *(role)* configuring their own checkout — see [../roles.md](../roles.md).

## Purpose

Give contributors who prefer the editor (over the terminal) a one-key
path to format, lint, scaffold, and visualize, without needing to
remember command flags.

## Entry point

`specs vscode init [--force]`

Or VS Code palette: **Specs: Init VS Code tasks**.

Pre-conditions: a writable `.vscode/` directory; `--force` overwrites
an existing `tasks.json`.

## Exit point

A `.vscode/tasks.json` containing one entry per supported specs task.

## Iteration

Re-run with `--force` after upgrading the engine to refresh task
definitions. The bundled VS Code extension picks up the same set of
commands automatically and does not require this file.
