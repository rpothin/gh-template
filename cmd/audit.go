package cmd

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/rpothin/gh-template/internal/config"
	ghapi "github.com/rpothin/gh-template/internal/github"
	"github.com/rpothin/gh-template/internal/ui"
	"github.com/rpothin/gh-template/internal/util"
	"github.com/spf13/cobra"
)

var (
	auditRepo       string
	auditConfigPath string
)

var auditCmd = &cobra.Command{
	Use:          "audit --repo <owner/repo> [--config <path>]",
	Short:        "Audit a repository against a template manifest",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		if auditRepo == "" {
			ui.Error("--repo is required")
			os.Exit(1)
		}

		manifest, err := config.LoadManifest(auditConfigPath)
		if err != nil {
			ui.Error("%v", err)
			os.Exit(1)
		}

		owner, repo, err := util.ParseOwnerRepo(auditRepo)
		if err != nil {
			ui.Error("%v", err)
			os.Exit(1)
		}

		client, err := ghapi.NewRESTClient()
		if err != nil {
			ui.Error("%v", err)
			os.Exit(1)
		}

		var (
			liveRepo    *ghapi.RepoInfo
			liveTopics  []string
			liveEnvs    []config.Environment
			repoErr     error
			topicsErr   error
			envErr      error
			fetchWaiter sync.WaitGroup
		)

		fetchWaiter.Add(3)

		go func() {
			defer fetchWaiter.Done()
			liveRepo, repoErr = ghapi.GetRepository(client, owner, repo)
		}()

		go func() {
			defer fetchWaiter.Done()
			liveTopics, topicsErr = ghapi.GetTopics(client, owner, repo)
		}()

		go func() {
			defer fetchWaiter.Done()
			liveEnvs, envErr = ghapi.GetEnvironments(client, owner, repo)
		}()

		fetchWaiter.Wait()

		if repoErr != nil || topicsErr != nil || envErr != nil {
			if repoErr != nil {
				ui.Error("%v", repoErr)
			}
			if topicsErr != nil {
				ui.Error("%v", topicsErr)
			}
			if envErr != nil {
				ui.Error("%v", envErr)
			}
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Auditing %s against %s\n", auditRepo, auditConfigPath)

		driftCount := 0
		auditSettings(manifest.Settings, liveRepo, &driftCount)
		auditTopics(manifest.Topics, liveTopics, &driftCount)
		auditEnvironments(manifest.Environments, liveEnvs, &driftCount)

		if driftCount == 0 {
			ui.SummaryLine("Summary: ✓ No drift detected. %s matches config.", auditRepo)
			return
		}

		ui.SummaryLine("Summary: %d drift(s) detected in %s", driftCount, auditRepo)
		os.Exit(1)
	},
}

func init() {
	auditCmd.Flags().StringVarP(&auditRepo, "repo", "r", "", "Repository in owner/repo format")
	auditCmd.Flags().StringVarP(&auditConfigPath, "config", "c", "./template-metadata.yml", "Path to config file")
	rootCmd.AddCommand(auditCmd)
}

func auditSettings(settings config.RepoSettings, live *ghapi.RepoInfo, driftCount *int) {
	fmt.Fprintln(os.Stderr, "\nSettings:")

	compareBoolSetting("has_wiki", settings.HasWiki, live.HasWiki, driftCount)
	compareBoolSetting("has_issues", settings.HasIssues, live.HasIssues, driftCount)
	compareBoolSetting("has_projects", settings.HasProjects, live.HasProjects, driftCount)
	compareBoolSetting("allow_squash_merge", settings.AllowSquashMerge, live.AllowSquashMerge, driftCount)
	compareBoolSetting("allow_merge_commit", settings.AllowMergeCommit, live.AllowMergeCommit, driftCount)
	compareBoolSetting("allow_rebase_merge", settings.AllowRebaseMerge, live.AllowRebaseMerge, driftCount)
	compareBoolSetting("allow_auto_merge", settings.AllowAutoMerge, live.AllowAutoMerge, driftCount)
	compareBoolSetting("delete_branch_on_merge", settings.DeleteBranchOnMerge, live.DeleteBranchOnMerge, driftCount)
	compareBoolSetting("allow_update_branch", settings.AllowUpdateBranch, live.AllowUpdateBranch, driftCount)
	compareStringSetting("visibility", settings.Visibility, live.Visibility, driftCount)
	compareStringSetting("description", settings.Description, live.Description, driftCount)
}

func auditTopics(configTopics, liveTopics []string, driftCount *int) {
	fmt.Fprintln(os.Stderr, "\nTopics:")

	configSet := make(map[string]struct{}, len(configTopics))
	liveSet := make(map[string]struct{}, len(liveTopics))
	seenConfig := make(map[string]struct{}, len(configTopics))

	for _, topic := range configTopics {
		configSet[topic] = struct{}{}
	}
	for _, topic := range liveTopics {
		liveSet[topic] = struct{}{}
	}

	for _, topic := range configTopics {
		if _, seen := seenConfig[topic]; seen {
			continue
		}
		seenConfig[topic] = struct{}{}

		if _, ok := liveSet[topic]; ok {
			ui.Success("%s", topic)
			continue
		}

		ui.Error("%s (missing from live repo)", topic)
		*driftCount++
	}

	extraTopics := make([]string, 0)
	for _, topic := range liveTopics {
		if _, ok := configSet[topic]; ok {
			continue
		}
		extraTopics = append(extraTopics, topic)
	}
	sort.Strings(extraTopics)

	seenExtra := make(map[string]struct{}, len(extraTopics))
	for _, topic := range extraTopics {
		if _, seen := seenExtra[topic]; seen {
			continue
		}
		seenExtra[topic] = struct{}{}
		ui.Warning("%s (in live repo, not in config)", topic)
	}
}

func auditEnvironments(configEnvs, liveEnvs []config.Environment, driftCount *int) {
	fmt.Fprintln(os.Stderr, "\nEnvironments:")

	liveByName := make(map[string]config.Environment, len(liveEnvs))
	configByName := make(map[string]struct{}, len(configEnvs))
	for _, env := range liveEnvs {
		liveByName[env.Name] = env
	}

	for _, env := range configEnvs {
		configByName[env.Name] = struct{}{}

		liveEnv, ok := liveByName[env.Name]
		if !ok {
			ui.Error("%s (missing from live repo; want wait_timer: %d)", env.Name, env.WaitTimer)
			*driftCount++
			continue
		}

		if env.WaitTimer == liveEnv.WaitTimer {
			ui.Success("%s (wait_timer: %d)", env.Name, env.WaitTimer)
			continue
		}

		ui.Error("%s (want wait_timer: %d, got %d)", env.Name, env.WaitTimer, liveEnv.WaitTimer)
		*driftCount++
	}

	extraEnvs := make([]config.Environment, 0)
	for _, env := range liveEnvs {
		if _, ok := configByName[env.Name]; ok {
			continue
		}
		extraEnvs = append(extraEnvs, env)
	}
	sort.Slice(extraEnvs, func(i, j int) bool {
		return extraEnvs[i].Name < extraEnvs[j].Name
	})

	seenExtra := make(map[string]struct{}, len(extraEnvs))
	for _, env := range extraEnvs {
		if _, seen := seenExtra[env.Name]; seen {
			continue
		}
		seenExtra[env.Name] = struct{}{}
		ui.Warning("%s (wait_timer: %d, not in config)", env.Name, env.WaitTimer)
	}
}

func compareBoolSetting(field string, want *bool, got bool, driftCount *int) {
	if want == nil {
		return
	}

	if *want == got {
		ui.AuditMatch(field, *want)
		return
	}

	ui.AuditDrift(field, *want, got)
	*driftCount++
}

func compareStringSetting(field, want, got string, driftCount *int) {
	if want == "" {
		return
	}

	if want == got {
		ui.AuditMatch(field, want)
		return
	}

	ui.AuditDrift(field, want, got)
	*driftCount++
}
