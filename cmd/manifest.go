package cmd

import (
	"fmt"
	"os"

	ghapi "github.com/rpothin/gh-template/internal/github"
	"github.com/rpothin/gh-template/internal/config"
	"github.com/rpothin/gh-template/internal/ui"
	"github.com/rpothin/gh-template/internal/util"
	"github.com/spf13/cobra"
)

var manifestFetchOutput string

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Commands for working with template manifests",
}

var manifestFetchCmd = &cobra.Command{
	Use:          "fetch <owner/repo>",
	Short:        "Fetch a template manifest from a repository",
	Long: `Fetches template-metadata.yml from the root of the given repository
and writes it to a local file for inspection and customisation before use.

This is useful when a template maintainer has shipped a recommended
template-metadata.yml alongside their template repository and you want
to review or tweak it before running 'gh template create'.

Example workflow:
  gh template manifest fetch owner/my-template
  # review / edit ./template-metadata.yml
  gh template create my-new-repo --manifest ./template-metadata.yml`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRef := args[0]
		owner, repo, err := util.ParseOwnerRepo(repoRef)
		if err != nil {
			return err
		}

		client, err := ghapi.NewRESTClient()
		if err != nil {
			return err
		}

		ui.Info("Fetching manifest from %s...", repoRef)
		manifest, err := ghapi.FetchManifestFromRepo(client, owner, repo)
		if err != nil {
			return err
		}

		yamlBytes, err := config.ManifestToYAML(manifest)
		if err != nil {
			return err
		}

		outPath := manifestFetchOutput
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
	manifestFetchCmd.Flags().StringVarP(&manifestFetchOutput, "output", "o", "", "File path to write the manifest (default: template-metadata.yml)")
	manifestCmd.AddCommand(manifestFetchCmd)
	rootCmd.AddCommand(manifestCmd)
}
