# Use cases

User-facing workflows exposed by the specs engine. Each file describes one use
case as: **Summary**, **Purpose**, **Entry point**, **Exit point**, and (where
the flow has decision points) a **Workflow** with iteration loops.

For the actor model these workflows reference, see [../actors.md](../actors.md).
For a one-page summary of what the engine delivers, see
[../overview.md](../overview.md).

## Authoring (day-to-day)

| Use case                                                         | Actor                        |
| ---------------------------------------------------------------- | ---------------------------- |
| [Author a change request](author-change-request.md)              | Stakeholder / Author         |
| [Scaffold a model artifact](scaffold-model-artifact.md)          | Author / Analyst / Architect |
| [Lint and format specifications](lint-and-format.md)             | any actor                    |
| [Verify requirement ↔ implementer links](verify-traceability.md) | Analyst / Architect          |
| [Maintain component baselines](maintain-baselines.md)            | Architect                    |
| [Visualize the traceability graph](visualize-traceability.md)    | any actor                    |

## Setup (one-off)

| Use case                                                      | When                                     |
| ------------------------------------------------------------- | ---------------------------------------- |
| [Set up a host](setup-host.md)                                | Starting a new repo or onboarding one    |
| [Diagnose the environment](diagnose-environment.md)           | Verifying the install or troubleshooting |
| [Update the framework content layer](update-framework.md)     | Picking up a new framework release       |
| [Configure VS Code integration](configure-vscode.md)          | Optional editor wiring                   |
| [Manage the framework registry](manage-framework-registry.md) | Per-machine framework shortcuts          |
| [Seed a brand-new framework](seed-new-framework.md)           | Authoring a bespoke framework            |
