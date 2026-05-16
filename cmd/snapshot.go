package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/rpothin/gh-template/internal/config"
	ghapi "github.com/rpothin/gh-template/internal/github"
	"github.com/rpothin/gh-template/internal/ui"
	"github.com/rpothin/gh-template/internal/util"
)

var (
	snapshotRepo   string
	snapshotOutput string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Snapshot a repository's settings into YAML",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if snapshotRepo == "" {
			return fmt.Errorf("--repo is required")
		}

		owner, repo, err := util.ParseOwnerRepo(snapshotRepo)
		if err != nil {
			return err
		}

		ui.Info("Snapshotting %s...", snapshotRepo)

		client, err := ghapi.NewRESTClient()
		if err != nil {
			return err
		}

		var (
			repoInfo  *ghapi.RepoInfo
			topics    []string
			envs      []config.Environment
			repoErr   error
			topicsErr error
			envsErr   error
		)

		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			repoInfo, repoErr = ghapi.GetRepository(client, owner, repo)
		}()

		go func() {
			defer wg.Done()
			topics, topicsErr = ghapi.GetTopics(client, owner, repo)
		}()

		go func() {
			defer wg.Done()
			envs, envsErr = ghapi.GetEnvironments(client, owner, repo)
		}()

		wg.Wait()

		for _, err := range []error{repoErr, topicsErr, envsErr} {
			if err != nil {
				return err
			}
		}

		manifest := &config.Manifest{
			Settings:     ghapi.RepoInfoToSettings(repoInfo),
			Topics:       topics,
			Environments: envs,
		}

		yamlBytes, err := config.ManifestToYAML(manifest)
		if err != nil {
			return err
		}

		if snapshotOutput != "" {
			if err := os.WriteFile(snapshotOutput, yamlBytes, 0644); err != nil {
				return fmt.Errorf("writing snapshot to %q: %w", snapshotOutput, err)
			}
			ui.Success("Snapshot written to %s", snapshotOutput)
			return nil
		}

		fmt.Print(string(yamlBytes))
		return nil
	},
}

func init() {
	snapshotCmd.Flags().StringVarP(&snapshotRepo, "repo", "r", "", "Repository in owner/repo format")
	snapshotCmd.Flags().StringVarP(&snapshotOutput, "output", "o", "", "File path to write the snapshot YAML to")
	_ = snapshotCmd.MarkFlagRequired("repo")
	rootCmd.AddCommand(snapshotCmd)
}
