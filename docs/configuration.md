# `.specs.yaml` reference

Lives next to the specs root.

## Managed mode (recommended default)

```yaml
tools_url: https://github.com/Color-Of-Code/specs-framework.git
tools_ref: v1.0.0 # tag, branch, or commit SHA
min_specs_version: 0.1.0
repos:
  redmine: container/redmine/redmine
  application_packages: container/redmine/application_packages
```

## Dev mode

Drop `tools_url` / `tools_ref` and point at a checkout instead:

```yaml
tools_dir: ../specs-framework # or .specs-framework (submodule/folder), or absolute path
min_specs_version: 0.1.0
repos: ...
```

## Using a named framework

When a [framework registry](#framework-registry) is configured, the toolchain resolves `tools_url`/`tools_ref` (or `tools_dir`) from the named entry. The `.specs.yaml` still stores the resolved values — the registry is only consulted at `init`/`bootstrap` time.

## Optional knobs

| Key                   | Purpose                                       |
| --------------------- | --------------------------------------------- |
| `change_requests_dir` | Override the default `change-requests/` path. |
| `model_dir`           | Override the default `model/` path.           |
| `baselines_file`      | Override the baselines table location.        |
| `markdownlint_config` | Path to a custom markdownlint config.         |
| `templates_schema`    | Path to a custom templates schema.            |

Defaults are sensible; only set these when overriding.

---

## Framework registry

The framework registry is a **user-level** configuration file that maps short names to framework sources. It is read by `specs init`, `specs bootstrap`, and `specs framework` commands.

### Location

| Platform | Path                                                  |
| -------- | ----------------------------------------------------- |
| Linux    | `~/.config/specs/frameworks.yaml`                     |
| macOS    | `~/Library/Application Support/specs/frameworks.yaml` |
| Windows  | `%APPDATA%\specs\frameworks.yaml`                     |

### Schema

```yaml
# frameworks.yaml
frameworks:
  <name>:
    url: <git-url>        # mutually exclusive with `path`
    ref: <tag-or-branch>  # optional; defaults to "main"
    path: <local-dir>     # mutually exclusive with `url`
```

### Example

```yaml
frameworks:
  default:
    url: https://github.com/Color-Of-Code/specs-framework.git
    ref: v1.0.0
  acme:
    url: https://git.example.com/acme/specs-framework.git
    ref: main
  local:
    path: ~/src/specs-framework
```

### Resolution order

When `specs init` or `specs bootstrap` determine which framework to use:

1. Explicit `--tools-url` / `--tools-dir` flags (highest priority).
2. `--framework <name>` flag — looked up in the registry.
3. The `default` registry entry (if it exists and no flags override).
4. Hard-coded fallback: `https://github.com/Color-Of-Code/specs-framework.git@main`.
