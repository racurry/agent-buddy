package mcpb

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvVar describes a user-configurable environment variable.
type EnvVar struct {
	Name        string
	Description string
	Required    bool
}

// BundleOptions configures the mcpb bundle generation.
type BundleOptions struct {
	PackageName string
	EnvVars     []EnvVar
	OutputPath  string // defaults to {package-name}.mcpb
}

// manifest mirrors the mcpb manifest.json structure.
type manifest struct {
	ManifestVersion string            `json:"manifest_version"`
	Name            string            `json:"name"`
	DisplayName     string            `json:"display_name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Author          manifestAuthor    `json:"author"`
	Server          manifestServer    `json:"server"`
	Compatibility   *compatibility    `json:"compatibility,omitempty"`
	UserConfig      map[string]ucField `json:"user_config,omitempty"`
}

type manifestAuthor struct {
	Name string `json:"name"`
}

type manifestServer struct {
	Type       string          `json:"type"`
	EntryPoint string          `json:"entry_point"`
	MCPConfig  mcpConfig       `json:"mcp_config"`
}

type mcpConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

type compatibility struct {
	Platforms []string           `json:"platforms,omitempty"`
	Runtimes  map[string]string  `json:"runtimes,omitempty"`
}

type ucField struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Required    bool   `json:"required,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
}

// Bundle creates a .mcpb file for a uvx-based MCP server.
func Bundle(info *PyPIInfo, opts BundleOptions) (string, error) {
	outPath := opts.OutputPath
	if outPath == "" {
		outDir := ".tmp"
		os.MkdirAll(outDir, 0o755)
		outPath = filepath.Join(outDir, opts.PackageName+".mcpb")
	}

	// Build manifest
	m := manifest{
		ManifestVersion: "0.4",
		Name:            info.Name,
		DisplayName:     displayName(info.Name),
		Version:         info.Version,
		Description:     info.Summary,
		Author:          manifestAuthor{Name: info.Author},
		Server: manifestServer{
			Type:       "uv",
			EntryPoint: "src/server.py",
			MCPConfig: mcpConfig{
				Command: "uv",
				Args:    []string{"run", "src/server.py"},
			},
		},
		Compatibility: &compatibility{
			Platforms: []string{"darwin", "linux", "win32"},
			Runtimes:  map[string]string{"python": info.PythonRequires},
		},
	}

	if len(opts.EnvVars) > 0 {
		m.UserConfig = make(map[string]ucField)
		m.Server.MCPConfig.Env = make(map[string]string)
		for _, ev := range opts.EnvVars {
			sensitive := isSensitive(ev.Name)
			m.UserConfig[ev.Name] = ucField{
				Type:        "string",
				Title:       envVarTitle(ev.Name),
				Description: ev.Description,
				Required:    ev.Required,
				Sensitive:   sensitive,
			}
			// Map user_config values to env vars via template substitution
			m.Server.MCPConfig.Env[ev.Name] = fmt.Sprintf("${user_config.%s}", ev.Name)
		}
	}

	manifestJSON, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling manifest: %w", err)
	}

	// Build pyproject.toml
	// Use hatchling with explicit config to avoid editable-build discovery issues.
	// The wrapper has no real package — just a script that imports the dependency.
	pyproject := fmt.Sprintf(`[project]
name = "%s-mcpb"
version = "%s"
description = "MCPB wrapper for %s"
requires-python = "%s"
dependencies = [
    "%s>=%s",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.hatch.build.targets.wheel]
packages = ["src"]
`, info.Name, info.Version, info.Name, info.PythonRequires, info.Name, info.Version)

	// Build server.py shim
	serverPy := generateShim(info, opts.EnvVars)

	// Create the ZIP
	outAbs, err := filepath.Abs(outPath)
	if err != nil {
		return "", err
	}

	f, err := os.Create(outAbs)
	if err != nil {
		return "", fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	files := map[string][]byte{
		"manifest.json":  manifestJSON,
		"pyproject.toml": []byte(pyproject),
		"src/server.py":  []byte(serverPy),
	}

	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			return "", fmt.Errorf("adding %s to zip: %w", name, err)
		}
		if _, err := fw.Write(content); err != nil {
			return "", fmt.Errorf("writing %s: %w", name, err)
		}
	}

	return outAbs, nil
}

func generateShim(info *PyPIInfo, envVars []EnvVar) string {
	var sb strings.Builder

	sb.WriteString("import importlib.metadata\n\n")

	// Discover and call the console_scripts entry point at runtime.
	// This works regardless of the package's internal module structure.
	// Env vars are handled by mcp_config.env in manifest.json.
	sb.WriteString(fmt.Sprintf(`ep = importlib.metadata.entry_points(group="console_scripts", name="%s")
entry = list(ep)[0]
fn = entry.load()
fn()
`, info.Name))

	return sb.String()
}

func displayName(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// envVarTitle converts "OBSIDIAN_API_KEY" to "Obsidian Api Key".
func envVarTitle(name string) string {
	parts := strings.Split(strings.ToLower(name), "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func isSensitive(name string) bool {
	lower := strings.ToLower(name)
	for _, keyword := range []string{"key", "token", "secret", "password", "credential"} {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}
