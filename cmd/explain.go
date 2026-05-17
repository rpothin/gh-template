package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rpothin/gh-template/internal/explain"
)

var explainAll bool

var explainCmd = &cobra.Command{
	Use:   "explain [field]",
	Short: "Show descriptions for all template-metadata.yml fields",
	Long: `Display descriptions for the fields used in template-metadata.yml.

Run without arguments to see the full reference table of all fields.
Use --all to see detailed descriptions for every field at once.
Provide a field name to see its detailed description and an example value.`,
	Example: `  # Show the reference table for all manifest fields
  $ gh template explain

  # Show detailed description and example for a specific field
  $ gh template explain visibility

  # Show detailed descriptions for all fields at once
  $ gh template explain --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if explainAll {
			printAllDetails()
			return nil
		}
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

func printAllDetails() {
	sections := []explain.Section{explain.SectionSettings, explain.SectionEnvironments}
	for _, section := range sections {
		fmt.Fprintf(os.Stdout, "\n"+bold+"%s"+reset+"\n", strings.ToUpper(string(section)))
		fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat(sep, 76))
		for _, f := range explain.Fields {
			if f.Section == section {
				printFieldEntry(f)
			}
		}
	}

	// Topics is a section without named fields — print a standalone entry.
	fmt.Fprintf(os.Stdout, "\n"+bold+"TOPICS"+reset+"\n")
	fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat(sep, 76))
	fmt.Fprintf(os.Stdout, "  A flat list of strings that label the repository on GitHub.\n\n")
	fmt.Fprintf(os.Stdout, "  "+bold+"Example:"+reset+"\n")
	fmt.Fprintf(os.Stdout, "    topics:\n      - go\n      - cli\n      - github-extension\n\n")
}

func printFieldEntry(f explain.FieldDef) {
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
	printFieldEntry(f)
	return nil
}

func init() {
	explainCmd.Flags().BoolVar(&explainAll, "all", false, "Show detailed descriptions for every field")
	rootCmd.AddCommand(explainCmd)
}
