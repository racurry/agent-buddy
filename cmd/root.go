package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent-buddy",
	Short: "Universal agent skill installer",
	Long: `Universal agent skill installer.

Fetches skills from GitHub repos and installs them into ~/.agents/skills/
so they're discoverable by any AgentSkills-compatible tool (Claude Code,
Gemini CLI, Cursor, Codex, etc.).

Skills are installed as directories named {org}__{repo}__{skill}, which
can be customized with --prefix on install.

Quick start:
  agent-buddy install anthropics/skills              Install all skills from a repo
  agent-buddy install anthropics/skills --only pdf   Install just one skill
  agent-buddy list                                   See what's installed
  agent-buddy uninstall anthropics__skills__pdf       Remove a skill`,
}

func Execute() error {
	return rootCmd.Execute()
}
