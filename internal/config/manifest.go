package config

// SecretPlaceholder is the literal string stored as the value for environment secrets
// in the manifest. It signals that the actual secret must be set manually after
// repository creation.
const SecretPlaceholder = "PLACEHOLDER"

// Manifest represents the structure of a template-metadata.yml file.
type Manifest struct {
	Settings     RepoSettings  `yaml:"settings"`
	Topics       []string      `yaml:"topics,omitempty"`
	Environments []Environment `yaml:"environments,omitempty"`
}

// RepoSettings maps to GitHub repository settings API fields.
// Pointer types allow distinguishing "not configured" (nil) from explicit false.
type RepoSettings struct {
	HasWiki             *bool  `yaml:"has_wiki,omitempty"`
	HasIssues           *bool  `yaml:"has_issues,omitempty"`
	HasProjects         *bool  `yaml:"has_projects,omitempty"`
	AllowSquashMerge    *bool  `yaml:"allow_squash_merge,omitempty"`
	AllowMergeCommit    *bool  `yaml:"allow_merge_commit,omitempty"`
	AllowRebaseMerge    *bool  `yaml:"allow_rebase_merge,omitempty"`
	AllowAutoMerge      *bool  `yaml:"allow_auto_merge,omitempty"`
	DeleteBranchOnMerge *bool  `yaml:"delete_branch_on_merge,omitempty"`
	AllowUpdateBranch   *bool  `yaml:"allow_update_branch,omitempty"`
	Visibility          string `yaml:"visibility,omitempty"`
	Description         string `yaml:"description,omitempty"`
}

// Environment represents a GitHub Actions deployment environment.
type Environment struct {
	Name                     string                `yaml:"name"`
	WaitTimer                int                   `yaml:"wait_timer,omitempty"`
	PreventSelfReview        *bool                 `yaml:"prevent_self_review,omitempty"`
	Reviewers                []string              `yaml:"reviewers,omitempty"`
	DeploymentBranchPolicy   string                `yaml:"deployment_branch_policy,omitempty"`
	DeploymentBranchPatterns []string              `yaml:"deployment_branch_patterns,omitempty"`
	Variables                []EnvironmentVariable `yaml:"variables,omitempty"`
	Secrets                  []EnvironmentSecret   `yaml:"secrets,omitempty"`
}

// EnvironmentVariable represents a key/value variable available in GitHub Actions
// for a deployment environment.
type EnvironmentVariable struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// EnvironmentSecret represents a named secret in a deployment environment.
// When captured via snapshot, Value is always SecretPlaceholder.
// When applied via create/sync, a missing secret is initialized with SecretPlaceholder.
type EnvironmentSecret struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
