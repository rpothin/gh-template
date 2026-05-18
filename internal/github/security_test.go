package github

import (
	"testing"

	"github.com/rpothin/gh-template/internal/config"
)

// helper to build a securityAnalysisEntry with the given status string.
func saEntry(status string) *securityAnalysisEntry {
	return &securityAnalysisEntry{Status: status}
}

func TestRepoInfoToSecurity_NilSecurityAndAnalysis(t *testing.T) {
	info := &RepoInfo{SecurityAndAnalysis: nil}
	s := RepoInfoToSecurity(info, true, false)

	if s == nil {
		t.Fatal("RepoInfoToSecurity() = nil, want non-nil")
	}
	if s.DependabotAlerts == nil || *s.DependabotAlerts != true {
		t.Errorf("DependabotAlerts = %v, want true", s.DependabotAlerts)
	}
	if s.PrivateVulnerabilityReporting == nil || *s.PrivateVulnerabilityReporting != false {
		t.Errorf("PrivateVulnerabilityReporting = %v, want false", s.PrivateVulnerabilityReporting)
	}
	// GHAS fields should be nil when SecurityAndAnalysis is nil.
	if s.DependabotSecurityUpdates != nil {
		t.Errorf("DependabotSecurityUpdates = %v, want nil", s.DependabotSecurityUpdates)
	}
	if s.SecretScanning != nil {
		t.Errorf("SecretScanning = %v, want nil", s.SecretScanning)
	}
	if s.SecretScanningPushProtection != nil {
		t.Errorf("SecretScanningPushProtection = %v, want nil", s.SecretScanningPushProtection)
	}
	if s.DependencyGraph != nil {
		t.Errorf("DependencyGraph = %v, want nil", s.DependencyGraph)
	}
}

func TestRepoInfoToSecurity_FullSecurityAndAnalysis(t *testing.T) {
	info := &RepoInfo{
		SecurityAndAnalysis: &SecurityAnalysis{
			DependabotSecurityUpdates:    saEntry("enabled"),
			SecretScanning:               saEntry("enabled"),
			SecretScanningPushProtection: saEntry("disabled"),
			DependencyGraph:              saEntry("enabled"),
		},
	}
	s := RepoInfoToSecurity(info, true, true)

	if s == nil {
		t.Fatal("RepoInfoToSecurity() = nil, want non-nil")
	}
	checks := []struct {
		name string
		got  *bool
		want bool
	}{
		{"DependabotAlerts", s.DependabotAlerts, true},
		{"PrivateVulnerabilityReporting", s.PrivateVulnerabilityReporting, true},
		{"DependabotSecurityUpdates", s.DependabotSecurityUpdates, true},
		{"SecretScanning", s.SecretScanning, true},
		{"SecretScanningPushProtection", s.SecretScanningPushProtection, false},
		{"DependencyGraph", s.DependencyGraph, true},
	}
	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got == nil {
				t.Fatalf("%s = nil, want %v", tc.name, tc.want)
			}
			if *tc.got != tc.want {
				t.Errorf("%s = %v, want %v", tc.name, *tc.got, tc.want)
			}
		})
	}
}

func TestRepoInfoToSecurity_UnknownStatus_TreatedAsFalse(t *testing.T) {
	// Any status string other than "enabled" must resolve to false.
	info := &RepoInfo{
		SecurityAndAnalysis: &SecurityAnalysis{
			SecretScanning: saEntry("not_available"),
			DependencyGraph: saEntry(""),
		},
	}
	s := RepoInfoToSecurity(info, false, false)

	if s.SecretScanning == nil || *s.SecretScanning != false {
		t.Errorf("SecretScanning = %v for status 'not_available', want false", s.SecretScanning)
	}
	if s.DependencyGraph == nil || *s.DependencyGraph != false {
		t.Errorf("DependencyGraph = %v for empty status, want false", s.DependencyGraph)
	}
}

func TestRepoInfoToSecurity_NilIndividualSAEntries(t *testing.T) {
	// SecurityAndAnalysis present but individual entries nil → fields remain nil.
	info := &RepoInfo{
		SecurityAndAnalysis: &SecurityAnalysis{
			// all fields left nil
		},
	}
	s := RepoInfoToSecurity(info, false, false)

	if s.DependabotSecurityUpdates != nil {
		t.Errorf("DependabotSecurityUpdates = %v, want nil for absent entry", s.DependabotSecurityUpdates)
	}
	if s.SecretScanning != nil {
		t.Errorf("SecretScanning = %v, want nil for absent entry", s.SecretScanning)
	}
	if s.SecretScanningPushProtection != nil {
		t.Errorf("SecretScanningPushProtection = %v, want nil for absent entry", s.SecretScanningPushProtection)
	}
	if s.DependencyGraph != nil {
		t.Errorf("DependencyGraph = %v, want nil for absent entry", s.DependencyGraph)
	}
}

func TestRepoInfoToSecurity_RepoInfoToSettings_Roundtrip(t *testing.T) {
	info := &RepoInfo{
		HasWiki:             true,
		HasIssues:           true,
		HasProjects:         false,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: true,
		Visibility:          "public",
		Description:         "my desc",
	}
	s := RepoInfoToSettings(info)

	check := func(name string, got *bool, want bool) {
		t.Helper()
		if got == nil {
			t.Errorf("%s = nil, want %v", name, want)
			return
		}
		if *got != want {
			t.Errorf("%s = %v, want %v", name, *got, want)
		}
	}
	check("HasWiki", s.HasWiki, true)
	check("HasIssues", s.HasIssues, true)
	check("HasProjects", s.HasProjects, false)
	check("AllowSquashMerge", s.AllowSquashMerge, true)
	check("DeleteBranchOnMerge", s.DeleteBranchOnMerge, true)
	if s.Visibility != "public" {
		t.Errorf("Visibility = %q, want %q", s.Visibility, "public")
	}
	if s.Description != "my desc" {
		t.Errorf("Description = %q, want %q", s.Description, "my desc")
	}
}

func TestRepoUpdatePayload_IncludesOnlyNonNilFields(t *testing.T) {
	cfg := config.RepoSettings{
		HasWiki:   boolPtr(true),
		Visibility: "private",
	}
	p := repoUpdatePayload(cfg)

	if p["has_wiki"] != true {
		t.Errorf("has_wiki = %v, want true", p["has_wiki"])
	}
	if p["visibility"] != "private" {
		t.Errorf("visibility = %v, want private", p["visibility"])
	}
	if p["private"] != true {
		t.Errorf("private = %v, want true (mirrors visibility=private)", p["private"])
	}
	// Fields that were not set should be absent.
	for _, absent := range []string{"has_issues", "description", "allow_squash_merge"} {
		if _, ok := p[absent]; ok {
			t.Errorf("payload unexpectedly contains key %q", absent)
		}
	}
}

func TestRepoUpdatePayload_PublicSetsPrivateFalse(t *testing.T) {
	cfg := config.RepoSettings{Visibility: "public"}
	p := repoUpdatePayload(cfg)

	if p["visibility"] != "public" {
		t.Errorf("visibility = %v, want public", p["visibility"])
	}
	if p["private"] != false {
		t.Errorf("private = %v, want false for visibility=public", p["private"])
	}
}

func TestRepoUpdatePayload_Empty(t *testing.T) {
	p := repoUpdatePayload(config.RepoSettings{})
	if len(p) != 0 {
		t.Errorf("payload len = %d, want 0 for empty settings", len(p))
	}
}
