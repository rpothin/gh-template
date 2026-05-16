package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rpothin/gh-template/internal/explain"
)

var explainCmd = &cobra.Command{
	Use:   "explain [field]",
	Short: "Show descriptions for all template-metadata.yml fields",
	Long: `Display descriptions for the fields used in template-metadata.yml.

Run without arguments to see the full reference table.
Provide a field name to see its detailed description and an example.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return printFieldDetail(args[0])
		}
		printFullTable()
		return nil
	},
}

const (
	bold  = "\033[1m"
	cyan  = "\033[36m"
	reset = "\033[0m"
	sep   = "─"
)

func printFullTable() {
	// Settings section
	fmt.Fprintf(os.Stdout, "\n"+bold+"SETTINGS"+reset+"\n")
	fmt.Fprintf(os.Stdout, "  %-24s  %-7s  %s\n", "Field", "Type", "Description")
	fmt.Fprintf(os.Stdout, "  %s\n", strings.Repeat(sep, 72))
	for _, f := range explain.Fields {
		if f.Section == explain.SectionSettings {
			fmt.Fprintf(os.Stdout, "  %-24s  %-7s  %s\n", f.Name, f.Type, f.Short)
		}
	}

	// Environments section
	fmt.Fprintf(os.Stdout, "\n"+bold+"ENVIRONMENTS"+reset+"\n")
	fmt.Fprintf(os.Stdout, "  %-12s  %-7s  %s\n", "Field", "Type", "Description")
	fmt.Fprintf(os.Stdout, "  %s\n", strings.Repeat(sep, 72))
	for _, f := range explain.Fields {
		if f.Section == explain.SectionEnvironments {
			fmt.Fprintf(os.Stdout, "  %-12s  %-7s  %s\n", f.Name, f.Type, f.Short)
		}
	}

	// Topics section
	fmt.Fprintf(os.Stdout, "\n"+bold+"TOPICS"+reset+"\n")
	fmt.Fprintf(os.Stdout, "  A flat list of strings that label the repository on GitHub.\n")
	fmt.Fprintf(os.Stdout, "  Example: [\"go\", \"cli\", \"github-extension\"]\n\n")
}

func printFieldDetail(name string) error {
	byName := explain.FieldsByName()
	f, ok := byName[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown field %q.\n\nAvailable fields:\n", name)
		for _, known := range explain.Fields {
			fmt.Fprintf(os.Stderr, "  %s  (%s)\n", known.Name, known.Section)
		}
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "\n"+bold+"%s"+reset+"  (%s · %s · optional)\n\n", f.Name, f.Section, f.Type)
	for _, line := range strings.Split(f.Long, "\n") {
		fmt.Fprintf(os.Stdout, "  %s\n", line)
	}
	fmt.Fprintf(os.Stdout, "\n  "+bold+"Example:"+reset+"\n")
	for _, line := range strings.Split(f.Example, "\n") {
		fmt.Fprintf(os.Stdout, "    %s\n", line)
	}
	if f.DocsURL != "" {
		fmt.Fprintf(os.Stdout, "\n  "+cyan+"Docs:"+reset+" %s\n", f.DocsURL)
	}
	fmt.Fprintln(os.Stdout)
	return nil
}

func init() {
	rootCmd.AddCommand(explainCmd)
}
