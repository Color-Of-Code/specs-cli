# `.specs.yaml` reference

Lives next to the specs root.

## Managed mode (recommended default)

```yaml
framework_url: https://github.com/Color-Of-Code/specs-framework.git
framework_ref: v1.0.0 # tag, branch, or commit SHA
min_specs_version: 0.1.0
repos:
  redmine: container/redmine/redmine
  application_packages: container/redmine/application_packages
```

## Dev mode

Drop `framework_url` / `framework_ref` and point at a checkout instead:

```yaml
framework_dir: ../specs-framework # or .specs-framework (submodule/folder), or absolute path
min_specs_version: 0.1.0
repos: ...
```

## Using a named framework

When a [framework registry](#framework-registry) is configured, the toolchain resolves `framework_url`/`framework_ref` (or `framework_dir`) from the named entry. The `.specs.yaml` still stores the resolved values — the registry is only consulted at `init` time.

## Optional knobs

| Key                   | Purpose                                                 |
| --------------------- | ------------------------------------------------------- |
| `change_requests_dir` | Override the default `change-requests/` path.           |
| `model_dir`           | Override the default `model/` path.                     |
| `baselines_file`      | Override the baselines table location.                  |
| `style_config`        | Path to a custom `style.yaml` for markdown style rules. |
| `templates_schema`    | Integer schema version expected by the engine.          |

Defaults are sensible; only set these when overriding.

### `style.yaml` schema

The `style.yaml` file controls markdown style checking with abstract, tool-independent rule names. If no `style_config` is set, the toolchain looks for `<framework_dir>/lint/style.yaml`, falling back to compiled-in defaults.

```yaml
rules:
  line_length: false # false (disabled) or integer max
  inline_html: false # allow inline HTML
  first_heading_h1: false # require first line to be h1
  heading_style: atx # "atx" (# Heading) or "setext"
  blank_lines_around_headings: true
  blank_lines_around_fences: true
  list_marker: dash # "dash" (-) or "asterisk" (*)
  no_trailing_whitespace: true
  no_consecutive_blank_lines: true
  fenced_code_language: true # require language on fenced code blocks
```

Only include keys you want to override — omitted keys use the compiled-in defaults.

---

## Framework registry

The framework registry is a **user-level** configuration file that maps short names to framework sources. It is read by `specs init` and `specs framework` commands.

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

When `specs init` determines which framework to use:

1. Explicit `--framework <name>[@ref]` value (highest priority): looked up in the registry; an `@ref` suffix overrides the registered ref for URL-based entries.
2. The registry's `default` entry (if it exists and `--framework` is omitted).
3. Hard-coded fallback: `https://github.com/Color-Of-Code/specs-framework.git@main`.
