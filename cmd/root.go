package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage GitHub repositories with configuration templates",
	Long: `gh-template is a GitHub CLI extension for individual developers.

It eliminates "ClickOps" by fully establishing repository configurations
(topics, features, environments) during creation, auditing live repositories
for settings drift, and snapshotting optimal configurations into shareable
YAML definitions.`,
}

// Execute runs the root cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
