# Concepts

## Paths

Three paths matter and are referenced throughout the documentation and in `specs doctor`:

- **specs root** — the directory that contains `.specs.yaml`, `model/`, and `change-requests/`. This is what `specs` operates on.
- **host repo** — the git repository that contains the specs root. It can _be_ the specs root (when the repo is dedicated to specs) or contain it as a subdirectory or submodule.
- **tools dir** — where the generic framework content (`templates/`, `process/`, `skills/`, `agents/`, lint config — collectively `.specs-framework`) is materialised. In **managed** mode this is the user cache dir; in **dev** mode it is whatever `tools_dir` points at.

`specs doctor` prints all three so you can verify what was detected.

## Framework sources

A **framework source** is the origin from which `.specs-framework` content is obtained. Every `specs init` or `specs bootstrap` invocation resolves a framework source to populate the tools layer. Three kinds exist:

### Remote URL (default)

A git URL that the CLI clones or caches. This is how most users consume the official `specs-framework` or a company fork:

```yaml
tools_url: https://github.com/Color-Of-Code/specs-framework.git
tools_ref: v1.0.0
```

### Local path

An existing directory on disk (your own checkout, a vendored snapshot, or a submodule):

```yaml
tools_dir: ../specs-framework
```

### Empty seed

For organisations that want to build a framework **from scratch** without forking an existing one. The CLI pre-seeds an empty directory with the minimal directory skeleton expected by the toolchain (`templates/`, `process/`, `skills/`, `agents/`). This is an advanced, low-level option — it produces only the bare structure; all content must be authored by the caller.

```bash
specs framework seed --out /path/to/my-framework
```

> **Important:** the seeded directory is _not_ managed by the toolchain after creation. It is the caller's responsibility to initialise a git repository in the output folder, push it to a remote, and maintain it going forward. The toolchain only creates the initial skeleton.

Once the seeded directory is pushed to a remote you can reference it like any other framework source:

```yaml
tools_url: https://git.example.com/my-org/my-framework.git
tools_ref: main
```

Or point at it locally during development:

```yaml
tools_dir: ../my-framework
```

## Framework registry

To avoid remembering URLs and to standardise framework selection across teams, the toolchain supports a **framework registry** — a name → source mapping stored in user-level or project-level configuration.

Registry entries live in `~/.config/specs/frameworks.yaml` (XDG-compliant path on Linux; platform equivalents elsewhere):

```yaml
# ~/.config/specs/frameworks.yaml
frameworks:
  default:
    url: https://github.com/Color-Of-Code/specs-framework.git
    ref: v1.0.0
  my-company:
    url: https://git.example.com/acme/specs-framework.git
    ref: main
  local-dev:
    path: ~/src/specs-framework
```

With a registry configured, `specs init` and `specs bootstrap` accept `--framework <name>` instead of raw URLs:

```bash
specs init --framework my-company
specs bootstrap --framework local-dev
```

The `default` entry is used when no `--framework`, `--tools-url`, or `--tools-dir` flag is provided.

### Managing the registry

```bash
specs framework list                       # show registered entries
specs framework add <name> --url <URL> [--ref <ref>]
specs framework add <name> --path <dir>
specs framework remove <name>
specs framework seed --out <dir>           # create an empty skeleton
```

## Two ways to use the framework content

Regardless of the source, the framework content does **not** have to live inside the host repo. Pick one of two consumption modes:

### managed (default) — hidden, CLI-managed, read-only

The CLI fetches `.specs-framework` once into the user data dir and re-uses it across every host project on the machine. End users never see the content, never commit it, never update it manually.

- Location: `os.UserCacheDir()` + `/specs-toolchain/tools/<ref>/`. On Linux that resolves to `${XDG_CACHE_HOME:-~/.cache}/specs-toolchain/tools/<ref>/`; on macOS `~/Library/Caches/specs-toolchain/tools/<ref>/`; on Windows `%LocalAppData%\specs-toolchain\tools\<ref>`.
- Version pin: `tools_ref` in `.specs.yaml` (a tag or commit). The host commits **only** `.specs.yaml`; nothing else.
- Refreshing: `specs tools update --to <ref>` rewrites `tools_ref` and re-fetches if needed.
- This is what `specs bootstrap` and `specs init` give you by default.

### dev — a regular checkout you can edit

Use this when you are working on the framework itself (editing templates, process docs, skills). Clone `specs-framework` anywhere and point `tools_dir` at it:

```yaml
# .specs.yaml
tools_dir: ../specs-framework # or any absolute/relative path
```

Or keep the historical submodule layout if you want every contributor on a host project to see the content in-tree. Both submodule and plain-folder checkouts are auto-detected.

### Quick decision

| You are…                               | Use this                         |
| -------------------------------------- | -------------------------------- |
| writing specs in a host project        | **managed**                      |
| editing templates / process docs       | **dev**                          |
| starting a brand-new framework         | **seed** then switch to **dev**  |
| working air-gapped, no internet at all | **dev** with a vendored snapshot |

Switch modes any time by editing `.specs.yaml`; nothing else changes.
