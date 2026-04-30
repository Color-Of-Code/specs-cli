package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Color-Of-Code/specs-toolchain/cli/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/cli/internal/toolsmanifest"
)

// doctorJSON is the stable schema consumed by the VS Code extension and
// other tooling. New fields may be added; existing fields are not
// removed or repurposed.
type doctorJSON struct {
	Version           string            `json:"version"`
	GOOS              string            `json:"goos"`
	GOARCH            string            `json:"goarch"`
	ConfigPath        string            `json:"config_path"`
	SpecsRoot         string            `json:"specs_root"`
	HostRoot          string            `json:"host_root"`
	SpecsMode         string            `json:"specs_mode"`
	ToolsDir          string            `json:"tools_dir"`
	ToolsMode         string            `json:"tools_mode"`
	ToolsURL          string            `json:"tools_url,omitempty"`
	ToolsRef          string            `json:"tools_ref,omitempty"`
	ToolsRev          string            `json:"tools_rev,omitempty"`
	ModelDir          string            `json:"model_dir"`
	ChangeRequestsDir string            `json:"change_requests_dir"`
	BaselinesFile     string            `json:"baselines_file"`
	MarkdownlintCfg   string            `json:"markdownlint_config"`
	MinSpecsVersion   string            `json:"min_specs_version,omitempty"`
	TemplatesSchema   int               `json:"templates_schema,omitempty"`
	Manifest          *manifestJSON     `json:"tools_manifest,omitempty"`
	Repos             map[string]string `json:"repos"`
	Compatible        bool              `json:"compatible"`
	CompatibleMessage string            `json:"compatible_message,omitempty"`
}

type manifestJSON struct {
	TemplatesSchema int    `json:"templates_schema"`
	Version         string `json:"version,omitempty"`
}

func cmdDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "emit machine-readable JSON to stdout (no human prose, no external-tool checks)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs doctor [--json]")
		fmt.Fprintln(os.Stderr, "Diagnose specs CLI environment, layout, and version drift.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	if *jsonOut {
		return emitDoctorJSON(cfg)
	}

	fmt.Printf("specs CLI:        %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
	if cfg.ConfigPath != "" {
		fmt.Printf("config file:      %s\n", cfg.ConfigPath)
	} else {
		fmt.Println("config file:      <not found> (using defaults; run `specs init` to write .specs.yaml)")
	}
	fmt.Printf("specs root:       %s\n", cfg.SpecsRoot)
	fmt.Printf("host root:        %s\n", cfg.HostRoot)
	fmt.Printf("specs mode:       %s\n", cfg.SpecsMode)
	if cfg.ToolsDir != "" {
		fmt.Printf("tools dir:        %s%s\n", cfg.ToolsDir, existsSuffix(cfg.ToolsDir))
		fmt.Printf("tools mode:       %s\n", cfg.ToolsMode)
		if cfg.ToolsMode == config.ToolsModeManaged {
			fmt.Printf("tools url:        %s\n", cfg.ToolsURL)
			ref := cfg.ToolsRef
			if ref == "" {
				ref = "(unset; defaults to main on next fetch)"
			}
			fmt.Printf("tools ref:        %s\n", ref)
		}
		if rev := gitShortRev(cfg.ToolsDir); rev != "" {
			fmt.Printf("tools rev:        %s\n", rev)
		}
	} else {
		fmt.Println("tools dir:        <missing> (run `specs bootstrap` or set tools_url/tools_dir)")
	}
	fmt.Printf("model dir:        %s\n", cfg.ModelDir)
	fmt.Printf("change-requests:  %s\n", cfg.ChangeRequestsDir)
	fmt.Printf("baselines file:   %s%s\n", cfg.BaselinesFile, existsSuffix(cfg.BaselinesFile))
	fmt.Printf("style config:     %s%s\n", cfg.StyleConfig, existsSuffix(cfg.StyleConfig))
	if cfg.MinSpecsVersion != "" {
		fmt.Printf("min_specs_version: %s\n", cfg.MinSpecsVersion)
	}
	if cfg.TemplatesSchema != 0 {
		fmt.Printf("templates_schema: %d (host requires)\n", cfg.TemplatesSchema)
	}
	if cfg.ToolsDir != "" {
		if m, err := toolsmanifest.Load(cfg.ToolsDir); err != nil {
			fmt.Printf("tools manifest:   error: %v\n", err)
		} else if m == nil {
			fmt.Println("tools manifest:   <not present>")
		} else {
			fmt.Printf("tools manifest:   templates_schema=%d version=%s\n", m.TemplatesSchema, m.Version)
			if ok, msg := toolsmanifest.Compatible(cfg.TemplatesSchema, m); !ok {
				return exitWith(1, "%s", msg)
			}
		}
	}
	fmt.Printf("repos configured: %d\n", len(cfg.Repos))

	fmt.Println("")
	fmt.Println("External tools:")
	reportTool("git", true)
	reportTool("markdownlint-cli2", false)
	reportTool("pnpm", false)
	reportTool("dot", false)

	if cfg.MinSpecsVersion != "" && Version != "dev" && Version < cfg.MinSpecsVersion {
		return exitWith(1, "installed CLI %s is older than min_specs_version %s", Version, cfg.MinSpecsVersion)
	}
	return nil
}

func existsSuffix(p string) string {
	if p == "" {
		return ""
	}
	if _, err := os.Stat(p); err != nil {
		return "  (missing)"
	}
	return ""
}

func reportTool(name string, required bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		mark := "optional"
		if required {
			mark = "REQUIRED"
		}
		fmt.Printf("  %-20s not found  [%s]\n", name, mark)
		return
	}
	fmt.Printf("  %-20s %s\n", name, path)
}

func gitShortRev(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(bytesTrim(out, "\n\r "))
}

// bytesTrim trims any of the bytes in cutset from both ends of b.
func bytesTrim(b []byte, cutset string) []byte {
	i, j := 0, len(b)
	for i < j && containsByte(cutset, b[i]) {
		i++
	}
	for j > i && containsByte(cutset, b[j-1]) {
		j--
	}
	return b[i:j]
}

func containsByte(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

// joinPath is a small helper used by other commands.
func joinPath(parts ...string) string { return filepath.Join(parts...) }

func emitDoctorJSON(cfg *config.Resolved) error {
	repos := map[string]string{}
	for name, p := range cfg.Repos {
		repos[name] = p
	}
	d := doctorJSON{
		Version:           Version,
		GOOS:              runtime.GOOS,
		GOARCH:            runtime.GOARCH,
		ConfigPath:        cfg.ConfigPath,
		SpecsRoot:         cfg.SpecsRoot,
		HostRoot:          cfg.HostRoot,
		SpecsMode:         string(cfg.SpecsMode),
		ToolsDir:          cfg.ToolsDir,
		ToolsMode:         string(cfg.ToolsMode),
		ToolsURL:          cfg.ToolsURL,
		ToolsRef:          cfg.ToolsRef,
		ModelDir:          cfg.ModelDir,
		ChangeRequestsDir: cfg.ChangeRequestsDir,
		BaselinesFile:     cfg.BaselinesFile,
		MarkdownlintCfg:   cfg.StyleConfig,
		MinSpecsVersion:   cfg.MinSpecsVersion,
		TemplatesSchema:   cfg.TemplatesSchema,
		Repos:             repos,
		Compatible:        true,
	}
	if cfg.ToolsDir != "" {
		if rev := gitShortRev(cfg.ToolsDir); rev != "" {
			d.ToolsRev = rev
		}
		if m, err := toolsmanifest.Load(cfg.ToolsDir); err == nil && m != nil {
			d.Manifest = &manifestJSON{TemplatesSchema: m.TemplatesSchema, Version: m.Version}
		}
		if ok, msg := toolsmanifest.Compatible(cfg.TemplatesSchema, d.Manifest.toToolsManifest()); !ok {
			d.Compatible = false
			d.CompatibleMessage = msg
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(d)
}

func (m *manifestJSON) toToolsManifest() *toolsmanifest.Manifest {
	if m == nil {
		return nil
	}
	return &toolsmanifest.Manifest{TemplatesSchema: m.TemplatesSchema, Version: m.Version}
}
