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
	syncRepo         string
	syncManifestPath string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync repository settings from a template manifest",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if syncRepo == "" {
			return fmt.Errorf("--repo is required")
		}

		manifest, err := config.LoadManifest(syncManifestPath)
		if err != nil {
			return err
		}

		owner, repo, err := util.ParseOwnerRepo(syncRepo)
		if err != nil {
			return err
		}

		client, err := ghapi.NewRESTClient()
		if err != nil {
			return err
		}

		ui.Info("Syncing %s/%s from %s...", owner, repo, syncManifestPath)

		var (
			errs []error
			mu   sync.Mutex
		)

		ui.Header("Settings:")
		if len(repoUpdatePayload(manifest.Settings)) == 0 {
			ui.Info("No settings configured, skipping.")
		} else if err := ghapi.UpdateRepository(client, owner, repo, manifest.Settings); err != nil {
			ui.Error("Failed to apply repository settings: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Applied repository settings")
		}

		ui.Header("Topics:")
		if len(manifest.Topics) == 0 {
			ui.Info("No topics configured, skipping.")
		} else if err := ghapi.SetTopics(client, owner, repo, manifest.Topics); err != nil {
			ui.Error("Failed to set topics: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Set topics: %v", manifest.Topics)
		}

		ui.Header("Environments:")
		if len(manifest.Environments) == 0 {
			ui.Info("No environments configured, skipping.")
		} else {
			var wg sync.WaitGroup
			wg.Add(len(manifest.Environments))

			for _, env := range manifest.Environments {
				env := env
				go func() {
					defer wg.Done()

					warns, err := ghapi.CreateOrUpdateEnvironment(client, owner, repo, env)
					mu.Lock()
					defer mu.Unlock()
					if err != nil {
						errs = append(errs, err)
						ui.Error("Failed to sync environment %s: %v", env.Name, err)
						return
					}
					ui.Success("Created/updated environment: %s", env.Name)
					for _, w := range warns {
						ui.Warning("%s", w)
					}
				}()
			}

			wg.Wait()
		}

		ui.Header("Actions:")
		if manifest.Actions == nil {
			ui.Info("No actions permissions configured, skipping.")
		} else if err := ghapi.UpdateActionsPermissions(client, owner, repo, manifest.Actions); err != nil {
			ui.Error("Failed to apply actions permissions: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Applied actions permissions")
		}

		ui.Header("Variables:")
		if len(manifest.Variables) == 0 {
			ui.Info("No repository variables configured, skipping.")
		} else if err := ghapi.ApplyRepoVariables(client, owner, repo, manifest.Variables); err != nil {
			ui.Error("Failed to apply repository variables: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Applied %d repository variable(s)", len(manifest.Variables))
		}

		ui.Header("Secrets:")
		if len(manifest.Secrets) == 0 {
			ui.Info("No repository secrets configured, skipping.")
		} else {
			warns, err := ghapi.ApplyRepoSecrets(client, owner, repo, manifest.Secrets)
			if err != nil {
				ui.Error("Failed to apply repository secrets: %v", err)
				errs = append(errs, err)
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
		} else if err := ghapi.UpdateSecuritySettings(client, owner, repo, manifest.Security); err != nil {
			ui.Error("Failed to apply security settings: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Applied security settings")
		}

		if len(errs) == 0 {
			ui.SummaryLine("Sync complete.")
			return nil
		}

		ui.SummaryLine("Sync completed with %d error(s).", len(errs))
		os.Exit(1)
		return nil
	},
}

func init() {
	syncCmd.Flags().StringVarP(&syncRepo, "repo", "r", "", "Repository in owner/repo format")
	syncCmd.Flags().StringVarP(&syncManifestPath, "manifest", "m", "./template-metadata.yml", "Path to the template manifest file")
	_ = syncCmd.MarkFlagRequired("repo")
	rootCmd.AddCommand(syncCmd)
}

func repoUpdatePayload(settings config.RepoSettings) map[string]interface{} {
	payload := make(map[string]interface{})

	if settings.HasWiki != nil {
		payload["has_wiki"] = *settings.HasWiki
	}
	if settings.HasIssues != nil {
		payload["has_issues"] = *settings.HasIssues
	}
	if settings.HasProjects != nil {
		payload["has_projects"] = *settings.HasProjects
	}
	if settings.AllowSquashMerge != nil {
		payload["allow_squash_merge"] = *settings.AllowSquashMerge
	}
	if settings.AllowMergeCommit != nil {
		payload["allow_merge_commit"] = *settings.AllowMergeCommit
	}
	if settings.AllowRebaseMerge != nil {
		payload["allow_rebase_merge"] = *settings.AllowRebaseMerge
	}
	if settings.AllowAutoMerge != nil {
		payload["allow_auto_merge"] = *settings.AllowAutoMerge
	}
	if settings.DeleteBranchOnMerge != nil {
		payload["delete_branch_on_merge"] = *settings.DeleteBranchOnMerge
	}
	if settings.AllowUpdateBranch != nil {
		payload["allow_update_branch"] = *settings.AllowUpdateBranch
	}
	if settings.Visibility != "" {
		payload["visibility"] = settings.Visibility
		if settings.Visibility == "private" {
			payload["private"] = true
		} else {
			payload["private"] = false
		}
	}
	if settings.Description != "" {
		payload["description"] = settings.Description
	}

	return payload
}
