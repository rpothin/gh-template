package config

// SecretPlaceholder is the literal string stored as the value for environment secrets
// in the manifest. It signals that the actual secret must be set manually after
// repository creation.
const SecretPlaceholder = "PLACEHOLDER"

// Manifest represents the structure of a template-metadata.yml file.
type Manifest struct {
	Template     string               `yaml:"template,omitempty"`
	Settings     RepoSettings         `yaml:"settings"`
	Topics       []string             `yaml:"topics,omitempty"`
	Environments []Environment        `yaml:"environments,omitempty"`
	Variables    []EnvironmentVariable `yaml:"variables,omitempty"`
	Secrets      []EnvironmentSecret   `yaml:"secrets,omitempty"`
	Actions      *ActionsSettings     `yaml:"actions,omitempty"`
	Security     *SecuritySettings    `yaml:"security,omitempty"`
}

// ActionsSettings maps to GitHub Actions repository permission settings.
// These are stored in two separate API endpoints (actions/permissions and
// actions/permissions/workflow) but presented as a single manifest section.
type ActionsSettings struct {
	CanApprovePullRequestReviews *bool  `yaml:"can_approve_pull_request_reviews,omitempty"`
	ShaPinningRequired           *bool  `yaml:"sha_pinning_required,omitempty"`
	DefaultWorkflowPermissions   string `yaml:"default_workflow_permissions,omitempty"`
}

// SecuritySettings maps to GitHub repository security analysis settings.
// Dependabot alerts use a dedicated /vulnerability-alerts endpoint;
// the remaining fields use the security_and_analysis object on PATCH /repos.
// Note: secret_scanning and secret_scanning_push_protection require the
// repository to be public or the account to have GitHub Advanced Security.
type SecuritySettings struct {
	DependabotAlerts             *bool `yaml:"dependabot_alerts,omitempty"`
	DependabotSecurityUpdates    *bool `yaml:"dependabot_security_updates,omitempty"`
	SecretScanning               *bool `yaml:"secret_scanning,omitempty"`
	SecretScanningPushProtection *bool `yaml:"secret_scanning_push_protection,omitempty"`
}

// RepoSettings maps to GitHub repository settings API fields.
// Pointer types allow distinguishing "not configured" (nil) from explicit false.
type RepoSettings struct {
	HasWiki                    *bool  `yaml:"has_wiki,omitempty"`
	HasIssues                  *bool  `yaml:"has_issues,omitempty"`
	HasProjects                *bool  `yaml:"has_projects,omitempty"`
	HasDiscussions             *bool  `yaml:"has_discussions,omitempty"`
	HasPullRequests            *bool  `yaml:"has_pull_requests,omitempty"`
	PullRequestCreationPolicy  string `yaml:"pull_request_creation_policy,omitempty"`
	AllowSquashMerge           *bool  `yaml:"allow_squash_merge,omitempty"`
	AllowMergeCommit           *bool  `yaml:"allow_merge_commit,omitempty"`
	AllowRebaseMerge           *bool  `yaml:"allow_rebase_merge,omitempty"`
	AllowAutoMerge             *bool  `yaml:"allow_auto_merge,omitempty"`
	DeleteBranchOnMerge        *bool  `yaml:"delete_branch_on_merge,omitempty"`
	AllowUpdateBranch          *bool  `yaml:"allow_update_branch,omitempty"`
	Visibility                 string `yaml:"visibility,omitempty"`
	Description                string `yaml:"description,omitempty"`
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
// The Value field is intentionally omitted from YAML output (secret values
// cannot be read from GitHub). When applied via create/sync, any missing
// secret is initialized with SecretPlaceholder.
type EnvironmentSecret struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value,omitempty"`
}
