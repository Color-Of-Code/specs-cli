# Development

The repository is split in two halves with separate toolchains:

- `cli/` — the Go CLI (the engine).
- `extension/` — the VS Code extension that wraps the CLI. It is a standalone pnpm project; all Node tooling lives there.

A small top-level [`Makefile`](../Makefile) ties them together for the common cases (build, lint, format, package, deploy-dev).

## Quick reference

```bash
make build              # build CLI + extension
make build-cli          # CLI binary only -> ./specs
make build-extension    # compile extension TypeScript
make package-extension  # produce a .vsix
make deploy-dev         # build + symlink extension into ~/.vscode/extensions
make check              # format-check + lint
make format             # format all .md files in place
make format-check       # exit 1 if any .md needs formatting
make lint               # specs lint --style
```

## CLI

```bash
cd cli
go test ./...
go build ./...
go install ./cmd/specs    # installs into $(go env GOBIN)
```

Or run without installing:

```bash
make build-cli            # produces ./specs at the repo root
```

## VS Code extension

The extension is its own pnpm project (`pnpm-workspace.yaml`, `pnpm-lock.yaml`, and `.npmrc` all live under `extension/`). It pins pnpm `10.33.2`, uses Node `24.15.0`, and enforces a 7-day minimum release age for new dependencies.

```bash
cd extension
pnpm install
pnpm run compile          # one-off TypeScript build
pnpm run watch            # incremental
```

See [extension/README.md](../extension/README.md) for the extension-specific settings, packaging notes, and platform matrix.

## Markdown lint & format

Repo-level docs are checked and formatted entirely by the Go CLI — no Node.js tools required. Style defaults are compiled into the binary ([`cli/internal/lint/style_defaults.yaml`](../cli/internal/lint/style_defaults.yaml)).

```bash
specs format             # format all .md files in place
specs format --check     # check formatting (exit 1 if changes needed)
specs lint --style       # run style lint rules
```

Or via the Makefile:

```bash
make format        # specs format
make format-check  # specs format --check
make lint          # specs lint --style
make check         # format-check + lint (what CI runs)
```

The `markdown` job in [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) runs the same checks on every push and pull request.

## Releases

Cross-platform release builds are produced by GoReleaser on git tags (`v*.*.*`). See [`cli/.goreleaser.yaml`](../cli/.goreleaser.yaml). Per-platform `.vsix` artifacts are produced by [`extension/scripts/build-extension.ts`](../extension/scripts/build-extension.ts), which stages the matching CLI binary into the extension before packaging:

```bash
cd extension
pnpm run package:bundled -- <target>
# targets: linux-x64 linux-arm64 darwin-x64 darwin-arm64 win32-x64
```

`make package-extension` runs the same script for the host platform.

## Status

Phase 1 — lint, layout auto-detection, `init` / `bootstrap` / `tools update`, **managed mode** (cache + auto-fetch).
**Phase 2** — `scaffold`, `cr {new,status,drain}`, `baseline update`, `link check`, `vscode init` shipped.
**Phase 3** — `visualize traceability` (DOT and Mermaid), `templates_schema` enforcement, `--layout submodule` shipped.
**Phase 4** — framework registry (`specs framework list/add/remove`, `--framework <name>` on `init`/`bootstrap`).
**VS Code extension** under `extension/` (see [extension/README.md](../extension/README.md)).

## Developing the VS Code Extension Locally

To incrementally test the extension without reinstalling on every iteration, use `make deploy-dev`. It builds and symlinks the extension folder into `~/.vscode/extensions`, so changes are picked up on the next window reload.

### Steps

1. **Build and symlink**

   ```bash
   make deploy-dev
   ```

   This will:
   - Build the `specs` CLI binary into `./specs` and (via the script) `extension/bin/`.
   - Compile the TypeScript extension source.
   - Symlink the `extension` folder into `~/.vscode/extensions/Color-Of-Code.specs`.

2. **Reload the VS Code window**

   After running the command, reload the VS Code window (Command Palette → **Developer: Reload Window**) to apply the changes.

3. **Iterate**

   Re-run `make build-extension` (or `cd extension && pnpm run watch` for live recompilation) and reload the window to pick up further edits.

If you previously installed a `.vsix` of this extension, uninstall it first so the symlinked one takes precedence.
