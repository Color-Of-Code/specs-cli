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

## Optional knobs

| Key                   | Purpose                                       |
| --------------------- | --------------------------------------------- |
| `change_requests_dir` | Override the default `change-requests/` path. |
| `model_dir`           | Override the default `model/` path.           |
| `baselines_file`      | Override the baselines table location.        |
| `markdownlint_config` | Path to a custom markdownlint config.         |
| `templates_schema`    | Path to a custom templates schema.            |

Defaults are sensible; only set these when overriding.
