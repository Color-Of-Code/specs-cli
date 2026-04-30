# Roles

Operational hats people put on while using the toolchain. One person typically wears several of them in the same repository. The model-authoring chain itself (stakeholder → author → analyst → architect) is described separately in [actors.md](actors.md); this page covers the surrounding setup, review, and maintenance work.

| Role             | Responsibility                                                                                | Typical commands                                                |
| ---------------- | --------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| Any user         | Get the engine working locally; curate the per-machine framework registry                     | `specs doctor`, `specs framework add` / `list` / `remove`       |
| Project owner    | Stand up a host repository, choose `managed` vs. `local` mode, and seed editor tasks          | `specs init`, `specs vscode init`                               |
| Reviewer         | Confirm a change request is structurally sound and traceable                                  | `specs link check`, `specs visualize traceability`              |
| Component owner  | Keep recorded component baselines aligned with the upstream repository revision they describe | `specs lint --baselines`, `specs baseline update`               |
| Framework author | Create, evolve, and distribute framework content                                              | `specs framework seed`, `specs framework update` on downstreams |

## Notes

- **Any user** owns their own machine's installation. Because `specs init` resolves frameworks exclusively through the registry, registering at least one entry (typically `default`) is part of the same one-time setup as installing the engine. There is no separate platform-admin role.
- **Project owner** picks one framework handling mode in `.specs.yaml`:
  - `managed` — the engine fetches the framework into the user cache; the host commits only `.specs.yaml`. This is the default.
  - `local` — `.specs.yaml` points at a directory on disk owned by the user (regular checkout, git submodule, or vendored snapshot — all treated the same).
- **Framework author** publishes a framework repository and registers it (or documents how consumers should). Day-to-day they also act as a project owner and reviewer on their own framework repo.
