package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/agenthubdev/agent-buddy/internal/github"
	"github.com/agenthubdev/agent-buddy/internal/skills"
	"github.com/spf13/cobra"
)

var installOnly []string
var installPrefix string

var installCmd = &cobra.Command{
	Use:   "install [org/repo]",
	Short: "Install skills from a GitHub repo",
	Long: `Fetch a GitHub repo and install its agent skills into ~/.agents/skills/.

The repo is downloaded as a tarball from the main branch, scanned for
directories containing a SKILL.md file, and each discovered skill is
copied to ~/.agents/skills/{prefix}__{skill-name}.

The default prefix is {org}__{repo} (e.g., anthropics__skills), but can
be overridden with --prefix. Use --only to pick specific skills.

Examples:
  agent-buddy install anthropics/skills
  agent-buddy install anthropics/skills --only pdf,docx
  agent-buddy install anthropics/skills --only pdf --prefix anthropics`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		orgRepo := args[0]
		parts := strings.SplitN(orgRepo, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid repo format %q, expected org/repo", orgRepo)
		}
		org, repo := parts[0], parts[1]

		fmt.Printf("Fetching %s...\n", orgRepo)
		tmpDir, err := github.FetchAndExtract(orgRepo)
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		fmt.Println("Installing skills...")
		prefix := installPrefix
		if prefix == "" {
			prefix = org + "__" + repo
		}

		installed, err := skills.Install(prefix, tmpDir, installOnly)
		if err != nil {
			return err
		}

		for _, name := range installed {
			fmt.Printf("  ✓ %s\n", name)
		}
		fmt.Printf("Installed %d skill(s)\n", len(installed))
		return nil
	},
}

func init() {
	installCmd.Flags().StringSliceVar(&installOnly, "only", nil, "Only install specific skills (comma-separated skill names)")
	installCmd.Flags().StringVar(&installPrefix, "prefix", "", "Custom prefix for skill directory names (default: org__repo)")
	rootCmd.AddCommand(installCmd)
}
