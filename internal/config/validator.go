package config

import (
	"errors"
	"fmt"
	"strings"
)

// ValidateManifest validates manifest fields with constrained values.
func ValidateManifest(m *Manifest) error {
	if m == nil {
		return errors.New("manifest is nil")
	}

	var errs []error
	if err := validateEnum("settings.visibility", m.Settings.Visibility, "public", "private", "internal"); err != nil {
		errs = append(errs, err)
	}
	if err := validateEnum("settings.pull_request_creation_policy", m.Settings.PullRequestCreationPolicy, "collaborators_only", "contributors_only", "all_users", "any_user"); err != nil {
		errs = append(errs, err)
	}
	if m.Actions != nil {
		if err := validateEnum("actions.default_workflow_permissions", m.Actions.DefaultWorkflowPermissions, "read", "write"); err != nil {
			errs = append(errs, err)
		}
	}
	for i, env := range m.Environments {
		field := fmt.Sprintf("environments[%d].deployment_branch_policy", i)
		if err := validateEnum(field, env.DeploymentBranchPolicy, "all", "protected", "selected"); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func validateEnum(field, value string, allowed ...string) error {
	if value == "" {
		return nil
	}
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of %s, got %q", field, quoteValues(allowed), value)
}

func quoteValues(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = fmt.Sprintf("%q", value)
	}
	return strings.Join(quoted, ", ")
}
