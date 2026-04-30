# specs-toolchain

Tooling for the [specs framework](https://github.com/Color-Of-Code/specs-framework): lint, scaffolding, change-request lifecycle, traceability links, and baseline updates for any host project that uses the framework.

It ships in two flavours that share the same engine:

1. **VS Code extension (recommended)** — bundles the `specs` binary and exposes every command from the palette, the Specs view, and tasks. Nothing extra to install.
2. **Engine only** — a single cross-platform Go binary for terminal users and CI.

You can use either one alone, or install both side by side.

## Quick start — VS Code extension

1. Download `specs-<your-platform>.vsix` from the latest [GitHub release](https://github.com/Color-Of-Code/specs-toolchain/releases).
2. Install it:

   ```bash
   code --install-extension specs-<your-platform>.vsix
   ```

3. Open a workspace and run **Specs: Bootstrap** (new project) or **Specs: Doctor** (existing project) from the Command Palette.

The extension uses its bundled binary by default, so no separate engine install is required. See [extension/README.md](extension/README.md) for the settings reference.

## Quick start — engine only

```bash
go install github.com/Color-Of-Code/specs-toolchain/engine/cmd/specs@latest
specs --version
specs doctor
```

Release tarballs from [GitHub Releases](https://github.com/Color-Of-Code/specs-toolchain/releases) work too — drop the `specs` binary anywhere on `PATH`.

Full installation notes (extension settings, combining both, platform paths) live in [docs/install.md](docs/install.md).

## Documentation

| Topic                                  | What it covers                                                                          |
| -------------------------------------- | --------------------------------------------------------------------------------------- |
| [Installation](docs/install.md)        | Extension and engine install paths, side-by-side usage.                                 |
| [Concepts](docs/concepts.md)           | specs root vs. host repo vs. tools dir; `managed` vs. `dev` mode for framework content. |
| [Commands](docs/commands.md)           | Reference for every `specs` subcommand (also reachable as **Specs: …** in VS Code).     |
| [Configuration](docs/configuration.md) | `.specs.yaml` keys, defaults, and overrides.                                            |
| [Development](docs/development.md)     | Building the engine and extension, release process.                                     |
| [Extension](extension/README.md)       | VS Code-specific settings, packaging notes, platform matrix.                            |
