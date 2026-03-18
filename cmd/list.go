package cmd

import (
	"fmt"
	"os"

	"github.com/agenthubdev/agent-buddy/internal/skills"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Long: `List all agent skills installed in ~/.agents/skills/.

Shows the directory name for each installed skill, which is the name
you'd pass to 'agent-buddy uninstall'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		destBase, err := skills.SkillsDir()
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(destBase)
		if os.IsNotExist(err) {
			fmt.Println("No skills installed.")
			return nil
		}
		if err != nil {
			return err
		}

		count := 0
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			fmt.Println(e.Name())
			count++
		}

		if count == 0 {
			fmt.Println("No skills installed.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
