# Concepts

## Paths

Three paths matter and are referenced throughout the documentation and in `specs doctor`:

- **specs root** — the directory that contains `.specs.yaml`, `model/`, and `change-requests/`. This is what `specs` operates on.
- **host repo** — the git repository that contains the specs root. It can _be_ the specs root (when the repo is dedicated to specs) or contain it as a subdirectory or submodule.
- **tools dir** — where the generic framework content (`templates/`, `process/`, `skills/`, `agents/`, lint config — collectively `.specs-tools`) is materialised. In **managed** mode this is the user cache dir; in **dev** mode it is whatever `tools_dir` points at.

`specs doctor` prints all three so you can verify what was detected.

## Two ways to use the framework content

The framework content does **not** have to live inside the host repo. Pick one:

### managed (default) — hidden, CLI-managed, read-only

The CLI fetches `.specs-tools` once into the user data dir and re-uses it across every host project on the machine. End users never see the content, never commit it, never update it manually.

- Location: `os.UserCacheDir()` + `/specs-cli/tools/<ref>/`. On Linux that resolves to `${XDG_CACHE_HOME:-~/.cache}/specs-cli/tools/<ref>/`; on macOS `~/Library/Caches/specs-cli/tools/<ref>/`; on Windows `%LocalAppData%\specs-cli\tools\<ref>\`.
- Version pin: `tools_ref` in `.specs.yaml` (a tag or commit). The host commits **only** `.specs.yaml`; nothing else.
- Refreshing: `specs tools update --to <ref>` rewrites `tools_ref` and re-fetches if needed.
- This is what `specs bootstrap` and `specs init` give you by default.

### dev — a regular checkout you can edit

Use this when you are working on the framework itself (editing templates, process docs, skills). Clone `specs-tools` anywhere and point `tools_dir` at it:

```yaml
# .specs.yaml
tools_dir: ../specs-tools # or any absolute/relative path
```

Or keep the historical submodule layout if you want every contributor on a host project to see the content in-tree. Both submodule and plain-folder checkouts are auto-detected.

### Quick decision

| You are…                               | Use this                         |
| -------------------------------------- | -------------------------------- |
| writing specs in a host project        | **managed**                      |
| editing templates / process docs       | **dev**                          |
| working air-gapped, no internet at all | **dev** with a vendored snapshot |

Switch modes any time by editing `.specs.yaml`; nothing else changes.
