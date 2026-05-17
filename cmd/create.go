package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/rpothin/gh-template/internal/config"
	"github.com/rpothin/gh-template/internal/github"
	"github.com/rpothin/gh-template/internal/ui"
	"github.com/rpothin/gh-template/internal/util"
	"github.com/spf13/cobra"
)

var createManifestPath string

type environmentResult struct {
	name     string
	warnings []string
	err      error
}

// isRepoRef returns true when s looks like "owner/repo" rather than a file path.
// A repo reference has exactly one "/" with no file extension on the second segment
// and no filesystem path indicators (leading dot or backslash).
func isRepoRef(s string) bool {
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "/") || strings.Contains(s, "\\") {
		return false
	}
	_, _, err := util.ParseOwnerRepo(s)
	if err != nil {
		return false
	}
	// If the repo segment contains a dot it likely has a file extension (e.g. template.yml).
	parts := strings.SplitN(s, "/", 2)
	return !strings.Contains(parts[1], ".")
}

var createCmd = &cobra.Command{
	Use:   "create <name-or-owner/repo>",
	Short: "Create a repository from a template",
	Long: `Create a repository from a template manifest.

The argument can be a plain repository name (uses the authenticated user as
owner) or an OWNER/REPO reference to create the repository under a specific
owner (user or organisation).

Repository visibility is controlled by the settings.visibility field in the
manifest. If the field is omitted, the repository is created as public.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := args[0]

		client, err := github.NewRESTClient()
		if err != nil {
			return err
		}

		// --manifest accepts either a local file path or an owner/repo reference.
		// When owner/repo is given, template-metadata.yml is fetched from the repo
		// root via the GitHub API and parsed in memory (no temp file).
		var manifest *config.Manifest
		if isRepoRef(createManifestPath) {
			mOwner, mRepo, _ := util.ParseOwnerRepo(createManifestPath)
			manifest, err = github.FetchManifestFromRepo(client, mOwner, mRepo)
			if err != nil {
				return fmt.Errorf("fetching manifest from %s: %w", createManifestPath, err)
			}
			ui.Info("Fetched manifest from %s", createManifestPath)
		} else {
			manifest, err = config.LoadManifest(createManifestPath)
			if err != nil {
				return fmt.Errorf("loading manifest %q: %w", createManifestPath, err)
			}
		}

		// Resolve the template repo: prefer manifest.Template; fall back to the
		// owner/repo value passed to --manifest when fetching remotely.
		templateRef := manifest.Template
		if templateRef == "" && isRepoRef(createManifestPath) {
			templateRef = createManifestPath
		}
		if templateRef == "" {
			return fmt.Errorf("manifest %q is missing the required 'template' field (owner/repo format)", createManifestPath)
		}

		templateOwner, templateRepo, err := util.ParseOwnerRepo(templateRef)
		if err != nil {
			return fmt.Errorf("invalid template field in manifest: %w", err)
		}

		// Parse the repository target: OWNER/REPO or just REPO (uses the
		// authenticated user as owner).
		var owner, repoName string
		if strings.Contains(arg, "/") {
			owner, repoName, err = util.ParseOwnerRepo(arg)
			if err != nil {
				return fmt.Errorf("invalid repository reference %q: %w", arg, err)
			}
		} else {
			repoName = arg
			owner, err = github.GetAuthenticatedUser(client)
			if err != nil {
				return err
			}
		}

		// Derive the private flag from the manifest visibility.
		// Empty visibility defaults to public (false).
		private := manifest.Settings.Visibility == "private" || manifest.Settings.Visibility == "internal"

		ui.Info("Creating repository %s/%s from template %s/%s...", owner, repoName, templateOwner, templateRepo)

		repo, err := github.CreateFromTemplate(client, templateOwner, templateRepo, owner, repoName, private)
		if err != nil {
			ui.Error("Failed to create repository: %v", err)
			os.Exit(1)
		}

		ui.Success("Repository created: %s", repo.HTMLURL)

		var errs []error

		ui.Header("Settings:")
		if err := github.UpdateRepository(client, owner, repoName, manifest.Settings); err != nil {
			ui.Error("Failed to apply repository settings: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Applied repository settings")
		}

		ui.Header("Topics:")
		if manifest.Topics == nil {
			ui.Info("No topics configured, skipping.")
		} else if err := github.SetTopics(client, owner, repoName, manifest.Topics); err != nil {
			ui.Error("Failed to set topics: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Set topics: [%s]", strings.Join(manifest.Topics, ", "))
		}

		ui.Header("Environments:")
		if len(manifest.Environments) == 0 {
			ui.Success("No environments to apply")
		} else {
			var wg sync.WaitGroup
			envCh := make(chan config.Environment)
			resultCh := make(chan environmentResult, len(manifest.Environments))

			workerCount := 4
			if len(manifest.Environments) < workerCount {
				workerCount = len(manifest.Environments)
			}

			wg.Add(workerCount)
			for i := 0; i < workerCount; i++ {
				go func() {
					defer wg.Done()
					for env := range envCh {
						warns, err := github.CreateOrUpdateEnvironment(client, owner, repoName, env)
						resultCh <- environmentResult{
							name:     env.Name,
							warnings: warns,
							err:      err,
						}
					}
				}()
			}

			for _, env := range manifest.Environments {
				envCh <- env
			}
			close(envCh)

			wg.Wait()
			close(resultCh)

			for result := range resultCh {
				if result.err != nil {
					ui.Error("Failed to create/update environment %s: %v", result.name, result.err)
					errs = append(errs, result.err)
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
		} else if err := github.UpdateActionsPermissions(client, owner, repoName, manifest.Actions); err != nil {
			ui.Error("Failed to apply actions permissions: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Applied actions permissions")
		}

		ui.Header("Variables:")
		if len(manifest.Variables) == 0 {
			ui.Success("No repository variables to apply")
		} else if err := github.ApplyRepoVariables(client, owner, repoName, manifest.Variables); err != nil {
			ui.Error("Failed to apply repository variables: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Applied %d repository variable(s)", len(manifest.Variables))
		}

		ui.Header("Secrets:")
		if len(manifest.Secrets) == 0 {
			ui.Success("No repository secrets to apply")
		} else {
			warns, err := github.ApplyRepoSecrets(client, owner, repoName, manifest.Secrets)
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
		} else if err := github.UpdateSecuritySettings(client, owner, repoName, manifest.Security); err != nil {
			ui.Error("Failed to apply security settings: %v", err)
			errs = append(errs, err)
		} else {
			ui.Success("Applied security settings")
		}

		if len(errs) > 0 {
			ui.SummaryLine("Done with %d error(s). Repository available at %s", len(errs), repo.HTMLURL)
			os.Exit(1)
		}

		ui.SummaryLine("Done! Repository available at %s", repo.HTMLURL)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createManifestPath, "manifest", "m", "./template-metadata.yml", "Path to the template manifest file")
	rootCmd.AddCommand(createCmd)
}
