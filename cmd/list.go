package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	ghapi "github.com/rpothin/gh-template/internal/github"
	"github.com/rpothin/gh-template/internal/ui"
	"github.com/spf13/cobra"
)

var listIncludeOrgs bool
var listIncludeArchived bool
var listFormat string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your template repositories",
	Long: `List template repositories owned by the authenticated user.

These are the repositories marked as templates in GitHub's repository settings —
the ones that appear in GitHub's "Choose a template" picker when creating a new
repository.

Use --include-orgs to also include template repositories from all organisations
you belong to.

Archived template repositories are excluded by default. Use --include-archived
to include them.

Use --format json to get machine-readable output suitable for scripting and
piping into other tools.`,
	Example: `  # List your template repositories
  $ gh template list

  # Include templates from your organizations
  $ gh template list --include-orgs

  # Include archived templates
  $ gh template list --include-orgs --include-archived

  # List in JSON format for scripting
  $ gh template list --format json | ConvertFrom-Json`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if listFormat != "table" && listFormat != "json" {
			return fmt.Errorf("invalid format %q: must be table or json", listFormat)
		}
		jsonMode := listFormat == "json"

		client, err := ghapi.NewRESTClient()
		if err != nil {
			return err
		}

		if !jsonMode {
			ui.Info("Fetching your template repositories...")
		}
		repos, err := ghapi.ListOwnedTemplateRepos(client, listIncludeArchived)
		if err != nil {
			return err
		}

		if listIncludeOrgs {
			if !jsonMode {
				ui.Info("Fetching organisation template repositories...")
			}
			orgRepos, err := ghapi.ListOrgTemplateRepos(client, listIncludeArchived, func(msg string) {
				ui.Warning("%s", msg)
			})
			if err != nil {
				return err
			}
			repos = append(repos, orgRepos...)
		}

		if jsonMode {
			if repos == nil {
				repos = []ghapi.TemplateSummary{}
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(repos)
		}

		if len(repos) == 0 {
			ui.Info("No template repositories found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "REPO\tVISIBILITY\tDESCRIPTION")
		fmt.Fprintln(w, "----\t----------\t-----------")
		for _, r := range repos {
			desc := r.Description
			if desc == "" {
				desc = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.FullName, r.Visibility, desc)
		}
		w.Flush()

		ui.Info("Found %d template repository(ies).", len(repos))
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listIncludeOrgs, "include-orgs", false, "Also list template repositories from your organisations")
	listCmd.Flags().BoolVar(&listIncludeArchived, "include-archived", false, "Include archived template repositories")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format: table or json")
	rootCmd.AddCommand(listCmd)
}

