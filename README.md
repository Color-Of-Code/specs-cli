# specs-cli

User-scope CLI for the [specs framework](https://github.com/jdehaan/specs-tools). A single cross-platform Go binary that handles lint, scaffolding, change-request lifecycle, traceability links, and baseline updates for any host project that uses the framework.

## Install

One binary per developer, shared across all host projects:

```bash
go install github.com/jdehaan/specs-cli/cmd/specs@latest
```

`go install` puts the binary at `$(go env GOBIN)` if set, otherwise `$(go env GOPATH)/bin` (typically `~/go/bin` on Linux/macOS, `%USERPROFILE%\go\bin` on Windows). Make sure that directory is on `PATH`. Release tarballs from GitHub Releases work too — drop `specs` anywhere on `PATH`.

Verify:

```bash
specs --version
specs doctor
```

## Two ways to use the framework content

The framework content (`templates/`, `process/`, `skills/`, `agents/`, lint config — collectively `.specs-tools`) does **not** have to live inside the host repo. Pick one:

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
tools_dir: ../specs-tools         # or any absolute/relative path
```

Or keep the historical submodule layout if you want every contributor on a host project to see the content in-tree (e.g. the current `redmine-deployment` repo). Both submodule and plain-folder checkouts are auto-detected.

### Quick decision

| You are…                              | Use this    |
| -------------------------------------- | ----------- |
| writing specs in a host project        | **managed** |
| editing templates / process docs       | **dev**     |
| working air-gapped, no internet at all | **dev** with a vendored snapshot |

Switch modes any time by editing `.specs.yaml`; nothing else changes.

## Concepts

Three paths matter and are referenced throughout this README and in `specs doctor`:

- **specs root** — the directory that contains `.specs.yaml`, `model/`, and `change-requests/`. This is what `specs` operates on.
- **host root** — the git repository that contains the specs root. It can *be* the specs root (when the repo is dedicated to specs) or contain it as a subdirectory or submodule.
- **tools dir** — where the generic framework content is materialised. In **managed** mode this is the user cache dir; in **dev** mode it is whatever `tools_dir` points at.

`specs doctor` prints all three so you can verify what was detected.

## Commands

| Command                                                                                               | Purpose                                           |
| ----------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| `specs version` / `--version`                                                                         | print the installed binary version                |
| `specs doctor`                                                                                        | diagnose environment, layout, version drift       |
| `specs init [--with-vscode] [--force]`                                                                | configure an existing host (writes `.specs.yaml`) |
| `specs bootstrap [--at <path>] [--layout folder\|submodule] [--tools-mode submodule\|folder\|vendor]` | scaffold a new host                               |
| `specs lint [--all\|--links\|--style\|--baselines]`                                                   | run lint checks                                   |
| `specs tools update [--to <ref>]`                                                                     | update the `.specs-tools` content layer           |

All write commands accept `--dry-run` where applicable.

## `.specs.yaml`

Lives next to the specs root. Minimal example for **managed** mode (recommended default):

```yaml
tools_url: https://github.com/jdehaan/specs-tools.git
tools_ref: v1.0.0          # tag, branch, or commit SHA
min_specs_version: 0.1.0
repos:
  redmine: container/redmine/redmine
  application_packages: container/redmine/application_packages
```

For **dev** mode, drop `tools_url`/`tools_ref` and point at a checkout instead:

```yaml
tools_dir: ../specs-tools  # or .specs-tools (submodule/folder), or absolute path
min_specs_version: 0.1.0
repos:
  ...
```

Other optional knobs: `change_requests_dir`, `model_dir`, `baselines_file`, `markdownlint_config`, `templates_schema`. Defaults are sensible; only set them when overriding.

> Status: `tools_url` / `tools_ref` (managed mode) is on the roadmap; today the binary supports the dev-mode `tools_dir` (auto-detecting submodule, plain folder, or vendor snapshot). The two-mode story above is what the docs and `specs bootstrap` will converge on.

## Status

Phase 1 — lint, layout auto-detection, `init`/`bootstrap`/`tools update` against a checkout. **Phase 1.5** — implement managed mode (cache + auto-fetch). **Phase 2** — authoring commands (`scaffold`, `cr`, `link`, `baseline`, `vscode`).

## Development

```bash
go test ./...
go build ./...
go install ./cmd/specs
```

Cross-platform release builds are produced by GoReleaser on git tags (`v*.*.*`). See [`.goreleaser.yaml`](./.goreleaser.yaml).
