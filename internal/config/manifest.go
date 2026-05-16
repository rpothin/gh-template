package config

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
	HasDownloads        *bool  `yaml:"has_downloads,omitempty"`
	AllowSquashMerge    *bool  `yaml:"allow_squash_merge,omitempty"`
	AllowMergeCommit    *bool  `yaml:"allow_merge_commit,omitempty"`
	AllowRebaseMerge    *bool  `yaml:"allow_rebase_merge,omitempty"`
	AllowAutoMerge      *bool  `yaml:"allow_auto_merge,omitempty"`
	DeleteBranchOnMerge *bool  `yaml:"delete_branch_on_merge,omitempty"`
	AllowUpdateBranch   *bool  `yaml:"allow_update_branch,omitempty"`
	IsTemplate          *bool  `yaml:"is_template,omitempty"`
	Visibility          string `yaml:"visibility,omitempty"`
	Description         string `yaml:"description,omitempty"`
}

// Environment represents a GitHub Actions deployment environment.
type Environment struct {
	Name      string `yaml:"name"`
	WaitTimer int    `yaml:"wait_timer"`
}
