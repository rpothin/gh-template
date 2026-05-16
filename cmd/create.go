package cmd

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/rpothin/gh-template/internal/config"
	"github.com/rpothin/gh-template/internal/github"
	"github.com/rpothin/gh-template/internal/ui"
	"github.com/rpothin/gh-template/internal/util"
	"github.com/spf13/cobra"
)

var (
	createTemplate   string
	createConfigPath string
	createPrivate    bool
)

type environmentResult struct {
	name     string
	warnings []string
	err      error
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a repository from a template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		templateOwner, templateRepo, err := util.ParseOwnerRepo(createTemplate)
		if err != nil {
			return err
		}

		client, err := github.NewRESTClient()
		if err != nil {
			return err
		}

		owner, err := github.GetAuthenticatedUser(client)
		if err != nil {
			return err
		}

		ui.Info("Creating repository %s/%s from template %s/%s...", owner, name, templateOwner, templateRepo)

		repo, err := github.CreateFromTemplate(client, templateOwner, templateRepo, owner, name, createPrivate)
		if err != nil {
			ui.Error("Failed to create repository: %v", err)
			os.Exit(1)
		}

		ui.Success("Repository created: %s", repo.HTMLURL)

		manifest, err := config.LoadManifest(createConfigPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				ui.Warning("Config file not found: %s (skipping config application)", createConfigPath)
				ui.SummaryLine("Done! Repository available at %s", repo.HTMLURL)
				return nil
			}
			return err
		}

		ui.Header("Settings:")
		if err := github.UpdateRepository(client, owner, name, manifest.Settings); err != nil {
			ui.Error("Failed to apply repository settings: %v", err)
		} else {
			ui.Success("Applied repository settings")
		}

		ui.Header("Topics:")
		if err := github.SetTopics(client, owner, name, manifest.Topics); err != nil {
			ui.Error("Failed to set topics: %v", err)
		} else {
			ui.Success("Set topics: [%s]", strings.Join(manifest.Topics, ", "))
		}

		ui.Header("Environments:")
		if len(manifest.Environments) == 0 {
			ui.Success("No environments to apply")
		} else {
			var wg sync.WaitGroup
			errCh := make(chan environmentResult, len(manifest.Environments))

			for _, env := range manifest.Environments {
				env := env
				wg.Add(1)
				go func() {
					defer wg.Done()
					warns, err := github.CreateOrUpdateEnvironment(client, owner, name, env)
					errCh <- environmentResult{
						name:     env.Name,
						warnings: warns,
						err:      err,
					}
				}()
			}

			wg.Wait()
			close(errCh)

			for result := range errCh {
				if result.err != nil {
					ui.Error("Failed to create/update environment %s: %v", result.name, result.err)
					continue
				}
				ui.Success("Created/updated environment: %s", result.name)
				for _, w := range result.warnings {
					ui.Warning("%s", w)
				}
			}
		}

		ui.Header("Actions:")
		if manifest.Actions == nil {
			ui.Info("No actions permissions configured, skipping.")
		} else if err := github.UpdateActionsPermissions(client, owner, name, manifest.Actions); err != nil {
			ui.Error("Failed to apply actions permissions: %v", err)
		} else {
			ui.Success("Applied actions permissions")
		}

		ui.Header("Variables:")
		if len(manifest.Variables) == 0 {
			ui.Success("No repository variables to apply")
		} else if err := github.ApplyRepoVariables(client, owner, name, manifest.Variables); err != nil {
			ui.Error("Failed to apply repository variables: %v", err)
		} else {
			ui.Success("Applied %d repository variable(s)", len(manifest.Variables))
		}

		ui.Header("Secrets:")
		if len(manifest.Secrets) == 0 {
			ui.Success("No repository secrets to apply")
		} else {
			warns, err := github.ApplyRepoSecrets(client, owner, name, manifest.Secrets)
			if err != nil {
				ui.Error("Failed to apply repository secrets: %v", err)
			} else {
				ui.Success("Applied %d repository secret(s)", len(manifest.Secrets))
				for _, w := range warns {
					ui.Warning("%s", w)
				}
			}
		}

		ui.Header("Security:")
		if manifest.Security == nil {
			ui.Info("No security settings configured, skipping.")
		} else if err := github.UpdateSecuritySettings(client, owner, name, manifest.Security); err != nil {
			ui.Error("Failed to apply security settings: %v", err)
		} else {
			ui.Success("Applied security settings")
		}

		ui.SummaryLine("Done! Repository available at %s", repo.HTMLURL)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createTemplate, "template", "t", "", "Template repository in owner/repo format")
	createCmd.Flags().StringVarP(&createConfigPath, "config", "c", "./template-metadata.yml", "Path to config file")
	createCmd.Flags().BoolVar(&createPrivate, "private", false, "Create as a private repository")
	_ = createCmd.MarkFlagRequired("template")
	rootCmd.AddCommand(createCmd)
}
