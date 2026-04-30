# Use cases

User-facing workflows exposed by the specs engine. Each file describes one use case as: **Summary**, **Owner**, **Purpose**, **Entry point**, **Exit point**, and (where the flow has decision points) a **Workflow** with iteration loops.

The **Owner** column below names whichever party is normally responsible. Authoring use cases are owned by an **actor** in the authoring chain (see [../actors.md](../actors.md)); setup, review, and maintenance use cases are owned by an operational **role** (see [../roles.md](../roles.md)). For a one-page summary of what the engine delivers, see [../overview.md](../overview.md).

## Authoring (day-to-day)

| Use case                                                         | Owner                                                        |
| ---------------------------------------------------------------- | ------------------------------------------------------------ |
| [Author a change request](author-change-request.md)              | Stakeholder → Author *(actors)*                              |
| [Scaffold a model artifact](scaffold-model-artifact.md)          | Author / Analyst / Architect *(actors)*                      |
| [Lint and format specifications](lint-and-format.md)             | Any authoring actor                                          |
| [Verify requirement ↔ implementer links](verify-traceability.md) | Reviewer *(role)* — typically Analyst / Architect themselves |
| [Maintain component baselines](maintain-baselines.md)            | Architect *(actor)*                                          |
| [Visualize the traceability graph](visualize-traceability.md)    | Any authoring actor; consumed by Reviewers and Stakeholders  |

## Setup and maintenance (one-off or occasional)

| Use case                                                      | Owner                              |
| ------------------------------------------------------------- | ---------------------------------- |
| [Set up a host](setup-host.md)                                | Project owner *(role)*             |
| [Diagnose the environment](diagnose-environment.md)           | Any user *(role)*                  |
| [Update the framework content layer](update-framework.md)     | Project owner *(role)*             |
| [Configure VS Code integration](configure-vscode.md)          | Project owner *(role)* or Any user |
| [Manage the framework registry](manage-framework-registry.md) | Any user *(role)*                  |
| [Seed a brand-new framework](seed-new-framework.md)           | Framework maintainer *(role)*      |
