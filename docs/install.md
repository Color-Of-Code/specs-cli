# Installation

There are two supported ways to use the framework. They share the same `specs` binary and behave identically — pick whichever fits your workflow. You can switch later without changing anything in the host project.

## Option 1 — VS Code extension (recommended)

The extension bundles a matching `specs` binary, so installing it is the only step needed. Every command is reachable from the Command Palette and from the Specs view, and the extension auto-detects the specs root in the open workspace.

1. Download `specs-<your-platform>.vsix` from the latest [GitHub release](https://github.com/Color-Of-Code/specs-toolchain/releases).
2. Install it:

   ```bash
   code --install-extension specs-<your-platform>.vsix
   ```

3. Open a workspace that contains `.specs.yaml` (or run **Specs: Bootstrap** from the palette to create one).

The extension never requires a separately installed CLI. If you happen to have one on `PATH`, you can opt in with the `specs.useGlobalBinary` setting; see [extension/README.md](../extension/README.md) for the full settings reference.

## Option 2 — CLI only

Use this if you do not work in VS Code, automate things in CI, or prefer a terminal-only workflow.

```bash
go install github.com/Color-Of-Code/specs-toolchain/cli/cmd/specs@latest
```

`go install` puts the binary at `$(go env GOBIN)` if set, otherwise `$(go env GOPATH)/bin` (typically `~/go/bin` on Linux/macOS, `%USERPROFILE%\go\bin` on Windows). Make sure that directory is on `PATH`. Release tarballs from [GitHub Releases](https://github.com/Color-Of-Code/specs-toolchain/releases) work too — drop the `specs` binary anywhere on `PATH`.

Verify the install:

```bash
specs --version
specs doctor
```

## Combining both

Installing the extension and the CLI side by side is fine and common: the extension uses its bundled binary by default, and you still get `specs` on the terminal for ad-hoc commands and CI. Set `specs.useGlobalBinary = true` to make the extension prefer the CLI on `PATH` so both stay on the same version.
