package cmd

import (
	"encoding/json"
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

// --- Audit report types ------------------------------------------------------

// AuditDriftItem represents a single field where the live value differs from
// the manifest value.
type AuditDriftItem struct {
	Section string `json:"section"`
	Field   string `json:"field"`
	Want    string `json:"want"`
	Got     string `json:"got"`
}

// AuditWarning represents a non-drift observation (e.g. a live value not
// present in the manifest).
type AuditWarning struct {
	Section string `json:"section"`
	Message string `json:"message"`
}

// AuditReport is the full structured result of an audit run.
type AuditReport struct {
	Repo       string           `json:"repo"`
	Manifest   string           `json:"manifest"`
	DriftCount int              `json:"drift_count"`
	Drifts     []AuditDriftItem `json:"drifts"`
	Warnings   []AuditWarning   `json:"warnings"`
}

// --- Audit collector ---------------------------------------------------------

// auditCollector accumulates drift items and warnings. In table mode it also
// prints them to stderr as they are recorded, matching the pre-existing UX.
type auditCollector struct {
	tableMode bool
	section   string
	Report    AuditReport
}

func newAuditCollector(tableMode bool, repo, manifest string) *auditCollector {
	return &auditCollector{
		tableMode: tableMode,
		Report: AuditReport{
			Repo:     repo,
			Manifest: manifest,
			Drifts:   []AuditDriftItem{},
			Warnings: []AuditWarning{},
		},
	}
}

func (c *auditCollector) setSection(name string) {
	c.section = name
	if c.tableMode {
		fmt.Fprintf(os.Stderr, "\n%s:\n", name)
	}
}

// enterSubSection sets the current section key (used in JSON) and optionally
// prints an indented sub-header in table mode.
func (c *auditCollector) enterSubSection(displayName, sectionKey string) {
	c.section = sectionKey
	if c.tableMode {
		fmt.Fprintf(os.Stderr, "\n  %s:\n", displayName)
	}
}

func (c *auditCollector) match(field string, val interface{}) {
	if c.tableMode {
		ui.AuditMatch(field, val)
	}
}

// successLine prints only in table mode (no JSON data recorded). Used for
// list-style sections like topics where the value IS the field name.
func (c *auditCollector) successLine(msg string) {
	if c.tableMode {
		ui.Success("%s", msg)
	}
}

// driftRaw records a drift item and optionally prints a freeform error line in
// table mode. Used for sections where the built-in "want X, got Y" format is
// not appropriate (e.g. topics).
func (c *auditCollector) driftRaw(field, want, got, tableMsg string) {
	c.Report.Drifts = append(c.Report.Drifts, AuditDriftItem{
		Section: c.section,
		Field:   field,
		Want:    want,
		Got:     got,
	})
	c.Report.DriftCount++
	if c.tableMode {
		ui.Error("%s", tableMsg)
	}
}

func (c *auditCollector) drift(field string, want, got interface{}) {
	c.Report.Drifts = append(c.Report.Drifts, AuditDriftItem{
		Section: c.section,
		Field:   field,
		Want:    fmt.Sprintf("%v", want),
		Got:     fmt.Sprintf("%v", got),
	})
	c.Report.DriftCount++
	if c.tableMode {
		ui.AuditDrift(field, want, got)
	}
}

func (c *auditCollector) warn(msg string) {
	c.Report.Warnings = append(c.Report.Warnings, AuditWarning{
		Section: c.section,
		Message: msg,
	})
	if c.tableMode {
		ui.Warning("%s", msg)
	}
}

func (c *auditCollector) info(msg string) {
	if c.tableMode {
		ui.Info("  %s", msg)
	}
}

// --- Command -----------------------------------------------------------------

var (
	auditRepo         string
	auditManifestPath string
	auditFormat       string
)

var auditCmd = &cobra.Command{
	Use:   "audit --repo <owner/repo> [--manifest <path>]",
	Short: "Audit a repository against a template manifest",
	Long: `Compares the live GitHub repository settings against a local manifest file,
visually detailing any configuration drift.

Use --format json to get a machine-readable report suitable for CI pipelines:
  gh template audit --repo owner/repo --format json | ConvertFrom-Json`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		if auditFormat != "table" && auditFormat != "json" {
			ui.Error("invalid format %q: must be table or json", auditFormat)
			os.Exit(1)
		}
		tableMode := auditFormat == "table"

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
			liveRepo      *ghapi.RepoInfo
			liveTopics    []string
			liveEnvs      []config.Environment
			liveActions   *config.ActionsSettings
			liveVars      []config.EnvironmentVariable
			liveSecrets   []config.EnvironmentSecret
			vulnAlerts    bool
			privVulnRep   bool
			repoErr       error
			topicsErr     error
			envErr        error
			actionsErr    error
			varsErr       error
			secretsErr    error
			vulnAlertsErr error
			privVulnRepErr error
			fetchWaiter   sync.WaitGroup
		)

		fetchWaiter.Add(8)

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

		go func() {
			defer fetchWaiter.Done()
			privVulnRep, privVulnRepErr = ghapi.GetPrivateVulnerabilityReportingEnabled(client, owner, repo)
		}()

		fetchWaiter.Wait()

		if repoErr != nil || topicsErr != nil || envErr != nil || actionsErr != nil || varsErr != nil || secretsErr != nil || vulnAlertsErr != nil || privVulnRepErr != nil {
			for _, e := range []error{repoErr, topicsErr, envErr, actionsErr, varsErr, secretsErr, vulnAlertsErr, privVulnRepErr} {
				if e != nil {
					ui.Error("%v", e)
				}
			}
			os.Exit(1)
		}

		if tableMode {
			fmt.Fprintf(os.Stderr, "Auditing %s against %s\n", auditRepo, auditManifestPath)
		}

		c := newAuditCollector(tableMode, auditRepo, auditManifestPath)

		auditSettings(manifest.Settings, liveRepo, c)
		auditTopics(manifest.Topics, liveTopics, c)
		auditEnvironments(manifest.Environments, liveEnvs, c)
		auditActions(manifest.Actions, liveActions, c)
		liveSecurity := ghapi.RepoInfoToSecurity(liveRepo, vulnAlerts, privVulnRep)
		auditSecurity(manifest.Security, liveSecurity, c)
		auditRepoVarsSecrets(manifest.Variables, manifest.Secrets, liveVars, liveSecrets, c)

		if !tableMode {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(c.Report)
		}

		if c.Report.DriftCount == 0 {
			if tableMode {
				ui.SummaryLine("Summary: ✓ No drift detected. %s matches config.", auditRepo)
			}
			return
		}

		if tableMode {
			ui.SummaryLine("Summary: %d drift(s) detected in %s", c.Report.DriftCount, auditRepo)
		}
		os.Exit(1)
	},
}

func init() {
	auditCmd.Flags().StringVarP(&auditRepo, "repo", "r", "", "Repository in owner/repo format")
	auditCmd.Flags().StringVarP(&auditManifestPath, "manifest", "m", "./template-metadata.yml", "Path to the template manifest file")
	auditCmd.Flags().StringVar(&auditFormat, "format", "table", "Output format: table or json")
	rootCmd.AddCommand(auditCmd)
}

// --- Audit helper functions --------------------------------------------------

func auditSettings(settings config.RepoSettings, live *ghapi.RepoInfo, c *auditCollector) {
	c.setSection("Settings")

	compareBoolSetting("has_wiki", settings.HasWiki, live.HasWiki, c)
	compareBoolSetting("has_issues", settings.HasIssues, live.HasIssues, c)
	compareBoolSetting("has_projects", settings.HasProjects, live.HasProjects, c)
	compareBoolSetting("has_discussions", settings.HasDiscussions, live.HasDiscussions, c)
	compareBoolSetting("has_pull_requests", settings.HasPullRequests, live.HasPullRequests, c)
	compareStringSetting("pull_request_creation_policy", settings.PullRequestCreationPolicy, live.PullRequestCreationPolicy, c)
	compareBoolSetting("allow_squash_merge", settings.AllowSquashMerge, live.AllowSquashMerge, c)
	compareBoolSetting("allow_merge_commit", settings.AllowMergeCommit, live.AllowMergeCommit, c)
	compareBoolSetting("allow_rebase_merge", settings.AllowRebaseMerge, live.AllowRebaseMerge, c)
	compareBoolSetting("allow_auto_merge", settings.AllowAutoMerge, live.AllowAutoMerge, c)
	compareBoolSetting("delete_branch_on_merge", settings.DeleteBranchOnMerge, live.DeleteBranchOnMerge, c)
	compareBoolSetting("allow_update_branch", settings.AllowUpdateBranch, live.AllowUpdateBranch, c)
	compareStringSetting("visibility", settings.Visibility, live.Visibility, c)
	// description is seed-only: only flag drift when the manifest specifies
	// one AND the live repo has none yet. Once the owner has set any
	// description it is treated as customised and silently skipped.
	if settings.Description != "" {
		if live.Description != "" {
			c.info("description (live value kept, seed-only)")
		} else {
			c.drift("description", settings.Description, "")
		}
	}
}

func auditTopics(configTopics, liveTopics []string, c *auditCollector) {
	c.setSection("Topics")

	// topics is seed-only: once the live repo has any topics the owner has
	// customised them, so skip the diff entirely.
	if len(configTopics) > 0 && len(liveTopics) > 0 {
		c.info("topics (live values kept, seed-only)")
		return
	}

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
			c.successLine(topic)
			continue
		}

		c.driftRaw(topic, topic, "(missing from live repo)", topic+" (missing from live repo)")
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
		c.warn(fmt.Sprintf("%s (in live repo, not in config)", topic))
	}
}

func auditEnvironments(configEnvs, liveEnvs []config.Environment, c *auditCollector) {
	c.setSection("Environments")

	liveByName := make(map[string]config.Environment, len(liveEnvs))
	configByName := make(map[string]struct{}, len(configEnvs))
	for _, env := range liveEnvs {
		liveByName[env.Name] = env
	}

	for _, env := range configEnvs {
		configByName[env.Name] = struct{}{}

		liveEnv, ok := liveByName[env.Name]
		if !ok {
			c.driftRaw(env.Name, env.Name, "(missing from live repo)", env.Name+" (missing from live repo)")
			continue
		}

		c.enterSubSection(env.Name, "environments/"+env.Name)
		auditEnvFields(env, liveEnv, c)
		c.setSection("Environments") // restore section after sub-section
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
		c.warn(fmt.Sprintf("%s (in live repo, not in config)", env.Name))
	}
}

func auditEnvFields(cfg, live config.Environment, c *auditCollector) {
	// wait_timer
	if cfg.WaitTimer != live.WaitTimer {
		c.drift("wait_timer", cfg.WaitTimer, live.WaitTimer)
	} else {
		c.match("wait_timer", cfg.WaitTimer)
	}

	// prevent_self_review
	if cfg.PreventSelfReview != nil {
		liveVal := false
		if live.PreventSelfReview != nil {
			liveVal = *live.PreventSelfReview
		}
		if *cfg.PreventSelfReview != liveVal {
			c.drift("prevent_self_review", *cfg.PreventSelfReview, liveVal)
		} else {
			c.match("prevent_self_review", *cfg.PreventSelfReview)
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
				c.match("reviewer", r)
			} else {
				c.drift("reviewer", r, "(missing)")
			}
		}
		cfgSet := make(map[string]struct{}, len(cfg.Reviewers))
		for _, r := range cfg.Reviewers {
			cfgSet[r] = struct{}{}
		}
		for _, r := range live.Reviewers {
			if _, ok := cfgSet[r]; !ok {
				c.warn(fmt.Sprintf("reviewer: %s (in live repo, not in config)", r))
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
			c.drift("deployment_branch_policy", cfg.DeploymentBranchPolicy, livePolicy)
		} else {
			c.match("deployment_branch_policy", cfg.DeploymentBranchPolicy)
		}
		if cfg.DeploymentBranchPolicy == "selected" {
			livePatternSet := make(map[string]struct{}, len(live.DeploymentBranchPatterns))
			for _, p := range live.DeploymentBranchPatterns {
				livePatternSet[p] = struct{}{}
			}
			for _, p := range cfg.DeploymentBranchPatterns {
				if _, ok := livePatternSet[p]; ok {
					c.match("deployment_branch_pattern", p)
				} else {
					c.drift("deployment_branch_pattern", p, "(missing)")
				}
			}
			cfgPatternSet := make(map[string]struct{}, len(cfg.DeploymentBranchPatterns))
			for _, p := range cfg.DeploymentBranchPatterns {
				cfgPatternSet[p] = struct{}{}
			}
			for _, p := range live.DeploymentBranchPatterns {
				if _, ok := cfgPatternSet[p]; !ok {
					c.warn(fmt.Sprintf("deployment_branch_pattern: %s (in live repo, not in config; sync would remove it)", p))
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
					c.match("variable:"+v.Name, v.Value)
				} else {
					c.drift("variable:"+v.Name, v.Value, liveVal)
				}
			} else {
				c.drift("variable:"+v.Name, v.Value, "(missing)")
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
				c.match("secret:"+s.Name, "(exists)")
			} else {
				c.drift("secret:"+s.Name, "(exists)", "(missing)")
			}
		}
	}
}

func compareBoolSetting(field string, want *bool, got bool, c *auditCollector) {
	if want == nil {
		return
	}
	if *want == got {
		c.match(field, *want)
		return
	}
	c.drift(field, *want, got)
}

func compareStringSetting(field, want, got string, c *auditCollector) {
	if want == "" {
		return
	}
	if want == got {
		c.match(field, want)
		return
	}
	c.drift(field, want, got)
}

func auditActions(cfg, live *config.ActionsSettings, c *auditCollector) {
	c.setSection("Actions")

	if cfg == nil {
		c.info("(no actions permissions configured in manifest)")
		return
	}
	if live == nil {
		c.warn("(could not read live actions permissions)")
		return
	}

	if cfg.ShaPinningRequired != nil {
		liveVal := false
		if live.ShaPinningRequired != nil {
			liveVal = *live.ShaPinningRequired
		}
		if *cfg.ShaPinningRequired != liveVal {
			c.drift("sha_pinning_required", *cfg.ShaPinningRequired, liveVal)
		} else {
			c.match("sha_pinning_required", *cfg.ShaPinningRequired)
		}
	}

	if cfg.CanApprovePullRequestReviews != nil {
		liveVal := false
		if live.CanApprovePullRequestReviews != nil {
			liveVal = *live.CanApprovePullRequestReviews
		}
		if *cfg.CanApprovePullRequestReviews != liveVal {
			c.drift("can_approve_pull_request_reviews", *cfg.CanApprovePullRequestReviews, liveVal)
		} else {
			c.match("can_approve_pull_request_reviews", *cfg.CanApprovePullRequestReviews)
		}
	}

	if cfg.DefaultWorkflowPermissions != "" {
		liveVal := live.DefaultWorkflowPermissions
		if cfg.DefaultWorkflowPermissions != liveVal {
			c.drift("default_workflow_permissions", cfg.DefaultWorkflowPermissions, liveVal)
		} else {
			c.match("default_workflow_permissions", cfg.DefaultWorkflowPermissions)
		}
	}
}

func auditSecurity(cfg, live *config.SecuritySettings, c *auditCollector) {
	c.setSection("Security")

	if cfg == nil {
		c.info("(no security settings configured in manifest)")
		return
	}

	liveAlerts := false
	if live != nil && live.DependabotAlerts != nil {
		liveAlerts = *live.DependabotAlerts
	}
	compareBoolSetting("dependabot_alerts", cfg.DependabotAlerts, liveAlerts, c)

	var liveUpdates bool
	if live != nil && live.DependabotSecurityUpdates != nil {
		liveUpdates = *live.DependabotSecurityUpdates
	}
	compareBoolSetting("dependabot_security_updates", cfg.DependabotSecurityUpdates, liveUpdates, c)

	var liveScanning bool
	if live != nil && live.SecretScanning != nil {
		liveScanning = *live.SecretScanning
	}
	compareBoolSetting("secret_scanning", cfg.SecretScanning, liveScanning, c)

	var livePushProtection bool
	if live != nil && live.SecretScanningPushProtection != nil {
		livePushProtection = *live.SecretScanningPushProtection
	}
	compareBoolSetting("secret_scanning_push_protection", cfg.SecretScanningPushProtection, livePushProtection, c)

	var livePrivVulnRep bool
	if live != nil && live.PrivateVulnerabilityReporting != nil {
		livePrivVulnRep = *live.PrivateVulnerabilityReporting
	}
	compareBoolSetting("private_vulnerability_reporting", cfg.PrivateVulnerabilityReporting, livePrivVulnRep, c)

	var liveDependencyGraph bool
	if live != nil && live.DependencyGraph != nil {
		liveDependencyGraph = *live.DependencyGraph
	}
	compareBoolSetting("dependency_graph", cfg.DependencyGraph, liveDependencyGraph, c)
}

func auditRepoVarsSecrets(cfgVars []config.EnvironmentVariable, cfgSecrets []config.EnvironmentSecret, liveVars []config.EnvironmentVariable, liveSecrets []config.EnvironmentSecret, c *auditCollector) {
	c.setSection("Repository Variables")

	if len(cfgVars) == 0 {
		c.info("(no repository variables configured in manifest)")
	} else {
		liveVarMap := make(map[string]string, len(liveVars))
		for _, v := range liveVars {
			liveVarMap[v.Name] = v.Value
		}
		for _, v := range cfgVars {
			if liveVal, ok := liveVarMap[v.Name]; ok {
				if v.Value == liveVal {
					c.match("variable:"+v.Name, v.Value)
				} else {
					c.drift("variable:"+v.Name, v.Value, liveVal)
				}
			} else {
				c.drift("variable:"+v.Name, v.Value, "(missing)")
			}
		}
	}

	c.setSection("Repository Secrets")

	if len(cfgSecrets) == 0 {
		c.info("(no repository secrets configured in manifest)")
	} else {
		liveSecretSet := make(map[string]struct{}, len(liveSecrets))
		for _, s := range liveSecrets {
			liveSecretSet[s.Name] = struct{}{}
		}
		for _, s := range cfgSecrets {
			if _, ok := liveSecretSet[s.Name]; ok {
				c.match("secret:"+s.Name, "(exists)")
			} else {
				c.drift("secret:"+s.Name, "(exists)", "(missing)")
			}
		}
	}
}
