# Roles

Operational hats people put on **outside** the authoring chain — setup, review, and framework distribution. For artifact ownership inside the authoring chain (Stakeholder, Author, Analyst, Architect), see [actors.md](actors.md).

One person typically wears several roles in the same repository.

| Role                 | Responsibility                                                                          | Typical commands                                                |
| -------------------- | --------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| Any user             | Get the engine working locally; curate the per-machine framework registry               | `specs doctor`, `specs framework add` / `list` / `remove`       |
| Project owner        | Stand up a host repository, choose `managed` vs. `local` mode, and seed editor tasks    | `specs init`, `specs vscode init`, `specs framework update`     |
| Reviewer             | Confirm a proposed change request is structurally sound and traceable before it lands   | `specs link check`, `specs visualize traceability`              |
| Framework maintainer | Create, evolve, and distribute framework content for downstream host repos to consume   | `specs framework seed`; publishes a framework repo and registry |

## Notes

- **Any user** owns their own machine's installation. Because `specs init` resolves frameworks exclusively through the registry, registering at least one entry (typically `default`) is part of the same one-time setup as installing the engine. There is no separate platform-admin role.
- **Project owner** picks one framework handling mode in `.specs.yaml`:
  - `managed` — the engine fetches the framework into the user cache; the host commits only `.specs.yaml`. This is the default.
  - `local` — `.specs.yaml` points at a directory on disk owned by the user (regular checkout, git submodule, or vendored snapshot — all treated the same).
- **Baselines** for tracked components are owned by the **Architect** actor (see [actors.md](actors.md)), not by a separate role — the same person who decomposes features into components keeps their pinned commits current.
- **Framework maintainer** is also a project owner and reviewer on the framework's own repository; the role here describes only the *downstream-facing* responsibility of publishing framework content.
