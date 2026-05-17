package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	ghapi "github.com/rpothin/gh-template/internal/github"
	"github.com/rpothin/gh-template/internal/ui"
	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for public template repositories",
	Long: `Searches GitHub for public template repositories matching the given query.

The query is automatically prefixed with "template:true". You can pass any
GitHub repository search qualifiers as part of the query:

  gh template search go cli
  gh template search starter language:go
  gh template search org:github

Results are sorted by star count (descending). Use --limit to control
how many results are returned (default: 30, max: 100).`,
	Args:         cobra.ArbitraryArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		client, err := ghapi.NewRESTClient()
		if err != nil {
			return err
		}

		if query == "" {
			ui.Info("Searching for popular public template repositories...")
		} else {
			ui.Info("Searching for template repositories matching %q...", query)
		}

		repos, err := ghapi.SearchTemplateRepos(client, query, searchLimit, func(msg string) {
			ui.Warning("%s", msg)
		})
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			ui.Info("No template repositories found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "REPO\t★\tDESCRIPTION")
		fmt.Fprintln(w, "----\t-\t-----------")
		for _, r := range repos {
			desc := r.Description
			if desc == "" {
				desc = "-"
			}
			fmt.Fprintf(w, "%s\t%d\t%s\n", r.FullName, r.StarCount, desc)
		}
		w.Flush()

		ui.Info("Found %d result(s).", len(repos))
		return nil
	},
}

func init() {
	searchCmd.Flags().IntVar(&searchLimit, "limit", 30, "Maximum number of results to return (max 100)")
	rootCmd.AddCommand(searchCmd)
}
