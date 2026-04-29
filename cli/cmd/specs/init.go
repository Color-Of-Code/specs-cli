package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Color-Of-Code/specs-cli/cli/internal/config"
)

// cmdInit configures an existing host: writes .specs.yaml and (optionally)
// .vscode/tasks.json. It auto-detects the current specs and tools layout
// rather than prescribing one.
func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite an existing .specs.yaml")
	withVSCode := fs.Bool("with-vscode", false, "also write .vscode/tasks.json")
	at := fs.String("at", "", "path to specs root (default: auto-detect from CWD)")
	toolsURL := fs.String("tools-url", "", "set tools_url (managed mode); leave empty to auto-detect a checkout via tools_dir")
	toolsRef := fs.String("tools-ref", "", "set tools_ref alongside --tools-url")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs init [--force] [--with-vscode] [--at <path>] [--tools-url URL --tools-ref REF]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*at)
	if err != nil {
		return err
	}
	specsRoot := cfg.SpecsRoot
	cfgPath := filepath.Join(specsRoot, config.FileName)

	if _, err := os.Stat(cfgPath); err == nil && !*force {
		return exitWith(1, "%s already exists (use --force to overwrite)", cfgPath)
	}

	f := &config.File{
		MinSpecsVersion: Version,
	}
	if *toolsURL != "" {
		f.ToolsURL = *toolsURL
		f.ToolsRef = *toolsRef
	} else {
		f.ToolsDir = "auto"
	}
	// Preserve any existing repos map / overrides if .specs.yaml already exists.
	if cfg.Source != nil {
		if cfg.Source.Repos != nil {
			f.Repos = cfg.Source.Repos
		}
		if cfg.Source.ChangeRequestsDir != "" {
			f.ChangeRequestsDir = cfg.Source.ChangeRequestsDir
		}
		if cfg.Source.ModelDir != "" {
			f.ModelDir = cfg.Source.ModelDir
		}
		if cfg.Source.BaselinesFile != "" {
			f.BaselinesFile = cfg.Source.BaselinesFile
		}
		// If we weren't told to set tools_url, preserve any pre-existing
		// managed-mode pin from the existing file.
		if *toolsURL == "" && cfg.Source.ToolsURL != "" {
			f.ToolsURL = cfg.Source.ToolsURL
			f.ToolsRef = cfg.Source.ToolsRef
			f.ToolsDir = ""
		}
	}
	if f.Repos == nil {
		// Pre-seed an empty map with a comment-friendly placeholder so the
		// host can fill it in. We write an actual empty map; users edit.
		f.Repos = map[string]string{}
	}

	if err := config.Save(cfgPath, f); err != nil {
		return err
	}
	fmt.Printf("wrote %s (specs_mode=%s, tools_mode=%s)\n", cfgPath, cfg.SpecsMode, cfg.ToolsMode)

	if *withVSCode {
		if err := writeVSCodeTasks(cfg.HostRoot); err != nil {
			return err
		}
		fmt.Println("wrote .vscode/tasks.json")
	}
	return nil
}

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
      "label": "Specs: Tools Update",
      "type": "shell",
      "command": "specs",
      "args": ["tools", "update"],
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
      "label": "Specs: Baseline Check",
      "type": "shell",
      "command": "specs",
      "args": ["baseline", "check"],
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
		// Don't clobber existing tasks; write next to it.
		target = filepath.Join(dir, "tasks.specs.json")
	}
	return os.WriteFile(target, []byte(vscodeTasksJSON), 0o644)
}
