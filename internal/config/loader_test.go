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
