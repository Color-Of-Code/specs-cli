# Overview

A single-page summary of what the **specs engine** delivers. Everything below
is a feature of the `specs` binary; the VS Code extension is a thin wrapper
and is not covered here.

## What the engine is for

A small CLI that lets a team author a structured specification model in
markdown — requirements, features, components, services, APIs — with
guaranteed formatting, traceability, and baseline integrity.

## The model

Four artifact kinds, scaffolded from templates:

| Kind          | Owned by    | Purpose                                                   |
| ------------- | ----------- | --------------------------------------------------------- |
| Requirement   | Author      | A single, testable statement of what the system shall do. |
| Feature       | Analyst     | A grouping that implements one or more requirements.      |
| Component     | Architect   | A unit of implementation pinned to an upstream repo.      |
| Service / API | Architect   | An interface between components or with the outside.      |

All artifacts live under `model/` and are written in markdown.

## The authoring chain

```text
Stakeholder ──► Author ──► Analyst ──► Architect
   input        requirements   features    components / services / APIs
```

Stakeholders describe needs as **change requests**; authors, analysts, and
architects refine that input into model artifacts. See
[actors.md](actors.md) for details.

## The change-request lifecycle

Work happens inside numbered change-request folders so it can be reviewed
in isolation before joining the canonical model.

```text
specs cr new      ──►  draft inside change-requests/NNN-slug/
specs scaffold    ──►  add requirements / features / components / ...
specs format      ──►  normalise markdown
specs lint        ──►  style + links + baselines
specs cr drain    ──►  git mv into canonical model paths
```

## What the engine guarantees

- **Consistent markdown** — `specs format` rewrites files in place,
  `specs format --check` is a CI gate.
- **Style compliance** — `specs lint --style` enforces the configured
  rules (defaults compiled in, overridable via `style.yaml`).
- **Bidirectional traceability** — `specs link check` verifies that every
  requirement listed as `Implemented By` is reciprocally listed in its
  feature/component, and vice versa.
- **Component baselines** — `specs lint --baselines` detects drift between
  recorded SHAs and the real upstream commits;
  `specs baseline update` rewrites the table.
- **Visual traceability** — `specs visualize traceability` renders the
  graph as DOT or Mermaid.
- **Diagnostics** — `specs doctor` prints engine version, resolved paths,
  framework mode, and version drift.

## What the engine does *not* do

- Render or publish documents — output is markdown, period.
- Manage tickets, sprints, or workflow approvals.
- Lock authors out of files — coordination is via change requests and
  normal git review.

## Where things live

| Path                        | Contents                                                                                                                   |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `.specs.yaml`               | Per-host configuration; resolves paths and framework.                                                                      |
| `model/`                    | Canonical model artifacts.                                                                                                 |
| `change-requests/NNN-slug/` | Work in progress, drained into `model/` on merge.                                                                          |
| `.specs-framework`          | Framework content (templates, lint config). Either fetched into the user cache (managed) or supplied as a local directory. |

## Getting started

1. Install the binary: see [install.md](install.md).
2. Initialise a host: `specs init`.
3. Verify the setup: `specs doctor`.
4. Open a change request and start authoring: see [actors.md](actors.md)
   and [use-cases/](use-cases/README.md).

## Reference

- [Concepts](concepts.md) — paths, framework sources, modes.
- [Commands](commands.md) — every `specs` subcommand.
- [Configuration](configuration.md) — `.specs.yaml` and registry schemas.
- [Use cases](use-cases/README.md) — task-oriented workflows.
