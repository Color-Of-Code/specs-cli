# specs-cli

User-scope CLI for the [specs framework](https://github.com/jdehaan/specs-tools). Replaces the bash hookup (`lint.sh`, `repo-map.sh`, manual `cp`/`ln -s`/`git mv` recipes) with a single cross-platform Go binary.

## Status

Phase 1 — lint parity, layout auto-detection, `init`/`bootstrap`/`tools update`. Authoring commands (`scaffold`, `cr`, `link`, `baseline`, `vscode`) land in Phase 2.

## Install

User-scope (one binary per developer, shared across all host projects):

```bash
go install github.com/jdehaan/specs-cli/cmd/specs@latest
# or grab a release binary from GitHub Releases and drop it into ~/.local/bin
```

Verify:

```bash
specs --version
specs doctor
```

## Layouts

`specs` auto-detects every supported combination of how `specs/` and `.specs-tools` are materialised in the host repo:

| `specs/` mode | `.specs-tools` mode | Notes |
|---|---|---|
| submodule of host | submodule of `specs/` | shared specs across hosts (current `redmine-deployment`) |
| submodule of host | plain folder | shared specs, vendored content |
| plain folder | submodule of host | private specs, version-pinned content (recommended greenfield default) |
| plain folder | plain folder | fully self-contained; no automatic content version pin |
| repo root (`--at .`) | any of the above | the host repo *is* the specs repo |

Detection drives `specs doctor` output and the behaviour of `specs tools update`.

## Commands

| Command | Purpose |
|---|---|
| `specs version` / `--version` | print the installed binary version |
| `specs doctor` | diagnose environment, layout, version drift |
| `specs init [--with-vscode] [--force]` | configure an existing host (writes `.specs.yaml`) |
| `specs bootstrap [--at <path>] [--layout folder\|submodule] [--tools-mode submodule\|folder\|vendor]` | scaffold a new host |
| `specs lint [--all\|--links\|--style\|--baselines]` | run lint checks (replaces `bash .lint/lint.sh`) |
| `specs tools update [--to <ref>]` | update the `.specs-tools` content layer |

All write commands accept `--dry-run` where applicable.

## `.specs.yaml`

Lives next to the specs root. Example:

```yaml
tools_dir: auto                  # or .specs-tools, /abs/path
min_specs_version: 0.1.0
templates_schema: 1              # optional, matched against tools-manifest.yaml
change_requests_dir: change-requests
model_dir: model
baselines_file: model/baselines/repo-baseline.md
markdownlint_config: ""          # default: <tools_dir>/lint/.markdownlint-cli2.jsonc
repos:
  redmine: container/redmine/redmine
  application_packages: container/redmine/application_packages
```

`tools_dir: auto` resolves to `<specs_root>/.specs-tools` first, then `<host_root>/.specs-tools`.

## Migrating from the bash hookup

For an existing host that uses `specs/.lint/lint.sh` + `repo-map.sh`:

1. Install the binary (`go install …` or release).
2. From the specs root, run `specs init`. Edit `.specs.yaml` to populate `repos:` from your old `repo-map.sh`.
3. Replace `bash specs/.lint/lint.sh` with `specs lint` in CI and docs.
4. Optionally keep `specs/.lint/lint.sh` as a one-line shim: `exec specs lint "$@"`.

## Development

```bash
go test ./...
go build ./...
go install ./cmd/specs
```

Cross-platform release builds are produced by GoReleaser on git tags (`v*.*.*`). See [`.goreleaser.yaml`](./.goreleaser.yaml).
