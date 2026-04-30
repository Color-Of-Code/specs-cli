# Author a change request

## Summary

Create a numbered change request (CR), draft requirements / features /
components inside it, then drain the finalised files into their
canonical homes under `model/`.

## Owner

The authoring chain itself — see [../actors.md](../actors.md):

- **Stakeholder** *(actor)* (initiator) opens the CR with input describing the need.
- **Author** *(actor)* drafts the requirements, **Analyst** *(actor)* the features, and **Architect** *(actor)* the components / services / APIs inside the same CR.

## Purpose

Keep work-in-progress isolated under `change-requests/NNN-slug/` while
it is being reviewed, and only promote it to the durable model once
merged. Avoid stepping on other authors editing the same area.

## Entry point

`specs cr new --id <NNN> --slug <slug> [--title <t>] [--force] [--dry-run]`

Or VS Code palette: **Specs: New change request**.

Pre-conditions: the host is initialised (`.specs.yaml` exists); the CR
id is not already taken (unless `--force`).

## Exit point

After **drain**: all CR-local model files are `git mv`d into the
canonical model paths and the original CR folder is empty (or removed),
ready for the merge commit.

## Workflow

1. **Create** the CR shell with `specs cr new`. This instantiates the
   `change-request` template tree under `change-requests/NNN-slug/`.
2. **Author** content. Use [`specs scaffold`](scaffold-model-artifact.md)
   with `--cr <NNN>` to add requirements, features, components, APIs,
   or services *inside* the CR folder.
3. **Iterate** locally:
    - Run [`specs format`](lint-and-format.md) to keep markdown tidy.
    - Run [`specs lint`](lint-and-format.md) for style and
      cross-reference checks.
    - Run [`specs cr status`](../commands.md) to list CRs and per-area file counts.
4. **Review** with collaborators on the CR branch.
5. **Drain** with `specs cr drain --id <NNN>` (use `--dry-run` first,
   then `--yes` to apply). The engine `git mv`s files into their
   canonical model locations.
6. **Verify** with [`specs link check`](verify-traceability.md) and
   [`specs lint --baselines`](maintain-baselines.md) before merging.

### Iteration

Steps 2–4 repeat until the CR is approved. After drain, if the merge
introduces conflicts, resolve them and re-run `lint` + `link check`.
