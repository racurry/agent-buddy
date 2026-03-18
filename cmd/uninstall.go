package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agenthubdev/agent-buddy/internal/skills"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [skill-name...]",
	Short: "Remove installed skills",
	Long: `Remove one or more installed skills from ~/.agents/skills/.

Pass the full directory name as shown by 'agent-buddy list'. Skills that
don't exist are reported but don't cause a failure.

Examples:
  agent-buddy uninstall anthropics__skills__pdf
  agent-buddy uninstall anthropics__skills__pdf anthropics__skills__docx`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		destBase, err := skills.SkillsDir()
		if err != nil {
			return err
		}

		for _, name := range args {
			target := filepath.Join(destBase, name)

			info, err := os.Stat(target)
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "  ✗ %s (not found)\n", name)
				continue
			}
			if err != nil {
				return err
			}
			if !info.IsDir() {
				fmt.Fprintf(os.Stderr, "  ✗ %s (not a skill directory)\n", name)
				continue
			}

			if err := os.RemoveAll(target); err != nil {
				return fmt.Errorf("removing skill %s: %w", name, err)
			}
			fmt.Printf("  ✓ %s\n", name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
