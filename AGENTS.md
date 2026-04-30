# AGENTS.md

Guidance for AI coding agents and human contributors working on this repo.

## Guiding principle: describe the present

Everything in this repo — code, comments, docs, this file — describes the project **as it is now**. Do not record history, migrations, renames, deprecations, or "phase" status in source or docs.

- No "formerly X", "previously Y", "this used to be …", "renamed from …", "no longer supported".
- No phase trackers, roadmap status notes, or "shipped in X" annotations in user-facing docs.
- No commented-out code kept "for reference".
- Breaking changes belong in the **commit message** and (if relevant) the changelog/release notes — never in the code or docs themselves.

History lives in git. Read it there when needed.

## Repo at a glance

- `engine/` — Go module (`github.com/Color-Of-Code/specs-toolchain/engine`). Builds the `specs` binary. Subcommands live under `engine/cmd/specs/`; reusable logic under `engine/internal/<package>/`.
- `extension/` — VS Code extension (TypeScript, standalone pnpm project). Wraps the engine binary; all Node tooling (`pnpm-workspace.yaml`, `pnpm-lock.yaml`, `.npmrc`) lives here.
- `docs/` — User-facing docs (`install.md`, `concepts.md`, `commands.md`, `configuration.md`, `development.md`).
- `Makefile` — Top-level entry point; ties Go and pnpm builds together.
- `.github/workflows/` — `ci.yml` (vet/test/build/markdown) and `release.yml` (GoReleaser + per-platform `.vsix`).

## Wording conventions

- The binary is the **engine**. Use "specs engine" everywhere user-visible: docs, comments, help text, settings.
- VS Code setting key for an explicit binary path is `specs.enginePath`.
- The internal Go package is `engine/`; module path `specs-toolchain/engine`.

## Build, test, lint

Use `make` from the repo root for the common cases:

```bash
make build              # engine + extension
make build-engine       # engine binary -> ./specs
make build-extension    # extension TypeScript
make package-extension  # produce a .vsix
make deploy-dev         # build + symlink extension into ~/.vscode/extensions
make check              # format-check + lint + vet + test (what CI runs)
make format             # specs format (in-place)
make format-check       # specs format --check (CI gate)
make lint               # specs lint --style
make vet                # go vet ./... in engine
make test               # go test ./... in engine
```

Direct invocations:

- Engine tests: `cd engine && go test ./...`
- Engine vet: `cd engine && go vet ./...`
- Extension type-check + compile: `cd extension && pnpm run compile`

After any change, **always** run the relevant build/tests for the surfaces you touched. Both Go and TypeScript builds must stay clean.

## Working on the engine (Go)

- All flag parsing uses the standard `flag` package with one `flag.NewFlagSet` per subcommand.
- Errors with a non-1 exit code go through `exitWith(code, format, args...)` which returns an `*exitError` understood by `main.main`.
- Subcommands live in `engine/cmd/specs/<name>.go`; register them in the `commands` slice in `main.go`.
- Reusable logic must go under `engine/internal/<package>/` with unit tests in the same package.
- Config schema is in [engine/internal/config/config.go](engine/internal/config/config.go); add new keys to both `File` and `Resolved` and preserve them on round-trip in `cmdInit`.
- Framework registry: see [engine/internal/registry/registry.go](engine/internal/registry/registry.go); URL/path entries are mutually exclusive.
- Path-exclusion helpers (`isExcludedPath` in `format.go`, `isExcludedRel` in `internal/lint/lint.go`) match by path **component** at any depth, not by prefix. Add new exclusions to the component set.
- Goldmark inline AST nodes (`*ast.RawHTML`, etc.) panic on `Lines()`. Always walk up to the nearest block ancestor before reading line ranges — see `lineOf` in [engine/internal/lint/rules.go](engine/internal/lint/rules.go).
- To distinguish a user-supplied flag from its default value, use the `flagWasSet` helper rather than comparing against the default literal.
- Go module is pinned at the version in [engine/go.mod](engine/go.mod). Avoid bumping unless required.

## Working on the extension (TypeScript)

- Engine binary is resolved by [extension/src/engine.ts](extension/src/engine.ts) via `specs.enginePath` → `useGlobalBinary` → bundled binary → PATH.
- Palette commands are registered in [extension/src/commands.ts](extension/src/commands.ts); their declarations live in [extension/package.json](extension/package.json) under `contributes.commands`. Both must be kept in sync.
- When adding a wrapper, match the engine's exact flag and positional-argument signature. Validate input in `showInputBox` callbacks rather than letting the engine reject it.
- Status bar refresh is event-driven (FileSystemWatchers + window state), not polled — see [extension/src/statusBar.ts](extension/src/statusBar.ts).
- Strict TypeScript; no implicit `any`. Run `pnpm run compile` before committing.

## Adding a new engine subcommand

1. Add `engine/cmd/specs/<name>.go` with a `cmd<Name>(args []string) error` entry point.
2. Register it in the `commands` slice in [engine/cmd/specs/main.go](engine/cmd/specs/main.go).
3. If it should be reachable from VS Code, add a wrapper in [extension/src/commands.ts](extension/src/commands.ts) and a corresponding `contributes.commands` entry in [extension/package.json](extension/package.json).
4. Document it in [docs/commands.md](docs/commands.md) (and any related docs).
5. If it generates a `.vscode/tasks.json` entry, update the `vscodeTasksJSON` template in [engine/cmd/specs/init.go](engine/cmd/specs/init.go).

## Documentation

- All Markdown is formatted by `specs format` and linted by `specs lint --style` — both are run by CI. Keep `make format-check` green.
- Don't introduce Node-based markdown tooling; the engine is the only formatter/linter for Markdown in this repo.
- When editing user-facing behaviour, update the matching doc in `docs/` *and* the relevant section of [README.md](README.md) or [extension/README.md](extension/README.md).
- The lint rule `inline_html` is **on** by default. Bare `<word>` placeholders in templates and docs are parsed as inline HTML and rejected; write `Word` (or use a code span) instead.
- Markdown tables don't support raw `|` inside cells. For commands with `flag1|flag2` choices, prefer a bullet list or escape with `\|`; the formatter cannot recover a malformed table.

## Commits

- Use Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, `test:` with an optional scope (e.g. `fix(extension): ...`).
- Group logically: one change per commit; keep CLI changes, extension changes, and docs in separate commits when independent.
- Note breaking changes prominently in the commit body (Go install path, VS Code setting renames, etc.).
- Run the relevant build/tests **before** committing.

## Things to avoid

- Don't recreate a top-level `package.json`; the root is intentionally pnpm-free. The `Makefile` is the entry point.
- Don't add subcommands that are thin aliases for existing flags.
- Don't poll the engine on a timer from the extension — use event-driven refresh (`FileSystemWatcher`, `onDid*` events).
- Don't widen types in TypeScript with `Record<string, string>` or `any`; derive from existing types or define precise interfaces.
- Don't bypass `make format` / `make lint` with `--no-verify` or by editing `style.yaml` to silence a rule without justification.

## Quick references

- Engine entry point: [engine/cmd/specs/main.go](engine/cmd/specs/main.go)
- Config schema: [engine/internal/config/config.go](engine/internal/config/config.go)
- Framework registry: [engine/internal/registry/registry.go](engine/internal/registry/registry.go)
- Extension activation: [extension/src/extension.ts](extension/src/extension.ts)
- Engine resolver: [extension/src/engine.ts](extension/src/engine.ts)
- Palette commands: [extension/src/commands.ts](extension/src/commands.ts) ↔ [extension/package.json](extension/package.json)
