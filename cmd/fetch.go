package cmd

import (
	"fmt"
	"os"

	"github.com/rpothin/gh-template/internal/config"
	ghapi "github.com/rpothin/gh-template/internal/github"
	"github.com/rpothin/gh-template/internal/ui"
	"github.com/rpothin/gh-template/internal/util"
	"github.com/spf13/cobra"
)

var (
	fetchRepo   string
	fetchOutput string
)

var fetchCmd = &cobra.Command{
	Use:   "fetch --repo <owner/repo> [--output <path>]",
	Short: "Fetch a template manifest from a repository",
	Long: `Fetch template-metadata.yml from the root of the given repository and write it
to a local file for inspection and customisation before use.

This is useful when a template maintainer has published a recommended manifest
alongside their template repository and you want to review or tweak it before
provisioning a new repository.

If no --output path is specified, the manifest is written to
template-metadata.yml in the current directory.

To create a repository using the fetched manifest, run:
  gh template create my-new-repo --manifest ./template-metadata.yml`,
	Example: `  # Fetch the manifest from a template repository
  $ gh template fetch --repo owner/my-template

  # Fetch and save to a custom path
  $ gh template fetch --repo owner/my-template --output ./configs/template.yml

  # Typical workflow: fetch, review, then create
  $ gh template fetch --repo owner/my-template
  $ gh template create my-new-repo`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, err := util.ParseOwnerRepo(fetchRepo)
		if err != nil {
			return err
		}

		client, err := ghapi.NewRESTClient()
		if err != nil {
			return err
		}

		ui.Info("Fetching manifest from %s...", fetchRepo)
		manifest, err := ghapi.FetchManifestFromRepo(client, owner, repo)
		if err != nil {
			return err
		}

		yamlBytes, err := config.ManifestToYAML(manifest)
		if err != nil {
			return err
		}

		outPath := fetchOutput
		if outPath == "" {
			outPath = "template-metadata.yml"
		}

		if err := os.WriteFile(outPath, yamlBytes, 0644); err != nil {
			return fmt.Errorf("writing manifest to %q: %w", outPath, err)
		}

		ui.Success("Manifest written to %s", outPath)
		return nil
	},
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchRepo, "repo", "r", "", "Repository in owner/repo format")
	fetchCmd.Flags().StringVarP(&fetchOutput, "output", "o", "", "File path to write the manifest (default: template-metadata.yml)")
	_ = fetchCmd.MarkFlagRequired("repo")
	rootCmd.AddCommand(fetchCmd)
}
