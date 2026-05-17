package cmd

import (
	"fmt"
	"os"
	"path/filepath"
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
			repoInfo       *ghapi.RepoInfo
			topics         []string
			envs           []config.Environment
			actionsSettings *config.ActionsSettings
			repoVars        []config.EnvironmentVariable
			repoSecrets     []config.EnvironmentSecret
			vulnAlerts      bool
			repoErr         error
			topicsErr       error
			envsErr         error
			actionsErr      error
			repoVarsErr     error
			repoSecretsErr  error
			vulnAlertsErr   error
		)

		var wg sync.WaitGroup
		wg.Add(7)

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

		go func() {
			defer wg.Done()
			actionsSettings, actionsErr = ghapi.GetActionsPermissions(client, owner, repo)
		}()

		go func() {
			defer wg.Done()
			repoVars, repoVarsErr = ghapi.GetRepoVariables(client, owner, repo)
		}()

		go func() {
			defer wg.Done()
			repoSecrets, repoSecretsErr = ghapi.GetRepoSecretNames(client, owner, repo)
		}()

		go func() {
			defer wg.Done()
			vulnAlerts, vulnAlertsErr = ghapi.GetVulnerabilityAlertsEnabled(client, owner, repo)
		}()

		wg.Wait()

		for _, err := range []error{repoErr, topicsErr, envsErr, actionsErr, repoVarsErr, repoSecretsErr, vulnAlertsErr} {
			if err != nil {
				return err
			}
		}

		manifest := &config.Manifest{
			Template:     snapshotRepo,
			Settings:     ghapi.RepoInfoToSettings(repoInfo),
			Topics:       topics,
			Environments: envs,
			Variables:    repoVars,
			Secrets:      repoSecrets,
			Actions:      actionsSettings,
			Security:     ghapi.RepoInfoToSecurity(repoInfo, vulnAlerts),
		}

		yamlBytes, err := config.ManifestToYAML(manifest)
		if err != nil {
			return err
		}

		if snapshotOutput != "" {
			// If the output path is an existing directory, write the default filename inside it.
			if info, statErr := os.Stat(snapshotOutput); statErr == nil && info.IsDir() {
				snapshotOutput = filepath.Join(snapshotOutput, "template-metadata.yml")
			}
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
