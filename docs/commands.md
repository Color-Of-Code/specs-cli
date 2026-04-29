# Command reference

Every command below is reachable both as `specs <command>` on the terminal and from the VS Code palette as **Specs: …**. All write commands accept `--dry-run` where applicable.

| Command                                                                                                        | Purpose                                                                                          |
| -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `specs version` / `--version`                                                                                  | print the installed binary version                                                               |
| `specs doctor`                                                                                                 | diagnose environment, layout, version drift                                                      |
| `specs init [--with-vscode] [--force] [--tools-url URL --tools-ref REF]`                                       | configure an existing host (writes `.specs.yaml`)                                                |
| `specs bootstrap [--at <path>] [--layout folder\|submodule] [--tools-mode managed\|submodule\|folder\|vendor]` | scaffold a new host (managed by default)                                                         |
| `specs lint [--all\|--links\|--style\|--baselines]`                                                            | run lint checks                                                                                  |
| `specs tools update [--to <ref>]`                                                                              | update the `.specs-framework` content layer                                                      |
| `specs scaffold <kind> [--cr <NNN>] [--title <t>] [--force] [--dry-run] <path>`                                | instantiate a template (`requirement\|feature\|component\|api\|service`)                         |
| `specs cr new --id <NNN> --slug <slug> [--title <t>] [--force] [--dry-run]`                                    | create a new change request from the template tree                                               |
| `specs cr status`                                                                                              | list change requests with file counts per area                                                   |
| `specs cr drain --id <NNN> [--yes] [--dry-run]`                                                                | interactively `git mv` CR-local files to canonical model homes                                   |
| `specs baseline check`                                                                                         | verify component baselines (alias for `lint --baselines`)                                        |
| `specs baseline update [--only <substr>] [--dry-run]`                                                          | rewrite stale SHAs in the Components table from `git log`                                        |
| `specs link check`                                                                                             | verify symmetry between requirements (`Implemented By`) and features/components (`Requirements`) |
| `specs visualize traceability [--format dot\|mermaid] [--out <path>]`                                          | render the requirement ↔ implementer graph                                                       |
| `specs vscode init [--force]`                                                                                  | write `.vscode/tasks.json` with every Specs task                                                 |
