package main

import (
	"os"
	"path/filepath"
)

const vscodeTasksJSON = `{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Specs: Lint (all)",
      "type": "shell",
      "command": "specs",
      "args": ["lint", "--all"],
      "problemMatcher": []
    },
    {
      "label": "Specs: Lint (links)",
      "type": "shell",
      "command": "specs",
      "args": ["lint", "--links"],
      "problemMatcher": []
    },
    {
      "label": "Specs: Lint (style)",
      "type": "shell",
      "command": "specs",
      "args": ["lint", "--style"],
      "problemMatcher": []
    },
    {
      "label": "Specs: Lint (baselines)",
      "type": "shell",
      "command": "specs",
      "args": ["lint", "--baselines"],
      "problemMatcher": []
    },
    {
      "label": "Specs: Doctor",
      "type": "shell",
      "command": "specs",
      "args": ["doctor"],
      "problemMatcher": []
    },
    {
      "label": "Specs: Framework Update",
      "type": "shell",
      "command": "specs",
      "args": ["framework", "update"],
      "problemMatcher": []
    },
    {
      "label": "Specs: Scaffold",
      "type": "shell",
      "command": "specs",
      "args": [
        "scaffold",
        "${input:specsKind}",
        "--title", "${input:specsTitle}",
        "${input:specsPath}"
      ],
      "problemMatcher": []
    },
    {
      "label": "Specs: CR New",
      "type": "shell",
      "command": "specs",
      "args": [
        "cr", "new",
        "--id", "${input:specsCRId}",
        "--slug", "${input:specsCRSlug}",
        "--title", "${input:specsTitle}"
      ],
      "problemMatcher": []
    },
    {
      "label": "Specs: CR Status",
      "type": "shell",
      "command": "specs",
      "args": ["cr", "status"],
      "problemMatcher": []
    },
    {
      "label": "Specs: CR Drain",
      "type": "shell",
      "command": "specs",
      "args": [
        "cr", "drain",
        "--id", "${input:specsCRId}"
      ],
      "problemMatcher": []
    },
    {
      "label": "Specs: Baseline Update (dry-run)",
      "type": "shell",
      "command": "specs",
      "args": ["baseline", "update", "--dry-run"],
      "problemMatcher": []
    },
    {
      "label": "Specs: Baseline Update",
      "type": "shell",
      "command": "specs",
      "args": ["baseline", "update"],
      "problemMatcher": []
    }
  ],
  "inputs": [
    {
      "id": "specsKind",
      "description": "Template kind",
      "type": "pickString",
      "options": ["requirement", "feature", "component", "api", "service"],
      "default": "requirement"
    },
    {
      "id": "specsPath",
      "description": "Path under model/<area>/ (e.g. security/099-mfa)",
      "type": "promptString"
    },
    {
      "id": "specsTitle",
      "description": "Human-readable title",
      "type": "promptString"
    },
    {
      "id": "specsCRId",
      "description": "CR id (e.g. 004)",
      "type": "promptString"
    },
    {
      "id": "specsCRSlug",
      "description": "CR slug (kebab-case)",
      "type": "promptString"
    }
  ]
}
`

// writeVSCodeTasks writes tasks.json (or tasks.specs.json if one exists).
func writeVSCodeTasks(hostRoot string) error {
	return writeVSCodeTasksAt(hostRoot, false)
}

// writeVSCodeTasksAt writes the tasks file. When force is true, an existing
// tasks.json is overwritten; otherwise the tasks land in tasks.specs.json.
func writeVSCodeTasksAt(hostRoot string, force bool) error {
	dir := filepath.Join(hostRoot, ".vscode")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	target := filepath.Join(dir, "tasks.json")
	if _, err := os.Stat(target); err == nil && !force {
		target = filepath.Join(dir, "tasks.specs.json")
	}
	return os.WriteFile(target, []byte(vscodeTasksJSON), 0o644)
}
