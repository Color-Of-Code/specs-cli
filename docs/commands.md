# Command reference

Every command below is reachable both as `specs <command>` on the terminal and from the VS Code palette as **Specs: …**. All write commands accept `--dry-run` where applicable.

## Core commands

| Command                                                                                                                             | Purpose                                                                                          |
| ----------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `specs version` / `--version`                                                                                                       | print the installed binary version                                                               |
| `specs doctor`                                                                                                                      | diagnose environment, layout, version drift                                                      |
| `specs init [--with-vscode] [--force] [--tools-url URL --tools-ref REF] [--framework <name>]`                                       | configure an existing host (writes `.specs.yaml`)                                                |
| `specs bootstrap [--at <path>] [--layout folder\|submodule] [--tools-mode managed\|submodule\|folder\|vendor] [--framework <name>]` | scaffold a new host (managed by default)                                                         |
| `specs lint [--all\|--links\|--style\|--baselines]`                                                                                 | run lint checks                                                                                  |
| `specs tools update [--to <ref>]`                                                                                                   | update the `.specs-framework` content layer                                                      |
| `specs scaffold <kind> [--cr <NNN>] [--title <t>] [--force] [--dry-run] <path>`                                                     | instantiate a template (`requirement\|feature\|component\|api\|service`)                         |
| `specs cr new --id <NNN> --slug <slug> [--title <t>] [--force] [--dry-run]`                                                         | create a new change request from the template tree                                               |
| `specs cr status`                                                                                                                   | list change requests with file counts per area                                                   |
| `specs cr drain --id <NNN> [--yes] [--dry-run]`                                                                                     | interactively `git mv` CR-local files to canonical model homes                                   |
| `specs baseline update [--only <substr>] [--dry-run]`                                                                               | rewrite stale SHAs in the Components table from `git log`                                        |
| `specs link check`                                                                                                                  | verify symmetry between requirements (`Implemented By`) and features/components (`Requirements`) |
| `specs visualize traceability [--format dot\|mermaid] [--out <path>]`                                                               | render the requirement ↔ implementer graph                                                       |
| `specs vscode init [--force]`                                                                                                       | write `.vscode/tasks.json` with every Specs task                                                 |

## Framework management commands

These commands manage the framework registry and support creating new frameworks from scratch. They are **not** needed for day-to-day specs authoring — only for framework maintainers and administrators.

| Command                                                | Purpose                                                   |
| ------------------------------------------------------ | --------------------------------------------------------- |
| `specs framework list`                                 | show all registered framework entries                     |
| `specs framework add <name> --url <URL> [--ref <ref>]` | register a remote framework source by name                |
| `specs framework add <name> --path <dir>`              | register a local directory as a framework source          |
| `specs framework remove <name>`                        | unregister a framework entry                              |
| `specs framework seed --out <dir>`                     | create an empty framework skeleton in the given directory |

### `specs framework seed`

Pre-seeds an empty directory with the minimal structure expected by the toolchain:

```text
<dir>/
├── templates/
├── process/
├── skills/
└── agents/
```

The command fails if the target directory already exists and is non-empty. After seeding, the caller is responsible for:

1. Running `git init` in the output directory.
2. Pushing it to a git remote for team use.
3. Registering it in the framework registry (or using `--tools-url` directly).

This is an **advanced** operation intended for organisations that need a bespoke framework rather than forking an existing one.
