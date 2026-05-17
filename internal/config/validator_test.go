package config

import (
	"strings"
	"testing"
)

func TestValidateManifest_Nil(t *testing.T) {
	err := ValidateManifest(nil)
	if err == nil {
		t.Fatal("ValidateManifest(nil) error = nil, want error")
	}
}

func TestValidateManifest_Empty_Valid(t *testing.T) {
	m := &Manifest{}
	if err := ValidateManifest(m); err != nil {
		t.Fatalf("ValidateManifest(&Manifest{}) error = %v, want nil", err)
	}
}

func TestValidateManifest_InvalidVisibility(t *testing.T) {
	m := &Manifest{Settings: RepoSettings{Visibility: "secret"}}
	err := ValidateManifest(m)
	if err == nil {
		t.Fatal("ValidateManifest() error = nil, want error for invalid visibility")
	}
	if !strings.Contains(err.Error(), `settings.visibility`) {
		t.Errorf("error = %q, want mention of settings.visibility", err)
	}
}

func TestValidateManifest_ValidVisibilities(t *testing.T) {
	for _, vis := range []string{"public", "private", "internal"} {
		m := &Manifest{Settings: RepoSettings{Visibility: vis}}
		if err := ValidateManifest(m); err != nil {
			t.Errorf("visibility %q: unexpected error %v", vis, err)
		}
	}
}

func TestValidateManifest_InvalidPullRequestPolicy(t *testing.T) {
	m := &Manifest{Settings: RepoSettings{PullRequestCreationPolicy: "everyone"}}
	err := ValidateManifest(m)
	if err == nil {
		t.Fatal("ValidateManifest() error = nil, want error for invalid pull_request_creation_policy")
	}
	if !strings.Contains(err.Error(), `settings.pull_request_creation_policy`) {
		t.Errorf("error = %q, want mention of settings.pull_request_creation_policy", err)
	}
}

func TestValidateManifest_InvalidActionsPermissions(t *testing.T) {
	m := &Manifest{
		Actions: &ActionsSettings{DefaultWorkflowPermissions: "admin"},
	}
	err := ValidateManifest(m)
	if err == nil {
		t.Fatal("ValidateManifest() error = nil, want error for invalid actions.default_workflow_permissions")
	}
	if !strings.Contains(err.Error(), `actions.default_workflow_permissions`) {
		t.Errorf("error = %q, want mention of actions.default_workflow_permissions", err)
	}
}

func TestValidateManifest_InvalidEnvironmentBranchPolicy(t *testing.T) {
	m := &Manifest{
		Environments: []Environment{
			{Name: "prod", DeploymentBranchPolicy: "custom"},
		},
	}
	err := ValidateManifest(m)
	if err == nil {
		t.Fatal("ValidateManifest() error = nil, want error for invalid deployment_branch_policy")
	}
	if !strings.Contains(err.Error(), `environments[0].deployment_branch_policy`) {
		t.Errorf("error = %q, want mention of environments[0].deployment_branch_policy", err)
	}
}

func TestValidateManifest_MultipleErrors_AllReported(t *testing.T) {
	m := &Manifest{
		Settings: RepoSettings{
			Visibility:                "bad-visibility",
			PullRequestCreationPolicy: "bad-policy",
		},
		Actions: &ActionsSettings{DefaultWorkflowPermissions: "admin"},
		Environments: []Environment{
			{Name: "env1", DeploymentBranchPolicy: "bad-policy"},
		},
	}
	err := ValidateManifest(m)
	if err == nil {
		t.Fatal("ValidateManifest() error = nil, want multi-error")
	}
	for _, want := range []string{
		"settings.visibility",
		"settings.pull_request_creation_policy",
		"actions.default_workflow_permissions",
		"environments[0].deployment_branch_policy",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want mention of %q", err, want)
		}
	}
}

func TestValidateManifest_ValidActionsPermissions(t *testing.T) {
	for _, perm := range []string{"read", "write"} {
		m := &Manifest{Actions: &ActionsSettings{DefaultWorkflowPermissions: perm}}
		if err := ValidateManifest(m); err != nil {
			t.Errorf("default_workflow_permissions %q: unexpected error %v", perm, err)
		}
	}
}

func TestValidateManifest_ValidDeploymentBranchPolicies(t *testing.T) {
	for _, policy := range []string{"all", "protected", "selected"} {
		m := &Manifest{
			Environments: []Environment{
				{Name: "env", DeploymentBranchPolicy: policy},
			},
		}
		if err := ValidateManifest(m); err != nil {
			t.Errorf("deployment_branch_policy %q: unexpected error %v", policy, err)
		}
	}
}
