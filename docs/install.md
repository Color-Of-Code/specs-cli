# Installation

There are two supported ways to use the framework. They share the same `specs` binary and behave identically — pick whichever fits your workflow. You can switch later without changing anything in the host project.

## Option 1 — VS Code extension (recommended)

The extension bundles a matching `specs` binary, so installing it is the only step needed. Every command is reachable from the Command Palette and from the Specs view, and the extension auto-detects the specs root in the open workspace.

1. Download `specs-<your-platform>.vsix` from the latest [GitHub release](https://github.com/Color-Of-Code/specs-toolchain/releases).
2. Install it:

   ```bash
   code --install-extension specs-<your-platform>.vsix
   ```

3. Open a workspace that contains `.specs.yaml` (or run **Specs: Init host** from the palette to create one).

The extension never requires a separately installed engine. If you happen to have one on `PATH`, you can opt in with the `specs.useGlobalBinary` setting; see [extension/README.md](../extension/README.md) for the full settings reference.

## Option 2 — engine only

Use this if you do not work in VS Code, automate things in CI, or prefer a terminal-only workflow.

```bash
go install github.com/Color-Of-Code/specs-toolchain/engine/cmd/specs@latest
```

`go install` puts the binary at `$(go env GOBIN)` if set, otherwise `$(go env GOPATH)/bin` (typically `~/go/bin` on Linux/macOS, `%USERPROFILE%\go\bin` on Windows). Make sure that directory is on `PATH`. Release tarballs from [GitHub Releases](https://github.com/Color-Of-Code/specs-toolchain/releases) work too — drop the `specs` binary anywhere on `PATH`.

Verify the install:

```bash
specs --version
specs doctor
```

## Combining both

Installing the extension and the engine side by side is fine and common: the extension uses its bundled binary by default, and you still get `specs` on the terminal for ad-hoc commands and CI. Set `specs.useGlobalBinary = true` to make the extension prefer the engine on `PATH` so both stay on the same version.

---

## Advanced: starting with a custom framework

By default, `specs init` pulls the official [specs-framework](https://github.com/Color-Of-Code/specs-framework) content. If your organisation maintains its own framework, pass `--framework <source>` (registered name, git URL, or local path).

If you need to create a **brand-new framework from scratch** rather than forking an existing one:

```bash
specs framework seed --out /path/to/my-framework
cd /path/to/my-framework
git init && git add -A && git commit -m "initial skeleton"
# push to your remote of choice
```

Then register it for convenient reuse:

```bash
specs framework add my-org --url https://git.example.com/my-org/my-framework.git
```

See [concepts → Framework sources](concepts.md#framework-sources) and [commands → Framework management](commands.md#framework-management-commands) for full details.
