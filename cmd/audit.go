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
	auditRepo         string
	auditManifestPath string
)

var auditCmd = &cobra.Command{
	Use:          "audit --repo <owner/repo> [--manifest <path>]",
	Short:        "Audit a repository against a template manifest",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		if auditRepo == "" {
			ui.Error("--repo is required")
			os.Exit(1)
		}

		manifest, err := config.LoadManifest(auditManifestPath)
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
			liveRepo       *ghapi.RepoInfo
			liveTopics     []string
			liveEnvs       []config.Environment
			liveActions    *config.ActionsSettings
			liveVars       []config.EnvironmentVariable
			liveSecrets    []config.EnvironmentSecret
			vulnAlerts     bool
			repoErr        error
			topicsErr      error
			envErr         error
			actionsErr     error
			varsErr        error
			secretsErr     error
			vulnAlertsErr  error
			fetchWaiter    sync.WaitGroup
		)

		fetchWaiter.Add(7)

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

		go func() {
			defer fetchWaiter.Done()
			liveActions, actionsErr = ghapi.GetActionsPermissions(client, owner, repo)
		}()

		go func() {
			defer fetchWaiter.Done()
			liveVars, varsErr = ghapi.GetRepoVariables(client, owner, repo)
		}()

		go func() {
			defer fetchWaiter.Done()
			liveSecrets, secretsErr = ghapi.GetRepoSecretNames(client, owner, repo)
		}()

		go func() {
			defer fetchWaiter.Done()
			vulnAlerts, vulnAlertsErr = ghapi.GetVulnerabilityAlertsEnabled(client, owner, repo)
		}()

		fetchWaiter.Wait()

		if repoErr != nil || topicsErr != nil || envErr != nil || actionsErr != nil || varsErr != nil || secretsErr != nil || vulnAlertsErr != nil {
			if repoErr != nil {
				ui.Error("%v", repoErr)
			}
			if topicsErr != nil {
				ui.Error("%v", topicsErr)
			}
			if envErr != nil {
				ui.Error("%v", envErr)
			}
			if actionsErr != nil {
				ui.Error("%v", actionsErr)
			}
			if varsErr != nil {
				ui.Error("%v", varsErr)
			}
			if secretsErr != nil {
				ui.Error("%v", secretsErr)
			}
			if vulnAlertsErr != nil {
				ui.Error("%v", vulnAlertsErr)
			}
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Auditing %s against %s\n", auditRepo, auditManifestPath)

		driftCount := 0
		auditSettings(manifest.Settings, liveRepo, &driftCount)
		auditTopics(manifest.Topics, liveTopics, &driftCount)
		auditEnvironments(manifest.Environments, liveEnvs, &driftCount)
		auditActions(manifest.Actions, liveActions, &driftCount)
		liveSecurity := ghapi.RepoInfoToSecurity(liveRepo, vulnAlerts)
		auditSecurity(manifest.Security, liveSecurity, &driftCount)
		auditRepoVarsSecrets(manifest.Variables, manifest.Secrets, liveVars, liveSecrets, &driftCount)

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
	auditCmd.Flags().StringVarP(&auditManifestPath, "manifest", "m", "./template-metadata.yml", "Path to the template manifest file")
	rootCmd.AddCommand(auditCmd)
}

func auditSettings(settings config.RepoSettings, live *ghapi.RepoInfo, driftCount *int) {
	fmt.Fprintln(os.Stderr, "\nSettings:")

	compareBoolSetting("has_wiki", settings.HasWiki, live.HasWiki, driftCount)
	compareBoolSetting("has_issues", settings.HasIssues, live.HasIssues, driftCount)
	compareBoolSetting("has_projects", settings.HasProjects, live.HasProjects, driftCount)
	compareBoolSetting("has_discussions", settings.HasDiscussions, live.HasDiscussions, driftCount)
	compareBoolSetting("has_pull_requests", settings.HasPullRequests, live.HasPullRequests, driftCount)
	compareStringSetting("pull_request_creation_policy", settings.PullRequestCreationPolicy, live.PullRequestCreationPolicy, driftCount)
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
			ui.Error("%s (missing from live repo)", env.Name)
			*driftCount++
			continue
		}

		fmt.Fprintf(os.Stderr, "\n  %s:\n", env.Name)
		auditEnvFields(env, liveEnv, driftCount)
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
	for _, env := range extraEnvs {
		ui.Warning("%s (in live repo, not in config)", env.Name)
	}
}

func auditEnvFields(cfg, live config.Environment, driftCount *int) {
	// wait_timer
	if cfg.WaitTimer != live.WaitTimer {
		ui.AuditDrift("wait_timer", cfg.WaitTimer, live.WaitTimer)
		*driftCount++
	} else {
		ui.AuditMatch("wait_timer", cfg.WaitTimer)
	}

	// prevent_self_review
	if cfg.PreventSelfReview != nil {
		liveVal := false
		if live.PreventSelfReview != nil {
			liveVal = *live.PreventSelfReview
		}
		if *cfg.PreventSelfReview != liveVal {
			ui.AuditDrift("prevent_self_review", *cfg.PreventSelfReview, liveVal)
			*driftCount++
		} else {
			ui.AuditMatch("prevent_self_review", *cfg.PreventSelfReview)
		}
	}

	// reviewers
	if cfg.Reviewers != nil {
		liveSet := make(map[string]struct{}, len(live.Reviewers))
		for _, r := range live.Reviewers {
			liveSet[r] = struct{}{}
		}
		for _, r := range cfg.Reviewers {
			if _, ok := liveSet[r]; ok {
				ui.AuditMatch("reviewer", r)
			} else {
				ui.AuditDrift("reviewer", r, "(missing)")
				*driftCount++
			}
		}
		// warn about live reviewers not in the manifest (no drift — manifest owns what it declares)
		cfgSet := make(map[string]struct{}, len(cfg.Reviewers))
		for _, r := range cfg.Reviewers {
			cfgSet[r] = struct{}{}
		}
		for _, r := range live.Reviewers {
			if _, ok := cfgSet[r]; !ok {
				ui.Warning("reviewer: %s (in live repo, not in config)", r)
			}
		}
	}

	// deployment_branch_policy
	if cfg.DeploymentBranchPolicy != "" {
		livePolicy := live.DeploymentBranchPolicy
		if livePolicy == "" {
			livePolicy = "all"
		}
		if cfg.DeploymentBranchPolicy != livePolicy {
			ui.AuditDrift("deployment_branch_policy", cfg.DeploymentBranchPolicy, livePolicy)
			*driftCount++
		} else {
			ui.AuditMatch("deployment_branch_policy", cfg.DeploymentBranchPolicy)
		}
		if cfg.DeploymentBranchPolicy == "custom" {
			livePatternSet := make(map[string]struct{}, len(live.DeploymentBranchPatterns))
			for _, p := range live.DeploymentBranchPatterns {
				livePatternSet[p] = struct{}{}
			}
			for _, p := range cfg.DeploymentBranchPatterns {
				if _, ok := livePatternSet[p]; ok {
					ui.AuditMatch("deployment_branch_pattern", p)
				} else {
					ui.AuditDrift("deployment_branch_pattern", p, "(missing)")
					*driftCount++
				}
			}
			// warn about extra live patterns not in config (sync would remove them)
			cfgPatternSet := make(map[string]struct{}, len(cfg.DeploymentBranchPatterns))
			for _, p := range cfg.DeploymentBranchPatterns {
				cfgPatternSet[p] = struct{}{}
			}
			for _, p := range live.DeploymentBranchPatterns {
				if _, ok := cfgPatternSet[p]; !ok {
					ui.Warning("deployment_branch_pattern: %s (in live repo, not in config; sync would remove it)", p)
				}
			}
		}
	}

	// variables
	if len(cfg.Variables) > 0 {
		liveVarMap := make(map[string]string, len(live.Variables))
		for _, v := range live.Variables {
			liveVarMap[v.Name] = v.Value
		}
		for _, v := range cfg.Variables {
			if liveVal, ok := liveVarMap[v.Name]; ok {
				if v.Value == liveVal {
					ui.AuditMatch("variable:"+v.Name, v.Value)
				} else {
					ui.AuditDrift("variable:"+v.Name, v.Value, liveVal)
					*driftCount++
				}
			} else {
				ui.AuditDrift("variable:"+v.Name, v.Value, "(missing)")
				*driftCount++
			}
		}
	}

	// secrets (name presence only — values are unreadable)
	if len(cfg.Secrets) > 0 {
		liveSecretSet := make(map[string]struct{}, len(live.Secrets))
		for _, s := range live.Secrets {
			liveSecretSet[s.Name] = struct{}{}
		}
		for _, s := range cfg.Secrets {
			if _, ok := liveSecretSet[s.Name]; ok {
				ui.AuditMatch("secret:"+s.Name, "(exists)")
			} else {
				ui.AuditDrift("secret:"+s.Name, "(exists)", "(missing)")
				*driftCount++
			}
		}
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

func auditActions(cfg, live *config.ActionsSettings, driftCount *int) {
	fmt.Fprintln(os.Stderr, "\nActions:")

	if cfg == nil {
		ui.Info("  (no actions permissions configured in manifest)")
		return
	}
	if live == nil {
		ui.Warning("  (could not read live actions permissions)")
		return
	}

	if cfg.ShaPinningRequired != nil {
		liveVal := false
		if live.ShaPinningRequired != nil {
			liveVal = *live.ShaPinningRequired
		}
		if *cfg.ShaPinningRequired != liveVal {
			ui.AuditDrift("sha_pinning_required", *cfg.ShaPinningRequired, liveVal)
			*driftCount++
		} else {
			ui.AuditMatch("sha_pinning_required", *cfg.ShaPinningRequired)
		}
	}

	if cfg.CanApprovePullRequestReviews != nil {
		liveVal := false
		if live.CanApprovePullRequestReviews != nil {
			liveVal = *live.CanApprovePullRequestReviews
		}
		if *cfg.CanApprovePullRequestReviews != liveVal {
			ui.AuditDrift("can_approve_pull_request_reviews", *cfg.CanApprovePullRequestReviews, liveVal)
			*driftCount++
		} else {
			ui.AuditMatch("can_approve_pull_request_reviews", *cfg.CanApprovePullRequestReviews)
		}
	}

	if cfg.DefaultWorkflowPermissions != "" {
		liveVal := live.DefaultWorkflowPermissions
		if cfg.DefaultWorkflowPermissions != liveVal {
			ui.AuditDrift("default_workflow_permissions", cfg.DefaultWorkflowPermissions, liveVal)
			*driftCount++
		} else {
			ui.AuditMatch("default_workflow_permissions", cfg.DefaultWorkflowPermissions)
		}
	}
}

func auditSecurity(cfg, live *config.SecuritySettings, driftCount *int) {
	fmt.Fprintln(os.Stderr, "\nSecurity:")

	if cfg == nil {
		ui.Info("  (no security settings configured in manifest)")
		return
	}

	liveAlerts := false
	if live != nil && live.DependabotAlerts != nil {
		liveAlerts = *live.DependabotAlerts
	}
	compareBoolSetting("dependabot_alerts", cfg.DependabotAlerts, liveAlerts, driftCount)

	var liveUpdates bool
	if live != nil && live.DependabotSecurityUpdates != nil {
		liveUpdates = *live.DependabotSecurityUpdates
	}
	compareBoolSetting("dependabot_security_updates", cfg.DependabotSecurityUpdates, liveUpdates, driftCount)

	var liveScanning bool
	if live != nil && live.SecretScanning != nil {
		liveScanning = *live.SecretScanning
	}
	compareBoolSetting("secret_scanning", cfg.SecretScanning, liveScanning, driftCount)

	var livePushProtection bool
	if live != nil && live.SecretScanningPushProtection != nil {
		livePushProtection = *live.SecretScanningPushProtection
	}
	compareBoolSetting("secret_scanning_push_protection", cfg.SecretScanningPushProtection, livePushProtection, driftCount)
}

func auditRepoVarsSecrets(cfgVars []config.EnvironmentVariable, cfgSecrets []config.EnvironmentSecret, liveVars []config.EnvironmentVariable, liveSecrets []config.EnvironmentSecret, driftCount *int) {
	fmt.Fprintln(os.Stderr, "\nRepository Variables:")

	if len(cfgVars) == 0 {
		ui.Info("  (no repository variables configured in manifest)")
	} else {
		liveVarMap := make(map[string]string, len(liveVars))
		for _, v := range liveVars {
			liveVarMap[v.Name] = v.Value
		}
		for _, v := range cfgVars {
			if liveVal, ok := liveVarMap[v.Name]; ok {
				if v.Value == liveVal {
					ui.AuditMatch("variable:"+v.Name, v.Value)
				} else {
					ui.AuditDrift("variable:"+v.Name, v.Value, liveVal)
					*driftCount++
				}
			} else {
				ui.AuditDrift("variable:"+v.Name, v.Value, "(missing)")
				*driftCount++
			}
		}
	}

	fmt.Fprintln(os.Stderr, "\nRepository Secrets:")

	if len(cfgSecrets) == 0 {
		ui.Info("  (no repository secrets configured in manifest)")
	} else {
		liveSecretSet := make(map[string]struct{}, len(liveSecrets))
		for _, s := range liveSecrets {
			liveSecretSet[s.Name] = struct{}{}
		}
		for _, s := range cfgSecrets {
			if _, ok := liveSecretSet[s.Name]; ok {
				ui.AuditMatch("secret:"+s.Name, "(exists)")
			} else {
				ui.AuditDrift("secret:"+s.Name, "(exists)", "(missing)")
				*driftCount++
			}
		}
	}
}
