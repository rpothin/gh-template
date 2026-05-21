package cmd

import (
	"testing"

	"github.com/rpothin/gh-template/internal/config"
	ghapi "github.com/rpothin/gh-template/internal/github"
)

// --- auditTopics seed-only tests ---

func TestAuditTopics_SeedOnly_NoReport(t *testing.T) {
	// When both the manifest and the live repo have topics, seed-only applies:
	// no drift and no warnings should be reported.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditTopics([]string{"go", "cli"}, []string{"custom-topic"}, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (seed-only should suppress drift)", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 0 {
		t.Errorf("Warnings = %d, want 0 (seed-only should suppress warnings)", len(c.Report.Warnings))
	}
}

func TestAuditTopics_NeedsSeeding_ReportsDrift(t *testing.T) {
	// When the manifest has topics but the live repo has none, drift should be
	// reported for each missing topic.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditTopics([]string{"go", "cli"}, []string{}, c)

	if c.Report.DriftCount != 2 {
		t.Errorf("DriftCount = %d, want 2", c.Report.DriftCount)
	}
}

func TestAuditTopics_EmptyConfig_ExtraTopicsWarn(t *testing.T) {
	// When the manifest has no topics but the live repo does, warn about extras
	// (existing behaviour — no seed-only guard applies here).
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditTopics([]string{}, []string{"custom-topic"}, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 1 {
		t.Errorf("Warnings = %d, want 1", len(c.Report.Warnings))
	}
}

func TestAuditTopics_BothEmpty_NoDrift(t *testing.T) {
	// Both config and live are empty: nothing to do.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditTopics([]string{}, []string{}, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

// --- auditSettings description seed-only tests ---

func TestAuditSettings_Description_SeedOnly_NoDrift(t *testing.T) {
	// Manifest has a description AND live repo already has one: no drift.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	settings := config.RepoSettings{Description: "Template description"}
	live := &ghapi.RepoInfo{Description: "Owner's custom description"}
	auditSettings(settings, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (seed-only should suppress drift)", c.Report.DriftCount)
	}
}

func TestAuditSettings_Description_NeedsSeeding_ReportsDrift(t *testing.T) {
	// Manifest has a description but live repo has none: drift reported so
	// the operator knows seeding is required.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	settings := config.RepoSettings{Description: "Template description"}
	live := &ghapi.RepoInfo{Description: ""}
	auditSettings(settings, live, c)

	if c.Report.DriftCount != 1 {
		t.Errorf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
}

func TestAuditSettings_Description_ManifestEmpty_NoDrift(t *testing.T) {
	// Manifest has no description: nothing to check regardless of live value.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	settings := config.RepoSettings{Description: ""}
	live := &ghapi.RepoInfo{Description: "Whatever the owner set"}
	auditSettings(settings, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

// --- auditSecurity tests ---

func boolPtr(b bool) *bool { return &b }

func TestAuditSecurity_NilConfig_NoReport(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditSecurity(nil, &config.SecuritySettings{}, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditSecurity_AllMatch_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.SecuritySettings{
		DependabotAlerts:              boolPtr(true),
		DependabotSecurityUpdates:     boolPtr(true),
		SecretScanning:                boolPtr(true),
		SecretScanningPushProtection:  boolPtr(true),
		PrivateVulnerabilityReporting: boolPtr(false),
		DependencyGraph:               boolPtr(true),
	}
	live := &config.SecuritySettings{
		DependabotAlerts:              boolPtr(true),
		DependabotSecurityUpdates:     boolPtr(true),
		SecretScanning:                boolPtr(true),
		SecretScanningPushProtection:  boolPtr(true),
		PrivateVulnerabilityReporting: boolPtr(false),
		DependencyGraph:               boolPtr(true),
	}
	auditSecurity(cfg, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditSecurity_PrivateVulnReporting_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.SecuritySettings{
		PrivateVulnerabilityReporting: boolPtr(true),
	}
	live := &config.SecuritySettings{
		PrivateVulnerabilityReporting: boolPtr(false),
	}
	auditSecurity(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Errorf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
}

func TestAuditSecurity_PrivateVulnReporting_NilLive_Drift(t *testing.T) {
	// If live is nil, treat private vulnerability reporting as false.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.SecuritySettings{
		PrivateVulnerabilityReporting: boolPtr(true),
	}
	auditSecurity(cfg, nil, c)

	if c.Report.DriftCount != 1 {
		t.Errorf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
}

func TestAuditSecurity_DependencyGraph_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.SecuritySettings{
		DependencyGraph: boolPtr(true),
	}
	live := &config.SecuritySettings{
		DependencyGraph: boolPtr(false),
	}
	auditSecurity(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Errorf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
}

func TestAuditSecurity_NilCfgFields_NoDrift(t *testing.T) {
	// Nil config fields means "not configured" — no drift should be recorded
	// regardless of live state.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.SecuritySettings{
		// PrivateVulnerabilityReporting and DependencyGraph are nil (not configured)
	}
	live := &config.SecuritySettings{
		PrivateVulnerabilityReporting: boolPtr(true),
		DependencyGraph:               boolPtr(true),
	}
	auditSecurity(cfg, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

// --- compareBoolSetting tests ---

func TestCompareBoolSetting_NilWant_Skipped(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	compareBoolSetting("has_wiki", nil, true, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (nil want skips comparison)", c.Report.DriftCount)
	}
}

func TestCompareBoolSetting_Match_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	want := true
	compareBoolSetting("has_wiki", &want, true, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestCompareBoolSetting_Mismatch_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	want := true
	compareBoolSetting("has_wiki", &want, false, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
	item := c.Report.Drifts[0]
	if item.Field != "has_wiki" {
		t.Errorf("Field = %q, want has_wiki", item.Field)
	}
	if item.Want != "true" {
		t.Errorf("Want = %q, want true", item.Want)
	}
	if item.Got != "false" {
		t.Errorf("Got = %q, want false", item.Got)
	}
}

// --- compareStringSetting tests ---

func TestCompareStringSetting_EmptyWant_Skipped(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	compareStringSetting("visibility", "", "public", c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (empty want skips comparison)", c.Report.DriftCount)
	}
}

func TestCompareStringSetting_Match_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	compareStringSetting("visibility", "public", "public", c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestCompareStringSetting_Mismatch_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	compareStringSetting("visibility", "public", "private", c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
	item := c.Report.Drifts[0]
	if item.Field != "visibility" {
		t.Errorf("Field = %q, want visibility", item.Field)
	}
	if item.Want != "public" {
		t.Errorf("Want = %q, want public", item.Want)
	}
	if item.Got != "private" {
		t.Errorf("Got = %q, want private", item.Got)
	}
}

// --- auditActions tests ---

func TestAuditActions_NilCfg_NoReport(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditActions(nil, &config.ActionsSettings{}, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 0 {
		t.Errorf("Warnings = %d, want 0", len(c.Report.Warnings))
	}
}

func TestAuditActions_NilLive_Warning(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditActions(&config.ActionsSettings{}, nil, c)

	if len(c.Report.Warnings) != 1 {
		t.Errorf("Warnings = %d, want 1 (could not read live actions permissions)", len(c.Report.Warnings))
	}
	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditActions_AllMatch_NoDrift(t *testing.T) {
	sha := true
	approve := false
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.ActionsSettings{
		ShaPinningRequired:           &sha,
		CanApprovePullRequestReviews: &approve,
		DefaultWorkflowPermissions:   "read",
	}
	live := &config.ActionsSettings{
		ShaPinningRequired:           &sha,
		CanApprovePullRequestReviews: &approve,
		DefaultWorkflowPermissions:   "read",
	}
	auditActions(cfg, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditActions_ShaPinning_Drift(t *testing.T) {
	want := true
	got := false
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.ActionsSettings{ShaPinningRequired: &want}
	live := &config.ActionsSettings{ShaPinningRequired: &got}
	auditActions(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
	item := c.Report.Drifts[0]
	if item.Field != "sha_pinning_required" {
		t.Errorf("Field = %q, want sha_pinning_required", item.Field)
	}
}

func TestAuditActions_DefaultWorkflowPermissions_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.ActionsSettings{DefaultWorkflowPermissions: "write"}
	live := &config.ActionsSettings{DefaultWorkflowPermissions: "read"}
	auditActions(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
	item := c.Report.Drifts[0]
	if item.Field != "default_workflow_permissions" {
		t.Errorf("Field = %q, want default_workflow_permissions", item.Field)
	}
	if item.Want != "write" {
		t.Errorf("Want = %q, want write", item.Want)
	}
	if item.Got != "read" {
		t.Errorf("Got = %q, want read", item.Got)
	}
}

func TestAuditActions_CanApprove_NilLiveDefaultsFalse(t *testing.T) {
	// When cfg sets a value but live has nil, live defaults to false.
	val := true
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := &config.ActionsSettings{CanApprovePullRequestReviews: &val}
	live := &config.ActionsSettings{} // nil field
	auditActions(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (cfg=true, live defaults to false)", c.Report.DriftCount)
	}
}

// --- auditEnvironments tests ---

func TestAuditEnvironments_BothEmpty_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditEnvironments([]config.Environment{}, []config.Environment{}, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditEnvironments_ConfigEnvMissingFromLive_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfgEnvs := []config.Environment{{Name: "production"}}
	auditEnvironments(cfgEnvs, []config.Environment{}, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
	if c.Report.Drifts[0].Field != "production" {
		t.Errorf("Drifts[0].Field = %q, want production", c.Report.Drifts[0].Field)
	}
}

func TestAuditEnvironments_ExtraLiveEnv_Warning(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	liveEnvs := []config.Environment{{Name: "staging"}}
	auditEnvironments([]config.Environment{}, liveEnvs, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 1 {
		t.Fatalf("Warnings = %d, want 1", len(c.Report.Warnings))
	}
	if c.Report.Warnings[0].Section != "Environments" {
		t.Errorf("Warning.Section = %q, want Environments", c.Report.Warnings[0].Section)
	}
}

// --- auditEnvFields tests ---

func TestAuditEnvFields_WaitTimer_Match_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{Name: "prod", WaitTimer: 10}
	live := config.Environment{Name: "prod", WaitTimer: 10}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditEnvFields_WaitTimer_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{Name: "prod", WaitTimer: 30}
	live := config.Environment{Name: "prod", WaitTimer: 0}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
	if c.Report.Drifts[0].Field != "wait_timer" {
		t.Errorf("Field = %q, want wait_timer", c.Report.Drifts[0].Field)
	}
}

func TestAuditEnvFields_PreventSelfReview_Drift(t *testing.T) {
	want := true
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{Name: "prod", PreventSelfReview: &want}
	live := config.Environment{Name: "prod"} // nil → false
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
	if c.Report.Drifts[0].Field != "prevent_self_review" {
		t.Errorf("Field = %q, want prevent_self_review", c.Report.Drifts[0].Field)
	}
}

func TestAuditEnvFields_Reviewers_Missing_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{Name: "prod", Reviewers: []string{"alice", "bob"}}
	live := config.Environment{Name: "prod", Reviewers: []string{"alice"}} // bob missing
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (bob missing)", c.Report.DriftCount)
	}
	if c.Report.Drifts[0].Field != "reviewer" {
		t.Errorf("Field = %q, want reviewer", c.Report.Drifts[0].Field)
	}
	if c.Report.Drifts[0].Want != "bob" {
		t.Errorf("Want = %q, want bob", c.Report.Drifts[0].Want)
	}
}

func TestAuditEnvFields_Reviewers_ExtraLive_Warning(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{Name: "prod", Reviewers: []string{"alice"}}
	live := config.Environment{Name: "prod", Reviewers: []string{"alice", "extra-reviewer"}}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 1 {
		t.Errorf("Warnings = %d, want 1 for extra live reviewer", len(c.Report.Warnings))
	}
}

func TestAuditEnvFields_DeploymentBranchPolicy_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{Name: "prod", DeploymentBranchPolicy: "protected"}
	live := config.Environment{Name: "prod", DeploymentBranchPolicy: "all"}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1", c.Report.DriftCount)
	}
	if c.Report.Drifts[0].Field != "deployment_branch_policy" {
		t.Errorf("Field = %q, want deployment_branch_policy", c.Report.Drifts[0].Field)
	}
}

func TestAuditEnvFields_DeploymentBranchPolicy_EmptyLiveNormalized(t *testing.T) {
	// An empty live deployment_branch_policy is normalised to "all".
	// So cfg="all" with live="" should NOT report drift.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{Name: "prod", DeploymentBranchPolicy: "all"}
	live := config.Environment{Name: "prod", DeploymentBranchPolicy: ""}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (empty live normalised to 'all')", c.Report.DriftCount)
	}
}

func TestAuditEnvFields_SelectedBranchPolicy_MissingPattern_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{
		Name:                     "prod",
		DeploymentBranchPolicy:   "selected",
		DeploymentBranchPatterns: []string{"main", "release/*"},
	}
	live := config.Environment{
		Name:                     "prod",
		DeploymentBranchPolicy:   "selected",
		DeploymentBranchPatterns: []string{"main"}, // release/* missing
	}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	// Expect: deployment_branch_policy matches (no drift), but release/* pattern is missing (1 drift).
	hasPatternDrift := false
	for _, d := range c.Report.Drifts {
		if d.Field == "deployment_branch_pattern" && d.Want == "release/*" {
			hasPatternDrift = true
		}
	}
	if !hasPatternDrift {
		t.Errorf("expected a drift for deployment_branch_pattern 'release/*', got drifts: %+v", c.Report.Drifts)
	}
}

func TestAuditEnvFields_SelectedBranchPolicy_ExtraLivePattern_Warning(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{
		Name:                     "prod",
		DeploymentBranchPolicy:   "selected",
		DeploymentBranchPatterns: []string{"main"},
	}
	live := config.Environment{
		Name:                     "prod",
		DeploymentBranchPolicy:   "selected",
		DeploymentBranchPatterns: []string{"main", "hotfix/*"},
	}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (extra live pattern only warns)", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 1 {
		t.Errorf("Warnings = %d, want 1 for extra live branch pattern", len(c.Report.Warnings))
	}
}

func TestAuditEnvFields_Variables_MatchAndDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{
		Name: "prod",
		Variables: []config.EnvironmentVariable{
			{Name: "APP_ENV", Value: "production"},
			{Name: "LOG_LEVEL", Value: "info"},
		},
	}
	live := config.Environment{
		Name: "prod",
		Variables: []config.EnvironmentVariable{
			{Name: "APP_ENV", Value: "production"},
			{Name: "LOG_LEVEL", Value: "debug"}, // drifted
		},
	}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (LOG_LEVEL mismatch)", c.Report.DriftCount)
	}
	if c.Report.Drifts[0].Field != "variable:LOG_LEVEL" {
		t.Errorf("Field = %q, want variable:LOG_LEVEL", c.Report.Drifts[0].Field)
	}
}

func TestAuditEnvFields_Secrets_Missing_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfg := config.Environment{
		Name:    "prod",
		Secrets: []config.EnvironmentSecret{{Name: "API_KEY"}, {Name: "DB_PASS"}},
	}
	live := config.Environment{
		Name:    "prod",
		Secrets: []config.EnvironmentSecret{{Name: "API_KEY"}}, // DB_PASS missing
	}
	c.enterSubSection("prod", "environments/prod")
	auditEnvFields(cfg, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (DB_PASS missing)", c.Report.DriftCount)
	}
	if c.Report.Drifts[0].Field != "secret:DB_PASS" {
		t.Errorf("Field = %q, want secret:DB_PASS", c.Report.Drifts[0].Field)
	}
}

// --- auditRepoVarsSecrets tests ---

func TestAuditRepoVarsSecrets_EmptyConfig_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditRepoVarsSecrets(nil, nil, nil, nil, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditRepoVarsSecrets_Variables_Match_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	vars := []config.EnvironmentVariable{{Name: "FOO", Value: "bar"}}
	auditRepoVarsSecrets(vars, nil, vars, nil, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditRepoVarsSecrets_Variables_Mismatch_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfgVars := []config.EnvironmentVariable{{Name: "ENV", Value: "prod"}}
	liveVars := []config.EnvironmentVariable{{Name: "ENV", Value: "staging"}}
	auditRepoVarsSecrets(cfgVars, nil, liveVars, nil, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (ENV value mismatch)", c.Report.DriftCount)
	}
	item := c.Report.Drifts[0]
	if item.Field != "variable:ENV" {
		t.Errorf("Field = %q, want variable:ENV", item.Field)
	}
	if item.Want != "prod" {
		t.Errorf("Want = %q, want prod", item.Want)
	}
	if item.Got != "staging" {
		t.Errorf("Got = %q, want staging", item.Got)
	}
}

func TestAuditRepoVarsSecrets_Variables_Missing_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfgVars := []config.EnvironmentVariable{{Name: "MISSING_VAR", Value: "x"}}
	auditRepoVarsSecrets(cfgVars, nil, nil, nil, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (variable absent from live)", c.Report.DriftCount)
	}
	if c.Report.Drifts[0].Got != "(missing)" {
		t.Errorf("Got = %q, want (missing)", c.Report.Drifts[0].Got)
	}
}

func TestAuditRepoVarsSecrets_Secrets_Present_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	secrets := []config.EnvironmentSecret{{Name: "TOKEN"}}
	auditRepoVarsSecrets(nil, secrets, nil, secrets, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (secret exists in live)", c.Report.DriftCount)
	}
}

func TestAuditRepoVarsSecrets_Secrets_Missing_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	cfgSecrets := []config.EnvironmentSecret{{Name: "DEPLOY_KEY"}}
	auditRepoVarsSecrets(nil, cfgSecrets, nil, nil, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (DEPLOY_KEY missing)", c.Report.DriftCount)
	}
	item := c.Report.Drifts[0]
	if item.Field != "secret:DEPLOY_KEY" {
		t.Errorf("Field = %q, want secret:DEPLOY_KEY", item.Field)
	}
	if item.Got != "(missing)" {
		t.Errorf("Got = %q, want (missing)", item.Got)
	}
}

// --- auditSettings bool field tests ---

func TestAuditSettings_BoolField_Match_NoDrift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	settings := config.RepoSettings{HasWiki: boolPtr(true)}
	live := &ghapi.RepoInfo{HasWiki: true}
	auditSettings(settings, live, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
}

func TestAuditSettings_BoolField_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	settings := config.RepoSettings{HasWiki: boolPtr(false)}
	live := &ghapi.RepoInfo{HasWiki: true}
	auditSettings(settings, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (has_wiki mismatch)", c.Report.DriftCount)
	}
	item := c.Report.Drifts[0]
	if item.Field != "has_wiki" {
		t.Errorf("Field = %q, want has_wiki", item.Field)
	}
}

func TestAuditSettings_Visibility_Drift(t *testing.T) {
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	settings := config.RepoSettings{Visibility: "public"}
	live := &ghapi.RepoInfo{Visibility: "private"}
	auditSettings(settings, live, c)

	if c.Report.DriftCount != 1 {
		t.Fatalf("DriftCount = %d, want 1 (visibility mismatch)", c.Report.DriftCount)
	}
	item := c.Report.Drifts[0]
	if item.Field != "visibility" {
		t.Errorf("Field = %q, want visibility", item.Field)
	}
	if item.Want != "public" {
		t.Errorf("Want = %q, want public", item.Want)
	}
	if item.Got != "private" {
		t.Errorf("Got = %q, want private", item.Got)
	}
}

// --- auditCommonFiles unit tests (no HTTP — tests pure branching logic) ---

func TestAuditCommonFiles_NoConfig_InfoOnly(t *testing.T) {
	// When common_files is empty, no drift and no warnings are recorded.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditCommonFiles([]string{}, "owner/template", "owner", "repo", nil, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (no common_files means nothing to check)", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 0 {
		t.Errorf("Warnings = %d, want 0", len(c.Report.Warnings))
	}
}

func TestAuditCommonFiles_MissingTemplate_Warn(t *testing.T) {
	// When common_files is set but template is empty, a warning must be recorded.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditCommonFiles([]string{"AGENTS.md"}, "", "owner", "repo", nil, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0 (no template means we can't compare)", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 1 {
		t.Fatalf("Warnings = %d, want 1", len(c.Report.Warnings))
	}
	if c.Report.Warnings[0].Section != "Common Files" {
		t.Errorf("Section = %q, want Common Files", c.Report.Warnings[0].Section)
	}
}

func TestAuditCommonFiles_InvalidTemplate_Warn(t *testing.T) {
	// A malformed template slug (missing slash) should produce a warning.
	c := newAuditCollector(false, "owner/repo", "manifest.yml")
	auditCommonFiles([]string{"AGENTS.md"}, "notavalidslug", "owner", "repo", nil, c)

	if c.Report.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", c.Report.DriftCount)
	}
	if len(c.Report.Warnings) != 1 {
		t.Fatalf("Warnings = %d, want 1", len(c.Report.Warnings))
	}
}

// --- shortSHA tests ---

func TestShortSHA_LongSHA_Truncated(t *testing.T) {
	sha := "abc123def456789012345678901234567890"
	got := shortSHA(sha)
	if got != "abc123de" {
		t.Errorf("shortSHA(%q) = %q, want abc123de", sha, got)
	}
}

func TestShortSHA_ShortSHA_Unchanged(t *testing.T) {
	sha := "abc12"
	got := shortSHA(sha)
	if got != sha {
		t.Errorf("shortSHA(%q) = %q, want %q", sha, got, sha)
	}
}
