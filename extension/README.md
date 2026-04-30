# Specs (VS Code extension)

VS Code support for the [specs-toolchain](https://github.com/Color-Of-Code/specs-toolchain)
framework. The extension bundles a matching `specs` binary; once
installed, every Specs command is reachable from the palette without
any separate engine installation.

This package is part of the [specs-toolchain](https://github.com/Color-Of-Code/specs-toolchain)
monorepo. Releases ship a per-platform `.vsix` attached to each
[GitHub release](https://github.com/Color-Of-Code/specs-toolchain/releases).

## Install

1. Download `specs-<your-platform>.vsix` from the latest release.
2. Install it:

   ```bash
   code --install-extension specs-<your-platform>.vsix
   ```

## Settings

| Setting                 | Default | Purpose                                                       |
| ----------------------- | ------- | ------------------------------------------------------------- |
| `specs.useGlobalBinary` | `false` | Prefer `specs` on `PATH` over the bundled binary.             |
| `specs.enginePath`      | `""`    | Explicit path to a specs engine binary. Overrides bundled and PATH. |
| `specs.toolsAutoUpdate` | `false` | Run `specs tools update` on activation.                       |
