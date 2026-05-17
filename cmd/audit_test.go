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
