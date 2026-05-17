package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestRejectsUnknownKeys(t *testing.T) {
	path := writeManifestFile(t, "settings:\n  has_wki: true\n")

	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("LoadManifest() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "field has_wki not found") {
		t.Fatalf("LoadManifest() error = %q, want unknown field details", err)
	}
}

func TestLoadManifestReportsAllValidationFailures(t *testing.T) {
	path := writeManifestFile(t, `settings:
  visibility: secret
  pull_request_creation_policy: everyone
actions:
  default_workflow_permissions: admin
environments:
  - name: production
    deployment_branch_policy: custom
`)

	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("LoadManifest() error = nil, want validation error")
	}

	for _, want := range []string{
		`settings.visibility must be one of "public", "private", "internal", got "secret"`,
		`settings.pull_request_creation_policy must be one of "collaborators_only", "contributors_only", "all_users", "any_user", got "everyone"`,
		`actions.default_workflow_permissions must be one of "read", "write", got "admin"`,
		`environments[0].deployment_branch_policy must be one of "all", "protected", "selected", got "custom"`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("LoadManifest() error = %q, want substring %q", err, want)
		}
	}
}

func writeManifestFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "template-metadata.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	return path
}

func TestLoadManifest_NewSecurityFields_RoundTrip(t *testing.T) {
	yaml := `security:
  dependabot_alerts: true
  dependabot_security_updates: false
  secret_scanning: true
  secret_scanning_push_protection: true
  private_vulnerability_reporting: true
  dependency_graph: false
`
	path := writeManifestFile(t, yaml)

	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if m.Security == nil {
		t.Fatal("Security is nil, want non-nil")
	}

	tests := []struct {
		name string
		got  *bool
		want bool
	}{
		{"private_vulnerability_reporting", m.Security.PrivateVulnerabilityReporting, true},
		{"dependency_graph", m.Security.DependencyGraph, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got == nil {
				t.Fatalf("%s = nil, want %v", tt.name, tt.want)
			}
			if *tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, *tt.got, tt.want)
			}
		})
	}
}

func TestLoadManifest_UnknownSecurityField_Rejected(t *testing.T) {
	path := writeManifestFile(t, "security:\n  unknown_field: true\n")

	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("LoadManifest() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "unknown_field") {
		t.Fatalf("LoadManifest() error = %q, want unknown field details", err)
	}
}
