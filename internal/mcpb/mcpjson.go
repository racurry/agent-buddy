package mcpb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// FlexArgs handles args as either a string or []string in JSON.
type FlexArgs []string

func (f *FlexArgs) UnmarshalJSON(data []byte) error {
	// Try array first
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*f = arr
		return nil
	}
	// Fall back to single string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*f = []string{s}
	return nil
}

// MCPServer represents a single server entry from a .mcp.json file.
type MCPServer struct {
	Command string            `json:"command"`
	Args    FlexArgs          `json:"args"`
	Env     map[string]string `json:"env"`
}

// MCPConfig represents the top-level .mcp.json structure.
type MCPConfig struct {
	Servers map[string]MCPServer `json:"mcpServers"`
}

// ParseMCPJSON reads a .mcp.json file and returns the parsed config.
func ParseMCPJSON(path string) (*MCPConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg MCPConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if len(cfg.Servers) == 0 {
		return nil, fmt.Errorf("no mcpServers found in %s", path)
	}

	return &cfg, nil
}

// ResolveServer picks a server from the config. If serverName is empty and
// there's exactly one server, it uses that. Otherwise it requires a match.
func ResolveServer(cfg *MCPConfig, serverName string) (string, *MCPServer, error) {
	if serverName != "" {
		srv, ok := cfg.Servers[serverName]
		if !ok {
			names := serverNames(cfg)
			return "", nil, fmt.Errorf("server %q not found in .mcp.json (available: %s)", serverName, strings.Join(names, ", "))
		}
		return serverName, &srv, nil
	}

	if len(cfg.Servers) == 1 {
		for name, srv := range cfg.Servers {
			return name, &srv, nil
		}
	}

	names := serverNames(cfg)
	return "", nil, fmt.Errorf("multiple servers in .mcp.json, specify one as argument: %s", strings.Join(names, ", "))
}

// PackageFromServer extracts the PyPI package name from a uvx server entry.
func PackageFromServer(srv *MCPServer) (string, error) {
	cmd := strings.ToLower(srv.Command)
	if cmd == "uvx" {
		if len(srv.Args) == 0 {
			return "", fmt.Errorf("uvx server has no args (expected package name)")
		}
		return srv.Args[0], nil
	}
	if cmd == "uv" {
		// Pattern: uv --directory <dir> run <package>
		for i, arg := range srv.Args {
			if arg == "run" && i+1 < len(srv.Args) {
				return srv.Args[i+1], nil
			}
		}
		return "", fmt.Errorf("could not find package name in uv args: %v", srv.Args)
	}
	return "", fmt.Errorf("unsupported command %q (expected uvx or uv)", srv.Command)
}

// EnvVarsFromServer extracts env var names from a server entry.
func EnvVarsFromServer(srv *MCPServer) []EnvVar {
	var vars []EnvVar
	for name := range srv.Env {
		vars = append(vars, EnvVar{
			Name:        name,
			Description: name,
			Required:    true,
		})
	}
	return vars
}

func serverNames(cfg *MCPConfig) []string {
	var names []string
	for name := range cfg.Servers {
		names = append(names, name)
	}
	return names
}
