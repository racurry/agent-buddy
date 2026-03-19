package cmd

import (
	"fmt"
	"strings"

	"github.com/agenthubdev/agent-buddy/internal/mcpb"
	"github.com/spf13/cobra"
)

var mcpbEnvVars []string
var mcpbOutput string
var mcpbFrom string

var mcpbCmd = &cobra.Command{
	Use:   "mcpb [pypi-package]",
	Short: "Create a .mcpb bundle from a uvx-based MCP server",
	Long: `Create a .mcpb (MCP Bundle) file for Claude Desktop from a uvx-based
MCP server published on PyPI.

The bundle is a thin wrapper that depends on the PyPI package — no source
code is cloned or bundled. Claude Desktop installs the dependency via UV
at runtime.

Use --from to read a .mcp.json file directly. The package name and env
vars are extracted automatically. If the file has multiple servers, all
are bundled unless you specify one by name.

Environment variables can also be declared manually with --env flags.
These become user_config fields in the manifest, so Claude Desktop shows
a configuration UI for them.

Examples:
  agent-buddy mcpb --from .mcp.json
  agent-buddy mcpb --from .mcp.json mcp-obsidian
  agent-buddy mcpb mcp-obsidian --env OBSIDIAN_API_KEY --env OBSIDIAN_HOST
  agent-buddy mcpb mcp-obsidian -o my-obsidian.mcpb`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if mcpbFrom != "" {
			return runFromMCPJSON(args)
		}

		if len(args) == 0 {
			return fmt.Errorf("provide a PyPI package name or use --from with a .mcp.json file")
		}
		return bundlePackage(args[0], extraEnvVars(), mcpbOutput)
	},
}

func runFromMCPJSON(args []string) error {
	cfg, err := mcpb.ParseMCPJSON(mcpbFrom)
	if err != nil {
		return err
	}

	// If a specific server was requested, just do that one
	if len(args) == 1 {
		name, srv, err := mcpb.ResolveServer(cfg, args[0])
		if err != nil {
			return err
		}
		return bundleFromServer(name, srv, mcpbOutput)
	}

	// Otherwise, bundle all uvx servers in the file
	var created int
	for name, srv := range cfg.Servers {
		srv := srv
		pkg, err := mcpb.PackageFromServer(&srv)
		if err != nil {
			fmt.Printf("Skipping %s: %v\n", name, err)
			continue
		}
		_ = pkg
		if err := bundleFromServer(name, &srv, ""); err != nil {
			fmt.Printf("Failed %s: %v\n", name, err)
			continue
		}
		created++
	}

	if created == 0 {
		return fmt.Errorf("no uvx servers found in %s", mcpbFrom)
	}
	fmt.Printf("\nCreated %d bundle(s)\n", created)
	return nil
}

func bundleFromServer(name string, srv *mcpb.MCPServer, output string) error {
	pkg, err := mcpb.PackageFromServer(srv)
	if err != nil {
		return err
	}

	envVars := mcpb.EnvVarsFromServer(srv)
	envVars = append(envVars, extraEnvVars()...)
	fmt.Printf("From server %q: package=%s, %d env var(s)\n", name, pkg, len(envVars))

	return bundlePackage(pkg, envVars, output)
}

func bundlePackage(packageName string, envVars []mcpb.EnvVar, output string) error {
	fmt.Printf("Fetching PyPI metadata for %s...\n", packageName)
	info, err := mcpb.FetchPyPI(packageName)
	if err != nil {
		return err
	}
	fmt.Printf("  %s v%s — %s\n", info.Name, info.Version, info.Summary)
	fmt.Printf("  Entry point: %s.%s\n", info.EntryModule, info.EntryFunc)

	opts := mcpb.BundleOptions{
		PackageName: packageName,
		EnvVars:     envVars,
		OutputPath:  output,
	}

	outPath, err := mcpb.Bundle(info, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Created %s\n", outPath)
	fmt.Println("Drag this file into Claude Desktop > Settings > Extensions to install.")
	return nil
}

func extraEnvVars() []mcpb.EnvVar {
	var vars []mcpb.EnvVar
	for _, e := range mcpbEnvVars {
		parts := strings.SplitN(e, "=", 2)
		name := parts[0]
		desc := name
		if len(parts) == 2 {
			desc = parts[1]
		}
		vars = append(vars, mcpb.EnvVar{
			Name:        name,
			Description: desc,
			Required:    true,
		})
	}
	return vars
}

func init() {
	mcpbCmd.Flags().StringVar(&mcpbFrom, "from", "", "Read server config from a .mcp.json file")
	mcpbCmd.Flags().StringSliceVarP(&mcpbEnvVars, "env", "e", nil, "Additional env vars to expose as user config")
	mcpbCmd.Flags().StringVarP(&mcpbOutput, "output", "o", "", "Output file path (default: {package}.mcpb)")
	rootCmd.AddCommand(mcpbCmd)
}
